// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod ai_service;
mod auth;
mod chat;
mod document;
mod image_gen;
mod image_generator;
mod security;
mod storage;
mod utils;

use ai_service::AIService;
use auth::AuthManager;
use chat::ChatManager;
use document::DocumentProcessor;
use image_gen::ImageGenerator;
use image_generator::ImageGenerator;
use security::SecurityManager;
use storage::StorageManager;

use std::sync::Arc;
use tauri::{Manager, State};
use tokio::sync::Mutex;

// 应用状态
pub struct AppState {
    ai_service: Arc<Mutex<AIService>>,
    auth_manager: Arc<Mutex<AuthManager>>,
    chat_manager: Arc<Mutex<ChatManager>>,
    document_processor: Arc<Mutex<DocumentProcessor>>,
    image_generator: Arc<Mutex<ImageGenerator>>,
    security_manager: Arc<Mutex<SecurityManager>>,
    storage_manager: Arc<Mutex<StorageManager>>,
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
    let auth_manager = state.auth_manager.lock().await;
    auth_manager
        .login(username, password)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn logout(state: State<'_, AppState>) -> Result<(), String> {
    let auth_manager = state.auth_manager.lock().await;
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
    let storage_manager = state.storage_manager.lock().await;
    storage_manager
        .save_file(path, content)
        .await
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

#[tokio::main]
async fn main() {
    env_logger::init();

    // 初始化应用状态
    let app_state = AppState {
        ai_service: Arc::new(Mutex::new(AIService::new().await.unwrap())),
        auth_manager: Arc::new(Mutex::new(AuthManager::new().await.unwrap())),
        chat_manager: Arc::new(Mutex::new(ChatManager::new().await.unwrap())),
        document_processor: Arc::new(Mutex::new(DocumentProcessor::new().await.unwrap())),
        image_generator: Arc::new(Mutex::new(ImageGenerator::new().await.unwrap())),
        security_manager: Arc::new(Mutex::new(SecurityManager::new().await.unwrap())),
        storage_manager: Arc::new(Mutex::new(StorageManager::new().await.unwrap())),
    };

    tauri::Builder::default()
        .manage(app_state)
        .invoke_handler(tauri::generate_handler![
            chat_with_ai,
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
            decrypt_data
        ])
        .setup(|app| {
            // 应用启动时的初始化逻辑
            log::info!("太上老君AI平台桌面应用启动");
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}