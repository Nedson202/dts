# DTS-Go: Distributed Task Scheduler

DTS-Go is a robust, scalable distributed task scheduler system built with Go. It provides a flexible and efficient way to schedule, manage, and execute tasks across distributed systems.

## Features

- Job creation, retrieval, updating, listing, and deletion
- Task scheduling with cron expressions
- Resource allocation and management
- Distributed execution of tasks
- Scalable architecture using Kafka and Cassandra
- gRPC and HTTP API support
- CLI tool for easy interaction with the system

## Architecture

DTS-Go consists of several microservices:

1. Job Service: Manages job CRUD operations
2. Scheduler Service: Handles task scheduling and resource allocation
3. Execution Service: Executes scheduled tasks

The system uses Apache Kafka for message queuing and Apache Cassandra for persistent storage.

## Prerequisites

- Go 1.16+
- Docker and Docker Compose
- Apache Kafka
- Apache Cassandra

## Getting Started

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/dts-go.git
   cd dts-go
   ```

2. Set up the environment variables in a `.env` file:
   ```
   KAFKA_BROKERS=localhost:9092
   CASSANDRA_HOSTS=localhost
   CASSANDRA_KEYSPACE=task_scheduler
   JOB_SERVICE_GRPC_PORT=50054
   JOB_SERVICE_HTTP_PORT=8080
   SCHEDULER_SERVICE_GRPC_PORT=50052
   SCHEDULER_SERVICE_HTTP_PORT=8081
   ```

3. Run database migrations:
   ```
   make migrate
   ```
   This will create necessary tables and apply any pending migrations to your Cassandra database.

4. Build and run the services using Docker Compose:
   ```
   docker-compose up --build
   ```

## Usage

### Using the CLI

The CLI tool provides an easy way to interact with the DTS-Go system. Here are some example commands:

1. Create a job:
   ```
   go run cmd/cli/main.go job create --name "My Job" --description "Description" --cron "*/5 * * * *" --metadata '{"key": "value"}'
   ```

2. Get a job:
   ```
   go run cmd/cli/main.go job get --id <job_id>
   ```

3. List jobs:
   ```
   go run cmd/cli/main.go job list --page-size 10 --status "active"
   ```

4. Update a job:
   ```
   go run cmd/cli/main.go job update --id <job_id> --name "Updated Job" --status "paused"
   ```

5. Delete a job:
   ```
   go run cmd/cli/main.go job delete --id <job_id>
   ```

### Using the API

The system exposes both gRPC and HTTP APIs. You can use tools like [grpcurl](https://github.com/fullstorydev/grpcurl) for gRPC or curl for HTTP to interact with the APIs.

Example HTTP request to create a job:

curl -X POST http://localhost:8080/v1/jobs \
-H "Content-Type: application/json" \
-d '{"name": "My Job", "description": "Description", "cron_expression": "/5 ", "metadata": {"key": "value"}}'

## Project Structure

- `cmd/`: Contains the main applications
  - `job-service/`: Job service implementation
  - `scheduler-service/`: Scheduler service implementation
  - `execution-service/`: Execution service implementation
  - `cli/`: Command-line interface tool
- `internal/`: Internal packages
  - `job/`: Job-related logic
  - `scheduler/`: Scheduler-related logic
- `pkg/`: Shared packages
  - `config/`: Configuration management
  - `database/`: Database clients and utilities
  - `models/`: Shared data models
  - `queue/`: Message queue clients and utilities
  - `services/`: gRPC service implementations
- `api/`: Protocol buffer definitions

## Development

### Adding a New Service

1. Create a new directory under `cmd/`
2. Implement the service logic
3. Add the service to `docker-compose.yml`
4. Update the `Makefile` if necessary

### Generating Protocol Buffers

After modifying the `.proto` files in the `api/proto/` directory, regenerate the Go code:

### Running Tests

Run the tests with:


## Configuration

The system can be configured using environment variables. Here are the available options:

- `KAFKA_BROKERS`: Comma-separated list of Kafka brokers (default: "localhost:9092")
- `CASSANDRA_HOSTS`: Comma-separated list of Cassandra hosts (default: "localhost")
- `CASSANDRA_KEYSPACE`: Cassandra keyspace name (default: "task_scheduler")
- `JOB_SERVICE_GRPC_PORT`: Job service gRPC port (default: "50054")
- `JOB_SERVICE_HTTP_PORT`: Job service HTTP port (default: "8080")
- `SCHEDULER_SERVICE_GRPC_PORT`: Scheduler service gRPC port (default: "50052")
- `SCHEDULER_SERVICE_HTTP_PORT`: Scheduler service HTTP port (default: "8081")

## API Documentation

### Job Service

- Create Job: `POST /v1/jobs`
- Get Job: `GET /v1/jobs/{id}`
- List Jobs: `GET /v1/jobs`
- Update Job: `PUT /v1/jobs/{id}`
- Delete Job: `DELETE /v1/jobs/{id}`

### Scheduler Service

- Schedule Job: `POST /v1/scheduler/jobs`
- Cancel Job: `DELETE /v1/scheduler/jobs/{job_id}`
- Get Scheduled Job: `GET /v1/scheduler/jobs/{job_id}`
- List Scheduled Jobs: `GET /v1/scheduler/jobs`

For detailed API documentation, please refer to the proto files in the `api/proto/` directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Apache Kafka](https://kafka.apache.org/)
- [Apache Cassandra](https://cassandra.apache.org/)

## Database Migrations

The project uses a simple migration system to manage database schema changes. Migrations are stored in the `migrations` directory and are executed in alphabetical order.

Current migrations:

1. `001_add_next_run_to_jobs.cql`: Adds the `next_run` column to the `jobs` table.
2. `002_add_last_run_to_jobs.cql`: Adds the `last_run` column to the `jobs` table.

To run migrations:

```
make migrate
```

To create a new migration:

1. Create a new `.cql` file in the `migrations` directory with a descriptive name, prefixed with a sequential number (e.g., `003_add_new_column.cql`).
2. Write your Cassandra CQL statements in the file.
3. Run the migration using the command above.

The migration system will keep track of applied migrations and only run new ones.
