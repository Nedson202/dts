package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nedson202/dts-go/pkg/config"
	"github.com/nedson202/dts-go/pkg/database"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize Cassandra client
	cassandraClient, err := database.NewCassandraClient(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}
	defer cassandraClient.Close()

	// Ensure the migrations table exists
	err = createMigrationsTable(cassandraClient)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get list of migration files
	files, err := ioutil.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".cql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort migration files
	sort.Strings(migrationFiles)

	// Execute migrations
	for _, file := range migrationFiles {
		if err := executeMigration(cassandraClient, file); err != nil {
			log.Printf("Failed to execute migration %s: %v", file, err)
			// Continue with the next migration instead of stopping
			continue
		}
	}

	fmt.Println("Migration process completed")
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
		fmt.Printf("Migration %s has already been applied, skipping\n", filename)
		return nil
	}

	// Read migration file
	content, err := ioutil.ReadFile(filepath.Join("migrations", filename))
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
				fmt.Printf("Column already exists, skipping statement: %s\n", statement)
				continue
			}
			fmt.Printf("Error executing statement: %s\nError: %v\n", statement, err)
			return err
		}
	}

	// Record migration as applied
	if err := client.Session.Query("INSERT INTO migrations (id, applied_at) VALUES (?, ?)", filename, time.Now()).Exec(); err != nil {
		return err
	}

	fmt.Printf("Successfully applied migration: %s\n", filename)
	return nil
}
