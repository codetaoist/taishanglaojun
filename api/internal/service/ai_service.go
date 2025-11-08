package service

import (
	"context"
	"fmt"
	"time"

	pb "github.com/codetaoist/taishanglaojun/api/proto"
	"github.com/codetaoist/taishanglaojun/api/internal/grpc"
)

// AIService wraps the AI gRPC client
type AIService struct {
	client *grpc.AIServiceClient
}

// NewAIService creates a new AI service
func NewAIService(client *grpc.AIServiceClient) *AIService {
	return &AIService{
		client: client,
	}
}

// Health checks the health of AI services
func (s *AIService) Health(ctx context.Context) (vectorHealthy, modelHealthy bool, err error) {
	// Check vector service health
	vectorResp, err := s.client.VectorHealthCheck(ctx)
	if err != nil {
		return false, false, fmt.Errorf("vector service health check failed: %w", err)
	}

	// Check model service health
	modelResp, err := s.client.ModelHealthCheck(ctx)
	if err != nil {
		return false, false, fmt.Errorf("model service health check failed: %w", err)
	}

	return vectorResp.Success, modelResp.Success, nil
}

// CreateCollection creates a new collection in the vector database
func (s *AIService) CreateCollection(ctx context.Context, name string, schema *pb.CollectionSchema) error {
	req := &pb.CreateCollectionRequest{
		CollectionName: name,
		Schema:         schema,
	}

	_, err := s.client.CreateCollection(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

// DropCollection drops a collection from the vector database
func (s *AIService) DropCollection(ctx context.Context, name string) error {
	req := &pb.DropCollectionRequest{
		CollectionName: name,
	}

	_, err := s.client.DropCollection(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	return nil
}

// HasCollection checks if a collection exists
func (s *AIService) HasCollection(ctx context.Context, name string) (bool, error) {
	req := &pb.HasCollectionRequest{
		CollectionName: name,
	}

	resp, err := s.client.HasCollection(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return resp.Has, nil
}

// ListCollections lists all collections
func (s *AIService) ListCollections(ctx context.Context) ([]*pb.Collection, error) {
	req := &pb.ListCollectionsRequest{}

	resp, err := s.client.ListCollections(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return resp.Collections, nil
}

// Search searches for vectors in a collection
func (s *AIService) Search(ctx context.Context, collectionName string, vector []float64, topK int64) ([]*pb.SearchResult, error) {
	// Convert float64 to float32
	vector32 := make([]float32, len(vector))
	for i, v := range vector {
		vector32[i] = float32(v)
	}

	req := &pb.SearchRequest{
		CollectionName: collectionName,
		Vector:         vector32,
		TopK:           topK,
	}

	resp, err := s.client.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	return resp.Results, nil
}

// Insert inserts data into a collection
func (s *AIService) Insert(ctx context.Context, collectionName string, data []*pb.VectorData) ([]string, error) {
	req := &pb.InsertRequest{
		CollectionName: collectionName,
		Data:           data,
	}

	resp, err := s.client.Insert(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to insert data: %w", err)
	}

	return resp.Ids, nil
}

// GenerateText generates text using a model
func (s *AIService) GenerateText(ctx context.Context, modelName, prompt string, maxLength int64, temperature float32) ([]string, error) {
	req := &pb.TextGenerationRequest{
		ModelName:           modelName,
		Prompt:              prompt,
		MaxLength:           maxLength,
		Temperature:         temperature,
		TopP:                0.9,
		NumReturnSequences:  1,
		DoSample:            true,
	}

	resp, err := s.client.GenerateText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("text generation failed: %s", resp.Message)
	}

	return resp.Texts, nil
}

// GenerateEmbedding generates embeddings using a model
func (s *AIService) GenerateEmbedding(ctx context.Context, modelName string, texts []string) ([][]float64, error) {
	req := &pb.EmbeddingRequest{
		ModelName: modelName,
		Texts:     texts,
	}

	resp, err := s.client.GenerateEmbedding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("embedding generation failed: %s", resp.Message)
	}

	// Note: The proto definition has a single-dimensional embeddings array
	// This is a limitation of the current proto definition
	// For now, we'll return a single embedding as a 2D array
	embedding := make([]float64, len(resp.Embeddings))
	for i, v := range resp.Embeddings {
		embedding[i] = float64(v)
	}

	return [][]float64{embedding}, nil
}

// DescribeCollection describes a collection
func (s *AIService) DescribeCollection(ctx context.Context, name string) (*pb.Collection, error) {
	req := &pb.DescribeCollectionRequest{
		CollectionName: name,
	}

	resp, err := s.client.DescribeCollection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("describe collection failed: %s", resp.Message)
	}

	// Note: DescribeCollectionResponse doesn't contain Collection field
	// We'll need to create a placeholder or modify the response structure
	return nil, fmt.Errorf("describe collection response structure needs to be updated")
}

// LoadCollection loads a collection into memory
func (s *AIService) LoadCollection(ctx context.Context, name string, replicaNumber int64) error {
	req := &pb.LoadCollectionRequest{
		CollectionName: name,
		ReplicaNumber:  replicaNumber,
	}

	resp, err := s.client.LoadCollection(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("load collection failed: %s", resp.Message)
	}

	return nil
}

// ReleaseCollection releases a collection from memory
func (s *AIService) ReleaseCollection(ctx context.Context, name string) error {
	req := &pb.ReleaseCollectionRequest{
		CollectionName: name,
	}

	resp, err := s.client.ReleaseCollection(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to release collection: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("release collection failed: %s", resp.Message)
	}

	return nil
}

// GetCollectionStatistics gets statistics about a collection
func (s *AIService) GetCollectionStatistics(ctx context.Context, name string) (string, error) {
	req := &pb.GetCollectionStatisticsRequest{
		CollectionName: name,
	}

	resp, err := s.client.GetCollectionStatistics(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get collection statistics: %w", err)
	}

	return resp.Stats, nil
}

// CreateIndex creates an index for a collection
func (s *AIService) CreateIndex(ctx context.Context, collectionName string, indexParams *pb.IndexParams) error {
	req := &pb.CreateIndexRequest{
		CollectionName: collectionName,
		IndexParams:    indexParams,
	}

	resp, err := s.client.CreateIndex(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("create index failed: %s", resp.Message)
	}

	return nil
}

// DropIndex drops an index from a collection
func (s *AIService) DropIndex(ctx context.Context, collectionName, indexName string) error {
	req := &pb.DropIndexRequest{
		CollectionName: collectionName,
		IndexName:      indexName,
	}

	resp, err := s.client.DropIndex(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("drop index failed: %s", resp.Message)
	}

	return nil
}

// DescribeIndex describes an index
func (s *AIService) DescribeIndex(ctx context.Context, collectionName, indexName string) ([]*pb.IndexParams, error) {
	req := &pb.DescribeIndexRequest{
		CollectionName: collectionName,
		IndexName:      indexName,
	}

	resp, err := s.client.DescribeIndex(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	return resp.Indexes, nil
}

// Delete deletes data from a collection
func (s *AIService) Delete(ctx context.Context, collectionName, expr string) ([]string, error) {
	req := &pb.DeleteRequest{
		CollectionName: collectionName,
		Expr:            expr,
	}

	_, err := s.client.Delete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete data: %w", err)
	}

	// Note: DeleteResponse returns delete_count, not IDs
	// We'll return an empty slice for now
	return []string{}, nil
}

// GetById gets data by IDs from a collection
func (s *AIService) GetById(ctx context.Context, collectionName string, ids, outputFields []string) ([]*pb.VectorData, error) {
	req := &pb.GetByIdRequest{
		CollectionName: collectionName,
		Ids:            ids,
		OutputFields:   outputFields,
	}

	resp, err := s.client.GetById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get data by IDs: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get data by IDs failed: %s", resp.Message)
	}

	return resp.Data, nil
}

// Compact compacts a collection
func (s *AIService) Compact(ctx context.Context, collectionName string) error {
	req := &pb.CompactRequest{
		CollectionName: collectionName,
	}

	_, err := s.client.Compact(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to compact collection: %w", err)
	}

	return nil
}

// ListModels lists all registered models
func (s *AIService) ListModels(ctx context.Context) ([]*pb.ModelConfig, error) {
	req := &pb.ListModelsRequest{}

	resp, err := s.client.ListModels(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	return resp.Models, nil
}

// LoadModel loads a model into memory
func (s *AIService) LoadModel(ctx context.Context, modelName string) error {
	req := &pb.LoadModelRequest{
		Name: modelName,
	}

	resp, err := s.client.LoadModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("model loading failed: %s", resp.Message)
	}

	return nil
}

// GetModelStatus gets the status of a model
func (s *AIService) GetModelStatus(ctx context.Context, modelName string) (bool, time.Time, error) {
	req := &pb.GetModelStatusRequest{
		Name: modelName,
	}

	resp, err := s.client.GetModelStatus(ctx, req)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to get model status: %w", err)
	}

	return resp.IsLoaded, time.Unix(resp.LoadTime, 0), nil
}

// RegisterModel registers a new model
func (s *AIService) RegisterModel(ctx context.Context, name string, provider pb.ModelProvider, modelPath, description string, config map[string]string) error {
	req := &pb.RegisterModelRequest{
		Name:        name,
		Provider:    provider,
		ModelPath:   modelPath,
		Description: description,
		Config:      config,
	}

	resp, err := s.client.RegisterModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register model: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("model registration failed: %s", resp.Message)
	}

	return nil
}

// UpdateModel updates an existing model
func (s *AIService) UpdateModel(ctx context.Context, name, modelPath, description string, config map[string]string) error {
	req := &pb.UpdateModelRequest{
		Name:        name,
		ModelPath:   modelPath,
		Description: description,
		Config:      config,
	}

	resp, err := s.client.UpdateModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("model update failed: %s", resp.Message)
	}

	return nil
}

// UnregisterModel unregisters a model
func (s *AIService) UnregisterModel(ctx context.Context, name string) error {
	req := &pb.UnregisterModelRequest{
		Name: name,
	}

	resp, err := s.client.UnregisterModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unregister model: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("model unregistration failed: %s", resp.Message)
	}

	return nil
}

// GetModel gets a model by name
func (s *AIService) GetModel(ctx context.Context, name string) (*pb.ModelConfig, error) {
	req := &pb.GetModelRequest{
		Name: name,
	}

	resp, err := s.client.GetModel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return resp.Model, nil
}

// UnloadModel unloads a model from memory
func (s *AIService) UnloadModel(ctx context.Context, name string) error {
	req := &pb.UnloadModelRequest{
		Name: name,
	}

	resp, err := s.client.UnloadModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unload model: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("model unloading failed: %s", resp.Message)
	}

	return nil
}