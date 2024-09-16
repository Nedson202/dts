package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/nedson202/dts-go/pkg/database"
	jobpb "github.com/nedson202/dts-go/pkg/job"
	"github.com/nedson202/dts-go/pkg/models"
	pb "github.com/nedson202/dts-go/pkg/scheduler"
	// "github.com/nedson202/dts-go/pkg/utils"
	"google.golang.org/grpc"
)

type PeriodicScheduler struct {
	cassandraClient *database.CassandraClient
	service         *Service
	checkInterval   time.Duration
	jobClient       jobpb.JobServiceClient
}

func NewPeriodicScheduler(cassandraClient *database.CassandraClient, service *Service, checkInterval time.Duration, jobServiceAddr string) (*PeriodicScheduler, error) {
	conn, err := grpc.Dial(jobServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	jobClient := jobpb.NewJobServiceClient(conn)

	log.Printf("Initializing PeriodicScheduler with check interval: %v", checkInterval)
	return &PeriodicScheduler{
		cassandraClient: cassandraClient,
		service:         service,
		checkInterval:   checkInterval,
		jobClient:       jobClient,
	}, nil
}

func (ps *PeriodicScheduler) Start(ctx context.Context) {
	log.Println("Starting PeriodicScheduler")
	ticker := time.NewTicker(ps.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("PeriodicScheduler stopped due to context cancellation")
			return
		case <-ticker.C:
			log.Println("Running periodic job check")
			ps.checkAndScheduleJobs()
		}
	}
}

func (ps *PeriodicScheduler) checkAndScheduleJobs() {
	startTime := time.Now()
	log.Println("Fetching pending jobs")
	jobs, err := models.GetPendingJobs(ps.cassandraClient, 100) // Limit to 100 jobs per cycle
	if err != nil {
		log.Printf("Error fetching pending jobs: %v", err)
		return
	}
	log.Printf("Found %d pending jobs", len(jobs))

	scheduledCount := 0
	for _, job := range jobs {
		log.Printf("Processing job: %s", job.ID)
		// Schedule the job
		_, err := ps.service.ScheduleJob(context.Background(), &pb.ScheduleJobRequest{
			Job: job.ToProto(),
			ResourceRequirements: &pb.Resources{
				// Set default or fetch from job configuration
				Cpu:     1,
				Memory:  1024,
				Storage: 1024,
			},
		})
		if err != nil {
			log.Printf("Error scheduling job %s: %v", job.ID, err)
		} else {
			scheduledCount++
			// Update the job's status, last run time, and next run time
			// job.LastRun = time.Now()
			// job.NextRun = utils.CalculateNextRun(job.CronExpression, job.LastRun)
			// log.Printf("Job %s scheduled. Next run: %v", job.ID, job.NextRun)
			
			// // Update job via Job Service
			// _, err = ps.jobClient.UpdateJob(context.Background(), &jobpb.UpdateJobRequest{
			// 	Id:             job.ID.String(),
			// 	Status:         jobpb.JobStatus_JOB_STATUS_SCHEDULED,
			// 	LastRun:        job.LastRun.Unix(),
			// 	NextRun:        job.NextRun.Unix(),
			// })
			// if err != nil {
			// 	log.Printf("Error updating job %s via Job Service: %v", job.ID, err)
			// }
		}
	}

	duration := time.Since(startTime)
	log.Printf("Periodic job check completed. Scheduled %d out of %d jobs. Duration: %v", scheduledCount, len(jobs), duration)
}
