package database

const (
	CreateKeyspaceQuery = `
		CREATE KEYSPACE IF NOT EXISTS %s 
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
	`

	CreateJobsTableQuery = `
		CREATE TABLE IF NOT EXISTS jobs (
			id UUID PRIMARY KEY,
			name TEXT,
			description TEXT,
			cron_expression TEXT,
			status TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			metadata MAP<TEXT, TEXT>
		)
	`
)
