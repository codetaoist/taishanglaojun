package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/response"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// DatabaseConnectionHandler 数据库连接管理处理器
type DatabaseConnectionHandler struct {
	service *services.DatabaseConnectionService
}

// NewDatabaseConnectionHandler 创建数据库连接管理处理器
func NewDatabaseConnectionHandler(service *services.DatabaseConnectionService) *DatabaseConnectionHandler {
	return &DatabaseConnectionHandler{
		service: service,
	}
}

// CreateConnection 创建数据库连接配置
// @Summary 创建数据库连接配置
// @Description 创建新的数据库连接配置
// @Tags database-connections
// @Accept json
// @Produce json
// @Param connection body models.DatabaseConnectionForm true "数据库连接配置"
// @Success 201 {object} response.Response{data=models.DatabaseConnection}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections [post]
func (h *DatabaseConnectionHandler) CreateConnection(c *gin.Context) {
	var form models.DatabaseConnectionForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// 验证必填字段
	if form.Name == "" || form.Type == "" || form.Host == "" || form.Username == "" {
		response.BadRequest(c, "Missing required fields: name, type, host, and username are required")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	connection, err := h.service.CreateConnection(c.Request.Context(), &form, userID.(string))
	if err != nil {
		response.InternalServerError(c, "Failed to create connection: "+err.Error())
		return
	}

	response.Success(c, connection)
}

// GetConnections 获取数据库连接列表
// @Summary 获取数据库连接列表
// @Description 获取数据库连接配置列表，支持搜索、过滤和分页
// @Tags database-connections
// @Accept json
// @Produce json
// @Param search query string false "搜索关键词"
// @Param type query string false "数据库类型"
// @Param tags query string false "标签（逗号分隔）"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]models.DatabaseConnection}}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections [get]
func (h *DatabaseConnectionHandler) GetConnections(c *gin.Context) {
	query := &models.DatabaseConnectionQuery{
		Search:    c.Query("search"),
		Type:      models.DatabaseType(c.Query("type")),
		Tags:      c.Query("tags"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// 解析分页参数
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			query.PageSize = ps
		}
	}

	connections, total, err := h.service.GetConnections(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get connections: "+err.Error())
		return
	}

	// 构建分页响应
	paginatedResponse := response.PaginatedResponse{
		Items:    connections,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	response.Success(c, paginatedResponse)
}

// GetConnection 获取单个数据库连接配置
// @Summary 获取数据库连接配置详情
// @Description 根据ID获取数据库连接配置详情
// @Tags database-connections
// @Accept json
// @Produce json
// @Param id path string true "连接ID"
// @Success 200 {object} response.Response{data=models.DatabaseConnection}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/{id} [get]
func (h *DatabaseConnectionHandler) GetConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Missing connection ID")
		return
	}

	connection, err := h.service.GetConnection(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "connection not found" {
			response.NotFound(c, "Connection not found")
			return
		}
		response.InternalServerError(c, "Failed to get connection: "+err.Error())
		return
	}

	response.Success(c, connection)
}

// UpdateConnection 更新数据库连接配置
// @Summary 更新数据库连接配置
// @Description 更新指定ID的数据库连接配置
// @Tags database-connections
// @Accept json
// @Produce json
// @Param id path string true "连接ID"
// @Param connection body models.DatabaseConnectionForm true "数据库连接配置"
// @Success 200 {object} response.Response{data=models.DatabaseConnection}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/{id} [put]
func (h *DatabaseConnectionHandler) UpdateConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Missing connection ID")
		return
	}

	var form models.DatabaseConnectionForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// 验证必填字段
	if form.Name == "" || form.Type == "" || form.Host == "" || form.Username == "" {
		response.BadRequest(c, "Missing required fields: name, type, host, and username are required")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	connection, err := h.service.UpdateConnection(c.Request.Context(), id, &form, userID.(string))
	if err != nil {
		if err.Error() == "connection not found" {
			response.NotFound(c, "Connection not found")
			return
		}
		response.InternalServerError(c, "Failed to update connection: "+err.Error())
		return
	}

	response.Success(c, connection)
}

// DeleteConnection 删除数据库连接配置
// @Summary 删除数据库连接配置
// @Description 删除指定ID的数据库连接配置
// @Tags database-connections
// @Accept json
// @Produce json
// @Param id path string true "连接ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/{id} [delete]
func (h *DatabaseConnectionHandler) DeleteConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Missing connection ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err := h.service.DeleteConnection(c.Request.Context(), id, userID.(string))
	if err != nil {
		if err.Error() == "connection not found" {
			response.NotFound(c, "Connection not found")
			return
		}
		response.InternalServerError(c, "Failed to delete connection: "+err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Connection deleted successfully"})
}

// TestConnection 测试数据库连接
// @Summary 测试数据库连接
// @Description 测试数据库连接配置是否有效
// @Tags database-connections
// @Accept json
// @Produce json
// @Param connection body models.DatabaseConnectionForm true "数据库连接配置"
// @Success 200 {object} response.Response{data=models.DatabaseConnectionTest}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/test [post]
func (h *DatabaseConnectionHandler) TestConnection(c *gin.Context) {
	var form models.DatabaseConnectionForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// 验证必填字段
	if form.Type == "" || form.Host == "" || form.Username == "" {
		response.BadRequest(c, "Missing required fields: type, host, and username are required")
		return
	}

	result, err := h.service.TestConnection(c.Request.Context(), &form)
	if err != nil {
		response.InternalServerError(c, "Failed to test connection: "+err.Error())
		return
	}

	response.Success(c, result)
}

// TestSavedConnection 测试已保存的数据库连接
// @Summary 测试已保存的数据库连接
// @Description 测试指定ID的已保存数据库连接
// @Tags database-connections
// @Accept json
// @Produce json
// @Param id path string true "连接ID"
// @Success 200 {object} response.Response{data=models.DatabaseConnectionTest}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/{id}/test [post]
func (h *DatabaseConnectionHandler) TestSavedConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Missing connection ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	result, err := h.service.TestSavedConnection(c.Request.Context(), id, userID.(string))
	if err != nil {
		if err.Error() == "connection not found" {
			response.NotFound(c, "Connection not found")
			return
		}
		response.InternalServerError(c, "Failed to test connection: "+err.Error())
		return
	}

	response.Success(c, result)
}

// GetConnectionsStatus 获取所有连接状态
// @Summary 获取所有连接状态
// @Description 获取所有数据库连接的状态信息
// @Tags database-connections
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.DatabaseConnectionStatus}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/status [get]
func (h *DatabaseConnectionHandler) GetConnectionsStatus(c *gin.Context) {
	statuses, err := h.service.GetConnectionsStatus(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to get connection statuses: "+err.Error())
		return
	}

	response.Success(c, statuses)
}

// RefreshConnectionStatus 刷新连接状态
// @Summary 刷新连接状态
// @Description 刷新指定连接的状态信息
// @Tags database-connections
// @Accept json
// @Produce json
// @Param id path string true "连接ID"
// @Success 200 {object} response.Response{data=models.DatabaseConnectionStatus}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/{id}/status [post]
func (h *DatabaseConnectionHandler) RefreshConnectionStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Missing connection ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	status, err := h.service.RefreshConnectionStatus(c.Request.Context(), id, userID.(string))
	if err != nil {
		if err.Error() == "connection not found" {
			response.NotFound(c, "Connection not found")
			return
		}
		response.InternalServerError(c, "Failed to refresh connection status: "+err.Error())
		return
	}

	response.Success(c, status)
}

// GetConnectionStats 获取连接统计信息
// @Summary 获取连接统计信息
// @Description 获取数据库连接的统计信息
// @Tags database-connections
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=models.DatabaseConnectionStats}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/database/connections/stats [get]
func (h *DatabaseConnectionHandler) GetConnectionStats(c *gin.Context) {
	stats, err := h.service.GetConnectionStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to get connection stats: "+err.Error())
		return
	}

	response.Success(c, stats)
}

// GetDatabaseTypes 获取支持的数据库类型
// @Summary 获取支持的数据库类型
// @Description 获取系统支持的数据库类型配置
// @Tags database-connections
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.DatabaseTypeConfig}
// @Router /api/v1/admin/database/types [get]
func (h *DatabaseConnectionHandler) GetDatabaseTypes(c *gin.Context) {
	types := models.GetDatabaseTypeConfigs()
	response.Success(c, types)
}