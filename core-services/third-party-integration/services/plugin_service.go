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

// PluginService 插件服务
type PluginService struct {
	repo        *repositories.PluginRepository
	pluginDir   string
	sandboxMode bool
}

// NewPluginService 创建新的插件服务
func NewPluginService(repo *repositories.PluginRepository) *PluginService {
	return &PluginService{
		repo:        repo,
		pluginDir:   "./plugins",
		sandboxMode: true,
	}
}

// InstallPlugin 安装插件
func (s *PluginService) InstallPlugin(pluginPath string) (*models.Plugin, error) {
	// 解压插件�?	extractPath, err := s.extractPlugin(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract plugin: %w", err)
	}

	// 读取插件清单
	manifest, err := s.readPluginManifest(extractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	// 验证插件
	if err := s.validatePlugin(manifest); err != nil {
		return nil, fmt.Errorf("plugin validation failed: %w", err)
	}

	// 检查插件是否已存在
	existing, _ := s.repo.GetByName(manifest.Name)
	if existing != nil {
		return nil, fmt.Errorf("plugin %s already installed", manifest.Name)
	}

	// 创建插件记录
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

	// 保存到数据库
	id, err := s.repo.Create(plugin)
	if err != nil {
		// 清理安装文件
		os.RemoveAll(extractPath)
		return nil, fmt.Errorf("failed to save plugin: %w", err)
	}

	plugin.ID = id
	return plugin, nil
}

// UninstallPlugin 卸载插件
func (s *PluginService) UninstallPlugin(id int64) error {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// 如果插件已启用，先禁�?	if plugin.IsEnabled {
		if err := s.DisablePlugin(id); err != nil {
			return fmt.Errorf("failed to disable plugin before uninstall: %w", err)
		}
	}

	// 更新状态为卸载�?	s.repo.Update(id, map[string]interface{}{
		"status":     models.PluginStatusUninstalling,
		"updated_at": time.Now(),
	})

	// 清理插件文件
	if plugin.InstallPath != "" {
		if err := os.RemoveAll(plugin.InstallPath); err != nil {
			// 记录错误但继续卸�?			fmt.Printf("Warning: failed to remove plugin files: %v\n", err)
		}
	}

	// 从数据库删除
	return s.repo.Delete(id)
}

// EnablePlugin 启用插件
func (s *PluginService) EnablePlugin(id int64) error {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	if plugin.IsEnabled {
		return fmt.Errorf("plugin is already enabled")
	}

	// 验证插件依赖
	if err := s.validateDependencies(plugin); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// 加载插件
	if err := s.loadPlugin(plugin); err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// 更新状�?	return s.repo.Update(id, map[string]interface{}{
		"is_enabled": true,
		"status":     models.PluginStatusEnabled,
		"updated_at": time.Now(),
	})
}

// DisablePlugin 禁用插件
func (s *PluginService) DisablePlugin(id int64) error {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	if !plugin.IsEnabled {
		return fmt.Errorf("plugin is already disabled")
	}

	// 卸载插件
	if err := s.unloadPlugin(plugin); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// 更新状�?	return s.repo.Update(id, map[string]interface{}{
		"is_enabled": false,
		"status":     models.PluginStatusDisabled,
		"updated_at": time.Now(),
	})
}

// GetPlugin 获取插件信息
func (s *PluginService) GetPlugin(id int64) (*models.Plugin, error) {
	return s.repo.GetByID(id)
}

// ListPlugins 获取插件列表
func (s *PluginService) ListPlugins(category string, status models.PluginStatus, limit, offset int) ([]*models.Plugin, int64, error) {
	return s.repo.List(category, status, limit, offset)
}

// UpdatePlugin 更新插件
func (s *PluginService) UpdatePlugin(id int64, pluginPath string) error {
	// 获取现有插件
	existingPlugin, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// 解压新版�?	extractPath, err := s.extractPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	// 读取新版本清�?	manifest, err := s.readPluginManifest(extractPath)
	if err != nil {
		os.RemoveAll(extractPath)
		return fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	// 验证是否为同一插件
	if manifest.Name != existingPlugin.Name {
		os.RemoveAll(extractPath)
		return fmt.Errorf("plugin name mismatch")
	}

	// 更新状态为更新�?	s.repo.Update(id, map[string]interface{}{
		"status":     models.PluginStatusUpdating,
		"updated_at": time.Now(),
	})

	// 如果插件已启用，先禁�?	wasEnabled := existingPlugin.IsEnabled
	if wasEnabled {
		if err := s.DisablePlugin(id); err != nil {
			return fmt.Errorf("failed to disable plugin for update: %w", err)
		}
	}

	// 备份旧版�?	backupPath := existingPlugin.InstallPath + ".backup"
	if err := os.Rename(existingPlugin.InstallPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old version: %w", err)
	}

	// 移动新版本到安装目录
	if err := os.Rename(extractPath, existingPlugin.InstallPath); err != nil {
		// 恢复备份
		os.Rename(backupPath, existingPlugin.InstallPath)
		return fmt.Errorf("failed to install new version: %w", err)
	}

	// 更新数据库记�?	updates := map[string]interface{}{
		"version":     manifest.Version,
		"description": manifest.Description,
		"config":      manifest.Config,
		"manifest":    *manifest,
		"status":      models.PluginStatusInstalled,
		"updated_at":  time.Now(),
	}

	if err := s.repo.Update(id, updates); err != nil {
		// 恢复旧版�?		os.RemoveAll(existingPlugin.InstallPath)
		os.Rename(backupPath, existingPlugin.InstallPath)
		return fmt.Errorf("failed to update plugin record: %w", err)
	}

	// 如果之前已启用，重新启用
	if wasEnabled {
		if err := s.EnablePlugin(id); err != nil {
			return fmt.Errorf("failed to re-enable plugin after update: %w", err)
		}
	}

	// 清理备份
	os.RemoveAll(backupPath)

	return nil
}

// GetPluginConfig 获取插件配置
func (s *PluginService) GetPluginConfig(id int64) (map[string]interface{}, error) {
	plugin, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}
	return plugin.Config, nil
}

// UpdatePluginConfig 更新插件配置
func (s *PluginService) UpdatePluginConfig(id int64, config map[string]interface{}) error {
	return s.repo.Update(id, map[string]interface{}{
		"config":     config,
		"updated_at": time.Now(),
	})
}

// extractPlugin 解压插件�?func (s *PluginService) extractPlugin(pluginPath string) (string, error) {
	// 创建临时目录
	extractPath := filepath.Join(s.pluginDir, fmt.Sprintf("temp_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return "", err
	}

	// 打开zip文件
	reader, err := zip.OpenReader(pluginPath)
	if err != nil {
		os.RemoveAll(extractPath)
		return "", err
	}
	defer reader.Close()

	// 解压文件
	for _, file := range reader.File {
		path := filepath.Join(extractPath, file.Name)
		
		// 安全检查：防止路径遍历攻击
		if !strings.HasPrefix(path, extractPath) {
			os.RemoveAll(extractPath)
			return "", fmt.Errorf("invalid file path in archive: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// 创建文件目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			os.RemoveAll(extractPath)
			return "", err
		}

		// 解压文件
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

// readPluginManifest 读取插件清单
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

// validatePlugin 验证插件
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

// validateDependencies 验证插件依赖
func (s *PluginService) validateDependencies(plugin *models.Plugin) error {
	for depName, depVersion := range plugin.Manifest.Dependencies {
		dep, err := s.repo.GetByName(depName)
		if err != nil {
			return fmt.Errorf("dependency %s not found", depName)
		}
		if !dep.IsEnabled {
			return fmt.Errorf("dependency %s is not enabled", depName)
		}
		// 这里可以添加版本兼容性检�?		_ = depVersion
	}
	return nil
}

// loadPlugin 加载插件
func (s *PluginService) loadPlugin(plugin *models.Plugin) error {
	// 这里实现插件加载逻辑
	// 可以使用Go的plugin包或其他插件系统
	fmt.Printf("Loading plugin: %s\n", plugin.Name)
	return nil
}

// unloadPlugin 卸载插件
func (s *PluginService) unloadPlugin(plugin *models.Plugin) error {
	// 这里实现插件卸载逻辑
	fmt.Printf("Unloading plugin: %s\n", plugin.Name)
	return nil
}

// 辅助函数
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
