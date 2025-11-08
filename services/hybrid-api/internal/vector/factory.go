package vector

import (
	"context"
	"fmt"
)

// DatabaseType 表示向量数据库类型
type DatabaseType string

const (
	DatabaseTypeMilvus DatabaseType = "milvus"
	DatabaseTypeWeaviate DatabaseType = "weaviate"
	DatabaseTypePinecone DatabaseType = "pinecone"
	DatabaseTypeQdrant DatabaseType = "qdrant"
)

// DatabaseConfig 表示向量数据库配置
type DatabaseConfig struct {
	Type     DatabaseType   `json:"type" yaml:"type"`
	Milvus   *MilvusConfig  `json:"milvus,omitempty" yaml:"milvus,omitempty"`
	// 这里可以添加其他向量数据库的配置
}

// VectorDatabaseFactory 向量数据库工厂
type VectorDatabaseFactory struct{}

// NewVectorDatabaseFactory 创建一个新的向量数据库工厂
func NewVectorDatabaseFactory() *VectorDatabaseFactory {
	return &VectorDatabaseFactory{}
}

// CreateVectorDatabase 根据配置创建向量数据库客户端
func (f *VectorDatabaseFactory) CreateVectorDatabase(config *DatabaseConfig) (VectorDatabase, error) {
	if config == nil {
		return nil, ErrInvalidConfig("config cannot be nil")
	}

	switch config.Type {
	case DatabaseTypeMilvus:
		if config.Milvus == nil {
			return nil, ErrInvalidConfig("milvus config cannot be nil")
		}
		return NewMilvusClient(config.Milvus)
	case DatabaseTypeWeaviate:
		return nil, ErrInvalidConfig("weaviate not implemented yet")
	case DatabaseTypePinecone:
		return nil, ErrInvalidConfig("pinecone not implemented yet")
	case DatabaseTypeQdrant:
		return nil, ErrInvalidConfig("qdrant not implemented yet")
	default:
		return nil, ErrInvalidConfig(fmt.Sprintf("unsupported database type: %s", config.Type))
	}
}

// ConnectAndCreate 创建并连接向量数据库客户端
// ConnectAndCreate 连接并创建数据库
func (f *VectorDatabaseFactory) ConnectAndCreate(ctx context.Context, config *DatabaseConfig) (VectorDatabase, error) {
	db, err := f.CreateVectorDatabase(config)
	if err != nil {
		return nil, err
	}
	
	// 检查连接状态
	if err := db.Health(ctx); err != nil {
		return nil, err
	}
	
	return db, nil
}

// DefaultVectorDatabaseConfig 返回默认的向量数据库配置
func DefaultVectorDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:   DatabaseTypeMilvus,
		Milvus: DefaultMilvusConfig(),
	}
}