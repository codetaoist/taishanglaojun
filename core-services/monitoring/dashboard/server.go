package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// DashboardServer 仪表板服务器
type DashboardServer struct {
	config          *DashboardConfig
	server          *http.Server
	router          *mux.Router
	upgrader        websocket.Upgrader
	storageManager  interfaces.StorageManager
	alertManager    interfaces.AlertManager
	wsConnections   map[string]*websocket.Conn
	wsConnMutex     sync.RWMutex
}

// DashboardConfig 仪表板配�?
type DashboardConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	
	// 静态文件配�?
	StaticDir    string `json:"static_dir" yaml:"static_dir"`
	TemplateDir  string `json:"template_dir" yaml:"template_dir"`
	
	// 安全配置
	EnableTLS    bool   `json:"enable_tls" yaml:"enable_tls"`
	CertFile     string `json:"cert_file" yaml:"cert_file"`
	KeyFile      string `json:"key_file" yaml:"key_file"`
	EnableAuth   bool   `json:"enable_auth" yaml:"enable_auth"`
	AuthToken    string `json:"auth_token" yaml:"auth_token"`
	
	// CORS配置
	EnableCORS   bool     `json:"enable_cors" yaml:"enable_cors"`
	AllowOrigins []string `json:"allow_origins" yaml:"allow_origins"`
	
	// WebSocket配置
	WSReadBufferSize  int           `json:"ws_read_buffer_size" yaml:"ws_read_buffer_size"`
	WSWriteBufferSize int           `json:"ws_write_buffer_size" yaml:"ws_write_buffer_size"`
	WSPingInterval    time.Duration `json:"ws_ping_interval" yaml:"ws_ping_interval"`
	WSPongTimeout     time.Duration `json:"ws_pong_timeout" yaml:"ws_pong_timeout"`
}

// NewDashboardServer 创建仪表板服务器
func NewDashboardServer(config *DashboardConfig, storageManager interfaces.StorageManager, alertManager interfaces.AlertManager) *DashboardServer {
	if config == nil {
		config = &DashboardConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			StaticDir:         "./static",
			TemplateDir:       "./templates",
			WSReadBufferSize:  1024,
			WSWriteBufferSize: 1024,
			WSPingInterval:    30 * time.Second,
			WSPongTimeout:     10 * time.Second,
		}
	}
	
	ds := &DashboardServer{
		config:         config,
		storageManager: storageManager,
		alertManager:   alertManager,
		wsConnections:  make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.WSReadBufferSize,
			WriteBufferSize: config.WSWriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				if !config.EnableCORS {
					return true
				}
				origin := r.Header.Get("Origin")
				for _, allowed := range config.AllowOrigins {
					if origin == allowed || allowed == "*" {
						return true
					}
				}
				return false
			},
		},
	}
	
	ds.setupRoutes()
	ds.setupServer()
	
	return ds
}

// setupRoutes 设置路由
func (ds *DashboardServer) setupRoutes() {
	ds.router = mux.NewRouter()
	
	// 中间�?
	ds.router.Use(ds.loggingMiddleware)
	if ds.config.EnableCORS {
		ds.router.Use(ds.corsMiddleware)
	}
	if ds.config.EnableAuth {
		ds.router.Use(ds.authMiddleware)
	}
	
	// API路由
	api := ds.router.PathPrefix("/api/v1").Subrouter()
	
	// 指标相关API
	api.HandleFunc("/metrics/query", ds.handleMetricsQuery).Methods("GET", "POST")
	api.HandleFunc("/metrics/query_range", ds.handleMetricsQueryRange).Methods("GET", "POST")
	api.HandleFunc("/metrics/labels", ds.handleMetricsLabels).Methods("GET")
	api.HandleFunc("/metrics/label/{name}/values", ds.handleMetricsLabelValues).Methods("GET")
	api.HandleFunc("/metrics/series", ds.handleMetricsSeries).Methods("GET", "POST")
	
	// 告警相关API
	api.HandleFunc("/alerts", ds.handleAlertsGet).Methods("GET")
	api.HandleFunc("/alerts", ds.handleAlertsCreate).Methods("POST")
	api.HandleFunc("/alerts/{id}", ds.handleAlertsUpdate).Methods("PUT")
	api.HandleFunc("/alerts/{id}", ds.handleAlertsDelete).Methods("DELETE")
	api.HandleFunc("/alerts/rules", ds.handleAlertRulesGet).Methods("GET")
	api.HandleFunc("/alerts/rules", ds.handleAlertRulesCreate).Methods("POST")
	api.HandleFunc("/alerts/rules/{id}", ds.handleAlertRulesUpdate).Methods("PUT")
	api.HandleFunc("/alerts/rules/{id}", ds.handleAlertRulesDelete).Methods("DELETE")
	
	// 静默相关API
	api.HandleFunc("/silences", ds.handleSilencesGet).Methods("GET")
	api.HandleFunc("/silences", ds.handleSilencesCreate).Methods("POST")
	api.HandleFunc("/silences/{id}", ds.handleSilencesDelete).Methods("DELETE")
	
	// 健康检查API
	api.HandleFunc("/health", ds.handleHealth).Methods("GET")
	api.HandleFunc("/status", ds.handleStatus).Methods("GET")
	
	// WebSocket
	ds.router.HandleFunc("/ws", ds.handleWebSocket)
	
	// 静态文�?
	if ds.config.StaticDir != "" {
		ds.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(ds.config.StaticDir))))
	}
	
	// 默认页面
	ds.router.HandleFunc("/", ds.handleIndex).Methods("GET")
	ds.router.HandleFunc("/dashboard", ds.handleDashboard).Methods("GET")
	ds.router.HandleFunc("/alerts", ds.handleAlertsPage).Methods("GET")
	ds.router.HandleFunc("/metrics", ds.handleMetricsPage).Methods("GET")
}

// setupServer 设置服务�?
func (ds *DashboardServer) setupServer() {
	addr := fmt.Sprintf("%s:%d", ds.config.Host, ds.config.Port)
	
	ds.server = &http.Server{
		Addr:         addr,
		Handler:      ds.router,
		ReadTimeout:  ds.config.ReadTimeout,
		WriteTimeout: ds.config.WriteTimeout,
		IdleTimeout:  ds.config.IdleTimeout,
	}
}

// Start 启动服务�?
func (ds *DashboardServer) Start() error {
	addr := fmt.Sprintf("%s:%d", ds.config.Host, ds.config.Port)
	
	if ds.config.EnableTLS {
		fmt.Printf("Dashboard server starting on https://%s\n", addr)
		return ds.server.ListenAndServeTLS(ds.config.CertFile, ds.config.KeyFile)
	} else {
		fmt.Printf("Dashboard server starting on http://%s\n", addr)
		return ds.server.ListenAndServe()
	}
}

// Stop 停止服务�?
func (ds *DashboardServer) Stop(ctx context.Context) error {
	// 关闭所有WebSocket连接
	ds.wsConnMutex.Lock()
	for id, conn := range ds.wsConnections {
		conn.Close()
		delete(ds.wsConnections, id)
	}
	ds.wsConnMutex.Unlock()
	
	return ds.server.Shutdown(ctx)
}

// 中间�?

// loggingMiddleware 日志中间�?
func (ds *DashboardServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		
		fmt.Printf("[%s] %s %s %v\n", 
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			duration,
		)
	})
}

// corsMiddleware CORS中间�?
func (ds *DashboardServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// 检查是否允许该来源
		allowed := false
		for _, allowedOrigin := range ds.config.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// authMiddleware 认证中间�?
func (ds *DashboardServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 跳过健康检查和静态文�?
		if strings.HasPrefix(r.URL.Path, "/health") || 
		   strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		
		// 检查Authorization�?
		auth := r.Header.Get("Authorization")
		if auth == "" {
			// 检查查询参�?
			token := r.URL.Query().Get("token")
			if token != "" {
				auth = "Bearer " + token
			}
		}
		
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		token := strings.TrimPrefix(auth, "Bearer ")
		if token != ds.config.AuthToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// API处理�?

// handleMetricsQuery 处理指标查询
func (ds *DashboardServer) handleMetricsQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}
	
	timeStr := r.URL.Query().Get("time")
	var queryTime time.Time
	if timeStr != "" {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			queryTime = t
		} else if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			queryTime = time.Unix(timestamp, 0)
		} else {
			http.Error(w, "Invalid time format", http.StatusBadRequest)
			return
		}
	} else {
		queryTime = time.Now()
	}
	
	result, err := ds.storageManager.QueryMetrics(context.Background(), query, queryTime)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "vector",
			"result":     result,
		},
	})
}

// handleMetricsQueryRange 处理指标范围查询
func (ds *DashboardServer) handleMetricsQueryRange(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}
	
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	stepStr := r.URL.Query().Get("step")
	
	if startStr == "" || endStr == "" {
		http.Error(w, "Missing start or end parameter", http.StatusBadRequest)
		return
	}
	
	start, err := ds.parseTime(startStr)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}
	
	end, err := ds.parseTime(endStr)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}
	
	step := time.Minute
	if stepStr != "" {
		if s, err := time.ParseDuration(stepStr); err == nil {
			step = s
		} else if seconds, err := strconv.ParseInt(stepStr, 10, 64); err == nil {
			step = time.Duration(seconds) * time.Second
		}
	}
	
	result, err := ds.storageManager.QueryRangeMetrics(context.Background(), query, start, end, step)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result":     result,
		},
	})
}

// handleMetricsLabels 处理标签查询
func (ds *DashboardServer) handleMetricsLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := ds.storageManager.GetLabels(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get labels: %v", err), http.StatusInternalServerError)
		return
	}
	
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data":   labels,
	})
}

// handleMetricsLabelValues 处理标签值查�?
func (ds *DashboardServer) handleMetricsLabelValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	labelName := vars["name"]
	
	values, err := ds.storageManager.GetLabelValues(context.Background(), labelName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get label values: %v", err), http.StatusInternalServerError)
		return
	}
	
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data":   values,
	})
}

// handleMetricsSeries 处理序列查询
func (ds *DashboardServer) handleMetricsSeries(w http.ResponseWriter, r *http.Request) {
	var matchers []string
	
	if r.Method == "POST" {
		var req struct {
			Match []string `json:"match[]"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		matchers = req.Match
	} else {
		matchers = r.URL.Query()["match[]"]
	}
	
	series, err := ds.storageManager.GetSeries(context.Background(), matchers)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get series: %v", err), http.StatusInternalServerError)
		return
	}
	
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data":   series,
	})
}

// handleAlertsGet 处理获取告警
func (ds *DashboardServer) handleAlertsGet(w http.ResponseWriter, r *http.Request) {
	alerts, err := ds.alertManager.GetAlerts(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get alerts: %v", err), http.StatusInternalServerError)
		return
	}
	
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data":   alerts,
	})
}

// handleHealth 处理健康检�?
func (ds *DashboardServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ds.writeJSONResponse(w, map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now().Unix(),
	})
}

// handleStatus 处理状态查�?
func (ds *DashboardServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now().Unix(),
		"version": "1.0.0",
		"uptime": time.Since(time.Now()).String(), // 这里应该是实际的启动时间
	}
	
	// 添加存储状�?
	if ds.storageManager != nil {
		if storageStats, err := ds.storageManager.GetStats(); err == nil {
			status["storage"] = storageStats
		}
	}
	
	// 添加告警状�?
	if ds.alertManager != nil {
		if alertStats, err := ds.alertManager.GetStats(); err == nil {
			status["alerts"] = alertStats
		}
	}
	
	ds.writeJSONResponse(w, status)
}

// 页面处理�?

// handleIndex 处理首页
func (ds *DashboardServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// handleDashboard 处理仪表板页�?
func (ds *DashboardServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ds.renderTemplate(w, "dashboard.html", nil)
}

// handleAlertsPage 处理告警页面
func (ds *DashboardServer) handleAlertsPage(w http.ResponseWriter, r *http.Request) {
	ds.renderTemplate(w, "alerts.html", nil)
}

// handleMetricsPage 处理指标页面
func (ds *DashboardServer) handleMetricsPage(w http.ResponseWriter, r *http.Request) {
	ds.renderTemplate(w, "metrics.html", nil)
}

// WebSocket处理�?

// handleWebSocket 处理WebSocket连接
func (ds *DashboardServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ds.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}
	
	// 生成连接ID
	connID := fmt.Sprintf("%s_%d", r.RemoteAddr, time.Now().UnixNano())
	
	// 保存连接
	ds.wsConnMutex.Lock()
	ds.wsConnections[connID] = conn
	ds.wsConnMutex.Unlock()
	
	// 设置关闭处理�?
	conn.SetCloseHandler(func(code int, text string) error {
		ds.wsConnMutex.Lock()
		delete(ds.wsConnections, connID)
		ds.wsConnMutex.Unlock()
		return nil
	})
	
	// 启动ping/pong
	go ds.handleWebSocketPing(conn, connID)
	
	// 处理消息
	ds.handleWebSocketMessages(conn, connID)
}

// handleWebSocketPing 处理WebSocket ping/pong
func (ds *DashboardServer) handleWebSocketPing(conn *websocket.Conn, connID string) {
	ticker := time.NewTicker(ds.config.WSPingInterval)
	defer ticker.Stop()
	
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(ds.config.WSPongTimeout))
		return nil
	})
	
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ds.wsConnMutex.Lock()
				delete(ds.wsConnections, connID)
				ds.wsConnMutex.Unlock()
				return
			}
		}
	}
}

// handleWebSocketMessages 处理WebSocket消息
func (ds *DashboardServer) handleWebSocketMessages(conn *websocket.Conn, connID string) {
	defer func() {
		ds.wsConnMutex.Lock()
		delete(ds.wsConnections, connID)
		ds.wsConnMutex.Unlock()
		conn.Close()
	}()
	
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}
		
		// 处理消息
		ds.processWebSocketMessage(conn, msg)
	}
}

// processWebSocketMessage 处理WebSocket消息
func (ds *DashboardServer) processWebSocketMessage(conn *websocket.Conn, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}
	
	switch msgType {
	case "subscribe":
		// 处理订阅请求
		ds.handleWebSocketSubscribe(conn, msg)
	case "unsubscribe":
		// 处理取消订阅请求
		ds.handleWebSocketUnsubscribe(conn, msg)
	case "query":
		// 处理查询请求
		ds.handleWebSocketQuery(conn, msg)
	}
}

// handleWebSocketSubscribe 处理WebSocket订阅
func (ds *DashboardServer) handleWebSocketSubscribe(conn *websocket.Conn, msg map[string]interface{}) {
	// 实现订阅逻辑
	response := map[string]interface{}{
		"type": "subscribed",
		"id":   msg["id"],
	}
	conn.WriteJSON(response)
}

// handleWebSocketUnsubscribe 处理WebSocket取消订阅
func (ds *DashboardServer) handleWebSocketUnsubscribe(conn *websocket.Conn, msg map[string]interface{}) {
	// 实现取消订阅逻辑
	response := map[string]interface{}{
		"type": "unsubscribed",
		"id":   msg["id"],
	}
	conn.WriteJSON(response)
}

// handleWebSocketQuery 处理WebSocket查询
func (ds *DashboardServer) handleWebSocketQuery(conn *websocket.Conn, msg map[string]interface{}) {
	// 实现查询逻辑
	response := map[string]interface{}{
		"type": "query_result",
		"id":   msg["id"],
		"data": nil,
	}
	conn.WriteJSON(response)
}

// 工具方法

// writeJSONResponse 写入JSON响应
func (ds *DashboardServer) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// parseTime 解析时间
func (ds *DashboardServer) parseTime(timeStr string) (time.Time, error) {
	// 尝试RFC3339格式
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}
	
	// 尝试Unix时间�?
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}
	
	// 尝试相对时间（如-1h, -30m�?
	if strings.HasPrefix(timeStr, "-") {
		if duration, err := time.ParseDuration(timeStr[1:]); err == nil {
			return time.Now().Add(-duration), nil
		}
	}
	
	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
}

// renderTemplate 渲染模板
func (ds *DashboardServer) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	// 这里应该实现模板渲染逻辑
	// 暂时返回简单的HTML
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Monitoring Dashboard</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .nav { margin-bottom: 20px; }
        .nav a { margin-right: 20px; text-decoration: none; color: #007bff; }
        .nav a:hover { text-decoration: underline; }
        .content { background-color: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    </style>
</head>
<body>
    <div class="header">
        <h1>监控仪表�?/h1>
    </div>
    <div class="nav">
        <a href="/dashboard">仪表�?/a>
        <a href="/metrics">指标</a>
        <a href="/alerts">告警</a>
    </div>
    <div class="content">
        <h2>%s</h2>
        <p>页面正在开发中...</p>
    </div>
</body>
</html>
`, templateName)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
