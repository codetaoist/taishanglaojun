from pydantic import BaseModel, Field
from typing import Dict, Any, Optional, List
from enum import Enum


class ModelProvider(str, Enum):
    """模型提供商"""
    OPENAI = "openai"
    HUGGINGFACE = "huggingface"
    OLLAMA = "ollama"
    CUSTOM = "custom"


class ModelConfigModel(BaseModel):
    """模型配置"""
    name: str = Field(..., description="模型名称")
    provider: ModelProvider = Field(..., description="模型提供商")
    model_id: str = Field(..., description="模型ID")
    parameters: Dict[str, Any] = Field(default_factory=dict, description="模型参数")
    enabled: bool = Field(True, description="是否启用")


class RegisterModelRequestModel(BaseModel):
    """注册模型请求"""
    config: ModelConfigModel = Field(..., description="模型配置")


class UpdateModelConfigRequestModel(BaseModel):
    """更新模型配置请求"""
    name: str = Field(..., description="模型名称")
    config: ModelConfigModel = Field(..., description="模型配置")


class GenerateTextRequestModel(BaseModel):
    """生成文本请求"""
    model_name: str = Field(..., description="模型名称")
    prompt: str = Field(..., description="提示词")
    parameters: Dict[str, Any] = Field(default_factory=dict, description="生成参数")


class GenerateTextResponseModel(BaseModel):
    """生成文本响应"""
    text: str = Field(..., description="生成的文本")
    tokens_used: int = Field(..., description="使用的令牌数")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="元数据")


class GenerateEmbeddingRequestModel(BaseModel):
    """生成嵌入请求"""
    model_name: str = Field(..., description="模型名称")
    text: str = Field(..., description="文本")


class GenerateEmbeddingResponseModel(BaseModel):
    """生成嵌入响应"""
    embedding: List[float] = Field(..., description="嵌入向量")
    dimension: int = Field(..., description="向量维度")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="元数据")


class StreamGenerateTextRequestModel(BaseModel):
    """流式生成文本请求"""
    model_name: str = Field(..., description="模型名称")
    prompt: str = Field(..., description="提示词")
    parameters: Dict[str, Any] = Field(default_factory=dict, description="生成参数")


class StreamGenerateTextResponseModel(BaseModel):
    """流式生成文本响应"""
    text_chunk: str = Field(..., description="文本片段")
    finished: bool = Field(False, description="是否完成")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="元数据")