package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
)

type TaskSchedule struct {
	NextExecutionTime time.Time
	Segment           string
	JobID             gocql.UUID
}

func GetTaskScheduleByJobID(client *database.CassandraClient, jobID gocql.UUID) (*TaskSchedule, error) {
	var taskSchedule TaskSchedule
	err := client.Session.Query(TaskScheduleGetByJobIDQuery, jobID).Scan(&taskSchedule.NextExecutionTime, &taskSchedule.Segment, &taskSchedule.JobID)
	if err != nil {
		return nil, err
	}
	return &taskSchedule, nil
}

// GetScheduledTasksDueForExecution retrieves jobs that are due for execution based on the current time and specified segments
func GetScheduledTasksDueForExecution(client *database.CassandraClient, segments []string) ([]*TaskSchedule, error) {
	nowTruncated := time.Now().Truncate(time.Minute)
	var scheduledTasks []*TaskSchedule

	for _, segment := range segments {
		logger.Info().Msgf("Fetching scheduled tasks for segment: %s", segment)
		logger.Info().Msgf("Type of segment: %T", segment)
		iter := client.Session.Query(TaskScheduleGetDueForExecutionQuery, nowTruncated, segment).Iter()
		for {
			var scheduledTask TaskSchedule
			if !iter.Scan(&scheduledTask.NextExecutionTime, &scheduledTask.Segment, &scheduledTask.JobID) {
				break
			}
			scheduledTasks = append(scheduledTasks, &scheduledTask)
		}
		if err := iter.Close(); err != nil {
			return nil, err
		}
	}

	return scheduledTasks, nil
}
