package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthAlertService 
type HealthAlertService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthAlertService 
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

// AlertType 
type AlertType string

const (
	AlertTypeAbnormal   AlertType = "abnormal"   // ?
	AlertTypeCritical   AlertType = "critical"   // ?
	AlertTypeEmergency  AlertType = "emergency"  // ?
	AlertTypeTrend      AlertType = "trend"      // 
	AlertTypeReminder   AlertType = "reminder"   // 
)

// AlertSeverity 
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"      // ?
	AlertSeverityMedium   AlertSeverity = "medium"   // ?
	AlertSeverityHigh     AlertSeverity = "high"     // ?
	AlertSeverityCritical AlertSeverity = "critical" // 
)

// HealthAlert 
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

// DetectAnomaliesRequest ?
type DetectAnomaliesRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	DataTypes []string  `json:"data_types,omitempty"`
	Days      int       `json:"days,omitempty"`
}

// DetectAnomaliesResponse ?
type DetectAnomaliesResponse struct {
	Alerts    []HealthAlert `json:"alerts"`
	Summary   string        `json:"summary"`
	Count     int           `json:"count"`
	Timestamp time.Time     `json:"timestamp"`
}

// CheckEmergencyRequest ?
type CheckEmergencyRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	DataType string    `json:"data_type" binding:"required"`
	Value    float64   `json:"value" binding:"required"`
	Unit     string    `json:"unit,omitempty"`
}

// CheckEmergencyResponse ?
type CheckEmergencyResponse struct {
	IsEmergency bool          `json:"is_emergency"`
	Alert       *HealthAlert  `json:"alert,omitempty"`
	Actions     []string      `json:"actions,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// GetAlertsRequest 
type GetAlertsRequest struct {
	UserID     uuid.UUID      `json:"user_id" binding:"required"`
	Types      []AlertType    `json:"types,omitempty"`
	Severities []AlertSeverity `json:"severities,omitempty"`
	IsRead     *bool          `json:"is_read,omitempty"`
	IsHandled  *bool          `json:"is_handled,omitempty"`
	Limit      int            `json:"limit,omitempty"`
	Offset     int            `json:"offset,omitempty"`
}

// GetAlertsResponse 
type GetAlertsResponse struct {
	Alerts []HealthAlert `json:"alerts"`
	Total  int           `json:"total"`
	Count  int           `json:"count"`
}

// DetectAnomalies ?
func (s *HealthAlertService) DetectAnomalies(ctx context.Context, req *DetectAnomaliesRequest) (*DetectAnomaliesResponse, error) {
	// ?
	if req.Days == 0 {
		req.Days = 30
	}
	
	if len(req.DataTypes) == 0 {
		req.DataTypes = []string{"heart_rate", "blood_pressure", "blood_sugar", "temperature"}
	}

	var allAlerts []HealthAlert
	
	// 
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		log.Printf("Failed to get health profile for user %s: %v", req.UserID, err)
	}

	// 
	for _, dataType := range req.DataTypes {
		alerts, err := s.detectDataTypeAnomalies(ctx, req.UserID, dataType, req.Days, profile)
		if err != nil {
			log.Printf("Failed to detect anomalies for data type %s: %v", dataType, err)
			continue
		}
		allAlerts = append(allAlerts, alerts...)
	}

	// 
	summary := s.generateAnomalySummary(allAlerts)

	return &DetectAnomaliesResponse{
		Alerts:    allAlerts,
		Summary:   summary,
		Count:     len(allAlerts),
		Timestamp: time.Now(),
	}, nil
}

// CheckEmergency ?
func (s *HealthAlertService) CheckEmergency(ctx context.Context, req *CheckEmergencyRequest) (*CheckEmergencyResponse, error) {
	// 
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		log.Printf("Failed to get health profile for user %s: %v", req.UserID, err)
	}

	// ?
	isEmergency, alert, actions := s.checkEmergencyCondition(req.UserID, req.DataType, req.Value, req.Unit, profile)

	response := &CheckEmergencyResponse{
		IsEmergency: isEmergency,
		Actions:     actions,
		Timestamp:   time.Now(),
	}

	if isEmergency && alert != nil {
		response.Alert = alert
		
		// ?
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

// GetAlerts 
func (s *HealthAlertService) GetAlerts(ctx context.Context, req *GetAlertsRequest) (*GetAlertsResponse, error) {
	// ?
	if req.Limit == 0 {
		req.Limit = 50
	}

	// 
	// ?
	alerts := s.getMockAlerts(req.UserID, req.Limit)

	// 
	filteredAlerts := s.filterAlerts(alerts, req)

	return &GetAlertsResponse{
		Alerts: filteredAlerts,
		Total:  len(filteredAlerts),
		Count:  len(filteredAlerts),
	}, nil
}

// detectDataTypeAnomalies 
func (s *HealthAlertService) detectDataTypeAnomalies(ctx context.Context, userID uuid.UUID, dataType string, days int, profile *domain.HealthProfile) ([]HealthAlert, error) {
	var alerts []HealthAlert

	// 
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	
	healthData, err := s.healthDataRepo.GetByUserIDAndType(ctx, userID, dataType, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get health data: %w", err)
	}

	if len(healthData) == 0 {
		return alerts, nil
	}

	// 
	normalRange := s.getNormalRange(dataType, profile)
	
	// ?
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

	// ?
	trendAlert := s.detectTrendAnomaly(userID, dataType, healthData, normalRange)
	if trendAlert != nil {
		alerts = append(alerts, *trendAlert)
	}

	return alerts, nil
}

// checkEmergencyCondition ?
func (s *HealthAlertService) checkEmergencyCondition(userID uuid.UUID, dataType string, value float64, unit string, profile *domain.HealthProfile) (bool, *HealthAlert, []string) {
	emergencyThresholds := s.getEmergencyThresholds(dataType, profile)
	
	isEmergency := false
	var actions []string

	switch dataType {
	case "heart_rate":
		if value < 40 || value > 180 {
			isEmergency = true
			actions = []string{
				"",
				"",
				"",
				"侲",
			}
		}
	case "blood_pressure":
		// value
		if value > 180 || value < 70 {
			isEmergency = true
			actions = []string{
				"",
				"?,
				"",
				"",
			}
		}
	case "blood_sugar":
		if value < 3.0 || value > 20.0 {
			isEmergency = true
			actions = []string{
				"",
				"?,
				"",
				"",
			}
		}
	case "temperature":
		if value > 40.0 || value < 35.0 {
			isEmergency = true
			actions = []string{
				"",
				"/",
				"仯",
				"",
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

// getNormalRange 
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

	// 
	if profile != nil {
		age := profile.GetAge()
		if age > 65 {
			// 
			if dataType == "heart_rate" {
				ranges[dataType]["min"] = 55
				ranges[dataType]["max"] = 95
			}
		}
	}

	return ranges[dataType]
}

// isAbnormalValue ?
func (s *HealthAlertService) isAbnormalValue(value float64, normalRange map[string]float64) bool {
	if normalRange == nil {
		return false
	}
	return value < normalRange["min"] || value > normalRange["max"]
}

// calculateSeverity 
func (s *HealthAlertService) calculateSeverity(value float64, normalRange map[string]float64) AlertSeverity {
	if normalRange == nil {
		return AlertSeverityLow
	}

	min := normalRange["min"]
	max := normalRange["max"]
	
	// 
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

// detectTrendAnomaly ?
func (s *HealthAlertService) detectTrendAnomaly(userID uuid.UUID, dataType string, data []*domain.HealthData, normalRange map[string]float64) *HealthAlert {
	if len(data) < 3 {
		return nil
	}

	// 
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
			Message:   fmt.Sprintf("%s?, s.getDataTypeDisplayName(dataType)),
			DataType:  dataType,
			IsRead:    false,
			IsHandled: false,
			CreatedAt: time.Now(),
		}
	}

	return nil
}

// getEmergencyThresholds ?
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

// getAlertTitle 
func (s *HealthAlertService) getAlertTitle(dataType string, alertType AlertType) string {
	dataName := s.getDataTypeDisplayName(dataType)
	
	switch alertType {
	case AlertTypeAbnormal:
		return fmt.Sprintf("%s", dataName)
	case AlertTypeCritical:
		return fmt.Sprintf("%s", dataName)
	case AlertTypeEmergency:
		return fmt.Sprintf("%s?, dataName)
	case AlertTypeTrend:
		return fmt.Sprintf("%s", dataName)
	case AlertTypeReminder:
		return fmt.Sprintf("%s", dataName)
	default:
		return fmt.Sprintf("%s", dataName)
	}
}

// getAlertMessage 
func (s *HealthAlertService) getAlertMessage(dataType string, value float64, unit string, normalRange map[string]float64) string {
	dataName := s.getDataTypeDisplayName(dataType)
	
	if normalRange != nil {
		return fmt.Sprintf("%s%.2f%s?%.2f-%.2f%s)?, 
			dataName, value, unit, normalRange["min"], normalRange["max"], unit)
	}
	
	return fmt.Sprintf("%s%.2f%s?, dataName, value, unit)
}

// getEmergencyMessage ?
func (s *HealthAlertService) getEmergencyMessage(dataType string, value float64, unit string) string {
	dataName := s.getDataTypeDisplayName(dataType)
	return fmt.Sprintf("%s%.2f%s", dataName, value, unit)
}

// getDataTypeDisplayName 
func (s *HealthAlertService) getDataTypeDisplayName(dataType string) string {
	names := map[string]string{
		"heart_rate":     "",
		"blood_pressure": "?,
		"blood_sugar":    "?,
		"temperature":    "",
		"steps":          "",
		"sleep_duration": "",
		"stress_level":   "",
	}
	
	if name, exists := names[dataType]; exists {
		return name
	}
	return dataType
}

// generateAnomalySummary 
func (s *HealthAlertService) generateAnomalySummary(alerts []HealthAlert) string {
	if len(alerts) == 0 {
		return ""
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

	summary := fmt.Sprintf("%d?, len(alerts))
	if criticalCount > 0 {
		summary += fmt.Sprintf("?d", criticalCount)
	}
	if highCount > 0 {
		summary += fmt.Sprintf("?d", highCount)
	}
	if mediumCount > 0 {
		summary += fmt.Sprintf("?d", mediumCount)
	}
	if lowCount > 0 {
		summary += fmt.Sprintf("?d", lowCount)
	}

	return summary
}

// getMockAlerts 
func (s *HealthAlertService) getMockAlerts(userID uuid.UUID, limit int) []HealthAlert {
	alerts := []HealthAlert{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      AlertTypeAbnormal,
			Severity:  AlertSeverityHigh,
			Title:     "?,
			Message:   "150/95mmHg",
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
			Title:     "",
			Message:   "?,
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

// filterAlerts 
func (s *HealthAlertService) filterAlerts(alerts []HealthAlert, req *GetAlertsRequest) []HealthAlert {
	var filtered []HealthAlert

	for _, alert := range alerts {
		// 
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

		// 
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

		// ?
		if req.IsRead != nil && alert.IsRead != *req.IsRead {
			continue
		}

		// ?
		if req.IsHandled != nil && alert.IsHandled != *req.IsHandled {
			continue
		}

		filtered = append(filtered, alert)
	}

	// 
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

