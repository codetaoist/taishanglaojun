import React, { useState, useEffect, useRef } from 'react';
import { 
  Wand2, 
  Lightbulb, 
  RefreshCw, 
  Sparkles,
  MessageSquare,
  ArrowRight,
  Clock,
  Target,
  Keyboard
} from 'lucide-react';
import { cn } from '../utils/cn';
import { aiService } from '../services/aiService';
// import { OptimizationResult, OptimizationSuggestion } from '../services/inputOptimizer';
import { platformManager, formatShortcut } from '../utils/platform';

interface InputOptimizerProps {
  value: string;
  onChange: (value: string) => void;
  onOptimize?: (optimizedText: string) => void;
  onSend?: () => void;
  placeholder?: string;
  disabled?: boolean;
  loading?: boolean;
  className?: string;
  showQuickSuggestions?: boolean;
  showIntentDetection?: boolean;
}

export default function InputOptimizer({
  value,
  onChange,
  onOptimize,
  onSend,
  placeholder = "输入您的消息...",
  disabled = false,
  loading = false,
  className = "",
  showQuickSuggestions = true,
  showIntentDetection = true
}: InputOptimizerProps) {
  const [isOptimizing, setIsOptimizing] = useState(false);
  const [quickSuggestions, setQuickSuggestions] = useState<string[]>([]);
  const [completionSuggestions, setCompletionSuggestions] = useState<string[]>([]);
  const [intentInfo, setIntentInfo] = useState<any>(null);
  const [platformShortcuts, setPlatformShortcuts] = useState<Record<string, string>>({});
  
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const optimizationTimeoutRef = useRef<NodeJS.Timeout>();
  const completionTimeoutRef = useRef<NodeJS.Timeout>();

  // 初始化平台功能
  useEffect(() => {
    const initializePlatform = async () => {
      try {
        await platformManager.initialize();
        const shortcuts = platformManager.getPlatformShortcuts();
        setPlatformShortcuts(shortcuts);
      } catch (error) {
        console.error('Failed to initialize platform manager:', error);
        // 使用默认快捷键
        setPlatformShortcuts({
          'optimize': 'Ctrl+Enter',
          'clear': 'Ctrl+L',
          'submit': 'Enter'
        });
      }
    };
    
    initializePlatform();
  }, []);

  // 自动调整文本框高度
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  }, [value]);

  // 自动优化建议
  useEffect(() => {
    if (value.trim() && showQuickSuggestions) {
      if (optimizationTimeoutRef.current) {
        clearTimeout(optimizationTimeoutRef.current);
      }
      
      optimizationTimeoutRef.current = setTimeout(async () => {
        try {
          const suggestions = await aiService.getQuickSuggestions(value);
          setQuickSuggestions(suggestions);
          
          if (showIntentDetection) {
            const intent = await aiService.detectInputIntent(value);
            setIntentInfo(intent);
          }
        } catch (error) {
          console.error('获取优化建议失败:', error);
        }
      }, 500);
    } else {
      setQuickSuggestions([]);
      setIntentInfo(null);
    }

    return () => {
      if (optimizationTimeoutRef.current) {
        clearTimeout(optimizationTimeoutRef.current);
      }
    };
  }, [value, showQuickSuggestions, showIntentDetection]);

  // 智能补全建议
  useEffect(() => {
    if (value.length > 3) {
      if (completionTimeoutRef.current) {
        clearTimeout(completionTimeoutRef.current);
      }
      
      completionTimeoutRef.current = setTimeout(async () => {
        try {
          const suggestions = await aiService.getQuickSuggestions(value);
          setCompletionSuggestions(suggestions);
        } catch (error) {
          console.error('Completion suggestions failed:', error);
        }
      }, 500);
    } else {
      setCompletionSuggestions([]);
    }

    return () => {
      if (completionTimeoutRef.current) {
        clearTimeout(completionTimeoutRef.current);
      }
    };
  }, [value]);

  const handleOptimize = async () => {
    if (!value.trim() || isOptimizing) return;

    setIsOptimizing(true);
    try {
      // 实现智能优化算法
      const optimizedText = await optimizeInputContent(value);
      
      // 直接更新输入框内容
      onChange(optimizedText);
      onOptimize?.(optimizedText);
    } catch (error) {
      console.error('Optimization failed:', error);
    } finally {
      setIsOptimizing(false);
    }
  };

  // 智能输入内容优化算法
  const optimizeInputContent = async (text: string): Promise<string> => {
    // 1. 基础文本清理
    let optimized = text.trim();
    
    // 2. 标点符号优化
    optimized = optimized
      .replace(/[，。！？；：]/g, (match) => {
        // 中文标点后添加适当空格
        return match + ' ';
      })
      .replace(/\s+/g, ' ') // 多个空格合并为一个
      .replace(/\s+([，。！？；：])/g, '$1'); // 标点前不要空格
    
    // 3. 语法优化
    optimized = optimized
      .replace(/([a-zA-Z])\s*([，。！？；：])/g, '$1$2') // 英文字母和标点之间不要空格
      .replace(/([，。！？；：])\s*([a-zA-Z])/g, '$1 $2') // 标点和英文字母之间要空格
      .replace(/([0-9])\s*([，。！？；：])/g, '$1$2') // 数字和标点之间不要空格
      .replace(/([，。！？；：])\s*([0-9])/g, '$1 $2'); // 标点和数字之间要空格
    
    // 4. 语义优化
    const semanticOptimizations = [
      // 常见口语化表达优化
      { pattern: /能不能/g, replacement: '是否可以' },
      { pattern: /怎么样/g, replacement: '如何' },
      { pattern: /有没有/g, replacement: '是否有' },
      { pattern: /好不好/g, replacement: '是否合适' },
      
      // 敬语优化
      { pattern: /请问/g, replacement: '请问' },
      { pattern: /麻烦/g, replacement: '请' },
      { pattern: /帮忙/g, replacement: '协助' },
      
      // 专业术语优化
      { pattern: /AI/gi, replacement: 'AI' },
      { pattern: /api/gi, replacement: 'API' },
      { pattern: /ui/gi, replacement: 'UI' },
      { pattern: /ux/gi, replacement: 'UX' },
    ];
    
    semanticOptimizations.forEach(({ pattern, replacement }) => {
      optimized = optimized.replace(pattern, replacement);
    });
    
    // 5. 长度和结构优化
    if (optimized.length > 200) {
      // 长文本分段
      const sentences = optimized.split(/[。！？]/).filter(s => s.trim());
      if (sentences.length > 1) {
        optimized = sentences.map(s => s.trim()).join('。\n') + '。';
      }
    }
    
    // 6. 最终清理
    optimized = optimized
      .replace(/\s+$/, '') // 移除末尾空格
      .replace(/^\s+/, '') // 移除开头空格
      .replace(/\n\s*\n/g, '\n'); // 多个换行合并
    
    return optimized;
  };



  const applyCompletionSuggestion = (suggestion: string) => {
    onChange(suggestion);
    setCompletionSuggestions([]);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    const isMac = platformManager.isMacOS();
    const modifierKey = isMac ? e.metaKey : e.ctrlKey;
    
    // Enter 发送消息
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (value.trim() && !loading && !disabled && onSend) {
        onSend();
      }
    }
    
    // Ctrl/Cmd + Enter 也可以发送消息
    if (e.key === 'Enter' && modifierKey) {
      e.preventDefault();
      if (value.trim() && !loading && !disabled && onSend) {
        onSend();
      }
    }
    
    // Ctrl/Cmd + Shift + O 优化输入
    if (e.key === 'O' && modifierKey && e.shiftKey) {
      e.preventDefault();
      handleOptimize();
    }
    
    // Ctrl/Cmd + L 清空输入
    if (e.key === 'L' || e.key === 'l') {
      if (modifierKey) {
        e.preventDefault();
        onChange('');
      }
    }
  };

  const handleSend = () => {
    if (value.trim() && !loading && !disabled && onSend) {
      onSend();
    }
  };

  const getIntentIcon = (intent: string) => {
    switch (intent) {
      case 'question':
        return <MessageSquare className="h-4 w-4 text-blue-500" />;
      case 'request':
        return <Target className="h-4 w-4 text-green-500" />;
      case 'command':
        return <ArrowRight className="h-4 w-4 text-purple-500" />;
      case 'conversation':
        return <Clock className="h-4 w-4 text-orange-500" />;
      default:
        return <Lightbulb className="h-4 w-4 text-gray-500" />;
    }
  };

  const getIntentLabel = (intent: string) => {
    switch (intent) {
      case 'question':
        return '提问';
      case 'request':
        return '请求';
      case 'command':
        return '指令';
      case 'conversation':
        return '对话';
      default:
        return '不明确';
    }
  };

  return (
    <div className={cn("relative", className)}>
      {/* 主输入区域 */}
      <div className="flex space-x-2">
        <div className="relative flex-1">
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder={placeholder}
            disabled={disabled || loading}
            className={cn(
              "w-full min-h-[44px] max-h-29 p-3 border border-border rounded-lg resize-none",
              "focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent",
              "bg-background text-foreground placeholder-muted-foreground",
              (disabled || loading) && "opacity-50 cursor-not-allowed"
            )}
            rows={1}
          />
        </div>

        {/* 发送按钮 */}
        {onSend && (
          <button
            onClick={handleSend}
            disabled={!value.trim() || loading || disabled}
            className={cn(
              "btn-primary px-4 flex-shrink-0",
              (!value.trim() || loading || disabled) && "opacity-50 cursor-not-allowed"
            )}
          >
            {loading ? (
              <RefreshCw className="h-4 w-4 animate-spin" />
            ) : (
              <ArrowRight className="h-4 w-4" />
            )}
          </button>
        )}
      </div>

      {/* 优化按钮 - 移到输入框下方左侧对齐 */}
      <div className="mt-2 flex justify-start">
        <button
          onClick={handleOptimize}
          disabled={!value.trim() || isOptimizing || disabled || loading}
          className={cn(
            "px-3 py-1.5 rounded-lg transition-all duration-200 text-sm",
            "bg-gradient-to-r from-purple-500/10 to-blue-500/10 hover:from-purple-500/20 hover:to-blue-500/20",
            "border border-purple-300/30 hover:border-purple-400/50",
            "hover:scale-105 hover:shadow-md focus:outline-none focus:ring-2 focus:ring-purple-400/50",
            "group relative overflow-hidden flex items-center space-x-2",
            (!value.trim() || isOptimizing || disabled || loading) && "opacity-50 cursor-not-allowed hover:scale-100"
          )}
          title="优化输入内容 (Ctrl+Shift+O)"
        >
          <div className="absolute inset-0 bg-gradient-to-r from-purple-400/20 to-blue-400/20 opacity-0 group-hover:opacity-100 transition-opacity duration-200"></div>
          {isOptimizing ? (
            <RefreshCw className="h-4 w-4 animate-spin text-purple-600 relative z-10" />
          ) : (
            <Wand2 className="h-4 w-4 text-purple-600 group-hover:text-purple-700 relative z-10 transition-colors duration-200" />
          )}
          <span className="text-purple-600 group-hover:text-purple-700 relative z-10 transition-colors duration-200">
            {isOptimizing ? '优化中...' : '优化输入内容'}
          </span>
        </button>
      </div>

      {/* 意图检测显示 */}
      {intentInfo && intentInfo.confidence > 0.5 && (
        <div className="mt-2 flex items-center justify-between">
          <div className="flex items-center space-x-2 text-sm">
            {getIntentIcon(intentInfo.intent)}
            <span className="text-muted-foreground">
              检测到: {getIntentLabel(intentInfo.intent)}
            </span>
            <span className="text-xs bg-secondary px-2 py-1 rounded">
              {Math.round(intentInfo.confidence * 100)}%
            </span>
          </div>
          
          {/* 快捷键提示 */}
          <div className="flex items-center space-x-2 text-xs text-muted-foreground">
            <Keyboard className="h-3 w-3" />
            <span>{formatShortcut(platformShortcuts.optimize || 'Ctrl+Shift+O')} 优化</span>
            <span>•</span>
            <span>{formatShortcut(platformShortcuts.send || 'Enter')} 发送</span>
          </div>
        </div>
      )}

      {/* 仅快捷键提示（当没有意图检测时） */}
      {(!intentInfo || intentInfo.confidence <= 0.5) && value.length > 0 && (
        <div className="mt-2 flex justify-end">
          <div className="flex items-center space-x-2 text-xs text-muted-foreground">
            <Keyboard className="h-3 w-3" />
            <span>{formatShortcut(platformShortcuts.optimize || 'Ctrl+Shift+O')} 优化</span>
            <span>•</span>
            <span>{formatShortcut(platformShortcuts.send || 'Enter')} 发送</span>
          </div>
        </div>
      )}

      {/* 快速建议 */}
      {quickSuggestions.length > 0 && (
        <div className="mt-2 space-y-1">
          <div className="flex items-center space-x-1 text-xs text-muted-foreground">
            <Lightbulb className="h-3 w-3" />
            <span>建议:</span>
          </div>
          {quickSuggestions.map((suggestion, index) => (
            <div
              key={index}
              className="text-xs text-amber-600 bg-amber-50 px-2 py-1 rounded border-l-2 border-amber-200"
            >
              {suggestion}
            </div>
          ))}
        </div>
      )}

      {/* 智能补全建议 */}
      {completionSuggestions.length > 0 && (
        <div className="mt-2 space-y-1">
          <div className="flex items-center space-x-1 text-xs text-muted-foreground">
            <Sparkles className="h-3 w-3" />
            <span>补全建议:</span>
          </div>
          <div className="space-y-1">
            {completionSuggestions.map((suggestion, index) => (
              <button
                key={index}
                onClick={() => applyCompletionSuggestion(suggestion)}
                className="block w-full text-left text-sm p-2 bg-secondary hover:bg-secondary/80 rounded border transition-colors"
              >
                {suggestion}
              </button>
            ))}
          </div>
        </div>
      )}


    </div>
  );
}