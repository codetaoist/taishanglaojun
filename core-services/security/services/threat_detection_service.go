package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"github.com/taishanglaojun/core-services/security/models"
)

// ThreatDetectionService 威胁检测服务
type ThreatDetectionService struct {
	db     *gorm.DB
	config *ThreatDetectionConfig
	rules  map[string]*models.DetectionRule
	mutex  sync.RWMutex
	
	// 检测器
	sqlInjectionDetector *SQLInjectionDetector
	xssDetector         *XSSDetector
	bruteForceDetector  *BruteForceDetector
	ddosDetector        *DDoSDetector
	
	// 统计信息
	requestCounts map[string]int
	lastReset     time.Time
	
	// 控制通道
	stopChan chan bool
	running  bool
}

// ThreatDetectionConfig 威胁检测配置
type ThreatDetectionConfig struct {
	Enabled           bool     `yaml:"enabled"`
	ScanInterval      int      `yaml:"scan_interval"`
	AlertThreshold    int      `yaml:"alert_threshold"`
	BlockedIPs        []string `yaml:"blocked_ips"`
	WhitelistedIPs    []string `yaml:"whitelisted_ips"`
	MaxRequestsPerMin int      `yaml:"max_requests_per_min"`
}

// SQLInjectionDetector SQL注入检测器
type SQLInjectionDetector struct {
	patterns []*regexp.Regexp
}

// XSSDetector XSS攻击检测器
type XSSDetector struct {
	patterns []*regexp.Regexp
}

// BruteForceDetector 暴力破解检测器
type BruteForceDetector struct {
	failedAttempts map[string][]time.Time
	mutex          sync.RWMutex
}

// DDoSDetector DDoS攻击检测器
type DDoSDetector struct {
	requestCounts map[string][]time.Time
	mutex         sync.RWMutex
}

// NewThreatDetectionService 创建威胁检测服务
func NewThreatDetectionService(db *gorm.DB, config *ThreatDetectionConfig) *ThreatDetectionService {
	service := &ThreatDetectionService{
		db:            db,
		config:        config,
		rules:         make(map[string]*models.DetectionRule),
		requestCounts: make(map[string]int),
		lastReset:     time.Now(),
		stopChan:      make(chan bool),
		running:       false,
	}
	
	// 初始化检测器
	service.initDetectors()
	
	// 加载检测规则
	service.loadRules()
	
	return service
}

// initDetectors 初始化各种检测器
func (tds *ThreatDetectionService) initDetectors() {
	// 初始化SQL注入检测器
	tds.sqlInjectionDetector = &SQLInjectionDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|update\s+.*\s+set|delete\s+from)`),
			regexp.MustCompile(`(?i)(\'\s*or\s*\'\s*=\s*\'|\'\s*or\s*1\s*=\s*1|admin\'\s*--)`),
			regexp.MustCompile(`(?i)(exec\s*\(|sp_executesql|xp_cmdshell)`),
		},
	}
	
	// 初始化XSS检测器
	tds.xssDetector = &XSSDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
			regexp.MustCompile(`(?i)javascript:`),
			regexp.MustCompile(`(?i)on\w+\s*=`),
			regexp.MustCompile(`(?i)<iframe[^>]*>`),
		},
	}
	
	// 初始化暴力破解检测器
	tds.bruteForceDetector = &BruteForceDetector{
		failedAttempts: make(map[string][]time.Time),
	}
	
	// 初始化DDoS检测器
	tds.ddosDetector = &DDoSDetector{
		requestCounts: make(map[string][]time.Time),
	}
}

// loadRules 加载检测规则
func (tds *ThreatDetectionService) loadRules() {
	var rules []models.DetectionRule
	if err := tds.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		log.Printf("Failed to load detection rules: %v", err)
		return
	}
	
	tds.mutex.Lock()
	defer tds.mutex.Unlock()
	
	for _, rule := range rules {
		tds.rules[rule.ID] = &rule
	}
	
	log.Printf("Loaded %d detection rules", len(rules))
}

// Start 启动威胁检测服务
func (tds *ThreatDetectionService) Start() {
	if tds.running {
		return
	}
	
	tds.running = true
	log.Println("Starting Threat Detection Service...")
	
	// 启动定期扫描
	go tds.periodicScan()
	
	// 启动统计重置
	go tds.resetStatistics()
	
	log.Println("Threat Detection Service started successfully")
}

// Stop 停止威胁检测服务
func (tds *ThreatDetectionService) Stop() {
	if !tds.running {
		return
	}
	
	log.Println("Stopping Threat Detection Service...")
	tds.stopChan <- true
	tds.running = false
	log.Println("Threat Detection Service stopped")
}

// periodicScan 定期扫描
func (tds *ThreatDetectionService) periodicScan() {
	ticker := time.NewTicker(time.Duration(tds.config.ScanInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			tds.performScan()
		case <-tds.stopChan:
			return
		}
	}
}

// resetStatistics 重置统计信息
func (tds *ThreatDetectionService) resetStatistics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			tds.mutex.Lock()
			tds.requestCounts = make(map[string]int)
			tds.lastReset = time.Now()
			tds.mutex.Unlock()
		case <-tds.stopChan:
			return
		}
	}
}

// performScan 执行扫描
func (tds *ThreatDetectionService) performScan() {
	log.Println("Performing threat detection scan...")
	
	// 检查未处理的安全事件
	var events []models.SecurityEvent
	if err := tds.db.Where("processed = ?", false).Find(&events).Error; err != nil {
		log.Printf("Failed to fetch security events: %v", err)
		return
	}
	
	for _, event := range events {
		tds.analyzeSecurityEvent(&event)
	}
}

// analyzeSecurityEvent 分析安全事件
func (tds *ThreatDetectionService) analyzeSecurityEvent(event *models.SecurityEvent) {
	// 应用检测规则
	tds.mutex.RLock()
	rules := tds.rules
	tds.mutex.RUnlock()
	
	for _, rule := range rules {
		if tds.matchRule(event, rule) {
			tds.createThreatAlert(event, rule)
		}
	}
	
	// 标记事件为已处理
	event.Processed = true
	tds.db.Save(event)
}

// matchRule 匹配检测规则
func (tds *ThreatDetectionService) matchRule(event *models.SecurityEvent, rule *models.DetectionRule) bool {
	// 简化的规则匹配逻辑
	conditions, ok := rule.Conditions["conditions"].([]interface{})
	if !ok {
		return false
	}
	
	for _, condition := range conditions {
		condMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}
		
		field, ok := condMap["field"].(string)
		if !ok {
			continue
		}
		
		operator, ok := condMap["operator"].(string)
		if !ok {
			continue
		}
		
		value, ok := condMap["value"].(string)
		if !ok {
			continue
		}
		
		if tds.evaluateCondition(event, field, operator, value) {
			return true
		}
	}
	
	return false
}

// evaluateCondition 评估条件
func (tds *ThreatDetectionService) evaluateCondition(event *models.SecurityEvent, field, operator, value string) bool {
	var fieldValue string
	
	switch field {
	case "event_type":
		fieldValue = event.EventType
	case "severity":
		fieldValue = event.Severity
	case "source_ip":
		fieldValue = event.SourceIP
	case "target_ip":
		fieldValue = event.TargetIP
	default:
		return false
	}
	
	switch operator {
	case "equals":
		return fieldValue == value
	case "contains":
		return strings.Contains(fieldValue, value)
	case "matches":
		matched, _ := regexp.MatchString(value, fieldValue)
		return matched
	default:
		return false
	}
}

// createThreatAlert 创建威胁告警
func (tds *ThreatDetectionService) createThreatAlert(event *models.SecurityEvent, rule *models.DetectionRule) {
	alert := &models.ThreatAlert{
		Title:       fmt.Sprintf("威胁检测: %s", rule.Name),
		Description: fmt.Sprintf("检测到威胁事件: %s", event.Description),
		Severity:    rule.Severity,
		Category:    rule.Category,
		SourceIP:    event.SourceIP,
		TargetIP:    event.TargetIP,
		UserID:      event.UserID,
		RuleID:      rule.ID,
		Status:      "open",
		RawData:     models.JSONB(map[string]interface{}{
			"event_id": event.ID,
			"rule_id":  rule.ID,
			"raw_data": event.RawData,
		}),
	}
	
	if err := tds.db.Create(alert).Error; err != nil {
		log.Printf("Failed to create threat alert: %v", err)
		return
	}
	
	log.Printf("Created threat alert: %s (Severity: %s)", alert.Title, alert.Severity)
	
	// 执行响应动作
	tds.executeActions(alert, rule)
}

// executeActions 执行响应动作
func (tds *ThreatDetectionService) executeActions(alert *models.ThreatAlert, rule *models.DetectionRule) {
	actions, ok := rule.Actions["actions"].([]interface{})
	if !ok {
		return
	}
	
	for _, action := range actions {
		actionMap, ok := action.(map[string]interface{})
		if !ok {
			continue
		}
		
		actionType, ok := actionMap["type"].(string)
		if !ok {
			continue
		}
		
		switch actionType {
		case "block_ip":
			tds.blockIP(alert.SourceIP)
		case "send_notification":
			tds.sendNotification(alert)
		case "log_event":
			tds.logSecurityEvent(alert)
		}
	}
}

// blockIP 阻止IP地址
func (tds *ThreatDetectionService) blockIP(ip string) {
	if ip == "" {
		return
	}
	
	log.Printf("Blocking IP address: %s", ip)
	// 这里可以集成防火墙API或其他安全设备
}

// sendNotification 发送通知
func (tds *ThreatDetectionService) sendNotification(alert *models.ThreatAlert) {
	log.Printf("Sending notification for alert: %s", alert.Title)
	// 这里可以集成邮件、短信、Slack等通知系统
}

// logSecurityEvent 记录安全事件
func (tds *ThreatDetectionService) logSecurityEvent(alert *models.ThreatAlert) {
	auditLog := &models.AuditLog{
		Action:       "threat_detected",
		ResourceType: "threat_alert",
		ResourceID:   alert.ID,
		Details: models.JSONB(map[string]interface{}{
			"alert_id":    alert.ID,
			"severity":    alert.Severity,
			"category":    alert.Category,
			"source_ip":   alert.SourceIP,
			"target_ip":   alert.TargetIP,
		}),
		Success: true,
	}
	
	tds.db.Create(auditLog)
}

// DetectSQLInjection 检测SQL注入
func (tds *ThreatDetectionService) DetectSQLInjection(req *http.Request) bool {
	// 检查URL参数
	for _, values := range req.URL.Query() {
		for _, value := range values {
			if tds.sqlInjectionDetector.detect(value) {
				tds.recordSecurityEvent("sql_injection", req)
				return true
			}
		}
	}
	
	// 检查POST数据
	if req.Method == "POST" {
		req.ParseForm()
		for _, values := range req.PostForm {
			for _, value := range values {
				if tds.sqlInjectionDetector.detect(value) {
					tds.recordSecurityEvent("sql_injection", req)
					return true
				}
			}
		}
	}
	
	return false
}

// DetectXSS 检测XSS攻击
func (tds *ThreatDetectionService) DetectXSS(req *http.Request) bool {
	// 检查URL参数
	for _, values := range req.URL.Query() {
		for _, value := range values {
			if tds.xssDetector.detect(value) {
				tds.recordSecurityEvent("xss_attack", req)
				return true
			}
		}
	}
	
	// 检查POST数据
	if req.Method == "POST" {
		req.ParseForm()
		for _, values := range req.PostForm {
			for _, value := range values {
				if tds.xssDetector.detect(value) {
					tds.recordSecurityEvent("xss_attack", req)
					return true
				}
			}
		}
	}
	
	return false
}

// DetectBruteForce 检测暴力破解
func (tds *ThreatDetectionService) DetectBruteForce(ip string, failed bool) bool {
	if !failed {
		return false
	}
	
	return tds.bruteForceDetector.detect(ip)
}

// DetectDDoS 检测DDoS攻击
func (tds *ThreatDetectionService) DetectDDoS(req *http.Request) bool {
	ip := tds.getClientIP(req)
	return tds.ddosDetector.detect(ip)
}

// recordSecurityEvent 记录安全事件
func (tds *ThreatDetectionService) recordSecurityEvent(eventType string, req *http.Request) {
	event := &models.SecurityEvent{
		EventType:   eventType,
		Severity:    "medium",
		SourceIP:    tds.getClientIP(req),
		Description: fmt.Sprintf("Detected %s from %s", eventType, tds.getClientIP(req)),
		RawData: models.JSONB(map[string]interface{}{
			"url":        req.URL.String(),
			"method":     req.Method,
			"user_agent": req.UserAgent(),
			"headers":    req.Header,
		}),
		Processed: false,
	}
	
	tds.db.Create(event)
}

// getClientIP 获取客户端IP
func (tds *ThreatDetectionService) getClientIP(req *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// 检查X-Real-IP头
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 使用RemoteAddr
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

// SQL注入检测器方法
func (sid *SQLInjectionDetector) detect(input string) bool {
	for _, pattern := range sid.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// XSS检测器方法
func (xd *XSSDetector) detect(input string) bool {
	for _, pattern := range xd.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// 暴力破解检测器方法
func (bfd *BruteForceDetector) detect(ip string) bool {
	bfd.mutex.Lock()
	defer bfd.mutex.Unlock()
	
	now := time.Now()
	attempts := bfd.failedAttempts[ip]
	
	// 清理过期的尝试记录（5分钟内）
	var validAttempts []time.Time
	for _, attempt := range attempts {
		if now.Sub(attempt) < 5*time.Minute {
			validAttempts = append(validAttempts, attempt)
		}
	}
	
	// 添加当前尝试
	validAttempts = append(validAttempts, now)
	bfd.failedAttempts[ip] = validAttempts
	
	// 如果5分钟内失败次数超过5次，认为是暴力破解
	return len(validAttempts) > 5
}

// DDoS检测器方法
func (dd *DDoSDetector) detect(ip string) bool {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()
	
	now := time.Now()
	requests := dd.requestCounts[ip]
	
	// 清理过期的请求记录（1分钟内）
	var validRequests []time.Time
	for _, request := range requests {
		if now.Sub(request) < 1*time.Minute {
			validRequests = append(validRequests, request)
		}
	}
	
	// 添加当前请求
	validRequests = append(validRequests, now)
	dd.requestCounts[ip] = validRequests
	
	// 如果1分钟内请求次数超过100次，认为是DDoS攻击
	return len(validRequests) > 100
}

// GetAlerts 获取威胁告警
func (tds *ThreatDetectionService) GetAlerts(ctx context.Context, limit, offset int) ([]models.ThreatAlert, error) {
	var alerts []models.ThreatAlert
	err := tds.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&alerts).Error
	
	return alerts, err
}

// GetRules 获取检测规则
func (tds *ThreatDetectionService) GetRules(ctx context.Context) ([]models.DetectionRule, error) {
	var rules []models.DetectionRule
	err := tds.db.WithContext(ctx).Find(&rules).Error
	return rules, err
}

// CreateRule 创建检测规则
func (tds *ThreatDetectionService) CreateRule(ctx context.Context, rule *models.DetectionRule) error {
	if err := tds.db.WithContext(ctx).Create(rule).Error; err != nil {
		return err
	}
	
	// 重新加载规则
	tds.loadRules()
	return nil
}

// UpdateRule 更新检测规则
func (tds *ThreatDetectionService) UpdateRule(ctx context.Context, id string, rule *models.DetectionRule) error {
	if err := tds.db.WithContext(ctx).Where("id = ?", id).Updates(rule).Error; err != nil {
		return err
	}
	
	// 重新加载规则
	tds.loadRules()
	return nil
}

// DeleteRule 删除检测规则
func (tds *ThreatDetectionService) DeleteRule(ctx context.Context, id string) error {
	if err := tds.db.WithContext(ctx).Delete(&models.DetectionRule{}, "id = ?", id).Error; err != nil {
		return err
	}
	
	// 重新加载规则
	tds.loadRules()
	return nil
}