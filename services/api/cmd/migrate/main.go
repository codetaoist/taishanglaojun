package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql

	"github.com/codetaoist/taishanglaojun/api/internal/config"
	"github.com/codetaoist/taishanglaojun/api/internal/db/migrate"
)

func main() {
	var (
		action   = flag.String("action", "", "Migration action: up, down, create")
		name     = flag.String("name", "", "Migration name (required for create)")
		dbURL    = flag.String("db", "", "Database URL (overrides config)")
		mPath    = flag.String("path", "./migrations", "Migrations directory path")
	)
	flag.Parse()

	if *action == "" {
		fmt.Println("Usage: migrate -action=<up|down|create> [-name=<migration_name>] [-db=<database_url>] [-path=<migrations_path>]")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override database URL if provided
	databaseURL := *dbURL
	if databaseURL == "" {
		databaseURL = cfg.DatabaseURL
	}

	if databaseURL == "" {
		log.Fatal("Database URL is required")
	}

	// Connect to database
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create migrator
	m := migrate.NewMigrator(db, *mPath)

	// Execute action
	switch *action {
	case "up":
		if err := m.Up(); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create action")
		}
		if err := migrate.CreateMigration(*mPath, *name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}