package domain

import (
	"time"

	"github.com/google/uuid"
)

// HealthDataType 健康数据类型枚举
type HealthDataType string

const (
	// 生理指标
	HeartRate       HealthDataType = "heart_rate"       // 心率
	BloodPressure   HealthDataType = "blood_pressure"   // 血压
	BloodSugar      HealthDataType = "blood_sugar"      // 血糖
	BodyTemperature HealthDataType = "body_temperature" // 体温
	Weight          HealthDataType = "weight"           // 体重
	Height          HealthDataType = "height"           // 身高
	BMI             HealthDataType = "bmi"              // 身体质量指数
	
	// 运动数据
	Steps         HealthDataType = "steps"          // 步数
	Distance      HealthDataType = "distance"       // 距离
	Calories      HealthDataType = "calories"       // 卡路里
	ExerciseTime  HealthDataType = "exercise_time"  // 运动时间
	
	// 睡眠数据
	SleepDuration HealthDataType = "sleep_duration" // 睡眠时长
	SleepQuality  HealthDataType = "sleep_quality"  // 睡眠质量
	DeepSleep     HealthDataType = "deep_sleep"     // 深度睡眠
	LightSleep    HealthDataType = "light_sleep"    // 浅度睡眠
	
	// 心理健康
	StressLevel   HealthDataType = "stress_level"   // 压力水平
	MoodScore     HealthDataType = "mood_score"     // 情绪评分
	AnxietyLevel  HealthDataType = "anxiety_level"  // 焦虑水平
)

// HealthDataSource 健康数据来源
type HealthDataSource string

const (
	ManualInput    HealthDataSource = "manual"      // 手动输入
	SmartWatch     HealthDataSource = "smart_watch" // 智能手表
	SmartPhone     HealthDataSource = "smart_phone" // 智能手机
	MedicalDevice  HealthDataSource = "medical"     // 医疗设备
	ThirdPartyApp  HealthDataSource = "third_party" // 第三方应用
)

// HealthData 健康数据聚合根
type HealthData struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	DataType    HealthDataType   `json:"data_type" gorm:"type:varchar(50);not null;index"`
	Value       float64          `json:"value" gorm:"not null"`
	Unit        string           `json:"unit" gorm:"type:varchar(20)"`
	Source      HealthDataSource `json:"source" gorm:"type:varchar(20);not null"`
	DeviceID    *string          `json:"device_id,omitempty" gorm:"type:varchar(100)"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	RecordedAt  time.Time        `json:"recorded_at" gorm:"not null;index"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 领域事件
	events []DomainEvent
}

// NewHealthData 创建新的健康数据
func NewHealthData(userID uuid.UUID, dataType HealthDataType, value float64, unit string, source HealthDataSource) *HealthData {
	id := uuid.New()
	now := time.Now()
	
	healthData := &HealthData{
		ID:         id,
		UserID:     userID,
		DataType:   dataType,
		Value:      value,
		Unit:       unit,
		Source:     source,
		RecordedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
		events:     make([]DomainEvent, 0),
	}
	
	// 发布健康数据创建事件
	healthData.publishEvent(NewHealthDataCreatedEvent(id, userID, dataType, value, unit, source))
	
	return healthData
}

// SetDeviceID 设置设备ID
func (h *HealthData) SetDeviceID(deviceID string) {
	h.DeviceID = &deviceID
	h.UpdatedAt = time.Now()
}

// SetMetadata 设置元数据
func (h *HealthData) SetMetadata(metadata map[string]interface{}) {
	h.Metadata = metadata
	h.UpdatedAt = time.Now()
}

// UpdateValue 更新数值
func (h *HealthData) UpdateValue(value float64, unit string) {
	oldValue := h.Value
	h.Value = value
	h.Unit = unit
	h.UpdatedAt = time.Now()
	
	// 发布数值更新事件
	h.publishEvent(NewHealthDataUpdatedEvent(h.ID, h.UserID, h.DataType, oldValue, value, unit))
}

// IsAbnormal 判断数据是否异常
func (h *HealthData) IsAbnormal() bool {
	switch h.DataType {
	case HeartRate:
		// 正常心率范围：60-100 bpm
		return h.Value < 60 || h.Value > 100
	case BloodPressure:
		// 这里简化处理，实际应该分收缩压和舒张压
		return h.Value > 140 || h.Value < 90
	case BloodSugar:
		// 正常血糖范围：3.9-6.1 mmol/L
		return h.Value < 3.9 || h.Value > 6.1
	case BodyTemperature:
		// 正常体温范围：36.1-37.2°C
		return h.Value < 36.1 || h.Value > 37.2
	case BMI:
		// BMI异常范围：<18.5 或 >24
		return h.Value < 18.5 || h.Value > 24
	default:
		return false
	}
}

// GetRiskLevel 获取风险等级
func (h *HealthData) GetRiskLevel() RiskLevel {
	if !h.IsAbnormal() {
		return RiskLevelNormal
	}
	
	switch h.DataType {
	case HeartRate:
		if h.Value < 50 || h.Value > 120 {
			return RiskLevelHigh
		}
		return RiskLevelMedium
	case BloodPressure:
		if h.Value > 160 || h.Value < 80 {
			return RiskLevelHigh
		}
		return RiskLevelMedium
	case BloodSugar:
		if h.Value < 3.0 || h.Value > 10.0 {
			return RiskLevelHigh
		}
		return RiskLevelMedium
	case BodyTemperature:
		if h.Value < 35.0 || h.Value > 39.0 {
			return RiskLevelHigh
		}
		return RiskLevelMedium
	default:
		return RiskLevelLow
	}
}

// publishEvent 发布领域事件
func (h *HealthData) publishEvent(event DomainEvent) {
	h.events = append(h.events, event)
}

// GetEvents 获取领域事件
func (h *HealthData) GetEvents() []DomainEvent {
	return h.events
}

// ClearEvents 清除领域事件
func (h *HealthData) ClearEvents() {
	h.events = make([]DomainEvent, 0)
}

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLevelNormal RiskLevel = "normal" // 正常
	RiskLevelLow    RiskLevel = "low"    // 低风险
	RiskLevelMedium RiskLevel = "medium" // 中风险
	RiskLevelHigh   RiskLevel = "high"   // 高风险
)