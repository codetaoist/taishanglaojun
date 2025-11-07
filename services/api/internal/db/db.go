package db

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql

    "github.com/codetaoist/taishanglaojun/api/internal/config"
    "github.com/codetaoist/taishanglaojun/api/internal/db/migrate"
)

// DB is the global connection pool initialised at startup.
var DB *sql.DB

// Init initialises the global DB pool using the provided configuration.
// It expects cfg.DatabaseURL to be in the standard PostgreSQL connection URI format, e.g.:
//   postgres://user:password@localhost:5432/laojun?sslmode=disable
func Init(cfg config.Config) error {
    if cfg.DatabaseURL == "" {
        // skip DB init in dev if no DSN provided
        return nil
    }
    db, err := sql.Open("pgx", cfg.DatabaseURL)
    if err != nil {
        return fmt.Errorf("open db: %w", err)
    }
    // reasonable defaults; can be overridden later via cfg if needed
    db.SetMaxOpenConns(16)
    db.SetMaxIdleConns(4)
    db.SetConnMaxLifetime(30 * time.Minute)

    // Validate connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        _ = db.Close()
        return fmt.Errorf("ping db: %w", err)
    }

    DB = db
    
    // Run migrations if enabled
    if cfg.AutoMigrate {
        log.Println("Running database migrations...")
        migrator := migrate.NewMigrator(DB, "./migrations")
        if err := migrator.Up(); err != nil {
            return fmt.Errorf("run migrations: %w", err)
        }
        log.Println("Database migrations completed successfully")
    }
    
    return nil
}

// Close closes the global DB pool.
func Close() error {
    if DB != nil {
        return DB.Close()
    }
    return nil
}