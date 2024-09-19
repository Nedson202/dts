package utils

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func CalculateNextRun(cronExpression string, from time.Time) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpression)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse cron expression: %w", err)
	}

	nextRun := schedule.Next(from)
	// Truncate to minute precision
	return nextRun.Truncate(time.Minute), nil
}

// ValidateCronExpression checks if the given cron expression is valid
func ValidateCronExpression(expression string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expression)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	return nil
}
