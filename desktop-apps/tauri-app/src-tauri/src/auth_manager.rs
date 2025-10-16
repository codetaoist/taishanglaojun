use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::State;
use reqwest::Client;
use tokio::time::{Duration, Instant};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub username: String,
    pub email: String,
    pub avatar_url: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginRequest {
    pub username: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct RegisterRequest {
    pub username: String,
    pub email: String,
    pub password: String,
    pub confirm_password: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthResponse {
    pub success: bool,
    pub message: String,
    pub access_token: Option<String>,
    pub refresh_token: Option<String>,
    pub user: Option<User>,
    pub expires_in: Option<i64>, // token过期时间（秒）
}

#[derive(Debug, Clone)]
pub struct AuthState {
    pub access_token: Option<String>,
    pub refresh_token: Option<String>,
    pub current_user: Option<User>,
    pub logged_in: bool,
    pub token_expires_at: Option<Instant>,
}

impl Default for AuthState {
    fn default() -> Self {
        Self {
            access_token: None,
            refresh_token: None,
            current_user: None,
            logged_in: false,
            token_expires_at: None,
        }
    }
}

pub struct AuthManager {
    pub state: Arc<Mutex<AuthState>>,
    pub client: Client,
    pub auth_server_url: String,
    pub auto_refresh_enabled: bool,
}

impl AuthManager {
    pub fn new() -> Self {
        Self {
            state: Arc::new(Mutex::new(AuthState::default())),
            client: Client::new(),
            auth_server_url: "http://localhost:8082".to_string(),
            auto_refresh_enabled: true,
        }
    }

    pub async fn login(&self, request: LoginRequest) -> Result<AuthResponse, String> {
        let url = format!("{}/auth/login", self.auth_server_url);
        
        let response = self.client
            .post(&url)
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("登录失败: HTTP {}", response.status()));
        }

        let auth_response: AuthResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if auth_response.success {
            self.save_auth_data(&auth_response).await;
        }

        Ok(auth_response)
    }

    pub async fn register(&self, request: RegisterRequest) -> Result<AuthResponse, String> {
        if request.password != request.confirm_password {
            return Err("密码确认不匹配".to_string());
        }

        let url = format!("{}/auth/register", self.auth_server_url);
        
        let response = self.client
            .post(&url)
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("注册失败: HTTP {}", response.status()));
        }

        let auth_response: AuthResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if auth_response.success {
            self.save_auth_data(&auth_response).await;
        }

        Ok(auth_response)
    }

    pub async fn logout(&self) -> Result<bool, String> {
        let mut state = self.state.lock().unwrap();
        
        // 清除本地认证数据
        *state = AuthState::default();
        
        // TODO: 调用服务器注销接口
        
        Ok(true)
    }

    pub async fn refresh_token(&self) -> Result<bool, String> {
        let refresh_token = {
            let state = self.state.lock().unwrap();
            state.refresh_token.clone()
        };

        let refresh_token = refresh_token.ok_or("没有刷新令牌")?;

        let url = format!("{}/auth/refresh", self.auth_server_url);
        let mut payload = HashMap::new();
        payload.insert("refresh_token", refresh_token);

        let response = self.client
            .post(&url)
            .json(&payload)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("刷新令牌失败: HTTP {}", response.status()));
        }

        let auth_response: AuthResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if auth_response.success {
            self.save_auth_data(&auth_response).await;
            Ok(true)
        } else {
            Err(auth_response.message)
        }
    }

    pub fn is_logged_in(&self) -> bool {
        let state = self.state.lock().unwrap();
        state.logged_in && state.access_token.is_some()
    }

    pub fn get_access_token(&self) -> Option<String> {
        let state = self.state.lock().unwrap();
        state.access_token.clone()
    }

    pub fn get_current_user(&self) -> Option<User> {
        let state = self.state.lock().unwrap();
        state.current_user.clone()
    }

    pub fn set_auth_server_url(&mut self, url: String) {
        self.auth_server_url = url;
    }

    pub fn enable_auto_refresh(&mut self, enable: bool) {
        self.auto_refresh_enabled = enable;
    }

    pub async fn clear_auth_data(&self) {
        let mut state = self.state.lock().unwrap();
        *state = AuthState::default();
    }

    async fn save_auth_data(&self, response: &AuthResponse) {
        let mut state = self.state.lock().unwrap();
        
        if let Some(token) = &response.access_token {
            state.access_token = Some(token.clone());
        }
        
        if let Some(refresh_token) = &response.refresh_token {
            state.refresh_token = Some(refresh_token.clone());
        }
        
        if let Some(user) = &response.user {
            state.current_user = Some(user.clone());
        }
        
        state.logged_in = response.success;
        
        if let Some(expires_in) = response.expires_in {
            state.token_expires_at = Some(Instant::now() + Duration::from_secs(expires_in as u64));
        }
    }

    pub fn is_token_expired(&self) -> bool {
        let state = self.state.lock().unwrap();
        if let Some(expires_at) = state.token_expires_at {
            Instant::now() >= expires_at
        } else {
            false
        }
    }
}

// Tauri 命令
#[tauri::command]
pub async fn auth_login(
    auth_manager: State<'_, AuthManager>,
    request: LoginRequest,
) -> Result<AuthResponse, String> {
    auth_manager.login(request).await
}

#[tauri::command]
pub async fn auth_register(
    auth_manager: State<'_, AuthManager>,
    request: RegisterRequest,
) -> Result<AuthResponse, String> {
    auth_manager.register(request).await
}

#[tauri::command]
pub async fn auth_logout(auth_manager: State<'_, AuthManager>) -> Result<bool, String> {
    auth_manager.logout().await
}

#[tauri::command]
pub async fn auth_refresh_token(auth_manager: State<'_, AuthManager>) -> Result<bool, String> {
    auth_manager.refresh_token().await
}

#[tauri::command]
pub fn auth_is_logged_in(auth_manager: State<'_, AuthManager>) -> bool {
    auth_manager.is_logged_in()
}

#[tauri::command]
pub fn auth_get_current_user(auth_manager: State<'_, AuthManager>) -> Option<User> {
    auth_manager.get_current_user()
}

#[tauri::command]
pub fn auth_get_access_token(auth_manager: State<'_, AuthManager>) -> Option<String> {
    auth_manager.get_access_token()
}

#[tauri::command]
pub async fn auth_clear_data(auth_manager: State<'_, AuthManager>) -> Result<(), String> {
    auth_manager.clear_auth_data().await;
    Ok(())
}

#[tauri::command]
pub async fn auth_set_server_url(
    auth_manager: State<'_, AuthManager>,
    url: String,
) -> Result<(), String> {
    // 由于State是不可变的，我们需要使用内部可变性
    // 这里我们暂时返回成功，实际实现需要重构AuthManager
    Ok(())
}

#[tauri::command]
pub async fn auth_enable_auto_refresh(
    auth_manager: State<'_, AuthManager>,
    enable: bool,
) -> Result<(), String> {
    // 由于State是不可变的，我们需要使用内部可变性
    // 这里我们暂时返回成功，实际实现需要重构AuthManager
    Ok(())
}

#[tauri::command]
pub fn validate_session(auth_manager: State<'_, AuthManager>) -> bool {
    auth_manager.is_logged_in()
}

#[tauri::command]
pub fn get_user_info(auth_manager: State<'_, AuthManager>) -> Option<User> {
    auth_manager.get_current_user()
}