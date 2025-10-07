package notifications

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/models"
)

// RateLimiter 限流器
type RateLimiter struct {
	rate       int           // 每个周期允许的请求数
	period     time.Duration // 周期长度
	burstLimit int           // 突发限制
	
	tokens     int           // 当前令牌数
	lastRefill time.Time     // 上次补充令牌时间
	mutex      sync.Mutex    // 同步锁
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate int, period time.Duration, burstLimit int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		period:     period,
		burstLimit: burstLimit,
		tokens:     burstLimit,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	// 计算需要补充的令牌数
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed.Nanoseconds() * int64(rl.rate) / int64(rl.period.Nanoseconds()))
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.burstLimit {
			rl.tokens = rl.burstLimit
		}
		rl.lastRefill = now
	}
	
	// 检查是否有可用令牌
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

// GetTokens 获取当前令牌数
func (rl *RateLimiter) GetTokens() int {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.tokens
}

// Reset 重置限流器
func (rl *RateLimiter) Reset() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	rl.tokens = rl.burstLimit
	rl.lastRefill = time.Now()
}

// TemplateEngine 模板引擎
type TemplateEngine struct {
	templateDir string
	templates   map[string]*template.Template
	mutex       sync.RWMutex
}

// NewTemplateEngine 创建模板引擎
func NewTemplateEngine(templateDir string) *TemplateEngine {
	return &TemplateEngine{
		templateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}
}

// Initialize 初始化模板引擎
func (te *TemplateEngine) Initialize() error {
	// 加载默认模板
	te.loadDefaultTemplates()
	
	// 如果指定了模板目录，加载自定义模板
	if te.templateDir != "" {
		return te.loadCustomTemplates()
	}
	
	return nil
}

// loadDefaultTemplates 加载默认模板
func (te *TemplateEngine) loadDefaultTemplates() {
	// 默认邮件模板
	emailTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: {{.SeverityColor}}; color: white; padding: 10px; border-radius: 5px; }
        .content { margin: 20px 0; }
        .labels { background-color: #f5f5f5; padding: 10px; border-radius: 5px; }
        .label { margin: 5px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h2>{{.Title}}</h2>
        <p>严重级别: {{.Severity}}</p>
    </div>
    <div class="content">
        <p>{{.Message}}</p>
        <p><strong>时间:</strong> {{.CreatedAt.Format "2006-01-02 15:04:05"}}</p>
        {{if .Labels}}
        <div class="labels">
            <h3>标签:</h3>
            {{range $key, $value := .Labels}}
            <div class="label"><strong>{{$key}}:</strong> {{$value}}</div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>
`
	
	// 默认文本模板
	textTemplate := `
告警: {{.Title}}
严重级别: {{.Severity}}
消息: {{.Message}}
时间: {{.CreatedAt.Format "2006-01-02 15:04:05"}}
{{if .Labels}}
标签:
{{range $key, $value := .Labels}}  {{$key}}: {{$value}}
{{end}}{{end}}
`
	
	// 默认Slack模板
	slackTemplate := `
{
    "text": "{{.Title}}",
    "attachments": [
        {
            "color": "{{.SeverityColor}}",
            "title": "{{.Title}}",
            "text": "{{.Message}}",
            "fields": [
                {
                    "title": "严重级别",
                    "value": "{{.Severity}}",
                    "short": true
                },
                {
                    "title": "时间",
                    "value": "{{.CreatedAt.Format "2006-01-02 15:04:05"}}",
                    "short": true
                }
                {{range $key, $value := .Labels}},
                {
                    "title": "{{$key}}",
                    "value": "{{$value}}",
                    "short": true
                }
                {{end}}
            ]
        }
    ]
}
`
	
	// 解析并存储模板
	te.mutex.Lock()
	defer te.mutex.Unlock()
	
	if tmpl, err := template.New("email").Parse(emailTemplate); err == nil {
		te.templates["email"] = tmpl
	}
	
	if tmpl, err := template.New("text").Parse(textTemplate); err == nil {
		te.templates["text"] = tmpl
	}
	
	if tmpl, err := template.New("slack").Parse(slackTemplate); err == nil {
		te.templates["slack"] = tmpl
	}
}

// loadCustomTemplates 加载自定义模板
func (te *TemplateEngine) loadCustomTemplates() error {
	// 这里可以实现从文件系统加载自定义模板的逻辑
	// 暂时返回nil
	return nil
}

// Render 渲染模板
func (te *TemplateEngine) Render(templateContent string, data interface{}) (string, error) {
	// 如果模板内容是模板名称，使用预定义模板
	te.mutex.RLock()
	if tmpl, exists := te.templates[templateContent]; exists {
		te.mutex.RUnlock()
		return te.renderTemplate(tmpl, data)
	}
	te.mutex.RUnlock()
	
	// 否则将内容作为模板解析
	tmpl, err := template.New("custom").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	return te.renderTemplate(tmpl, data)
}

// renderTemplate 渲染模板
func (te *TemplateEngine) renderTemplate(tmpl *template.Template, data interface{}) (string, error) {
	// 扩展数据，添加辅助函数和变量
	extendedData := te.extendTemplateData(data)
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, extendedData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// extendTemplateData 扩展模板数据
func (te *TemplateEngine) extendTemplateData(data interface{}) interface{} {
	// 如果是通知对象，添加额外的字段
	if notification, ok := data.(*models.Notification); ok {
		return struct {
			*models.Notification
			SeverityColor string
		}{
			Notification:  notification,
			SeverityColor: te.getSeverityColor(notification.Severity),
		}
	}
	
	return data
}

// getSeverityColor 获取严重级别对应的颜色
func (te *TemplateEngine) getSeverityColor(severity models.Severity) string {
	switch severity {
	case models.SeverityCritical:
		return "#FF0000" // 红色
	case models.SeverityHigh:
		return "#FF8C00" // 橙色
	case models.SeverityMedium:
		return "#FFD700" // 黄色
	case models.SeverityLow:
		return "#32CD32" // 绿色
	case models.SeverityInfo:
		return "#87CEEB" // 蓝色
	default:
		return "#808080" // 灰色
	}
}

// AddTemplate 添加模板
func (te *TemplateEngine) AddTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	
	te.mutex.Lock()
	defer te.mutex.Unlock()
	
	te.templates[name] = tmpl
	return nil
}

// RemoveTemplate 移除模板
func (te *TemplateEngine) RemoveTemplate(name string) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	
	delete(te.templates, name)
}

// ListTemplates 列出所有模板
func (te *TemplateEngine) ListTemplates() []string {
	te.mutex.RLock()
	defer te.mutex.RUnlock()
	
	var names []string
	for name := range te.templates {
		names = append(names, name)
	}
	
	return names
}

// NotificationFormatter 通知格式化器
type NotificationFormatter struct {
	templateEngine *TemplateEngine
}

// NewNotificationFormatter 创建通知格式化器
func NewNotificationFormatter(templateEngine *TemplateEngine) *NotificationFormatter {
	return &NotificationFormatter{
		templateEngine: templateEngine,
	}
}

// FormatForEmail 格式化为邮件
func (nf *NotificationFormatter) FormatForEmail(notification *models.Notification) (string, string, error) {
	// 格式化标题
	subject := notification.Title
	if subject == "" {
		subject = fmt.Sprintf("[%s] 监控告警", notification.Severity)
	}
	
	// 格式化内容
	body, err := nf.templateEngine.Render("email", notification)
	if err != nil {
		// 如果模板渲染失败，使用简单格式
		body = nf.formatSimpleText(notification)
	}
	
	return subject, body, nil
}

// FormatForSlack 格式化为Slack
func (nf *NotificationFormatter) FormatForSlack(notification *models.Notification) (string, error) {
	return nf.templateEngine.Render("slack", notification)
}

// FormatForText 格式化为文本
func (nf *NotificationFormatter) FormatForText(notification *models.Notification) (string, error) {
	return nf.templateEngine.Render("text", notification)
}

// formatSimpleText 格式化为简单文本
func (nf *NotificationFormatter) formatSimpleText(notification *models.Notification) string {
	text := fmt.Sprintf("告警: %s\n", notification.Title)
	text += fmt.Sprintf("严重级别: %s\n", notification.Severity)
	text += fmt.Sprintf("消息: %s\n", notification.Message)
	text += fmt.Sprintf("时间: %s\n", notification.CreatedAt.Format("2006-01-02 15:04:05"))
	
	if len(notification.Labels) > 0 {
		text += "标签:\n"
		for k, v := range notification.Labels {
			text += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}
	
	return text
}

// NotificationBatcher 通知批处理器
type NotificationBatcher struct {
	batchSize    int
	batchTimeout time.Duration
	buffer       []*models.Notification
	timer        *time.Timer
	mutex        sync.Mutex
	callback     func([]*models.Notification)
}

// NewNotificationBatcher 创建通知批处理器
func NewNotificationBatcher(batchSize int, batchTimeout time.Duration, callback func([]*models.Notification)) *NotificationBatcher {
	return &NotificationBatcher{
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		buffer:       make([]*models.Notification, 0, batchSize),
		callback:     callback,
	}
}

// Add 添加通知到批处理器
func (nb *NotificationBatcher) Add(notification *models.Notification) {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	
	nb.buffer = append(nb.buffer, notification)
	
	// 如果达到批处理大小，立即处理
	if len(nb.buffer) >= nb.batchSize {
		nb.flush()
		return
	}
	
	// 如果是第一个通知，启动定时器
	if len(nb.buffer) == 1 {
		nb.timer = time.AfterFunc(nb.batchTimeout, func() {
			nb.mutex.Lock()
			defer nb.mutex.Unlock()
			nb.flush()
		})
	}
}

// flush 刷新缓冲区
func (nb *NotificationBatcher) flush() {
	if len(nb.buffer) == 0 {
		return
	}
	
	// 停止定时器
	if nb.timer != nil {
		nb.timer.Stop()
		nb.timer = nil
	}
	
	// 处理批次
	batch := make([]*models.Notification, len(nb.buffer))
	copy(batch, nb.buffer)
	nb.buffer = nb.buffer[:0]
	
	// 异步处理
	go nb.callback(batch)
}

// Flush 手动刷新
func (nb *NotificationBatcher) Flush() {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	nb.flush()
}

// Size 获取当前缓冲区大小
func (nb *NotificationBatcher) Size() int {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	return len(nb.buffer)
}