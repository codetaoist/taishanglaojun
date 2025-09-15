package auth

import (
	"fmt"
	"time"
)

// TokenManager 管理认证令牌
type TokenManager struct {
	token     string
	expiresAt time.Time
}

// NewTokenManager 创建新的令牌管理器
func NewTokenManager() *TokenManager {
	return &TokenManager{}
}

// SetToken 设置认证令牌
func (tm *TokenManager) SetToken(token string, expiresAt time.Time) {
	tm.token = token
	tm.expiresAt = expiresAt
}

// GetToken 获取当前有效的令牌
func (tm *TokenManager) GetToken() (string, error) {
	if tm.token == "" {
		return "", fmt.Errorf("未找到认证令牌")
	}
	
	if time.Now().After(tm.expiresAt) {
		return "", fmt.Errorf("认证令牌已过期")
	}
	
	return tm.token, nil
}

// IsAuthenticated 检查是否已认证
func (tm *TokenManager) IsAuthenticated() bool {
	_, err := tm.GetToken()
	return err == nil
}

// ClearToken 清除令牌
func (tm *TokenManager) ClearToken() {
	tm.token = ""
	tm.expiresAt = time.Time{}
}