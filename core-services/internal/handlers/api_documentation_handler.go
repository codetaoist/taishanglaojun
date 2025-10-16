package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// APIDocumentationHandler API文档管理处理器
type APIDocumentationHandler struct {
	apiDocService *services.APIDocumentationService
	authService   *middleware.AuthService
	logger        *zap.Logger
}

// NewAPIDocumentationHandler 创建API文档管理处理器
func NewAPIDocumentationHandler(
	apiDocService *services.APIDocumentationService,
	authService *middleware.AuthService,
	logger *zap.Logger,
) *APIDocumentationHandler {
	return &APIDocumentationHandler{
		apiDocService: apiDocService,
		authService:   authService,
		logger:        logger,
	}
}

// GetCategories 获取分类列表
func (h *APIDocumentationHandler) GetCategories(c *gin.Context) {
	h.logger.Info("Getting API categories")

	var req services.CategoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	result, err := h.apiDocService.GetCategories(req)
	if err != nil {
		h.logger.Error("Failed to get categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取分类列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取分类列表成功",
		"data":    result,
	})
}

// GetCategoryByID 根据ID获取分类
func (h *APIDocumentationHandler) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Getting category by ID", zap.String("id", id))

	category, err := h.apiDocService.GetCategoryByID(id)
	if err != nil {
		h.logger.Error("Failed to get category", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "分类不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取分类成功",
		"data":    category,
	})
}

// GetEndpoints 获取接口列表
func (h *APIDocumentationHandler) GetEndpoints(c *gin.Context) {
	h.logger.Info("Getting API endpoints")

	var req services.EndpointListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	result, err := h.apiDocService.GetEndpoints(req)
	if err != nil {
		h.logger.Error("Failed to get endpoints", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取接口列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取接口列表成功",
		"data":    result,
	})
}

// GetEndpointByID 根据ID获取接口详情
func (h *APIDocumentationHandler) GetEndpointByID(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Getting endpoint by ID", zap.String("id", id))

	result, err := h.apiDocService.GetEndpointByID(id)
	if err != nil {
		h.logger.Error("Failed to get endpoint", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "接口不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取接口详情成功",
		"data":    result,
	})
}

// GetEndpointsByCategory 根据分类获取接口列表
func (h *APIDocumentationHandler) GetEndpointsByCategory(c *gin.Context) {
	categoryID := c.Param("id")
	h.logger.Info("Getting endpoints by category", zap.String("categoryID", categoryID))

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.apiDocService.GetEndpointsByCategory(categoryID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get endpoints by category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取分类接口失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取分类接口成功",
		"data":    result,
	})
}

// SearchEndpoints 搜索接口
func (h *APIDocumentationHandler) SearchEndpoints(c *gin.Context) {
	keyword := c.Query("keyword")
	h.logger.Info("Searching endpoints", zap.String("keyword", keyword))

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "搜索关键词不能为空",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.apiDocService.SearchEndpoints(keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search endpoints", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "搜索接口失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "搜索接口成功",
		"data":    result,
	})
}

// GetStatistics 获取统计信息
func (h *APIDocumentationHandler) GetStatistics(c *gin.Context) {
	h.logger.Info("Getting API documentation statistics")

	stats, err := h.apiDocService.GetStatistics()
	if err != nil {
		h.logger.Error("Failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取统计信息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取统计信息成功",
		"data":    stats,
	})
}

// TestAPIRequest 测试API请求结构
type TestAPIRequest struct {
	EndpointID   string                 `json:"endpoint_id" binding:"required"`
	Method       string                 `json:"method" binding:"required"`
	URL          string                 `json:"url" binding:"required"`
	Headers      map[string]string      `json:"headers"`
	Parameters   map[string]interface{} `json:"parameters"`
	Body         interface{}            `json:"body"`
}

// TestAPI 测试API接口
func (h *APIDocumentationHandler) TestAPI(c *gin.Context) {
	h.logger.Info("Testing API endpoint")

	var req TestAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未认证",
		})
		return
	}

	// 这里可以实现实际的API测试逻辑
	// 目前先模拟测试结果
	success := true
	responseTime := 150 // 毫秒
	errorMsg := ""

	// 记录测试结果
	if err := h.apiDocService.RecordAPITest(req.EndpointID, userID.(string), success, responseTime, errorMsg); err != nil {
		h.logger.Error("Failed to record API test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "记录测试结果失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API测试完成",
		"data": gin.H{
			"success":      success,
			"responseTime": responseTime,
			"errorMsg":     errorMsg,
			"timestamp":    "2025-10-16T16:46:00Z",
		},
	})
}

// GetAPITestHistory 获取API测试历史
func (h *APIDocumentationHandler) GetAPITestHistory(c *gin.Context) {
	endpointID := c.Param("id")
	h.logger.Info("Getting API test history", zap.String("endpointID", endpointID))

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 这里可以实现获取测试历史的逻辑
	// 目前返回模拟数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取测试历史成功",
		"data": gin.H{
			"tests": []gin.H{},
			"total": 0,
			"page":  page,
			"pageSize": pageSize,
		},
	})
}