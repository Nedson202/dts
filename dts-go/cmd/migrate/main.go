package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
)

func main() {
	logger.Init()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	// Initialize Cassandra client
	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Cassandra client")
	}
	defer cassandraClient.Close()

	// Ensure the migrations table exists
	err = createMigrationsTable(cassandraClient)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create migrations table")
	}

	// Get list of migration files
	files, err := os.ReadDir("migrations")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read migrations directory")
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort migration files
	sort.Strings(migrationFiles)

	// Execute migrations
	for _, file := range migrationFiles {
		if err := executeMigration(cassandraClient, file); err != nil {
			logger.Error().Err(err).Msgf("Failed to execute migration %s", file)
			// Continue with the next migration instead of stopping
			continue
		}
	}

	logger.Info().Msg("Migration process completed")
}

func createMigrationsTable(client *database.CassandraClient) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id text PRIMARY KEY,
			applied_at timestamp
		)`
	return client.Session.Query(query).Exec()
}

func executeMigration(client *database.CassandraClient, filename string) error {
	// Check if migration has already been applied
	var count int
	if err := client.Session.Query("SELECT COUNT(*) FROM migrations WHERE id = ?", filename).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		logger.Info().Msgf("Migration %s has already been applied, skipping", filename)
		return nil
	}

	// Read migration file
	content, err := os.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return err
	}

	// Execute migration
	for _, statement := range strings.Split(string(content), ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if err := client.Session.Query(statement).Exec(); err != nil {
			// Check if the error is because the column already exists
			if strings.Contains(err.Error(), "already exists") {
				logger.Warn().Msgf("Column already exists, skipping statement: %s", statement)
				continue
			}
			logger.Error().Err(err).Msgf("Error executing statement: %s", statement)
			return err
		}
	}

	// Record migration as applied
	if err := client.Session.Query("INSERT INTO migrations (id, applied_at) VALUES (?, ?)", filename, time.Now()).Exec(); err != nil {
		return err
	}

	logger.Info().Msgf("Successfully applied migration: %s", filename)
	return nil
}
