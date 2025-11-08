package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/service"
	"go.uber.org/zap"
)

// PluginRuntime 管理插件的运行时环境
type PluginRuntime struct {
	plugins     map[string]*PluginInstance
	mu          sync.RWMutex
	logger      *zap.Logger
	directory   string
	tempDir     string
	auditLogger service.AuditLogger
}

// PluginInstance 表示一个运行中的插件实例
type PluginInstance struct {
	ID          string
	Name        string
	Version     string
	Status      models.PluginStatus
	Config      map[string]interface{}
	Process     *os.Process
	StartTime   time.Time
	LastPing    time.Time
	LogFile     *os.File
	Manifest    *PluginManifest
	Environment map[string]string
}

// PluginManifest 插件清单
type PluginManifest struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Entry        string                 `json:"entry"`
	Permissions  []string               `json:"permissions"`
	Dependencies []string               `json:"dependencies"`
	Checksum     string                 `json:"checksum"`
	Signature    string                 `json:"signature"`
	Metadata     map[string]interface{} `json:"metadata"`
	Compatibility Compatibility         `json:"compatibility"`
}

// Compatibility 兼容性信息
type Compatibility struct {
	API     string `json:"api"`
	Backend string `json:"backend"`
}

// NewPluginRuntime 创建新的插件运行时
func NewPluginRuntime(logger *zap.Logger, directory, tempDir string, auditLogger service.AuditLogger) *PluginRuntime {
	return &PluginRuntime{
		plugins:     make(map[string]*PluginInstance),
		logger:      logger,
		directory:   directory,
		tempDir:     tempDir,
		auditLogger: auditLogger,
	}
}

// Install 安装插件
func (r *PluginRuntime) Install(ctx context.Context, plugin *models.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查插件是否已存在
	if _, exists := r.plugins[plugin.ID]; exists {
		return fmt.Errorf("plugin %s already installed", plugin.ID)
	}

	// 创建插件目录
	pluginPath := filepath.Join(r.directory, plugin.ID)
	if err := os.MkdirAll(pluginPath, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// 下载插件包
	packagePath := filepath.Join(r.tempDir, fmt.Sprintf("%s-%s.tar.gz", plugin.ID, plugin.Version))
	if err := r.downloadPlugin(plugin.Source, packagePath); err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}

	// 解压插件包
	if err := r.extractPlugin(packagePath, pluginPath); err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	// 验证插件清单
	manifestPath := filepath.Join(pluginPath, "manifest.json")
	manifest, err := r.loadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	// 验证插件签名
	if err := r.verifyPlugin(manifest); err != nil {
		return fmt.Errorf("plugin verification failed: %w", err)
	}

	// 创建插件实例
	instance := &PluginInstance{
		ID:          plugin.ID,
		Name:        plugin.Name,
		Version:     plugin.Version,
		Status:      models.PluginStatusInstalled,
		Config:      plugin.Config,
		Manifest:    manifest,
		Environment: make(map[string]string),
	}

	// 设置环境变量
	instance.Environment["PLUGIN_ID"] = plugin.ID
	instance.Environment["PLUGIN_DIR"] = pluginPath
	instance.Environment["PLUGIN_VERSION"] = plugin.Version

	// 将配置转换为环境变量
	for key, value := range plugin.Config {
		if str, ok := value.(string); ok {
			instance.Environment[fmt.Sprintf("PLUGIN_CONFIG_%s", key)] = str
		}
	}

	r.plugins[plugin.ID] = instance

	// 记录审计日志
	r.auditLogger.Log(ctx, &models.AuditLog{
		Action:     "install",
		Resource:   "plugin",
		ResourceID: plugin.ID,
		Details:    map[string]interface{}{"name": plugin.Name, "version": plugin.Version},
		Result:     "success",
	})

	r.logger.Info("Plugin installed successfully",
		zap.String("plugin_id", plugin.ID),
		zap.String("plugin_name", plugin.Name),
		zap.String("version", plugin.Version))

	return nil
}

// Start 启动插件
func (r *PluginRuntime) Start(ctx context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if instance.Status == models.PluginStatusRunning {
		return fmt.Errorf("plugin %s is already running", pluginID)
	}

	// 创建日志文件
	logPath := filepath.Join(r.tempDir, fmt.Sprintf("%s.log", pluginID))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// 准备启动命令
	pluginPath := filepath.Join(r.directory, pluginID)
	entryPath := filepath.Join(pluginPath, instance.Manifest.Entry)

	// 创建子进程
	cmd := r.createPluginCommand(entryPath, instance.Environment)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start plugin process: %w", err)
	}

	// 更新实例状态
	instance.Process = cmd.Process
	instance.Status = models.PluginStatusRunning
	instance.StartTime = time.Now()
	instance.LastPing = time.Now()
	instance.LogFile = logFile

	// 记录审计日志
	r.auditLogger.Log(ctx, &models.AuditLog{
		Action:     "start",
		Resource:   "plugin",
		ResourceID: pluginID,
		Details:    map[string]interface{}{"name": instance.Name},
		Result:     "success",
	})

	r.logger.Info("Plugin started successfully",
		zap.String("plugin_id", pluginID),
		zap.Int("pid", instance.Process.Pid))

	return nil
}

// Stop 停止插件
func (r *PluginRuntime) Stop(ctx context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if instance.Status != models.PluginStatusRunning {
		return fmt.Errorf("plugin %s is not running", pluginID)
	}

	// 发送SIGTERM信号
	if err := instance.Process.Signal(os.Interrupt); err != nil {
		r.logger.Warn("Failed to send SIGTERM to plugin",
			zap.String("plugin_id", pluginID),
			zap.Error(err))
	}

	// 等待进程退出
	done := make(chan error, 1)
	go func() {
		_, err := instance.Process.Wait()
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			r.logger.Warn("Plugin process exited with error",
				zap.String("plugin_id", pluginID),
				zap.Error(err))
		}
	case <-time.After(10 * time.Second):
		// 超时后强制杀死进程
		if err := instance.Process.Kill(); err != nil {
			r.logger.Error("Failed to kill plugin process",
				zap.String("plugin_id", pluginID),
				zap.Error(err))
		}
		<-done // 等待进程真正退出
	}

	// 关闭日志文件
	if instance.LogFile != nil {
		instance.LogFile.Close()
	}

	// 更新实例状态
	instance.Status = models.PluginStatusStopped
	instance.Process = nil

	// 记录审计日志
	r.auditLogger.Log(ctx, &models.AuditLog{
		Action:     "stop",
		Resource:   "plugin",
		ResourceID: pluginID,
		Details:    map[string]interface{}{"name": instance.Name},
		Result:     "success",
	})

	r.logger.Info("Plugin stopped successfully",
		zap.String("plugin_id", pluginID))

	return nil
}

// Uninstall 卸载插件
func (r *PluginRuntime) Uninstall(ctx context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 如果插件正在运行，先停止它
	if instance.Status == models.PluginStatusRunning {
		if err := r.Stop(ctx, pluginID); err != nil {
			r.logger.Warn("Failed to stop plugin before uninstall",
				zap.String("plugin_id", pluginID),
				zap.Error(err))
		}
	}

	// 删除插件目录
	pluginPath := filepath.Join(r.directory, pluginID)
	if err := os.RemoveAll(pluginPath); err != nil {
		return fmt.Errorf("failed to remove plugin directory: %w", err)
	}

	// 删除插件实例
	delete(r.plugins, pluginID)

	// 记录审计日志
	r.auditLogger.Log(ctx, &models.AuditLog{
		Action:     "uninstall",
		Resource:   "plugin",
		ResourceID: pluginID,
		Details:    map[string]interface{}{"name": instance.Name},
		Result:     "success",
	})

	r.logger.Info("Plugin uninstalled successfully",
		zap.String("plugin_id", pluginID))

	return nil
}

// GetStatus 获取插件状态
func (r *PluginRuntime) GetStatus(pluginID string) (models.PluginStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, exists := r.plugins[pluginID]
	if !exists {
		return "", fmt.Errorf("plugin %s not found", pluginID)
	}

	return instance.Status, nil
}

// List 列出所有插件
func (r *PluginRuntime) List() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plugins []PluginInfo
	for _, instance := range r.plugins {
		plugins = append(plugins, PluginInfo{
			ID:        instance.ID,
			Name:      instance.Name,
			Version:   instance.Version,
			Status:    instance.Status,
			StartTime: instance.StartTime,
			LastPing:  instance.LastPing,
		})
	}

	return plugins
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Version   string                `json:"version"`
	Status    models.PluginStatus   `json:"status"`
	StartTime time.Time             `json:"startTime"`
	LastPing  time.Time             `json:"lastPing"`
}

// 以下是辅助方法

func (r *PluginRuntime) downloadPlugin(source, dest string) error {
	// 实现插件下载逻辑
	// 这里可以是HTTP下载、Git克隆等
	return nil
}

func (r *PluginRuntime) extractPlugin(src, dest string) error {
	// 实现插件解压逻辑
	// 这里可以使用tar/gzip等工具
	return nil
}

func (r *PluginRuntime) loadManifest(path string) (*PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (r *PluginRuntime) verifyPlugin(manifest *PluginManifest) error {
	// 实现插件验证逻辑
	// 这里可以验证签名、校验和等
	return nil
}

func (r *PluginRuntime) createPluginCommand(entry string, env map[string]string) *os.Cmd {
	// 创建插件启动命令
	// 这里可以根据插件类型选择不同的启动方式
	return nil
}

// HealthCheck 健康检查
func (r *PluginRuntime) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for pluginID, instance := range r.plugins {
		if instance.Status == models.PluginStatusRunning {
			// 检查进程是否还在运行
			if instance.Process == nil {
				instance.Status = models.PluginStatusError
				r.logger.Error("Plugin process is nil but status is running",
					zap.String("plugin_id", pluginID))
				continue
			}

			// 发送心跳检查
			if err := instance.Process.Signal(os.Signal(nil)); err != nil {
				instance.Status = models.PluginStatusError
				r.logger.Error("Plugin process is not responding",
					zap.String("plugin_id", pluginID),
					zap.Error(err))
			}
		}
	}

	return nil
}