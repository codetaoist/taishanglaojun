#!/usr/bin/env python3
"""
Milvus client wrapper for vector database operations
"""

import logging
import json
from typing import Dict, List, Any, Optional, Union
from pymilvus import (
    connections, 
    Collection, 
    CollectionSchema, 
    FieldSchema, 
    DataType,
    utility,
    MilvusException
)
import numpy as np

logger = logging.getLogger(__name__)

class MilvusClient:
    """Milvus client wrapper for vector database operations"""
    
    def __init__(self, host='localhost', port='19530'):
        """
        Initialize Milvus client
        
        Args:
            host: Milvus server host
            port: Milvus server port
        """
        self.host = host
        self.port = port
        self.connected = False
        self.collections = {}  # Cache for collection objects
        
    def connect(self):
        """Connect to Milvus server"""
        try:
            connections.connect(
                alias="default",
                host=self.host,
                port=self.port
            )
            self.connected = True
            logger.info(f"Connected to Milvus at {self.host}:{self.port}")
            return True
        except Exception as e:
            logger.error(f"Failed to connect to Milvus: {str(e)}")
            self.connected = False
            return False
    
    def disconnect(self):
        """Disconnect from Milvus server"""
        try:
            connections.disconnect("default")
            self.connected = False
            logger.info("Disconnected from Milvus")
        except Exception as e:
            logger.error(f"Error disconnecting from Milvus: {str(e)}")
    
    def create_collection(self, collection_name: str, schema: Dict[str, Any]) -> bool:
        """
        Create a collection with the given schema
        
        Args:
            collection_name: Name of the collection
            schema: Schema definition as a dictionary
            
        Returns:
            True if successful, False otherwise
        """
        if not self.connected:
            logger.error("Not connected to Milvus")
            return False
            
        try:
            # Check if collection already exists
            if utility.has_collection(collection_name):
                logger.warning(f"Collection {collection_name} already exists")
                return True
                
            # Parse schema from dictionary
            fields = []
            for field in schema.get('fields', []):
                field_name = field['name']
                field_type = self._get_data_type(field['type'])
                
                # Create field schema
                field_schema = FieldSchema(
                    name=field_name,
                    dtype=field_type,
                    is_primary=field.get('is_primary', False),
                    auto_id=field.get('auto_id', False),
                    max_length=field.get('max_length', 65535) if field_type == DataType.VARCHAR else None
                )
                fields.append(field_schema)
            
            # Create collection schema
            collection_schema = CollectionSchema(
                fields=fields,
                description=schema.get('description', '')
            )
            
            # Create collection
            collection = Collection(
                name=collection_name,
                schema=collection_schema
            )
            
            # Create index for vector field
            vector_field_name = None
            for field in schema.get('fields', []):
                if field['type'] in ['FLOAT_VECTOR', 'BINARY_VECTOR']:
                    vector_field_name = field['name']
                    break
                    
            if vector_field_name:
                index_params = {
                    "metric_type": "L2",
                    "index_type": "HNSW",
                    "params": {"M": 8, "efConstruction": 64}
                }
                collection.create_index(
                    field_name=vector_field_name,
                    index_params=index_params
                )
            
            logger.info(f"Created collection: {collection_name}")
            return True
            
        except Exception as e:
            logger.error(f"Error creating collection {collection_name}: {str(e)}")
            return False
    
    def drop_collection(self, collection_name: str) -> bool:
        """
        Drop a collection
        
        Args:
            collection_name: Name of the collection to drop
            
        Returns:
            True if successful, False otherwise
        """
        if not self.connected:
            logger.error("Not connected to Milvus")
            return False
            
        try:
            if utility.has_collection(collection_name):
                utility.drop_collection(collection_name)
                if collection_name in self.collections:
                    del self.collections[collection_name]
                logger.info(f"Dropped collection: {collection_name}")
            return True
            
        except Exception as e:
            logger.error(f"Error dropping collection {collection_name}: {str(e)}")
            return False
    
    def has_collection(self, collection_name: str) -> bool:
        """
        Check if a collection exists
        
        Args:
            collection_name: Name of the collection to check
            
        Returns:
            True if collection exists, False otherwise
        """
        if not self.connected:
            logger.error("Not connected to Milvus")
            return False
            
        try:
            return utility.has_collection(collection_name)
        except Exception as e:
            logger.error(f"Error checking collection {collection_name}: {str(e)}")
            return False
    
    def insert_vectors(self, collection_name: str, data: List[Dict[str, Any]]) -> bool:
        """
        Insert vectors into a collection
        
        Args:
            collection_name: Name of the collection
            data: List of dictionaries containing vector data
            
        Returns:
            True if successful, False otherwise
        """
        if not self.connected:
            logger.error("Not connected to Milvus")
            return False
            
        try:
            collection = self._get_collection(collection_name)
            if not collection:
                return False
                
            # Convert data to the format expected by Milvus
            entities = []
            schema = collection.schema
            fields = {field.name: field for field in schema.fields}
            
            for field_name in fields:
                field_data = []
                for item in data:
                    field_data.append(item.get(field_name))
                entities.append(field_data)
            
            # Insert data
            insert_result = collection.insert(entities)
            collection.flush()
            
            logger.info(f"Inserted {len(data)} vectors into {collection_name}")
            return True
            
        except Exception as e:
            logger.error(f"Error inserting vectors into {collection_name}: {str(e)}")
            return False
    
    def search_vectors(self, collection_name: str, query_vectors: List[List[float]], 
                      top_k: int = 10, expr: str = None) -> List[Dict[str, Any]]:
        """
        Search for similar vectors
        
        Args:
            collection_name: Name of the collection
            query_vectors: List of query vectors
            top_k: Number of results to return
            expr: Filter expression (optional)
            
        Returns:
            List of search results
        """
        if not self.connected:
            logger.error("Not connected to Milvus")
            return []
            
        try:
            collection = self._get_collection(collection_name)
            if not collection:
                return []
                
            # Load collection into memory
            collection.load()
            
            # Define search parameters
            search_params = {
                "metric_type": "L2",
                "params": {"nprobe": 10}
            }
            
            # Perform search
            results = collection.search(
                data=query_vectors,
                anns_field="vector",  # Assuming vector field is named "vector"
                param=search_params,
                limit=top_k,
                expr=expr,
                output_fields=["*"]  # Return all fields
            )
            
            # Process results
            search_results = []
            for query_result in results:
                for hit in query_result:
                    result_item = {
                        "id": str(hit.id),
                        "score": float(hit.score),
                        "metadata": {}
                    }
                    
                    # Add metadata fields
                    for field_name, field_value in hit.entity.items():
                        if field_name != "vector":  # Skip the vector field
                            result_item["metadata"][field_name] = field_value
                    
                    search_results.append(result_item)
            
            logger.info(f"Search in {collection_name} returned {len(search_results)} results")
            return search_results
            
        except Exception as e:
            logger.error(f"Error searching in {collection_name}: {str(e)}")
            return []
    
    def _get_collection(self, collection_name: str) -> Optional[Collection]:
        """Get a collection object, using cache if available"""
        if collection_name in self.collections:
            return self.collections[collection_name]
            
        try:
            if utility.has_collection(collection_name):
                collection = Collection(collection_name)
                self.collections[collection_name] = collection
                return collection
            else:
                logger.error(f"Collection {collection_name} does not exist")
                return None
        except Exception as e:
            logger.error(f"Error getting collection {collection_name}: {str(e)}")
            return None
    
    def _get_data_type(self, type_str: str) -> DataType:
        """Convert string type to Milvus DataType"""
        type_map = {
            "BOOL": DataType.BOOL,
            "INT8": DataType.INT8,
            "INT16": DataType.INT16,
            "INT32": DataType.INT32,
            "INT64": DataType.INT64,
            "FLOAT": DataType.FLOAT,
            "DOUBLE": DataType.DOUBLE,
            "STRING": DataType.STRING,
            "VARCHAR": DataType.VARCHAR,
            "FLOAT_VECTOR": DataType.FLOAT_VECTOR,
            "BINARY_VECTOR": DataType.BINARY_VECTOR
        }
        
        return type_map.get(type_str, DataType.FLOAT_VECTOR)  # Default to FLOAT_VECTOR