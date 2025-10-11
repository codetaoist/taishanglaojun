"""
太上老君AI平台 Python SDK 基础使用示例

本示例展示了如何使用Python SDK进行基本的API调用
"""

import os
import asyncio
import time
from typing import List, Dict, Any
from taishanglaojun import TaiShangLaoJunAPI, TaiShangLaoJunError

# 初始化API客户端
api = TaiShangLaoJunAPI(
    api_key=os.getenv('TAISHANGLAOJUN_API_KEY'),
    base_url='https://api.taishanglaojun.com/v1',
    timeout=30  # 30秒超时
)


def basic_chat_example():
    """基础聊天示例"""
    print('=== 基础聊天示例 ===')
    
    try:
        response = api.chat.send(
            message="你好，请介绍一下太上老君AI平台的主要功能",
            model="taishanglaojun-v1"
        )

        print(f'AI回复: {response.message}')
        print(f'对话ID: {response.conversation_id}')
        print(f'使用的模型: {response.model}')
        print(f'消耗的token数: {response.usage}')
    except TaiShangLaoJunError as e:
        print(f'聊天失败: {e.message}')


def stream_chat_example():
    """流式聊天示例"""
    print('\n=== 流式聊天示例 ===')
    
    try:
        stream = api.chat.stream(
            message="请写一首关于人工智能的诗",
            model="taishanglaojun-v1"
        )

        print('AI正在回复...')
        full_response = ''
        
        for chunk in stream:
            if chunk.content:
                print(chunk.content, end='', flush=True)
                full_response += chunk.content
        
        print(f'\n完整回复: {full_response}')
    except TaiShangLaoJunError as e:
        print(f'流式聊天失败: {e.message}')


def knowledge_base_example():
    """知识库操作示例"""
    print('\n=== 知识库操作示例 ===')
    
    try:
        # 创建知识库
        knowledge_base = api.knowledge.create(
            name="产品文档库",
            description="存储产品相关文档和FAQ",
            type="document"
        )
        
        print(f'创建的知识库: {knowledge_base}')

        # 搜索知识库
        search_results = api.knowledge.search(
            query="如何使用AI对话功能",
            knowledge_base_id=knowledge_base.id,
            limit=5
        )
        
        print(f'搜索结果: {search_results}')

        # 获取知识库列表
        knowledge_bases = api.knowledge.list(
            page=1,
            limit=10
        )
        
        print(f'知识库列表: {knowledge_bases}')
    except TaiShangLaoJunError as e:
        print(f'知识库操作失败: {e.message}')


def user_management_example():
    """用户管理示例"""
    print('\n=== 用户管理示例 ===')
    
    try:
        # 创建用户
        user = api.users.create(
            email="test@example.com",
            name="测试用户",
            role="user",
            metadata={
                "department": "技术部",
                "position": "工程师"
            }
        )
        
        print(f'创建的用户: {user}')

        # 获取用户信息
        user_info = api.users.get(user.id)
        print(f'用户信息: {user_info}')

        # 更新用户信息
        updated_user = api.users.update(
            user.id,
            name="更新后的用户名",
            metadata={
                "department": "产品部",
                "position": "产品经理"
            }
        )
        
        print(f'更新后的用户: {updated_user}')

        # 获取用户列表
        users = api.users.list(
            page=1,
            limit=10,
            role="user"
        )
        
        print(f'用户列表: {users}')
    except TaiShangLaoJunError as e:
        print(f'用户管理失败: {e.message}')


def api_key_management_example():
    """API密钥管理示例"""
    print('\n=== API密钥管理示例 ===')
    
    try:
        # 创建API密钥
        from datetime import datetime, timedelta
        
        api_key = api.api_keys.create(
            name="测试API密钥",
            description="用于测试的API密钥",
            permissions=["chat:read", "chat:write", "knowledge:read"],
            expires_at=datetime.now() + timedelta(days=30)  # 30天后过期
        )
        
        print(f'创建的API密钥: {api_key}')

        # 获取API密钥列表
        api_keys = api.api_keys.list(
            page=1,
            limit=10
        )
        
        print(f'API密钥列表: {api_keys}')

        # 更新API密钥
        updated_api_key = api.api_keys.update(
            api_key.id,
            name="更新后的API密钥",
            permissions=["chat:read", "knowledge:read"]
        )
        
        print(f'更新后的API密钥: {updated_api_key}')
    except TaiShangLaoJunError as e:
        print(f'API密钥管理失败: {e.message}')


def error_handling_example():
    """错误处理示例"""
    print('\n=== 错误处理示例 ===')
    
    try:
        # 故意使用无效的API密钥
        invalid_api = TaiShangLaoJunAPI(
            api_key='invalid_key',
            base_url='https://api.taishanglaojun.com/v1'
        )

        invalid_api.chat.send(message="这个请求会失败")
    except TaiShangLaoJunError as e:
        print('捕获到错误:')
        print(f'- 错误代码: {e.code}')
        print(f'- 错误消息: {e.message}')
        print(f'- HTTP状态码: {e.status}')
        
        # 根据错误类型进行不同处理
        if e.code == 'UNAUTHORIZED':
            print('处理方案: 检查API密钥是否正确')
        elif e.code == 'RATE_LIMITED':
            print('处理方案: 等待一段时间后重试')
        elif e.code == 'VALIDATION_ERROR':
            print('处理方案: 检查请求参数是否正确')
        else:
            print('处理方案: 联系技术支持')


def retry_example():
    """重试机制示例"""
    print('\n=== 重试机制示例 ===')
    
    def api_call_with_retry(api_call, max_retries=3):
        """带重试机制的API调用"""
        for i in range(max_retries):
            try:
                return api_call()
            except TaiShangLaoJunError as e:
                print(f'第{i + 1}次尝试失败: {e.message}')
                
                if e.code == 'RATE_LIMITED' and i < max_retries - 1:
                    delay = 2 ** i  # 指数退避
                    print(f'等待{delay}秒后重试...')
                    time.sleep(delay)
                    continue
                
                if i == max_retries - 1:
                    raise e  # 最后一次尝试失败，抛出错误

    try:
        result = api_call_with_retry(
            lambda: api.chat.send(message="测试重试机制")
        )
        print(f'重试成功: {result.message}')
    except TaiShangLaoJunError as e:
        print(f'重试失败: {e.message}')


def batch_operation_example():
    """批量操作示例"""
    print('\n=== 批量操作示例 ===')
    
    try:
        # 批量发送消息
        messages = [
            "什么是人工智能？",
            "机器学习的基本原理是什么？",
            "深度学习和机器学习有什么区别？"
        ]

        batch_response = api.chat.batch(
            messages=[{"message": msg} for msg in messages],
            model="taishanglaojun-v1"
        )

        print('批量聊天结果:')
        for i, (question, response) in enumerate(zip(messages, batch_response)):
            print(f'问题{i + 1}: {question}')
            print(f'回答{i + 1}: {response.message}')
            print('---')
    except TaiShangLaoJunError as e:
        print(f'批量操作失败: {e.message}')


async def async_example():
    """异步操作示例"""
    print('\n=== 异步操作示例 ===')
    
    try:
        # 创建异步API客户端
        async_api = TaiShangLaoJunAPI(
            api_key=os.getenv('TAISHANGLAOJUN_API_KEY'),
            base_url='https://api.taishanglaojun.com/v1',
            async_mode=True
        )

        # 并发发送多个请求
        tasks = [
            async_api.chat.send(message="什么是人工智能？"),
            async_api.chat.send(message="机器学习的基本原理是什么？"),
            async_api.chat.send(message="深度学习和机器学习有什么区别？")
        ]

        responses = await asyncio.gather(*tasks)
        
        print('并发聊天结果:')
        for i, response in enumerate(responses):
            print(f'回答{i + 1}: {response.message}')
            print('---')
    except TaiShangLaoJunError as e:
        print(f'异步操作失败: {e.message}')


class ChatBot:
    """聊天机器人示例类"""
    
    def __init__(self, api_key: str):
        self.api = TaiShangLaoJunAPI(api_key=api_key)
        self.conversations = {}
    
    def handle_message(self, user_id: str, message: str) -> str:
        """处理用户消息"""
        try:
            conversation_id = self.conversations.get(user_id)
            
            response = self.api.chat.send(
                message=message,
                conversation_id=conversation_id,
                user_id=user_id
            )
            
            self.conversations[user_id] = response.conversation_id
            return response.message
        except TaiShangLaoJunError as e:
            return f"抱歉，处理消息时出现错误: {e.message}"
    
    def clear_conversation(self, user_id: str):
        """清除用户对话历史"""
        if user_id in self.conversations:
            del self.conversations[user_id]


def chatbot_example():
    """聊天机器人示例"""
    print('\n=== 聊天机器人示例 ===')
    
    if not os.getenv('TAISHANGLAOJUN_API_KEY'):
        print('跳过聊天机器人示例: 未设置API密钥')
        return
    
    bot = ChatBot(os.getenv('TAISHANGLAOJUN_API_KEY'))
    
    # 模拟用户对话
    user_id = "user_123"
    messages = [
        "你好！",
        "你能帮我解释一下什么是机器学习吗？",
        "那深度学习呢？",
        "谢谢你的解释！"
    ]
    
    for message in messages:
        print(f'用户: {message}')
        response = bot.handle_message(user_id, message)
        print(f'机器人: {response}')
        print('---')


def main():
    """主函数 - 运行所有示例"""
    print('太上老君AI平台 Python SDK 示例')
    print('===================================')

    # 检查API密钥
    if not os.getenv('TAISHANGLAOJUN_API_KEY'):
        print('错误: 请设置环境变量 TAISHANGLAOJUN_API_KEY')
        return

    try:
        basic_chat_example()
        stream_chat_example()
        knowledge_base_example()
        user_management_example()
        api_key_management_example()
        error_handling_example()
        retry_example()
        batch_operation_example()
        chatbot_example()
        
        # 运行异步示例
        asyncio.run(async_example())
        
    except Exception as e:
        print(f'示例运行失败: {e}')

    print('\n所有示例运行完成！')


if __name__ == '__main__':
    main()