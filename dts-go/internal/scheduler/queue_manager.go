package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nedson202/dts-go/pkg/models"
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

func (qm *QueueManager) EnqueueJob(ctx context.Context, job *models.ScheduledJob) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %v", err)
	}

	err = qm.kafkaClient.PublishMessage(ctx, []byte(job.JobID.String()), jobJSON)
	if err != nil {
		return fmt.Errorf("failed to publish job to Kafka: %v", err)
	}

	log.Printf("Job %s enqueued successfully", job.JobID)
	return nil
}

func (qm *QueueManager) DequeueJob(ctx context.Context) (*models.ScheduledJob, error) {
	msg, err := qm.kafkaClient.ConsumeMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to consume message from Kafka: %v", err)
	}

	var job models.ScheduledJob
	err = json.Unmarshal(msg, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %v", err)
	}

	log.Printf("Job %s dequeued successfully", job.JobID)
	return &job, nil
}

func (qm *QueueManager) Close() error {
	return qm.kafkaClient.Close()
}
