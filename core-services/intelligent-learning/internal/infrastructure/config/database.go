package config

import (
	"fmt"
	"time"
)

// DatabaseConfig 数据库配�?
type DatabaseConfig struct {
	Host            string        `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port            int           `yaml:"port" env:"DB_PORT" default:"5432"`
	Username        string        `yaml:"username" env:"DB_USERNAME" default:"postgres"`
	Password        string        `yaml:"password" env:"DB_PASSWORD" default:""`
	Database        string        `yaml:"database" env:"DB_DATABASE" default:"intelligent_learning"`
	SSLMode         string        `yaml:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env:"DB_CONN_MAX_IDLE_TIME" default:"1m"`
}

// DSN 生成数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `yaml:"host" env:"REDIS_HOST" default:"localhost"`
	Port         int           `yaml:"port" env:"REDIS_PORT" default:"6379"`
	Password     string        `yaml:"password" env:"REDIS_PASSWORD" default:""`
	Database     int           `yaml:"database" env:"REDIS_DATABASE" default:"0"`
	PoolSize     int           `yaml:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int           `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" default:"5"`
	DialTimeout  time.Duration `yaml:"dial_timeout" env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"REDIS_WRITE_TIMEOUT" default:"3s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"REDIS_IDLE_TIMEOUT" default:"5m"`
}

// Address 生成Redis地址
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	URLs     []string `yaml:"urls" env:"ES_URLS" default:"http://localhost:9200"`
	Username string   `yaml:"username" env:"ES_USERNAME" default:""`
	Password string   `yaml:"password" env:"ES_PASSWORD" default:""`
	Index    string   `yaml:"index" env:"ES_INDEX" default:"intelligent_learning"`
}

// Neo4jConfig Neo4j配置
type Neo4jConfig struct {
	URI      string `yaml:"uri" env:"NEO4J_URI" default:"bolt://localhost:7687"`
	Username string `yaml:"username" env:"NEO4J_USERNAME" default:"neo4j"`
	Password string `yaml:"password" env:"NEO4J_PASSWORD" default:"password"`
	Database string `yaml:"database" env:"NEO4J_DATABASE" default:"neo4j"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Neo4j         Neo4jConfig         `yaml:"neo4j"`
}
