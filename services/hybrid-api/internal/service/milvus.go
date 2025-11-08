package service

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// MilvusService implements VectorService interface using Milvus
type MilvusService struct {
	client     MilvusClient
	config     models.VectorDBConfig
	collections map[string]bool // Track initialized collections
}

// NewMilvusService creates a new MilvusService
func NewMilvusService(config models.VectorDBConfig) (*MilvusService, error) {
	client, err := NewMilvusClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Milvus client: %w", err)
	}

	return &MilvusService{
		client:     client,
		config:     config,
		collections: make(map[string]bool),
	}, nil
}

// Connect establishes connection to Milvus
func (s *MilvusService) Connect(ctx context.Context) error {
	return s.client.Connect(ctx)
}

// Disconnect closes connection to Milvus
func (s *MilvusService) Disconnect(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

// Health checks the health of Milvus
func (s *MilvusService) Health(ctx context.Context) error {
	return s.client.Health(ctx)
}

// CreateCollection creates a new vector collection
func (s *MilvusService) CreateCollection(ctx context.Context, req *models.CreateCollectionRequest) error {
	// Convert to Milvus collection schema
	schema := &MilvusCollectionSchema{
		Name:        req.Name,
		Description: req.Description,
		AutoID:      req.AutoID,
		Fields: []MilvusFieldSchema{
			{
				Name:       "id",
				Type:       MilvusFieldTypeInt64,
				PrimaryKey: true,
				AutoID:     req.AutoID,
			},
			{
				Name:       "external_id",
				Type:       MilvusFieldTypeVarChar,
				MaxLength:  65535,
			},
		},
	}

	// Add metadata fields if enabled
	if req.EnableMetadata {
		schema.Fields = append(schema.Fields, MilvusFieldSchema{
			Name:       "metadata",
			Type:       MilvusFieldTypeJSON,
		})
	}

	// Add vector field
	schema.Fields = append(schema.Fields, MilvusFieldSchema{
		Name:       "vector",
		Type:       MilvusFieldTypeFloatVector,
		Dimensions: req.Dimension,
	})

	// Create collection in Milvus
	err := s.client.CreateCollection(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Create index if requested
	if req.IndexType != "" {
		indexParams := &MilvusIndexParams{
			IndexType: req.IndexType,
			MetricType: req.MetricType,
		}

		if req.IndexType == "IVF_FLAT" || req.IndexType == "IVF_SQ8" || req.IndexType == "IVF_PQ" {
			indexParams.ExtraParams = map[string]interface{}{
				"nlist": req.IndexParams.Nlist,
			}
		} else if req.IndexType == "HNSW" {
			indexParams.ExtraParams = map[string]interface{}{
				"M":               req.IndexParams.M,
				"efConstruction": req.IndexParams.EfConstruction,
			}
		}

		err = s.client.CreateIndex(ctx, req.Name, "vector", indexParams)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Mark collection as initialized
	s.collections[req.Name] = true

	return nil
}

// DropCollection drops a vector collection
func (s *MilvusService) DropCollection(ctx context.Context, collectionName string) error {
	err := s.client.DropCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	// Remove from initialized collections
	delete(s.collections, collectionName)

	return nil
}

// HasCollection checks if a collection exists
func (s *MilvusService) HasCollection(ctx context.Context, collectionName string) (bool, error) {
	return s.client.HasCollection(ctx, collectionName)
}

// DescribeCollection describes a collection
func (s *MilvusService) DescribeCollection(ctx context.Context, collectionName string) (*models.VectorCollection, error) {
	schema, err := s.client.DescribeCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}

	// Convert Milvus schema to our model
	collection := &models.VectorCollection{
		Name:        schema.Name,
		Dims:        0, // Will be set below
		MetricType:  "", // Will be set below
		IndexType:   "", // Will be set below
		CreatedAt:   time.Now(), // Placeholder
	}

	// Extract dimension and other info from fields
	for _, field := range schema.Fields {
		if field.Name == "vector" {
			collection.Dims = field.Dimensions
		}
	}

	// Get index info to determine index type and metric type
	indexes, err := s.client.DescribeIndex(ctx, collectionName)
	if err == nil && len(indexes) > 0 {
		collection.IndexType = indexes[0].IndexType
		collection.MetricType = indexes[0].MetricType
	}

	return collection, nil
}

// ListCollections lists all collections
func (s *MilvusService) ListCollections(ctx context.Context) ([]string, error) {
	return s.client.ListCollections(ctx)
}

// UpsertVectors inserts or updates vectors in a collection
func (s *MilvusService) UpsertVectors(ctx context.Context, req *models.UpsertVectorsRequest) (*models.UpsertResponse, error) {
	// Initialize collection if not already done
	if !s.collections[req.CollectionName] {
		hasCollection, err := s.client.HasCollection(ctx, req.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to check collection existence: %w", err)
		}
		if !hasCollection {
			return nil, fmt.Errorf("collection %s does not exist", req.CollectionName)
		}
		s.collections[req.CollectionName] = true
	}

	// Load collection if not loaded
	loaded, err := s.client.IsCollectionLoaded(ctx, req.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if collection is loaded: %w", err)
	}
	if !loaded {
		err = s.client.LoadCollection(ctx, req.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to load collection: %w", err)
		}
	}

	// Convert vectors to Milvus format
	milvusVectors := make([]MilvusVector, len(req.Vectors))
	for i, vector := range req.Vectors {
		milvusVectors[i] = MilvusVector{
			ID:         vector.ID,
			Vector:     vector.Embedding,
			ExternalID: vector.ExternalID,
		}

		// Add metadata if available
		if vector.Metadata != nil {
			milvusVectors[i].Metadata = vector.Metadata
		}
	}

	// Upsert vectors
	ids, err := s.client.Upsert(ctx, req.CollectionName, milvusVectors)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert vectors: %w", err)
	}

	return &models.UpsertResponse{
		SuccessCount: len(ids),
		FailedCount:  0,
		InsertIDs:    ids,
	}, nil
}

// SearchVectors searches for similar vectors in a collection
func (s *MilvusService) SearchVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// Initialize collection if not already done
	if !s.collections[req.CollectionName] {
		hasCollection, err := s.client.HasCollection(ctx, req.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to check collection existence: %w", err)
		}
		if !hasCollection {
			return nil, fmt.Errorf("collection %s does not exist", req.CollectionName)
		}
		s.collections[req.CollectionName] = true
	}

	// Load collection if not loaded
	loaded, err := s.client.IsCollectionLoaded(ctx, req.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if collection is loaded: %w", err)
	}
	if !loaded {
		err = s.client.LoadCollection(ctx, req.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to load collection: %w", err)
		}
	}

	// Convert search request to Milvus format
	searchParams := &MilvusSearchParams{
		Vector:      req.QueryVector,
		TopK:        req.TopK,
		MetricType:  req.MetricType,
		SearchParams: map[string]interface{}{},
	}

	// Add search parameters based on index type
	if req.SearchParams != nil {
		searchParams.SearchParams = req.SearchParams
	}

	// Add filter expression if provided
	if req.Filter != nil {
		searchParams.Filter = fmt.Sprintf("%v", req.Filter)
	}

	// Search vectors
	results, err := s.client.Search(ctx, req.CollectionName, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// Convert results to our format
	searchResults := make([]models.SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = models.SearchResult{
			ID:         result.ID,
			Score:      result.Score,
			ExternalID: result.ExternalID,
		}

		// Add metadata if available
		if result.Metadata != nil {
			searchResults[i].Metadata = result.Metadata
		}
	}

	return &models.SearchResponse{
		Results: searchResults,
		Total:   len(searchResults),
	}, nil
}

// GetVector retrieves a vector by ID
func (s *MilvusService) GetVector(ctx context.Context, collectionName, vectorID string) (*models.VectorData, error) {
	// Initialize collection if not already done
	if !s.collections[collectionName] {
		hasCollection, err := s.client.HasCollection(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to check collection existence: %w", err)
		}
		if !hasCollection {
			return nil, fmt.Errorf("collection %s does not exist", collectionName)
		}
		s.collections[collectionName] = true
	}

	// Load collection if not loaded
	loaded, err := s.client.IsCollectionLoaded(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if collection is loaded: %w", err)
	}
	if !loaded {
		err = s.client.LoadCollection(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to load collection: %w", err)
		}
	}

	// Query by ID
	vectors, err := s.client.QueryByID(ctx, collectionName, vectorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query vector: %w", err)
	}

	if len(vectors) == 0 {
		return nil, fmt.Errorf("vector with ID %s not found", vectorID)
	}

	// Convert to our format
	vector := &models.VectorData{
		ID:         vectors[0].ID,
		Embedding:  vectors[0].Vector,
		ExternalID: vectors[0].ExternalID,
	}

	// Add metadata if available
	if vectors[0].Metadata != nil {
		vector.Metadata = vectors[0].Metadata
	}

	return vector, nil
}

// DeleteVectors deletes vectors by IDs
func (s *MilvusService) DeleteVectors(ctx context.Context, req *models.DeleteVectorsRequest) (*models.DeleteResponse, error) {
	// Initialize collection if not already done
	if !s.collections[req.CollectionName] {
		hasCollection, err := s.client.HasCollection(ctx, req.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to check collection existence: %w", err)
		}
		if !hasCollection {
			return nil, fmt.Errorf("collection %s does not exist", req.CollectionName)
		}
		s.collections[req.CollectionName] = true
	}

	// Delete vectors
	err := s.client.DeleteByID(ctx, req.CollectionName, req.Ids)
	if err != nil {
		return nil, fmt.Errorf("failed to delete vectors: %w", err)
	}

	return &models.DeleteResponse{
		SuccessCount: len(req.Ids),
		FailedCount:  0,
	}, nil
}

// CreateIndex creates an index on a collection
func (s *MilvusService) CreateIndex(ctx context.Context, req *models.CreateIndexRequest) error {
	// Convert to Milvus index params
	indexParams := &MilvusIndexParams{
		IndexType: req.IndexType,
		MetricType: req.MetricType,
	}

	// Add extra params based on index type
	if req.Params != nil {
		indexParams.ExtraParams = req.Params
	}

	return s.client.CreateIndex(ctx, req.CollectionName, req.FieldName, indexParams)
}

// DropIndex drops an index on a collection
func (s *MilvusService) DropIndex(ctx context.Context, collectionName, fieldName string) error {
	return s.client.DropIndex(ctx, collectionName, fieldName)
}

// DescribeIndex describes an index on a collection
func (s *MilvusService) DescribeIndex(ctx context.Context, collectionName, fieldName string) (*models.VectorIndex, error) {
	indexes, err := s.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	if len(indexes) == 0 {
		return nil, fmt.Errorf("no index found on field %s", fieldName)
	}

	// Convert to our format
	index := &models.VectorIndex{
		CollectionName: collectionName,
		FieldName:      fieldName,
		IndexType:      indexes[0].IndexType,
		MetricType:     indexes[0].MetricType,
	}

	// Add extra params if available
	if indexes[0].ExtraParams != nil {
		index.Params = indexes[0].ExtraParams
	}

	return index, nil
}

// GetCollectionStats gets statistics about a collection
func (s *MilvusService) GetCollectionStats(ctx context.Context, collectionName string) (*models.CollectionStats, error) {
	stats, err := s.client.GetCollectionStats(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	// Convert to our format
	collectionStats := &models.CollectionStats{
		CollectionName: collectionName,
		RowCount:       0, // Will be set below
	}

	// Extract row count from stats
	if rowCount, ok := stats["row_count"]; ok {
		if count, ok := rowCount.(int64); ok {
			collectionStats.RowCount = count
		}
	}

	return collectionStats, nil
}

// RebuildIndex rebuilds an index on a collection
func (s *MilvusService) RebuildIndex(ctx context.Context, collectionName, fieldName string) error {
	return s.client.RebuildIndex(ctx, collectionName, fieldName)
}

// Compact compacts a collection to reclaim space
func (s *MilvusService) Compact(ctx context.Context, collectionName string) error {
	return s.client.Compact(ctx, collectionName)
}