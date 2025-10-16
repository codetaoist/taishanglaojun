package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/audit"
)

// AuditMiddleware 
type AuditMiddleware struct {
	service audit.AuditService
	config  AuditMiddlewareConfig
	logger  *zap.Logger
}

// AuditMiddlewareConfig 
type AuditMiddlewareConfig struct {
	// 
	Enabled        bool     `json:"enabled"`
	SkipPaths      []string `json:"skip_paths"`
	SkipMethods    []string `json:"skip_methods"`
	SkipUserAgents []string `json:"skip_user_agents"`

	// 
	LogRequestBody  bool `json:"log_request_body"`
	LogResponseBody bool `json:"log_response_body"`
	MaxBodySize     int  `json:"max_body_size"`

	// 
	SensitiveFields []string `json:"sensitive_fields"`
	MaskSensitive   bool     `json:"mask_sensitive"`

	// 
	AsyncLogging bool `json:"async_logging"`

	// 
	FilterEnabled bool     `json:"filter_enabled"`
	FilterRules   []string `json:"filter_rules"`

	// 
	LogIPAddress bool `json:"log_ip_address"`
	LogUserAgent bool `json:"log_user_agent"`
	LogHeaders   bool `json:"log_headers"`

	// 
	ComplianceMode bool     `json:"compliance_mode"`
	ComplianceTags []string `json:"compliance_tags"`

	// 
	CustomFields map[string]string `json:"custom_fields"`

	// 
	ContinueOnError bool `json:"continue_on_error"`
}

// ResponseWriter 
type ResponseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	size   int64
}

// NewAuditMiddleware 
func NewAuditMiddleware(service audit.AuditService, config AuditMiddlewareConfig, logger *zap.Logger) *AuditMiddleware {
	// 
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 64 * 1024 // 64KB
	}

	return &AuditMiddleware{
		service: service,
		config:  config,
		logger:  logger,
	}
}

// Handler 
func (m *AuditMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// 
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		start := time.Now()

		// 
		writer := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
			status:         200,
		}
		c.Writer = writer

		// 
		var requestBody []byte
		if m.config.LogRequestBody && c.Request.Body != nil {
			requestBody = m.readRequestBody(c)
		}

		// 
		c.Next()

		// 
		m.logAuditEvent(c, writer, requestBody, start)
	}
}

// RequireAudit 
func (m *AuditMiddleware) RequireAudit(eventType audit.EventType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		start := time.Now()

		// 
		writer := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
			status:         200,
		}
		c.Writer = writer

		// 
		var requestBody []byte
		if m.config.LogRequestBody && c.Request.Body != nil {
			requestBody = m.readRequestBody(c)
		}

		// 
		c.Next()

		// 
		m.logSpecificAuditEvent(c, writer, requestBody, start, eventType, action)
	}
}

// LogEvent 
func (m *AuditMiddleware) LogEvent(c *gin.Context, eventType audit.EventType, action string, details map[string]interface{}) {
	if !m.config.Enabled {
		return
	}

	event := m.createBaseEvent(c, eventType, action)

	// 
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}
	for key, value := range details {
		event.Metadata[key] = value
	}

	// 
	m.recordEvent(c.Request.Context(), event)
}

// 

// shouldSkip 
func (m *AuditMiddleware) shouldSkip(c *gin.Context) bool {
	// 
	for _, path := range m.config.SkipPaths {
		if strings.HasPrefix(c.Request.URL.Path, path) {
			return true
		}
	}

	// 鷽
	for _, method := range m.config.SkipMethods {
		if c.Request.Method == method {
			return true
		}
	}

	// User-Agent
	userAgent := c.Request.UserAgent()
	for _, ua := range m.config.SkipUserAgents {
		if strings.Contains(userAgent, ua) {
			return true
		}
	}

	// 
	if m.config.FilterEnabled {
		return m.applyFilterRules(c)
	}

	return false
}

// applyFilterRules 
func (m *AuditMiddleware) applyFilterRules(c *gin.Context) bool {
	// 
	// false
	return false
}

// readRequestBody 
func (m *AuditMiddleware) readRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}

	// 
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, int64(m.config.MaxBodySize)))
	if err != nil {
		m.logger.Warn("Failed to read request body", zap.Error(err))
		return nil
	}

	// 
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	return body
}

// logAuditEvent 
func (m *AuditMiddleware) logAuditEvent(c *gin.Context, writer *ResponseWriter, requestBody []byte, start time.Time) {
	// 
	eventType, action := m.determineEventTypeAndAction(c)

	// 
	event := m.createAuditEvent(c, writer, requestBody, start, eventType, action)

	// 
	m.recordEvent(c.Request.Context(), event)
}

// logSpecificAuditEvent 
func (m *AuditMiddleware) logSpecificAuditEvent(c *gin.Context, writer *ResponseWriter, requestBody []byte, start time.Time, eventType audit.EventType, action string) {
	// 
	event := m.createAuditEvent(c, writer, requestBody, start, eventType, action)

	// 
	m.recordEvent(c.Request.Context(), event)
}

// determineEventTypeAndAction 
func (m *AuditMiddleware) determineEventTypeAndAction(c *gin.Context) (audit.EventType, string) {
	method := c.Request.Method
	path := c.Request.URL.Path
	status := c.Writer.Status()

	// HTTP
	var eventType audit.EventType
	var action string

	switch {
	case strings.Contains(path, "/auth/login"):
		eventType = audit.EventTypeUserLogin
		action = "user_login"
	case strings.Contains(path, "/auth/logout"):
		eventType = audit.EventTypeUserLogout
		action = "user_logout"
	case strings.Contains(path, "/users") && method == "POST":
		eventType = audit.EventTypeUserCreate
		action = "user_create"
	case strings.Contains(path, "/users") && method == "PUT":
		eventType = audit.EventTypeUserUpdate
		action = "user_update"
	case strings.Contains(path, "/users") && method == "DELETE":
		eventType = audit.EventTypeUserDelete
		action = "user_delete"
	case method == "GET":
		eventType = audit.EventTypeDataAccess
		action = "data_read"
	case method == "POST":
		eventType = audit.EventTypeDataCreate
		action = "data_create"
	case method == "PUT" || method == "PATCH":
		eventType = audit.EventTypeDataUpdate
		action = "data_update"
	case method == "DELETE":
		eventType = audit.EventTypeDataDelete
		action = "data_delete"
	default:
		eventType = audit.EventTypeAPICall
		action = fmt.Sprintf("api_%s", strings.ToLower(method))
	}

	// 
	if status >= 400 {
		action += "_failed"
	} else {
		action += "_success"
	}

	return eventType, action
}

// createAuditEvent 
func (m *AuditMiddleware) createAuditEvent(c *gin.Context, writer *ResponseWriter, requestBody []byte, start time.Time, eventType audit.EventType, action string) *audit.AuditEvent {
	event := m.createBaseEvent(c, eventType, action)

	// 
	event.RequestMethod = c.Request.Method
	event.RequestURL = c.Request.URL.String()

	// 
	event.ResponseStatus = writer.status
	event.ResponseSize = writer.size

	// 
	duration := time.Since(start)
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}
	event.Metadata["duration_ms"] = duration.Milliseconds()

	// 
	if m.config.LogRequestBody && len(requestBody) > 0 {
		event.Metadata["request_body"] = m.processBody(requestBody)
	}

	// 
	if m.config.LogResponseBody && writer.body.Len() > 0 {
		event.Metadata["response_body"] = m.processBody(writer.body.Bytes())
	}

	// 
	if m.config.LogHeaders {
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		event.Metadata["request_headers"] = m.maskSensitiveHeaders(headers)
	}

	// 
	if len(c.Request.URL.RawQuery) > 0 {
		event.Metadata["query_params"] = c.Request.URL.RawQuery
	}

	// 
	if len(c.Params) > 0 {
		params := make(map[string]string)
		for _, param := range c.Params {
			params[param.Key] = param.Value
		}
		event.Metadata["path_params"] = params
	}

	// 
	for key, value := range m.config.CustomFields {
		event.Metadata[key] = value
	}

	// 
	if m.config.ComplianceMode {
		event.ComplianceTags = append(event.ComplianceTags, m.config.ComplianceTags...)
	}

	// 
	event.SecurityLevel = m.determineSecurityLevel(c, writer.status)

	// 
	event.RiskScore = m.calculateRiskScore(c, writer.status)

	return event
}

// createBaseEvent 
func (m *AuditMiddleware) createBaseEvent(c *gin.Context, eventType audit.EventType, action string) *audit.AuditEvent {
	event := &audit.AuditEvent{
		ID:            audit.GenerateEventID(),
		Timestamp:     time.Now(),
		EventType:     eventType,
		EventAction:   action,
		EventCategory: "HTTP_REQUEST",
	}

	// 
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			event.UserID = uid
		}
	}

	if userName, exists := c.Get("user_name"); exists {
		if name, ok := userName.(string); ok {
			event.UserName = name
		}
	}

	if userEmail, exists := c.Get("user_email"); exists {
		if email, ok := userEmail.(string); ok {
			event.UserEmail = email
		}
	}

	if userRole, exists := c.Get("user_role"); exists {
		if role, ok := userRole.(string); ok {
			event.UserRole = role
		}
	}

	// 
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			event.TenantID = tid
		}
	}

	if tenantName, exists := c.Get("tenant_name"); exists {
		if name, ok := tenantName.(string); ok {
			event.TenantName = name
		}
	}

	// 
	if requestID, exists := c.Get("request_id"); exists {
		if rid, ok := requestID.(string); ok {
			event.RequestID = rid
		}
	}

	if sessionID, exists := c.Get("session_id"); exists {
		if sid, ok := sessionID.(string); ok {
			event.SessionID = sid
		}
	}

	if correlationID, exists := c.Get("correlation_id"); exists {
		if cid, ok := correlationID.(string); ok {
			event.CorrelationID = cid
		}
	}

	// 
	if m.config.LogIPAddress {
		event.IPAddress = m.getClientIP(c)
	}

	if m.config.LogUserAgent {
		event.UserAgent = c.Request.UserAgent()
	}

	// 
	event.SourceSystem = "taishanglaojun"
	event.SourceComponent = "audit-middleware"

	return event
}

// getClientIP IP
func (m *AuditMiddleware) getClientIP(c *gin.Context) string {
	// X-Forwarded-For
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// X-Real-IP
	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// RemoteAddr
	return c.ClientIP()
}

// processBody /
func (m *AuditMiddleware) processBody(body []byte) interface{} {
	if len(body) == 0 {
		return nil
	}

	// JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		// 
		if m.config.MaskSensitive {
			return m.maskSensitiveData(jsonData)
		}
		return jsonData
	}

	// JSON
	bodyStr := string(body)
	if m.config.MaskSensitive {
		return m.maskSensitiveString(bodyStr)
	}

	return bodyStr
}

// maskSensitiveData 
func (m *AuditMiddleware) maskSensitiveData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if m.isSensitiveField(key) {
				result[key] = "***MASKED***"
			} else {
				result[key] = m.maskSensitiveData(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = m.maskSensitiveData(item)
		}
		return result
	default:
		return v
	}
}

// maskSensitiveString 
func (m *AuditMiddleware) maskSensitiveString(str string) string {
	for _, field := range m.config.SensitiveFields {
		if strings.Contains(strings.ToLower(str), strings.ToLower(field)) {
			return "***MASKED***"
		}
	}
	return str
}

// maskSensitiveHeaders 
func (m *AuditMiddleware) maskSensitiveHeaders(headers map[string]string) map[string]string {
	result := make(map[string]string)
	sensitiveHeaders := []string{"authorization", "cookie", "x-api-key", "x-auth-token"}

	for key, value := range headers {
		lowerKey := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveHeaders {
			if strings.Contains(lowerKey, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive || m.isSensitiveField(key) {
			result[key] = "***MASKED***"
		} else {
			result[key] = value
		}
	}

	return result
}

// isSensitiveField 
func (m *AuditMiddleware) isSensitiveField(field string) bool {
	lowerField := strings.ToLower(field)
	for _, sensitive := range m.config.SensitiveFields {
		if strings.Contains(lowerField, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// determineSecurityLevel 
func (m *AuditMiddleware) determineSecurityLevel(c *gin.Context, status int) audit.SecurityLevel {
	// 
	path := c.Request.URL.Path

	switch {
	case strings.Contains(path, "/auth") || strings.Contains(path, "/login"):
		return audit.SecurityLevelHigh
	case status >= 400:
		return audit.SecurityLevelMedium
	case strings.Contains(path, "/admin"):
		return audit.SecurityLevelHigh
	default:
		return audit.SecurityLevelLow
	}
}

// calculateRiskScore 
func (m *AuditMiddleware) calculateRiskScore(c *gin.Context, status int) float64 {
	score := 0.0

	// 
	if status >= 500 {
		score += 0.8
	} else if status >= 400 {
		score += 0.5
	}

	// 
	path := c.Request.URL.Path
	if strings.Contains(path, "/admin") {
		score += 0.3
	}
	if strings.Contains(path, "/auth") {
		score += 0.2
	}

	// 
	if c.Request.Method == "DELETE" {
		score += 0.3
	} else if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		score += 0.1
	}

	// 0-1
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// recordEvent 
func (m *AuditMiddleware) recordEvent(ctx context.Context, event *audit.AuditEvent) {
	if m.config.AsyncLogging {
		// 
		go func() {
			if err := m.service.LogEvent(ctx, event); err != nil {
				m.logger.Error("Failed to log audit event",
					zap.String("event_id", event.ID),
					zap.Error(err))
			}
		}()
	} else {
		// 
		if err := m.service.LogEvent(ctx, event); err != nil {
			m.logger.Error("Failed to log audit event",
				zap.String("event_id", event.ID),
				zap.Error(err))

			if !m.config.ContinueOnError {
				// 
			}
		}
	}
}

// ResponseWriter 

func (w *ResponseWriter) Write(data []byte) (int, error) {
	// 
	if w.body != nil {
		w.body.Write(data)
	}

	// 
	n, err := w.ResponseWriter.Write(data)
	w.size += int64(n)
	return n, err
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// 

// GetAuditEvent 
func GetAuditEvent(c *gin.Context) (*audit.AuditEvent, bool) {
	if event, exists := c.Get("audit_event"); exists {
		if auditEvent, ok := event.(*audit.AuditEvent); ok {
			return auditEvent, true
		}
	}
	return nil, false
}

// SetAuditEvent 
func SetAuditEvent(c *gin.Context, event *audit.AuditEvent) {
	c.Set("audit_event", event)
}

// AddAuditMetadata 
func AddAuditMetadata(c *gin.Context, key string, value interface{}) {
	if event, exists := GetAuditEvent(c); exists {
		if event.Metadata == nil {
			event.Metadata = make(map[string]interface{})
		}
		event.Metadata[key] = value
	}
}

// SetAuditResource 
func SetAuditResource(c *gin.Context, resourceID, resourceType, resourceName string) {
	if event, exists := GetAuditEvent(c); exists {
		event.ResourceID = resourceID
		event.ResourceType = resourceType
		event.ResourceName = resourceName
	}
}

