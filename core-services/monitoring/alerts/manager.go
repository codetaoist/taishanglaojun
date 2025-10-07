package alerts

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/taishanglaojun/core-services/monitoring/models"
)

// AlertManager 告警管理器
type AlertManager struct {
	// 配置
	config *AlertManagerConfig
	
	// 存储
	storage interfaces.MetricStorage
	
	// 通知管理器
	notificationManager interfaces.NotificationManager
	
	// 告警规则
	rules map[string]*models.AlertRule
	
	// 活跃告警
	activeAlerts map[string]*models.Alert
	
	// 告警历史
	alertHistory []*models.AlertHistory
	
	// 静默规则
	silences map[string]*models.Silence
	
	// 抑制规则
	inhibitions []*models.Inhibition
	
	// 升级策略
	escalationPolicies map[string]*models.EscalationPolicy
	
	// 值班计划
	onCallSchedules map[string]*models.OnCallSchedule
	
	// 同步锁
	mutex sync.RWMutex
	
	// 运行状态
	running bool
	
	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
	
	// 评估器
	evaluator *AlertEvaluator
	
	// 统计信息
	stats *AlertManagerStats
}

// AlertManagerConfig 告警管理器配置
type AlertManagerConfig struct {
	// 评估配置
	EvaluationInterval time.Duration `yaml:"evaluation_interval"`
	EvaluationTimeout  time.Duration `yaml:"evaluation_timeout"`
	
	// 通知配置
	NotificationTimeout time.Duration `yaml:"notification_timeout"`
	NotificationRetries int           `yaml:"notification_retries"`
	
	// 历史配置
	HistoryRetention time.Duration `yaml:"history_retention"`
	MaxHistorySize   int           `yaml:"max_history_size"`
	
	// 批处理配置
	BatchSize     int           `yaml:"batch_size"`
	BatchTimeout  time.Duration `yaml:"batch_timeout"`
	
	// 去重配置
	GroupWait    time.Duration `yaml:"group_wait"`
	GroupInterval time.Duration `yaml:"group_interval"`
	RepeatInterval time.Duration `yaml:"repeat_interval"`
	
	// 路由配置
	Routes []*RouteConfig `yaml:"routes"`
	
	// 接收器配置
	Receivers []*ReceiverConfig `yaml:"receivers"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Match    map[string]string `yaml:"match"`
	MatchRE  map[string]string `yaml:"match_re"`
	Receiver string            `yaml:"receiver"`
	Continue bool              `yaml:"continue"`
	Routes   []*RouteConfig    `yaml:"routes"`
	
	GroupBy       []string      `yaml:"group_by"`
	GroupWait     time.Duration `yaml:"group_wait"`
	GroupInterval time.Duration `yaml:"group_interval"`
	RepeatInterval time.Duration `yaml:"repeat_interval"`
}

// ReceiverConfig 接收器配置
type ReceiverConfig struct {
	Name            string                    `yaml:"name"`
	EmailConfigs    []*EmailConfig           `yaml:"email_configs"`
	WebhookConfigs  []*WebhookConfig         `yaml:"webhook_configs"`
	SlackConfigs    []*SlackConfig           `yaml:"slack_configs"`
	DingTalkConfigs []*DingTalkConfig        `yaml:"dingtalk_configs"`
	WeChatConfigs   []*WeChatConfig          `yaml:"wechat_configs"`
	SMSConfigs      []*SMSConfig             `yaml:"sms_configs"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	To       []string `yaml:"to"`
	From     string   `yaml:"from"`
	Subject  string   `yaml:"subject"`
	Body     string   `yaml:"body"`
	HTML     string   `yaml:"html"`
	Headers  map[string]string `yaml:"headers"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL         string            `yaml:"url"`
	Method      string            `yaml:"method"`
	Headers     map[string]string `yaml:"headers"`
	Body        string            `yaml:"body"`
	Timeout     time.Duration     `yaml:"timeout"`
	MaxRetries  int               `yaml:"max_retries"`
}

// SlackConfig Slack配置
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
	Title      string `yaml:"title"`
	Text       string `yaml:"text"`
	Color      string `yaml:"color"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Secret     string `yaml:"secret"`
	Title      string `yaml:"title"`
	Text       string `yaml:"text"`
	AtMobiles  []string `yaml:"at_mobiles"`
	AtAll      bool   `yaml:"at_all"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	MsgType    string `yaml:"msg_type"`
	Content    string `yaml:"content"`
	AtUsers    []string `yaml:"at_users"`
	AtAll      bool   `yaml:"at_all"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider    string   `yaml:"provider"`
	AccessKey   string   `yaml:"access_key"`
	SecretKey   string   `yaml:"secret_key"`
	SignName    string   `yaml:"sign_name"`
	TemplateCode string  `yaml:"template_code"`
	PhoneNumbers []string `yaml:"phone_numbers"`
}

// AlertManagerStats 告警管理器统计信息
type AlertManagerStats struct {
	Running           bool      `json:"running"`
	RulesCount        int       `json:"rules_count"`
	ActiveAlertsCount int       `json:"active_alerts_count"`
	SilencesCount     int       `json:"silences_count"`
	TotalEvaluations  uint64    `json:"total_evaluations"`
	TotalAlerts       uint64    `json:"total_alerts"`
	TotalNotifications uint64   `json:"total_notifications"`
	LastEvaluationTime time.Time `json:"last_evaluation_time"`
	LastUpdateTime    time.Time `json:"last_update_time"`
}

// NewAlertManager 创建告警管理器
func NewAlertManager(config *AlertManagerConfig, storage interfaces.MetricStorage, notificationManager interfaces.NotificationManager) *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AlertManager{
		config:              config,
		storage:             storage,
		notificationManager: notificationManager,
		rules:               make(map[string]*models.AlertRule),
		activeAlerts:        make(map[string]*models.Alert),
		alertHistory:        make([]*models.AlertHistory, 0),
		silences:            make(map[string]*models.Silence),
		inhibitions:         make([]*models.Inhibition, 0),
		escalationPolicies:  make(map[string]*models.EscalationPolicy),
		onCallSchedules:     make(map[string]*models.OnCallSchedule),
		ctx:                 ctx,
		cancel:              cancel,
		running:             false,
		stats:               &AlertManagerStats{},
	}
}

// Initialize 初始化告警管理器
func (am *AlertManager) Initialize() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	// 创建评估器
	am.evaluator = NewAlertEvaluator(am.storage, am.config.EvaluationTimeout)
	
	return nil
}

// Start 启动告警管理器
func (am *AlertManager) Start() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if am.running {
		return fmt.Errorf("alert manager is already running")
	}
	
	// 启动评估循环
	go am.evaluationLoop()
	
	// 启动清理循环
	go am.cleanupLoop()
	
	am.running = true
	am.stats.Running = true
	
	return nil
}

// Stop 停止告警管理器
func (am *AlertManager) Stop() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if !am.running {
		return nil
	}
	
	// 取消上下文
	am.cancel()
	
	am.running = false
	am.stats.Running = false
	
	return nil
}

// evaluationLoop 评估循环
func (am *AlertManager) evaluationLoop() {
	ticker := time.NewTicker(am.config.EvaluationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.evaluateRules()
		}
	}
}

// evaluateRules 评估所有规则
func (am *AlertManager) evaluateRules() {
	am.mutex.RLock()
	rules := make([]*models.AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}
	am.mutex.RUnlock()
	
	for _, rule := range rules {
		if rule.Enabled {
			am.evaluateRule(rule)
		}
	}
	
	am.stats.LastEvaluationTime = time.Now()
	am.stats.TotalEvaluations++
}

// evaluateRule 评估单个规则
func (am *AlertManager) evaluateRule(rule *models.AlertRule) {
	// 执行查询
	result, err := am.evaluator.Evaluate(am.ctx, rule)
	if err != nil {
		fmt.Printf("Failed to evaluate rule %s: %v\n", rule.Name, err)
		return
	}
	
	// 处理评估结果
	for _, alertResult := range result {
		am.processAlertResult(rule, alertResult)
	}
}

// processAlertResult 处理告警结果
func (am *AlertManager) processAlertResult(rule *models.AlertRule, result *AlertEvaluationResult) {
	alertKey := am.generateAlertKey(rule, result.Labels)
	
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	existingAlert, exists := am.activeAlerts[alertKey]
	
	if result.Firing {
		if exists {
			// 更新现有告警
			existingAlert.Value = result.Value
			existingAlert.UpdatedAt = time.Now()
		} else {
			// 创建新告警
			alert := &models.Alert{
				ID:          am.generateAlertID(),
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Labels:      result.Labels,
				Annotations: rule.Annotations,
				Status:      models.AlertStatusFiring,
				Severity:    rule.Severity,
				Value:       result.Value,
				StartsAt:    time.Now(),
				UpdatedAt:   time.Now(),
				Fingerprint: am.generateFingerprint(result.Labels),
			}
			
			am.activeAlerts[alertKey] = alert
			am.stats.TotalAlerts++
			
			// 发送通知
			am.sendNotification(alert, models.AlertStatusFiring)
			
			// 记录历史
			am.recordAlertHistory(alert, models.AlertStatusFiring, "Alert fired")
		}
	} else {
		if exists {
			// 解决告警
			existingAlert.Status = models.AlertStatusResolved
			existingAlert.EndsAt = time.Now()
			existingAlert.UpdatedAt = time.Now()
			
			// 发送解决通知
			am.sendNotification(existingAlert, models.AlertStatusResolved)
			
			// 记录历史
			am.recordAlertHistory(existingAlert, models.AlertStatusResolved, "Alert resolved")
			
			// 从活跃告警中移除
			delete(am.activeAlerts, alertKey)
		}
	}
}

// sendNotification 发送通知
func (am *AlertManager) sendNotification(alert *models.Alert, status models.AlertStatus) {
	// 检查是否被静默
	if am.isAlertSilenced(alert) {
		return
	}
	
	// 检查是否被抑制
	if am.isAlertInhibited(alert) {
		return
	}
	
	// 创建通知
	notification := &models.Notification{
		ID:        am.generateNotificationID(),
		AlertID:   alert.ID,
		Type:      "alert",
		Status:    string(status),
		Title:     am.buildNotificationTitle(alert, status),
		Message:   am.buildNotificationMessage(alert, status),
		Labels:    alert.Labels,
		Severity:  alert.Severity,
		CreatedAt: time.Now(),
	}
	
	// 发送通知
	if err := am.notificationManager.Send(am.ctx, notification); err != nil {
		fmt.Printf("Failed to send notification for alert %s: %v\n", alert.ID, err)
	} else {
		am.stats.TotalNotifications++
	}
}

// isAlertSilenced 检查告警是否被静默
func (am *AlertManager) isAlertSilenced(alert *models.Alert) bool {
	for _, silence := range am.silences {
		if silence.IsActive() && silence.Matches(alert.Labels) {
			return true
		}
	}
	return false
}

// isAlertInhibited 检查告警是否被抑制
func (am *AlertManager) isAlertInhibited(alert *models.Alert) bool {
	for _, inhibition := range am.inhibitions {
		if am.checkInhibition(inhibition, alert) {
			return true
		}
	}
	return false
}

// checkInhibition 检查抑制规则
func (am *AlertManager) checkInhibition(inhibition *models.Inhibition, alert *models.Alert) bool {
	// 检查是否有匹配的源告警
	for _, activeAlert := range am.activeAlerts {
		if activeAlert.Status == models.AlertStatusFiring {
			// 检查源匹配器
			if am.matchLabels(activeAlert.Labels, inhibition.SourceMatchers) {
				// 检查目标匹配器
				if am.matchLabels(alert.Labels, inhibition.TargetMatchers) {
					// 检查相等标签
					if am.checkEqualLabels(activeAlert.Labels, alert.Labels, inhibition.Equal) {
						return true
					}
				}
			}
		}
	}
	return false
}

// matchLabels 匹配标签
func (am *AlertManager) matchLabels(labels map[string]string, matchers []models.Matcher) bool {
	for _, matcher := range matchers {
		value, exists := labels[matcher.Name]
		if !exists {
			return false
		}
		
		switch matcher.Type {
		case models.MatcherTypeEqual:
			if value != matcher.Value {
				return false
			}
		case models.MatcherTypeNotEqual:
			if value == matcher.Value {
				return false
			}
		case models.MatcherTypeRegex:
			// 这里需要实现正则匹配
			// matched, _ := regexp.MatchString(matcher.Value, value)
			// if !matched {
			//     return false
			// }
		case models.MatcherTypeNotRegex:
			// 这里需要实现正则匹配
			// matched, _ := regexp.MatchString(matcher.Value, value)
			// if matched {
			//     return false
			// }
		}
	}
	return true
}

// checkEqualLabels 检查相等标签
func (am *AlertManager) checkEqualLabels(labels1, labels2 map[string]string, equalLabels []string) bool {
	for _, label := range equalLabels {
		if labels1[label] != labels2[label] {
			return false
		}
	}
	return true
}

// buildNotificationTitle 构建通知标题
func (am *AlertManager) buildNotificationTitle(alert *models.Alert, status models.AlertStatus) string {
	switch status {
	case models.AlertStatusFiring:
		return fmt.Sprintf("[%s] %s", alert.Severity, alert.RuleName)
	case models.AlertStatusResolved:
		return fmt.Sprintf("[RESOLVED] %s", alert.RuleName)
	default:
		return alert.RuleName
	}
}

// buildNotificationMessage 构建通知消息
func (am *AlertManager) buildNotificationMessage(alert *models.Alert, status models.AlertStatus) string {
	message := fmt.Sprintf("Alert: %s\n", alert.RuleName)
	message += fmt.Sprintf("Status: %s\n", status)
	message += fmt.Sprintf("Severity: %s\n", alert.Severity)
	message += fmt.Sprintf("Value: %.2f\n", alert.Value)
	message += fmt.Sprintf("Time: %s\n", alert.StartsAt.Format(time.RFC3339))
	
	if len(alert.Labels) > 0 {
		message += "Labels:\n"
		for k, v := range alert.Labels {
			message += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}
	
	if len(alert.Annotations) > 0 {
		message += "Annotations:\n"
		for k, v := range alert.Annotations {
			message += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}
	
	return message
}

// recordAlertHistory 记录告警历史
func (am *AlertManager) recordAlertHistory(alert *models.Alert, status models.AlertStatus, message string) {
	history := &models.AlertHistory{
		ID:        am.generateHistoryID(),
		AlertID:   alert.ID,
		RuleID:    alert.RuleID,
		Status:    status,
		Message:   message,
		Labels:    alert.Labels,
		Value:     alert.Value,
		Timestamp: time.Now(),
	}
	
	am.alertHistory = append(am.alertHistory, history)
	
	// 限制历史记录数量
	if len(am.alertHistory) > am.config.MaxHistorySize {
		am.alertHistory = am.alertHistory[1:]
	}
}

// cleanupLoop 清理循环
func (am *AlertManager) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.cleanup()
		}
	}
}

// cleanup 清理过期数据
func (am *AlertManager) cleanup() {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	now := time.Now()
	
	// 清理过期静默
	for id, silence := range am.silences {
		if silence.IsExpired() {
			delete(am.silences, id)
		}
	}
	
	// 清理过期历史
	if am.config.HistoryRetention > 0 {
		cutoff := now.Add(-am.config.HistoryRetention)
		var newHistory []*models.AlertHistory
		for _, history := range am.alertHistory {
			if history.Timestamp.After(cutoff) {
				newHistory = append(newHistory, history)
			}
		}
		am.alertHistory = newHistory
	}
}

// generateAlertKey 生成告警键
func (am *AlertManager) generateAlertKey(rule *models.AlertRule, labels map[string]string) string {
	return fmt.Sprintf("%s:%s", rule.ID, am.generateFingerprint(labels))
}

// generateFingerprint 生成指纹
func (am *AlertManager) generateFingerprint(labels map[string]string) string {
	// 对标签进行排序以确保一致性
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	
	return fmt.Sprintf("%x", parts) // 这里应该使用哈希函数
}

// generateAlertID 生成告警ID
func (am *AlertManager) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// generateNotificationID 生成通知ID
func (am *AlertManager) generateNotificationID() string {
	return fmt.Sprintf("notification_%d", time.Now().UnixNano())
}

// generateHistoryID 生成历史ID
func (am *AlertManager) generateHistoryID() string {
	return fmt.Sprintf("history_%d", time.Now().UnixNano())
}

// 实现 interfaces.AlertManager 接口

// EvaluateRule 评估规则
func (am *AlertManager) EvaluateRule(ctx context.Context, rule *models.AlertRule) ([]*models.Alert, error) {
	result, err := am.evaluator.Evaluate(ctx, rule)
	if err != nil {
		return nil, err
	}
	
	var alerts []*models.Alert
	for _, alertResult := range result {
		if alertResult.Firing {
			alert := &models.Alert{
				ID:          am.generateAlertID(),
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Labels:      alertResult.Labels,
				Annotations: rule.Annotations,
				Status:      models.AlertStatusFiring,
				Severity:    rule.Severity,
				Value:       alertResult.Value,
				StartsAt:    time.Now(),
				UpdatedAt:   time.Now(),
				Fingerprint: am.generateFingerprint(alertResult.Labels),
			}
			alerts = append(alerts, alert)
		}
	}
	
	return alerts, nil
}

// CreateRule 创建规则
func (am *AlertManager) CreateRule(ctx context.Context, rule *models.AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if _, exists := am.rules[rule.ID]; exists {
		return fmt.Errorf("rule %s already exists", rule.ID)
	}
	
	am.rules[rule.ID] = rule
	am.stats.RulesCount = len(am.rules)
	
	return nil
}

// UpdateRule 更新规则
func (am *AlertManager) UpdateRule(ctx context.Context, rule *models.AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if _, exists := am.rules[rule.ID]; !exists {
		return fmt.Errorf("rule %s not found", rule.ID)
	}
	
	am.rules[rule.ID] = rule
	
	return nil
}

// DeleteRule 删除规则
func (am *AlertManager) DeleteRule(ctx context.Context, ruleID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if _, exists := am.rules[ruleID]; !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}
	
	delete(am.rules, ruleID)
	am.stats.RulesCount = len(am.rules)
	
	return nil
}

// GetRule 获取规则
func (am *AlertManager) GetRule(ctx context.Context, ruleID string) (*models.AlertRule, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	rule, exists := am.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}
	
	return rule, nil
}

// ListRules 列出规则
func (am *AlertManager) ListRules(ctx context.Context) ([]*models.AlertRule, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	rules := make([]*models.AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// FireAlert 触发告警
func (am *AlertManager) FireAlert(ctx context.Context, alert *models.Alert) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	alertKey := fmt.Sprintf("%s:%s", alert.RuleID, alert.Fingerprint)
	am.activeAlerts[alertKey] = alert
	am.stats.ActiveAlertsCount = len(am.activeAlerts)
	
	// 发送通知
	am.sendNotification(alert, models.AlertStatusFiring)
	
	return nil
}

// ResolveAlert 解决告警
func (am *AlertManager) ResolveAlert(ctx context.Context, alertID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	// 查找告警
	var alert *models.Alert
	var alertKey string
	for key, a := range am.activeAlerts {
		if a.ID == alertID {
			alert = a
			alertKey = key
			break
		}
	}
	
	if alert == nil {
		return fmt.Errorf("alert %s not found", alertID)
	}
	
	// 更新状态
	alert.Status = models.AlertStatusResolved
	alert.EndsAt = time.Now()
	alert.UpdatedAt = time.Now()
	
	// 发送通知
	am.sendNotification(alert, models.AlertStatusResolved)
	
	// 从活跃告警中移除
	delete(am.activeAlerts, alertKey)
	am.stats.ActiveAlertsCount = len(am.activeAlerts)
	
	return nil
}

// AcknowledgeAlert 确认告警
func (am *AlertManager) AcknowledgeAlert(ctx context.Context, alertID string, user string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	// 查找告警
	var alert *models.Alert
	for _, a := range am.activeAlerts {
		if a.ID == alertID {
			alert = a
			break
		}
	}
	
	if alert == nil {
		return fmt.Errorf("alert %s not found", alertID)
	}
	
	// 更新状态
	alert.Status = models.AlertStatusAcknowledged
	alert.UpdatedAt = time.Now()
	
	// 记录历史
	am.recordAlertHistory(alert, models.AlertStatusAcknowledged, fmt.Sprintf("Acknowledged by %s", user))
	
	return nil
}

// GetAlert 获取告警
func (am *AlertManager) GetAlert(ctx context.Context, alertID string) (*models.Alert, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	for _, alert := range am.activeAlerts {
		if alert.ID == alertID {
			return alert, nil
		}
	}
	
	return nil, fmt.Errorf("alert %s not found", alertID)
}

// ListAlerts 列出告警
func (am *AlertManager) ListAlerts(ctx context.Context, filters map[string]string) ([]*models.Alert, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	var alerts []*models.Alert
	for _, alert := range am.activeAlerts {
		if am.matchFilters(alert, filters) {
			alerts = append(alerts, alert)
		}
	}
	
	return alerts, nil
}

// matchFilters 匹配过滤器
func (am *AlertManager) matchFilters(alert *models.Alert, filters map[string]string) bool {
	for key, value := range filters {
		switch key {
		case "status":
			if string(alert.Status) != value {
				return false
			}
		case "severity":
			if string(alert.Severity) != value {
				return false
			}
		case "rule_name":
			if alert.RuleName != value {
				return false
			}
		default:
			// 检查标签
			if labelValue, exists := alert.Labels[key]; !exists || labelValue != value {
				return false
			}
		}
	}
	return true
}

// CreateSilence 创建静默
func (am *AlertManager) CreateSilence(ctx context.Context, silence *models.Silence) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	am.silences[silence.ID] = silence
	am.stats.SilencesCount = len(am.silences)
	
	return nil
}

// DeleteSilence 删除静默
func (am *AlertManager) DeleteSilence(ctx context.Context, silenceID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if _, exists := am.silences[silenceID]; !exists {
		return fmt.Errorf("silence %s not found", silenceID)
	}
	
	delete(am.silences, silenceID)
	am.stats.SilencesCount = len(am.silences)
	
	return nil
}

// ListSilences 列出静默
func (am *AlertManager) ListSilences(ctx context.Context) ([]*models.Silence, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	silences := make([]*models.Silence, 0, len(am.silences))
	for _, silence := range am.silences {
		silences = append(silences, silence)
	}
	
	return silences, nil
}

// GetStats 获取统计信息
func (am *AlertManager) GetStats() *AlertManagerStats {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	am.stats.RulesCount = len(am.rules)
	am.stats.ActiveAlertsCount = len(am.activeAlerts)
	am.stats.SilencesCount = len(am.silences)
	am.stats.LastUpdateTime = time.Now()
	
	return am.stats
}

// IsRunning 检查是否运行中
func (am *AlertManager) IsRunning() bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.running
}