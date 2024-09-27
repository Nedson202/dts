package execution

import (
	"github.com/nedson202/dts-go/pkg/client"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/queue"
)

var _ TaskProcessor = (*TaskConsumer)(nil)

type TaskConsumer struct {
	kafkaClient *queue.KafkaClient
	executor    *TaskExecutor
}

type TaskConsumerArgs struct {
	CassandraClient *database.CassandraClient
	Brokers         []string
	GroupID         string
	JobClient       *client.JobClient
	Topic           string
}

func NewTaskConsumer(args TaskConsumerArgs) (*TaskConsumer, error) {
	kafkaClient, err := queue.NewKafkaClient(args.Brokers, args.GroupID, args.Topic)
	if err != nil {
		return nil, err
	}
	
	executor := NewTaskExecutor(args.CassandraClient, args.JobClient, kafkaClient)
	return &TaskConsumer{kafkaClient: kafkaClient, executor: executor}, nil
}

func (tc *TaskConsumer) Start(topic string) error {
	logger.Info().Msgf("Starting TaskConsumer for topic: %s", topic)
	if err := tc.kafkaClient.Consume(); err != nil {
		return err
	}

	go func() {
		for message := range tc.kafkaClient.Messages() {
			if err := tc.executor.executeTask(message); err != nil {
				logger.Error().Msgf("Error executing task: %v", err)
			}
		}
	}()

	return nil
}

func (tc *TaskConsumer) Stop() error {
	logger.Info().Msgf("Stopping TaskConsumer")
	return tc.kafkaClient.Close()
}
