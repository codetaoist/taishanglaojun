package models

import (
	"time"
)

// VectorCollectionConfig represents the configuration for a vector collection
type VectorCollectionConfig struct {
	CollectionName string                 `json:"collectionName"`
	Dimension      int                    `json:"dimension"`
	IndexType      string                 `json:"indexType"`
	MetricType     string                 `json:"metricType"`
	ExtraParams    map[string]interface{} `json:"extraParams"`
}

// VectorIndex represents a vector index configuration
type VectorIndex struct {
	CollectionName string                 `json:"collectionName"`
	IndexName      string                 `json:"indexName"`
	IndexType      string                 `json:"indexType"`
	MetricType     string                 `json:"metricType"`
	ExtraParams    map[string]interface{} `json:"extraParams"`
	
	// 兼容字段
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Type     string                 `json:"type,omitempty"`
	State    string                 `json:"state,omitempty"`
	FieldName string                 `json:"fieldName,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
}

// VectorData represents a vector with its metadata
type VectorData struct {
	ID         string                 `json:"id"`
	Vector     []float64              `json:"vector"`
	Metadata   map[string]interface{} `json:"metadata"`
	
	// 兼容字段
	Embedding  []float64              `json:"embedding,omitempty"`
	ExternalID string                 `json:"externalId,omitempty"`
}

// VectorSearchResult 搜索结果项
type VectorSearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Vector   []float64              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	
	// 兼容字段
	ExternalID string `json:"externalId,omitempty"`
}

// SearchResult 搜索结果（别名）
type SearchResult = VectorSearchResult

// VectorSearchRequest represents a search request
type VectorSearchRequest struct {
	CollectionName string                 `json:"collectionName"`
	QueryVector    []float64              `json:"queryVector"`
	TopK           int                    `json:"topK"`
	Filter         map[string]interface{} `json:"filter,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
	
	// 兼容字段
	MetricType     string                 `json:"metricType,omitempty"`
	SearchParams   map[string]interface{} `json:"searchParams,omitempty"`
}

// VectorSearchResponse represents a search response
type VectorSearchResponse struct {
	Results []VectorSearchResult `json:"results"`
	Total   int                  `json:"total"`
}

// VectorCollectionStats represents statistics for a vector collection
type VectorCollectionStats struct {
	CollectionName string    `json:"collectionName"`
	VectorCount    int64     `json:"vectorCount"`
	IndexSize      int64     `json:"indexSize"`
	LastUpdated    time.Time `json:"lastUpdated"`
}

// VectorDatabaseConfig represents the configuration for a vector database
type VectorDatabaseConfig struct {
	Type     string                 `json:"type"` // milvus, weaviate, pinecone, qdrant
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Username string                 `json:"username,omitempty"`
	Password string                 `json:"password,omitempty"`
	Database string                 `json:"database,omitempty"`
	TLS      bool                   `json:"tls"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
	
	// 兼容字段
	Endpoint string `json:"endpoint,omitempty"`
}

// VectorDatabaseStatus represents the status of a vector database
type VectorDatabaseStatus struct {
	Connected   bool      `json:"connected"`
	LastChecked time.Time `json:"lastChecked"`
	Error       string    `json:"error,omitempty"`
}

// VectorDatabaseInfo represents information about a vector database
type VectorDatabaseInfo struct {
	Type        string   `json:"type"`
	Version     string   `json:"version"`
	Collections []string `json:"collections"`
}

// IndexParams 索引参数
type IndexParams struct {
	Nlist          int `json:"nlist,omitempty"`
	M              int `json:"m,omitempty"`
	EfConstruction int `json:"efConstruction,omitempty"`
}

// VectorOperation represents an operation on vectors
type VectorOperation struct {
	Type        string                 `json:"type"` // insert, update, delete, search
	Collection  string                 `json:"collection"`
	Data        interface{}            `json:"data"`
	Params      map[string]interface{} `json:"params,omitempty"`
	RequestedAt time.Time              `json:"requestedAt"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	Status      string                 `json:"status"` // pending, running, completed, failed
	Error       string                 `json:"error,omitempty"`
}

// VectorDBConfig 向量数据库配置（别名）
type VectorDBConfig = VectorDatabaseConfig

// CreateCollectionRequest 创建集合请求
type CreateCollectionRequest struct {
	CollectionName string                 `json:"collectionName"`
	Dimension      int                    `json:"dimension"`
	IndexType      string                 `json:"indexType"`
	MetricType     string                 `json:"metricType"`
	ExtraParams    map[string]interface{} `json:"extraParams,omitempty"`
	
	// 兼容字段
	Name           string                 `json:"name,omitempty"`
	Description    string                 `json:"description,omitempty"`
	AutoID         bool                   `json:"autoId,omitempty"`
	EnableMetadata bool                   `json:"enableMetadata,omitempty"`
	IndexParams    IndexParams            `json:"indexParams,omitempty"`
}

// UpsertVectorsRequest 插入/更新向量请求
type UpsertVectorsRequest struct {
	CollectionName string                 `json:"collectionName"`
	Vectors        []VectorData           `json:"vectors"`
	ExtraParams    map[string]interface{} `json:"extraParams,omitempty"`
}

// UpsertResponse 插入/更新响应
type UpsertResponse struct {
	InsertedCount int      `json:"insertedCount"`
	UpdatedCount  int      `json:"updatedCount"`
	Ids           []string `json:"ids"`
	
	// 兼容字段
	SuccessCount  int      `json:"successCount,omitempty"`
	FailedCount   int      `json:"failedCount,omitempty"`
	InsertIDs     []string `json:"insertIds,omitempty"`
}

// SearchRequest 搜索请求（别名）
type SearchRequest = VectorSearchRequest

// SearchResponse 搜索响应（别名）
type SearchResponse = VectorSearchResponse

// DeleteVectorsRequest 删除向量请求
type DeleteVectorsRequest struct {
	CollectionName string   `json:"collectionName"`
	Ids            []string `json:"ids"`
	Filter         map[string]interface{} `json:"filter,omitempty"`
	ExtraParams    map[string]interface{} `json:"extraParams,omitempty"`
}

// DeleteResponse 删除响应
type DeleteResponse struct {
	DeletedCount int `json:"deletedCount"`
	
	// 兼容字段
	SuccessCount int `json:"successCount,omitempty"`
	FailedCount  int `json:"failedCount,omitempty"`
}

// CreateIndexRequest 创建索引请求
type CreateIndexRequest struct {
	CollectionName string                 `json:"collectionName"`
	IndexName      string                 `json:"indexName,omitempty"`
	IndexType      string                 `json:"indexType"`
	MetricType     string                 `json:"metricType"`
	ExtraParams    map[string]interface{} `json:"extraParams,omitempty"`
	
	// 兼容字段
	FieldName      string                 `json:"fieldName,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
}

// CollectionStats 集合统计信息
type CollectionStats struct {
	Name      string `json:"name"`
	Count     int64  `json:"count"`
	Size      int64  `json:"size"`
	Indexed   bool   `json:"indexed"`
	
	// 兼容字段
	RowCount       int64  `json:"rowCount,omitempty"`
	CollectionName string `json:"collectionName,omitempty"`
}

// UsageInfo 使用信息（别名，与models.Usage兼容）
type UsageInfo = Usage

// Collection 表示向量集合
type Collection struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Dimension   int       `json:"dimension"`
	MetricType  string    `json:"metricType"`
	VectorCount int64     `json:"vectorCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CollectionModel 表示集合在关系数据库中的模型
type CollectionModel struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TenantID         string    `json:"tenantId" gorm:"index"`
	Name             string    `json:"name" gorm:"not null"`
	Description      string    `json:"description"`
	ModelID          string    `json:"modelId"`
	Dimension        int       `json:"dimension"`
	IndexType        string    `json:"indexType"`
	MetricType       string    `json:"metricType"`
	ExtraIndexArgs   string    `json:"extraIndexArgs"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// VectorCollectionInfo 表示向量数据库中的集合信息
type VectorCollectionInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Dimension   int       `json:"dimension"`
	MetricType  string    `json:"metricType"`
	VectorCount int64     `json:"vectorCount"`
	IndexSize   int64     `json:"indexSize"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}