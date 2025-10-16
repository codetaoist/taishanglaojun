package domain

import (
	"time"

	"github.com/google/uuid"
)

// HealthDataType 
type HealthDataType string

const (
	// 
	HeartRate       HealthDataType = "heart_rate"       // 
	BloodPressure   HealthDataType = "blood_pressure"   // ?
	BloodSugar      HealthDataType = "blood_sugar"      // ?
	BodyTemperature HealthDataType = "body_temperature" // 
	Weight          HealthDataType = "weight"           // 
	Height          HealthDataType = "height"           // 
	BMI             HealthDataType = "bmi"              // 
	
	// 
	Steps         HealthDataType = "steps"          // 
	Distance      HealthDataType = "distance"       // 
	Calories      HealthDataType = "calories"       // ?
	ExerciseTime  HealthDataType = "exercise_time"  // 
	
	// 
	SleepDuration HealthDataType = "sleep_duration" // 
	SleepQuality  HealthDataType = "sleep_quality"  // 
	DeepSleep     HealthDataType = "deep_sleep"     // 
	LightSleep    HealthDataType = "light_sleep"    // 
	
	// 
	StressLevel   HealthDataType = "stress_level"   // 
	MoodScore     HealthDataType = "mood_score"     // 
	AnxietyLevel  HealthDataType = "anxiety_level"  // 
)

// HealthDataSource 
type HealthDataSource string

const (
	ManualInput    HealthDataSource = "manual"      // 
	SmartWatch     HealthDataSource = "smart_watch" // 
	SmartPhone     HealthDataSource = "smart_phone" // 
	MedicalDevice  HealthDataSource = "medical"     // 豸
	ThirdPartyApp  HealthDataSource = "third_party" // ?
)

// HealthData ?
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
	
	// 
	events []DomainEvent
}

// NewHealthData 
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
	
	// 
	healthData.publishEvent(NewHealthDataCreatedEvent(id, userID, dataType, value, unit, source))
	
	return healthData
}

// SetDeviceID 豸ID
func (h *HealthData) SetDeviceID(deviceID string) {
	h.DeviceID = &deviceID
	h.UpdatedAt = time.Now()
}

// SetMetadata ?
func (h *HealthData) SetMetadata(metadata map[string]interface{}) {
	h.Metadata = metadata
	h.UpdatedAt = time.Now()
}

// UpdateValue ?
func (h *HealthData) UpdateValue(value float64, unit string) {
	oldValue := h.Value
	h.Value = value
	h.Unit = unit
	h.UpdatedAt = time.Now()
	
	// ?
	h.publishEvent(NewHealthDataUpdatedEvent(h.ID, h.UserID, h.DataType, oldValue, value, unit))
}

// IsAbnormal 
func (h *HealthData) IsAbnormal() bool {
	switch h.DataType {
	case HeartRate:
		// ?0-100 bpm
		return h.Value < 60 || h.Value > 100
	case BloodPressure:
		// 
		return h.Value > 140 || h.Value < 90
	case BloodSugar:
		// 3.9-6.1 mmol/L
		return h.Value < 3.9 || h.Value > 6.1
	case BodyTemperature:
		// ?6.1-37.2C
		return h.Value < 36.1 || h.Value > 37.2
	case BMI:
		// BMI?18.5 ?>24
		return h.Value < 18.5 || h.Value > 24
	default:
		return false
	}
}

// GetRiskLevel 
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

// publishEvent 
func (h *HealthData) publishEvent(event DomainEvent) {
	h.events = append(h.events, event)
}

// GetEvents 
func (h *HealthData) GetEvents() []DomainEvent {
	return h.events
}

// ClearEvents 
func (h *HealthData) ClearEvents() {
	h.events = make([]DomainEvent, 0)
}

// RiskLevel 
type RiskLevel string

const (
	RiskLevelNormal RiskLevel = "normal" // 
	RiskLevelLow    RiskLevel = "low"    // ?
	RiskLevelMedium RiskLevel = "medium" // ?
	RiskLevelHigh   RiskLevel = "high"   // ?
)

