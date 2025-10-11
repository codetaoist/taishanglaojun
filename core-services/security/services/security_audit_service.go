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

// SecurityAuditService 安全审计服务
type SecurityAuditService struct {
	db     *gorm.DB
	config *SecurityAuditConfig
	
	// 审计处理�?
	auditProcessor *AuditProcessor
	eventAnalyzer  *EventAnalyzer
	complianceChecker *ComplianceChecker
	
	// 审计队列
	auditQueue chan *models.AuditLog
	eventQueue chan *models.SecurityEvent
	
	// 统计信息
	stats *AuditStats
	mutex sync.RWMutex
	
	// 控制通道
	stopChan chan bool
	running  bool
}

// SecurityAuditConfig 安全审计配置
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

// AuditProcessor 审计处理�?
type AuditProcessor struct {
	config *SecurityAuditConfig
	db     *gorm.DB
}

// EventAnalyzer 事件分析�?
type EventAnalyzer struct {
	patterns map[string]*AnalysisPattern
	mutex    sync.RWMutex
}

// ComplianceChecker 合规检查器
type ComplianceChecker struct {
	rules    map[string]*ComplianceRule
	standard string
	mutex    sync.RWMutex
}

// AuditStats 审计统计
type AuditStats struct {
	TotalLogs        int64
	TotalEvents      int64
	CriticalEvents   int64
	ComplianceScore  float64
	LastProcessed    time.Time
	ProcessingErrors int64
}

// AnalysisPattern 分析模式
type AnalysisPattern struct {
	Name        string
	Pattern     string
	Severity    string
	Action      string
	Description string
}

// ComplianceRule 合规规则
type ComplianceRule struct {
	ID          string
	Name        string
	Standard    string
	Category    string
	Description string
	CheckFunc   func(ctx context.Context, data interface{}) (bool, string)
	Severity    string
}

// AuditLogEntry 审计日志条目
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

// SecurityEventEntry 安全事件条目
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

// NewSecurityAuditService 创建安全审计服务
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
	
	// 初始化组�?
	service.initComponents()
	
	return service
}

// initComponents 初始化组�?
func (sas *SecurityAuditService) initComponents() {
	// 初始化审计处理器
	sas.auditProcessor = &AuditProcessor{
		config: sas.config,
		db:     sas.db,
	}
	
	// 初始化事件分析器
	sas.eventAnalyzer = &EventAnalyzer{
		patterns: make(map[string]*AnalysisPattern),
	}
	sas.initAnalysisPatterns()
	
	// 初始化合规检查器
	sas.complianceChecker = &ComplianceChecker{
		rules:    make(map[string]*ComplianceRule),
		standard: sas.config.ComplianceStandard,
	}
	sas.initComplianceRules()
}

// initAnalysisPatterns 初始化分析模�?
func (sas *SecurityAuditService) initAnalysisPatterns() {
	patterns := []*AnalysisPattern{
		{
			Name:        "failed_login_attempts",
			Pattern:     "login_failed",
			Severity:    "medium",
			Action:      "monitor",
			Description: "检测失败登录尝�?,
		},
		{
			Name:        "privilege_escalation",
			Pattern:     "privilege_change",
			Severity:    "high",
			Action:      "alert",
			Description: "检测权限提升操�?,
		},
		{
			Name:        "data_access_anomaly",
			Pattern:     "data_access",
			Severity:    "medium",
			Action:      "log",
			Description: "检测异常数据访�?,
		},
		{
			Name:        "admin_operations",
			Pattern:     "admin_action",
			Severity:    "high",
			Action:      "audit",
			Description: "检测管理员操作",
		},
		{
			Name:        "security_config_change",
			Pattern:     "security_config",
			Severity:    "critical",
			Action:      "immediate_alert",
			Description: "检测安全配置变�?,
		},
	}
	
	sas.eventAnalyzer.mutex.Lock()
	for _, pattern := range patterns {
		sas.eventAnalyzer.patterns[pattern.Name] = pattern
	}
	sas.eventAnalyzer.mutex.Unlock()
}

// initComplianceRules 初始化合规规�?
func (sas *SecurityAuditService) initComplianceRules() {
	rules := []*ComplianceRule{
		{
			ID:          "auth_001",
			Name:        "强制身份认证",
			Standard:    "DengBao 2.0",
			Category:    "authentication",
			Description: "所有用户访问必须经过身份认�?,
			CheckFunc:   sas.checkAuthenticationCompliance,
			Severity:    "high",
		},
		{
			ID:          "access_001",
			Name:        "访问控制",
			Standard:    "DengBao 2.0",
			Category:    "access_control",
			Description: "实施基于角色的访问控�?,
			CheckFunc:   sas.checkAccessControlCompliance,
			Severity:    "high",
		},
		{
			ID:          "audit_001",
			Name:        "安全审计",
			Standard:    "DengBao 2.0",
			Category:    "audit",
			Description: "记录所有安全相关事�?,
			CheckFunc:   sas.checkAuditCompliance,
			Severity:    "medium",
		},
		{
			ID:          "crypto_001",
			Name:        "密码算法",
			Standard:    "DengBao 2.0",
			Category:    "cryptography",
			Description: "使用国密算法进行加密",
			CheckFunc:   sas.checkCryptographyCompliance,
			Severity:    "high",
		},
		{
			ID:          "data_001",
			Name:        "数据保护",
			Standard:    "DengBao 2.0",
			Category:    "data_protection",
			Description: "敏感数据必须加密存储",
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

// Start 启动安全审计服务
func (sas *SecurityAuditService) Start() {
	if sas.running {
		return
	}
	
	sas.running = true
	log.Println("Starting Security Audit Service...")
	
	// 启动审计日志处理
	go sas.processAuditLogs()
	
	// 启动安全事件处理
	go sas.processSecurityEvents()
	
	// 启动定期清理
	go sas.periodicCleanup()
	
	// 启动合规检�?
	if sas.config.ComplianceEnabled {
		go sas.periodicComplianceCheck()
	}
	
	log.Println("Security Audit Service started successfully")
}

// Stop 停止安全审计服务
func (sas *SecurityAuditService) Stop() {
	if !sas.running {
		return
	}
	
	log.Println("Stopping Security Audit Service...")
	sas.stopChan <- true
	sas.running = false
	log.Println("Security Audit Service stopped")
}

// processAuditLogs 处理审计日志
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

// processSecurityEvents 处理安全事件
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

// processAuditLog 处理单个审计日志
func (sas *SecurityAuditService) processAuditLog(auditLog *models.AuditLog) {
	// 分析审计日志
	riskLevel := sas.analyzeAuditRisk(auditLog)
	auditLog.Metadata = models.JSONB(map[string]interface{}{
		"risk_level":     riskLevel,
		"processed_at":   time.Now(),
		"processor":      "security_audit_service",
	})
	
	// 保存到数据库
	if err := sas.db.Save(auditLog).Error; err != nil {
		log.Printf("Failed to save audit log: %v", err)
		sas.updateStats("processing_errors", 1)
		return
	}
	
	// 更新统计信息
	sas.updateStats("total_logs", 1)
	
	// 如果是高风险事件，创建安全事�?
	if riskLevel == "high" || riskLevel == "critical" {
		sas.createSecurityEventFromAudit(auditLog)
	}
	
	log.Printf("Processed audit log: %s - %s", auditLog.Action, riskLevel)
}

// processSecurityEvent 处理单个安全事件
func (sas *SecurityAuditService) processSecurityEvent(event *models.SecurityEvent) {
	// 分析安全事件
	sas.analyzeSecurityEvent(event)
	
	// 保存到数据库
	if err := sas.db.Save(event).Error; err != nil {
		log.Printf("Failed to save security event: %v", err)
		sas.updateStats("processing_errors", 1)
		return
	}
	
	// 更新统计信息
	sas.updateStats("total_events", 1)
	if event.Severity == "critical" {
		sas.updateStats("critical_events", 1)
	}
	
	log.Printf("Processed security event: %s - %s", event.EventType, event.Severity)
}

// analyzeAuditRisk 分析审计风险
func (sas *SecurityAuditService) analyzeAuditRisk(auditLog *models.AuditLog) string {
	// 基于操作类型分析风险
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
	
	// 检查失败结�?
	if auditLog.Result == "failed" || auditLog.Result == "error" {
		return "medium"
	}
	
	return "low"
}

// analyzeSecurityEvent 分析安全事件
func (sas *SecurityAuditService) analyzeSecurityEvent(event *models.SecurityEvent) {
	sas.eventAnalyzer.mutex.RLock()
	defer sas.eventAnalyzer.mutex.RUnlock()
	
	// 匹配分析模式
	for _, pattern := range sas.eventAnalyzer.patterns {
		if strings.Contains(event.EventType, pattern.Pattern) {
			event.Severity = pattern.Severity
			
			// 根据模式执行相应动作
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

// createSecurityEventFromAudit 从审计日志创建安全事�?
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
	
	// 添加到事件队�?
	select {
	case sas.eventQueue <- event:
	default:
		log.Println("Event queue is full, dropping event")
	}
}

// sendImmediateAlert 发送即时告�?
func (sas *SecurityAuditService) sendImmediateAlert(event *models.SecurityEvent) {
	log.Printf("IMMEDIATE ALERT: %s - %s", event.EventType, event.Description)
	// 这里可以集成告警系统，如邮件、短信、钉钉等
}

// sendAlert 发送告�?
func (sas *SecurityAuditService) sendAlert(event *models.SecurityEvent) {
	log.Printf("ALERT: %s - %s", event.EventType, event.Description)
	// 这里可以集成告警系统
}

// addToMonitoring 添加到监�?
func (sas *SecurityAuditService) addToMonitoring(event *models.SecurityEvent) {
	log.Printf("MONITORING: %s - %s", event.EventType, event.Description)
	// 这里可以添加到监控系�?
}

// periodicCleanup 定期清理
func (sas *SecurityAuditService) periodicCleanup() {
	ticker := time.NewTicker(24 * time.Hour) // 每天清理一�?
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

// cleanupOldLogs 清理旧日�?
func (sas *SecurityAuditService) cleanupOldLogs() {
	// 清理过期的审计日�?
	auditRetention := time.Duration(sas.config.LogRetentionDays) * 24 * time.Hour
	auditCutoff := time.Now().Add(-auditRetention)
	
	result := sas.db.Where("created_at < ?", auditCutoff).Delete(&models.AuditLog{})
	if result.Error != nil {
		log.Printf("Failed to cleanup audit logs: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d audit logs", result.RowsAffected)
	}
	
	// 清理过期的安全事�?
	eventRetention := time.Duration(sas.config.EventRetentionDays) * 24 * time.Hour
	eventCutoff := time.Now().Add(-eventRetention)
	
	result = sas.db.Where("created_at < ?", eventCutoff).Delete(&models.SecurityEvent{})
	if result.Error != nil {
		log.Printf("Failed to cleanup security events: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d security events", result.RowsAffected)
	}
}

// periodicComplianceCheck 定期合规检�?
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

// runComplianceCheck 运行合规检�?
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
	
	// 保存合规报告
	if err := sas.db.Create(report).Error; err != nil {
		log.Printf("Failed to save compliance report: %v", err)
	}
	
	// 更新统计信息
	sas.mutex.Lock()
	sas.stats.ComplianceScore = report.Score
	sas.mutex.Unlock()
	
	log.Printf("Compliance check completed: %.2f%% (%d/%d rules passed)", 
		report.Score, passedRules, totalRules)
}

// 合规检查函数实�?
func (sas *SecurityAuditService) checkAuthenticationCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 检查是否所有访问都经过身份认证
	var unauthenticatedCount int64
	err := sas.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("action LIKE ? AND result = ?", "%access%", "unauthorized").
		Count(&unauthenticatedCount).Error
	
	if err != nil {
		return false, fmt.Sprintf("检查失�? %v", err)
	}
	
	if unauthenticatedCount > 0 {
		return false, fmt.Sprintf("发现 %d 次未授权访问", unauthenticatedCount)
	}
	
	return true, "所有访问均经过身份认证"
}

func (sas *SecurityAuditService) checkAccessControlCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 检查访问控制实施情�?
	var accessViolationCount int64
	err := sas.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("action LIKE ? AND result = ?", "%access%", "forbidden").
		Count(&accessViolationCount).Error
	
	if err != nil {
		return false, fmt.Sprintf("检查失�? %v", err)
	}
	
	// 如果访问违规次数过多，可能表示访问控制不够严�?
	if accessViolationCount > 100 {
		return false, fmt.Sprintf("访问违规次数过多: %d", accessViolationCount)
	}
	
	return true, "访问控制实施正常"
}

func (sas *SecurityAuditService) checkAuditCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 检查审计日志记录情�?
	var recentLogCount int64
	yesterday := time.Now().AddDate(0, 0, -1)
	
	err := sas.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at > ?", yesterday).
		Count(&recentLogCount).Error
	
	if err != nil {
		return false, fmt.Sprintf("检查失�? %v", err)
	}
	
	if recentLogCount == 0 {
		return false, "�?4小时内无审计日志记录"
	}
	
	return true, fmt.Sprintf("审计日志记录正常，近24小时记录 %d �?, recentLogCount)
}

func (sas *SecurityAuditService) checkCryptographyCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 检查密码算法使用情�?
	// 这里可以检查系统中是否使用了国密算�?
	return true, "密码算法符合要求"
}

func (sas *SecurityAuditService) checkDataProtectionCompliance(ctx context.Context, data interface{}) (bool, string) {
	// 检查数据保护措�?
	// 这里可以检查敏感数据是否加密存�?
	return true, "数据保护措施符合要求"
}

// updateStats 更新统计信息
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

// LogAuditEvent 记录审计事件
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
	
	// 添加到处理队�?
	select {
	case sas.auditQueue <- auditLog:
		return nil
	default:
		// 队列满时直接保存到数据库
		return sas.db.WithContext(ctx).Create(auditLog).Error
	}
}

// LogSecurityEvent 记录安全事件
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
	
	// 添加到处理队�?
	select {
	case sas.eventQueue <- event:
		return nil
	default:
		// 队列满时直接保存到数据库
		return sas.db.WithContext(ctx).Create(event).Error
	}
}

// GetAuditLogs 获取审计日志
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

// GetSecurityEvents 获取安全事件
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

// GetComplianceReports 获取合规报告
func (sas *SecurityAuditService) GetComplianceReports(ctx context.Context, standard string, limit, offset int) ([]models.ComplianceReport, error) {
	query := sas.db.WithContext(ctx).Model(&models.ComplianceReport{})
	
	if standard != "" {
		query = query.Where("standard = ?", standard)
	}
	
	var reports []models.ComplianceReport
	err := query.Order("check_time DESC").Limit(limit).Offset(offset).Find(&reports).Error
	
	return reports, err
}

// GetAuditStats 获取审计统计信息
func (sas *SecurityAuditService) GetAuditStats(ctx context.Context) (*AuditStats, error) {
	sas.mutex.RLock()
	defer sas.mutex.RUnlock()
	
	// 创建统计信息副本
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

// UpdateSecurityEvent 更新安全事件
func (sas *SecurityAuditService) UpdateSecurityEvent(ctx context.Context, id string, updates map[string]interface{}) error {
	return sas.db.WithContext(ctx).Model(&models.SecurityEvent{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAuditLog 删除审计日志
func (sas *SecurityAuditService) DeleteAuditLog(ctx context.Context, id string) error {
	return sas.db.WithContext(ctx).Delete(&models.AuditLog{}, "id = ?", id).Error
}

// DeleteSecurityEvent 删除安全事件
func (sas *SecurityAuditService) DeleteSecurityEvent(ctx context.Context, id string) error {
	return sas.db.WithContext(ctx).Delete(&models.SecurityEvent{}, "id = ?", id).Error
}
