use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::net::{IpAddr, SocketAddr};
use std::path::PathBuf;
use std::sync::Arc;
use tokio::sync::{Mutex, RwLock};
use tokio::net::{TcpListener, TcpStream, UdpSocket};
use uuid::Uuid;

// 传输协议类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TransferProtocol {
    DirectWiFi,      // 直连WiFi
    Bluetooth,       // 蓝牙传输
    LocalNetwork,    // 局域网
    CloudRelay,      // 云端中继
}

// 设备信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeviceInfo {
    pub id: String,
    pub name: String,
    pub device_type: String,
    pub ip_address: IpAddr,
    pub port: u16,
    pub protocol: TransferProtocol,
    pub capabilities: Vec<String>,
    pub last_seen: String,
    pub is_trusted: bool,
}

// 传输任务状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TransferStatus {
    Pending,
    Connecting,
    Transferring,
    Paused,
    Completed,
    Failed,
    Cancelled,
}

// 文件传输任务
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferTask {
    pub id: String,
    pub file_path: String,
    pub file_name: String,
    pub file_size: u64,
    pub target_device: DeviceInfo,
    pub source_account: String,
    pub target_account: String,
    pub status: TransferStatus,
    pub progress: f64,
    pub speed: u64,
    pub created_at: String,
    pub updated_at: String,
    pub chunks_total: u32,
    pub chunks_completed: u32,
    pub error_message: Option<String>,
}

// 文件块信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileChunk {
    pub chunk_id: u32,
    pub offset: u64,
    pub size: u32,
    pub hash: String,
    pub data: Vec<u8>,
}

// 传输请求
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferRequest {
    pub request_id: String,
    pub file_info: FileInfo,
    pub source_device: DeviceInfo,
    pub source_account: String,
    pub target_account: String,
    pub encryption_key: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileInfo {
    pub name: String,
    pub size: u64,
    pub hash: String,
    pub mime_type: String,
    pub created_at: String,
}

// 设备发现服务
pub struct DeviceDiscovery {
    local_device: DeviceInfo,
    discovered_devices: Arc<RwLock<HashMap<String, DeviceInfo>>>,
    udp_socket: Arc<Mutex<Option<UdpSocket>>>,
    discovery_port: u16,
}

impl DeviceDiscovery {
    pub fn new(local_device: DeviceInfo) -> Self {
        Self {
            local_device,
            discovered_devices: Arc::new(RwLock::new(HashMap::new())),
            udp_socket: Arc::new(Mutex::new(None)),
            discovery_port: 8888,
        }
    }

    // 启动设备发现服务
    pub async fn start_discovery(&self) -> Result<()> {
        let socket = UdpSocket::bind(format!("0.0.0.0:{}", self.discovery_port)).await?;
        socket.set_broadcast(true)?;
        
        let mut socket_guard = self.udp_socket.lock().await;
        *socket_guard = Some(socket);
        drop(socket_guard);

        // 启动广播线程
        self.start_broadcast().await?;
        
        // 启动监听线程
        self.start_listening().await?;

        Ok(())
    }

    // 广播设备信息
    async fn start_broadcast(&self) -> Result<()> {
        let device_info = serde_json::to_string(&self.local_device)?;
        let broadcast_addr = format!("255.255.255.255:{}", self.discovery_port);
        
        let socket_guard = self.udp_socket.lock().await;
        if let Some(socket) = socket_guard.as_ref() {
            socket.send_to(device_info.as_bytes(), &broadcast_addr).await?;
        }
        
        Ok(())
    }

    // 监听设备发现消息
    async fn start_listening(&self) -> Result<()> {
        let socket_guard = self.udp_socket.lock().await;
        if let Some(socket) = socket_guard.as_ref() {
            let mut buf = [0; 1024];
            match socket.recv_from(&mut buf).await {
                Ok((len, _addr)) => {
                    let data = String::from_utf8_lossy(&buf[..len]);
                    if let Ok(device_info) = serde_json::from_str::<DeviceInfo>(&data) {
                        if device_info.id != self.local_device.id {
                            let mut devices = self.discovered_devices.write().await;
                            devices.insert(device_info.id.clone(), device_info);
                        }
                    }
                }
                Err(e) => {
                    eprintln!("设备发现监听错误: {}", e);
                }
            }
        }
        
        Ok(())
    }

    // 停止设备发现服务
    pub async fn stop_discovery(&self) -> Result<()> {
        let mut socket_guard = self.udp_socket.lock().await;
        *socket_guard = None;
        Ok(())
    }

    // 获取发现的设备列表
    pub async fn get_discovered_devices(&self) -> Vec<DeviceInfo> {
        let devices = self.discovered_devices.read().await;
        devices.values().cloned().collect()
    }
}

// P2P网络管理器
pub struct P2PNetworkManager {
    local_device: DeviceInfo,
    active_connections: Arc<RwLock<HashMap<String, TcpStream>>>,
    tcp_listener: Arc<Mutex<Option<TcpListener>>>,
}

impl P2PNetworkManager {
    pub fn new(local_device: DeviceInfo) -> Self {
        Self {
            local_device,
            active_connections: Arc::new(RwLock::new(HashMap::new())),
            tcp_listener: Arc::new(Mutex::new(None)),
        }
    }

    // 启动P2P服务
    pub async fn start_service(&self) -> Result<()> {
        let addr = SocketAddr::new(self.local_device.ip_address, self.local_device.port);
        let listener = TcpListener::bind(addr).await?;
        
        let mut listener_guard = self.tcp_listener.lock().await;
        *listener_guard = Some(listener);
        
        Ok(())
    }

    // 连接到远程设备
    pub async fn connect_to_device(&self, device: &DeviceInfo) -> Result<()> {
        let addr = SocketAddr::new(device.ip_address, device.port);
        let stream = TcpStream::connect(addr).await?;
        
        let mut connections = self.active_connections.write().await;
        connections.insert(device.id.clone(), stream);
        
        Ok(())
    }

    // 断开设备连接
    pub async fn disconnect_device(&self, device_id: &str) -> Result<()> {
        let mut connections = self.active_connections.write().await;
        connections.remove(device_id);
        Ok(())
    }
}

// 文件传输管理器
pub struct FileTransferManager {
    active_transfers: Arc<RwLock<HashMap<String, TransferTask>>>,
    device_discovery: Arc<DeviceDiscovery>,
    p2p_manager: Arc<P2PNetworkManager>,
    account_manager: Arc<MultiAccountManager>,
    chunk_size: u32,
}

impl FileTransferManager {
    pub fn new(local_device: DeviceInfo) -> Self {
        let device_discovery = Arc::new(DeviceDiscovery::new(local_device.clone()));
        let p2p_manager = Arc::new(P2PNetworkManager::new(local_device));
        let account_manager = Arc::new(MultiAccountManager::new());
        
        Self {
            active_transfers: Arc::new(RwLock::new(HashMap::new())),
            device_discovery,
            p2p_manager,
            account_manager,
            chunk_size: 64 * 1024, // 64KB chunks
        }
    }

    // 启动文件传输服务
    pub async fn start_service(&self) -> Result<()> {
        self.device_discovery.start_discovery().await?;
        self.p2p_manager.start_service().await?;
        Ok(())
    }

    // 发起文件传输
    pub async fn initiate_transfer(
        &self,
        file_path: PathBuf,
        target_device: DeviceInfo,
        source_account: String,
        target_account: String,
    ) -> Result<String> {
        let file_metadata = tokio::fs::metadata(&file_path).await?;
        let file_name = file_path.file_name()
            .ok_or_else(|| anyhow!("无效的文件路径"))?
            .to_string_lossy()
            .to_string();

        let task_id = Uuid::new_v4().to_string();
        let transfer_task = TransferTask {
            id: task_id.clone(),
            file_path: file_path.to_string_lossy().to_string(),
            file_name,
            file_size: file_metadata.len(),
            target_device,
            source_account,
            target_account,
            status: TransferStatus::Pending,
            progress: 0.0,
            speed: 0,
            created_at: chrono::Utc::now().to_rfc3339(),
            updated_at: chrono::Utc::now().to_rfc3339(),
            chunks_total: ((file_metadata.len() + self.chunk_size as u64 - 1) / self.chunk_size as u64) as u32,
            chunks_completed: 0,
            error_message: None,
        };

        let mut transfers = self.active_transfers.write().await;
        transfers.insert(task_id.clone(), transfer_task);

        Ok(task_id)
    }

    // 获取传输任务状态
    pub async fn get_transfer_status(&self, task_id: &str) -> Option<TransferTask> {
        let transfers = self.active_transfers.read().await;
        transfers.get(task_id).cloned()
    }

    // 获取所有传输任务
    pub async fn get_all_transfers(&self) -> Vec<TransferTask> {
        let transfers = self.active_transfers.read().await;
        transfers.values().cloned().collect()
    }

    // 暂停传输
    pub async fn pause_transfer(&self, task_id: &str) -> Result<()> {
        let mut transfers = self.active_transfers.write().await;
        if let Some(task) = transfers.get_mut(task_id) {
            task.status = TransferStatus::Paused;
            task.updated_at = chrono::Utc::now().to_rfc3339();
        }
        Ok(())
    }

    // 恢复传输
    pub async fn resume_transfer(&self, task_id: &str) -> Result<()> {
        let mut transfers = self.active_transfers.write().await;
        if let Some(task) = transfers.get_mut(task_id) {
            task.status = TransferStatus::Transferring;
            task.updated_at = chrono::Utc::now().to_rfc3339();
        }
        Ok(())
    }

    // 取消传输
    pub async fn cancel_transfer(&self, task_id: &str) -> Result<()> {
        let mut transfers = self.active_transfers.write().await;
        if let Some(task) = transfers.get_mut(task_id) {
            task.status = TransferStatus::Cancelled;
            task.updated_at = chrono::Utc::now().to_rfc3339();
        }
        Ok(())
    }

    // 切换账号
    pub async fn switch_account(&self, account_id: String) -> Result<()> {
        self.account_manager.switch_account(account_id).await
    }

    // 获取当前账号
    pub async fn get_current_account(&self) -> Option<AccountInfo> {
        self.account_manager.get_active_account().await
    }

    // 获取所有传输任务
    pub async fn get_transfer_tasks(&self) -> Vec<TransferTask> {
        let transfers = self.active_transfers.read().await;
        transfers.values().cloned().collect()
    }

    // 启动设备发现
    pub async fn start_device_discovery(&self) -> Result<()> {
        self.device_discovery.start_discovery().await
    }

    // 停止设备发现
    pub async fn stop_device_discovery(&self) -> Result<()> {
        self.device_discovery.stop_discovery().await
    }

    // 获取发现的设备
    pub async fn get_discovered_devices(&self) -> Vec<DeviceInfo> {
        self.device_discovery.get_discovered_devices().await
    }
}

// 多账号管理器
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AccountInfo {
    pub id: String,
    pub username: String,
    pub display_name: String,
    pub avatar_url: Option<String>,
    pub is_active: bool,
    pub permissions: Vec<String>,
}

pub struct MultiAccountManager {
    accounts: Arc<RwLock<HashMap<String, AccountInfo>>>,
    active_account: Arc<RwLock<Option<String>>>,
}

impl MultiAccountManager {
    pub fn new() -> Self {
        Self {
            accounts: Arc::new(RwLock::new(HashMap::new())),
            active_account: Arc::new(RwLock::new(None)),
        }
    }

    // 添加账号
    pub async fn add_account(&self, account: AccountInfo) -> Result<()> {
        let mut accounts = self.accounts.write().await;
        accounts.insert(account.id.clone(), account);
        Ok(())
    }

    // 切换活跃账号
    pub async fn switch_account(&self, account_id: String) -> Result<()> {
        let accounts = self.accounts.read().await;
        if accounts.contains_key(&account_id) {
            let mut active = self.active_account.write().await;
            *active = Some(account_id);
            Ok(())
        } else {
            Err(anyhow!("账号不存在"))
        }
    }

    // 获取当前活跃账号
    pub async fn get_active_account(&self) -> Option<AccountInfo> {
        let active_id = self.active_account.read().await;
        if let Some(id) = active_id.as_ref() {
            let accounts = self.accounts.read().await;
            accounts.get(id).cloned()
        } else {
            None
        }
    }

    // 获取所有账号
    pub async fn get_all_accounts(&self) -> Vec<AccountInfo> {
        let accounts = self.accounts.read().await;
        accounts.values().cloned().collect()
    }
}