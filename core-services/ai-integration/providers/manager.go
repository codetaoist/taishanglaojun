package providers

import (
	"context"
	"fmt"
	"sync"
)

// Manager AI提供商管理器
type Manager struct {
	providers map[string]AIProvider
	default   string
	mu        sync.RWMutex
}

// NewManager 创建提供商管理器
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]AIProvider),
	}
}

// RegisterProvider 注册提供�?func (m *Manager) RegisterProvider(name string, provider AIProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid provider config for %s: %w", name, err)
	}
	
	m.providers[name] = provider
	
	// 如果是第一个提供商，设为默�?	if m.default == "" {
		m.default = name
	}
	
	return nil
}

// GetProvider 获取提供�?func (m *Manager) GetProvider(name string) (AIProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if name == "" {
		name = m.default
	}
	
	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	
	return provider, nil
}

// GetDefaultProvider 获取默认提供�?func (m *Manager) GetDefaultProvider() (AIProvider, error) {
	return m.GetProvider("")
}

// SetDefault 设置默认提供�?func (m *Manager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}
	
	m.default = name
	return nil
}

// ListProviders 列出所有提供商
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	
	return names
}

// Chat 发送对话消�?func (m *Manager) Chat(ctx context.Context, providerName string, req *ChatRequest) (*ChatResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.Chat(ctx, req)
}

// GetProviderModels 获取提供商支持的模型
func (m *Manager) GetProviderModels(providerName string) ([]string, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.GetModels(), nil
}

// GetAllModels 获取所有提供商的模�?func (m *Manager) GetAllModels() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	models := make(map[string][]string)
	for name, provider := range m.providers {
		models[name] = provider.GetModels()
	}
	
	return models
}

// HealthCheck 健康检�?func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	results := make(map[string]error)
	
	for name, provider := range m.providers {
		// 发送一个简单的测试请求
		testReq := &ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			Model: provider.GetModels()[0], // 使用第一个模�?		}
		
		_, err := provider.Chat(ctx, testReq)
		results[name] = err
	}
	
	return results
}
