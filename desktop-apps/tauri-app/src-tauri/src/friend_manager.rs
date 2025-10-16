use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::State;
use reqwest::Client;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum FriendStatus {
    Pending,    // 待确认
    Accepted,   // 已接受
    Blocked,    // 已屏蔽
    Declined,   // 已拒绝
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum OnlineStatus {
    Online,     // 在线
    Offline,    // 离线
    Away,       // 离开
    Busy,       // 忙碌
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Friend {
    pub id: String,
    pub username: String,
    pub email: String,
    pub avatar_url: Option<String>,
    pub status: FriendStatus,
    pub online_status: OnlineStatus,
    pub last_seen: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FriendRequest {
    pub id: String,
    pub from_user_id: String,
    pub to_user_id: String,
    pub from_username: String,
    pub to_username: String,
    pub message: Option<String>,
    pub status: FriendStatus,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AddFriendRequest {
    pub username: String,
    pub message: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FriendResponse {
    pub success: bool,
    pub message: String,
    pub friends: Option<Vec<Friend>>,
    pub requests: Option<Vec<FriendRequest>>,
}

#[derive(Debug, Clone)]
pub struct FriendManagerState {
    pub friends: Vec<Friend>,
    pub friend_requests: Vec<FriendRequest>,
    pub last_sync: Option<std::time::Instant>,
}

impl Default for FriendManagerState {
    fn default() -> Self {
        Self {
            friends: Vec::new(),
            friend_requests: Vec::new(),
            last_sync: None,
        }
    }
}

pub struct FriendManager {
    pub state: Arc<Mutex<FriendManagerState>>,
    pub client: Client,
    pub api_base_url: String,
}

impl FriendManager {
    pub fn new() -> Self {
        Self {
            state: Arc::new(Mutex::new(FriendManagerState::default())),
            client: Client::new(),
            api_base_url: "http://localhost:8082".to_string(),
        }
    }

    pub async fn get_friend_list(&self, auth_token: &str) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends", self.api_base_url);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取好友列表失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if friend_response.success {
            if let Some(friends) = &friend_response.friends {
                let mut state = self.state.lock().unwrap();
                state.friends = friends.clone();
                state.last_sync = Some(std::time::Instant::now());
            }
        }

        Ok(friend_response)
    }

    pub async fn get_friend_requests(&self, auth_token: &str) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends/requests", self.api_base_url);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取好友请求失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if friend_response.success {
            if let Some(requests) = &friend_response.requests {
                let mut state = self.state.lock().unwrap();
                state.friend_requests = requests.clone();
            }
        }

        Ok(friend_response)
    }

    pub async fn add_friend(&self, auth_token: &str, request: AddFriendRequest) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends/add", self.api_base_url);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("添加好友失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(friend_response)
    }

    pub async fn respond_to_friend_request(&self, auth_token: &str, request_id: &str, accept: bool) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends/requests/{}/respond", self.api_base_url, request_id);
        
        let mut payload = HashMap::new();
        payload.insert("accept", accept);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&payload)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("响应好友请求失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 如果成功，刷新本地数据
        if friend_response.success {
            self.refresh_data(auth_token).await.ok();
        }

        Ok(friend_response)
    }

    pub async fn remove_friend(&self, auth_token: &str, friend_id: &str) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends/{}", self.api_base_url, friend_id);
        
        let response = self.client
            .delete(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("删除好友失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 如果成功，从本地状态中移除
        if friend_response.success {
            let mut state = self.state.lock().unwrap();
            state.friends.retain(|f| f.id != friend_id);
        }

        Ok(friend_response)
    }

    pub async fn block_friend(&self, auth_token: &str, friend_id: &str) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends/{}/block", self.api_base_url, friend_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("屏蔽好友失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 如果成功，更新本地状态
        if friend_response.success {
            let mut state = self.state.lock().unwrap();
            if let Some(friend) = state.friends.iter_mut().find(|f| f.id == friend_id) {
                friend.status = FriendStatus::Blocked;
            }
        }

        Ok(friend_response)
    }

    pub async fn unblock_friend(&self, auth_token: &str, friend_id: &str) -> Result<FriendResponse, String> {
        let url = format!("{}/api/friends/{}/unblock", self.api_base_url, friend_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("取消屏蔽好友失败: HTTP {}", response.status()));
        }

        let friend_response: FriendResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 如果成功，更新本地状态
        if friend_response.success {
            let mut state = self.state.lock().unwrap();
            if let Some(friend) = state.friends.iter_mut().find(|f| f.id == friend_id) {
                friend.status = FriendStatus::Accepted;
            }
        }

        Ok(friend_response)
    }

    pub fn get_cached_friends(&self) -> Vec<Friend> {
        let state = self.state.lock().unwrap();
        state.friends.clone()
    }

    pub fn get_cached_requests(&self) -> Vec<FriendRequest> {
        let state = self.state.lock().unwrap();
        state.friend_requests.clone()
    }

    pub async fn refresh_data(&self, auth_token: &str) -> Result<(), String> {
        // 并行获取好友列表和好友请求
        let friends_future = self.get_friend_list(auth_token);
        let requests_future = self.get_friend_requests(auth_token);
        
        let (friends_result, requests_result) = tokio::join!(friends_future, requests_future);
        
        friends_result?;
        requests_result?;
        
        Ok(())
    }

    pub fn set_api_base_url(&mut self, url: String) {
        self.api_base_url = url;
    }
}

// Tauri 命令
#[tauri::command]
pub async fn friend_get_list(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
) -> Result<FriendResponse, String> {
    friend_manager.get_friend_list(&auth_token).await
}

#[tauri::command]
pub async fn friend_get_requests(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
) -> Result<FriendResponse, String> {
    friend_manager.get_friend_requests(&auth_token).await
}

#[tauri::command]
pub async fn friend_add(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
    request: AddFriendRequest,
) -> Result<FriendResponse, String> {
    friend_manager.add_friend(&auth_token, request).await
}

#[tauri::command]
pub async fn friend_respond_request(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
    request_id: String,
    accept: bool,
) -> Result<FriendResponse, String> {
    friend_manager.respond_to_friend_request(&auth_token, &request_id, accept).await
}

#[tauri::command]
pub async fn friend_remove(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
    friend_id: String,
) -> Result<FriendResponse, String> {
    friend_manager.remove_friend(&auth_token, &friend_id).await
}

#[tauri::command]
pub async fn friend_block(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
    friend_id: String,
) -> Result<FriendResponse, String> {
    friend_manager.block_friend(&auth_token, &friend_id).await
}

#[tauri::command]
pub async fn friend_unblock(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
    friend_id: String,
) -> Result<FriendResponse, String> {
    friend_manager.unblock_friend(&auth_token, &friend_id).await
}

#[tauri::command]
pub fn friend_get_cached_list(friend_manager: State<'_, FriendManager>) -> Vec<Friend> {
    friend_manager.get_cached_friends()
}

#[tauri::command]
pub fn friend_get_cached_requests(friend_manager: State<'_, FriendManager>) -> Vec<FriendRequest> {
    friend_manager.get_cached_requests()
}

#[tauri::command]
pub async fn friend_refresh_data(
    friend_manager: State<'_, FriendManager>,
    auth_token: String,
) -> Result<(), String> {
    friend_manager.refresh_data(&auth_token).await
}