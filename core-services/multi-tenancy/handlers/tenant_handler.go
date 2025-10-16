package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/models"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/services"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/utils"
)

// TenantHandler HTTP?
type TenantHandler struct {
	tenantService services.TenantService
	logger        utils.Logger
	validator     utils.Validator
}

// NewTenantHandler HTTP?
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

// CreateTenant 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body models.CreateTenantRequest true ""
// @Success 201 {object} models.TenantResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 409 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Invalid create tenant request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 
	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create tenant", "error", err)
		
		// 
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

// GetTenant 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.TenantResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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

// UpdateTenant 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body models.UpdateTenantRequest true ""
// @Success 200 {object} models.TenantResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Invalid update tenant request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 
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

// DeleteTenant 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.SuccessResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
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

// ListTenants 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param name query string false ""
// @Param status query string false "?
// @Param search query string false "?
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Param order_by query string false "" default(created_at)
// @Param order query string false "" default(desc)
// @Success 200 {object} models.TenantListResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
	if err := h.validator.Validate(&query); err != nil {
		h.logger.Error("Invalid list tenants query", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 
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

// GetTenantBySubdomain ?
// @Summary ?
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param subdomain path string true "?
// @Success 200 {object} models.TenantResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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

// GetTenantByDomain 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param domain path string true ""
// @Success 200 {object} models.TenantResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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

// AddTenantUser 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body models.AddTenantUserRequest true ""
// @Success 200 {object} models.SuccessResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
	if err := h.validator.Validate(&req); err != nil {
		h.logger.Error("Invalid add tenant user request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 
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

// RemoveTenantUser 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param user_id path string true "ID"
// @Success 200 {object} models.SuccessResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse ""
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
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

// ListTenantUsers 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param role query string false ""
// @Param status query string false "?
// @Param search query string false "?
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Param order_by query string false "" default(created_at)
// @Param order query string false "" default(desc)
// @Success 200 {object} models.TenantUserListResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
	if err := h.validator.Validate(&query); err != nil {
		h.logger.Error("Invalid list tenant users query", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// 
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

// GetTenantStats 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.TenantStatsResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
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

// GetTenantHealth ?
// @Summary ?
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.TenantHealthResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// ?
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

// ActivateTenant ?
// @Summary ?
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.SuccessResponse "?
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// ?
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

// SuspendTenant 
// @Summary 
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param reason query string false ""
// @Success 200 {object} models.SuccessResponse ""
// @Failure 400 {object} models.ErrorResponse ""
// @Failure 404 {object} models.ErrorResponse "?
// @Failure 500 {object} models.ErrorResponse "?
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
	
	// 
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

// 

// parsePageQuery 
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

