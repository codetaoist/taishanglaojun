use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::{broadcast, RwLock};
use tokio_tungstenite::{connect_async, tungstenite::Message};
use futures_util::{SinkExt, StreamExt};
use uuid::Uuid;
use chrono::{DateTime, Utc};
use anyhow::Result;

use crate::sync_service::{MultiDeviceSyncService, SyncRecord, DeviceInfo};

/// 实时同步消息类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RealtimeSyncMessage {
    /// 设备上线
    DeviceOnline { device_id: String, user_id: String },
    /// 设备下线
    DeviceOffline { device_id: String, user_id: String },
    /// 数据更新
    DataUpdate { sync_record: SyncRecord },
    /// 聊天消息
    ChatMessage { 
        message_id: String,
        session_id: String,
        sender_id: String,
        content: String,
        timestamp: DateTime<Utc>,
    },
    /// 好友状态更新
    FriendStatusUpdate {
        friend_id: String,
        status: String,
        last_seen: DateTime<Utc>,
    },
    /// 同步请求
    SyncRequest {
        device_id: String,
        last_sync: DateTime<Utc>,
    },
    /// 同步响应
    SyncResponse {
        records: Vec<SyncRecord>,
        next_sync_token: String,
    },
    /// 心跳
    Heartbeat { device_id: String, timestamp: DateTime<Utc> },
}

/// 设备连接状态
#[derive(Debug, Clone)]
pub struct DeviceConnection {
    pub device_id: String,
    pub user_id: String,
    pub connected_at: DateTime<Utc>,
    pub last_heartbeat: DateTime<Utc>,
    pub sender: broadcast::Sender<RealtimeSyncMessage>,
}

/// 实时同步管理器
pub struct RealtimeSyncManager {
    sync_service: Arc<MultiDeviceSyncService>,
    connections: Arc<RwLock<HashMap<String, DeviceConnection>>>,
    user_devices: Arc<RwLock<HashMap<String, Vec<String>>>>, // user_id -> device_ids
    message_broadcaster: broadcast::Sender<RealtimeSyncMessage>,
}

impl RealtimeSyncManager {
    pub fn new(sync_service: Arc<MultiDeviceSyncService>) -> Self {
        let (message_broadcaster, _) = broadcast::channel(1000);
        
        Self {
            sync_service,
            connections: Arc::new(RwLock::new(HashMap::new())),
            user_devices: Arc::new(RwLock::new(HashMap::new())),
            message_broadcaster,
        }
    }

    /// 启动实时同步服务
    pub async fn start_service(&self, port: u16) -> Result<()> {
        let addr = format!("127.0.0.1:{}", port);
        let listener = tokio::net::TcpListener::bind(&addr).await?;
        
        println!("实时同步服务启动在: {}", addr);

        while let Ok((stream, _)) = listener.accept().await {
            let connections = Arc::clone(&self.connections);
            let user_devices = Arc::clone(&self.user_devices);
            let sync_service = Arc::clone(&self.sync_service);
            let broadcaster = self.message_broadcaster.clone();

            tokio::spawn(async move {
                if let Err(e) = Self::handle_connection(
                    stream, 
                    connections, 
                    user_devices, 
                    sync_service, 
                    broadcaster
                ).await {
                    eprintln!("连接处理错误: {}", e);
                }
            });
        }

        Ok(())
    }

    /// 处理WebSocket连接
    async fn handle_connection(
        stream: tokio::net::TcpStream,
        connections: Arc<RwLock<HashMap<String, DeviceConnection>>>,
        user_devices: Arc<RwLock<HashMap<String, Vec<String>>>>,
        sync_service: Arc<MultiDeviceSyncService>,
        broadcaster: broadcast::Sender<RealtimeSyncMessage>,
    ) -> Result<()> {
        let ws_stream = tokio_tungstenite::accept_async(stream).await?;
        let (mut ws_sender, mut ws_receiver) = ws_stream.split();

        let mut device_id: Option<String> = None;
        let mut user_id: Option<String> = None;
        let mut receiver = broadcaster.subscribe();

        loop {
            tokio::select! {
                // 处理来自客户端的消息
                msg = ws_receiver.next() => {
                    match msg {
                        Some(Ok(Message::Text(text))) => {
                            if let Ok(sync_msg) = serde_json::from_str::<RealtimeSyncMessage>(&text) {
                                match sync_msg {
                                    RealtimeSyncMessage::DeviceOnline { device_id: did, user_id: uid } => {
                                        device_id = Some(did.clone());
                                        user_id = Some(uid.clone());

                                        // 注册设备连接
                                        let (sender, _) = broadcast::channel(100);
                                        let connection = DeviceConnection {
                                            device_id: did.clone(),
                                            user_id: uid.clone(),
                                            connected_at: Utc::now(),
                                            last_heartbeat: Utc::now(),
                                            sender: sender.clone(),
                                        };

                                        {
                                            let mut conns = connections.write().await;
                                            conns.insert(did.clone(), connection);
                                        }

                                        {
                                            let mut user_devs = user_devices.write().await;
                                            user_devs.entry(uid.clone())
                                                .or_insert_with(Vec::new)
                                                .push(did.clone());
                                        }

                                        // 广播设备上线消息
                                        let _ = broadcaster.send(sync_msg);
                                    }
                                    RealtimeSyncMessage::SyncRequest { device_id: req_device_id, last_sync } => {
                                        if let Some(uid) = &user_id {
                                            // 获取增量同步数据
                                            if let Ok(records) = sync_service.get_incremental_sync(
                                                uid, 
                                                &req_device_id, 
                                                last_sync
                                            ).await {
                                                let response = RealtimeSyncMessage::SyncResponse {
                                                    records,
                                                    next_sync_token: Uuid::new_v4().to_string(),
                                                };
                                                
                                                let response_text = serde_json::to_string(&response)?;
                                                ws_sender.send(Message::Text(response_text)).await?;
                                            }
                                        }
                                    }
                                    RealtimeSyncMessage::ChatMessage { .. } => {
                                        // 转发聊天消息给用户的其他设备
                                        if let Some(uid) = &user_id {
                                            Self::broadcast_to_user_devices(
                                                &user_devices,
                                                &connections,
                                                uid,
                                                &device_id.as_ref().unwrap_or(&String::new()),
                                                sync_msg
                                            ).await;
                                        }
                                    }
                                    RealtimeSyncMessage::Heartbeat { device_id: hb_device_id, .. } => {
                                        // 更新心跳时间
                                        let mut conns = connections.write().await;
                                        if let Some(conn) = conns.get_mut(&hb_device_id) {
                                            conn.last_heartbeat = Utc::now();
                                        }
                                    }
                                    _ => {
                                        // 其他消息类型的处理
                                        let _ = broadcaster.send(sync_msg);
                                    }
                                }
                            }
                        }
                        Some(Ok(Message::Close(_))) | None => {
                            break;
                        }
                        Some(Err(e)) => {
                            eprintln!("WebSocket错误: {}", e);
                            break;
                        }
                        _ => {}
                    }
                }
                // 处理广播消息
                broadcast_msg = receiver.recv() => {
                    if let Ok(msg) = broadcast_msg {
                        // 只转发给相关的设备
                        if Self::should_forward_message(&msg, &device_id, &user_id) {
                            let msg_text = serde_json::to_string(&msg)?;
                            if let Err(e) = ws_sender.send(Message::Text(msg_text)).await {
                                eprintln!("发送消息失败: {}", e);
                                break;
                            }
                        }
                    }
                }
            }
        }

        // 清理连接
        if let (Some(did), Some(uid)) = (device_id, user_id) {
            {
                let mut conns = connections.write().await;
                conns.remove(&did);
            }

            {
                let mut user_devs = user_devices.write().await;
                if let Some(devices) = user_devs.get_mut(&uid) {
                    devices.retain(|d| d != &did);
                    if devices.is_empty() {
                        user_devs.remove(&uid);
                    }
                }
            }

            // 广播设备下线消息
            let offline_msg = RealtimeSyncMessage::DeviceOffline { 
                device_id: did, 
                user_id: uid 
            };
            let _ = broadcaster.send(offline_msg);
        }

        Ok(())
    }

    /// 向用户的其他设备广播消息
    async fn broadcast_to_user_devices(
        user_devices: &Arc<RwLock<HashMap<String, Vec<String>>>>,
        connections: &Arc<RwLock<HashMap<String, DeviceConnection>>>,
        user_id: &str,
        exclude_device_id: &str,
        message: RealtimeSyncMessage,
    ) {
        let user_devs = user_devices.read().await;
        if let Some(device_ids) = user_devs.get(user_id) {
            let conns = connections.read().await;
            
            for device_id in device_ids {
                if device_id != exclude_device_id {
                    if let Some(conn) = conns.get(device_id) {
                        let _ = conn.sender.send(message.clone());
                    }
                }
            }
        }
    }

    /// 判断是否应该转发消息
    fn should_forward_message(
        msg: &RealtimeSyncMessage,
        device_id: &Option<String>,
        user_id: &Option<String>,
    ) -> bool {
        match msg {
            RealtimeSyncMessage::ChatMessage { .. } => true,
            RealtimeSyncMessage::FriendStatusUpdate { .. } => true,
            RealtimeSyncMessage::DataUpdate { sync_record } => {
                // 只转发给同一用户的其他设备
                if let Some(uid) = user_id {
                    sync_record.user_id == *uid && 
                    Some(&sync_record.device_id) != device_id.as_ref()
                } else {
                    false
                }
            }
            RealtimeSyncMessage::DeviceOnline { user_id: msg_user_id, .. } |
            RealtimeSyncMessage::DeviceOffline { user_id: msg_user_id, .. } => {
                user_id.as_ref() == Some(msg_user_id)
            }
            _ => false,
        }
    }

    /// 获取在线设备列表
    pub async fn get_online_devices(&self, user_id: &str) -> Vec<DeviceInfo> {
        let user_devs = self.user_devices.read().await;
        let conns = self.connections.read().await;
        
        let mut online_devices = Vec::new();
        
        if let Some(device_ids) = user_devs.get(user_id) {
            for device_id in device_ids {
                if let Some(conn) = conns.get(device_id) {
                    // 检查心跳是否在5分钟内
                    let now = Utc::now();
                    if (now - conn.last_heartbeat).num_minutes() < 5 {
                        if let Ok(devices) = self.sync_service.get_user_devices(user_id).await {
                            for device in devices {
                                if device.device_id == *device_id {
                                    online_devices.push(device);
                                    break;
                                }
                            }
                        }
                    }
                }
            }
        }
        
        online_devices
    }

    /// 发送消息到特定设备
    pub async fn send_to_device(
        &self,
        device_id: &str,
        message: RealtimeSyncMessage,
    ) -> Result<()> {
        let conns = self.connections.read().await;
        if let Some(conn) = conns.get(device_id) {
            conn.sender.send(message)?;
        }
        Ok(())
    }

    /// 发送消息到用户的所有设备
    pub async fn send_to_user(
        &self,
        user_id: &str,
        message: RealtimeSyncMessage,
    ) -> Result<()> {
        self.message_broadcaster.send(message)?;
        Ok(())
    }
}

// Tauri命令函数
#[tauri::command]
pub async fn start_realtime_sync(
    port: Option<u16>,
    realtime_sync: tauri::State<'_, std::sync::Arc<RealtimeSyncManager>>,
) -> Result<(), String> {
    let port = port.unwrap_or(8080);
    realtime_sync.start_service(port).await
        .map_err(|e| format!("Failed to start realtime sync: {}", e))
}

#[tauri::command]
pub async fn stop_realtime_sync(
    realtime_sync: tauri::State<'_, std::sync::Arc<RealtimeSyncManager>>,
) -> Result<(), String> {
    // 停止实时同步服务
    println!("Stopping realtime sync service");
    Ok(())
}

#[tauri::command]
pub async fn get_online_devices(
    user_id: String,
    realtime_sync: tauri::State<'_, std::sync::Arc<RealtimeSyncManager>>,
) -> Result<Vec<DeviceInfo>, String> {
    Ok(realtime_sync.get_online_devices(&user_id).await)
}

#[tauri::command]
pub async fn send_message_to_device(
    device_id: String,
    message: RealtimeSyncMessage,
    realtime_sync: tauri::State<'_, std::sync::Arc<RealtimeSyncManager>>,
) -> Result<(), String> {
    realtime_sync.send_to_device(&device_id, message).await
        .map_err(|e| format!("Failed to send message to device: {}", e))
}

#[tauri::command]
pub async fn send_message_to_user(
    user_id: String,
    message: RealtimeSyncMessage,
    realtime_sync: tauri::State<'_, std::sync::Arc<RealtimeSyncManager>>,
) -> Result<(), String> {
    realtime_sync.send_to_user(&user_id, message).await
        .map_err(|e| format!("Failed to send message to user: {}", e))
}