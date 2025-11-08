package service

import (
	"context"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// VectorService defines the interface for vector database operations
type VectorService interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Health(ctx context.Context) error

	// Collection management
	CreateCollection(ctx context.Context, req *models.CreateCollectionRequest) error
	DropCollection(ctx context.Context, collectionName string) error
	HasCollection(ctx context.Context, collectionName string) (bool, error)
	DescribeCollection(ctx context.Context, collectionName string) (*models.VectorCollection, error)
	ListCollections(ctx context.Context) ([]string, error)

	// Vector operations
	UpsertVectors(ctx context.Context, req *models.UpsertVectorsRequest) (*models.UpsertResponse, error)
	SearchVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	GetVector(ctx context.Context, collectionName, vectorID string) (*models.VectorData, error)
	DeleteVectors(ctx context.Context, req *models.DeleteVectorsRequest) (*models.DeleteResponse, error)

	// Index operations
	CreateIndex(ctx context.Context, req *models.CreateIndexRequest) error
	DropIndex(ctx context.Context, collectionName, fieldName string) error
	DescribeIndex(ctx context.Context, collectionName, fieldName string) (*models.VectorIndex, error)

	// Collection operations
	GetCollectionStats(ctx context.Context, collectionName string) (*models.CollectionStats, error)
	RebuildIndex(ctx context.Context, collectionName, fieldName string) error
	Compact(ctx context.Context, collectionName string) error
}

// VectorServiceFactory creates vector service instances based on configuration
type VectorServiceFactory struct{}

// NewVectorServiceFactory creates a new VectorServiceFactory
func NewVectorServiceFactory() *VectorServiceFactory {
	return &VectorServiceFactory{}
}

// CreateService creates a vector service based on the provided configuration
func (f *VectorServiceFactory) CreateService(config models.VectorDBConfig) (VectorService, error) {
	switch config.Type {
	case "milvus":
		return NewMilvusService(config)
	case "qdrant":
		return NewQdrantService(config)
	case "weaviate":
		return NewWeaviateService(config)
	default:
		return NewMilvusService(config) // Default to Milvus
	}
}