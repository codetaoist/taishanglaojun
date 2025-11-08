package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 测试Ollama模型服务API
func testOllamaMain() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	
	// API基础URL
	baseURL := "http://localhost:8082/api/v1/taishang/models"
	
	// 测试创建Ollama模型配置
	fmt.Println("=== 测试创建Ollama模型配置 ===")
	testCreateOllamaModel(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试获取模型列表
	fmt.Println("\n=== 测试获取模型列表 ===")
	testListModelsOllama(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试获取Ollama服务中的模型
	fmt.Println("\n=== 测试获取Ollama服务中的模型 ===")
	testListOllamaServiceModels(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试文本生成
	fmt.Println("\n=== 测试文本生成 ===")
	testTextGenerationOllama(baseURL)
}

// 测试创建Ollama模型配置
func testCreateOllamaModel(baseURL string) {
	// 创建Ollama模型配置请求
	modelConfig := map[string]interface{}{
		"name":        "test-ollama-llama2",
		"provider":    "ollama",
		"model":       "llama2:7b",
		"baseURL":     "http://localhost:11434",
		"enabled":     true,
		"maxTokens":   2048,
		"temperature": 0.7,
		"description": "测试用Ollama Llama2 7B模型",
		"timeout":     60,
	}
	
	// 转换为JSON
	jsonData, err := json.Marshal(modelConfig)
	if err != nil {
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}
	
	// 发送HTTP请求
	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("HTTP请求错误: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		return
	}
	
	// 打印结果
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试获取模型列表
func testListModelsOllama(baseURL string) {
	// 发送HTTP请求
	resp, err := http.Get(baseURL)
	if err != nil {
		fmt.Printf("HTTP请求错误: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		return
	}
	
	// 打印结果
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试获取Ollama服务中的模型
func testListOllamaServiceModels(baseURL string) {
	// 发送HTTP请求
	resp, err := http.Get(baseURL + "/service/Ollama/models")
	if err != nil {
		fmt.Printf("HTTP请求错误: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		return
	}
	
	// 打印结果
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试文本生成
func testTextGenerationOllama(baseURL string) {
	// 创建文本生成请求
	textGenReq := map[string]interface{}{
		"model": "test-ollama-llama2",
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个有用的助手。"},
			{"role": "user", "content": "请简单介绍一下人工智能。"},
		},
		"maxTokens":   500,
		"temperature": 0.7,
		"stream":      false,
	}
	
	// 转换为JSON
	jsonData, err := json.Marshal(textGenReq)
	if err != nil {
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}
	
	// 发送HTTP请求
	resp, err := http.Post(baseURL+"/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("HTTP请求错误: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		return
	}
	
	// 打印结果
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
}