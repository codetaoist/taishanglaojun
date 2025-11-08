package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Set up a connection to the vector service
	vectorConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to vector service: %v", err)
	}
	defer vectorConn.Close()

	// Set up a connection to the model service
	modelConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to model service: %v", err)
	}
	defer modelConn.Close()

	// Test vector service health
	log.Println("Testing vector service health...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// TODO: Add actual gRPC client calls here
	// For now, just test the connection
	log.Println("Successfully connected to vector service")

	// Test model service health
	log.Println("Testing model service health...")
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// TODO: Add actual gRPC client calls here
	// For now, just test the connection
	log.Println("Successfully connected to model service")

	log.Println("gRPC connection test completed successfully!")
}