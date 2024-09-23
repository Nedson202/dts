package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/nedson202/dts-go/pkg/client"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
)

type TaskManager struct {
	processors map[string]TaskProcessor
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

type TaskProcessorArgs struct {
	Topic           string
	CassandraClient *database.CassandraClient
	Brokers         []string
	GroupID         string
	JobClient       *client.JobClient
}

func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		processors: make(map[string]TaskProcessor),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (tm *TaskManager) AddTaskProcessor(args TaskProcessorArgs) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	processor, err := NewTaskConsumer(TaskConsumerArgs{
		CassandraClient: args.CassandraClient,
		Brokers:         args.Brokers,
		GroupID:         args.GroupID,
		JobClient:       args.JobClient,
		Topic:           args.Topic,
	})
	if err != nil {
		return fmt.Errorf("failed to create task processor: %w", err)
	}

	tm.processors[args.Topic] = processor
	return nil
}

func (tm *TaskManager) AddTaskRetryProcessor(args TaskProcessorArgs) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	processor, err := NewTaskRetryConsumer(TaskRetryConsumerArgs{
		CassandraClient: args.CassandraClient,
		Brokers:         args.Brokers,
		GroupID:         args.GroupID,
		JobClient:       args.JobClient,
		Topic:           args.Topic,
	})
	if err != nil {
		return fmt.Errorf("failed to create task retry processor: %w", err)
	}
	
	tm.processors[args.Topic] = processor
	return nil
}

func (tm *TaskManager) StartTaskManager(ctx context.Context) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	for topic, processor := range tm.processors {
		go func(t string, p TaskProcessor) {
			if err := p.Start(t); err != nil {
				logger.Error().Err(err).Msgf("Failed to start processor for topic %s", t)
			}
		}(topic, processor)
	}

	logger.Info().Msg("Task Manager started successfully")
	return nil
}

func (tm *TaskManager) StopTaskManager() error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tm.cancel() // Cancel the context to stop all processors

	var errs []error
	for topic, processor := range tm.processors {
		if err := processor.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop processor %s: %w", topic, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while stopping processors: %v", errs)
	}

	logger.Info().Msg("Task Manager stopped successfully")
	return nil
}

func (tm *TaskManager) GetProcessor(topic string) (TaskProcessor, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	processor, exists := tm.processors[topic]
	return processor, exists
}
