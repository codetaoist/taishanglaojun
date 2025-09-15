package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client API客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient 创建新的API客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken 设置认证令牌
func (c *Client) SetToken(token string) {
	c.token = token
}

// Request 发送HTTP请求
func (c *Client) Request(method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path
	
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ct-cli/1.0.0")
	
	// 添加认证头
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	
	return resp, nil
}

// Get 发送GET请求
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Request("GET", path, nil)
}

// Post 发送POST请求
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.Request("POST", path, body)
}

// Put 发送PUT请求
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	return c.Request("PUT", path, body)
}

// Delete 发送DELETE请求
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Request("DELETE", path, nil)
}