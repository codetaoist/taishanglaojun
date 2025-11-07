package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
)

// Migrator handles database migrations
type Migrator struct {
	db        *sql.DB
	migrationsPath string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:        db,
		migrationsPath: migrationsPath,
	}
}

// CreateMigrationsTable creates the migrations table if it doesn't exist
func (m *Migrator) CreateMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);
	`
	_, err := m.db.Exec(query)
	return err
}

// GetAppliedMigrations returns a list of applied migration versions
func (m *Migrator) GetAppliedMigrations() ([]string, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// GetPendingMigrations returns a list of pending migration files
func (m *Migrator) GetPendingMigrations(appliedVersions []string) ([]string, error) {
	files, err := ioutil.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	var pendingMigrations []string
	appliedMap := make(map[string]bool)
	for _, version := range appliedVersions {
		appliedMap[version] = true
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(filename, ".up.sql") {
			continue
		}

		version := strings.TrimSuffix(filename, ".up.sql")
		if !appliedMap[version] {
			pendingMigrations = append(pendingMigrations, version)
		}
	}

	// Sort migrations by version
	sort.Strings(pendingMigrations)

	return pendingMigrations, nil
}

// ApplyMigration applies a single migration
func (m *Migrator) ApplyMigration(version string) error {
	// Read migration file
	upFile := filepath.Join(m.migrationsPath, version+".up.sql")
	downFile := filepath.Join(m.migrationsPath, version+".down.sql")

	// Read up migration
	upSQL, err := ioutil.ReadFile(upFile)
	if err != nil {
		return fmt.Errorf("failed to read up migration file %s: %w", upFile, err)
	}

	// Begin transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute up migration
	if _, err := tx.Exec(string(upSQL)); err != nil {
		return fmt.Errorf("failed to execute up migration %s: %w", version, err)
	}

	// Record migration
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		version,
	); err != nil {
		return fmt.Errorf("failed to record migration %s: %w", version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %s: %w", version, err)
	}

	return nil
}

// RollbackMigration rolls back a single migration
func (m *Migrator) RollbackMigration(version string) error {
	// Read down migration file
	downFile := filepath.Join(m.migrationsPath, version+".down.sql")

	downSQL, err := ioutil.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("failed to read down migration file %s: %w", downFile, err)
	}

	// Begin transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute down migration
	if _, err := tx.Exec(string(downSQL)); err != nil {
		return fmt.Errorf("failed to execute down migration %s: %w", version, err)
	}

	// Remove migration record
	if _, err := tx.Exec(
		"DELETE FROM schema_migrations WHERE version = $1",
		version,
	); err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", version, err)
	}

	return nil
}

// Up applies all pending migrations
func (m *Migrator) Up() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	appliedVersions, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	pendingMigrations, err := m.GetPendingMigrations(appliedVersions)
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pendingMigrations) == 0 {
		fmt.Println("No pending migrations to apply")
		return nil
	}

	fmt.Printf("Applying %d migrations...\n", len(pendingMigrations))
	for _, version := range pendingMigrations {
		if err := m.ApplyMigration(version); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}
		fmt.Printf("Applied migration: %s\n", version)
	}

	fmt.Println("All migrations applied successfully")
	return nil
}

// Down rolls back the last applied migration
func (m *Migrator) Down() error {
	appliedVersions, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedVersions) == 0 {
		fmt.Println("No migrations to rollback")
		return nil
	}

	// Get the last applied migration
	lastVersion := appliedVersions[len(appliedVersions)-1]

	if err := m.RollbackMigration(lastVersion); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastVersion, err)
	}

	fmt.Printf("Rolled back migration: %s\n", lastVersion)
	return nil
}

// CreateMigration creates a new migration file pair
func CreateMigration(migrationsPath, name string) error {
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	version := fmt.Sprintf("%s_%s", timestamp, name)

	upFile := filepath.Join(migrationsPath, version+".up.sql")
	downFile := filepath.Join(migrationsPath, version+".down.sql")

	// Create up migration file
	upContent := fmt.Sprintf("-- Migration: %s\n-- Up\n\n", version)
	if err := ioutil.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downContent := fmt.Sprintf("-- Migration: %s\n-- Down\n\n", version)
	if err := ioutil.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n%s\n%s\n", upFile, downFile)
	return nil
}