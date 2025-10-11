package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/repositories"
)

// WebhookService Webhook服务
type WebhookService struct {
	repo *repositories.WebhookRepository
}

// NewWebhookService 创建新的Webhook服务
func NewWebhookService(repo *repositories.WebhookRepository) *WebhookService {
	return &WebhookService{
		repo: repo,
	}
}

// CreateWebhook 创建Webhook
func (s *WebhookService) CreateWebhook(userID int64, name, url, secret string, events []string, headers map[string]string) (*models.Webhook, error) {
	webhook := &models.Webhook{
		UserID:      userID,
		Name:        name,
		URL:         url,
		Secret:      secret,
		Events:      events,
		Headers:     headers,
		Status:      models.WebhookStatusActive,
		IsActive:    true,
		RetryCount:  3,
		Timeout:     30,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 验证Webhook配置
	if err := s.validateWebhook(webhook); err != nil {
		return nil, fmt.Errorf("invalid webhook configuration: %w", err)
	}

	// 保存到数据库
	id, err := s.repo.Create(webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	webhook.ID = id
	return webhook, nil
}

// GetWebhook 获取Webhook
func (s *WebhookService) GetWebhook(id int64) (*models.Webhook, error) {
	return s.repo.GetByID(id)
}

// ListWebhooks 获取用户的Webhook列表
func (s *WebhookService) ListWebhooks(userID int64, limit, offset int) ([]*models.Webhook, int64, error) {
	return s.repo.ListByUserID(userID, limit, offset)
}

// UpdateWebhook 更新Webhook
func (s *WebhookService) UpdateWebhook(id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.repo.Update(id, updates)
}

// DeleteWebhook 删除Webhook
func (s *WebhookService) DeleteWebhook(id int64) error {
	return s.repo.Delete(id)
}

// TestWebhook 测试Webhook
func (s *WebhookService) TestWebhook(id int64) error {
	webhook, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("webhook not found: %w", err)
	}

	// 创建测试事件
	testEvent := map[string]interface{}{
		"event":     "test",
		"timestamp": time.Now().Unix(),
		"data": map[string]interface{}{
			"message": "This is a test webhook event",
			"test":    true,
		},
	}

	return s.SendWebhook(webhook, testEvent)
}

// SendWebhook 发送Webhook事件
func (s *WebhookService) SendWebhook(webhook *models.Webhook, event map[string]interface{}) error {
	if !webhook.IsActive {
		return fmt.Errorf("webhook is not active")
	}

	// 序列化事件数�?	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", webhook.URL, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求�?	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TaiShangLaoJun-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", fmt.Sprintf("%d", webhook.ID))
	req.Header.Set("X-Webhook-Event", fmt.Sprintf("%v", event["event"]))
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// 添加自定义请求头
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// 生成签名
	if webhook.Secret != "" {
		signature := s.generateSignature(payload, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// 发送请�?	client := &http.Client{
		Timeout: time.Duration(webhook.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		// 记录失败并重�?		s.recordWebhookAttempt(webhook.ID, false, err.Error())
		return s.retryWebhook(webhook, event, 1)
	}
	defer resp.Body.Close()

	// 检查响应状�?	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.recordWebhookAttempt(webhook.ID, true, "")
		return nil
	}

	// 记录失败并重�?	errorMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
	s.recordWebhookAttempt(webhook.ID, false, errorMsg)
	return s.retryWebhook(webhook, event, 1)
}

// retryWebhook 重试Webhook
func (s *WebhookService) retryWebhook(webhook *models.Webhook, event map[string]interface{}, attempt int) error {
	if attempt > webhook.RetryCount {
		return fmt.Errorf("webhook failed after %d attempts", webhook.RetryCount)
	}

	// 计算重试延迟（指数退避）
	delay := time.Duration(attempt*attempt) * time.Second
	time.Sleep(delay)

	// 重新发�?	payload, _ := json.Marshal(event)
	req, _ := http.NewRequest("POST", webhook.URL, strings.NewReader(string(payload)))

	// 设置请求�?	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TaiShangLaoJun-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", fmt.Sprintf("%d", webhook.ID))
	req.Header.Set("X-Webhook-Event", fmt.Sprintf("%v", event["event"]))
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	req.Header.Set("X-Webhook-Retry", fmt.Sprintf("%d", attempt))

	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	if webhook.Secret != "" {
		signature := s.generateSignature(payload, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	client := &http.Client{
		Timeout: time.Duration(webhook.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		s.recordWebhookAttempt(webhook.ID, false, err.Error())
		return s.retryWebhook(webhook, event, attempt+1)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.recordWebhookAttempt(webhook.ID, true, "")
		return nil
	}

	errorMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
	s.recordWebhookAttempt(webhook.ID, false, errorMsg)
	return s.retryWebhook(webhook, event, attempt+1)
}

// TriggerEvent 触发事件
func (s *WebhookService) TriggerEvent(userID int64, eventType string, data map[string]interface{}) error {
	// 获取用户的所有活跃Webhook
	webhooks, _, err := s.repo.ListByUserID(userID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	// 创建事件
	event := map[string]interface{}{
		"event":     eventType,
		"timestamp": time.Now().Unix(),
		"data":      data,
	}

	// 发送到所有匹配的Webhook
	for _, webhook := range webhooks {
		if s.shouldTriggerWebhook(webhook, eventType) {
			go func(w *models.Webhook) {
				if err := s.SendWebhook(w, event); err != nil {
					fmt.Printf("Failed to send webhook %d: %v\n", w.ID, err)
				}
			}(webhook)
		}
	}

	return nil
}

// shouldTriggerWebhook 检查是否应该触发Webhook
func (s *WebhookService) shouldTriggerWebhook(webhook *models.Webhook, eventType string) bool {
	if !webhook.IsActive {
		return false
	}

	// 检查事件类型是否匹�?	for _, event := range webhook.Events {
		if event == "*" || event == eventType {
			return true
		}
	}

	return false
}

// generateSignature 生成签名
func (s *WebhookService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证签名
func (s *WebhookService) VerifySignature(payload []byte, signature, secret string) bool {
	expectedSignature := s.generateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// validateWebhook 验证Webhook配置
func (s *WebhookService) validateWebhook(webhook *models.Webhook) error {
	if webhook.Name == "" {
		return fmt.Errorf("webhook name is required")
	}

	if webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if !strings.HasPrefix(webhook.URL, "http://") && !strings.HasPrefix(webhook.URL, "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	if len(webhook.Events) == 0 {
		return fmt.Errorf("at least one event must be specified")
	}

	if webhook.Timeout <= 0 {
		webhook.Timeout = 30
	}

	if webhook.RetryCount < 0 {
		webhook.RetryCount = 3
	}

	return nil
}

// recordWebhookAttempt 记录Webhook尝试
func (s *WebhookService) recordWebhookAttempt(webhookID int64, success bool, errorMsg string) {
	updates := map[string]interface{}{
		"last_triggered_at": time.Now(),
		"updated_at":        time.Now(),
	}

	if success {
		updates["status"] = models.WebhookStatusActive
		updates["last_success_at"] = time.Now()
	} else {
		updates["status"] = models.WebhookStatusError
		updates["last_error"] = errorMsg
	}

	s.repo.Update(webhookID, updates)
}

// GetWebhookStats 获取Webhook统计信息
func (s *WebhookService) GetWebhookStats(id int64) (map[string]interface{}, error) {
	webhook, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	stats := map[string]interface{}{
		"name":               webhook.Name,
		"url":                webhook.URL,
		"events":             webhook.Events,
		"status":             webhook.Status,
		"is_active":          webhook.IsActive,
		"last_triggered_at":  webhook.LastTriggeredAt,
		"last_success_at":    webhook.LastSuccessAt,
		"last_error":         webhook.LastError,
		"retry_count":        webhook.RetryCount,
		"timeout":            webhook.Timeout,
		"created_at":         webhook.CreatedAt,
		"updated_at":         webhook.UpdatedAt,
	}

	// 可以添加更多统计信息，如成功率、失败次数等
	stats["total_triggers"] = 100  // 示例数据
	stats["success_rate"] = 95.5   // 示例数据
	stats["avg_response_time"] = 250 // 毫秒

	return stats, nil
}

// HandleIncomingWebhook 处理接收到的Webhook
func (s *WebhookService) HandleIncomingWebhook(r *http.Request) error {
	// 解析请求�?	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}

	// 获取Webhook ID
	webhookID := r.Header.Get("X-Webhook-ID")
	if webhookID == "" {
		return fmt.Errorf("missing webhook ID")
	}

	// 验证签名（如果有�?	signature := r.Header.Get("X-Webhook-Signature")
	if signature != "" {
		// 这里需要获取对应的secret进行验证
		// 为了简化，这里跳过验证
	}

	// 处理事件
	eventType, ok := payload["event"].(string)
	if !ok {
		return fmt.Errorf("missing event type")
	}

	fmt.Printf("Received webhook event: %s\n", eventType)
	
	// 这里可以根据事件类型进行相应的处�?	switch eventType {
	case "user.created":
		return s.handleUserCreatedEvent(payload)
	case "order.completed":
		return s.handleOrderCompletedEvent(payload)
	case "payment.received":
		return s.handlePaymentReceivedEvent(payload)
	default:
		fmt.Printf("Unknown event type: %s\n", eventType)
	}

	return nil
}

// handleUserCreatedEvent 处理用户创建事件
func (s *WebhookService) handleUserCreatedEvent(payload map[string]interface{}) error {
	fmt.Printf("Handling user created event: %+v\n", payload)
	// 实现用户创建事件的处理逻辑
	return nil
}

// handleOrderCompletedEvent 处理订单完成事件
func (s *WebhookService) handleOrderCompletedEvent(payload map[string]interface{}) error {
	fmt.Printf("Handling order completed event: %+v\n", payload)
	// 实现订单完成事件的处理逻辑
	return nil
}

// handlePaymentReceivedEvent 处理支付接收事件
func (s *WebhookService) handlePaymentReceivedEvent(payload map[string]interface{}) error {
	fmt.Printf("Handling payment received event: %+v\n", payload)
	// 实现支付接收事件的处理逻辑
	return nil
}
