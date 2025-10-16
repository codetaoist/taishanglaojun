use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use sqlx::{SqlitePool, Row};
use std::collections::HashMap;
use std::sync::Arc;

use crate::data_sync_manager::{DataSyncManager, DataSyncEvent};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub username: String,
    pub email: String,
    pub display_name: Option<String>,
    pub avatar_url: Option<String>,
    pub status: String,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatSessionWithMessages {
    pub session: crate::chat::ChatSession,
    pub messages: Vec<crate::chat::ChatMessage>,
    pub user_info: Option<User>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectWithFiles {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub owner: User,
    pub files: Vec<crate::storage::FileInfo>,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

pub struct DataAccessLayer {
    main_db: SqlitePool,
    chat_db: SqlitePool,
    storage_db: SqlitePool,
    sync_manager: Arc<DataSyncManager>,
}

impl DataAccessLayer {
    pub fn new(
        main_db: SqlitePool,
        chat_db: SqlitePool,
        storage_db: SqlitePool,
        sync_manager: Arc<DataSyncManager>,
    ) -> Self {
        Self {
            main_db,
            chat_db,
            storage_db,
            sync_manager,
        }
    }

    /// 获取用户完整信息（包含聊天统计）
    pub async fn get_user_with_stats(&self, user_id: &str) -> Result<Option<User>> {
        // 从主数据库获取用户基本信息
        let user_row = sqlx::query(
            "SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at 
             FROM users WHERE id = ?"
        )
        .bind(user_id)
        .fetch_optional(&self.main_db)
        .await?;

        if let Some(row) = user_row {
            let user = User {
                id: row.get("id"),
                username: row.get("username"),
                email: row.get("email"),
                display_name: row.get("display_name"),
                avatar_url: row.get("avatar_url"),
                status: row.get("status"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
            };

            // 可以在这里添加聊天统计信息
            // let chat_stats = self.get_user_chat_stats(user_id).await?;

            Ok(Some(user))
        } else {
            Ok(None)
        }
    }

    /// 创建用户（跨数据库事务）
    pub async fn create_user(&self, user: &User) -> Result<()> {
        // 开始事务处理
        let mut tx = self.main_db.begin().await?;

        // 在主数据库中创建用户
        sqlx::query(
            "INSERT INTO users (id, username, email, display_name, avatar_url, status, created_at, updated_at)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
        )
        .bind(&user.id)
        .bind(&user.username)
        .bind(&user.email)
        .bind(&user.display_name)
        .bind(&user.avatar_url)
        .bind(&user.status)
        .bind(&user.created_at)
        .bind(&user.updated_at)
        .execute(&mut *tx)
        .await?;

        // 提交主数据库事务
        tx.commit().await?;

        // 记录同步事件
        let sync_event = DataSyncEvent {
            event_type: "user_create".to_string(),
            table_name: "users".to_string(),
            record_id: user.id.clone(),
            data: serde_json::to_value(user)?,
            timestamp: chrono::Utc::now(),
        };

        self.sync_manager.record_change(sync_event).await?;

        Ok(())
    }

    /// 获取聊天会话及相关信息
    pub async fn get_chat_session_with_context(&self, session_id: &str) -> Result<Option<ChatSessionWithMessages>> {
        // 从聊天数据库获取会话信息
        let session_row = sqlx::query(
            "SELECT id, title, chat_type, created_at, updated_at, message_count 
             FROM chat_sessions WHERE id = ?"
        )
        .bind(session_id)
        .fetch_optional(&self.chat_db)
        .await?;

        if let Some(session_row) = session_row {
            let session = crate::chat::ChatSession {
                id: session_row.get("id"),
                title: session_row.get("title"),
                chat_type: session_row.get("chat_type"),
                created_at: session_row.get("created_at"),
                updated_at: session_row.get("updated_at"),
                message_count: session_row.get("message_count"),
            };

            // 获取消息列表
            let message_rows = sqlx::query(
                "SELECT id, session_id, role, content, message_type, metadata, created_at 
                 FROM chat_messages WHERE session_id = ? ORDER BY created_at ASC"
            )
            .bind(session_id)
            .fetch_all(&self.chat_db)
            .await?;

            let messages: Vec<crate::chat::ChatMessage> = message_rows
                .into_iter()
                .map(|row| crate::chat::ChatMessage {
                    id: row.get("id"),
                    session_id: row.get("session_id"),
                    role: row.get("role"),
                    content: row.get("content"),
                    message_type: row.get("message_type"),
                    metadata: row.get("metadata"),
                    created_at: row.get("created_at"),
                })
                .collect();

            // 如果需要用户信息，可以从主数据库获取
            // let user_info = self.get_user_with_stats(&user_id).await?;

            Ok(Some(ChatSessionWithMessages {
                session,
                messages,
                user_info: None, // 根据需要填充
            }))
        } else {
            Ok(None)
        }
    }

    /// 获取项目及其文件信息
    pub async fn get_project_with_files(&self, project_id: &str) -> Result<Option<ProjectWithFiles>> {
        // 从主数据库获取项目信息
        let project_row = sqlx::query(
            "SELECT p.id, p.name, p.description, p.created_at, p.owner_id,
                    u.username, u.email, u.display_name, u.avatar_url, u.status,
                    u.created_at as user_created_at, u.updated_at as user_updated_at
             FROM projects p 
             JOIN users u ON p.owner_id = u.id 
             WHERE p.id = ?"
        )
        .bind(project_id)
        .fetch_optional(&self.main_db)
        .await?;

        if let Some(row) = project_row {
            let owner = User {
                id: row.get("owner_id"),
                username: row.get("username"),
                email: row.get("email"),
                display_name: row.get("display_name"),
                avatar_url: row.get("avatar_url"),
                status: row.get("status"),
                created_at: row.get("user_created_at"),
                updated_at: row.get("user_updated_at"),
            };

            // 从存储数据库获取文件信息
            let file_rows = sqlx::query(
                "SELECT path, name, size, created_at, modified_at, file_type, hash 
                 FROM files WHERE path LIKE ?"
            )
            .bind(format!("%/projects/{}/%", project_id))
            .fetch_all(&self.storage_db)
            .await?;

            let files: Vec<crate::storage::FileInfo> = file_rows
                .into_iter()
                .map(|row| crate::storage::FileInfo {
                    path: row.get("path"),
                    name: row.get("name"),
                    size: row.get::<i64, _>("size") as u64,
                    created_at: row.get("created_at"),
                    modified_at: row.get("modified_at"),
                    file_type: row.get("file_type"),
                    hash: row.get("hash"),
                })
                .collect();

            Ok(Some(ProjectWithFiles {
                id: row.get("id"),
                name: row.get("name"),
                description: row.get("description"),
                owner,
                files,
                created_at: row.get("created_at"),
            }))
        } else {
            Ok(None)
        }
    }

    /// 执行跨数据库搜索
    pub async fn search_across_databases(&self, query: &str) -> Result<HashMap<String, Vec<serde_json::Value>>> {
        let mut results = HashMap::new();

        // 搜索用户
        let user_results = sqlx::query(
            "SELECT id, username, email, display_name FROM users 
             WHERE username LIKE ? OR email LIKE ? OR display_name LIKE ?"
        )
        .bind(format!("%{}%", query))
        .bind(format!("%{}%", query))
        .bind(format!("%{}%", query))
        .fetch_all(&self.main_db)
        .await?;

        results.insert("users".to_string(), 
            user_results.into_iter()
                .map(|row| serde_json::json!({
                    "id": row.get::<String, _>("id"),
                    "username": row.get::<String, _>("username"),
                    "email": row.get::<String, _>("email"),
                    "display_name": row.get::<Option<String>, _>("display_name"),
                }))
                .collect()
        );

        // 搜索聊天会话
        let chat_results = sqlx::query(
            "SELECT id, title, chat_type FROM chat_sessions 
             WHERE title LIKE ?"
        )
        .bind(format!("%{}%", query))
        .fetch_all(&self.chat_db)
        .await?;

        results.insert("chat_sessions".to_string(),
            chat_results.into_iter()
                .map(|row| serde_json::json!({
                    "id": row.get::<String, _>("id"),
                    "title": row.get::<String, _>("title"),
                    "chat_type": row.get::<String, _>("chat_type"),
                }))
                .collect()
        );

        // 搜索文件
        let file_results = sqlx::query(
            "SELECT path, name, file_type FROM files 
             WHERE name LIKE ?"
        )
        .bind(format!("%{}%", query))
        .fetch_all(&self.storage_db)
        .await?;

        results.insert("files".to_string(),
            file_results.into_iter()
                .map(|row| serde_json::json!({
                    "path": row.get::<String, _>("path"),
                    "name": row.get::<String, _>("name"),
                    "file_type": row.get::<String, _>("file_type"),
                }))
                .collect()
        );

        Ok(results)
    }

    /// 获取数据库统计信息
    pub async fn get_database_stats(&self) -> Result<HashMap<String, serde_json::Value>> {
        let mut stats = HashMap::new();

        // 主数据库统计
        let user_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM users")
            .fetch_one(&self.main_db).await?;
        let project_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM projects")
            .fetch_one(&self.main_db).await?;

        stats.insert("main_db".to_string(), serde_json::json!({
            "users": user_count,
            "projects": project_count,
        }));

        // 聊天数据库统计
        let session_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM chat_sessions")
            .fetch_one(&self.chat_db).await?;
        let message_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM chat_messages")
            .fetch_one(&self.chat_db).await?;

        stats.insert("chat_db".to_string(), serde_json::json!({
            "sessions": session_count,
            "messages": message_count,
        }));

        // 存储数据库统计
        let file_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM files")
            .fetch_one(&self.storage_db).await?;
        let total_size: Option<i64> = sqlx::query_scalar("SELECT SUM(size) FROM files")
            .fetch_one(&self.storage_db).await?;

        stats.insert("storage_db".to_string(), serde_json::json!({
            "files": file_count,
            "total_size": total_size.unwrap_or(0),
        }));

        Ok(stats)
    }
}

// Tauri 命令
#[tauri::command]
pub async fn get_user_with_stats(
    dal: tauri::State<'_, Arc<DataAccessLayer>>,
    user_id: String,
) -> Result<Option<User>, String> {
    dal.get_user_with_stats(&user_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn search_all_data(
    dal: tauri::State<'_, Arc<DataAccessLayer>>,
    query: String,
) -> Result<HashMap<String, Vec<serde_json::Value>>, String> {
    dal.search_across_databases(&query).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_database_statistics(
    dal: tauri::State<'_, Arc<DataAccessLayer>>,
) -> Result<HashMap<String, serde_json::Value>, String> {
    dal.get_database_stats().await.map_err(|e| e.to_string())
}