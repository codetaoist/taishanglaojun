package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// SecurityConfig 安全模块配置
type SecurityConfig struct {
	Server           ServerConfig           `yaml:"server" json:"server"`
	Environment      string                 `yaml:"environment" json:"environment"`
	Database         DatabaseConfig         `yaml:"database" json:"database"`
	ThreatDetection  ThreatDetectionConfig  `yaml:"threat_detection" json:"threat_detection"`
	Vulnerability    VulnerabilityConfig    `yaml:"vulnerability" json:"vulnerability"`
	Pentest          PentestConfig          `yaml:"pentest" json:"pentest"`
	SecurityEducation SecurityEducationConfig `yaml:"security_education" json:"security_education"`
	SecurityAudit    SecurityAuditConfig    `yaml:"security_audit" json:"security_audit"`
	Middleware       MiddlewareConfig       `yaml:"middleware" json:"middleware"`
	Encryption       EncryptionConfig       `yaml:"encryption" json:"encryption"`
	Logging          LoggingConfig          `yaml:"logging" json:"logging"`
	Monitoring       MonitoringConfig       `yaml:"monitoring" json:"monitoring"`
}

// ServerConfig 服务器配�?
type ServerConfig struct {
	Port         int `yaml:"port" json:"port"`
	ReadTimeout  int `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  int `yaml:"idle_timeout" json:"idle_timeout"`
}

// DatabaseConfig 数据库配�?
type DatabaseConfig struct {
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
	Database     string `yaml:"database" json:"database"`
	SSLMode      string `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxLifetime  string `yaml:"max_lifetime" json:"max_lifetime"`
}

// ThreatDetectionConfig 威胁检测配�?
type ThreatDetectionConfig struct {
	Enabled              bool          `yaml:"enabled" json:"enabled"`
	ScanInterval         time.Duration `yaml:"scan_interval" json:"scan_interval"`
	MaxConcurrentScans   int           `yaml:"max_concurrent_scans" json:"max_concurrent_scans"`
	AlertThreshold       int           `yaml:"alert_threshold" json:"alert_threshold"`
	AutoBlock            bool          `yaml:"auto_block" json:"auto_block"`
	BlockDuration        time.Duration `yaml:"block_duration" json:"block_duration"`
	SQLInjection         DetectionRule `yaml:"sql_injection" json:"sql_injection"`
	XSSAttack            DetectionRule `yaml:"xss_attack" json:"xss_attack"`
	BruteForce           DetectionRule `yaml:"brute_force" json:"brute_force"`
	DDoSAttack           DetectionRule `yaml:"ddos_attack" json:"ddos_attack"`
	PathTraversal        DetectionRule `yaml:"path_traversal" json:"path_traversal"`
	CommandInjection     DetectionRule `yaml:"command_injection" json:"command_injection"`
	NotificationChannels []string      `yaml:"notification_channels" json:"notification_channels"`
}

// DetectionRule 检测规则配�?
type DetectionRule struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Severity    string   `yaml:"severity" json:"severity"`
	Patterns    []string `yaml:"patterns" json:"patterns"`
	Threshold   int      `yaml:"threshold" json:"threshold"`
	TimeWindow  string   `yaml:"time_window" json:"time_window"`
	Action      string   `yaml:"action" json:"action"`
	Description string   `yaml:"description" json:"description"`
}

// VulnerabilityConfig 漏洞扫描配置
type VulnerabilityConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	ScanInterval       time.Duration `yaml:"scan_interval" json:"scan_interval"`
	MaxConcurrentScans int           `yaml:"max_concurrent_scans" json:"max_concurrent_scans"`
	ScanTimeout        time.Duration `yaml:"scan_timeout" json:"scan_timeout"`
	WebScan            WebScanConfig `yaml:"web_scan" json:"web_scan"`
	NetworkScan        NetworkScanConfig `yaml:"network_scan" json:"network_scan"`
	ReportRetention    time.Duration `yaml:"report_retention" json:"report_retention"`
}

// WebScanConfig Web扫描配置
type WebScanConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	MaxDepth        int      `yaml:"max_depth" json:"max_depth"`
	MaxPages        int      `yaml:"max_pages" json:"max_pages"`
	RequestTimeout  string   `yaml:"request_timeout" json:"request_timeout"`
	UserAgent       string   `yaml:"user_agent" json:"user_agent"`
	FollowRedirects bool     `yaml:"follow_redirects" json:"follow_redirects"`
	CheckSSL        bool     `yaml:"check_ssl" json:"check_ssl"`
	CheckHeaders    bool     `yaml:"check_headers" json:"check_headers"`
	ExcludePaths    []string `yaml:"exclude_paths" json:"exclude_paths"`
}

// NetworkScanConfig 网络扫描配置
type NetworkScanConfig struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	PortRange      string   `yaml:"port_range" json:"port_range"`
	ScanTimeout    string   `yaml:"scan_timeout" json:"scan_timeout"`
	MaxConcurrent  int      `yaml:"max_concurrent" json:"max_concurrent"`
	ServiceDetection bool   `yaml:"service_detection" json:"service_detection"`
	OSDetection    bool     `yaml:"os_detection" json:"os_detection"`
	ExcludeHosts   []string `yaml:"exclude_hosts" json:"exclude_hosts"`
}

// PentestConfig 渗透测试配�?
type PentestConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	MaxConcurrentJobs int           `yaml:"max_concurrent_jobs" json:"max_concurrent_jobs"`
	JobTimeout        time.Duration `yaml:"job_timeout" json:"job_timeout"`
	AllowedTargets    []string      `yaml:"allowed_targets" json:"allowed_targets"`
	ForbiddenTargets  []string      `yaml:"forbidden_targets" json:"forbidden_targets"`
	Tools             ToolsConfig   `yaml:"tools" json:"tools"`
	ReportRetention   time.Duration `yaml:"report_retention" json:"report_retention"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	Nmap    NmapConfig    `yaml:"nmap" json:"nmap"`
	DirBuster DirBusterConfig `yaml:"dirbuster" json:"dirbuster"`
	SQLMap  SQLMapConfig  `yaml:"sqlmap" json:"sqlmap"`
}

// NmapConfig Nmap配置
type NmapConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	BinaryPath string `yaml:"binary_path" json:"binary_path"`
	Options    string `yaml:"options" json:"options"`
	Timeout    string `yaml:"timeout" json:"timeout"`
}

// DirBusterConfig DirBuster配置
type DirBusterConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Wordlists   []string `yaml:"wordlists" json:"wordlists"`
	Extensions  []string `yaml:"extensions" json:"extensions"`
	Threads     int      `yaml:"threads" json:"threads"`
	Timeout     string   `yaml:"timeout" json:"timeout"`
}

// SQLMapConfig SQLMap配置
type SQLMapConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	BinaryPath string `yaml:"binary_path" json:"binary_path"`
	Options    string `yaml:"options" json:"options"`
	Timeout    string `yaml:"timeout" json:"timeout"`
}

// SecurityEducationConfig 安全教育配置
type SecurityEducationConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	LabTimeout      time.Duration `yaml:"lab_timeout" json:"lab_timeout"`
	MaxConcurrentLabs int         `yaml:"max_concurrent_labs" json:"max_concurrent_labs"`
	CourseRetention time.Duration `yaml:"course_retention" json:"course_retention"`
	CertificationConfig CertificationConfig `yaml:"certification" json:"certification"`
}

// CertificationConfig 认证配置
type CertificationConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	ValidityPeriod string `yaml:"validity_period" json:"validity_period"`
	IssuerName     string `yaml:"issuer_name" json:"issuer_name"`
	SigningKey     string `yaml:"signing_key" json:"signing_key"`
}

// SecurityAuditConfig 安全审计配置
type SecurityAuditConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	LogRetention      time.Duration `yaml:"log_retention" json:"log_retention"`
	EventRetention    time.Duration `yaml:"event_retention" json:"event_retention"`
	ComplianceCheck   ComplianceConfig `yaml:"compliance_check" json:"compliance_check"`
	AlertThreshold    int           `yaml:"alert_threshold" json:"alert_threshold"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// ComplianceConfig 合规配置
type ComplianceConfig struct {
	Enabled    bool     `yaml:"enabled" json:"enabled"`
	Standards  []string `yaml:"standards" json:"standards"`
	CheckInterval string `yaml:"check_interval" json:"check_interval"`
	AutoRemediation bool `yaml:"auto_remediation" json:"auto_remediation"`
}

// MiddlewareConfig 中间件配�?
type MiddlewareConfig struct {
	RateLimit    RateLimitConfig    `yaml:"rate_limit" json:"rate_limit"`
	CORS         CORSConfig         `yaml:"cors" json:"cors"`
	Security     SecurityHeadersConfig `yaml:"security" json:"security"`
	IPFilter     IPFilterConfig     `yaml:"ip_filter" json:"ip_filter"`
	InputValidation InputValidationConfig `yaml:"input_validation" json:"input_validation"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	MaxRequests  int    `yaml:"max_requests" json:"max_requests"`
	TimeWindow   string `yaml:"time_window" json:"time_window"`
	BlockDuration string `yaml:"block_duration" json:"block_duration"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers" json:"allowed_headers"`
	AllowCredentials bool   `yaml:"allow_credentials" json:"allow_credentials"`
}

// SecurityHeadersConfig 安全头配�?
type SecurityHeadersConfig struct {
	Enabled                bool   `yaml:"enabled" json:"enabled"`
	ContentTypeOptions     string `yaml:"content_type_options" json:"content_type_options"`
	FrameOptions           string `yaml:"frame_options" json:"frame_options"`
	XSSProtection          string `yaml:"xss_protection" json:"xss_protection"`
	StrictTransportSecurity string `yaml:"strict_transport_security" json:"strict_transport_security"`
	ContentSecurityPolicy  string `yaml:"content_security_policy" json:"content_security_policy"`
	ReferrerPolicy         string `yaml:"referrer_policy" json:"referrer_policy"`
	PermissionsPolicy      string `yaml:"permissions_policy" json:"permissions_policy"`
}

// IPFilterConfig IP过滤配置
type IPFilterConfig struct {
	Enabled    bool     `yaml:"enabled" json:"enabled"`
	Whitelist  []string `yaml:"whitelist" json:"whitelist"`
	Blacklist  []string `yaml:"blacklist" json:"blacklist"`
	Mode       string   `yaml:"mode" json:"mode"` // whitelist, blacklist, or disabled
}

// InputValidationConfig 输入验证配置
type InputValidationConfig struct {
	Enabled           bool     `yaml:"enabled" json:"enabled"`
	MaxInputLength    int      `yaml:"max_input_length" json:"max_input_length"`
	BlockMaliciousInput bool   `yaml:"block_malicious_input" json:"block_malicious_input"`
	SanitizeInput     bool     `yaml:"sanitize_input" json:"sanitize_input"`
	MaliciousPatterns []string `yaml:"malicious_patterns" json:"malicious_patterns"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Algorithm    string `yaml:"algorithm" json:"algorithm"`
	KeySize      int    `yaml:"key_size" json:"key_size"`
	DefaultKey   string `yaml:"default_key" json:"default_key"`
	KeyRotation  bool   `yaml:"key_rotation" json:"key_rotation"`
	RotationInterval string `yaml:"rotation_interval" json:"rotation_interval"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	MetricsPort int   `yaml:"metrics_port" json:"metrics_port"`
	MetricsPath string `yaml:"metrics_path" json:"metrics_path"`
	HealthCheck HealthCheckConfig `yaml:"health_check" json:"health_check"`
}

// HealthCheckConfig 健康检查配�?
type HealthCheckConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Port     int    `yaml:"port" json:"port"`
	Path     string `yaml:"path" json:"path"`
	Interval string `yaml:"interval" json:"interval"`
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *SecurityConfig {
	return &SecurityConfig{
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  120,
		},
		Environment: "development",
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			Username:     "security_user",
			Password:     "security_password",
			Database:     "security_db",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  "5m",
		},
		ThreatDetection: ThreatDetectionConfig{
			Enabled:            true,
			ScanInterval:       time.Minute * 5,
			MaxConcurrentScans: 10,
			AlertThreshold:     5,
			AutoBlock:          true,
			BlockDuration:      time.Hour * 24,
			SQLInjection: DetectionRule{
				Enabled:     true,
				Severity:    "high",
				Patterns:    []string{"union select", "drop table", "insert into"},
				Threshold:   3,
				TimeWindow:  "5m",
				Action:      "block",
				Description: "SQL注入攻击检�?,
			},
			XSSAttack: DetectionRule{
				Enabled:     true,
				Severity:    "medium",
				Patterns:    []string{"<script", "javascript:", "onload="},
				Threshold:   3,
				TimeWindow:  "5m",
				Action:      "alert",
				Description: "XSS攻击检�?,
			},
			BruteForce: DetectionRule{
				Enabled:     true,
				Severity:    "high",
				Patterns:    []string{},
				Threshold:   10,
				TimeWindow:  "1m",
				Action:      "block",
				Description: "暴力破解攻击检�?,
			},
			DDoSAttack: DetectionRule{
				Enabled:     true,
				Severity:    "critical",
				Patterns:    []string{},
				Threshold:   100,
				TimeWindow:  "1m",
				Action:      "block",
				Description: "DDoS攻击检�?,
			},
			NotificationChannels: []string{"email", "webhook"},
		},
		Vulnerability: VulnerabilityConfig{
			Enabled:            true,
			ScanInterval:       time.Hour * 24,
			MaxConcurrentScans: 5,
			ScanTimeout:        time.Hour * 2,
			WebScan: WebScanConfig{
				Enabled:         true,
				MaxDepth:        3,
				MaxPages:        1000,
				RequestTimeout:  "30s",
				UserAgent:       "SecurityScanner/1.0",
				FollowRedirects: true,
				CheckSSL:        true,
				CheckHeaders:    true,
				ExcludePaths:    []string{"/admin", "/private"},
			},
			NetworkScan: NetworkScanConfig{
				Enabled:          true,
				PortRange:        "1-65535",
				ScanTimeout:      "5m",
				MaxConcurrent:    100,
				ServiceDetection: true,
				OSDetection:      false,
				ExcludeHosts:     []string{"127.0.0.1", "localhost"},
			},
			ReportRetention: time.Hour * 24 * 30,
		},
		Pentest: PentestConfig{
			Enabled:           true,
			MaxConcurrentJobs: 3,
			JobTimeout:        time.Hour * 8,
			AllowedTargets:    []string{"192.168.1.0/24", "10.0.0.0/8"},
			ForbiddenTargets:  []string{"127.0.0.1", "localhost"},
			Tools: ToolsConfig{
				Nmap: NmapConfig{
					Enabled:    true,
					BinaryPath: "/usr/bin/nmap",
					Options:    "-sS -O",
					Timeout:    "10m",
				},
				DirBuster: DirBusterConfig{
					Enabled:    true,
					Wordlists:  []string{"/usr/share/wordlists/dirb/common.txt"},
					Extensions: []string{".php", ".html", ".js", ".css"},
					Threads:    10,
					Timeout:    "5m",
				},
				SQLMap: SQLMapConfig{
					Enabled:    true,
					BinaryPath: "/usr/bin/sqlmap",
					Options:    "--batch --random-agent",
					Timeout:    "30m",
				},
			},
			ReportRetention: time.Hour * 24 * 90,
		},
		SecurityEducation: SecurityEducationConfig{
			Enabled:           true,
			LabTimeout:        time.Hour * 4,
			MaxConcurrentLabs: 50,
			CourseRetention:   time.Hour * 24 * 365,
			CertificationConfig: CertificationConfig{
				Enabled:        true,
				ValidityPeriod: "1y",
				IssuerName:     "太上老君安全学院",
				SigningKey:     "default_signing_key",
			},
		},
		SecurityAudit: SecurityAuditConfig{
			Enabled:        true,
			LogRetention:   time.Hour * 24 * 90,
			EventRetention: time.Hour * 24 * 30,
			ComplianceCheck: ComplianceConfig{
				Enabled:         true,
				Standards:       []string{"DengBao2.0", "ISO27001", "NIST"},
				CheckInterval:   "24h",
				AutoRemediation: false,
			},
			AlertThreshold:  10,
			CleanupInterval: time.Hour * 24,
		},
		Middleware: MiddlewareConfig{
			RateLimit: RateLimitConfig{
				Enabled:       true,
				MaxRequests:   100,
				TimeWindow:    "1m",
				BlockDuration: "5m",
			},
			CORS: CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
				AllowCredentials: true,
			},
			Security: SecurityHeadersConfig{
				Enabled:                 true,
				ContentTypeOptions:      "nosniff",
				FrameOptions:            "DENY",
				XSSProtection:           "1; mode=block",
				StrictTransportSecurity: "max-age=31536000; includeSubDomains",
				ContentSecurityPolicy:   "default-src 'self'",
				ReferrerPolicy:          "strict-origin-when-cross-origin",
				PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
			},
			IPFilter: IPFilterConfig{
				Enabled:   false,
				Whitelist: []string{},
				Blacklist: []string{},
				Mode:      "disabled",
			},
			InputValidation: InputValidationConfig{
				Enabled:             true,
				MaxInputLength:      10000,
				BlockMaliciousInput: true,
				SanitizeInput:       true,
				MaliciousPatterns: []string{
					"<script", "javascript:", "onload=", "onerror=",
					"union select", "drop table", "insert into",
					"../", "..\\", "%2e%2e%2f",
				},
			},
		},
		Encryption: EncryptionConfig{
			Algorithm:        "AES-256-GCM",
			KeySize:          32,
			DefaultKey:       "default_encryption_key_32_bytes",
			KeyRotation:      true,
			RotationInterval: "30d",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		Monitoring: MonitoringConfig{
			Enabled:     true,
			MetricsPort: 9090,
			MetricsPath: "/metrics",
			HealthCheck: HealthCheckConfig{
				Enabled:  true,
				Port:     8080,
				Path:     "/health",
				Interval: "30s",
			},
		},
	}
}

// LoadConfig 加载配置
func LoadConfig() (*SecurityConfig, error) {
	config := GetDefaultConfig()
	
	// 从环境变量加载配�?
	loadFromEnv(config)
	
	return config, nil
}

// loadFromEnv 从环境变量加载配�?
func loadFromEnv(config *SecurityConfig) {
	// 数据库配�?
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if username := os.Getenv("DB_USERNAME"); username != "" {
		config.Database.Username = username
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if database := os.Getenv("DB_DATABASE"); database != "" {
		config.Database.Database = database
	}
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// 威胁检测配�?
	if enabled := os.Getenv("THREAT_DETECTION_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.ThreatDetection.Enabled = e
		}
	}
	if autoBlock := os.Getenv("THREAT_DETECTION_AUTO_BLOCK"); autoBlock != "" {
		if ab, err := strconv.ParseBool(autoBlock); err == nil {
			config.ThreatDetection.AutoBlock = ab
		}
	}

	// 漏洞扫描配置
	if enabled := os.Getenv("VULNERABILITY_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Vulnerability.Enabled = e
		}
	}

	// 渗透测试配�?
	if enabled := os.Getenv("PENTEST_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Pentest.Enabled = e
		}
	}
	if allowedTargets := os.Getenv("PENTEST_ALLOWED_TARGETS"); allowedTargets != "" {
		config.Pentest.AllowedTargets = strings.Split(allowedTargets, ",")
	}

	// 安全教育配置
	if enabled := os.Getenv("SECURITY_EDUCATION_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.SecurityEducation.Enabled = e
		}
	}

	// 安全审计配置
	if enabled := os.Getenv("SECURITY_AUDIT_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.SecurityAudit.Enabled = e
		}
	}

	// 中间件配�?
	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Middleware.RateLimit.Enabled = e
		}
	}
	if maxRequests := os.Getenv("RATE_LIMIT_MAX_REQUESTS"); maxRequests != "" {
		if mr, err := strconv.Atoi(maxRequests); err == nil {
			config.Middleware.RateLimit.MaxRequests = mr
		}
	}

	// CORS配置
	if enabled := os.Getenv("CORS_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Middleware.CORS.Enabled = e
		}
	}
	if allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); allowedOrigins != "" {
		config.Middleware.CORS.AllowedOrigins = strings.Split(allowedOrigins, ",")
	}

	// 安全头配�?
	if enabled := os.Getenv("SECURITY_HEADERS_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Middleware.Security.Enabled = e
		}
	}

	// IP过滤配置
	if enabled := os.Getenv("IP_FILTER_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Middleware.IPFilter.Enabled = e
		}
	}
	if whitelist := os.Getenv("IP_FILTER_WHITELIST"); whitelist != "" {
		config.Middleware.IPFilter.Whitelist = strings.Split(whitelist, ",")
	}
	if blacklist := os.Getenv("IP_FILTER_BLACKLIST"); blacklist != "" {
		config.Middleware.IPFilter.Blacklist = strings.Split(blacklist, ",")
	}

	// 加密配置
	if algorithm := os.Getenv("ENCRYPTION_ALGORITHM"); algorithm != "" {
		config.Encryption.Algorithm = algorithm
	}
	if keySize := os.Getenv("ENCRYPTION_KEY_SIZE"); keySize != "" {
		if ks, err := strconv.Atoi(keySize); err == nil {
			config.Encryption.KeySize = ks
		}
	}
	if defaultKey := os.Getenv("ENCRYPTION_DEFAULT_KEY"); defaultKey != "" {
		config.Encryption.DefaultKey = defaultKey
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logging.Format = format
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Logging.Output = output
	}

	// 监控配置
	if enabled := os.Getenv("MONITORING_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Monitoring.Enabled = e
		}
	}
	if metricsPort := os.Getenv("MONITORING_METRICS_PORT"); metricsPort != "" {
		if mp, err := strconv.Atoi(metricsPort); err == nil {
			config.Monitoring.MetricsPort = mp
		}
	}
}
