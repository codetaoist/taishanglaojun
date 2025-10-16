package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmI2NzNmOTAtZTc1My00NGU4LWE1ZWItNzUxNmQ0YzY4M2FkIiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJBRE1JTiIsImxldmVsIjo1LCJpc3MiOiJ0YWlzaGFuZy1sYW9qdW4iLCJleHAiOjE3NjA2Nzc4NDAsIm5iZiI6MTc2MDU5MTQ0MCwiaWF0IjoxNzYwNTkxNDQwfQ.zyAlXixegNqdtyJ0kF279CBRgvkXfLbSBwMMgiWLjMg"
	
	// 测试不同的查询参数
	testCases := []struct {
		name string
		url  string
	}{
		{"基础查询", "http://localhost:8080/api/v1/roles"},
		{"搜索admin", "http://localhost:8080/api/v1/roles?search=admin"},
		{"搜索manager", "http://localhost:8080/api/v1/roles?search=manager"},
		{"类型筛选-system", "http://localhost:8080/api/v1/roles?type=system"},
		{"类型筛选-custom", "http://localhost:8080/api/v1/roles?type=custom"},
		{"状态筛选-active", "http://localhost:8080/api/v1/roles?is_active=true"},
		{"状态筛选-inactive", "http://localhost:8080/api/v1/roles?is_active=false"},
		{"组合查询", "http://localhost:8080/api/v1/roles?search=admin&type=system"},
	}
	
	for _, tc := range testCases {
		fmt.Printf("\n=== %s ===\n", tc.name)
		
		req, err := http.NewRequest("GET", tc.url, nil)
		if err != nil {
			fmt.Printf("创建请求失败: %v\n", err)
			continue
		}
		
		req.Header.Set("Authorization", "Bearer "+token)
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("读取响应失败: %v\n", err)
			continue
		}
		
		fmt.Printf("状态码: %d\n", resp.StatusCode)
		
		if resp.StatusCode == 200 {
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err == nil {
				if roles, ok := result["roles"].([]interface{}); ok {
					fmt.Printf("角色数量: %d\n", len(roles))
					if len(roles) > 0 {
						fmt.Println("角色列表:")
						for i, role := range roles {
							if roleMap, ok := role.(map[string]interface{}); ok {
								name := roleMap["name"]
								code := roleMap["code"]
								roleType := roleMap["type"]
								fmt.Printf("  %d. %s (%s) - %s\n", i+1, name, code, roleType)
							}
						}
					} else {
						fmt.Println("没有找到匹配的角色")
					}
				}
			}
		} else {
			fmt.Printf("错误响应: %s\n", string(body))
		}
	}
}