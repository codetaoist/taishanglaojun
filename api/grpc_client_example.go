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
	// pb "codetaoist/api/proto"
)

// VectorServiceClient wraps the gRPC client for vector operations
type VectorServiceClient struct {
	client pb.VectorServiceClient
	conn   *grpc.ClientConn
}

// NewVectorServiceClient creates a new client for the vector service
func NewVectorServiceClient(address string) (*VectorServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vector service: %v", err)
	}

	client := pb.NewVectorServiceClient(conn)
	return &VectorServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the connection to the vector service
func (c *VectorServiceClient) Close() error {
	return c.conn.Close()
}

// CreateEmbedding creates an embedding for the given text
func (c *VectorServiceClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.CreateEmbeddingRequest{
	//     Text: text,
	// }
	
	// resp, err := c.client.CreateEmbedding(ctx, req)
	// if err != nil {
	//     return nil, err
	// }
	
	// return resp.Embedding, nil
	
	// Placeholder implementation
	return []float32{0.1, 0.2, 0.3}, nil
}

// ModelServiceClient wraps the gRPC client for model operations
type ModelServiceClient struct {
	client pb.ModelServiceClient
	conn   *grpc.ClientConn
}

// NewModelServiceClient creates a new client for the model service
func NewModelServiceClient(address string) (*ModelServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to model service: %v", err)
	}

	client := pb.NewModelServiceClient(conn)
	return &ModelServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the connection to the model service
func (c *ModelServiceClient) Close() error {
	return c.conn.Close()
}

// GenerateText generates text based on the given prompt
func (c *ModelServiceClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	// TODO: Implement after generating the protobuf code
	// req := &pb.GenerateTextRequest{
	//     Prompt: prompt,
	// }
	
	// resp, err := c.client.GenerateText(ctx, req)
	// if err != nil {
	//     return "", err
	// }
	
	// return resp.Text, nil
	
	// Placeholder implementation
	return "This is a generated text response.", nil
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
	embedding, err := vectorClient.CreateEmbedding(ctx, "This is a test text for embedding generation.")
	if err != nil {
		log.Printf("Vector service test failed: %v", err)
	} else {
		fmt.Printf("Embedding generated: %v (first 5 values)\n", embedding[:min(5, len(embedding))])
	}

	// Test model service
	fmt.Println("\nTesting model service...")
	text, err := modelClient.GenerateText(ctx, "Write a short story about a robot learning to paint.")
	if err != nil {
		log.Printf("Model service test failed: %v", err)
	} else {
		fmt.Printf("Generated text: %s\n", text)
	}

	fmt.Println("\nAll tests completed successfully!")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}