#!/bin/bash

# Generate protobuf code for Go
echo "Generating protobuf code..."

# Create output directory if it doesn't exist
mkdir -p ../gen/go

# Generate vector service
protoc --go_out=../gen/go --go-grpc_out=../gen/go \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  vector/vector.proto

# Generate model service
protoc --go_out=../gen/go --go-grpc_out=../gen/go \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  model/model.proto

echo "Protobuf code generation completed!"