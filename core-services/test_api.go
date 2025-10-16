package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 测试角色列表API
	fmt.Println("测试角色列表API...")
	
	// 创建请求
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/roles", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	
	// 添加Authorization头
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmI2NzNmOTAtZTc1My00NGU4LWE1ZWItNzUxNmQ0YzY4M2FkIiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJBRE1JTiIsImxldmVsIjo1LCJpc3MiOiJ0YWlzaGFuZy1sYW9qdW4iLCJleHAiOjE3NjA2Nzc4NDAsIm5iZiI6MTc2MDU5MTQ0MCwiaWF0IjoxNzYwNTkxNDQwfQ.zyAlXixegNqdtyJ0kF279CBRgvkXfLbSBwMMgiWLjMg"
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
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
	fmt.Printf("响应内容: %s\n", string(body))

	// 尝试格式化JSON
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			prettyJSON, _ := json.MarshalIndent(result, "", "  ")
			fmt.Printf("格式化JSON:\n%s\n", string(prettyJSON))
		}
	}
}