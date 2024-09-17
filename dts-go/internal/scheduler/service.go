package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/proto/job/v1"
	"github.com/nedson202/dts-go/pkg/models"
	"github.com/nedson202/dts-go/pkg/queue"
	pb "github.com/nedson202/dts-go/proto/scheduler/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedSchedulerServiceServer
	cassandraClient *database.CassandraClient
	kafkaClient     *queue.KafkaClient
	resourceManager *ResourceManager
	queueManager    *QueueManager
}

func NewService(cassandraClient *database.CassandraClient, kafkaClient *queue.KafkaClient, rm *ResourceManager, qm *QueueManager) *Service {
	if kafkaClient == nil {
		log.Fatal("Kafka client is nil")
	}
	return &Service{
		cassandraClient: cassandraClient,
		kafkaClient:     kafkaClient,
		resourceManager: rm,
		queueManager:    qm,
	}
}

func (s *Service) ScheduleJob(ctx context.Context, req *pb.ScheduleJobRequest) (*pb.ScheduleJobResponse, error) {
	// Convert pb.Job to models.Job if necessary
	job, err := models.JobFromProto(req.Job)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job data: %v", err)
	}

	// Allocate resources
	err = models.AllocateResources(s.cassandraClient, models.Resources{
		CPU:     int32(req.ResourceRequirements.Cpu),
		Memory:  int32(req.ResourceRequirements.Memory),
		Storage: int32(req.ResourceRequirements.Storage),
	})
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "Failed to allocate resources: %v", err)
	}

	// Create a scheduled job
	scheduledJob := &models.ScheduledJob{
		JobID:     job.ID,
		Job:       *job,
		Resources: models.Resources{
			CPU:     int32(req.ResourceRequirements.Cpu),
			Memory:  int32(req.ResourceRequirements.Memory),
			Storage: int32(req.ResourceRequirements.Storage),
		},
		StartTime: time.Now(),
	}

	// Save the scheduled job
	if err := models.SaveScheduledJob(s.cassandraClient, scheduledJob); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save scheduled job: %v", err)
	}

	// Enqueue the job for execution
	if err := s.queueManager.EnqueueJob(ctx, scheduledJob); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to enqueue job: %v", err)
	}

	return &pb.ScheduleJobResponse{
		ScheduleId: scheduledJob.JobID.String(),
	}, nil
}

func (s *Service) CancelJob(ctx context.Context, req *pb.CancelJobRequest) (*pb.CancelJobResponse, error) {
	jobID, err := gocql.ParseUUID(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID: %v", err)
	}

	// Get the scheduled job
	scheduledJob, err := models.GetScheduledJob(s.cassandraClient, jobID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Scheduled job not found: %v", err)
	}

	// Delete the scheduled job
	err = models.DeleteScheduledJob(s.cassandraClient, jobID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete scheduled job: %v", err)
	}

	// Release the resources
	err = models.ReleaseResources(s.cassandraClient, scheduledJob.Resources)
	if err != nil {
		log.Printf("Failed to release resources: %v", err)
		// Continue execution, as the job is already cancelled
	}

	// Publish cancellation message to Kafka
	cancelMessage := struct {
		JobID     string    `json:"job_id"`
		Cancelled time.Time `json:"cancelled_at"`
	}{
		JobID:     req.JobId,
		Cancelled: time.Now(),
	}
	messageJSON, err := json.Marshal(cancelMessage)
	if err != nil {
		log.Printf("Error marshaling cancel message: %v", err)
	} else {
		err = s.kafkaClient.PublishMessage(ctx, []byte(req.JobId), messageJSON)
		if err != nil {
			log.Printf("Error publishing job cancellation to Kafka: %v", err)
		}
	}

	return &pb.CancelJobResponse{Success: true}, nil
}

func (s *Service) GetScheduledJob(ctx context.Context, req *pb.GetScheduledJobRequest) (*pb.GetScheduledJobResponse, error) {
	jobID, err := gocql.ParseUUID(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID: %v", err)
	}

	scheduledJob, err := models.GetScheduledJob(s.cassandraClient, jobID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Scheduled job not found: %v", err)
	}

	return scheduledJob.ToProto(), nil
}

func (s *Service) ListScheduledJobs(ctx context.Context, req *pb.ListScheduledJobsRequest) (*pb.ListScheduledJobsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100 // Default page size
	}

	var lastID gocql.UUID
	if req.PageToken != "" {
		var err error
		lastID, err = gocql.ParseUUID(req.PageToken)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid page token: %v", err)
		}
	}

	scheduledJobs, err := models.ListScheduledJobs(s.cassandraClient, pageSize, lastID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list scheduled jobs: %v", err)
	}

	var pbJobs []*pb.GetScheduledJobResponse
	for _, job := range scheduledJobs {
		pbJobs = append(pbJobs, job.ToProto())
	}

	var nextPageToken string
	if len(scheduledJobs) == pageSize {
		nextPageToken = scheduledJobs[len(scheduledJobs)-1].JobID.String()
	}

	return &pb.ListScheduledJobsResponse{
		Jobs:         pbJobs,
		NextPageToken: nextPageToken,
		TotalCount:    int32(len(pbJobs)),
	}, nil
}

type ScheduledJob struct {
	JobID     gocql.UUID
	Job       *jobv1.GetJobResponse
	Resources *models.Resources
	StartTime time.Time
}

