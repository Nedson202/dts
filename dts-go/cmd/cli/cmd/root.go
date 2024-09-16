package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "dts-cli",
	Short: "CLI for interacting with the Distributed Task Scheduler",
	Long:  `A command-line interface for testing and interacting with the Job, Scheduler, and Execution services of the Distributed Task Scheduler.`,
}

func init() {
	RootCmd.AddCommand(jobCmd)
	RootCmd.AddCommand(schedulerCmd)
	RootCmd.AddCommand(executionCmd)
}
