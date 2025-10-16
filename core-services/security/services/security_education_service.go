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
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// SecurityEducationService 
type SecurityEducationService struct {
	db     *gorm.DB
	config *SecurityEducationConfig
	
	// 黷
	labManager     *LabManager
	courseManager  *CourseManager
	certManager    *CertificationManager
	
	// 
	userProgress map[string]*UserProgress
	mutex        sync.RWMutex
	
	// 
	stopChan chan bool
	running  bool
}

// SecurityEducationConfig 
type SecurityEducationConfig struct {
	Enabled              bool   `yaml:"enabled"`
	LabEnvironmentPath   string `yaml:"lab_environment_path"`
	MaxConcurrentLabs    int    `yaml:"max_concurrent_labs"`
	LabTimeout           int    `yaml:"lab_timeout"`
	CertificationEnabled bool   `yaml:"certification_enabled"`
	DefaultLanguage      string `yaml:"default_language"`
}

// LabManager 黷?
type LabManager struct {
	activeLabs map[string]*models.LabEnvironment
	mutex      sync.RWMutex
	config     *SecurityEducationConfig
}

// CourseManager ?
type CourseManager struct {
	courses map[string]*models.SecurityCourse
	mutex   sync.RWMutex
}

// CertificationManager ?
type CertificationManager struct {
	certifications map[string]*models.SecurityCertification
	mutex          sync.RWMutex
}

// UserProgress 
type UserProgress struct {
	UserID           string
	CompletedCourses []string
	CurrentCourse    string
	Progress         int
	LastActivity     time.Time
	LabSessions      []string
	Achievements     []string
}

// LabSession 
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

// NewSecurityEducationService 
func NewSecurityEducationService(db *gorm.DB, config *SecurityEducationConfig) *SecurityEducationService {
	service := &SecurityEducationService{
		db:           db,
		config:       config,
		userProgress: make(map[string]*UserProgress),
		stopChan:     make(chan bool),
		running:      false,
	}
	
	// 
	service.initManagers()
	
	// 黷?
	service.loadCourses()
	service.loadLabEnvironments()
	
	return service
}

// initManagers 
func (ses *SecurityEducationService) initManagers() {
	// 黷
	ses.labManager = &LabManager{
		activeLabs: make(map[string]*models.LabEnvironment),
		config:     ses.config,
	}
	
	// 
	ses.courseManager = &CourseManager{
		courses: make(map[string]*models.SecurityCourse),
	}
	
	// 
	ses.certManager = &CertificationManager{
		certifications: make(map[string]*models.SecurityCertification),
	}
}

// Start 
func (ses *SecurityEducationService) Start() {
	if ses.running {
		return
	}
	
	ses.running = true
	log.Println("Starting Security Education Service...")
	
	// 黷
	go ses.monitorLabEnvironments()
	
	// 
	go ses.syncUserProgress()
	
	log.Println("Security Education Service started successfully")
}

// Stop 
func (ses *SecurityEducationService) Stop() {
	if !ses.running {
		return
	}
	
	log.Println("Stopping Security Education Service...")
	ses.stopChan <- true
	ses.running = false
	log.Println("Security Education Service stopped")
}

// monitorLabEnvironments 黷
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

// syncUserProgress 
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

// checkLabStatus 黷?
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

// timeoutLab 黷
func (ses *SecurityEducationService) timeoutLab(lab *models.LabEnvironment) {
	ses.labManager.mutex.Lock()
	delete(ses.labManager.activeLabs, lab.ID)
	ses.labManager.mutex.Unlock()
	
	lab.Status = "timeout"
	ses.db.Save(lab)
	
	log.Printf("Lab environment %s timed out", lab.ID)
}

// saveUserProgress 
func (ses *SecurityEducationService) saveUserProgress() {
	ses.mutex.RLock()
	progressData := make(map[string]*UserProgress)
	for k, v := range ses.userProgress {
		progressData[k] = v
	}
	ses.mutex.RUnlock()
	
	for userID, progress := range progressData {
		// 浽
		log.Printf("Saving progress for user %s: %d%% completed", userID, progress.Progress)
	}
}

// loadCourses 
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

// loadLabEnvironments 黷
func (ses *SecurityEducationService) loadLabEnvironments() {
	var labs []models.LabEnvironment
	if err := ses.db.Where("status = ?", "available").Find(&labs).Error; err != nil {
		log.Printf("Failed to load lab environments: %v", err)
		return
	}
	
	log.Printf("Loaded %d lab environments", len(labs))
}

// CreateCourse 
func (ses *SecurityEducationService) CreateCourse(ctx context.Context, course *models.SecurityCourse) error {
	if err := ses.db.WithContext(ctx).Create(course).Error; err != nil {
		return err
	}
	
	// 
	ses.courseManager.mutex.Lock()
	ses.courseManager.courses[course.ID] = course
	ses.courseManager.mutex.Unlock()
	
	log.Printf("Created security course: %s", course.Title)
	return nil
}

// GetCourses 
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

// GetCourse 
func (ses *SecurityEducationService) GetCourse(ctx context.Context, id string) (*models.SecurityCourse, error) {
	var course models.SecurityCourse
	err := ses.db.WithContext(ctx).Where("id = ?", id).First(&course).Error
	if err != nil {
		return nil, err
	}
	
	return &course, nil
}

// EnrollCourse 
func (ses *SecurityEducationService) EnrollCourse(ctx context.Context, userID, courseID string) error {
	// ?
	course, err := ses.GetCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("course not found: %v", err)
	}
	
	// ?
	// 
	
	// 
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

// UpdateCourseProgress 
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
	
	// ?
	if progress >= 100 {
		userProgress.CompletedCourses = append(userProgress.CompletedCourses, courseID)
		userProgress.CurrentCourse = ""
		userProgress.Progress = 0
		
		// 
		userProgress.Achievements = append(userProgress.Achievements, fmt.Sprintf("completed_course_%s", courseID))
		
		log.Printf("User %s completed course %s", userID, courseID)
	}
	
	return nil
}

// CreateLabEnvironment 黷
func (ses *SecurityEducationService) CreateLabEnvironment(ctx context.Context, userID, labType string) (*models.LabEnvironment, error) {
	// 鲢黷?
	ses.labManager.mutex.RLock()
	activeCount := len(ses.labManager.activeLabs)
	ses.labManager.mutex.RUnlock()
	
	if activeCount >= ses.config.MaxConcurrentLabs {
		return nil, fmt.Errorf("maximum concurrent lab environments reached: %d", ses.config.MaxConcurrentLabs)
	}
	
	// 黷
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
	
	// 黷
	go ses.setupLabEnvironment(lab)
	
	// 黷?
	ses.labManager.mutex.Lock()
	ses.labManager.activeLabs[lab.ID] = lab
	ses.labManager.mutex.Unlock()
	
	log.Printf("Creating lab environment %s for user %s", lab.Name, userID)
	return lab, nil
}

// setupLabEnvironment 黷
func (ses *SecurityEducationService) setupLabEnvironment(lab *models.LabEnvironment) {
	// 黷
	time.Sleep(10 * time.Second)
	
	// ?
	lab.Status = "running"
	lab.AccessURL = fmt.Sprintf("https://lab.example.com/%s", lab.ID)
	lab.Credentials = models.JSONB(map[string]interface{}{
		"username": "student",
		"password": ses.generateRandomPassword(),
	})
	
	ses.db.Save(lab)
	
	log.Printf("Lab environment %s is now running", lab.Name)
}

// getLabDescription 黷
func (ses *SecurityEducationService) getLabDescription(labType string) string {
	descriptions := map[string]string{
		"web_security":     "Web黷Web",
		"network_security": "簲黷繥?,
		"crypto":           "黷",
		"forensics":        "黷?,
		"reverse_eng":      "黷",
		"malware_analysis": "黷?,
	}
	
	if desc, exists := descriptions[labType]; exists {
		return desc
	}
	
	return "黷"
}

// generateRandomPassword 
func (ses *SecurityEducationService) generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GetLabEnvironments 黷
func (ses *SecurityEducationService) GetLabEnvironments(ctx context.Context, userID string) ([]models.LabEnvironment, error) {
	var labs []models.LabEnvironment
	query := ses.db.WithContext(ctx)
	
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	err := query.Order("created_at DESC").Find(&labs).Error
	return labs, err
}

// GetLabEnvironment 黷
func (ses *SecurityEducationService) GetLabEnvironment(ctx context.Context, id string) (*models.LabEnvironment, error) {
	var lab models.LabEnvironment
	err := ses.db.WithContext(ctx).Where("id = ?", id).First(&lab).Error
	if err != nil {
		return nil, err
	}
	
	return &lab, nil
}

// DeleteLabEnvironment 黷
func (ses *SecurityEducationService) DeleteLabEnvironment(ctx context.Context, id string) error {
	// 
	ses.labManager.mutex.Lock()
	delete(ses.labManager.activeLabs, id)
	ses.labManager.mutex.Unlock()
	
	// ?
	return ses.db.WithContext(ctx).Delete(&models.LabEnvironment{}, "id = ?", id).Error
}

// CreateCertification 
func (ses *SecurityEducationService) CreateCertification(ctx context.Context, cert *models.SecurityCertification) error {
	if err := ses.db.WithContext(ctx).Create(cert).Error; err != nil {
		return err
	}
	
	// 
	ses.certManager.mutex.Lock()
	ses.certManager.certifications[cert.ID] = cert
	ses.certManager.mutex.Unlock()
	
	log.Printf("Created security certification: %s", cert.Name)
	return nil
}

// GetCertifications 
func (ses *SecurityEducationService) GetCertifications(ctx context.Context, userID string) ([]models.SecurityCertification, error) {
	var certs []models.SecurityCertification
	query := ses.db.WithContext(ctx)
	
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	err := query.Order("issued_at DESC").Find(&certs).Error
	return certs, err
}

// IssueCertification 
func (ses *SecurityEducationService) IssueCertification(ctx context.Context, userID, certType string) (*models.SecurityCertification, error) {
	// ?
	if !ses.checkCertificationRequirements(userID, certType) {
		return nil, fmt.Errorf("user does not meet certification requirements")
	}
	
	// 
	cert := &models.SecurityCertification{
		Name:        ses.getCertificationName(certType),
		Type:        certType,
		UserID:      userID,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().AddDate(1, 0, 0), // 1
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

// checkCertificationRequirements ?
func (ses *SecurityEducationService) checkCertificationRequirements(userID, certType string) bool {
	ses.mutex.RLock()
	progress, exists := ses.userProgress[userID]
	ses.mutex.RUnlock()
	
	if !exists {
		return false
	}
	
	// 鲻
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

// getCertificationName 
func (ses *SecurityEducationService) getCertificationName(certType string) string {
	names := map[string]string{
		"web_security":     "Web",
		"network_security": "簲",
		"basic_security":   "",
		"crypto":           "?,
		"forensics":        "",
	}
	
	if name, exists := names[certType]; exists {
		return name
	}
	
	return ""
}

// getCertificationDescription 
func (ses *SecurityEducationService) getCertificationDescription(certType string) string {
	descriptions := map[string]string{
		"web_security":     "Web",
		"network_security": "簲",
		"basic_security":   "?,
		"crypto":           "?,
		"forensics":        "",
	}
	
	if desc, exists := descriptions[certType]; exists {
		return desc
	}
	
	return "?
}

// generateCertNumber 
func (ses *SecurityEducationService) generateCertNumber() string {
	return fmt.Sprintf("SEC-%d-%04d", time.Now().Year(), rand.Intn(9999))
}

// GetUserProgress 
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

// GetUserAchievements 
func (ses *SecurityEducationService) GetUserAchievements(ctx context.Context, userID string) ([]string, error) {
	progress, err := ses.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return progress.Achievements, nil
}

// AddAchievement 
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
	
	// 
	for _, existing := range progress.Achievements {
		if existing == achievement {
			return nil // 
		}
	}
	
	progress.Achievements = append(progress.Achievements, achievement)
	progress.LastActivity = time.Now()
	
	log.Printf("Added achievement %s to user %s", achievement, userID)
	return nil
}

// GetEducationStats 
func (ses *SecurityEducationService) GetEducationStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 
	var courseCount int64
	ses.db.WithContext(ctx).Model(&models.SecurityCourse{}).Count(&courseCount)
	stats["total_courses"] = courseCount
	
	// 黷
	var labCount int64
	ses.db.WithContext(ctx).Model(&models.LabEnvironment{}).Count(&labCount)
	stats["total_labs"] = labCount
	
	// 
	var certCount int64
	ses.db.WithContext(ctx).Model(&models.SecurityCertification{}).Count(&certCount)
	stats["total_certifications"] = certCount
	
	// 
	ses.mutex.RLock()
	activeUsers := len(ses.userProgress)
	ses.mutex.RUnlock()
	stats["active_users"] = activeUsers
	
	// ?
	var courseStats []map[string]interface{}
	ses.db.WithContext(ctx).Model(&models.SecurityCourse{}).
		Select("category, count(*) as count").
		Group("category").
		Scan(&courseStats)
	stats["courses_by_category"] = courseStats
	
	// 黷?
	var labStats []map[string]interface{}
	ses.db.WithContext(ctx).Model(&models.LabEnvironment{}).
		Select("type, count(*) as count").
		Group("type").
		Scan(&labStats)
	stats["labs_by_type"] = labStats
	
	return stats, nil
}

// UpdateCourse 
func (ses *SecurityEducationService) UpdateCourse(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := ses.db.WithContext(ctx).Model(&models.SecurityCourse{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	
	// 
	ses.loadCourses()
	return nil
}

// DeleteCourse 
func (ses *SecurityEducationService) DeleteCourse(ctx context.Context, id string) error {
	// ?
	ses.courseManager.mutex.Lock()
	delete(ses.courseManager.courses, id)
	ses.courseManager.mutex.Unlock()
	
	// ?
	return ses.db.WithContext(ctx).Delete(&models.SecurityCourse{}, "id = ?", id).Error
}

