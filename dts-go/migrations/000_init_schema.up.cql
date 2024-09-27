CREATE KEYSPACE IF NOT EXISTS task_scheduler
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

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

CREATE TABLE IF NOT EXISTS task_schedule (
    next_execution_time timestamp,
    segment text,
    job_id uuid,
    PRIMARY KEY ((segment, next_execution_time), job_id)
) WITH CLUSTERING ORDER BY (job_id ASC);

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
