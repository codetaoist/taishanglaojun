package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// ModelManager 模型服务管理器
type ModelManager struct {
	services map[string]ModelService
	factory  ModelServiceFactory
	configs  map[string]*models.ModelConfig
	mu       sync.RWMutex
}

// NewModelManager 创建模型服务管理器
func NewModelManager() *ModelManager {
	return &ModelManager{
		services: make(map[string]ModelService),
		factory:  NewModelServiceFactory(),
		configs:  make(map[string]*models.ModelConfig),
	}
}

// RegisterService 注册模型服务
func (m *ModelManager) RegisterService(name string, service ModelService) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.services[name]; exists {
		return fmt.Errorf("service already registered: %s", name)
	}
	
	m.services[name] = service
	return nil
}

// GetService 获取模型服务
func (m *ModelManager) GetService(name string) (ModelService, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	service, exists := m.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}
	
	return service, nil
}

// CreateService 创建模型服务
func (m *ModelManager) CreateService(config *models.ModelConfig) (ModelService, error) {
	service, err := m.factory.CreateService(config)
	if err != nil {
		return nil, err
	}
	
	// 注册服务
	if err := m.RegisterService(config.Name, service); err != nil {
		return nil, err
	}
	
	// 保存配置
	m.mu.Lock()
	m.configs[config.Name] = config
	m.mu.Unlock()
	
	return service, nil
}

// ListServices 列出所有服务
func (m *ModelManager) ListServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	names := make([]string, 0, len(m.services))
	for name := range m.services {
		names = append(names, name)
	}
	
	return names
}

// RemoveService 移除服务
func (m *ModelManager) RemoveService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	service, exists := m.services[name]
	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}
	
	// 断开连接
	if err := service.Disconnect(context.Background()); err != nil {
		return fmt.Errorf("failed to disconnect service: %w", err)
	}
	
	delete(m.services, name)
	delete(m.configs, name)
	
	return nil
}

// GetConfig 获取服务配置
func (m *ModelManager) GetConfig(name string) (*models.ModelConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("config not found: %s", name)
	}
	
	return config, nil
}

// UpdateConfig 更新服务配置
func (m *ModelManager) UpdateConfig(config *models.ModelConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 检查服务是否存在
	service, exists := m.services[config.Name]
	if !exists {
		return fmt.Errorf("service not found: %s", config.Name)
	}
	
	// 更新配置
	m.configs[config.Name] = config
	
	// 重新连接服务
	if err := service.Disconnect(context.Background()); err != nil {
		return fmt.Errorf("failed to disconnect service: %w", err)
	}
	
	if err := service.Connect(context.Background(), config); err != nil {
		return fmt.Errorf("failed to connect service: %w", err)
	}
	
	return nil
}

// HealthCheck 健康检查
func (m *ModelManager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	results := make(map[string]error)
	for name, service := range m.services {
		results[name] = service.Health(ctx)
	}
	
	return results
}

// Close 关闭所有服务
func (m *ModelManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var lastErr error
	for name, service := range m.services {
		if err := service.Disconnect(context.Background()); err != nil {
			lastErr = fmt.Errorf("failed to disconnect service %s: %w", name, err)
		}
	}
	
	// 清空所有服务
	m.services = make(map[string]ModelService)
	m.configs = make(map[string]*models.ModelConfig)
	
	return lastErr
}

// GetDefaultService 获取默认服务
func (m *ModelManager) GetDefaultService() (ModelService, error) {
	// 优先查找名为 "default" 的服务
	if service, err := m.GetService("default"); err == nil {
		return service, nil
	}
	
	// 如果没有默认服务，返回第一个注册的服务
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, service := range m.services {
		return service, nil
	}
	
	return nil, fmt.Errorf("no services available")
}

// GetServiceByProvider 根据提供商获取服务
func (m *ModelManager) GetServiceByProvider(provider string) (ModelService, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for name, config := range m.configs {
		if config.Provider() == provider {
			if service, exists := m.services[name]; exists {
				return service, nil
			}
		}
	}
	
	return nil, fmt.Errorf("service not found for provider: %s", provider)
}

// GetServiceByModel 根据模型名称获取服务
func (m *ModelManager) GetServiceByModel(modelID string) (ModelService, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for name, config := range m.configs {
		if config.ModelID == modelID {
			if service, exists := m.services[name]; exists {
				return service, nil
			}
		}
	}
	
	return nil, fmt.Errorf("service not found for model: %s", modelID)
}