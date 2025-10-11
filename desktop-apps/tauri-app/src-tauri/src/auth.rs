use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use crate::security::SecurityManager;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub username: String,
    pub email: Option<String>,
    pub display_name: String,
    pub avatar_url: Option<String>,
    pub permissions: Vec<String>,
    pub created_at: String,
    pub last_login: Option<String>,
    pub is_active: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoginRequest {
    pub username: String,
    pub password: String,
    pub remember_me: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoginResponse {
    pub success: bool,
    pub token: Option<String>,
    pub user: Option<User>,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterRequest {
    pub username: String,
    pub password: String,
    pub email: Option<String>,
    pub display_name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthConfig {
    pub enable_registration: bool,
    pub require_email_verification: bool,
    pub default_permissions: Vec<String>,
    pub admin_users: Vec<String>,
}

pub struct AuthManager {
    security_manager: SecurityManager,
    config: AuthConfig,
    users: HashMap<String, User>,
    current_user: Option<User>,
    current_token: Option<String>,
}

impl AuthManager {
    pub async fn new() -> Result<Self> {
        let security_manager = SecurityManager::new().await?;
        
        let config = AuthConfig {
            enable_registration: true,
            require_email_verification: false,
            default_permissions: vec![
                "chat".to_string(),
                "document_read".to_string(),
                "image_generate".to_string(),
            ],
            admin_users: vec!["admin".to_string()],
        };

        let mut auth_manager = Self {
            security_manager,
            config,
            users: HashMap::new(),
            current_user: None,
            current_token: None,
        };

        // 创建默认管理员用户
        auth_manager.create_default_admin().await?;

        Ok(auth_manager)
    }

    // 用户登录
    pub async fn login(&mut self, username: String, password: String) -> Result<String> {
        let token = self.security_manager.login(username.clone(), password).await?;
        
        // 获取用户信息
        if let Some(user) = self.users.get(&username) {
            self.current_user = Some(user.clone());
            self.current_token = Some(token.clone());
        }

        Ok(token)
    }

    // 用户登出
    pub async fn logout(&mut self) -> Result<()> {
        if let Some(token) = &self.current_token {
            self.security_manager.logout(token.clone()).await?;
        }
        
        self.current_user = None;
        self.current_token = None;
        
        Ok(())
    }

    // 用户注册
    pub async fn register(&mut self, request: RegisterRequest) -> Result<User> {
        if !self.config.enable_registration {
            return Err(anyhow!("注册功能已禁用"));
        }

        if self.users.contains_key(&request.username) {
            return Err(anyhow!("用户名已存在"));
        }

        // 确定用户权限
        let mut permissions = self.config.default_permissions.clone();
        if self.config.admin_users.contains(&request.username) {
            permissions.push("admin".to_string());
        }

        // 在安全管理器中注册用户
        self.security_manager.register_user(
            request.username.clone(),
            request.password,
            permissions.clone(),
        ).await?;

        // 创建用户对象
        let user = User {
            id: uuid::Uuid::new_v4().to_string(),
            username: request.username.clone(),
            email: request.email,
            display_name: request.display_name,
            avatar_url: None,
            permissions,
            created_at: chrono::Utc::now().to_rfc3339(),
            last_login: None,
            is_active: true,
        };

        self.users.insert(request.username, user.clone());
        Ok(user)
    }

    // 获取当前用户
    pub async fn get_current_user(&self) -> Option<User> {
        self.current_user.clone()
    }

    // 验证令牌
    pub async fn validate_token(&self, token: &str) -> Result<User> {
        let session = self.security_manager.validate_token(token).await?;
        
        if let Some(user) = self.users.get(&session.username) {
            Ok(user.clone())
        } else {
            Err(anyhow!("用户不存在"))
        }
    }

    // 检查权限
    pub async fn check_permission(&self, permission: &str) -> Result<bool> {
        if let Some(token) = &self.current_token {
            self.security_manager.check_permission(token, permission).await
        } else {
            Ok(false)
        }
    }

    // 更新用户信息
    pub async fn update_user_profile(
        &mut self,
        username: String,
        display_name: Option<String>,
        email: Option<String>,
        avatar_url: Option<String>,
    ) -> Result<User> {
        if let Some(user) = self.users.get_mut(&username) {
            if let Some(name) = display_name {
                user.display_name = name;
            }
            if let Some(email_addr) = email {
                user.email = Some(email_addr);
            }
            if let Some(avatar) = avatar_url {
                user.avatar_url = Some(avatar);
            }
            
            Ok(user.clone())
        } else {
            Err(anyhow!("用户不存在"))
        }
    }

    // 更改密码
    pub async fn change_password(
        &mut self,
        username: String,
        old_password: String,
        new_password: String,
    ) -> Result<()> {
        self.security_manager.change_password(username, old_password, new_password).await
    }

    // 获取用户列表（管理员功能）
    pub async fn get_users(&self) -> Result<Vec<User>> {
        if !self.check_permission("admin").await? {
            return Err(anyhow!("权限不足"));
        }

        Ok(self.users.values().cloned().collect())
    }

    // 禁用/启用用户（管理员功能）
    pub async fn set_user_active(&mut self, username: String, is_active: bool) -> Result<()> {
        if !self.check_permission("admin").await? {
            return Err(anyhow!("权限不足"));
        }

        if let Some(user) = self.users.get_mut(&username) {
            user.is_active = is_active;
            Ok(())
        } else {
            Err(anyhow!("用户不存在"))
        }
    }

    // 更新用户权限（管理员功能）
    pub async fn update_user_permissions(
        &mut self,
        username: String,
        permissions: Vec<String>,
    ) -> Result<()> {
        if !self.check_permission("admin").await? {
            return Err(anyhow!("权限不足"));
        }

        if let Some(user) = self.users.get_mut(&username) {
            user.permissions = permissions;
            Ok(())
        } else {
            Err(anyhow!("用户不存在"))
        }
    }

    // 删除用户（管理员功能）
    pub async fn delete_user(&mut self, username: String) -> Result<()> {
        if !self.check_permission("admin").await? {
            return Err(anyhow!("权限不足"));
        }

        if self.config.admin_users.contains(&username) {
            return Err(anyhow!("不能删除管理员用户"));
        }

        self.users.remove(&username);
        Ok(())
    }

    // 获取用户统计信息
    pub async fn get_user_stats(&self) -> Result<serde_json::Value> {
        let total_users = self.users.len();
        let active_users = self.users.values().filter(|u| u.is_active).count();
        let admin_users = self.users.values()
            .filter(|u| u.permissions.contains(&"admin".to_string()))
            .count();

        Ok(serde_json::json!({
            "total_users": total_users,
            "active_users": active_users,
            "admin_users": admin_users,
            "registration_enabled": self.config.enable_registration
        }))
    }

    // 刷新令牌
    pub async fn refresh_token(&mut self) -> Result<String> {
        if let Some(current_user) = &self.current_user {
            // 先登出当前会话
            if let Some(token) = &self.current_token {
                let _ = self.security_manager.logout(token.clone()).await;
            }

            // 重新生成令牌（这里简化处理，实际应该有专门的刷新机制）
            let username = current_user.username.clone();
            let permissions = current_user.permissions.clone();
            
            // 这里需要重新实现令牌刷新逻辑
            // 暂时返回错误，提示需要重新登录
            Err(anyhow!("令牌已过期，请重新登录"))
        } else {
            Err(anyhow!("未登录"))
        }
    }

    // 检查用户是否在线
    pub async fn is_user_online(&self, username: &str) -> bool {
        if let Some(current_user) = &self.current_user {
            current_user.username == username && self.current_token.is_some()
        } else {
            false
        }
    }

    // 获取在线用户列表
    pub async fn get_online_users(&self) -> Vec<String> {
        if let Some(current_user) = &self.current_user {
            vec![current_user.username.clone()]
        } else {
            vec![]
        }
    }

    // 私有方法

    async fn create_default_admin(&mut self) -> Result<()> {
        let admin_username = "admin".to_string();
        let admin_password = "admin123".to_string(); // 默认密码，应该在首次登录时强制更改

        if !self.users.contains_key(&admin_username) {
            let admin_permissions = vec![
                "admin".to_string(),
                "chat".to_string(),
                "document_read".to_string(),
                "document_write".to_string(),
                "image_generate".to_string(),
                "image_edit".to_string(),
                "user_management".to_string(),
                "system_management".to_string(),
            ];

            // 在安全管理器中注册管理员
            self.security_manager.register_user(
                admin_username.clone(),
                admin_password,
                admin_permissions.clone(),
            ).await?;

            // 创建管理员用户对象
            let admin_user = User {
                id: uuid::Uuid::new_v4().to_string(),
                username: admin_username.clone(),
                email: Some("admin@taishanglaojun.com".to_string()),
                display_name: "系统管理员".to_string(),
                avatar_url: None,
                permissions: admin_permissions,
                created_at: chrono::Utc::now().to_rfc3339(),
                last_login: None,
                is_active: true,
            };

            self.users.insert(admin_username, admin_user);
        }

        Ok(())
    }
}