# AI对话功能 - 共享数据模型定义

## 数据模型规范

### 1. 消息模型 (Message)

#### Android (Kotlin)
```kotlin
data class ChatMessage(
    val id: String = UUID.randomUUID().toString(),
    val conversationId: String,
    val content: String,
    val messageType: MessageType,
    val sender: MessageSender,
    val timestamp: Long = System.currentTimeMillis(),
    val status: MessageStatus = MessageStatus.SENDING,
    val metadata: Map<String, Any> = emptyMap()
)

enum class MessageType {
    TEXT,
    IMAGE,
    AUDIO,
    SYSTEM
}

enum class MessageSender {
    USER,
    AI,
    SYSTEM
}

enum class MessageStatus {
    SENDING,
    SENT,
    DELIVERED,
    READ,
    FAILED
}
```

#### iOS (Swift)
```swift
struct ChatMessage: Codable, Identifiable {
    let id: String = UUID().uuidString
    let conversationId: String
    let content: String
    let messageType: MessageType
    let sender: MessageSender
    let timestamp: TimeInterval = Date().timeIntervalSince1970
    var status: MessageStatus = .sending
    let metadata: [String: Any] = [:]
}

enum MessageType: String, Codable, CaseIterable {
    case text = "TEXT"
    case image = "IMAGE"
    case audio = "AUDIO"
    case system = "SYSTEM"
}

enum MessageSender: String, Codable, CaseIterable {
    case user = "USER"
    case ai = "AI"
    case system = "SYSTEM"
}

enum MessageStatus: String, Codable, CaseIterable {
    case sending = "SENDING"
    case sent = "SENT"
    case delivered = "DELIVERED"
    case read = "READ"
    case failed = "FAILED"
}
```

#### 鸿蒙 (ArkTS)
```typescript
export interface ChatMessage {
  id: string;
  conversationId: string;
  content: string;
  messageType: MessageType;
  sender: MessageSender;
  timestamp: number;
  status: MessageStatus;
  metadata: Record<string, any>;
}

export enum MessageType {
  TEXT = 'TEXT',
  IMAGE = 'IMAGE',
  AUDIO = 'AUDIO',
  SYSTEM = 'SYSTEM'
}

export enum MessageSender {
  USER = 'USER',
  AI = 'AI',
  SYSTEM = 'SYSTEM'
}

export enum MessageStatus {
  SENDING = 'SENDING',
  SENT = 'SENT',
  DELIVERED = 'DELIVERED',
  READ = 'READ',
  FAILED = 'FAILED'
}
```

### 2. 对话会话模型 (Conversation)

#### Android (Kotlin)
```kotlin
data class Conversation(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val createdAt: Long = System.currentTimeMillis(),
    val updatedAt: Long = System.currentTimeMillis(),
    val lastMessage: ChatMessage? = null,
    val messageCount: Int = 0,
    val isArchived: Boolean = false,
    val aiPersonality: AIPersonality = AIPersonality.DEFAULT
)

enum class AIPersonality {
    DEFAULT,
    WISE_SAGE,      // 智慧长者
    FRIENDLY_GUIDE, // 友善向导
    SCHOLARLY,      // 学者风格
    POETIC         // 诗意风格
}
```

#### iOS (Swift)
```swift
struct Conversation: Codable, Identifiable {
    let id: String = UUID().uuidString
    var title: String
    let createdAt: TimeInterval = Date().timeIntervalSince1970
    var updatedAt: TimeInterval = Date().timeIntervalSince1970
    var lastMessage: ChatMessage?
    var messageCount: Int = 0
    var isArchived: Bool = false
    var aiPersonality: AIPersonality = .default
}

enum AIPersonality: String, Codable, CaseIterable {
    case `default` = "DEFAULT"
    case wiseSage = "WISE_SAGE"
    case friendlyGuide = "FRIENDLY_GUIDE"
    case scholarly = "SCHOLARLY"
    case poetic = "POETIC"
}
```

#### 鸿蒙 (ArkTS)
```typescript
export interface Conversation {
  id: string;
  title: string;
  createdAt: number;
  updatedAt: number;
  lastMessage?: ChatMessage;
  messageCount: number;
  isArchived: boolean;
  aiPersonality: AIPersonality;
}

export enum AIPersonality {
  DEFAULT = 'DEFAULT',
  WISE_SAGE = 'WISE_SAGE',
  FRIENDLY_GUIDE = 'FRIENDLY_GUIDE',
  SCHOLARLY = 'SCHOLARLY',
  POETIC = 'POETIC'
}
```

### 3. AI响应模型 (AIResponse)

#### 通用JSON格式
```json
{
  "success": true,
  "message": {
    "content": "您好！我是太上老君AI助手，很高兴为您服务。",
    "messageType": "TEXT",
    "timestamp": 1704067200000,
    "metadata": {
      "model": "taishang-v1",
      "tokens_used": 25,
      "response_time": 1200,
      "personality": "WISE_SAGE"
    }
  },
  "suggestions": [
    "请问您想了解什么文化知识？",
    "我可以为您解答哲学问题",
    "需要我推荐一些经典著作吗？"
  ],
  "error": null
}
```

## API接口规范

### 1. 发送消息接口
```
POST /api/ai/chat/send
Content-Type: application/json
Authorization: Bearer <token>

Request Body:
{
  "conversationId": "conv_123",
  "message": "什么是道？",
  "messageType": "TEXT",
  "aiPersonality": "WISE_SAGE"
}

Response:
{
  "success": true,
  "data": {
    "messageId": "msg_456",
    "response": "道，可道，非常道...",
    "conversationId": "conv_123",
    "timestamp": 1704067200000
  }
}
```

### 2. 获取对话历史
```
GET /api/ai/chat/conversations/{conversationId}/messages
Authorization: Bearer <token>

Query Parameters:
- page: 页码 (默认: 1)
- limit: 每页数量 (默认: 50)
- before: 获取指定时间之前的消息

Response:
{
  "success": true,
  "data": {
    "messages": [...],
    "hasMore": true,
    "total": 150
  }
}
```

### 3. 创建新对话
```
POST /api/ai/chat/conversations
Authorization: Bearer <token>

Request Body:
{
  "title": "关于道德经的讨论",
  "aiPersonality": "WISE_SAGE"
}

Response:
{
  "success": true,
  "data": {
    "conversationId": "conv_789",
    "title": "关于道德经的讨论",
    "createdAt": 1704067200000
  }
}
```

## 本地存储规范

### Android (Room Database)
```kotlin
@Entity(tableName = "chat_messages")
data class ChatMessageEntity(
    @PrimaryKey val id: String,
    val conversationId: String,
    val content: String,
    val messageType: String,
    val sender: String,
    val timestamp: Long,
    val status: String,
    val metadata: String // JSON字符串
)

@Entity(tableName = "conversations")
data class ConversationEntity(
    @PrimaryKey val id: String,
    val title: String,
    val createdAt: Long,
    val updatedAt: Long,
    val lastMessageId: String?,
    val messageCount: Int,
    val isArchived: Boolean,
    val aiPersonality: String
)
```

### iOS (Core Data)
```swift
// ChatMessage.xcdatamodeld
entity ChatMessage {
    id: String
    conversationId: String
    content: String
    messageType: String
    sender: String
    timestamp: Double
    status: String
    metadata: String
}

entity Conversation {
    id: String
    title: String
    createdAt: Double
    updatedAt: Double
    messageCount: Int32
    isArchived: Bool
    aiPersonality: String
    messages: Set<ChatMessage>
}
```

### 鸿蒙 (关系型数据库)
```typescript
// 消息表结构
const CREATE_MESSAGES_TABLE = `
  CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    content TEXT NOT NULL,
    message_type TEXT NOT NULL,
    sender TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    status TEXT NOT NULL,
    metadata TEXT
  )
`;

// 对话表结构
const CREATE_CONVERSATIONS_TABLE = `
  CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    message_count INTEGER DEFAULT 0,
    is_archived INTEGER DEFAULT 0,
    ai_personality TEXT DEFAULT 'DEFAULT'
  )
`;
```

## 网络请求规范

### 统一错误处理
```json
{
  "success": false,
  "error": {
    "code": "AI_SERVICE_UNAVAILABLE",
    "message": "AI服务暂时不可用，请稍后重试",
    "details": {
      "timestamp": 1704067200000,
      "requestId": "req_123456"
    }
  }
}
```

### 常见错误码
- `INVALID_MESSAGE`: 消息格式无效
- `CONVERSATION_NOT_FOUND`: 对话不存在
- `AI_SERVICE_UNAVAILABLE`: AI服务不可用
- `RATE_LIMIT_EXCEEDED`: 请求频率超限
- `AUTHENTICATION_FAILED`: 认证失败
- `INSUFFICIENT_PERMISSIONS`: 权限不足

## 安全和隐私规范

### 1. 数据加密
- 本地存储的敏感消息内容需要加密
- 网络传输使用HTTPS/TLS
- API密钥安全存储

### 2. 隐私保护
- 用户可以删除对话历史
- 支持匿名模式对话
- 敏感信息过滤和脱敏

### 3. 权限管理
- 网络访问权限
- 存储权限
- 麦克风权限（语音输入）
- 相机权限（图片输入）