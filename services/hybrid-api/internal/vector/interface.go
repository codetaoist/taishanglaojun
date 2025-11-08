package vector

import (
	"context"
	"time"
)

// Vector 表示一个向量数据
type Vector struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResult 表示向量搜索结果
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Vector   []float32              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchOptions 表示向量搜索选项
type SearchOptions struct {
	TopK          int                    `json:"topk"`
	IncludeVector bool                   `json:"include_vector"`
	Filter        map[string]interface{} `json:"filter,omitempty"`
}

// IndexType 表示向量索引类型
type IndexType string

const (
	IndexTypeFlat  IndexType = "FLAT"
	IndexTypeIVF   IndexType = "IVF_FLAT"
	IndexTypeIVFSQ IndexType = "IVF_SQ8"
	IndexTypeHNSW  IndexType = "HNSW"
)

// MetricType 表示距离度量类型
type MetricType string

const (
	MetricTypeL2         MetricType = "L2"
	MetricTypeIP         MetricType = "IP"
	MetricTypeCosine     MetricType = "COSINE"
	MetricTypeHamming    MetricType = "HAMMING"
	MetricTypeJaccard    MetricType = "JACCARD"
	MetricTypeTanimoto   MetricType = "TANIMOTO"
	MetricTypeSubstructure MetricType = "SUBSTRUCTURE"
	MetricTypeSuperstructure MetricType = "SUPERSTRUCTURE"
)

// IndexParams 表示索引参数
type IndexParams struct {
	IndexType IndexType            `json:"index_type"`
	MetricType MetricType          `json:"metric_type"`
	Params     map[string]interface{} `json:"params"`
}

// CollectionInfo 表示集合信息
type CollectionInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	VectorDim   int          `json:"vector_dim"`
	IndexParams IndexParams  `json:"index_params"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CollectionStats 表示集合统计信息
type CollectionStats struct {
	Name        string `json:"name"`
	VectorCount int64  `json:"vector_count"`
	SizeInBytes int64  `json:"size_in_bytes"`
}

// VectorDatabase 定义向量数据库接口
type VectorDatabase interface {
	// 集合管理
	CreateCollection(ctx context.Context, name, description string, vectorDim int, indexParams IndexParams) error
	DropCollection(ctx context.Context, name string) error
	HasCollection(ctx context.Context, name string) (bool, error)
	ListCollections(ctx context.Context) ([]string, error)
	GetCollectionInfo(ctx context.Context, name string) (*CollectionInfo, error)
	GetCollectionStats(ctx context.Context, name string) (*CollectionStats, error)
	
	// 索引管理
	CreateIndex(ctx context.Context, req interface{}) error // 使用interface{}以支持不同类型的请求
	DropIndex(ctx context.Context, collectionName string) error
	HasIndex(ctx context.Context, collectionName string) (bool, error)
	
	// 向量操作
	Insert(ctx context.Context, collectionName string, vectors []Vector) error
	Upsert(ctx context.Context, collectionName string, vectors []Vector) error
	Delete(ctx context.Context, collectionName string, ids []string) error
	
	// 向量查询
	Search(ctx context.Context, collectionName string, queryVector []float32, opts SearchOptions) ([]SearchResult, error)
	GetByID(ctx context.Context, collectionName string, id string) (*Vector, error)
	
	// 数据库管理
	Health(ctx context.Context) error
	Close() error
	Compact(ctx context.Context, collectionName string) error
}