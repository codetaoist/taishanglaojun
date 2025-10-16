package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// SetupMenuRoutes 设置菜单相关路由
func SetupMenuRoutes(apiV1 *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware, db *gorm.DB, logger *zap.Logger) {
    // 创建菜单处理器
    menuHandler := handlers.NewMenuHandler(db, logger)

    // 菜单路由组
    menuGroup := apiV1.Group("/menus")
    // 启用认证，确保用户上下文与权限一致
    menuGroup.Use(jwtMiddleware.AuthRequired())
    {
        // 获取菜单树
        menuGroup.GET("/tree", menuHandler.GetMenuTree)
        
        // 获取菜单列表
        menuGroup.GET("", menuHandler.GetMenuList)
        
        // 根据ID获取菜单
        menuGroup.GET("/:id", menuHandler.GetMenuById)

        // 初始化菜单数据
        menuGroup.POST("/seed", menuHandler.SeedMenus)
        
        // 调试端点：获取所有菜单数据
        menuGroup.GET("/debug/all", menuHandler.DebugGetAllMenus)
    }

    // 管理员菜单管理路由
    adminMenuGroup := apiV1.Group("/admin/menus")
    adminMenuGroup.Use(jwtMiddleware.AuthRequired())
    {
        // 创建菜单
        adminMenuGroup.POST("", menuHandler.CreateMenu)
        
        // 更新菜单
        adminMenuGroup.PUT("/:id", menuHandler.UpdateMenu)
        
        // 删除菜单
        adminMenuGroup.DELETE("/:id", menuHandler.DeleteMenu)
    }

    logger.Info("Menu routes registered successfully")
}