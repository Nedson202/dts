-- Migration: Add next_run column to jobs table
-- Filename: 001_add_next_run_to_jobs.cql

-- Add next_run column
ALTER TABLE task_scheduler.jobs ADD next_run timestamp;

-- Create an index on next_run for efficient querying
CREATE INDEX IF NOT EXISTS jobs_next_run_idx ON task_scheduler.jobs (next_run);

