package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	// 创建测试管理员用户
	fmt.Println("创建测试管理员用户...")
	
	resp, err := http.Post("http://localhost:8080/api/v1/auth/test-user?username=admin&role=admin&password=admin123", 
		"application/json", 
		strings.NewReader("{}"))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}