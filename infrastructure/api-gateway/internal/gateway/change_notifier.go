package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
)

// LogChangeNotifier 日志变更通知器
type LogChangeNotifier struct {
	logger logger.Logger
}

// NewLogChangeNotifier 创建日志变更通知器
func NewLogChangeNotifier(log logger.Logger) *LogChangeNotifier {
	return &LogChangeNotifier{
		logger: log,
	}
}

// NotifyChange 通知路由变更到日志
func (lcn *LogChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	lcn.logger.WithFields(map[string]interface{}{
		"change_type": string(changeType),
		"method":      route.Method,
		"path":        route.Path,
		"service":     route.Service,
		"enabled":     route.Enabled,
	}).Info("Route change notification")
	
	return nil
}

// WebhookChangeNotifier Webhook变更通知器
type WebhookChangeNotifier struct {
	webhookURL string
	timeout    time.Duration
	logger     logger.Logger
	client     *http.Client
}

// WebhookPayload Webhook负载
type WebhookPayload struct {
	Timestamp  time.Time       `json:"timestamp"`
	ChangeType RouteChangeType `json:"change_type"`
	Route      *RouteConfig    `json:"route"`
	Source     string          `json:"source"`
}

// NewWebhookChangeNotifier 创建Webhook变更通知器
func NewWebhookChangeNotifier(webhookURL string, timeout time.Duration, log logger.Logger) *WebhookChangeNotifier {
	return &WebhookChangeNotifier{
		webhookURL: webhookURL,
		timeout:    timeout,
		logger:     log,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// NotifyChange 通知路由变更到Webhook
func (wcn *WebhookChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	payload := WebhookPayload{
		Timestamp:  time.Now(),
		ChangeType: changeType,
		Route:      route,
		Source:     "api-gateway",
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", wcn.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "api-gateway-notifier/1.0")
	
	resp, err := wcn.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}
	
	wcn.logger.WithFields(map[string]interface{}{
		"webhook_url": wcn.webhookURL,
		"change_type": string(changeType),
		"method":      route.Method,
		"path":        route.Path,
		"status_code": resp.StatusCode,
	}).Info("Route change webhook notification sent")
	
	return nil
}

// MetricsChangeNotifier 指标变更通知器
type MetricsChangeNotifier struct {
	logger logger.Logger
	// TODO: 添加Prometheus指标
}

// NewMetricsChangeNotifier 创建指标变更通知器
func NewMetricsChangeNotifier(log logger.Logger) *MetricsChangeNotifier {
	return &MetricsChangeNotifier{
		logger: log,
	}
}

// NotifyChange 通知路由变更到指标系统
func (mcn *MetricsChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	// TODO: 更新Prometheus指标
	mcn.logger.WithFields(map[string]interface{}{
		"change_type": string(changeType),
		"method":      route.Method,
		"path":        route.Path,
		"service":     route.Service,
	}).Debug("Route change metrics updated")
	
	return nil
}

// EventStoreChangeNotifier 事件存储变更通知器
type EventStoreChangeNotifier struct {
	logger logger.Logger
	// TODO: 添加事件存储客户端
}

// RouteChangeEvent 路由变更事件
type RouteChangeEvent struct {
	ID         string          `json:"id"`
	Timestamp  time.Time       `json:"timestamp"`
	ChangeType RouteChangeType `json:"change_type"`
	Route      *RouteConfig    `json:"route"`
	Source     string          `json:"source"`
	Version    int             `json:"version"`
}

// NewEventStoreChangeNotifier 创建事件存储变更通知器
func NewEventStoreChangeNotifier(log logger.Logger) *EventStoreChangeNotifier {
	return &EventStoreChangeNotifier{
		logger: log,
	}
}

// NotifyChange 通知路由变更到事件存储
func (escn *EventStoreChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	event := RouteChangeEvent{
		ID:         fmt.Sprintf("route-change-%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		ChangeType: changeType,
		Route:      route,
		Source:     "api-gateway",
		Version:    1,
	}
	
	// TODO: 存储事件到事件存储系统
	escn.logger.WithFields(map[string]interface{}{
		"event_id":    event.ID,
		"change_type": string(changeType),
		"method":      route.Method,
		"path":        route.Path,
	}).Debug("Route change event stored")
	
	return nil
}

// CompositeChangeNotifier 复合变更通知器
type CompositeChangeNotifier struct {
	notifiers []RouteChangeNotifier
	logger    logger.Logger
}

// NewCompositeChangeNotifier 创建复合变更通知器
func NewCompositeChangeNotifier(log logger.Logger) *CompositeChangeNotifier {
	return &CompositeChangeNotifier{
		notifiers: make([]RouteChangeNotifier, 0),
		logger:    log,
	}
}

// AddNotifier 添加通知器
func (ccn *CompositeChangeNotifier) AddNotifier(notifier RouteChangeNotifier) {
	ccn.notifiers = append(ccn.notifiers, notifier)
	ccn.logger.Info("Change notifier added to composite notifier")
}

// NotifyChange 通知路由变更到所有通知器
func (ccn *CompositeChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	var errors []error
	
	for _, notifier := range ccn.notifiers {
		if err := notifier.NotifyChange(ctx, changeType, route); err != nil {
			ccn.logger.Errorf("Failed to notify route change: %v", err)
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to notify %d notifiers", len(errors))
	}
	
	return nil
}

// AsyncChangeNotifier 异步变更通知器
type AsyncChangeNotifier struct {
	notifier RouteChangeNotifier
	logger   logger.Logger
	queue    chan notificationTask
	workers  int
	stopCh   chan struct{}
}

// notificationTask 通知任务
type notificationTask struct {
	ctx        context.Context
	changeType RouteChangeType
	route      *RouteConfig
}

// NewAsyncChangeNotifier 创建异步变更通知器
func NewAsyncChangeNotifier(notifier RouteChangeNotifier, workers int, log logger.Logger) *AsyncChangeNotifier {
	return &AsyncChangeNotifier{
		notifier: notifier,
		logger:   log,
		queue:    make(chan notificationTask, 1000), // 缓冲队列
		workers:  workers,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动异步通知器
func (acn *AsyncChangeNotifier) Start() {
	for i := 0; i < acn.workers; i++ {
		go acn.worker(i)
	}
	
	acn.logger.WithFields(map[string]interface{}{
		"workers": acn.workers,
	}).Info("Async change notifier started")
}

// Stop 停止异步通知器
func (acn *AsyncChangeNotifier) Stop() {
	close(acn.stopCh)
	acn.logger.Info("Async change notifier stopped")
}

// worker 工作协程
func (acn *AsyncChangeNotifier) worker(id int) {
	acn.logger.WithFields(map[string]interface{}{
		"worker_id": id,
	}).Debug("Async notification worker started")
	
	for {
		select {
		case <-acn.stopCh:
			return
		case task := <-acn.queue:
			if err := acn.notifier.NotifyChange(task.ctx, task.changeType, task.route); err != nil {
				acn.logger.WithFields(map[string]interface{}{
					"worker_id":   id,
					"change_type": string(task.changeType),
					"method":      task.route.Method,
					"path":        task.route.Path,
					"error":       err.Error(),
				}).Error("Failed to process notification task")
			}
		}
	}
}

// NotifyChange 异步通知路由变更
func (acn *AsyncChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	task := notificationTask{
		ctx:        ctx,
		changeType: changeType,
		route:      route,
	}
	
	select {
	case acn.queue <- task:
		return nil
	default:
		return fmt.Errorf("notification queue is full")
	}
}

// RetryChangeNotifier 重试变更通知器
type RetryChangeNotifier struct {
	notifier   RouteChangeNotifier
	maxRetries int
	retryDelay time.Duration
	logger     logger.Logger
}

// NewRetryChangeNotifier 创建重试变更通知器
func NewRetryChangeNotifier(notifier RouteChangeNotifier, maxRetries int, retryDelay time.Duration, log logger.Logger) *RetryChangeNotifier {
	return &RetryChangeNotifier{
		notifier:   notifier,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		logger:     log,
	}
}

// NotifyChange 带重试的通知路由变更
func (rcn *RetryChangeNotifier) NotifyChange(ctx context.Context, changeType RouteChangeType, route *RouteConfig) error {
	var lastErr error
	
	for i := 0; i <= rcn.maxRetries; i++ {
		if i > 0 {
			// 等待重试延迟
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rcn.retryDelay):
			}
		}
		
		err := rcn.notifier.NotifyChange(ctx, changeType, route)
		if err == nil {
			if i > 0 {
				rcn.logger.WithFields(map[string]interface{}{
					"retry_count": i,
					"change_type": string(changeType),
					"method":      route.Method,
					"path":        route.Path,
				}).Info("Route change notification succeeded after retry")
			}
			return nil
		}
		
		lastErr = err
		rcn.logger.WithFields(map[string]interface{}{
			"retry_count": i,
			"max_retries": rcn.maxRetries,
			"change_type": string(changeType),
			"method":      route.Method,
			"path":        route.Path,
			"error":       err.Error(),
		}).Warn("Route change notification failed, will retry")
	}
	
	return fmt.Errorf("failed to notify after %d retries: %w", rcn.maxRetries, lastErr)
}