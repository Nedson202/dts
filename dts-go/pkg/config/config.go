package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	KafkaBrokers              []string
	TaskTopic                 string
	TaskRetryTopic            string
	CassandraHosts            []string
	CassandraKeyspace         string
	SchedulerServicePort      string
	ExecutionServiceGRPCPort  string
	ExecutionServiceHTTPPort  string
	JobServiceHost            string
	JobServiceGRPCPort        string
	JobServiceHTTPPort        string
	SchedulerServiceGRPCPort  string
	SchedulerServiceHTTPPort  string
	CassandraDataRetentionDays int
	JobServiceAddr            string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		KafkaBrokers:              getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		TaskTopic:                 getEnv("KAFKA_TASK_TOPIC", "jobs"),
		TaskRetryTopic:            getEnv("KAFKA_TASK_RETRY_TOPIC", "jobs-retry"),
		CassandraHosts:            getEnvAsSlice("CASSANDRA_HOSTS", []string{"localhost"}),
		CassandraKeyspace:         getEnv("CASSANDRA_KEYSPACE", "task_scheduler"),
		SchedulerServicePort:      getEnv("SCHEDULER_SERVICE_PORT", "50052"),
		ExecutionServiceGRPCPort:  getEnv("EXECUTION_SERVICE_GRPC_PORT", "50053"),
		ExecutionServiceHTTPPort:  getEnv("EXECUTION_SERVICE_HTTP_PORT", "8082"),
		JobServiceHost:            getEnv("JOB_SERVICE_HOST", "localhost"),
		JobServiceGRPCPort:        getEnv("JOB_SERVICE_GRPC_PORT", "50054"),
		JobServiceHTTPPort:        getEnv("JOB_SERVICE_HTTP_PORT", "8080"),
		SchedulerServiceGRPCPort:  getEnv("SCHEDULER_SERVICE_GRPC_PORT", "50052"),
		SchedulerServiceHTTPPort:  getEnv("SCHEDULER_SERVICE_HTTP_PORT", "8081"),
		CassandraDataRetentionDays: getEnvAsInt("CASSANDRA_DATA_RETENTION_DAYS", 30),
		JobServiceAddr:            getEnv("JOB_SERVICE_ADDR", "localhost:50054"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid integer value for %s, using default: %d\n", key, defaultValue)
		return defaultValue
	}
	return value
}
