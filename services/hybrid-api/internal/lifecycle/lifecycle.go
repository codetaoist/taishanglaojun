package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/runtime"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/sandbox"
	"go.uber.org/zap"
)

// PluginState 插件状态
type PluginState string

const (
	StateInstalled PluginState = "installed"
	StateStarting  PluginState = "starting"
	StateRunning   PluginState = "running"
	StateStopping  PluginState = "stopping"
	StateStopped   PluginState = "stopped"
	StateError     PluginState = "error"
	StateUpgrading PluginState = "upgrading"
)

// PluginEvent 插件事件
type PluginEvent struct {
	PluginID   string                 `json:"pluginId"`
	EventType  string                 `json:"eventType"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// PluginEventHandler 插件事件处理器
type PluginEventHandler interface {
	HandleEvent(ctx context.Context, event PluginEvent) error
}

// PluginLifecycleManager 插件生命周期管理器
type PluginLifecycleManager struct {
	runtime     *runtime.PluginRuntime
	sandboxMgr  *sandbox.PluginSandboxManager
	plugins     map[string]*PluginInstance
	handlers    []PluginEventHandler
	mu          sync.RWMutex
	logger      *zap.Logger
	eventChan   chan PluginEvent
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// PluginInstance 插件实例
type PluginInstance struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	State       PluginState       `json:"state"`
	Sandbox     *sandbox.Sandbox  `json:"-"`
	Config      map[string]interface{} `json:"config"`
	Metadata    map[string]string `json:"metadata"`
	StartTime   time.Time         `json:"startTime,omitempty"`
	StopTime    time.Time         `json:"stopTime,omitempty"`
	Error       string            `json:"error,omitempty"`
	RestartCount int              `json:"restartCount"`
	HealthCheck HealthCheck       `json:"healthCheck"`
}

// HealthCheck 健康检查配置
type HealthCheck struct {
	Enabled     bool          `json:"enabled"`
	Path        string        `json:"path"`
	Method      string        `json:"method"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	MaxFailures int           `json:"maxFailures"`
	FailureCount int          `json:"failureCount"`
	LastCheck   time.Time     `json:"lastCheck"`
}

// NewPluginLifecycleManager 创建插件生命周期管理器
func NewPluginLifecycleManager(
	runtime *runtime.PluginRuntime,
	sandboxMgr *sandbox.PluginSandboxManager,
	logger *zap.Logger,
) *PluginLifecycleManager {
	mgr := &PluginLifecycleManager{
		runtime:    runtime,
		sandboxMgr: sandboxMgr,
		plugins:    make(map[string]*PluginInstance),
		handlers:   make([]PluginEventHandler, 0),
		logger:     logger,
		eventChan:  make(chan PluginEvent, 100),
		stopChan:   make(chan struct{}),
	}

	// 启动事件处理循环
	mgr.wg.Add(1)
	go mgr.eventLoop()

	return mgr
}

// AddEventHandler 添加事件处理器
func (m *PluginLifecycleManager) AddEventHandler(handler PluginEventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// InstallPlugin 安装插件
func (m *PluginLifecycleManager) InstallPlugin(ctx context.Context, plugin *models.Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查插件是否已存在
	if _, exists := m.plugins[plugin.ID]; exists {
		return fmt.Errorf("plugin %s already exists", plugin.ID)
	}

	// 创建插件实例
	instance := &PluginInstance{
		ID:          plugin.ID,
		Name:        plugin.Name,
		Version:     plugin.Version,
		State:       StateInstalled,
		Config:      make(map[string]interface{}),
		Metadata:    make(map[string]string),
		RestartCount: 0,
		HealthCheck: HealthCheck{
			Enabled:      false,
			Interval:     30 * time.Second,
			Timeout:      5 * time.Second,
			MaxFailures:  3,
			FailureCount: 0,
		},
	}

	// 安装插件
	if err := m.runtime.Install(ctx, plugin); err != nil {
		instance.State = StateError
		instance.Error = err.Error()
		m.plugins[plugin.ID] = instance
		m.emitEvent(PluginEvent{
			PluginID:  plugin.ID,
			EventType: "install_failed",
			Data:      map[string]interface{}{"error": err.Error()},
		})
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	m.plugins[plugin.ID] = instance
	m.emitEvent(PluginEvent{
		PluginID:  plugin.ID,
		EventType: "installed",
	})

	m.logger.Info("Plugin installed successfully",
		zap.String("plugin_id", plugin.ID),
		zap.String("plugin_name", plugin.Name),
		zap.String("version", plugin.Version))

	return nil
}

// StartPlugin 启动插件
func (m *PluginLifecycleManager) StartPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if instance.State == StateRunning {
		return fmt.Errorf("plugin %s is already running", pluginID)
	}

	// 更新状态为启动中
	instance.State = StateStarting
	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "starting",
	})

	// 创建沙箱
	sandbox, err := m.sandboxMgr.CreateSandbox(pluginID, sandbox.DefaultResourceLimits())
	if err != nil {
		instance.State = StateError
		instance.Error = err.Error()
		m.emitEvent(PluginEvent{
			PluginID:  pluginID,
			EventType: "start_failed",
			Data:      map[string]interface{}{"error": err.Error()},
		})
		return fmt.Errorf("failed to create sandbox: %w", err)
	}

	instance.Sandbox = sandbox

	// 启动插件
	if err := m.runtime.Start(ctx, pluginID); err != nil {
		instance.State = StateError
		instance.Error = err.Error()
		m.emitEvent(PluginEvent{
			PluginID:  pluginID,
			EventType: "start_failed",
			Data:      map[string]interface{}{"error": err.Error()},
		})
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// 更新状态
	instance.State = StateRunning
	instance.StartTime = time.Now()
	instance.Error = ""

	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "started",
	})

	m.logger.Info("Plugin started successfully",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", instance.Name))

	return nil
}

// StopPlugin 停止插件
func (m *PluginLifecycleManager) StopPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if instance.State != StateRunning {
		return fmt.Errorf("plugin %s is not running", pluginID)
	}

	// 更新状态为停止中
	instance.State = StateStopping
	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "stopping",
	})

	// 停止插件
	if err := m.runtime.Stop(ctx, pluginID); err != nil {
		instance.State = StateError
		instance.Error = err.Error()
		m.emitEvent(PluginEvent{
			PluginID:  pluginID,
			EventType: "stop_failed",
			Data:      map[string]interface{}{"error": err.Error()},
		})
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// 销毁沙箱
	if instance.Sandbox != nil {
		if err := m.sandboxMgr.DestroySandbox(pluginID); err != nil {
			m.logger.Error("Failed to destroy sandbox",
				zap.String("plugin_id", pluginID),
				zap.Error(err))
		}
		instance.Sandbox = nil
	}

	// 更新状态
	instance.State = StateStopped
	instance.StopTime = time.Now()
	instance.Error = ""

	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "stopped",
	})

	m.logger.Info("Plugin stopped successfully",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", instance.Name))

	return nil
}

// UpgradePlugin 升级插件
func (m *PluginLifecycleManager) UpgradePlugin(ctx context.Context, pluginID, newVersion string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 如果插件正在运行，先停止
	if instance.State == StateRunning {
		if err := m.StopPlugin(ctx, pluginID); err != nil {
			return fmt.Errorf("failed to stop plugin before upgrade: %w", err)
		}
	}

	// 更新状态为升级中
	instance.State = StateUpgrading
	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "upgrading",
		Data:      map[string]interface{}{"new_version": newVersion},
	})

	// 升级插件
	if err := m.runtime.Upgrade(ctx, pluginID, newVersion); err != nil {
		instance.State = StateError
		instance.Error = err.Error()
		m.emitEvent(PluginEvent{
			PluginID:  pluginID,
			EventType: "upgrade_failed",
			Data:      map[string]interface{}{"error": err.Error(), "new_version": newVersion},
		})
		return fmt.Errorf("failed to upgrade plugin: %w", err)
	}

	// 更新版本
	instance.Version = newVersion
	instance.State = StateInstalled
	instance.Error = ""

	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "upgraded",
		Data:      map[string]interface{}{"new_version": newVersion},
	})

	m.logger.Info("Plugin upgraded successfully",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", instance.Name),
		zap.String("new_version", newVersion))

	return nil
}

// UninstallPlugin 卸载插件
func (m *PluginLifecycleManager) UninstallPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 如果插件正在运行，先停止
	if instance.State == StateRunning {
		if err := m.StopPlugin(ctx, pluginID); err != nil {
			return fmt.Errorf("failed to stop plugin before uninstall: %w", err)
		}
	}

	// 卸载插件
	if err := m.runtime.Uninstall(ctx, pluginID); err != nil {
		instance.State = StateError
		instance.Error = err.Error()
		m.emitEvent(PluginEvent{
			PluginID:  pluginID,
			EventType: "uninstall_failed",
			Data:      map[string]interface{}{"error": err.Error()},
		})
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	// 从管理器中移除
	delete(m.plugins, pluginID)

	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "uninstalled",
	})

	m.logger.Info("Plugin uninstalled successfully",
		zap.String("plugin_id", pluginID),
		zap.String("plugin_name", instance.Name))

	return nil
}

// GetPluginInstance 获取插件实例
func (m *PluginLifecycleManager) GetPluginInstance(pluginID string) (*PluginInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return instance, nil
}

// ListPlugins 列出所有插件
func (m *PluginLifecycleManager) ListPlugins() []*PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*PluginInstance, 0, len(m.plugins))
	for _, instance := range m.plugins {
		plugins = append(plugins, instance)
	}

	return plugins
}

// RestartPlugin 重启插件
func (m *PluginLifecycleManager) RestartPlugin(ctx context.Context, pluginID string) error {
	instance, exists := m.GetPluginInstance(pluginID)
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 如果插件正在运行，先停止
	if instance.State == StateRunning {
		if err := m.StopPlugin(ctx, pluginID); err != nil {
			return fmt.Errorf("failed to stop plugin before restart: %w", err)
		}
	}

	// 增加重启计数
	instance.RestartCount++

	// 启动插件
	if err := m.StartPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to start plugin after restart: %w", err)
	}

	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "restarted",
		Data:      map[string]interface{}{"restart_count": instance.RestartCount},
	})

	return nil
}

// UpdatePluginConfig 更新插件配置
func (m *PluginLifecycleManager) UpdatePluginConfig(pluginID string, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	instance.Config = config

	m.emitEvent(PluginEvent{
		PluginID:  pluginID,
		EventType: "config_updated",
		Data:      map[string]interface{}{"config": config},
	})

	return nil
}

// emitEvent 发出事件
func (m *PluginLifecycleManager) emitEvent(event PluginEvent) {
	event.Timestamp = time.Now()
	select {
	case m.eventChan <- event:
	default:
		m.logger.Warn("Event channel is full, dropping event",
			zap.String("plugin_id", event.PluginID),
			zap.String("event_type", event.EventType))
	}
}

// eventLoop 事件处理循环
func (m *PluginLifecycleManager) eventLoop() {
	defer m.wg.Done()

	for {
		select {
		case event := <-m.eventChan:
			m.handleEvent(event)
		case <-m.stopChan:
			return
		}
	}
}

// handleEvent 处理事件
func (m *PluginLifecycleManager) handleEvent(event PluginEvent) {
	m.mu.RLock()
	handlers := make([]PluginEventHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.HandleEvent(context.Background(), event); err != nil {
			m.logger.Error("Failed to handle event",
				zap.String("plugin_id", event.PluginID),
				zap.String("event_type", event.EventType),
				zap.Error(err))
		}
	}
}

// Shutdown 关闭生命周期管理器
func (m *PluginLifecycleManager) Shutdown(ctx context.Context) error {
	// 停止所有运行中的插件
	for _, instance := range m.ListPlugins() {
		if instance.State == StateRunning {
			if err := m.StopPlugin(ctx, instance.ID); err != nil {
				m.logger.Error("Failed to stop plugin during shutdown",
					zap.String("plugin_id", instance.ID),
					zap.Error(err))
			}
		}
	}

	// 停止事件处理循环
	close(m.stopChan)
	m.wg.Wait()

	return nil
}