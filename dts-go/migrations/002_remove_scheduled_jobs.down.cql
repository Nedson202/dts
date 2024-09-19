-- Drop the index first
DROP INDEX IF EXISTS scheduled_jobs_status_idx;

-- Drop the scheduled_jobs table
DROP TABLE IF EXISTS scheduled_jobs;
