package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SimpleService 定义简单服务结构
type SimpleService struct {
	Name        string
	Path        string
	Description string
	Port        string
	Command     []string
}

// 获取所有简单服务列表
func getAllSimpleServices() []SimpleService {
	return []SimpleService{
		{
			Name:        "auth",
			Path:        "services/auth",
			Description: "认证服务",
			Port:        "8081",
			Command:     []string{"go", "run", "main.go"},
		},
		{
			Name:        "api",
			Path:        "services/api",
			Description: "API服务",
			Port:        "8082",
			Command:     []string{"go", "run", "main.go"},
		},
		{
			Name:        "gateway",
			Path:        "services/gateway",
			Description: "网关服务",
			Port:        "8080",
			Command:     []string{"go", "run", "main.go"},
		},
		{
			Name:        "frontend",
			Path:        "services/frontend",
			Description: "前端服务",
			Port:        "3000",
			Command:     []string{"npm", "run", "dev"},
		},
	}
}

// killProcessByPort 根据端口杀死进程
func killProcessByPort(port string) error {
	// 使用lsof命令查找占用端口的进程
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%s", port))
	output, err := cmd.Output()
	if err != nil {
		// 如果没有找到进程，这不是错误
		return nil
	}
	
	// 解析进程ID
	pids := strings.Fields(string(output))
	if len(pids) == 0 {
		return nil
	}
	
	fmt.Printf("发现端口 %s 被以下进程占用: %v\n", port, pids)
	
	// 杀死所有占用该端口的进程
	for _, pidStr := range pids {
		pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
		if err != nil {
			continue
		}
		
		killCmd := exec.Command("kill", "-TERM", strconv.Itoa(pid))
		if err := killCmd.Run(); err != nil {
			// 如果TERM信号失败，尝试KILL信号
			killCmd = exec.Command("kill", "-KILL", strconv.Itoa(pid))
			if err := killCmd.Run(); err != nil {
				fmt.Printf("警告: 无法杀死进程 %d: %v\n", pid, err)
			}
		} else {
			fmt.Printf("已终止进程 %d\n", pid)
		}
	}
	
	// 给进程一些时间退出
	time.Sleep(1 * time.Second)
	return nil
}

// 启动单个简单服务
func startSimpleService(service SimpleService) (*exec.Cmd, error) {
	// 检查服务目录是否存在
	if _, err := os.Stat(service.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("服务目录不存在: %s", service.Path)
	}
	
	// 对于Go服务，检查main.go是否存在
	if service.Command[0] == "go" {
		mainPath := filepath.Join(service.Path, "main.go")
		if _, err := os.Stat(mainPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("服务文件不存在: %s", mainPath)
		}
	}
	
	cmd := exec.Command(service.Command[0], service.Command[1:]...)
	cmd.Dir = service.Path
	
	// 创建管道用于捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建标准输出管道失败: %v", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建标准错误管道失败: %v", err)
	}
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动服务 %s 失败: %v", service.Name, err)
	}
	
	// 启动goroutine读取输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("[%s] %s\n", service.Name, scanner.Text())
		}
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[%s ERROR] %s\n", service.Name, scanner.Text())
		}
	}()
	
	return cmd, nil
}

func main() {
	fmt.Println("=== 全模块启动器 (简单版) ===")
	
	// 获取所有简单服务
	services := getAllSimpleServices()
	if len(services) == 0 {
		fmt.Println("错误: 没有找到任何可用的服务配置")
		os.Exit(1)
	}
	
	// 从命令行参数获取要启动的服务列表
	args := os.Args[1:]
	var selectedServices []SimpleService
	var unknownServices []string
	
	if len(args) == 0 {
		// 没有指定服务，启动所有服务
		selectedServices = services
		fmt.Println("未指定服务，将启动所有服务")
	} else {
		// 根据参数选择服务
		serviceMap := make(map[string]SimpleService)
		for _, service := range services {
			serviceMap[service.Name] = service
		}
		
		for _, arg := range args {
			if service, exists := serviceMap[arg]; exists {
				selectedServices = append(selectedServices, service)
			} else {
				unknownServices = append(unknownServices, arg)
			}
		}
		
		if len(unknownServices) > 0 {
			fmt.Printf("警告: 发现未知服务: %v\n", unknownServices)
		}
	}
	
	if len(selectedServices) == 0 {
		fmt.Println("错误: 没有选择任何有效的服务")
		os.Exit(1)
	}
	
	fmt.Printf("\n将启动以下服务:\n")
	for _, service := range selectedServices {
		fmt.Printf("- %s (%s) - 端口: %s\n", service.Name, service.Description, service.Port)
	}
	
	// 在启动服务前，先检查并终止占用端口的进程
	fmt.Println("\n检查端口占用情况...")
	for _, service := range selectedServices {
		if err := killProcessByPort(service.Port); err != nil {
			fmt.Printf("警告: 检查端口 %s 时出错: %v\n", service.Port, err)
		}
	}
	
	// 创建通道用于接收中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 启动所有选中的服务
	var runningServices []*exec.Cmd
	var failedServices []string
	
	for _, service := range selectedServices {
		fmt.Printf("正在启动 %s 服务...\n", service.Name)
		cmd, err := startSimpleService(service)
		if err != nil {
			fmt.Printf("错误: 启动服务 %s 失败: %v\n", service.Name, err)
			failedServices = append(failedServices, service.Name)
			continue
		}
		runningServices = append(runningServices, cmd)
		fmt.Printf("服务 %s 已启动 (PID: %d)\n", service.Name, cmd.Process.Pid)
	}
	
	if len(runningServices) == 0 {
		fmt.Printf("错误: 没有成功启动任何服务。失败的服务: %v\n", failedServices)
		os.Exit(1)
	}
	
	if len(failedServices) > 0 {
		fmt.Printf("警告: 以下服务启动失败: %v\n", failedServices)
	}
	
	fmt.Println("\n所有服务已启动，按 Ctrl+C 停止所有服务")
	
	// 等待中断信号
	<-sigChan
	
	fmt.Println("\n正在停止所有服务...")
	
	// 停止所有服务
	for _, cmd := range runningServices {
		if cmd.Process != nil {
			fmt.Printf("正在停止服务 (PID: %d)...\n", cmd.Process.Pid)
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				fmt.Printf("警告: 发送停止信号失败: %v\n", err)
				continue
			}
			
			// 等待一段时间，如果服务没有停止，则强制杀死
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			
			select {
			case <-time.After(5 * time.Second):
				fmt.Println("服务未在5秒内停止，强制终止")
				if err := cmd.Process.Kill(); err != nil {
					fmt.Printf("错误: 强制终止服务失败: %v\n", err)
				}
			case err := <-done:
				if err != nil {
					fmt.Printf("警告: 服务停止时出错: %v\n", err)
				} else {
					fmt.Printf("服务已成功停止\n")
				}
			}
		}
	}
	
	fmt.Println("所有服务已停止")
}