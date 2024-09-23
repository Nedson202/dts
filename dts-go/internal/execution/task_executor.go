package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/client"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/models"
	"github.com/nedson202/dts-go/pkg/queue"
	jobpb "github.com/nedson202/dts-go/proto/job/v1"
)

type TaskExecutor struct {
	cassandraClient *database.CassandraClient
	jobClient       *client.JobClient
	kafkaClient     *queue.KafkaClient
	maxRetries      int
}

func NewTaskExecutor(cassandraClient *database.CassandraClient, jobClient *client.JobClient, kafkaClient *queue.KafkaClient) *TaskExecutor {
	return &TaskExecutor{
		cassandraClient: cassandraClient,
		jobClient:       jobClient,
		kafkaClient:     kafkaClient,
		maxRetries:      3,
	}
}

func (tc *TaskExecutor) executeTask(message []byte) error {
	var scheduledJob ScheduledJob
	if err := json.Unmarshal(message, &scheduledJob); err != nil {
		return fmt.Errorf("error unmarshaling message: %w", err)
	}

	return tc.processAndRetry(scheduledJob)
}

func (tc *TaskExecutor) executeRetryTask(message []byte) error {
	var scheduledJob ScheduledJob
	if err := json.Unmarshal(message, &scheduledJob); err != nil {
		return fmt.Errorf("error unmarshaling message: %w", err)
	}

	if scheduledJob.RetryCount >= tc.maxRetries {
		logger.Info().Msgf("Max retries reached for idempotency key %s. Retry count: %d", scheduledJob.IdempotencyKey, scheduledJob.RetryCount)
		return nil
	}

	return tc.processAndRetry(scheduledJob)
}

func (tc *TaskExecutor) processAndRetry(scheduledJob ScheduledJob) error {
	err := tc.processTask(scheduledJob)
	if err != nil {
		logger.Error().Err(err).Msgf("Error processing task %s", scheduledJob.JobID)

		scheduledJob.RetryCount++
		return tc.enqueueForRetry(scheduledJob)
	}
	return nil
}

func (tc *TaskExecutor) enqueueForRetry(scheduledJob ScheduledJob) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	jobJSON, err := json.Marshal(scheduledJob)
	if err != nil {
		return fmt.Errorf("error marshaling job for retry: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := tc.kafkaClient.Produce(ctx, cfg.TaskRetryTopic, []byte(scheduledJob.IdempotencyKey), jobJSON); err != nil {
		return fmt.Errorf("failed to publish retry message: %w", err)
	}

	return nil
}

func (tc *TaskExecutor) processTask(scheduledJob ScheduledJob) error {
	if scheduledJob.JobID == "" {
		return fmt.Errorf("job ID is empty in the message")
	}

	jobID, err := gocql.ParseUUID(scheduledJob.JobID)
	if err != nil {
		return fmt.Errorf("error parsing job ID '%s': %w", scheduledJob.JobID, err)
	}

	// Create execution record
	execution := &models.Execution{
		ID:        gocql.TimeUUID(),
		JobID:     jobID,
		Status:    "RUNNING",
		StartTime: scheduledJob.StartTime,
	}
	logger.Info().Msgf("Creating execution for job %s", scheduledJob.JobID)

	if err := models.CreateExecution(tc.cassandraClient, execution); err != nil {
		return fmt.Errorf("error creating execution for job %s: %w", scheduledJob.JobID, err)
	}
	logger.Info().Msgf("Execution created for job %s", scheduledJob.JobID)

	// // Simulate job execution (replace this with actual job execution logic)
	// time.Sleep(30 * time.Second)

	// Update execution record
	execution.Status = "COMPLETED"
	now := time.Now()
	execution.EndTime = &now
	if err := models.UpdateExecution(tc.cassandraClient, execution); err != nil {
		return fmt.Errorf("error updating execution for job %s: %w", scheduledJob.JobID, err)
	}
	logger.Info().Msgf("Execution updated for job %s", scheduledJob.JobID)

	// Update job status to COMPLETED
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = tc.jobClient.UpdateJob(ctx, scheduledJob.JobID, jobpb.JobStatus_COMPLETED, now)
	if err != nil {
		return fmt.Errorf("error updating status for job %s: %w", scheduledJob.JobID, err)
	}
	logger.Info().Msgf("Job %s status updated to COMPLETED", scheduledJob.JobID)

	return nil
}
