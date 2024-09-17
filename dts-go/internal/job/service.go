package job

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/models"
	"github.com/nedson202/dts-go/pkg/queue"
	pb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedJobServiceServer
	cassandraClient *database.CassandraClient
	kafkaClient     *queue.KafkaClient
}

func NewService(cassandraClient *database.CassandraClient, kafkaClient *queue.KafkaClient) *Service {
	return &Service{
		cassandraClient: cassandraClient,
		kafkaClient:     kafkaClient,
	}
}

func (s *Service) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	job := &models.Job{
		ID:             gocql.TimeUUID(),
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		Status:         req.Status.String(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       req.Metadata,
	}

	if job.Status == pb.JobStatus_JOB_STATUS_UNSPECIFIED.String() {
		job.Status = pb.JobStatus_JOB_STATUS_PENDING.String()
	}

	err := models.CreateJob(s.cassandraClient, job)
	if err != nil {
		log.Printf("Error inserting job into Cassandra: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create job")
	}

	// Publish the new job to Kafka
	jobJSON, err := json.Marshal(job)
	if err != nil {
		log.Printf("Error marshaling job: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to process job")
	}

	err = s.kafkaClient.PublishMessage(ctx, []byte(job.ID.String()), jobJSON)
	if err != nil {
		log.Printf("Error publishing job to Kafka: %v", err)
		// Note: We're not returning an error here as the job is already created in Cassandra
	}

	return &pb.CreateJobResponse{JobId: job.ID.String()}, nil
}

func (s *Service) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID")
	}

	job, err := models.GetJob(s.cassandraClient, id)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Job not found")
		}
		log.Printf("Error retrieving job from Cassandra: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve job")
	}

	return job.ToProto(), nil
}

func (s *Service) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 250 {
		pageSize = 250
	}

	nilUUID := gocql.UUID{}
	var lastID gocql.UUID
	var err error
	if req.LastId != "" {
		lastID, err = gocql.ParseUUID(req.LastId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid last ID")
		}
	} else {
		lastID = nilUUID
	}

	jobs, err := models.ListJobs(s.cassandraClient, pageSize, lastID, req.Status)
	if err != nil {
		log.Printf("Error listing jobs from Cassandra: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list jobs")
	}

	var pbJobs []*pb.GetJobResponse
	for _, job := range jobs {
		pbJobs = append(pbJobs, job.ToProto())
	}

	var nextLastID string
	if len(jobs) > 0 {
		nextLastID = jobs[len(jobs)-1].ID.String()
	}

	return &pb.ListJobsResponse{
		Jobs:     pbJobs,
		Total:    int32(len(pbJobs)),
		NextPage: nextLastID,
	}, nil
}

func (s *Service) UpdateJob(ctx context.Context, req *pb.UpdateJobRequest) (*pb.GetJobResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID")
	}

	job, err := models.GetJob(s.cassandraClient, id)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Job not found")
		}
		log.Printf("Error retrieving job from Cassandra: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve job")
	}

	job.Name = req.Name
	job.Description = req.Description
	job.CronExpression = req.CronExpression
	job.UpdatedAt = time.Now()
	job.Metadata = req.Metadata

	// Check if the status is being updated
	if req.Status != pb.JobStatus_JOB_STATUS_UNSPECIFIED {
		// Validate the state transition

		if !isValidStateTransition(pb.JobStatus(pb.JobStatus_value[job.Status]), req.Status) {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid state transition from %s to %s", job.Status, req.Status)
		}
		job.Status = req.Status.String()
	}

	err = models.UpdateJob(s.cassandraClient, job)
	if err != nil {
		log.Printf("Error updating job in Cassandra: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to update job")
	}

	// Publish the updated job to Kafka
	jobJSON, err := json.Marshal(job)
	if err != nil {
		log.Printf("Error marshaling updated job: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to process updated job")
	}

	err = s.kafkaClient.PublishMessage(ctx, []byte(job.ID.String()), jobJSON)
	if err != nil {
		log.Printf("Error publishing updated job to Kafka: %v", err)
		// Note: We're not returning an error here as the job is already updated in Cassandra
	}

	return job.ToProto(), nil
}

func (s *Service) DeleteJob(ctx context.Context, req *pb.DeleteJobRequest) (*pb.DeleteJobResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID")
	}

	err = models.DeleteJob(s.cassandraClient, id)
	if err != nil {
		log.Printf("Error deleting job from Cassandra: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to delete job")
	}

	// Publish the deleted job ID to Kafka
	err = s.kafkaClient.PublishMessage(ctx, []byte(id.String()), []byte("deleted"))
	if err != nil {
		log.Printf("Error publishing deleted job to Kafka: %v", err)
		// Note: We're not returning an error here as the job is already deleted from Cassandra
	}

	return &pb.DeleteJobResponse{Success: true}, nil
}

func isValidStateTransition(from, to pb.JobStatus) bool {
	// Define valid state transitions
	validTransitions := map[pb.JobStatus][]pb.JobStatus{
		pb.JobStatus_JOB_STATUS_PENDING:   {pb.JobStatus_JOB_STATUS_SCHEDULED, pb.JobStatus_JOB_STATUS_CANCELLED},
		pb.JobStatus_JOB_STATUS_SCHEDULED: {pb.JobStatus_JOB_STATUS_RUNNING, pb.JobStatus_JOB_STATUS_CANCELLED},
		pb.JobStatus_JOB_STATUS_RUNNING:   {pb.JobStatus_JOB_STATUS_COMPLETED, pb.JobStatus_JOB_STATUS_FAILED, pb.JobStatus_JOB_STATUS_CANCELLED},
		pb.JobStatus_JOB_STATUS_FAILED:    {pb.JobStatus_JOB_STATUS_RETRYING, pb.JobStatus_JOB_STATUS_CANCELLED},
		pb.JobStatus_JOB_STATUS_PAUSED:    {pb.JobStatus_JOB_STATUS_SCHEDULED, pb.JobStatus_JOB_STATUS_CANCELLED},
	}

	validToStates, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, validState := range validToStates {
		if to == validState {
			return true
		}
	}

	return false
}
