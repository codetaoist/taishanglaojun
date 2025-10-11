use anyhow::{anyhow, Result};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use sqlx::{Row, SqlitePool};
use std::path::PathBuf;

use crate::ai_service::AIService;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub id: String,
    pub session_id: String,
    pub role: String, // "user" or "assistant"
    pub content: String,
    pub message_type: String, // "text", "image", "file", etc.
    pub metadata: Value,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatSession {
    pub id: String,
    pub title: String,
    pub chat_type: String, // "general", "reasoning", "multimodal", etc.
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub message_count: i32,
}

pub struct ChatManager {
    db_pool: SqlitePool,
    ai_service: AIService,
}

impl ChatManager {
    pub async fn new() -> Result<Self> {
        // 获取应用数据目录
        let app_data_dir = dirs::data_dir()
            .ok_or_else(|| anyhow!("Failed to get app data directory"))?
            .join("taishang-laojun");
        
        std::fs::create_dir_all(&app_data_dir)?;
        
        let db_path = app_data_dir.join("chat.db");
        let db_url = format!("sqlite:{}", db_path.display());
        
        let db_pool = SqlitePool::connect(&db_url).await?;
        
        // 创建表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS chat_sessions (
                id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                chat_type TEXT NOT NULL,
                created_at TEXT NOT NULL,
                updated_at TEXT NOT NULL,
                message_count INTEGER DEFAULT 0
            )
            "#,
        )
        .execute(&db_pool)
        .await?;

        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS chat_messages (
                id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                message_type TEXT NOT NULL,
                metadata TEXT NOT NULL,
                created_at TEXT NOT NULL,
                FOREIGN KEY (session_id) REFERENCES chat_sessions (id)
            )
            "#,
        )
        .execute(&db_pool)
        .await?;

        let ai_service = AIService::new().await?;

        Ok(Self {
            db_pool,
            ai_service,
        })
    }

    // 创建新的聊天会话
    pub async fn create_session(&self, title: String, chat_type: String) -> Result<ChatSession> {
        let session = ChatSession {
            id: uuid::Uuid::new_v4().to_string(),
            title,
            chat_type,
            created_at: Utc::now(),
            updated_at: Utc::now(),
            message_count: 0,
        };

        sqlx::query(
            r#"
            INSERT INTO chat_sessions (id, title, chat_type, created_at, updated_at, message_count)
            VALUES (?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&session.id)
        .bind(&session.title)
        .bind(&session.chat_type)
        .bind(session.created_at.to_rfc3339())
        .bind(session.updated_at.to_rfc3339())
        .bind(session.message_count)
        .execute(&self.db_pool)
        .await?;

        Ok(session)
    }

    // 发送消息并获取AI回复
    pub async fn send_message(&self, message: String, chat_type: String) -> Result<String> {
        // 获取或创建会话
        let session = self.get_or_create_session(chat_type.clone()).await?;

        // 保存用户消息
        let user_message = ChatMessage {
            id: uuid::Uuid::new_v4().to_string(),
            session_id: session.id.clone(),
            role: "user".to_string(),
            content: message.clone(),
            message_type: "text".to_string(),
            metadata: serde_json::json!({}),
            created_at: Utc::now(),
        };

        self.save_message(&user_message).await?;

        // 根据聊天类型选择AI处理方式
        let ai_response = match chat_type.as_str() {
            "reasoning" => {
                self.ai_service
                    .reasoning(
                        message,
                        vec![], // 可以从历史消息中提取前提
                        "deductive".to_string(),
                    )
                    .await?
            }
            "multimodal" => {
                self.ai_service
                    .process_multimodal(serde_json::json!({
                        "text": message
                    }))
                    .await?
            }
            "nlp" => {
                self.ai_service
                    .process_nlp(
                        message,
                        vec!["sentiment".to_string(), "entities".to_string()],
                        None,
                    )
                    .await?
            }
            _ => {
                // 默认使用AGI处理
                self.ai_service
                    .process_agi(
                        "dialogue".to_string(),
                        serde_json::json!({
                            "message": message,
                            "context": self.get_recent_context(&session.id, 5).await?
                        }),
                        serde_json::json!({
                            "session_id": session.id,
                            "chat_type": chat_type
                        }),
                    )
                    .await?
            }
        };

        // 提取AI回复内容
        let ai_content = if ai_response.success {
            match &ai_response.result {
                Value::String(s) => s.clone(),
                Value::Object(obj) => {
                    if let Some(content) = obj.get("content") {
                        content.as_str().unwrap_or("AI回复解析失败").to_string()
                    } else if let Some(conclusion) = obj.get("conclusion") {
                        conclusion.as_str().unwrap_or("AI回复解析失败").to_string()
                    } else {
                        serde_json::to_string_pretty(&ai_response.result)?
                    }
                }
                _ => serde_json::to_string_pretty(&ai_response.result)?,
            }
        } else {
            format!("AI处理失败: {}", ai_response.error.unwrap_or_else(|| "未知错误".to_string()))
        };

        // 保存AI回复
        let ai_message = ChatMessage {
            id: uuid::Uuid::new_v4().to_string(),
            session_id: session.id.clone(),
            role: "assistant".to_string(),
            content: ai_content.clone(),
            message_type: "text".to_string(),
            metadata: serde_json::json!({
                "confidence": ai_response.confidence,
                "used_capabilities": ai_response.used_capabilities,
                "process_time": ai_response.process_time
            }),
            created_at: Utc::now(),
        };

        self.save_message(&ai_message).await?;

        // 更新会话
        self.update_session(&session.id).await?;

        Ok(ai_content)
    }

    // 获取聊天历史
    pub async fn get_history(&self, limit: usize) -> Result<Vec<Value>> {
        let rows = sqlx::query(
            r#"
            SELECT m.*, s.title, s.chat_type
            FROM chat_messages m
            JOIN chat_sessions s ON m.session_id = s.id
            ORDER BY m.created_at DESC
            LIMIT ?
            "#,
        )
        .bind(limit as i64)
        .fetch_all(&self.db_pool)
        .await?;

        let mut messages = Vec::new();
        for row in rows {
            let message = serde_json::json!({
                "id": row.get::<String, _>("id"),
                "session_id": row.get::<String, _>("session_id"),
                "role": row.get::<String, _>("role"),
                "content": row.get::<String, _>("content"),
                "message_type": row.get::<String, _>("message_type"),
                "metadata": serde_json::from_str::<Value>(&row.get::<String, _>("metadata")).unwrap_or_default(),
                "created_at": row.get::<String, _>("created_at"),
                "session_title": row.get::<String, _>("title"),
                "chat_type": row.get::<String, _>("chat_type")
            });
            messages.push(message);
        }

        Ok(messages)
    }

    // 获取会话列表
    pub async fn get_sessions(&self, limit: Option<usize>) -> Result<Vec<ChatSession>> {
        let limit = limit.unwrap_or(50);
        
        let rows = sqlx::query(
            r#"
            SELECT * FROM chat_sessions
            ORDER BY updated_at DESC
            LIMIT ?
            "#,
        )
        .bind(limit as i64)
        .fetch_all(&self.db_pool)
        .await?;

        let mut sessions = Vec::new();
        for row in rows {
            let session = ChatSession {
                id: row.get("id"),
                title: row.get("title"),
                chat_type: row.get("chat_type"),
                created_at: DateTime::parse_from_rfc3339(&row.get::<String, _>("created_at"))?.with_timezone(&Utc),
                updated_at: DateTime::parse_from_rfc3339(&row.get::<String, _>("updated_at"))?.with_timezone(&Utc),
                message_count: row.get("message_count"),
            };
            sessions.push(session);
        }

        Ok(sessions)
    }

    // 删除会话
    pub async fn delete_session(&self, session_id: String) -> Result<()> {
        // 删除消息
        sqlx::query("DELETE FROM chat_messages WHERE session_id = ?")
            .bind(&session_id)
            .execute(&self.db_pool)
            .await?;

        // 删除会话
        sqlx::query("DELETE FROM chat_sessions WHERE id = ?")
            .bind(&session_id)
            .execute(&self.db_pool)
            .await?;

        Ok(())
    }

    // 私有方法

    async fn get_or_create_session(&self, chat_type: String) -> Result<ChatSession> {
        // 尝试获取最近的会话
        let row = sqlx::query(
            r#"
            SELECT * FROM chat_sessions
            WHERE chat_type = ?
            ORDER BY updated_at DESC
            LIMIT 1
            "#,
        )
        .bind(&chat_type)
        .fetch_optional(&self.db_pool)
        .await?;

        if let Some(row) = row {
            Ok(ChatSession {
                id: row.get("id"),
                title: row.get("title"),
                chat_type: row.get("chat_type"),
                created_at: DateTime::parse_from_rfc3339(&row.get::<String, _>("created_at"))?.with_timezone(&Utc),
                updated_at: DateTime::parse_from_rfc3339(&row.get::<String, _>("updated_at"))?.with_timezone(&Utc),
                message_count: row.get("message_count"),
            })
        } else {
            // 创建新会话
            let title = match chat_type.as_str() {
                "reasoning" => "智能推理对话",
                "multimodal" => "多模态对话",
                "nlp" => "自然语言处理",
                _ => "AI对话",
            };
            self.create_session(title.to_string(), chat_type).await
        }
    }

    async fn save_message(&self, message: &ChatMessage) -> Result<()> {
        sqlx::query(
            r#"
            INSERT INTO chat_messages (id, session_id, role, content, message_type, metadata, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&message.id)
        .bind(&message.session_id)
        .bind(&message.role)
        .bind(&message.content)
        .bind(&message.message_type)
        .bind(serde_json::to_string(&message.metadata)?)
        .bind(message.created_at.to_rfc3339())
        .execute(&self.db_pool)
        .await?;

        Ok(())
    }

    async fn update_session(&self, session_id: &str) -> Result<()> {
        sqlx::query(
            r#"
            UPDATE chat_sessions
            SET updated_at = ?, message_count = message_count + 2
            WHERE id = ?
            "#,
        )
        .bind(Utc::now().to_rfc3339())
        .bind(session_id)
        .execute(&self.db_pool)
        .await?;

        Ok(())
    }

    async fn get_recent_context(&self, session_id: &str, limit: usize) -> Result<Vec<Value>> {
        let rows = sqlx::query(
            r#"
            SELECT role, content FROM chat_messages
            WHERE session_id = ?
            ORDER BY created_at DESC
            LIMIT ?
            "#,
        )
        .bind(session_id)
        .bind(limit as i64)
        .fetch_all(&self.db_pool)
        .await?;

        let mut context = Vec::new();
        for row in rows.into_iter().rev() {
            context.push(serde_json::json!({
                "role": row.get::<String, _>("role"),
                "content": row.get::<String, _>("content")
            }));
        }

        Ok(context)
    }
}