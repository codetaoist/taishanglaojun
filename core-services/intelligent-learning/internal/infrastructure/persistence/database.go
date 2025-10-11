package persistence

import (
	"context"
	"fmt"
	"time"

	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	_ "github.com/lib/pq"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/infrastructure/config"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	PostgreSQL    *sql.DB
	Redis         *redis.Client
	Elasticsearch *elasticsearch.Client
	Neo4j         neo4j.DriverWithContext
	config        *config.StorageConfig
}

// NewDatabaseManager 创建新的数据库管理器
func NewDatabaseManager(cfg *config.StorageConfig) (*DatabaseManager, error) {
	dm := &DatabaseManager{
		config: cfg,
	}

	// 初始化PostgreSQL
	if err := dm.initPostgreSQL(); err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// 初始化Redis
	if err := dm.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// 初始化Elasticsearch
	if err := dm.initElasticsearch(); err != nil {
		return nil, fmt.Errorf("failed to initialize Elasticsearch: %w", err)
	}

	// 初始化Neo4j
	if err := dm.initNeo4j(); err != nil {
		return nil, fmt.Errorf("failed to initialize Neo4j: %w", err)
	}

	return dm, nil
}

// initPostgreSQL 初始化PostgreSQL连接
func (dm *DatabaseManager) initPostgreSQL() error {
	db, err := sql.Open("postgres", dm.config.Database.DSN())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接�?
	db.SetMaxOpenConns(dm.config.Database.MaxOpenConns)
	db.SetMaxIdleConns(dm.config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(dm.config.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(dm.config.Database.ConnMaxIdleTime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	dm.PostgreSQL = db
	return nil
}

// initRedis 初始化Redis连接
func (dm *DatabaseManager) initRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         dm.config.Redis.Address(),
		Password:     dm.config.Redis.Password,
		DB:           dm.config.Redis.Database,
		PoolSize:     dm.config.Redis.PoolSize,
		MinIdleConns: dm.config.Redis.MinIdleConns,
		DialTimeout:  dm.config.Redis.DialTimeout,
		ReadTimeout:  dm.config.Redis.ReadTimeout,
		WriteTimeout: dm.config.Redis.WriteTimeout,
		IdleTimeout:  dm.config.Redis.IdleTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	dm.Redis = rdb
	return nil
}

// initElasticsearch 初始化Elasticsearch连接
func (dm *DatabaseManager) initElasticsearch() error {
	cfg := elasticsearch.Config{
		Addresses: dm.config.Elasticsearch.URLs,
		Username:  dm.config.Elasticsearch.Username,
		Password:  dm.config.Elasticsearch.Password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := es.Info(es.Info.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	dm.Elasticsearch = es
	return nil
}

// initNeo4j 初始化Neo4j连接
func (dm *DatabaseManager) initNeo4j() error {
	driver, err := neo4j.NewDriverWithContext(
		dm.config.Neo4j.URI,
		neo4j.BasicAuth(dm.config.Neo4j.Username, dm.config.Neo4j.Password, ""),
	)
	if err != nil {
		return fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	dm.Neo4j = driver
	return nil
}

// Close 关闭所有数据库连接
func (dm *DatabaseManager) Close() error {
	var errors []error

	// 关闭PostgreSQL
	if dm.PostgreSQL != nil {
		if err := dm.PostgreSQL.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close PostgreSQL: %w", err))
		}
	}

	// 关闭Redis
	if dm.Redis != nil {
		if err := dm.Redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	// 关闭Neo4j
	if dm.Neo4j != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := dm.Neo4j.Close(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Neo4j: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// Health 检查所有数据库连接健康状�?
func (dm *DatabaseManager) Health(ctx context.Context) map[string]error {
	health := make(map[string]error)

	// 检查PostgreSQL
	if dm.PostgreSQL != nil {
		health["postgresql"] = dm.PostgreSQL.PingContext(ctx)
	}

	// 检查Redis
	if dm.Redis != nil {
		health["redis"] = dm.Redis.Ping(ctx).Err()
	}

	// 检查Elasticsearch
	if dm.Elasticsearch != nil {
		res, err := dm.Elasticsearch.Info(dm.Elasticsearch.Info.WithContext(ctx))
		if err != nil {
			health["elasticsearch"] = err
		} else {
			res.Body.Close()
			if res.IsError() {
				health["elasticsearch"] = fmt.Errorf("elasticsearch error: %s", res.Status())
			} else {
				health["elasticsearch"] = nil
			}
		}
	}

	// 检查Neo4j
	if dm.Neo4j != nil {
		health["neo4j"] = dm.Neo4j.VerifyConnectivity(ctx)
	}

	return health
}

// GetPostgreSQL 获取PostgreSQL连接
func (dm *DatabaseManager) GetPostgreSQL() *sql.DB {
	return dm.PostgreSQL
}

// GetRedis 获取Redis客户�?
func (dm *DatabaseManager) GetRedis() *redis.Client {
	return dm.Redis
}

// GetElasticsearch 获取Elasticsearch客户�?
func (dm *DatabaseManager) GetElasticsearch() *elasticsearch.Client {
	return dm.Elasticsearch
}

// GetNeo4j 获取Neo4j驱动
func (dm *DatabaseManager) GetNeo4j() neo4j.DriverWithContext {
	return dm.Neo4j
}

// Transaction 执行PostgreSQL事务
func (dm *DatabaseManager) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := dm.PostgreSQL.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Neo4jSession 创建Neo4j会话并执行函�?
func (dm *DatabaseManager) Neo4jSession(ctx context.Context, fn func(neo4j.SessionWithContext) error) error {
	session := dm.Neo4j.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: dm.config.Neo4j.Database,
	})
	defer session.Close(ctx)

	return fn(session)
}

// Neo4jReadTransaction 执行Neo4j读事�?
func (dm *DatabaseManager) Neo4jReadTransaction(ctx context.Context, fn func(neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := dm.Neo4j.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: dm.config.Neo4j.Database,
	})
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, fn)
}

// Neo4jWriteTransaction 执行Neo4j写事�?
func (dm *DatabaseManager) Neo4jWriteTransaction(ctx context.Context, fn func(neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := dm.Neo4j.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: dm.config.Neo4j.Database,
	})
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, fn)
}
