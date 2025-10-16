use anyhow::{anyhow, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::time::Duration;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AIRequest {
    pub id: String,
    pub capability: String,
    pub request_type: String,
    pub input: Value,
    pub context: Value,
    pub requirements: Vec<String>,
    pub timeout: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AIResponse {
    pub request_id: String,
    pub success: bool,
    pub result: Value,
    pub confidence: f64,
    pub used_capabilities: Vec<String>,
    pub metadata: Value,
    pub error: Option<String>,
    pub process_time: Option<u64>,
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemStatus {
    pub active_requests: usize,
    pub total_requests: u64,
    pub success_rate: f64,
    pub avg_response_time: u64,
    pub overall_health: f64,
    pub last_updated: String,
}

pub struct AIService {
    client: Client,
    base_url: String,
    api_key: Option<String>,
}

impl AIService {
    pub async fn new() -> Result<Self> {
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .build()?;

        Ok(Self {
            client,
            base_url: "http://localhost:8080/api/v1".to_string(),
            api_key: None,
        })
    }

    pub fn set_api_key(&mut self, api_key: String) {
        self.api_key = Some(api_key);
    }

    pub fn set_base_url(&mut self, base_url: String) {
        self.base_url = base_url;
    }

    // 发送AI请求
    pub async fn send_request(&self, request: AIRequest) -> Result<AIResponse> {
        let url = format!("{}/ai/advanced/process", self.base_url);
        
        let mut req_builder = self.client.post(&url).json(&request);
        
        if let Some(api_key) = &self.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }

        let response = req_builder.send().await?;

        if response.status().is_success() {
            let ai_response: AIResponse = response.json().await?;
            Ok(ai_response)
        } else {
            let error_text = response.text().await?;
            Err(anyhow!("AI service error: {}", error_text))
        }
    }

    // 多模态处理
    pub async fn process_multimodal(&self, data: Value) -> Result<AIResponse> {
        let request = AIRequest {
            id: uuid::Uuid::new_v4().to_string(),
            capability: "multimodal".to_string(),
            request_type: "multimodal".to_string(),
            input: data,
            context: serde_json::json!({}),
            requirements: vec![],
            timeout: Some(30),
        };

        self.send_request(request).await
    }

    // 智能推理
    pub async fn reasoning(&self, query: String, premises: Vec<String>, reasoning_type: String) -> Result<AIResponse> {
        let request = AIRequest {
            id: uuid::Uuid::new_v4().to_string(),
            capability: "reasoning".to_string(),
            request_type: "reasoning".to_string(),
            input: serde_json::json!({
                "query": query,
                "premises": premises
            }),
            context: serde_json::json!({
                "reasoning_type": reasoning_type
            }),
            requirements: vec![],
            timeout: Some(30),
        };

        self.send_request(request).await
    }

    // NLP处理
    pub async fn process_nlp(&self, text: String, tasks: Vec<String>, language: Option<String>) -> Result<AIResponse> {
        let request = AIRequest {
            id: uuid::Uuid::new_v4().to_string(),
            capability: "nlp".to_string(),
            request_type: "nlp".to_string(),
            input: serde_json::json!({
                "text": text,
                "tasks": tasks
            }),
            context: serde_json::json!({
                "language": language.unwrap_or_else(|| "auto".to_string())
            }),
            requirements: vec![],
            timeout: Some(20),
        };

        self.send_request(request).await
    }

    // AGI处理
    pub async fn process_agi(&self, task_type: String, input: Value, context: Value) -> Result<AIResponse> {
        let request = AIRequest {
            id: uuid::Uuid::new_v4().to_string(),
            capability: "agi".to_string(),
            request_type: task_type,
            input,
            context,
            requirements: vec![],
            timeout: Some(60),
        };

        self.send_request(request).await
    }

    // 混合能力处理
    pub async fn process_hybrid(&self, input: Value, context: Value) -> Result<AIResponse> {
        let request = AIRequest {
            id: uuid::Uuid::new_v4().to_string(),
            capability: "hybrid".to_string(),
            request_type: "hybrid".to_string(),
            input,
            context,
            requirements: vec![],
            timeout: Some(90),
        };

        self.send_request(request).await
    }

    // 获取系统状态
    pub async fn get_status(&self) -> Result<Value> {
        let url = format!("{}/ai/advanced/status", self.base_url);
        
        let mut req_builder = self.client.get(&url);
        
        if let Some(api_key) = &self.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }

        let response = req_builder.send().await?;

        if response.status().is_success() {
            let status: Value = response.json().await?;
            Ok(status)
        } else {
            let error_text = response.text().await?;
            Err(anyhow!("Failed to get system status: {}", error_text))
        }
    }

    // 获取性能指标
    pub async fn get_performance_metrics(&self, limit: Option<usize>) -> Result<Value> {
        let url = format!("{}/ai/advanced/metrics", self.base_url);
        
        let mut req_builder = self.client.get(&url);
        
        if let Some(limit) = limit {
            req_builder = req_builder.query(&[("limit", limit)]);
        }
        
        if let Some(api_key) = &self.api_key {
            req_builder = req_builder.header("Authorization", format!("Bearer {}", api_key));
        }

        let response = req_builder.send().await?;

        if response.status().is_success() {
            let metrics: Value = response.json().await?;
            Ok(metrics)
        } else {
            let error_text = response.text().await?;
            Err(anyhow!("Failed to get performance metrics: {}", error_text))
        }
    }

    // 健康检查
    pub async fn health_check(&self) -> Result<bool> {
        let url = format!("{}/health", self.base_url);
        
        let response = self.client.get(&url).send().await?;
        Ok(response.status().is_success())
    }
}