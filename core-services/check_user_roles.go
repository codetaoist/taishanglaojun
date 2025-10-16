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
	fmt.Println("🔍 检查用户角色和权限配置...")

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

	userID := "1c54d101-7840-46bb-81d9-4f430b0a41de"

	// 1. 检查users表中的用户信息
	fmt.Println("\n1. 检查users表中的用户信息:")
	var user struct {
		ID       string
		Username string
		Email    string
		Role     string
	}
	
	err = db.GetDB().Table("users").Where("id = ?", userID).First(&user).Error
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}
	
	fmt.Printf("   用户ID: %s\n", user.ID)
	fmt.Printf("   用户名: %s\n", user.Username)
	fmt.Printf("   邮箱: %s\n", user.Email)
	fmt.Printf("   角色: %s\n", user.Role)

	// 2. 检查user_roles表中的关联关系
	fmt.Println("\n2. 检查user_roles表中的关联关系:")
	var userRoles []struct {
		UserID string
		RoleID string
	}
	
	err = db.GetDB().Table("user_roles").Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		fmt.Printf("   查询user_roles失败: %v\n", err)
	} else if len(userRoles) == 0 {
		fmt.Println("   ❌ user_roles表中没有找到该用户的角色关联")
	} else {
		fmt.Printf("   ✅ 找到 %d 个角色关联:\n", len(userRoles))
		for _, ur := range userRoles {
			fmt.Printf("     - 用户ID: %s, 角色ID: %s\n", ur.UserID, ur.RoleID)
		}
	}

	// 3. 检查roles表中的角色信息
	fmt.Println("\n3. 检查roles表中的角色信息:")
	var roles []struct {
		ID     string
		Name   string
		Code   string
		Status string
	}
	
	err = db.GetDB().Table("roles").Find(&roles).Error
	if err != nil {
		fmt.Printf("   查询roles失败: %v\n", err)
	} else {
		fmt.Printf("   找到 %d 个角色:\n", len(roles))
		for _, role := range roles {
			fmt.Printf("     - ID: %s, 名称: %s, 代码: %s, 状态: %s\n", 
				role.ID, role.Name, role.Code, role.Status)
		}
	}

	// 4. 查找super_admin角色
	fmt.Println("\n4. 查找super_admin角色:")
	var superAdminRole struct {
		ID     string
		Name   string
		Code   string
		Status string
	}
	
	err = db.GetDB().Table("roles").Where("code = ? OR name = ?", "super_admin", "super_admin").First(&superAdminRole).Error
	if err != nil {
		fmt.Printf("   ❌ 未找到super_admin角色: %v\n", err)
		
		// 创建super_admin角色
		fmt.Println("\n5. 创建super_admin角色:")
		createRoleSQL := `
			INSERT INTO roles (id, name, code, description, status, created_at, updated_at) 
			VALUES ('super-admin-role-id', 'super_admin', 'super_admin', '超级管理员角色', 'active', datetime('now'), datetime('now'))
		`
		err = db.GetDB().Exec(createRoleSQL).Error
		if err != nil {
			fmt.Printf("   ❌ 创建super_admin角色失败: %v\n", err)
		} else {
			fmt.Println("   ✅ 成功创建super_admin角色")
			superAdminRole.ID = "super-admin-role-id"
		}
	} else {
		fmt.Printf("   ✅ 找到super_admin角色:\n")
		fmt.Printf("     - ID: %s\n", superAdminRole.ID)
		fmt.Printf("     - 名称: %s\n", superAdminRole.Name)
		fmt.Printf("     - 代码: %s\n", superAdminRole.Code)
		fmt.Printf("     - 状态: %s\n", superAdminRole.Status)
	}

	// 6. 检查并创建user_roles关联
	if superAdminRole.ID != "" {
		fmt.Println("\n6. 检查user_roles关联:")
		var existingRelation struct {
			UserID string
			RoleID string
		}
		
		err = db.GetDB().Table("user_roles").Where("user_id = ? AND role_id = ?", userID, superAdminRole.ID).First(&existingRelation).Error
		if err != nil {
			fmt.Println("   ❌ user_roles表中没有admin用户与super_admin角色的关联")
			fmt.Println("   💡 创建关联...")
			
			createRelationSQL := `
				INSERT INTO user_roles (user_id, role_id, created_at, updated_at) 
				VALUES (?, ?, datetime('now'), datetime('now'))
			`
			err = db.GetDB().Exec(createRelationSQL, userID, superAdminRole.ID).Error
			if err != nil {
				fmt.Printf("   ❌ 创建user_roles关联失败: %v\n", err)
			} else {
				fmt.Println("   ✅ 成功创建user_roles关联")
			}
		} else {
			fmt.Println("   ✅ user_roles表中已存在关联")
		}
	}

	fmt.Println("\n🔍 检查完成")
}