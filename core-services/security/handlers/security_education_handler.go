package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/security/services"
	"github.com/taishanglaojun/core-services/security/models"
)

// SecurityEducationHandler 安全教育处理器
type SecurityEducationHandler struct {
	educationService *services.SecurityEducationService
}

// NewSecurityEducationHandler 创建安全教育处理器
func NewSecurityEducationHandler(educationService *services.SecurityEducationService) *SecurityEducationHandler {
	return &SecurityEducationHandler{
		educationService: educationService,
	}
}

// CreateCourse 创建安全课程
// @Summary 创建安全课程
// @Description 创建新的安全培训课程
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param course body models.SecurityCourse true "课程信息"
// @Success 201 {object} models.SecurityCourse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/courses [post]
func (h *SecurityEducationHandler) CreateCourse(c *gin.Context) {
	var course models.SecurityCourse
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证必要字段
	if course.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Course title is required",
		})
		return
	}

	if course.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Course category is required",
		})
		return
	}

	if err := h.educationService.CreateCourse(c.Request.Context(), &course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create course",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, course)
}

// GetCourses 获取安全课程列表
// @Summary 获取安全课程列表
// @Description 获取安全培训课程列表，支持分页和筛选
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param category query string false "课程分类"
// @Param level query string false "难度级别"
// @Param status query string false "课程状态"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/courses [get]
func (h *SecurityEducationHandler) GetCourses(c *gin.Context) {
	category := c.Query("category")
	level := c.Query("level")
	status := c.Query("status")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit

	courses, err := h.educationService.GetCourses(c.Request.Context(), category, level, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get courses",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": courses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(courses),
		},
	})
}

// GetCourse 获取安全课程详情
// @Summary 获取安全课程详情
// @Description 根据ID获取安全课程详情
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "课程ID"
// @Success 200 {object} models.SecurityCourse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/courses/{id} [get]
func (h *SecurityEducationHandler) GetCourse(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Course ID is required",
		})
		return
	}

	course, err := h.educationService.GetCourse(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Course not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, course)
}

// UpdateCourse 更新安全课程
// @Summary 更新安全课程
// @Description 更新安全课程信息
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "课程ID"
// @Param updates body map[string]interface{} true "更新内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/courses/{id} [put]
func (h *SecurityEducationHandler) UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Course ID is required",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.educationService.UpdateCourse(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update course",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course updated successfully",
	})
}

// DeleteCourse 删除安全课程
// @Summary 删除安全课程
// @Description 根据ID删除安全课程
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "课程ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/courses/{id} [delete]
func (h *SecurityEducationHandler) DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Course ID is required",
		})
		return
	}

	if err := h.educationService.DeleteCourse(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete course",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course deleted successfully",
	})
}

// EnrollCourse 报名课程
// @Summary 报名课程
// @Description 用户报名参加安全课程
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "课程ID"
// @Param enrollment body map[string]interface{} true "报名信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/courses/{id}/enroll [post]
func (h *SecurityEducationHandler) EnrollCourse(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Course ID is required",
		})
		return
	}

	var enrollment map[string]interface{}
	if err := c.ShouldBindJSON(&enrollment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userID, exists := enrollment["user_id"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	if err := h.educationService.EnrollCourse(c.Request.Context(), id, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to enroll course",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course enrollment successful",
	})
}

// CreateLabEnvironment 创建实验环境
// @Summary 创建实验环境
// @Description 创建新的安全实验环境
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param lab body models.LabEnvironment true "实验环境信息"
// @Success 201 {object} models.LabEnvironment
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/labs [post]
func (h *SecurityEducationHandler) CreateLabEnvironment(c *gin.Context) {
	var lab models.LabEnvironment
	if err := c.ShouldBindJSON(&lab); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证必要字段
	if lab.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Lab name is required",
		})
		return
	}

	if lab.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Lab type is required",
		})
		return
	}

	if err := h.educationService.CreateLabEnvironment(c.Request.Context(), &lab); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create lab environment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, lab)
}

// GetLabEnvironments 获取实验环境列表
// @Summary 获取实验环境列表
// @Description 获取安全实验环境列表，支持分页和筛选
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param type query string false "实验类型"
// @Param status query string false "环境状态"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/labs [get]
func (h *SecurityEducationHandler) GetLabEnvironments(c *gin.Context) {
	labType := c.Query("type")
	status := c.Query("status")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit

	labs, err := h.educationService.GetLabEnvironments(c.Request.Context(), labType, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get lab environments",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": labs,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(labs),
		},
	})
}

// GetLabEnvironment 获取实验环境详情
// @Summary 获取实验环境详情
// @Description 根据ID获取实验环境详情
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "实验环境ID"
// @Success 200 {object} models.LabEnvironment
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/labs/{id} [get]
func (h *SecurityEducationHandler) GetLabEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Lab environment ID is required",
		})
		return
	}

	lab, err := h.educationService.GetLabEnvironment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Lab environment not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, lab)
}

// DeleteLabEnvironment 删除实验环境
// @Summary 删除实验环境
// @Description 根据ID删除实验环境
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "实验环境ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/labs/{id} [delete]
func (h *SecurityEducationHandler) DeleteLabEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Lab environment ID is required",
		})
		return
	}

	if err := h.educationService.DeleteLabEnvironment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete lab environment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lab environment deleted successfully",
	})
}

// CreateCertification 创建安全认证
// @Summary 创建安全认证
// @Description 创建新的安全认证
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param certification body models.SecurityCertification true "认证信息"
// @Success 201 {object} models.SecurityCertification
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/certifications [post]
func (h *SecurityEducationHandler) CreateCertification(c *gin.Context) {
	var certification models.SecurityCertification
	if err := c.ShouldBindJSON(&certification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证必要字段
	if certification.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Certification name is required",
		})
		return
	}

	if certification.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Certification type is required",
		})
		return
	}

	if err := h.educationService.CreateCertification(c.Request.Context(), &certification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create certification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, certification)
}

// GetCertifications 获取安全认证列表
// @Summary 获取安全认证列表
// @Description 获取安全认证列表，支持分页和筛选
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param type query string false "认证类型"
// @Param status query string false "认证状态"
// @Param user_id query string false "用户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/certifications [get]
func (h *SecurityEducationHandler) GetCertifications(c *gin.Context) {
	certType := c.Query("type")
	status := c.Query("status")
	userID := c.Query("user_id")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit

	certifications, err := h.educationService.GetCertifications(c.Request.Context(), certType, status, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get certifications",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": certifications,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(certifications),
		},
	})
}

// GetCertification 获取安全认证详情
// @Summary 获取安全认证详情
// @Description 根据ID获取安全认证详情
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "认证ID"
// @Success 200 {object} models.SecurityCertification
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/certifications/{id} [get]
func (h *SecurityEducationHandler) GetCertification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Certification ID is required",
		})
		return
	}

	certification, err := h.educationService.GetCertification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Certification not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, certification)
}

// IssueCertification 颁发认证
// @Summary 颁发认证
// @Description 为用户颁发安全认证
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param id path string true "认证ID"
// @Param issue body map[string]interface{} true "颁发信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/certifications/{id}/issue [post]
func (h *SecurityEducationHandler) IssueCertification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Certification ID is required",
		})
		return
	}

	var issue map[string]interface{}
	if err := c.ShouldBindJSON(&issue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userID, exists := issue["user_id"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	if err := h.educationService.IssueCertification(c.Request.Context(), id, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to issue certification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certification issued successfully",
	})
}

// GetUserProgress 获取用户学习进度
// @Summary 获取用户学习进度
// @Description 获取指定用户的学习进度
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/users/{user_id}/progress [get]
func (h *SecurityEducationHandler) GetUserProgress(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	progress, err := h.educationService.GetUserProgress(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user progress",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetUserAchievements 获取用户成就
// @Summary 获取用户成就
// @Description 获取指定用户的学习成就
// @Tags 安全教育
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/users/{user_id}/achievements [get]
func (h *SecurityEducationHandler) GetUserAchievements(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	achievements, err := h.educationService.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user achievements",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, achievements)
}

// GetEducationStats 获取教育统计信息
// @Summary 获取教育统计信息
// @Description 获取安全教育相关的统计信息
// @Tags 安全教育
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/stats [get]
func (h *SecurityEducationHandler) GetEducationStats(c *gin.Context) {
	stats, err := h.educationService.GetEducationStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get education statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// StartEducationService 启动安全教育服务
// @Summary 启动安全教育服务
// @Description 启动安全教育服务
// @Tags 安全教育
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/service/start [post]
func (h *SecurityEducationHandler) StartEducationService(c *gin.Context) {
	h.educationService.Start()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Security education service started successfully",
	})
}

// StopEducationService 停止安全教育服务
// @Summary 停止安全教育服务
// @Description 停止安全教育服务
// @Tags 安全教育
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/education/service/stop [post]
func (h *SecurityEducationHandler) StopEducationService(c *gin.Context) {
	h.educationService.Stop()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Security education service stopped successfully",
	})
}

// GetServiceStatus 获取服务状态
// @Summary 获取服务状态
// @Description 获取安全教育服务的运行状态
// @Tags 安全教育
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/security/education/service/status [get]
func (h *SecurityEducationHandler) GetServiceStatus(c *gin.Context) {
	// 这里可以添加获取服务状态的逻辑
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now(),
	})
}