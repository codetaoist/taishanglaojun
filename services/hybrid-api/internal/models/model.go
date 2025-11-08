package models

import (
	"time"

	"gorm.io/datatypes"
)

// ModelConfig 模型配置
type ModelConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	ServiceType string    `json:"serviceType" gorm:"not null"` // 对应数据库中的service_type
	Endpoint    string    `json:"endpoint" gorm:"not null"`   // 对应数据库中的endpoint
	APIKey      string    `json:"apiKey,omitempty"`
	ModelID     string    `json:"modelId" gorm:"not null"`    // 对应数据库中的model_id
	MaxTokens   int       `json:"maxTokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	IsDefault   bool      `json:"isDefault" gorm:"default:false"` // 对应数据库中的is_default
	IsActive    bool      `json:"isActive" gorm:"default:true"`   // 对应数据库中的is_active
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (ModelConfig) TableName() string {
	return "tai_model_configs"
}

// Provider 为了兼容性，提供Provider方法
func (m *ModelConfig) Provider() string {
	return m.ServiceType
}

// Model 为了兼容性，提供Model方法
func (m *ModelConfig) Model() string {
	return m.ModelID
}

// BaseURL 为了兼容性，提供BaseURL方法
func (m *ModelConfig) BaseURL() string {
	return m.Endpoint
}

// Enabled 为了兼容性，提供Enabled方法
func (m *ModelConfig) Enabled() bool {
	return m.IsActive
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Provider    string            `json:"provider"`
	Type        string            `json:"type"` // text, embedding, multimodal
	MaxTokens   int               `json:"maxTokens,omitempty"`
	CostPer1K   float64           `json:"costPer1K,omitempty"`
	Capabilities []string         `json:"capabilities,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt,omitempty"`
}

// TextGenerationRequest 文本生成请求
type TextGenerationRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	Prompt      string                 `json:"prompt,omitempty"` // 兼容单提示词模式
	MaxTokens   int                    `json:"maxTokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"topP,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// TextGenerationResponse 文本生成响应
type TextGenerationResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Content string    `json:"content,omitempty"` // 生成的内容
}

// TextGenerationChunk 文本生成流式响应块
type TextGenerationChunk struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Content string    `json:"content,omitempty"` // 流式内容
	Done    bool      `json:"done,omitempty"`    // 是否结束
}

// EmbeddingRequest 嵌入生成请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	Text  string   `json:"text,omitempty"`  // 兼容单文本输入
	User  string   `json:"user,omitempty"`  // 用户标识
}

// EmbeddingResponse 嵌入生成响应
type EmbeddingResponse struct {
	Object    string      `json:"object"`
	Data      []Embedding `json:"data"`
	Model     string      `json:"model"`
	Usage     Usage       `json:"usage"`
	Embedding []float64   `json:"embedding,omitempty"` // 兼容单嵌入向量
}

// Message 对话消息
type Message struct {
	ID             string          `json:"id" gorm:"primaryKey"`
	ConversationID string          `json:"conversationId"`
	UserID         string          `json:"userId"`
	Role           string          `json:"role"` // system, user, assistant
	Content        string          `json:"content"`
	Metadata       datatypes.JSON  `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Message
func (Message) TableName() string {
	return "tai_messages"
}

// Choice 选择项
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Text         string  `json:"text,omitempty"`
	FinishReason string  `json:"finish_reason,omitempty"`
	Delta        Message `json:"delta,omitempty"` // 用于流式响应
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Embedding 嵌入向量
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// Conversation 对话
type Conversation struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	UserID      string                 `json:"userId"`
	Title       string                 `json:"title"`
	ModelConfig datatypes.JSON         `json:"modelConfig" gorm:"type:jsonb"` // JSON格式的模型配置
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	
	// 关联字段，不存储在数据库中
	Messages []Message `json:"messages,omitempty" gorm:"-"`
}

// TableName returns the table name for Conversation
func (Conversation) TableName() string {
	return "tai_conversations"
}

// FineTuningJob 微调作业
type FineTuningJob struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	BaseModel   string    `json:"baseModel"`
	TrainingData string   `json:"trainingData"` // JSON格式的训练数据
	Status      string    `json:"status"`       // preparing, running, completed, failed
	Progress    int       `json:"progress"`     // 0-100
	ErrorMessage string   `json:"errorMessage,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	// 兼容字段
	Model string `json:"model,omitempty"`
}

// EmbeddingsRequest 嵌入生成请求（别名）
type EmbeddingsRequest struct {
	Model string   `json:"model"`
	Texts []string `json:"texts"` // 批量文本输入
	User  string   `json:"user,omitempty"` // 用户标识
}

// EmbeddingsResponse 嵌入生成响应（别名）
type EmbeddingsResponse struct {
	Embeddings [][]float64 `json:"embeddings"` // 批量嵌入向量
	Model      string      `json:"model"`
	
	// 兼容字段
	Usage UsageInfo `json:"usage,omitempty"`
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	Title       string                 `json:"title"`
	ModelConfig map[string]interface{} `json:"modelConfig"` // JSON格式的模型配置
	Messages    []Message              `json:"messages,omitempty"`
	
	// 兼容字段
	UserID string `json:"userId,omitempty"`
	Model  string `json:"model,omitempty"`
}

// UpdateConversationRequest 更新对话请求
type UpdateConversationRequest struct {
	Title       string                 `json:"title,omitempty"`
	Model       string                 `json:"model,omitempty"`        // 兼容字段
	ModelConfig map[string]interface{} `json:"modelConfig,omitempty"` // JSON格式的模型配置
	Messages    []Message              `json:"messages,omitempty"`
}

// AddMessageRequest 添加消息请求
type AddMessageRequest struct {
	ConversationID string    `json:"conversationId"`
	Message        Message   `json:"message"`
	
	// 兼容字段
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// UpdateMessageRequest 更新消息请求
type UpdateMessageRequest struct {
	ConversationID string    `json:"conversationId"`
	MessageIndex   int       `json:"messageIndex"`
	Message        Message   `json:"message"`
	
	// 兼容字段
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ToolExecutionRequest 工具执行请求
type ToolExecutionRequest struct {
	ConversationID string `json:"conversationId"`
	ToolName       string `json:"toolName"`
	Parameters     string `json:"parameters"` // JSON格式的参数
}

// ToolExecutionResponse 工具执行响应
type ToolExecutionResponse struct {
	ID        string    `json:"id"`
	RequestID string    `json:"requestId"`
	ToolName  string    `json:"toolName"`
	Result    string    `json:"result"` // JSON格式的结果
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateFineTuningJobRequest 创建微调作业请求
type CreateFineTuningJobRequest struct {
	Name         string `json:"name"`
	BaseModel    string `json:"baseModel"`
	TrainingData string `json:"trainingData"` // JSON格式的训练数据
	
	// 兼容字段
	Model string `json:"model,omitempty"`
}

// ModelServiceInfo 模型服务信息
type ModelServiceInfo struct {
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	Status       string   `json:"status"`
	Models       []string `json:"models"`
	Capabilities []string `json:"capabilities"`
	
	// 兼容字段
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Features    []string `json:"features,omitempty"`
}