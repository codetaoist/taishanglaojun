package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/taishanglaojun/auth_system/internal/middleware"
	"github.com/taishanglaojun/auth_system/internal/models"
	"github.com/taishanglaojun/auth_system/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "注册请求"
// @Success 201 {object} models.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := h.validateRegisterRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to register user")
		return
	}

	h.logger.Info("User registered successfully", 
		zap.String("username", req.Username),
		zap.String("email", req.Email),
	)

	c.JSON(http.StatusCreated, resp)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户身份验证并获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录请求"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// 设置客户端信息
	req.UserAgent = c.GetHeader("User-Agent")
	req.IPAddress = c.ClientIP()

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to login")
		return
	}

	h.logger.Info("User logged in successfully", 
		zap.String("username", req.Username),
		zap.String("user_id", resp.User.ID.String()),
	)

	c.JSON(http.StatusOK, resp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 撤销用户会话
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.LogoutRequest true "登出请求"
// @Success 200 {object} models.LogoutResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从上下文获取会话ID
		if sessionID, exists := c.Get("session_id"); exists {
			if sid, ok := sessionID.(uuid.UUID); ok {
				req.SessionID = sid
			}
		}
		
		if req.SessionID == uuid.Nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid request format or missing session ID",
			})
			return
		}
	}

	// 获取访问令牌用于黑名单（注意：LogoutRequest结构体中没有AccessToken字段）
	// 如果需要处理访问令牌，应该在服务层通过其他方式处理
	authHeader := c.GetHeader("Authorization")
	_ = authHeader // 暂时忽略，避免未使用变量警告

	resp, err := h.authService.Logout(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to logout")
		return
	}

	h.logger.Info("User logged out successfully", 
		zap.String("session_id", req.SessionID.String()),
	)

	c.JSON(http.StatusOK, resp)
}

// RefreshToken 刷新令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to refresh token")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ValidateToken 验证令牌
// @Summary 验证访问令牌
// @Description 验证令牌的有效性并返回用户信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.ValidateTokenRequest true "验证令牌请求"
// @Success 200 {object} models.ValidateTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req models.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.authService.ValidateToken(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to validate token")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), uid, &req); err != nil {
		h.handleServiceError(c, err, "Failed to change password")
		return
	}

	h.logger.Info("Password changed successfully", zap.String("user_id", uid.String()))

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Password changed successfully",
	})
}

// ForgotPassword 忘记密码
// @Summary 忘记密码
// @Description 发送密码重置链接
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "忘记密码请求"
// @Success 200 {object} models.ForgotPasswordResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.authService.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to process forgot password request")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用重置令牌重置密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} models.ResetPasswordResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.authService.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to reset password")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyEmail 验证邮箱
// @Summary 验证邮箱
// @Description 使用验证令牌验证邮箱地址
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.VerifyEmailRequest true "验证邮箱请求"
// @Success 200 {object} models.VerifyEmailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.authService.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to verify email")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ResendVerification 重新发送验证邮件
// @Summary 重新发送验证邮件
// @Description 重新发送邮箱验证链接
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.ResendVerificationRequest true "重新发送验证请求"
// @Success 200 {object} models.ResendVerificationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.authService.ResendVerification(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to resend verification")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前用户的资料信息
// @Tags 用户
// @Produce json
// @Success 200 {object} models.PublicUser
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	profile, err := h.authService.GetUserProfile(c.Request.Context(), uid)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user profile")
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body models.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} models.PublicUser
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	profile, err := h.authService.UpdateUserProfile(c.Request.Context(), uid, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update user profile")
		return
	}

	h.logger.Info("User profile updated successfully", zap.String("user_id", uid.String()))

	c.JSON(http.StatusOK, profile)
}

// GetSessions 获取用户会话
// @Summary 获取用户会话
// @Description 获取当前用户的所有活跃会话
// @Tags 会话
// @Produce json
// @Success 200 {array} models.Session
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	sessions, err := h.authService.GetUserSessions(c.Request.Context(), uid)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user sessions")
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// RevokeSession 撤销会话
// @Summary 撤销会话
// @Description 撤销指定的用户会话
// @Tags 会话
// @Param session_id path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/sessions/{session_id} [delete]
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid session ID format",
		})
		return
	}

	if err := h.authService.RevokeSession(c.Request.Context(), sessionID); err != nil {
		h.handleServiceError(c, err, "Failed to revoke session")
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Session revoked successfully",
	})
}

// RevokeAllSessions 撤销所有会话
// @Summary 撤销所有会话
// @Description 撤销当前用户的所有会话
// @Tags 会话
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/sessions [delete]
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	if err := h.authService.RevokeAllSessions(c.Request.Context(), uid); err != nil {
		h.handleServiceError(c, err, "Failed to revoke all sessions")
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "All sessions revoked successfully",
	})
}

// Me 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前认证用户的基本信息
// @Tags 用户
// @Produce json
// @Success 200 {object} middleware.CurrentUser
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// validateRegisterRequest 验证注册请求
func (h *AuthHandler) validateRegisterRequest(req *models.RegisterRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

// handleServiceError 处理服务错误
func (h *AuthHandler) handleServiceError(c *gin.Context, err error, logMessage string) {
	h.logger.Error(logMessage, zap.Error(err))

	switch err {
	case service.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	case service.ErrUserExists:
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "user_exists",
			Message: "User already exists",
		})
	case service.ErrUserInactive:
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "user_inactive",
			Message: "User account is inactive",
		})
	case service.ErrUserSuspended:
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "user_suspended",
			Message: "User account is suspended",
		})
	case service.ErrInvalidToken:
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid or expired token",
		})
	case service.ErrTokenExpired:
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "token_expired",
			Message: "Token has expired",
		})
	case service.ErrSessionNotFound:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "session_not_found",
			Message: "Session not found",
		})
	case service.ErrSessionExpired:
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "session_expired",
			Message: "Session has expired",
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "An internal error occurred",
		})
	}
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}