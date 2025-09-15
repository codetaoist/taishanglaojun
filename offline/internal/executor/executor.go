package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Executor 命令执行器
type Executor struct {
	workDir string
	dryRun  bool
}

// NewExecutor 创建新的执行器
func NewExecutor(workDir string) *Executor {
	return &Executor{
		workDir: workDir,
		dryRun:  false,
	}
}

// SetDryRun 设置是否为试运行模式
func (e *Executor) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

// ExecuteCommand 执行系统命令
func (e *Executor) ExecuteCommand(command string, args ...string) error {
	if e.dryRun {
		fmt.Printf("[DRY RUN] 执行命令: %s %s\n", command, strings.Join(args, " "))
		return nil
	}
	
	cmd := exec.Command(command, args...)
	cmd.Dir = e.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// CreateFile 创建文件
func (e *Executor) CreateFile(filename, content string) error {
	filePath := filepath.Join(e.workDir, filename)
	
	if e.dryRun {
		fmt.Printf("[DRY RUN] 创建文件: %s\n", filePath)
		return nil
	}
	
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ModifyFile 修改文件
func (e *Executor) ModifyFile(filename string, modifier func(string) string) error {
	filePath := filepath.Join(e.workDir, filename)
	
	if e.dryRun {
		fmt.Printf("[DRY RUN] 修改文件: %s\n", filePath)
		return nil
	}
	
	// 读取原文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	
	// 应用修改
	newContent := modifier(string(content))
	
	// 写回文件
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// DeleteFile 删除文件
func (e *Executor) DeleteFile(filename string) error {
	filePath := filepath.Join(e.workDir, filename)
	
	if e.dryRun {
		fmt.Printf("[DRY RUN] 删除文件: %s\n", filePath)
		return nil
	}
	
	return os.Remove(filePath)
}

// GitCommit 执行Git提交
func (e *Executor) GitCommit(message string) error {
	if err := e.ExecuteCommand("git", "add", "."); err != nil {
		return fmt.Errorf("git add 失败: %w", err)
	}
	
	if err := e.ExecuteCommand("git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit 失败: %w", err)
	}
	
	return nil
}