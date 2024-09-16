package models

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	pb "github.com/nedson202/dts-go/pkg/job"
	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Job struct {
	ID             gocql.UUID
	Name           string
	Description    string
	CronExpression string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastRun        time.Time
	Metadata       map[string]string
	NextRun        time.Time
}

func (j *Job) ToProto() *pb.Job {
	return &pb.Job{
		Id:             j.ID.String(),
		Name:           j.Name,
		Description:    j.Description,
		CronExpression: j.CronExpression,
		Status:         pb.JobStatus(pb.JobStatus_value[j.Status]), // Convert string to enum
		CreatedAt:      timestamppb.New(j.CreatedAt),
		UpdatedAt:      timestamppb.New(j.UpdatedAt),
		 LastRun:        timestamppb.New(j.LastRun),
		Metadata:       j.Metadata,
	}
}

func JobFromProto(pbJob *pb.Job) (*Job, error) {
	id, err := gocql.ParseUUID(pbJob.Id)
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:             id,
		Name:           pbJob.Name,
		Description:    pbJob.Description,
		CronExpression: pbJob.CronExpression,
		Status:         pbJob.Status.String(), // Convert enum to string
		CreatedAt:      pbJob.CreatedAt.AsTime(),
		UpdatedAt:      pbJob.UpdatedAt.AsTime(),
		 LastRun:        pbJob.LastRun.AsTime(),
		Metadata:       pbJob.Metadata,
	}, nil
}

func CreateJob(cassandraClient *database.CassandraClient, job *Job) error {
	if job.Status == pb.JobStatus_JOB_STATUS_UNSPECIFIED.String() {
		job.Status = pb.JobStatus_JOB_STATUS_PENDING.String()
	}
	job.NextRun = calculateNextRun(job.CronExpression, time.Now())
	return cassandraClient.Session.Query(
		"INSERT INTO jobs (id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		job.ID, job.Name, job.Description, job.CronExpression, job.Status, job.CreatedAt, job.UpdatedAt, job.LastRun, job.NextRun, job.Metadata,
	).Exec()
}

func GetJob(cassandraClient *database.CassandraClient, id gocql.UUID) (*Job, error) {
	var job Job
	err := cassandraClient.Session.Query(
		"SELECT id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata FROM jobs WHERE id = ?",
		id,
	).Scan(&job.ID, &job.Name, &job.Description, &job.CronExpression, &job.Status, &job.CreatedAt, &job.UpdatedAt, &job.LastRun, &job.NextRun, &job.Metadata)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func ListJobs(cassandraClient *database.CassandraClient, pageSize int, lastID gocql.UUID, status string) ([]*Job, error) {
	var jobs []*Job
	var query string
	var args []interface{}

	nilUUID := gocql.UUID{}
	if status != "" {
		if lastID != nilUUID {
			query = "SELECT id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata FROM jobs WHERE status_text = ? AND token(id) > token(?) LIMIT ? ALLOW FILTERING"
			args = []interface{}{status, lastID, pageSize}
		} else {
			query = "SELECT id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata FROM jobs WHERE status_text = ? LIMIT ? ALLOW FILTERING"
			args = []interface{}{status, pageSize}
		}
	} else {
		if lastID != nilUUID {
			query = "SELECT id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata FROM jobs WHERE token(id) > token(?) LIMIT ?"
			args = []interface{}{lastID, pageSize}
		} else {
			query = "SELECT id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata FROM jobs LIMIT ?"
			args = []interface{}{pageSize}
		}
	}

	iter := cassandraClient.Session.Query(query, args...).Iter()
	for {
		var job Job
		var statusStr string
		if !iter.Scan(&job.ID, &job.Name, &job.Description, &job.CronExpression, &statusStr, &job.CreatedAt, &job.UpdatedAt, &job.LastRun, &job.NextRun, &job.Metadata) {
			break
		}
		job.Status = statusStr
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func UpdateJob(cassandraClient *database.CassandraClient, job *Job) error {
	job.NextRun = calculateNextRun(job.CronExpression, time.Now())
	return cassandraClient.Session.Query(
		"UPDATE jobs SET name = ?, description = ?, cron_expression = ?, status_text = ?, updated_at = ?, last_run = ?, next_run = ?, metadata = ? WHERE id = ?",
		job.Name, job.Description, job.CronExpression, job.Status, job.UpdatedAt, job.LastRun, job.NextRun, job.Metadata, job.ID,
	).Exec()
}

func DeleteJob(cassandraClient *database.CassandraClient, id gocql.UUID) error {
	return cassandraClient.Session.Query("DELETE FROM jobs WHERE id = ?", id).Exec()
}

func GetPendingJobs(client *database.CassandraClient, limit int) ([]*Job, error) {
	query := "SELECT id, name, description, cron_expression, status_text, created_at, updated_at, last_run, next_run, metadata FROM jobs WHERE status_text = ? AND next_run <= ? LIMIT ? ALLOW FILTERING"
	iter := client.Session.Query(query, pb.JobStatus_JOB_STATUS_PENDING.String(), time.Now(), limit).Iter()
	var jobs []*Job
	var job Job
	for iter.Scan(&job.ID, &job.Name, &job.Description, &job.CronExpression, &job.Status, &job.CreatedAt, &job.UpdatedAt, &job.LastRun, &job.NextRun, &job.Metadata) {
		jobs = append(jobs, &job)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func UpdateJobLastRun(client *database.CassandraClient, jobID gocql.UUID, lastRun time.Time) error {
	query := "UPDATE jobs SET last_run = ? WHERE id = ?"
	return client.Session.Query(query, lastRun, jobID).Exec()
}

func calculateNextRun(cronExpression string, from time.Time) time.Time {
	schedule, err := cron.ParseStandard(cronExpression)
	if err != nil {
		log.Printf("Error parsing cron expression: %v", err)
		return time.Time{}
	}
	return schedule.Next(from)
}
