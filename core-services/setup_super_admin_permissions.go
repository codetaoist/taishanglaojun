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
	fmt.Println("🔧 设置超级管理员权限...")

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
		log.Fatal("初始化日志失败:", err)
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
		log.Fatal("连接数据库失败:", err)
	}

	// 1. 创建权限表（如果不存在）
	fmt.Println("\n1. 确保权限表存在:")
	createPermissionsTableSQL := `
		CREATE TABLE IF NOT EXISTS permissions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			code TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	err = db.GetDB().Exec(createPermissionsTableSQL).Error
	if err != nil {
		log.Fatal("创建permissions表失败:", err)
	}
	fmt.Println("   ✅ permissions表已确保存在")

	// 2. 创建角色权限关联表（如果不存在）
	fmt.Println("\n2. 确保角色权限关联表存在:")
	createRolePermissionsTableSQL := `
		CREATE TABLE IF NOT EXISTS role_permissions (
			role_id TEXT NOT NULL,
			permission_id TEXT NOT NULL,
			PRIMARY KEY(role_id, permission_id)
		)
	`
	err = db.GetDB().Exec(createRolePermissionsTableSQL).Error
	if err != nil {
		log.Fatal("创建role_permissions表失败:", err)
	}
	fmt.Println("   ✅ role_permissions表已确保存在")

	// 定义权限结构
	type Permission struct {
		ID          string
		Name        string
		Code        string
		Description string
	}

	// 3. 创建超级管理员权限
	fmt.Println("\n3. 创建超级管理员权限:")
	permissions := []Permission{
		{"perm-user-read", "用户查看", "user:read", "查看用户信息"},
		{"perm-user-write", "用户编辑", "user:write", "编辑用户信息"},
		{"perm-user-delete", "用户删除", "user:delete", "删除用户"},
		{"perm-user-admin", "用户管理", "user:admin", "完整用户管理权限"},
		{"perm-role-read", "角色查看", "role:read", "查看角色信息"},
		{"perm-role-write", "角色编辑", "role:write", "编辑角色信息"},
		{"perm-role-delete", "角色删除", "role:delete", "删除角色"},
		{"perm-role-admin", "角色管理", "role:admin", "完整角色管理权限"},
		{"perm-permission-read", "权限查看", "permission:read", "查看权限信息"},
		{"perm-permission-write", "权限编辑", "permission:write", "编辑权限信息"},
		{"perm-permission-delete", "权限删除", "permission:delete", "删除权限"},
		{"perm-permission-admin", "权限管理", "permission:admin", "完整权限管理权限"},
		{"perm-menu-read", "菜单查看", "menu:read", "查看菜单信息"},
		{"perm-menu-write", "菜单编辑", "menu:write", "编辑菜单信息"},
		{"perm-menu-delete", "菜单删除", "menu:delete", "删除菜单"},
		{"perm-menu-admin", "菜单管理", "menu:admin", "完整菜单管理权限"},
		{"perm-system-admin", "系统管理", "system:admin", "完整系统管理权限"},
		{"perm-super-admin", "超级管理员", "super:admin", "超级管理员权限"},
	}

	for _, perm := range permissions {
		// 检查权限是否已存在
		var existingPerm struct {
			ID string
		}
		err = db.GetDB().Table("permissions").Where("code = ?", perm.Code).First(&existingPerm).Error
		if err != nil {
			// 权限不存在，创建它
			createPermSQL := `
				INSERT INTO permissions (id, name, code, description, created_at, updated_at) 
				VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
			`
			err = db.GetDB().Exec(createPermSQL, perm.ID, perm.Name, perm.Code, perm.Description).Error
			if err != nil {
				fmt.Printf("   ❌ 创建权限 %s 失败: %v\n", perm.Code, err)
			} else {
				fmt.Printf("   ✅ 创建权限: %s (%s)\n", perm.Code, perm.Name)
			}
		} else {
			fmt.Printf("   ⚡ 权限已存在: %s (%s)\n", perm.Code, perm.Name)
		}
	}

	// 4. 为super_admin角色分配所有权限
	fmt.Println("\n4. 为super_admin角色分配权限:")
	
	// 获取super_admin角色ID
	var superAdminRole struct {
		ID string
	}
	err = db.GetDB().Table("roles").Where("code = ?", "super_admin").First(&superAdminRole).Error
	if err != nil {
		log.Fatal("未找到super_admin角色:", err)
	}

	// 获取所有权限
	var allPermissions []struct {
		ID string
	}
	err = db.GetDB().Table("permissions").Find(&allPermissions).Error
	if err != nil {
		log.Fatal("获取权限列表失败:", err)
	}

	// 为每个权限创建角色权限关联
	for _, perm := range allPermissions {
		// 检查关联是否已存在
		var count int64
		err = db.GetDB().Table("role_permissions").Where("role_id = ? AND permission_id = ?", superAdminRole.ID, perm.ID).Count(&count).Error
		if err != nil {
			fmt.Printf("   ❌ 检查权限关联失败: %v\n", err)
			continue
		}
		
		if count == 0 {
			// 关联不存在，创建它
			createRelationSQL := `
				INSERT INTO role_permissions (role_id, permission_id)
				VALUES (?, ?)
			`
			err = db.GetDB().Exec(createRelationSQL, superAdminRole.ID, perm.ID).Error
			if err != nil {
				fmt.Printf("   ❌ 创建角色权限关联失败: %v\n", err)
			} else {
				fmt.Printf("   ✅ 为super_admin分配权限: %s\n", perm.ID)
			}
		} else {
			fmt.Printf("   ⚡ 权限关联已存在: %s\n", perm.ID)
		}
	}

	fmt.Println("\n🔧 超级管理员权限设置完成")
}