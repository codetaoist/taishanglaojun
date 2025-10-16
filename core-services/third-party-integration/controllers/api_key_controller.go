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

// APIKeyController API?type APIKeyController struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyController API?func NewAPIKeyController(apiKeyService *services.APIKeyService) *APIKeyController {
	return &APIKeyController{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKeyRequest API
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse API
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

// APIKeyResponse API?type APIKeyResponse struct {
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

// CreateAPIKey API
func (c *APIKeyController) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// ID
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "Invalid expires_at format", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	// API
	apiKey, key, err := c.apiKeyService.GenerateAPIKey(userID, req.Name, req.Description, req.Permissions, expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 
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

// ListAPIKeys API
func (c *APIKeyController) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 
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

	// API
	apiKeys, total, err := c.apiKeyService.ListAPIKeys(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 
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

// GetAPIKey API
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

	// API
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 
	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 
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

// UpdateAPIKeyRequest API
type UpdateAPIKeyRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

// UpdateAPIKey API
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

	// API
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 
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

	// API
	err = c.apiKeyService.UpdateAPIKey(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAPIKey API
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

	// API
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// API
	err = c.apiKeyService.RevokeAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAPIKeyUsage API
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

	// API
	apiKey, err := c.apiKeyService.GetAPIKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if apiKey.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 
	usage, err := c.apiKeyService.GetUsageStats(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// getUserIDFromContext ID
// JWT
func getUserIDFromContext(r *http.Request) int64 {
	// ID
	// ?	if userID := r.Header.Get("X-User-ID"); userID != "" {
		if id, err := strconv.ParseInt(userID, 10, 64); err == nil {
			return id
		}
	}
	return 0
}

