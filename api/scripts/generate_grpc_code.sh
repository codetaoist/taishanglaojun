#!/bin/bash

# Script to generate gRPC code for Go and Python after protobuf installation

set -e

echo "Starting gRPC code generation..."

# Check if protobuf is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install protobuf first."
    exit 1
fi

# Check if Go protobuf plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
fi

# Add GOPATH to PATH if not already present
if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
    export PATH="$PATH:$GOPATH/bin"
fi

# Create directories for generated code if they don't exist
mkdir -p /Users/lida/Documents/work/codetaoist/api/proto
mkdir -p /Users/lida/Documents/work/codetaoist/services/ai/proto

# Generate Go gRPC code
echo "Generating Go gRPC code..."
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --proto_path=/Users/lida/Documents/work/codetaoist/api/proto \
    /Users/lida/Documents/work/codetaoist/api/proto/ai_service.proto

# Generate Python gRPC code
echo "Generating Python gRPC code..."
python3 -m grpc_tools.protoc \
    --python_out=. \
    --grpc_python_out=. \
    --proto_path=/Users/lida/Documents/work/codetaoist/api/proto \
    /Users/lida/Documents/work/codetaoist/api/proto/ai_service.proto

# Copy generated Python code to the services/ai directory if needed
if [ -f "ai_service_pb2.py" ]; then
    mkdir -p /Users/lida/Documents/work/codetaoist/services/ai/proto
    cp ai_service_pb2.py /Users/lida/Documents/work/codetaoist/services/ai/proto/
    cp ai_service_pb2_grpc.py /Users/lida/Documents/work/codetaoist/services/ai/proto/
    echo "Python gRPC code copied to services/ai/proto/"
fi

echo "gRPC code generation completed successfully!"

# Update the Go and Python files to uncomment the protobuf code
echo "Updating Go files to use generated protobuf code..."

# Update ai_grpc_services.go
sed -i '' 's|// pb "codetaoist/api/proto"|pb "codetaoist/api/proto"|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go
sed -i '' 's|// pb.UnimplementedVectorServiceServer|pb.UnimplementedVectorServiceServer|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go
sed -i '' 's|// pb.UnimplementedModelServiceServer|pb.UnimplementedModelServiceServer|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go

# Update ai_grpc_client.go
sed -i '' 's|// pb "codetaoist/api/proto"|pb "codetaoist/api/proto"|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_client.go

# Register the services
sed -i '' 's|// pb.RegisterVectorServiceServer(s, NewVectorServiceServer())|pb.RegisterVectorServiceServer(s, NewVectorServiceServer())|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go
sed -i '' 's|// pb.RegisterModelServiceServer(s, NewModelServiceServer())|pb.RegisterModelServiceServer(s, NewModelServiceServer())|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go

# Create clients
sed -i '' 's|// vectorClient := pb.NewVectorServiceClient(vectorConn)|vectorClient := pb.NewVectorServiceClient(vectorConn)|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go
sed -i '' 's|// modelClient := pb.NewModelServiceClient(modelConn)|modelClient := pb.NewModelServiceClient(modelConn)|g' /Users/lida/Documents/work/codetaoist/api/ai_grpc_services.go

# Update Python files to use generated protobuf code
echo "Updating Python files to use generated protobuf code..."

# Update test_grpc_client.py
sed -i '' 's|# import app.proto.ai_service_pb2 as ai_service_pb2|import sys\nsys.path.append("..\/..\/api")\nimport proto.ai_service_pb2 as ai_service_pb2|g' /Users/lida/Documents/work/codetaoist/services/ai/test_grpc_client.py
sed -i '' 's|# import app.proto.ai_service_pb2_grpc as ai_service_pb2_grpc|import proto.ai_service_pb2_grpc as ai_service_pb2_grpc|g' /Users/lida/Documents/work/codetaoist/services/ai/test_grpc_client.py

# Uncomment the protobuf code in the Python files
sed -i '' 's|# TODO: Uncomment after generating the protobuf code|# TODO: Uncomment after generating the protobuf code|g' /Users/lida/Documents/work/codetaoist/services/ai/test_grpc_client.py
sed -i '' '/# self.channel = grpc.insecure_channel(address)/s/# //' /Users/lida/Documents/work/codetaoist/services/ai/test_grpc_client.py
sed -i '' '/# self.stub = ai_service_pb2_grpc.VectorServiceStub(self.channel)/s/# //' /Users/lida/Documents/work/codetaoist/services/ai/test_grpc_client.py
sed -i '' '/# self.channel.close()/s/# //' /Users/lida/Documents/work/codetaoist/services/ai/test_grpc_client.py

# Update ai_grpc_services.py
sed -i '' 's|# TODO: Import the generated gRPC modules|import sys\nsys.path.append("..\/..\/api")\nfrom proto import ai_service_pb2\nfrom proto import ai_service_pb2_grpc|g' /Users/lida/Documents/work/codetaoist/services/ai/ai_grpc_services.py

echo "Files updated to use generated protobuf code!"

# Create a simple test script to verify the gRPC communication
echo "Creating test script..."

cat > /Users/lida/Documents/work/codetaoist/test_grpc_communication.sh << 'EOF'
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
EOF

chmod +x /Users/lida/Documents/work/codetaoist/test_grpc_communication.sh

echo "All setup completed! You can now test the gRPC communication by running:"
echo "/Users/lida/Documents/work/codetaoist/test_grpc_communication.sh"