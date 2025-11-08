package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// 测试模型服务API
func main() {
	// 检查命令行参数
	if len(os.Args) > 1 && os.Args[1] == "ollama" {
		testOllamaMain()
		return
	}
	
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	
	// API基础URL
	baseURL := "http://localhost:8080/api/v1/taishang/models"
	
	// 测试创建模型配置
	fmt.Println("=== 测试创建模型配置 ===")
	testCreateModel(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试获取模型列表
	fmt.Println("\n=== 测试获取模型列表 ===")
	testListModels(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试文本生成
	fmt.Println("\n=== 测试文本生成 ===")
	testTextGeneration(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试流式文本生成
	fmt.Println("\n=== 测试流式文本生成 ===")
	testTextGenerationStream(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试嵌入生成
	fmt.Println("\n=== 测试嵌入生成 ===")
	testEmbeddingGeneration(baseURL)
	
	// 等待一秒
	time.Sleep(1 * time.Second)
	
	// 测试批量嵌入生成
	fmt.Println("\n=== 测试批量嵌入生成 ===")
	testBatchEmbeddingGeneration(baseURL)
}

// 测试创建模型配置
func testCreateModel(baseURL string) {
	// 创建模型配置请求
	modelConfig := map[string]interface{}{
		"name":        "test-gpt-3.5-turbo",
		"provider":    "openai",
		"model":       "gpt-3.5-turbo",
		"apiKey":      "sk-test-key",
		"baseURL":     "https://api.openai.com/v1",
		"enabled":     true,
		"maxTokens":   2048,
		"temperature": 0.7,
		"description": "测试用OpenAI GPT-3.5 Turbo模型",
		"timeout":     30,
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
func testListModels(baseURL string) {
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

// 测试文本生成
func testTextGeneration(baseURL string) {
	// 创建文本生成请求
	textGenReq := map[string]interface{}{
		"model": "test-gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个有用的助手。"},
			{"role": "user", "content": "请介绍一下人工智能的发展历史。"},
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

// 测试流式文本生成
func testTextGenerationStream(baseURL string) {
	// 创建文本生成请求
	textGenReq := map[string]interface{}{
		"model": "test-gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个有用的助手。"},
			{"role": "user", "content": "请简单介绍一下机器学习。"},
		},
		"maxTokens":   300,
		"temperature": 0.7,
		"stream":      true,
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
	
	// 打印结果
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Println("流式响应:")
	
	// 读取流式响应
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("读取流式响应错误: %v\n", err)
			}
			break
		}
		
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
	}
	fmt.Println()
}

// 测试嵌入生成
func testEmbeddingGeneration(baseURL string) {
	// 创建嵌入生成请求
	embeddingReq := map[string]interface{}{
		"model": "test-gpt-3.5-turbo",
		"text":  "这是一个测试文本，用于生成嵌入向量。",
	}
	
	// 转换为JSON
	jsonData, err := json.Marshal(embeddingReq)
	if err != nil {
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}
	
	// 发送HTTP请求
	resp, err := http.Post(baseURL+"/embeddings", "application/json", bytes.NewBuffer(jsonData))
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

// 测试批量嵌入生成
func testBatchEmbeddingGeneration(baseURL string) {
	// 创建批量嵌入生成请求
	batchEmbeddingReq := map[string]interface{}{
		"model": "test-gpt-3.5-turbo",
		"texts": []string{
			"这是第一个测试文本。",
			"这是第二个测试文本。",
			"这是第三个测试文本。",
		},
	}
	
	// 转换为JSON
	jsonData, err := json.Marshal(batchEmbeddingReq)
	if err != nil {
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}
	
	// 发送HTTP请求
	resp, err := http.Post(baseURL+"/embeddings/batch", "application/json", bytes.NewBuffer(jsonData))
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