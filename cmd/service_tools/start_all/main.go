package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// InteractiveService 定义交互式服务结构
type InteractiveService struct {
	Name        string
	Path        string
	Description string
	Port        string
}

// 获取所有交互式服务列表
func getAllInteractiveServices() []InteractiveService {
	return []InteractiveService{
		{
			Name:        "auth",
			Path:        "services/auth",
			Description: "认证服务",
			Port:        "8081",
		},
		{
			Name:        "api",
			Path:        "services/api",
			Description: "API服务",
			Port:        "8082",
		},
		{
			Name:        "gateway",
			Path:        "services/gateway",
			Description: "网关服务",
			Port:        "8080",
		},
		{
			Name:        "frontend",
			Path:        "services/frontend",
			Description: "前端服务",
			Port:        "3000",
		},
	}
}

// 启动单个交互式服务
func startInteractiveService(service InteractiveService) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	
	// 前端服务使用npm启动，其他服务使用go run
	if service.Name == "frontend" {
		cmd = exec.Command("npm", "run", "dev")
	} else {
		// 检查是否有main.go文件
		mainPath := fmt.Sprintf("%s/main.go", service.Path)
		if _, err := os.Stat(mainPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("服务 %s 没有找到main.go文件", service.Name)
		}
		cmd = exec.Command("go", "run", "main.go")
	}
	
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

// 显示交互式服务列表
func showInteractiveServices(services []InteractiveService) {
	fmt.Println("可用服务列表:")
	for i, service := range services {
		fmt.Printf("%d. %s - %s (端口: %s)\n", i+1, service.Name, service.Description, service.Port)
	}
}

// 获取用户选择的交互式服务
func getSelectedInteractiveServices(services []InteractiveService) []InteractiveService {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Println("\n请选择要启动的服务 (输入数字，多个用逗号分隔，all表示全部，q退出):")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "q" {
			os.Exit(0)
		}
		
		if input == "all" {
			return services
		}
		
		var selected []InteractiveService
		parts := strings.Split(input, ",")
		valid := true
		
		for _, part := range parts {
			var index int
			_, err := fmt.Sscanf(part, "%d", &index)
			if err != nil || index < 1 || index > len(services) {
				fmt.Printf("无效的选择: %s\n", part)
				valid = false
				break
			}
			selected = append(selected, services[index-1])
		}
		
		if valid {
			return selected
		}
	}
}

func main() {
	fmt.Println("=== 全模块启动器 ===")
	
	// 获取所有交互式服务
	services := getAllInteractiveServices()
	
	// 显示交互式服务列表
	showInteractiveServices(services)
	
	// 获取用户选择
	selectedServices := getSelectedInteractiveServices(services)
	
	fmt.Printf("\n将启动以下服务:\n")
	for _, service := range selectedServices {
		fmt.Printf("- %s (%s)\n", service.Name, service.Description)
	}
	
	// 创建通道用于接收中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 启动所有选中的服务
	var runningServices []*exec.Cmd
	for _, service := range selectedServices {
		fmt.Printf("正在启动 %s 服务...\n", service.Name)
		cmd, err := startInteractiveService(service)
		if err != nil {
			fmt.Printf("启动服务 %s 失败: %v\n", service.Name, err)
			continue
		}
		runningServices = append(runningServices, cmd)
		fmt.Printf("服务 %s 已启动 (PID: %d)\n", service.Name, cmd.Process.Pid)
	}
	
	if len(runningServices) == 0 {
		fmt.Println("没有成功启动任何服务")
		return
	}
	
	fmt.Println("\n所有服务已启动，按 Ctrl+C 停止所有服务")
	
	// 等待中断信号
	<-sigChan
	
	fmt.Println("\n正在停止所有服务...")
	
	// 停止所有服务
	for _, cmd := range runningServices {
		if cmd.Process != nil {
			fmt.Printf("正在停止服务 (PID: %d)...\n", cmd.Process.Pid)
			cmd.Process.Signal(syscall.SIGTERM)
			
			// 等待一段时间，如果服务没有停止，则强制杀死
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			
			select {
			case <-time.After(5 * time.Second):
				fmt.Println("服务未在5秒内停止，强制终止")
				cmd.Process.Kill()
			case err := <-done:
				if err != nil {
					fmt.Printf("服务停止时出错: %v\n", err)
				}
			}
		}
	}
	
	fmt.Println("所有服务已停止")
}