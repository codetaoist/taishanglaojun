// 多租户工具包
package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// GenerateTenantID 生成租户ID
func GenerateTenantID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "tenant_" + hex.EncodeToString(bytes)
}

// ValidateTenantDomain 验证租户域名
func ValidateTenantDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}
	
	if len(domain) < 3 {
		return fmt.Errorf("域名长度不能少于3个字符")
	}
	
	if strings.Contains(domain, " ") {
		return fmt.Errorf("域名不能包含空格")
	}
	
	return nil
}

// NormalizeTenantName 标准化租户名称
func NormalizeTenantName(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}

// CacheService 缓存服务接口
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// Logger 日志记录接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
	WithContext(ctx context.Context) Logger
}