package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS middleware to handle Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		c.Abort()
	})
}

// Auth middleware for authentication and authorization
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for health check and auth endpoints
		if c.Request.URL.Path == "/health" || 
		   c.Request.URL.Path == "/api/v1/auth/login" || 
		   c.Request.URL.Path == "/api/v1/auth/refresh" {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the token format is correct (Bearer token)
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token
		token := authHeader[7:]

		// Validate token (this is a placeholder implementation)
		// In a real implementation, you would validate the JWT token here
		if !validateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract user information from token and set in context
		userID, err := extractUserIDFromToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("userID", userID)
		c.Next()
	}
}

// validateToken validates the JWT token (placeholder implementation)
func validateToken(token string) bool {
	// This is a placeholder implementation
	// In a real implementation, you would validate the JWT token here
	// For now, we'll just check if the token is not empty
	return token != ""
}

// extractUserIDFromToken extracts user ID from JWT token (placeholder implementation)
func extractUserIDFromToken(token string) (string, error) {
	// This is a placeholder implementation
	// In a real implementation, you would extract the user ID from the JWT token
	// For now, we'll just return a fixed user ID
	return "user123", nil
}

// Login handles user login and returns a JWT token
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate credentials (placeholder implementation)
	// In a real implementation, you would validate the credentials against a database
	if req.Username != "admin" || req.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Generate JWT token (placeholder implementation)
	// In a real implementation, you would generate a proper JWT token
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"type":  "Bearer",
	})
}

// RefreshToken handles token refresh
func RefreshToken(c *gin.Context) {
	// Get Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header is required",
		})
		return
	}

	// Check if the token format is correct (Bearer token)
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid authorization header format",
		})
		return
	}

	// Extract token
	token := authHeader[7:]

	// Validate token (placeholder implementation)
	if !validateToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired token",
		})
		return
	}

	// Generate new JWT token (placeholder implementation)
	// In a real implementation, you would generate a proper JWT token
	newToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
		"type":  "Bearer",
	})
}