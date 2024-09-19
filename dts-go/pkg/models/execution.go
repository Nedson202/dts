package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	pb "github.com/nedson202/dts-go/proto/execution/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Execution struct {
	ID        gocql.UUID `json:"id"`
	JobID     gocql.UUID `json:"job_id"`
	Status    string     `json:"status"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"` // Change this to a pointer
	Result    string     `json:"result"`
	Error     string     `json:"error"`
}

func (e *Execution) ToProto() *pb.ExecutionResponse {
	resp := &pb.ExecutionResponse{
		Id:        e.ID.String(),
		JobId:     e.JobID.String(),
		Status:    e.Status,
		StartTime: timestamppb.New(e.StartTime),
		Result:    e.Result,
		Error:     e.Error,
	}
	if e.EndTime != nil {
		resp.EndTime = timestamppb.New(*e.EndTime)
	}
	return resp
}

func CreateExecution(client *database.CassandraClient, execution *Execution) error {
	query := `INSERT INTO job_executions (id, job_id, status, start_time, end_time, result, error) VALUES (?, ?, ?, ?, ?, ?, ?)`
	return client.Session.Query(query, execution.ID, execution.JobID, execution.Status, execution.StartTime, execution.EndTime, execution.Result, execution.Error).Exec()
}

func GetExecution(client *database.CassandraClient, id gocql.UUID) (*Execution, error) {
	var execution Execution
	var endTime time.Time
	query := `SELECT id, job_id, status, start_time, end_time, result, error FROM job_executions WHERE id = ?`
	err := client.Session.Query(query, id).Scan(&execution.ID, &execution.JobID, &execution.Status, &execution.StartTime, &endTime, &execution.Result, &execution.Error)
	if err != nil {
		return nil, err
	}
	if !endTime.IsZero() {
		execution.EndTime = &endTime
	}
	return &execution, nil
}

func ListExecutions(client *database.CassandraClient, pageSize int, lastID gocql.UUID, jobID, status string) ([]*Execution, error) {
	var executions []*Execution
	var query string
	var args []interface{}

	if jobID != "" && status != "" {
		query = `SELECT id, job_id, status, start_time, end_time, result, error FROM job_executions WHERE job_id = ? AND status = ? AND id > ? ORDER BY id DESC LIMIT ?`
		args = []interface{}{jobID, status, lastID, pageSize}
	} else if jobID != "" {
		query = `SELECT id, job_id, status, start_time, end_time, result, error FROM job_executions WHERE job_id = ? AND id > ? ORDER BY id DESC LIMIT ?`
		args = []interface{}{jobID, lastID, pageSize}
	} else if status != "" {
		query = `SELECT id, job_id, status, start_time, end_time, result, error FROM job_executions WHERE status = ? AND id > ? ORDER BY id DESC LIMIT ?`
		args = []interface{}{status, lastID, pageSize}
	} else {
		query = `SELECT id, job_id, status, start_time, end_time, result, error FROM job_executions WHERE id > ? ORDER BY id DESC LIMIT ?`
		args = []interface{}{lastID, pageSize}
	}

	iter := client.Session.Query(query, args...).Iter()
	for {
		var execution Execution
		if !iter.Scan(&execution.ID, &execution.JobID, &execution.Status, &execution.StartTime, &execution.EndTime, &execution.Result, &execution.Error) {
			break
		}
		executions = append(executions, &execution)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return executions, nil
}

func UpdateExecution(client *database.CassandraClient, execution *Execution) error {
	query := `UPDATE job_executions SET status = ?, end_time = ?, result = ?, error = ? WHERE id = ? AND job_id = ?`
	return client.Session.Query(query, execution.Status, execution.EndTime, execution.Result, execution.Error, execution.ID, execution.JobID).Exec()
}
