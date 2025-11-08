package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "github.com/codetaoist/taishanglaojun/api/proto"
)

// AIServiceClient wraps the gRPC client for AI services
type AIServiceClient struct {
	vectorConn   *grpc.ClientConn
	modelConn    *grpc.ClientConn
	vectorClient pb.VectorServiceClient
	modelClient  pb.ModelServiceClient
}

// NewAIServiceClient creates a new AI service client
func NewAIServiceClient(vectorAddr, modelAddr string) (*AIServiceClient, error) {
	// Configure gRPC client options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Connect to Vector service
	vectorConn, err := grpc.Dial(vectorAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vector service: %w", err)
	}

	// Connect to Model service
	modelConn, err := grpc.Dial(modelAddr, opts...)
	if err != nil {
		vectorConn.Close()
		return nil, fmt.Errorf("failed to connect to model service: %w", err)
	}

	return &AIServiceClient{
		vectorConn:   vectorConn,
		modelConn:    modelConn,
		vectorClient: pb.NewVectorServiceClient(vectorConn),
		modelClient:  pb.NewModelServiceClient(modelConn),
	}, nil
}

// Close closes the gRPC connections
func (c *AIServiceClient) Close() error {
	var errs []error

	if c.vectorConn != nil {
		if err := c.vectorConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("vector connection close error: %w", err))
		}
	}

	if c.modelConn != nil {
		if err := c.modelConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("model connection close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// HealthCheck checks the health of the vector service
func (c *AIServiceClient) VectorHealthCheck(ctx context.Context) (*pb.HealthCheckResponse, error) {
	return c.vectorClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
}

// HealthCheck checks the health of the model service
func (c *AIServiceClient) ModelHealthCheck(ctx context.Context) (*pb.HealthCheckResponse, error) {
	return c.modelClient.HealthCheck(ctx, &pb.HealthCheckRequest{})
}

// CreateCollection creates a new collection in the vector database
func (c *AIServiceClient) CreateCollection(ctx context.Context, req *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	return c.vectorClient.CreateCollection(ctx, req)
}

// DropCollection drops a collection from the vector database
func (c *AIServiceClient) DropCollection(ctx context.Context, req *pb.DropCollectionRequest) (*pb.DropCollectionResponse, error) {
	return c.vectorClient.DropCollection(ctx, req)
}

// HasCollection checks if a collection exists
func (c *AIServiceClient) HasCollection(ctx context.Context, req *pb.HasCollectionRequest) (*pb.HasCollectionResponse, error) {
	return c.vectorClient.HasCollection(ctx, req)
}

// DescribeCollection describes a collection
func (c *AIServiceClient) DescribeCollection(ctx context.Context, req *pb.DescribeCollectionRequest) (*pb.DescribeCollectionResponse, error) {
	return c.vectorClient.DescribeCollection(ctx, req)
}

// LoadCollection loads a collection into memory
func (c *AIServiceClient) LoadCollection(ctx context.Context, req *pb.LoadCollectionRequest) (*pb.LoadCollectionResponse, error) {
	return c.vectorClient.LoadCollection(ctx, req)
}

// ReleaseCollection releases a collection from memory
func (c *AIServiceClient) ReleaseCollection(ctx context.Context, req *pb.ReleaseCollectionRequest) (*pb.ReleaseCollectionResponse, error) {
	return c.vectorClient.ReleaseCollection(ctx, req)
}

// GetCollectionStatistics gets statistics about a collection
func (c *AIServiceClient) GetCollectionStatistics(ctx context.Context, req *pb.GetCollectionStatisticsRequest) (*pb.GetCollectionStatisticsResponse, error) {
	return c.vectorClient.GetCollectionStatistics(ctx, req)
}

// ListCollections lists all collections
func (c *AIServiceClient) ListCollections(ctx context.Context, req *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	return c.vectorClient.ListCollections(ctx, req)
}

// CreateIndex creates an index for a collection
func (c *AIServiceClient) CreateIndex(ctx context.Context, req *pb.CreateIndexRequest) (*pb.CreateIndexResponse, error) {
	return c.vectorClient.CreateIndex(ctx, req)
}

// DropIndex drops an index from a collection
func (c *AIServiceClient) DropIndex(ctx context.Context, req *pb.DropIndexRequest) (*pb.DropIndexResponse, error) {
	return c.vectorClient.DropIndex(ctx, req)
}

// DescribeIndex describes an index
func (c *AIServiceClient) DescribeIndex(ctx context.Context, req *pb.DescribeIndexRequest) (*pb.DescribeIndexResponse, error) {
	return c.vectorClient.DescribeIndex(ctx, req)
}

// Insert inserts data into a collection
func (c *AIServiceClient) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResponse, error) {
	return c.vectorClient.Insert(ctx, req)
}

// Delete deletes data from a collection
func (c *AIServiceClient) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return c.vectorClient.Delete(ctx, req)
}

// Search searches for vectors in a collection
func (c *AIServiceClient) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	return c.vectorClient.Search(ctx, req)
}

// GetById gets data by IDs from a collection
func (c *AIServiceClient) GetById(ctx context.Context, req *pb.GetByIdRequest) (*pb.GetByIdResponse, error) {
	return c.vectorClient.GetById(ctx, req)
}

// Compact compacts a collection
func (c *AIServiceClient) Compact(ctx context.Context, req *pb.CompactRequest) (*pb.CompactResponse, error) {
	return c.vectorClient.Compact(ctx, req)
}

// RegisterModel registers a new model
func (c *AIServiceClient) RegisterModel(ctx context.Context, req *pb.RegisterModelRequest) (*pb.RegisterModelResponse, error) {
	return c.modelClient.RegisterModel(ctx, req)
}

// UpdateModel updates an existing model
func (c *AIServiceClient) UpdateModel(ctx context.Context, req *pb.UpdateModelRequest) (*pb.UpdateModelResponse, error) {
	return c.modelClient.UpdateModel(ctx, req)
}

// UnregisterModel unregisters a model
func (c *AIServiceClient) UnregisterModel(ctx context.Context, req *pb.UnregisterModelRequest) (*pb.UnregisterModelResponse, error) {
	return c.modelClient.UnregisterModel(ctx, req)
}

// ListModels lists all registered models
func (c *AIServiceClient) ListModels(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	return c.modelClient.ListModels(ctx, req)
}

// GetModel gets a model by name
func (c *AIServiceClient) GetModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	return c.modelClient.GetModel(ctx, req)
}

// LoadModel loads a model into memory
func (c *AIServiceClient) LoadModel(ctx context.Context, req *pb.LoadModelRequest) (*pb.LoadModelResponse, error) {
	return c.modelClient.LoadModel(ctx, req)
}

// UnloadModel unloads a model from memory
func (c *AIServiceClient) UnloadModel(ctx context.Context, req *pb.UnloadModelRequest) (*pb.UnloadModelResponse, error) {
	return c.modelClient.UnloadModel(ctx, req)
}

// GetModelStatus gets the status of a model
func (c *AIServiceClient) GetModelStatus(ctx context.Context, req *pb.GetModelStatusRequest) (*pb.GetModelStatusResponse, error) {
	return c.modelClient.GetModelStatus(ctx, req)
}

// GenerateText generates text using a model
func (c *AIServiceClient) GenerateText(ctx context.Context, req *pb.TextGenerationRequest) (*pb.TextGenerationResponse, error) {
	return c.modelClient.GenerateText(ctx, req)
}

// GenerateEmbedding generates embeddings using a model
func (c *AIServiceClient) GenerateEmbedding(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	return c.modelClient.GenerateEmbedding(ctx, req)
}