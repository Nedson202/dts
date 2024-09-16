package cmd

import (
	"github.com/spf13/cobra"
)

var executionCmd = &cobra.Command{
	Use:   "execution",
	Short: "Interact with the Execution service",
	Long:  `Execute jobs and retrieve execution status using the Execution service.`,
}

// Add subcommands for execution operations here
