#!/bin/bash

# Test script for gRPC communication between Go and Python

echo "Testing gRPC communication between Go and Python services..."

# Start the Go services in the background
echo "Starting Go gRPC services..."
cd /Users/lida/Documents/work/codetaoist/api
go run ai_grpc_services.go &
GO_PID=$!

# Wait for the services to start
sleep 3

# Test the Python client
echo "Testing Python gRPC client..."
cd /Users/lida/Documents/work/codetaoist/services/ai
python3 test_grpc_client.py

# Stop the Go services
echo "Stopping Go gRPC services..."
kill $GO_PID

echo "gRPC communication test completed!"
