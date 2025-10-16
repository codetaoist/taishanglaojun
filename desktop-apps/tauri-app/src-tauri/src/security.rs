use anyhow::{anyhow, Result};
use base64::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};
use aes_gcm::{
    aead::{Aead, KeyInit},
    Aes256Gcm, Nonce,
};
use rsa::{RsaPrivateKey, RsaPublicKey, Pkcs1v15Encrypt};
use sha2::{Digest, Sha256};
use rand::{rngs::OsRng, RngCore};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserCredentials {
    pub username: String,
    pub password_hash: String,
    pub salt: String,
    pub created_at: u64,
    pub last_login: Option<u64>,
    pub permissions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SessionToken {
    pub token: String,
    pub username: String,
    pub created_at: u64,
    pub expires_at: u64,
    pub permissions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncryptionResult {
    pub encrypted_data: String,
    pub nonce: String,
    pub algorithm: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityConfig {
    pub session_timeout: u64,
    pub max_login_attempts: u32,
    pub password_min_length: usize,
    pub require_special_chars: bool,
    pub encryption_algorithm: String,
}

pub struct SecurityManager {
    config: SecurityConfig,
    users: HashMap<String, UserCredentials>,
    sessions: HashMap<String, SessionToken>,
    login_attempts: HashMap<String, u32>,
    aes_key: [u8; 32],
    rsa_private_key: RsaPrivateKey,
    rsa_public_key: RsaPublicKey,
}

impl SecurityManager {
    pub async fn new() -> Result<Self> {
        let config = SecurityConfig {
            session_timeout: 3600, // 1小时
            max_login_attempts: 5,
            password_min_length: 8,
            require_special_chars: true,
            encryption_algorithm: "AES-256-GCM".to_string(),
        };

        // 生成AES密钥
        let mut aes_key = [0u8; 32];
        OsRng.fill_bytes(&mut aes_key);

        // 生成RSA密钥对
        let mut rng = rand::thread_rng();
        let rsa_private_key = RsaPrivateKey::new(&mut rng, 2048)?;
        let rsa_public_key = RsaPublicKey::from(&rsa_private_key);

        Ok(Self {
            config,
            users: HashMap::new(),
            sessions: HashMap::new(),
            login_attempts: HashMap::new(),
            aes_key,
            rsa_private_key,
            rsa_public_key,
        })
    }

    // 用户注册
    pub async fn register_user(
        &mut self,
        username: String,
        password: String,
        permissions: Vec<String>,
    ) -> Result<()> {
        if self.users.contains_key(&username) {
            return Err(anyhow!("用户已存在"));
        }

        if !self.validate_password(&password) {
            return Err(anyhow!("密码不符合安全要求"));
        }

        let salt = self.generate_salt();
        let password_hash = self.hash_password(&password, &salt)?;

        let credentials = UserCredentials {
            username: username.clone(),
            password_hash,
            salt,
            created_at: self.current_timestamp(),
            last_login: None,
            permissions,
        };

        self.users.insert(username, credentials);
        Ok(())
    }

    // 用户登录
    pub async fn login(&mut self, username: String, password: String) -> Result<String> {
        // 检查登录尝试次数
        let attempts = self.login_attempts.get(&username).unwrap_or(&0);
        if *attempts >= self.config.max_login_attempts {
            return Err(anyhow!("登录尝试次数过多，账户已被锁定"));
        }

        // 验证用户凭据
        let (user_salt, user_password_hash, user_permissions) = if let Some(user) = self.users.get(&username) {
            (user.salt.clone(), user.password_hash.clone(), user.permissions.clone())
        } else {
            // 用户不存在
            let attempts = self.login_attempts.entry(username).or_insert(0);
            *attempts += 1;
            return Err(anyhow!("用户名或密码错误"));
        };

        let password_hash = self.hash_password(&password, &user_salt)?;
        if password_hash == user_password_hash {
            // 登录成功
            let current_time = self.current_timestamp();
            if let Some(user) = self.users.get_mut(&username) {
                user.last_login = Some(current_time);
            }
            self.login_attempts.remove(&username);
            
            // 创建会话令牌
            let token = self.create_session_token(&username, &user_permissions)?;
            Ok(token)
        } else {
            // 密码错误
            let attempts = self.login_attempts.entry(username).or_insert(0);
            *attempts += 1;
            Err(anyhow!("用户名或密码错误"))
        }
    }

    // 用户登出
    pub async fn logout(&mut self, token: String) -> Result<()> {
        self.sessions.remove(&token);
        Ok(())
    }

    // 验证会话令牌
    pub async fn validate_token(&self, token: &str) -> Result<SessionToken> {
        if let Some(session) = self.sessions.get(token) {
            let current_time = self.current_timestamp();
            if current_time <= session.expires_at {
                Ok(session.clone())
            } else {
                Err(anyhow!("会话已过期"))
            }
        } else {
            Err(anyhow!("无效的会话令牌"))
        }
    }

    // 检查权限
    pub async fn check_permission(&self, token: &str, permission: &str) -> Result<bool> {
        let session = self.validate_token(token).await?;
        Ok(session.permissions.contains(&permission.to_string()) || 
           session.permissions.contains(&"admin".to_string()))
    }

    // AES加密
    pub async fn encrypt(&self, data: String) -> Result<String> {
        let cipher = Aes256Gcm::new_from_slice(&self.aes_key)?;
        let mut nonce_bytes = [0u8; 12];
        OsRng.fill_bytes(&mut nonce_bytes);
        let nonce = Nonce::from_slice(&nonce_bytes);

        let ciphertext = cipher
            .encrypt(nonce, data.as_bytes())
            .map_err(|e| anyhow!("加密失败: {}", e))?;

        let result = EncryptionResult {
            encrypted_data: base64::prelude::BASE64_STANDARD.encode(&ciphertext),
            nonce: base64::prelude::BASE64_STANDARD.encode(&nonce_bytes),
            algorithm: self.config.encryption_algorithm.clone(),
        };

        Ok(serde_json::to_string(&result)?)
    }

    // AES解密
    pub async fn decrypt(&self, encrypted_data: String) -> Result<String> {
        let result: EncryptionResult = serde_json::from_str(&encrypted_data)?;
        
        let cipher = Aes256Gcm::new_from_slice(&self.aes_key)?;
        let nonce_bytes = base64::prelude::BASE64_STANDARD.decode(&result.nonce)?;
        let nonce = Nonce::from_slice(&nonce_bytes);
        let ciphertext = base64::prelude::BASE64_STANDARD.decode(&result.encrypted_data)?;

        let plaintext = cipher
            .decrypt(nonce, ciphertext.as_ref())
            .map_err(|e| anyhow!("解密失败: {}", e))?;

        Ok(String::from_utf8(plaintext)?)
    }

    // RSA加密（用于密钥交换）
    pub async fn rsa_encrypt(&self, data: String) -> Result<String> {
        let mut rng = rand::thread_rng();
        let encrypted = self.rsa_public_key
            .encrypt(&mut rng, Pkcs1v15Encrypt, data.as_bytes())
            .map_err(|e| anyhow!("RSA加密失败: {}", e))?;
        
        Ok(base64::encode(&encrypted))
    }

    // RSA解密
    pub async fn rsa_decrypt(&self, encrypted_data: String) -> Result<String> {
        let ciphertext = base64::decode(&encrypted_data)?;
        let decrypted = self.rsa_private_key
            .decrypt(Pkcs1v15Encrypt, &ciphertext)
            .map_err(|e| anyhow!("RSA解密失败: {}", e))?;
        
        Ok(String::from_utf8(decrypted)?)
    }

    // 生成文件哈希
    pub async fn generate_file_hash(&self, file_path: String) -> Result<String> {
        let content = std::fs::read(&file_path)?;
        let mut hasher = Sha256::new();
        hasher.update(&content);
        let hash = hasher.finalize();
        Ok(format!("{:x}", hash))
    }

    // 验证文件完整性
    pub async fn verify_file_integrity(&self, file_path: String, expected_hash: String) -> Result<bool> {
        let actual_hash = self.generate_file_hash(file_path).await?;
        Ok(actual_hash == expected_hash)
    }

    // 清理过期会话
    pub async fn cleanup_expired_sessions(&mut self) -> Result<usize> {
        let current_time = self.current_timestamp();
        let initial_count = self.sessions.len();
        
        self.sessions.retain(|_, session| session.expires_at > current_time);
        
        Ok(initial_count - self.sessions.len())
    }

    // 更改密码
    pub async fn change_password(
        &mut self,
        username: String,
        old_password: String,
        new_password: String,
    ) -> Result<()> {
        // 先获取用户信息进行验证
        let _user_salt = if let Some(user) = self.users.get(&username) {
            let old_password_hash = self.hash_password(&old_password, &user.salt)?;
            if old_password_hash != user.password_hash {
                return Err(anyhow!("原密码错误"));
            }
            user.salt.clone()
        } else {
            return Err(anyhow!("用户不存在"));
        };

        if !self.validate_password(&new_password) {
            return Err(anyhow!("新密码不符合安全要求"));
        }

        let new_salt = self.generate_salt();
        let new_password_hash = self.hash_password(&new_password, &new_salt)?;
        
        // 现在更新用户信息
        if let Some(user) = self.users.get_mut(&username) {
            user.password_hash = new_password_hash;
            user.salt = new_salt;
        }
        
        Ok(())
    }

    // 获取公钥（用于客户端加密）
    pub async fn get_public_key(&self) -> Result<String> {
        use rsa::pkcs8::EncodePublicKey;
        let pem = self.rsa_public_key.to_public_key_pem(rsa::pkcs8::LineEnding::LF)?;
        Ok(pem)
    }

    // 私有方法

    fn validate_password(&self, password: &str) -> bool {
        if password.len() < self.config.password_min_length {
            return false;
        }

        if self.config.require_special_chars {
            let has_upper = password.chars().any(|c| c.is_uppercase());
            let has_lower = password.chars().any(|c| c.is_lowercase());
            let has_digit = password.chars().any(|c| c.is_numeric());
            let has_special = password.chars().any(|c| "!@#$%^&*()_+-=[]{}|;:,.<>?".contains(c));
            
            has_upper && has_lower && has_digit && has_special
        } else {
            true
        }
    }

    fn generate_salt(&self) -> String {
        let mut salt = [0u8; 32];
        OsRng.fill_bytes(&mut salt);
        base64::encode(&salt)
    }

    fn hash_password(&self, password: &str, salt: &str) -> Result<String> {
        let mut hasher = Sha256::new();
        hasher.update(password.as_bytes());
        hasher.update(salt.as_bytes());
        let hash = hasher.finalize();
        Ok(format!("{:x}", hash))
    }

    fn create_session_token(&mut self, username: &str, permissions: &[String]) -> Result<String> {
        let token = self.generate_token();
        let current_time = self.current_timestamp();
        let expires_at = current_time + self.config.session_timeout;

        let session = SessionToken {
            token: token.clone(),
            username: username.to_string(),
            created_at: current_time,
            expires_at,
            permissions: permissions.to_vec(),
        };

        self.sessions.insert(token.clone(), session);
        Ok(token)
    }

    fn generate_token(&self) -> String {
        let mut token_bytes = [0u8; 32];
        OsRng.fill_bytes(&mut token_bytes);
        base64::encode(&token_bytes)
    }

    fn current_timestamp(&self) -> u64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs()
    }
}