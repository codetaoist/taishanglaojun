package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/services"
)

// APIKeyController API密钥控制�?type APIKeyController struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyController 创建新的API密钥控制�?func NewAPIKeyController(apiKeyService *services.APIKeyService) *APIKeyController {
	return &APIKeyController{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKeyRequest 创建API密钥请求
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse 创建API密钥响应
type CreateAPIKeyResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Key         string    `json:"key"`
	Prefix      string    `json:"prefix"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   *string   `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIKeyResponse API密钥响应（不包含完整密钥�?type APIKeyResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prefix      string    `json:"prefix"`
	Permissions []string  `json:"permissions"`
	Status      string    `json:"status"`
	LastUsedAt  *string   `json:"last_used_at,omitempty"`
	ExpiresAt   *string   `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAPIKey 创建API密钥
func (c *APIKeyController) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// 从上下文获取用户ID（假设已通过中间件设置）
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 解析过期时间
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "Invalid expires_at format", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	// 创建API密钥
	apiKey, key, err := c.apiKeyService.GenerateAPIKey(userID, req.Name, req.Description, req.Permissions, expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := CreateAPIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Description: apiKey.Description,
		Key:         key,
		Prefix:      apiKey.Prefix,
		Permissions: apiKey.Permissions,
		CreatedAt:   apiKey.CreatedAt,
	}

	if apiKey.ExpiresAt != nil {
		expiresAtStr := apiKey.ExpiresAt.Format(time.RFC3339)
		response.ExpiresAt = &expiresAtStr
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListAPIKeys 获取API密钥列表
func (c *APIKeyController) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 解析分页参数
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// 获取API密钥列表
	apiKeys, total, err := c.apiKeyService.ListAPIKeys(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应
	var responses []APIKeyResponse
	for _, apiKey := range apiKeys {
		response := APIKeyResponse{
			ID:          apiKey.ID,
			Name:        apiKey.Name,
			Description: apiKey.Description,
			Prefix:      apiKey.Prefix,
			Permissions: apiKey.Permissions,
			Status:      apiKey.Status,
			CreatedAt:   apiKey.CreatedAt,
			UpdatedAt:   apiKey.UpdatedAt,
		}

		if apiKey.LastUsedAt != nil {
			lastUsedAtStr := apiKey.LastUsedAt.Format(time.RFC3339)
			response.LastUsedAt = &lastUsedAtStr
		}

		if apiKey.ExpiresAt != nil {
			expiresAtStr := apiKey.ExpiresAt.Format(time.RFC3339)
			response.ExpiresAt = &expiresAtStr
		}

		responses = append(responses, response)
	}

	result := map[string]interface{}{
		"data":   responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetAPIKey 获取单个API密钥
func (c *APIKeyController) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取API密钥
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 验证所有权
	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 构建响应
	response := APIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Description: apiKey.Description,
		Prefix:      apiKey.Prefix,
		Permissions: apiKey.Permissions,
		Status:      apiKey.Status,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}

	if apiKey.LastUsedAt != nil {
		lastUsedAtStr := apiKey.LastUsedAt.Format(time.RFC3339)
		response.LastUsedAt = &lastUsedAtStr
	}

	if apiKey.ExpiresAt != nil {
		expiresAtStr := apiKey.ExpiresAt.Format(time.RFC3339)
		response.ExpiresAt = &expiresAtStr
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateAPIKeyRequest 更新API密钥请求
type UpdateAPIKeyRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

// UpdateAPIKey 更新API密钥
func (c *APIKeyController) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}

	var req UpdateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取API密钥验证所有权
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Permissions != nil {
		updates["permissions"] = req.Permissions
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" {
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
		updates["status"] = *req.Status
	}

	// 更新API密钥
	err = c.apiKeyService.UpdateAPIKey(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAPIKey 删除API密钥
func (c *APIKeyController) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取API密钥验证所有权
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 删除API密钥
	err = c.apiKeyService.RevokeAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAPIKeyUsage 获取API密钥使用统计
func (c *APIKeyController) GetAPIKeyUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取API密钥验证所有权
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 获取使用统计
	usage, err := c.apiKeyService.GetUsageStats(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// getUserIDFromContext 从请求上下文获取用户ID
// 这是一个示例函数，实际实现应该从JWT令牌或会话中获取
func getUserIDFromContext(r *http.Request) int64 {
	// 这里应该从认证中间件设置的上下文中获取用户ID
	// 示例实现�?	if userID := r.Header.Get("X-User-ID"); userID != "" {
		if id, err := strconv.ParseInt(userID, 10, 64); err == nil {
			return id
		}
	}
	return 0
}
