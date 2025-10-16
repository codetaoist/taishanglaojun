package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// AppModuleHandler 应用模块处理器
type AppModuleHandler struct {
	service *services.AppModuleService
	logger  *zap.Logger
}

// NewAppModuleHandler 创建应用模块处理器
func NewAppModuleHandler(service *services.AppModuleService, logger *zap.Logger) *AppModuleHandler {
	return &AppModuleHandler{
		service: service,
		logger:  logger,
	}
}

// GetUserModulesRequest 获取用户模块请求
type GetUserModulesRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// GetUserModulesResponse 获取用户模块响应
type GetUserModulesResponse struct {
	Modules []models.AppModule `json:"modules"`
}

// GetUserModules 获取用户模块
func (h *AppModuleHandler) GetUserModules(c *gin.Context) {
	// JWT
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	_, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	//
	roleStr, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	userRole := models.UserRole(roleStr.(string))

	//
	modules, err := h.service.GetModulesByUserRole(userRole)
	if err != nil {
		h.logger.Error("Failed to get user modules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, GetUserModulesResponse{
		Modules: modules,
	})
}

// CreateModuleRequest 创建模块请求
type CreateModuleRequest struct {
	Name         string                `json:"name" binding:"required"`
	DisplayName  string                `json:"display_name" binding:"required"`
	Description  string                `json:"description"`
	Category     models.ModuleCategory `json:"category" binding:"required"`
	Icon         string                `json:"icon"`
	Path         string                `json:"path" binding:"required"`
	RequiredRole models.UserRole       `json:"required_role" binding:"required"`
	IsCore       bool                  `json:"is_core"`
	IsEnabled    bool                  `json:"is_enabled"`
	AutoStart    bool                  `json:"auto_start"`
	Priority     int                   `json:"priority"`
	Version      string                `json:"version"`
}

// CreateModule 创建模块
func (h *AppModuleHandler) CreateModule(c *gin.Context) {
	//
	if !h.checkAdminPermission(c) {
		return
	}

	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CreateModuleRequest: " + err.Error()})
		return
	}

	module := &models.AppModule{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		Category:     req.Category,
		Icon:         req.Icon,
		Path:         req.Path,
		RequiredRole: req.RequiredRole,
		IsCore:       req.IsCore,
		IsEnabled:    req.IsEnabled,
		AutoStart:    req.AutoStart,
		Priority:     req.Priority,
		Version:      req.Version,
	}

	if err := h.service.CreateModule(module); err != nil {
		h.logger.Error("Failed to create module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "模块创建成功",
		"module":  module,
	})
}

// UpdateModuleRequest 更新模块请求
type UpdateModuleRequest struct {
	DisplayName  string                `json:"display_name"`
	Description  string                `json:"description"`
	Category     models.ModuleCategory `json:"category"`
	Icon         string                `json:"icon"`
	Path         string                `json:"path"`
	RequiredRole models.UserRole       `json:"required_role"`
	IsCore       bool                  `json:"is_core"`
	IsEnabled    bool                  `json:"is_enabled"`
	AutoStart    bool                  `json:"auto_start"`
	Priority     int                   `json:"priority"`
	Version      string                `json:"version"`
}

// UpdateModule 更新模块
func (h *AppModuleHandler) UpdateModule(c *gin.Context) {
	//
	if !h.checkAdminPermission(c) {
		return
	}

	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UpdateModuleRequest: " + err.Error()})
		return
	}

	//
	module, err := h.service.GetModuleByID(moduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模块不存在"})
		return
	}

	//
	if req.DisplayName != "" {
		module.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		module.Description = req.Description
	}
	if req.Category != "" {
		module.Category = req.Category
	}
	if req.Icon != "" {
		module.Icon = req.Icon
	}
	if req.Path != "" {
		module.Path = req.Path
	}
	if req.RequiredRole != "" {
		module.RequiredRole = req.RequiredRole
	}
	module.IsCore = req.IsCore
	module.IsEnabled = req.IsEnabled
	module.AutoStart = req.AutoStart
	module.Priority = req.Priority
	if req.Version != "" {
		module.Version = req.Version
	}

	if err := h.service.UpdateModule(module); err != nil {
		h.logger.Error("Failed to update module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"module":  module,
	})
}

// DeleteModule 删除模块
func (h *AppModuleHandler) DeleteModule(c *gin.Context) {
	//
	if !h.checkAdminPermission(c) {
		return
	}

	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	if err := h.service.DeleteModule(moduleID); err != nil {
		h.logger.Error("Failed to delete module", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// GetModule 获取模块
func (h *AppModuleHandler) GetModule(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	module, err := h.service.GetModuleByID(moduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模块不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"module": module})
}

// SetUserModulePermissionRequest 设置用户模块权限请求
type SetUserModulePermissionRequest struct {
	ModuleID string `json:"module_id" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

// SetUserModulePermission 设置用户模块权限
func (h *AppModuleHandler) SetUserModulePermission(c *gin.Context) {
	// JWT
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req SetUserModulePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ": " + err.Error()})
		return
	}

	moduleID, err := uuid.Parse(req.ModuleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	if err := h.service.SetUserModulePermission(userID, moduleID, req.Enabled); err != nil {
		h.logger.Error("Failed to set user module permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// GetUserPreference 获取用户偏好
func (h *AppModuleHandler) GetUserPreference(c *gin.Context) {
	// JWT
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	preference, err := h.service.GetUserPreference(userID)
	if err != nil {
		h.logger.Error("Failed to get user preference", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preference": preference})
}

// UpdateUserPreferenceRequest 更新用户偏好请求
type UpdateUserPreferenceRequest struct {
	Theme     string `json:"theme"`
	Language  string `json:"language"`
	MenuStyle string `json:"menu_style"`
	AutoStart bool   `json:"auto_start"`
	Settings  string `json:"settings"`
}

// UpdateUserPreference 更新用户偏好
func (h *AppModuleHandler) UpdateUserPreference(c *gin.Context) {
	// JWT
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req UpdateUserPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ": " + err.Error()})
		return
	}

	//
	preference, err := h.service.GetUserPreference(userID)
	if err != nil {
		h.logger.Error("Failed to get user preference", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	//
	if req.Theme != "" {
		preference.Theme = req.Theme
	}
	if req.Language != "" {
		preference.Language = req.Language
	}
	if req.MenuStyle != "" {
		preference.MenuStyle = req.MenuStyle
	}
	preference.AutoStart = req.AutoStart
	if req.Settings != "" {
		preference.Settings = req.Settings
	}

	if err := h.service.UpdateUserPreference(preference); err != nil {
		h.logger.Error("Failed to update user preference", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "",
		"preference": preference,
	})
}

// InitializeDefaultModules 初始化默认模块
func (h *AppModuleHandler) InitializeDefaultModules(c *gin.Context) {
	// JWT
	if !h.checkSuperAdminPermission(c) {
		return
	}

	if err := h.service.InitializeDefaultModules(); err != nil {
		h.logger.Error("Failed to initialize default modules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// checkAdminPermission 检查管理员权限
func (h *AppModuleHandler) checkAdminPermission(c *gin.Context) bool {
	roleStr, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return false
	}

	role := models.UserRole(roleStr.(string))
	if role != models.RoleAdmin && role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": ""})
		return false
	}

	return true
}

// checkSuperAdminPermission 检查超级管理员权限
func (h *AppModuleHandler) checkSuperAdminPermission(c *gin.Context) bool {
	roleStr, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return false
	}

	role := models.UserRole(roleStr.(string))
	if role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": ""})
		return false
	}

	return true
}
