use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::State;
use reqwest::Client;
use tokio_tungstenite::{connect_async, tungstenite::protocol::Message as WsMessage};
use futures_util::{SinkExt, StreamExt};
use tokio::sync::mpsc;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MessageType {
    Text,
    Image,
    File,
    System,
    Emoji,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChatType {
    Private,
    Group,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MessageStatus {
    Sending,
    Sent,
    Delivered,
    Read,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub id: String,
    pub chat_id: String,
    pub sender_id: String,
    pub sender_username: String,
    pub content: String,
    pub message_type: MessageType,
    pub status: MessageStatus,
    pub timestamp: String,
    pub created_at: String,
    pub updated_at: String,
    
    // 文件消息相关
    pub file_name: Option<String>,
    pub file_url: Option<String>,
    pub file_size: Option<u64>,
    
    // 回复消息相关
    pub reply_to_message_id: Option<String>,
    pub reply_to_content: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Chat {
    pub id: String,
    pub name: String,
    pub chat_type: ChatType,
    pub avatar_url: Option<String>,
    pub last_message: Option<String>,
    pub last_message_time: Option<String>,
    pub unread_count: i32,
    pub participants: Vec<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SendMessageRequest {
    pub chat_id: String,
    pub content: String,
    pub message_type: MessageType,
    pub reply_to_message_id: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateChatRequest {
    pub chat_type: ChatType,
    pub name: String,
    pub participants: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatResponse {
    pub success: bool,
    pub message: String,
    pub chats: Option<Vec<Chat>>,
    pub messages: Option<Vec<Message>>,
    pub chat: Option<Chat>,
}

#[derive(Debug, Clone)]
pub struct ChatManagerState {
    pub chats: HashMap<String, Chat>,
    pub messages: HashMap<String, Vec<Message>>, // chat_id -> messages
    pub connected: bool,
    pub current_user_id: Option<String>,
}

impl Default for ChatManagerState {
    fn default() -> Self {
        Self {
            chats: HashMap::new(),
            messages: HashMap::new(),
            connected: false,
            current_user_id: None,
        }
    }
}

pub struct ChatManager {
    pub state: Arc<Mutex<ChatManagerState>>,
    pub client: Client,
    pub api_base_url: String,
    pub ws_url: String,
    pub message_sender: Option<mpsc::UnboundedSender<String>>,
}

impl ChatManager {
    pub fn new() -> Self {
        Self {
            state: Arc::new(Mutex::new(ChatManagerState::default())),
            client: Client::new(),
            api_base_url: "http://localhost:8082".to_string(),
            ws_url: "ws://localhost:8082/ws/chat".to_string(),
            message_sender: None,
        }
    }

    pub async fn get_chat_list(&self, auth_token: &str) -> Result<ChatResponse, String> {
        let url = format!("{}/api/chats", self.api_base_url);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取聊天列表失败: HTTP {}", response.status()));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if chat_response.success {
            if let Some(chats) = &chat_response.chats {
                let mut state = self.state.lock().unwrap();
                for chat in chats {
                    state.chats.insert(chat.id.clone(), chat.clone());
                }
            }
        }

        Ok(chat_response)
    }

    pub async fn get_chat_messages(&self, auth_token: &str, chat_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<ChatResponse, String> {
        let mut url = format!("{}/api/chats/{}/messages", self.api_base_url, chat_id);
        
        let mut params = Vec::new();
        if let Some(limit) = limit {
            params.push(format!("limit={}", limit));
        }
        if let Some(offset) = offset {
            params.push(format!("offset={}", offset));
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
            return Err(format!("获取聊天消息失败: HTTP {}", response.status()));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if chat_response.success {
            if let Some(messages) = &chat_response.messages {
                let mut state = self.state.lock().unwrap();
                state.messages.insert(chat_id.to_string(), messages.clone());
            }
        }

        Ok(chat_response)
    }

    pub async fn send_message(&self, auth_token: &str, request: SendMessageRequest) -> Result<ChatResponse, String> {
        let url = format!("{}/api/chats/{}/messages", self.api_base_url, request.chat_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("发送消息失败: HTTP {}", response.status()));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(chat_response)
    }

    pub async fn create_chat(&self, auth_token: &str, request: CreateChatRequest) -> Result<ChatResponse, String> {
        let url = format!("{}/api/chats", self.api_base_url);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("创建聊天失败: HTTP {}", response.status()));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if chat_response.success {
            if let Some(chat) = &chat_response.chat {
                let mut state = self.state.lock().unwrap();
                state.chats.insert(chat.id.clone(), chat.clone());
            }
        }

        Ok(chat_response)
    }

    pub async fn delete_chat(&self, auth_token: &str, chat_id: &str) -> Result<ChatResponse, String> {
        let url = format!("{}/api/chats/{}", self.api_base_url, chat_id);
        
        let response = self.client
            .delete(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("删除聊天失败: HTTP {}", response.status()));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 从本地状态中移除
        if chat_response.success {
            let mut state = self.state.lock().unwrap();
            state.chats.remove(chat_id);
            state.messages.remove(chat_id);
        }

        Ok(chat_response)
    }

    pub async fn mark_messages_read(&self, auth_token: &str, chat_id: &str, message_ids: Vec<String>) -> Result<ChatResponse, String> {
        let url = format!("{}/api/chats/{}/messages/read", self.api_base_url, chat_id);
        
        let mut payload = HashMap::new();
        payload.insert("message_ids", message_ids);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&payload)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("标记消息已读失败: HTTP {}", response.status()));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(chat_response)
    }

    pub async fn connect_websocket(&mut self, auth_token: &str) -> Result<(), String> {
        let ws_url = format!("{}?token={}", self.ws_url, auth_token);
        
        let (ws_stream, _) = connect_async(&ws_url)
            .await
            .map_err(|e| format!("WebSocket连接失败: {}", e))?;

        let (mut ws_sender, mut ws_receiver) = ws_stream.split();
        let (tx, mut rx) = mpsc::unbounded_channel::<String>();
        
        self.message_sender = Some(tx);
        
        // 发送消息任务
        let send_task = tokio::spawn(async move {
            while let Some(message) = rx.recv().await {
                if let Err(e) = ws_sender.send(WsMessage::Text(message)).await {
                    eprintln!("发送WebSocket消息失败: {}", e);
                    break;
                }
            }
        });

        // 接收消息任务
        let state = Arc::clone(&self.state);
        let receive_task = tokio::spawn(async move {
            while let Some(message) = ws_receiver.next().await {
                match message {
                    Ok(WsMessage::Text(text)) => {
                        // 解析接收到的消息并更新状态
                        if let Ok(message) = serde_json::from_str::<Message>(&text) {
                            let mut state = state.lock().unwrap();
                            let chat_messages = state.messages.entry(message.chat_id.clone()).or_insert_with(Vec::new);
                            chat_messages.push(message);
                        }
                    }
                    Ok(WsMessage::Close(_)) => {
                        println!("WebSocket连接关闭");
                        break;
                    }
                    Err(e) => {
                        eprintln!("WebSocket错误: {}", e);
                        break;
                    }
                    _ => {}
                }
            }
        });

        // 更新连接状态
        {
            let mut state = self.state.lock().unwrap();
            state.connected = true;
        }

        // 等待任务完成（这里可以根据需要调整）
        tokio::select! {
            _ = send_task => {},
            _ = receive_task => {},
        }

        Ok(())
    }

    pub fn disconnect_websocket(&mut self) {
        self.message_sender = None;
        let mut state = self.state.lock().unwrap();
        state.connected = false;
    }

    pub fn get_cached_chats(&self) -> Vec<Chat> {
        let state = self.state.lock().unwrap();
        state.chats.values().cloned().collect()
    }

    pub fn get_cached_messages(&self, chat_id: &str) -> Vec<Message> {
        let state = self.state.lock().unwrap();
        state.messages.get(chat_id).cloned().unwrap_or_default()
    }

    pub fn is_connected(&self) -> bool {
        let state = self.state.lock().unwrap();
        state.connected
    }

    pub fn set_current_user_id(&self, user_id: String) {
        let mut state = self.state.lock().unwrap();
        state.current_user_id = Some(user_id);
    }
}

// Tauri 命令
#[tauri::command]
pub async fn chat_get_list(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
) -> Result<ChatResponse, String> {
    chat_manager.get_chat_list(&auth_token).await
}

#[tauri::command]
pub async fn chat_get_messages(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
    chat_id: String,
    limit: Option<i32>,
    offset: Option<i32>,
) -> Result<ChatResponse, String> {
    chat_manager.get_chat_messages(&auth_token, &chat_id, limit, offset).await
}

#[tauri::command]
pub async fn chat_send_message(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
    request: SendMessageRequest,
) -> Result<ChatResponse, String> {
    chat_manager.send_message(&auth_token, request).await
}

#[tauri::command]
pub async fn chat_create(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
    request: CreateChatRequest,
) -> Result<ChatResponse, String> {
    chat_manager.create_chat(&auth_token, request).await
}

#[tauri::command]
pub async fn chat_delete(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
    chat_id: String,
) -> Result<ChatResponse, String> {
    chat_manager.delete_chat(&auth_token, &chat_id).await
}

#[tauri::command]
pub async fn chat_mark_read(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
    chat_id: String,
    message_ids: Vec<String>,
) -> Result<ChatResponse, String> {
    chat_manager.mark_messages_read(&auth_token, &chat_id, message_ids).await
}

#[tauri::command]
pub fn chat_get_cached_list(chat_manager: State<'_, ChatManager>) -> Vec<Chat> {
    chat_manager.get_cached_chats()
}

#[tauri::command]
pub fn chat_get_cached_messages(
    chat_manager: State<'_, ChatManager>,
    chat_id: String,
) -> Vec<Message> {
    chat_manager.get_cached_messages(&chat_id)
}

#[tauri::command]
pub fn chat_is_connected(chat_manager: State<'_, ChatManager>) -> bool {
    chat_manager.is_connected()
}

#[tauri::command]
pub async fn chat_connect_websocket(
    chat_manager: State<'_, ChatManager>,
    auth_token: String,
) -> Result<ChatResponse, String> {
    let mut manager = chat_manager.inner().clone();
    match manager.connect_websocket(&auth_token).await {
        Ok(_) => Ok(ChatResponse {
            success: true,
            message: "WebSocket connected successfully".to_string(),
            chats: None,
            messages: None,
            chat: None,
        }),
        Err(e) => Err(e),
    }
}

#[tauri::command]
pub async fn chat_disconnect_websocket(
    chat_manager: State<'_, ChatManager>,
) -> Result<ChatResponse, String> {
    let mut manager = chat_manager.inner().clone();
    manager.disconnect_websocket();
    Ok(ChatResponse {
        success: true,
        message: "WebSocket disconnected successfully".to_string(),
        chats: None,
        messages: None,
        chat: None,
    })
}

#[tauri::command]
pub fn chat_set_current_user(
    chat_manager: State<'_, ChatManager>,
    user_id: String,
) -> Result<ChatResponse, String> {
    chat_manager.set_current_user_id(user_id);
    Ok(ChatResponse {
        success: true,
        message: "Current user set successfully".to_string(),
        chats: None,
        messages: None,
        chat: None,
    })
}