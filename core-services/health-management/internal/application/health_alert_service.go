package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthAlertService 健康预警服务
type HealthAlertService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthAlertService 创建健康预警服务
func NewHealthAlertService(
	healthDataRepo domain.HealthDataRepository,
	healthProfileRepo domain.HealthProfileRepository,
	eventPublisher EventPublisher,
) *HealthAlertService {
	return &HealthAlertService{
		healthDataRepo:    healthDataRepo,
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// AlertType 预警类型
type AlertType string

const (
	AlertTypeAbnormal   AlertType = "abnormal"   // 异常值
	AlertTypeCritical   AlertType = "critical"   // 危险值
	AlertTypeEmergency  AlertType = "emergency"  // 紧急情况
	AlertTypeTrend      AlertType = "trend"      // 趋势预警
	AlertTypeReminder   AlertType = "reminder"   // 提醒
)

// AlertSeverity 预警严重程度
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"      // 低
	AlertSeverityMedium   AlertSeverity = "medium"   // 中
	AlertSeverityHigh     AlertSeverity = "high"     // 高
	AlertSeverityCritical AlertSeverity = "critical" // 危险
)

// HealthAlert 健康预警
type HealthAlert struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Title       string        `json:"title"`
	Message     string        `json:"message"`
	DataType    string        `json:"data_type,omitempty"`
	Value       float64       `json:"value,omitempty"`
	Unit        string        `json:"unit,omitempty"`
	Threshold   float64       `json:"threshold,omitempty"`
	IsRead      bool          `json:"is_read"`
	IsHandled   bool          `json:"is_handled"`
	CreatedAt   time.Time     `json:"created_at"`
	HandledAt   *time.Time    `json:"handled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DetectAnomaliesRequest 异常检测请求
type DetectAnomaliesRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	DataTypes []string  `json:"data_types,omitempty"`
	Days      int       `json:"days,omitempty"`
}

// DetectAnomaliesResponse 异常检测响应
type DetectAnomaliesResponse struct {
	Alerts    []HealthAlert `json:"alerts"`
	Summary   string        `json:"summary"`
	Count     int           `json:"count"`
	Timestamp time.Time     `json:"timestamp"`
}

// CheckEmergencyRequest 紧急情况检查请求
type CheckEmergencyRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	DataType string    `json:"data_type" binding:"required"`
	Value    float64   `json:"value" binding:"required"`
	Unit     string    `json:"unit,omitempty"`
}

// CheckEmergencyResponse 紧急情况检查响应
type CheckEmergencyResponse struct {
	IsEmergency bool          `json:"is_emergency"`
	Alert       *HealthAlert  `json:"alert,omitempty"`
	Actions     []string      `json:"actions,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// GetAlertsRequest 获取预警请求
type GetAlertsRequest struct {
	UserID     uuid.UUID      `json:"user_id" binding:"required"`
	Types      []AlertType    `json:"types,omitempty"`
	Severities []AlertSeverity `json:"severities,omitempty"`
	IsRead     *bool          `json:"is_read,omitempty"`
	IsHandled  *bool          `json:"is_handled,omitempty"`
	Limit      int            `json:"limit,omitempty"`
	Offset     int            `json:"offset,omitempty"`
}

// GetAlertsResponse 获取预警响应
type GetAlertsResponse struct {
	Alerts []HealthAlert `json:"alerts"`
	Total  int           `json:"total"`
	Count  int           `json:"count"`
}

// DetectAnomalies 检测健康数据异常
func (s *HealthAlertService) DetectAnomalies(ctx context.Context, req *DetectAnomaliesRequest) (*DetectAnomaliesResponse, error) {
	// 设置默认值
	if req.Days == 0 {
		req.Days = 30
	}
	
	if len(req.DataTypes) == 0 {
		req.DataTypes = []string{"heart_rate", "blood_pressure", "blood_sugar", "temperature"}
	}

	var allAlerts []HealthAlert
	
	// 获取用户健康档案
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		log.Printf("Failed to get health profile for user %s: %v", req.UserID, err)
	}

	// 检测每种数据类型的异常
	for _, dataType := range req.DataTypes {
		alerts, err := s.detectDataTypeAnomalies(ctx, req.UserID, dataType, req.Days, profile)
		if err != nil {
			log.Printf("Failed to detect anomalies for data type %s: %v", dataType, err)
			continue
		}
		allAlerts = append(allAlerts, alerts...)
	}

	// 生成摘要
	summary := s.generateAnomalySummary(allAlerts)

	return &DetectAnomaliesResponse{
		Alerts:    allAlerts,
		Summary:   summary,
		Count:     len(allAlerts),
		Timestamp: time.Now(),
	}, nil
}

// CheckEmergency 检查紧急情况
func (s *HealthAlertService) CheckEmergency(ctx context.Context, req *CheckEmergencyRequest) (*CheckEmergencyResponse, error) {
	// 获取用户健康档案
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		log.Printf("Failed to get health profile for user %s: %v", req.UserID, err)
	}

	// 检查是否为紧急情况
	isEmergency, alert, actions := s.checkEmergencyCondition(req.UserID, req.DataType, req.Value, req.Unit, profile)

	response := &CheckEmergencyResponse{
		IsEmergency: isEmergency,
		Actions:     actions,
		Timestamp:   time.Now(),
	}

	if isEmergency && alert != nil {
		response.Alert = alert
		
		// 发布紧急预警事件
		if s.eventPublisher != nil {
			event := map[string]interface{}{
				"type":      "emergency_alert",
				"user_id":   req.UserID,
				"alert_id":  alert.ID,
				"data_type": req.DataType,
				"value":     req.Value,
				"severity":  alert.Severity,
			}
			s.eventPublisher.Publish("health.emergency", event)
		}
	}

	return response, nil
}

// GetAlerts 获取用户预警
func (s *HealthAlertService) GetAlerts(ctx context.Context, req *GetAlertsRequest) (*GetAlertsResponse, error) {
	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 50
	}

	// 这里应该从数据库获取预警记录
	// 由于没有预警表，我们模拟返回一些数据
	alerts := s.getMockAlerts(req.UserID, req.Limit)

	// 根据条件过滤
	filteredAlerts := s.filterAlerts(alerts, req)

	return &GetAlertsResponse{
		Alerts: filteredAlerts,
		Total:  len(filteredAlerts),
		Count:  len(filteredAlerts),
	}, nil
}

// detectDataTypeAnomalies 检测特定数据类型的异常
func (s *HealthAlertService) detectDataTypeAnomalies(ctx context.Context, userID uuid.UUID, dataType string, days int, profile *domain.HealthProfile) ([]HealthAlert, error) {
	var alerts []HealthAlert

	// 获取最近的健康数据
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	
	healthData, err := s.healthDataRepo.GetByUserIDAndType(ctx, userID, dataType, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get health data: %w", err)
	}

	if len(healthData) == 0 {
		return alerts, nil
	}

	// 获取正常范围
	normalRange := s.getNormalRange(dataType, profile)
	
	// 检测异常值
	for _, data := range healthData {
		if s.isAbnormalValue(data.Value, normalRange) {
			severity := s.calculateSeverity(data.Value, normalRange)
			alert := HealthAlert{
				ID:        uuid.New(),
				UserID:    userID,
				Type:      AlertTypeAbnormal,
				Severity:  severity,
				Title:     s.getAlertTitle(dataType, AlertTypeAbnormal),
				Message:   s.getAlertMessage(dataType, data.Value, data.Unit, normalRange),
				DataType:  dataType,
				Value:     data.Value,
				Unit:      data.Unit,
				IsRead:    false,
				IsHandled: false,
				CreatedAt: data.RecordedAt,
			}
			alerts = append(alerts, alert)
		}
	}

	// 检测趋势异常
	trendAlert := s.detectTrendAnomaly(userID, dataType, healthData, normalRange)
	if trendAlert != nil {
		alerts = append(alerts, *trendAlert)
	}

	return alerts, nil
}

// checkEmergencyCondition 检查紧急情况条件
func (s *HealthAlertService) checkEmergencyCondition(userID uuid.UUID, dataType string, value float64, unit string, profile *domain.HealthProfile) (bool, *HealthAlert, []string) {
	emergencyThresholds := s.getEmergencyThresholds(dataType, profile)
	
	isEmergency := false
	var actions []string

	switch dataType {
	case "heart_rate":
		if value < 40 || value > 180 {
			isEmergency = true
			actions = []string{
				"立即就医",
				"联系紧急联系人",
				"避免剧烈运动",
				"保持冷静",
			}
		}
	case "blood_pressure":
		// 假设value是收缩压
		if value > 180 || value < 70 {
			isEmergency = true
			actions = []string{
				"立即就医",
				"测量血压确认",
				"联系医生",
				"避免剧烈活动",
			}
		}
	case "blood_sugar":
		if value < 3.0 || value > 20.0 {
			isEmergency = true
			actions = []string{
				"立即就医",
				"检查血糖仪准确性",
				"联系医生",
				"准备急救药物",
			}
		}
	case "temperature":
		if value > 40.0 || value < 35.0 {
			isEmergency = true
			actions = []string{
				"立即就医",
				"物理降温/保温",
				"监测体温变化",
				"联系医生",
			}
		}
	}

	var alert *HealthAlert
	if isEmergency {
		alert = &HealthAlert{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      AlertTypeEmergency,
			Severity:  AlertSeverityCritical,
			Title:     s.getAlertTitle(dataType, AlertTypeEmergency),
			Message:   s.getEmergencyMessage(dataType, value, unit),
			DataType:  dataType,
			Value:     value,
			Unit:      unit,
			IsRead:    false,
			IsHandled: false,
			CreatedAt: time.Now(),
		}
	}

	return isEmergency, alert, actions
}

// getNormalRange 获取正常范围
func (s *HealthAlertService) getNormalRange(dataType string, profile *domain.HealthProfile) map[string]float64 {
	ranges := map[string]map[string]float64{
		"heart_rate": {
			"min": 60,
			"max": 100,
		},
		"blood_pressure": {
			"min": 90,
			"max": 140,
		},
		"blood_sugar": {
			"min": 4.0,
			"max": 7.0,
		},
		"temperature": {
			"min": 36.0,
			"max": 37.5,
		},
	}

	// 根据用户档案调整范围
	if profile != nil {
		age := profile.GetAge()
		if age > 65 {
			// 老年人的正常范围可能不同
			if dataType == "heart_rate" {
				ranges[dataType]["min"] = 55
				ranges[dataType]["max"] = 95
			}
		}
	}

	return ranges[dataType]
}

// isAbnormalValue 判断是否为异常值
func (s *HealthAlertService) isAbnormalValue(value float64, normalRange map[string]float64) bool {
	if normalRange == nil {
		return false
	}
	return value < normalRange["min"] || value > normalRange["max"]
}

// calculateSeverity 计算严重程度
func (s *HealthAlertService) calculateSeverity(value float64, normalRange map[string]float64) AlertSeverity {
	if normalRange == nil {
		return AlertSeverityLow
	}

	min := normalRange["min"]
	max := normalRange["max"]
	
	// 计算偏离程度
	var deviation float64
	if value < min {
		deviation = (min - value) / min
	} else {
		deviation = (value - max) / max
	}

	if deviation > 0.5 {
		return AlertSeverityCritical
	} else if deviation > 0.3 {
		return AlertSeverityHigh
	} else if deviation > 0.1 {
		return AlertSeverityMedium
	}
	return AlertSeverityLow
}

// detectTrendAnomaly 检测趋势异常
func (s *HealthAlertService) detectTrendAnomaly(userID uuid.UUID, dataType string, data []*domain.HealthData, normalRange map[string]float64) *HealthAlert {
	if len(data) < 3 {
		return nil
	}

	// 检查连续上升或下降趋势
	consecutiveAbnormal := 0
	for i := len(data) - 3; i < len(data); i++ {
		if s.isAbnormalValue(data[i].Value, normalRange) {
			consecutiveAbnormal++
		}
	}

	if consecutiveAbnormal >= 3 {
		return &HealthAlert{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      AlertTypeTrend,
			Severity:  AlertSeverityMedium,
			Title:     s.getAlertTitle(dataType, AlertTypeTrend),
			Message:   fmt.Sprintf("%s连续异常，建议关注", s.getDataTypeDisplayName(dataType)),
			DataType:  dataType,
			IsRead:    false,
			IsHandled: false,
			CreatedAt: time.Now(),
		}
	}

	return nil
}

// getEmergencyThresholds 获取紧急情况阈值
func (s *HealthAlertService) getEmergencyThresholds(dataType string, profile *domain.HealthProfile) map[string]float64 {
	thresholds := map[string]map[string]float64{
		"heart_rate": {
			"critical_min": 40,
			"critical_max": 180,
		},
		"blood_pressure": {
			"critical_min": 70,
			"critical_max": 180,
		},
		"blood_sugar": {
			"critical_min": 3.0,
			"critical_max": 20.0,
		},
		"temperature": {
			"critical_min": 35.0,
			"critical_max": 40.0,
		},
	}

	return thresholds[dataType]
}

// getAlertTitle 获取预警标题
func (s *HealthAlertService) getAlertTitle(dataType string, alertType AlertType) string {
	dataName := s.getDataTypeDisplayName(dataType)
	
	switch alertType {
	case AlertTypeAbnormal:
		return fmt.Sprintf("%s异常", dataName)
	case AlertTypeCritical:
		return fmt.Sprintf("%s危险", dataName)
	case AlertTypeEmergency:
		return fmt.Sprintf("%s紧急情况", dataName)
	case AlertTypeTrend:
		return fmt.Sprintf("%s趋势异常", dataName)
	case AlertTypeReminder:
		return fmt.Sprintf("%s提醒", dataName)
	default:
		return fmt.Sprintf("%s预警", dataName)
	}
}

// getAlertMessage 获取预警消息
func (s *HealthAlertService) getAlertMessage(dataType string, value float64, unit string, normalRange map[string]float64) string {
	dataName := s.getDataTypeDisplayName(dataType)
	
	if normalRange != nil {
		return fmt.Sprintf("%s值为%.2f%s，超出正常范围(%.2f-%.2f%s)，建议关注", 
			dataName, value, unit, normalRange["min"], normalRange["max"], unit)
	}
	
	return fmt.Sprintf("%s值为%.2f%s，检测到异常，建议关注", dataName, value, unit)
}

// getEmergencyMessage 获取紧急情况消息
func (s *HealthAlertService) getEmergencyMessage(dataType string, value float64, unit string) string {
	dataName := s.getDataTypeDisplayName(dataType)
	return fmt.Sprintf("%s值为%.2f%s，达到危险水平，请立即就医！", dataName, value, unit)
}

// getDataTypeDisplayName 获取数据类型显示名称
func (s *HealthAlertService) getDataTypeDisplayName(dataType string) string {
	names := map[string]string{
		"heart_rate":     "心率",
		"blood_pressure": "血压",
		"blood_sugar":    "血糖",
		"temperature":    "体温",
		"steps":          "步数",
		"sleep_duration": "睡眠时长",
		"stress_level":   "压力水平",
	}
	
	if name, exists := names[dataType]; exists {
		return name
	}
	return dataType
}

// generateAnomalySummary 生成异常摘要
func (s *HealthAlertService) generateAnomalySummary(alerts []HealthAlert) string {
	if len(alerts) == 0 {
		return "未检测到健康数据异常"
	}

	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, alert := range alerts {
		switch alert.Severity {
		case AlertSeverityCritical:
			criticalCount++
		case AlertSeverityHigh:
			highCount++
		case AlertSeverityMedium:
			mediumCount++
		case AlertSeverityLow:
			lowCount++
		}
	}

	summary := fmt.Sprintf("检测到%d个健康异常", len(alerts))
	if criticalCount > 0 {
		summary += fmt.Sprintf("，其中%d个危险级", criticalCount)
	}
	if highCount > 0 {
		summary += fmt.Sprintf("，%d个高风险", highCount)
	}
	if mediumCount > 0 {
		summary += fmt.Sprintf("，%d个中风险", mediumCount)
	}
	if lowCount > 0 {
		summary += fmt.Sprintf("，%d个低风险", lowCount)
	}

	return summary
}

// getMockAlerts 获取模拟预警数据
func (s *HealthAlertService) getMockAlerts(userID uuid.UUID, limit int) []HealthAlert {
	alerts := []HealthAlert{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      AlertTypeAbnormal,
			Severity:  AlertSeverityHigh,
			Title:     "血压异常",
			Message:   "血压值为150/95mmHg，超出正常范围，建议关注",
			DataType:  "blood_pressure",
			Value:     150,
			Unit:      "mmHg",
			IsRead:    false,
			IsHandled: false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      AlertTypeTrend,
			Severity:  AlertSeverityMedium,
			Title:     "心率趋势异常",
			Message:   "心率连续异常，建议关注",
			DataType:  "heart_rate",
			IsRead:    true,
			IsHandled: false,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	if len(alerts) > limit {
		return alerts[:limit]
	}
	return alerts
}

// filterAlerts 过滤预警
func (s *HealthAlertService) filterAlerts(alerts []HealthAlert, req *GetAlertsRequest) []HealthAlert {
	var filtered []HealthAlert

	for _, alert := range alerts {
		// 过滤类型
		if len(req.Types) > 0 {
			found := false
			for _, t := range req.Types {
				if alert.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 过滤严重程度
		if len(req.Severities) > 0 {
			found := false
			for _, s := range req.Severities {
				if alert.Severity == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 过滤已读状态
		if req.IsRead != nil && alert.IsRead != *req.IsRead {
			continue
		}

		// 过滤处理状态
		if req.IsHandled != nil && alert.IsHandled != *req.IsHandled {
			continue
		}

		filtered = append(filtered, alert)
	}

	// 应用分页
	start := req.Offset
	if start >= len(filtered) {
		return []HealthAlert{}
	}

	end := start + req.Limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}