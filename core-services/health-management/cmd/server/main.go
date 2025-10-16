package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/taishanglaojun/health-management/internal/application"
	"github.com/taishanglaojun/health-management/internal/domain"
	"github.com/taishanglaojun/health-management/internal/infrastructure/repository"
	httpHandler "github.com/taishanglaojun/health-management/internal/interfaces/http"
)

// Config 
type Config struct {
	Port     string
	Database DatabaseConfig
}

// DatabaseConfig ?
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// EventPublisher ?
type SimpleEventPublisher struct{}

func (p *SimpleEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	// 
	log.Printf("Event published: %s - %s", event.GetEventType(), event.GetEventID())
	return nil
}

func main() {
	// 
	config := loadConfig()
	
	// 
	db, err := initDatabase(config.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// 
	if err := migrateDatabase(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	// ?
	healthDataRepo := repository.NewPostgreSQLHealthDataRepository(db)
	healthProfileRepo := repository.NewPostgreSQLHealthProfileRepository(db)
	
	// 
	eventPublisher := &SimpleEventPublisher{}
	
	// ?
	healthDataService := application.NewHealthDataService(healthDataRepo, eventPublisher)
	healthProfileService := application.NewHealthProfileService(healthProfileRepo, eventPublisher)
	
	// HTTP
	router := httpHandler.NewRouter(healthDataService, healthProfileService)
	
	// Gin
	engine := gin.New()
	
	// ?
	httpHandler.SetupMiddlewares(engine)
	
	// ?
	httpHandler.SetupHealthCheck(engine)
	
	// API
	router.SetupRoutes(engine)
	
	// HTTP?
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: engine,
	}
	
	// ?
	go func() {
		log.Printf("Health Management Service starting on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// 
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// ?
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}

// loadConfig 
func loadConfig() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "health_management"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

// getEnv ?
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initDatabase 
func initDatabase(config DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// ?
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	return db, nil
}

// migrateDatabase ?
func migrateDatabase(db *gorm.DB) error {
	log.Println("Starting database migration...")
	
	// ?
	err := db.AutoMigrate(
		&domain.HealthData{},
		&domain.HealthProfile{},
		&domain.HealthReport{},
		&domain.HealthAlert{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	log.Println("Database migration completed successfully")
	return nil
}

// 麯?
func healthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// 
func initSampleData(db *gorm.DB) error {
	// ?
	var count int64
	db.Model(&domain.HealthProfile{}).Count(&count)
	if count > 0 {
		log.Println("Sample data already exists, skipping initialization")
		return nil
	}
	
	log.Println("Initializing sample data...")
	
	// ?
	// 紴
	
	log.Println("Sample data initialization completed")
	return nil
}

// ?
func performanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
		// ?
		if duration > 1*time.Second {
			log.Printf("Slow request: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// ?
func errorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ?
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// IP
		// 
		c.Next()
	}
}

// ?
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// JWT?
		// 
		c.Next()
	}
}

// 
func setupLogging() {
	// ?
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// 
	// logFile, err := os.OpenFile("health-management.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	//     log.Fatalln("Failed to open log file:", err)
	// }
	// log.SetOutput(logFile)
}

// 
func gracefulShutdown(server *http.Server, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// HTTP?
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	// ?
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
	
	log.Println("Server exited")
}

