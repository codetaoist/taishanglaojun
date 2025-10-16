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
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// ThreatDetectionService ?
type ThreatDetectionService struct {
	db     *gorm.DB
	config *ThreatDetectionConfig
	rules  map[string]*models.DetectionRule
	mutex  sync.RWMutex
	
	// 
	sqlInjectionDetector *SQLInjectionDetector
	xssDetector         *XSSDetector
	bruteForceDetector  *BruteForceDetector
	ddosDetector        *DDoSDetector
	
	// 
	requestCounts map[string]int
	lastReset     time.Time
	
	// 
	stopChan chan bool
	running  bool
}

// ThreatDetectionConfig ?
type ThreatDetectionConfig struct {
	Enabled           bool     `yaml:"enabled"`
	ScanInterval      int      `yaml:"scan_interval"`
	AlertThreshold    int      `yaml:"alert_threshold"`
	BlockedIPs        []string `yaml:"blocked_ips"`
	WhitelistedIPs    []string `yaml:"whitelisted_ips"`
	MaxRequestsPerMin int      `yaml:"max_requests_per_min"`
}

// SQLInjectionDetector SQL
type SQLInjectionDetector struct {
	patterns []*regexp.Regexp
}

// XSSDetector XSS
type XSSDetector struct {
	patterns []*regexp.Regexp
}

// BruteForceDetector 
type BruteForceDetector struct {
	failedAttempts map[string][]time.Time
	mutex          sync.RWMutex
}

// DDoSDetector DDoS
type DDoSDetector struct {
	requestCounts map[string][]time.Time
	mutex         sync.RWMutex
}

// NewThreatDetectionService ?
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
	
	// 
	service.initDetectors()
	
	// ?
	service.loadRules()
	
	return service
}

// initDetectors 
func (tds *ThreatDetectionService) initDetectors() {
	// SQL
	tds.sqlInjectionDetector = &SQLInjectionDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|update\s+.*\s+set|delete\s+from)`),
			regexp.MustCompile(`(?i)(\'\s*or\s*\'\s*=\s*\'|\'\s*or\s*1\s*=\s*1|admin\'\s*--)`),
			regexp.MustCompile(`(?i)(exec\s*\(|sp_executesql|xp_cmdshell)`),
		},
	}
	
	// XSS
	tds.xssDetector = &XSSDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
			regexp.MustCompile(`(?i)javascript:`),
			regexp.MustCompile(`(?i)on\w+\s*=`),
			regexp.MustCompile(`(?i)<iframe[^>]*>`),
		},
	}
	
	// 
	tds.bruteForceDetector = &BruteForceDetector{
		failedAttempts: make(map[string][]time.Time),
	}
	
	// DDoS
	tds.ddosDetector = &DDoSDetector{
		requestCounts: make(map[string][]time.Time),
	}
}

// loadRules ?
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

// Start ?
func (tds *ThreatDetectionService) Start() {
	if tds.running {
		return
	}
	
	tds.running = true
	log.Println("Starting Threat Detection Service...")
	
	// 
	go tds.periodicScan()
	
	// 
	go tds.resetStatistics()
	
	log.Println("Threat Detection Service started successfully")
}

// Stop ?
func (tds *ThreatDetectionService) Stop() {
	if !tds.running {
		return
	}
	
	log.Println("Stopping Threat Detection Service...")
	tds.stopChan <- true
	tds.running = false
	log.Println("Threat Detection Service stopped")
}

// periodicScan 
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

// resetStatistics 
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

// performScan 
func (tds *ThreatDetectionService) performScan() {
	log.Println("Performing threat detection scan...")
	
	// İ?
	var events []models.SecurityEvent
	if err := tds.db.Where("processed = ?", false).Find(&events).Error; err != nil {
		log.Printf("Failed to fetch security events: %v", err)
		return
	}
	
	for _, event := range events {
		tds.analyzeSecurityEvent(&event)
	}
}

// analyzeSecurityEvent 
func (tds *ThreatDetectionService) analyzeSecurityEvent(event *models.SecurityEvent) {
	// ?
	tds.mutex.RLock()
	rules := tds.rules
	tds.mutex.RUnlock()
	
	for _, rule := range rules {
		if tds.matchRule(event, rule) {
			tds.createThreatAlert(event, rule)
		}
	}
	
	// 
	event.Processed = true
	tds.db.Save(event)
}

// matchRule ?
func (tds *ThreatDetectionService) matchRule(event *models.SecurityEvent, rule *models.DetectionRule) bool {
	// 
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

// evaluateCondition 
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

// createThreatAlert 澯
func (tds *ThreatDetectionService) createThreatAlert(event *models.SecurityEvent, rule *models.DetectionRule) {
	alert := &models.ThreatAlert{
		Title:       fmt.Sprintf("? %s", rule.Name),
		Description: fmt.Sprintf(": %s", event.Description),
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
	
	// 
	tds.executeActions(alert, rule)
}

// executeActions 
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

// blockIP IP
func (tds *ThreatDetectionService) blockIP(ip string) {
	if ip == "" {
		return
	}
	
	log.Printf("Blocking IP address: %s", ip)
	// API?
}

// sendNotification 
func (tds *ThreatDetectionService) sendNotification(alert *models.ThreatAlert) {
	log.Printf("Sending notification for alert: %s", alert.Title)
	// Slack
}

// logSecurityEvent 
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

// DetectSQLInjection SQL
func (tds *ThreatDetectionService) DetectSQLInjection(req *http.Request) bool {
	// URL
	for _, values := range req.URL.Query() {
		for _, value := range values {
			if tds.sqlInjectionDetector.detect(value) {
				tds.recordSecurityEvent("sql_injection", req)
				return true
			}
		}
	}
	
	// POST
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

// DetectXSS XSS
func (tds *ThreatDetectionService) DetectXSS(req *http.Request) bool {
	// URL
	for _, values := range req.URL.Query() {
		for _, value := range values {
			if tds.xssDetector.detect(value) {
				tds.recordSecurityEvent("xss_attack", req)
				return true
			}
		}
	}
	
	// POST
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

// DetectBruteForce ?
func (tds *ThreatDetectionService) DetectBruteForce(ip string, failed bool) bool {
	if !failed {
		return false
	}
	
	return tds.bruteForceDetector.detect(ip)
}

// DetectDDoS DDoS
func (tds *ThreatDetectionService) DetectDDoS(req *http.Request) bool {
	ip := tds.getClientIP(req)
	return tds.ddosDetector.detect(ip)
}

// recordSecurityEvent 
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

// getClientIP IP
func (tds *ThreatDetectionService) getClientIP(req *http.Request) string {
	// X-Forwarded-For?
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// X-Real-IP?
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// RemoteAddr
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

// SQL
func (sid *SQLInjectionDetector) detect(input string) bool {
	for _, pattern := range sid.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// XSS
func (xd *XSSDetector) detect(input string) bool {
	for _, pattern := range xd.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// 
func (bfd *BruteForceDetector) detect(ip string) bool {
	bfd.mutex.Lock()
	defer bfd.mutex.Unlock()
	
	now := time.Now()
	attempts := bfd.failedAttempts[ip]
	
	// 5
	var validAttempts []time.Time
	for _, attempt := range attempts {
		if now.Sub(attempt) < 5*time.Minute {
			validAttempts = append(validAttempts, attempt)
		}
	}
	
	// 
	validAttempts = append(validAttempts, now)
	bfd.failedAttempts[ip] = validAttempts
	
	// 5??
	return len(validAttempts) > 5
}

// DDoS
func (dd *DDoSDetector) detect(ip string) bool {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()
	
	now := time.Now()
	requests := dd.requestCounts[ip]
	
	// 1
	var validRequests []time.Time
	for _, request := range requests {
		if now.Sub(request) < 1*time.Minute {
			validRequests = append(validRequests, request)
		}
	}
	
	// 
	validRequests = append(validRequests, now)
	dd.requestCounts[ip] = validRequests
	
	// 1?00DDoS
	return len(validRequests) > 100
}

// GetAlerts 澯
func (tds *ThreatDetectionService) GetAlerts(ctx context.Context, limit, offset int) ([]models.ThreatAlert, error) {
	var alerts []models.ThreatAlert
	err := tds.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&alerts).Error
	
	return alerts, err
}

// GetRules ?
func (tds *ThreatDetectionService) GetRules(ctx context.Context) ([]models.DetectionRule, error) {
	var rules []models.DetectionRule
	err := tds.db.WithContext(ctx).Find(&rules).Error
	return rules, err
}

// CreateRule ?
func (tds *ThreatDetectionService) CreateRule(ctx context.Context, rule *models.DetectionRule) error {
	if err := tds.db.WithContext(ctx).Create(rule).Error; err != nil {
		return err
	}
	
	// 
	tds.loadRules()
	return nil
}

// UpdateRule ?
func (tds *ThreatDetectionService) UpdateRule(ctx context.Context, id string, rule *models.DetectionRule) error {
	if err := tds.db.WithContext(ctx).Where("id = ?", id).Updates(rule).Error; err != nil {
		return err
	}
	
	// 
	tds.loadRules()
	return nil
}

// DeleteRule ?
func (tds *ThreatDetectionService) DeleteRule(ctx context.Context, id string) error {
	if err := tds.db.WithContext(ctx).Delete(&models.DetectionRule{}, "id = ?", id).Error; err != nil {
		return err
	}
	
	// 
	tds.loadRules()
	return nil
}

