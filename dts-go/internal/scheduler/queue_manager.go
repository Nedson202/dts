package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nedson202/dts-go/pkg/config"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal job")
		return fmt.Errorf("failed to marshal job: %v", err)
	}

	err = qm.kafkaClient.Produce(ctx, cfg.TaskTopic, []byte(job.IdempotencyKey), jobJSON)
	if err != nil {
		logger.Error().Err(err).Msgf("Failed to publish job %s to Kafka", job.JobID)
		return fmt.Errorf("failed to publish job to Kafka: %v", err)
	}

	logger.Info().Msgf("Job enqueued successfully %s", job.JobID)
	return nil
}

func (qm *QueueManager) Close() error {
	return qm.kafkaClient.Close()
}
