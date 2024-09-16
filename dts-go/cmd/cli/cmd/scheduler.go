package cmd

import (
	"github.com/spf13/cobra"
)

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Interact with the Scheduler service",
	Long:  `Schedule jobs and retrieve schedules using the Scheduler service.`,
}

// Add subcommands for scheduler operations here
