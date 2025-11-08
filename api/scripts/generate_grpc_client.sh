#!/bin/bash

# Generate gRPC client code for Go
# This script generates the Go code for the gRPC client based on the protobuf definition

# Set the working directory
cd "$(dirname "$0")"

# Create the proto directory if it doesn't exist
mkdir -p proto

# Generate the Go code for the protobuf
protoc --go_out=. --go-grpc_out=. proto/ai_service.proto

echo "Go gRPC client code generated successfully!"