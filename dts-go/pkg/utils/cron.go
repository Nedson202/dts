package utils

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// CalculateNextRun calculates the next run time based on a cron expression and a reference time
func CalculateNextRun(cronExpression string, from time.Time) time.Time {
	schedule, err := cron.ParseStandard(cronExpression)
	if err != nil {
		log.Printf("Error parsing cron expression: %v", err)
		return time.Time{}
	}
	return schedule.Next(from)
}
