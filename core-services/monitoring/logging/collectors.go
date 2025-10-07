package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// BaseCollector 基础收集器
type BaseCollector struct {
	config  CollectorConfig
	handler func(*LogEntry) error
	stats   *CollectorStats
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewBaseCollector 创建基础收集器
func NewBaseCollector(config CollectorConfig) *BaseCollector {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &BaseCollector{
		config: config,
		stats: &CollectorStats{
			SourceInfo: config.Name,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetLogHandler 设置日志处理回调
func (bc *BaseCollector) SetLogHandler(handler func(*LogEntry) error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	bc.handler = handler
}

// GetStats 获取统计信息
func (bc *BaseCollector) GetStats() *CollectorStats {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	stats := *bc.stats
	return &stats
}

// HealthCheck 健康检查
func (bc *BaseCollector) HealthCheck() error {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	if !bc.stats.IsActive {
		return fmt.Errorf("collector is not active")
	}
	
	// 检查最近是否有收集活动
	if time.Since(bc.stats.LastCollected) > 5*time.Minute {
		return fmt.Errorf("no recent collection activity")
	}
	
	return nil
}

// handleLog 处理日志
func (bc *BaseCollector) handleLog(entry *LogEntry) {
	bc.mutex.RLock()
	handler := bc.handler
	bc.mutex.RUnlock()
	
	if handler != nil {
		start := time.Now()
		
		if err := handler(entry); err != nil {
			bc.recordError()
		} else {
			bc.recordCollected(time.Since(start))
		}
	}
}

// recordCollected 记录收集统计
func (bc *BaseCollector) recordCollected(duration time.Duration) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.stats.CollectedLogs++
	bc.stats.LastCollected = time.Now()
	bc.stats.CollectionTime = duration
	bc.stats.IsActive = true
}

// recordError 记录错误统计
func (bc *BaseCollector) recordError() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.stats.ErrorLogs++
}

// setActive 设置活跃状态
func (bc *BaseCollector) setActive(active bool) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.stats.IsActive = active
}

// FileCollector 文件收集器
type FileCollector struct {
	*BaseCollector
	filePaths []string
	watcher   *fsnotify.Watcher
	files     map[string]*os.File
	positions map[string]int64
}

// NewFileCollector 创建文件收集器
func NewFileCollector(config CollectorConfig) (*FileCollector, error) {
	base := NewBaseCollector(config)
	
	// 获取文件路径
	paths, ok := config.Settings["paths"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("file paths not specified")
	}
	
	filePaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if pathStr, ok := path.(string); ok {
			filePaths = append(filePaths, pathStr)
		}
	}
	
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no valid file paths specified")
	}
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	
	return &FileCollector{
		BaseCollector: base,
		filePaths:     filePaths,
		watcher:       watcher,
		files:         make(map[string]*os.File),
		positions:     make(map[string]int64),
	}, nil
}

// Start 启动收集器
func (fc *FileCollector) Start() error {
	// 打开文件并添加监听
	for _, path := range fc.filePaths {
		if err := fc.watchFile(path); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", path, err)
		}
	}
	
	// 启动文件监听
	fc.wg.Add(1)
	go fc.watchFiles()
	
	fc.setActive(true)
	return nil
}

// Stop 停止收集器
func (fc *FileCollector) Stop() error {
	fc.cancel()
	fc.watcher.Close()
	
	// 关闭所有文件
	for _, file := range fc.files {
		file.Close()
	}
	
	fc.wg.Wait()
	fc.setActive(false)
	return nil
}

// watchFile 监听文件
func (fc *FileCollector) watchFile(path string) error {
	// 展开通配符
	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}
	
	for _, match := range matches {
		// 打开文件
		file, err := os.Open(match)
		if err != nil {
			continue
		}
		
		// 移动到文件末尾
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			file.Close()
			continue
		}
		
		fc.files[match] = file
		fc.positions[match] = 0
		
		// 添加文件监听
		if err := fc.watcher.Add(match); err != nil {
			file.Close()
			delete(fc.files, match)
			continue
		}
		
		// 读取现有内容
		fc.wg.Add(1)
		go fc.readFile(match)
	}
	
	return nil
}

// watchFiles 监听文件变化
func (fc *FileCollector) watchFiles() {
	defer fc.wg.Done()
	
	for {
		select {
		case <-fc.ctx.Done():
			return
			
		case event, ok := <-fc.watcher.Events:
			if !ok {
				return
			}
			
			if event.Op&fsnotify.Write == fsnotify.Write {
				// 文件被写入，读取新内容
				fc.wg.Add(1)
				go fc.readFile(event.Name)
			}
			
		case err, ok := <-fc.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)
		}
	}
}

// readFile 读取文件
func (fc *FileCollector) readFile(path string) {
	defer fc.wg.Done()
	
	file, exists := fc.files[path]
	if !exists {
		return
	}
	
	// 从上次位置开始读取
	if _, err := file.Seek(fc.positions[path], io.SeekStart); err != nil {
		return
	}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		// 创建日志条目
		entry := &LogEntry{
			ID:        generateID(),
			Timestamp: time.Now(),
			Level:     LogLevelInfo,
			Message:   line,
			Source:    path,
			Service:   fc.config.Name,
			Fields: map[string]interface{}{
				"file": path,
			},
		}
		
		// 尝试解析JSON
		if strings.HasPrefix(line, "{") {
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
				entry.Fields = jsonData
				if msg, ok := jsonData["message"].(string); ok {
					entry.Message = msg
				}
				if level, ok := jsonData["level"].(string); ok {
					entry.Level = parseLogLevel(level)
				}
			}
		}
		
		fc.handleLog(entry)
	}
	
	// 更新位置
	if pos, err := file.Seek(0, io.SeekCurrent); err == nil {
		fc.positions[path] = pos
	}
}

// SyslogCollector 系统日志收集器
type SyslogCollector struct {
	*BaseCollector
	listener net.Listener
	address  string
}

// NewSyslogCollector 创建系统日志收集器
func NewSyslogCollector(config CollectorConfig) (*SyslogCollector, error) {
	base := NewBaseCollector(config)
	
	address, ok := config.Settings["address"].(string)
	if !ok {
		address = ":514" // 默认syslog端口
	}
	
	return &SyslogCollector{
		BaseCollector: base,
		address:       address,
	}, nil
}

// Start 启动收集器
func (sc *SyslogCollector) Start() error {
	listener, err := net.Listen("tcp", sc.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", sc.address, err)
	}
	
	sc.listener = listener
	
	// 启动接收循环
	sc.wg.Add(1)
	go sc.acceptConnections()
	
	sc.setActive(true)
	return nil
}

// Stop 停止收集器
func (sc *SyslogCollector) Stop() error {
	sc.cancel()
	
	if sc.listener != nil {
		sc.listener.Close()
	}
	
	sc.wg.Wait()
	sc.setActive(false)
	return nil
}

// acceptConnections 接受连接
func (sc *SyslogCollector) acceptConnections() {
	defer sc.wg.Done()
	
	for {
		conn, err := sc.listener.Accept()
		if err != nil {
			select {
			case <-sc.ctx.Done():
				return
			default:
				continue
			}
		}
		
		sc.wg.Add(1)
		go sc.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (sc *SyslogCollector) handleConnection(conn net.Conn) {
	defer sc.wg.Done()
	defer conn.Close()
	
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		// 解析syslog消息
		entry := sc.parseSyslogMessage(line)
		sc.handleLog(entry)
	}
}

// parseSyslogMessage 解析syslog消息
func (sc *SyslogCollector) parseSyslogMessage(message string) *LogEntry {
	// 简单的syslog解析
	entry := &LogEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Message:   message,
		Source:    "syslog",
		Service:   sc.config.Name,
		Fields: map[string]interface{}{
			"protocol": "syslog",
		},
	}
	
	// 这里可以添加更复杂的syslog解析逻辑
	
	return entry
}

// HTTPCollector HTTP收集器
type HTTPCollector struct {
	*BaseCollector
	server *http.Server
	port   int
}

// NewHTTPCollector 创建HTTP收集器
func NewHTTPCollector(config CollectorConfig) (*HTTPCollector, error) {
	base := NewBaseCollector(config)
	
	port, ok := config.Settings["port"].(int)
	if !ok {
		port = 8080 // 默认端口
	}
	
	collector := &HTTPCollector{
		BaseCollector: base,
		port:          port,
	}
	
	// 创建HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/logs", collector.handleLogs)
	mux.HandleFunc("/health", collector.handleHealth)
	
	collector.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	return collector, nil
}

// Start 启动收集器
func (hc *HTTPCollector) Start() error {
	// 启动HTTP服务器
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		
		if err := hc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP collector server error: %v\n", err)
		}
	}()
	
	hc.setActive(true)
	return nil
}

// Stop 停止收集器
func (hc *HTTPCollector) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := hc.server.Shutdown(ctx); err != nil {
		return err
	}
	
	hc.wg.Wait()
	hc.setActive(false)
	return nil
}

// handleLogs 处理日志请求
func (hc *HTTPCollector) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	
	// 解析日志条目
	var entries []*LogEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		// 尝试解析单个条目
		var entry LogEntry
		if err := json.Unmarshal(body, &entry); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		entries = []*LogEntry{&entry}
	}
	
	// 处理每个日志条目
	for _, entry := range entries {
		if entry.ID == "" {
			entry.ID = generateID()
		}
		if entry.Timestamp.IsZero() {
			entry.Timestamp = time.Now()
		}
		if entry.Source == "" {
			entry.Source = "http"
		}
		if entry.Service == "" {
			entry.Service = hc.config.Name
		}
		
		hc.handleLog(entry)
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleHealth 处理健康检查请求
func (hc *HTTPCollector) handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := hc.GetStats()
	
	response := map[string]interface{}{
		"status":          "ok",
		"collected_logs":  stats.CollectedLogs,
		"error_logs":      stats.ErrorLogs,
		"last_collected":  stats.LastCollected,
		"is_active":       stats.IsActive,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 占位符实现，实际项目中需要根据具体需求实现

// NewJournaldCollector 创建journald收集器
func NewJournaldCollector(config CollectorConfig) (LogCollector, error) {
	// 占位符实现
	return NewBaseCollector(config), nil
}

// NewDockerCollector 创建Docker收集器
func NewDockerCollector(config CollectorConfig) (LogCollector, error) {
	// 占位符实现
	return NewBaseCollector(config), nil
}

// NewKubernetesCollector 创建Kubernetes收集器
func NewKubernetesCollector(config CollectorConfig) (LogCollector, error) {
	// 占位符实现
	return NewBaseCollector(config), nil
}

// NewTCPCollector 创建TCP收集器
func NewTCPCollector(config CollectorConfig) (LogCollector, error) {
	// 占位符实现
	return NewBaseCollector(config), nil
}

// NewUDPCollector 创建UDP收集器
func NewUDPCollector(config CollectorConfig) (LogCollector, error) {
	// 占位符实现
	return NewBaseCollector(config), nil
}

// 辅助函数

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "TRACE":
		return LogLevelTrace
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	case "FATAL":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}