package execution

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/models"
	"github.com/nedson202/dts-go/pkg/queue"
	jobpb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskConsumer struct {
	cassandraClient *database.CassandraClient
	kafkaConsumer   *queue.KafkaConsumer
	jobClient       jobpb.JobServiceClient
}

func NewTaskConsumer(cassandraClient *database.CassandraClient, brokers []string, groupID string, jobServiceAddr string) (*TaskConsumer, error) {
	consumer := &TaskConsumer{
		cassandraClient: cassandraClient,
	}

	kafkaConsumer, err := queue.NewKafkaConsumer(brokers, groupID, consumer.handleMessage)
	if err != nil {
		return nil, err
	}
	consumer.kafkaConsumer = kafkaConsumer

	// Set up connection to job service
	conn, err := grpc.Dial(jobServiceAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to job service. Will retry in background.")
	} else {
		consumer.jobClient = jobpb.NewJobServiceClient(conn)
		logger.Info().Msgf("Successfully connected to job service at %s", jobServiceAddr)
	}

	// Attempt to connect to the job service in a separate goroutine if initial connection failed
	if consumer.jobClient == nil {
		go consumer.connectJobService(jobServiceAddr)
	}

	return consumer, nil
}

func (tc *TaskConsumer) connectJobService(jobServiceAddr string) {
	for {
		conn, err := grpc.Dial(jobServiceAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to connect to job service. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		tc.jobClient = jobpb.NewJobServiceClient(conn)
		logger.Info().Msgf("Successfully connected to job service at %s", jobServiceAddr)
		break
	}
}

func (tc *TaskConsumer) handleMessage(message []byte) error {
	var scheduledJob struct {
		JobID     string    `json:"JobID"`
		StartTime time.Time `json:"StartTime"`
	}
	if err := json.Unmarshal(message, &scheduledJob); err != nil {
		logger.Error().Err(err).Msg("Error unmarshaling message")
		return err
	}

	if scheduledJob.JobID == "" {
		logger.Error().Msg("Error: Job ID is empty in the message")
		return errors.New("error: Job ID is empty in the message")
	}

	jobID, err := gocql.ParseUUID(scheduledJob.JobID)
	if err != nil {
		logger.Error().Err(err).Msgf("Error parsing job ID '%s'", scheduledJob.JobID)
		return err
	}

	// Create execution record
	execution := &models.Execution{
		ID:        gocql.TimeUUID(),
		JobID:     jobID,
		Status:    "RUNNING",
		StartTime: scheduledJob.StartTime,
	}

	if err := models.CreateExecution(tc.cassandraClient, execution); err != nil {
		logger.Error().Err(err).Msgf("Error creating execution for job %s", scheduledJob.JobID)
		return err
	}

	// Simulate job execution (replace this with actual job execution logic)
	time.Sleep(30 * time.Second)

	// Update execution record
	execution.Status = "COMPLETED"
	now := time.Now()
	execution.EndTime = &now
	if err := models.UpdateExecution(tc.cassandraClient, execution); err != nil {
		logger.Error().Err(err).Msgf("Error updating execution for job %s", scheduledJob.JobID)
		return err
	}

	// Update job status to COMPLETED
	if tc.jobClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = tc.jobClient.UpdateJob(ctx, &jobpb.UpdateJobRequest{
			Id:     scheduledJob.JobID,
			Status: jobpb.JobStatus_COMPLETED,
			LastRun: timestamppb.New(execution.StartTime),
		})
		cancel()
		if err != nil {
			logger.Error().Err(err).Msgf("Error updating status for job %s", scheduledJob.JobID)
			return err
		}
	} else {
		logger.Warn().Msgf("Job client is not initialized, skipping status update for job %s", scheduledJob.JobID)
	}


	return nil
}

func (tc *TaskConsumer) Start(ctx context.Context, topic string) error {
	// Start consuming messages
	return tc.kafkaConsumer.Consume(ctx, []string{topic})
}

func (tc *TaskConsumer) Stop() error {
	return tc.kafkaConsumer.Close()
}
