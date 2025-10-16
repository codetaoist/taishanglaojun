package handlers

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.uber.org/zap"
    "gorm.io/gorm"

    "github.com/codetaoist/taishanglaojun/core-services/internal/models"
    "github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	menuService *services.MenuService
	logger      *zap.Logger
}

// NewMenuHandler 创建新的菜单处理器
func NewMenuHandler(db *gorm.DB, logger *zap.Logger) *MenuHandler {
	return &MenuHandler{
		menuService: services.NewMenuService(db, logger),
		logger:      logger,
	}
}

// GetMenuTree 获取菜单树
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
    h.logger.Info("Getting menu tree")

    // 获取用户角色，默认为USER；兼容JWT中设置的"role"键并规范大小写
    userRole := models.RoleUser
    if roleVal, exists := c.Get("userRole"); exists {
        if roleStr, ok := roleVal.(string); ok && roleStr != "" {
            userRole = models.UserRole(strings.ToUpper(roleStr))
        }
    } else if roleVal, exists := c.Get("role"); exists {
        if roleStr, ok := roleVal.(string); ok && roleStr != "" {
            userRole = models.UserRole(strings.ToUpper(roleStr))
        }
    }

	menuTree, err := h.menuService.GetMenuTree(userRole)
	if err != nil {
		h.logger.Error("Failed to get menu tree", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取菜单树失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    menuTree,
		"message": "success",
	})
}

// GetMenuList 获取菜单列表
func (h *MenuHandler) GetMenuList(c *gin.Context) {
    h.logger.Info("Getting menu list")

    // 获取用户角色，默认为USER；兼容JWT中设置的"role"键并规范大小写
    userRole := models.RoleUser
    if roleVal, exists := c.Get("userRole"); exists {
        if roleStr, ok := roleVal.(string); ok && roleStr != "" {
            userRole = models.UserRole(strings.ToUpper(roleStr))
        }
    } else if roleVal, exists := c.Get("role"); exists {
        if roleStr, ok := roleVal.(string); ok && roleStr != "" {
            userRole = models.UserRole(strings.ToUpper(roleStr))
        }
    }

	menuList, err := h.menuService.GetMenuList(userRole)
	if err != nil {
		h.logger.Error("Failed to get menu list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取菜单列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    menuList,
		"message": "success",
	})
}

// GetMenuById 根据ID获取菜单
func (h *MenuHandler) GetMenuById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid menu ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的菜单ID",
		})
		return
	}

	h.logger.Info("Getting menu by ID", zap.String("id", id.String()))

    // 获取用户角色，默认为USER；兼容JWT中设置的"role"键并规范大小写
    userRole := models.RoleUser
    if roleVal, exists := c.Get("userRole"); exists {
        if roleStr, ok := roleVal.(string); ok && roleStr != "" {
            userRole = models.UserRole(strings.ToUpper(roleStr))
        }
    } else if roleVal, exists := c.Get("role"); exists {
        if roleStr, ok := roleVal.(string); ok && roleStr != "" {
            userRole = models.UserRole(strings.ToUpper(roleStr))
        }
    }

	menu, err := h.menuService.GetMenuByID(id, userRole)
	if err != nil {
		h.logger.Error("Failed to get menu by ID", zap.String("id", id.String()), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    menu,
		"message": "success",
	})
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	h.logger.Info("Creating menu")

	var menu models.Menu
	if err := c.ShouldBindJSON(&menu); err != nil {
		h.logger.Error("Invalid menu data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的菜单数据",
			"error":   err.Error(),
		})
		return
	}

	if err := h.menuService.CreateMenu(&menu); err != nil {
		h.logger.Error("Failed to create menu", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建菜单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"data":    menu.ToResponse(),
		"message": "菜单创建成功",
	})
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid menu ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的菜单ID",
		})
		return
	}

	h.logger.Info("Updating menu", zap.String("id", id.String()))

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的更新数据",
			"error":   err.Error(),
		})
		return
	}

	if err := h.menuService.UpdateMenu(id, updates); err != nil {
		h.logger.Error("Failed to update menu", zap.String("id", id.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新菜单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "菜单更新成功",
	})
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid menu ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的菜单ID",
		})
		return
	}

	h.logger.Info("Deleting menu", zap.String("id", id.String()))

	if err := h.menuService.DeleteMenu(id); err != nil {
		h.logger.Error("Failed to delete menu", zap.String("id", id.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除菜单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "菜单删除成功",
	})
}

// SeedMenus 初始化菜单数据
func (h *MenuHandler) SeedMenus(c *gin.Context) {
	h.logger.Info("Seeding menu data")

	if err := h.menuService.SeedDefaultMenus(); err != nil {
		h.logger.Error("Failed to seed menu data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "初始化菜单数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "菜单数据初始化成功",
	})
}

// DebugGetAllMenus 调试端点：获取所有菜单数据（包括软删除的）
func (h *MenuHandler) DebugGetAllMenus(c *gin.Context) {
	h.logger.Info("Debug: Getting all menus")

	menus, err := h.menuService.DebugGetAllMenus()
	if err != nil {
		h.logger.Error("Failed to get all menus for debug", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取调试菜单数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": menus,
		"message": "success",
	})
}