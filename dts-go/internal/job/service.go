package job

import (
	"context"
	"math/rand"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/models"
	"github.com/nedson202/dts-go/pkg/utils"
	pb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedJobServiceServer
	cassandraClient *database.CassandraClient
	segments        []string
}

func NewService(cassandraClient *database.CassandraClient, segments []string) *Service {
	return &Service{
		cassandraClient: cassandraClient,
		segments:        segments,
	}
}

func (s *Service) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	// Validate cron expression
	if err := utils.ValidateCronExpression(req.CronExpression); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid cron expression: %v", err)
	}
	
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

	if job.Status == pb.JobStatus_UNSPECIFIED.String() {
		job.Status = pb.JobStatus_PENDING.String()
	}

	nextRun, err := utils.CalculateNextRun(job.CronExpression, time.Now())
	if err != nil {
		logger.Error().Err(err).Msgf("Error calculating next run time for job %s", job.ID)
		return nil, status.Errorf(codes.Internal, "Failed to calculate next run time")
	}
	job.NextRun = nextRun

	// Create a task schedule
	randomSegment := s.segments[rand.Intn(len(s.segments))]
	taskSchedule := &models.TaskSchedule{
		NextExecutionTime: job.NextRun,
		Segment:           randomSegment,
		JobID:             job.ID,
	}

	// Start a batch operation
	batch := s.cassandraClient.Session.NewBatch(gocql.LoggedBatch)

	// Add job creation query to batch
	batch.Query(
		models.JobCreationQuery,
		job.ID, job.Name, job.Description, job.CronExpression, job.Status, job.CreatedAt, job.UpdatedAt, job.LastRun, job.NextRun, job.Metadata,
	)

	// Add task schedule creation query to batch
	batch.Query(
		models.TaskScheduleCreationQuery,
		taskSchedule.NextExecutionTime, taskSchedule.Segment, taskSchedule.JobID,
	)

	// Execute the batch
	if err := s.cassandraClient.Session.ExecuteBatch(batch); err != nil {
		logger.Error().Err(err).Msg("Error inserting job and task schedule into Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to create job")
	}

	return &pb.CreateJobResponse{JobId: job.ID.String()}, nil
}

func (s *Service) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.JobResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID")
	}

	job, err := models.GetJob(s.cassandraClient, id)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Job not found")
		}
		logger.Error().Err(err).Msg("Error retrieving job from Cassandra")
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
		logger.Error().Err(err).Msg("Error listing jobs from Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to list jobs")
	}

	var pbJobs []*pb.JobResponse
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

func (s *Service) UpdateJob(ctx context.Context, req *pb.UpdateJobRequest) (*pb.JobResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID")
	}

	// Fetch the existing job
	existingJob, err := models.GetJob(s.cassandraClient, id)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Job not found")
		}
		logger.Error().Err(err).Msg("Error retrieving job from Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to retrieve job")
	}

	// Update only the fields that are provided in the request
	if req.Name != "" {
		existingJob.Name = req.Name
	}
	if req.Description != "" {
		existingJob.Description = req.Description
	}
	if req.CronExpression != "" {
		if err := utils.ValidateCronExpression(req.CronExpression); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid cron expression: %v", err)
		}
		existingJob.CronExpression = req.CronExpression
	}
	if req.Status != pb.JobStatus_UNSPECIFIED {
		existingJob.Status = req.Status.String()
	}
	if req.Metadata != nil {
		existingJob.Metadata = req.Metadata
	}

	existingJob.UpdatedAt = time.Now()

	// Calculate new next run time
	nextRun, err := utils.CalculateNextRun(existingJob.CronExpression, time.Now())
	if err != nil {
		logger.Error().Err(err).Msgf("Error calculating next run time for job %s", existingJob.ID)
		return nil, status.Errorf(codes.Internal, "Failed to calculate next run time")
	}
	existingJob.NextRun = nextRun

	// Create a new task schedule
	existingTaskSchedule, err := models.GetTaskScheduleByJobID(s.cassandraClient, existingJob.ID)
	if err != nil {
		logger.Error().Err(err).Msg("Error retrieving task schedule from Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to retrieve task schedule")
	}
	randomSegment := s.segments[rand.Intn(len(s.segments))]
	newTaskSchedule := &models.TaskSchedule{
		NextExecutionTime: existingJob.NextRun,
		Segment:           randomSegment,
		JobID:             existingJob.ID,
	}

	// Start a batch operation
	batch := s.cassandraClient.Session.NewBatch(gocql.LoggedBatch)

	// Add job update query to batch
	batch.Query(
		models.JobUpdateQuery,
		existingJob.Name, existingJob.Description, existingJob.CronExpression, existingJob.Status,
		existingJob.UpdatedAt, existingJob.LastRun, existingJob.NextRun, existingJob.Metadata, existingJob.ID,
	)

	// Delete old task schedule
	batch.Query(models.TaskScheduleDeleteQuery, existingJob.ID, existingTaskSchedule.NextExecutionTime, existingTaskSchedule.Segment)

	// Add new task schedule creation query to batch
	batch.Query(
		models.TaskScheduleCreationQuery,
		newTaskSchedule.NextExecutionTime, newTaskSchedule.Segment, newTaskSchedule.JobID,
	)

	// Execute the batch
	if err := s.cassandraClient.Session.ExecuteBatch(batch); err != nil {
		logger.Error().Err(err).Msg("Error updating job and task schedule in Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to update job")
	}

	return existingJob.ToProto(), nil
}

func (s *Service) DeleteJob(ctx context.Context, req *pb.DeleteJobRequest) (*pb.DeleteJobResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid job ID")
	}

	err = models.DeleteJob(s.cassandraClient, id)
	if err != nil {
		logger.Error().Err(err).Msg("Error deleting job from Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to delete job")
	}

	return &pb.DeleteJobResponse{Success: true}, nil
}
