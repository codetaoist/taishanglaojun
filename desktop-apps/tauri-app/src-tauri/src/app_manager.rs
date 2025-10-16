use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::State;
use reqwest::Client;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AppStatus {
    Installed,
    Running,
    Stopped,
    Updating,
    Uninstalled,
    Error,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AppCategory {
    Productivity,
    Development,
    Communication,
    Entertainment,
    Education,
    Business,
    Utilities,
    Games,
    Other,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AppType {
    Desktop,
    Web,
    Mobile,
    Service,
    Plugin,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Application {
    pub id: String,
    pub name: String,
    pub display_name: String,
    pub description: String,
    pub version: String,
    pub app_type: AppType,
    pub category: AppCategory,
    pub status: AppStatus,
    pub icon_url: Option<String>,
    pub executable_path: Option<String>,
    pub install_path: Option<String>,
    pub config_path: Option<String>,
    pub log_path: Option<String>,
    pub pid: Option<u32>,
    pub port: Option<u16>,
    pub url: Option<String>,
    pub auto_start: bool,
    pub dependencies: Vec<String>,
    pub permissions: Vec<String>,
    pub environment_vars: HashMap<String, String>,
    pub command_args: Vec<String>,
    pub working_directory: Option<String>,
    pub memory_usage: Option<u64>,
    pub cpu_usage: Option<f64>,
    pub disk_usage: Option<u64>,
    pub network_usage: Option<u64>,
    pub last_started: Option<String>,
    pub last_stopped: Option<String>,
    pub install_date: Option<String>,
    pub update_date: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppProcess {
    pub app_id: String,
    pub pid: u32,
    pub name: String,
    pub status: String,
    pub memory_usage: u64,
    pub cpu_usage: f64,
    pub start_time: String,
    pub command_line: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppLog {
    pub id: String,
    pub app_id: String,
    pub level: String,
    pub message: String,
    pub timestamp: String,
    pub source: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub app_id: String,
    pub key: String,
    pub value: String,
    pub description: Option<String>,
    pub is_sensitive: bool,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct InstallAppRequest {
    pub name: String,
    pub display_name: String,
    pub description: String,
    pub app_type: AppType,
    pub category: AppCategory,
    pub executable_path: Option<String>,
    pub install_path: Option<String>,
    pub auto_start: bool,
    pub dependencies: Vec<String>,
    pub permissions: Vec<String>,
    pub environment_vars: HashMap<String, String>,
    pub command_args: Vec<String>,
    pub working_directory: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateAppRequest {
    pub display_name: Option<String>,
    pub description: Option<String>,
    pub version: Option<String>,
    pub status: Option<AppStatus>,
    pub executable_path: Option<String>,
    pub auto_start: Option<bool>,
    pub environment_vars: Option<HashMap<String, String>>,
    pub command_args: Option<Vec<String>>,
    pub working_directory: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct StartAppRequest {
    pub app_id: String,
    pub environment_vars: Option<HashMap<String, String>>,
    pub command_args: Option<Vec<String>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct StopAppRequest {
    pub app_id: String,
    pub force: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateConfigRequest {
    pub app_id: String,
    pub configs: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppResponse {
    pub success: bool,
    pub message: String,
    pub applications: Option<Vec<Application>>,
    pub application: Option<Application>,
    pub processes: Option<Vec<AppProcess>>,
    pub logs: Option<Vec<AppLog>>,
    pub configs: Option<Vec<AppConfig>>,
}

#[derive(Debug, Clone)]
pub struct AppManagerState {
    pub applications: HashMap<String, Application>,
    pub processes: HashMap<String, Vec<AppProcess>>, // app_id -> processes
    pub logs: HashMap<String, Vec<AppLog>>, // app_id -> logs
    pub configs: HashMap<String, Vec<AppConfig>>, // app_id -> configs
    pub running_apps: HashMap<String, u32>, // app_id -> pid
}

impl Default for AppManagerState {
    fn default() -> Self {
        Self {
            applications: HashMap::new(),
            processes: HashMap::new(),
            logs: HashMap::new(),
            configs: HashMap::new(),
            running_apps: HashMap::new(),
        }
    }
}

pub struct AppManager {
    pub state: Arc<Mutex<AppManagerState>>,
    pub client: Client,
    pub api_base_url: String,
}

impl AppManager {
    pub fn new() -> Self {
        Self {
            state: Arc::new(Mutex::new(AppManagerState::default())),
            client: Client::new(),
            api_base_url: "http://localhost:8082".to_string(),
        }
    }

    pub async fn install_app(&self, auth_token: &str, request: InstallAppRequest) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps", self.api_base_url);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("安装应用失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(app) = &app_response.application {
                let mut state = self.state.lock().unwrap();
                state.applications.insert(app.id.clone(), app.clone());
            }
        }

        Ok(app_response)
    }

    pub async fn uninstall_app(&self, auth_token: &str, app_id: &str) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}", self.api_base_url, app_id);
        
        let response = self.client
            .delete(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("卸载应用失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 从本地状态中移除
        if app_response.success {
            let mut state = self.state.lock().unwrap();
            state.applications.remove(app_id);
            state.processes.remove(app_id);
            state.logs.remove(app_id);
            state.configs.remove(app_id);
            state.running_apps.remove(app_id);
        }

        Ok(app_response)
    }

    pub async fn update_app(&self, auth_token: &str, app_id: &str, request: UpdateAppRequest) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}", self.api_base_url, app_id);
        
        let response = self.client
            .put(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("更新应用失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(app) = &app_response.application {
                let mut state = self.state.lock().unwrap();
                state.applications.insert(app.id.clone(), app.clone());
            }
        }

        Ok(app_response)
    }

    pub async fn start_app(&self, auth_token: &str, request: StartAppRequest) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}/start", self.api_base_url, request.app_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("启动应用失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(app) = &app_response.application {
                let mut state = self.state.lock().unwrap();
                state.applications.insert(app.id.clone(), app.clone());
                if let Some(pid) = app.pid {
                    state.running_apps.insert(app.id.clone(), pid);
                }
            }
        }

        Ok(app_response)
    }

    pub async fn stop_app(&self, auth_token: &str, request: StopAppRequest) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}/stop", self.api_base_url, request.app_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("停止应用失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(app) = &app_response.application {
                let mut state = self.state.lock().unwrap();
                state.applications.insert(app.id.clone(), app.clone());
                state.running_apps.remove(&app.id);
            }
        }

        Ok(app_response)
    }

    pub async fn restart_app(&self, auth_token: &str, app_id: &str) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}/restart", self.api_base_url, app_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("重启应用失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(app) = &app_response.application {
                let mut state = self.state.lock().unwrap();
                state.applications.insert(app.id.clone(), app.clone());
                if let Some(pid) = app.pid {
                    state.running_apps.insert(app.id.clone(), pid);
                } else {
                    state.running_apps.remove(&app.id);
                }
            }
        }

        Ok(app_response)
    }

    pub async fn get_app(&self, auth_token: &str, app_id: &str) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}", self.api_base_url, app_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取应用信息失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(app) = &app_response.application {
                let mut state = self.state.lock().unwrap();
                state.applications.insert(app.id.clone(), app.clone());
            }
        }

        Ok(app_response)
    }

    pub async fn list_apps(&self, auth_token: &str, category: Option<AppCategory>, status: Option<AppStatus>) -> Result<AppResponse, String> {
        let mut url = format!("{}/api/apps", self.api_base_url);
        
        let mut params = Vec::new();
        if let Some(category) = category {
            params.push(format!("category={:?}", category));
        }
        if let Some(status) = status {
            params.push(format!("status={:?}", status));
        }
        
        if !params.is_empty() {
            url.push('?');
            url.push_str(&params.join("&"));
        }
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取应用列表失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(apps) = &app_response.applications {
                let mut state = self.state.lock().unwrap();
                for app in apps {
                    state.applications.insert(app.id.clone(), app.clone());
                    if let Some(pid) = app.pid {
                        if app.status == AppStatus::Running {
                            state.running_apps.insert(app.id.clone(), pid);
                        }
                    }
                }
            }
        }

        Ok(app_response)
    }

    pub async fn get_app_processes(&self, auth_token: &str, app_id: &str) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}/processes", self.api_base_url, app_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取应用进程失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(processes) = &app_response.processes {
                let mut state = self.state.lock().unwrap();
                state.processes.insert(app_id.to_string(), processes.clone());
            }
        }

        Ok(app_response)
    }

    pub async fn get_app_logs(&self, auth_token: &str, app_id: &str, limit: Option<i32>, level: Option<String>) -> Result<AppResponse, String> {
        let mut url = format!("{}/api/apps/{}/logs", self.api_base_url, app_id);
        
        let mut params = Vec::new();
        if let Some(limit) = limit {
            params.push(format!("limit={}", limit));
        }
        if let Some(level) = level {
            params.push(format!("level={}", level));
        }
        
        if !params.is_empty() {
            url.push('?');
            url.push_str(&params.join("&"));
        }
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取应用日志失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(logs) = &app_response.logs {
                let mut state = self.state.lock().unwrap();
                state.logs.insert(app_id.to_string(), logs.clone());
            }
        }

        Ok(app_response)
    }

    pub async fn get_app_configs(&self, auth_token: &str, app_id: &str) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}/configs", self.api_base_url, app_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取应用配置失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if app_response.success {
            if let Some(configs) = &app_response.configs {
                let mut state = self.state.lock().unwrap();
                state.configs.insert(app_id.to_string(), configs.clone());
            }
        }

        Ok(app_response)
    }

    pub async fn update_app_configs(&self, auth_token: &str, request: UpdateConfigRequest) -> Result<AppResponse, String> {
        let url = format!("{}/api/apps/{}/configs", self.api_base_url, request.app_id);
        
        let response = self.client
            .put(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("更新应用配置失败: HTTP {}", response.status()));
        }

        let app_response: AppResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(app_response)
    }

    pub fn get_cached_apps(&self) -> Vec<Application> {
        let state = self.state.lock().unwrap();
        state.applications.values().cloned().collect()
    }

    pub fn get_cached_processes(&self, app_id: &str) -> Vec<AppProcess> {
        let state = self.state.lock().unwrap();
        state.processes.get(app_id).cloned().unwrap_or_default()
    }

    pub fn get_cached_logs(&self, app_id: &str) -> Vec<AppLog> {
        let state = self.state.lock().unwrap();
        state.logs.get(app_id).cloned().unwrap_or_default()
    }

    pub fn get_cached_configs(&self, app_id: &str) -> Vec<AppConfig> {
        let state = self.state.lock().unwrap();
        state.configs.get(app_id).cloned().unwrap_or_default()
    }

    pub fn get_running_apps(&self) -> HashMap<String, u32> {
        let state = self.state.lock().unwrap();
        state.running_apps.clone()
    }

    pub fn is_app_running(&self, app_id: &str) -> bool {
        let state = self.state.lock().unwrap();
        state.running_apps.contains_key(app_id)
    }

    pub fn set_api_base_url(&mut self, url: String) {
        self.api_base_url = url;
    }
}

// Tauri 命令
#[tauri::command]
pub async fn app_install(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    request: InstallAppRequest,
) -> Result<AppResponse, String> {
    app_manager.install_app(&auth_token, request).await
}

#[tauri::command]
pub async fn app_uninstall(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
) -> Result<AppResponse, String> {
    app_manager.uninstall_app(&auth_token, &app_id).await
}

#[tauri::command]
pub async fn app_update(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
    request: UpdateAppRequest,
) -> Result<AppResponse, String> {
    app_manager.update_app(&auth_token, &app_id, request).await
}

#[tauri::command]
pub async fn app_start(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    request: StartAppRequest,
) -> Result<AppResponse, String> {
    app_manager.start_app(&auth_token, request).await
}

#[tauri::command]
pub async fn app_stop(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    request: StopAppRequest,
) -> Result<AppResponse, String> {
    app_manager.stop_app(&auth_token, request).await
}

#[tauri::command]
pub async fn app_restart(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
) -> Result<AppResponse, String> {
    app_manager.restart_app(&auth_token, &app_id).await
}

#[tauri::command]
pub async fn app_get(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
) -> Result<AppResponse, String> {
    app_manager.get_app(&auth_token, &app_id).await
}

#[tauri::command]
pub async fn app_list(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    category: Option<AppCategory>,
    status: Option<AppStatus>,
) -> Result<AppResponse, String> {
    app_manager.list_apps(&auth_token, category, status).await
}

#[tauri::command]
pub async fn app_get_processes(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
) -> Result<AppResponse, String> {
    app_manager.get_app_processes(&auth_token, &app_id).await
}

#[tauri::command]
pub async fn app_get_logs(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
    limit: Option<i32>,
    level: Option<String>,
) -> Result<AppResponse, String> {
    app_manager.get_app_logs(&auth_token, &app_id, limit, level).await
}

#[tauri::command]
pub async fn app_get_configs(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    app_id: String,
) -> Result<AppResponse, String> {
    app_manager.get_app_configs(&auth_token, &app_id).await
}

#[tauri::command]
pub async fn app_update_configs(
    app_manager: State<'_, AppManager>,
    auth_token: String,
    request: UpdateConfigRequest,
) -> Result<AppResponse, String> {
    app_manager.update_app_configs(&auth_token, request).await
}

#[tauri::command]
pub fn app_get_cached_list(app_manager: State<'_, AppManager>) -> Vec<Application> {
    app_manager.get_cached_apps()
}

#[tauri::command]
pub fn app_get_cached_processes(
    app_manager: State<'_, AppManager>,
    app_id: String,
) -> Vec<AppProcess> {
    app_manager.get_cached_processes(&app_id)
}

#[tauri::command]
pub fn app_get_cached_logs(
    app_manager: State<'_, AppManager>,
    app_id: String,
) -> Vec<AppLog> {
    app_manager.get_cached_logs(&app_id)
}

#[tauri::command]
pub fn app_get_cached_configs(
    app_manager: State<'_, AppManager>,
    app_id: String,
) -> Vec<AppConfig> {
    app_manager.get_cached_configs(&app_id)
}

#[tauri::command]
pub fn app_get_running_list(app_manager: State<'_, AppManager>) -> HashMap<String, u32> {
    app_manager.get_running_apps()
}

#[tauri::command]
pub fn app_is_running(
    app_manager: State<'_, AppManager>,
    app_id: String,
) -> bool {
    app_manager.is_app_running(&app_id)
}