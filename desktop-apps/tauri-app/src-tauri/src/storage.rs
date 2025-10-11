use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::{Path, PathBuf};
use std::collections::HashMap;
use sqlx::{SqlitePool, Row};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileInfo {
    pub path: String,
    pub name: String,
    pub size: u64,
    pub created_at: String,
    pub modified_at: String,
    pub file_type: String,
    pub hash: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StorageConfig {
    pub base_path: String,
    pub max_file_size: u64,
    pub allowed_extensions: Vec<String>,
    pub auto_backup: bool,
    pub backup_interval: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackupInfo {
    pub id: String,
    pub original_path: String,
    pub backup_path: String,
    pub created_at: String,
    pub size: u64,
}

pub struct StorageManager {
    config: StorageConfig,
    db_pool: SqlitePool,
    file_cache: HashMap<String, FileInfo>,
}

impl StorageManager {
    pub async fn new() -> Result<Self> {
        let config = StorageConfig {
            base_path: "./data".to_string(),
            max_file_size: 100 * 1024 * 1024, // 100MB
            allowed_extensions: vec![
                "txt".to_string(),
                "md".to_string(),
                "json".to_string(),
                "csv".to_string(),
                "pdf".to_string(),
                "docx".to_string(),
                "xlsx".to_string(),
                "png".to_string(),
                "jpg".to_string(),
                "jpeg".to_string(),
                "gif".to_string(),
                "mp3".to_string(),
                "mp4".to_string(),
                "wav".to_string(),
            ],
            auto_backup: true,
            backup_interval: 3600, // 1小时
        };

        // 确保基础目录存在
        fs::create_dir_all(&config.base_path)?;
        fs::create_dir_all(format!("{}/backups", config.base_path))?;
        fs::create_dir_all(format!("{}/temp", config.base_path))?;

        // 初始化数据库
        let db_path = format!("{}/storage.db", config.base_path);
        let db_pool = SqlitePool::connect(&format!("sqlite:{}", db_path)).await?;
        
        // 创建表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS files (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                path TEXT UNIQUE NOT NULL,
                name TEXT NOT NULL,
                size INTEGER NOT NULL,
                created_at TEXT NOT NULL,
                modified_at TEXT NOT NULL,
                file_type TEXT NOT NULL,
                hash TEXT
            )
            "#,
        )
        .execute(&db_pool)
        .await?;

        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS backups (
                id TEXT PRIMARY KEY,
                original_path TEXT NOT NULL,
                backup_path TEXT NOT NULL,
                created_at TEXT NOT NULL,
                size INTEGER NOT NULL
            )
            "#,
        )
        .execute(&db_pool)
        .await?;

        Ok(Self {
            config,
            db_pool,
            file_cache: HashMap::new(),
        })
    }

    // 保存文件
    pub async fn save_file(&mut self, path: String, content: String) -> Result<FileInfo> {
        let full_path = self.get_full_path(&path)?;
        
        // 检查文件大小
        if content.len() as u64 > self.config.max_file_size {
            return Err(anyhow!("文件大小超过限制"));
        }

        // 检查文件扩展名
        if !self.is_allowed_extension(&path) {
            return Err(anyhow!("不支持的文件类型"));
        }

        // 确保目录存在
        if let Some(parent) = full_path.parent() {
            fs::create_dir_all(parent)?;
        }

        // 写入文件
        fs::write(&full_path, content)?;

        // 获取文件信息
        let file_info = self.get_file_info(&path).await?;

        // 保存到数据库
        self.save_file_info_to_db(&file_info).await?;

        // 更新缓存
        self.file_cache.insert(path.clone(), file_info.clone());

        // 自动备份
        if self.config.auto_backup {
            let _ = self.create_backup(&path).await;
        }

        Ok(file_info)
    }

    // 读取文件
    pub async fn read_file(&self, path: String) -> Result<String> {
        let full_path = self.get_full_path(&path)?;
        
        if !full_path.exists() {
            return Err(anyhow!("文件不存在: {}", path));
        }

        let content = fs::read_to_string(&full_path)?;
        Ok(content)
    }

    // 读取二进制文件
    pub async fn read_binary_file(&self, path: String) -> Result<Vec<u8>> {
        let full_path = self.get_full_path(&path)?;
        
        if !full_path.exists() {
            return Err(anyhow!("文件不存在: {}", path));
        }

        let content = fs::read(&full_path)?;
        Ok(content)
    }

    // 保存二进制文件
    pub async fn save_binary_file(&mut self, path: String, data: Vec<u8>) -> Result<FileInfo> {
        let full_path = self.get_full_path(&path)?;
        
        // 检查文件大小
        if data.len() as u64 > self.config.max_file_size {
            return Err(anyhow!("文件大小超过限制"));
        }

        // 检查文件扩展名
        if !self.is_allowed_extension(&path) {
            return Err(anyhow!("不支持的文件类型"));
        }

        // 确保目录存在
        if let Some(parent) = full_path.parent() {
            fs::create_dir_all(parent)?;
        }

        // 写入文件
        fs::write(&full_path, data)?;

        // 获取文件信息
        let file_info = self.get_file_info(&path).await?;

        // 保存到数据库
        self.save_file_info_to_db(&file_info).await?;

        // 更新缓存
        self.file_cache.insert(path.clone(), file_info.clone());

        Ok(file_info)
    }

    // 删除文件
    pub async fn delete_file(&mut self, path: String) -> Result<()> {
        let full_path = self.get_full_path(&path)?;
        
        if full_path.exists() {
            fs::remove_file(&full_path)?;
        }

        // 从数据库删除
        sqlx::query("DELETE FROM files WHERE path = ?")
            .bind(&path)
            .execute(&self.db_pool)
            .await?;

        // 从缓存删除
        self.file_cache.remove(&path);

        Ok(())
    }

    // 移动文件
    pub async fn move_file(&mut self, from_path: String, to_path: String) -> Result<FileInfo> {
        let from_full_path = self.get_full_path(&from_path)?;
        let to_full_path = self.get_full_path(&to_path)?;

        if !from_full_path.exists() {
            return Err(anyhow!("源文件不存在: {}", from_path));
        }

        // 确保目标目录存在
        if let Some(parent) = to_full_path.parent() {
            fs::create_dir_all(parent)?;
        }

        // 移动文件
        fs::rename(&from_full_path, &to_full_path)?;

        // 更新数据库
        sqlx::query("UPDATE files SET path = ? WHERE path = ?")
            .bind(&to_path)
            .bind(&from_path)
            .execute(&self.db_pool)
            .await?;

        // 更新缓存
        if let Some(file_info) = self.file_cache.remove(&from_path) {
            let mut new_file_info = file_info;
            new_file_info.path = to_path.clone();
            new_file_info.name = Path::new(&to_path)
                .file_name()
                .and_then(|n| n.to_str())
                .unwrap_or("unknown")
                .to_string();
            self.file_cache.insert(to_path, new_file_info.clone());
            Ok(new_file_info)
        } else {
            self.get_file_info(&to_path).await
        }
    }

    // 复制文件
    pub async fn copy_file(&mut self, from_path: String, to_path: String) -> Result<FileInfo> {
        let from_full_path = self.get_full_path(&from_path)?;
        let to_full_path = self.get_full_path(&to_path)?;

        if !from_full_path.exists() {
            return Err(anyhow!("源文件不存在: {}", from_path));
        }

        // 确保目标目录存在
        if let Some(parent) = to_full_path.parent() {
            fs::create_dir_all(parent)?;
        }

        // 复制文件
        fs::copy(&from_full_path, &to_full_path)?;

        // 获取文件信息
        let file_info = self.get_file_info(&to_path).await?;

        // 保存到数据库
        self.save_file_info_to_db(&file_info).await?;

        // 更新缓存
        self.file_cache.insert(to_path, file_info.clone());

        Ok(file_info)
    }

    // 列出目录文件
    pub async fn list_files(&self, directory: Option<String>) -> Result<Vec<FileInfo>> {
        let dir_path = if let Some(dir) = directory {
            self.get_full_path(&dir)?
        } else {
            PathBuf::from(&self.config.base_path)
        };

        if !dir_path.exists() || !dir_path.is_dir() {
            return Err(anyhow!("目录不存在或不是有效目录"));
        }

        let mut files = Vec::new();
        for entry in fs::read_dir(&dir_path)? {
            let entry = entry?;
            let path = entry.path();
            
            if path.is_file() {
                let relative_path = path.strip_prefix(&self.config.base_path)
                    .unwrap_or(&path)
                    .to_string_lossy()
                    .replace('\\', "/");
                
                let file_info = self.get_file_info(&relative_path).await?;
                files.push(file_info);
            }
        }

        Ok(files)
    }

    // 创建备份
    pub async fn create_backup(&self, path: String) -> Result<BackupInfo> {
        let full_path = self.get_full_path(&path)?;
        
        if !full_path.exists() {
            return Err(anyhow!("文件不存在: {}", path));
        }

        let backup_id = format!("backup_{}_{}", 
            chrono::Utc::now().timestamp(),
            uuid::Uuid::new_v4().to_string()[..8].to_string()
        );
        
        let backup_path = format!("{}/backups/{}", self.config.base_path, backup_id);
        
        // 复制文件到备份目录
        fs::copy(&full_path, &backup_path)?;
        
        let metadata = fs::metadata(&backup_path)?;
        let backup_info = BackupInfo {
            id: backup_id,
            original_path: path,
            backup_path,
            created_at: chrono::Utc::now().to_rfc3339(),
            size: metadata.len(),
        };

        // 保存备份信息到数据库
        sqlx::query(
            "INSERT INTO backups (id, original_path, backup_path, created_at, size) VALUES (?, ?, ?, ?, ?)"
        )
        .bind(&backup_info.id)
        .bind(&backup_info.original_path)
        .bind(&backup_info.backup_path)
        .bind(&backup_info.created_at)
        .bind(backup_info.size as i64)
        .execute(&self.db_pool)
        .await?;

        Ok(backup_info)
    }

    // 恢复备份
    pub async fn restore_backup(&mut self, backup_id: String) -> Result<String> {
        let row = sqlx::query("SELECT * FROM backups WHERE id = ?")
            .bind(&backup_id)
            .fetch_one(&self.db_pool)
            .await?;

        let original_path: String = row.get("original_path");
        let backup_path: String = row.get("backup_path");

        let full_original_path = self.get_full_path(&original_path)?;
        
        // 确保目标目录存在
        if let Some(parent) = full_original_path.parent() {
            fs::create_dir_all(parent)?;
        }

        // 复制备份文件到原位置
        fs::copy(&backup_path, &full_original_path)?;

        // 更新文件信息
        let file_info = self.get_file_info(&original_path).await?;
        self.save_file_info_to_db(&file_info).await?;
        self.file_cache.insert(original_path.clone(), file_info);

        Ok(original_path)
    }

    // 获取存储统计信息
    pub async fn get_storage_stats(&self) -> Result<serde_json::Value> {
        let total_files: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM files")
            .fetch_one(&self.db_pool)
            .await?;

        let total_size: Option<i64> = sqlx::query_scalar("SELECT SUM(size) FROM files")
            .fetch_one(&self.db_pool)
            .await?;

        let backup_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM backups")
            .fetch_one(&self.db_pool)
            .await?;

        let backup_size: Option<i64> = sqlx::query_scalar("SELECT SUM(size) FROM backups")
            .fetch_one(&self.db_pool)
            .await?;

        Ok(serde_json::json!({
            "total_files": total_files,
            "total_size": total_size.unwrap_or(0),
            "backup_count": backup_count,
            "backup_size": backup_size.unwrap_or(0),
            "base_path": self.config.base_path,
            "max_file_size": self.config.max_file_size
        }))
    }

    // 私有方法

    fn get_full_path(&self, relative_path: &str) -> Result<PathBuf> {
        let path = Path::new(&self.config.base_path).join(relative_path);
        Ok(path)
    }

    fn is_allowed_extension(&self, path: &str) -> bool {
        if let Some(extension) = Path::new(path).extension().and_then(|ext| ext.to_str()) {
            self.config.allowed_extensions.contains(&extension.to_lowercase())
        } else {
            false
        }
    }

    async fn get_file_info(&self, path: &str) -> Result<FileInfo> {
        let full_path = self.get_full_path(path)?;
        let metadata = fs::metadata(&full_path)?;
        
        let name = Path::new(path)
            .file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("unknown")
            .to_string();

        let file_type = Path::new(path)
            .extension()
            .and_then(|ext| ext.to_str())
            .unwrap_or("unknown")
            .to_string();

        Ok(FileInfo {
            path: path.to_string(),
            name,
            size: metadata.len(),
            created_at: chrono::Utc::now().to_rfc3339(),
            modified_at: chrono::Utc::now().to_rfc3339(),
            file_type,
            hash: None,
        })
    }

    async fn save_file_info_to_db(&self, file_info: &FileInfo) -> Result<()> {
        sqlx::query(
            r#"
            INSERT OR REPLACE INTO files 
            (path, name, size, created_at, modified_at, file_type, hash) 
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&file_info.path)
        .bind(&file_info.name)
        .bind(file_info.size as i64)
        .bind(&file_info.created_at)
        .bind(&file_info.modified_at)
        .bind(&file_info.file_type)
        .bind(&file_info.hash)
        .execute(&self.db_pool)
        .await?;

        Ok(())
    }
}