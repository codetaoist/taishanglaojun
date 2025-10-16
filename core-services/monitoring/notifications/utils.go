package notifications

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// RateLimiter ?
type RateLimiter struct {
	rate       int           // 
	period     time.Duration // 
	burstLimit int           // 
	
	tokens     int           // ?
	lastRefill time.Time     // 
	mutex      sync.Mutex    // ?
}

// NewRateLimiter ?
func NewRateLimiter(rate int, period time.Duration, burstLimit int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		period:     period,
		burstLimit: burstLimit,
		tokens:     burstLimit,
		lastRefill: time.Now(),
	}
}

// Allow ?
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	// ?
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed.Nanoseconds() * int64(rl.rate) / int64(rl.period.Nanoseconds()))
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.burstLimit {
			rl.tokens = rl.burstLimit
		}
		rl.lastRefill = now
	}
	
	// 
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

// GetTokens ?
func (rl *RateLimiter) GetTokens() int {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.tokens
}

// Reset ?
func (rl *RateLimiter) Reset() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	rl.tokens = rl.burstLimit
	rl.lastRefill = time.Now()
}

// TemplateEngine 
type TemplateEngine struct {
	templateDir string
	templates   map[string]*template.Template
	mutex       sync.RWMutex
}

// NewTemplateEngine 
func NewTemplateEngine(templateDir string) *TemplateEngine {
	return &TemplateEngine{
		templateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}
}

// Initialize ?
func (te *TemplateEngine) Initialize() error {
	// 
	te.loadDefaultTemplates()
	
	// ?
	if te.templateDir != "" {
		return te.loadCustomTemplates()
	}
	
	return nil
}

// loadDefaultTemplates 
func (te *TemplateEngine) loadDefaultTemplates() {
	// 
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
        <p>: {{.Severity}}</p>
    </div>
    <div class="content">
        <p>{{.Message}}</p>
        <p><strong>:</strong> {{.CreatedAt.Format "2006-01-02 15:04:05"}}</p>
        {{if .Labels}}
        <div class="labels">
            <h3>:</h3>
            {{range $key, $value := .Labels}}
            <div class="label"><strong>{{$key}}:</strong> {{$value}}</div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>
`
	
	// 
	textTemplate := `
澯: {{.Title}}
: {{.Severity}}
: {{.Message}}
: {{.CreatedAt.Format "2006-01-02 15:04:05"}}
{{if .Labels}}
:
{{range $key, $value := .Labels}}  {{$key}}: {{$value}}
{{end}}{{end}}
`
	
	// Slack
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
                    "title": "",
                    "value": "{{.Severity}}",
                    "short": true
                },
                {
                    "title": "",
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
	
	// 洢?
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

// loadCustomTemplates ?
func (te *TemplateEngine) loadCustomTemplates() error {
	// 
	// nil
	return nil
}

// Render 
func (te *TemplateEngine) Render(templateContent string, data interface{}) (string, error) {
	// ?
	te.mutex.RLock()
	if tmpl, exists := te.templates[templateContent]; exists {
		te.mutex.RUnlock()
		return te.renderTemplate(tmpl, data)
	}
	te.mutex.RUnlock()
	
	// ?
	tmpl, err := template.New("custom").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	return te.renderTemplate(tmpl, data)
}

// renderTemplate 
func (te *TemplateEngine) renderTemplate(tmpl *template.Template, data interface{}) (string, error) {
	// 
	extendedData := te.extendTemplateData(data)
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, extendedData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// extendTemplateData 
func (te *TemplateEngine) extendTemplateData(data interface{}) interface{} {
	// 
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

// getSeverityColor ?
func (te *TemplateEngine) getSeverityColor(severity models.Severity) string {
	switch severity {
	case models.SeverityCritical:
		return "#FF0000" // 
	case models.SeverityHigh:
		return "#FF8C00" // 
	case models.SeverityMedium:
		return "#FFD700" // 
	case models.SeverityLow:
		return "#32CD32" // 
	case models.SeverityInfo:
		return "#87CEEB" // 
	default:
		return "#808080" // 
	}
}

// AddTemplate 
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

// RemoveTemplate 
func (te *TemplateEngine) RemoveTemplate(name string) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	
	delete(te.templates, name)
}

// ListTemplates ?
func (te *TemplateEngine) ListTemplates() []string {
	te.mutex.RLock()
	defer te.mutex.RUnlock()
	
	var names []string
	for name := range te.templates {
		names = append(names, name)
	}
	
	return names
}

// NotificationFormatter 
type NotificationFormatter struct {
	templateEngine *TemplateEngine
}

// NewNotificationFormatter 
func NewNotificationFormatter(templateEngine *TemplateEngine) *NotificationFormatter {
	return &NotificationFormatter{
		templateEngine: templateEngine,
	}
}

// FormatForEmail 
func (nf *NotificationFormatter) FormatForEmail(notification *models.Notification) (string, string, error) {
	// ?
	subject := notification.Title
	if subject == "" {
		subject = fmt.Sprintf("[%s] 澯", notification.Severity)
	}
	
	// ?
	body, err := nf.templateEngine.Render("email", notification)
	if err != nil {
		// ?
		body = nf.formatSimpleText(notification)
	}
	
	return subject, body, nil
}

// FormatForSlack Slack
func (nf *NotificationFormatter) FormatForSlack(notification *models.Notification) (string, error) {
	return nf.templateEngine.Render("slack", notification)
}

// FormatForText 
func (nf *NotificationFormatter) FormatForText(notification *models.Notification) (string, error) {
	return nf.templateEngine.Render("text", notification)
}

// formatSimpleText ?
func (nf *NotificationFormatter) formatSimpleText(notification *models.Notification) string {
	text := fmt.Sprintf("澯: %s\n", notification.Title)
	text += fmt.Sprintf(": %s\n", notification.Severity)
	text += fmt.Sprintf(": %s\n", notification.Message)
	text += fmt.Sprintf(": %s\n", notification.CreatedAt.Format("2006-01-02 15:04:05"))
	
	if len(notification.Labels) > 0 {
		text += ":\n"
		for k, v := range notification.Labels {
			text += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}
	
	return text
}

// NotificationBatcher 
type NotificationBatcher struct {
	batchSize    int
	batchTimeout time.Duration
	buffer       []*models.Notification
	timer        *time.Timer
	mutex        sync.Mutex
	callback     func([]*models.Notification)
}

// NewNotificationBatcher 
func NewNotificationBatcher(batchSize int, batchTimeout time.Duration, callback func([]*models.Notification)) *NotificationBatcher {
	return &NotificationBatcher{
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		buffer:       make([]*models.Notification, 0, batchSize),
		callback:     callback,
	}
}

// Add ?
func (nb *NotificationBatcher) Add(notification *models.Notification) {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	
	nb.buffer = append(nb.buffer, notification)
	
	// 
	if len(nb.buffer) >= nb.batchSize {
		nb.flush()
		return
	}
	
	// 
	if len(nb.buffer) == 1 {
		nb.timer = time.AfterFunc(nb.batchTimeout, func() {
			nb.mutex.Lock()
			defer nb.mutex.Unlock()
			nb.flush()
		})
	}
}

// flush ?
func (nb *NotificationBatcher) flush() {
	if len(nb.buffer) == 0 {
		return
	}
	
	// ?
	if nb.timer != nil {
		nb.timer.Stop()
		nb.timer = nil
	}
	
	// 
	batch := make([]*models.Notification, len(nb.buffer))
	copy(batch, nb.buffer)
	nb.buffer = nb.buffer[:0]
	
	// 
	go nb.callback(batch)
}

// Flush 
func (nb *NotificationBatcher) Flush() {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	nb.flush()
}

// Size ?
func (nb *NotificationBatcher) Size() int {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	return len(nb.buffer)
}

