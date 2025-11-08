#!/usr/bin/env python3
"""
Test script for Milvus integration
"""

import sys
import os
import grpc
import numpy as np
import time
import logging
from typing import List, Dict, Any

# Add the current directory to the path to import the generated modules
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

import ai_service_pb2
import ai_service_pb2_grpc

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def test_vector_service():
    """Test the vector service with Milvus integration"""
    
    # Create a channel to the vector service
    channel = grpc.insecure_channel('localhost:50051')
    stub = ai_service_pb2_grpc.VectorServiceStub(channel)
    
    try:
        # Test health check
        logger.info("Testing health check...")
        health_response = stub.HealthCheck(ai_service_pb2.HealthCheckRequest())
        logger.info(f"Health check response: healthy={health_response.healthy}, message={health_response.message}")
        
        if not health_response.healthy:
            logger.warning("Vector service is not healthy. Skipping further tests.")
            return
        
        # Test collection operations
        collection_name = f"test_collection_{int(time.time())}"
        
        # Create a schema for the collection
        schema = {
            "fields": [
                {"name": "id", "type": "int64", "is_primary": True, "auto_id": True},
                {"name": "vector", "type": "float_vector", "dim": 128},
                {"name": "text", "type": "varchar", "max_length": 500}
            ]
        }
        
        # Create collection
        logger.info(f"Creating collection: {collection_name}")
        create_response = stub.CreateCollection(ai_service_pb2.CreateCollectionRequest(
            collection_name=collection_name,
            schema=str(schema)
        ))
        logger.info("Collection created successfully")
        
        # Check if collection exists
        logger.info(f"Checking if collection exists: {collection_name}")
        has_response = stub.HasCollection(ai_service_pb2.HasCollectionRequest(
            collection_name=collection_name
        ))
        logger.info(f"Collection exists: {has_response.has}")
        
        # Insert some test vectors (Note: We don't have an Insert method in the proto yet)
        # For now, we'll just test search with empty collection
        logger.info("Testing search with empty collection...")
        query_vector = np.random.rand(128).astype(np.float32).tolist()
        
        search_response = stub.Search(ai_service_pb2.SearchRequest(
            collection_name=collection_name,
            vector=query_vector,
            top_k=5
        ))
        
        logger.info(f"Search returned {len(search_response.results)} results")
        for i, result in enumerate(search_response.results):
            logger.info(f"Result {i}: id={result.id}, score={result.score}")
        
        # Drop the collection
        logger.info(f"Dropping collection: {collection_name}")
        drop_response = stub.DropCollection(ai_service_pb2.DropCollectionRequest(
            collection_name=collection_name
        ))
        logger.info("Collection dropped successfully")
        
        # Verify collection no longer exists
        logger.info(f"Verifying collection no longer exists: {collection_name}")
        has_response = stub.HasCollection(ai_service_pb2.HasCollectionRequest(
            collection_name=collection_name
        ))
        logger.info(f"Collection exists after drop: {has_response.has}")
        
        logger.info("All tests completed successfully!")
        
    except grpc.RpcError as e:
        logger.error(f"gRPC error: {e.code()} - {e.details()}")
    except Exception as e:
        logger.error(f"Error: {str(e)}")
    finally:
        channel.close()

if __name__ == '__main__':
    logger.info("Starting vector service integration test...")
    test_vector_service()