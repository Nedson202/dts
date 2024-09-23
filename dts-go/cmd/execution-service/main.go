package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nedson202/dts-go/internal/execution"
	"github.com/nedson202/dts-go/pkg/client"
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

	jobClient, err := client.NewJobClient(cfg.JobServiceAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create job client")
	}


	service, err := execution.NewService(execution.ServiceConfig{
		CassandraClient: cassandraClient,
		JobClient:       jobClient,
		Brokers:         cfg.KafkaBrokers,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create execution service")
	}

	// Start the task manager
	if err := service.StartTaskManager(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start task manager")
	}

	// Create and run server
	server := executionServer.NewServer(service, cfg.ExecutionServiceGRPCPort, cfg.ExecutionServiceHTTPPort)

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

	// Stop the task manager
	if err := service.StopTaskManager(); err != nil {
		logger.Error().Err(err).Msg("Error stopping task manager")
	}

	logger.Info().Msg("Server exiting")
}
