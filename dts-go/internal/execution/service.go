package execution

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nedson202/dts-go/pkg/database"
	pb "github.com/nedson202/dts-go/pkg/execution"
	"github.com/nedson202/dts-go/pkg/queue"
)

type Service struct {
	pb.UnimplementedExecutionServiceServer
	cassandraClient *database.CassandraClient
	kafkaClient     *queue.KafkaClient
}

func NewService(cassandraClient *database.CassandraClient, kafkaClient *queue.KafkaClient) *Service {
	return &Service{
		cassandraClient: cassandraClient,
		kafkaClient:     kafkaClient,
	}
}

func (s *Service) ExecuteJob(ctx context.Context, req *pb.ExecuteJobRequest) (*pb.ExecuteJobResponse, error) {
	executionID := "generated-execution-id" // In a real implementation, generate a unique ID

	executionJSON, err := json.Marshal(req.Job)
	if err != nil {
		log.Printf("Error marshaling job: %v", err)
		return nil, err
	}

	err = s.kafkaClient.PublishMessage(ctx, []byte(executionID), executionJSON)
	if err != nil {
		log.Printf("Error publishing job execution to Kafka: %v", err)
		return nil, err
	}

	return &pb.ExecuteJobResponse{
		ExecutionId: executionID,
	}, nil
}

func (s *Service) GetExecutionStatus(ctx context.Context, req *pb.GetExecutionStatusRequest) (*pb.GetExecutionStatusResponse, error) {
	// In a real implementation, retrieve the execution status from Cassandra
	// For this example, we'll return a mock response
	return &pb.GetExecutionStatusResponse{
		Status: "completed",
		Result: "Job executed successfully",
	}, nil
}
