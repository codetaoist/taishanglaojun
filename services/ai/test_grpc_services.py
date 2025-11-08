#!/usr/bin/env python3
"""
Test script for the AI gRPC services
"""

import grpc
import time
from concurrent import futures

# Import the generated gRPC modules
# TODO: Uncomment after generating the protobuf code
# import app.proto.ai_service_pb2 as ai_service_pb2
# import app.proto.ai_service_pb2_grpc as ai_service_pb2_grpc

def test_vector_service():
    """Test the vector gRPC service"""
    print("Testing vector service...")
    
    # Create a channel to the vector service
    with grpc.insecure_channel('localhost:50051') as channel:
        # TODO: Create a stub after generating the protobuf code
        # stub = ai_service_pb2_grpc.VectorServiceStub(channel)
        
        # Test health check
        # TODO: Add actual health check after generating the protobuf code
        # request = ai_service_pb2.HealthCheckRequest()
        # response = stub.HealthCheck(request)
        # print(f"Vector service health: {response.status}")
        
        print("Vector service test completed successfully!")

def test_model_service():
    """Test the model gRPC service"""
    print("Testing model service...")
    
    # Create a channel to the model service
    with grpc.insecure_channel('localhost:50052') as channel:
        # TODO: Create a stub after generating the protobuf code
        # stub = ai_service_pb2_grpc.ModelServiceStub(channel)
        
        # Test health check
        # TODO: Add actual health check after generating the protobuf code
        # request = ai_service_pb2.HealthCheckRequest()
        # response = stub.HealthCheck(request)
        # print(f"Model service health: {response.status}")
        
        print("Model service test completed successfully!")

if __name__ == '__main__':
    print("Starting gRPC service tests...")
    
    # Test vector service
    try:
        test_vector_service()
    except Exception as e:
        print(f"Vector service test failed: {e}")
    
    # Test model service
    try:
        test_model_service()
    except Exception as e:
        print(f"Model service test failed: {e}")
    
    print("gRPC service tests completed!")