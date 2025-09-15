package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Project 项目信息
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Manager 项目管理器
type Manager struct {
	configPath string
	projects   map[string]*Project
}

// NewManager 创建新的项目管理器
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
		projects:   make(map[string]*Project),
	}
}

// LoadProjects 加载项目配置
func (m *Manager) LoadProjects() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil // 配置文件不存在，返回空项目列表
	}
	
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("读取项目配置失败: %w", err)
	}
	
	var projects []*Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("解析项目配置失败: %w", err)
	}
	
	for _, project := range projects {
		m.projects[project.ID] = project
	}
	
	return nil
}

// SaveProjects 保存项目配置
func (m *Manager) SaveProjects() error {
	var projects []*Project
	for _, project := range m.projects {
		projects = append(projects, project)
	}
	
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化项目配置失败: %w", err)
	}
	
	// 确保配置目录存在
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("保存项目配置失败: %w", err)
	}
	
	return nil
}

// AddProject 添加项目
func (m *Manager) AddProject(project *Project) {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	m.projects[project.ID] = project
}

// GetProject 获取项目
func (m *Manager) GetProject(id string) (*Project, bool) {
	project, exists := m.projects[id]
	return project, exists
}

// ListProjects 列出所有项目
func (m *Manager) ListProjects() []*Project {
	var projects []*Project
	for _, project := range m.projects {
		projects = append(projects, project)
	}
	return projects
}

// RemoveProject 移除项目
func (m *Manager) RemoveProject(id string) bool {
	if _, exists := m.projects[id]; exists {
		delete(m.projects, id)
		return true
	}
	return false
}

// UpdateProject 更新项目
func (m *Manager) UpdateProject(project *Project) {
	project.UpdatedAt = time.Now()
	m.projects[project.ID] = project
}

// GetCurrentProject 获取当前目录绑定的项目
func (m *Manager) GetCurrentProject() (*Project, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %w", err)
	}
	
	for _, project := range m.projects {
		if project.Path == currentDir {
			return project, nil
		}
	}
	
	return nil, fmt.Errorf("当前目录未绑定任何项目")
}

// BindProject 绑定项目到当前目录
func (m *Manager) BindProject(projectID string) error {
	project, exists := m.projects[projectID]
	if !exists {
		return fmt.Errorf("项目 %s 不存在", projectID)
	}
	
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %w", err)
	}
	
	project.Path = currentDir
	project.UpdatedAt = time.Now()
	
	return m.SaveProjects()
}