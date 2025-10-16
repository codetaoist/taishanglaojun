use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::error::Error;
use std::fmt;
use regex::Regex;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizationRequest {
    pub text: String,
    pub target_audience: String,
    pub optimization_type: String,
    pub language: String,
    pub platform: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizationSuggestion {
    pub optimized_text: String,
    pub confidence: f64,
    pub improvements: Vec<String>,
    pub optimization_type: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizationResult {
    pub original_text: String,
    pub suggestions: Vec<OptimizationSuggestion>,
    pub best_suggestion: Option<OptimizationSuggestion>,
    pub processing_time_ms: u64,
}

#[derive(Debug)]
pub enum OptimizerError {
    PlatformNotSupported(String),
    OptimizationFailed(String),
    NetworkError(String),
    InvalidInput(String),
}

impl fmt::Display for OptimizerError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            OptimizerError::PlatformNotSupported(msg) => write!(f, "Platform not supported: {}", msg),
            OptimizerError::OptimizationFailed(msg) => write!(f, "Optimization failed: {}", msg),
            OptimizerError::NetworkError(msg) => write!(f, "Network error: {}", msg),
            OptimizerError::InvalidInput(msg) => write!(f, "Invalid input: {}", msg),
        }
    }
}

impl Error for OptimizerError {}

pub struct InputOptimizer {
    platform: String,
    optimization_cache: HashMap<String, OptimizationResult>,
    common_patterns: HashMap<String, String>,
}

impl InputOptimizer {
    pub async fn new() -> Result<Self, OptimizerError> {
        let platform = Self::detect_platform();
        let mut optimizer = Self {
            platform,
            optimization_cache: HashMap::new(),
            common_patterns: HashMap::new(),
        };
        
        optimizer.initialize_patterns().await?;
        Ok(optimizer)
    }

    fn detect_platform() -> String {
        #[cfg(target_os = "windows")]
        return "windows".to_string();
        
        #[cfg(target_os = "macos")]
        return "macos".to_string();
        
        #[cfg(target_os = "linux")]
        return "linux".to_string();
        
        #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
        return "unknown".to_string();
    }

    async fn initialize_patterns(&mut self) -> Result<(), OptimizerError> {
        // 初始化常见的优化模式
        self.common_patterns.insert(
            "unclear_question".to_string(),
            "请明确您的问题，提供更多上下文信息".to_string(),
        );
        self.common_patterns.insert(
            "vague_request".to_string(),
            "请具体说明您需要什么帮助".to_string(),
        );
        self.common_patterns.insert(
            "incomplete_sentence".to_string(),
            "请完整表达您的想法".to_string(),
        );
        
        Ok(())
    }

    pub async fn optimize_input(&mut self, request: OptimizationRequest) -> Result<OptimizationResult, OptimizerError> {
        let start_time = std::time::Instant::now();
        
        // 检查缓存
        if let Some(cached_result) = self.optimization_cache.get(&request.text) {
            return Ok(cached_result.clone());
        }

        // 验证输入
        if request.text.trim().is_empty() {
            return Err(OptimizerError::InvalidInput("Empty input text".to_string()));
        }

        let mut suggestions = Vec::new();

        // 基础优化
        if let Ok(basic_suggestion) = self.basic_optimization(&request.text).await {
            suggestions.push(basic_suggestion);
        }

        // 平台特定优化
        if let Ok(platform_suggestion) = self.platform_specific_optimization(&request).await {
            suggestions.push(platform_suggestion);
        }

        // 语言特定优化
        if let Ok(language_suggestion) = self.language_specific_optimization(&request).await {
            suggestions.push(language_suggestion);
        }

        // 选择最佳建议
        let best_suggestion = self.select_best_suggestion(&suggestions);

        let processing_time = start_time.elapsed().as_millis() as u64;

        let result = OptimizationResult {
            original_text: request.text.clone(),
            suggestions: suggestions.clone(),
            best_suggestion,
            processing_time_ms: processing_time,
        };

        // 缓存结果
        self.optimization_cache.insert(request.text, result.clone());

        Ok(result)
    }

    async fn basic_optimization(&self, text: &str) -> Result<OptimizationSuggestion, OptimizerError> {
        let mut optimized_text = text.to_string();
        let mut improvements = Vec::new();

        // 基础文本清理
        optimized_text = self.clean_text(&optimized_text);
        if optimized_text != text {
            improvements.push("清理了多余的空格和特殊字符".to_string());
        }

        // 语法检查和修正
        optimized_text = self.fix_grammar(&optimized_text);
        if optimized_text != text {
            improvements.push("修正了语法错误".to_string());
        }

        // 增强表达清晰度
        optimized_text = self.enhance_clarity(&optimized_text);
        if optimized_text != text {
            improvements.push("增强了表达的清晰度".to_string());
        }

        Ok(OptimizationSuggestion {
            optimized_text,
            confidence: 0.8,
            improvements,
            optimization_type: "basic".to_string(),
        })
    }

    async fn platform_specific_optimization(&self, request: &OptimizationRequest) -> Result<OptimizationSuggestion, OptimizerError> {
        match self.platform.as_str() {
            "windows" => self.optimize_for_windows(request).await,
            "macos" => self.optimize_for_macos(request).await,
            "linux" => self.optimize_for_linux(request).await,
            _ => Err(OptimizerError::PlatformNotSupported(self.platform.clone())),
        }
    }

    async fn optimize_for_windows(&self, request: &OptimizationRequest) -> Result<OptimizationSuggestion, OptimizerError> {
        let mut optimized_text = request.text.clone();
        let mut improvements = Vec::new();

        // Windows 特定的优化逻辑
        // 例如：处理 Windows 路径格式
        if optimized_text.contains("\\") {
            optimized_text = optimized_text.replace("\\", "/");
            improvements.push("统一了路径分隔符格式".to_string());
        }

        // 处理 Windows 特定的术语
        optimized_text = self.normalize_windows_terms(&optimized_text);
        if optimized_text != request.text {
            improvements.push("标准化了 Windows 相关术语".to_string());
        }

        Ok(OptimizationSuggestion {
            optimized_text,
            confidence: 0.9,
            improvements,
            optimization_type: "windows_specific".to_string(),
        })
    }

    async fn optimize_for_macos(&self, request: &OptimizationRequest) -> Result<OptimizationSuggestion, OptimizerError> {
        let mut optimized_text = request.text.clone();
        let mut improvements = Vec::new();

        // macOS 特定的优化逻辑
        // 处理 macOS 快捷键格式
        optimized_text = self.normalize_macos_shortcuts(&optimized_text);
        if optimized_text != request.text {
            improvements.push("标准化了 macOS 快捷键格式".to_string());
        }

        // 处理 macOS 特定术语
        optimized_text = self.normalize_macos_terms(&optimized_text);
        if optimized_text != request.text {
            improvements.push("标准化了 macOS 相关术语".to_string());
        }

        Ok(OptimizationSuggestion {
            optimized_text,
            confidence: 0.9,
            improvements,
            optimization_type: "macos_specific".to_string(),
        })
    }

    async fn optimize_for_linux(&self, request: &OptimizationRequest) -> Result<OptimizationSuggestion, OptimizerError> {
        let mut optimized_text = request.text.clone();
        let mut improvements = Vec::new();

        // Linux 特定的优化逻辑
        // 处理命令行相关内容
        optimized_text = self.normalize_linux_commands(&optimized_text);
        if optimized_text != request.text {
            improvements.push("标准化了 Linux 命令格式".to_string());
        }

        // 处理包管理器相关内容
        optimized_text = self.normalize_package_managers(&optimized_text);
        if optimized_text != request.text {
            improvements.push("标准化了包管理器命令".to_string());
        }

        Ok(OptimizationSuggestion {
            optimized_text,
            confidence: 0.85,
            improvements,
            optimization_type: "linux_specific".to_string(),
        })
    }

    async fn language_specific_optimization(&self, request: &OptimizationRequest) -> Result<OptimizationSuggestion, OptimizerError> {
        match request.language.as_str() {
            "zh" | "zh-CN" => self.optimize_chinese(&request.text).await,
            "en" | "en-US" => self.optimize_english(&request.text).await,
            _ => Err(OptimizerError::InvalidInput(format!("Unsupported language: {}", request.language))),
        }
    }

    async fn optimize_chinese(&self, text: &str) -> Result<OptimizationSuggestion, OptimizerError> {
        let mut optimized_text = text.to_string();
        let mut improvements = Vec::new();

        // 中文标点符号标准化
        optimized_text = self.normalize_chinese_punctuation(&optimized_text);
        if optimized_text != text {
            improvements.push("标准化了中文标点符号".to_string());
        }

        // 中文语序优化
        optimized_text = self.optimize_chinese_word_order(&optimized_text);
        if optimized_text != text {
            improvements.push("优化了中文语序".to_string());
        }

        Ok(OptimizationSuggestion {
            optimized_text,
            confidence: 0.85,
            improvements,
            optimization_type: "chinese_specific".to_string(),
        })
    }

    async fn optimize_english(&self, text: &str) -> Result<OptimizationSuggestion, OptimizerError> {
        let mut optimized_text = text.to_string();
        let mut improvements = Vec::new();

        // 英文语法检查
        optimized_text = self.check_english_grammar(&optimized_text);
        if optimized_text != text {
            improvements.push("修正了英文语法".to_string());
        }

        // 英文拼写检查
        optimized_text = self.check_english_spelling(&optimized_text);
        if optimized_text != text {
            improvements.push("修正了英文拼写".to_string());
        }

        Ok(OptimizationSuggestion {
            optimized_text,
            confidence: 0.9,
            improvements,
            optimization_type: "english_specific".to_string(),
        })
    }

    fn select_best_suggestion(&self, suggestions: &[OptimizationSuggestion]) -> Option<OptimizationSuggestion> {
        suggestions
            .iter()
            .max_by(|a, b| a.confidence.partial_cmp(&b.confidence).unwrap())
            .cloned()
    }

    // 辅助方法
    fn clean_text(&self, text: &str) -> String {
        // 清理多余的空格
        let re = Regex::new(r"\s+").unwrap();
        re.replace_all(text.trim(), " ").to_string()
    }

    fn fix_grammar(&self, text: &str) -> String {
        // 基础语法修正
        text.to_string()
    }

    fn enhance_clarity(&self, text: &str) -> String {
        // 增强清晰度
        text.to_string()
    }

    fn normalize_windows_terms(&self, text: &str) -> String {
        text.replace("文件夹", "目录")
            .replace("右键", "右键单击")
    }

    fn normalize_macos_shortcuts(&self, text: &str) -> String {
        text.replace("Ctrl", "Cmd")
            .replace("Alt", "Option")
    }

    fn normalize_macos_terms(&self, text: &str) -> String {
        text.replace("文件夹", "文件夹")
            .replace("右键", "右键点击")
    }

    fn normalize_linux_commands(&self, text: &str) -> String {
        // 标准化 Linux 命令格式
        text.to_string()
    }

    fn normalize_package_managers(&self, text: &str) -> String {
        // 标准化包管理器命令
        text.to_string()
    }

    fn normalize_chinese_punctuation(&self, text: &str) -> String {
        text.replace("，", "，")
            .replace("。", "。")
            .replace("？", "？")
            .replace("！", "！")
    }

    fn optimize_chinese_word_order(&self, text: &str) -> String {
        // 中文语序优化
        text.to_string()
    }

    fn check_english_grammar(&self, text: &str) -> String {
        // 英文语法检查
        text.to_string()
    }

    fn check_english_spelling(&self, text: &str) -> String {
        // 英文拼写检查
        text.to_string()
    }

    pub async fn get_quick_suggestions(&self, text: &str) -> Result<Vec<String>, OptimizerError> {
        let mut suggestions = Vec::new();

        if text.len() < 5 {
            suggestions.push("输入内容太短，请提供更多信息".to_string());
        }

        if !text.contains("?") && !text.contains("？") && text.len() > 20 {
            suggestions.push("考虑将长句分解为多个短句".to_string());
        }

        if text.chars().filter(|c| c.is_uppercase()).count() > text.len() / 2 {
            suggestions.push("避免过多使用大写字母".to_string());
        }

        Ok(suggestions)
    }

    pub fn detect_intent(&self, text: &str) -> serde_json::Value {
        let mut intent = "unknown";
        let mut confidence = 0.0;

        if text.contains("?") || text.contains("？") || text.starts_with("什么") || text.starts_with("如何") {
            intent = "question";
            confidence = 0.8;
        } else if text.contains("请") || text.contains("帮助") || text.contains("需要") {
            intent = "request";
            confidence = 0.7;
        } else if text.starts_with("创建") || text.starts_with("生成") || text.starts_with("制作") {
            intent = "command";
            confidence = 0.9;
        } else if text.len() > 10 {
            intent = "conversation";
            confidence = 0.6;
        }

        serde_json::json!({
            "intent": intent,
            "confidence": confidence
        })
    }
}