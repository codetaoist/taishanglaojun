#!/bin/bash

# Install protobuf and plugins
# This script installs protobuf and the necessary Go plugins

# Install protobuf using Homebrew
echo "Installing protobuf..."
brew install protobuf

# Install Go protobuf plugins
echo "Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# Add the Go bin directory to PATH if not already added
export PATH="$PATH:$(go env GOPATH)/bin"

# Generate the Go code for the protobuf
echo "Generating Go gRPC client code..."
protoc --go_out=. --go-grpc_out=. proto/ai_service.proto

echo "Go gRPC client code generated successfully!"