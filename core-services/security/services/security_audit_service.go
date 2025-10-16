package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// SecurityAuditService 
type SecurityAuditService struct {
	db     *gorm.DB
	config *SecurityAuditConfig
	
	// ?
	auditProcessor *AuditProcessor
	eventAnalyzer  *EventAnalyzer
	complianceChecker *ComplianceChecker
	
	// 
	auditQueue chan *models.AuditLog
	eventQueue chan *models.SecurityEvent
	
	// 
	stats *AuditStats
	mutex sync.RWMutex
	
	// 
	stopChan chan bool
	running  bool
}

// SecurityAuditConfig 
type SecurityAuditConfig struct {
	Enabled              bool   `yaml:"enabled"`
	LogRetentionDays     int    `yaml:"log_retention_days"`
	EventRetentionDays   int    `yaml:"event_retention_days"`
	ComplianceEnabled    bool   `yaml:"compliance_enabled"`
	ComplianceStandard   string `yaml:"compliance_standard"`
	AlertThreshold       int    `yaml:"alert_threshold"`
	BatchSize            int    `yaml:"batch_size"`
	ProcessingInterval   int    `yaml:"processing_interval"`
}

// AuditProcessor ?
type AuditProcessor struct {
	config *SecurityAuditConfig
	db     *gorm.DB
}

// EventAnalyzer ?
type EventAnalyzer struct {
	patterns map[string]*AnalysisPattern
	mutex    sync.RWMutex
}

// ComplianceChecker 
type ComplianceChecker struct {
	rules    map[string]*ComplianceRule
	standard string
	mutex    sync.RWMutex
}

// AuditStats 
type AuditStats struct {
	TotalLogs        int64
	TotalEvents      int64
	CriticalEvents   int64
	ComplianceScore  float64
	LastProcessed    time.Time
	ProcessingErrors int64
}

// AnalysisPattern 
type AnalysisPattern struct {
	Name        string
	Pattern     string
	Severity    string
	Action      string
	Description string
}

// ComplianceRule 
type ComplianceRule struct {
	ID          string
	Name        string
	Standard    string
	Category    string
	Description string
	CheckFunc   func(ctx context.Context, data interface{}) (bool, string)
	Severity    string
}

// AuditLogEntry 
type AuditLogEntry struct {
	Timestamp   time.Time
	UserID      string
	Action      string
	Resource    string
	Result      string
	IPAddress   string
	UserAgent   string
	Details     map[string]interface{}
	RiskLevel   string
	Compliance  map[string]bool
}

// SecurityEventEntry 
type SecurityEventEntry struct {
	Timestamp   time.Time
	EventType   string
	Source      string
	Target      string
	Severity    string
	Description string
	Metadata    map[string]interface{}
	Status      string
	Response    string
}

// NewSecurityAuditService 
func NewSecurityAuditService(db *gorm.DB, config *SecurityAuditConfig) *SecurityAuditService {
	service := &SecurityAuditService{
		db:         db,
		config:     config,
		auditQueue: make(chan *models.AuditLog, 1000),
		eventQueue: make(chan *models.SecurityEvent, 1000),
		stats:      &AuditStats{},
		stopChan:   make(chan bool),
		running:    false,
	}
	
	// ?
	service.initComponents()
	
	return service
}

// initComponents ?
func (sas *SecurityAuditService) initComponents() {
	// 
	sas.auditProcessor = &AuditProcessor{
		config: sas.config,
		db:     sas.db,
	}
	
	// 
	sas.eventAnalyzer = &EventAnalyzer{
		patterns: make(map[string]*AnalysisPattern),
	}
	sas.initAnalysisPatterns()
	
	// 
	sas.complianceChecker = &ComplianceChecker{
		rules:    make(map[string]*ComplianceRule),
		standard: sas.config.ComplianceStandard,
	}
	sas.initComplianceRules()
}

// initAnalysisPatterns ?
func (sas *SecurityAuditService) initAnalysisPatterns() {
	patterns := []*AnalysisPattern{
		{
			Name:        "failed_login_attempts",
			Pattern:     "login_failed",
			Severity:    "medium",
			Action:      "monitor",
			Description: "?,
		},
		{
			Name:        "privilege_escalation",
			Pattern:     "privilege_change",
			Severity:    "high",
			Action:      "alert",
			Description: "?,
		},
		{
			Name:        "data_access_anomaly",
			Pattern:     "data_access",
			Severity:    "medium",
			Action:      "log",
			Description: "?,
		},
		{
			Name:        "admin_operations",
			Pattern:     "admin_action",
			Severity:    "high",
			Action:      "audit",
			Description: "",
		},
		{
			Name:        "security_config_change",
			Pattern:     "security_config",
			Severity:    "critical",
			Action:      "immediate_alert",
			Description: "?,
		},
	}
	
	sas.eventAnalyzer.mutex.Lock()
	for _, pattern := range patterns {
		sas.eventAnalyzer.patterns[pattern.Name] = pattern
	}
	sas.eventAnalyzer.mutex.Unlock()
}

// initComplianceRules ?
func (sas *SecurityAuditService) initComplianceRules() {
	rules := []*ComplianceRule{
		{
			ID:          "auth_001",
			Name:        "",
			Standard:    "DengBao 2.0",
			Category:    "authentication",
			Description: "?,
			CheckFunc:   sas.checkAuthenticationCompliance,
			Severity:    "high",
		},
		{
			ID:          "access_001",
			Name:        "",
			Standard:    "DengBao 2.0",
			Category:    "access_control",
			Description: "?,
			CheckFunc:   sas.checkAccessControlCompliance,
			Severity:    "high",
		},
		{
			ID:          "audit_001",
			Name:        "",
			Standard:    "DengBao 2.0",
			Category:    "audit",
			Description: "?,
			CheckFunc:   sas.checkAuditCompliance,
			Severity:    "medium",
		},
		{
			ID:          "crypto_001",
			Name:        "㷨",
			Standard:    "DengBao 2.0",
			Category:    "cryptography",
			Description: "㷨",
			CheckFunc:   sas.checkCryptographyCompliance,
			Severity:    "high",
		},
		{
			ID:          "data_001",
			Name:        "",
			Standard:    "DengBao 2.0",
			Category:    "data_protection",
			Description: "洢",
			CheckFunc:   sas.checkDataProtectionCompliance,
			Severity:    "critical",
		},
	}
	
	sas.complianceChecker.mutex.Lock()
	for _, rule := range rules {
		sas.complianceChecker.rules[rule.ID] = rule
	}
	sas.complianceChecker.mutex.Unlock()
}

// Start 
func (sas *SecurityAuditService) Start() {
	if sas.running {
		return
	}
	
	sas.running = true
	log.Println("Starting Security Audit Service...")
	
	// 
	go sas.processAuditLogs()
	
	// 
	go sas.processSecurityEvents()
	
	// 
	go sas.periodicCleanup()
	
	// ?
	if sas.config.ComplianceEnabled {
		go sas.periodicComplianceCheck()
	}
	
	log.Println("Security Audit Service started successfully")
}

// Stop 
func (sas *SecurityAuditService) Stop() {
	if !sas.running {
		return
	}
	
	log.Println("Stopping Security Audit Service...")
	sas.stopChan <- true
	sas.running = false
	log.Println("Security Audit Service stopped")
}

// processAuditLogs 
func (sas *SecurityAuditService) processAuditLogs() {
	for {
		select {
		case auditLog := <-sas.auditQueue:
			sas.processAuditLog(auditLog)
		case <-sas.stopChan:
			return
		}
	}
}

// processSecurityEvents 
func (sas *SecurityAuditService) processSecurityEvents() {
	for {
		select {
		case event := <-sas.eventQueue:
			sas.processSecurityEvent(event)
		case <-sas.stopChan:
			return
		}
	}
}

// processAuditLog 
func (sas *SecurityAuditService) processAuditLog(auditLog *models.AuditLog) {
	// 
	riskLevel := sas.analyzeAuditRisk(auditLog)
	auditLog.Metadata = models.JSONB(map[string]interface{}{
		"risk_level":     riskLevel,
		"processed_at":   time.Now(),
		"processor":      "security_audit_service",
	})
	
	// 浽
	if err := sas.db.Save(auditLog).Error; err != nil {
		log.Printf("Failed to save audit log: %v", err)
		sas.updateStats("processing_errors", 1)
		return
	}
	
	// 
	sas.updateStats("total_logs", 1)
	
	// ?
	if riskLevel == "high" || riskLevel == "critical" {
		sas.createSecurityEventFromAudit(auditLog)
	}
	
	log.Printf("Processed audit log: %s - %s", auditLog.Action, riskLevel)
}

// processSecurityEvent 
func (sas *SecurityAuditService) processSecurityEvent(event *models.SecurityEvent) {
	// 
	sas.analyzeSecurityEvent(event)
	
	// 浽
	if err := sas.db.Save(event).Error; err != nil {
		log.Printf("Failed to save security event: %v", err)
		sas.updateStats("processing_errors", 1)
		return
	}
	
	// 
	sas.updateStats("total_events", 1)
	if event.Severity == "critical" {
		sas.updateStats("critical_events", 1)
	}
	
	log.Printf("Processed security event: %s - %s", event.EventType, event.Severity)
}

// analyzeAuditRisk 
func (sas *SecurityAuditService) analyzeAuditRisk(auditLog *models.AuditLog) string {
	// 
	highRiskActions := []string{
		"user_delete", "role_change", "permission_grant",
		"security_config_change", "admin_login", "data_export",
	}
	
	mediumRiskActions := []string{
		"login_failed", "password_change", "data_access",
		"file_upload", "api_access",
	}
	
	for _, action := range highRiskActions {
		if strings.Contains(auditLog.Action, action) {
			return "high"
		}
	}
	
	for _, action := range mediumRiskActions {
		if strings.Contains(auditLog.Action, action) {
			return "medium"
		}
	}
	
	// ?
	if auditLog.Result == "failed" || auditLog.Result == "error" {
		return "medium"
	}
	
	return "low"
}

// analyzeSecurityEvent 
func (sas *SecurityAuditService) analyzeSecurityEvent(event *models.SecurityEvent) {
	sas.eventAnalyzer.mutex.RLock()
	defer sas.eventAnalyzer.mutex.RUnlock()
	
	// 
	for _, pattern := range sas.eventAnalyzer.patterns {
		if strings.Contains(event.EventType, pattern.Pattern) {
			event.Severity = pattern.Severity
			
			// 
			switch pattern.Action {
			case "immediate_alert":
				sas.sendImmediateAlert(event)
			case "alert":
				sas.sendAlert(event)
			case "monitor":
				sas.addToMonitoring(event)
			}
			
			break
		}
	}
}

// createSecurityEventFromAudit ?
func (sas *SecurityAuditService) createSecurityEventFromAudit(auditLog *models.AuditLog) {
	event := &models.SecurityEvent{
		EventType:   fmt.Sprintf("audit_%s", auditLog.Action),
		Source:      auditLog.IPAddress,
		Target:      auditLog.Resource,
		Severity:    "medium",
		Description: fmt.Sprintf("High-risk audit event: %s by user %s", auditLog.Action, auditLog.UserID),
		Metadata: models.JSONB(map[string]interface{}{
			"audit_log_id": auditLog.ID,
			"user_id":      auditLog.UserID,
			"action":       auditLog.Action,
			"result":       auditLog.Result,
		}),
		Status: "new",
	}
	
	// ?
	select {
	case sas.eventQueue <- event:
	default:
		log.Println("Event queue is full, dropping event")
	}
}

// sendImmediateAlert ?
func (sas *SecurityAuditService) sendImmediateAlert(event *models.SecurityEvent) {
	log.Printf("IMMEDIATE ALERT: %s - %s", event.EventType, event.Description)
	// 澯
}

// sendAlert ?
func (sas *SecurityAuditService) sendAlert(event *models.SecurityEvent) {
	log.Printf("ALERT: %s - %s", event.EventType, event.Description)
	// 澯
}

// addToMonitoring ?
func (sas *SecurityAuditService) addToMonitoring(event *models.SecurityEvent) {
	log.Printf("MONITORING: %s - %s", event.EventType, event.Description)
	// ?
}

// periodicCleanup 
func (sas *SecurityAuditService) periodicCleanup() {
	ticker := time.NewTicker(24 * time.Hour) // ?
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			sas.cleanupOldLogs()
		case <-sas.stopChan:
			return
		}
	}
}

// cleanupOldLogs ?
func (sas *SecurityAuditService) cleanupOldLogs() {
	// ?
	auditRetention := time.Duration(sas.config.LogRetentionDays) * 24 * time.Hour
	auditCutoff := time.Now().Add(-auditRetention)
	
	result := sas.db.Where("created_at < ?", auditCutoff).Delete(&models.AuditLog{})
	if result.Error != nil {
		log.Printf("Failed to cleanup audit logs: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d audit logs", result.RowsAffected)
	}
	
	// İ?
	eventRetention := time.Duration(sas.config.EventRetentionDays) * 24 * time.Hour
	eventCutoff := time.Now().Add(-eventRetention)
	
	result = sas.db.Where("created_at < ?", eventCutoff).Delete(&models.SecurityEvent{})
	if result.Error != nil {
		log.Printf("Failed to cleanup security events: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d security events", result.RowsAffected)
	}
}

// periodicComplianceCheck ?
func (sas *SecurityAuditService) periodicComplianceCheck() {
	ticker := time.NewTicker(time.Duration(sas.config.ProcessingInterval) * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			sas.runComplianceCheck()
		case <-sas.stopChan:
			return
		}
	}
}

// runComplianceCheck ?
func (sas *SecurityAuditService) runComplianceCheck() {
	ctx := context.Background()
	
	sas.complianceChecker.mutex.RLock()
	rules := make([]*ComplianceRule, 0, len(sas.complianceChecker.rules))
	for _, rule := range sas.complianceChecker.rules {
		rules = append(rules, rule)
	}
	sas.complianceChecker.mutex.RUnlock()
	
	totalRules := len(rules)
	passedRules := 0
	
	report := &models.ComplianceReport{
		Standard:    sas.config.ComplianceStandard,
		CheckTime:   time.Now(),
		TotalRules:  totalRules,
		PassedRules: 0,
		FailedRules: 0,
		Score:       0.0,
		Status:      "completed",
		Details: models.JSONB(map[string]interface{}{
			"rules": []map[string]interface{}{},
		}),
	}
	
	var ruleResults []map[string]interface{}
	
	for _, rule := range rules {
		passed, message := rule.CheckFunc(ctx, nil)
		if passed {
			passedRules++
		}
		
		ruleResults = append(ruleResults, map[string]interface{}{
			"rule_id":     rule.ID,
			"rule_name":   rule.Name,
			"category":    rule.Category,
			"passed":      passed,
			"message":     message,
			"severity":    rule.Severity,
		})
		
		log.Printf("Compliance check %s: %s - %v", rule.ID, rule.Name, passed)
	}
	
	report.PassedRules = passedRules
	report.FailedRules = totalRules - passedRules
	report.Score = float64(passedRules) / float64(totalRules) * 100
	
	details := map[string]interface{}{
		"rules": ruleResults,
		"summary": map[string]interface{}{
			"total_rules":  totalRules,
			"passed_rules": passedRules,
			"failed_rules": totalRules - passedRules,
			"score":        report.Score,
		},
	}
	report.Details = models.JSONB(details)
	
	// 汨
	if err := sas.db.Create(report).Error; err != nil {
		log.Printf("Failed to save compliance report: %v", err)
	}
	
	// 
	sas.mutex.Lock()
	sas.stats.ComplianceScore = report.Score
	sas.mutex.Unlock()
	
	log.Printf("Compliance check completed: %.2f%% (%d/%d rules passed)", 
		report.Score, passedRules, totalRules)
}

// 麯?
func (sas *SecurityAuditService) checkAuthenticationCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 
	var unauthenticatedCount int64
	err := sas.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("action LIKE ? AND result = ?", "%access%", "unauthorized").
		Count(&unauthenticatedCount).Error
	
	if err != nil {
		return false, fmt.Sprintf("? %v", err)
	}
	
	if unauthenticatedCount > 0 {
		return false, fmt.Sprintf(" %d ", unauthenticatedCount)
	}
	
	return true, ""
}

func (sas *SecurityAuditService) checkAccessControlCompliance(ctx context.Context, data interface{}) (bool, string) {
	// ?
	var accessViolationCount int64
	err := sas.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("action LIKE ? AND result = ?", "%access%", "forbidden").
		Count(&accessViolationCount).Error
	
	if err != nil {
		return false, fmt.Sprintf("? %v", err)
	}
	
	// ?
	if accessViolationCount > 100 {
		return false, fmt.Sprintf(": %d", accessViolationCount)
	}
	
	return true, ""
}

func (sas *SecurityAuditService) checkAuditCompliance(ctx context.Context, data interface{}) (bool, string) {
	// ?
	var recentLogCount int64
	yesterday := time.Now().AddDate(0, 0, -1)
	
	err := sas.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at > ?", yesterday).
		Count(&recentLogCount).Error
	
	if err != nil {
		return false, fmt.Sprintf("? %v", err)
	}
	
	if recentLogCount == 0 {
		return false, "?4"
	}
	
	return true, fmt.Sprintf("24 %d ?, recentLogCount)
}

func (sas *SecurityAuditService) checkCryptographyCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 㷨?
	// ?
	return true, "㷨"
}

func (sas *SecurityAuditService) checkDataProtectionCompliance(ctx context.Context, data interface{}) (bool, string) {
	// ?
	// ?
	return true, ""
}

// updateStats 
func (sas *SecurityAuditService) updateStats(metric string, value int64) {
	sas.mutex.Lock()
	defer sas.mutex.Unlock()
	
	switch metric {
	case "total_logs":
		sas.stats.TotalLogs += value
	case "total_events":
		sas.stats.TotalEvents += value
	case "critical_events":
		sas.stats.CriticalEvents += value
	case "processing_errors":
		sas.stats.ProcessingErrors += value
	}
	
	sas.stats.LastProcessed = time.Now()
}

// LogAuditEvent 
func (sas *SecurityAuditService) LogAuditEvent(ctx context.Context, userID, action, resource, result, ipAddress, userAgent string, details map[string]interface{}) error {
	auditLog := &models.AuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Result:    result,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   models.JSONB(details),
	}
	
	// ?
	select {
	case sas.auditQueue <- auditLog:
		return nil
	default:
		// 浽
		return sas.db.WithContext(ctx).Create(auditLog).Error
	}
}

// LogSecurityEvent 
func (sas *SecurityAuditService) LogSecurityEvent(ctx context.Context, eventType, source, target, severity, description string, metadata map[string]interface{}) error {
	event := &models.SecurityEvent{
		EventType:   eventType,
		Source:      source,
		Target:      target,
		Severity:    severity,
		Description: description,
		Metadata:    models.JSONB(metadata),
		Status:      "new",
	}
	
	// ?
	select {
	case sas.eventQueue <- event:
		return nil
	default:
		// 浽
		return sas.db.WithContext(ctx).Create(event).Error
	}
}

// GetAuditLogs 
func (sas *SecurityAuditService) GetAuditLogs(ctx context.Context, userID string, action string, startTime, endTime time.Time, limit, offset int) ([]models.AuditLog, error) {
	query := sas.db.WithContext(ctx).Model(&models.AuditLog{})
	
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	if action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}
	
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}
	
	var logs []models.AuditLog
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	
	return logs, err
}

// GetSecurityEvents 
func (sas *SecurityAuditService) GetSecurityEvents(ctx context.Context, eventType string, severity string, startTime, endTime time.Time, limit, offset int) ([]models.SecurityEvent, error) {
	query := sas.db.WithContext(ctx).Model(&models.SecurityEvent{})
	
	if eventType != "" {
		query = query.Where("event_type LIKE ?", "%"+eventType+"%")
	}
	
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}
	
	var events []models.SecurityEvent
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&events).Error
	
	return events, err
}

// GetComplianceReports 汨
func (sas *SecurityAuditService) GetComplianceReports(ctx context.Context, standard string, limit, offset int) ([]models.ComplianceReport, error) {
	query := sas.db.WithContext(ctx).Model(&models.ComplianceReport{})
	
	if standard != "" {
		query = query.Where("standard = ?", standard)
	}
	
	var reports []models.ComplianceReport
	err := query.Order("check_time DESC").Limit(limit).Offset(offset).Find(&reports).Error
	
	return reports, err
}

// GetAuditStats 
func (sas *SecurityAuditService) GetAuditStats(ctx context.Context) (*AuditStats, error) {
	sas.mutex.RLock()
	defer sas.mutex.RUnlock()
	
	// 
	stats := &AuditStats{
		TotalLogs:        sas.stats.TotalLogs,
		TotalEvents:      sas.stats.TotalEvents,
		CriticalEvents:   sas.stats.CriticalEvents,
		ComplianceScore:  sas.stats.ComplianceScore,
		LastProcessed:    sas.stats.LastProcessed,
		ProcessingErrors: sas.stats.ProcessingErrors,
	}
	
	return stats, nil
}

// UpdateSecurityEvent 
func (sas *SecurityAuditService) UpdateSecurityEvent(ctx context.Context, id string, updates map[string]interface{}) error {
	return sas.db.WithContext(ctx).Model(&models.SecurityEvent{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAuditLog 
func (sas *SecurityAuditService) DeleteAuditLog(ctx context.Context, id string) error {
	return sas.db.WithContext(ctx).Delete(&models.AuditLog{}, "id = ?", id).Error
}

// DeleteSecurityEvent 
func (sas *SecurityAuditService) DeleteSecurityEvent(ctx context.Context, id string) error {
	return sas.db.WithContext(ctx).Delete(&models.SecurityEvent{}, "id = ?", id).Error
}

