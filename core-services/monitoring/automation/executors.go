package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// BaseExecutor 基础执行?
type BaseExecutor struct {
	config ExecutorConfig
	stats  *ExecutorStats
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBaseExecutor 创建基础执行?
func NewBaseExecutor(config ExecutorConfig) *BaseExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseExecutor{
		config: config,
		stats: &ExecutorStats{
			ExecutedTasks:   0,
			SuccessfulTasks: 0,
			FailedTasks:     0,
			AverageExecTime: 0,
			LastExecution:   time.Time{},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动执行?
func (e *BaseExecutor) Start() error {
	return nil
}

// Stop 停止执行?
func (e *BaseExecutor) Stop() error {
	e.cancel()
	return nil
}

// GetStats 获取统计信息
func (e *BaseExecutor) GetStats() *ExecutorStats {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	stats := *e.stats
	return &stats
}

// HealthCheck 健康检?
func (e *BaseExecutor) HealthCheck() error {
	return nil
}

// updateStats 更新统计信息
func (e *BaseExecutor) updateStats(success bool, duration time.Duration) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	e.stats.ExecutedTasks++
	if success {
		e.stats.SuccessfulTasks++
	} else {
		e.stats.FailedTasks++
	}
	
	// 计算平均执行时间
	if e.stats.ExecutedTasks == 1 {
		e.stats.AverageExecTime = duration
	} else {
		e.stats.AverageExecTime = time.Duration(
			(int64(e.stats.AverageExecTime)*(e.stats.ExecutedTasks-1) + int64(duration)) / e.stats.ExecutedTasks,
		)
	}
	
	e.stats.LastExecution = time.Now()
}

// ShellExecutor Shell执行?
type ShellExecutor struct {
	*BaseExecutor
}

// NewShellExecutor 创建Shell执行?
func NewShellExecutor(config ExecutorConfig) *ShellExecutor {
	return &ShellExecutor{
		BaseExecutor: NewBaseExecutor(config),
	}
}

// ExecuteTask 执行任务
func (e *ShellExecutor) ExecuteTask(task *Task, inputs map[string]interface{}) error {
	startTime := time.Now()
	
	// 构建命令
	command := task.Definition.Command
	if command == "" {
		command = task.Definition.Script
	}
	
	// 替换变量
	for key, value := range inputs {
		placeholder := fmt.Sprintf("${%s}", key)
		command = strings.ReplaceAll(command, placeholder, fmt.Sprintf("%v", value))
	}
	
	// 创建命令上下?
	ctx := e.ctx
	if task.Definition.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, task.Definition.Timeout)
		defer cancel()
	}
	
	// 执行命令
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	
	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range task.Definition.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// 设置工作目录
	if task.Definition.WorkingDir != "" {
		cmd.Dir = task.Definition.WorkingDir
	}
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	
	// 记录日志
	task.mutex.Lock()
	task.Logs = append(task.Logs, string(output))
	task.mutex.Unlock()
	
	// 更新统计信息
	duration := time.Since(startTime)
	e.updateStats(err == nil, duration)
	
	return err
}

// DockerExecutor Docker执行?
type DockerExecutor struct {
	*BaseExecutor
}

// NewDockerExecutor 创建Docker执行?
func NewDockerExecutor(config ExecutorConfig) *DockerExecutor {
	return &DockerExecutor{
		BaseExecutor: NewBaseExecutor(config),
	}
}

// ExecuteTask 执行任务
func (e *DockerExecutor) ExecuteTask(task *Task, inputs map[string]interface{}) error {
	startTime := time.Now()
	
	// 构建Docker命令
	dockerCmd := []string{"docker", "run", "--rm"}
	
	// 添加环境变量
	for key, value := range task.Definition.Environment {
		dockerCmd = append(dockerCmd, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	
	// 添加工作目录
	if task.Definition.WorkingDir != "" {
		dockerCmd = append(dockerCmd, "-w", task.Definition.WorkingDir)
	}
	
	// 添加镜像
	if task.Definition.Image != "" {
		dockerCmd = append(dockerCmd, task.Definition.Image)
	}
	
	// 添加命令
	if task.Definition.Command != "" {
		dockerCmd = append(dockerCmd, "sh", "-c", task.Definition.Command)
	}
	
	// 创建命令上下?
	ctx := e.ctx
	if task.Definition.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, task.Definition.Timeout)
		defer cancel()
	}
	
	// 执行Docker命令
	cmd := exec.CommandContext(ctx, dockerCmd[0], dockerCmd[1:]...)
	output, err := cmd.CombinedOutput()
	
	// 记录日志
	task.mutex.Lock()
	task.Logs = append(task.Logs, string(output))
	task.mutex.Unlock()
	
	// 更新统计信息
	duration := time.Since(startTime)
	e.updateStats(err == nil, duration)
	
	return err
}

// KubernetesExecutor Kubernetes执行?
type KubernetesExecutor struct {
	*BaseExecutor
}

// NewKubernetesExecutor 创建Kubernetes执行?
func NewKubernetesExecutor(config ExecutorConfig) *KubernetesExecutor {
	return &KubernetesExecutor{
		BaseExecutor: NewBaseExecutor(config),
	}
}

// ExecuteTask 执行任务
func (e *KubernetesExecutor) ExecuteTask(task *Task, inputs map[string]interface{}) error {
	startTime := time.Now()
	
	// 构建Job定义
	jobDef := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("task-%s", task.ID),
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"restartPolicy": "Never",
					"containers": []map[string]interface{}{
						{
							"name":    "task",
							"image":   task.Definition.Image,
							"command": []string{"sh", "-c", task.Definition.Command},
							"env":     buildEnvVars(task.Definition.Environment),
						},
					},
				},
			},
		},
	}
	
	// 序列化Job定义
	jobJSON, err := json.Marshal(jobDef)
	if err != nil {
		e.updateStats(false, time.Since(startTime))
		return fmt.Errorf("failed to marshal job definition: %w", err)
	}
	
	// 创建Job
	createCmd := exec.Command("kubectl", "apply", "-f", "-")
	createCmd.Stdin = strings.NewReader(string(jobJSON))
	
	output, err := createCmd.CombinedOutput()
	if err != nil {
		e.updateStats(false, time.Since(startTime))
		return fmt.Errorf("failed to create job: %w", err)
	}
	
	// 等待Job完成
	jobName := fmt.Sprintf("task-%s", task.ID)
	waitCmd := exec.Command("kubectl", "wait", "--for=condition=complete", "--timeout=300s", "job/"+jobName)
	
	ctx := e.ctx
	if task.Definition.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, task.Definition.Timeout)
		defer cancel()
	}
	
	waitCmd = exec.CommandContext(ctx, waitCmd.Path, waitCmd.Args[1:]...)
	waitOutput, waitErr := waitCmd.CombinedOutput()
	
	// 获取Job日志
	logsCmd := exec.Command("kubectl", "logs", "job/"+jobName)
	logsOutput, _ := logsCmd.CombinedOutput()
	
	// 清理Job
	deleteCmd := exec.Command("kubectl", "delete", "job", jobName)
	deleteCmd.Run()
	
	// 记录日志
	task.mutex.Lock()
	task.Logs = append(task.Logs, string(output))
	task.Logs = append(task.Logs, string(waitOutput))
	task.Logs = append(task.Logs, string(logsOutput))
	task.mutex.Unlock()
	
	// 更新统计信息
	duration := time.Since(startTime)
	e.updateStats(waitErr == nil, duration)
	
	return waitErr
}

// HTTPExecutor HTTP执行?
type HTTPExecutor struct {
	*BaseExecutor
}

// NewHTTPExecutor 创建HTTP执行?
func NewHTTPExecutor(config ExecutorConfig) *HTTPExecutor {
	return &HTTPExecutor{
		BaseExecutor: NewBaseExecutor(config),
	}
}

// ExecuteTask 执行任务
func (e *HTTPExecutor) ExecuteTask(task *Task, inputs map[string]interface{}) error {
	startTime := time.Now()
	
	// 构建curl命令
	curlCmd := []string{"curl", "-s"}
	
	// 添加方法
	method := "GET"
	if m, ok := task.Definition.Environment["METHOD"]; ok {
		method = m
	}
	curlCmd = append(curlCmd, "-X", method)
	
	// 添加头部
	for key, value := range task.Definition.Environment {
		if strings.HasPrefix(key, "HEADER_") {
			headerName := strings.TrimPrefix(key, "HEADER_")
			curlCmd = append(curlCmd, "-H", fmt.Sprintf("%s: %s", headerName, value))
		}
	}
	
	// 添加数据
	if data, ok := task.Definition.Environment["DATA"]; ok {
		curlCmd = append(curlCmd, "-d", data)
	}
	
	// 添加URL
	url := task.Definition.Command
	if url == "" {
		url = task.Definition.Environment["URL"]
	}
	curlCmd = append(curlCmd, url)
	
	// 创建命令上下?
	ctx := e.ctx
	if task.Definition.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, task.Definition.Timeout)
		defer cancel()
	}
	
	// 执行curl命令
	cmd := exec.CommandContext(ctx, curlCmd[0], curlCmd[1:]...)
	output, err := cmd.CombinedOutput()
	
	// 记录日志
	task.mutex.Lock()
	task.Logs = append(task.Logs, string(output))
	task.mutex.Unlock()
	
	// 更新统计信息
	duration := time.Since(startTime)
	e.updateStats(err == nil, duration)
	
	return err
}

// ScriptExecutor 脚本执行?
type ScriptExecutor struct {
	*BaseExecutor
}

// NewScriptExecutor 创建脚本执行?
func NewScriptExecutor(config ExecutorConfig) *ScriptExecutor {
	return &ScriptExecutor{
		BaseExecutor: NewBaseExecutor(config),
	}
}

// ExecuteTask 执行任务
func (e *ScriptExecutor) ExecuteTask(task *Task, inputs map[string]interface{}) error {
	startTime := time.Now()
	
	// 创建临时脚本文件
	script := task.Definition.Script
	if script == "" {
		script = task.Definition.Command
	}
	
	// 替换变量
	for key, value := range inputs {
		placeholder := fmt.Sprintf("${%s}", key)
		script = strings.ReplaceAll(script, placeholder, fmt.Sprintf("%v", value))
	}
	
	// 写入临时文件
	tmpFile, err := os.CreateTemp("", "script-*.sh")
	if err != nil {
		e.updateStats(false, time.Since(startTime))
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(script); err != nil {
		e.updateStats(false, time.Since(startTime))
		return fmt.Errorf("failed to write script: %w", err)
	}
	tmpFile.Close()
	
	// 设置执行权限
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		e.updateStats(false, time.Since(startTime))
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	
	// 创建命令上下?
	ctx := e.ctx
	if task.Definition.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, task.Definition.Timeout)
		defer cancel()
	}
	
	// 执行脚本
	cmd := exec.CommandContext(ctx, tmpFile.Name())
	
	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range task.Definition.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// 设置工作目录
	if task.Definition.WorkingDir != "" {
		cmd.Dir = task.Definition.WorkingDir
	}
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	
	// 记录日志
	task.mutex.Lock()
	task.Logs = append(task.Logs, string(output))
	task.mutex.Unlock()
	
	// 更新统计信息
	duration := time.Since(startTime)
	e.updateStats(err == nil, duration)
	
	return err
}

// CreateExecutor 创建执行?
func CreateExecutor(config ExecutorConfig) (Executor, error) {
	switch config.Type {
	case "shell":
		return NewShellExecutor(config), nil
	case "docker":
		return NewDockerExecutor(config), nil
	case "kubernetes":
		return NewKubernetesExecutor(config), nil
	case "http":
		return NewHTTPExecutor(config), nil
	case "script":
		return NewScriptExecutor(config), nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", config.Type)
	}
}

// buildEnvVars 构建环境变量
func buildEnvVars(env map[string]string) []map[string]interface{} {
	var envVars []map[string]interface{}
	for key, value := range env {
		envVars = append(envVars, map[string]interface{}{
			"name":  key,
			"value": value,
		})
	}
	return envVars
}

