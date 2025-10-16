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

// PluginController 插件控制器
type PluginController struct {
	pluginService *services.PluginService
}

// NewPluginController 创建新的插件控制器
func NewPluginController(pluginService *services.PluginService) *PluginController {
	return &PluginController{
		pluginService: pluginService,
	}
}

// InstallPluginRequest 安装插件请求
type InstallPluginRequest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Source      string                 `json:"source"`
	Config      map[string]interface{} `json:"config,omitempty"`
	AutoEnable  bool                   `json:"auto_enable"`
}

// PluginResponse 插件响应
type PluginResponse struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Status      string                 `json:"status"`
	IsEnabled   bool                   `json:"is_enabled"`
	Config      map[string]interface{} `json:"config"`
	Manifest    map[string]interface{} `json:"manifest"`
	InstalledAt time.Time              `json:"installed_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// InstallPlugin 安装插件
func (c *PluginController) InstallPlugin(w http.ResponseWriter, r *http.Request) {
	var req InstallPluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求
	if req.Name == "" || req.Version == "" || req.Source == "" {
		http.Error(w, "Name, version, and source are required", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 安装插件
	plugin, err := c.pluginService.InstallPlugin(userID, req.Name, req.Version, req.Source, req.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 如果需要自动启?	if req.AutoEnable {
		err = c.pluginService.EnablePlugin(plugin.ID)
		if err != nil {
			// 记录错误但不失败
			// log.Printf("Failed to auto-enable plugin %d: %v", plugin.ID, err)
		}
	}

	// 重新获取插件信息
	plugin, err = c.pluginService.GetPlugin(plugin.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := c.buildPluginResponse(plugin)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListPlugins 获取插件列表
func (c *PluginController) ListPlugins(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 解析查询参数
	limit := 20
	offset := 0
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

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

	// 获取插件列表
	plugins, total, err := c.pluginService.ListPlugins(userID, status, search, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应
	var responses []PluginResponse
	for _, plugin := range plugins {
		responses = append(responses, c.buildPluginResponse(plugin))
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

// GetPlugin 获取单个插件
func (c *PluginController) GetPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件
	plugin, err := c.pluginService.GetPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 验证所有权
	if plugin.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	response := c.buildPluginResponse(plugin)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdatePluginRequest 更新插件请求
type UpdatePluginRequest struct {
	Config map[string]interface{} `json:"config,omitempty"`
}

// UpdatePlugin 更新插件配置
func (c *PluginController) UpdatePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
		return
	}

	var req UpdatePluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件验证所有权
	plugin, err := c.pluginService.GetPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if plugin.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 更新插件配置
	err = c.pluginService.UpdatePluginConfig(id, req.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnablePlugin 启用插件
func (c *PluginController) EnablePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件验证所有权
	plugin, err := c.pluginService.GetPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if plugin.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 启用插件
	err = c.pluginService.EnablePlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DisablePlugin 禁用插件
func (c *PluginController) DisablePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件验证所有权
	plugin, err := c.pluginService.GetPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if plugin.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 禁用插件
	err = c.pluginService.DisablePlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UninstallPlugin 卸载插件
func (c *PluginController) UninstallPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件验证所有权
	plugin, err := c.pluginService.GetPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if plugin.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 卸载插件
	err = c.pluginService.UninstallPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePluginVersion 更新插件版本
func (c *PluginController) UpdatePluginVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
		return
	}

	type UpdateVersionRequest struct {
		Version string `json:"version"`
		Source  string `json:"source"`
	}

	var req UpdateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Version == "" || req.Source == "" {
		http.Error(w, "Version and source are required", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件验证所有权
	plugin, err := c.pluginService.GetPlugin(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if plugin.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 更新插件版本
	err = c.pluginService.UpdatePlugin(id, req.Version, req.Source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPluginStats 获取插件统计信息
func (c *PluginController) GetPluginStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取插件统计
	stats, err := c.pluginService.GetPluginStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// buildPluginResponse 构建插件响应
func (c *PluginController) buildPluginResponse(plugin *models.Plugin) PluginResponse {
	return PluginResponse{
		ID:          plugin.ID,
		Name:        plugin.Name,
		Version:     plugin.Version,
		Description: plugin.Description,
		Author:      plugin.Author,
		Status:      plugin.Status,
		IsEnabled:   plugin.IsEnabled,
		Config:      plugin.Config,
		Manifest:    plugin.Manifest,
		InstalledAt: plugin.InstalledAt,
		UpdatedAt:   plugin.UpdatedAt,
	}
}

