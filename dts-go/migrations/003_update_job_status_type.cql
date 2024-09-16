-- Migration: Update job status type
-- Filename: 003_update_job_status_type.cql

-- Note: After this migration, you'll need to update your application code to use 'status_text' instead of 'status'
-- The application should handle populating the status_text column for existing and new jobs
-- Once the application is updated and deployed, you can create another migration to drop the old 'status' column

-- Add a new column for the text status
ALTER TABLE task_scheduler.jobs ADD status_text text;

-- Create an index on the new status column
CREATE INDEX IF NOT EXISTS jobs_status_text_idx ON task_scheduler.jobs (status_text);
