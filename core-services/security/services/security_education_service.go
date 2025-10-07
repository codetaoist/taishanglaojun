package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"github.com/taishanglaojun/core-services/security/models"
)

// SecurityEducationService 安全教育服务
type SecurityEducationService struct {
	db     *gorm.DB
	config *SecurityEducationConfig
	
	// 实验环境管理
	labManager     *LabManager
	courseManager  *CourseManager
	certManager    *CertificationManager
	
	// 用户进度跟踪
	userProgress map[string]*UserProgress
	mutex        sync.RWMutex
	
	// 控制通道
	stopChan chan bool
	running  bool
}

// SecurityEducationConfig 安全教育配置
type SecurityEducationConfig struct {
	Enabled              bool   `yaml:"enabled"`
	LabEnvironmentPath   string `yaml:"lab_environment_path"`
	MaxConcurrentLabs    int    `yaml:"max_concurrent_labs"`
	LabTimeout           int    `yaml:"lab_timeout"`
	CertificationEnabled bool   `yaml:"certification_enabled"`
	DefaultLanguage      string `yaml:"default_language"`
}

// LabManager 实验环境管理器
type LabManager struct {
	activeLabs map[string]*models.LabEnvironment
	mutex      sync.RWMutex
	config     *SecurityEducationConfig
}

// CourseManager 课程管理器
type CourseManager struct {
	courses map[string]*models.SecurityCourse
	mutex   sync.RWMutex
}

// CertificationManager 认证管理器
type CertificationManager struct {
	certifications map[string]*models.SecurityCertification
	mutex          sync.RWMutex
}

// UserProgress 用户进度
type UserProgress struct {
	UserID           string
	CompletedCourses []string
	CurrentCourse    string
	Progress         int
	LastActivity     time.Time
	LabSessions      []string
	Achievements     []string
}

// LabSession 实验会话
type LabSession struct {
	ID           string
	UserID       string
	LabID        string
	StartTime    time.Time
	EndTime      time.Time
	Status       string
	Progress     int
	Achievements []string
}

// NewSecurityEducationService 创建安全教育服务
func NewSecurityEducationService(db *gorm.DB, config *SecurityEducationConfig) *SecurityEducationService {
	service := &SecurityEducationService{
		db:           db,
		config:       config,
		userProgress: make(map[string]*UserProgress),
		stopChan:     make(chan bool),
		running:      false,
	}
	
	// 初始化管理器
	service.initManagers()
	
	// 加载课程和实验环境
	service.loadCourses()
	service.loadLabEnvironments()
	
	return service
}

// initManagers 初始化管理器
func (ses *SecurityEducationService) initManagers() {
	// 初始化实验环境管理器
	ses.labManager = &LabManager{
		activeLabs: make(map[string]*models.LabEnvironment),
		config:     ses.config,
	}
	
	// 初始化课程管理器
	ses.courseManager = &CourseManager{
		courses: make(map[string]*models.SecurityCourse),
	}
	
	// 初始化认证管理器
	ses.certManager = &CertificationManager{
		certifications: make(map[string]*models.SecurityCertification),
	}
}

// Start 启动安全教育服务
func (ses *SecurityEducationService) Start() {
	if ses.running {
		return
	}
	
	ses.running = true
	log.Println("Starting Security Education Service...")
	
	// 启动实验环境监控
	go ses.monitorLabEnvironments()
	
	// 启动用户进度同步
	go ses.syncUserProgress()
	
	log.Println("Security Education Service started successfully")
}

// Stop 停止安全教育服务
func (ses *SecurityEducationService) Stop() {
	if !ses.running {
		return
	}
	
	log.Println("Stopping Security Education Service...")
	ses.stopChan <- true
	ses.running = false
	log.Println("Security Education Service stopped")
}

// monitorLabEnvironments 监控实验环境
func (ses *SecurityEducationService) monitorLabEnvironments() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ses.checkLabStatus()
		case <-ses.stopChan:
			return
		}
	}
}

// syncUserProgress 同步用户进度
func (ses *SecurityEducationService) syncUserProgress() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ses.saveUserProgress()
		case <-ses.stopChan:
			return
		}
	}
}

// checkLabStatus 检查实验环境状态
func (ses *SecurityEducationService) checkLabStatus() {
	ses.labManager.mutex.RLock()
	labs := make([]*models.LabEnvironment, 0, len(ses.labManager.activeLabs))
	for _, lab := range ses.labManager.activeLabs {
		labs = append(labs, lab)
	}
	ses.labManager.mutex.RUnlock()
	
	for _, lab := range labs {
		if time.Since(lab.CreatedAt) > time.Duration(ses.config.LabTimeout)*time.Minute {
			ses.timeoutLab(lab)
		}
	}
}

// timeoutLab 超时实验环境
func (ses *SecurityEducationService) timeoutLab(lab *models.LabEnvironment) {
	ses.labManager.mutex.Lock()
	delete(ses.labManager.activeLabs, lab.ID)
	ses.labManager.mutex.Unlock()
	
	lab.Status = "timeout"
	ses.db.Save(lab)
	
	log.Printf("Lab environment %s timed out", lab.ID)
}

// saveUserProgress 保存用户进度
func (ses *SecurityEducationService) saveUserProgress() {
	ses.mutex.RLock()
	progressData := make(map[string]*UserProgress)
	for k, v := range ses.userProgress {
		progressData[k] = v
	}
	ses.mutex.RUnlock()
	
	for userID, progress := range progressData {
		// 这里可以将进度保存到数据库或缓存
		log.Printf("Saving progress for user %s: %d%% completed", userID, progress.Progress)
	}
}

// loadCourses 加载课程
func (ses *SecurityEducationService) loadCourses() {
	var courses []models.SecurityCourse
	if err := ses.db.Where("enabled = ?", true).Find(&courses).Error; err != nil {
		log.Printf("Failed to load security courses: %v", err)
		return
	}
	
	ses.courseManager.mutex.Lock()
	defer ses.courseManager.mutex.Unlock()
	
	for _, course := range courses {
		ses.courseManager.courses[course.ID] = &course
	}
	
	log.Printf("Loaded %d security courses", len(courses))
}

// loadLabEnvironments 加载实验环境
func (ses *SecurityEducationService) loadLabEnvironments() {
	var labs []models.LabEnvironment
	if err := ses.db.Where("status = ?", "available").Find(&labs).Error; err != nil {
		log.Printf("Failed to load lab environments: %v", err)
		return
	}
	
	log.Printf("Loaded %d lab environments", len(labs))
}

// CreateCourse 创建安全课程
func (ses *SecurityEducationService) CreateCourse(ctx context.Context, course *models.SecurityCourse) error {
	if err := ses.db.WithContext(ctx).Create(course).Error; err != nil {
		return err
	}
	
	// 添加到课程管理器
	ses.courseManager.mutex.Lock()
	ses.courseManager.courses[course.ID] = course
	ses.courseManager.mutex.Unlock()
	
	log.Printf("Created security course: %s", course.Title)
	return nil
}

// GetCourses 获取课程列表
func (ses *SecurityEducationService) GetCourses(ctx context.Context, category string, difficulty string) ([]models.SecurityCourse, error) {
	query := ses.db.WithContext(ctx).Where("enabled = ?", true)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	
	var courses []models.SecurityCourse
	err := query.Order("created_at DESC").Find(&courses).Error
	
	return courses, err
}

// GetCourse 获取课程详情
func (ses *SecurityEducationService) GetCourse(ctx context.Context, id string) (*models.SecurityCourse, error) {
	var course models.SecurityCourse
	err := ses.db.WithContext(ctx).Where("id = ?", id).First(&course).Error
	if err != nil {
		return nil, err
	}
	
	return &course, nil
}

// EnrollCourse 注册课程
func (ses *SecurityEducationService) EnrollCourse(ctx context.Context, userID, courseID string) error {
	// 检查课程是否存在
	course, err := ses.GetCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("course not found: %v", err)
	}
	
	// 检查用户是否已经注册
	// 这里可以添加注册记录表的检查逻辑
	
	// 更新用户进度
	ses.mutex.Lock()
	if progress, exists := ses.userProgress[userID]; exists {
		progress.CurrentCourse = courseID
		progress.LastActivity = time.Now()
	} else {
		ses.userProgress[userID] = &UserProgress{
			UserID:           userID,
			CompletedCourses: []string{},
			CurrentCourse:    courseID,
			Progress:         0,
			LastActivity:     time.Now(),
			LabSessions:      []string{},
			Achievements:     []string{},
		}
	}
	ses.mutex.Unlock()
	
	log.Printf("User %s enrolled in course %s", userID, course.Title)
	return nil
}

// UpdateCourseProgress 更新课程进度
func (ses *SecurityEducationService) UpdateCourseProgress(ctx context.Context, userID, courseID string, progress int) error {
	ses.mutex.Lock()
	defer ses.mutex.Unlock()
	
	userProgress, exists := ses.userProgress[userID]
	if !exists {
		return fmt.Errorf("user not enrolled in any course")
	}
	
	if userProgress.CurrentCourse != courseID {
		return fmt.Errorf("user not enrolled in this course")
	}
	
	userProgress.Progress = progress
	userProgress.LastActivity = time.Now()
	
	// 检查是否完成课程
	if progress >= 100 {
		userProgress.CompletedCourses = append(userProgress.CompletedCourses, courseID)
		userProgress.CurrentCourse = ""
		userProgress.Progress = 0
		
		// 添加完成成就
		userProgress.Achievements = append(userProgress.Achievements, fmt.Sprintf("completed_course_%s", courseID))
		
		log.Printf("User %s completed course %s", userID, courseID)
	}
	
	return nil
}

// CreateLabEnvironment 创建实验环境
func (ses *SecurityEducationService) CreateLabEnvironment(ctx context.Context, userID, labType string) (*models.LabEnvironment, error) {
	// 检查并发实验环境数量
	ses.labManager.mutex.RLock()
	activeCount := len(ses.labManager.activeLabs)
	ses.labManager.mutex.RUnlock()
	
	if activeCount >= ses.config.MaxConcurrentLabs {
		return nil, fmt.Errorf("maximum concurrent lab environments reached: %d", ses.config.MaxConcurrentLabs)
	}
	
	// 创建实验环境
	lab := &models.LabEnvironment{
		Name:        fmt.Sprintf("%s-lab-%d", labType, time.Now().Unix()),
		Type:        labType,
		Status:      "creating",
		UserID:      userID,
		Description: ses.getLabDescription(labType),
		Config: models.JSONB(map[string]interface{}{
			"type":       labType,
			"user_id":    userID,
			"created_at": time.Now(),
		}),
	}
	
	if err := ses.db.WithContext(ctx).Create(lab).Error; err != nil {
		return nil, err
	}
	
	// 异步创建实验环境
	go ses.setupLabEnvironment(lab)
	
	// 添加到活跃实验环境列表
	ses.labManager.mutex.Lock()
	ses.labManager.activeLabs[lab.ID] = lab
	ses.labManager.mutex.Unlock()
	
	log.Printf("Creating lab environment %s for user %s", lab.Name, userID)
	return lab, nil
}

// setupLabEnvironment 设置实验环境
func (ses *SecurityEducationService) setupLabEnvironment(lab *models.LabEnvironment) {
	// 模拟实验环境创建过程
	time.Sleep(10 * time.Second)
	
	// 更新状态为运行中
	lab.Status = "running"
	lab.AccessURL = fmt.Sprintf("https://lab.example.com/%s", lab.ID)
	lab.Credentials = models.JSONB(map[string]interface{}{
		"username": "student",
		"password": ses.generateRandomPassword(),
	})
	
	ses.db.Save(lab)
	
	log.Printf("Lab environment %s is now running", lab.Name)
}

// getLabDescription 获取实验环境描述
func (ses *SecurityEducationService) getLabDescription(labType string) string {
	descriptions := map[string]string{
		"web_security":     "Web应用安全实验环境，包含常见的Web漏洞练习",
		"network_security": "网络安全实验环境，包含网络攻防练习",
		"crypto":           "密码学实验环境，包含加密解密练习",
		"forensics":        "数字取证实验环境，包含取证分析练习",
		"reverse_eng":      "逆向工程实验环境，包含二进制分析练习",
		"malware_analysis": "恶意软件分析实验环境，包含恶意软件分析练习",
	}
	
	if desc, exists := descriptions[labType]; exists {
		return desc
	}
	
	return "通用安全实验环境"
}

// generateRandomPassword 生成随机密码
func (ses *SecurityEducationService) generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GetLabEnvironments 获取实验环境列表
func (ses *SecurityEducationService) GetLabEnvironments(ctx context.Context, userID string) ([]models.LabEnvironment, error) {
	var labs []models.LabEnvironment
	query := ses.db.WithContext(ctx)
	
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	err := query.Order("created_at DESC").Find(&labs).Error
	return labs, err
}

// GetLabEnvironment 获取实验环境详情
func (ses *SecurityEducationService) GetLabEnvironment(ctx context.Context, id string) (*models.LabEnvironment, error) {
	var lab models.LabEnvironment
	err := ses.db.WithContext(ctx).Where("id = ?", id).First(&lab).Error
	if err != nil {
		return nil, err
	}
	
	return &lab, nil
}

// DeleteLabEnvironment 删除实验环境
func (ses *SecurityEducationService) DeleteLabEnvironment(ctx context.Context, id string) error {
	// 从活跃列表中移除
	ses.labManager.mutex.Lock()
	delete(ses.labManager.activeLabs, id)
	ses.labManager.mutex.Unlock()
	
	// 从数据库中删除
	return ses.db.WithContext(ctx).Delete(&models.LabEnvironment{}, "id = ?", id).Error
}

// CreateCertification 创建认证
func (ses *SecurityEducationService) CreateCertification(ctx context.Context, cert *models.SecurityCertification) error {
	if err := ses.db.WithContext(ctx).Create(cert).Error; err != nil {
		return err
	}
	
	// 添加到认证管理器
	ses.certManager.mutex.Lock()
	ses.certManager.certifications[cert.ID] = cert
	ses.certManager.mutex.Unlock()
	
	log.Printf("Created security certification: %s", cert.Name)
	return nil
}

// GetCertifications 获取认证列表
func (ses *SecurityEducationService) GetCertifications(ctx context.Context, userID string) ([]models.SecurityCertification, error) {
	var certs []models.SecurityCertification
	query := ses.db.WithContext(ctx)
	
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	err := query.Order("issued_at DESC").Find(&certs).Error
	return certs, err
}

// IssueCertification 颁发认证
func (ses *SecurityEducationService) IssueCertification(ctx context.Context, userID, certType string) (*models.SecurityCertification, error) {
	// 检查用户是否满足认证条件
	if !ses.checkCertificationRequirements(userID, certType) {
		return nil, fmt.Errorf("user does not meet certification requirements")
	}
	
	// 创建认证
	cert := &models.SecurityCertification{
		Name:        ses.getCertificationName(certType),
		Type:        certType,
		UserID:      userID,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().AddDate(1, 0, 0), // 1年有效期
		Status:      "active",
		CertNumber:  ses.generateCertNumber(),
		Description: ses.getCertificationDescription(certType),
	}
	
	if err := ses.db.WithContext(ctx).Create(cert).Error; err != nil {
		return nil, err
	}
	
	log.Printf("Issued certification %s to user %s", cert.Name, userID)
	return cert, nil
}

// checkCertificationRequirements 检查认证要求
func (ses *SecurityEducationService) checkCertificationRequirements(userID, certType string) bool {
	ses.mutex.RLock()
	progress, exists := ses.userProgress[userID]
	ses.mutex.RUnlock()
	
	if !exists {
		return false
	}
	
	// 根据认证类型检查不同的要求
	switch certType {
	case "web_security":
		return len(progress.CompletedCourses) >= 3 && len(progress.LabSessions) >= 5
	case "network_security":
		return len(progress.CompletedCourses) >= 2 && len(progress.LabSessions) >= 3
	case "basic_security":
		return len(progress.CompletedCourses) >= 1
	default:
		return false
	}
}

// getCertificationName 获取认证名称
func (ses *SecurityEducationService) getCertificationName(certType string) string {
	names := map[string]string{
		"web_security":     "Web安全专家认证",
		"network_security": "网络安全专家认证",
		"basic_security":   "信息安全基础认证",
		"crypto":           "密码学专家认证",
		"forensics":        "数字取证专家认证",
	}
	
	if name, exists := names[certType]; exists {
		return name
	}
	
	return "安全专业认证"
}

// getCertificationDescription 获取认证描述
func (ses *SecurityEducationService) getCertificationDescription(certType string) string {
	descriptions := map[string]string{
		"web_security":     "证明持有者具备Web应用安全测试和防护的专业能力",
		"network_security": "证明持有者具备网络安全防护和渗透测试的专业能力",
		"basic_security":   "证明持有者具备信息安全基础知识和技能",
		"crypto":           "证明持有者具备密码学理论和实践应用能力",
		"forensics":        "证明持有者具备数字取证分析和调查能力",
	}
	
	if desc, exists := descriptions[certType]; exists {
		return desc
	}
	
	return "证明持有者具备相关安全专业能力"
}

// generateCertNumber 生成认证编号
func (ses *SecurityEducationService) generateCertNumber() string {
	return fmt.Sprintf("SEC-%d-%04d", time.Now().Year(), rand.Intn(9999))
}

// GetUserProgress 获取用户进度
func (ses *SecurityEducationService) GetUserProgress(ctx context.Context, userID string) (*UserProgress, error) {
	ses.mutex.RLock()
	progress, exists := ses.userProgress[userID]
	ses.mutex.RUnlock()
	
	if !exists {
		return &UserProgress{
			UserID:           userID,
			CompletedCourses: []string{},
			CurrentCourse:    "",
			Progress:         0,
			LastActivity:     time.Time{},
			LabSessions:      []string{},
			Achievements:     []string{},
		}, nil
	}
	
	return progress, nil
}

// GetUserAchievements 获取用户成就
func (ses *SecurityEducationService) GetUserAchievements(ctx context.Context, userID string) ([]string, error) {
	progress, err := ses.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return progress.Achievements, nil
}

// AddAchievement 添加成就
func (ses *SecurityEducationService) AddAchievement(ctx context.Context, userID, achievement string) error {
	ses.mutex.Lock()
	defer ses.mutex.Unlock()
	
	progress, exists := ses.userProgress[userID]
	if !exists {
		ses.userProgress[userID] = &UserProgress{
			UserID:           userID,
			CompletedCourses: []string{},
			CurrentCourse:    "",
			Progress:         0,
			LastActivity:     time.Now(),
			LabSessions:      []string{},
			Achievements:     []string{achievement},
		}
		return nil
	}
	
	// 检查是否已经有这个成就
	for _, existing := range progress.Achievements {
		if existing == achievement {
			return nil // 已经存在
		}
	}
	
	progress.Achievements = append(progress.Achievements, achievement)
	progress.LastActivity = time.Now()
	
	log.Printf("Added achievement %s to user %s", achievement, userID)
	return nil
}

// GetEducationStats 获取教育统计信息
func (ses *SecurityEducationService) GetEducationStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 课程统计
	var courseCount int64
	ses.db.WithContext(ctx).Model(&models.SecurityCourse{}).Count(&courseCount)
	stats["total_courses"] = courseCount
	
	// 实验环境统计
	var labCount int64
	ses.db.WithContext(ctx).Model(&models.LabEnvironment{}).Count(&labCount)
	stats["total_labs"] = labCount
	
	// 认证统计
	var certCount int64
	ses.db.WithContext(ctx).Model(&models.SecurityCertification{}).Count(&certCount)
	stats["total_certifications"] = certCount
	
	// 活跃用户统计
	ses.mutex.RLock()
	activeUsers := len(ses.userProgress)
	ses.mutex.RUnlock()
	stats["active_users"] = activeUsers
	
	// 按类型统计课程
	var courseStats []map[string]interface{}
	ses.db.WithContext(ctx).Model(&models.SecurityCourse{}).
		Select("category, count(*) as count").
		Group("category").
		Scan(&courseStats)
	stats["courses_by_category"] = courseStats
	
	// 按类型统计实验环境
	var labStats []map[string]interface{}
	ses.db.WithContext(ctx).Model(&models.LabEnvironment{}).
		Select("type, count(*) as count").
		Group("type").
		Scan(&labStats)
	stats["labs_by_type"] = labStats
	
	return stats, nil
}

// UpdateCourse 更新课程
func (ses *SecurityEducationService) UpdateCourse(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := ses.db.WithContext(ctx).Model(&models.SecurityCourse{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	
	// 重新加载课程
	ses.loadCourses()
	return nil
}

// DeleteCourse 删除课程
func (ses *SecurityEducationService) DeleteCourse(ctx context.Context, id string) error {
	// 从课程管理器中移除
	ses.courseManager.mutex.Lock()
	delete(ses.courseManager.courses, id)
	ses.courseManager.mutex.Unlock()
	
	// 从数据库中删除
	return ses.db.WithContext(ctx).Delete(&models.SecurityCourse{}, "id = ?", id).Error
}