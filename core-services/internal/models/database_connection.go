package models

import (
	"time"
	"gorm.io/gorm"
)

// DatabaseType 数据库类型枚举
type DatabaseType string

const (
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeOracle     DatabaseType = "oracle"
	DatabaseTypeSQLServer  DatabaseType = "sqlserver"
	DatabaseTypeMariaDB    DatabaseType = "mariadb"
)

// ConnectionStatus 连接状态枚举
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusError        ConnectionStatus = "error"
	ConnectionStatusUnknown      ConnectionStatus = "unknown"
)

// DatabaseConnection 数据库连接配置模型
type DatabaseConnection struct {
	ID                string           `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name              string           `json:"name" gorm:"type:varchar(255);not null;index"`
	Type              DatabaseType     `json:"type" gorm:"type:varchar(50);not null;index"`
	Host              string           `json:"host" gorm:"type:varchar(255);not null"`
	Port              int              `json:"port" gorm:"not null"`
	Database          string           `json:"database" gorm:"type:varchar(255);not null"`
	Username          string           `json:"username" gorm:"type:varchar(255);not null"`
	Password          string           `json:"password" gorm:"type:text;not null"` // 加密存储
	SSL               bool             `json:"ssl" gorm:"default:false"`
	ConnectionTimeout int              `json:"connection_timeout" gorm:"default:30"`
	MaxConnections    int              `json:"max_connections" gorm:"default:10"`
	Description       string           `json:"description" gorm:"type:text"`
	Tags              string           `json:"tags" gorm:"type:text"` // JSON数组字符串
	IsDefault         bool             `json:"is_default" gorm:"default:false;index"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	CreatedBy         string           `json:"created_by" gorm:"type:varchar(36);not null;index"`
	LastConnectedAt   *time.Time       `json:"last_connected_at"`
	
	// 关联的连接状态
	Status *DatabaseConnectionStatus `json:"status,omitempty" gorm:"foreignKey:ConnectionID"`
}

// DatabaseConnectionStatus 数据库连接状态模型
type DatabaseConnectionStatus struct {
	ID                string           `json:"id" gorm:"type:varchar(36);primaryKey"`
	ConnectionID      string           `json:"connection_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	Status            ConnectionStatus `json:"status" gorm:"type:varchar(50);not null;index"`
	LastChecked       time.Time        `json:"last_checked"`
	ResponseTime      *int             `json:"response_time"` // 毫秒
	ErrorMessage      string           `json:"error_message" gorm:"type:text"`
	ServerVersion     string           `json:"server_version" gorm:"type:varchar(255)"`
	DatabaseSize      string           `json:"database_size" gorm:"type:varchar(100)"`
	ActiveConnections *int             `json:"active_connections"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	
	// 关联的数据库连接
	Connection DatabaseConnection `json:"connection,omitempty" gorm:"foreignKey:ConnectionID"`
}

// DatabaseConnectionEvent 数据库连接事件日志
type DatabaseConnectionEvent struct {
	ID           string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	ConnectionID string    `json:"connection_id" gorm:"type:varchar(36);not null;index"`
	Action       string    `json:"action" gorm:"type:varchar(50);not null;index"`
	Success      bool      `json:"success" gorm:"not null;index"`
	Message      string    `json:"message" gorm:"type:text"`
	Details      string    `json:"details" gorm:"type:text"` // JSON字符串
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	IPAddress    string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent    string    `json:"user_agent" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	
	// 关联的数据库连接
	Connection DatabaseConnection `json:"connection,omitempty" gorm:"foreignKey:ConnectionID"`
}

// BeforeCreate 创建前钩子
func (dc *DatabaseConnection) BeforeCreate(tx *gorm.DB) error {
	if dc.ID == "" {
		dc.ID = generateUUID()
	}
	return nil
}

// BeforeCreate 创建前钩子
func (dcs *DatabaseConnectionStatus) BeforeCreate(tx *gorm.DB) error {
	if dcs.ID == "" {
		dcs.ID = generateUUID()
	}
	return nil
}

// BeforeCreate 创建前钩子
func (dce *DatabaseConnectionEvent) BeforeCreate(tx *gorm.DB) error {
	if dce.ID == "" {
		dce.ID = generateUUID()
	}
	return nil
}

// DatabaseConnectionForm 数据库连接表单结构
type DatabaseConnectionForm struct {
	Name              string       `json:"name" binding:"required,min=1,max=255"`
	Type              DatabaseType `json:"type" binding:"required"`
	Host              string       `json:"host" binding:"required,min=1,max=255"`
	Port              int          `json:"port" binding:"required,min=1,max=65535"`
	Database          string       `json:"database" binding:"required,min=1,max=255"`
	Username          string       `json:"username" binding:"required,min=1,max=255"`
	Password          string       `json:"password" binding:"required,min=1"`
	SSL               *bool        `json:"ssl"`
	ConnectionTimeout *int         `json:"connection_timeout" binding:"omitempty,min=1,max=300"`
	MaxConnections    *int         `json:"max_connections" binding:"omitempty,min=1,max=1000"`
	Description       string       `json:"description" binding:"max=1000"`
	Tags              []string     `json:"tags"`
	IsDefault         *bool        `json:"is_default"`
}

// DatabaseConnectionQuery 数据库连接查询参数
type DatabaseConnectionQuery struct {
	Page      int              `form:"page" binding:"omitempty,min=1"`
	PageSize  int              `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Search    string           `form:"search"`
	Type      DatabaseType     `form:"type"`
	Status    ConnectionStatus `form:"status"`
	Tags      string           `form:"tags"`
	SortBy    string           `form:"sortBy" binding:"omitempty,oneof=name type created_at last_connected_at"`
	SortOrder string           `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
}

// DatabaseConnectionTest 数据库连接测试结果
type DatabaseConnectionTest struct {
	Success      bool                   `json:"success"`
	ResponseTime int                    `json:"response_time"` // 毫秒
	ErrorMessage string                 `json:"error_message,omitempty"`
	ServerInfo   *DatabaseServerInfo    `json:"server_info,omitempty"`
}

// DatabaseServerInfo 数据库服务器信息
type DatabaseServerInfo struct {
	Version  string `json:"version"`
	Charset  string `json:"charset,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// DatabaseConnectionStats 数据库连接统计信息
type DatabaseConnectionStats struct {
	TotalConnections      int                              `json:"total_connections"`
	ActiveConnections     int                              `json:"active_connections"`
	ConnectionsByType     map[DatabaseType]int             `json:"connections_by_type"`
	ConnectionsByStatus   map[ConnectionStatus]int         `json:"connections_by_status"`
	AverageResponseTime   float64                          `json:"average_response_time"`
	LastUpdated           time.Time                        `json:"last_updated"`
}

// DatabaseTypeConfig 数据库类型配置
type DatabaseTypeConfig struct {
	Type                     DatabaseType `json:"type"`
	Name                     string       `json:"name"`
	Icon                     string       `json:"icon"`
	DefaultPort              int          `json:"default_port"`
	SupportsSsl              bool         `json:"supports_ssl"`
	SupportsConnectionPool   bool         `json:"supports_connection_pool"`
	ConnectionStringTemplate string       `json:"connection_string_template"`
	Description              string       `json:"description"`
	DocumentationURL         string       `json:"documentation_url,omitempty"`
}

// GetDatabaseTypeConfigs 获取所有数据库类型配置
func GetDatabaseTypeConfigs() map[DatabaseType]DatabaseTypeConfig {
	return map[DatabaseType]DatabaseTypeConfig{
		DatabaseTypeMySQL: {
			Type:                     DatabaseTypeMySQL,
			Name:                     "MySQL",
			Icon:                     "mysql",
			DefaultPort:              3306,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "mysql://{username}:{password}@{host}:{port}/{database}",
			Description:              "MySQL 关系型数据库",
			DocumentationURL:         "https://dev.mysql.com/doc/",
		},
		DatabaseTypePostgreSQL: {
			Type:                     DatabaseTypePostgreSQL,
			Name:                     "PostgreSQL",
			Icon:                     "postgresql",
			DefaultPort:              5432,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "postgresql://{username}:{password}@{host}:{port}/{database}",
			Description:              "PostgreSQL 关系型数据库",
			DocumentationURL:         "https://www.postgresql.org/docs/",
		},
		DatabaseTypeSQLite: {
			Type:                     DatabaseTypeSQLite,
			Name:                     "SQLite",
			Icon:                     "sqlite",
			DefaultPort:              0,
			SupportsSsl:              false,
			SupportsConnectionPool:   false,
			ConnectionStringTemplate: "sqlite://{database}",
			Description:              "SQLite 轻量级数据库",
			DocumentationURL:         "https://www.sqlite.org/docs.html",
		},
		DatabaseTypeMongoDB: {
			Type:                     DatabaseTypeMongoDB,
			Name:                     "MongoDB",
			Icon:                     "mongodb",
			DefaultPort:              27017,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "mongodb://{username}:{password}@{host}:{port}/{database}",
			Description:              "MongoDB 文档数据库",
			DocumentationURL:         "https://docs.mongodb.com/",
		},
		DatabaseTypeRedis: {
			Type:                     DatabaseTypeRedis,
			Name:                     "Redis",
			Icon:                     "redis",
			DefaultPort:              6379,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "redis://{username}:{password}@{host}:{port}/{database}",
			Description:              "Redis 内存数据库",
			DocumentationURL:         "https://redis.io/documentation",
		},
		DatabaseTypeOracle: {
			Type:                     DatabaseTypeOracle,
			Name:                     "Oracle",
			Icon:                     "oracle",
			DefaultPort:              1521,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "oracle://{username}:{password}@{host}:{port}/{database}",
			Description:              "Oracle 企业级数据库",
			DocumentationURL:         "https://docs.oracle.com/database/",
		},
		DatabaseTypeSQLServer: {
			Type:                     DatabaseTypeSQLServer,
			Name:                     "SQL Server",
			Icon:                     "sqlserver",
			DefaultPort:              1433,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "sqlserver://{username}:{password}@{host}:{port}/{database}",
			Description:              "Microsoft SQL Server",
			DocumentationURL:         "https://docs.microsoft.com/sql/",
		},
		DatabaseTypeMariaDB: {
			Type:                     DatabaseTypeMariaDB,
			Name:                     "MariaDB",
			Icon:                     "mariadb",
			DefaultPort:              3306,
			SupportsSsl:              true,
			SupportsConnectionPool:   true,
			ConnectionStringTemplate: "mariadb://{username}:{password}@{host}:{port}/{database}",
			Description:              "MariaDB 关系型数据库",
			DocumentationURL:         "https://mariadb.org/documentation/",
		},
	}
}

// generateUUID 生成UUID的辅助函数（需要在其他地方实现）
func generateUUID() string {
	// 这里应该使用实际的UUID生成库，比如 github.com/google/uuid
	// 暂时返回一个占位符
	return "placeholder-uuid"
}