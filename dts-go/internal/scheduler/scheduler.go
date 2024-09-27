package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/nedson202/dts-go/pkg/client"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/models"
	jobpb "github.com/nedson202/dts-go/proto/job/v1"
)

type Scheduler struct {
	cassandraClient *database.CassandraClient
	checkInterval   time.Duration
	queueManager    *QueueManager
	jobClient       *client.JobClient
}

func NewScheduler(cassandraClient *database.CassandraClient, checkInterval time.Duration, queueManager *QueueManager) (*Scheduler, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	jobServiceAddr := cfg.JobServiceAddr
	if jobServiceAddr == "" {
		return nil, fmt.Errorf("job service address is not set")
	}

	jobClient, err := client.NewJobClient(jobServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create job client: %w", err)
	}

	// Create a scheduler without the job client first
	scheduler := &Scheduler{
		cassandraClient: cassandraClient,
		checkInterval:   checkInterval,
		queueManager:    queueManager,
		jobClient:       jobClient,
	}

	logger.Info().Msgf("Initializing Scheduler with check interval: %v", checkInterval)
	return scheduler, nil
}

func (s *Scheduler) Start(ctx context.Context, segments []string) {
	logger.Info().Msgf("Starting Scheduler with segments: %v", segments)
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Scheduler stopped due to context cancellation")
			return
		case <-ticker.C:
			logger.Info().Msg("Running periodic job check")
			s.ProcessPendingJobs(ctx, segments)
		}
	}
}

func (s *Scheduler) ProcessPendingJobs(ctx context.Context, segments []string) error {
	if s.jobClient == nil {
		logger.Warn().Msg("Job client not yet connected. Skipping job processing.")
		return nil
	}

	startTime := time.Now().Truncate(time.Minute)
	logger.Info().Msg("Fetching pending tasks")
	tasks, err := models.GetScheduledTasksDueForExecution(s.cassandraClient, segments)
	if err != nil {
		logger.Error().Err(err).Msg("Error fetching pending tasks")
		return err
	}
	logger.Info().Msgf("Found %d pending tasks", len(tasks))

	scheduledCount := 0
	for _, task := range tasks {
		logger.Info().Msgf("Processing task: %s", task.JobID)
		if err := s.scheduleTask(ctx, task); err != nil {
			logger.Error().Err(err).Msgf("Error scheduling task %s", task.JobID)
		} else {
			scheduledCount++
		}
	}

	duration := time.Since(startTime)
	logger.Info().Msgf("Periodic task check completed. Scheduled %d out of %d tasks. Duration: %v", scheduledCount, len(tasks), duration)
	return nil
}

func (s *Scheduler) scheduleTask(ctx context.Context, task *models.TaskSchedule) error {
	// Update the job status to SCHEDULED
	_, err := s.jobClient.UpdateJob(ctx, task.JobID.String(), jobpb.JobStatus_SCHEDULED, time.Time{})
	if err != nil {
		logger.Error().Err(err).Msgf("Error updating job %s to SCHEDULED", task.JobID)
		return err
	}

	idempotencyKey, err := uuid.NewV4()
	if err != nil {
		logger.Error().Err(err).Msgf("Error generating unique ID for job %s", task.JobID)
		return err
	}

	// Use QueueManager to enqueue the job
	scheduledJob := &ScheduledJob{
		IdempotencyKey: idempotencyKey.String(),
		JobID:          task.JobID.String(),
		StartTime:      time.Now(),
	}
	err = s.queueManager.EnqueueJob(ctx, scheduledJob)
	if err != nil {
		logger.Error().Err(err).Msgf("Error enqueueing job %s", task.JobID)
		// Revert the job status to PENDING if enqueueing fails
		revertErr := s.revertJobStatus(ctx, task.JobID.String(), jobpb.JobStatus_PENDING)
		if revertErr != nil {
			logger.Error().Err(revertErr).Msgf("Failed to revert job %s status to PENDING", task.JobID)
		}
		return err
	}

	return nil
}

func (s *Scheduler) revertJobStatus(ctx context.Context, jobID string, status jobpb.JobStatus) error {
	_, err := s.jobClient.UpdateJob(ctx, jobID, status, time.Time{})
	return err
}

type ScheduledJob struct {
	IdempotencyKey string
	JobID          string
	StartTime      time.Time
}
