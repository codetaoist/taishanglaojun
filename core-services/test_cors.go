package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	// 测试CORS和API连接
	fmt.Println("测试CORS和API连接...")
	
	// 创建OPTIONS请求测试CORS
	req, err := http.NewRequest("OPTIONS", "http://localhost:8080/api/v1/roles", nil)
	if err != nil {
		fmt.Printf("创建OPTIONS请求失败: %v\n", err)
		return
	}
	
	// 添加CORS相关头
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization,Content-Type")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("OPTIONS请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("OPTIONS响应状态码: %d\n", resp.StatusCode)
	fmt.Println("CORS响应头:")
	for key, values := range resp.Header {
		if strings.Contains(strings.ToLower(key), "access-control") {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}
	}
	
	// 测试实际的GET请求
	fmt.Println("\n测试GET请求...")
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmI2NzNmOTAtZTc1My00NGU4LWE1ZWItNzUxNmQ0YzY4M2FkIiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJBRE1JTiIsImxldmVsIjo1LCJpc3MiOiJ0YWlzaGFuZy1sYW9qdW4iLCJleHAiOjE3NjA2Nzc4NDAsIm5iZiI6MTc2MDU5MTQ0MCwiaWF0IjoxNzYwNTkxNDQwfQ.zyAlXixegNqdtyJ0kF279CBRgvkXfLbSBwMMgiWLjMg"
	
	getReq, err := http.NewRequest("GET", "http://localhost:8080/api/v1/roles", nil)
	if err != nil {
		fmt.Printf("创建GET请求失败: %v\n", err)
		return
	}
	
	getReq.Header.Set("Authorization", "Bearer "+token)
	getReq.Header.Set("Origin", "http://localhost:5173")
	
	getResp, err := client.Do(getReq)
	if err != nil {
		fmt.Printf("GET请求失败: %v\n", err)
		return
	}
	defer getResp.Body.Close()
	
	fmt.Printf("GET响应状态码: %d\n", getResp.StatusCode)
	fmt.Println("GET响应头:")
	for key, values := range getResp.Header {
		if strings.Contains(strings.ToLower(key), "access-control") || strings.Contains(strings.ToLower(key), "content-type") {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}
	}
	
	// 读取响应内容
	body, err := io.ReadAll(getResp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}
	
	fmt.Printf("响应内容长度: %d 字节\n", len(body))
	if len(body) > 0 && len(body) < 1000 {
		fmt.Printf("响应内容: %s\n", string(body))
	} else if len(body) > 0 {
		fmt.Printf("响应内容（前500字符）: %s...\n", string(body[:500]))
	}
}