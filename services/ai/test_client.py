#!/usr/bin/env python3
"""
Simple gRPC client to test the AI services
"""

import grpc
import sys
import time
import numpy as np

# Import the generated gRPC modules
import ai_service_pb2
import ai_service_pb2_grpc

def test_vector_service():
    """Test the vector service"""
    print("Testing Vector Service...")
    
    # Create a channel to connect to the server
    channel = grpc.insecure_channel('localhost:50051')
    stub = ai_service_pb2_grpc.VectorServiceStub(channel)
    
    try:
        # Test HealthCheck
        print("\n1. Testing HealthCheck...")
        response = stub.HealthCheck(ai_service_pb2.HealthCheckRequest())
        print(f"Health check response: healthy={response.healthy}, message='{response.message}'")
        
        # Test CreateCollection
        print("\n2. Testing CreateCollection...")
        schema = ai_service_pb2.CollectionSchema()
        field = ai_service_pb2.FieldSchema()
        field.name = "id"
        field.is_primary_key = True
        field.data_type = "INT64"
        schema.fields.append(field)
        
        field = ai_service_pb2.FieldSchema()
        field.name = "vector"
        field.is_primary_key = False
        field.data_type = "FLOAT_VECTOR"
        field.type_params["dim"] = "128"
        schema.fields.append(field)
        
        create_request = ai_service_pb2.CreateCollectionRequest()
        create_request.collection_name = "test_collection"
        create_request.schema = schema
        
        response = stub.CreateCollection(create_request)
        print(f"Create collection response: success={response.success}, message='{response.message}'")
        
        # Test HasCollection
        print("\n3. Testing HasCollection...")
        has_request = ai_service_pb2.HasCollectionRequest()
        has_request.collection_name = "test_collection"
        
        response = stub.HasCollection(has_request)
        print(f"Has collection response: has={response.has}")
        
        # Test Search
        print("\n4. Testing Search...")
        search_request = ai_service_pb2.SearchRequest()
        search_request.collection_name = "test_collection"
        search_request.vector.extend(np.random.rand(128).astype(np.float32))
        search_request.top_k = 5
        
        response = stub.Search(search_request)
        print(f"Search response: {len(response.results)} results")
        for i, result in enumerate(response.results):
            print(f"  Result {i+1}: id={result.id}, score={result.score}")
        
        print("\nVector Service tests completed successfully!")
        
    except grpc.RpcError as e:
        print(f"Error testing Vector Service: {e.code()}: {e.details()}")
    finally:
        channel.close()

def test_model_service():
    """Test the model service"""
    print("\nTesting Model Service...")
    
    # Create a channel to connect to the server
    channel = grpc.insecure_channel('localhost:50052')
    stub = ai_service_pb2_grpc.ModelServiceStub(channel)
    
    try:
        # Test HealthCheck
        print("\n1. Testing HealthCheck...")
        response = stub.HealthCheck(ai_service_pb2.HealthCheckRequest())
        print(f"Health check response: healthy={response.healthy}, message='{response.message}'")
        
        # Test RegisterModel
        print("\n2. Testing RegisterModel...")
        register_request = ai_service_pb2.RegisterModelRequest()
        register_request.name = "test_model"
        register_request.provider = ai_service_pb2.ModelProvider.HUGGINGFACE
        register_request.model_path = "sentence-transformers/all-MiniLM-L6-v2"
        register_request.description = "A test model for embeddings"
        
        response = stub.RegisterModel(register_request)
        print(f"Register model response: success={response.success}, message='{response.message}'")
        
        # Test GenerateText
        print("\n3. Testing GenerateText...")
        text_request = ai_service_pb2.TextGenerationRequest()
        text_request.model_name = "test_model"
        text_request.prompt = "What is the meaning of life?"
        text_request.max_length = 100
        text_request.temperature = 0.7
        
        response = stub.GenerateText(text_request)
        print(f"Generate text response: success={response.success}, message='{response.message}'")
        print(f"Generated text: {response.texts[0] if response.texts else 'No text generated'}")
        
        # Test GenerateEmbedding
        print("\n4. Testing GenerateEmbedding...")
        embedding_request = ai_service_pb2.EmbeddingRequest()
        embedding_request.model_name = "test_model"
        embedding_request.texts.append("Hello, world!")
        embedding_request.texts.append("How are you?")
        
        response = stub.GenerateEmbedding(embedding_request)
        print(f"Generate embedding response: success={response.success}, message='{response.message}'")
        print(f"Generated {len(response.embeddings)} embeddings")
        for i, embedding in enumerate(response.embeddings):
            print(f"  Embedding {i+1}: dimension={len(embedding.values)}")
        
        print("\nModel Service tests completed successfully!")
        
    except grpc.RpcError as e:
        print(f"Error testing Model Service: {e.code()}: {e.details()}")
    finally:
        channel.close()

if __name__ == '__main__':
    if len(sys.argv) > 1:
        service_type = sys.argv[1]
        if service_type == "vector":
            test_vector_service()
        elif service_type == "model":
            test_model_service()
        elif service_type == "all":
            test_vector_service()
            test_model_service()
        else:
            print(f"Unknown service type: {service_type}")
            print("Usage: python test_client.py [vector|model|all]")
    else:
        print("Please specify a service type: vector, model, or all")
        print("Usage: python test_client.py [vector|model|all]")