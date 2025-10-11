package crossmodal

import (
	"time"
)

// CrossModalInferenceRequest 跨模态推理请�?
type CrossModalInferenceRequest struct {
	Type        string                 `json:"type"`        // semantic_search, content_matching, emotion_analysis, etc.
	Data        map[string]interface{} `json:"data"`        // 请求数据
	Options     map[string]interface{} `json:"options"`     // 可选参�?
	Context     map[string]interface{} `json:"context"`     // 上下文信�?
	Timestamp   time.Time              `json:"timestamp"`   // 请求时间�?
}

// CrossModalInferenceResponse 跨模态推理响�?
type CrossModalInferenceResponse struct {
	Success     bool                   `json:"success"`     // 是否成功
	Result      map[string]interface{} `json:"result"`      // 结果数据
	Error       string                 `json:"error"`       // 错误信息
	Confidence  float64                `json:"confidence"`  // 置信�?
	Metadata    map[string]interface{} `json:"metadata"`    // 元数�?
	ProcessTime int64                  `json:"process_time"` // 处理时间(毫秒)
	Timestamp   time.Time              `json:"timestamp"`   // 响应时间�?
}

// CrossModalServiceConfig 跨模态服务配�?
type CrossModalServiceConfig struct {
	APIEndpoint     string        `json:"api_endpoint"`     // API端点
	APIKey          string        `json:"api_key"`          // API密钥
	Timeout         time.Duration `json:"timeout"`          // 超时时间
	MaxRetries      int           `json:"max_retries"`      // 最大重试次�?
	EnableCache     bool          `json:"enable_cache"`     // 是否启用缓存
	CacheExpiry     time.Duration `json:"cache_expiry"`     // 缓存过期时间
	ModelVersion    string        `json:"model_version"`    // 模型版本
	BatchSize       int           `json:"batch_size"`       // 批处理大�?
}
