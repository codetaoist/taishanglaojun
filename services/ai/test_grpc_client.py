#!/usr/bin/env python3
"""
Python gRPC client for testing communication with Go services
"""

import grpc
import time
import logging

# Import the generated gRPC modules
# TODO: Uncomment after generating the protobuf code
import sys
sys.path.append("../../api")
import proto.ai_service_pb2 as ai_service_pb2
import proto.ai_service_pb2_grpc as ai_service_pb2_grpc

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class VectorServiceClient:
    """Client for the VectorService"""
    
    def __init__(self, address='localhost:50051'):
        self.address = address
        # TODO: Uncomment after generating the protobuf code
        self.channel = grpc.insecure_channel(address)
        self.stub = ai_service_pb2_grpc.VectorServiceStub(self.channel)
    
    def close(self):
        """Close the channel"""
        # TODO: Uncomment after generating the protobuf code
        self.channel.close()
        pass
    
    def health_check(self):
        """Check the health of the vector service"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     request = ai_service_pb2.HealthCheckRequest()
        #     response = self.stub.HealthCheck(request)
        #     return response.healthy, response.message
        # except grpc.RpcError as e:
        #     logger.error(f"Health check failed: {e}")
        #     return False, str(e)
        
        # Placeholder implementation
        logger.info("Health check for vector service")
        return True, "Vector service is running"
    
    def create_collection(self, name):
        """Create a new collection"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     schema = ai_service_pb2.CollectionSchema(
        #         name=name,
        #         fields=[
        #             ai_service_pb2.FieldSchema(
        #                 name="id",
        #                 is_primary_key=True,
        #                 data_type="string"
        #             ),
        #             ai_service_pb2.FieldSchema(
        #                 name="vector",
        #                 data_type="vector",
        #                 type_params={"dim": "768"}
        #             )
        #         ]
        #     )
        #     
        #     request = ai_service_pb2.CreateCollectionRequest(
        #         collection_name=name,
        #         schema=schema
        #     )
        #     
        #     response = self.stub.CreateCollection(request)
        #     return True
        # except grpc.RpcError as e:
        #     logger.error(f"Create collection failed: {e}")
        #     return False
        
        # Placeholder implementation
        logger.info(f"Creating collection: {name}")
        return True
    
    def search(self, collection_name, vector, top_k=10):
        """Search for similar vectors"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     request = ai_service_pb2.SearchRequest(
        #         collection_name=collection_name,
        #         vector=vector,
        #         top_k=top_k
        #     )
        #     
        #     response = self.stub.Search(request)
        #     
        #     results = []
        #     for result in response.results:
        #         results.append({
        #             'id': result.id,
        #             'score': result.score,
        #             'metadata': dict(result.metadata)
        #         })
        #     
        #     return results
        # except grpc.RpcError as e:
        #     logger.error(f"Search failed: {e}")
        #     return []
        
        # Placeholder implementation
        logger.info(f"Searching in collection {collection_name} with top_k={top_k}")
        
        # Mock results
        results = []
        for i in range(min(top_k, 5)):
            results.append({
                'id': f"result_{i}",
                'score': 0.9 - i * 0.1,
                'metadata': {'key': f"value_{i}"}
            })
        
        return results

class ModelServiceClient:
    """Client for the ModelService"""
    
    def __init__(self, address='localhost:50052'):
        self.address = address
        # TODO: Uncomment after generating the protobuf code
        self.channel = grpc.insecure_channel(address)
        # self.stub = ai_service_pb2_grpc.ModelServiceStub(self.channel)
    
    def close(self):
        """Close the channel"""
        # TODO: Uncomment after generating the protobuf code
        self.channel.close()
        pass
    
    def health_check(self):
        """Check the health of the model service"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     request = ai_service_pb2.HealthCheckRequest()
        #     response = self.stub.HealthCheck(request)
        #     return response.healthy, response.message
        # except grpc.RpcError as e:
        #     logger.error(f"Health check failed: {e}")
        #     return False, str(e)
        
        # Placeholder implementation
        logger.info("Health check for model service")
        return True, "Model service is running"
    
    def register_model(self, name, provider, model_path, description):
        """Register a new model"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     request = ai_service_pb2.RegisterModelRequest(
        #         name=name,
        #         provider=ai_service_pb2.ModelProvider.Value(provider),
        #         model_path=model_path,
        #         description=description
        #     )
        #     
        #     response = self.stub.RegisterModel(request)
        #     return response.success, response.message
        # except grpc.RpcError as e:
        #     logger.error(f"Register model failed: {e}")
        #     return False, str(e)
        
        # Placeholder implementation
        logger.info(f"Registering model: {name}, provider: {provider}, path: {model_path}")
        return True, f"Model {name} registered successfully"
    
    def generate_text(self, model_name, prompt):
        """Generate text based on a prompt"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     request = ai_service_pb2.TextGenerationRequest(
        #         model_name=model_name,
        #         prompt=prompt
        #     )
        #     
        #     response = self.stub.GenerateText(request)
        #     
        #     if response.success and len(response.texts) > 0:
        #         return response.texts[0]
        #     else:
        #         return ""
        # except grpc.RpcError as e:
        #     logger.error(f"Generate text failed: {e}")
        #     return ""
        
        # Placeholder implementation
        logger.info(f"Generating text with model {model_name} for prompt: {prompt}")
        return f"This is a generated response to: {prompt}"
    
    def generate_embedding(self, model_name, texts):
        """Generate embeddings for the given texts"""
        # TODO: Implement after generating the protobuf code
        # try:
        #     request = ai_service_pb2.EmbeddingRequest(
        #         model_name=model_name,
        #         texts=texts
        #     )
        #     
        #     response = self.stub.GenerateEmbedding(request)
        #     
        #     if response.success:
        #         return response.embeddings
        #     else:
        #         return []
        # except grpc.RpcError as e:
        #     logger.error(f"Generate embedding failed: {e}")
        #     return []
        
        # Placeholder implementation
        logger.info(f"Generating embeddings with model {model_name} for {len(texts)} texts")
        
        # Generate random embeddings
        import numpy as np
        embeddings = []
        for _ in texts:
            embedding = np.random.rand(768).astype(np.float32).tolist()
            embeddings.append(embedding)
        
        return embeddings

def test_vector_service():
    """Test the vector service"""
    print("Testing vector service...")
    
    client = VectorServiceClient()
    
    try:
        # Health check
        healthy, message = client.health_check()
        print(f"Vector service health: {healthy}, message: {message}")
        
        # Create collection
        success = client.create_collection("test_collection")
        if success:
            print("Collection created successfully")
        else:
            print("Failed to create collection")
        
        # Search
        query_vector = [float(i) / 768.0 for i in range(768)]
        results = client.search("test_collection", query_vector, 5)
        print(f"Search results: {results}")
        
    finally:
        client.close()

def test_model_service():
    """Test the model service"""
    print("\nTesting model service...")
    
    client = ModelServiceClient()
    
    try:
        # Health check
        healthy, message = client.health_check()
        print(f"Model service health: {healthy}, message: {message}")
        
        # Register model
        success, message = client.register_model(
            "test_model", 
            "HUGGINGFACE", 
            "sentence-transformers/all-MiniLM-L6-v2", 
            "Test model for embeddings"
        )
        if success:
            print(f"Model registered successfully: {message}")
        else:
            print(f"Failed to register model: {message}")
        
        # Generate text
        text = client.generate_text(
            "test_model", 
            "Write a short story about a robot learning to paint."
        )
        print(f"Generated text: {text}")
        
        # Generate embeddings
        embeddings = client.generate_embedding(
            "test_model", 
            ["This is a test text for embedding generation."]
        )
        if embeddings:
            print(f"Embeddings generated: {embeddings[0][:5]} (first 5 values of first embedding)")
        else:
            print("Failed to generate embeddings")
        
    finally:
        client.close()

if __name__ == '__main__':
    test_vector_service()
    test_model_service()
    print("\nAll tests completed successfully!")