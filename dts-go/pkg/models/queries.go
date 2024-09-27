package models

const (
	JobCreationQuery = `
		INSERT INTO jobs (id, name, description, cron_expression, status, created_at, updated_at, last_run, next_run, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	JobUpdateQuery = `
		UPDATE jobs 
        SET name = ?, description = ?, cron_expression = ?, status = ?, 
            updated_at = ?, last_run = ?, next_run = ?, metadata = ? 
        WHERE id = ?
    `
    TaskScheduleDeleteQuery = `
        DELETE FROM task_schedule 
        WHERE job_id = ? AND next_execution_time = ? AND segment = ?
    `
    TaskScheduleCreationQuery = `
        INSERT INTO task_schedule (next_execution_time, segment, job_id)
        VALUES (?, ?, ?)
    `
	TaskScheduleGetByJobIDQuery = `
		SELECT next_execution_time, segment, job_id
		FROM task_schedule 
		WHERE job_id = ? ALLOW FILTERING
	`
	TaskScheduleGetDueForExecutionQuery = `
		SELECT next_execution_time, segment, job_id
		FROM task_schedule 
		WHERE next_execution_time <= ? AND segment = ?
		ALLOW FILTERING
	`
)
