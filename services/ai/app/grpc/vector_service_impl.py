import grpc
from typing import List, Dict, Any, Optional
from concurrent import futures
import asyncio

# 这些导入将在生成protobuf代码后添加
# from app.proto import vector_service_pb2
# from app.proto import vector_service_pb2_grpc

from app.services.vector_db import vector_db_service
from app.utils.logger import get_logger
from app.models.vector import (
    VectorModel, SearchResultModel, SearchOptionsModel, 
    IndexParamsModel, CollectionInfoModel, CollectionStatsModel,
    IndexType, MetricType
)

logger = get_logger(__name__)


class VectorServiceImpl:
    """向量数据库gRPC服务实现"""
    
    def __init__(self):
        self.vector_db_service = vector_db_service
    
    async def HealthCheck(self, request, context):
        """健康检查"""
        try:
            is_healthy, message = self.vector_db_service.health_check()
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "healthy": is_healthy,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Health check failed", error=str(e))
            # 返回错误响应（将在生成protobuf代码后实现）
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Health check failed: {str(e)}")
            return {}
    
    async def CreateCollection(self, request, context):
        """创建集合"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            description = request.description
            vector_dim = request.vector_dim
            
            # 解析索引参数
            index_type = IndexType(request.index_params.index_type)
            metric_type = MetricType(request.index_params.metric_type)
            index_params = IndexParamsModel(
                index_type=index_type,
                metric_type=metric_type,
                params=dict(request.index_params.params)
            )
            
            # 创建集合
            success, message = self.vector_db_service.create_collection(
                name=name,
                description=description,
                vector_dim=vector_dim,
                index_params=index_params
            )
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Create collection failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Create collection failed: {str(e)}")
            return {}
    
    async def DropCollection(self, request, context):
        """删除集合"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 删除集合
            success, message = self.vector_db_service.drop_collection(name=name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Drop collection failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Drop collection failed: {str(e)}")
            return {}
    
    async def HasCollection(self, request, context):
        """检查集合是否存在"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 检查集合是否存在
            exists = self.vector_db_service.has_collection(name=name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "exists": exists
            }
            
            return response
            
        except Exception as e:
            logger.error("Has collection check failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Has collection check failed: {str(e)}")
            return {}
    
    async def ListCollections(self, request, context):
        """列出所有集合"""
        try:
            # 获取集合列表
            collections = self.vector_db_service.list_collections()
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "collections": collections
            }
            
            return response
            
        except Exception as e:
            logger.error("List collections failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"List collections failed: {str(e)}")
            return {}
    
    async def GetCollectionInfo(self, request, context):
        """获取集合信息"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 获取集合信息
            collection_info = self.vector_db_service.get_collection_info(name=name)
            
            if not collection_info:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Collection {name} not found")
                return {}
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "name": collection_info.name,
                "description": collection_info.description,
                "vector_dim": collection_info.vector_dim,
                "created_at": collection_info.created_at,
                "updated_at": collection_info.updated_at
            }
            
            # 添加索引参数（如果有）
            if collection_info.index_params:
                response["index_params"] = {
                    "index_type": collection_info.index_params.index_type.value,
                    "metric_type": collection_info.index_params.metric_type.value,
                    "params": collection_info.index_params.params
                }
            
            return response
            
        except Exception as e:
            logger.error("Get collection info failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Get collection info failed: {str(e)}")
            return {}
    
    async def GetCollectionStats(self, request, context):
        """获取集合统计信息"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 获取集合统计信息
            collection_stats = self.vector_db_service.get_collection_stats(name=name)
            
            if not collection_stats:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Collection {name} not found")
                return {}
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "name": collection_stats.name,
                "vector_count": collection_stats.vector_count,
                "size_in_bytes": collection_stats.size_in_bytes
            }
            
            return response
            
        except Exception as e:
            logger.error("Get collection stats failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Get collection stats failed: {str(e)}")
            return {}
    
    async def CreateIndex(self, request, context):
        """创建索引"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            
            # 解析索引参数
            index_type = IndexType(request.index_params.index_type)
            metric_type = MetricType(request.index_params.metric_type)
            index_params = IndexParamsModel(
                index_type=index_type,
                metric_type=metric_type,
                params=dict(request.index_params.params)
            )
            
            # 创建索引
            success, message = self.vector_db_service.create_index(
                collection_name=collection_name,
                index_params=index_params
            )
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Create index failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Create index failed: {str(e)}")
            return {}
    
    async def DropIndex(self, request, context):
        """删除索引"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            
            # 删除索引
            success, message = self.vector_db_service.drop_index(collection_name=collection_name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Drop index failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Drop index failed: {str(e)}")
            return {}
    
    async def HasIndex(self, request, context):
        """检查集合是否有索引"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            
            # 检查索引是否存在
            has_index = self.vector_db_service.has_index(collection_name=collection_name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "has_index": has_index
            }
            
            return response
            
        except Exception as e:
            logger.error("Has index check failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Has index check failed: {str(e)}")
            return {}
    
    async def Insert(self, request, context):
        """插入向量"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            
            # 解析向量数据
            vectors = []
            for vector_data in request.vectors:
                vector = VectorModel(
                    id=vector_data.id,
                    vector=list(vector_data.vector),
                    metadata=dict(vector_data.metadata) if vector_data.metadata else {}
                )
                vectors.append(vector)
            
            # 插入向量
            success, message, ids = self.vector_db_service.insert(
                collection_name=collection_name,
                vectors=vectors
            )
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message,
                "ids": ids
            }
            
            return response
            
        except Exception as e:
            logger.error("Insert vectors failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Insert vectors failed: {str(e)}")
            return {}
    
    async def Upsert(self, request, context):
        """更新或插入向量"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            
            # 解析向量数据
            vectors = []
            for vector_data in request.vectors:
                vector = VectorModel(
                    id=vector_data.id,
                    vector=list(vector_data.vector),
                    metadata=dict(vector_data.metadata) if vector_data.metadata else {}
                )
                vectors.append(vector)
            
            # 更新或插入向量
            success, message, ids = self.vector_db_service.upsert(
                collection_name=collection_name,
                vectors=vectors
            )
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message,
                "ids": ids
            }
            
            return response
            
        except Exception as e:
            logger.error("Upsert vectors failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Upsert vectors failed: {str(e)}")
            return {}
    
    async def Delete(self, request, context):
        """删除向量"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            ids = list(request.ids)
            
            # 删除向量
            success, message = self.vector_db_service.delete(
                collection_name=collection_name,
                ids=ids
            )
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Delete vectors failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Delete vectors failed: {str(e)}")
            return {}
    
    async def Search(self, request, context):
        """搜索向量"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            query_vector = list(request.query_vector)
            
            # 解析搜索选项
            search_options = SearchOptionsModel(
                topk=request.topk,
                include_vector=request.include_vector if hasattr(request, 'include_vector') else False
            )
            
            # 搜索向量
            success, message, search_results = self.vector_db_service.search(
                collection_name=collection_name,
                query_vector=query_vector,
                options=search_options
            )
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message,
                "results": []
            }
            
            # 添加搜索结果
            for result in search_results:
                result_data = {
                    "id": result.id,
                    "score": result.score
                }
                
                if result.vector:
                    result_data["vector"] = result.vector
                
                if result.metadata:
                    result_data["metadata"] = result.metadata
                
                response["results"].append(result_data)
            
            return response
            
        except Exception as e:
            logger.error("Search vectors failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Search vectors failed: {str(e)}")
            return {}
    
    async def GetById(self, request, context):
        """根据ID获取向量"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            id = request.id
            
            # 获取向量
            success, message, vector = self.vector_db_service.get_by_id(
                collection_name=collection_name,
                id=id
            )
            
            if not success:
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details(message)
                return {}
            
            if not vector:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Vector with ID {id} not found")
                return {}
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "id": vector.id,
                "vector": vector.vector,
                "metadata": vector.metadata
            }
            
            return response
            
        except Exception as e:
            logger.error("Get vector by ID failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Get vector by ID failed: {str(e)}")
            return {}
    
    async def Compact(self, request, context):
        """压缩集合"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            collection_name = request.collection_name
            
            # 压缩集合
            success, message = self.vector_db_service.compact(collection_name=collection_name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Compact collection failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Compact collection failed: {str(e)}")
            return {}