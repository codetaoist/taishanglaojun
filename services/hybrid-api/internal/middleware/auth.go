package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// ErrorCode and helpers copied minimal to avoid import cycles; keep consistent with router.
type ErrorCode string

const (
	CodeOK              ErrorCode = "OK"
	CodeUnauthenticated ErrorCode = "UNAUTHENTICATED"
)

func errorJSON(c *gin.Context, httpStatus int, code ErrorCode, msg string) {
	traceID := ""
	if v, exists := c.Get("traceID"); exists {
		if s, ok := v.(string); ok {
			traceID = s
		}
	}
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": msg,
		"traceId": traceID,
	})
}

// UserInfo represents user information extracted from the auth service
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// AuthResponse represents the response from the auth service
type AuthResponse struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Data    UserInfo  `json:"data"`
}

// Auth returns a middleware that enforces JWT on protected routes.
// It validates tokens against the auth service for better security.
func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Health 或公开路由可在路由层选择不使用该中间件
		fmt.Printf("DevSkipSignature: %v\n", cfg.DevSkipSignature)
		if cfg.DevSkipSignature {
			c.Set("actor", "dev")
			c.Set("user_id", "dev")
			c.Set("username", "dev")
			c.Set("role", "dev")
			c.Next()
			return
		}
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			errorJSON(c, http.StatusUnauthorized, CodeUnauthenticated, "missing bearer token")
			c.Abort()
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if tokenStr == "" {
			errorJSON(c, http.StatusUnauthorized, CodeUnauthenticated, "empty token")
			c.Abort()
			return
		}

		// Try to validate with auth service first
		userInfo, err := validateWithAuthService(tokenStr, cfg)
		if err != nil {
			// If auth service is not available, fall back to local validation
			if strings.Contains(err.Error(), "connection refused") || 
			   strings.Contains(err.Error(), "no such host") ||
			   strings.Contains(err.Error(), "timeout") {
				userInfo, err = validateLocally(tokenStr, cfg)
				if err != nil {
					errorJSON(c, http.StatusUnauthorized, CodeUnauthenticated, fmt.Sprintf("invalid token: %v", err))
					c.Abort()
					return
				}
			} else {
				errorJSON(c, http.StatusUnauthorized, CodeUnauthenticated, fmt.Sprintf("token validation failed: %v", err))
				c.Abort()
				return
			}
		}

		// Set user information in context
		c.Set("actor", fmt.Sprintf("%d", userInfo.ID))
		c.Set("user_id", userInfo.ID)
		c.Set("username", userInfo.Username)
		c.Set("role", userInfo.Role)
		c.Next()
	}
}

// validateWithAuthService validates the token with the auth service
func validateWithAuthService(token string, cfg config.Config) (*UserInfo, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create request to auth service
	req, err := http.NewRequest("GET", cfg.AuthServiceURL+"/api/v1/profile", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	// Parse response
	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Check if auth was successful
	if authResp.Code != "SUCCESS" {
		return nil, fmt.Errorf("auth service error: %s", authResp.Message)
	}

	return &authResp.Data, nil
}

// validateLocally validates the token locally (fallback method)
func validateLocally(tokenStr string, cfg config.Config) (*UserInfo, error) {
	// HS256 默认；如提供 RS256 公钥则选择 RS256 验签
	var parsed *jwt.Token
	var err error
	if cfg.JWTPublicKeyPEM != "" {
		key, e := jwt.ParseRSAPublicKeyFromPEM([]byte(cfg.JWTPublicKeyPEM))
		if e != nil {
			return nil, fmt.Errorf("invalid jwt public key: %w", e)
		}
		parsed, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return key, nil
		})
	} else {
		if cfg.JWTSecret == "" {
			return nil, fmt.Errorf("jwt secret not configured")
		}
		parsed, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
	}
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	
	// Extract user ID from sub claim
	userID := 0
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		fmt.Sscanf(sub, "%d", &userID)
	}
	
	// Create user info from claims
	userInfo := &UserInfo{
		ID: userID,
	}
	
	// Extract username and role if available
	if username, ok := claims["username"].(string); ok {
		userInfo.Username = username
	}
	if role, ok := claims["role"].(string); ok {
		userInfo.Role = role
	}
	
	return userInfo, nil
}