package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
    "gorm.io/gorm"

    "github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
    "github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
)

// SetupPermissionRoutes 注册权限相关路由
func SetupPermissionRoutes(apiV1 *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) {
    // 创建权限处理器
    permissionHandler := handlers.NewPermissionHandler(db, redisClient, logger)

    // 权限路由组（需要认证）
    permissionsGroup := apiV1.Group("/permissions")
    permissionsGroup.Use(jwtMiddleware.AuthRequired())
    {
        // 权限检查
        permissionsGroup.POST("/check", permissionHandler.CheckPermission)
        permissionsGroup.POST("/check/batch", permissionHandler.CheckPermissions)

        // 权限 CRUD
        permissionsGroup.GET("", permissionHandler.ListPermissions)
        permissionsGroup.POST("", permissionHandler.CreatePermission)
        permissionsGroup.GET("/:id", permissionHandler.GetPermission)
        permissionsGroup.PUT("/:id", permissionHandler.UpdatePermission)
        permissionsGroup.DELETE("/:id", permissionHandler.DeletePermission)
    }

    // 角色路由组（需要认证）
    rolesGroup := apiV1.Group("/roles")
    rolesGroup.Use(jwtMiddleware.AuthRequired())
    {
        // 角色 CRUD
        rolesGroup.GET("", permissionHandler.ListRoles)
        rolesGroup.POST("", permissionHandler.CreateRole)
        rolesGroup.GET("/:id", permissionHandler.GetRole)
        rolesGroup.PUT("/:id", permissionHandler.UpdateRole)
        rolesGroup.DELETE("/:id", permissionHandler.DeleteRole)

        // 角色权限管理
        rolesGroup.GET("/:id/permissions", permissionHandler.GetRolePermissions)
        rolesGroup.POST("/:id/permissions", permissionHandler.AssignPermissionsToRole)
        rolesGroup.DELETE("/:id/permissions/:permissionId", permissionHandler.RemovePermissionFromRole)
    }

    // 用户角色路由组（需要认证）
    userRolesGroup := apiV1.Group("/user-roles")
    userRolesGroup.Use(jwtMiddleware.AuthRequired())
    {
        // 用户角色管理
        userRolesGroup.GET("/:userId/roles", permissionHandler.GetUserRoles)
        userRolesGroup.POST("/:userId/roles", permissionHandler.AssignRolesToUser)
        userRolesGroup.DELETE("/:userId/roles/:roleId", permissionHandler.RemoveRoleFromUser)
    }

    logger.Info("Permission routes registered successfully")
}