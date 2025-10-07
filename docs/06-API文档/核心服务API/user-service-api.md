# 用户服务API文档

## 1. 服务概述

用户服务是太上老君AI平台的核心服务之一，负责用户账户管理、身份认证、权限控制和用户配置文件管理。基于S×C×T三轴理论设计，提供完整的用户生命周期管理功能。

### 1.1 服务架构

```mermaid
graph TB
    subgraph "用户服务架构"
        subgraph "API层"
            REST[REST API]
            GraphQL[GraphQL API]
            WS[WebSocket API]
        end
        
        subgraph "业务逻辑层"
            AUTH[认证服务]
            PROFILE[用户配置]
            PERM[权限管理]
            SOCIAL[社交功能]
        end
        
        subgraph "数据层"
            PG[(PostgreSQL)]
            REDIS[(Redis缓存)]
            ES[(Elasticsearch)]
        end
        
        subgraph "外部集成"
            OAUTH[OAuth提供商]
            EMAIL[邮件服务]
            SMS[短信服务]
        end
    end
    
    REST --> AUTH
    GraphQL --> AUTH
    WS --> AUTH
    
    AUTH --> PROFILE
    AUTH --> PERM
    AUTH --> SOCIAL
    
    PROFILE --> PG
    PERM --> PG
    SOCIAL --> PG
    
    AUTH --> REDIS
    PROFILE --> REDIS
    
    SOCIAL --> ES
    
    AUTH --> OAUTH
    AUTH --> EMAIL
    AUTH --> SMS
```

### 1.2 核心功能

- **用户注册与登录**：支持多种注册方式和认证方法
- **用户配置文件管理**：个人信息、偏好设置、学习档案
- **权限与角色管理**：基于RBAC的权限控制系统
- **社交功能**：好友关系、关注/粉丝、社交图谱
- **多租户支持**：企业级多租户架构
- **安全功能**：账户安全、隐私保护、审计日志

## 2. 认证与授权

### 2.1 用户注册

#### 2.1.1 邮箱注册

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "username": "johndoe",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "birth_date": "1990-01-01",
  "gender": "male",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai",
  "terms_accepted": true,
  "privacy_accepted": true,
  "marketing_consent": false
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "user_id": "usr_1234567890abcdef",
    "email": "user@example.com",
    "username": "johndoe",
    "status": "pending_verification",
    "verification_token": "vrf_abcdef1234567890",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "api_version": "v1",
    "request_id": "req_1234567890abcdef"
  }
}
```

#### 2.1.2 手机号注册

```http
POST /api/v1/auth/register/phone
Content-Type: application/json

{
  "phone": "+8613800138000",
  "verification_code": "123456",
  "password": "SecurePassword123!",
  "username": "johndoe",
  "first_name": "张",
  "last_name": "三",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai"
}
```

#### 2.1.3 社交账号注册

```http
POST /api/v1/auth/register/social
Content-Type: application/json

{
  "provider": "google",
  "access_token": "ya29.a0AfH6SMC...",
  "username": "johndoe",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai"
}
```

### 2.2 用户登录

#### 2.2.1 邮箱/用户名登录

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "identifier": "user@example.com",  // 邮箱或用户名
  "password": "SecurePassword123!",
  "remember_me": true,
  "device_info": {
    "device_id": "dev_1234567890",
    "device_name": "iPhone 15 Pro",
    "os": "iOS",
    "os_version": "17.0",
    "app_version": "1.0.0"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "rt_1234567890abcdef",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "user_id": "usr_1234567890abcdef",
      "email": "user@example.com",
      "username": "johndoe",
      "first_name": "John",
      "last_name": "Doe",
      "avatar_url": "https://cdn.taishanglaojun.com/avatars/usr_1234567890abcdef.jpg",
      "roles": ["user"],
      "permissions": ["profile:read", "profile:write", "learning:read"],
      "last_login_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 2.2.2 手机号登录

```http
POST /api/v1/auth/login/phone
Content-Type: application/json

{
  "phone": "+8613800138000",
  "verification_code": "123456",
  "device_info": {
    "device_id": "dev_1234567890",
    "device_name": "华为 Mate 60 Pro",
    "os": "Android",
    "os_version": "14.0",
    "app_version": "1.0.0"
  }
}
```

#### 2.2.3 社交账号登录

```http
POST /api/v1/auth/login/social
Content-Type: application/json

{
  "provider": "wechat",
  "access_token": "ACCESS_TOKEN",
  "openid": "OPENID"
}
```

### 2.3 令牌管理

#### 2.3.1 刷新访问令牌

```http
POST /api/v1/auth/refresh
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "refresh_token": "rt_1234567890abcdef"
}
```

#### 2.3.2 注销登录

```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "all_devices": false  // 是否注销所有设备
}
```

#### 2.3.3 撤销令牌

```http
POST /api/v1/auth/revoke
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "token": "rt_1234567890abcdef",
  "token_type": "refresh_token"  // access_token, refresh_token, all
}
```

## 3. 用户配置文件管理

### 3.1 获取用户信息

#### 3.1.1 获取当前用户信息

```http
GET /api/v1/users/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "user_id": "usr_1234567890abcdef",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "display_name": "John Doe",
    "avatar_url": "https://cdn.taishanglaojun.com/avatars/usr_1234567890abcdef.jpg",
    "phone": "+1234567890",
    "birth_date": "1990-01-01",
    "gender": "male",
    "language": "zh-CN",
    "timezone": "Asia/Shanghai",
    "bio": "AI enthusiast and lifelong learner",
    "location": {
      "country": "China",
      "province": "Beijing",
      "city": "Beijing"
    },
    "preferences": {
      "theme": "auto",
      "notifications": {
        "email": true,
        "push": true,
        "sms": false
      },
      "privacy": {
        "profile_visibility": "public",
        "activity_visibility": "friends",
        "search_visibility": true
      }
    },
    "statistics": {
      "total_learning_hours": 120,
      "completed_courses": 5,
      "achievement_points": 1500,
      "social_connections": 25
    },
    "status": "active",
    "email_verified": true,
    "phone_verified": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "last_login_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 3.1.2 获取指定用户信息

```http
GET /api/v1/users/{user_id}
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 3.1.3 搜索用户

```http
GET /api/v1/users/search?q=john&limit=20&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**查询参数：**

- `q`: 搜索关键词（用户名、显示名称、邮箱）
- `limit`: 返回结果数量限制（默认20，最大100）
- `offset`: 偏移量（用于分页）
- `filters`: 过滤条件（JSON格式）
  - `location`: 地理位置过滤
  - `age_range`: 年龄范围
  - `interests`: 兴趣标签
  - `activity_level`: 活跃度

### 3.2 更新用户信息

#### 3.2.1 更新基本信息

```http
PATCH /api/v1/users/me
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "first_name": "John",
  "last_name": "Smith",
  "display_name": "John Smith",
  "bio": "Updated bio description",
  "location": {
    "country": "China",
    "province": "Shanghai",
    "city": "Shanghai"
  },
  "interests": ["AI", "Machine Learning", "Philosophy", "Health"]
}
```

#### 3.2.2 更新头像

```http
POST /api/v1/users/me/avatar
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

avatar: [binary file data]
```

#### 3.2.3 更新偏好设置

```http
PATCH /api/v1/users/me/preferences
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "theme": "dark",
  "language": "en-US",
  "timezone": "America/New_York",
  "notifications": {
    "email": true,
    "push": false,
    "sms": false,
    "learning_reminders": true,
    "social_updates": true,
    "system_announcements": true
  },
  "privacy": {
    "profile_visibility": "friends",
    "activity_visibility": "private",
    "search_visibility": false,
    "data_sharing": false
  },
  "learning": {
    "difficulty_preference": "intermediate",
    "learning_pace": "moderate",
    "preferred_content_types": ["video", "text", "interactive"],
    "study_reminders": {
      "enabled": true,
      "frequency": "daily",
      "time": "09:00"
    }
  }
}
```

### 3.3 账户安全

#### 3.3.1 修改密码

```http
POST /api/v1/users/me/change-password
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "current_password": "CurrentPassword123!",
  "new_password": "NewSecurePassword456!",
  "confirm_password": "NewSecurePassword456!"
}
```

#### 3.3.2 启用/禁用双因素认证

```http
POST /api/v1/users/me/2fa/enable
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "method": "totp",  // totp, sms, email
  "phone": "+8613800138000"  // SMS方式时需要
}
```

**响应示例（TOTP）：**

```json
{
  "success": true,
  "data": {
    "secret": "JBSWY3DPEHPK3PXP",
    "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "backup_codes": [
      "12345678",
      "87654321",
      "11223344",
      "44332211",
      "55667788"
    ]
  }
}
```

#### 3.3.3 验证双因素认证

```http
POST /api/v1/users/me/2fa/verify
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "code": "123456"
}
```

#### 3.3.4 获取登录设备列表

```http
GET /api/v1/users/me/devices
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": [
    {
      "device_id": "dev_1234567890",
      "device_name": "iPhone 15 Pro",
      "os": "iOS",
      "os_version": "17.0",
      "app_version": "1.0.0",
      "ip_address": "192.168.1.100",
      "location": "Beijing, China",
      "is_current": true,
      "last_active": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 3.3.5 注销指定设备

```http
DELETE /api/v1/users/me/devices/{device_id}
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 4. 社交功能

### 4.1 好友管理

#### 4.1.1 发送好友请求

```http
POST /api/v1/users/me/friends/requests
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "target_user_id": "usr_abcdef1234567890",
  "message": "Hi, I'd like to connect with you!"
}
```

#### 4.1.2 获取好友请求列表

```http
GET /api/v1/users/me/friends/requests?type=received&status=pending
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**查询参数：**

- `type`: `sent` | `received`
- `status`: `pending` | `accepted` | `rejected`

#### 4.1.3 处理好友请求

```http
PATCH /api/v1/users/me/friends/requests/{request_id}
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "action": "accept"  // accept, reject
}
```

#### 4.1.4 获取好友列表

```http
GET /api/v1/users/me/friends?limit=50&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 4.1.5 删除好友

```http
DELETE /api/v1/users/me/friends/{user_id}
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 4.2 关注系统

#### 4.2.1 关注用户

```http
POST /api/v1/users/me/following
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "target_user_id": "usr_abcdef1234567890"
}
```

#### 4.2.2 取消关注

```http
DELETE /api/v1/users/me/following/{user_id}
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 4.2.3 获取关注列表

```http
GET /api/v1/users/me/following?limit=50&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 4.2.4 获取粉丝列表

```http
GET /api/v1/users/me/followers?limit=50&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 5. 权限与角色管理

### 5.1 获取用户权限

```http
GET /api/v1/users/me/permissions
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "roles": [
      {
        "role_id": "role_user",
        "name": "user",
        "display_name": "普通用户",
        "description": "平台普通用户权限",
        "assigned_at": "2024-01-01T00:00:00Z"
      }
    ],
    "permissions": [
      {
        "permission_id": "perm_profile_read",
        "name": "profile:read",
        "resource": "profile",
        "action": "read",
        "granted_by": "role_user"
      },
      {
        "permission_id": "perm_profile_write",
        "name": "profile:write",
        "resource": "profile",
        "action": "write",
        "granted_by": "role_user"
      }
    ],
    "effective_permissions": [
      "profile:read",
      "profile:write",
      "learning:read",
      "health:read",
      "culture:read",
      "community:read",
      "community:write"
    ]
  }
}
```

### 5.2 检查权限

```http
POST /api/v1/users/me/permissions/check
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "permissions": ["learning:write", "admin:read"],
  "resource_id": "course_123"  // 可选，检查特定资源权限
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "results": [
      {
        "permission": "learning:write",
        "granted": true,
        "reason": "granted_by_role"
      },
      {
        "permission": "admin:read",
        "granted": false,
        "reason": "insufficient_role"
      }
    ]
  }
}
```

## 6. 多租户支持

### 6.1 租户管理

#### 6.1.1 获取用户租户信息

```http
GET /api/v1/users/me/tenants
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 6.1.2 切换租户

```http
POST /api/v1/users/me/tenants/switch
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "tenant_id": "tenant_enterprise_123"
}
```

#### 6.1.3 邀请用户加入租户

```http
POST /api/v1/tenants/{tenant_id}/invitations
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "email": "newuser@company.com",
  "role": "member",
  "message": "Welcome to our organization!"
}
```

## 7. 数据模型

### 7.1 用户实体

```typescript
interface User {
  user_id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  display_name: string;
  avatar_url?: string;
  phone?: string;
  birth_date?: string;
  gender?: 'male' | 'female' | 'other' | 'prefer_not_to_say';
  language: string;
  timezone: string;
  bio?: string;
  location?: UserLocation;
  interests?: string[];
  preferences: UserPreferences;
  statistics: UserStatistics;
  status: 'active' | 'inactive' | 'suspended' | 'deleted';
  email_verified: boolean;
  phone_verified: boolean;
  two_factor_enabled: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

interface UserLocation {
  country?: string;
  province?: string;
  city?: string;
  latitude?: number;
  longitude?: number;
}

interface UserPreferences {
  theme: 'light' | 'dark' | 'auto';
  notifications: NotificationPreferences;
  privacy: PrivacyPreferences;
  learning?: LearningPreferences;
}

interface NotificationPreferences {
  email: boolean;
  push: boolean;
  sms: boolean;
  learning_reminders?: boolean;
  social_updates?: boolean;
  system_announcements?: boolean;
}

interface PrivacyPreferences {
  profile_visibility: 'public' | 'friends' | 'private';
  activity_visibility: 'public' | 'friends' | 'private';
  search_visibility: boolean;
  data_sharing: boolean;
}

interface LearningPreferences {
  difficulty_preference: 'beginner' | 'intermediate' | 'advanced';
  learning_pace: 'slow' | 'moderate' | 'fast';
  preferred_content_types: string[];
  study_reminders: StudyReminders;
}

interface StudyReminders {
  enabled: boolean;
  frequency: 'daily' | 'weekly' | 'custom';
  time: string;
  days?: string[];
}

interface UserStatistics {
  total_learning_hours: number;
  completed_courses: number;
  achievement_points: number;
  social_connections: number;
  content_created: number;
  community_contributions: number;
}
```

### 7.2 认证实体

```typescript
interface AuthSession {
  session_id: string;
  user_id: string;
  device_id: string;
  device_info: DeviceInfo;
  ip_address: string;
  location?: string;
  access_token: string;
  refresh_token: string;
  expires_at: string;
  is_active: boolean;
  created_at: string;
  last_active: string;
}

interface DeviceInfo {
  device_name: string;
  os: string;
  os_version: string;
  app_version: string;
  browser?: string;
  browser_version?: string;
}

interface TwoFactorAuth {
  user_id: string;
  method: 'totp' | 'sms' | 'email';
  secret?: string;
  phone?: string;
  backup_codes: string[];
  enabled: boolean;
  verified: boolean;
  created_at: string;
}
```

### 7.3 社交实体

```typescript
interface Friendship {
  friendship_id: string;
  requester_id: string;
  addressee_id: string;
  status: 'pending' | 'accepted' | 'rejected' | 'blocked';
  message?: string;
  created_at: string;
  updated_at: string;
}

interface Following {
  following_id: string;
  follower_id: string;
  following_id: string;
  created_at: string;
}

interface SocialGraph {
  user_id: string;
  friends_count: number;
  followers_count: number;
  following_count: number;
  mutual_friends: string[];
  social_score: number;
  updated_at: string;
}
```

## 8. 错误处理

### 8.1 用户服务特定错误

```typescript
enum UserServiceErrorCode {
  // 认证错误
  INVALID_CREDENTIALS = 'U2001',
  ACCOUNT_LOCKED = 'U2002',
  ACCOUNT_SUSPENDED = 'U2003',
  EMAIL_NOT_VERIFIED = 'U2004',
  TWO_FACTOR_REQUIRED = 'U2005',
  INVALID_2FA_CODE = 'U2006',
  
  // 注册错误
  EMAIL_ALREADY_EXISTS = 'U3001',
  USERNAME_ALREADY_EXISTS = 'U3002',
  PHONE_ALREADY_EXISTS = 'U3003',
  INVALID_INVITATION_CODE = 'U3004',
  REGISTRATION_DISABLED = 'U3005',
  
  // 用户信息错误
  USER_NOT_FOUND = 'U4001',
  INVALID_USER_DATA = 'U4002',
  PROFILE_UPDATE_FAILED = 'U4003',
  AVATAR_UPLOAD_FAILED = 'U4004',
  
  // 社交功能错误
  FRIEND_REQUEST_EXISTS = 'U5001',
  CANNOT_FRIEND_SELF = 'U5002',
  FRIENDSHIP_NOT_FOUND = 'U5003',
  ALREADY_FOLLOWING = 'U5004',
  CANNOT_FOLLOW_SELF = 'U5005',
  
  // 权限错误
  INSUFFICIENT_PERMISSIONS = 'U6001',
  ROLE_NOT_FOUND = 'U6002',
  PERMISSION_DENIED = 'U6003',
  
  // 租户错误
  TENANT_NOT_FOUND = 'U7001',
  TENANT_ACCESS_DENIED = 'U7002',
  INVITATION_EXPIRED = 'U7003',
  INVITATION_ALREADY_USED = 'U7004',
}
```

### 8.2 错误响应示例

```json
{
  "success": false,
  "error": {
    "code": "U2001",
    "message": "Invalid email or password",
    "details": {
      "field": "credentials",
      "attempts_remaining": 2,
      "lockout_time": 300
    },
    "trace_id": "trace_1234567890abcdef",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## 9. 性能优化

### 9.1 缓存策略

```yaml
# 用户服务缓存配置
cache_strategy:
  user_profile:
    ttl: 3600  # 1小时
    key_pattern: "user:profile:{user_id}"
    invalidation_events:
      - "user.profile.updated"
      - "user.preferences.updated"
      
  user_permissions:
    ttl: 1800  # 30分钟
    key_pattern: "user:permissions:{user_id}"
    invalidation_events:
      - "user.role.changed"
      - "role.permissions.updated"
      
  social_graph:
    ttl: 7200  # 2小时
    key_pattern: "user:social:{user_id}"
    invalidation_events:
      - "friendship.created"
      - "friendship.deleted"
      - "following.created"
      - "following.deleted"
      
  session_data:
    ttl: 86400  # 24小时
    key_pattern: "session:{session_id}"
    invalidation_events:
      - "user.logout"
      - "session.expired"
```

### 9.2 数据库优化

```sql
-- 用户表索引优化
CREATE INDEX CONCURRENTLY idx_users_email ON users(email) WHERE status = 'active';
CREATE INDEX CONCURRENTLY idx_users_username ON users(username) WHERE status = 'active';
CREATE INDEX CONCURRENTLY idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_users_created_at ON users(created_at);
CREATE INDEX CONCURRENTLY idx_users_last_login ON users(last_login_at) WHERE last_login_at IS NOT NULL;

-- 社交关系索引
CREATE INDEX CONCURRENTLY idx_friendships_requester ON friendships(requester_id, status);
CREATE INDEX CONCURRENTLY idx_friendships_addressee ON friendships(addressee_id, status);
CREATE INDEX CONCURRENTLY idx_following_follower ON following(follower_id);
CREATE INDEX CONCURRENTLY idx_following_following ON following(following_id);

-- 会话表索引
CREATE INDEX CONCURRENTLY idx_sessions_user_active ON auth_sessions(user_id, is_active) WHERE is_active = true;
CREATE INDEX CONCURRENTLY idx_sessions_device ON auth_sessions(device_id);
CREATE INDEX CONCURRENTLY idx_sessions_expires ON auth_sessions(expires_at) WHERE is_active = true;
```

## 10. 监控与指标

### 10.1 关键指标

```yaml
# 用户服务监控指标
metrics:
  authentication:
    - name: "login_success_rate"
      description: "登录成功率"
      type: "percentage"
      target: "> 95%"
      
    - name: "login_response_time"
      description: "登录响应时间"
      type: "duration"
      target: "< 500ms"
      
    - name: "failed_login_attempts"
      description: "失败登录尝试次数"
      type: "counter"
      alert_threshold: "> 100/hour"
      
  user_management:
    - name: "user_registration_rate"
      description: "用户注册率"
      type: "rate"
      
    - name: "profile_update_success_rate"
      description: "用户信息更新成功率"
      type: "percentage"
      target: "> 99%"
      
    - name: "avatar_upload_success_rate"
      description: "头像上传成功率"
      type: "percentage"
      target: "> 95%"
      
  social_features:
    - name: "friend_request_response_rate"
      description: "好友请求响应率"
      type: "percentage"
      
    - name: "social_interaction_rate"
      description: "社交互动率"
      type: "rate"
      
  performance:
    - name: "cache_hit_rate"
      description: "缓存命中率"
      type: "percentage"
      target: "> 90%"
      
    - name: "database_query_time"
      description: "数据库查询时间"
      type: "duration"
      target: "< 100ms"
```

## 11. 安全考虑

### 11.1 数据保护

- **密码安全**：使用bcrypt加密，最小复杂度要求
- **敏感数据加密**：PII数据使用AES-256加密存储
- **访问日志**：记录所有敏感操作的审计日志
- **数据脱敏**：API响应中敏感字段脱敏处理

### 11.2 防护措施

- **暴力破解防护**：登录失败次数限制和账户锁定
- **会话管理**：安全的会话生成和过期机制
- **输入验证**：严格的输入验证和SQL注入防护
- **CSRF防护**：使用CSRF令牌保护状态变更操作

## 12. 相关文档

- [API概览文档](../api-overview.md)
- [AI服务API](./ai-service-api.md)
- [学习服务API](./learning-service-api.md)
- [GraphQL Schema](../graphql-schema.md)
- [认证与授权指南](../auth-guide.md)