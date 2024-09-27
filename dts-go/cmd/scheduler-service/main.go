package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/election"
	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/nedson202/dts-go/pkg/queue"
	"github.com/nedson202/dts-go/pkg/services/scheduler"
)

func main() {
	logger.Init()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load config")
		os.Exit(1)
	}

	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Cassandra client")
		os.Exit(1)
	}
	defer cassandraClient.Close()

	kafkaClient, err := queue.NewKafkaClient(cfg.KafkaBrokers, "scheduler-service", "")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Kafka client")
		os.Exit(1)
	}
	defer kafkaClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instanceID := uuid.New().String()
	electionManager := election.NewElection(cassandraClient, instanceID)

	var wg sync.WaitGroup

	// leadershipChangeChan := make(chan bool)

	// Run the election process
	wg.Add(1)
	go func() {
		defer wg.Done()
		// defer close(leadershipChangeChan)
		
		if err := electionManager.RunElection(ctx); err != nil {
			if err != context.Canceled {
				logger.Error().Err(err).Msg("Election process failed")
			}
			return
		}
	}()

	// // Listen for election results
	// go func() {
	// 	for isLeader := range electionManager.ResultChan() {
	// 		logger.Info().Msgf("Election result: %v", isLeader)
	// 		leadershipChangeChan <- isLeader
	// 	}
	// }()

	// Run the scheduler service
	wg.Add(1)
	go func() {
		defer wg.Done()
		checkInterval := 1 * time.Minute
		server, err := scheduler.NewServer(cassandraClient, kafkaClient, checkInterval)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create scheduler server")
			cancel() // Cancel context to initiate shutdown
			return
		}

		var serverCtx context.Context
		var serverCancel context.CancelFunc
		var serverWg sync.WaitGroup

		for isLeader := range electionManager.ResultChan() {
			if serverCancel != nil {
				serverCancel()
				serverWg.Wait()
			}

			serverCtx, serverCancel = context.WithCancel(ctx)

			if isLeader {
				logger.Info().Msgf("Instance %s became the leader, starting scheduler", instanceID)
				serverWg.Add(1)
				go func() {
					defer serverWg.Done()
					logger.Info().Msgf("Starting scheduler service as coordinator for instance %s", instanceID)
					if err := server.RunAsCoordinator(serverCtx); err != nil {
						if err != context.Canceled {
							logger.Error().Err(err).Msg("Leader scheduler stopped unexpectedly")
							cancel() // Cancel main context to initiate shutdown
						}
					}
				}()
			} else {
				logger.Info().Msgf("Instance %s is now a follower", instanceID)
				serverWg.Add(1)
				go func() {
					defer serverWg.Done()
					if err := server.RunAsWorker(serverCtx); err != nil {
						if err != context.Canceled {
							logger.Error().Err(err).Msg("Follower scheduler stopped unexpectedly")
							cancel() // Cancel main context to initiate shutdown
						}
					}
				}()
			}
		}

		if serverCancel != nil {
			serverCancel()
			serverWg.Wait()
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info().Msg("Shutting down...")
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()
}
