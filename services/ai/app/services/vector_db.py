from typing import List, Dict, Any, Optional, Tuple
from pymilvus import connections, Collection, FieldSchema, CollectionSchema, DataType, utility
from app.core.config import settings
from app.utils.logger import get_logger
from app.models.vector import (
    VectorModel, SearchResultModel, SearchOptionsModel, 
    IndexParamsModel, CollectionInfoModel, CollectionStatsModel,
    IndexType, MetricType
)

logger = get_logger(__name__)


class VectorDBService:
    """向量数据库服务"""
    
    def __init__(self):
        self._connected = False
        self._connect()
    
    def _connect(self):
        """连接到Milvus"""
        try:
            connections.connect(
                alias="default",
                host=settings.milvus_host,
                port=settings.milvus_port,
                user=settings.milvus_user,
                password=settings.milvus_password,
                db_name=settings.milvus_database
            )
            self._connected = True
            logger.info("Connected to Milvus", host=settings.milvus_host, port=settings.milvus_port)
        except Exception as e:
            logger.error("Failed to connect to Milvus", error=str(e))
            self._connected = False
    
    def is_connected(self) -> bool:
        """检查是否已连接"""
        return self._connected
    
    def health_check(self) -> Tuple[bool, str]:
        """健康检查"""
        if not self._connected:
            return False, "Not connected to Milvus"
        
        try:
            # 尝试列出集合来检查连接状态
            utility.list_collections()
            return True, "Healthy"
        except Exception as e:
            return False, f"Health check failed: {str(e)}"
    
    def create_collection(
        self, 
        name: str, 
        description: str, 
        vector_dim: int, 
        index_params: IndexParamsModel
    ) -> Tuple[bool, str]:
        """创建集合"""
        try:
            # 检查集合是否已存在
            if utility.has_collection(name):
                return False, f"Collection {name} already exists"
            
            # 定义字段
            id_field = FieldSchema(name="id", dtype=DataType.VARCHAR, max_length=65535, is_primary=True)
            vector_field = FieldSchema(name="vector", dtype=DataType.FLOAT_VECTOR, dim=vector_dim)
            
            # 创建集合模式
            schema = CollectionSchema(
                fields=[id_field, vector_field],
                description=description
            )
            
            # 创建集合
            collection = Collection(name, schema)
            
            # 创建索引
            index_params_dict = {
                "index_type": index_params.index_type.value,
                "metric_type": index_params.metric_type.value,
                "params": index_params.params
            }
            collection.create_index("vector", index_params_dict)
            
            logger.info("Created collection", name=name, dimension=vector_dim)
            return True, f"Collection {name} created successfully"
            
        except Exception as e:
            logger.error("Failed to create collection", name=name, error=str(e))
            return False, f"Failed to create collection: {str(e)}"
    
    def drop_collection(self, name: str) -> Tuple[bool, str]:
        """删除集合"""
        try:
            if not utility.has_collection(name):
                return False, f"Collection {name} does not exist"
            
            utility.drop_collection(name)
            logger.info("Dropped collection", name=name)
            return True, f"Collection {name} dropped successfully"
            
        except Exception as e:
            logger.error("Failed to drop collection", name=name, error=str(e))
            return False, f"Failed to drop collection: {str(e)}"
    
    def has_collection(self, name: str) -> bool:
        """检查集合是否存在"""
        try:
            return utility.has_collection(name)
        except Exception as e:
            logger.error("Failed to check collection existence", name=name, error=str(e))
            return False
    
    def list_collections(self) -> List[str]:
        """列出所有集合"""
        try:
            return utility.list_collections()
        except Exception as e:
            logger.error("Failed to list collections", error=str(e))
            return []
    
    def get_collection_info(self, name: str) -> Optional[CollectionInfoModel]:
        """获取集合信息"""
        try:
            if not utility.has_collection(name):
                return None
            
            collection = Collection(name)
            schema = collection.schema
            
            # 获取索引信息
            indexes = collection.indexes
            index_params = None
            if indexes:
                index = indexes[0]
                index_params = IndexParamsModel(
                    index_type=IndexType(index.params["index_type"]),
                    metric_type=MetricType(index.params["metric_type"]),
                    params=index.params.get("params", {})
                )
            
            return CollectionInfoModel(
                name=name,
                description=schema.description,
                vector_dim=schema.fields[1].params["dim"],
                index_params=index_params,
                created_at=int(collection.description.get("created_at", 0)),
                updated_at=int(collection.description.get("updated_at", 0))
            )
            
        except Exception as e:
            logger.error("Failed to get collection info", name=name, error=str(e))
            return None
    
    def get_collection_stats(self, name: str) -> Optional[CollectionStatsModel]:
        """获取集合统计信息"""
        try:
            if not utility.has_collection(name):
                return None
            
            collection = Collection(name)
            collection.load()
            
            # 获取向量数量
            num_entities = collection.num_entities
            
            # 估算大小（简化计算）
            vector_dim = collection.schema.fields[1].params["dim"]
            size_in_bytes = num_entities * vector_dim * 4  # 假设每个float32占4字节
            
            return CollectionStatsModel(
                name=name,
                vector_count=num_entities,
                size_in_bytes=size_in_bytes
            )
            
        except Exception as e:
            logger.error("Failed to get collection stats", name=name, error=str(e))
            return None
    
    def create_index(self, collection_name: str, index_params: IndexParamsModel) -> Tuple[bool, str]:
        """创建索引"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist"
            
            collection = Collection(collection_name)
            
            # 检查是否已有索引
            if collection.has_index():
                return False, f"Collection {collection_name} already has an index"
            
            # 创建索引
            index_params_dict = {
                "index_type": index_params.index_type.value,
                "metric_type": index_params.metric_type.value,
                "params": index_params.params
            }
            collection.create_index("vector", index_params_dict)
            
            logger.info("Created index", collection=collection_name, index_type=index_params.index_type.value)
            return True, f"Index created successfully for collection {collection_name}"
            
        except Exception as e:
            logger.error("Failed to create index", collection=collection_name, error=str(e))
            return False, f"Failed to create index: {str(e)}"
    
    def drop_index(self, collection_name: str) -> Tuple[bool, str]:
        """删除索引"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist"
            
            collection = Collection(collection_name)
            
            # 检查是否有索引
            if not collection.has_index():
                return False, f"Collection {collection_name} has no index"
            
            collection.drop_index()
            
            logger.info("Dropped index", collection=collection_name)
            return True, f"Index dropped successfully for collection {collection_name}"
            
        except Exception as e:
            logger.error("Failed to drop index", collection=collection_name, error=str(e))
            return False, f"Failed to drop index: {str(e)}"
    
    def has_index(self, collection_name: str) -> bool:
        """检查集合是否有索引"""
        try:
            if not utility.has_collection(collection_name):
                return False
            
            collection = Collection(collection_name)
            return collection.has_index()
            
        except Exception as e:
            logger.error("Failed to check index existence", collection=collection_name, error=str(e))
            return False
    
    def insert(self, collection_name: str, vectors: List[VectorModel]) -> Tuple[bool, str, List[str]]:
        """插入向量"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist", []
            
            collection = Collection(collection_name)
            
            # 准备数据
            ids = [v.id for v in vectors]
            vectors_data = [v.vector for v in vectors]
            
            # 插入数据
            insert_result = collection.insert([ids, vectors_data])
            
            logger.info("Inserted vectors", collection=collection_name, count=len(vectors))
            return True, f"Inserted {len(vectors)} vectors successfully", insert_result.primary_keys
            
        except Exception as e:
            logger.error("Failed to insert vectors", collection=collection_name, error=str(e))
            return False, f"Failed to insert vectors: {str(e)}", []
    
    def upsert(self, collection_name: str, vectors: List[VectorModel]) -> Tuple[bool, str, List[str]]:
        """更新或插入向量"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist", []
            
            collection = Collection(collection_name)
            
            # 准备数据
            ids = [v.id for v in vectors]
            vectors_data = [v.vector for v in vectors]
            
            # 更新或插入数据
            upsert_result = collection.upsert([ids, vectors_data])
            
            logger.info("Upserted vectors", collection=collection_name, count=len(vectors))
            return True, f"Upserted {len(vectors)} vectors successfully", upsert_result.primary_keys
            
        except Exception as e:
            logger.error("Failed to upsert vectors", collection=collection_name, error=str(e))
            return False, f"Failed to upsert vectors: {str(e)}", []
    
    def delete(self, collection_name: str, ids: List[str]) -> Tuple[bool, str]:
        """删除向量"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist"
            
            collection = Collection(collection_name)
            
            # 删除数据
            collection.delete(ids)
            
            logger.info("Deleted vectors", collection=collection_name, count=len(ids))
            return True, f"Deleted {len(ids)} vectors successfully"
            
        except Exception as e:
            logger.error("Failed to delete vectors", collection=collection_name, error=str(e))
            return False, f"Failed to delete vectors: {str(e)}"
    
    def search(
        self, 
        collection_name: str, 
        query_vector: List[float], 
        options: SearchOptionsModel
    ) -> Tuple[bool, str, List[SearchResultModel]]:
        """搜索向量"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist", []
            
            collection = Collection(collection_name)
            collection.load()
            
            # 执行搜索
            search_params = {
                "metric_type": "L2",  # 默认使用L2距离
                "params": {"nprobe": 10}
            }
            
            results = collection.search(
                data=[query_vector],
                anns_field="vector",
                param=search_params,
                limit=options.topk,
                expr=None  # 可以后续添加过滤条件
            )
            
            # 转换结果
            search_results = []
            for hit in results[0]:
                search_results.append(SearchResultModel(
                    id=hit.entity.get("id"),
                    score=hit.score,
                    vector=query_vector if options.include_vector else None,
                    metadata={}  # 可以后续添加元数据
                ))
            
            logger.info("Searched vectors", collection=collection_name, results_count=len(search_results))
            return True, f"Found {len(search_results)} results", search_results
            
        except Exception as e:
            logger.error("Failed to search vectors", collection=collection_name, error=str(e))
            return False, f"Failed to search vectors: {str(e)}", []
    
    def get_by_id(self, collection_name: str, id: str) -> Tuple[bool, str, Optional[VectorModel]]:
        """根据ID获取向量"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist", None
            
            collection = Collection(collection_name)
            collection.load()
            
            # 查询数据
            result = collection.query(expr=f"id == '{id}'", output_fields=["vector"])
            
            if not result:
                return False, f"Vector with ID {id} not found", None
            
            vector_data = result[0]
            vector_model = VectorModel(
                id=id,
                vector=vector_data["vector"],
                metadata={}  # 可以后续添加元数据
            )
            
            logger.info("Retrieved vector", collection=collection_name, id=id)
            return True, f"Retrieved vector with ID {id}", vector_model
            
        except Exception as e:
            logger.error("Failed to get vector by ID", collection=collection_name, id=id, error=str(e))
            return False, f"Failed to get vector by ID: {str(e)}", None
    
    def compact(self, collection_name: str) -> Tuple[bool, str]:
        """压缩集合"""
        try:
            if not utility.has_collection(collection_name):
                return False, f"Collection {collection_name} does not exist"
            
            collection = Collection(collection_name)
            collection.compact()
            
            logger.info("Compacted collection", name=collection_name)
            return True, f"Collection {collection_name} compacted successfully"
            
        except Exception as e:
            logger.error("Failed to compact collection", collection=collection_name, error=str(e))
            return False, f"Failed to compact collection: {str(e)}"


# 全局向量数据库服务实例
vector_db_service = VectorDBService()