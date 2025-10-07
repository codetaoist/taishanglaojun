# 太上老君AI平台 - 安全配置指南

## 概述

本文档详细介绍太上老君AI平台的安全配置，包括网络安全、身份认证、数据加密、访问控制、安全监控等方面的配置和最佳实践。

## 安全架构

```mermaid
graph TB
    subgraph "外部访问层"
        A[CDN/WAF]
        B[负载均衡器]
        C[防火墙]
    end
    
    subgraph "网络安全层"
        D[Ingress Controller]
        E[Network Policies]
        F[Service Mesh]
    end
    
    subgraph "身份认证层"
        G[OAuth 2.0]
        H[JWT Token]
        I[RBAC]
        J[API Gateway]
    end
    
    subgraph "应用安全层"
        K[前端应用]
        L[后端API]
        M[AI服务]
        N[数据库]
    end
    
    subgraph "数据安全层"
        O[数据加密]
        P[密钥管理]
        Q[备份加密]
        R[审计日志]
    end
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
    G --> H
    H --> I
    I --> J
    J --> K
    J --> L
    J --> M
    L --> N
    N --> O
    O --> P
    P --> Q
    Q --> R
```

## 网络安全

### 1. 防火墙配置

```yaml
# security/firewall-rules.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: taishanglaojun-prod
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend
  namespace: taishanglaojun-prod
spec:
  podSelector:
    matchLabels:
      app: frontend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 3000
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 8080
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-backend
  namespace: taishanglaojun-prod
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    - podSelector:
        matchLabels:
          app: ai-service
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
  - to:
    - podSelector:
        matchLabels:
          app: ai-service
    ports:
    - protocol: TCP
      port: 8000

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-ai-service
  namespace: taishanglaojun-prod
spec:
  podSelector:
    matchLabels:
      app: ai-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 8000
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: qdrant
    ports:
    - protocol: TCP
      port: 6333
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 80

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-database
  namespace: taishanglaojun-prod
spec:
  podSelector:
    matchLabels:
      app: postgres
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 5432
```

### 2. TLS/SSL配置

```yaml
# security/tls-certificates.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@taishanglaojun.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
    - dns01:
        cloudflare:
          email: admin@taishanglaojun.com
          apiTokenSecretRef:
            name: cloudflare-api-token
            key: api-token

---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: taishanglaojun-tls
  namespace: taishanglaojun-prod
spec:
  secretName: taishanglaojun-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - taishanglaojun.com
  - www.taishanglaojun.com
  - api.taishanglaojun.com
  - ai.taishanglaojun.com

---
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-api-token
  namespace: cert-manager
type: Opaque
stringData:
  api-token: "your-cloudflare-api-token"
```

### 3. WAF配置

```yaml
# security/waf-config.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: taishanglaojun-ingress
  namespace: taishanglaojun-prod
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    
    # WAF规则
    nginx.ingress.kubernetes.io/enable-modsecurity: "true"
    nginx.ingress.kubernetes.io/enable-owasp-core-rules: "true"
    nginx.ingress.kubernetes.io/modsecurity-transaction-id: "$request_id"
    nginx.ingress.kubernetes.io/modsecurity-snippet: |
      SecRuleEngine On
      SecRequestBodyAccess On
      SecRule REQUEST_HEADERS:Content-Type "text/xml" \
        "id:'200001',phase:1,t:none,t:lowercase,pass,nolog,ctl:requestBodyProcessor=XML"
      SecRule REQUEST_HEADERS:Content-Type "application/json" \
        "id:'200002',phase:1,t:none,t:lowercase,pass,nolog,ctl:requestBodyProcessor=JSON"
      SecRule REQBODY_ERROR "!@eq 0" \
        "id:'200003',phase:2,t:none,log,deny,status:400,msg:'Failed to parse request body.',logdata:'Error %{REQBODY_ERROR_MSG}',severity:2"
      SecRule MULTIPART_STRICT_ERROR "!@eq 0" \
        "id:'200004',phase:2,t:none,log,deny,status:400,msg:'Multipart request body failed strict validation: \
        PE %{REQBODY_PROCESSOR_ERROR}, \
        BQ %{MULTIPART_BOUNDARY_QUOTED}, \
        BW %{MULTIPART_BOUNDARY_WHITESPACE}, \
        DB %{MULTIPART_DATA_BEFORE}, \
        DA %{MULTIPART_DATA_AFTER}, \
        HF %{MULTIPART_HEADER_FOLDING}, \
        LF %{MULTIPART_LF_LINE}, \
        SM %{MULTIPART_MISSING_SEMICOLON}, \
        IQ %{MULTIPART_INVALID_QUOTING}, \
        IP %{MULTIPART_INVALID_PART}, \
        IH %{MULTIPART_INVALID_HEADER_FOLDING}, \
        FL %{MULTIPART_FILE_LIMIT_EXCEEDED}'"
    
    # 速率限制
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/rate-limit-connections: "10"
    
    # 请求大小限制
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/client-body-buffer-size: "1m"
    
    # 安全头
    nginx.ingress.kubernetes.io/configuration-snippet: |
      more_set_headers "X-Frame-Options: DENY";
      more_set_headers "X-Content-Type-Options: nosniff";
      more_set_headers "X-XSS-Protection: 1; mode=block";
      more_set_headers "Strict-Transport-Security: max-age=31536000; includeSubDomains";
      more_set_headers "Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss: https:";
      more_set_headers "Referrer-Policy: strict-origin-when-cross-origin";
      more_set_headers "Permissions-Policy: geolocation=(), microphone=(), camera=()";

spec:
  tls:
  - hosts:
    - taishanglaojun.com
    - www.taishanglaojun.com
    - api.taishanglaojun.com
    - ai.taishanglaojun.com
    secretName: taishanglaojun-tls
  
  rules:
  - host: taishanglaojun.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
  
  - host: api.taishanglaojun.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 8080
  
  - host: ai.taishanglaojun.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ai-service
            port:
              number: 8000
```

## 身份认证与授权

### 1. OAuth 2.0配置

```go
// pkg/auth/oauth.go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "golang.org/x/oauth2/github"
)

type OAuthConfig struct {
    GoogleConfig *oauth2.Config
    GitHubConfig *oauth2.Config
    StateStore   map[string]time.Time
}

func NewOAuthConfig() *OAuthConfig {
    return &OAuthConfig{
        GoogleConfig: &oauth2.Config{
            ClientID:     getEnv("GOOGLE_CLIENT_ID"),
            ClientSecret: getEnv("GOOGLE_CLIENT_SECRET"),
            RedirectURL:  getEnv("GOOGLE_REDIRECT_URL"),
            Scopes:       []string{"openid", "profile", "email"},
            Endpoint:     google.Endpoint,
        },
        GitHubConfig: &oauth2.Config{
            ClientID:     getEnv("GITHUB_CLIENT_ID"),
            ClientSecret: getEnv("GITHUB_CLIENT_SECRET"),
            RedirectURL:  getEnv("GITHUB_REDIRECT_URL"),
            Scopes:       []string{"user:email"},
            Endpoint:     github.Endpoint,
        },
        StateStore: make(map[string]time.Time),
    }
}

func (o *OAuthConfig) GenerateState() string {
    b := make([]byte, 32)
    rand.Read(b)
    state := base64.URLEncoding.EncodeToString(b)
    o.StateStore[state] = time.Now().Add(10 * time.Minute)
    return state
}

func (o *OAuthConfig) ValidateState(state string) bool {
    expiry, exists := o.StateStore[state]
    if !exists {
        return false
    }
    
    if time.Now().After(expiry) {
        delete(o.StateStore, state)
        return false
    }
    
    delete(o.StateStore, state)
    return true
}

func (o *OAuthConfig) GoogleLogin(c *gin.Context) {
    state := o.GenerateState()
    url := o.GoogleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func (o *OAuthConfig) GoogleCallback(c *gin.Context) {
    state := c.Query("state")
    if !o.ValidateState(state) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
        return
    }
    
    code := c.Query("code")
    token, err := o.GoogleConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
        return
    }
    
    // 获取用户信息
    client := o.GoogleConfig.Client(context.Background(), token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
        return
    }
    defer resp.Body.Close()
    
    var userInfo map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&userInfo)
    
    // 创建JWT Token
    jwtToken, err := GenerateJWT(userInfo["id"].(string), userInfo["email"].(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "token": jwtToken,
        "user":  userInfo,
    })
}

func (o *OAuthConfig) GitHubLogin(c *gin.Context) {
    state := o.GenerateState()
    url := o.GitHubConfig.AuthCodeURL(state)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func (o *OAuthConfig) GitHubCallback(c *gin.Context) {
    state := c.Query("state")
    if !o.ValidateState(state) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
        return
    }
    
    code := c.Query("code")
    token, err := o.GitHubConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
        return
    }
    
    // 获取用户信息
    client := o.GitHubConfig.Client(context.Background(), token)
    resp, err := client.Get("https://api.github.com/user")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
        return
    }
    defer resp.Body.Close()
    
    var userInfo map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&userInfo)
    
    // 获取邮箱信息
    emailResp, err := client.Get("https://api.github.com/user/emails")
    if err == nil {
        defer emailResp.Body.Close()
        var emails []map[string]interface{}
        json.NewDecoder(emailResp.Body).Decode(&emails)
        
        for _, email := range emails {
            if email["primary"].(bool) {
                userInfo["email"] = email["email"]
                break
            }
        }
    }
    
    // 创建JWT Token
    jwtToken, err := GenerateJWT(fmt.Sprintf("%v", userInfo["id"]), userInfo["email"].(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "token": jwtToken,
        "user":  userInfo,
    })
}
```

### 2. JWT Token管理

```go
// pkg/auth/jwt.go
package auth

import (
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v4"
)

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

var jwtSecret = []byte(getEnv("JWT_SECRET"))

func GenerateJWT(userID, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        Role:   "user", // 默认角色
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "taishanglaojun",
            Subject:   userID,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return jwtSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}

func RefreshJWT(tokenString string) (string, error) {
    claims, err := ValidateJWT(tokenString)
    if err != nil {
        return "", err
    }
    
    // 检查是否在刷新窗口内
    if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
        return "", errors.New("token not eligible for refresh")
    }
    
    return GenerateJWT(claims.UserID, claims.Email)
}

// JWT中间件
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
            c.Abort()
            return
        }
        
        // 移除Bearer前缀
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        }
        
        claims, err := ValidateJWT(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)
        c.Next()
    }
}

// 角色检查中间件
func RequireRole(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "No role information"})
            c.Abort()
            return
        }
        
        if role != requiredRole && role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 3. RBAC权限控制

```yaml
# security/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: taishanglaojun-backend
  namespace: taishanglaojun-prod

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: taishanglaojun-prod
  name: taishanglaojun-backend-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: taishanglaojun-backend-binding
  namespace: taishanglaojun-prod
subjects:
- kind: ServiceAccount
  name: taishanglaojun-backend
  namespace: taishanglaojun-prod
roleRef:
  kind: Role
  name: taishanglaojun-backend-role
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: taishanglaojun-ai-service
  namespace: taishanglaojun-prod

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: taishanglaojun-prod
  name: taishanglaojun-ai-service-role
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: taishanglaojun-ai-service-binding
  namespace: taishanglaojun-prod
subjects:
- kind: ServiceAccount
  name: taishanglaojun-ai-service
  namespace: taishanglaojun-prod
roleRef:
  kind: Role
  name: taishanglaojun-ai-service-role
  apiGroup: rbac.authorization.k8s.io
```

## 数据加密

### 1. 数据库加密

```sql
-- PostgreSQL加密配置
-- 启用透明数据加密(TDE)
ALTER SYSTEM SET ssl = on;
ALTER SYSTEM SET ssl_cert_file = '/etc/ssl/certs/server.crt';
ALTER SYSTEM SET ssl_key_file = '/etc/ssl/private/server.key';
ALTER SYSTEM SET ssl_ca_file = '/etc/ssl/certs/ca.crt';

-- 字段级加密
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 创建加密函数
CREATE OR REPLACE FUNCTION encrypt_sensitive_data(data TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN encode(encrypt(data::bytea, 'encryption_key', 'aes'), 'base64');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION decrypt_sensitive_data(encrypted_data TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN convert_from(decrypt(decode(encrypted_data, 'base64'), 'encryption_key', 'aes'), 'UTF8');
END;
$$ LANGUAGE plpgsql;

-- 使用加密字段
CREATE TABLE users_encrypted (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_encrypted TEXT, -- 加密存储
    id_card_encrypted TEXT, -- 加密存储
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入加密数据
INSERT INTO users_encrypted (username, email, phone_encrypted, id_card_encrypted)
VALUES (
    'testuser',
    'test@example.com',
    encrypt_sensitive_data('13800138000'),
    encrypt_sensitive_data('123456789012345678')
);

-- 查询解密数据
SELECT 
    id,
    username,
    email,
    decrypt_sensitive_data(phone_encrypted) AS phone,
    decrypt_sensitive_data(id_card_encrypted) AS id_card
FROM users_encrypted;
```

### 2. 应用层加密

```go
// pkg/crypto/encryption.go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "io"
    
    "golang.org/x/crypto/pbkdf2"
)

type Encryptor struct {
    key []byte
}

func NewEncryptor(password string, salt []byte) *Encryptor {
    key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
    return &Encryptor{key: key}
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}

// 敏感数据处理中间件
func SensitiveDataMiddleware(encryptor *Encryptor) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 请求数据解密
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            var requestData map[string]interface{}
            if err := c.ShouldBindJSON(&requestData); err == nil {
                // 解密敏感字段
                sensitiveFields := []string{"phone", "id_card", "bank_account"}
                for _, field := range sensitiveFields {
                    if encryptedValue, exists := requestData[field]; exists {
                        if encryptedStr, ok := encryptedValue.(string); ok {
                            if decryptedValue, err := encryptor.Decrypt(encryptedStr); err == nil {
                                requestData[field] = decryptedValue
                            }
                        }
                    }
                }
                c.Set("decrypted_data", requestData)
            }
        }
        
        c.Next()
        
        // 响应数据加密
        if c.Writer.Header().Get("Content-Type") == "application/json" {
            // 这里可以添加响应数据加密逻辑
        }
    }
}
```

### 3. 密钥管理

```yaml
# security/vault-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vault-config
  namespace: vault
data:
  vault.hcl: |
    ui = true
    
    listener "tcp" {
      address = "0.0.0.0:8200"
      tls_disable = false
      tls_cert_file = "/vault/tls/tls.crt"
      tls_key_file = "/vault/tls/tls.key"
    }
    
    storage "postgresql" {
      connection_url = "postgres://vault:vault_password@postgres:5432/vault?sslmode=require"
      table = "vault_kv_store"
      max_parallel = "128"
    }
    
    seal "awskms" {
      region = "us-west-2"
      kms_key_id = "your-kms-key-id"
    }
    
    api_addr = "https://vault.taishanglaojun.com:8200"
    cluster_addr = "https://vault.taishanglaojun.com:8201"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vault
  namespace: vault
spec:
  replicas: 3
  selector:
    matchLabels:
      app: vault
  template:
    metadata:
      labels:
        app: vault
    spec:
      serviceAccountName: vault
      containers:
      - name: vault
        image: vault:1.15.0
        ports:
        - containerPort: 8200
          name: vault-port
        - containerPort: 8201
          name: cluster-port
        env:
        - name: VAULT_ADDR
          value: "https://127.0.0.1:8200"
        - name: VAULT_API_ADDR
          value: "https://vault.taishanglaojun.com:8200"
        - name: VAULT_CLUSTER_ADDR
          value: "https://vault.taishanglaojun.com:8201"
        - name: VAULT_RAFT_NODE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        volumeMounts:
        - name: vault-config
          mountPath: /vault/config
        - name: vault-tls
          mountPath: /vault/tls
        command:
        - vault
        - server
        - -config=/vault/config/vault.hcl
        securityContext:
          capabilities:
            add:
            - IPC_LOCK
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: vault-config
        configMap:
          name: vault-config
      - name: vault-tls
        secret:
          secretName: vault-tls
```

### 4. External Secrets Operator配置

```yaml
# security/external-secrets.yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: taishanglaojun-prod
spec:
  provider:
    vault:
      server: "https://vault.taishanglaojun.com:8200"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "taishanglaojun"
          serviceAccountRef:
            name: "external-secrets"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-credentials
  namespace: taishanglaojun-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: database-credentials
    creationPolicy: Owner
  data:
  - secretKey: postgres-password
    remoteRef:
      key: database
      property: postgres_password
  - secretKey: redis-password
    remoteRef:
      key: database
      property: redis_password

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: api-keys
  namespace: taishanglaojun-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: api-keys
    creationPolicy: Owner
  data:
  - secretKey: openai-api-key
    remoteRef:
      key: api-keys
      property: openai_api_key
  - secretKey: anthropic-api-key
    remoteRef:
      key: api-keys
      property: anthropic_api_key
  - secretKey: jwt-secret
    remoteRef:
      key: api-keys
      property: jwt_secret

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: oauth-credentials
  namespace: taishanglaojun-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: oauth-credentials
    creationPolicy: Owner
  data:
  - secretKey: google-client-id
    remoteRef:
      key: oauth
      property: google_client_id
  - secretKey: google-client-secret
    remoteRef:
      key: oauth
      property: google_client_secret
  - secretKey: github-client-id
    remoteRef:
      key: oauth
      property: github_client_id
  - secretKey: github-client-secret
    remoteRef:
      key: oauth
      property: github_client_secret
```

## 安全监控

### 1. 安全事件监控

```yaml
# security/security-monitoring.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: falco-config
  namespace: falco
data:
  falco.yaml: |
    rules_file:
      - /etc/falco/falco_rules.yaml
      - /etc/falco/falco_rules.local.yaml
      - /etc/falco/k8s_audit_rules.yaml
      - /etc/falco/rules.d
    
    time_format_iso_8601: true
    json_output: true
    json_include_output_property: true
    
    log_stderr: true
    log_syslog: true
    log_level: info
    
    priority: debug
    
    buffered_outputs: false
    
    outputs:
      rate: 1
      max_burst: 1000
    
    syslog_output:
      enabled: true
    
    file_output:
      enabled: true
      keep_alive: false
      filename: /var/log/falco.log
    
    stdout_output:
      enabled: true
    
    webserver:
      enabled: true
      listen_port: 8765
      k8s_healthz_endpoint: /healthz
      ssl_enabled: false
      ssl_certificate: /etc/ssl/falco/falco.pem
    
    grpc:
      enabled: false
      bind_address: "0.0.0.0:5060"
      threadiness: 8
    
    grpc_output:
      enabled: false

  falco_rules.local.yaml: |
    - rule: Detect Crypto Mining
      desc: Detect cryptocurrency mining activities
      condition: >
        spawned_process and
        (proc.name in (xmrig, cpuminer, ccminer, cgminer, bfgminer) or
         proc.cmdline contains "stratum+tcp" or
         proc.cmdline contains "mining.pool" or
         proc.cmdline contains "cryptonight")
      output: >
        Cryptocurrency mining detected (user=%user.name command=%proc.cmdline
        container=%container.name image=%container.image.repository)
      priority: CRITICAL
      tags: [cryptocurrency, mining, malware]
    
    - rule: Detect Privilege Escalation
      desc: Detect attempts to escalate privileges
      condition: >
        spawned_process and
        (proc.name in (sudo, su, pkexec, doas) or
         proc.cmdline contains "chmod +s" or
         proc.cmdline contains "setuid" or
         proc.cmdline contains "setgid")
      output: >
        Privilege escalation attempt detected (user=%user.name command=%proc.cmdline
        container=%container.name image=%container.image.repository)
      priority: WARNING
      tags: [privilege_escalation, security]
    
    - rule: Detect Suspicious Network Activity
      desc: Detect suspicious network connections
      condition: >
        (inbound_outbound) and
        (fd.sport in (4444, 5555, 6666, 7777, 8888, 9999) or
         fd.dport in (4444, 5555, 6666, 7777, 8888, 9999) or
         fd.sip.name contains "tor" or
         fd.sip.name contains "onion")
      output: >
        Suspicious network activity detected (connection=%fd.name direction=%evt.type
        container=%container.name image=%container.image.repository)
      priority: WARNING
      tags: [network, suspicious, security]
    
    - rule: Detect File System Changes in Sensitive Directories
      desc: Detect unauthorized changes to sensitive directories
      condition: >
        (open_write or rename or unlink) and
        (fd.name startswith /etc/ or
         fd.name startswith /usr/bin/ or
         fd.name startswith /usr/sbin/ or
         fd.name startswith /bin/ or
         fd.name startswith /sbin/)
      output: >
        Sensitive file system change detected (file=%fd.name operation=%evt.type
        user=%user.name container=%container.name image=%container.image.repository)
      priority: WARNING
      tags: [filesystem, security, sensitive]

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: falco
  namespace: falco
spec:
  selector:
    matchLabels:
      app: falco
  template:
    metadata:
      labels:
        app: falco
    spec:
      serviceAccount: falco
      hostNetwork: true
      hostPID: true
      containers:
      - name: falco
        image: falcosecurity/falco:0.36.0
        securityContext:
          privileged: true
        volumeMounts:
        - name: dev
          mountPath: /host/dev
        - name: proc
          mountPath: /host/proc
          readOnly: true
        - name: boot
          mountPath: /host/boot
          readOnly: true
        - name: lib-modules
          mountPath: /host/lib/modules
          readOnly: true
        - name: usr
          mountPath: /host/usr
          readOnly: true
        - name: etc
          mountPath: /host/etc
          readOnly: true
        - name: falco-config
          mountPath: /etc/falco
        env:
        - name: FALCO_K8S_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        resources:
          limits:
            memory: 512Mi
            cpu: 500m
          requests:
            memory: 256Mi
            cpu: 100m
      volumes:
      - name: dev
        hostPath:
          path: /dev
      - name: proc
        hostPath:
          path: /proc
      - name: boot
        hostPath:
          path: /boot
      - name: lib-modules
        hostPath:
          path: /lib/modules
      - name: usr
        hostPath:
          path: /usr
      - name: etc
        hostPath:
          path: /etc
      - name: falco-config
        configMap:
          name: falco-config
```

### 2. 安全审计日志

```go
// pkg/audit/audit.go
package audit

import (
    "encoding/json"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

type AuditEvent struct {
    Timestamp   time.Time   `json:"timestamp"`
    UserID      string      `json:"user_id"`
    Email       string      `json:"email"`
    Action      string      `json:"action"`
    Resource    string      `json:"resource"`
    Method      string      `json:"method"`
    Path        string      `json:"path"`
    IP          string      `json:"ip"`
    UserAgent   string      `json:"user_agent"`
    StatusCode  int         `json:"status_code"`
    Duration    int64       `json:"duration_ms"`
    RequestBody interface{} `json:"request_body,omitempty"`
    Error       string      `json:"error,omitempty"`
}

type AuditLogger struct {
    logger *logrus.Logger
}

func NewAuditLogger() *AuditLogger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    
    // 配置日志输出到文件
    file, err := os.OpenFile("/var/log/audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err == nil {
        logger.SetOutput(file)
    }
    
    return &AuditLogger{logger: logger}
}

func (a *AuditLogger) AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 记录请求体（敏感信息需要脱敏）
        var requestBody interface{}
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            var body map[string]interface{}
            if err := c.ShouldBindJSON(&body); err == nil {
                // 脱敏处理
                requestBody = a.sanitizeRequestBody(body)
            }
        }
        
        c.Next()
        
        // 记录审计事件
        event := AuditEvent{
            Timestamp:   start,
            UserID:      c.GetString("user_id"),
            Email:       c.GetString("email"),
            Action:      a.getActionFromPath(c.Request.URL.Path),
            Resource:    a.getResourceFromPath(c.Request.URL.Path),
            Method:      c.Request.Method,
            Path:        c.Request.URL.Path,
            IP:          c.ClientIP(),
            UserAgent:   c.Request.UserAgent(),
            StatusCode:  c.Writer.Status(),
            Duration:    time.Since(start).Milliseconds(),
            RequestBody: requestBody,
        }
        
        // 记录错误信息
        if c.Writer.Status() >= 400 {
            if err, exists := c.Get("error"); exists {
                event.Error = err.(string)
            }
        }
        
        a.logAuditEvent(event)
    }
}

func (a *AuditLogger) sanitizeRequestBody(body map[string]interface{}) map[string]interface{} {
    sensitiveFields := []string{"password", "token", "secret", "key", "phone", "id_card"}
    
    sanitized := make(map[string]interface{})
    for k, v := range body {
        isSensitive := false
        for _, field := range sensitiveFields {
            if strings.Contains(strings.ToLower(k), field) {
                isSensitive = true
                break
            }
        }
        
        if isSensitive {
            sanitized[k] = "***REDACTED***"
        } else {
            sanitized[k] = v
        }
    }
    
    return sanitized
}

func (a *AuditLogger) getActionFromPath(path string) string {
    if strings.Contains(path, "/login") {
        return "LOGIN"
    } else if strings.Contains(path, "/logout") {
        return "LOGOUT"
    } else if strings.Contains(path, "/chat") {
        return "CHAT"
    } else if strings.Contains(path, "/upload") {
        return "UPLOAD"
    } else if strings.Contains(path, "/download") {
        return "DOWNLOAD"
    }
    return "UNKNOWN"
}

func (a *AuditLogger) getResourceFromPath(path string) string {
    parts := strings.Split(path, "/")
    if len(parts) >= 3 {
        return parts[2]
    }
    return "UNKNOWN"
}

func (a *AuditLogger) logAuditEvent(event AuditEvent) {
    eventJSON, _ := json.Marshal(event)
    a.logger.Info(string(eventJSON))
    
    // 高风险事件额外处理
    if a.isHighRiskEvent(event) {
        a.handleHighRiskEvent(event)
    }
}

func (a *AuditLogger) isHighRiskEvent(event AuditEvent) bool {
    // 定义高风险事件
    highRiskActions := []string{"LOGIN_FAILED", "PRIVILEGE_ESCALATION", "SENSITIVE_DATA_ACCESS"}
    
    for _, action := range highRiskActions {
        if event.Action == action {
            return true
        }
    }
    
    // 多次失败登录
    if event.Action == "LOGIN" && event.StatusCode >= 400 {
        return true
    }
    
    // 异常IP访问
    if a.isAnomalousIP(event.IP) {
        return true
    }
    
    return false
}

func (a *AuditLogger) handleHighRiskEvent(event AuditEvent) {
    // 发送告警
    alert := map[string]interface{}{
        "type":        "security_alert",
        "severity":    "high",
        "event":       event,
        "timestamp":   time.Now(),
        "description": "High risk security event detected",
    }
    
    // 发送到监控系统
    a.sendSecurityAlert(alert)
}

func (a *AuditLogger) sendSecurityAlert(alert map[string]interface{}) {
    // 实现告警发送逻辑
    // 可以发送到Slack、邮件、短信等
}

func (a *AuditLogger) isAnomalousIP(ip string) bool {
    // 实现IP异常检测逻辑
    // 可以基于地理位置、历史访问记录等
    return false
}
```

## 相关文档

- [部署概览](./deployment-overview.md)
- [Docker部署](./docker-deployment.md)
- [Kubernetes部署](./kubernetes-deployment.md)
- [监控运维](./monitoring-operations.md)
- [性能调优](./performance-tuning.md)