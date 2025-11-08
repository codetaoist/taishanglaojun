from typing import Dict, List, Any, Optional, Tuple
import torch
from transformers import AutoTokenizer, AutoModel, AutoModelForCausalLM, pipeline
from sentence_transformers import SentenceTransformer
from openai import OpenAI
from app.core.config import settings
from app.utils.logger import get_logger
from app.models.model import (
    ModelConfigModel, RegisterModelRequestModel, UpdateModelRequestModel,
    TextGenerationRequestModel, TextGenerationResponseModel,
    EmbeddingRequestModel, EmbeddingResponseModel,
    ModelProvider
)

logger = get_logger(__name__)


class ModelService:
    """模型服务"""
    
    def __init__(self):
        self.models: Dict[str, ModelConfigModel] = {}
        self.loaded_models: Dict[str, Any] = {}
        self.openai_client = None
        self._init_openai()
    
    def _init_openai(self):
        """初始化OpenAI客户端"""
        if settings.openai_api_key:
            self.openai_client = OpenAI(api_key=settings.openai_api_key)
            logger.info("OpenAI client initialized")
        else:
            logger.warning("OpenAI API key not provided")
    
    def register_model(self, request: RegisterModelRequestModel) -> Tuple[bool, str]:
        """注册模型"""
        try:
            # 检查模型是否已注册
            if request.name in self.models:
                return False, f"Model {request.name} already registered"
            
            # 创建模型配置
            model_config = ModelConfigModel(
                name=request.name,
                provider=request.provider,
                model_path=request.model_path,
                model_type=request.model_type,
                description=request.description,
                is_default=request.is_default,
                config=request.config or {}
            )
            
            # 注册模型
            self.models[request.name] = model_config
            
            # 如果是默认模型，更新其他模型的默认状态
            if request.is_default:
                for name, config in self.models.items():
                    if name != request.name:
                        config.is_default = False
            
            logger.info("Registered model", name=request.name, provider=request.provider.value)
            return True, f"Model {request.name} registered successfully"
            
        except Exception as e:
            logger.error("Failed to register model", name=request.name, error=str(e))
            return False, f"Failed to register model: {str(e)}"
    
    def update_model(self, request: UpdateModelRequestModel) -> Tuple[bool, str]:
        """更新模型"""
        try:
            # 检查模型是否存在
            if request.name not in self.models:
                return False, f"Model {request.name} not found"
            
            # 更新模型配置
            model_config = self.models[request.name]
            
            if request.provider is not None:
                model_config.provider = request.provider
            if request.model_path is not None:
                model_config.model_path = request.model_path
            if request.model_type is not None:
                model_config.model_type = request.model_type
            if request.description is not None:
                model_config.description = request.description
            if request.is_default is not None:
                model_config.is_default = request.is_default
                # 如果是默认模型，更新其他模型的默认状态
                if request.is_default:
                    for name, config in self.models.items():
                        if name != request.name:
                            config.is_default = False
            if request.config is not None:
                model_config.config.update(request.config)
            
            # 如果模型已加载，卸载它以便重新加载
            if request.name in self.loaded_models:
                self.unload_model(request.name)
            
            logger.info("Updated model", name=request.name)
            return True, f"Model {request.name} updated successfully"
            
        except Exception as e:
            logger.error("Failed to update model", name=request.name, error=str(e))
            return False, f"Failed to update model: {str(e)}"
    
    def unregister_model(self, name: str) -> Tuple[bool, str]:
        """注销模型"""
        try:
            # 检查模型是否存在
            if name not in self.models:
                return False, f"Model {name} not found"
            
            # 如果模型已加载，卸载它
            if name in self.loaded_models:
                self.unload_model(name)
            
            # 删除模型
            del self.models[name]
            
            logger.info("Unregistered model", name=name)
            return True, f"Model {name} unregistered successfully"
            
        except Exception as e:
            logger.error("Failed to unregister model", name=name, error=str(e))
            return False, f"Failed to unregister model: {str(e)}"
    
    def list_models(self) -> List[ModelConfigModel]:
        """列出所有模型"""
        return list(self.models.values())
    
    def get_model(self, name: str) -> Optional[ModelConfigModel]:
        """获取模型配置"""
        return self.models.get(name)
    
    def get_default_model(self, model_type: str) -> Optional[ModelConfigModel]:
        """获取默认模型"""
        for config in self.models.values():
            if config.model_type == model_type and config.is_default:
                return config
        return None
    
    def load_model(self, name: str) -> Tuple[bool, str]:
        """加载模型"""
        try:
            # 检查模型是否存在
            if name not in self.models:
                return False, f"Model {name} not found"
            
            # 检查模型是否已加载
            if name in self.loaded_models:
                return True, f"Model {name} already loaded"
            
            model_config = self.models[name]
            
            # 根据提供商加载模型
            if model_config.provider == ModelProvider.HUGGINGFACE:
                model = self._load_huggingface_model(model_config)
            elif model_config.provider == ModelProvider.OPENAI:
                # OpenAI模型不需要加载，只需验证API密钥
                if not self.openai_client:
                    return False, "OpenAI client not initialized"
                model = "openai"
            elif model_config.provider == ModelProvider.SENTENCE_TRANSFORMERS:
                model = self._load_sentence_transformers_model(model_config)
            else:
                return False, f"Unsupported provider: {model_config.provider.value}"
            
            # 存储加载的模型
            self.loaded_models[name] = model
            
            logger.info("Loaded model", name=name, provider=model_config.provider.value)
            return True, f"Model {name} loaded successfully"
            
        except Exception as e:
            logger.error("Failed to load model", name=name, error=str(e))
            return False, f"Failed to load model: {str(e)}"
    
    def _load_huggingface_model(self, config: ModelConfigModel) -> Any:
        """加载HuggingFace模型"""
        if config.model_type == "embedding":
            # 加载嵌入模型
            tokenizer = AutoTokenizer.from_pretrained(config.model_path)
            model = AutoModel.from_pretrained(config.model_path)
            return {"tokenizer": tokenizer, "model": model}
        elif config.model_type == "generation":
            # 加载生成模型
            tokenizer = AutoTokenizer.from_pretrained(config.model_path)
            model = AutoModelForCausalLM.from_pretrained(config.model_path)
            return {"tokenizer": tokenizer, "model": model}
        else:
            raise ValueError(f"Unsupported model type: {config.model_type}")
    
    def _load_sentence_transformers_model(self, config: ModelConfigModel) -> Any:
        """加载SentenceTransformers模型"""
        return SentenceTransformer(config.model_path)
    
    def unload_model(self, name: str) -> Tuple[bool, str]:
        """卸载模型"""
        try:
            # 检查模型是否已加载
            if name not in self.loaded_models:
                return True, f"Model {name} not loaded"
            
            # 卸载模型
            del self.loaded_models[name]
            
            logger.info("Unloaded model", name=name)
            return True, f"Model {name} unloaded successfully"
            
        except Exception as e:
            logger.error("Failed to unload model", name=name, error=str(e))
            return False, f"Failed to unload model: {str(e)}"
    
    def is_model_loaded(self, name: str) -> bool:
        """检查模型是否已加载"""
        return name in self.loaded_models
    
    def generate_text(self, request: TextGenerationRequestModel) -> Tuple[bool, str, Optional[TextGenerationResponseModel]]:
        """生成文本"""
        try:
            # 确定使用的模型
            model_name = request.model_name
            if not model_name:
                # 使用默认生成模型
                default_model = self.get_default_model("generation")
                if not default_model:
                    return False, "No default generation model available", None
                model_name = default_model.name
            
            # 检查模型是否存在
            if model_name not in self.models:
                return False, f"Model {model_name} not found", None
            
            # 加载模型（如果尚未加载）
            if model_name not in self.loaded_models:
                success, message = self.load_model(model_name)
                if not success:
                    return False, message, None
            
            model_config = self.models[model_name]
            loaded_model = self.loaded_models[model_name]
            
            # 根据提供商生成文本
            if model_config.provider == ModelProvider.HUGGINGFACE:
                text, tokens = self._generate_with_huggingface(loaded_model, request)
            elif model_config.provider == ModelProvider.OPENAI:
                text, tokens = self._generate_with_openai(model_config.model_path, request)
            else:
                return False, f"Text generation not supported for provider: {model_config.provider.value}", None
            
            response = TextGenerationResponseModel(
                text=text,
                model_name=model_name,
                tokens_used=tokens,
                finish_reason="completed"
            )
            
            logger.info("Generated text", model=model_name, tokens=tokens)
            return True, "Text generated successfully", response
            
        except Exception as e:
            logger.error("Failed to generate text", model=request.model_name, error=str(e))
            return False, f"Failed to generate text: {str(e)}", None
    
    def _generate_with_huggingface(self, model: Dict[str, Any], request: TextGenerationRequestModel) -> Tuple[str, int]:
        """使用HuggingFace模型生成文本"""
        tokenizer = model["tokenizer"]
        model_instance = model["model"]
        
        # 编码输入
        inputs = tokenizer.encode(request.prompt, return_tensors="pt")
        
        # 生成文本
        with torch.no_grad():
            outputs = model_instance.generate(
                inputs,
                max_length=request.max_tokens,
                temperature=request.temperature,
                top_p=request.top_p,
                do_sample=True,
                pad_token_id=tokenizer.eos_token_id
            )
        
        # 解码输出
        generated_text = tokenizer.decode(outputs[0], skip_special_tokens=True)
        
        # 计算使用的token数
        tokens_used = len(outputs[0]) - len(inputs[0])
        
        return generated_text, tokens_used
    
    def _generate_with_openai(self, model_path: str, request: TextGenerationRequestModel) -> Tuple[str, int]:
        """使用OpenAI模型生成文本"""
        response = self.openai_client.chat.completions.create(
            model=model_path,
            messages=[{"role": "user", "content": request.prompt}],
            max_tokens=request.max_tokens,
            temperature=request.temperature,
            top_p=request.top_p
        )
        
        text = response.choices[0].message.content
        tokens_used = response.usage.total_tokens
        
        return text, tokens_used
    
    def generate_embedding(self, request: EmbeddingRequestModel) -> Tuple[bool, str, Optional[EmbeddingResponseModel]]:
        """生成嵌入"""
        try:
            # 确定使用的模型
            model_name = request.model_name
            if not model_name:
                # 使用默认嵌入模型
                default_model = self.get_default_model("embedding")
                if not default_model:
                    return False, "No default embedding model available", None
                model_name = default_model.name
            
            # 检查模型是否存在
            if model_name not in self.models:
                return False, f"Model {model_name} not found", None
            
            # 加载模型（如果尚未加载）
            if model_name not in self.loaded_models:
                success, message = self.load_model(model_name)
                if not success:
                    return False, message, None
            
            model_config = self.models[model_name]
            loaded_model = self.loaded_models[model_name]
            
            # 根据提供商生成嵌入
            if model_config.provider == ModelProvider.HUGGINGFACE:
                embedding = self._embed_with_huggingface(loaded_model, request.text)
            elif model_config.provider == ModelProvider.OPENAI:
                embedding = self._embed_with_openai(model_config.model_path, request.text)
            elif model_config.provider == ModelProvider.SENTENCE_TRANSFORMERS:
                embedding = self._embed_with_sentence_transformers(loaded_model, request.text)
            else:
                return False, f"Embedding not supported for provider: {model_config.provider.value}", None
            
            response = EmbeddingResponseModel(
                embedding=embedding,
                model_name=model_name,
                dimension=len(embedding)
            )
            
            logger.info("Generated embedding", model=model_name, dimension=len(embedding))
            return True, "Embedding generated successfully", response
            
        except Exception as e:
            logger.error("Failed to generate embedding", model=request.model_name, error=str(e))
            return False, f"Failed to generate embedding: {str(e)}", None
    
    def _embed_with_huggingface(self, model: Dict[str, Any], text: str) -> List[float]:
        """使用HuggingFace模型生成嵌入"""
        tokenizer = model["tokenizer"]
        model_instance = model["model"]
        
        # 编码输入
        inputs = tokenizer(text, return_tensors="pt", padding=True, truncation=True)
        
        # 生成嵌入
        with torch.no_grad():
            outputs = model_instance(**inputs)
            # 使用[CLS] token的嵌入作为句子嵌入
            embedding = outputs.last_hidden_state[:, 0, :].squeeze().tolist()
        
        return embedding
    
    def _embed_with_openai(self, model_path: str, text: str) -> List[float]:
        """使用OpenAI模型生成嵌入"""
        response = self.openai_client.embeddings.create(
            model=model_path,
            input=text
        )
        
        return response.data[0].embedding
    
    def _embed_with_sentence_transformers(self, model: Any, text: str) -> List[float]:
        """使用SentenceTransformers模型生成嵌入"""
        embedding = model.encode(text)
        return embedding.tolist()


# 全局模型服务实例
model_service = ModelService()