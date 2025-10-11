// AI服务配置
export interface AIProviderConfig {
  id: string;
  name: string;
  baseUrl: string;
  apiKey?: string;
  models: AIModelConfig[];
  enabled: boolean;
  timeout: number;
  maxRetries: number;
}

export interface AIModelConfig {
  id: string;
  name: string;
  description: string;
  capabilities: string[];
  maxTokens: number;
  costPerToken?: {
    input: number;
    output: number;
  };
}

export interface AIServiceConfig {
  providers: AIProviderConfig[];
  defaultProvider: string;
  defaultModel: string;
  fallbackProvider?: string;
  globalTimeout: number;
  maxConcurrentRequests: number;
  enableCaching: boolean;
  cacheExpiry: number;
  enableMetrics: boolean;
  logLevel: 'debug' | 'info' | 'warn' | 'error';
}

// 默认AI服务配置
export const defaultAIConfig: AIServiceConfig = {
  providers: [
    {
      id: 'openai',
      name: 'OpenAI',
      baseUrl: 'https://api.openai.com/v1',
      enabled: true,
      timeout: 30000,
      maxRetries: 3,
      models: [
        {
          id: 'gpt-4',
          name: 'GPT-4',
          description: '最先进的大语言模型',
          capabilities: ['chat', 'completion', 'reasoning'],
          maxTokens: 8192,
          costPerToken: {
            input: 0.03,
            output: 0.06
          }
        },
        {
          id: 'gpt-3.5-turbo',
          name: 'GPT-3.5 Turbo',
          description: '快速且经济的语言模型',
          capabilities: ['chat', 'completion'],
          maxTokens: 4096,
          costPerToken: {
            input: 0.001,
            output: 0.002
          }
        },
        {
          id: 'dall-e-3',
          name: 'DALL-E 3',
          description: '高质量图像生成模型',
          capabilities: ['image-generation'],
          maxTokens: 0
        }
      ]
    },
    {
      id: 'anthropic',
      name: 'Anthropic',
      baseUrl: 'https://api.anthropic.com/v1',
      enabled: false,
      timeout: 30000,
      maxRetries: 3,
      models: [
        {
          id: 'claude-3-opus',
          name: 'Claude 3 Opus',
          description: 'Anthropic最强大的模型',
          capabilities: ['chat', 'completion', 'reasoning', 'analysis'],
          maxTokens: 200000
        },
        {
          id: 'claude-3-sonnet',
          name: 'Claude 3 Sonnet',
          description: '平衡性能和速度的模型',
          capabilities: ['chat', 'completion', 'reasoning'],
          maxTokens: 200000
        }
      ]
    },
    {
      id: 'local',
      name: '本地服务',
      baseUrl: 'http://localhost:8080/api/v1',
      enabled: true,
      timeout: 60000,
      maxRetries: 2,
      models: [
        {
          id: 'taishanglaojun-chat',
          name: '太上老君聊天模型',
          description: '本地部署的聊天模型',
          capabilities: ['chat', 'completion', 'cultural-wisdom'],
          maxTokens: 4096
        },
        {
          id: 'taishanglaojun-vision',
          name: '太上老君视觉模型',
          description: '本地部署的图像理解模型',
          capabilities: ['image-analysis', 'ocr', 'visual-qa'],
          maxTokens: 2048
        }
      ]
    }
  ],
  defaultProvider: 'local',
  defaultModel: 'taishanglaojun-chat',
  fallbackProvider: 'openai',
  globalTimeout: 30000,
  maxConcurrentRequests: 10,
  enableCaching: true,
  cacheExpiry: 3600000, // 1小时
  enableMetrics: true,
  logLevel: 'info'
};

// 环境变量配置
export const getAIConfigFromEnv = (): Partial<AIServiceConfig> => {
  const config: Partial<AIServiceConfig> = {};
  
  // 从环境变量或本地存储获取配置
  if (typeof window !== 'undefined') {
    const savedConfig = localStorage.getItem('ai-service-config');
    if (savedConfig) {
      try {
        return JSON.parse(savedConfig);
      } catch (error) {
        console.warn('Failed to parse saved AI config:', error);
      }
    }
  }
  
  return config;
};

// 保存配置到本地存储
export const saveAIConfig = (config: AIServiceConfig): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('ai-service-config', JSON.stringify(config));
  }
};

// 合并配置
export const mergeAIConfig = (defaultConfig: AIServiceConfig, userConfig: Partial<AIServiceConfig>): AIServiceConfig => {
  return {
    ...defaultConfig,
    ...userConfig,
    providers: userConfig.providers || defaultConfig.providers
  };
};

// 获取最终配置
export const getAIConfig = (): AIServiceConfig => {
  const userConfig = getAIConfigFromEnv();
  return mergeAIConfig(defaultAIConfig, userConfig);
};