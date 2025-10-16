// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod ai_service;
mod auth;
mod chat;
mod document;
mod file_transfer;
mod image_generator;
mod input_optimizer;
mod security;
mod storage;
mod transfer_security;
mod auth_manager;
mod friend_manager;
mod project_manager;
mod app_manager;
mod chat_manager;
mod module_manager;
mod data_sync_manager;
mod data_access_layer;
mod database_config;
mod sync_service;
mod realtime_sync;
mod offline_manager;
mod sync_config;

use ai_service::AIService;
use auth::AuthManager;
use chat::ChatManager;
use document::DocumentProcessor;
use file_transfer::FileTransferManager;
use image_generator::ImageGenerator;
use input_optimizer::InputOptimizer;
use security::SecurityManager;
use storage::StorageManager;
use transfer_security::TransferSecurityManager;

use std::sync::Arc;
use tauri::State;
use tokio::sync::Mutex;

// 应用状态
pub struct AppState {
    ai_service: Arc<Mutex<AIService>>,
    auth_manager: Arc<Mutex<AuthManager>>,
    chat_manager: Arc<Mutex<ChatManager>>,
    document_processor: Arc<Mutex<DocumentProcessor>>,
    file_transfer_manager: Arc<Mutex<FileTransferManager>>,
    image_generator: Arc<Mutex<ImageGenerator>>,
    input_optimizer: Arc<Mutex<InputOptimizer>>,
    security_manager: Arc<Mutex<SecurityManager>>,
    storage_manager: Arc<Mutex<StorageManager>>,
    transfer_security_manager: Arc<Mutex<TransferSecurityManager>>,
    // 新的管理器
    new_auth_manager: auth_manager::AuthManager,
    new_friend_manager: friend_manager::FriendManager,
    new_project_manager: project_manager::ProjectManager,
    new_app_manager: app_manager::AppManager,
    new_chat_manager: chat_manager::ChatManager,
    module_manager: Arc<Mutex<module_manager::ModuleManager>>,
    // 数据管理层
    data_sync_manager: Arc<data_sync_manager::DataSyncManager>,
    data_access_layer: Arc<data_access_layer::DataAccessLayer>,
    // 多设备同步组件
    sync_service: Arc<sync_service::MultiDeviceSyncService>,
    realtime_sync: Arc<realtime_sync::RealtimeSyncManager>,
    offline_manager: Arc<offline_manager::OfflineDataManager>,
}

// AI对话命令
#[tauri::command]
async fn chat_with_ai(
    message: String,
    chat_type: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let chat_manager = state.chat_manager.lock().await;
    chat_manager
        .send_message(message, chat_type)
        .await
        .map_err(|e| e.to_string())
}

// 输入优化命令
#[tauri::command]
async fn optimize_input(
    text: String,
    target_audience: Option<String>,
    optimization_type: Option<String>,
    language: Option<String>,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    use input_optimizer::OptimizationRequest;
    
    let mut input_optimizer = state.input_optimizer.lock().await;
    let request = OptimizationRequest {
        text,
        target_audience: target_audience.unwrap_or_else(|| "general".to_string()),
        optimization_type: optimization_type.unwrap_or_else(|| "clarity".to_string()),
        language: language.unwrap_or_else(|| "zh".to_string()),
        platform: None,
    };
    
    let result = input_optimizer
        .optimize_input(request)
        .await
        .map_err(|e| e.to_string())?;
    
    serde_json::to_value(result).map_err(|e| e.to_string())
}

#[tauri::command]
async fn optimize_input_windows(
    text: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    use input_optimizer::OptimizationRequest;
    
    let mut input_optimizer = state.input_optimizer.lock().await;
    let request = OptimizationRequest {
        text: text.clone(),
        target_audience: "general".to_string(),
        optimization_type: "platform_specific".to_string(),
        language: "zh".to_string(),
        platform: Some("windows".to_string()),
    };
    
    let result = input_optimizer
        .optimize_input(request)
        .await
        .map_err(|e| e.to_string())?;
    
    Ok(result.best_suggestion
        .map(|s| s.optimized_text)
        .unwrap_or(text))
}

#[tauri::command]
async fn optimize_input_macos(
    text: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    use input_optimizer::OptimizationRequest;
    
    let mut input_optimizer = state.input_optimizer.lock().await;
    let request = OptimizationRequest {
        text: text.clone(),
        target_audience: "general".to_string(),
        optimization_type: "platform_specific".to_string(),
        language: "zh".to_string(),
        platform: Some("macos".to_string()),
    };
    
    let result = input_optimizer
        .optimize_input(request)
        .await
        .map_err(|e| e.to_string())?;
    
    Ok(result.best_suggestion
        .map(|s| s.optimized_text)
        .unwrap_or(text))
}

#[tauri::command]
async fn optimize_input_linux(
    text: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    use input_optimizer::OptimizationRequest;
    
    let mut input_optimizer = state.input_optimizer.lock().await;
    let request = OptimizationRequest {
        text: text.clone(),
        target_audience: "general".to_string(),
        optimization_type: "platform_specific".to_string(),
        language: "zh".to_string(),
        platform: Some("linux".to_string()),
    };
    
    let result = input_optimizer
        .optimize_input(request)
        .await
        .map_err(|e| e.to_string())?;
    
    Ok(result.best_suggestion
        .map(|s| s.optimized_text)
        .unwrap_or(text))
}

#[tauri::command]
async fn get_quick_suggestions(
    text: String,
    state: State<'_, AppState>,
) -> Result<Vec<String>, String> {
    let input_optimizer = state.input_optimizer.lock().await;
    input_optimizer
        .get_quick_suggestions(&text)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn detect_input_intent(
    text: String,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    let input_optimizer = state.input_optimizer.lock().await;
    Ok(input_optimizer.detect_intent(&text))
}

// 文档处理命令
#[tauri::command]
async fn process_document(
    file_path: String,
    operation: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let document_processor = state.document_processor.lock().await;
    document_processor
        .process_file(file_path, operation)
        .await
        .map_err(|e| e.to_string())
}

// 图像生成命令
#[tauri::command]
async fn generate_image(
    prompt: String,
    style: Option<String>,
    width: Option<u32>,
    height: Option<u32>,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    use image_generator::ImageGenerationRequest;
    
    let image_generator = state.image_generator.lock().await;
    let request = ImageGenerationRequest {
        prompt,
        negative_prompt: None,
        style,
        width,
        height,
        steps: None,
        guidance_scale: None,
        seed: None,
        batch_size: None,
    };
    
    let result = image_generator
        .generate_image(request)
        .await
        .map_err(|e| e.to_string())?;
    
    serde_json::to_value(result).map_err(|e| e.to_string())
}

// 图像分析命令
#[tauri::command]
async fn analyze_image(
    image_path: String,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    let image_generator = state.image_generator.lock().await;
    let result = image_generator
        .analyze_image(image_path)
        .await
        .map_err(|e| e.to_string())?;
    
    serde_json::to_value(result).map_err(|e| e.to_string())
}

// 图像编辑命令
#[tauri::command]
async fn edit_image(
    image_path: String,
    operation: String,
    parameters: serde_json::Value,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    use image_generator::ImageEditRequest;
    
    let image_generator = state.image_generator.lock().await;
    let request = ImageEditRequest {
        image_path,
        operation,
        parameters,
    };
    
    let result = image_generator
        .edit_image(request)
        .await
        .map_err(|e| e.to_string())?;
    
    serde_json::to_value(result).map_err(|e| e.to_string())
}

// 保存图像命令
#[tauri::command]
async fn save_generated_image(
    image_data: serde_json::Value,
    save_path: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    use image_generator::GeneratedImage;
    
    let image_generator = state.image_generator.lock().await;
    let image: GeneratedImage = serde_json::from_value(image_data)
        .map_err(|e| format!("图像数据解析失败: {}", e))?;
    
    image_generator
        .save_image(&image, &save_path)
        .await
        .map_err(|e| e.to_string())
}

// 用户认证命令
#[tauri::command]
async fn login(
    username: String,
    password: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let mut auth_manager = state.auth_manager.lock().await;
    auth_manager
        .login(username, password)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn logout(state: State<'_, AppState>) -> Result<(), String> {
    let mut auth_manager = state.auth_manager.lock().await;
    auth_manager.logout().await.map_err(|e| e.to_string())
}

// 获取聊天历史
#[tauri::command]
async fn get_chat_history(
    limit: Option<usize>,
    state: State<'_, AppState>,
) -> Result<Vec<serde_json::Value>, String> {
    let chat_manager = state.chat_manager.lock().await;
    chat_manager
        .get_history(limit.unwrap_or(50))
        .await
        .map_err(|e| e.to_string())
}

// 保存文件
#[tauri::command]
async fn save_file(
    path: String,
    content: String,
    state: State<'_, AppState>,
) -> Result<(), String> {
    let mut storage_manager = state.storage_manager.lock().await;
    storage_manager
        .save_file(path, content)
        .await
        .map(|_| ())
        .map_err(|e| e.to_string())
}

// 读取文件
#[tauri::command]
async fn read_file(path: String, state: State<'_, AppState>) -> Result<String, String> {
    let storage_manager = state.storage_manager.lock().await;
    storage_manager
        .read_file(path)
        .await
        .map_err(|e| e.to_string())
}

// 获取系统状态
#[tauri::command]
async fn get_system_status(state: State<'_, AppState>) -> Result<serde_json::Value, String> {
    let ai_service = state.ai_service.lock().await;
    ai_service.get_status().await.map_err(|e| e.to_string())
}

// 加密数据
#[tauri::command]
async fn encrypt_data(
    data: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let security_manager = state.security_manager.lock().await;
    security_manager
        .encrypt(data)
        .await
        .map_err(|e| e.to_string())
}

// 解密数据
#[tauri::command]
async fn decrypt_data(
    encrypted_data: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let security_manager = state.security_manager.lock().await;
    security_manager
        .decrypt(encrypted_data)
        .await
        .map_err(|e| e.to_string())
}

// 文件传输命令

// 启动设备发现
#[tauri::command]
async fn start_device_discovery(state: State<'_, AppState>) -> Result<(), String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    file_transfer_manager
        .start_device_discovery()
        .await
        .map_err(|e| e.to_string())
}

// 停止设备发现
#[tauri::command]
async fn stop_device_discovery(state: State<'_, AppState>) -> Result<(), String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    file_transfer_manager
        .stop_device_discovery()
        .await
        .map_err(|e| e.to_string())
}

// 获取发现的设备列表
#[tauri::command]
async fn get_discovered_devices(state: State<'_, AppState>) -> Result<Vec<serde_json::Value>, String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    let devices = file_transfer_manager.get_discovered_devices().await;
    devices.into_iter()
        .map(|device| serde_json::to_value(device).map_err(|e| e.to_string()))
        .collect()
}

// 发起文件传输
#[tauri::command]
async fn initiate_file_transfer(
    file_path: String,
    target_device_id: String,
    source_account: String,
    target_account: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    
    // 创建目标设备信息（这里简化处理，实际应该从设备发现中获取）
    let target_device = file_transfer::DeviceInfo {
        id: target_device_id,
        name: "目标设备".to_string(),
        device_type: "Unknown".to_string(),
        ip_address: "127.0.0.1".parse().unwrap(),
        port: 8080,
        protocol: file_transfer::TransferProtocol::LocalNetwork,
        capabilities: vec!["file_transfer".to_string()],
        last_seen: chrono::Utc::now().to_rfc3339(),
        is_trusted: false,
    };
    
    let task_id = file_transfer_manager
        .initiate_transfer(
            std::path::PathBuf::from(file_path),
            target_device,
            source_account,
            target_account,
        )
        .await
        .map_err(|e| e.to_string())?;
    Ok(task_id)
}

// 获取传输任务列表
#[tauri::command]
async fn get_transfer_tasks(state: State<'_, AppState>) -> Result<Vec<serde_json::Value>, String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    let tasks = file_transfer_manager.get_transfer_tasks().await;
    tasks.into_iter()
        .map(|task| serde_json::to_value(task).map_err(|e| e.to_string()))
        .collect()
}

// 暂停传输任务
#[tauri::command]
async fn pause_transfer(task_id: String, state: State<'_, AppState>) -> Result<(), String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    file_transfer_manager
        .pause_transfer(&task_id)
        .await
        .map_err(|e| e.to_string())
}

// 恢复传输任务
#[tauri::command]
async fn resume_transfer(task_id: String, state: State<'_, AppState>) -> Result<(), String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    file_transfer_manager
        .resume_transfer(&task_id)
        .await
        .map_err(|e| e.to_string())
}

// 取消传输任务
#[tauri::command]
async fn cancel_transfer(task_id: String, state: State<'_, AppState>) -> Result<(), String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    file_transfer_manager
        .cancel_transfer(&task_id)
        .await
        .map_err(|e| e.to_string())
}

// 切换账号
#[tauri::command]
async fn switch_account(account_id: String, state: State<'_, AppState>) -> Result<(), String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    file_transfer_manager
        .switch_account(account_id)
        .await
        .map_err(|e| e.to_string())
}

// 获取当前账号信息
#[tauri::command]
async fn get_current_account(state: State<'_, AppState>) -> Result<serde_json::Value, String> {
    let file_transfer_manager = state.file_transfer_manager.lock().await;
    let account = file_transfer_manager.get_current_account().await;
    serde_json::to_value(account).map_err(|e| e.to_string())
}

// 添加信任设备
#[tauri::command]
async fn add_trusted_device(device_id: String, state: State<'_, AppState>) -> Result<(), String> {
    let transfer_security_manager = state.transfer_security_manager.lock().await;
    transfer_security_manager
        .add_trusted_device(device_id)
        .await
        .map_err(|e| e.to_string())
}

// 移除信任设备
#[tauri::command]
async fn remove_trusted_device(device_id: String, state: State<'_, AppState>) -> Result<(), String> {
    let transfer_security_manager = state.transfer_security_manager.lock().await;
    transfer_security_manager
        .remove_trusted_device(&device_id)
        .await
        .map_err(|e| e.to_string())
}

// 生成认证令牌
#[tauri::command]
async fn generate_auth_token(
    device_id: String,
    account_id: String,
    permissions: Vec<String>,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    let transfer_security_manager = state.transfer_security_manager.lock().await;
    let token = transfer_security_manager
        .generate_auth_token(device_id, account_id, permissions)
        .await
        .map_err(|e| e.to_string())?;
    serde_json::to_value(token).map_err(|e| e.to_string())
}

#[tokio::main]
async fn main() {
    env_logger::init();

    // 创建本地设备信息
    let local_device = file_transfer::DeviceInfo {
        id: uuid::Uuid::new_v4().to_string(),
        name: "太上老君桌面应用".to_string(),
        device_type: "Desktop".to_string(),
        ip_address: "127.0.0.1".parse().unwrap(),
        port: 8080,
        protocol: file_transfer::TransferProtocol::LocalNetwork,
        capabilities: vec!["file_transfer".to_string(), "encryption".to_string()],
        last_seen: chrono::Utc::now().to_rfc3339(),
        is_trusted: true,
    };

    // 初始化数据库管理器
    let app_data_dir = dirs::data_dir()
        .unwrap_or_else(|| std::env::current_dir().unwrap())
        .join("taishang-laojun");
    
    let mut db_manager = database_config::DatabaseManager::with_app_data_dir(app_data_dir);
    db_manager.initialize().await.expect("Failed to initialize database manager");
    
    // 创建数据同步管理器
    let data_sync_manager = Arc::new(
        data_sync_manager::DataSyncManager::new(
            db_manager.get_main_pool().unwrap().clone()
        ).await.expect("Failed to create data sync manager")
    );
    
    // 创建数据访问层
    let data_access_layer = Arc::new(
        data_access_layer::DataAccessLayer::new(
            db_manager.get_main_pool().unwrap().clone(),
            db_manager.get_chat_pool().unwrap().clone(),
            db_manager.get_storage_pool().unwrap().clone(),
            data_sync_manager.clone(),
        )
    );

    // 初始化多设备同步组件
    let sync_service = Arc::new(
        sync_service::MultiDeviceSyncService::new(
            db_manager.get_main_pool().unwrap().clone(),
            db_manager.get_chat_pool().unwrap().clone(),
            db_manager.get_storage_pool().unwrap().clone()
        )
    );
    
    let realtime_sync = Arc::new(
        realtime_sync::RealtimeSyncManager::new(sync_service.clone())
    );
    
    let offline_manager = Arc::new(
        offline_manager::OfflineDataManager::new(
            db_manager.get_main_pool().unwrap().clone()
        )
    );

    // 初始化应用状态
    let app_state = AppState {
        ai_service: Arc::new(Mutex::new(AIService::new().await.unwrap())),
        auth_manager: Arc::new(Mutex::new(AuthManager::new().await.unwrap())),
        chat_manager: Arc::new(Mutex::new(ChatManager::new().await.unwrap())),
        document_processor: Arc::new(Mutex::new(DocumentProcessor::new().await.unwrap())),
        file_transfer_manager: Arc::new(Mutex::new(FileTransferManager::new(local_device))),
        image_generator: Arc::new(Mutex::new(ImageGenerator::new().await.unwrap())),
        input_optimizer: Arc::new(Mutex::new(InputOptimizer::new().await.unwrap())),
        security_manager: Arc::new(Mutex::new(SecurityManager::new().await.unwrap())),
        storage_manager: Arc::new(Mutex::new(StorageManager::new().await.unwrap())),
        transfer_security_manager: Arc::new(Mutex::new(TransferSecurityManager::new())),
        // 初始化新的管理器
        new_auth_manager: auth_manager::AuthManager::new(),
        new_friend_manager: friend_manager::FriendManager::new(),
        new_project_manager: project_manager::ProjectManager::new(),
        new_app_manager: app_manager::AppManager::new(),
        new_chat_manager: chat_manager::ChatManager::new(),
        module_manager: Arc::new(Mutex::new(module_manager::ModuleManager::new())),
        // 数据管理层
        data_sync_manager,
        data_access_layer,
        // 多设备同步组件
        sync_service,
        realtime_sync,
        offline_manager,
    };

    tauri::Builder::default()
        .manage(app_state.new_auth_manager)
        .manage(app_state.new_friend_manager)
        .manage(app_state.new_project_manager)
        .manage(app_state.new_app_manager)
        .manage(app_state.new_chat_manager)
        .manage(app_state.module_manager)
        .manage(app_state.ai_service)
        .manage(app_state.auth_manager)
        .manage(app_state.chat_manager)
        .manage(app_state.document_processor)
        .manage(app_state.file_transfer_manager)
        .manage(app_state.image_generator)
        .manage(app_state.input_optimizer)
        .manage(app_state.security_manager)
        .manage(app_state.storage_manager)
        .manage(app_state.transfer_security_manager)
        .manage(app_state.data_sync_manager)
        .manage(app_state.data_access_layer)
        .manage(app_state.sync_service)
        .manage(app_state.realtime_sync)
        .manage(app_state.offline_manager)
        .invoke_handler(tauri::generate_handler![
            chat_with_ai,
            optimize_input,
            optimize_input_windows,
            optimize_input_macos,
            optimize_input_linux,
            get_quick_suggestions,
            detect_input_intent,
            process_document,
            generate_image,
            analyze_image,
            edit_image,
            save_generated_image,
            login,
            logout,
            get_chat_history,
            save_file,
            read_file,
            get_system_status,
            encrypt_data,
            decrypt_data,
            start_device_discovery,
            stop_device_discovery,
            get_discovered_devices,
            initiate_file_transfer,
            get_transfer_tasks,
            pause_transfer,
            resume_transfer,
            cancel_transfer,
            switch_account,
            get_current_account,
            add_trusted_device,
            remove_trusted_device,
            generate_auth_token,
            // 模块管理命令
            module_manager::get_user_modules,
            module_manager::health_check,
            // 认证管理命令
            auth_manager::auth_login,
            auth_manager::auth_register,
            auth_manager::auth_logout,
            auth_manager::auth_refresh_token,
            auth_manager::auth_is_logged_in,
            auth_manager::auth_get_access_token,
            auth_manager::auth_get_current_user,
            auth_manager::auth_set_server_url,
            auth_manager::auth_enable_auto_refresh,
            auth_manager::auth_clear_data,
            auth_manager::validate_session,
            auth_manager::get_user_info,
            // 好友管理命令
            friend_manager::friend_get_list,
            friend_manager::friend_get_requests,
            friend_manager::friend_add,
            friend_manager::friend_respond_request,
            friend_manager::friend_remove,
            friend_manager::friend_block,
            friend_manager::friend_unblock,
            friend_manager::friend_get_cached_list,
            friend_manager::friend_get_cached_requests,
            friend_manager::friend_refresh_data,
            // 项目管理命令
            project_manager::project_create,
            project_manager::project_update,
            project_manager::project_delete,
            project_manager::project_get,
            project_manager::project_list,
            project_manager::issue_create,
            project_manager::issue_update,
            project_manager::issue_delete,
            project_manager::issue_get,
            project_manager::issue_list,
            project_manager::member_add,
            project_manager::member_remove,
            project_manager::member_update_role,
            project_manager::member_list,
            project_manager::milestone_create,
            project_manager::milestone_update,
            project_manager::milestone_delete,
            project_manager::milestone_list,
            project_manager::comment_add,
            project_manager::comment_update,
            project_manager::comment_delete,
            project_manager::comment_get,
            project_manager::project_get_cached_list,
            project_manager::issue_get_cached_list,
            // 应用程序管理命令
            app_manager::app_install,
            app_manager::app_uninstall,
            app_manager::app_update,
            app_manager::app_start,
            app_manager::app_stop,
            app_manager::app_restart,
            app_manager::app_get,
            app_manager::app_list,
            app_manager::app_get_processes,
            app_manager::app_get_logs,
            app_manager::app_get_configs,
            app_manager::app_update_configs,
            app_manager::app_get_cached_list,
            app_manager::app_get_cached_processes,
            app_manager::app_get_cached_logs,
            app_manager::app_get_cached_configs,
            app_manager::app_get_running_list,
            app_manager::app_is_running,
            // 聊天管理命令
            chat_manager::chat_get_list,
            chat_manager::chat_get_messages,
            chat_manager::chat_send_message,
            chat_manager::chat_create,
            chat_manager::chat_delete,
            chat_manager::chat_mark_read,
            chat_manager::chat_connect_websocket,
            chat_manager::chat_disconnect_websocket,
            chat_manager::chat_get_cached_list,
            chat_manager::chat_get_cached_messages,
            chat_manager::chat_is_connected,
            chat_manager::chat_set_current_user,
            // 数据管理命令
            data_access_layer::get_user_with_stats,
            data_access_layer::search_all_data,
            data_access_layer::get_database_statistics,
            database_config::check_database_health,
            // 多设备同步命令
            sync_service::register_device,
            sync_service::get_user_devices,
            sync_service::sync_incremental,
            sync_service::sync_chat_messages,
            sync_service::sync_friend_data,
            realtime_sync::start_realtime_sync,
            realtime_sync::stop_realtime_sync,
            realtime_sync::get_online_devices,
            offline_manager::add_offline_operation,
            offline_manager::process_offline_queue,
            offline_manager::get_offline_queue,
            offline_manager::cache_data,
            offline_manager::get_cached_data,
            offline_manager::remove_cached_data,
            offline_manager::get_unresolved_conflicts,
            offline_manager::resolve_conflict,
            offline_manager::cleanup_expired_data,
            // 同步配置命令
            sync_config::get_sync_config,
            sync_config::update_sync_config,
            sync_config::get_device_config,
            sync_config::update_device_config,
            sync_config::reset_sync_config
        ])
        .setup(|_app| {
            // 应用启动时的初始化逻辑
            log::info!("太上老君AI平台桌面应用启动");
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}