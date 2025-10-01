package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/taishanglaojun/database-layer/internal/database"
	"github.com/taishanglaojun/database-layer/internal/models"
	"github.com/taishanglaojun/database-layer/internal/repository"
)

func main() {
	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// 创建数据库配置
	config := &database.Config{
		Postgres: database.PostgresConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         5432,
			Username:     getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "password"),
			Database:     getEnv("DB_NAME", "taishang_test"),
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: database.RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         6379,
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
	}

	// 初始化数据库管理器
	manager, err := database.NewManager(config, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer manager.Close()

	// 测试数据库连接
	if err := testDatabaseConnections(manager, logger); err != nil {
		logger.Fatal("Database connection test failed", zap.Error(err))
	}

	// 演示基础仓储操作
	if err := demonstrateRepository(manager, logger); err != nil {
		logger.Fatal("Repository demonstration failed", zap.Error(err))
	}

	// 演示缓存操作
	if err := demonstrateCache(manager, logger); err != nil {
		logger.Fatal("Cache demonstration failed", zap.Error(err))
	}

	logger.Info("Database layer demonstration completed successfully")
}

// testDatabaseConnections 测试数据库连接
func testDatabaseConnections(manager *database.Manager, logger *zap.Logger) error {
	logger.Info("Testing database connections...")

	// 检查健康状态
	health := manager.GetHealthStatus()
	
	for dbType, status := range health {
		logger.Info("Database health check", 
			zap.String("type", dbType),
			zap.Any("status", status))
	}

	// 测试PostgreSQL连接
	if postgres := manager.GetPostgresDB(); postgres != nil {
		if err := postgres.Health(); err != nil {
			return fmt.Errorf("PostgreSQL health check failed: %w", err)
		}
		logger.Info("PostgreSQL connection successful")
	}

	// 测试Redis连接
	if redis := manager.GetRedisDB(); redis != nil {
		if err := redis.Health(); err != nil {
			return fmt.Errorf("Redis health check failed: %w", err)
		}
		logger.Info("Redis connection successful")
	}

	return nil
}

// TestUser 测试用户模型
type TestUser struct {
	models.BaseModel
	Name     string `json:"name" gorm:"size:100;not null"`
	Email    string `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

// TableName 指定表名
func (TestUser) TableName() string {
	return "test_users"
}

// demonstrateRepository 演示仓储操作
func demonstrateRepository(manager *database.Manager, logger *zap.Logger) error {
	logger.Info("Demonstrating repository operations...")

	postgres := manager.GetPostgresDB()
	if postgres == nil {
		return fmt.Errorf("PostgreSQL not available")
	}

	// 自动迁移测试表
	if err := postgres.AutoMigrate(&TestUser{}); err != nil {
		return fmt.Errorf("failed to migrate test table: %w", err)
	}

	// 创建仓储
	repo := repository.NewBaseRepository[TestUser](postgres.GetDB(), logger)
	ctx := context.Background()

	// 创建测试用户
	user := &TestUser{
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Age:      25,
		IsActive: true,
	}

	if err := repo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	logger.Info("User created", zap.Uint("id", user.ID))

	// 根据ID获取用户
	retrievedUser, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	logger.Info("User retrieved", zap.String("name", retrievedUser.Name))

	// 更新用户
	retrievedUser.Age = 26
	if err := repo.Update(ctx, retrievedUser); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	logger.Info("User updated", zap.Int("new_age", retrievedUser.Age))

	// 批量创建用户
	users := []*TestUser{
		{Name: "李四", Email: "lisi@example.com", Age: 30},
		{Name: "王五", Email: "wangwu@example.com", Age: 28},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 32},
	}

	if err := repo.BatchCreate(ctx, users); err != nil {
		return fmt.Errorf("failed to batch create users: %w", err)
	}
	logger.Info("Users batch created", zap.Int("count", len(users)))

	// 分页查询
	opts := &models.QueryOptions{
		Pagination: &models.PaginationQuery{
			Page:     1,
			PageSize: 2,
			OrderBy:  "created_at",
			Sort:     "desc",
		},
	}

	result, err := repo.Paginate(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to paginate users: %w", err)
	}
	logger.Info("Users paginated", 
		zap.Int64("total", result.Total),
		zap.Int("page_size", result.PageSize))

	// 搜索用户
	searchOpts := &models.QueryOptions{
		Search: &models.SearchQuery{
			Keyword: "张",
			Fields:  []string{"name"},
		},
	}

	searchResults, err := repo.List(ctx, searchOpts)
	if err != nil {
		return fmt.Errorf("failed to search users: %w", err)
	}
	logger.Info("Users searched", zap.Int("count", len(searchResults)))

	return nil
}

// demonstrateCache 演示缓存操作
func demonstrateCache(manager *database.Manager, logger *zap.Logger) error {
	logger.Info("Demonstrating cache operations...")

	cache := manager.GetCacheService()
	ctx := context.Background()

	// 设置缓存
	if err := cache.Set(ctx, "test:key1", "Hello, World!", 5*time.Minute); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	logger.Info("Cache set successfully")

	// 获取缓存
	value, err := cache.Get(ctx, "test:key1")
	if err != nil {
		return fmt.Errorf("failed to get cache: %w", err)
	}
	logger.Info("Cache retrieved", zap.String("value", value))

	// 检查缓存是否存在
	exists, err := cache.Exists(ctx, "test:key1")
	if err != nil {
		return fmt.Errorf("failed to check cache existence: %w", err)
	}
	logger.Info("Cache existence checked", zap.Int64("exists", exists))

	// 删除缓存
	if err := cache.Del(ctx, "test:key1"); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	logger.Info("Cache deleted successfully")

	// Redis特定操作（如果Redis可用）
	if redis := manager.GetRedisDB(); redis != nil {
		// Hash操作
		if err := redis.HSet(ctx, "test:hash", "field1", "value1"); err != nil {
			return fmt.Errorf("failed to set hash: %w", err)
		}

		hashValue, err := redis.HGet(ctx, "test:hash", "field1")
		if err != nil {
			return fmt.Errorf("failed to get hash: %w", err)
		}
		logger.Info("Hash operation successful", zap.String("value", hashValue))

		// List操作
		if err := redis.LPush(ctx, "test:list", "item1", "item2"); err != nil {
			return fmt.Errorf("failed to push to list: %w", err)
		}
		logger.Info("List operation successful")

		// Set操作
		if err := redis.SAdd(ctx, "test:set", "member1", "member2"); err != nil {
			return fmt.Errorf("failed to add to set: %w", err)
		}

		members, err := redis.SMembers(ctx, "test:set")
		if err != nil {
			return fmt.Errorf("failed to get set members: %w", err)
		}
		logger.Info("Set operation successful", zap.Strings("members", members))

		// 计数器操作
		count, err := redis.Incr(ctx, "test:counter")
		if err != nil {
			return fmt.Errorf("failed to increment counter: %w", err)
		}
		logger.Info("Counter incremented", zap.Int64("count", count))
	}

	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}