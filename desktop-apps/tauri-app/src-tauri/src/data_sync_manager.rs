use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use sqlx::SqlitePool;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DataSyncEvent {
    pub event_type: String,
    pub table_name: String,
    pub record_id: String,
    pub data: serde_json::Value,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone)]
pub struct DatabaseConnection {
    pub name: String,
    pub pool: SqlitePool,
    pub priority: i32, // 数据库优先级，用于冲突解决
}

pub struct DataSyncManager {
    databases: Arc<RwLock<HashMap<String, DatabaseConnection>>>,
    sync_queue: Arc<RwLock<Vec<DataSyncEvent>>>,
    is_syncing: Arc<RwLock<bool>>,
}

impl DataSyncManager {
    pub fn new() -> Self {
        Self {
            databases: Arc::new(RwLock::new(HashMap::new())),
            sync_queue: Arc::new(RwLock::new(Vec::new())),
            is_syncing: Arc::new(RwLock::new(false)),
        }
    }

    /// 注册数据库连接
    pub async fn register_database(&self, name: String, pool: SqlitePool, priority: i32) -> Result<()> {
        let mut databases = self.databases.write().await;
        databases.insert(name.clone(), DatabaseConnection {
            name,
            pool,
            priority,
        });
        Ok(())
    }

    /// 记录数据变更事件
    pub async fn record_change(&self, event: DataSyncEvent) -> Result<()> {
        let mut queue = self.sync_queue.write().await;
        queue.push(event);
        
        // 如果队列过长，触发同步
        if queue.len() > 100 {
            drop(queue);
            self.sync_data().await?;
        }
        
        Ok(())
    }

    /// 执行数据同步
    pub async fn sync_data(&self) -> Result<()> {
        let mut is_syncing = self.is_syncing.write().await;
        if *is_syncing {
            return Ok(()); // 已在同步中
        }
        *is_syncing = true;
        drop(is_syncing);

        let mut queue = self.sync_queue.write().await;
        let events = queue.drain(..).collect::<Vec<_>>();
        drop(queue);

        // 按时间戳排序事件
        let mut sorted_events = events;
        sorted_events.sort_by(|a, b| a.timestamp.cmp(&b.timestamp));

        // 处理每个事件
        for event in sorted_events {
            if let Err(e) = self.process_sync_event(event).await {
                eprintln!("同步事件处理失败: {}", e);
            }
        }

        let mut is_syncing = self.is_syncing.write().await;
        *is_syncing = false;

        Ok(())
    }

    /// 处理单个同步事件
    async fn process_sync_event(&self, event: DataSyncEvent) -> Result<()> {
        let databases = self.databases.read().await;
        
        match event.event_type.as_str() {
            "user_update" => {
                // 用户数据更新需要同步到聊天数据库
                if let Some(chat_db) = databases.get("chat") {
                    self.sync_user_to_chat(&event, &chat_db.pool).await?;
                }
            },
            "session_create" => {
                // 聊天会话创建需要更新用户活动状态
                if let Some(main_db) = databases.get("main") {
                    self.update_user_activity(&event, &main_db.pool).await?;
                }
            },
            "file_upload" => {
                // 文件上传需要同步到主数据库
                if let Some(main_db) = databases.get("main") {
                    self.sync_file_to_main(&event, &main_db.pool).await?;
                }
            },
            _ => {
                // 其他事件类型的处理
            }
        }

        Ok(())
    }

    /// 同步用户数据到聊天数据库
    async fn sync_user_to_chat(&self, event: &DataSyncEvent, chat_pool: &SqlitePool) -> Result<()> {
        // 这里实现用户数据到聊天数据库的同步逻辑
        // 例如：更新聊天会话中的用户信息
        Ok(())
    }

    /// 更新用户活动状态
    async fn update_user_activity(&self, event: &DataSyncEvent, main_pool: &SqlitePool) -> Result<()> {
        // 这里实现用户活动状态的更新逻辑
        Ok(())
    }

    /// 同步文件信息到主数据库
    async fn sync_file_to_main(&self, event: &DataSyncEvent, main_pool: &SqlitePool) -> Result<()> {
        // 这里实现文件信息到主数据库的同步逻辑
        Ok(())
    }

    /// 检查数据一致性
    pub async fn check_consistency(&self) -> Result<Vec<String>> {
        let mut inconsistencies = Vec::new();
        let databases = self.databases.read().await;

        // 检查用户数据一致性
        if let (Some(main_db), Some(chat_db)) = (databases.get("main"), databases.get("chat")) {
            let main_users = self.get_user_ids(&main_db.pool).await?;
            let chat_users = self.get_chat_user_ids(&chat_db.pool).await?;
            
            for user_id in &chat_users {
                if !main_users.contains(user_id) {
                    inconsistencies.push(format!("聊天数据库中存在未知用户: {}", user_id));
                }
            }
        }

        Ok(inconsistencies)
    }

    /// 获取主数据库中的用户ID列表
    async fn get_user_ids(&self, pool: &SqlitePool) -> Result<Vec<String>> {
        let rows = sqlx::query("SELECT id FROM users")
            .fetch_all(pool)
            .await?;
        
        Ok(rows.into_iter()
            .map(|row| row.get::<i32, _>("id").to_string())
            .collect())
    }

    /// 获取聊天数据库中涉及的用户ID列表
    async fn get_chat_user_ids(&self, pool: &SqlitePool) -> Result<Vec<String>> {
        // 这里需要根据实际的聊天数据库结构来实现
        // 假设聊天消息中有user_id字段
        Ok(Vec::new())
    }

    /// 修复数据不一致问题
    pub async fn repair_inconsistencies(&self) -> Result<()> {
        let inconsistencies = self.check_consistency().await?;
        
        for issue in inconsistencies {
            println!("修复数据不一致: {}", issue);
            // 这里实现具体的修复逻辑
        }

        Ok(())
    }
}

// Tauri 命令
#[tauri::command]
pub async fn sync_databases(
    sync_manager: tauri::State<'_, Arc<DataSyncManager>>,
) -> Result<(), String> {
    sync_manager.sync_data().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn check_data_consistency(
    sync_manager: tauri::State<'_, Arc<DataSyncManager>>,
) -> Result<Vec<String>, String> {
    sync_manager.check_consistency().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn repair_data_inconsistencies(
    sync_manager: tauri::State<'_, Arc<DataSyncManager>>,
) -> Result<(), String> {
    sync_manager.repair_inconsistencies().await.map_err(|e| e.to_string())
}