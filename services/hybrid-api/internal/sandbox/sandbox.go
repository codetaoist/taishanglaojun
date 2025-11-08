package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"go.uber.org/zap"
)

// Sandbox 插件沙箱环境
type Sandbox struct {
	id         string
	rootDir    string
	workDir    string
	network    bool
	resources  ResourceLimits
	allowedPaths []string
	mu         sync.Mutex
	logger     *zap.Logger
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	CPUQuota    int64         // CPU配额 (百分比)
	MemoryLimit int64         // 内存限制 (字节)
	DiskLimit   int64         // 磁盘限制 (字节)
	ProcessLimit int          // 进程数限制
	Network     bool          // 是否允许网络访问
	Timeout     time.Duration // 执行超时
}

// SandboxResult 沙箱执行结果
type SandboxResult struct {
	ExitCode int         `json:"exitCode"`
	Stdout   string      `json:"stdout"`
	Stderr   string      `json:"stderr"`
	Error    error       `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Metrics  ResourceUsage `json:"metrics"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUTime    time.Duration `json:"cpuTime"`
	MemoryUsed int64         `json:"memoryUsed"`
	DiskUsed   int64         `json:"diskUsed"`
	Processes  int           `json:"processes"`
}

// NewSandbox 创建新的沙箱实例
func NewSandbox(id string, rootDir string, limits ResourceLimits, logger *zap.Logger) (*Sandbox, error) {
	// 创建沙箱根目录
	sandboxDir := filepath.Join(rootDir, fmt.Sprintf("sandbox_%s", id))
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	// 创建工作目录
	workDir := filepath.Join(sandboxDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// 创建临时目录
	tmpDir := filepath.Join(sandboxDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tmp directory: %w", err)
	}

	// 设置目录权限
	if err := os.Chmod(sandboxDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to set sandbox directory permissions: %w", err)
	}

	return &Sandbox{
		id:          id,
		rootDir:     sandboxDir,
		workDir:     workDir,
		network:     limits.Network,
		resources:   limits,
		allowedPaths: []string{workDir, tmpDir},
		logger:      logger,
	}, nil
}

// Execute 在沙箱中执行命令
func (s *Sandbox) Execute(ctx context.Context, cmd string, args []string, env map[string]string) (*SandboxResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	startTime := time.Now()
	
	// 创建命令
	command := exec.CommandContext(ctx, cmd, args...)
	command.Dir = s.workDir

	// 设置环境变量
	command.Env = os.Environ()
	for k, v := range env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 添加沙箱特定环境变量
	command.Env = append(command.Env, fmt.Sprintf("SANDBOX_ID=%s", s.id))
	command.Env = append(command.Env, fmt.Sprintf("SANDBOX_ROOT=%s", s.rootDir))

	// 设置资源限制
	if err := s.setResourceLimits(command); err != nil {
		return nil, fmt.Errorf("failed to set resource limits: %w", err)
	}

	// 设置网络命名空间（如果禁用网络）
	if !s.network {
		if err := s.setupNetworkNamespace(command); err != nil {
			return nil, fmt.Errorf("failed to setup network namespace: %w", err)
		}
	}

	// 捕获输出
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	// 执行命令
	err := command.Run()
	duration := time.Since(startTime)

	// 收集资源使用情况
	metrics, _ := s.collectResourceUsage()

	result := &SandboxResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
		Metrics:  metrics,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.Error = err
			result.ExitCode = -1
		}
	}

	s.logger.Info("Command executed in sandbox",
		zap.String("sandbox_id", s.id),
		zap.String("command", cmd),
		zap.Strings("args", args),
		zap.Int("exit_code", result.ExitCode),
		zap.Duration("duration", duration))

	return result, nil
}

// CopyFile 将文件复制到沙箱中
func (s *Sandbox) CopyFile(src, dst string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保目标路径在允许的路径内
	dstPath := filepath.Join(s.workDir, dst)
	if !strings.HasPrefix(dstPath, s.workDir) {
		return fmt.Errorf("destination path is outside sandbox work directory")
	}

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 复制文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// ReadFile 从沙箱中读取文件
func (s *Sandbox) ReadFile(path string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保路径在允许的路径内
	filePath := filepath.Join(s.workDir, path)
	if !strings.HasPrefix(filePath, s.workDir) {
		return nil, fmt.Errorf("file path is outside sandbox work directory")
	}

	return os.ReadFile(filePath)
}

// WriteFile 将文件写入沙箱
func (s *Sandbox) WriteFile(path string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保路径在允许的路径内
	filePath := filepath.Join(s.workDir, path)
	if !strings.HasPrefix(filePath, s.workDir) {
		return fmt.Errorf("file path is outside sandbox work directory")
	}

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// Cleanup 清理沙箱资源
func (s *Sandbox) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.RemoveAll(s.rootDir)
}

// setResourceLimits 设置资源限制
func (s *Sandbox) setResourceLimits(cmd *exec.Cmd) error {
	// 设置进程属性
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS,
	}

	// 设置资源限制
	if s.resources.MemoryLimit > 0 {
		// 这里需要根据操作系统设置内存限制
		// 在Linux上可以使用setrlimit
	}

	if s.resources.ProcessLimit > 0 {
		// 设置进程数限制
	}

	return nil
}

// setupNetworkNamespace 设置网络命名空间
func (s *Sandbox) setupNetworkNamespace(cmd *exec.Cmd) error {
	// 在Linux上，可以创建新的网络命名空间来隔离网络
	// 这里简化处理，实际实现需要更复杂的网络配置
	return nil
}

// collectResourceUsage 收集资源使用情况
func (s *Sandbox) collectResourceUsage) (ResourceUsage, error) {
	// 实现资源使用情况收集
	// 这里需要根据操作系统实现具体的资源监控
	return ResourceUsage{}, nil
}

// PluginSandboxManager 插件沙箱管理器
type PluginSandboxManager struct {
	sandboxes map[string]*Sandbox
	mu        sync.RWMutex
	rootDir   string
	logger    *zap.Logger
}

// NewPluginSandboxManager 创建插件沙箱管理器
func NewPluginSandboxManager(rootDir string, logger *zap.Logger) *PluginSandboxManager {
	return &PluginSandboxManager{
		sandboxes: make(map[string]*Sandbox),
		rootDir:   rootDir,
		logger:    logger,
	}
}

// CreateSandbox 为插件创建沙箱
func (m *PluginSandboxManager) CreateSandbox(pluginID string, limits ResourceLimits) (*Sandbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if _, exists := m.sandboxes[pluginID]; exists {
		return nil, fmt.Errorf("sandbox for plugin %s already exists", pluginID)
	}

	// 创建新沙箱
	sandbox, err := NewSandbox(pluginID, m.rootDir, limits, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	m.sandboxes[pluginID] = sandbox
	return sandbox, nil
}

// GetSandbox 获取插件沙箱
func (m *PluginSandboxManager) GetSandbox(pluginID string) (*Sandbox, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sandbox, exists := m.sandboxes[pluginID]
	if !exists {
		return nil, fmt.Errorf("sandbox for plugin %s not found", pluginID)
	}

	return sandbox, nil
}

// DestroySandbox 销毁插件沙箱
func (m *PluginSandboxManager) DestroySandbox(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sandbox, exists := m.sandboxes[pluginID]
	if !exists {
		return fmt.Errorf("sandbox for plugin %s not found", pluginID)
	}

	// 清理沙箱
	if err := sandbox.Cleanup(); err != nil {
		m.logger.Error("Failed to cleanup sandbox",
			zap.String("plugin_id", pluginID),
			zap.Error(err))
	}

	delete(m.sandboxes, pluginID)
	return nil
}

// DefaultResourceLimits 返回默认资源限制
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		CPUQuota:     50,          // 50% CPU
		MemoryLimit:  512 * 1024 * 1024, // 512MB
		DiskLimit:    100 * 1024 * 1024, // 100MB
		ProcessLimit: 10,          // 10个进程
		Network:      false,       // 默认禁用网络
		Timeout:      30 * time.Second, // 30秒超时
	}
}

// HighResourceLimits 返回高资源限制
func HighResourceLimits() ResourceLimits {
	return ResourceLimits{
		CPUQuota:     80,          // 80% CPU
		MemoryLimit:  1024 * 1024 * 1024, // 1GB
		DiskLimit:    500 * 1024 * 1024, // 500MB
		ProcessLimit: 20,          // 20个进程
		Network:      true,        // 允许网络
		Timeout:      60 * time.Second, // 60秒超时
	}
}