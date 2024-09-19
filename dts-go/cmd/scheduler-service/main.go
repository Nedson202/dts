package main

import (
	"context"
	"log"
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
		log.Fatalf("Failed to load config: %v", err)
	}

	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}
	defer cassandraClient.Close()

	kafkaClient, err := queue.NewKafkaClient(cfg.KafkaBrokers, cfg.TaskTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	
	checkInterval := 1 * time.Minute
	server, err := scheduler.NewServer(cassandraClient, kafkaClient, checkInterval)
	if err != nil {
		log.Fatalf("Failed to create scheduler server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Received shutdown signal")
		cancel()
	}()

	if err := server.Run(ctx); err != nil {
		log.Fatalf("Scheduler service error: %v", err)
	}
}
