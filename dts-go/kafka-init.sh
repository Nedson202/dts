#!/bin/bash
set -e

# Wait for Kafka to be ready
echo "Waiting for Kafka to be ready..."
cub kafka-ready -b kafka:29092 1 30

# Create topics
echo "Creating Kafka topics..."
kafka-topics --create --if-not-exists --bootstrap-server kafka:29092 --replication-factor 1 --partitions 1 --topic jobs
kafka-topics --create --if-not-exists --bootstrap-server kafka:29092 --replication-factor 1 --partitions 1 --topic scheduled-jobs
kafka-topics --create --if-not-exists --bootstrap-server kafka:29092 --replication-factor 1 --partitions 1 --topic job-executions

echo "Kafka topics created."
