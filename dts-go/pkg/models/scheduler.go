package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	pb "github.com/nedson202/dts-go/pkg/scheduler"
)

type ScheduledJob struct {
	JobID     gocql.UUID `json:"job_id"`
	Job       Job        `json:"job"`
	Resources Resources  `json:"resources"`
	StartTime time.Time  `json:"start_time"`
}

type Resources struct {
	CPU     int32 `json:"cpu"`
	Memory  int32 `json:"memory"`
	Storage int32 `json:"storage"`
}

func SaveScheduledJob(client *database.CassandraClient, job *ScheduledJob) error {
	jobData, err := json.Marshal(job.Job)
	if err != nil {
		return err
	}
	return client.Session.Query(`
		INSERT INTO scheduled_jobs (id, job_data, cpu, memory, storage, start_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`, job.JobID, string(jobData), job.Resources.CPU, job.Resources.Memory, job.Resources.Storage, job.StartTime).Exec()
}

func GetScheduledJob(client *database.CassandraClient, jobID gocql.UUID) (*ScheduledJob, error) {
	var job ScheduledJob
	var jobData string
	err := client.Session.Query(`
		SELECT id, job_data, cpu, memory, storage, start_time 
		FROM scheduled_jobs 
		WHERE id = ?
	`, jobID).Scan(&job.JobID, &jobData, &job.Resources.CPU, &job.Resources.Memory, &job.Resources.Storage, &job.StartTime)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(jobData), &job.Job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func DeleteScheduledJob(client *database.CassandraClient, jobID gocql.UUID) error {
	return client.Session.Query(`
		DELETE FROM scheduled_jobs 
		WHERE id = ?
	`, jobID).Exec()
}

func ListScheduledJobs(client *database.CassandraClient, pageSize int, lastID gocql.UUID) ([]*ScheduledJob, error) {
	var query string
	var args []interface{}

	nilUUID := gocql.UUID{}
	if lastID == nilUUID {
		query = `
			SELECT id, job_data, resources, start_time 
			FROM scheduled_jobs 
			LIMIT ?
		`
		args = []interface{}{pageSize}
	} else {
		query = `
			SELECT id, job_data, resources, start_time 
			FROM scheduled_jobs 
			WHERE token(id) > token(?) 
			LIMIT ?
		`
		args = []interface{}{lastID, pageSize}
	}

	iter := client.Session.Query(query, args...).Iter()
	var jobs []*ScheduledJob
	var job ScheduledJob
	var jobData string
	for iter.Scan(&job.JobID, &jobData, &job.Resources.CPU, &job.Resources.Memory, &job.Resources.Storage, &job.StartTime) {
		err := json.Unmarshal([]byte(jobData), &job.Job)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func GetAvailableResources(client *database.CassandraClient) (*Resources, error) {
	var resources Resources
	err := client.Session.Query(`
		SELECT cpu, memory, storage FROM available_resources WHERE id = 'global'
	`).Scan(&resources.CPU, &resources.Memory, &resources.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to get available resources: %v", err)
	}
	return &resources, nil
}

func AllocateResources(client *database.CassandraClient, required Resources) error {
	// First, get the current available resources
	available, err := GetAvailableResources(client)
	if err != nil {
		return fmt.Errorf("failed to get available resources: %v", err)
	}

	// Check if we have enough resources
	if available.CPU < required.CPU || available.Memory < required.Memory || available.Storage < required.Storage {
		return fmt.Errorf("insufficient resources")
	}

	// If we have enough, update the resources
	newCPU := available.CPU - required.CPU
	newMemory := available.Memory - required.Memory
	newStorage := available.Storage - required.Storage

	err = client.Session.Query(`
		UPDATE available_resources
		SET cpu = ?, memory = ?, storage = ?
		WHERE id = 'global'
	`, newCPU, newMemory, newStorage).Exec()

	if err != nil {
		return fmt.Errorf("failed to update resources: %v", err)
	}

	return nil
}

func ReleaseResources(client *database.CassandraClient, resources Resources) error {
	// First, get the current available resources
	available, err := GetAvailableResources(client)
	if err != nil {
		return fmt.Errorf("failed to get available resources: %v", err)
	}

	// Calculate new resource values
	newCPU := available.CPU + resources.CPU
	newMemory := available.Memory + resources.Memory
	newStorage := available.Storage + resources.Storage

	// Update the resources
	err = client.Session.Query(`
		UPDATE available_resources
		SET cpu = ?, memory = ?, storage = ?
		WHERE id = 'global'
	`, newCPU, newMemory, newStorage).Exec()

	if err != nil {
		return fmt.Errorf("failed to release resources: %v", err)
	}

	return nil
}

func (s *ScheduledJob) ToProto() *pb.ScheduledJob {
	return &pb.ScheduledJob{
		JobId:                s.JobID.String(),
		NextExecutionTime:    s.StartTime.Format(time.RFC3339),
		ResourceRequirements: s.Resources.ToProto(),
	}
}

func (r *Resources) ToProto() *pb.Resources {
	return &pb.Resources{
		Cpu:     r.CPU,
		Memory:  r.Memory,
		Storage: r.Storage,
	}
}

func ScheduledJobFromProto(p *pb.ScheduledJob) (*ScheduledJob, error) {
	jobID, err := gocql.ParseUUID(p.JobId)
	if err != nil {
		return nil, err
	}
	startTime, err := time.Parse(time.RFC3339, p.NextExecutionTime)
	if err != nil {
		return nil, err
	}
	return &ScheduledJob{
		JobID:     jobID,
		StartTime: startTime,
		Resources: ResourcesFromProto(p.ResourceRequirements),
	}, nil
}

func ResourcesFromProto(p *pb.Resources) Resources {
	return Resources{
		CPU:     p.Cpu,
		Memory:  p.Memory,
		Storage: p.Storage,
	}
}
