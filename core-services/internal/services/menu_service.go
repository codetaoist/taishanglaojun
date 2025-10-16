package services

import (
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// MenuService 菜单服务
type MenuService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMenuService 创建新的菜单服务
func NewMenuService(db *gorm.DB, logger *zap.Logger) *MenuService {
	return &MenuService{
		db:     db,
		logger: logger,
	}
}

// GetMenuTree 获取菜单树结构
func (s *MenuService) GetMenuTree(userRole models.UserRole) ([]models.MenuResponse, error) {
    s.logger.Info("Getting menu tree", zap.String("userRole", string(userRole)))

    // 确保系统管理下的静态子菜单存在（数据库管理/日志管理/问题跟踪）
    if err := s.ensureStaticAdminMenus(); err != nil {
        s.logger.Warn("ensureStaticAdminMenus failed", zap.Error(err))
    }

    // 获取允许的角色
    allowedRoles := s.getAllowedRoles(userRole)
    s.logger.Info("Allowed roles for user", zap.Any("allowedRoles", allowedRoles))

	var menus []models.Menu
	
	// 查询所有可见且启用的菜单，按排序字段排序
    err := s.db.Where("is_visible = ? AND is_enabled = ? AND (required_role = ? OR required_role IN (?))", 
        true, true, userRole, allowedRoles).
        Order("sort DESC, level ASC").
        Find(&menus).Error
	
	if err != nil {
		s.logger.Error("Failed to get menus from database", zap.Error(err))
		return nil, fmt.Errorf("获取菜单失败: %v", err)
	}

	s.logger.Info("Found menus for user", zap.Int("count", len(menus)), zap.Any("menus", menus))

	// 构建树形结构
	return s.buildMenuTree(menus, nil), nil
}

// GetMenuList 获取扁平化菜单列表
func (s *MenuService) GetMenuList(userRole models.UserRole) ([]models.MenuResponse, error) {
    s.logger.Info("Getting menu list", zap.String("userRole", string(userRole)))

    // 确保系统管理下的静态子菜单存在（数据库管理/日志管理/问题跟踪）
    if err := s.ensureStaticAdminMenus(); err != nil {
        s.logger.Warn("ensureStaticAdminMenus failed", zap.Error(err))
    }

    var menus []models.Menu

    err := s.db.Where("is_visible = ? AND is_enabled = ? AND (required_role = ? OR required_role IN (?))", 
        true, true, userRole, s.getAllowedRoles(userRole)).
        Order("sort DESC").
        Find(&menus).Error
	
	if err != nil {
		s.logger.Error("Failed to get menu list from database", zap.Error(err))
		return nil, fmt.Errorf("获取菜单列表失败: %v", err)
	}

	// 转换为响应格式
	var menuResponses []models.MenuResponse
	for _, menu := range menus {
		menuResponses = append(menuResponses, menu.ToResponse())
	}

	return menuResponses, nil
}

// GetMenuByID 根据ID获取菜单
func (s *MenuService) GetMenuByID(id uuid.UUID, userRole models.UserRole) (*models.MenuResponse, error) {
	s.logger.Info("Getting menu by ID", zap.String("id", id.String()), zap.String("userRole", string(userRole)))

	var menu models.Menu
	
	err := s.db.Where("id = ? AND is_visible = ? AND is_enabled = ? AND (required_role = ? OR required_role IN (?))", 
		id, true, true, userRole, s.getAllowedRoles(userRole)).
		First(&menu).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("菜单不存在或无权限访问")
		}
		s.logger.Error("Failed to get menu by ID from database", zap.Error(err))
		return nil, fmt.Errorf("获取菜单失败: %v", err)
	}

	response := menu.ToResponse()
	return &response, nil
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(menu *models.Menu) error {
	s.logger.Info("Creating menu", zap.String("name", menu.Name))

	// 如果有父菜单，设置层级
	if menu.ParentID != nil {
		var parent models.Menu
		if err := s.db.First(&parent, "id = ?", *menu.ParentID).Error; err != nil {
			return fmt.Errorf("父菜单不存在: %v", err)
		}
		menu.Level = parent.Level + 1
	} else {
		menu.Level = 1
	}

	if err := s.db.Create(menu).Error; err != nil {
		s.logger.Error("Failed to create menu", zap.Error(err))
		return fmt.Errorf("创建菜单失败: %v", err)
	}

	return nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(id uuid.UUID, updates map[string]interface{}) error {
	s.logger.Info("Updating menu", zap.String("id", id.String()))

	result := s.db.Model(&models.Menu{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		s.logger.Error("Failed to update menu", zap.Error(result.Error))
		return fmt.Errorf("更新菜单失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("菜单不存在")
	}

	return nil
}

// DeleteMenu 删除菜单（软删除）
func (s *MenuService) DeleteMenu(id uuid.UUID) error {
	s.logger.Info("Deleting menu", zap.String("id", id.String()))

	// 检查是否有子菜单
	var count int64
	s.db.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("存在子菜单，无法删除")
	}

	result := s.db.Delete(&models.Menu{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete menu", zap.Error(result.Error))
		return fmt.Errorf("删除菜单失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("菜单不存在")
	}

	return nil
}

// buildMenuTree 构建菜单树
func (s *MenuService) buildMenuTree(menus []models.Menu, parentID *uuid.UUID) []models.MenuResponse {
	var tree []models.MenuResponse

	for _, menu := range menus {
		// 检查是否为当前层级的菜单
		if (parentID == nil && menu.ParentID == nil) || 
		   (parentID != nil && menu.ParentID != nil && *menu.ParentID == *parentID) {
			
			menuResponse := menu.ToResponse()
			
			// 递归获取子菜单
			children := s.buildMenuTree(menus, &menu.ID)
			if len(children) > 0 {
				menuResponse.Children = children
			}
			
			tree = append(tree, menuResponse)
		}
	}

	return tree
}

// getAllowedRoles 获取用户角色允许访问的所有角色
func (s *MenuService) getAllowedRoles(userRole models.UserRole) []models.UserRole {
	roleHierarchy := map[models.UserRole][]models.UserRole{
		models.RoleGuest:      {models.RoleGuest},
		models.RoleUser:       {models.RoleGuest, models.RoleUser},
		models.RolePremium:    {models.RoleGuest, models.RoleUser, models.RolePremium},
		models.RoleAdmin:      {models.RoleGuest, models.RoleUser, models.RolePremium, models.RoleAdmin},
		models.RoleSuperAdmin: {models.RoleGuest, models.RoleUser, models.RolePremium, models.RoleAdmin, models.RoleSuperAdmin},
	}

	if roles, exists := roleHierarchy[userRole]; exists {
		return roles
	}

	return []models.UserRole{models.RoleGuest}
}

// SeedDefaultMenus 初始化默认菜单数据
func (s *MenuService) SeedDefaultMenus() error {
    s.logger.Info("Attempting to seed default menus")

    // 检查是否已有任何菜单数据
    var count int64
    s.db.Model(&models.Menu{}).Count(&count)
    if count > 0 {
        s.logger.Info("Menus already exist, skipping seed")
        return nil
    }

    // 创建默认菜单数据
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 顶级菜单
        dashboard := &models.Menu{
            Name:        "dashboard",
            Title:       "仪表板",
            Path:        "/dashboard",
            Icon:        "DashboardOutlined",
            Sort:        1,
            Level:       1,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleUser,
        }
        if err := tx.Create(dashboard).Error; err != nil {
            s.logger.Error("Failed to seed dashboard menu", zap.Error(err))
            return fmt.Errorf("创建仪表板菜单失败: %v", err)
        }

        profile := &models.Menu{
            Name:        "profile",
            Title:       "个人资料",
            Path:        "/profile",
            Icon:        "UserOutlined",
            Sort:        2,
            Level:       1,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleUser,
        }
        if err := tx.Create(profile).Error; err != nil {
            s.logger.Error("Failed to seed profile menu", zap.Error(err))
            return fmt.Errorf("创建个人资料菜单失败: %v", err)
        }

        projects := &models.Menu{
            Name:        "projects",
            Title:       "项目管理",
            Path:        "/projects",
            Icon:        "ProjectOutlined",
            Sort:        3,
            Level:       1,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleUser,
        }
        if err := tx.Create(projects).Error; err != nil {
            s.logger.Error("Failed to seed projects root menu", zap.Error(err))
            return fmt.Errorf("创建项目管理菜单失败: %v", err)
        }

        // 项目子菜单
        workspace := &models.Menu{
            Name:        "projects-workspace",
            Title:       "项目工作台",
            Path:        "/projects/workspace",
            Icon:        "DesktopOutlined",
            ParentID:    &projects.ID,
            Sort:        1,
            Level:       2,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleUser,
        }
        if err := tx.Create(workspace).Error; err != nil {
            s.logger.Error("Failed to seed projects workspace menu", zap.Error(err))
            return fmt.Errorf("创建项目工作台菜单失败: %v", err)
        }

        management := &models.Menu{
            Name:        "projects-management",
            Title:       "项目配置",
            Path:        "/projects/management",
            Icon:        "SettingOutlined",
            ParentID:    &projects.ID,
            Sort:        2,
            Level:       2,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleUser,
        }
        if err := tx.Create(management).Error; err != nil {
            s.logger.Error("Failed to seed projects management menu", zap.Error(err))
            return fmt.Errorf("创建项目配置菜单失败: %v", err)
        }

        // 管理后台
        admin := &models.Menu{
            Name:        "admin",
            Title:       "系统管理",
            Path:        "/admin",
            Icon:        "SettingOutlined",
            Sort:        20,
            Level:       1,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleAdmin,
        }
        if err := tx.Create(admin).Error; err != nil {
            s.logger.Error("Failed to seed admin root menu", zap.Error(err))
            return fmt.Errorf("创建系统管理菜单失败: %v", err)
        }

        adminMenus := &models.Menu{
            Name:        "admin-menus",
            Title:       "菜单管理",
            Path:        "/admin/menus",
            Icon:        "MenuOutlined",
            ParentID:    &admin.ID,
            Sort:        1,
            Level:       2,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleAdmin,
        }
        if err := tx.Create(adminMenus).Error; err != nil {
            s.logger.Error("Failed to seed admin menus page", zap.Error(err))
            return fmt.Errorf("创建菜单管理页面失败: %v", err)
        }

        adminPerms := &models.Menu{
            Name:        "admin-permissions",
            Title:       "权限管理",
            Path:        "/admin/permissions",
            Icon:        "SafetyOutlined",
            ParentID:    &admin.ID,
            Sort:        2,
            Level:       2,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleAdmin,
        }
        if err := tx.Create(adminPerms).Error; err != nil {
            s.logger.Error("Failed to seed admin permissions page", zap.Error(err))
            return fmt.Errorf("创建权限管理页面失败: %v", err)
        }

        s.logger.Info("Default menus seeded successfully")
        return nil
    })
}

// ensureStaticAdminMenus 确保系统管理下的静态子菜单存在
func (s *MenuService) ensureStaticAdminMenus() error {
    // 查找系统管理根菜单
    var admin models.Menu
    if err := s.db.Where("path = ? AND deleted_at IS NULL", "/admin").First(&admin).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            // 如果不存在，则创建一个系统管理根菜单
            root := &models.Menu{
                Name:        "admin",
                Title:       "系统管理",
                Path:        "/admin",
                Icon:        "SettingOutlined",
                Sort:        20,
                Level:       1,
                IsVisible:   true,
                IsEnabled:   true,
                RequiredRole: models.RoleAdmin,
            }
            if err := s.db.Create(root).Error; err != nil {
                return fmt.Errorf("创建系统管理根菜单失败: %v", err)
            }
            admin = *root
        } else {
            return err
        }
    }

    // 需要补齐的静态子菜单定义
    type childSpec struct {
        name  string
        title string
        path  string
        icon  string
        sort  int
    }
    specs := []childSpec{
        {name: "admin-database", title: "数据库管理", path: "/admin/database", icon: "DatabaseOutlined", sort: 3},
        {name: "admin-logs", title: "日志管理", path: "/admin/logs", icon: "FileTextOutlined", sort: 4},
        {name: "admin-issues", title: "问题跟踪", path: "/admin/issues", icon: "AuditOutlined", sort: 5},
    }

    for _, sp := range specs {
        // 检查是否已存在对应子菜单
        var exists models.Menu
        err := s.db.Where("path = ? AND deleted_at IS NULL", sp.path).First(&exists).Error
        if err == nil {
            // 已存在，跳过
            continue
        }
        if err != nil && err != gorm.ErrRecordNotFound {
            s.logger.Warn("Failed to query admin child menu", zap.String("path", sp.path), zap.Error(err))
            continue
        }

        // 创建缺失的子菜单
        m := &models.Menu{
            Name:        sp.name,
            Title:       sp.title,
            Path:        sp.path,
            Icon:        sp.icon,
            ParentID:    &admin.ID,
            Sort:        sp.sort,
            Level:       admin.Level + 1,
            IsVisible:   true,
            IsEnabled:   true,
            RequiredRole: models.RoleAdmin,
        }
        if err := s.db.Create(m).Error; err != nil {
            s.logger.Warn("Failed to create static admin child menu", zap.String("name", sp.name), zap.Error(err))
            // 不中断整个流程
        }
    }

    return nil
}

// DebugGetAllMenus 调试方法：获取所有菜单数据（包括软删除的）
func (s *MenuService) DebugGetAllMenus() ([]models.Menu, error) {
	s.logger.Info("Debug: Getting all menus including soft deleted")

	var menus []models.Menu
	// 使用 Unscoped() 来包含软删除的记录
	if err := s.db.Unscoped().Find(&menus).Error; err != nil {
		s.logger.Error("Failed to get all menus for debug", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Debug: Found menus", zap.Int("count", len(menus)))
	return menus, nil
}