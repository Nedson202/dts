package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/queue"
)

type QueueManager struct {
	kafkaClient *queue.KafkaClient
}

func NewQueueManager(kafkaClient *queue.KafkaClient) *QueueManager {
	return &QueueManager{
		kafkaClient: kafkaClient,
	}
}

func (qm *QueueManager) EnqueueJob(ctx context.Context, job *ScheduledJob) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal job")
		return fmt.Errorf("failed to marshal job: %v", err)
	}

	err = qm.kafkaClient.PublishMessage(ctx, []byte(job.JobID.String()), jobJSON)
	if err != nil {
		logger.Error().Err(err).Msgf("Failed to publish job %s to Kafka", job.JobID)
		return fmt.Errorf("failed to publish job to Kafka: %v", err)
	}

	logger.Info().Msgf("Job %s enqueued successfully", job.JobID)
	return nil
}

func (qm *QueueManager) DequeueJob(ctx context.Context) (*ScheduledJob, error) {
	msg, err := qm.kafkaClient.ConsumeMessage(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to consume message from Kafka")
		return nil, fmt.Errorf("failed to consume message from Kafka: %v", err)
	}

	var job ScheduledJob
	err = json.Unmarshal(msg, &job)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to unmarshal job")
		return nil, fmt.Errorf("failed to unmarshal job: %v", err)
	}

	logger.Info().Msgf("Job %s dequeued successfully", job.JobID)
	return &job, nil
}

func (qm *QueueManager) Close() error {
	return qm.kafkaClient.Close()
}
