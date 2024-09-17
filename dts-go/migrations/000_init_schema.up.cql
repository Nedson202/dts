CREATE KEYSPACE IF NOT EXISTS task_scheduler
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

-- Create the jobs table
CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY,
  name TEXT,
  description TEXT,
  cron_expression TEXT,
  status_text TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  last_run TIMESTAMP,
  next_run TIMESTAMP,
  metadata MAP<TEXT, TEXT>
);

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
