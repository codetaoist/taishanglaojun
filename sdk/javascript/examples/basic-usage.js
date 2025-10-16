/**
 * 太上老君AI平台 JavaScript SDK 基础使用示例
 * 
 * 本示例展示了如何使用JavaScript SDK进行基本的API调用
 */

const { TaiShangLaoJunAPI } = require('@taishanglaojun/sdk');

// 初始化API客户端
const api = new TaiShangLaoJunAPI({
  apiKey: process.env.TAISHANGLAOJUN_API_KEY,
  baseURL: 'https://api.taishanglaojun.com/v1',
  timeout: 30000 // 30秒超时
});

/**
 * 基础聊天示例
 */
async function basicChatExample() {
  console.log('=== 基础聊天示例 ===');
  
  try {
    const response = await api.chat.send({
      message: "你好，请介绍一下太上老君AI平台的主要功能",
      model: "taishanglaojun-v1"
    });

    console.log('AI回复:', response.message);
    console.log('对话ID:', response.conversation_id);
    console.log('使用的模型:', response.model);
    console.log('消耗的token数:', response.usage);
  } catch (error) {
    console.error('聊天失败:', error.message);
  }
}

/**
 * 流式聊天示例
 */
async function streamChatExample() {
  console.log('\n=== 流式聊天示例 ===');
  
  try {
    const stream = api.chat.stream({
      message: "请写一首关于人工智能的诗",
      model: "taishanglaojun-v1"
    });

    console.log('AI正在回复...');
    let fullResponse = '';
    
    for await (const chunk of stream) {
      if (chunk.content) {
        process.stdout.write(chunk.content);
        fullResponse += chunk.content;
      }
    }
    
    console.log('\n完整回复:', fullResponse);
  } catch (error) {
    console.error('流式聊天失败:', error.message);
  }
}

/**
 * 知识库操作示例
 */
async function knowledgeBaseExample() {
  console.log('\n=== 知识库操作示例 ===');
  
  try {
    // 创建知识库
    const knowledgeBase = await api.knowledge.create({
      name: "产品文档库",
      description: "存储产品相关文档和FAQ",
      type: "document"
    });
    
    console.log('创建的知识库:', knowledgeBase);

    // 搜索知识库
    const searchResults = await api.knowledge.search({
      query: "如何使用AI对话功能",
      knowledge_base_id: knowledgeBase.id,
      limit: 5
    });
    
    console.log('搜索结果:', searchResults);

    // 获取知识库列表
    const knowledgeBases = await api.knowledge.list({
      page: 1,
      limit: 10
    });
    
    console.log('知识库列表:', knowledgeBases);
  } catch (error) {
    console.error('知识库操作失败:', error.message);
  }
}

/**
 * 用户管理示例
 */
async function userManagementExample() {
  console.log('\n=== 用户管理示例 ===');
  
  try {
    // 创建用户
    const user = await api.users.create({
      email: "test@example.com",
      name: "测试用户",
      role: "user",
      metadata: {
        department: "技术部",
        position: "工程师"
      }
    });
    
    console.log('创建的用户:', user);

    // 获取用户信息
    const userInfo = await api.users.get(user.id);
    console.log('用户信息:', userInfo);

    // 更新用户信息
    const updatedUser = await api.users.update(user.id, {
      name: "更新后的用户名",
      metadata: {
        department: "产品部",
        position: "产品经理"
      }
    });
    
    console.log('更新后的用户:', updatedUser);

    // 获取用户列表
    const users = await api.users.list({
      page: 1,
      limit: 10,
      role: "user"
    });
    
    console.log('用户列表:', users);
  } catch (error) {
    console.error('用户管理失败:', error.message);
  }
}

/**
 * API密钥管理示例
 */
async function apiKeyManagementExample() {
  console.log('\n=== API密钥管理示例 ===');
  
  try {
    // 创建API密钥
    const apiKey = await api.apiKeys.create({
      name: "测试API密钥",
      description: "用于测试的API密钥",
      permissions: ["chat:read", "chat:write", "knowledge:read"],
      expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000) // 30天后过期
    });
    
    console.log('创建的API密钥:', apiKey);

    // 获取API密钥列表
    const apiKeys = await api.apiKeys.list({
      page: 1,
      limit: 10
    });
    
    console.log('API密钥列表:', apiKeys);

    // 更新API密钥
    const updatedApiKey = await api.apiKeys.update(apiKey.id, {
      name: "更新后的API密钥",
      permissions: ["chat:read", "knowledge:read"]
    });
    
    console.log('更新后的API密钥:', updatedApiKey);
  } catch (error) {
    console.error('API密钥管理失败:', error.message);
  }
}

/**
 * 错误处理示例
 */
async function errorHandlingExample() {
  console.log('\n=== 错误处理示例 ===');
  
  try {
    // 故意使用无效的API密钥
    const invalidAPI = new TaiShangLaoJunAPI({
      apiKey: 'invalid_key',
      baseURL: 'https://api.taishanglaojun.com/v1'
    });

    await invalidAPI.chat.send({
      message: "这个请求会失败"
    });
  } catch (error) {
    console.log('捕获到错误:');
    console.log('- 错误代码:', error.code);
    console.log('- 错误消息:', error.message);
    console.log('- HTTP状态码:', error.status);
    
    // 根据错误类型进行不同处理
    switch (error.code) {
      case 'UNAUTHORIZED':
        console.log('处理方案: 检查API密钥是否正确');
        break;
      case 'RATE_LIMITED':
        console.log('处理方案: 等待一段时间后重试');
        break;
      case 'VALIDATION_ERROR':
        console.log('处理方案: 检查请求参数是否正确');
        break;
      default:
        console.log('处理方案: 联系技术支持');
    }
  }
}

/**
 * 重试机制示例
 */
async function retryExample() {
  console.log('\n=== 重试机制示例 ===');
  
  async function apiCallWithRetry(apiCall, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
      try {
        return await apiCall();
      } catch (error) {
        console.log(`第${i + 1}次尝试失败:`, error.message);
        
        if (error.code === 'RATE_LIMITED' && i < maxRetries - 1) {
          const delay = Math.pow(2, i) * 1000; // 指数退避
          console.log(`等待${delay}ms后重试...`);
          await new Promise(resolve => setTimeout(resolve, delay));
          continue;
        }
        
        if (i === maxRetries - 1) {
          throw error; // 最后一次尝试失败，抛出错误
        }
      }
    }
  }

  try {
    const result = await apiCallWithRetry(() => 
      api.chat.send({ message: "测试重试机制" })
    );
    console.log('重试成功:', result.message);
  } catch (error) {
    console.error('重试失败:', error.message);
  }
}

/**
 * 批量操作示例
 */
async function batchOperationExample() {
  console.log('\n=== 批量操作示例 ===');
  
  try {
    // 批量发送消息
    const messages = [
      "什么是人工智能？",
      "机器学习的基本原理是什么？",
      "深度学习和机器学习有什么区别？"
    ];

    const batchResponse = await api.chat.batch({
      messages: messages.map(message => ({ message })),
      model: "taishanglaojun-v1"
    });

    console.log('批量聊天结果:');
    batchResponse.forEach((response, index) => {
      console.log(`问题${index + 1}: ${messages[index]}`);
      console.log(`回答${index + 1}: ${response.message}`);
      console.log('---');
    });
  } catch (error) {
    console.error('批量操作失败:', error.message);
  }
}

/**
 * 主函数 - 运行所有示例
 */
async function main() {
  console.log('太上老君AI平台 JavaScript SDK 示例');
  console.log('=====================================');

  // 检查API密钥
  if (!process.env.TAISHANGLAOJUN_API_KEY) {
    console.error('错误: 请设置环境变量 TAISHANGLAOJUN_API_KEY');
    process.exit(1);
  }

  try {
    await basicChatExample();
    await streamChatExample();
    await knowledgeBaseExample();
    await userManagementExample();
    await apiKeyManagementExample();
    await errorHandlingExample();
    await retryExample();
    await batchOperationExample();
  } catch (error) {
    console.error('示例运行失败:', error);
  }

  console.log('\n所有示例运行完成！');
}

// 如果直接运行此文件，则执行主函数
if (require.main === module) {
  main().catch(console.error);
}

module.exports = {
  basicChatExample,
  streamChatExample,
  knowledgeBaseExample,
  userManagementExample,
  apiKeyManagementExample,
  errorHandlingExample,
  retryExample,
  batchOperationExample
};