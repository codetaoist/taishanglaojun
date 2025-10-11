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

// Config 应用配置
type Config struct {
	Port     string
	Database DatabaseConfig
}

// DatabaseConfig 数据库配�?
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// EventPublisher 简单的事件发布器实�?
type SimpleEventPublisher struct{}

func (p *SimpleEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	// 这里简化处理，实际应该发送到消息队列
	log.Printf("Event published: %s - %s", event.GetEventType(), event.GetEventID())
	return nil
}

func main() {
	// 加载配置
	config := loadConfig()
	
	// 初始化数据库
	db, err := initDatabase(config.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// 自动迁移数据库表
	if err := migrateDatabase(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	// 初始化仓�?
	healthDataRepo := repository.NewPostgreSQLHealthDataRepository(db)
	healthProfileRepo := repository.NewPostgreSQLHealthProfileRepository(db)
	
	// 初始化事件发布器
	eventPublisher := &SimpleEventPublisher{}
	
	// 初始化应用服�?
	healthDataService := application.NewHealthDataService(healthDataRepo, eventPublisher)
	healthProfileService := application.NewHealthProfileService(healthProfileRepo, eventPublisher)
	
	// 初始化HTTP路由
	router := httpHandler.NewRouter(healthDataService, healthProfileService)
	
	// 创建Gin引擎
	engine := gin.New()
	
	// 设置中间�?
	httpHandler.SetupMiddlewares(engine)
	
	// 设置健康检�?
	httpHandler.SetupHealthCheck(engine)
	
	// 设置API路由
	router.SetupRoutes(engine)
	
	// 创建HTTP服务�?
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: engine,
	}
	
	// 启动服务�?
	go func() {
		log.Printf("Health Management Service starting on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// 优雅关闭服务�?
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}

// loadConfig 加载配置
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

// getEnv 获取环境变量，如果不存在则返回默认�?
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initDatabase 初始化数据库连接
func initDatabase(config DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 配置连接�?
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	return db, nil
}

// migrateDatabase 数据库迁�?
func migrateDatabase(db *gorm.DB) error {
	log.Println("Starting database migration...")
	
	// 自动迁移所有模�?
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

// 健康检查函�?
func healthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// 初始化示例数据（可选）
func initSampleData(db *gorm.DB) error {
	// 检查是否已有数�?
	var count int64
	db.Model(&domain.HealthProfile{}).Count(&count)
	if count > 0 {
		log.Println("Sample data already exists, skipping initialization")
		return nil
	}
	
	log.Println("Initializing sample data...")
	
	// 这里可以添加一些示例数�?
	// 例如创建测试用户的健康档案等
	
	log.Println("Sample data initialization completed")
	return nil
}

// 性能监控中间�?
func performanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
		// 记录慢请�?
		if duration > 1*time.Second {
			log.Printf("Slow request: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// 错误处理中间�?
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

// 请求限流中间件（简化版�?
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现基于IP或用户的请求限流
		// 简化处理，直接通过
		c.Next()
	}
}

// 认证中间件（简化版�?
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现JWT认证或其他认证方�?
		// 简化处理，直接通过
		c.Next()
	}
}

// 日志配置
func setupLogging() {
	// 配置日志格式和输�?
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// 可以配置日志文件输出
	// logFile, err := os.OpenFile("health-management.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	//     log.Fatalln("Failed to open log file:", err)
	// }
	// log.SetOutput(logFile)
}

// 优雅关闭处理
func gracefulShutdown(server *http.Server, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// 关闭HTTP服务�?
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	// 关闭数据库连�?
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
	
	log.Println("Server exited")
}
