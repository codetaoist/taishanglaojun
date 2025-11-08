package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codetaoist/taishanglaojun/auth/internal/model"
	"github.com/codetaoist/taishanglaojun/auth/internal/service"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles login requests
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "AUTH_FAILED",
			"message": "Authentication failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Login successful",
		"data":    resp,
	})
}

// Register handles registration requests
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.authService.Register(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "REGISTRATION_FAILED",
			"message": "Registration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "SUCCESS",
		"message": "Registration successful",
	})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.RefreshToken(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "TOKEN_REFRESH_FAILED",
			"message": "Token refresh failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Token refreshed successfully",
		"data":    resp,
	})
}

// Logout handles logout requests
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "MISSING_TOKEN",
			"message": "Authorization header is missing",
		})
		return
	}

	// Extract token from "Bearer <token>"
	token := authHeader[7:] // Remove "Bearer " prefix
	if len(token) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_TOKEN",
			"message": "Invalid token format",
		})
		return
	}

	if err := h.authService.Logout(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "LOGOUT_FAILED",
			"message": "Logout failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Logout successful",
	})
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Unauthorized",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.authService.ChangePassword(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "PASSWORD_CHANGE_FAILED",
			"message": "Password change failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Password changed successfully",
	})
}

// GetProfile handles user profile requests
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Unauthorized",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
		return
	}

	user, err := h.authService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "USER_NOT_FOUND",
			"message": "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "User profile retrieved successfully",
		"data":    user,
	})
}

// GetUser handles user retrieval requests (admin only)
func (h *AuthHandler) GetUser(c *gin.Context) {
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_USER_ID",
			"message": "Invalid user ID",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "USER_NOT_FOUND",
			"message": "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "User retrieved successfully",
		"data":    user,
	})
}

// RevokeToken handles token revocation requests
func (h *AuthHandler) RevokeToken(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "MISSING_TOKEN",
			"message": "Authorization header is missing",
		})
		return
	}

	// Extract token from "Bearer <token>"
	token := authHeader[7:] // Remove "Bearer " prefix
	if len(token) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_TOKEN",
			"message": "Invalid token format",
		})
		return
	}

	// Get reason from request body
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body is provided, use default reason
		req.Reason = "revoked"
	}

	if err := h.authService.RevokeToken(token, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "TOKEN_REVOCATION_FAILED",
			"message": "Token revocation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Token revoked successfully",
	})
}