import aiService from './aiService';

export interface OptimizationSuggestion {
  id: string;
  originalText: string;
  optimizedText: string;
  improvements: string[];
  confidence: number;
  category: 'clarity' | 'specificity' | 'context' | 'grammar' | 'structure';
}

export interface OptimizationRequest {
  text: string;
  context?: string;
  targetAudience?: 'technical' | 'general' | 'academic' | 'casual';
  optimizationType?: 'clarity' | 'conciseness' | 'detail' | 'professional';
  language?: 'zh' | 'en';
}

export interface OptimizationResult {
  suggestions: OptimizationSuggestion[];
  bestSuggestion?: OptimizationSuggestion;
  originalLength: number;
  optimizedLength: number;
  improvementScore: number;
}

class InputOptimizer {
  private optimizationCache: Map<string, OptimizationResult> = new Map();
  private readonly CACHE_EXPIRY = 5 * 60 * 1000; // 5分钟缓存

  /**
   * 优化用户输入内容
   */
  async optimizeInput(request: OptimizationRequest): Promise<OptimizationResult> {
    const cacheKey = this.generateCacheKey(request);
    const cached = this.optimizationCache.get(cacheKey);
    
    if (cached) {
      return cached;
    }

    try {
      const prompt = this.buildOptimizationPrompt(request);
      
      const aiResponse = await aiService.sendChatMessage(prompt, 'input_optimization', {
        model: 'gpt-4',
        provider: 'openai',
        temperature: 0.3,
        maxTokens: 1000
      });

      if (aiResponse.success) {
        const result = this.parseOptimizationResponse(aiResponse.result.content, request.text);
        
        // 缓存结果
        this.optimizationCache.set(cacheKey, result);
        setTimeout(() => {
          this.optimizationCache.delete(cacheKey);
        }, this.CACHE_EXPIRY);

        return result;
      } else {
        throw new Error(aiResponse.error || '优化请求失败');
      }
    } catch (error) {
      console.error('Input optimization failed:', error);
      return this.createFallbackResult(request.text);
    }
  }

  /**
   * 快速优化 - 提供即时的基础优化建议
   */
  async quickOptimize(text: string): Promise<string[]> {
    const suggestions: string[] = [];

    // 基础语法和表达优化
    if (text.length < 10) {
      suggestions.push('建议提供更多细节以获得更好的回答');
    }

    if (!text.includes('?') && !text.includes('？') && !text.includes('请') && !text.includes('帮')) {
      suggestions.push('可以明确表达您的需求，如"请帮我..."或"如何..."');
    }

    if (text.includes('这个') || text.includes('那个')) {
      suggestions.push('建议具体说明"这个"或"那个"指的是什么');
    }

    if (text.split('').filter(char => char === '。' || char === '.' || char === '!' || char === '?').length === 0) {
      suggestions.push('建议添加适当的标点符号');
    }

    return suggestions;
  }

  /**
   * 智能补全建议
   */
  async getCompletionSuggestions(partialText: string, context?: string): Promise<string[]> {
    if (partialText.length < 3) {
      return [];
    }

    try {
      const prompt = `
基于以下部分输入，提供3-5个可能的补全建议：
输入: "${partialText}"
${context ? `上下文: ${context}` : ''}

请提供简洁、实用的补全建议，每行一个建议。
`;

      const aiResponse = await aiService.sendChatMessage(prompt, 'completion_suggestions', {
        model: 'gpt-3.5-turbo',
        provider: 'openai',
        temperature: 0.5,
        maxTokens: 200
      });

      if (aiResponse.success) {
        return aiResponse.result.content
          .split('\n')
          .filter((line: string) => line.trim())
          .slice(0, 5);
      }
    } catch (error) {
      console.error('Completion suggestions failed:', error);
    }

    return [];
  }

  /**
   * 检测输入意图
   */
  detectIntent(text: string): {
    intent: 'question' | 'request' | 'command' | 'conversation' | 'unclear';
    confidence: number;
    suggestions: string[];
  } {
    const questionWords = ['什么', '如何', '怎么', '为什么', '哪里', '谁', '何时', '?', '？'];
    const requestWords = ['请', '帮', '能否', '可以', '希望', '想要'];
    const commandWords = ['生成', '创建', '制作', '写', '画', '计算', '分析'];

    let intent: 'question' | 'request' | 'command' | 'conversation' | 'unclear' = 'unclear';
    let confidence = 0;
    const suggestions: string[] = [];

    if (questionWords.some(word => text.includes(word))) {
      intent = 'question';
      confidence = 0.8;
    } else if (requestWords.some(word => text.includes(word))) {
      intent = 'request';
      confidence = 0.7;
    } else if (commandWords.some(word => text.includes(word))) {
      intent = 'command';
      confidence = 0.6;
    } else if (text.length > 20) {
      intent = 'conversation';
      confidence = 0.5;
    }

    if (confidence < 0.6) {
      suggestions.push('建议明确表达您的需求类型（提问、请求帮助、或执行任务）');
    }

    return { intent, confidence, suggestions };
  }

  private buildOptimizationPrompt(request: OptimizationRequest): string {
    const { text, context, targetAudience, optimizationType, language } = request;
    
    return `
请优化以下用户输入，使其更清晰、准确和有效：

原始输入: "${text}"
${context ? `上下文: ${context}` : ''}
目标受众: ${targetAudience || '通用'}
优化类型: ${optimizationType || '清晰度'}
语言: ${language || 'zh'}

请提供：
1. 优化后的文本
2. 具体改进点
3. 改进原因
4. 置信度评分(0-1)

格式：
优化文本: [优化后的内容]
改进点: [具体改进说明]
置信度: [0-1的数值]
`;
  }

  private parseOptimizationResponse(response: string, originalText: string): OptimizationResult {
    const lines = response.split('\n').filter(line => line.trim());
    
    let optimizedText = originalText;
    let improvements: string[] = [];
    let confidence = 0.5;

    for (const line of lines) {
      if (line.includes('优化文本:')) {
        optimizedText = line.replace('优化文本:', '').trim();
      } else if (line.includes('改进点:')) {
        improvements.push(line.replace('改进点:', '').trim());
      } else if (line.includes('置信度:')) {
        const confidenceMatch = line.match(/[\d.]+/);
        if (confidenceMatch) {
          confidence = parseFloat(confidenceMatch[0]);
        }
      }
    }

    const suggestion: OptimizationSuggestion = {
      id: `opt_${Date.now()}`,
      originalText,
      optimizedText,
      improvements,
      confidence,
      category: 'clarity'
    };

    return {
      suggestions: [suggestion],
      bestSuggestion: suggestion,
      originalLength: originalText.length,
      optimizedLength: optimizedText.length,
      improvementScore: confidence
    };
  }

  private createFallbackResult(text: string): OptimizationResult {
    return {
      suggestions: [],
      originalLength: text.length,
      optimizedLength: text.length,
      improvementScore: 0
    };
  }

  private generateCacheKey(request: OptimizationRequest): string {
    return `${request.text}_${request.context || ''}_${request.targetAudience || ''}_${request.optimizationType || ''}`;
  }

  /**
   * 清理缓存
   */
  clearCache(): void {
    this.optimizationCache.clear();
  }

  /**
   * 获取优化统计
   */
  getStats(): {
    cacheSize: number;
    totalOptimizations: number;
  } {
    return {
      cacheSize: this.optimizationCache.size,
      totalOptimizations: this.optimizationCache.size
    };
  }
}

export default new InputOptimizer();