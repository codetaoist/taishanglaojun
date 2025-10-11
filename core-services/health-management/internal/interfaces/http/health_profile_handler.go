package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/application"
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthProfileHandler 健康档案HTTP处理�?
type HealthProfileHandler struct {
	healthProfileService *application.HealthProfileService
}

// NewHealthProfileHandler 创建健康档案HTTP处理�?
func NewHealthProfileHandler(healthProfileService *application.HealthProfileService) *HealthProfileHandler {
	return &HealthProfileHandler{
		healthProfileService: healthProfileService,
	}
}

// CreateHealthProfileRequest 创建健康档案请求
type CreateHealthProfileRequest struct {
	UserID            string                 `json:"user_id" binding:"required"`
	Gender            domain.Gender          `json:"gender" binding:"required"`
	DateOfBirth       time.Time              `json:"date_of_birth" binding:"required"`
	Height            float64                `json:"height" binding:"required"`
	BloodType         domain.BloodType       `json:"blood_type" binding:"required"`
	EmergencyContact  string                 `json:"emergency_contact"`
	MedicalHistory    []string               `json:"medical_history"`
	Allergies         []string               `json:"allergies"`
	Medications       []string               `json:"medications"`
	HealthGoals       []string               `json:"health_goals"`
	PreferredUnits    map[string]string      `json:"preferred_units"`
	NotificationPrefs map[string]interface{} `json:"notification_preferences"`
}

// UpdateHealthProfileRequest 更新健康档案请求
type UpdateHealthProfileRequest struct {
	Gender            *domain.Gender         `json:"gender,omitempty"`
	DateOfBirth       *time.Time             `json:"date_of_birth,omitempty"`
	Height            *float64               `json:"height,omitempty"`
	BloodType         *domain.BloodType      `json:"blood_type,omitempty"`
	EmergencyContact  *string                `json:"emergency_contact,omitempty"`
	PreferredUnits    map[string]string      `json:"preferred_units,omitempty"`
	NotificationPrefs map[string]interface{} `json:"notification_preferences,omitempty"`
}

// AddMedicalHistoryRequest 添加病史请求
type AddMedicalHistoryRequest struct {
	Condition string `json:"condition" binding:"required"`
}

// AddAllergyRequest 添加过敏史请�?
type AddAllergyRequest struct {
	Allergen string `json:"allergen" binding:"required"`
}

// AddMedicationRequest 添加药物请求
type AddMedicationRequest struct {
	Medication string `json:"medication" binding:"required"`
}

// SetHealthGoalsRequest 设置健康目标请求
type SetHealthGoalsRequest struct {
	Goals []string `json:"goals" binding:"required"`
}

// HealthProfileResponse 健康档案响应
type HealthProfileResponse struct {
	ID                string                 `json:"id"`
	UserID            string                 `json:"user_id"`
	Gender            domain.Gender          `json:"gender"`
	DateOfBirth       time.Time              `json:"date_of_birth"`
	Age               int                    `json:"age"`
	Height            float64                `json:"height"`
	BloodType         domain.BloodType       `json:"blood_type"`
	EmergencyContact  string                 `json:"emergency_contact"`
	MedicalHistory    []string               `json:"medical_history"`
	Allergies         []string               `json:"allergies"`
	Medications       []string               `json:"medications"`
	HealthGoals       []string               `json:"health_goals"`
	PreferredUnits    map[string]string      `json:"preferred_units"`
	NotificationPrefs map[string]interface{} `json:"notification_preferences"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// HealthProfileListResponse 健康档案列表响应
type HealthProfileListResponse struct {
	Data       []HealthProfileResponse `json:"data"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

// BMIResponse BMI响应
type BMIResponse struct {
	UserID string  `json:"user_id"`
	BMI    float64 `json:"bmi"`
	Status string  `json:"status"`
}

// CreateHealthProfile 创建健康档案
// @Summary 创建健康档案
// @Description 创建新的健康档案
// @Tags health-profile
// @Accept json
// @Produce json
// @Param request body CreateHealthProfileRequest true "创建健康档案请求"
// @Success 201 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles [post]
func (h *HealthProfileHandler) CreateHealthProfile(c *gin.Context) {
	var req CreateHealthProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	createReq := &application.CreateHealthProfileRequest{
		UserID:            userID,
		Gender:            req.Gender,
		DateOfBirth:       req.DateOfBirth,
		Height:            req.Height,
		BloodType:         req.BloodType,
		EmergencyContact:  req.EmergencyContact,
		MedicalHistory:    req.MedicalHistory,
		Allergies:         req.Allergies,
		Medications:       req.Medications,
		HealthGoals:       req.HealthGoals,
		PreferredUnits:    req.PreferredUnits,
		NotificationPrefs: req.NotificationPrefs,
	}

	resp, err := h.healthProfileService.CreateHealthProfile(c.Request.Context(), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_FAILED",
			Message: "创建健康档案失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, h.toHealthProfileResponse(resp))
}

// GetHealthProfile 获取健康档案
// @Summary 获取健康档案
// @Description 根据ID获取健康档案
// @Tags health-profile
// @Produce json
// @Param id path string true "健康档案ID"
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id} [get]
func (h *HealthProfileHandler) GetHealthProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	resp, err := h.healthProfileService.GetHealthProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_FAILED",
			Message: "获取健康档案失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// GetHealthProfileByUser 获取用户的健康档�?
// @Summary 获取用户的健康档�?
// @Description 根据用户ID获取健康档案
// @Tags health-profile
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-profile [get]
func (h *HealthProfileHandler) GetHealthProfileByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	resp, err := h.healthProfileService.GetHealthProfileByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_FAILED",
			Message: "获取健康档案失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// UpdateHealthProfile 更新健康档案
// @Summary 更新健康档案
// @Description 更新健康档案基本信息
// @Tags health-profile
// @Accept json
// @Produce json
// @Param id path string true "健康档案ID"
// @Param request body UpdateHealthProfileRequest true "更新健康档案请求"
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id} [put]
func (h *HealthProfileHandler) UpdateHealthProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	var req UpdateHealthProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	updateReq := &application.UpdateHealthProfileRequest{
		ID:                id,
		Gender:            req.Gender,
		DateOfBirth:       req.DateOfBirth,
		Height:            req.Height,
		BloodType:         req.BloodType,
		EmergencyContact:  req.EmergencyContact,
		PreferredUnits:    req.PreferredUnits,
		NotificationPrefs: req.NotificationPrefs,
	}

	resp, err := h.healthProfileService.UpdateHealthProfile(c.Request.Context(), updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: "更新健康档案失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// DeleteHealthProfile 删除健康档案
// @Summary 删除健康档案
// @Description 删除健康档案
// @Tags health-profile
// @Produce json
// @Param id path string true "健康档案ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id} [delete]
func (h *HealthProfileHandler) DeleteHealthProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	err = h.healthProfileService.DeleteHealthProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_FAILED",
			Message: "删除健康档案失败",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddMedicalHistory 添加病史
// @Summary 添加病史
// @Description 为健康档案添加病史记�?
// @Tags health-profile
// @Accept json
// @Produce json
// @Param id path string true "健康档案ID"
// @Param request body AddMedicalHistoryRequest true "添加病史请求"
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id}/medical-history [post]
func (h *HealthProfileHandler) AddMedicalHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	var req AddMedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	addReq := &application.AddMedicalHistoryRequest{
		ProfileID: id,
		Condition: req.Condition,
	}

	resp, err := h.healthProfileService.AddMedicalHistory(c.Request.Context(), addReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ADD_MEDICAL_HISTORY_FAILED",
			Message: "添加病史失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// RemoveMedicalHistory 删除病史
// @Summary 删除病史
// @Description 从健康档案中删除病史记录
// @Tags health-profile
// @Produce json
// @Param id path string true "健康档案ID"
// @Param condition path string true "病史条件"
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id}/medical-history/{condition} [delete]
func (h *HealthProfileHandler) RemoveMedicalHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	condition := c.Param("condition")

	removeReq := &application.RemoveMedicalHistoryRequest{
		ProfileID: id,
		Condition: condition,
	}

	resp, err := h.healthProfileService.RemoveMedicalHistory(c.Request.Context(), removeReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "REMOVE_MEDICAL_HISTORY_FAILED",
			Message: "删除病史失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// AddAllergy 添加过敏�?
// @Summary 添加过敏�?
// @Description 为健康档案添加过敏史记录
// @Tags health-profile
// @Accept json
// @Produce json
// @Param id path string true "健康档案ID"
// @Param request body AddAllergyRequest true "添加过敏史请�?
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id}/allergies [post]
func (h *HealthProfileHandler) AddAllergy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	var req AddAllergyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	addReq := &application.AddAllergyRequest{
		ProfileID: id,
		Allergen:  req.Allergen,
	}

	resp, err := h.healthProfileService.AddAllergy(c.Request.Context(), addReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ADD_ALLERGY_FAILED",
			Message: "添加过敏史失�?,
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// RemoveAllergy 删除过敏�?
// @Summary 删除过敏�?
// @Description 从健康档案中删除过敏史记�?
// @Tags health-profile
// @Produce json
// @Param id path string true "健康档案ID"
// @Param allergen path string true "过敏�?
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id}/allergies/{allergen} [delete]
func (h *HealthProfileHandler) RemoveAllergy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	allergen := c.Param("allergen")

	removeReq := &application.RemoveAllergyRequest{
		ProfileID: id,
		Allergen:  allergen,
	}

	resp, err := h.healthProfileService.RemoveAllergy(c.Request.Context(), removeReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "REMOVE_ALLERGY_FAILED",
			Message: "删除过敏史失�?,
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// SetHealthGoals 设置健康目标
// @Summary 设置健康目标
// @Description 为健康档案设置健康目�?
// @Tags health-profile
// @Accept json
// @Produce json
// @Param id path string true "健康档案ID"
// @Param request body SetHealthGoalsRequest true "设置健康目标请求"
// @Success 200 {object} HealthProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id}/health-goals [put]
func (h *HealthProfileHandler) SetHealthGoals(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	var req SetHealthGoalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	setReq := &application.SetHealthGoalsRequest{
		ProfileID: id,
		Goals:     req.Goals,
	}

	resp, err := h.healthProfileService.SetHealthGoals(c.Request.Context(), setReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SET_HEALTH_GOALS_FAILED",
			Message: "设置健康目标失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileResponse(resp))
}

// CalculateBMI 计算BMI
// @Summary 计算BMI
// @Description 根据健康档案计算BMI
// @Tags health-profile
// @Produce json
// @Param id path string true "健康档案ID"
// @Param weight query float64 true "体重(kg)"
// @Success 200 {object} BMIResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles/{id}/bmi [get]
func (h *HealthProfileHandler) CalculateBMI(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康档案ID格式无效",
		})
		return
	}

	weightStr := c.Query("weight")
	if weightStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_WEIGHT",
			Message: "体重参数不能为空",
		})
		return
	}

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_WEIGHT",
			Message: "体重参数格式无效",
		})
		return
	}

	bmiReq := &application.CalculateBMIRequest{
		ProfileID: id,
		Weight:    weight,
	}

	resp, err := h.healthProfileService.CalculateBMI(c.Request.Context(), bmiReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CALCULATE_BMI_FAILED",
			Message: "计算BMI失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康档案不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, BMIResponse{
		UserID: resp.UserID.String(),
		BMI:    resp.BMI,
		Status: resp.Status,
	})
}

// ListHealthProfiles 获取健康档案列表
// @Summary 获取健康档案列表
// @Description 分页获取健康档案列表
// @Tags health-profile
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} HealthProfileListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-profiles [get]
func (h *HealthProfileHandler) ListHealthProfiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &application.ListHealthProfilesRequest{
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.healthProfileService.ListHealthProfiles(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "LIST_FAILED",
			Message: "获取健康档案列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthProfileListResponse(resp, page, pageSize))
}

// 辅助方法

func (h *HealthProfileHandler) toHealthProfileResponse(profile *application.HealthProfileResponse) HealthProfileResponse {
	return HealthProfileResponse{
		ID:                profile.ID.String(),
		UserID:            profile.UserID.String(),
		Gender:            profile.Gender,
		DateOfBirth:       profile.DateOfBirth,
		Age:               profile.Age,
		Height:            profile.Height,
		BloodType:         profile.BloodType,
		EmergencyContact:  profile.EmergencyContact,
		MedicalHistory:    profile.MedicalHistory,
		Allergies:         profile.Allergies,
		Medications:       profile.Medications,
		HealthGoals:       profile.HealthGoals,
		PreferredUnits:    profile.PreferredUnits,
		NotificationPrefs: profile.NotificationPrefs,
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}
}

func (h *HealthProfileHandler) toHealthProfileListResponse(profiles []*application.HealthProfileResponse, page, pageSize int) HealthProfileListResponse {
	responses := make([]HealthProfileResponse, len(profiles))
	for i, profile := range profiles {
		responses[i] = h.toHealthProfileResponse(profile)
	}

	total := int64(len(profiles)) // 这里简化处理，实际应该从服务层获取总数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return HealthProfileListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
