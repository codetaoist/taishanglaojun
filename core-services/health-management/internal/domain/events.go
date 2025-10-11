package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent 领域事件接口
type DomainEvent interface {
	GetEventID() uuid.UUID
	GetEventType() string
	GetAggregateID() uuid.UUID
	GetOccurredAt() time.Time
	GetEventData() map[string]interface{}
}

// BaseDomainEvent 基础领域事件
type BaseDomainEvent struct {
	EventID     uuid.UUID              `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID uuid.UUID              `json:"aggregate_id"`
	OccurredAt  time.Time              `json:"occurred_at"`
	EventData   map[string]interface{} `json:"event_data"`
}

func (e BaseDomainEvent) GetEventID() uuid.UUID                 { return e.EventID }
func (e BaseDomainEvent) GetEventType() string                  { return e.EventType }
func (e BaseDomainEvent) GetAggregateID() uuid.UUID             { return e.AggregateID }
func (e BaseDomainEvent) GetOccurredAt() time.Time              { return e.OccurredAt }
func (e BaseDomainEvent) GetEventData() map[string]interface{}  { return e.EventData }

// 健康数据相关事件

// HealthDataCreatedEvent 健康数据创建事件
type HealthDataCreatedEvent struct {
	BaseDomainEvent
}

func NewHealthDataCreatedEvent(healthDataID, userID uuid.UUID, dataType HealthDataType, value float64, unit string, source HealthDataSource) *HealthDataCreatedEvent {
	return &HealthDataCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_data.created",
			AggregateID: healthDataID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":   userID,
				"data_type": dataType,
				"value":     value,
				"unit":      unit,
				"source":    source,
			},
		},
	}
}

// HealthDataUpdatedEvent 健康数据更新事件
type HealthDataUpdatedEvent struct {
	BaseDomainEvent
}

func NewHealthDataUpdatedEvent(healthDataID, userID uuid.UUID, dataType HealthDataType, oldValue, newValue float64, unit string) *HealthDataUpdatedEvent {
	return &HealthDataUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_data.updated",
			AggregateID: healthDataID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":   userID,
				"data_type": dataType,
				"old_value": oldValue,
				"new_value": newValue,
				"unit":      unit,
			},
		},
	}
}

// AbnormalHealthDataDetectedEvent 异常健康数据检测事�?
type AbnormalHealthDataDetectedEvent struct {
	BaseDomainEvent
}

func NewAbnormalHealthDataDetectedEvent(healthDataID, userID uuid.UUID, dataType HealthDataType, value float64, riskLevel RiskLevel) *AbnormalHealthDataDetectedEvent {
	return &AbnormalHealthDataDetectedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_data.abnormal_detected",
			AggregateID: healthDataID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":    userID,
				"data_type":  dataType,
				"value":      value,
				"risk_level": riskLevel,
			},
		},
	}
}

// 健康档案相关事件

// HealthProfileCreatedEvent 健康档案创建事件
type HealthProfileCreatedEvent struct {
	BaseDomainEvent
}

func NewHealthProfileCreatedEvent(profileID, userID uuid.UUID, gender Gender) *HealthProfileCreatedEvent {
	return &HealthProfileCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_profile.created",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id": userID,
				"gender":  gender,
			},
		},
	}
}

// HealthProfileUpdatedEvent 健康档案更新事件
type HealthProfileUpdatedEvent struct {
	BaseDomainEvent
}

func NewHealthProfileUpdatedEvent(profileID, userID uuid.UUID, updateType string) *HealthProfileUpdatedEvent {
	return &HealthProfileUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_profile.updated",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":     userID,
				"update_type": updateType,
			},
		},
	}
}

// MedicalHistoryAddedEvent 病史添加事件
type MedicalHistoryAddedEvent struct {
	BaseDomainEvent
}

func NewMedicalHistoryAddedEvent(profileID, userID uuid.UUID, condition string) *MedicalHistoryAddedEvent {
	return &MedicalHistoryAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "medical_history.added",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":   userID,
				"condition": condition,
			},
		},
	}
}

// MedicalHistoryRemovedEvent 病史移除事件
type MedicalHistoryRemovedEvent struct {
	BaseDomainEvent
}

func NewMedicalHistoryRemovedEvent(profileID, userID uuid.UUID, condition string) *MedicalHistoryRemovedEvent {
	return &MedicalHistoryRemovedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "medical_history.removed",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":   userID,
				"condition": condition,
			},
		},
	}
}

// AllergyAddedEvent 过敏史添加事�?
type AllergyAddedEvent struct {
	BaseDomainEvent
}

func NewAllergyAddedEvent(profileID, userID uuid.UUID, allergen string) *AllergyAddedEvent {
	return &AllergyAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "allergy.added",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":  userID,
				"allergen": allergen,
			},
		},
	}
}

// AllergyRemovedEvent 过敏史移除事�?
type AllergyRemovedEvent struct {
	BaseDomainEvent
}

func NewAllergyRemovedEvent(profileID, userID uuid.UUID, allergen string) *AllergyRemovedEvent {
	return &AllergyRemovedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "allergy.removed",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":  userID,
				"allergen": allergen,
			},
		},
	}
}

// MedicationAddedEvent 用药史添加事�?
type MedicationAddedEvent struct {
	BaseDomainEvent
}

func NewMedicationAddedEvent(profileID, userID uuid.UUID, medication string) *MedicationAddedEvent {
	return &MedicationAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "medication.added",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":    userID,
				"medication": medication,
			},
		},
	}
}

// MedicationRemovedEvent 用药史移除事�?
type MedicationRemovedEvent struct {
	BaseDomainEvent
}

func NewMedicationRemovedEvent(profileID, userID uuid.UUID, medication string) *MedicationRemovedEvent {
	return &MedicationRemovedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "medication.removed",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":    userID,
				"medication": medication,
			},
		},
	}
}

// HealthGoalsSetEvent 健康目标设置事件
type HealthGoalsSetEvent struct {
	BaseDomainEvent
}

func NewHealthGoalsSetEvent(profileID, userID uuid.UUID, goals []string) *HealthGoalsSetEvent {
	return &HealthGoalsSetEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_goals.set",
			AggregateID: profileID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id": userID,
				"goals":   goals,
			},
		},
	}
}

// 健康报告相关事件

// HealthReportGeneratedEvent 健康报告生成事件
type HealthReportGeneratedEvent struct {
	BaseDomainEvent
}

func NewHealthReportGeneratedEvent(reportID, userID uuid.UUID, reportType string, period string) *HealthReportGeneratedEvent {
	return &HealthReportGeneratedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_report.generated",
			AggregateID: reportID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":     userID,
				"report_type": reportType,
				"period":      period,
			},
		},
	}
}

// HealthAlertTriggeredEvent 健康警报触发事件
type HealthAlertTriggeredEvent struct {
	BaseDomainEvent
}

func NewHealthAlertTriggeredEvent(alertID, userID uuid.UUID, alertType string, severity string, message string) *HealthAlertTriggeredEvent {
	return &HealthAlertTriggeredEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "health_alert.triggered",
			AggregateID: alertID,
			OccurredAt:  time.Now(),
			EventData: map[string]interface{}{
				"user_id":    userID,
				"alert_type": alertType,
				"severity":   severity,
				"message":    message,
			},
		},
	}
}
