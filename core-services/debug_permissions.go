package main

import (
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化日志记录器
	zapLogger, err := logger.New(logger.LogConfig{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		Filename:   cfg.Logger.Filename,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	})
	if err != nil {
		log.Fatal("初始化日志记录器失败:", err)
	}

	// 初始化数据库
	db, err := database.New(database.Config{
		Driver:          cfg.Database.Type,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Database,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.MaxLifetime) * time.Second,
		SSLMode:         cfg.Database.SSLMode,
		ConnectTimeout:  30 * time.Second,
	}, zapLogger)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	gormDB := db.GetDB()

	fmt.Println("🔍 调试权限系统数据:")

	// 1. 检查用户信息
	fmt.Println("\n1. 用户信息:")
	var users []struct {
		ID       string
		Username string
		Role     string
	}
	err = gormDB.Table("users").Select("id, username, role").Find(&users).Error
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}
	for _, user := range users {
		fmt.Printf("   用户: %s (%s) - 角色: %s\n", user.Username, user.ID, user.Role)
	}

	// 2. 检查角色表
	fmt.Println("\n2. 角色表:")
	var roles []struct {
		ID   string
		Name string
		Code string
	}
	err = gormDB.Table("roles").Select("id, name, code").Find(&roles).Error
	if err != nil {
		log.Fatal("查询角色失败:", err)
	}
	for _, role := range roles {
		fmt.Printf("   角色: %s (%s) - 代码: %s\n", role.Name, role.ID, role.Code)
	}

	// 3. 检查用户角色关联
	fmt.Println("\n3. 用户角色关联:")
	var userRoles []struct {
		UserID string
		RoleID string
	}
	err = gormDB.Table("user_roles").Select("user_id, role_id").Find(&userRoles).Error
	if err != nil {
		log.Fatal("查询用户角色关联失败:", err)
	}
	for _, ur := range userRoles {
		fmt.Printf("   用户: %s -> 角色: %s\n", ur.UserID, ur.RoleID)
	}

	// 4. 检查权限表
	fmt.Println("\n4. 权限表:")
	var permissions []struct {
		ID   string
		Name string
		Code string
	}
	err = gormDB.Table("permissions").Select("id, name, code").Find(&permissions).Error
	if err != nil {
		log.Fatal("查询权限失败:", err)
	}
	for _, perm := range permissions {
		fmt.Printf("   权限: %s (%s) - 代码: %s\n", perm.Name, perm.ID, perm.Code)
	}

	// 5. 检查角色权限关联
	fmt.Println("\n5. 角色权限关联:")
	var rolePermissions []struct {
		RoleID       string
		PermissionID string
	}
	err = gormDB.Table("role_permissions").Select("role_id, permission_id").Find(&rolePermissions).Error
	if err != nil {
		log.Fatal("查询角色权限关联失败:", err)
	}
	for _, rp := range rolePermissions {
		fmt.Printf("   角色: %s -> 权限: %s\n", rp.RoleID, rp.PermissionID)
	}

	// 6. 模拟getUserPermissionsFromDB函数的查询
	fmt.Println("\n6. 模拟权限查询 (admin用户):")
	adminUserID := "1c54d101-7840-46bb-81d9-4f430b0a41de"

	// 先检查roles表的结构
	fmt.Println("\n6.1 检查roles表结构:")
	var roleStructure []map[string]interface{}
	err = gormDB.Raw("PRAGMA table_info(roles)").Scan(&roleStructure).Error
	if err != nil {
		fmt.Printf("   ❌ 获取roles表结构失败: %v\n", err)
	} else {
		fmt.Printf("   roles表结构: %+v\n", roleStructure)
	}

	// 查询用户角色 - 使用简化的查询
	fmt.Println("\n6.2 查询用户角色:")
	var userRoleQuery []struct {
		RoleID   string `gorm:"column:role_id"`
		RoleIDDB string `gorm:"column:id"`
		RoleName string `gorm:"column:name"`
		RoleCode string `gorm:"column:code"`
	}

	err = gormDB.Table("user_roles").
		Select("user_roles.role_id, roles.id, roles.name, roles.code").
		Joins("LEFT JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", adminUserID).
		Scan(&userRoleQuery).Error

	if err != nil {
		fmt.Printf("   ❌ 查询用户角色失败: %v\n", err)
	} else {
		fmt.Printf("   用户角色查询结果: %+v\n", userRoleQuery)
		
		// 模拟handlers.go中的逻辑
		var roleNames []string
		var roleIDs []string
		
		for _, ur := range userRoleQuery {
			roleNames = append(roleNames, ur.RoleName)
			roleIDs = append(roleIDs, ur.RoleID)
		}
		
		fmt.Printf("   提取的角色名称: %v\n", roleNames)
		fmt.Printf("   提取的角色ID: %v\n", roleIDs)
	}

	if len(userRoleQuery) > 0 {
		var roleIDs []string
		for _, ur := range userRoleQuery {
			roleIDs = append(roleIDs, ur.RoleID)
		}

		// 查询角色权限
		var permissionQuery []struct {
			Code string
		}

		err = gormDB.Table("role_permissions").
			Select("permissions.code").
			Joins("LEFT JOIN permissions ON role_permissions.permission_id = permissions.id").
			Where("role_permissions.role_id IN ?", roleIDs).
			Scan(&permissionQuery).Error

		if err != nil {
			fmt.Printf("   ❌ 查询角色权限失败: %v\n", err)
		} else {
			fmt.Printf("   权限查询结果: %+v\n", permissionQuery)
		}
	}

	fmt.Println("\n🔧 调试完成")
}