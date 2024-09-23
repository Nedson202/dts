package scheduler

import (
	"context"
	"time"

	"github.com/nedson202/dts-go/internal/scheduler"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/queue"
)

type Server struct {
	cassandraClient *database.CassandraClient
	kafkaClient     *queue.KafkaClient
	scheduler       *scheduler.Scheduler
}

func NewServer(cassandraClient *database.CassandraClient, kafkaClient *queue.KafkaClient, checkInterval time.Duration) (*Server, error) {
	queueManager := scheduler.NewQueueManager(kafkaClient)
	scheduler, err := scheduler.NewScheduler(cassandraClient, checkInterval, queueManager)
	if err != nil {
		return nil, err
	}

	return &Server{
		cassandraClient: cassandraClient,
		kafkaClient:     kafkaClient,
		scheduler:       scheduler,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	logger.Info().Msg("Starting scheduler service...")

	// Start the scheduler
	go s.scheduler.Start(ctx)

	return nil
}
