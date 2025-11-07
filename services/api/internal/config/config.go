package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/pgvector/pgvector"
)

type Config struct {
    Env              string
    LogLevel         string
    TraceHeader      string
    Port             string
    AllowedOrigins   []string
    JWTSecret        string
    JWTPrivateKeyPEM string
    JWTPublicKeyPEM  string
    DevSkipSignature bool
    AutoMigrate      bool
    AuthServiceURL   string

    DatabaseURL      string // Postgres DSN, optional in dev
    DB               *sql.DB // Database connection
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getbool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func splitCSV(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Close closes the database connection
func (c *Config) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// Load reads environment variables and returns runtime configuration.
func Load() Config {
	cfg := Config{
		Env:            getenv("ENV", "dev"),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		TraceHeader:    getenv("TRACE_HEADER", "X-Trace-Id"),
        Port:           getenv("LAOJUN_API_PORT", "8081"),
        AllowedOrigins: splitCSV("ALLOWED_ORIGINS"),
        JWTSecret:      getenv("JWT_SECRET", ""),
        DevSkipSignature: getbool("DEV_SKIP_SIGNATURE", true),
        AutoMigrate:    getbool("AUTO_MIGRATE", true),
        DatabaseURL:    getenv("DATABASE_URL", ""),
        AuthServiceURL: getenv("AUTH_SERVICE_URL", "http://localhost:8081"),
    }
	// Optional key paths (RS256) — read file contents if provided
	if pk := getenv("JWT_PRIVATE_KEY_PATH", ""); pk != "" {
		if b, err := os.ReadFile(pk); err == nil {
			cfg.JWTPrivateKeyPEM = string(b)
		}
	}
	if pub := getenv("JWT_PUBLIC_KEY_PATH", ""); pub != "" {
		if b, err := os.ReadFile(pub); err == nil {
			cfg.JWTPublicKeyPEM = string(b)
		}
	}

	// Initialize database connection
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost/codetaoist?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}

	cfg.DB = db

	return cfg
}