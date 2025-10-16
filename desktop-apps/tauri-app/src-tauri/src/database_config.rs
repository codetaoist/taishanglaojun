use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use sqlx::{SqlitePool, sqlite::SqliteConnectOptions};
use std::path::PathBuf;
use std::str::FromStr;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DatabaseConfig {
    pub main_db_path: PathBuf,
    pub chat_db_path: PathBuf,
    pub storage_db_path: PathBuf,
    pub max_connections: u32,
    pub connection_timeout: u64,
    pub enable_wal_mode: bool,
    pub enable_foreign_keys: bool,
    pub backup_enabled: bool,
    pub backup_interval_hours: u64,
}

impl Default for DatabaseConfig {
    fn default() -> Self {
        Self {
            main_db_path: PathBuf::from("taishang.db"),
            chat_db_path: PathBuf::from("chat.db"),
            storage_db_path: PathBuf::from("storage.db"),
            max_connections: 10,
            connection_timeout: 30,
            enable_wal_mode: true,
            enable_foreign_keys: true,
            backup_enabled: true,
            backup_interval_hours: 24,
        }
    }
}

pub struct DatabaseManager {
    config: DatabaseConfig,
    main_pool: Option<SqlitePool>,
    chat_pool: Option<SqlitePool>,
    storage_pool: Option<SqlitePool>,
}

impl DatabaseManager {
    pub fn new(config: DatabaseConfig) -> Self {
        Self {
            config,
            main_pool: None,
            chat_pool: None,
            storage_pool: None,
        }
    }

    pub fn with_app_data_dir(app_data_dir: PathBuf) -> Self {
        let config = DatabaseConfig {
            main_db_path: app_data_dir.join("taishang.db"),
            chat_db_path: app_data_dir.join("chat.db"),
            storage_db_path: app_data_dir.join("storage.db"),
            ..Default::default()
        };
        Self::new(config)
    }

    async fn create_connection_pool(&self, db_path: &PathBuf) -> Result<SqlitePool> {
        // 确保数据库目录存在
        if let Some(parent) = db_path.parent() {
            tokio::fs::create_dir_all(parent).await?;
        }

        let mut options = SqliteConnectOptions::from_str(&format!("sqlite:{}", db_path.display()))?
            .create_if_missing(true);

        if self.config.enable_foreign_keys {
            options = options.pragma("foreign_keys", "ON");
        }

        if self.config.enable_wal_mode {
            options = options.pragma("journal_mode", "WAL");
        }

        // 性能优化设置
        options = options
            .pragma("synchronous", "NORMAL")
            .pragma("cache_size", "10000")
            .pragma("temp_store", "memory")
            .pragma("mmap_size", "268435456"); // 256MB

        let pool = SqlitePool::connect_with(options).await?;
        
        Ok(pool)
    }

    pub async fn initialize(&mut self) -> Result<()> {
        // 初始化主数据库连接池
        self.main_pool = Some(self.create_connection_pool(&self.config.main_db_path).await?);
        
        // 初始化聊天数据库连接池
        self.chat_pool = Some(self.create_connection_pool(&self.config.chat_db_path).await?);
        
        // 初始化存储数据库连接池
        self.storage_pool = Some(self.create_connection_pool(&self.config.storage_db_path).await?);

        // 初始化数据库表结构
        self.initialize_schemas().await?;

        Ok(())
    }

    async fn initialize_schemas(&self) -> Result<()> {
        // 初始化主数据库表结构
        if let Some(pool) = &self.main_pool {
            self.initialize_main_db_schema(pool).await?;
        }

        // 初始化聊天数据库表结构
        if let Some(pool) = &self.chat_pool {
            self.initialize_chat_db_schema(pool).await?;
        }

        // 初始化存储数据库表结构
        if let Some(pool) = &self.storage_pool {
            self.initialize_storage_db_schema(pool).await?;
        }

        Ok(())
    }

    async fn initialize_main_db_schema(&self, pool: &SqlitePool) -> Result<()> {
        // 用户表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS users (
                id TEXT PRIMARY KEY,
                username TEXT UNIQUE NOT NULL,
                email TEXT UNIQUE NOT NULL,
                display_name TEXT,
                avatar_url TEXT,
                status TEXT DEFAULT 'active',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
            "#,
        )
        .execute(pool)
        .await?;

        // 项目表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS projects (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL,
                description TEXT,
                owner_id TEXT NOT NULL,
                status TEXT DEFAULT 'active',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (owner_id) REFERENCES users(id)
            )
            "#,
        )
        .execute(pool)
        .await?;

        // 好友表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS friends (
                id TEXT PRIMARY KEY,
                user_id TEXT NOT NULL,
                friend_id TEXT NOT NULL,
                status TEXT DEFAULT 'pending',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id),
                FOREIGN KEY (friend_id) REFERENCES users(id),
                UNIQUE(user_id, friend_id)
            )
            "#,
        )
        .execute(pool)
        .await?;

        // 设置表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS settings (
                id TEXT PRIMARY KEY,
                user_id TEXT,
                key TEXT NOT NULL,
                value TEXT,
                category TEXT DEFAULT 'general',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id)
            )
            "#,
        )
        .execute(pool)
        .await?;

        Ok(())
    }

    async fn initialize_chat_db_schema(&self, pool: &SqlitePool) -> Result<()> {
        // 聊天会话表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS chat_sessions (
                id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                chat_type TEXT DEFAULT 'general',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                message_count INTEGER DEFAULT 0
            )
            "#,
        )
        .execute(pool)
        .await?;

        // 聊天消息表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS chat_messages (
                id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                message_type TEXT DEFAULT 'text',
                metadata TEXT,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (session_id) REFERENCES chat_sessions(id)
            )
            "#,
        )
        .execute(pool)
        .await?;

        Ok(())
    }

    async fn initialize_storage_db_schema(&self, pool: &SqlitePool) -> Result<()> {
        // 文件信息表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS files (
                path TEXT PRIMARY KEY,
                name TEXT NOT NULL,
                size INTEGER NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                file_type TEXT,
                hash TEXT,
                backup_path TEXT
            )
            "#,
        )
        .execute(pool)
        .await?;

        // 备份记录表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS backups (
                id TEXT PRIMARY KEY,
                source_path TEXT NOT NULL,
                backup_path TEXT NOT NULL,
                backup_type TEXT DEFAULT 'full',
                status TEXT DEFAULT 'completed',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                size INTEGER
            )
            "#,
        )
        .execute(pool)
        .await?;

        Ok(())
    }

    pub fn get_main_pool(&self) -> Result<&SqlitePool> {
        self.main_pool.as_ref().ok_or_else(|| anyhow!("Main database pool not initialized"))
    }

    pub fn get_chat_pool(&self) -> Result<&SqlitePool> {
        self.chat_pool.as_ref().ok_or_else(|| anyhow!("Chat database pool not initialized"))
    }

    pub fn get_storage_pool(&self) -> Result<&SqlitePool> {
        self.storage_pool.as_ref().ok_or_else(|| anyhow!("Storage database pool not initialized"))
    }

    pub async fn health_check(&self) -> Result<DatabaseHealthStatus> {
        let mut status = DatabaseHealthStatus::default();

        // 检查主数据库
        if let Ok(pool) = self.get_main_pool() {
            match sqlx::query("SELECT 1").fetch_one(pool).await {
                Ok(_) => status.main_db_healthy = true,
                Err(e) => status.main_db_error = Some(e.to_string()),
            }
        }

        // 检查聊天数据库
        if let Ok(pool) = self.get_chat_pool() {
            match sqlx::query("SELECT 1").fetch_one(pool).await {
                Ok(_) => status.chat_db_healthy = true,
                Err(e) => status.chat_db_error = Some(e.to_string()),
            }
        }

        // 检查存储数据库
        if let Ok(pool) = self.get_storage_pool() {
            match sqlx::query("SELECT 1").fetch_one(pool).await {
                Ok(_) => status.storage_db_healthy = true,
                Err(e) => status.storage_db_error = Some(e.to_string()),
            }
        }

        Ok(status)
    }

    pub async fn close(&mut self) -> Result<()> {
        if let Some(pool) = self.main_pool.take() {
            pool.close().await;
        }
        if let Some(pool) = self.chat_pool.take() {
            pool.close().await;
        }
        if let Some(pool) = self.storage_pool.take() {
            pool.close().await;
        }
        Ok(())
    }
}

#[derive(Debug, Default, Serialize)]
pub struct DatabaseHealthStatus {
    pub main_db_healthy: bool,
    pub chat_db_healthy: bool,
    pub storage_db_healthy: bool,
    pub main_db_error: Option<String>,
    pub chat_db_error: Option<String>,
    pub storage_db_error: Option<String>,
}

// Tauri 命令
#[tauri::command]
pub async fn check_database_health(
    db_manager: tauri::State<'_, std::sync::Arc<tokio::sync::Mutex<DatabaseManager>>>,
) -> Result<DatabaseHealthStatus, String> {
    let manager = db_manager.lock().await;
    manager.health_check().await.map_err(|e| e.to_string())
}