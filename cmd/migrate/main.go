package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
)

func main() {
	var (
		action   = flag.String("action", "", "Migration action: up, down, create")
		name     = flag.String("name", "", "Migration name (required for create)")
		dbURL    = flag.String("db", "", "Database URL (overrides config)")
		mPath    = flag.String("path", "./db/migrations", "Migrations directory path")
	)
	flag.Parse()

	if *action == "" {
		fmt.Println("Usage: migrate -action=<up|down|create> [-name=<migration_name>] [-db=<database_url>] [-path=<migrations_path>]")
		os.Exit(1)
	}

	// Default database URL if not provided
	databaseURL := *dbURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			databaseURL = "postgres://postgres:password@localhost/taishanglaojun?sslmode=disable"
		}
	}

	// Connect to database
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Execute action
	switch *action {
	case "up":
		if err := migrateUp(db, *mPath); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := migrateDown(db, *mPath); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create action")
		}
		if err := createMigration(*mPath, *name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

// Migrator represents a simple database migrator
type Migrator struct {
	db        *sql.DB
	mPath     string
	applied   map[string]bool
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB, mPath string) *Migrator {
	return &Migrator{
		db:    db,
		mPath: mPath,
	}
}

// Up applies all pending migrations
func (m *Migrator) Up() error {
	// Create migrations table if not exists
	if _, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get applied migrations
	rows, err := m.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %v", err)
		}
		applied[version] = true
	}

	// Get migration files
	files, err := ioutil.ReadDir(m.mPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort files by name (which should include version prefix)
	var migrationFiles []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		migrationFiles = append(migrationFiles, file.Name())
	}
	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, file := range migrationFiles {
		version := strings.TrimSuffix(file, ".sql")
		if applied[version] {
			continue // Already applied
		}

		// Read migration file
		content, err := ioutil.ReadFile(filepath.Join(m.mPath, file))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file, err)
		}

		// Apply migration
		if _, err := m.db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %v", file, err)
		}

		// Record migration
		if _, err := m.db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %v", file, err)
		}

		fmt.Printf("Applied migration: %s\n", file)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	// Get the latest applied migration
	var version string
	err := m.db.QueryRow("SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get latest migration: %v", err)
	}

	// Check if there's a down migration file
	downFile := filepath.Join(m.mPath, version+".down.sql")
	if _, err := os.Stat(downFile); os.IsNotExist(err) {
		return fmt.Errorf("no down migration file found for version %s", version)
	}

	// Read down migration file
	content, err := ioutil.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("failed to read down migration file %s: %v", downFile, err)
	}

	// Apply down migration
	if _, err := m.db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to apply down migration %s: %v", version, err)
	}

	// Remove migration record
	if _, err := m.db.Exec("DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		return fmt.Errorf("failed to remove migration record %s: %v", version, err)
	}

	fmt.Printf("Rolled back migration: %s\n", version)
	return nil
}

// migrateUp applies all pending migrations
func migrateUp(db *sql.DB, mPath string) error {
	m := NewMigrator(db, mPath)
	return m.Up()
}

// migrateDown rolls back the last migration
func migrateDown(db *sql.DB, mPath string) error {
	m := NewMigrator(db, mPath)
	return m.Down()
}

// createMigration creates a new migration file
func createMigration(mPath, name string) error {
	// Ensure migrations directory exists
	if err := os.MkdirAll(mPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %v", err)
	}

	// Generate timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	
	// Create up migration file
	upFile := filepath.Join(mPath, fmt.Sprintf("%s__%s.up.sql", timestamp, name))
	upContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n-- Add your UP migration SQL here\n", name, time.Now().Format("2006-01-02 15:04:05"))
	if err := ioutil.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %v", err)
	}

	// Create down migration file
	downFile := filepath.Join(mPath, fmt.Sprintf("%s__%s.down.sql", timestamp, name))
	downContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n-- Add your DOWN migration SQL here\n", name, time.Now().Format("2006-01-02 15:04:05"))
	if err := ioutil.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %v", err)
	}

	fmt.Printf("Created migration files:\n%s\n%s\n", upFile, downFile)
	return nil
}