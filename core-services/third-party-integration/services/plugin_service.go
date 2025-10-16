package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/repositories"
)

// PluginService 
type PluginService struct {
	repo        *repositories.PluginRepository
	pluginDir   string
	sandboxMode bool
}

// NewPluginService 
func NewPluginService(repo *repositories.PluginRepository) *PluginService {
	return &PluginService{
		repo:        repo,
		pluginDir:   "./plugins",
		sandboxMode: true,
	}
}

// InstallPlugin 
func (s *PluginService) InstallPlugin(pluginPath string) (*models.Plugin, error) {
	// ?	extractPath, err := s.extractPlugin(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract plugin: %w", err)
	}

	// 嵥
	manifest, err := s.readPluginManifest(extractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	// 
	if err := s.validatePlugin(manifest); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %w", err)
	}

	// 
	existing, _ := s.repo.GetByName(manifest.Name)
	if existing != nil {
		return nil, fmt.Errorf("plugin %s already installed", manifest.Name)
	}

	// 
	plugin := &models.Plugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      extractAuthorFromManifest(manifest),
		License:     extractLicenseFromManifest(manifest),
		Category:    extractCategoryFromManifest(manifest),
		Status:      models.PluginStatusInstalled,
		Config:      manifest.Config,
		Manifest:    *manifest,
		InstallPath: extractPath,
		IsEnabled:   false,
		InstallDate: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 浽
	id, err := s.repo.Create(plugin)
	if err != nil {
		// 
		os.RemoveAll(extractPath)
		return nil, fmt.Errorf("failed to save plugin: %w", err)
	}

	plugin.ID = id
	return plugin, nil
}

// UninstallPlugin 
func (s *PluginService) UninstallPlugin(id int64) error {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// ?	if plugin.IsEnabled {
		if err := s.DisablePlugin(id); err != nil {
			return fmt.Errorf("failed to disable plugin before uninstall: %w", err)
		}
	}

	// ?	s.repo.Update(id, map[string]interface{}{
		"status":     models.PluginStatusUninstalling,
		"updated_at": time.Now(),
	})

	// 
	if plugin.InstallPath != "" {
		if err := os.RemoveAll(plugin.InstallPath); err != nil {
			// ?			fmt.Printf("Warning: failed to remove plugin files: %v\n", err)
		}
	}

	// 
	return s.repo.Delete(id)
}

// EnablePlugin 
func (s *PluginService) EnablePlugin(id int64) error {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	if plugin.IsEnabled {
		return fmt.Errorf("plugin is already enabled")
	}

	// 
	if err := s.validateDependencies(plugin); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// 
	if err := s.loadPlugin(plugin); err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// ?	return s.repo.Update(id, map[string]interface{}{
		"is_enabled": true,
		"status":     models.PluginStatusEnabled,
		"updated_at": time.Now(),
	})
}

// DisablePlugin 
func (s *PluginService) DisablePlugin(id int64) error {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	if !plugin.IsEnabled {
		return fmt.Errorf("plugin is already disabled")
	}

	// 
	if err := s.unloadPlugin(plugin); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// ?	return s.repo.Update(id, map[string]interface{}{
		"is_enabled": false,
		"status":     models.PluginStatusDisabled,
		"updated_at": time.Now(),
	})
}

// GetPlugin 
func (s *PluginService) GetPlugin(id int64) (*models.Plugin, error) {
	return s.repo.GetByID(id)
}

// ListPlugins 
func (s *PluginService) ListPlugins(category string, status models.PluginStatus, limit, offset int) ([]*models.Plugin, int64, error) {
	return s.repo.List(category, status, limit, offset)
}

// UpdatePlugin 
func (s *PluginService) UpdatePlugin(id int64, pluginPath string) error {
	// 
	existingPlugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// ?	extractPath, err := s.extractPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	// 汾?	manifest, err := s.readPluginManifest(extractPath)
	if err != nil {
		os.RemoveAll(extractPath)
		return fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	// 
	if manifest.Name != existingPlugin.Name {
		os.RemoveAll(extractPath)
		return fmt.Errorf("plugin name mismatch")
	}

	// ?	s.repo.Update(id, map[string]interface{}{
		"status":     models.PluginStatusUpdating,
		"updated_at": time.Now(),
	})

	// ?	wasEnabled := existingPlugin.IsEnabled
	if wasEnabled {
		if err := s.DisablePlugin(id); err != nil {
			return fmt.Errorf("failed to disable plugin for update: %w", err)
		}
	}

	// ?	backupPath := existingPlugin.InstallPath + ".backup"
	if err := os.Rename(existingPlugin.InstallPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old version: %w", err)
	}

	// 汾
	if err := os.Rename(extractPath, existingPlugin.InstallPath); err != nil {
		// 
		os.Rename(backupPath, existingPlugin.InstallPath)
		return fmt.Errorf("failed to install new version: %w", err)
	}

	// ?	updates := map[string]interface{}{
		"version":     manifest.Version,
		"description": manifest.Description,
		"config":      manifest.Config,
		"manifest":    *manifest,
		"status":      models.PluginStatusInstalled,
		"updated_at":  time.Now(),
	}

	if err := s.repo.Update(id, updates); err != nil {
		// ?		os.RemoveAll(existingPlugin.InstallPath)
		os.Rename(backupPath, existingPlugin.InstallPath)
		return fmt.Errorf("failed to update plugin record: %w", err)
	}

	// 
	if wasEnabled {
		if err := s.EnablePlugin(id); err != nil {
			return fmt.Errorf("failed to re-enable plugin after update: %w", err)
		}
	}

	// 
	os.RemoveAll(backupPath)

	return nil
}

// GetPluginConfig 
func (s *PluginService) GetPluginConfig(id int64) (map[string]interface{}, error) {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}
	return plugin.Config, nil
}

// UpdatePluginConfig 
func (s *PluginService) UpdatePluginConfig(id int64, config map[string]interface{}) error {
	return s.repo.Update(id, map[string]interface{}{
		"config":     config,
		"updated_at": time.Now(),
	})
}

// extractPlugin ?func (s *PluginService) extractPlugin(pluginPath string) (string, error) {
	// 
	extractPath := filepath.Join(s.pluginDir, fmt.Sprintf("temp_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return "", err
	}

	// zip
	reader, err := zip.OpenReader(pluginPath)
	if err != nil {
		os.RemoveAll(extractPath)
		return "", err
	}
	defer reader.Close()

	// 
	for _, file := range reader.File {
		path := filepath.Join(extractPath, file.Name)
		
		// 
		if !strings.HasPrefix(path, extractPath) {
			os.RemoveAll(extractPath)
			return "", fmt.Errorf("invalid file path in archive: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// 
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			os.RemoveAll(extractPath)
			return "", err
		}

		// 
		fileReader, err := file.Open()
		if err != nil {
			os.RemoveAll(extractPath)
			return "", err
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			os.RemoveAll(extractPath)
			return "", err
		}

		_, err = io.Copy(targetFile, fileReader)
		fileReader.Close()
		targetFile.Close()

		if err != nil {
			os.RemoveAll(extractPath)
			return "", err
		}
	}

	return extractPath, nil
}

// readPluginManifest 嵥
func (s *PluginService) readPluginManifest(pluginPath string) (*models.PluginManifest, error) {
	manifestPath := filepath.Join(pluginPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest models.PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// validatePlugin 
func (s *PluginService) validatePlugin(manifest *models.PluginManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.Main == "" {
		return fmt.Errorf("plugin main file is required")
	}
	return nil
}

// validateDependencies 
func (s *PluginService) validateDependencies(plugin *models.Plugin) error {
	for depName, depVersion := range plugin.Manifest.Dependencies {
		dep, err := s.repo.GetByName(depName)
		if err != nil {
			return fmt.Errorf("dependency %s not found", depName)
		}
		if !dep.IsEnabled {
			return fmt.Errorf("dependency %s is not enabled", depName)
		}
		// 汾?		_ = depVersion
	}
	return nil
}

// loadPlugin 
func (s *PluginService) loadPlugin(plugin *models.Plugin) error {
	// 
	// Goplugin
	fmt.Printf("Loading plugin: %s\n", plugin.Name)
	return nil
}

// unloadPlugin 
func (s *PluginService) unloadPlugin(plugin *models.Plugin) error {
	// 
	fmt.Printf("Unloading plugin: %s\n", plugin.Name)
	return nil
}

// 
func extractAuthorFromManifest(manifest *models.PluginManifest) string {
	if author, ok := manifest.Config["author"].(string); ok {
		return author
	}
	return ""
}

func extractLicenseFromManifest(manifest *models.PluginManifest) string {
	if license, ok := manifest.Config["license"].(string); ok {
		return license
	}
	return ""
}

func extractCategoryFromManifest(manifest *models.PluginManifest) string {
	if category, ok := manifest.Config["category"].(string); ok {
		return category
	}
	return "general"
}

