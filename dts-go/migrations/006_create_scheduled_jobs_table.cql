-- Migration: Create scheduled_jobs table
-- Filename: 006_create_scheduled_jobs_table.cql

-- Create the scheduled_jobs table
CREATE TABLE IF NOT EXISTS task_scheduler.scheduled_jobs (
    id uuid PRIMARY KEY,
    job_data text,
    cpu int,
    memory int,
    storage int,
    start_time timestamp
);

-- Create an index on start_time for efficient querying
CREATE INDEX IF NOT EXISTS scheduled_jobs_start_time_idx ON task_scheduler.scheduled_jobs (start_time);
