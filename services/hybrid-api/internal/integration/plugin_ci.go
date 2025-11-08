package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/services/api/internal/ci"
	"github.com/codetaoist/services/api/internal/middleware"
	"github.com/codetaoist/services/api/internal/plugin"
)

// PluginCIIntegration handles integration between plugin system and CI/CD pipeline
type PluginCIIntegration struct {
	pluginManager *plugin.PluginManager
	ciPipeline    *ci.CIPipeline
	logger        *middleware.Logger
	mu            sync.RWMutex
	
	// Plugin build tracking
	pluginBuilds map[string]*PluginBuildStatus
	
	// CI/CD triggers for plugins
	ciTriggers map[string]*CITriggerConfig
	
	// Integration configuration
	config *IntegrationConfig
}

// PluginBuildStatus tracks the build status of a plugin
type PluginBuildStatus struct {
	PluginID      string
	Version       string
	BuildID       string
	Status        string // "pending", "running", "success", "failed"
	StartTime     time.Time
	EndTime       *time.Time
	BuildLog      string
	ArtifactURL   string
	Error         error
}

// CITriggerConfig defines CI/CD triggers for plugin events
type CITriggerConfig struct {
	PluginID      string
	TriggerEvent  string // "install", "update", "uninstall"
	PipelineID    string
	AutoTrigger   bool
	Parameters    map[string]interface{}
}

// IntegrationConfig holds configuration for plugin-CI integration
type IntegrationConfig struct {
	// Auto-build plugins on install/update
	AutoBuildPlugins bool
	
	// Build timeout in seconds
	BuildTimeout int
	
	// Artifact retention in days
	ArtifactRetention int
	
	// Enable plugin versioning with CI
	EnableVersioning bool
	
	// CI pipeline template for plugins
	DefaultPipelineTemplate string
	
	// Plugin registry for built artifacts
	RegistryURL string
}

// NewPluginCIIntegration creates a new instance of PluginCIIntegration
func NewPluginCIIntegration(
	pluginManager *plugin.PluginManager,
	ciPipeline *ci.CIPipeline,
	logger *middleware.Logger,
	config *IntegrationConfig,
) *PluginCIIntegration {
	if config == nil {
		config = &IntegrationConfig{
			AutoBuildPlugins:       true,
			BuildTimeout:          1800, // 30 minutes
			ArtifactRetention:     30,   // 30 days
			EnableVersioning:       true,
			DefaultPipelineTemplate: "plugin-build",
			RegistryURL:           "",
		}
	}
	
	return &PluginCIIntegration{
		pluginManager: pluginManager,
		ciPipeline:    ciPipeline,
		logger:        logger,
		pluginBuilds:  make(map[string]*PluginBuildStatus),
		ciTriggers:    make(map[string]*CITriggerConfig),
		config:        config,
	}
}

// Start initializes the integration
func (i *PluginCIIntegration) Start(ctx context.Context) error {
	i.logger.Info("Starting plugin-CI/CD integration")
	
	// Register plugin event handlers
	if err := i.registerPluginEventHandlers(); err != nil {
		return fmt.Errorf("failed to register plugin event handlers: %w", err)
	}
	
	// Initialize default CI triggers
	if err := i.initializeDefaultTriggers(); err != nil {
		return fmt.Errorf("failed to initialize default triggers: %w", err)
	}
	
	// Start cleanup routine for old build artifacts
	go i.cleanupRoutine(ctx)
	
	return nil
}

// Stop stops the integration
func (i *PluginCIIntegration) Stop(ctx context.Context) error {
	i.logger.Info("Stopping plugin-CI/CD integration")
	return nil
}

// registerPluginEventHandlers registers handlers for plugin lifecycle events
func (i *PluginCIIntegration) registerPluginEventHandlers() error {
	// Register plugin install handler
	i.pluginManager.RegisterEventHandler("install", i.handlePluginInstall)
	
	// Register plugin update handler
	i.pluginManager.RegisterEventHandler("update", i.handlePluginUpdate)
	
	// Register plugin uninstall handler
	i.pluginManager.RegisterEventHandler("uninstall", i.handlePluginUninstall)
	
	return nil
}

// initializeDefaultTriggers sets up default CI triggers for common plugin events
func (i *PluginCIIntegration) initializeDefaultTriggers() error {
	// Default trigger for plugin install
	installTrigger := &CITriggerConfig{
		TriggerEvent:  "install",
		PipelineID:    i.config.DefaultPipelineTemplate,
		AutoTrigger:   i.config.AutoBuildPlugins,
		Parameters: map[string]interface{}{
			"build_timeout": i.config.BuildTimeout,
			"versioning":    i.config.EnableVersioning,
		},
	}
	i.ciTriggers["install"] = installTrigger
	
	// Default trigger for plugin update
	updateTrigger := &CITriggerConfig{
		TriggerEvent:  "update",
		PipelineID:    i.config.DefaultPipelineTemplate,
		AutoTrigger:   i.config.AutoBuildPlugins,
		Parameters: map[string]interface{}{
			"build_timeout": i.config.BuildTimeout,
			"versioning":    i.config.EnableVersioning,
		},
	}
	i.ciTriggers["update"] = updateTrigger
	
	return nil
}

// handlePluginInstall handles plugin install events
func (i *PluginCIIntegration) handlePluginInstall(ctx context.Context, pluginID string, metadata map[string]interface{}) error {
	i.logger.Infof("Handling install event for plugin: %s", pluginID)
	
	trigger, exists := i.ciTriggers["install"]
	if !exists || !trigger.AutoTrigger {
		i.logger.Debugf("No auto-trigger configured for plugin install: %s", pluginID)
		return nil
	}
	
	// Create CI pipeline for plugin build
	pipelineID, err := i.createPluginBuildPipeline(ctx, pluginID, "install", metadata)
	if err != nil {
		return fmt.Errorf("failed to create plugin build pipeline: %w", err)
	}
	
	// Start the build
	buildID, err := i.ciPipeline.RunBuild(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to start plugin build: %w", err)
	}
	
	// Track the build status
	i.trackPluginBuild(pluginID, buildID, "install")
	
	return nil
}

// handlePluginUpdate handles plugin update events
func (i *PluginCIIntegration) handlePluginUpdate(ctx context.Context, pluginID string, metadata map[string]interface{}) error {
	i.logger.Infof("Handling update event for plugin: %s", pluginID)
	
	trigger, exists := i.ciTriggers["update"]
	if !exists || !trigger.AutoTrigger {
		i.logger.Debugf("No auto-trigger configured for plugin update: %s", pluginID)
		return nil
	}
	
	// Create CI pipeline for plugin build
	pipelineID, err := i.createPluginBuildPipeline(ctx, pluginID, "update", metadata)
	if err != nil {
		return fmt.Errorf("failed to create plugin build pipeline: %w", err)
	}
	
	// Start the build
	buildID, err := i.ciPipeline.RunBuild(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to start plugin build: %w", err)
	}
	
	// Track the build status
	i.trackPluginBuild(pluginID, buildID, "update")
	
	return nil
}

// handlePluginUninstall handles plugin uninstall events
func (i *PluginCIIntegration) handlePluginUninstall(ctx context.Context, pluginID string, metadata map[string]interface{}) error {
	i.logger.Infof("Handling uninstall event for plugin: %s", pluginID)
	
	// Clean up build artifacts for the plugin
	if err := i.cleanupPluginArtifacts(ctx, pluginID); err != nil {
		i.logger.Errorf("Failed to cleanup artifacts for plugin %s: %v", pluginID, err)
	}
	
	return nil
}

// createPluginBuildPipeline creates a CI pipeline for building a plugin
func (i *PluginCIIntegration) createPluginBuildPipeline(ctx context.Context, pluginID, action string, metadata map[string]interface{}) (string, error) {
	// Get plugin details
	pluginInfo, err := i.pluginManager.GetPluginInfo(pluginID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin info: %w", err)
	}
	
	// Create pipeline configuration
	pipelineConfig := &ci.Pipeline{
		Name:        fmt.Sprintf("plugin-%s-%s", pluginID, action),
		Description: fmt.Sprintf("Build pipeline for plugin %s on %s", pluginID, action),
		Enabled:     true,
		Build: ci.BuildConfig{
			BuildDir:    fmt.Sprintf("plugins/%s", pluginID),
			Dockerfile:  "Dockerfile",
			ImageName:   fmt.Sprintf("plugin-%s", pluginID),
			ImageTag:    pluginInfo.Version,
			Environment: map[string]string{
				"PLUGIN_ID":   pluginID,
				"PLUGIN_VERSION": pluginInfo.Version,
				"ACTION":      action,
			},
			Timeout: time.Duration(i.config.BuildTimeout) * time.Second,
		},
		Test: ci.TestConfig{
			TestDir:     fmt.Sprintf("plugins/%s/tests", pluginID),
			TestPattern: "*_test.go",
			Environment: map[string]string{
				"PLUGIN_ID": pluginID,
			},
			Timeout: time.Duration(i.config.BuildTimeout/3) * time.Second,
		},
		Deploy: ci.DeployConfig{
			Environment: "staging",
			Variables: map[string]string{
				"PLUGIN_ID": pluginID,
				"VERSION":   pluginInfo.Version,
			},
		},
	}
	
	// Create the pipeline
	pipelineID, err := i.ciPipeline.CreatePipeline(ctx, pipelineConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create pipeline: %w", err)
	}
	
	return pipelineID, nil
}

// trackPluginBuild tracks the status of a plugin build
func (i *PluginCIIntegration) trackPluginBuild(pluginID, buildID, action string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	status := &PluginBuildStatus{
		PluginID:  pluginID,
		BuildID:   buildID,
		Status:    "pending",
		StartTime: time.Now(),
	}
	
	i.pluginBuilds[buildID] = status
	
	// Start monitoring the build in a goroutine
	go i.monitorBuildStatus(pluginID, buildID)
}

// monitorBuildStatus monitors the status of a build and updates the plugin build status
func (i *PluginCIIntegration) monitorBuildStatus(pluginID, buildID string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Get build status
			buildInfo, err := i.ciPipeline.GetBuild(context.Background(), buildID)
			if err != nil {
				i.logger.Errorf("Failed to get build status for %s: %v", buildID, err)
				continue
			}
			
			// Update status
			i.mu.Lock()
			status, exists := i.pluginBuilds[buildID]
			if !exists {
				i.mu.Unlock()
				return
			}
			
			status.Status = buildInfo.Status
			if buildInfo.Status == "success" || buildInfo.Status == "failed" {
				now := time.Now()
				status.EndTime = &now
				if buildInfo.Status == "failed" {
					status.Error = fmt.Errorf("build failed")
				}
				i.mu.Unlock()
				
				// Handle build completion
				i.handleBuildCompletion(pluginID, buildID, buildInfo.Status)
				return
			}
			i.mu.Unlock()
		}
	}
}

// handleBuildCompletion handles the completion of a plugin build
func (i *PluginCIIntegration) handleBuildCompletion(pluginID, buildID, status string) {
	i.logger.Infof("Plugin build %s completed with status: %s", buildID, status)
	
	if status == "success" {
		// Get build artifacts
		artifacts, err := i.ciPipeline.GetBuildArtifacts(context.Background(), buildID)
		if err != nil {
			i.logger.Errorf("Failed to get build artifacts for %s: %v", buildID, err)
			return
		}
		
		// Update plugin with artifact information
		i.mu.Lock()
		buildStatus, exists := i.pluginBuilds[buildID]
		if exists && len(artifacts) > 0 {
			buildStatus.ArtifactURL = artifacts[0].URL
		}
		i.mu.Unlock()
		
		// Notify plugin manager of successful build
		if err := i.pluginManager.NotifyBuildCompletion(pluginID, buildID, status, artifacts); err != nil {
			i.logger.Errorf("Failed to notify plugin manager of build completion: %v", err)
		}
	}
}

// cleanupPluginArtifacts cleans up build artifacts for a plugin
func (i *PluginCIIntegration) cleanupPluginArtifacts(ctx context.Context, pluginID string) error {
	i.logger.Infof("Cleaning up artifacts for plugin: %s", pluginID)
	
	// Find all builds for this plugin
	var buildsToDelete []string
	i.mu.RLock()
	for buildID, status := range i.pluginBuilds {
		if status.PluginID == pluginID {
			buildsToDelete = append(buildsToDelete, buildID)
		}
	}
	i.mu.RUnlock()
	
	// Delete builds
	for _, buildID := range buildsToDelete {
		if err := i.ciPipeline.DeleteBuild(ctx, buildID); err != nil {
			i.logger.Errorf("Failed to delete build %s: %v", buildID, err)
		}
		
		// Remove from tracking
		i.mu.Lock()
		delete(i.pluginBuilds, buildID)
		i.mu.Unlock()
	}
	
	return nil
}

// cleanupRoutine runs periodically to clean up old build artifacts
func (i *PluginCIIntegration) cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i.cleanupOldArtifacts(ctx)
		}
	}
}

// cleanupOldArtifacts cleans up build artifacts older than the retention period
func (i *PluginCIIntegration) cleanupOldArtifacts(ctx context.Context) {
	i.logger.Info("Cleaning up old build artifacts")
	
	cutoff := time.Now().AddDate(0, 0, -i.config.ArtifactRetention)
	
	i.mu.Lock()
	defer i.mu.Unlock()
	
	for buildID, status := range i.pluginBuilds {
		if status.StartTime.Before(cutoff) {
			if err := i.ciPipeline.DeleteBuild(ctx, buildID); err != nil {
				i.logger.Errorf("Failed to delete old build %s: %v", buildID, err)
			} else {
				delete(i.pluginBuilds, buildID)
			}
		}
	}
}

// GetPluginBuildStatus returns the build status for a plugin
func (i *PluginCIIntegration) GetPluginBuildStatus(pluginID string) ([]*PluginBuildStatus, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	var statuses []*PluginBuildStatus
	for _, status := range i.pluginBuilds {
		if status.PluginID == pluginID {
			statuses = append(statuses, status)
		}
	}
	
	return statuses, nil
}

// TriggerPluginBuild manually triggers a build for a plugin
func (i *PluginCIIntegration) TriggerPluginBuild(ctx context.Context, pluginID string) (string, error) {
	// Get plugin details
	pluginInfo, err := i.pluginManager.GetPluginInfo(pluginID)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin info: %w", err)
	}
	
	// Create CI pipeline for plugin build
	pipelineID, err := i.createPluginBuildPipeline(ctx, pluginID, "manual", map[string]interface{}{
		"version": pluginInfo.Version,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create plugin build pipeline: %w", err)
	}
	
	// Start the build
	buildID, err := i.ciPipeline.RunBuild(ctx, pipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to start plugin build: %w", err)
	}
	
	// Track the build status
	i.trackPluginBuild(pluginID, buildID, "manual")
	
	return buildID, nil
}

// RegisterCITrigger registers a CI trigger for a plugin event
func (i *PluginCIIntegration) RegisterCITrigger(trigger *CITriggerConfig) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	i.ciTriggers[trigger.TriggerEvent] = trigger
	i.logger.Infof("Registered CI trigger for event: %s", trigger.TriggerEvent)
	
	return nil
}

// UnregisterCITrigger unregisters a CI trigger for a plugin event
func (i *PluginCIIntegration) UnregisterCITrigger(event string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	if _, exists := i.ciTriggers[event]; !exists {
		return fmt.Errorf("CI trigger for event %s not found", event)
	}
	
	delete(i.ciTriggers, event)
	i.logger.Infof("Unregistered CI trigger for event: %s", event)
	
	return nil
}