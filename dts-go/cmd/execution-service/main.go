package main

import (
	"log"
	"net"
	"os"

	"github.com/nedson202/dts-go/internal/execution"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	pb "github.com/nedson202/dts-go/pkg/execution"
	"github.com/nedson202/dts-go/pkg/queue"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}
	defer cassandraClient.Close()

	kafkaClient, err := queue.NewKafkaClient(cfg.GetKafkaBrokers(), cfg.ExecutionTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	executionService := execution.NewService(cassandraClient, kafkaClient)

	port := os.Getenv("EXECUTION_SERVICE_PORT")
	if port == "" {
		port = ":50053"
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterExecutionServiceServer(s, executionService)

	log.Printf("Execution service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
