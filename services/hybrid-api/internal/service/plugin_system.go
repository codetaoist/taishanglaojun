package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/communication"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/lifecycle"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/runtime"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/sandbox"
	"go.uber.org/zap"
)

// PluginSystemService 插件系统服务
type PluginSystemService struct {
	runtime          *runtime.PluginRuntime
	sandboxManager   *sandbox.PluginSandboxManager
	lifecycleManager *lifecycle.PluginLifecycleManager
	commManager      *communication.PluginCommunicationManager
	pluginDAO        *dao.PluginDAO
	auditLogDAO      *dao.AuditLogDAO
	config           *config.Config
	logger           *zap.Logger
	mu               sync.RWMutex
	started          bool
}

// NewPluginSystemService 创建插件系统服务
func NewPluginSystemService(
	cfg *config.Config,
	pluginDAO *dao.PluginDAO,
	auditLogDAO *dao.AuditLogDAO,
	logger *zap.Logger,
) (*PluginSystemService, error) {
	// 创建插件运行时
	pluginRuntime := runtime.NewPluginRuntime(cfg, logger)

	// 创建沙箱管理器
	sandboxManager := sandbox.NewPluginSandboxManager(cfg.Plugin.SandboxDir, logger)

	// 创建生命周期管理器
	lifecycleManager := lifecycle.NewPluginLifecycleManager(pluginRuntime, sandboxManager, logger)

	// 创建通信管理器
	commManager := communication.NewPluginCommunicationManager(lifecycleManager, logger)

	// 创建插件系统服务
	service := &PluginSystemService{
		runtime:          pluginRuntime,
		sandboxManager:   sandboxManager,
		lifecycleManager: lifecycleManager,
		commManager:      commManager,
		pluginDAO:        pluginDAO,
		auditLogDAO:      auditLogDAO,
		config:           cfg,
		logger:           logger,
		started:          false,
	}

	// 注册事件处理器
	lifecycleManager.AddEventHandler(service)

	return service, nil
}

// Start 启动插件系统服务
func (s *PluginSystemService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("plugin system service already started")
	}

	s.logger.Info("Starting plugin system service")

	// 加载已安装的插件
	plugins, err := s.pluginDAO.List(ctx, "default", "", "", 1, 1000)
	if err != nil {
		return fmt.Errorf("failed to list installed plugins: %w", err)
	}

	// 恢复插件状态
	for _, plugin := range plugins {
		instance := &lifecycle.PluginInstance{
			ID:          plugin.ID,
			Name:        plugin.Name,
			Version:     plugin.Version,
			State:       lifecycle.PluginState(plugin.Status),
			Config:      make(map[string]interface{}),
			Metadata:    make(map[string]string),
			RestartCount: 0,
			HealthCheck: lifecycle.HealthCheck{
				Enabled:      false,
				Interval:     30 * 0,
				Timeout:      5 * 0,
				MaxFailures:  3,
				FailureCount: 0,
			},
		}

		// 如果插件状态为运行中，尝试启动
		if plugin.Status == "running" {
			if err := s.lifecycleManager.StartPlugin(ctx, plugin.ID); err != nil {
				s.logger.Error("Failed to start plugin during service startup",
					zap.String("plugin_id", plugin.ID),
					zap.Error(err))
			}
		}
	}

	s.started = true
	s.logger.Info("Plugin system service started successfully")
	return nil
}

// Stop 停止插件系统服务
func (s *PluginSystemService) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	s.logger.Info("Stopping plugin system service")

	// 停止生命周期管理器
	if err := s.lifecycleManager.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown lifecycle manager", zap.Error(err))
	}

	s.started = false
	s.logger.Info("Plugin system service stopped")
	return nil
}

// InstallPlugin 安装插件
func (s *PluginSystemService) InstallPlugin(ctx context.Context, name, version, source string, config map[string]string) error {
	// 创建插件模型
	plugin := &models.Plugin{
		ID:       fmt.Sprintf("%s-%s", name, version),
		Name:     name,
		Version:  version,
		Status:   "installed",
		Checksum: "",
	}

	// 通过生命周期管理器安装插件
	if err := s.lifecycleManager.InstallPlugin(ctx, plugin); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	// 记录审计日志
	if err := s.auditLogDAO.Create(ctx, &models.AuditLog{
		TenantID: "default",
		Action:   "install_plugin",
		Resource: plugin.ID,
		Details:  fmt.Sprintf("Installed plugin %s version %s from %s", name, version, source),
	}); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	return nil
}

// StartPlugin 启动插件
func (s *PluginSystemService) StartPlugin(ctx context.Context, pluginID string) error {
	if err := s.lifecycleManager.StartPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// 更新数据库状态
	if err := s.pluginDAO.SetStatus(ctx, pluginID, "running"); err != nil {
		s.logger.Error("Failed to update plugin status in database",
			zap.String("plugin_id", pluginID),
			zap.Error(err))
	}

	// 记录审计日志
	if err := s.auditLogDAO.Create(ctx, &models.AuditLog{
		TenantID: "default",
		Action:   "start_plugin",
		Resource: pluginID,
		Details:  fmt.Sprintf("Started plugin %s", pluginID),
	}); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	return nil
}

// StopPlugin 停止插件
func (s *PluginSystemService) StopPlugin(ctx context.Context, pluginID string) error {
	if err := s.lifecycleManager.StopPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// 更新数据库状态
	if err := s.pluginDAO.SetStatus(ctx, pluginID, "stopped"); err != nil {
		s.logger.Error("Failed to update plugin status in database",
			zap.String("plugin_id", pluginID),
			zap.Error(err))
	}

	// 记录审计日志
	if err := s.auditLogDAO.Create(ctx, &models.AuditLog{
		TenantID: "default",
		Action:   "stop_plugin",
		Resource: pluginID,
		Details:  fmt.Sprintf("Stopped plugin %s", pluginID),
	}); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	return nil
}

// UpgradePlugin 升级插件
func (s *PluginSystemService) UpgradePlugin(ctx context.Context, pluginID, newVersion string) error {
	if err := s.lifecycleManager.UpgradePlugin(ctx, pluginID, newVersion); err != nil {
		return fmt.Errorf("failed to upgrade plugin: %w", err)
	}

	// 更新数据库版本
	if err := s.pluginDAO.Upgrade(ctx, pluginID, newVersion); err != nil {
		s.logger.Error("Failed to update plugin version in database",
			zap.String("plugin_id", pluginID),
			zap.Error(err))
	}

	// 记录审计日志
	if err := s.auditLogDAO.Create(ctx, &models.AuditLog{
		TenantID: "default",
		Action:   "upgrade_plugin",
		Resource: pluginID,
		Details:  fmt.Sprintf("Upgraded plugin %s to version %s", pluginID, newVersion),
	}); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	return nil
}

// UninstallPlugin 卸载插件
func (s *PluginSystemService) UninstallPlugin(ctx context.Context, pluginID string) error {
	if err := s.lifecycleManager.UninstallPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	// 从数据库删除插件
	if err := s.pluginDAO.Delete(ctx, pluginID); err != nil {
		s.logger.Error("Failed to delete plugin from database",
			zap.String("plugin_id", pluginID),
			zap.Error(err))
	}

	// 记录审计日志
	if err := s.auditLogDAO.Create(ctx, &models.AuditLog{
		TenantID: "default",
		Action:   "uninstall_plugin",
		Resource: pluginID,
		Details:  fmt.Sprintf("Uninstalled plugin %s", pluginID),
	}); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	return nil
}

// GetPluginInstance 获取插件实例
func (s *PluginSystemService) GetPluginInstance(pluginID string) (*lifecycle.PluginInstance, error) {
	return s.lifecycleManager.GetPluginInstance(pluginID)
}

// ListPlugins 列出所有插件
func (s *PluginSystemService) ListPlugins() []*lifecycle.PluginInstance {
	return s.lifecycleManager.ListPlugins()
}

// GetCommunicationManager 获取通信管理器
func (s *PluginSystemService) GetCommunicationManager() *communication.PluginCommunicationManager {
	return s.commManager
}

// HandleEvent 实现PluginEventHandler接口
func (s *PluginSystemService) HandleEvent(ctx context.Context, event lifecycle.PluginEvent) error {
	s.logger.Info("Plugin event received",
		zap.String("plugin_id", event.PluginID),
		zap.String("event_type", event.EventType))

	// 根据事件类型执行不同的操作
	switch event.EventType {
	case "installed", "started", "stopped", "upgraded", "uninstalled":
		// 这些事件已经在对应的操作方法中处理了数据库更新和审计日志
		// 这里可以添加额外的处理逻辑，比如通知其他系统
	default:
		// 其他事件类型的处理
	}

	return nil
}