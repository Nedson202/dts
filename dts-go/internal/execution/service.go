package execution

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/client"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/models"
	pb "github.com/nedson202/dts-go/proto/execution/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedExecutionServiceServer
	cassandraClient *database.CassandraClient
	taskManager     *TaskManager
}

type ServiceConfig struct {
	CassandraClient *database.CassandraClient
	JobClient       *client.JobClient
	Brokers         []string
}

func NewService(serviceConfig ServiceConfig) (*Service, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	taskManager := NewTaskManager()

	// Add regular task processor
	taskManager.AddTaskProcessor(TaskProcessorArgs{
		Topic:            cfg.TaskTopic,
		CassandraClient: serviceConfig.CassandraClient,
		Brokers:         serviceConfig.Brokers,
		GroupID:         "task_execution_group",
		JobClient:       serviceConfig.JobClient,
	})

	// Add retry task processor
	err = taskManager.AddTaskRetryProcessor(TaskProcessorArgs{
		Topic:            cfg.TaskRetryTopic,
		CassandraClient: serviceConfig.CassandraClient,
		Brokers:          serviceConfig.Brokers,
		GroupID:          "task_retry_execution_group",
		JobClient:        serviceConfig.JobClient,
	})

	if err != nil {
		return nil, err
	}

	return &Service{
		cassandraClient: serviceConfig.CassandraClient,
		taskManager:     taskManager,
	}, nil
}

func (s *Service) GetExecution(ctx context.Context, req *pb.GetExecutionRequest) (*pb.ExecutionResponse, error) {
	id, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid execution ID")
	}

	execution, err := models.GetExecution(s.cassandraClient, id)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "Execution not found")
		}
		logger.Error().Err(err).Msg("Error retrieving execution from Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to retrieve execution")
	}

	return execution.ToProto(), nil
}

func (s *Service) ListExecutions(ctx context.Context, req *pb.ListExecutionsRequest) (*pb.ListExecutionsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 250 {
		pageSize = 250
	}

	var lastID gocql.UUID
	var err error
	if req.LastId != "" {
		lastID, err = gocql.ParseUUID(req.LastId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid last ID")
		}
	}

	executions, err := models.ListExecutions(s.cassandraClient, pageSize, lastID, req.JobId, req.Status)
	if err != nil {
		logger.Error().Err(err).Msg("Error listing executions from Cassandra")
		return nil, status.Errorf(codes.Internal, "Failed to list executions")
	}

	var pbExecutions []*pb.ExecutionResponse
	for _, execution := range executions {
		pbExecutions = append(pbExecutions, execution.ToProto())
	}

	var nextLastID string
	if len(executions) > 0 {
		nextLastID = executions[len(executions)-1].ID.String()
	}

	return &pb.ListExecutionsResponse{
		Executions: pbExecutions,
		Total:      int32(len(pbExecutions)),
		NextPage:   nextLastID,
	}, nil
}

func (s *Service) StartTaskManager(ctx context.Context) error {
	return s.taskManager.StartTaskManager(ctx)
}

func (s *Service) StopTaskManager() error {
	return s.taskManager.StopTaskManager()
}

type ScheduledJob struct {
	IdempotencyKey string    `json:"IdempotencyKey"`
	JobID          string    `json:"JobID"`
	StartTime      time.Time `json:"StartTime"`
	RetryCount     int       `json:"RetryCount"`
}
