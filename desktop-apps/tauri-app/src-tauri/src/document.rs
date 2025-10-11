use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::fs;
use std::path::Path;

use crate::ai_service::AIService;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DocumentInfo {
    pub path: String,
    pub name: String,
    pub size: u64,
    pub format: String,
    pub created_at: String,
    pub modified_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessingResult {
    pub success: bool,
    pub result: Value,
    pub processing_time: u64,
    pub error: Option<String>,
}

pub struct DocumentProcessor {
    ai_service: AIService,
}

impl DocumentProcessor {
    pub async fn new() -> Result<Self> {
        let ai_service = AIService::new().await?;
        
        Ok(Self {
            ai_service,
        })
    }

    // 处理文档
    pub async fn process_file(&self, file_path: String, operation: String) -> Result<String> {
        let path = Path::new(&file_path);
        
        if !path.exists() {
            return Err(anyhow!("文件不存在: {}", file_path));
        }

        let content = self.read_file_content(&file_path).await?;
        let file_info = self.get_file_info(&file_path)?;

        match operation.as_str() {
            "summarize" => self.summarize_document(content, file_info).await,
            "extract_keywords" => self.extract_keywords(content).await,
            "analyze_sentiment" => self.analyze_sentiment(content).await,
            "translate" => self.translate_document(content).await,
            "extract_entities" => self.extract_entities(content).await,
            "generate_outline" => self.generate_outline(content).await,
            "check_grammar" => self.check_grammar(content).await,
            "improve_writing" => self.improve_writing(content).await,
            _ => Err(anyhow!("不支持的操作: {}", operation)),
        }
    }

    // 文档摘要
    async fn summarize_document(&self, content: String, file_info: DocumentInfo) -> Result<String> {
        let response = self.ai_service.process_nlp(
            content,
            vec!["summarization".to_string()],
            Some("auto".to_string()),
        ).await?;

        if response.success {
            if let Some(summary) = response.result.get("summary") {
                Ok(summary.as_str().unwrap_or("摘要生成失败").to_string())
            } else {
                Ok("摘要生成失败".to_string())
            }
        } else {
            Err(anyhow!("摘要生成失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 提取关键词
    async fn extract_keywords(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_nlp(
            content,
            vec!["keyword_extraction".to_string()],
            Some("auto".to_string()),
        ).await?;

        if response.success {
            if let Some(keywords) = response.result.get("keywords") {
                if let Some(keywords_array) = keywords.as_array() {
                    let keywords_str: Vec<String> = keywords_array
                        .iter()
                        .filter_map(|k| k.as_str())
                        .map(|s| s.to_string())
                        .collect();
                    Ok(keywords_str.join(", "))
                } else {
                    Ok(keywords.as_str().unwrap_or("关键词提取失败").to_string())
                }
            } else {
                Ok("关键词提取失败".to_string())
            }
        } else {
            Err(anyhow!("关键词提取失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 情感分析
    async fn analyze_sentiment(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_nlp(
            content,
            vec!["sentiment_analysis".to_string()],
            Some("auto".to_string()),
        ).await?;

        if response.success {
            if let Some(sentiment) = response.result.get("sentiment") {
                let sentiment_obj = sentiment.as_object().unwrap_or(&serde_json::Map::new());
                let label = sentiment_obj.get("label").and_then(|v| v.as_str()).unwrap_or("未知");
                let score = sentiment_obj.get("score").and_then(|v| v.as_f64()).unwrap_or(0.0);
                Ok(format!("情感倾向: {} (置信度: {:.2})", label, score))
            } else {
                Ok("情感分析失败".to_string())
            }
        } else {
            Err(anyhow!("情感分析失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 文档翻译
    async fn translate_document(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_nlp(
            content,
            vec!["translation".to_string()],
            Some("auto".to_string()),
        ).await?;

        if response.success {
            if let Some(translation) = response.result.get("translation") {
                Ok(translation.as_str().unwrap_or("翻译失败").to_string())
            } else {
                Ok("翻译失败".to_string())
            }
        } else {
            Err(anyhow!("翻译失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 实体提取
    async fn extract_entities(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_nlp(
            content,
            vec!["named_entity_recognition".to_string()],
            Some("auto".to_string()),
        ).await?;

        if response.success {
            if let Some(entities) = response.result.get("entities") {
                if let Some(entities_array) = entities.as_array() {
                    let mut result = String::new();
                    for entity in entities_array {
                        if let Some(entity_obj) = entity.as_object() {
                            let text = entity_obj.get("text").and_then(|v| v.as_str()).unwrap_or("");
                            let label = entity_obj.get("label").and_then(|v| v.as_str()).unwrap_or("");
                            result.push_str(&format!("{} ({})\n", text, label));
                        }
                    }
                    Ok(result)
                } else {
                    Ok("实体提取失败".to_string())
                }
            } else {
                Ok("实体提取失败".to_string())
            }
        } else {
            Err(anyhow!("实体提取失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 生成大纲
    async fn generate_outline(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_agi(
            "outline_generation".to_string(),
            serde_json::json!({
                "content": content,
                "task": "generate_outline"
            }),
            serde_json::json!({
                "format": "hierarchical",
                "max_levels": 3
            }),
        ).await?;

        if response.success {
            match &response.result {
                Value::String(s) => Ok(s.clone()),
                Value::Object(obj) => {
                    if let Some(outline) = obj.get("outline") {
                        Ok(outline.as_str().unwrap_or("大纲生成失败").to_string())
                    } else {
                        Ok(serde_json::to_string_pretty(&response.result)?)
                    }
                }
                _ => Ok(serde_json::to_string_pretty(&response.result)?),
            }
        } else {
            Err(anyhow!("大纲生成失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 语法检查
    async fn check_grammar(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_nlp(
            content,
            vec!["grammar_check".to_string()],
            Some("auto".to_string()),
        ).await?;

        if response.success {
            if let Some(grammar_result) = response.result.get("grammar") {
                Ok(serde_json::to_string_pretty(grammar_result)?)
            } else {
                Ok("语法检查完成，未发现问题".to_string())
            }
        } else {
            Err(anyhow!("语法检查失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 写作改进
    async fn improve_writing(&self, content: String) -> Result<String> {
        let response = self.ai_service.process_agi(
            "writing_improvement".to_string(),
            serde_json::json!({
                "content": content,
                "task": "improve_writing"
            }),
            serde_json::json!({
                "focus": ["clarity", "conciseness", "style"],
                "preserve_meaning": true
            }),
        ).await?;

        if response.success {
            match &response.result {
                Value::String(s) => Ok(s.clone()),
                Value::Object(obj) => {
                    if let Some(improved) = obj.get("improved_content") {
                        Ok(improved.as_str().unwrap_or("写作改进失败").to_string())
                    } else {
                        Ok(serde_json::to_string_pretty(&response.result)?)
                    }
                }
                _ => Ok(serde_json::to_string_pretty(&response.result)?),
            }
        } else {
            Err(anyhow!("写作改进失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 读取文件内容
    async fn read_file_content(&self, file_path: &str) -> Result<String> {
        let path = Path::new(file_path);
        let extension = path.extension()
            .and_then(|ext| ext.to_str())
            .unwrap_or("")
            .to_lowercase();

        match extension.as_str() {
            "txt" | "md" | "markdown" => {
                Ok(fs::read_to_string(file_path)?)
            }
            "pdf" => {
                // 这里应该使用PDF解析库，简化实现
                Err(anyhow!("PDF文件处理暂未实现"))
            }
            "docx" => {
                // 这里应该使用DOCX解析库，简化实现
                Err(anyhow!("DOCX文件处理暂未实现"))
            }
            "html" | "htm" => {
                let content = fs::read_to_string(file_path)?;
                // 简单的HTML标签移除
                let text = content
                    .replace("<br>", "\n")
                    .replace("<p>", "\n")
                    .replace("</p>", "\n");
                // 这里应该使用HTML解析库进行更好的处理
                Ok(text)
            }
            _ => {
                // 尝试作为文本文件读取
                match fs::read_to_string(file_path) {
                    Ok(content) => Ok(content),
                    Err(_) => Err(anyhow!("不支持的文件格式: {}", extension)),
                }
            }
        }
    }

    // 获取文件信息
    fn get_file_info(&self, file_path: &str) -> Result<DocumentInfo> {
        let path = Path::new(file_path);
        let metadata = fs::metadata(file_path)?;
        
        let name = path.file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("unknown")
            .to_string();
            
        let format = path.extension()
            .and_then(|ext| ext.to_str())
            .unwrap_or("unknown")
            .to_string();

        Ok(DocumentInfo {
            path: file_path.to_string(),
            name,
            size: metadata.len(),
            format,
            created_at: chrono::Utc::now().to_rfc3339(),
            modified_at: chrono::Utc::now().to_rfc3339(),
        })
    }

    // 支持的文件格式
    pub fn supported_formats() -> Vec<String> {
        vec![
            "txt".to_string(),
            "md".to_string(),
            "markdown".to_string(),
            "html".to_string(),
            "htm".to_string(),
            "pdf".to_string(),
            "docx".to_string(),
            "doc".to_string(),
        ]
    }

    // 检查文件格式是否支持
    pub fn is_supported_format(file_path: &str) -> bool {
        let path = Path::new(file_path);
        if let Some(extension) = path.extension().and_then(|ext| ext.to_str()) {
            Self::supported_formats().contains(&extension.to_lowercase())
        } else {
            false
        }
    }
}