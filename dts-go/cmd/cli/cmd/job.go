package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nedson202/dts-go/pkg/logger"
	jobv1 "github.com/nedson202/dts-go/proto/job/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Interact with the Job service",
	Long:  `Create, read, update, list, and delete jobs using the Job service.`,
}

var createJobCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new job",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		cronExpression, _ := cmd.Flags().GetString("cron")
		metadata, _ := cmd.Flags().GetString("metadata")

		metadataMap := make(map[string]string)
		if metadata != "" {
			if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
				logger.Fatal().Err(err).Msg("Failed to parse metadata")
			}
		}

		conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect")
		}
		defer conn.Close()

		client := jobv1.NewJobServiceClient(conn)

		resp, err := client.CreateJob(context.Background(), &jobv1.CreateJobRequest{
			Name:           name,
			Description:    description,
			CronExpression: cronExpression,
			Metadata:       metadataMap,
		})

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create job")
		}

		fmt.Printf("Job created with ID: %s\n", resp.JobId)
	},
}

var getJobCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a job by ID",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")

		conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect")
		}
		defer conn.Close()

		client := jobv1.NewJobServiceClient(conn)

		resp, err := client.GetJob(context.Background(), &jobv1.GetJobRequest{
			Id: id,
		})

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to get job")
		}

		printJobResponse(resp)
	},
}

var listJobsCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs",
	Run: func(cmd *cobra.Command, args []string) {
		pageSize, _ := cmd.Flags().GetInt32("page-size")
		if pageSize <= 0 || pageSize > 250 {
			pageSize = 250
		}
		status, _ := cmd.Flags().GetString("status")
		lastID, _ := cmd.Flags().GetString("last-id")

		conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect")
		}
		defer conn.Close()

		client := jobv1.NewJobServiceClient(conn)

		resp, err := client.ListJobs(context.Background(), &jobv1.ListJobsRequest{
			PageSize: pageSize,
			Status:   status,
			LastId:   lastID,
		})

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to list jobs")
		}

		fmt.Printf("Total jobs: %d\n", resp.Total)
		for _, j := range resp.Jobs {
			printJobResponse(j)
			fmt.Println("---")
		}
		fmt.Printf("Next page token: %s\n", resp.NextPage)
	},
}

var updateJobCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a job",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		cronExpression, _ := cmd.Flags().GetString("cron")
		status, _ := cmd.Flags().GetString("status")
		metadata, _ := cmd.Flags().GetString("metadata")

		metadataMap := make(map[string]string)
		if metadata != "" {
			if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
				logger.Fatal().Err(err).Msg("Failed to parse metadata")
			}
		}

		conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect")
		}
		defer conn.Close()

		client := jobv1.NewJobServiceClient(conn)

		resp, err := client.UpdateJob(context.Background(), &jobv1.UpdateJobRequest{
			Id:             id,
			Name:           name,
			Description:    description,
			CronExpression: cronExpression,
			Status:         jobv1.JobStatus(jobv1.JobStatus_value[status]),
			Metadata:       metadataMap,
		})

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to update job")
		}

		printJobResponse(resp)
	},
}

var deleteJobCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a job",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")

		conn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect")
		}
		defer conn.Close()

		client := jobv1.NewJobServiceClient(conn)

		resp, err := client.DeleteJob(context.Background(), &jobv1.DeleteJobRequest{
			Id: id,
		})

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to delete job")
		}

		fmt.Printf("Job deleted: %v\n", resp.Success)
	},
}

func init() {
	jobCmd.AddCommand(createJobCmd)
	jobCmd.AddCommand(getJobCmd)
	jobCmd.AddCommand(listJobsCmd)
	jobCmd.AddCommand(updateJobCmd)
	jobCmd.AddCommand(deleteJobCmd)

	createJobCmd.Flags().String("name", "", "Name of the job")
	createJobCmd.Flags().String("description", "", "Description of the job")
	createJobCmd.Flags().String("cron", "", "Cron expression for the job")
	createJobCmd.Flags().String("metadata", "", "Metadata for the job (JSON format)")

	getJobCmd.Flags().String("id", "", "ID of the job")

	listJobsCmd.Flags().Int32("page-size", 250, "Page size (1-250, default 250)")
	listJobsCmd.Flags().String("status", "", "Status filter")
	listJobsCmd.Flags().String("last-id", "", "Last ID for pagination")

	updateJobCmd.Flags().String("id", "", "ID of the job")
	updateJobCmd.Flags().String("name", "", "Name of the job")
	updateJobCmd.Flags().String("description", "", "Description of the job")
	updateJobCmd.Flags().String("cron", "", "Cron expression for the job")
	updateJobCmd.Flags().String("status", "", "Status of the job")
	updateJobCmd.Flags().String("metadata", "", "Metadata for the job (JSON format)")

	deleteJobCmd.Flags().String("id", "", "ID of the job")
}

func printJobResponse(j *jobv1.JobResponse) {
	m := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	jsonBytes, err := m.Marshal(j)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to marshal job to JSON")
	}
	fmt.Println(string(jsonBytes))
}
