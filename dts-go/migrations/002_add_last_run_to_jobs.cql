-- Migration: Add last_run column to jobs table
-- Filename: 002_add_last_run_to_jobs.cql

-- Add last_run column
ALTER TABLE task_scheduler.jobs ADD last_run timestamp;

-- Create an index on last_run for efficient querying
CREATE INDEX IF NOT EXISTS jobs_last_run_idx ON task_scheduler.jobs (last_run);
