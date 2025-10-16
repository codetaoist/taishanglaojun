use anyhow::{anyhow, Result};
use rand::{RngCore, thread_rng};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

// 加密配置
#[derive(Debug, Clone)]
pub struct EncryptionConfig {
    pub algorithm: String,
    pub key_size: usize,
    pub block_size: usize,
}

impl Default for EncryptionConfig {
    fn default() -> Self {
        Self {
            algorithm: "AES-256-GCM".to_string(),
            key_size: 32, // 256 bits
            block_size: 16, // 128 bits
        }
    }
}

// 设备密钥对
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeviceKeyPair {
    pub device_id: String,
    pub public_key: Vec<u8>,
    pub private_key: Vec<u8>,
    pub created_at: String,
    pub expires_at: Option<String>,
}

// 会话密钥
#[derive(Debug, Clone)]
pub struct SessionKey {
    pub session_id: String,
    pub key: Vec<u8>,
    pub nonce: Vec<u8>,
    pub created_at: String,
    pub expires_at: String,
}

// 认证令牌
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthToken {
    pub token_id: String,
    pub device_id: String,
    pub account_id: String,
    pub permissions: Vec<String>,
    pub issued_at: String,
    pub expires_at: String,
    pub signature: String,
}

// 传输安全管理器
pub struct TransferSecurityManager {
    config: EncryptionConfig,
    device_keys: Arc<RwLock<HashMap<String, DeviceKeyPair>>>,
    session_keys: Arc<RwLock<HashMap<String, SessionKey>>>,
    trusted_devices: Arc<RwLock<HashMap<String, bool>>>,
    auth_tokens: Arc<RwLock<HashMap<String, AuthToken>>>,
}

impl TransferSecurityManager {
    pub fn new() -> Self {
        Self {
            config: EncryptionConfig::default(),
            device_keys: Arc::new(RwLock::new(HashMap::new())),
            session_keys: Arc::new(RwLock::new(HashMap::new())),
            trusted_devices: Arc::new(RwLock::new(HashMap::new())),
            auth_tokens: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    // 生成设备密钥对 (简化实现)
    pub async fn generate_device_keypair(&self, device_id: String) -> Result<DeviceKeyPair> {
        // 简化实现，返回空的密钥对
        // 在实际项目中应该实现真正的RSA密钥生成
        let keypair = DeviceKeyPair {
            device_id: device_id.clone(),
            public_key: vec![],
            private_key: vec![],
            created_at: chrono::Utc::now().to_rfc3339(),
            expires_at: None,
        };

        let mut keys = self.device_keys.write().await;
        keys.insert(device_id, keypair.clone());

        Ok(keypair)
    }

    // 生成会话密钥
    pub async fn generate_session_key(&self, session_id: String) -> Result<SessionKey> {
        let mut key = vec![0u8; self.config.key_size];
        let mut nonce = vec![0u8; 12]; // GCM nonce size
        
        thread_rng().fill_bytes(&mut key);
        thread_rng().fill_bytes(&mut nonce);

        let session_key = SessionKey {
            session_id: session_id.clone(),
            key,
            nonce,
            created_at: chrono::Utc::now().to_rfc3339(),
            expires_at: (chrono::Utc::now() + chrono::Duration::hours(24)).to_rfc3339(),
        };

        let mut keys = self.session_keys.write().await;
        keys.insert(session_id, session_key.clone());

        Ok(session_key)
    }

    // 生成AES密钥
    pub fn generate_aes_key() -> [u8; 32] {
        let mut key = [0u8; 32];
        thread_rng().fill_bytes(&mut key);
        key
    }

    // RSA加密
    pub fn rsa_encrypt(&self, data: &[u8], _public_key: &[u8]) -> Result<Vec<u8>> {
        // 简化实现，直接返回原数据
        // 在实际项目中应该实现真正的RSA加密
        Ok(data.to_vec())
    }

    // RSA解密
    pub fn rsa_decrypt(&self, encrypted_data: &[u8], _private_key: &[u8]) -> Result<Vec<u8>> {
        // 简化实现，直接返回原数据
        // 在实际项目中应该实现真正的RSA解密
        Ok(encrypted_data.to_vec())
    }

    // AES加密
    pub fn aes_encrypt(&self, data: &[u8], _key: &[u8], _nonce: &[u8]) -> Result<Vec<u8>> {
        // 简化实现，直接返回原数据
        // 在实际项目中应该实现真正的AES加密
        Ok(data.to_vec())
    }

    // AES解密
    pub fn aes_decrypt(&self, encrypted_data: &[u8], _key: &[u8], _nonce: &[u8]) -> Result<Vec<u8>> {
        // 简化实现，直接返回原数据
        // 在实际项目中应该实现真正的AES解密
        Ok(encrypted_data.to_vec())
    }

    // 加密数据 (简化实现)
    pub async fn encrypt_data(&self, data: &[u8], _session_id: &str) -> Result<Vec<u8>> {
        // 简化实现，直接返回原数据
        // 在实际项目中应该实现真正的加密
        Ok(data.to_vec())
    }

    // 解密数据 (简化实现)
    pub async fn decrypt_data(&self, encrypted_data: &[u8], _session_id: &str) -> Result<Vec<u8>> {
        // 简化实现，直接返回原数据
        // 在实际项目中应该实现真正的解密
        Ok(encrypted_data.to_vec())
    }

    // 验证设备身份
    pub async fn verify_device_identity(
        &self,
        _device_id: String,
        _public_key: Vec<u8>,
    ) -> Result<bool> {
        // 简化实现，总是返回true
        // 在实际项目中应该验证设备的公钥和身份
        Ok(true)
    }

    // 验证文件完整性
    pub async fn verify_file_integrity(
        &self,
        _file_info: crate::file_transfer::FileInfo,
    ) -> Result<bool> {
        // 简化实现，总是返回true
        // 在实际项目中应该验证文件的哈希值
        Ok(true)
    }

    // 计算文件哈希
    pub fn calculate_file_hash(&self, data: &[u8]) -> String {
        let mut hasher = Sha256::new();
        hasher.update(data);
        format!("{:x}", hasher.finalize())
    }

    // 验证数据完整性
    pub fn verify_data_integrity(&self, data: &[u8], expected_hash: &str) -> bool {
        let actual_hash = self.calculate_file_hash(data);
        actual_hash == expected_hash
    }

    // 添加信任设备
    pub async fn add_trusted_device(&self, device_id: String) -> Result<()> {
        let mut trusted = self.trusted_devices.write().await;
        trusted.insert(device_id, true);
        Ok(())
    }

    // 移除信任设备
    pub async fn remove_trusted_device(&self, device_id: &str) -> Result<()> {
        let mut trusted = self.trusted_devices.write().await;
        trusted.remove(device_id);
        Ok(())
    }

    // 检查设备是否受信任
    pub async fn is_device_trusted(&self, device_id: &str) -> bool {
        let trusted = self.trusted_devices.read().await;
        trusted.get(device_id).copied().unwrap_or(false)
    }

    // 生成认证令牌
    pub async fn generate_auth_token(
        &self,
        device_id: String,
        account_id: String,
        permissions: Vec<String>,
    ) -> Result<AuthToken> {
        let token_id = uuid::Uuid::new_v4().to_string();
        let issued_at = chrono::Utc::now().to_rfc3339();
        let expires_at = (chrono::Utc::now() + chrono::Duration::hours(24)).to_rfc3339();

        // 创建签名数据
        let signature_data = format!("{}:{}:{}:{}:{}", 
            token_id, device_id, account_id, issued_at, expires_at);
        let signature = self.calculate_file_hash(signature_data.as_bytes());

        let token = AuthToken {
            token_id: token_id.clone(),
            device_id,
            account_id,
            permissions,
            issued_at,
            expires_at,
            signature,
        };

        let mut tokens = self.auth_tokens.write().await;
        tokens.insert(token_id, token.clone());

        Ok(token)
    }

    // 验证认证令牌
    pub async fn verify_auth_token(&self, token_id: &str) -> Result<bool> {
        let tokens = self.auth_tokens.read().await;
        let token = tokens.get(token_id)
            .ok_or_else(|| anyhow!("令牌不存在"))?;

        // 检查过期时间
        let expires_at = chrono::DateTime::parse_from_rfc3339(&token.expires_at)?;
        if chrono::Utc::now() > expires_at {
            return Ok(false);
        }

        // 验证签名
        let signature_data = format!("{}:{}:{}:{}:{}", 
            token.token_id, token.device_id, token.account_id, 
            token.issued_at, token.expires_at);
        let expected_signature = self.calculate_file_hash(signature_data.as_bytes());

        Ok(token.signature == expected_signature)
    }

    // 获取令牌信息
    pub async fn get_token_info(&self, token_id: &str) -> Option<AuthToken> {
        let tokens = self.auth_tokens.read().await;
        tokens.get(token_id).cloned()
    }

    // 撤销令牌
    pub async fn revoke_token(&self, token_id: &str) -> Result<()> {
        let mut tokens = self.auth_tokens.write().await;
        tokens.remove(token_id);
        Ok(())
    }

    // 清理过期令牌
    pub async fn cleanup_expired_tokens(&self) -> Result<()> {
        let mut tokens = self.auth_tokens.write().await;
        let now = chrono::Utc::now();
        
        tokens.retain(|_, token| {
            if let Ok(expires_at) = chrono::DateTime::parse_from_rfc3339(&token.expires_at) {
                now <= expires_at
            } else {
                false
            }
        });

        Ok(())
    }

    // 清理过期会话密钥
    pub async fn cleanup_expired_sessions(&self) -> Result<()> {
        let mut keys = self.session_keys.write().await;
        let now = chrono::Utc::now();
        
        keys.retain(|_, session| {
            if let Ok(expires_at) = chrono::DateTime::parse_from_rfc3339(&session.expires_at) {
                now <= expires_at
            } else {
                false
            }
        });

        Ok(())
    }
}

// 安全传输协议
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SecureMessage {
    HandshakeRequest {
        device_id: String,
        public_key: Vec<u8>,
        timestamp: String,
    },
    HandshakeResponse {
        device_id: String,
        public_key: Vec<u8>,
        session_key: Vec<u8>,
        timestamp: String,
    },
    TransferRequest {
        session_id: String,
        file_info: crate::file_transfer::FileInfo,
        auth_token: String,
    },
    TransferResponse {
        session_id: String,
        accepted: bool,
        reason: Option<String>,
    },
    FileChunk {
        session_id: String,
        chunk_id: u32,
        encrypted_data: Vec<u8>,
        hash: String,
    },
    ChunkAck {
        session_id: String,
        chunk_id: u32,
        success: bool,
    },
    TransferComplete {
        session_id: String,
        file_hash: String,
    },
    Error {
        session_id: String,
        error_code: u32,
        message: String,
    },
}

// 安全传输处理器
pub struct SecureTransferHandler {
    security_manager: Arc<TransferSecurityManager>,
}

impl SecureTransferHandler {
    pub fn new(security_manager: Arc<TransferSecurityManager>) -> Self {
        Self {
            security_manager,
        }
    }

    // 处理握手请求
    pub async fn handle_handshake_request(
        &self,
        device_id: String,
        public_key: Vec<u8>,
    ) -> Result<SecureMessage> {
        // 生成会话密钥
        let session_id = uuid::Uuid::new_v4().to_string();
        let session_key = self.security_manager.generate_session_key(session_id.clone()).await?;

        // 获取本地设备密钥
        let local_keypair = self.security_manager.generate_device_keypair("local".to_string()).await?;

        Ok(SecureMessage::HandshakeResponse {
            device_id: "local".to_string(),
            public_key: local_keypair.public_key,
            session_key: session_key.key,
            timestamp: chrono::Utc::now().to_rfc3339(),
        })
    }

    // 处理传输请求
    pub async fn handle_transfer_request(
        &self,
        session_id: String,
        file_info: crate::file_transfer::FileInfo,
        auth_token: String,
    ) -> Result<SecureMessage> {
        // 验证认证令牌
        let is_valid = self.security_manager.verify_auth_token(&auth_token).await?;
        
        if !is_valid {
            return Ok(SecureMessage::TransferResponse {
                session_id,
                accepted: false,
                reason: Some("无效的认证令牌".to_string()),
            });
        }

        Ok(SecureMessage::TransferResponse {
            session_id,
            accepted: true,
            reason: None,
        })
    }

    // 处理文件块
    pub async fn handle_file_chunk(
        &self,
        session_id: String,
        chunk_id: u32,
        encrypted_data: Vec<u8>,
        expected_hash: String,
    ) -> Result<SecureMessage> {
        // 解密数据
        match self.security_manager.decrypt_data(&encrypted_data, &session_id).await {
            Ok(decrypted_data) => {
                // 验证哈希
                let actual_hash = self.security_manager.calculate_file_hash(&decrypted_data);
                let success = actual_hash == expected_hash;

                Ok(SecureMessage::ChunkAck {
                    session_id,
                    chunk_id,
                    success,
                })
            }
            Err(_) => {
                Ok(SecureMessage::ChunkAck {
                    session_id,
                    chunk_id,
                    success: false,
                })
            }
        }
    }
}