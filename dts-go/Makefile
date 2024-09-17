.PHONY: proto build run docker-build docker-run docker-stop cli dev dev-job dev-scheduler dev-execution migrate

proto:
	buf generate

build: proto
	go build -o bin/job-service ./cmd/job-service
	go build -o bin/scheduler-service ./cmd/scheduler-service
	go build -o bin/execution-service ./cmd/execution-service

run: build
	./bin/job-service & ./bin/scheduler-service & ./bin/execution-service

docker-build:
	docker-compose build

docker-run:
	docker-compose up

docker-stop:
	docker-compose down

test:
	go test ./...

clean:
	rm -rf bin
	docker-compose down -v

cli:
	go build -o bin/dts-cli ./cmd/cli

run-job-service:
	go run cmd/job-service/main.go

migrate:
	go run cmd/migrate/main.go

# Update these commands to use 'docker compose' instead of 'docker-compose'
dev:
	docker compose -f docker-compose.dev.yml up -d

dev-down:
	docker compose -f docker-compose.dev.yml down

dev-logs:
	docker compose -f docker-compose.dev.yml logs -f

dev-build:
	docker compose -f docker-compose.dev.yml build

dev-restart:
	docker compose -f docker-compose.dev.yml restart

migrate-up:
	migrate -path ./migrations -database cassandra://${CASSANDRA_HOSTS}/${CASSANDRA_KEYSPACE} up

migrate-down:
	migrate -path ./migrations -database cassandra://${CASSANDRA_HOSTS}/${CASSANDRA_KEYSPACE} down

migrate-create:
	migrate create -ext cql -dir ./migrations $(name)
