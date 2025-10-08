package services

import (
	"time"
)

// CrossModalInferenceRequest 跨模态推理请求
type CrossModalInferenceRequest struct {
	Type        string                 `json:"type"`        // semantic_search, content_matching, emotion_analysis, etc.
	Data        map[string]interface{} `json:"data"`        // 请求数据
	Options     map[string]interface{} `json:"options"`     // 可选参数
	Context     map[string]interface{} `json:"context"`     // 上下文信息
	Timestamp   time.Time              `json:"timestamp"`   // 请求时间戳
}

// CrossModalInferenceResponse 跨模态推理响应
type CrossModalInferenceResponse struct {
	Success     bool                   `json:"success"`     // 是否成功
	Result      map[string]interface{} `json:"result"`      // 结果数据
	Error       string                 `json:"error"`       // 错误信息
	Confidence  float64                `json:"confidence"`  // 置信度
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
	ProcessTime int64                  `json:"process_time"` // 处理时间(毫秒)
	Timestamp   time.Time              `json:"timestamp"`   // 响应时间戳
}