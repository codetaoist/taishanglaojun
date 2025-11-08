#!/usr/bin/env python3
"""
测试OpenAI集成
"""
import os
import sys
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from app.services.model_service import model_service
from app.models.model import (
    RegisterModelRequestModel,
    TextGenerationRequestModel,
    EmbeddingRequestModel,
    ModelProvider
)

def test_openai_integration():
    """测试OpenAI集成"""
    print("开始测试OpenAI集成...")
    
    # 检查是否设置了OpenAI API密钥
    if not model_service.openai_client:
        print("错误: 未设置OpenAI API密钥。请在.env文件中设置OPENAI_API_KEY。")
        return False
    
    # 注册OpenAI文本生成模型
    print("\n1. 注册OpenAI文本生成模型...")
    register_request = RegisterModelRequestModel(
        name="gpt-3.5-turbo",
        provider=ModelProvider.OPENAI,
        model_path="gpt-3.5-turbo",
        model_type="generation",
        description="OpenAI GPT-3.5 Turbo模型",
        is_default=True
    )
    
    success, message = model_service.register_model(register_request)
    print(f"注册结果: {success}, 消息: {message}")
    
    # 注册OpenAI嵌入模型
    print("\n2. 注册OpenAI嵌入模型...")
    register_request = RegisterModelRequestModel(
        name="text-embedding-ada-002",
        provider=ModelProvider.OPENAI,
        model_path="text-embedding-ada-002",
        model_type="embedding",
        description="OpenAI文本嵌入模型",
        is_default=True
    )
    
    success, message = model_service.register_model(register_request)
    print(f"注册结果: {success}, 消息: {message}")
    
    # 测试文本生成
    print("\n3. 测试文本生成...")
    text_request = TextGenerationRequestModel(
        prompt="请简单介绍一下人工智能。",
        model_name="gpt-3.5-turbo",
        max_tokens=100,
        temperature=0.7
    )
    
    success, message, response = model_service.generate_text(text_request)
    if success:
        print(f"生成的文本: {response.text}")
        print(f"使用的模型: {response.model_name}")
        print(f"使用的token数: {response.tokens_used}")
    else:
        print(f"文本生成失败: {message}")
    
    # 测试嵌入生成
    print("\n4. 测试嵌入生成...")
    embedding_request = EmbeddingRequestModel(
        text="这是一个测试句子。",
        model_name="text-embedding-ada-002"
    )
    
    success, message, response = model_service.generate_embedding(embedding_request)
    if success:
        print(f"嵌入向量维度: {response.dimension}")
        print(f"嵌入向量前5个值: {response.embedding[:5]}")
    else:
        print(f"嵌入生成失败: {message}")
    
    print("\nOpenAI集成测试完成!")
    return True

if __name__ == "__main__":
    test_openai_integration()