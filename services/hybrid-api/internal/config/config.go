package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
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
    
    // Vector database configuration
    VectorDBType     string // Type of vector database (milvus, weaviate, etc.)
    VectorDBHost     string // Host of vector database
    VectorDBPort     string // Port of vector database
    VectorDBUser     string // Username for vector database
    VectorDBPassword string // Password for vector database
    VectorDBDatabase string // Database name for vector database
    
    // AI service configuration
    AIService        AIServiceConfig
    
    // Plugin system configuration
    Plugin           PluginConfig
}

type AIServiceConfig struct {
    VectorAddr string // Address of the vector service
    ModelAddr  string // Address of the model service
}

type PluginConfig struct {
    SandboxDir      string // Directory for plugin sandboxes
    PluginDir       string // Directory for plugin files
    MaxPlugins      int    // Maximum number of plugins
    DefaultTimeout  int    // Default timeout for plugin operations (seconds)
    EnableNetwork   bool   // Whether plugins can access network
    ResourceLimits  ResourceLimits
}

type ResourceLimits struct {
    CPUQuota     int64 // CPU quota (percentage)
    MemoryLimit  int64 // Memory limit (bytes)
    DiskLimit    int64 // Disk limit (bytes)
    ProcessLimit int   // Process limit
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

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return i
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
        Port:           getenv("LAOJUN_API_PORT", "8082"),
        AllowedOrigins: splitCSV("ALLOWED_ORIGINS"),
        JWTSecret:      getenv("JWT_SECRET", ""),
        DevSkipSignature: getbool("DEV_SKIP_SIGNATURE", true),
        AutoMigrate:    getbool("AUTO_MIGRATE", true),
        DatabaseURL:    getenv("DATABASE_URL", ""),
        AuthServiceURL: getenv("AUTH_SERVICE_URL", "http://localhost:8081"),
        
        // Vector database configuration
        VectorDBType:     getenv("VECTOR_DB_TYPE", "milvus"),
        VectorDBHost:     getenv("VECTOR_DB_HOST", "localhost"),
        VectorDBPort:     getenv("VECTOR_DB_PORT", "19530"),
        VectorDBUser:     getenv("VECTOR_DB_USER", ""),
        VectorDBPassword: getenv("VECTOR_DB_PASSWORD", ""),
        VectorDBDatabase: getenv("VECTOR_DB_DATABASE", ""),
        
        // AI service configuration
        AIService: AIServiceConfig{
            VectorAddr: getenv("AI_VECTOR_SERVICE_ADDR", "localhost:50051"),
            ModelAddr:  getenv("AI_MODEL_SERVICE_ADDR", "localhost:50052"),
        },
        
        // Plugin system configuration
        Plugin: PluginConfig{
            SandboxDir:     getenv("PLUGIN_SANDBOX_DIR", "/tmp/taishanglaojun/sandboxes"),
            PluginDir:      getenv("PLUGIN_DIR", "/tmp/taishanglaojun/plugins"),
            MaxPlugins:     getInt("PLUGIN_MAX_PLUGINS", 10),
            DefaultTimeout: getInt("PLUGIN_DEFAULT_TIMEOUT", 30),
            EnableNetwork:  getbool("PLUGIN_ENABLE_NETWORK", false),
            ResourceLimits: ResourceLimits{
                CPUQuota:     getInt64("PLUGIN_CPU_QUOTA", 50),
                MemoryLimit:  getInt64("PLUGIN_MEMORY_LIMIT", 512*1024*1024), // 512MB
                DiskLimit:    getInt64("PLUGIN_DISK_LIMIT", 100*1024*1024),   // 100MB
                ProcessLimit: getInt("PLUGIN_PROCESS_LIMIT", 10),
            },
        },
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
		dbURL = "postgres://postgres:password@localhost/taishanglaojun?sslmode=disable"
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