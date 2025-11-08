import asyncio
import uvicorn
from fastapi import FastAPI
from app.grpc.server import serve as grpc_serve
from app.core.config import settings
from app.utils.logger import get_logger

logger = get_logger(__name__)

# 创建FastAPI应用
app = FastAPI(
    title=settings.app_name,
    description="AI Service for Taishang",
    version="1.0.0"
)

@app.get("/")
async def root():
    """根端点"""
    return {"message": "AI Service is running"}

@app.get("/health")
async def health_check():
    """健康检查端点"""
    return {"status": "healthy"}

async def run_servers():
    """同时运行HTTP和gRPC服务器"""
    # 创建HTTP服务器任务
    http_server = uvicorn.Server(
        uvicorn.Config(
            app=app,
            host=settings.host,
            port=settings.port,
            log_level="info"
        )
    )
    
    # 创建gRPC服务器任务
    grpc_server = grpc_serve()
    
    # 同时运行两个服务器
    await asyncio.gather(
        http_server.serve(),
        grpc_server
    )

if __name__ == "__main__":
    logger.info("Starting AI Service servers", http_port=settings.port, grpc_port=settings.grpc_port)
    asyncio.run(run_servers())