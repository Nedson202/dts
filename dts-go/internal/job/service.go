package job

import (
	"context"
	"fmt"
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
}

func NewService(cassandraClient *database.CassandraClient) *Service {
	return &Service{
		cassandraClient: cassandraClient,
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

	err := models.CreateJob(s.cassandraClient, job)
	if err != nil {
		logger.Error().Err(err).Msg("Error inserting job into Cassandra")
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

	if req.LastRun != nil {
		lastRunTime := req.LastRun.AsTime()
		existingJob.LastRun = &lastRunTime
	}

	existingJob.UpdatedAt = time.Now()

	err = models.UpdateJob(s.cassandraClient, existingJob)
	if err != nil {
		logger.Error().Err(err).Msg("Error updating job in Cassandra")
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

func (s *Service) CancelJob(ctx context.Context, req *pb.CancelJobRequest) (*pb.CancelJobResponse, error) {
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

	if job.Status == pb.JobStatus_COMPLETED.String() || job.Status == pb.JobStatus_FAILED.String() || job.Status == pb.JobStatus_CANCELLED.String() {
		return nil, status.Errorf(codes.FailedPrecondition, "Cannot cancel job with status: %s", job.Status)
	}

	job.Status = pb.JobStatus_CANCELLED.String()
	job.UpdatedAt = time.Now()

	err = models.UpdateJob(s.cassandraClient, job)
	if err != nil {
		logger.Error().Err(err).Msg("Error updating job in Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to cancel job")
	}

	return &pb.CancelJobResponse{
		Success: true,
		Message: fmt.Sprintf("Job %s has been cancelled", job.ID),
	}, nil
}
