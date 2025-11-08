import grpc
import asyncio
from concurrent import futures
from app.core.config import settings
from app.utils.logger import get_logger
from app.grpc.vector_service_impl import VectorServiceImpl
from app.grpc.model_service_impl import ModelServiceImpl

# 这些导入将在生成protobuf代码后添加
# from app.proto import ai_service_pb2_grpc

logger = get_logger(__name__)


async def serve():
    """启动gRPC服务器"""
    # 创建gRPC服务器
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # 添加向量数据库服务
    vector_service = VectorServiceImpl()
    # ai_service_pb2_grpc.add_VectorServiceServicer_to_server(vector_service, server)
    
    # 添加模型服务
    model_service = ModelServiceImpl()
    # ai_service_pb2_grpc.add_ModelServiceServicer_to_server(model_service, server)
    
    # 绑定端口
    listen_addr = f'[::]:{settings.grpc_port}'
    server.add_insecure_port(listen_addr)
    
    logger.info("Starting gRPC server", port=settings.grpc_port)
    
    # 启动服务器
    await server.start()
    
    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Shutting down gRPC server")
        await server.stop(5)


if __name__ == '__main__':
    asyncio.run(serve())