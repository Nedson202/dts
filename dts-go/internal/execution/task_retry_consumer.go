package execution

import (
	"github.com/nedson202/dts-go/pkg/client"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/queue"
)

var _ TaskProcessor = (*TaskRetryConsumer)(nil)

type TaskRetryConsumer struct {
	kafkaClient *queue.KafkaClient
	executor    *TaskExecutor
}

type TaskRetryConsumerArgs struct {
	CassandraClient *database.CassandraClient
	Brokers         []string
	GroupID         string
	JobClient       *client.JobClient
	Topic           string
}

func NewTaskRetryConsumer(args TaskRetryConsumerArgs) (*TaskRetryConsumer, error) {
	kafkaClient, err := queue.NewKafkaClient(args.Brokers, args.GroupID, args.Topic)
	executor := NewTaskExecutor(args.CassandraClient, args.JobClient, kafkaClient)
	if err != nil {
		return nil, err
	}

	return &TaskRetryConsumer{kafkaClient: kafkaClient, executor: executor}, nil
}

func (tc *TaskRetryConsumer) Start(topic string) error {
	logger.Info().Msgf("Starting TaskRetryConsumer for topic: %s", topic)
	if err := tc.kafkaClient.Consume(); err != nil {
		return err
	}

	go func() {
		for message := range tc.kafkaClient.Messages() {
			if err := tc.executor.executeRetryTask(message); err != nil {
				logger.Error().Msgf("Error executing task: %v", err)
			}
		}
	}()

	return nil
}

func (tc *TaskRetryConsumer) Stop() error {
	logger.Info().Msgf("Stopping TaskRetryConsumer")
	return tc.kafkaClient.Close()
}
