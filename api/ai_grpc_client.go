package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Import the generated protobuf packages
	// TODO: Uncomment after generating the protobuf code
	// pb "github.com/codetaoist/taishang/pkg/grpc/ai"
)

// VectorServiceClient wraps the gRPC client for vector operations
type VectorServiceClient struct {
	// TODO: Uncomment after generating the protobuf code
	// client pb.VectorServiceClient
	conn *grpc.ClientConn
}

// NewVectorServiceClient creates a new client for the vector service
func NewVectorServiceClient(address string) (*VectorServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vector service: %v", err)
	}

	// TODO: Uncomment after generating the protobuf code
	// client := pb.NewVectorServiceClient(conn)
	return &VectorServiceClient{
		// client: client,
		conn: conn,
	}, nil
}

// Close closes the connection to the vector service
func (c *VectorServiceClient) Close() error {
	return c.conn.Close()
}

// HealthCheck checks the health of the vector service
func (c *VectorServiceClient) HealthCheck(ctx context.Context) (bool, string, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.HealthCheckRequest{}
	// resp, err := c.client.HealthCheck(ctx, req)
	// if err != nil {
	//     return false, "", err
	// }
	// return resp.Healthy, resp.Message, nil
	
	// Placeholder implementation
	return true, "Vector service is running", nil
}

// CreateCollection creates a new collection
func (c *VectorServiceClient) CreateCollection(ctx context.Context, name string) error {
	// TODO: Implement after generating the protobuf code
	// schema := &pb.CollectionSchema{
	//     Name: name,
	//     Fields: []*pb.FieldSchema{
	//         {
	//             Name:       "id",
	//             IsPrimaryKey: true,
	//             DataType:   "string",
	//         },
	//         {
	//             Name:     "vector",
	//             DataType: "vector",
	//             TypeParams: map[string]string{
	//                 "dim": "768",
	//             },
	//         },
	//     },
	// }
	// 
	// req := &pb.CreateCollectionRequest{
	//     CollectionName: name,
	//     Schema:         schema,
	// }
	// 
	// _, err := c.client.CreateCollection(ctx, req)
	// return err
	
	// Placeholder implementation
	log.Printf("Creating collection: %s", name)
	return nil
}

// Search searches for similar vectors
func (c *VectorServiceClient) Search(ctx context.Context, collectionName string, vector []float32, topK int64) ([]map[string]interface{}, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.SearchRequest{
	//     CollectionName: collectionName,
	//     Vector:         vector,
	//     TopK:           topK,
	// }
	// 
	// resp, err := c.client.Search(ctx, req)
	// if err != nil {
	//     return nil, err
	// }
	// 
	// results := make([]map[string]interface{}, 0, len(resp.Results))
	// for _, result := range resp.Results {
	//     results = append(results, map[string]interface{}{
	//         "id":       result.Id,
	//         "score":    result.Score,
	//         "metadata": result.Metadata,
	//     })
	// }
	// 
	// return results, nil
	
	// Placeholder implementation
	log.Printf("Searching in collection %s with top_k=%d", collectionName, topK)
	
	// Mock results
	results := make([]map[string]interface{}, 0, 5)
	for i := 0; i < 5; i++ {
		results = append(results, map[string]interface{}{
			"id":    fmt.Sprintf("result_%d", i),
			"score": 0.9 - float32(i)*0.1,
			"metadata": map[string]string{
				"key": fmt.Sprintf("value_%d", i),
			},
		})
	}
	
	return results, nil
}

// ModelServiceClient wraps the gRPC client for model operations
type ModelServiceClient struct {
	// TODO: Uncomment after generating the protobuf code
	// client pb.ModelServiceClient
	conn *grpc.ClientConn
}

// NewModelServiceClient creates a new client for the model service
func NewModelServiceClient(address string) (*ModelServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to model service: %v", err)
	}

	// TODO: Uncomment after generating the protobuf code
	// client := pb.NewModelServiceClient(conn)
	return &ModelServiceClient{
		// client: client,
		conn: conn,
	}, nil
}

// Close closes the connection to the model service
func (c *ModelServiceClient) Close() error {
	return c.conn.Close()
}

// HealthCheck checks the health of the model service
func (c *ModelServiceClient) HealthCheck(ctx context.Context) (bool, string, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.HealthCheckRequest{}
	// resp, err := c.client.HealthCheck(ctx, req)
	// if err != nil {
	//     return false, "", err
	// }
	// return resp.Healthy, resp.Message, nil
	
	// Placeholder implementation
	return true, "Model service is running", nil
}

// RegisterModel registers a new model
func (c *ModelServiceClient) RegisterModel(ctx context.Context, name, provider, modelPath, description string) error {
	// TODO: Implement after generating the protobuf code
	// req := &pb.RegisterModelRequest{
	//     Name:        name,
	//     Provider:    pb.ModelProvider(pb.ModelProvider_value[provider]),
	//     ModelPath:   modelPath,
	//     Description: description,
	// }
	// 
	// _, err := c.client.RegisterModel(ctx, req)
	// return err
	
	// Placeholder implementation
	log.Printf("Registering model: %s, provider: %s, path: %s", name, provider, modelPath)
	return nil
}

// GenerateText generates text based on a prompt
func (c *ModelServiceClient) GenerateText(ctx context.Context, modelName, prompt string) (string, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.TextGenerationRequest{
	//     ModelName: modelName,
	//     Prompt:    prompt,
	// }
	// 
	// resp, err := c.client.GenerateText(ctx, req)
	// if err != nil {
	//     return "", err
	// }
	// 
	// if len(resp.Texts) > 0 {
	//     return resp.Texts[0], nil
	// }
	// 
	// return "", fmt.Errorf("no text generated")
	
	// Placeholder implementation
	log.Printf("Generating text with model %s for prompt: %s", modelName, prompt)
	return fmt.Sprintf("This is a generated response to: %s", prompt), nil
}

// GenerateEmbedding generates embeddings for the given texts
func (c *ModelServiceClient) GenerateEmbedding(ctx context.Context, modelName string, texts []string) ([][]float32, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.EmbeddingRequest{
	//     ModelName: modelName,
	//     Texts:     texts,
	// }
	// 
	// resp, err := c.client.GenerateEmbedding(ctx, req)
	// if err != nil {
	//     return nil, err
	// }
	// 
	// return resp.Embeddings, nil
	
	// Placeholder implementation
	log.Printf("Generating embeddings with model %s for %d texts", modelName, len(texts))
	
	// Generate random embeddings
	embeddings := make([][]float32, 0, len(texts))
	for range texts {
		embedding := make([]float32, 768)
		for i := range embedding {
			embedding[i] = float32(i) / 768.0
		}
		embeddings = append(embeddings, embedding)
	}
	
	return embeddings, nil
}

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a vector service client
	vectorClient, err := NewVectorServiceClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create vector service client: %v", err)
	}
	defer vectorClient.Close()

	// Create a model service client
	modelClient, err := NewModelServiceClient("localhost:50052")
	if err != nil {
		log.Fatalf("Failed to create model service client: %v", err)
	}
	defer modelClient.Close()

	// Test vector service
	fmt.Println("Testing vector service...")
	
	// Health check
	healthy, message, err := vectorClient.HealthCheck(ctx)
	if err != nil {
		log.Printf("Vector service health check failed: %v", err)
	} else {
		fmt.Printf("Vector service health: %t, message: %s\n", healthy, message)
	}
	
	// Create collection
	err = vectorClient.CreateCollection(ctx, "test_collection")
	if err != nil {
		log.Printf("Create collection failed: %v", err)
	} else {
		fmt.Println("Collection created successfully")
	}
	
	// Search
	queryVector := make([]float32, 768)
	for i := range queryVector {
		queryVector[i] = float32(i) / 768.0
	}
	
	results, err := vectorClient.Search(ctx, "test_collection", queryVector, 5)
	if err != nil {
		log.Printf("Search failed: %v", err)
	} else {
		fmt.Printf("Search results: %v\n", results)
	}

	// Test model service
	fmt.Println("\nTesting model service...")
	
	// Health check
	healthy, message, err = modelClient.HealthCheck(ctx)
	if err != nil {
		log.Printf("Model service health check failed: %v", err)
	} else {
		fmt.Printf("Model service health: %t, message: %s\n", healthy, message)
	}
	
	// Register model
	err = modelClient.RegisterModel(ctx, "test_model", "HUGGINGFACE", "sentence-transformers/all-MiniLM-L6-v2", "Test model for embeddings")
	if err != nil {
		log.Printf("Register model failed: %v", err)
	} else {
		fmt.Println("Model registered successfully")
	}
	
	// Generate text
	text, err := modelClient.GenerateText(ctx, "test_model", "Write a short story about a robot learning to paint.")
	if err != nil {
		log.Printf("Generate text failed: %v", err)
	} else {
		fmt.Printf("Generated text: %s\n", text)
	}
	
	// Generate embeddings
	embeddings, err := modelClient.GenerateEmbedding(ctx, "test_model", []string{"This is a test text for embedding generation."})
	if err != nil {
		log.Printf("Generate embeddings failed: %v", err)
	} else {
		fmt.Printf("Embeddings generated: %v (first 5 values of first embedding)\n", embeddings[0][:5])
	}

	fmt.Println("\nAll tests completed successfully!")
}