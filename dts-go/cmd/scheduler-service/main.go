package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nedson202/dts-go/internal/scheduler"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/queue"
	schedulerServer "github.com/nedson202/dts-go/pkg/services/scheduler"
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
	kafkaClient, err := queue.NewKafkaClient(cfg.KafkaBrokers, cfg.SchedulerTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	resourceManager := scheduler.NewResourceManager(cassandraClient)
	queueManager := scheduler.NewQueueManager(kafkaClient)

	// Create scheduler service
	schedulerService := scheduler.NewService(cassandraClient, kafkaClient, resourceManager, queueManager)

	// Create and start the periodic scheduler
	jobServiceAddr := fmt.Sprintf("%s:%s", cfg.JobServiceHost, cfg.JobServiceGRPCPort)
	periodicScheduler, err := scheduler.NewPeriodicScheduler(cassandraClient, schedulerService, 1*time.Minute, jobServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create periodic scheduler: %v", err)
	}
	go periodicScheduler.Start(context.Background())

	// Use separate ports for gRPC and HTTP
	grpcPort := cfg.SchedulerServiceGRPCPort
	httpPort := cfg.SchedulerServiceHTTPPort

	log.Printf("Starting scheduler server on gRPC port %s and HTTP port %s", grpcPort, httpPort)

	// Create and run server
	server := schedulerServer.NewServer(schedulerService, grpcPort, httpPort)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
