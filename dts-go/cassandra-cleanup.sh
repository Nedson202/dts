#!/bin/bash
set -e

echo "Cleaning up Cassandra snapshots..."
nodetool clearsnapshot

echo "Cleaning up Cassandra data..."
find /var/lib/cassandra/data/task_scheduler -name "*-Compacted" -type d -exec rm -rf {} +

echo "Cleanup completed."
