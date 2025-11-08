from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    # 应用配置
    app_name: str = "Taishang AI Service"
    app_version: str = "1.0.0"
    debug: bool = False
    log_level: str = "info"
    
    # 服务端口
    http_port: int = 8083
    grpc_port: int = 50051
    
    # Milvus配置
    milvus_host: str = "localhost"
    milvus_port: int = 19530
    milvus_user: str = ""
    milvus_password: str = ""
    milvus_database: str = ""
    
    # 模型配置
    model_cache_dir: str = "/app/models"
    default_embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    
    # OpenAI配置
    openai_api_key: Optional[str] = None
    
    # 安全配置
    jwt_secret_key: str = "your-secret-key-change-in-production"
    
    class Config:
        env_file = ".env"
        case_sensitive = False


settings = Settings()