package vector

import (
	"time"
)

// MilvusConfig 表示Milvus客户端配置
type MilvusConfig struct {
	// Milvus服务器地址
	Address string `json:"address" yaml:"address"`
	
	// Milvus服务器端口
	Port int `json:"port" yaml:"port"`
	
	// 用户名
	Username string `json:"username" yaml:"username"`
	
	// 密码
	Password string `json:"password" yaml:"password"`
	
	// 数据库名称
	Database string `json:"database" yaml:"database"`
	
	// 连接超时时间
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	
	// 请求超时时间
	RequestTimeout time.Duration `json:"request_timeout" yaml:"request_timeout"`
	
	// 最大重试次数
	MaxRetry int `json:"max_retry" yaml:"max_retry"`
	
	// 是否启用TLS
	EnableTLS bool `json:"enable_tls" yaml:"enable_tls"`
	
	// TLS证书路径
	TLSCertPath string `json:"tls_cert_path" yaml:"tls_cert_path"`
	
	// TLS私钥路径
	TLSKeyPath string `json:"tls_key_path" yaml:"tls_key_path"`
	
	// TLS CA证书路径
	TLSCAPath string `json:"tls_ca_path" yaml:"tls_ca_path"`
	
	// 是否跳过TLS验证
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// DefaultMilvusConfig 返回默认的Milvus配置
func DefaultMilvusConfig() *MilvusConfig {
	return &MilvusConfig{
		Address:            "localhost",
		Port:               19530,
		Username:           "",
		Password:           "",
		Database:           "",
		ConnectTimeout:     10 * time.Second,
		RequestTimeout:     30 * time.Second,
		MaxRetry:           3,
		EnableTLS:          false,
		TLSCertPath:        "",
		TLSKeyPath:         "",
		TLSCAPath:          "",
		InsecureSkipVerify: false,
	}
}

// Validate 验证配置的有效性
func (c *MilvusConfig) Validate() error {
	if c.Address == "" {
		return ErrInvalidConfig("address cannot be empty")
	}
	
	if c.Port <= 0 || c.Port > 65535 {
		return ErrInvalidConfig("port must be between 1 and 65535")
	}
	
	if c.ConnectTimeout <= 0 {
		return ErrInvalidConfig("connect_timeout must be positive")
	}
	
	if c.RequestTimeout <= 0 {
		return ErrInvalidConfig("request_timeout must be positive")
	}
	
	if c.MaxRetry < 0 {
		return ErrInvalidConfig("max_retry cannot be negative")
	}
	
	return nil
}