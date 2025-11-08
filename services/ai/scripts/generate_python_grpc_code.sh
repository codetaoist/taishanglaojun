#!/bin/bash

# Python gRPC code generation script

# Check if protobuf is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install protobuf first."
    exit 1
fi

# Check if Python grpcio-tools is installed
if ! python3 -c "import grpc_tools" &> /dev/null; then
    echo "Error: Python grpcio-tools is not installed. Please install it first."
    exit 1
fi

# Set paths
PROTO_DIR="/Users/lida/Documents/work/codetaoist/api/proto"
PYTHON_OUT_DIR="/Users/lida/Documents/work/codetaoist/services/ai"

# Generate Python gRPC code
echo "Generating Python gRPC code..."
python3 -m grpc_tools.protoc \
    --proto_path=${PROTO_DIR} \
    --python_out=${PYTHON_OUT_DIR} \
    --grpc_python_out=${PYTHON_OUT_DIR} \
    ${PROTO_DIR}/ai_service.proto

# Check if generation was successful
if [ $? -eq 0 ]; then
    echo "Python gRPC code generated successfully!"
    echo "Generated files:"
    ls -la ${PYTHON_OUT_DIR}/ai_service_pb2.py
    ls -la ${PYTHON_OUT_DIR}/ai_service_pb2_grpc.py
else
    echo "Error: Failed to generate Python gRPC code."
    exit 1
fi

# Create __init__.py files if they don't exist
if [ ! -f ${PYTHON_OUT_DIR}/__init__.py ]; then
    touch ${PYTHON_OUT_DIR}/__init__.py
    echo "Created __init__.py in ${PYTHON_OUT_DIR}"
fi

echo "Python gRPC code generation completed."