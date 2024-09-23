package client

import (
	"context"
	"fmt"
	"time"

	"github.com/nedson202/dts-go/pkg/logger"
	pb "github.com/nedson202/dts-go/proto/job/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type JobClient struct {
	client pb.JobServiceClient
	conn   *grpc.ClientConn
}

func NewJobClient(jobServiceAddr string) (*JobClient, error) {
	if jobServiceAddr == "" {
		return nil, fmt.Errorf("job service address is not set")
	}

	logger.Info().Msgf("Attempting to connect to job service at %s", jobServiceAddr)

	var conn *grpc.ClientConn
	var err error

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		conn, err = grpc.DialContext(ctx, jobServiceAddr,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithConnectParams(grpc.ConnectParams{
				Backoff: backoff.Config{
					BaseDelay:  100 * time.Millisecond,
					Multiplier: 1.6,
					Jitter:     0.2,
					MaxDelay:   3 * time.Second,
				},
				MinConnectTimeout: 5 * time.Second,
			}),
		)
		cancel()

		if err != nil {
			logger.Error().Err(err).Msgf("Failed to connect to job service. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		logger.Info().Msgf("Successfully connected to job service at %s", jobServiceAddr)
		break
	}

	return &JobClient{
		client: pb.NewJobServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *JobClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *JobClient) UpdateJob(ctx context.Context, id string, status pb.JobStatus, lastRun time.Time) (*pb.JobResponse, error) {
	// handle lastRun as optional
	var lastRunPb *timestamppb.Timestamp
	if !lastRun.IsZero() {
		lastRunPb = timestamppb.New(lastRun)
	}

	return c.client.UpdateJob(ctx, &pb.UpdateJobRequest{
		Id:     id,
		Status:   status,
		LastRun:  lastRunPb,
	})
}
