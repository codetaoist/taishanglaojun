package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// APIDocumentationService API文档管理服务
type APIDocumentationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAPIDocumentationService 创建API文档管理服务
func NewAPIDocumentationService(db *gorm.DB, logger *zap.Logger) *APIDocumentationService {
	return &APIDocumentationService{
		db:     db,
		logger: logger,
	}
}

// CategoryListRequest 分类列表请求
type CategoryListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	IsActive *bool  `json:"is_active" form:"is_active"`
}

// CategoryListResponse 分类列表响应
type CategoryListResponse struct {
	Categories []models.APICategory `json:"categories"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
}

// EndpointListRequest 接口列表请求
type EndpointListRequest struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	Keyword    string `json:"keyword" form:"keyword"`
	CategoryID string `json:"category_id" form:"category_id"`
	Method     string `json:"method" form:"method"`
	Status     string `json:"status" form:"status"`
}

// EndpointListResponse 接口列表响应
type EndpointListResponse struct {
	Endpoints []models.APIEndpoint `json:"endpoints"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"page_size"`
}

// EndpointDetailResponse 接口详情响应
type EndpointDetailResponse struct {
	Endpoint models.APIEndpoint `json:"endpoint"`
	Category models.APICategory `json:"category"`
}

// GetCategories 获取分类列表
func (s *APIDocumentationService) GetCategories(req CategoryListRequest) (*CategoryListResponse, error) {
	s.logger.Info("Getting API categories", zap.Any("request", req))

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.APICategory{})

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
	}

	// 状态过滤
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count categories", zap.Error(err))
		return nil, fmt.Errorf("获取分类总数失败: %v", err)
	}

	// 分页查询
	var categories []models.APICategory
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&categories).Error; err != nil {
		s.logger.Error("Failed to get categories", zap.Error(err))
		return nil, fmt.Errorf("获取分类列表失败: %v", err)
	}

	return &CategoryListResponse{
		Categories: categories,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// GetCategoryByID 根据ID获取分类
func (s *APIDocumentationService) GetCategoryByID(id string) (*models.APICategory, error) {
	s.logger.Info("Getting category by ID", zap.String("id", id))

	var category models.APICategory
	if err := s.db.Where("id = ?", id).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("分类不存在")
		}
		s.logger.Error("Failed to get category", zap.Error(err))
		return nil, fmt.Errorf("获取分类失败: %v", err)
	}

	return &category, nil
}

// GetEndpoints 获取接口列表
func (s *APIDocumentationService) GetEndpoints(req EndpointListRequest) (*EndpointListResponse, error) {
	s.logger.Info("Getting API endpoints", zap.Any("request", req))

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.APIEndpoint{}).Preload("Category")

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR path LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}

	// 分类过滤
	if req.CategoryID != "" {
		query = query.Where("category_id = ?", req.CategoryID)
	}

	// 方法过滤
	if req.Method != "" {
		query = query.Where("method = ?", strings.ToUpper(req.Method))
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count endpoints", zap.Error(err))
		return nil, fmt.Errorf("获取接口总数失败: %v", err)
	}

	// 分页查询
	var endpoints []models.APIEndpoint
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&endpoints).Error; err != nil {
		s.logger.Error("Failed to get endpoints", zap.Error(err))
		return nil, fmt.Errorf("获取接口列表失败: %v", err)
	}

	return &EndpointListResponse{
		Endpoints: endpoints,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// GetEndpointByID 根据ID获取接口详情
func (s *APIDocumentationService) GetEndpointByID(id string) (*EndpointDetailResponse, error) {
	s.logger.Info("Getting endpoint by ID", zap.String("id", id))

	var endpoint models.APIEndpoint
	if err := s.db.Preload("Category").Where("id = ?", id).First(&endpoint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("接口不存在")
		}
		s.logger.Error("Failed to get endpoint", zap.Error(err))
		return nil, fmt.Errorf("获取接口失败: %v", err)
	}

	// 增加查看次数
	s.db.Model(&endpoint).Update("view_count", gorm.Expr("view_count + 1"))

	return &EndpointDetailResponse{
		Endpoint: endpoint,
		Category: *endpoint.Category,
	}, nil
}

// GetEndpointsByCategory 根据分类获取接口列表
func (s *APIDocumentationService) GetEndpointsByCategory(categoryID string, page, pageSize int) (*EndpointListResponse, error) {
	s.logger.Info("Getting endpoints by category", zap.String("categoryID", categoryID))

	req := EndpointListRequest{
		Page:       page,
		PageSize:   pageSize,
		CategoryID: categoryID,
	}

	return s.GetEndpoints(req)
}

// SearchEndpoints 搜索接口
func (s *APIDocumentationService) SearchEndpoints(keyword string, page, pageSize int) (*EndpointListResponse, error) {
	s.logger.Info("Searching endpoints", zap.String("keyword", keyword))

	req := EndpointListRequest{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	}

	return s.GetEndpoints(req)
}

// GetStatistics 获取统计信息
func (s *APIDocumentationService) GetStatistics() (map[string]interface{}, error) {
	s.logger.Info("Getting API documentation statistics")

	stats := make(map[string]interface{})

	// 分类总数
	var categoryCount int64
	if err := s.db.Model(&models.APICategory{}).Where("is_active = ?", true).Count(&categoryCount).Error; err != nil {
		s.logger.Error("Failed to count categories", zap.Error(err))
		return nil, fmt.Errorf("获取分类统计失败: %v", err)
	}
	stats["categoryCount"] = categoryCount

	// 接口总数
	var endpointCount int64
	if err := s.db.Model(&models.APIEndpoint{}).Count(&endpointCount).Error; err != nil {
		s.logger.Error("Failed to count endpoints", zap.Error(err))
		return nil, fmt.Errorf("获取接口统计失败: %v", err)
	}
	stats["endpointCount"] = endpointCount

	// 按方法统计
	var methodStats []struct {
		Method string `json:"method"`
		Count  int64  `json:"count"`
	}
	if err := s.db.Model(&models.APIEndpoint{}).
		Select("method, COUNT(*) as count").
		Group("method").
		Find(&methodStats).Error; err != nil {
		s.logger.Error("Failed to get method statistics", zap.Error(err))
		return nil, fmt.Errorf("获取方法统计失败: %v", err)
	}
	stats["methodStats"] = methodStats

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	if err := s.db.Model(&models.APIEndpoint{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		s.logger.Error("Failed to get status statistics", zap.Error(err))
		return nil, fmt.Errorf("获取状态统计失败: %v", err)
	}
	stats["statusStats"] = statusStats

	// 按分类统计
	var categoryStats []struct {
		CategoryID   string `json:"category_id"`
		CategoryName string `json:"category_name"`
		Count        int64  `json:"count"`
	}
	if err := s.db.Table("api_endpoints").
		Select("api_endpoints.category_id, api_categories.name as category_name, COUNT(*) as count").
		Joins("LEFT JOIN api_categories ON api_endpoints.category_id = api_categories.id").
		Group("api_endpoints.category_id, api_categories.name").
		Find(&categoryStats).Error; err != nil {
		s.logger.Error("Failed to get category statistics", zap.Error(err))
		return nil, fmt.Errorf("获取分类统计失败: %v", err)
	}
	stats["categoryStats"] = categoryStats

	return stats, nil
}

// RecordAPITest 记录API测试
func (s *APIDocumentationService) RecordAPITest(endpointID, userID string, success bool, responseTime int, errorMsg string) error {
	s.logger.Info("Recording API test", zap.String("endpointID", endpointID), zap.Bool("success", success))

	testRecord := models.APITestRecord{
		ID:           uuid.New().String(),
		EndpointID:   endpointID,
		TestType:     "manual",
		RequestData:  "",
		ResponseData: "",
		StatusCode:   200,
		ResponseTime: responseTime,
		IsSuccess:    success,
		ErrorMsg:     errorMsg,
		Environment:  "dev",
		CreatedAt:    time.Now(),
		CreatedBy:    userID,
	}

	if err := s.db.Create(&testRecord).Error; err != nil {
		s.logger.Error("Failed to record API test", zap.Error(err))
		return fmt.Errorf("记录API测试失败: %v", err)
	}

	// 更新接口的测试统计
	updates := map[string]interface{}{
		"test_count":      gorm.Expr("test_count + 1"),
		"last_tested_at":  time.Now(),
	}

	if !success {
		updates["error_count"] = gorm.Expr("error_count + 1")
	}

	if err := s.db.Model(&models.APIEndpoint{}).Where("id = ?", endpointID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update endpoint test stats", zap.Error(err))
		return fmt.Errorf("更新接口测试统计失败: %v", err)
	}

	return nil
}