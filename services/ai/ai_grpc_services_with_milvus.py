#!/usr/bin/env python3
"""
Implementation of the AI gRPC services based on the existing proto file
"""

import grpc
import numpy as np
import logging
import json
from concurrent import futures
import time
import os
from typing import Dict, List, Any, Optional

# Import the generated gRPC modules
import ai_service_pb2 as ai_service_pb2
import ai_service_pb2_grpc as ai_service_pb2_grpc

# Import Milvus client
from milvus_client import MilvusClient

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Mock data for demonstration
models = {}
loaded_models = {}

class VectorServiceImpl(ai_service_pb2_grpc.VectorServiceServicer):
    """Implementation of the VectorService using Milvus"""
    
    def __init__(self, milvus_host='localhost', milvus_port='19530'):
        """Initialize the VectorService with Milvus client"""
        self.milvus_client = MilvusClient(host=milvus_host, port=milvus_port)
        self.connected = self.milvus_client.connect()
        if not self.connected:
            logger.error("Failed to connect to Milvus")
        else:
            logger.info("Connected to Milvus successfully")
    
    def HealthCheck(self, request, context):
        """Check the health of the vector service"""
        logger.info("Health check received for vector service")
        
        # Check if Milvus is connected
        if not self.connected:
            return ai_service_pb2.HealthCheckResponse(
                healthy=False,
                message="Vector service is not connected to Milvus"
            )
        
        return ai_service_pb2.HealthCheckResponse(
            healthy=True,
            message="Vector service is running with Milvus backend"
        )
    
    def CreateCollection(self, request, context):
        """Create a new collection"""
        collection_name = request.collection_name
        schema_json = request.schema
        
        try:
            # Parse schema from JSON
            schema = json.loads(schema_json) if isinstance(schema_json, str) else schema_json
            
            # Create collection using Milvus client
            success = self.milvus_client.create_collection(collection_name, schema)
            
            if success:
                logger.info(f"Created collection: {collection_name}")
                return ai_service_pb2.CreateCollectionResponse()
            else:
                logger.error(f"Failed to create collection: {collection_name}")
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details(f"Failed to create collection: {collection_name}")
                return ai_service_pb2.CreateCollectionResponse()
                
        except Exception as e:
            logger.error(f"Error creating collection {collection_name}: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error creating collection: {str(e)}")
            return ai_service_pb2.CreateCollectionResponse()
    
    def DropCollection(self, request, context):
        """Drop a collection"""
        collection_name = request.collection_name
        
        try:
            # Drop collection using Milvus client
            success = self.milvus_client.drop_collection(collection_name)
            
            if success:
                logger.info(f"Dropped collection: {collection_name}")
            else:
                logger.warning(f"Failed to drop collection or collection doesn't exist: {collection_name}")
            
            return ai_service_pb2.DropCollectionResponse()
            
        except Exception as e:
            logger.error(f"Error dropping collection {collection_name}: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error dropping collection: {str(e)}")
            return ai_service_pb2.DropCollectionResponse()
    
    def HasCollection(self, request, context):
        """Check if a collection exists"""
        collection_name = request.collection_name
        
        try:
            # Check if collection exists using Milvus client
            has = self.milvus_client.has_collection(collection_name)
            
            logger.info(f"Checked collection {collection_name}: {has}")
            return ai_service_pb2.HasCollectionResponse(has=has)
            
        except Exception as e:
            logger.error(f"Error checking collection {collection_name}: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error checking collection: {str(e)}")
            return ai_service_pb2.HasCollectionResponse(has=False)
    
    def Search(self, request, context):
        """Search for similar vectors"""
        collection_name = request.collection_name
        query_vector = list(request.vector)
        top_k = request.top_k
        
        try:
            # Search vectors using Milvus client
            search_results = self.milvus_client.search_vectors(
                collection_name=collection_name,
                query_vectors=[query_vector],
                top_k=top_k
            )
            
            # Convert search results to gRPC response format
            results = []
            for result in search_results:
                result_item = ai_service_pb2.SearchResult(
                    id=result["id"],
                    score=result["score"],
                    metadata=result["metadata"]
                )
                results.append(result_item)
            
            logger.info(f"Search in collection {collection_name} returned {len(results)} results")
            return ai_service_pb2.SearchResponse(results=results)
            
        except Exception as e:
            logger.error(f"Error searching in collection {collection_name}: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error searching: {str(e)}")
            return ai_service_pb2.SearchResponse(results=[])

class ModelServiceImpl(ai_service_pb2_grpc.ModelServiceServicer):
    """Implementation of the ModelService"""
    
    def HealthCheck(self, request, context):
        """Check the health of the model service"""
        logger.info("Health check received for model service")
        return ai_service_pb2.HealthCheckResponse(
            healthy=True,
            message="Model service is running"
        )
    
    def RegisterModel(self, request, context):
        """Register a new model"""
        name = request.name
        provider = request.provider
        model_path = request.model_path
        config = dict(request.config)
        description = request.description
        
        models[name] = {
            "name": name,
            "provider": provider,
            "model_path": model_path,
            "config": config,
            "description": description,
            "is_loaded": False
        }
        
        logger.info(f"Registered model: {name}")
        return ai_service_pb2.RegisterModelResponse(
            success=True,
            message=f"Model {name} registered successfully"
        )
    
    def GenerateText(self, request, context):
        """Generate text based on a prompt"""
        model_name = request.model_name
        prompt = request.prompt
        max_length = request.max_length
        temperature = request.temperature
        
        # For now, just return a simple response
        response_text = f"This is a generated response to: {prompt}"
        
        logger.info(f"Generated text using model: {model_name}")
        return ai_service_pb2.TextGenerationResponse(
            texts=[response_text],
            success=True,
            message="Text generated successfully"
        )
    
    def GenerateEmbedding(self, request, context):
        """Generate embeddings for the given texts"""
        model_name = request.model_name
        texts = list(request.texts)
        
        # For now, generate random embeddings
        embeddings = []
        for text in texts:
            embedding = np.random.rand(768).astype(np.float32).tolist()
            embeddings.append(embedding)
        
        logger.info(f"Generated {len(embeddings)} embeddings using model: {model_name}")
        return ai_service_pb2.EmbeddingResponse(
            embeddings=embeddings,
            success=True,
            message="Embeddings generated successfully"
        )

def serve_vector_service():
    """Start the vector service on port 50051"""
    # Create a gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Add vector service with Milvus client
    ai_service_pb2_grpc.add_VectorServiceServicer_to_server(VectorServiceImpl(), server)
    
    # Bind the server to the port 50051
    vector_port = '[::]:50051'
    server.add_insecure_port(vector_port)
    logger.info(f"Vector service starting on port {vector_port}")
    
    # Start the server
    server.start()
    logger.info("Vector service started")
    
    try:
        while True:
            time.sleep(86400)  # One day
    except KeyboardInterrupt:
        server.stop(0)

def serve_model_service():
    """Start the model service on port 50052"""
    # Create a gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Add model service
    ai_service_pb2_grpc.add_ModelServiceServicer_to_server(ModelServiceImpl(), server)
    
    # Bind the server to the port 50052
    model_port = '[::]:50052'
    server.add_insecure_port(model_port)
    logger.info(f"Model service starting on port {model_port}")
    
    # Start the server
    server.start()
    logger.info("Model service started")
    
    try:
        while True:
            time.sleep(86400)  # One day
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == "__main__":
    # Start both services
    import threading
    
    vector_thread = threading.Thread(target=serve_vector_service)
    model_thread = threading.Thread(target=serve_model_service)
    
    vector_thread.start()
    model_thread.start()
    
    vector_thread.join()
    model_thread.join()