package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nedson202/dts-go/internal/execution"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	executionServer "github.com/nedson202/dts-go/pkg/services/execution"
)

func main() {
	logger.Init()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Cassandra client")
	}
	defer cassandraClient.Close()

	executionService, err := execution.NewService(cassandraClient, cfg.KafkaBrokers, "execution-group", cfg.JobServiceAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create execution service")
	}

	// Create and run server
	server := executionServer.NewServer(executionService, cfg.ExecutionServiceGRPCPort, cfg.ExecutionServiceHTTPPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the TaskConsumer
	go func() {
		if err := executionService.StartTaskConsumer(ctx); err != nil {
			logger.Fatal().Err(err).Msg("Error from TaskConsumer")
		}
	}()

	// Start the server in a new goroutine
	go func() {
		if err := server.Run(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutting down server...")

	// Cancel the context to stop the TaskConsumer
	cancel()

	// Stop the TaskConsumer
	if err := executionService.StopTaskConsumer(); err != nil {
		logger.Error().Err(err).Msg("Error stopping TaskConsumer")
	}

	logger.Info().Msg("Server exiting")
}
