package service

import (
	"context"
)

// MilvusClient defines the interface for Milvus operations
type MilvusClient interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Health(ctx context.Context) error

	// Collection operations
	CreateCollection(ctx context.Context, schema *MilvusCollectionSchema) error
	DropCollection(ctx context.Context, collectionName string) error
	HasCollection(ctx context.Context, collectionName string) (bool, error)
	DescribeCollection(ctx context.Context, collectionName string) (*MilvusCollectionSchema, error)
	ListCollections(ctx context.Context) ([]string, error)
	LoadCollection(ctx context.Context, collectionName string) error
	ReleaseCollection(ctx context.Context, collectionName string) error
	IsCollectionLoaded(ctx context.Context, collectionName string) (bool, error)
	GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error)

	// Index operations
	CreateIndex(ctx context.Context, collectionName, fieldName string, params *MilvusIndexParams) error
	DropIndex(ctx context.Context, collectionName, fieldName string) error
	DescribeIndex(ctx context.Context, collectionName string) ([]*MilvusIndexInfo, error)
	RebuildIndex(ctx context.Context, collectionName, fieldName string) error

	// Vector operations
	Upsert(ctx context.Context, collectionName string, vectors []MilvusVector) ([]string, error)
	Search(ctx context.Context, collectionName string, params *MilvusSearchParams) ([]*MilvusSearchResult, error)
	QueryByID(ctx context.Context, collectionName, id string) ([]MilvusVector, error)
	DeleteByID(ctx context.Context, collectionName string, ids []string) error

	// Collection operations
	Compact(ctx context.Context, collectionName string) error
}

// MilvusCollectionSchema represents a Milvus collection schema
type MilvusCollectionSchema struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	AutoID      bool                `json:"auto_id"`
	Fields      []MilvusFieldSchema `json:"fields"`
}

// MilvusFieldSchema represents a field in a Milvus collection
type MilvusFieldSchema struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	PrimaryKey bool        `json:"primary_key,omitempty"`
	AutoID     bool        `json:"auto_id,omitempty"`
	MaxLength  int         `json:"max_length,omitempty"`
	Dimensions int         `json:"dimensions,omitempty"`
}

// MilvusFieldType constants
const (
	MilvusFieldTypeBool       = "BOOL"
	MilvusFieldTypeInt8       = "INT8"
	MilvusFieldTypeInt16      = "INT16"
	MilvusFieldTypeInt32      = "INT32"
	MilvusFieldTypeInt64      = "INT64"
	MilvusFieldTypeFloat      = "FLOAT"
	MilvusFieldTypeDouble     = "DOUBLE"
	MilvusFieldTypeVarChar    = "VARCHAR"
	MilvusFieldTypeJSON       = "JSON"
	MilvusFieldTypeFloatVector = "FLOAT_VECTOR"
	MilvusFieldTypeBinaryVector = "BINARY_VECTOR"
)

// MilvusIndexParams represents index parameters for Milvus
type MilvusIndexParams struct {
	IndexType   string                 `json:"index_type"`
	MetricType  string                 `json:"metric_type"`
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
}

// MilvusIndexInfo represents index information in Milvus
type MilvusIndexInfo struct {
	IndexType   string                 `json:"index_type"`
	MetricType  string                 `json:"metric_type"`
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
}

// MilvusVector represents a vector in Milvus
type MilvusVector struct {
	ID         string                 `json:"id"`
	Vector     []float64              `json:"vector"`
	ExternalID string                 `json:"external_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MilvusSearchParams represents search parameters for Milvus
type MilvusSearchParams struct {
	Vector      []float64              `json:"vector"`
	TopK        int                    `json:"top_k"`
	MetricType  string                 `json:"metric_type"`
	SearchParams map[string]interface{} `json:"search_params,omitempty"`
	Filter      string                 `json:"filter,omitempty"`
}

// MilvusSearchResult represents a search result from Milvus
type MilvusSearchResult struct {
	ID         string                 `json:"id"`
	Score      float64                `json:"score"`
	ExternalID string                 `json:"external_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}