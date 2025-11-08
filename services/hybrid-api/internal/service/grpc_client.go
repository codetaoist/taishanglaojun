package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbVector "github.com/codetaoist/api/proto/vector"
	pbModel "github.com/codetaoist/api/proto/model"
)

// VectorServiceClientImpl implements the VectorServiceClient interface
type VectorServiceClientImpl struct {
	conn   *grpc.ClientConn
	client pbVector.VectorServiceClient
}

// NewVectorServiceClientImpl creates a new VectorServiceClientImpl
func NewVectorServiceClientImpl(ctx context.Context, address string) (*VectorServiceClientImpl, error) {
	// Create gRPC connection
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vector service: %v", err)
	}

	// Create client
	client := pbVector.NewVectorServiceClient(conn)

	return &VectorServiceClientImpl{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the connection
func (c *VectorServiceClientImpl) Close() error {
	return c.conn.Close()
}

// HealthCheck performs a health check
func (c *VectorServiceClientImpl) HealthCheck(ctx context.Context) error {
	req := &pbVector.HealthCheckRequest{
		Service: "vector-service",
	}

	_, err := c.client.HealthCheck(ctx, req)
	return err
}

// CreateCollection creates a new collection
func (c *VectorServiceClientImpl) CreateCollection(ctx context.Context, request *pbVector.CreateCollectionRequest) (*pbVector.CreateCollectionResponse, error) {
	return c.client.CreateCollection(ctx, request)
}

// DropCollection drops a collection
func (c *VectorServiceClientImpl) DropCollection(ctx context.Context, request *pbVector.DropCollectionRequest) (*pbVector.DropCollectionResponse, error) {
	return c.client.DropCollection(ctx, request)
}

// HasCollection checks if a collection exists
func (c *VectorServiceClientImpl) HasCollection(ctx context.Context, request *pbVector.HasCollectionRequest) (*pbVector.HasCollectionResponse, error) {
	return c.client.HasCollection(ctx, request)
}

// ListCollections lists all collections
func (c *VectorServiceClientImpl) ListCollections(ctx context.Context, request *pbVector.ListCollectionsRequest) (*pbVector.ListCollectionsResponse, error) {
	return c.client.ListCollections(ctx, request)
}

// GetCollectionInfo gets collection information
func (c *VectorServiceClientImpl) GetCollectionInfo(ctx context.Context, request *pbVector.GetCollectionInfoRequest) (*pbVector.GetCollectionInfoResponse, error) {
	return c.client.GetCollectionInfo(ctx, request)
}

// GetCollectionStats gets collection statistics
func (c *VectorServiceClientImpl) GetCollectionStats(ctx context.Context, request *pbVector.GetCollectionStatsRequest) (*pbVector.GetCollectionStatsResponse, error) {
	return c.client.GetCollectionStats(ctx, request)
}

// CreateIndex creates an index
func (c *VectorServiceClientImpl) CreateIndex(ctx context.Context, request *pbVector.CreateIndexRequest) (*pbVector.CreateIndexResponse, error) {
	return c.client.CreateIndex(ctx, request)
}

// DropIndex drops an index
func (c *VectorServiceClientImpl) DropIndex(ctx context.Context, request *pbVector.DropIndexRequest) (*pbVector.DropIndexResponse, error) {
	return c.client.DropIndex(ctx, request)
}

// HasIndex checks if an index exists
func (c *VectorServiceClientImpl) HasIndex(ctx context.Context, request *pbVector.HasIndexRequest) (*pbVector.HasIndexResponse, error) {
	return c.client.HasIndex(ctx, request)
}

// DescribeIndex describes an index
func (c *VectorServiceClientImpl) DescribeIndex(ctx context.Context, request *pbVector.DescribeIndexRequest) (*pbVector.DescribeIndexResponse, error) {
	return c.client.DescribeIndex(ctx, request)
}

// Insert inserts data into a collection
func (c *VectorServiceClientImpl) Insert(ctx context.Context, request *pbVector.InsertRequest) (*pbVector.InsertResponse, error) {
	return c.client.Insert(ctx, request)
}

// Delete deletes data from a collection
func (c *VectorServiceClientImpl) Delete(ctx context.Context, request *pbVector.DeleteRequest) (*pbVector.DeleteResponse, error) {
	return c.client.Delete(ctx, request)
}

// Upsert upserts data into a collection
func (c *VectorServiceClientImpl) Upsert(ctx context.Context, request *pbVector.UpsertRequest) (*pbVector.UpsertResponse, error) {
	return c.client.Upsert(ctx, request)
}

// Search performs vector search
func (c *VectorServiceClientImpl) Search(ctx context.Context, request *pbVector.SearchRequest) (*pbVector.SearchResponse, error) {
	return c.client.Search(ctx, request)
}

// Query performs data query
func (c *VectorServiceClientImpl) Query(ctx context.Context, request *pbVector.QueryRequest) (*pbVector.QueryResponse, error) {
	return c.client.Query(ctx, request)
}

// ModelServiceClientImpl implements the ModelServiceClient interface
type ModelServiceClientImpl struct {
	conn   *grpc.ClientConn
	client pbModel.ModelServiceClient
}

// NewModelServiceClientImpl creates a new ModelServiceClientImpl
func NewModelServiceClientImpl(ctx context.Context, address string) (*ModelServiceClientImpl, error) {
	// Create gRPC connection
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to model service: %v", err)
	}

	// Create client
	client := pbModel.NewModelServiceClient(conn)

	return &ModelServiceClientImpl{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the connection
func (c *ModelServiceClientImpl) Close() error {
	return c.conn.Close()
}

// HealthCheck performs a health check
func (c *ModelServiceClientImpl) HealthCheck(ctx context.Context) error {
	req := &pbModel.HealthCheckRequest{
		Service: "model-service",
	}

	_, err := c.client.HealthCheck(ctx, req)
	return err
}

// RegisterModel registers a new model
func (c *ModelServiceClientImpl) RegisterModel(ctx context.Context, request *pbModel.RegisterModelRequest) (*pbModel.RegisterModelResponse, error) {
	return c.client.RegisterModel(ctx, request)
}

// UpdateModel updates a model
func (c *ModelServiceClientImpl) UpdateModel(ctx context.Context, request *pbModel.UpdateModelRequest) (*pbModel.UpdateModelResponse, error) {
	return c.client.UpdateModel(ctx, request)
}

// UnregisterModel unregisters a model
func (c *ModelServiceClientImpl) UnregisterModel(ctx context.Context, request *pbModel.UnregisterModelRequest) (*pbModel.UnregisterModelResponse, error) {
	return c.client.UnregisterModel(ctx, request)
}

// ListModels lists all models
func (c *ModelServiceClientImpl) ListModels(ctx context.Context, request *pbModel.ListModelsRequest) (*pbModel.ListModelsResponse, error) {
	return c.client.ListModels(ctx, request)
}

// GetModel gets a model
func (c *ModelServiceClientImpl) GetModel(ctx context.Context, request *pbModel.GetModelRequest) (*pbModel.GetModelResponse, error) {
	return c.client.GetModel(ctx, request)
}

// LoadModel loads a model
func (c *ModelServiceClientImpl) LoadModel(ctx context.Context, request *pbModel.LoadModelRequest) (*pbModel.LoadModelResponse, error) {
	return c.client.LoadModel(ctx, request)
}

// UnloadModel unloads a model
func (c *ModelServiceClientImpl) UnloadModel(ctx context.Context, request *pbModel.UnloadModelRequest) (*pbModel.UnloadModelResponse, error) {
	return c.client.UnloadModel(ctx, request)
}

// IsModelLoaded checks if a model is loaded
func (c *ModelServiceClientImpl) IsModelLoaded(ctx context.Context, request *pbModel.IsModelLoadedRequest) (*pbModel.IsModelLoadedResponse, error) {
	return c.client.IsModelLoaded(ctx, request)
}

// GenerateText generates text using a model
func (c *ModelServiceClientImpl) GenerateText(ctx context.Context, request *pbModel.GenerateTextRequest) (*pbModel.GenerateTextResponse, error) {
	return c.client.GenerateText(ctx, request)
}

// GenerateEmbedding generates embeddings using a model
func (c *ModelServiceClientImpl) GenerateEmbedding(ctx context.Context, request *pbModel.GenerateEmbeddingRequest) (*pbModel.GenerateEmbeddingResponse, error) {
	return c.client.GenerateEmbedding(ctx, request)
}

// ClassifyText classifies text using a model
func (c *ModelServiceClientImpl) ClassifyText(ctx context.Context, request *pbModel.ClassifyTextRequest) (*pbModel.ClassifyTextResponse, error) {
	return c.client.ClassifyText(ctx, request)
}

// SummarizeText summarizes text using a model
func (c *ModelServiceClientImpl) SummarizeText(ctx context.Context, request *pbModel.SummarizeTextRequest) (*pbModel.SummarizeTextResponse, error) {
	return c.client.SummarizeText(ctx, request)
}

// TranslateText translates text using a model
func (c *ModelServiceClientImpl) TranslateText(ctx context.Context, request *pbModel.TranslateTextRequest) (*pbModel.TranslateTextResponse, error) {
	return c.client.TranslateText(ctx, request)
}

// AIServiceClientImpl implements the AIServiceClient interface
type AIServiceClientImpl struct {
	vectorClient *VectorServiceClientImpl
	modelClient  *ModelServiceClientImpl
}

// NewAIServiceClientImpl creates a new AIServiceClientImpl
func NewAIServiceClientImpl(ctx context.Context, vectorAddress, modelAddress string) (*AIServiceClientImpl, error) {
	vectorClient, err := NewVectorServiceClientImpl(ctx, vectorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector service client: %v", err)
	}

	modelClient, err := NewModelServiceClientImpl(ctx, modelAddress)
	if err != nil {
		vectorClient.Close()
		return nil, fmt.Errorf("failed to create model service client: %v", err)
	}

	return &AIServiceClientImpl{
		vectorClient: vectorClient,
		modelClient:  modelClient,
	}, nil
}

// Close closes all connections
func (c *AIServiceClientImpl) Close() error {
	var errors []error

	if err := c.vectorClient.Close(); err != nil {
		errors = append(errors, fmt.Errorf("vector client close error: %v", err))
	}

	if err := c.modelClient.Close(); err != nil {
		errors = append(errors, fmt.Errorf("model client close error: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", errors)
	}

	return nil
}

// GetVectorClient returns the vector service client
func (c *AIServiceClientImpl) GetVectorClient() *VectorServiceClientImpl {
	return c.vectorClient
}

// GetModelClient returns the model service client
func (c *AIServiceClientImpl) GetModelClient() *ModelServiceClientImpl {
	return c.modelClient
}

// HybridAIServiceManager manages AI services in the hybrid architecture
type HybridAIServiceManager struct {
	aiClient      *AIServiceClientImpl
	serviceManager *ServiceManager
}

// NewHybridAIServiceManager creates a new HybridAIServiceManager
func NewHybridAIServiceManager(serviceManager *ServiceManager) *HybridAIServiceManager {
	return &HybridAIServiceManager{
		serviceManager: serviceManager,
	}
}

// Start starts the AI service manager
func (m *HybridAIServiceManager) Start(ctx context.Context) error {
	// Get vector service
	vectorService, err := m.serviceManager.GetService(ctx, "vector-service")
	if err != nil {
		return fmt.Errorf("failed to get vector service: %v", err)
	}

	// Get model service
	modelService, err := m.serviceManager.GetService(ctx, "model-service")
	if err != nil {
		return fmt.Errorf("failed to get model service: %v", err)
	}

	// Create AI service client
	vectorAddress := fmt.Sprintf("%s:%d", vectorService.Address, vectorService.Port)
	modelAddress := fmt.Sprintf("%s:%d", modelService.Address, modelService.Port)

	aiClient, err := NewAIServiceClientImpl(ctx, vectorAddress, modelAddress)
	if err != nil {
		return fmt.Errorf("failed to create AI service client: %v", err)
	}

	m.aiClient = aiClient

	return nil
}

// Stop stops the AI service manager
func (m *HybridAIServiceManager) Stop() error {
	if m.aiClient != nil {
		return m.aiClient.Close()
	}
	return nil
}

// GetAIClient returns the AI service client
func (m *HybridAIServiceManager) GetAIClient() *AIServiceClientImpl {
	return m.aiClient
}

// Reconnect reconnects to the AI services
func (m *HybridAIServiceManager) Reconnect(ctx context.Context) error {
	// Close existing connections
	if m.aiClient != nil {
		if err := m.aiClient.Close(); err != nil {
			return fmt.Errorf("failed to close existing AI client: %v", err)
		}
	}

	// Reconnect
	return m.Start(ctx)
}

// HealthCheck performs health checks on all AI services
func (m *HybridAIServiceManager) HealthCheck(ctx context.Context) error {
	if m.aiClient == nil {
		return fmt.Errorf("AI service client not initialized")
	}

	// Check vector service
	if err := m.aiClient.vectorClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("vector service health check failed: %v", err)
	}

	// Check model service
	if err := m.aiClient.modelClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("model service health check failed: %v", err)
	}

	return nil
}

// ServiceConnectorImpl implements the ServiceConnector interface
type ServiceConnectorImpl struct{}

// NewServiceConnectorImpl creates a new ServiceConnectorImpl
func NewServiceConnectorImpl() *ServiceConnectorImpl {
	return &ServiceConnectorImpl{}
}

// ConnectHTTP connects to an HTTP service
func (c *ServiceConnectorImpl) ConnectHTTP(ctx context.Context, service *ServiceInfo) (*http.Client, error) {
	if service.Protocol != "http" {
		return nil, fmt.Errorf("service %s is not an HTTP service", service.ID)
	}

	// Create a simple HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return client, nil
}

// ConnectGRPC connects to a gRPC service
func (c *ServiceConnectorImpl) ConnectGRPC(ctx context.Context, service *ServiceInfo) (*grpc.ClientConn, error) {
	if service.Protocol != "grpc" {
		return nil, fmt.Errorf("service %s is not a gRPC service", service.ID)
	}

	// Create gRPC connection
	address := fmt.Sprintf("%s:%d", service.Address, service.Port)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC service %s: %v", service.ID, err)
	}

	return conn, nil
}