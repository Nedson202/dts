package execution

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/models"
	jobpb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KafkaConsumer struct {
	consumer        sarama.ConsumerGroup
	cassandraClient *database.CassandraClient
	jobClient       jobpb.JobServiceClient
	ready           chan bool
}

func NewKafkaConsumer(brokers []string, groupID string, cassandraClient *database.CassandraClient, jobServiceAddr string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0 // Use an appropriate Kafka version

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	kafkaConsumer := &KafkaConsumer{
		consumer:        consumer,
		cassandraClient: cassandraClient,
		ready:           make(chan bool),
	}

	// Set up connection to job service
	conn, err := grpc.Dial(jobServiceAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to job service. Will retry in background.")
	} else {
		kafkaConsumer.jobClient = jobpb.NewJobServiceClient(conn)
		logger.Info().Msgf("Successfully connected to job service at %s", jobServiceAddr)
	}

	// Attempt to connect to the job service in a separate goroutine if initial connection failed
	if kafkaConsumer.jobClient == nil {
		go kafkaConsumer.connectJobService(jobServiceAddr)
	}

	return kafkaConsumer, nil
}

func (kc *KafkaConsumer) connectJobService(jobServiceAddr string) {
	for {
		conn, err := grpc.Dial(jobServiceAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to connect to job service. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		kc.jobClient = jobpb.NewJobServiceClient(conn)
		logger.Info().Msgf("Successfully connected to job service at %s", jobServiceAddr)
		break
	}
}

func (kc *KafkaConsumer) Consume(ctx context.Context, topics []string) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := kc.consumer.Consume(ctx, topics, kc); err != nil {
				logger.Error().Err(err).Msg("Error from consumer")
			}
			if ctx.Err() != nil {
				return
			}
			kc.ready = make(chan bool)
		}
	}()

	<-kc.ready // Wait till the consumer has been set up
	logger.Info().Msg("Kafka consumer up and running")

	<-ctx.Done()
	logger.Info().Msg("Terminating: context cancelled")
	wg.Wait()
	return nil
}

func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(kc.ready)
	return nil
}

func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		logger.Info().Msgf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		
		var scheduledJob struct {
			JobID     string    `json:"JobID"`
			StartTime time.Time `json:"StartTime"`
		}
		if err := json.Unmarshal(message.Value, &scheduledJob); err != nil {
			logger.Error().Err(err).Msg("Error unmarshaling message")
			continue
		}

		if scheduledJob.JobID == "" {
			logger.Error().Msg("Error: Job ID is empty in the message")
			continue
		}

		jobID, err := gocql.ParseUUID(scheduledJob.JobID)
		if err != nil {
			logger.Error().Err(err).Msgf("Error parsing job ID '%s'", scheduledJob.JobID)
			continue
		}

		// Create execution record
		execution := &models.Execution{
			ID:        gocql.TimeUUID(),
			JobID:     jobID,
			Status:    "RUNNING",
			StartTime: scheduledJob.StartTime,
		}

		if err := models.CreateExecution(kc.cassandraClient, execution); err != nil {
			logger.Error().Err(err).Msgf("Error creating execution for job %s", scheduledJob.JobID)
			continue
		}

		// Simulate job execution (replace this with actual job execution logic)
		time.Sleep(30 * time.Second)

		// Update execution record
		execution.Status = "COMPLETED"
		now := time.Now()
		execution.EndTime = &now
		if err := models.UpdateExecution(kc.cassandraClient, execution); err != nil {
			logger.Error().Err(err).Msgf("Error updating execution for job %s", scheduledJob.JobID)
		}

		// Update job status to COMPLETED
		if kc.jobClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err = kc.jobClient.UpdateJob(ctx, &jobpb.UpdateJobRequest{
				Id:     scheduledJob.JobID,
				Status: jobpb.JobStatus_COMPLETED,
				LastRun: timestamppb.New(execution.StartTime),
			})
			cancel()
			if err != nil {
				logger.Error().Err(err).Msgf("Error updating status for job %s", scheduledJob.JobID)
				continue
			}
		} else {
			logger.Warn().Msgf("Job client is not initialized, skipping status update for job %s", scheduledJob.JobID)
		}

		session.MarkMessage(message, "")
	}
	return nil
}
