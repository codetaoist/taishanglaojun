package security

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"github.com/taishanglaojun/core-services/security/config"
	"github.com/taishanglaojun/core-services/security/handlers"
	"github.com/taishanglaojun/core-services/security/models"
	"github.com/taishanglaojun/core-services/security/services"
)

// SecurityModule 安全模块结构
type SecurityModule struct {
	DB     *gorm.DB
	Router *gin.RouterGroup
	Config *config.SecurityConfig
	
	// 服务组件
	ThreatDetectionService   *services.ThreatDetectionService
	VulnerabilityService     *services.VulnerabilityService
	PentestService          *services.PentestService
	SecurityEducationService *services.SecurityEducationService
	SecurityAuditService    *services.SecurityAuditService
	
	// 处理器
	ThreatDetectionHandler   *handlers.ThreatDetectionHandler
	VulnerabilityHandler     *handlers.VulnerabilityHandler
	PentestHandler          *handlers.PentestHandler
	SecurityEducationHandler *handlers.SecurityEducationHandler
	SecurityAuditHandler    *handlers.SecurityAuditHandler
}

// NewSecurityModule 创建新的安全模块实例
func NewSecurityModule(db *gorm.DB, router *gin.RouterGroup, cfg *config.SecurityConfig) *SecurityModule {
	module := &SecurityModule{
		DB:     db,
		Router: router,
		Config: cfg,
	}
	
	// 初始化服务
	module.initServices()
	
	// 初始化处理器
	module.initHandlers()
	
	// 注册路由
	module.registerRoutes()
	
	// 运行数据库迁移
	module.migrate()
	
	return module
}

// initServices 初始化服务组件
func (sm *SecurityModule) initServices() {
	// 初始化威胁检测服务
	sm.ThreatDetectionService = services.NewThreatDetectionService(sm.DB, &sm.Config.ThreatDetection)
	
	// 初始化漏洞管理服务
	sm.VulnerabilityService = services.NewVulnerabilityService(sm.DB, &sm.Config.Vulnerability)
	
	// 初始化渗透测试服务
	sm.PentestService = services.NewPentestService(sm.DB, &sm.Config.Pentest)
	
	// 初始化安全教育服务
	sm.SecurityEducationService = services.NewSecurityEducationService(sm.DB, &sm.Config.SecurityEducation)
	
	// 初始化安全审计服务
	sm.SecurityAuditService = services.NewSecurityAuditService(sm.DB, &sm.Config.SecurityAudit)
}

// initHandlers 初始化处理器
func (sm *SecurityModule) initHandlers() {
	// 初始化威胁检测处理器
	sm.ThreatDetectionHandler = handlers.NewThreatDetectionHandler(sm.ThreatDetectionService)
	
	// 初始化漏洞管理处理器
	sm.VulnerabilityHandler = handlers.NewVulnerabilityHandler(sm.VulnerabilityService)
	
	// 初始化渗透测试处理器
	sm.PentestHandler = handlers.NewPentestHandler(sm.PentestService)
	
	// 初始化安全教育处理器
	sm.SecurityEducationHandler = handlers.NewSecurityEducationHandler(sm.SecurityEducationService)
	
	// 初始化安全审计处理器
	sm.SecurityAuditHandler = handlers.NewSecurityAuditHandler(sm.SecurityAuditService)
}

// registerRoutes 注册路由
func (sm *SecurityModule) registerRoutes() {
	// 威胁检测路由
	threatGroup := sm.Router.Group("/threat-detection")
	{
		threatGroup.GET("/alerts", sm.ThreatDetectionHandler.GetAlerts)
		threatGroup.GET("/rules", sm.ThreatDetectionHandler.GetRules)
		threatGroup.POST("/rules", sm.ThreatDetectionHandler.CreateRule)
		threatGroup.PUT("/rules/:id", sm.ThreatDetectionHandler.UpdateRule)
		threatGroup.DELETE("/rules/:id", sm.ThreatDetectionHandler.DeleteRule)
		threatGroup.POST("/scan", sm.ThreatDetectionHandler.StartScan)
	}
	
	// 漏洞管理路由
	vulnGroup := sm.Router.Group("/vulnerabilities")
	{
		vulnGroup.GET("/", sm.VulnerabilityHandler.GetVulnerabilities)
		vulnGroup.GET("/:id", sm.VulnerabilityHandler.GetVulnerability)
		vulnGroup.POST("/scan", sm.VulnerabilityHandler.StartScan)
		vulnGroup.PUT("/:id/status", sm.VulnerabilityHandler.UpdateStatus)
		vulnGroup.GET("/reports", sm.VulnerabilityHandler.GetReports)
	}
	
	// 渗透测试路由
	pentestGroup := sm.Router.Group("/pentest")
	{
		pentestGroup.GET("/projects", sm.PentestHandler.GetProjects)
		pentestGroup.POST("/projects", sm.PentestHandler.CreateProject)
		pentestGroup.GET("/projects/:id", sm.PentestHandler.GetProject)
		pentestGroup.PUT("/projects/:id", sm.PentestHandler.UpdateProject)
		pentestGroup.DELETE("/projects/:id", sm.PentestHandler.DeleteProject)
		pentestGroup.POST("/projects/:id/start", sm.PentestHandler.StartTest)
		pentestGroup.GET("/projects/:id/report", sm.PentestHandler.GetReport)
	}
	
	// 安全教育路由
	educationGroup := sm.Router.Group("/education")
	{
		educationGroup.GET("/courses", sm.SecurityEducationHandler.GetCourses)
		educationGroup.GET("/courses/:id", sm.SecurityEducationHandler.GetCourse)
		educationGroup.POST("/courses/:id/enroll", sm.SecurityEducationHandler.EnrollCourse)
		educationGroup.GET("/labs", sm.SecurityEducationHandler.GetLabs)
		educationGroup.POST("/labs/:id/start", sm.SecurityEducationHandler.StartLab)
		educationGroup.GET("/certifications", sm.SecurityEducationHandler.GetCertifications)
	}
	
	// 安全审计路由
	auditGroup := sm.Router.Group("/audit")
	{
		auditGroup.GET("/logs", sm.SecurityAuditHandler.GetAuditLogs)
		auditGroup.GET("/reports", sm.SecurityAuditHandler.GetAuditReports)
		auditGroup.POST("/reports", sm.SecurityAuditHandler.GenerateReport)
		auditGroup.GET("/compliance", sm.SecurityAuditHandler.GetComplianceStatus)
	}
}

// migrate 运行数据库迁移
func (sm *SecurityModule) migrate() {
	err := sm.DB.AutoMigrate(
		&models.ThreatAlert{},
		&models.DetectionRule{},
		&models.Vulnerability{},
		&models.ScanJob{},
		&models.PentestProject{},
		&models.PentestResult{},
		&models.SecurityCourse{},
		&models.LabEnvironment{},
		&models.SecurityCertification{},
		&models.AuditLog{},
		&models.SecurityEvent{},
		&models.ComplianceReport{},
		&models.UserSecurityProfile{},
		&models.ThreatIntelligence{},
	)
	
	if err != nil {
		log.Fatalf("Failed to migrate security module database: %v", err)
	}
	
	log.Println("Security module database migration completed successfully")
}

// Start 启动安全模块
func (sm *SecurityModule) Start() error {
	log.Println("Starting Security Module...")
	
	// 启动威胁检测服务
	if sm.Config.ThreatDetection.Enabled {
		go sm.ThreatDetectionService.Start()
		log.Println("Threat Detection Service started")
	}
	
	// 启动漏洞扫描服务
	if sm.Config.Vulnerability.Enabled {
		go sm.VulnerabilityService.Start()
		log.Println("Vulnerability Scanning Service started")
	}
	
	// 启动安全教育服务
	if sm.Config.SecurityEducation.Enabled {
		go sm.SecurityEducationService.Start()
		log.Println("Security Education Service started")
	}
	
	// 启动安全审计服务
	if sm.Config.SecurityAudit.Enabled {
		go sm.SecurityAuditService.Start()
		log.Println("Security Audit Service started")
	}
	
	log.Println("Security Module started successfully")
	return nil
}

// Stop 停止安全模块
func (sm *SecurityModule) Stop() error {
	log.Println("Stopping Security Module...")
	
	// 停止各个服务
	if sm.ThreatDetectionService != nil {
		sm.ThreatDetectionService.Stop()
		log.Println("Threat Detection Service stopped")
	}
	
	if sm.VulnerabilityService != nil {
		sm.VulnerabilityService.Stop()
		log.Println("Vulnerability Service stopped")
	}
	
	if sm.PentestService != nil {
		sm.PentestService.Stop()
		log.Println("Pentest Service stopped")
	}
	
	if sm.SecurityEducationService != nil {
		sm.SecurityEducationService.Stop()
		log.Println("Security Education Service stopped")
	}
	
	if sm.SecurityAuditService != nil {
		sm.SecurityAuditService.Stop()
		log.Println("Security Audit Service stopped")
	}
	
	log.Println("Security Module stopped successfully")
	return nil
}

// GetModuleInfo 获取模块信息
func (sm *SecurityModule) GetModuleInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "security",
		"version":     "1.0.0",
		"description": "太上老君AI平台安全模块",
		"services": []string{
			"threat-detection",
			"vulnerability-management",
			"penetration-testing",
			"security-education",
			"security-audit",
		},
		"status": "active",
	}
}