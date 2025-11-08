import grpc
from typing import List, Dict, Any, Optional
from concurrent import futures
import asyncio

# 这些导入将在生成protobuf代码后添加
# from app.proto import model_service_pb2
# from app.proto import model_service_pb2_grpc

from app.services.model_service import model_service
from app.utils.logger import get_logger
from app.models.model import (
    ModelConfigModel, RegisterModelRequestModel, UpdateModelRequestModel,
    TextGenerationRequestModel, TextGenerationResponseModel,
    EmbeddingRequestModel, EmbeddingResponseModel,
    ModelProvider
)

logger = get_logger(__name__)


class ModelServiceImpl:
    """模型gRPC服务实现"""
    
    def __init__(self):
        self.model_service = model_service
    
    async def HealthCheck(self, request, context):
        """健康检查"""
        try:
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "healthy": True,
                "message": "Model service is healthy"
            }
            
            return response
            
        except Exception as e:
            logger.error("Health check failed", error=str(e))
            # 返回错误响应（将在生成protobuf代码后实现）
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Health check failed: {str(e)}")
            return {}
    
    async def RegisterModel(self, request, context):
        """注册模型"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            provider = ModelProvider(request.provider)
            
            register_request = RegisterModelRequestModel(
                name=request.name,
                provider=provider,
                model_path=request.model_path,
                model_type=request.model_type,
                description=request.description,
                is_default=request.is_default,
                config=dict(request.config) if request.config else {}
            )
            
            # 注册模型
            success, message = self.model_service.register_model(register_request)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Register model failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Register model failed: {str(e)}")
            return {}
    
    async def UpdateModel(self, request, context):
        """更新模型"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            provider = ModelProvider(request.provider) if request.provider else None
            
            update_request = UpdateModelRequestModel(
                name=request.name,
                provider=provider,
                model_path=request.model_path,
                model_type=request.model_type,
                description=request.description,
                is_default=request.is_default,
                config=dict(request.config) if request.config else {}
            )
            
            # 更新模型
            success, message = self.model_service.update_model(update_request)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Update model failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Update model failed: {str(e)}")
            return {}
    
    async def UnregisterModel(self, request, context):
        """注销模型"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 注销模型
            success, message = self.model_service.unregister_model(name=name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Unregister model failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Unregister model failed: {str(e)}")
            return {}
    
    async def ListModels(self, request, context):
        """列出所有模型"""
        try:
            # 获取模型列表
            models = self.model_service.list_models()
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "models": []
            }
            
            # 添加模型信息
            for model in models:
                model_data = {
                    "name": model.name,
                    "provider": model.provider.value,
                    "model_path": model.model_path,
                    "model_type": model.model_type,
                    "description": model.description,
                    "is_default": model.is_default,
                    "config": model.config
                }
                response["models"].append(model_data)
            
            return response
            
        except Exception as e:
            logger.error("List models failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"List models failed: {str(e)}")
            return {}
    
    async def GetModel(self, request, context):
        """获取模型配置"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 获取模型配置
            model = self.model_service.get_model(name=name)
            
            if not model:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Model {name} not found")
                return {}
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "name": model.name,
                "provider": model.provider.value,
                "model_path": model.model_path,
                "model_type": model.model_type,
                "description": model.description,
                "is_default": model.is_default,
                "config": model.config
            }
            
            return response
            
        except Exception as e:
            logger.error("Get model failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Get model failed: {str(e)}")
            return {}
    
    async def LoadModel(self, request, context):
        """加载模型"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 加载模型
            success, message = self.model_service.load_model(name=name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Load model failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Load model failed: {str(e)}")
            return {}
    
    async def UnloadModel(self, request, context):
        """卸载模型"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 卸载模型
            success, message = self.model_service.unload_model(name=name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "success": success,
                "message": message
            }
            
            return response
            
        except Exception as e:
            logger.error("Unload model failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Unload model failed: {str(e)}")
            return {}
    
    async def IsModelLoaded(self, request, context):
        """检查模型是否已加载"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            name = request.name
            
            # 检查模型是否已加载
            is_loaded = self.model_service.is_model_loaded(name=name)
            
            # 构建响应（将在生成protobuf代码后实现）
            response = {
                "is_loaded": is_loaded
            }
            
            return response
            
        except Exception as e:
            logger.error("Check model loaded status failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Check model loaded status failed: {str(e)}")
            return {}
    
    async def GenerateText(self, request, context):
        """生成文本"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            text_request = TextGenerationRequestModel(
                prompt=request.prompt,
                model_name=request.model_name if request.model_name else "",
                max_tokens=request.max_tokens if request.max_tokens else 100,
                temperature=request.temperature if request.temperature else 0.7,
                top_p=request.top_p if request.top_p else 1.0
            )
            
            # 生成文本
            success, message, response = self.model_service.generate_text(text_request)
            
            if not success:
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details(message)
                return {}
            
            # 构建响应（将在生成protobuf代码后实现）
            grpc_response = {
                "text": response.text,
                "model_name": response.model_name,
                "tokens_used": response.tokens_used,
                "finish_reason": response.finish_reason
            }
            
            return grpc_response
            
        except Exception as e:
            logger.error("Generate text failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Generate text failed: {str(e)}")
            return {}
    
    async def GenerateEmbedding(self, request, context):
        """生成嵌入"""
        try:
            # 解析请求（将在生成protobuf代码后实现）
            embedding_request = EmbeddingRequestModel(
                text=request.text,
                model_name=request.model_name if request.model_name else ""
            )
            
            # 生成嵌入
            success, message, response = self.model_service.generate_embedding(embedding_request)
            
            if not success:
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details(message)
                return {}
            
            # 构建响应（将在生成protobuf代码后实现）
            grpc_response = {
                "embedding": response.embedding,
                "model_name": response.model_name,
                "dimension": response.dimension
            }
            
            return grpc_response
            
        except Exception as e:
            logger.error("Generate embedding failed", error=str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Generate embedding failed: {str(e)}")
            return {}