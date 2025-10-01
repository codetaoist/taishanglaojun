package permission

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Service 权限服务结构体
type Service struct {
	db *sql.DB
}

// Permission 权限结构体
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserPermission 用户权限结构体
type UserPermission struct {
	UserID       int        `json:"user_id"`
	PermissionID int        `json:"permission_id"`
	Permission   Permission `json:"permission"`
	GrantedAt    time.Time  `json:"granted_at"`
	GrantedBy    int        `json:"granted_by"`
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID         int    `json:"user_id" binding:"required"`
	PermissionName string `json:"permission_name" binding:"required"`
	Resource       string `json:"resource,omitempty"`
}

// BatchPermissionCheckRequest 批量权限检查请求
type BatchPermissionCheckRequest struct {
	UserID      int      `json:"user_id" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
	Level   int    `json:"level,omitempty"`
}

// BatchPermissionCheckResponse 批量权限检查响应
type BatchPermissionCheckResponse struct {
	Results map[string]PermissionCheckResponse `json:"results"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Level       int    `json:"level" binding:"required,min=1,max=9"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Level       int    `json:"level,omitempty"`
}

// GrantPermissionRequest 授予权限请求
type GrantPermissionRequest struct {
	PermissionID int `json:"permission_id" binding:"required"`
	GrantedBy    int `json:"granted_by" binding:"required"`
}

// UpdateUserLevelRequest 更新用户等级请求
type UpdateUserLevelRequest struct {
	Level     int `json:"level" binding:"required,min=1,max=9"`
	UpdatedBy int `json:"updated_by" binding:"required"`
}

// NewService 创建新的权限服务实例
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CheckPermission 检查用户权限
func (s *Service) CheckPermission(c *gin.Context) {
	var req PermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	allowed, reason, level := s.checkUserPermission(req.UserID, req.PermissionName)

	c.JSON(http.StatusOK, PermissionCheckResponse{
		Allowed: allowed,
		Reason:  reason,
		Level:   level,
	})

	// 记录权限检查审计
	s.logPermissionCheck(req.UserID, req.PermissionName, allowed, reason)
}

// BatchCheckPermissions 批量检查用户权限
func (s *Service) BatchCheckPermissions(c *gin.Context) {
	var req BatchPermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	results := make(map[string]PermissionCheckResponse)
	for _, permissionName := range req.Permissions {
		allowed, reason, level := s.checkUserPermission(req.UserID, permissionName)
		results[permissionName] = PermissionCheckResponse{
			Allowed: allowed,
			Reason:  reason,
			Level:   level,
		}
	}

	c.JSON(http.StatusOK, BatchPermissionCheckResponse{
		Results: results,
	})
}

// ListPermissions 获取权限列表
func (s *Service) ListPermissions(c *gin.Context) {
	rows, err := s.db.Query(`
		SELECT id, name, description, level, created_at 
		FROM permissions 
		ORDER BY level ASC, name ASC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败"})
		return
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Level, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描权限数据失败"})
			return
		}
		permissions = append(permissions, p)
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// GetPermission 获取单个权限
func (s *Service) GetPermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	var p Permission
	err = s.db.QueryRow(`
		SELECT id, name, description, level, created_at 
		FROM permissions WHERE id = $1`, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Level, &p.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败"})
		}
		return
	}

	c.JSON(http.StatusOK, p)
}

// CreatePermission 创建权限
func (s *Service) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 检查权限名称是否已存在
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM permissions WHERE name = $1)", req.Name).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "权限名称已存在"})
		return
	}

	var permissionID int
	err = s.db.QueryRow(`
		INSERT INTO permissions (name, description, level) 
		VALUES ($1, $2, $3) RETURNING id`,
		req.Name, req.Description, req.Level).Scan(&permissionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建权限失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "权限创建成功",
		"permission_id": permissionID,
	})
}

// UpdatePermission 更新权限
func (s *Service) UpdatePermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 构建更新语句
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
		argIndex++
	}
	if req.Description != "" {
		updateFields = append(updateFields, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.Description)
		argIndex++
	}
	if req.Level > 0 {
		updateFields = append(updateFields, fmt.Sprintf("level = $%d", argIndex))
		args = append(args, req.Level)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供更新字段"})
		return
	}

	args = append(args, id)
	updateSQL := fmt.Sprintf("UPDATE permissions SET %s WHERE id = $%d",
		string(updateFields[0]), argIndex)
	for i := 1; i < len(updateFields); i++ {
		updateSQL = fmt.Sprintf("%s, %s", updateSQL, updateFields[i])
	}

	result, err := s.db.Exec(updateSQL, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新权限失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限更新成功"})
}

// DeletePermission 删除权限
func (s *Service) DeletePermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	result, err := s.db.Exec("DELETE FROM permissions WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除权限失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "权限不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限删除成功"})
}

// GetUserPermissions 获取用户权限
func (s *Service) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取用户权限等级
	var userLevel int
	err = s.db.QueryRow("SELECT permission_level FROM users WHERE id = $1", userID).Scan(&userLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		}
		return
	}

	// 获取用户可访问的所有权限（等级小于等于用户等级的权限）
	rows, err := s.db.Query(`
		SELECT id, name, description, level, created_at 
		FROM permissions 
		WHERE level <= $1 
		ORDER BY level ASC, name ASC`, userLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败"})
		return
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Level, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描权限数据失败"})
			return
		}
		permissions = append(permissions, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"user_level":  userLevel,
		"permissions": permissions,
	})
}

// GrantPermission 授予权限（预留接口，当前基于等级系统）
func (s *Service) GrantPermission(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "当前系统基于等级权限，请使用更新用户等级接口",
	})
}

// RevokePermission 撤销权限（预留接口，当前基于等级系统）
func (s *Service) RevokePermission(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "当前系统基于等级权限，请使用更新用户等级接口",
	})
}

// GetUserLevel 获取用户权限等级
func (s *Service) GetUserLevel(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var level int
	var username string
	err = s.db.QueryRow(`
		SELECT username, permission_level 
		FROM users WHERE id = $1`, userID).Scan(&username, &level)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
		"level":    level,
	})
}

// UpdateUserLevel 更新用户权限等级
func (s *Service) UpdateUserLevel(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req UpdateUserLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "details": err.Error()})
		return
	}

	// 检查用户是否存在
	var currentLevel int
	var username string
	err = s.db.QueryRow("SELECT username, permission_level FROM users WHERE id = $1", userID).Scan(&username, &currentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		}
		return
	}

	// 更新用户权限等级
	_, err = s.db.Exec(`
		UPDATE users 
		SET permission_level = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`, req.Level, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户等级失败"})
		return
	}

	// 记录审计日志
	s.logAuditEvent(req.UpdatedBy, "UPDATE_USER_LEVEL",
		fmt.Sprintf("用户 %s (ID: %d) 权限等级从 %d 更新为 %d", username, userID, currentLevel, req.Level))

	c.JSON(http.StatusOK, gin.H{
		"message":      "用户权限等级更新成功",
		"user_id":      userID,
		"username":     username,
		"old_level":    currentLevel,
		"new_level":    req.Level,
	})
}

// GetPermissionAudit 获取权限审计日志
func (s *Service) GetPermissionAudit(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	rows, err := s.db.Query(`
		SELECT al.id, al.user_id, u.username, al.action, al.details, al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.action LIKE '%PERMISSION%' OR al.action LIKE '%LEVEL%'
		ORDER BY al.created_at DESC
		LIMIT $1`, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询审计日志失败"})
		return
	}
	defer rows.Close()

	type AuditLog struct {
		ID        int       `json:"id"`
		UserID    *int      `json:"user_id"`
		Username  *string   `json:"username"`
		Action    string    `json:"action"`
		Details   string    `json:"details"`
		CreatedAt time.Time `json:"created_at"`
	}

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Username, &log.Action, &log.Details, &log.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描审计日志失败"})
			return
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
}

// GetUserAudit 获取用户审计日志
func (s *Service) GetUserAudit(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	rows, err := s.db.Query(`
		SELECT id, action, details, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户审计日志失败"})
		return
	}
	defer rows.Close()

	type UserAuditLog struct {
		ID        int       `json:"id"`
		Action    string    `json:"action"`
		Details   string    `json:"details"`
		CreatedAt time.Time `json:"created_at"`
	}

	var logs []UserAuditLog
	for rows.Next() {
		var log UserAuditLog
		if err := rows.Scan(&log.ID, &log.Action, &log.Details, &log.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描审计日志失败"})
			return
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"audit_logs": logs,
	})
}

// checkUserPermission 检查用户是否具有指定权限
func (s *Service) checkUserPermission(userID int, permissionName string) (bool, string, int) {
	// 获取用户权限等级
	var userLevel int
	var isActive bool
	err := s.db.QueryRow(`
		SELECT permission_level, is_active 
		FROM users WHERE id = $1`, userID).Scan(&userLevel, &isActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, "用户不存在", 0
		}
		return false, "数据库查询失败", 0
	}

	if !isActive {
		return false, "用户已被禁用", userLevel
	}

	// 获取权限要求的等级
	var requiredLevel int
	err = s.db.QueryRow(`
		SELECT level FROM permissions WHERE name = $1`, permissionName).Scan(&requiredLevel)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, "权限不存在", userLevel
		}
		return false, "权限查询失败", userLevel
	}

	// 检查用户等级是否满足权限要求
	if userLevel >= requiredLevel {
		return true, "权限验证通过", userLevel
	}

	return false, fmt.Sprintf("权限等级不足，需要等级 %d，当前等级 %d", requiredLevel, userLevel), userLevel
}

// logPermissionCheck 记录权限检查审计
func (s *Service) logPermissionCheck(userID int, permissionName string, allowed bool, reason string) {
	action := "PERMISSION_CHECK"
	details := fmt.Sprintf("权限检查: %s, 结果: %t, 原因: %s", permissionName, allowed, reason)
	s.logAuditEvent(userID, action, details)
}

// logAuditEvent 记录审计事件
func (s *Service) logAuditEvent(userID int, action, details string) {
	_, err := s.db.Exec(`
		INSERT INTO audit_logs (user_id, action, details) 
		VALUES ($1, $2, $3)`, userID, action, details)
	if err != nil {
		// 记录日志但不影响主要流程
		fmt.Printf("审计日志记录失败: %v\n", err)
	}
}