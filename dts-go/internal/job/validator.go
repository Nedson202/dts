package job

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

// ValidateCronExpression checks if the given cron expression is valid
func ValidateCronExpression(cronExpr string) error {
    parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
    _, err := parser.Parse(cronExpr)
    if err != nil {
        return fmt.Errorf("invalid cron expression: %w", err)
    }
    return nil
}
