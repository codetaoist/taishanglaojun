package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// HealthDataRepository 健康数据仓储接口
type HealthDataRepository interface {
	// 基础CRUD操作
	Save(ctx context.Context, healthData *HealthData) error
	FindByID(ctx context.Context, id uuid.UUID) (*HealthData, error)
	Update(ctx context.Context, healthData *HealthData) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// 查询操作
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*HealthData, error)
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType HealthDataType, limit, offset int) ([]*HealthData, error)
	FindByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*HealthData, error)
	FindByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType HealthDataType, startTime, endTime time.Time) ([]*HealthData, error)
	
	// 统计操作
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType HealthDataType) (int64, error)
	
	// 聚合操作
	GetLatestByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType HealthDataType) (*HealthData, error)
	GetAverageByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType HealthDataType, startTime, endTime time.Time) (float64, error)
	GetMinMaxByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType HealthDataType, startTime, endTime time.Time) (min, max float64, err error)
	
	// 异常数据查询
	FindAbnormalDataByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*HealthData, error)
	FindAbnormalDataByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*HealthData, error)
}

// HealthProfileRepository 健康档案仓储接口
type HealthProfileRepository interface {
	// 基础CRUD操作
	Save(ctx context.Context, profile *HealthProfile) error
	FindByID(ctx context.Context, id uuid.UUID) (*HealthProfile, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*HealthProfile, error)
	Update(ctx context.Context, profile *HealthProfile) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// 查询操作
	ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
	FindByGender(ctx context.Context, gender Gender, limit, offset int) ([]*HealthProfile, error)
	FindByAgeRange(ctx context.Context, minAge, maxAge int, limit, offset int) ([]*HealthProfile, error)
	
	// 统计操作
	CountTotal(ctx context.Context) (int64, error)
	CountByGender(ctx context.Context, gender Gender) (int64, error)
}

// HealthReportRepository 健康报告仓储接口
type HealthReportRepository interface {
	// 基础CRUD操作
	Save(ctx context.Context, report *HealthReport) error
	FindByID(ctx context.Context, id uuid.UUID) (*HealthReport, error)
	Update(ctx context.Context, report *HealthReport) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// 查询操作
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*HealthReport, error)
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, reportType string, limit, offset int) ([]*HealthReport, error)
	FindByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*HealthReport, error)
	
	// 统计操作
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	GetLatestByUserIDAndType(ctx context.Context, userID uuid.UUID, reportType string) (*HealthReport, error)
}

// HealthAlertRepository 健康警报仓储接口
type HealthAlertRepository interface {
	// 基础CRUD操作
	Save(ctx context.Context, alert *HealthAlert) error
	FindByID(ctx context.Context, id uuid.UUID) (*HealthAlert, error)
	Update(ctx context.Context, alert *HealthAlert) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// 查询操作
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*HealthAlert, error)
	FindByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*HealthAlert, error)
	FindByUserIDAndSeverity(ctx context.Context, userID uuid.UUID, severity string, limit, offset int) ([]*HealthAlert, error)
	FindUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*HealthAlert, error)
	
	// 统计操作
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// HealthReport 健康报告聚合�?
type HealthReport struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	ReportType  string                 `json:"report_type" gorm:"type:varchar(50);not null"`
	Period      string                 `json:"period" gorm:"type:varchar(20);not null"`
	StartDate   time.Time              `json:"start_date" gorm:"not null"`
	EndDate     time.Time              `json:"end_date" gorm:"not null"`
	Summary     map[string]interface{} `json:"summary" gorm:"type:jsonb"`
	Insights    []string               `json:"insights" gorm:"type:jsonb"`
	Recommendations []string           `json:"recommendations" gorm:"type:jsonb"`
	GeneratedAt time.Time              `json:"generated_at" gorm:"not null"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 领域事件
	events []DomainEvent
}

// HealthAlert 健康警报聚合�?
type HealthAlert struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	AlertType   string                 `json:"alert_type" gorm:"type:varchar(50);not null"`
	Severity    string                 `json:"severity" gorm:"type:varchar(20);not null"`
	Title       string                 `json:"title" gorm:"type:varchar(200);not null"`
	Message     string                 `json:"message" gorm:"type:text;not null"`
	Status      string                 `json:"status" gorm:"type:varchar(20);not null;default:'unread'"`
	SourceID    *uuid.UUID             `json:"source_id,omitempty" gorm:"type:uuid"`
	SourceType  *string                `json:"source_type,omitempty" gorm:"type:varchar(50)"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	TriggeredAt time.Time              `json:"triggered_at" gorm:"not null"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 领域事件
	events []DomainEvent
}

// NewHealthReport 创建新的健康报告
func NewHealthReport(userID uuid.UUID, reportType, period string, startDate, endDate time.Time) *HealthReport {
	id := uuid.New()
	now := time.Now()
	
	report := &HealthReport{
		ID:          id,
		UserID:      userID,
		ReportType:  reportType,
		Period:      period,
		StartDate:   startDate,
		EndDate:     endDate,
		GeneratedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		events:      make([]DomainEvent, 0),
	}
	
	// 发布健康报告生成事件
	report.publishEvent(NewHealthReportGeneratedEvent(id, userID, reportType, period))
	
	return report
}

// NewHealthAlert 创建新的健康警报
func NewHealthAlert(userID uuid.UUID, alertType, severity, title, message string) *HealthAlert {
	id := uuid.New()
	now := time.Now()
	
	alert := &HealthAlert{
		ID:          id,
		UserID:      userID,
		AlertType:   alertType,
		Severity:    severity,
		Title:       title,
		Message:     message,
		Status:      "unread",
		TriggeredAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		events:      make([]DomainEvent, 0),
	}
	
	// 发布健康警报触发事件
	alert.publishEvent(NewHealthAlertTriggeredEvent(id, userID, alertType, severity, message))
	
	return alert
}

// MarkAsRead 标记警报为已�?
func (h *HealthAlert) MarkAsRead() {
	if h.Status != "read" {
		h.Status = "read"
		now := time.Now()
		h.ReadAt = &now
		h.UpdatedAt = now
	}
}

// publishEvent 发布领域事件
func (h *HealthReport) publishEvent(event DomainEvent) {
	h.events = append(h.events, event)
}

// GetEvents 获取领域事件
func (h *HealthReport) GetEvents() []DomainEvent {
	return h.events
}

// ClearEvents 清除领域事件
func (h *HealthReport) ClearEvents() {
	h.events = make([]DomainEvent, 0)
}

// publishEvent 发布领域事件
func (h *HealthAlert) publishEvent(event DomainEvent) {
	h.events = append(h.events, event)
}

// GetEvents 获取领域事件
func (h *HealthAlert) GetEvents() []DomainEvent {
	return h.events
}

// ClearEvents 清除领域事件
func (h *HealthAlert) ClearEvents() {
	h.events = make([]DomainEvent, 0)
}
