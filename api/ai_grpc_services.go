package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"sort"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/codetaoist/taishanglaojun/api/proto"
)

// VectorServiceServer implements the VectorService interface
type VectorServiceServer struct {
	pb.UnimplementedVectorServiceServer
	
	// In-memory storage for collections (for demonstration purposes)
	collections map[string]*Collection
	mu          sync.RWMutex
}

// Collection represents a collection in the vector database
type Collection struct {
	Name   string
	Fields []Field
	Items  []Item
}

// Field represents a field in a collection
type Field struct {
	Name         string
	IsPrimaryKey bool
	DataType     string
	TypeParams   map[string]string
}

// Item represents an item in a collection
type Item struct {
	ID       string
	Vector   []float32
	Metadata map[string]string
}

// ModelServiceServer implements the ModelService interface
type ModelServiceServer struct {
	pb.UnimplementedModelServiceServer
	
	// In-memory storage for models (for demonstration purposes)
	models map[string]*Model
	mu     sync.RWMutex
}

// Model represents a registered model
type Model struct {
	Name        string
	Provider    pb.ModelProvider
	ModelPath   string
	Description string
}

// NewVectorServiceServer creates a new VectorServiceServer
func NewVectorServiceServer() *VectorServiceServer {
	return &VectorServiceServer{
		collections: make(map[string]*Collection),
	}
}

// NewModelServiceServer creates a new ModelServiceServer
func NewModelServiceServer() *ModelServiceServer {
	return &ModelServiceServer{
		models: make(map[string]*Model),
	}
}

// HealthCheck checks the health of the VectorService
func (s *VectorServiceServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Success: true,
		Message: "Vector service is running",
	}, nil
}

// CreateCollection creates a new collection
func (s *VectorServiceServer) CreateCollection(ctx context.Context, req *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	collectionName := req.GetCollectionName()
	schema := req.GetSchema()
	
	// Convert proto fields to internal representation
	fields := make([]Field, len(schema.GetFields()))
	for i, field := range schema.GetFields() {
		fields[i] = Field{
			Name:         field.GetName(),
			IsPrimaryKey: field.GetIsPrimaryKey(),
			DataType:     field.GetDataType(),
			TypeParams:   field.GetTypeParams(),
		}
	}
	
	// Create the collection
	collection := &Collection{
		Name:   collectionName,
		Fields: fields,
		Items:  []Item{},
	}
	
	s.collections[collectionName] = collection
	
	log.Printf("Created collection: %s", collectionName)
	
	return &pb.CreateCollectionResponse{
		Success: true,
		Message: fmt.Sprintf("Collection %s created successfully", collectionName),
	}, nil
}

// Search searches for similar vectors in a collection
func (s *VectorServiceServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	collectionName := req.GetCollectionName()
	queryVector := req.GetVector()
	topK := int(req.GetTopK())
	
	collection, exists := s.collections[collectionName]
	if !exists {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}
	
	// Simple cosine similarity calculation (for demonstration purposes)
	results := []*pb.SearchResult{}
	for _, item := range collection.Items {
		similarity := cosineSimilarity(queryVector, item.Vector)
		results = append(results, &pb.SearchResult{
			Id:       item.ID,
			Score:    similarity,
			Metadata: item.Metadata,
		})
	}
	
	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// Return topK results
	if len(results) > topK {
		results = results[:topK]
	}
	
	log.Printf("Search in collection %s returned %d results", collectionName, len(results))
	
	return &pb.SearchResponse{
		Results: results,
	}, nil
}

// HealthCheck checks the health of the ModelService
func (s *ModelServiceServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Success: true,
		Message: "Model service is running",
	}, nil
}

// RegisterModel registers a new model
func (s *ModelServiceServer) RegisterModel(ctx context.Context, req *pb.RegisterModelRequest) (*pb.RegisterModelResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	modelName := req.GetName()
	provider := req.GetProvider()
	modelPath := req.GetModelPath()
	description := req.GetDescription()
	
	// Create the model
	model := &Model{
		Name:        modelName,
		Provider:    provider,
		ModelPath:   modelPath,
		Description: description,
	}
	
	s.models[modelName] = model
	
	log.Printf("Registered model: %s", modelName)
	
	return &pb.RegisterModelResponse{
		Success: true,
		Message: fmt.Sprintf("Model %s registered successfully", modelName),
	}, nil
}



// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// main function to start the gRPC server
func main() {
	// Create a listener on TCP port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	// Create a new gRPC server
	s := grpc.NewServer()
	
	// Create service instances
	vectorServiceServer := NewVectorServiceServer()
	modelServiceServer := NewModelServiceServer()
	
	// Register services with the gRPC server
	pb.RegisterVectorServiceServer(s, vectorServiceServer)
	pb.RegisterModelServiceServer(s, modelServiceServer)
	
	// Register the health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	
	// Set the serving status for all services to SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("vector.VectorService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("model.ModelService", grpc_health_v1.HealthCheckResponse_SERVING)
	
	log.Println("gRPC server is running on port 50051...")
	
	// Start the server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}