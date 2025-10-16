package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/security/services"
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// SecurityEducationHandler ?
type SecurityEducationHandler struct {
	educationService *services.SecurityEducationService
}

// NewSecurityEducationHandler ?
func NewSecurityEducationHandler(educationService *services.SecurityEducationService) *SecurityEducationHandler {
	return &SecurityEducationHandler{
		educationService: educationService,
	}
}

// CreateCourse 
// @Summary 
// @Description İ
// @Tags 
// @Accept json
// @Produce json
// @Param course body models.SecurityCourse true ""
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

	// 
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

// GetCourses 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param category query string false ""
// @Param level query string false ""
// @Param status query string false "?
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
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

// GetCourse 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
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

// UpdateCourse 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param updates body map[string]interface{} true ""
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

// DeleteCourse 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
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

// EnrollCourse 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param enrollment body map[string]interface{} true ""
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

// CreateLabEnvironment 黷
// @Summary 黷
// @Description İ黷
// @Tags 
// @Accept json
// @Produce json
// @Param lab body models.LabEnvironment true "黷"
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

	// 
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

// GetLabEnvironments 黷
// @Summary 黷
// @Description 黷?
// @Tags 
// @Accept json
// @Produce json
// @Param type query string false ""
// @Param status query string false "?
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
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

// GetLabEnvironment 黷
// @Summary 黷
// @Description ID黷
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "黷ID"
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

// DeleteLabEnvironment 黷
// @Summary 黷
// @Description ID黷
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "黷ID"
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

// CreateCertification 
// @Summary 
// @Description İ
// @Tags 
// @Accept json
// @Produce json
// @Param certification body models.SecurityCertification true ""
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

	// 
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

// GetCertifications 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param type query string false ""
// @Param status query string false "?
// @Param user_id query string false "ID"
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
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

// GetCertification 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
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

// IssueCertification 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param issue body map[string]interface{} true ""
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

// GetUserProgress 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
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

// GetUserAchievements 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
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

// GetEducationStats 
// @Summary 
// @Description ?
// @Tags 
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

// StartEducationService 
// @Summary 
// @Description 
// @Tags 
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

// StopEducationService 
// @Summary 
// @Description 
// @Tags 
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

// GetServiceStatus ?
// @Summary ?
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/security/education/service/status [get]
func (h *SecurityEducationHandler) GetServiceStatus(c *gin.Context) {
	// 
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now(),
	})
}

