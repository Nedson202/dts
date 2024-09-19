package main

import (
	"github.com/nedson202/dts-go/internal/job"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	jobServer "github.com/nedson202/dts-go/pkg/services/job"
)

func main() {
	logger.Init()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	// Use localhost if no Cassandra hosts are provided
	if len(cfg.CassandraHosts) == 0 {
		cfg.CassandraHosts = []string{"localhost"}
		logger.Info().Msg("No Cassandra hosts provided, using localhost")
	}

	// Initialize Cassandra client
	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Cassandra client")
	}
	defer cassandraClient.Close()

	// Log Kafka brokers for debugging
	logger.Info().Msgf("Kafka Brokers: %v", cfg.KafkaBrokers)

	// Create job service
	jobService := job.NewService(cassandraClient)

	// Use separate ports for gRPC and HTTP
	grpcPort := cfg.JobServiceGRPCPort
	httpPort := cfg.JobServiceHTTPPort

	logger.Info().Msgf("Starting server on gRPC port %s and HTTP port %s", grpcPort, httpPort)

	// Create and run server
	server := jobServer.NewServer(jobService, grpcPort, httpPort)
	if err := server.Run(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run server")
	}
}
