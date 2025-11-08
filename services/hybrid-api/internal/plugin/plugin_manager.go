package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	"time"

	"github.com/codetaoist/services/api/internal/middleware"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeVector     PluginType = "vector"
	PluginTypeModel      PluginType = "model"
	PluginTypeMiddleware PluginType = "middleware"
	PluginTypeWorkflow   PluginType = "workflow"
	PluginTypeAI         PluginType = "ai"
)

// PluginInfo contains information about a plugin
type PluginInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Type        PluginType             `json:"type"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Homepage    string                 `json:"homepage"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	LoadedAt    time.Time              `json:"loaded_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Plugin represents a plugin interface
type Plugin interface {
	// Initialize initializes the plugin with the given configuration
	Initialize(ctx context.Context, config map[string]interface{}) error
	
	// Start starts the plugin
	Start(ctx context.Context) error
	
	// Stop stops the plugin
	Stop(ctx context.Context) error
	
	// GetInfo returns plugin information
	GetInfo() *PluginInfo
	
	// HealthCheck checks if the plugin is healthy
	HealthCheck(ctx context.Context) error
}

// PluginManager manages plugin lifecycle
type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
	logger  *zap.Logger
	config  *PluginConfig
}

// PluginConfig contains plugin manager configuration
type PluginConfig struct {
	Directory     string        `yaml:"directory"`
	Repository    struct {
		URL    string `yaml:"url"`
		Branch string `yaml:"branch"`
	} `yaml:"repository"`
	AutoLoad      bool          `yaml:"auto_load"`
	AutoBuild     bool          `yaml:"auto_build"`
	Versioning    bool          `yaml:"versioning"`
	UpdateInterval time.Duration `yaml:"update_interval"`
	Timeout       time.Duration `yaml:"timeout"`
	ResourceLimits struct {
		Memory string `yaml:"memory"`
		CPU    string `yaml:"cpu"`
	} `yaml:"resource_limits"`
	HealthCheck  bool          `yaml:"health_check"`
	CheckInterval time.Duration `yaml:"check_interval"`
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(config *PluginConfig, logger *zap.Logger) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		logger:  logger,
		config:  config,
	}
}

// LoadPlugin loads a plugin from the given path
func (pm *PluginManager) LoadPlugin(ctx context.Context, path string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("plugin file does not exist: %s", path)
	}

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %v", err)
	}

	// Look up the NewPlugin function
	newPluginSymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin does not export NewPlugin function: %v", err)
	}

	// Type assert the symbol to the correct type
	newPlugin, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("unexpected type from module symbol")
	}

	// Create a new plugin instance
	pluginInstance := newPlugin()

	// Get plugin info
	info := pluginInstance.GetInfo()

	// Check if plugin with same ID already exists
	if _, exists := pm.plugins[info.ID]; exists {
		return fmt.Errorf("plugin with ID %s already loaded", info.ID)
	}

	// Initialize the plugin
	if err := pluginInstance.Initialize(ctx, info.Config); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %v", info.ID, err)
	}

	// Start the plugin
	if err := pluginInstance.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin %s: %v", info.ID, err)
	}

	// Add to plugins map
	pm.plugins[info.ID] = pluginInstance

	pm.logger.Info("Successfully loaded plugin", 
		zap.String("plugin_id", info.ID),
		zap.String("plugin_name", info.Name),
		zap.String("version", info.Version))
	return nil
}

// UnloadPlugin unloads a plugin with the given ID
func (pm *PluginManager) UnloadPlugin(ctx context.Context, id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[id]
	if !exists {
		return fmt.Errorf("plugin with ID %s not found", id)
	}

	// Stop the plugin
	if err := plugin.Stop(ctx); err != nil {
		pm.logger.Error("Error stopping plugin", 
			zap.String("plugin_id", id),
			zap.Error(err))
	}

	// Remove from plugins map
	delete(pm.plugins, id)

	pm.logger.Info("Successfully unloaded plugin", zap.String("plugin_id", id))
	return nil
}

// GetPlugin returns a plugin with the given ID
func (pm *PluginManager) GetPlugin(id string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[id]
	return plugin, exists
}

// ListPlugins returns a list of all loaded plugins
func (pm *PluginManager) ListPlugins() []*PluginInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var plugins []*PluginInfo
	for _, plugin := range pm.plugins {
		info := plugin.GetInfo()
		plugins = append(plugins, info)
	}

	return plugins
}

// LoadAllPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadAllPlugins(ctx context.Context) error {
	if pm.config.Directory == "" {
		return fmt.Errorf("plugin directory not configured")
	}

	// Walk through the plugin directory
	return filepath.Walk(pm.config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only load .so files
		if filepath.Ext(path) != ".so" {
			return nil
		}

		// Load the plugin
		if err := pm.LoadPlugin(ctx, path); err != nil {
			pm.logger.Error("Failed to load plugin", 
				zap.String("path", path),
				zap.Error(err))
			// Continue loading other plugins even if one fails
			return nil
		}

		return nil
	})
}

// Start starts the plugin manager
func (pm *PluginManager) Start(ctx context.Context) error {
	pm.logger.Info("Starting plugin manager")

	// Auto-load plugins if configured
	if pm.config.AutoLoad {
		if err := pm.LoadAllPlugins(ctx); err != nil {
			return fmt.Errorf("failed to load plugins: %v", err)
		}
	}

	// Start health check routine if configured
	if pm.config.HealthCheck {
		go pm.healthCheckRoutine(ctx)
	}

	pm.logger.Info("Plugin manager started")
	return nil
}

// Stop stops the plugin manager
func (pm *PluginManager) Stop(ctx context.Context) error {
	pm.logger.Info("Stopping plugin manager")

	// Stop all plugins
	for id, plugin := range pm.plugins {
		if err := plugin.Stop(ctx); err != nil {
			pm.logger.Error("Error stopping plugin", 
				zap.String("plugin_id", id),
				zap.Error(err))
		}
	}

	pm.logger.Info("Plugin manager stopped")
	return nil
}

// healthCheckRoutine performs periodic health checks on all plugins
func (pm *PluginManager) healthCheckRoutine(ctx context.Context) {
	ticker := time.NewTicker(pm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks performs health checks on all plugins
func (pm *PluginManager) performHealthChecks(ctx context.Context) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for id, plugin := range pm.plugins {
		if err := plugin.HealthCheck(ctx); err != nil {
			pm.logger.Error("Health check failed for plugin", 
				zap.String("plugin_id", id),
				zap.Error(err))
		}
	}
}

// EnablePlugin enables a plugin
func (pm *PluginManager) EnablePlugin(ctx context.Context, id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[id]
	if !exists {
		return fmt.Errorf("plugin with ID %s not found", id)
	}

	info := plugin.GetInfo()
	info.Enabled = true
	info.UpdatedAt = time.Now()

	pm.logger.Info("Enabled plugin", zap.String("plugin_id", id))
	return nil
}

// DisablePlugin disables a plugin
func (pm *PluginManager) DisablePlugin(ctx context.Context, id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[id]
	if !exists {
		return fmt.Errorf("plugin with ID %s not found", id)
	}

	info := plugin.GetInfo()
	info.Enabled = false
	info.UpdatedAt = time.Now()

	pm.logger.Info("Disabled plugin", zap.String("plugin_id", id))
	return nil
}

// GetPluginsByType returns plugins of the specified type
func (pm *PluginManager) GetPluginsByType(pluginType PluginType) []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var plugins []Plugin
	for _, plugin := range pm.plugins {
		info := plugin.GetInfo()
		if info.Type == pluginType && info.Enabled {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}