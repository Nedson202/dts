CREATE KEYSPACE IF NOT EXISTS task_scheduler
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE task_scheduler;

-- Create the jobs table
CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY,
  name TEXT,
  description TEXT,
  cron_expression TEXT,
  status TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  last_run TIMESTAMP,
  next_run TIMESTAMP,
  metadata MAP<TEXT, TEXT>
);

-- Create the scheduled_jobs table
CREATE TABLE IF NOT EXISTS scheduled_jobs (
  id UUID PRIMARY KEY,
  job_id UUID,
  start_time TIMESTAMP,
  cpu INT,
  memory INT,
  storage INT
);

-- Create an index on start_time for efficient querying
CREATE INDEX IF NOT EXISTS scheduled_jobs_start_time_idx ON scheduled_jobs (start_time);

-- Create the job_executions table
CREATE TABLE IF NOT EXISTS job_executions (
  id UUID,
  job_id UUID,
  status TEXT,
  start_time TIMESTAMP,
  end_time TIMESTAMP,
  result TEXT,
  PRIMARY KEY ((job_id), id)
);
