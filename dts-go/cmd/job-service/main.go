package main

import (
	"log"

	"github.com/nedson202/dts-go/internal/job"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/queue"
	jobServer "github.com/nedson202/dts-go/pkg/services/job"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Log Cassandra hosts for debugging
	log.Printf("Cassandra Hosts: %v", cfg.CassandraHosts)

	// Use localhost if no Cassandra hosts are provided
	if len(cfg.CassandraHosts) == 0 {
		cfg.CassandraHosts = []string{"localhost"}
		log.Println("No Cassandra hosts provided, using localhost")
	}

	// Initialize Cassandra client
	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}
	defer cassandraClient.Close()

	// Log Kafka brokers for debugging
	log.Printf("Kafka Brokers: %v", cfg.KafkaBrokers)

	// Initialize Kafka client
	kafkaClient, err := queue.NewKafkaClient(cfg.KafkaBrokers, cfg.JobTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	// Create job service
	jobService := job.NewService(cassandraClient, kafkaClient)

	// Use separate ports for gRPC and HTTP
	grpcPort := cfg.JobServiceGRPCPort
	httpPort := cfg.JobServiceHTTPPort

	log.Printf("Starting server on gRPC port %s and HTTP port %s", grpcPort, httpPort)

	// Create and run server
	server := jobServer.NewServer(jobService, grpcPort, httpPort)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
