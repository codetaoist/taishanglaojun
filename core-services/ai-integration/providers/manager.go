package providers

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Manager AI提供商管理器
type Manager struct {
	providers       map[string]AIProvider
	defaultProvider string
	logger          *zap.Logger
	mu              sync.RWMutex
}

// NewManager 创建新的管理�?
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		providers: make(map[string]AIProvider),
		logger:    logger,
	}
}

// RegisterProvider 注册提供�?
func (m *Manager) RegisterProvider(name string, provider AIProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	m.providers[name] = provider
	m.logger.Info("Provider registered", zap.String("name", name))
	return nil
}

// SetDefaultProvider 设置默认提供�?
func (m *Manager) SetDefaultProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	m.defaultProvider = name
	m.logger.Info("Default provider set", zap.String("name", name))
	return nil
}

// GetProvider 获取提供�?
func (m *Manager) GetProvider(name string) (AIProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetDefaultProvider 获取默认提供�?
func (m *Manager) GetDefaultProvider() (AIProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultProvider == "" {
		return nil, fmt.Errorf("no default provider set")
	}

	provider, exists := m.providers[m.defaultProvider]
	if !exists {
		return nil, fmt.Errorf("default provider %s not found", m.defaultProvider)
	}

	return provider, nil
}

// Chat 使用指定提供商进行对�?
func (m *Manager) Chat(ctx context.Context, providerName string, request ChatRequest) (*ChatResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.Chat(ctx, request)
}

// Generate 使用指定提供商生成内�?
func (m *Manager) Generate(ctx context.Context, providerName string, request GenerateRequest) (*GenerateResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.Generate(ctx, request)
}

// Analyze 使用指定提供商分析内�?
func (m *Manager) Analyze(ctx context.Context, providerName string, request AnalyzeRequest) (*AnalyzeResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.Analyze(ctx, request)
}

// Embed 使用指定提供商生成嵌入向�?
func (m *Manager) Embed(ctx context.Context, providerName string, text string) ([]float32, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.Embed(ctx, text)
}

// RemoveProvider 移除提供�?
func (m *Manager) RemoveProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	delete(m.providers, name)

	// 如果移除的是默认提供商，清空默认设置
	if m.defaultProvider == name {
		m.defaultProvider = ""
	}

	m.logger.Info("Provider removed", zap.String("name", name))
	return nil
}

// IsHealthy 检查提供商健康状�?
func (m *Manager) IsHealthy(name string) bool {
	provider, err := m.GetProvider(name)
	if err != nil {
		return false
	}

	// 这里可以添加健康检查逻辑
	// 暂时返回true
	_ = provider
	return true
}

// GetProviders 获取所有提供商名称
func (m *Manager) GetProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetModels 获取所有提供商的模型列�?
func (m *Manager) GetModels() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	models := make(map[string][]string)
	for name, provider := range m.providers {
		models[name] = provider.GetModels()
	}
	return models
}

// GetProviderInfo 获取提供商信�?
func (m *Manager) GetProviderInfo() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := make(map[string]interface{})
	info["total_providers"] = len(m.providers)
	info["default_provider"] = m.defaultProvider

	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}
	info["providers"] = providers

	return info
}

