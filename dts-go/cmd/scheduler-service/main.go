package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/queue"
	"github.com/nedson202/dts-go/pkg/services/scheduler"
)

func main() {
	logger.Init()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load config")
		os.Exit(1)
	}

	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Cassandra client")
		os.Exit(1)
	}
	defer cassandraClient.Close()

	kafkaClient, err := queue.NewKafkaClient(cfg.KafkaBrokers, "scheduler-service", "")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Kafka client")
		os.Exit(1)
	}
	defer kafkaClient.Close()

	checkInterval := 1 * time.Minute
	server, err := scheduler.NewServer(cassandraClient, kafkaClient, checkInterval)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create scheduler server")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Run(ctx); err != nil {
			logger.Error().Err(err).Msg("Scheduler stopped unexpectedly")
			cancel() // Cancel context to initiate shutdown
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutdown signal received, cancelling context...")
	cancel()
}
