use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use tauri::State;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppModule {
    pub id: String,
    pub name: String,
    pub description: String,
    pub category: String,
    pub required_role: String,
    pub is_active: bool,
    pub icon: Option<String>,
    pub route: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UserModules {
    pub modules: Vec<AppModule>,
    pub user_permissions: HashMap<String, bool>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HealthStatus {
    pub status: String,
    pub timestamp: String,
    pub version: String,
    pub uptime: u64,
}

pub struct ModuleManager {
    modules: Vec<AppModule>,
}

impl ModuleManager {
    pub fn new() -> Self {
        let modules = vec![
            AppModule {
                id: "ai_chat".to_string(),
                name: "AI对话".to_string(),
                description: "与AI进行智能对话".to_string(),
                category: "AI".to_string(),
                required_role: "USER".to_string(),
                is_active: true,
                icon: Some("chat".to_string()),
                route: Some("/chat".to_string()),
            },
            AppModule {
                id: "file_transfer".to_string(),
                name: "文件传输".to_string(),
                description: "文件上传下载管理".to_string(),
                category: "工具".to_string(),
                required_role: "USER".to_string(),
                is_active: true,
                icon: Some("file".to_string()),
                route: Some("/files".to_string()),
            },
            AppModule {
                id: "project_manager".to_string(),
                name: "项目管理".to_string(),
                description: "项目和任务管理".to_string(),
                category: "管理".to_string(),
                required_role: "USER".to_string(),
                is_active: true,
                icon: Some("project".to_string()),
                route: Some("/projects".to_string()),
            },
            AppModule {
                id: "friend_manager".to_string(),
                name: "好友管理".to_string(),
                description: "好友和社交功能".to_string(),
                category: "社交".to_string(),
                required_role: "USER".to_string(),
                is_active: true,
                icon: Some("users".to_string()),
                route: Some("/friends".to_string()),
            },
            AppModule {
                id: "app_manager".to_string(),
                name: "应用管理".to_string(),
                description: "应用程序管理和监控".to_string(),
                category: "系统".to_string(),
                required_role: "ADMIN".to_string(),
                is_active: true,
                icon: Some("apps".to_string()),
                route: Some("/apps".to_string()),
            },
        ];

        Self { modules }
    }

    pub fn get_user_modules(&self, user_role: &str) -> UserModules {
        let available_modules: Vec<AppModule> = self
            .modules
            .iter()
            .filter(|module| {
                module.is_active && self.has_permission(user_role, &module.required_role)
            })
            .cloned()
            .collect();

        let mut user_permissions = HashMap::new();
        for module in &available_modules {
            user_permissions.insert(module.id.clone(), true);
        }

        UserModules {
            modules: available_modules,
            user_permissions,
        }
    }

    fn has_permission(&self, user_role: &str, required_role: &str) -> bool {
        match (user_role, required_role) {
            ("ADMIN", _) => true,
            ("USER", "USER") => true,
            ("GUEST", "GUEST") => true,
            _ => false,
        }
    }

    pub fn get_health_status(&self) -> HealthStatus {
        HealthStatus {
            status: "healthy".to_string(),
            timestamp: chrono::Utc::now().to_rfc3339(),
            version: env!("CARGO_PKG_VERSION").to_string(),
            uptime: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
        }
    }
}

// Tauri 命令
#[tauri::command]
pub async fn get_user_modules(
    module_manager: State<'_, std::sync::Arc<tokio::sync::Mutex<ModuleManager>>>,
    user_role: Option<String>,
) -> Result<UserModules, String> {
    let manager = module_manager.lock().await;
    let role = user_role.unwrap_or_else(|| "USER".to_string());
    Ok(manager.get_user_modules(&role))
}

#[tauri::command]
pub async fn health_check(
    module_manager: State<'_, std::sync::Arc<tokio::sync::Mutex<ModuleManager>>>,
) -> Result<HealthStatus, String> {
    let manager = module_manager.lock().await;
    Ok(manager.get_health_status())
}