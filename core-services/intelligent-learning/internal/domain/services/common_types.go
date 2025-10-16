package services

import (
	"time"
	"github.com/google/uuid"
)

// EmotionalState 情绪状?
type EmotionalState struct {
	Mood        string  `json:"mood"`        // 情绪
	Stress      float64 `json:"stress"`      // 压力水平
	Motivation  float64 `json:"motivation"`  // 动机水平
	Confidence  float64 `json:"confidence"`  // 自信水平
	Engagement  float64 `json:"engagement"`  // 参与?
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	Accuracy        float64 `json:"accuracy"`         // 准确?
	Speed           float64 `json:"speed"`            // 速度
	Efficiency      float64 `json:"efficiency"`       // 效率
	CompletionRate  float64 `json:"completion_rate"`  // 完成?
	ErrorRate       float64 `json:"error_rate"`       // 错误?
	Consistency     float64 `json:"consistency"`      // 一致?
	Timeline        string  `json:"timeline"`         // 时间?
	ExpectedOutcome string  `json:"expected_outcome"` // 预期结果
}

// Achievement 成就
type Achievement struct {
	ID          uuid.UUID `json:"id"`           // 成就ID
	Name        string    `json:"name"`         // 成就名称
	Description string    `json:"description"`  // 成就描述
	Type        string    `json:"type"`         // 成就类型
	Points      int       `json:"points"`       // 积分
	UnlockedAt  time.Time `json:"unlocked_at"`  // 解锁时间
}

// Position 位置
type Position struct {
	X float64 `json:"x"` // X坐标
	Y float64 `json:"y"` // Y坐标
	Z float64 `json:"z"` // Z坐标
}

// TrendDirection 趋势方向
type TrendDirection string

const (
	TrendDirectionUp    TrendDirection = "up"    // 上升
	TrendDirectionDown  TrendDirection = "down"  // 下降
	TrendDirectionFlat  TrendDirection = "flat"  // 平稳
)

// ConfidenceInterval 置信区间
type ConfidenceInterval struct {
	Lower      float64 `json:"lower"`       // 下界
	Upper      float64 `json:"upper"`       // 上界
	Confidence float64 `json:"confidence"`  // 置信?
}

// Evidence 证据
type Evidence struct {
	ID          uuid.UUID              `json:"id"`          // 证据ID
	Type        string                 `json:"type"`        // 证据类型
	Source      string                 `json:"source"`      // 证据来源
	Content     string                 `json:"content"`     // 证据内容
	Reliability float64                `json:"reliability"` // 可靠?
	Timestamp   time.Time              `json:"timestamp"`   // 时间?
	Metadata    map[string]interface{} `json:"metadata"`    // 元数?
}

// CachedInferenceResult 缓存的推理结?
type CachedInferenceResult struct {
	QueryID     string                 `json:"query_id"`     // 查询ID
	Result      interface{}            `json:"result"`       // 推理结果
	Confidence  float64                `json:"confidence"`   // 置信?
	CachedAt    time.Time              `json:"cached_at"`    // 缓存时间
	ExpiresAt   time.Time              `json:"expires_at"`   // 过期时间
	Metadata    map[string]interface{} `json:"metadata"`     // 元数?
}

// RecommendationType 推荐类型
type RecommendationType string

const (
	RecommendationTypeContent   RecommendationType = "content"   // 内容推荐
	RecommendationTypeActivity  RecommendationType = "activity"  // 活动推荐
	RecommendationTypePath      RecommendationType = "path"      // 路径推荐
	RecommendationTypeResource  RecommendationType = "resource"  // 资源推荐
)

// ValidationMethod 验证方法
type ValidationMethod string

const (
	ValidationMethodCrossValidation ValidationMethod = "cross_validation" // 交叉验证
	ValidationMethodHoldout         ValidationMethod = "holdout"          // 留出验证
	ValidationMethodBootstrap       ValidationMethod = "bootstrap"        // 自助验证
	ValidationMethodTimeSeriesSplit ValidationMethod = "time_series_split" // 时间序列分割
)

// Challenge 挑战
type Challenge struct {
	ChallengeID uuid.UUID                  `json:"challenge_id"` // 挑战ID
	Type        ChallengeType              `json:"type"`         // 挑战类型
	Name        string                     `json:"name"`         // 名称
	Description string                     `json:"description"`  // 描述
	Difficulty  float64                   `json:"difficulty"`   // 难度
	Reward      *Reward                   `json:"reward"`       // 奖励
	StartTime   time.Time                 `json:"start_time"`   // 开始时?
	EndTime     *time.Time                `json:"end_time"`     // 结束时间
	Progress    float64                   `json:"progress"`     // 进度
	Status      ChallengeStatus           `json:"status"`       // 状?
	Metadata    map[string]interface{}     `json:"metadata"`     // 元数?
}

// ChallengeType 挑战类型
type ChallengeType string

const (
	ChallengeTypeDaily    ChallengeType = "daily"    // 每日
	ChallengeTypeWeekly   ChallengeType = "weekly"   // 每周
	ChallengeTypeMonthly  ChallengeType = "monthly"  // 每月
	ChallengeTypeSpecial  ChallengeType = "special"  // 特殊
	ChallengeTypePersonal ChallengeType = "personal" // 个人
	ChallengeTypeGroup    ChallengeType = "group"    // 群组
)

// ChallengeStatus 挑战状?
type ChallengeStatus string

const (
	ChallengeStatusActive    ChallengeStatus = "active"    // 活跃
	ChallengeStatusCompleted ChallengeStatus = "completed" // 完成
	ChallengeStatusFailed    ChallengeStatus = "failed"    // 失败
	ChallengeStatusExpired   ChallengeStatus = "expired"   // 过期
)

// Reward 奖励
type Reward struct {
	Type        RewardType                 `json:"type"`        // 奖励类型
	Value       interface{}                `json:"value"`       // 奖励?
	Description string                     `json:"description"` // 描述
	Metadata    map[string]interface{}     `json:"metadata"`    // 元数?
}

// RewardType 奖励类型
type RewardType string

const (
	RewardTypePoints      RewardType = "points"      // 积分
	RewardTypeBadge       RewardType = "badge"       // 徽章
	RewardTypeCertificate RewardType = "certificate" // 证书
	RewardTypeUnlock      RewardType = "unlock"      // 解锁
)

// ModalityType 模态类?
type ModalityType string

const (
	// 基础模态类?
	ModalityTypeText        ModalityType = "text"        // 文本
	ModalityTypeImage       ModalityType = "image"       // 图像
	ModalityTypeAudio       ModalityType = "audio"       // 音频
	ModalityTypeVideo       ModalityType = "video"       // 视频
	ModalityTypeGraph       ModalityType = "graph"       // 图表
	ModalityTypeTabular     ModalityType = "tabular"     // 表格
	ModalityTypeTime        ModalityType = "time_series" // 时间序列
	ModalityTypeSpatial     ModalityType = "spatial"     // 空间
	ModalityTypeMultimodal  ModalityType = "multimodal"  // 多模?
	
	// 学习模态类?
	ModalityTypeVisual      ModalityType = "visual"      // 视觉
	ModalityTypeAuditory    ModalityType = "auditory"    // 听觉
	ModalityTypeKinesthetic ModalityType = "kinesthetic" // 动觉
	ModalityTypeReading     ModalityType = "reading"     // 阅读
	ModalityTypeWriting     ModalityType = "writing"     // 写作
)

// ImplementationPlan 实施计划
type ImplementationPlan struct {
	PlanID      uuid.UUID                  `json:"plan_id"`      // 计划ID
	Name        string                     `json:"name"`         // 名称
	Description string                     `json:"description"`  // 描述
	Steps       []ImplementationStep       `json:"steps"`        // 步骤
	Timeline    *Timeline                  `json:"timeline"`     // 时间?
	Resources   []Resource                 `json:"resources"`    // 资源
	Status      PlanStatus                 `json:"status"`       // 状?
	Progress    float64                    `json:"progress"`     // 进度
	Metadata    map[string]interface{}     `json:"metadata"`     // 元数?
}

// ImplementationStep 实施步骤
type ImplementationStep struct {
	StepID      uuid.UUID                  `json:"step_id"`      // 步骤ID
	Name        string                     `json:"name"`         // 名称
	Description string                     `json:"description"`  // 描述
	Order       int                        `json:"order"`        // 顺序
	Duration    time.Duration              `json:"duration"`     // 持续时间
	Dependencies []uuid.UUID               `json:"dependencies"` // 依赖
	Status      StepStatus                 `json:"status"`       // 状?
	Metadata    map[string]interface{}     `json:"metadata"`     // 元数?
}

// Timeline 时间?
type Timeline struct {
	StartTime   time.Time                  `json:"start_time"`   // 开始时?
	EndTime     time.Time                  `json:"end_time"`     // 结束时间
	Milestones  []Milestone                `json:"milestones"`   // 里程?
	Metadata    map[string]interface{}     `json:"metadata"`     // 元数?
}

// Milestone 里程?
type Milestone struct {
	MilestoneID uuid.UUID                  `json:"milestone_id"` // 里程碑ID
	Name        string                     `json:"name"`         // 名称
	Description string                     `json:"description"`  // 描述
	TargetDate  time.Time                  `json:"target_date"`  // 目标日期
	Status      MilestoneStatus            `json:"status"`       // 状?
	Metadata    map[string]interface{}     `json:"metadata"`     // 元数?
}

// Resource 资源
type Resource struct {
	ResourceID  uuid.UUID                  `json:"resource_id"`  // 资源ID
	Type        ResourceType               `json:"type"`         // 类型
	Name        string                     `json:"name"`         // 名称
	Description string                     `json:"description"`  // 描述
	Quantity    float64                    `json:"quantity"`     // 数量
	Unit        string                     `json:"unit"`         // 单位
	Status      ResourceStatus             `json:"status"`       // 状?
	Metadata    map[string]interface{}     `json:"metadata"`     // 元数?
}

// PlanStatus 计划状?
type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "draft"      // 草稿
	PlanStatusActive     PlanStatus = "active"     // 活跃
	PlanStatusPaused     PlanStatus = "paused"     // 暂停
	PlanStatusCompleted  PlanStatus = "completed"  // 完成
	PlanStatusCancelled  PlanStatus = "cancelled"  // 取消
)

// StepStatus 步骤状?
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"    // 待处?
	StepStatusInProgress StepStatus = "in_progress" // 进行?
	StepStatusCompleted  StepStatus = "completed"  // 完成
	StepStatusSkipped    StepStatus = "skipped"    // 跳过
	StepStatusFailed     StepStatus = "failed"     // 失败
)

// MilestoneStatus 里程碑状?
type MilestoneStatus string

const (
	MilestoneStatusPending   MilestoneStatus = "pending"   // 待处?
	MilestoneStatusAchieved  MilestoneStatus = "achieved"  // 已达?
	MilestoneStatusMissed    MilestoneStatus = "missed"    // 错过
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeHuman     ResourceType = "human"     // 人力
	ResourceTypeMaterial  ResourceType = "material"  // 物料
	ResourceTypeFinancial ResourceType = "financial" // 财务
	ResourceTypeTechnical ResourceType = "technical" // 技?
)

// ResourceStatus 资源状?
type ResourceStatus string

const (
	ResourceStatusAvailable ResourceStatus = "available" // 可用
	ResourceStatusAllocated ResourceStatus = "allocated" // 已分?
	ResourceStatusExhausted ResourceStatus = "exhausted" // 耗尽
)

