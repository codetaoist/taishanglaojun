package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// AppModuleService 应用模块服务
type AppModuleService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAppModuleService 创建应用模块服务
func NewAppModuleService(db *gorm.DB, logger *zap.Logger) *AppModuleService {
	return &AppModuleService{
		db:     db,
		logger: logger,
	}
}

// GetModulesByUserRole 根据用户角色获取模块列表
// 该方法根据用户角色查询所有已启用的模块，返回一个模块列表。
// 模块列表按优先级降序排序，相同优先级按名称升序排序。
func (s *AppModuleService) GetModulesByUserRole(userRole models.UserRole) ([]models.AppModule, error) {
	var modules []models.AppModule

	// 定义角色权重映射，用于确定用户角色的权限范围
	roleWeights := map[models.UserRole]int{
		models.RoleGuest:      0,
		models.RoleUser:       1,
		models.RolePremium:    2,
		models.RoleAdmin:      3,
		models.RoleSuperAdmin: 4,
	}

	userWeight := roleWeights[userRole]

	// 查询所有已启用的模块，筛选出用户角色有权限访问的模块
	err := s.db.Where("is_enabled = ? AND required_role IN (?)", true, s.getRolesForWeight(userWeight)).
		Order("priority DESC, name ASC").
		Find(&modules).Error

	if err != nil {
		s.logger.Error("Failed to get modules by user role", zap.Error(err), zap.String("role", string(userRole)))
		return nil, fmt.Errorf("获取模块失败: %v", err)
	}

	return modules, nil
}

// getRolesForWeight 根据角色权重获取角色列表
// 该方法根据角色权重返回一个角色列表，用于筛选模块访问权限。
func (s *AppModuleService) getRolesForWeight(weight int) []models.UserRole {
	roles := []models.UserRole{}

	if weight >= 0 {
		roles = append(roles, models.RoleGuest)
	}
	if weight >= 1 {
		roles = append(roles, models.RoleUser)
	}
	if weight >= 2 {
		roles = append(roles, models.RolePremium)
	}
	if weight >= 3 {
		roles = append(roles, models.RoleAdmin)
	}
	if weight >= 4 {
		roles = append(roles, models.RoleSuperAdmin)
	}

	return roles
}

// GetUserModules 根据用户ID获取用户模块列表
// 该方法根据用户ID查询所有已启用的模块，返回一个模块列表。
// 模块列表按优先级降序排序，相同优先级按名称升序排序。
func (s *AppModuleService) GetUserModules(userID uuid.UUID) ([]models.AppModule, error) {
	var modules []models.AppModule

	// 查询用户模块权限，筛选出用户角色有权限访问的模块
	query := `
		SELECT m.* FROM app_modules m
		LEFT JOIN user_module_permissions ump ON m.id = ump.module_id AND ump.user_id = ?
		WHERE m.is_enabled = true 
		AND (ump.enabled IS NULL OR ump.enabled = true)
		ORDER BY m.priority DESC, m.name ASC
	`

	err := s.db.Raw(query, userID).Scan(&modules).Error
	if err != nil {
		s.logger.Error("Failed to get user modules", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf(": %v", err)
	}

	return modules, nil
}

// CreateModule 创建应用模块
// 该方法用于创建一个新的应用模块，将模块信息保存到数据库中。
func (s *AppModuleService) CreateModule(module *models.AppModule) error {
	if err := s.db.Create(module).Error; err != nil {
		s.logger.Error("Failed to create module", zap.Error(err))
		return fmt.Errorf("创建模块失败: %v", err)
	}

	s.logger.Info("Module created successfully", zap.String("module_id", module.ID.String()))
	return nil
}

// UpdateModule 更新应用模块
// 该方法用于更新已存在的应用模块，将新的模块信息保存到数据库中。
func (s *AppModuleService) UpdateModule(module *models.AppModule) error {
	if err := s.db.Save(module).Error; err != nil {
		s.logger.Error("Failed to update module", zap.Error(err))
		return fmt.Errorf(": %v", err)
	}

	s.logger.Info("Module updated successfully", zap.String("module_id", module.ID.String()))
	return nil
}

// DeleteModule 删除应用模块
// 该方法用于删除已存在的应用模块，将模块从数据库中移除。
func (s *AppModuleService) DeleteModule(moduleID uuid.UUID) error {
	if err := s.db.Delete(&models.AppModule{}, moduleID).Error; err != nil {
		s.logger.Error("Failed to delete module", zap.Error(err))
		return fmt.Errorf("删除模块失败: %v", err)
	}

	s.logger.Info("Module deleted successfully", zap.String("module_id", moduleID.String()))
	return nil
}

// GetModuleByID 根据ID获取应用模块
// 该方法根据模块ID查询已存在的应用模块，返回一个模块实例。
// 如果模块不存在，将返回一个错误。
func (s *AppModuleService) GetModuleByID(moduleID uuid.UUID) (*models.AppModule, error) {
	var module models.AppModule

	if err := s.db.First(&module, moduleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模块不存在: %w", err)
		}
		s.logger.Error("Failed to get module by ID", zap.Error(err))
		return nil, fmt.Errorf("获取模块失败: %v", err)
	}

	return &module, nil
}

// SetUserModulePermission 设置用户模块权限
// 该方法用于设置用户对应用模块的访问权限，将权限信息保存到数据库中。
func (s *AppModuleService) SetUserModulePermission(userID, moduleID uuid.UUID, enabled bool) error {
	permission := &models.UserModulePermission{
		UserID:   userID,
		ModuleID: moduleID,
		Enabled:  enabled,
	}

	//  UPSERT 操作，若存在则更新，否则创建新记录
	err := s.db.Where("user_id = ? AND module_id = ?", userID, moduleID).
		Assign(models.UserModulePermission{Enabled: enabled, UpdatedAt: time.Now()}).
		FirstOrCreate(permission).Error

	if err != nil {
		s.logger.Error("Failed to set user module permission", zap.Error(err))
		return fmt.Errorf("设置用户模块权限失败: %v", err)
	}

	s.logger.Info("User module permission set successfully",
		zap.String("user_id", userID.String()),
		zap.String("module_id", moduleID.String()),
		zap.Bool("enabled", enabled))

	return nil
}

// GetUserPreference 获取用户偏好设置
// 该方法根据用户ID查询用户的偏好设置，若不存在则创建默认设置。
// 默认设置包括主题、语言、菜单样式、自动启动等。
func (s *AppModuleService) GetUserPreference(userID uuid.UUID) (*models.UserPreference, error) {
	var preference models.UserPreference

	err := s.db.Where("user_id = ?", userID).First(&preference).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 若不存在，则创建默认设置
			preference = models.UserPreference{
				UserID:    userID,
				Theme:     "light",
				Language:  "zh-CN",
				MenuStyle: "sidebar",
				AutoStart: false,
				Settings:  "{}",
			}

			if createErr := s.db.Create(&preference).Error; createErr != nil {
				s.logger.Error("Failed to create default user preference", zap.Error(createErr))
				return nil, fmt.Errorf(": %v", createErr)
			}

			return &preference, nil
		}

		s.logger.Error("Failed to get user preference", zap.Error(err))
		return nil, fmt.Errorf(": %v", err)
	}

	return &preference, nil
}

// UpdateUserPreference 更新用户偏好设置
// 该方法用于更新用户的偏好设置，将新的设置保存到数据库中。
func (s *AppModuleService) UpdateUserPreference(preference *models.UserPreference) error {
	if err := s.db.Save(preference).Error; err != nil {
		s.logger.Error("Failed to update user preference", zap.Error(err))
		return fmt.Errorf("更新用户偏好设置失败: %v", err)
	}

	s.logger.Info("User preference updated successfully", zap.String("user_id", preference.UserID.String()))
	return nil
}

// InitializeDefaultModules 初始化默认应用模块
// 该方法用于初始化应用的默认模块，将核心模块和基础模块保存到数据库中。
func (s *AppModuleService) InitializeDefaultModules() error {
	defaultModules := []models.AppModule{
		{
			Name:         "ai-chat",
			DisplayName:  "AI",
			Description:  "AI",
			Category:     models.CategoryAI,
			Icon:         "MessageCircle",
			Path:         "/ai-chat",
			RequiredRole: models.RoleUser,
			IsCore:       true,
			IsEnabled:    true,
			Priority:     100,
		},
		{
			Name:         "document-processing",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryFile,
			Icon:         "FileText",
			Path:         "/document-processing",
			RequiredRole: models.RoleUser,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     90,
		},
		{
			Name:         "file-transfer",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryFile,
			Icon:         "Upload",
			Path:         "/file-transfer",
			RequiredRole: models.RoleUser,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     80,
		},
		{
			Name:         "image-generation",
			DisplayName:  "",
			Description:  "AI",
			Category:     models.CategoryCreative,
			Icon:         "Image",
			Path:         "/image-generation",
			RequiredRole: models.RolePremium,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     70,
		},
		{
			Name:         "desktop-pet",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryCreative,
			Icon:         "Heart",
			Path:         "/desktop-pet",
			RequiredRole: models.RoleUser,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     60,
		},
		{
			Name:         "system-monitor",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategorySystem,
			Icon:         "Activity",
			Path:         "/system-monitor",
			RequiredRole: models.RoleAdmin,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     50,
		},
		{
			Name:         "connection-test",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategorySystem,
			Icon:         "Wifi",
			Path:         "/connection-test",
			RequiredRole: models.RoleUser,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     40,
		},
		{
			Name:         "user-auth",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryUser,
			Icon:         "User",
			Path:         "/user-auth",
			RequiredRole: models.RoleGuest,
			IsCore:       true,
			IsEnabled:    true,
			Priority:     30,
		},
		{
			Name:         "friend-management",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryUser,
			Icon:         "Users",
			Path:         "/friend-management",
			RequiredRole: models.RoleUser,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     20,
		},
		{
			Name:         "project-management",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryBusiness,
			Icon:         "Folder",
			Path:         "/project-management",
			RequiredRole: models.RolePremium,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     10,
		},
		{
			Name:         "app-management",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategorySystem,
			Icon:         "Settings",
			Path:         "/app-management",
			RequiredRole: models.RoleAdmin,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     5,
		},
		{
			Name:         "chat-management",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategoryUser,
			Icon:         "MessageSquare",
			Path:         "/chat-management",
			RequiredRole: models.RoleUser,
			IsCore:       false,
			IsEnabled:    true,
			Priority:     15,
		},
		{
			Name:         "settings",
			DisplayName:  "",
			Description:  "",
			Category:     models.CategorySystem,
			Icon:         "Settings",
			Path:         "/settings",
			RequiredRole: models.RoleUser,
			IsCore:       true,
			IsEnabled:    true,
			Priority:     1,
		},
	}

	for _, module := range defaultModules {
		//
		var existingModule models.AppModule
		err := s.db.Where("name = ?", module.Name).First(&existingModule).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新模块
			if createErr := s.db.Create(&module).Error; createErr != nil {
				s.logger.Error("Failed to create default module",
					zap.Error(createErr),
					zap.String("module_name", module.Name))
				return fmt.Errorf("创建默认模块失败: %v", createErr)
			}
			s.logger.Info("Default module created", zap.String("module_name", module.Name))
		} else if err != nil {
			s.logger.Error("Failed to check existing module", zap.Error(err))
			return fmt.Errorf("检查现有模块失败: %v", err)
		}
	}

	s.logger.Info("Default modules initialized successfully")
	return nil
}
