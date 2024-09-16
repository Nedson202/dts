package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	KafkaBrokers       []string
	JobTopic           string
	SchedulerTopic     string
	ExecutionTopic     string
	CassandraHosts     []string
	CassandraKeyspace  string
	SchedulerServicePort string
	ExecutionServicePort string
	JobServiceHost string
	JobServiceGRPCPort string
	JobServiceHTTPPort string
	SchedulerServiceGRPCPort string
	SchedulerServiceHTTPPort string
	CassandraDataRetentionDays int
}

func LoadConfig() *Config {
	return &Config{
		KafkaBrokers: getEnvAsSlice("KAFKA_BROKERS", []string{"kafka:29092"}),
		JobTopic:           getEnv("KAFKA_JOB_TOPIC", "jobs"),
		SchedulerTopic:     getEnv("KAFKA_SCHEDULER_TOPIC", "scheduled-jobs"),
		ExecutionTopic:     getEnv("KAFKA_EXECUTION_TOPIC", "job-executions"),
		CassandraHosts:     getEnvAsSlice("CASSANDRA_HOSTS", []string{"localhost"}),
		CassandraKeyspace:  getEnv("CASSANDRA_KEYSPACE", "task_scheduler"),
		SchedulerServicePort: getEnv("SCHEDULER_SERVICE_PORT", ":50052"),
		ExecutionServicePort: getEnv("EXECUTION_SERVICE_PORT", ":50053"),
		JobServiceHost: getEnv("JOB_SERVICE_HOST", "localhost"),
		JobServiceGRPCPort: getEnv("JOB_SERVICE_GRPC_PORT", "50054"),
		JobServiceHTTPPort: getEnv("JOB_SERVICE_HTTP_PORT", "8080"),
		SchedulerServiceGRPCPort: getEnv("SCHEDULER_SERVICE_GRPC_PORT", "50052"),
		SchedulerServiceHTTPPort: getEnv("SCHEDULER_SERVICE_HTTP_PORT", "8081"),
		CassandraDataRetentionDays: getEnvAsInt("CASSANDRA_DATA_RETENTION_DAYS", 30),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (c *Config) GetKafkaBrokers() []string {
	return c.KafkaBrokers
}
