package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"gorm.io/gorm"
)

// ModelConfigManager 模型配置管理器
type ModelConfigManager struct {
	db     *gorm.DB
	config *Config
}

// NewModelConfigManager 创建模型配置管理器
func NewModelConfigManager(db *gorm.DB, config *Config) *ModelConfigManager {
	return &ModelConfigManager{
		db:     db,
		config: config,
	}
}

// InitializeDefaultConfigs 初始化默认模型配置
func (m *ModelConfigManager) InitializeDefaultConfigs() error {
	// 检查是否已有配置
	var count int64
	if err := m.db.Model(&models.ModelConfig{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check model configs: %w", err)
	}

	if count > 0 {
		// 已有配置，跳过初始化
		return nil
	}

	// 创建默认配置
	defaultConfigs := []*models.ModelConfig{
		{
			Name:        "default-openai",
			ServiceType: "openai",
			ModelID:     "gpt-3.5-turbo",
			APIKey:      os.Getenv("OPENAI_API_KEY"),
			Endpoint:    "https://api.openai.com/v1",
			MaxTokens:   4096,
			Temperature: 0.7,
			IsDefault:   false,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "default-ollama",
			ServiceType: "ollama",
			ModelID:     "llama2",
			Endpoint:    "http://localhost:11434",
			MaxTokens:   4096,
			Temperature: 0.7,
			IsDefault:   false,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, config := range defaultConfigs {
		if err := m.db.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create default model config %s: %w", config.Name, err)
		}
	}

	return nil
}

// LoadConfigFromFile 从文件加载配置
func (m *ModelConfigManager) LoadConfigFromFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，创建默认配置文件
		return m.createDefaultConfigFile(filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析JSON
	var configs []*models.ModelConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// 保存到数据库
	for _, config := range configs {
		// 检查是否已存在
		var existing models.ModelConfig
		err := m.db.Where("name = ?", config.Name).First(&existing).Error
		if err == nil {
			// 已存在，更新
			config.ID = existing.ID
			config.UpdatedAt = time.Now()
			if err := m.db.Save(config).Error; err != nil {
				return fmt.Errorf("failed to update model config %s: %w", config.Name, err)
			}
		} else if err == gorm.ErrRecordNotFound {
			// 不存在，创建
			config.CreatedAt = time.Now()
			config.UpdatedAt = time.Now()
			if err := m.db.Create(config).Error; err != nil {
				return fmt.Errorf("failed to create model config %s: %w", config.Name, err)
			}
		} else {
			return fmt.Errorf("failed to check model config %s: %w", config.Name, err)
		}
	}

	return nil
}

// SaveConfigToFile 保存配置到文件
func (m *ModelConfigManager) SaveConfigToFile(filePath string) error {
	// 获取所有配置
	var configs []*models.ModelConfig
	if err := m.db.Find(&configs).Error; err != nil {
		return fmt.Errorf("failed to get model configs: %w", err)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configs: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// createDefaultConfigFile 创建默认配置文件
func (m *ModelConfigManager) createDefaultConfigFile(filePath string) error {
	defaultConfigs := []*models.ModelConfig{
		{
			Name:        "default-openai",
			ServiceType: "openai",
			ModelID:     "gpt-3.5-turbo",
			APIKey:      os.Getenv("OPENAI_API_KEY"),
			Endpoint:    "https://api.openai.com/v1",
			MaxTokens:   4096,
			Temperature: 0.7,
			IsDefault:   false,
			IsActive:    true,
		},
		{
			Name:        "default-ollama",
			ServiceType: "ollama",
			ModelID:     "llama2",
			Endpoint:    "http://localhost:11434",
			MaxTokens:   4096,
			Temperature: 0.7,
			IsDefault:   false,
			IsActive:    true,
		},
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(defaultConfigs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default configs: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}

// GetEnabledConfigs 获取启用的配置
func (m *ModelConfigManager) GetEnabledConfigs() ([]*models.ModelConfig, error) {
	var configs []*models.ModelConfig
	if err := m.db.Where("enabled = ?", true).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get enabled configs: %w", err)
	}
	return configs, nil
}

// GetConfigByName 根据名称获取配置
func (m *ModelConfigManager) GetConfigByName(name string) (*models.ModelConfig, error) {
	var config models.ModelConfig
	if err := m.db.Where("name = ?", name).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("config not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return &config, nil
}

// UpdateConfig 更新配置
func (m *ModelConfigManager) UpdateConfig(config *models.ModelConfig) error {
	config.UpdatedAt = time.Now()
	if err := m.db.Save(config).Error; err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	return nil
}

// DeleteConfig 删除配置
func (m *ModelConfigManager) DeleteConfig(name string) error {
	if err := m.db.Where("name = ?", name).Delete(&models.ModelConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	return nil
}