package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"taishanglaojun/core-services/multi-tenancy/models"
	"taishanglaojun/core-services/multi-tenancy/services"
	"taishanglaojun/core-services/multi-tenancy/utils"
)

// TenantHandler 租户HTTP处理器
type TenantHandler struct {
	tenantService services.TenantService
	logger        utils.Logger
	validator     utils.Validator
}

// NewTenantHandler 创建租户HTTP处理器
func NewTenantHandler(
	tenantService services.TenantService,
	logger utils.Logger,
	validator utils.Validator,
) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger,
		validator:     validator,
	}
}

// CreateTenant 创建租户
// @Summary 创建租户
// @Description 创建新的租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param request body models.CreateTenantRequest true "创建租户请求"
// @Success 201 {object} models.TenantResponse "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 409 {object} models.ErrorResponse "租户已存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req models.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create tenant request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	// 验证请求参数
	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Invalid create tenant request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 创建租户
	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create tenant", "error", err)
		
		// 根据错误类型返回不同的状态码
		if err.Error() == "subdomain already exists" || err.Error() == "domain already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to create tenant",
		})
		return
	}
	
	c.JSON(http.StatusCreated, tenant)
}

// GetTenant 获取租户详情
// @Summary 获取租户详情
// @Description 根据租户ID获取租户详细信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Success 200 {object} models.TenantResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	tenant, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, tenant)
}

// UpdateTenant 更新租户
// @Summary 更新租户
// @Description 更新租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Param request body models.UpdateTenantRequest true "更新租户请求"
// @Success 200 {object} models.TenantResponse "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	var req models.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update tenant request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	// 验证请求参数
	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Invalid update tenant request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 更新租户
	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to update tenant", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to update tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, tenant)
}

// DeleteTenant 删除租户
// @Summary 删除租户
// @Description 删除指定的租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Success 200 {object} models.SuccessResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	// 删除租户
	if err := h.tenantService.DeleteTenant(c.Request.Context(), tenantID); err != nil {
		h.logger.Error("Failed to delete tenant", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to delete tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Tenant deleted successfully",
	})
}

// ListTenants 列出租户
// @Summary 列出租户
// @Description 获取租户列表，支持分页和过滤
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param name query string false "租户名称过滤"
// @Param status query string false "租户状态过滤"
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param order_by query string false "排序字段" default(created_at)
// @Param order query string false "排序方向" default(desc)
// @Success 200 {object} models.TenantListResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	var query models.TenantQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.logger.Error("Failed to bind list tenants query", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}
	
	// 验证查询参数
	if err := h.validator.Validate(&query); err != nil {
		h.logger.Error("Invalid list tenants query", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 获取租户列表
	result, err := h.tenantService.ListTenants(c.Request.Context(), &query)
	if err != nil {
		h.logger.Error("Failed to list tenants", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list tenants",
		})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// GetTenantBySubdomain 通过子域名获取租户
// @Summary 通过子域名获取租户
// @Description 根据子域名获取租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param subdomain path string true "子域名"
// @Success 200 {object} models.TenantResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/subdomain/{subdomain} [get]
func (h *TenantHandler) GetTenantBySubdomain(c *gin.Context) {
	subdomain := c.Param("subdomain")
	if subdomain == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid subdomain",
			Message: "Subdomain cannot be empty",
		})
		return
	}
	
	tenant, err := h.tenantService.GetTenantBySubdomain(c.Request.Context(), subdomain)
	if err != nil {
		h.logger.Error("Failed to get tenant by subdomain", "error", err, "subdomain", subdomain)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, tenant)
}

// GetTenantByDomain 通过域名获取租户
// @Summary 通过域名获取租户
// @Description 根据域名获取租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Param domain path string true "域名"
// @Success 200 {object} models.TenantResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/domain/{domain} [get]
func (h *TenantHandler) GetTenantByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid domain",
			Message: "Domain cannot be empty",
		})
		return
	}
	
	tenant, err := h.tenantService.GetTenantByDomain(c.Request.Context(), domain)
	if err != nil {
		h.logger.Error("Failed to get tenant by domain", "error", err, "domain", domain)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, tenant)
}

// AddTenantUser 添加租户用户
// @Summary 添加租户用户
// @Description 将用户添加到租户
// @Tags 租户用户管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Param request body models.AddTenantUserRequest true "添加用户请求"
// @Success 200 {object} models.SuccessResponse "添加成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/users [post]
func (h *TenantHandler) AddTenantUser(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	var req models.AddTenantUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind add tenant user request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	// 验证请求参数
	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Invalid add tenant user request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 添加租户用户
	if err := h.tenantService.AddTenantUser(c.Request.Context(), tenantID, &req); err != nil {
		h.logger.Error("Failed to add tenant user", "error", err, "tenant_id", tenantID, "user_id", req.UserID)
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to add tenant user",
		})
		return
	}
	
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User added to tenant successfully",
	})
}

// RemoveTenantUser 移除租户用户
// @Summary 移除租户用户
// @Description 从租户中移除用户
// @Tags 租户用户管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Param user_id path string true "用户ID"
// @Success 200 {object} models.SuccessResponse "移除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户或用户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/users/{user_id} [delete]
func (h *TenantHandler) RemoveTenantUser(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid UUID",
		})
		return
	}
	
	// 移除租户用户
	if err := h.tenantService.RemoveTenantUser(c.Request.Context(), tenantID, userID); err != nil {
		h.logger.Error("Failed to remove tenant user", "error", err, "tenant_id", tenantID, "user_id", userID)
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to remove tenant user",
		})
		return
	}
	
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User removed from tenant successfully",
	})
}

// ListTenantUsers 列出租户用户
// @Summary 列出租户用户
// @Description 获取租户的用户列表
// @Tags 租户用户管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Param role query string false "角色过滤"
// @Param status query string false "状态过滤"
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param order_by query string false "排序字段" default(created_at)
// @Param order query string false "排序方向" default(desc)
// @Success 200 {object} models.TenantUserListResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/users [get]
func (h *TenantHandler) ListTenantUsers(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	var query models.TenantUserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.logger.Error("Failed to bind list tenant users query", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}
	
	// 验证查询参数
	if err := h.validator.Validate(&query); err != nil {
		h.logger.Error("Invalid list tenant users query", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 获取租户用户列表
	result, err := h.tenantService.ListTenantUsers(c.Request.Context(), tenantID, &query)
	if err != nil {
		h.logger.Error("Failed to list tenant users", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list tenant users",
		})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// GetTenantStats 获取租户统计
// @Summary 获取租户统计
// @Description 获取租户的使用统计信息
// @Tags 租户统计
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Success 200 {object} models.TenantStatsResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/stats [get]
func (h *TenantHandler) GetTenantStats(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	// 获取租户统计
	stats, err := h.tenantService.GetTenantStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant stats", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get tenant stats",
		})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

// GetTenantHealth 获取租户健康状态
// @Summary 获取租户健康状态
// @Description 获取租户的健康检查结果
// @Tags 租户监控
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Success 200 {object} models.TenantHealthResponse "获取成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/health [get]
func (h *TenantHandler) GetTenantHealth(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	// 获取租户健康状态
	health, err := h.tenantService.GetTenantHealth(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant health", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get tenant health",
		})
		return
	}
	
	c.JSON(http.StatusOK, health)
}

// ActivateTenant 激活租户
// @Summary 激活租户
// @Description 激活指定的租户
// @Tags 租户状态管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Success 200 {object} models.SuccessResponse "激活成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/activate [post]
func (h *TenantHandler) ActivateTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	// 激活租户
	if err := h.tenantService.ActivateTenant(c.Request.Context(), tenantID); err != nil {
		h.logger.Error("Failed to activate tenant", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to activate tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Tenant activated successfully",
	})
}

// SuspendTenant 暂停租户
// @Summary 暂停租户
// @Description 暂停指定的租户
// @Tags 租户状态管理
// @Accept json
// @Produce json
// @Param id path string true "租户ID"
// @Param reason query string false "暂停原因"
// @Success 200 {object} models.SuccessResponse "暂停成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 404 {object} models.ErrorResponse "租户不存在"
// @Failure 500 {object} models.ErrorResponse "内部服务器错误"
// @Router /api/v1/tenants/{id}/suspend [post]
func (h *TenantHandler) SuspendTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid tenant ID",
			Message: "Tenant ID must be a valid UUID",
		})
		return
	}
	
	reason := c.Query("reason")
	
	// 暂停租户
	if err := h.tenantService.SuspendTenant(c.Request.Context(), tenantID, reason); err != nil {
		h.logger.Error("Failed to suspend tenant", "error", err, "tenant_id", tenantID)
		
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not found",
				Message: "Tenant not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to suspend tenant",
		})
		return
	}
	
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Tenant suspended successfully",
	})
}

// 辅助方法

// parsePageQuery 解析分页查询参数
func (h *TenantHandler) parsePageQuery(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}
	
	return page, pageSize
}