package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID       string `gorm:"column:id;primaryKey"`
	Username string `gorm:"column:username"`
	Email    string `gorm:"column:email"`
	Status   string `gorm:"column:status"`
	IsActive bool   `gorm:"column:is_active"`
}

type Role struct {
	ID          string `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	Code        string `gorm:"column:code"`
	Description string `gorm:"column:description"`
	Type        string `gorm:"column:type"`
	Level       int    `gorm:"column:level"`
	Status      string `gorm:"column:status"`
	IsActive    bool   `gorm:"column:is_active"`
}

type UserRole struct {
	ID        string    `gorm:"column:id;primaryKey"`
	UserID    string    `gorm:"column:user_id"`
	RoleID    string    `gorm:"column:role_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Permission struct {
	ID          string `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	Code        string `gorm:"column:code"`
	Description string `gorm:"column:description"`
	Resource    string `gorm:"column:resource"`
	Action      string `gorm:"column:action"`
	Status      string `gorm:"column:status"`
	IsActive    bool   `gorm:"column:is_active"`
}

type RolePermission struct {
	RoleID       string `gorm:"column:role_id;primaryKey"`
	PermissionID string `gorm:"column:permission_id;primaryKey"`
}

func main() {
	// 数据库连接字符串
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	fmt.Println("MySQL数据库连接成功!")

	// 查找admin用户
	var adminUser User
	if err := db.Table("users").Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		log.Fatal("查找admin用户失败:", err)
	}
	fmt.Printf("✅ 找到admin用户: %s (%s)\n", adminUser.Username, adminUser.Email)

	// 查找admin角色
	var adminRole Role
	if err := db.Table("roles").Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		log.Fatal("查找admin角色失败:", err)
	}
	fmt.Printf("✅ 找到admin角色: %s (%s)\n", adminRole.Name, adminRole.Code)

	// 检查用户是否已经有admin角色
	var existingUserRole UserRole
	if err := db.Table("user_roles").Where("user_id = ? AND role_id = ?", adminUser.ID, adminRole.ID).First(&existingUserRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 分配admin角色给用户
			userRole := UserRole{
				ID:        uuid.New().String(),
				UserID:    adminUser.ID,
				RoleID:    adminRole.ID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := db.Table("user_roles").Create(&userRole).Error; err != nil {
				log.Fatal("分配admin角色失败:", err)
			}
			fmt.Println("✅ admin角色分配成功")
		} else {
			log.Fatal("检查用户角色失败:", err)
		}
	} else {
		fmt.Println("✅ admin用户已经拥有admin角色")
	}

	// 查找所有权限
	var allPermissions []Permission
	if err := db.Table("permissions").Find(&allPermissions).Error; err != nil {
		log.Fatal("查询权限失败:", err)
	}
	fmt.Printf("📋 系统中共有 %d 个权限\n", len(allPermissions))

	// 为admin角色分配所有权限
	fmt.Println("🔐 正在为admin角色分配所有权限...")
	assignedCount := 0
	for _, permission := range allPermissions {
		// 检查权限是否已经分配
		var existingRolePermission RolePermission
		if err := db.Table("role_permissions").Where("role_id = ? AND permission_id = ?", adminRole.ID, permission.ID).First(&existingRolePermission).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 分配权限
				rolePermission := RolePermission{
					RoleID:       adminRole.ID,
					PermissionID: permission.ID,
				}
				if err := db.Table("role_permissions").Create(&rolePermission).Error; err != nil {
					fmt.Printf("  ❌ 分配权限失败 %s: %v\n", permission.Name, err)
					continue
				}
				assignedCount++
				fmt.Printf("  ✅ 分配权限: %s (%s)\n", permission.Name, permission.Code)
			}
		}
	}
	fmt.Printf("✅ 成功分配 %d 个权限给admin角色\n", assignedCount)

	// 验证分配结果
	fmt.Println("\n🎯 验证admin用户的最终权限配置:")
	
	// 查询用户角色
	var userRoles []UserRole
	if err := db.Table("user_roles").Where("user_id = ?", adminUser.ID).Find(&userRoles).Error; err != nil {
		log.Fatal("查询用户角色失败:", err)
	}

	fmt.Printf("  📋 用户角色 (%d个):\n", len(userRoles))
	
	totalPermissions := 0
	for _, userRole := range userRoles {
		// 查询角色详情
		var role Role
		if err := db.Table("roles").Where("id = ?", userRole.RoleID).First(&role).Error; err != nil {
			continue
		}
		
		fmt.Printf("    - %s (%s) - 级别: %d\n", role.Name, role.Code, role.Level)
		
		// 查询角色权限数量
		var permissionCount int64
		if err := db.Table("role_permissions").Where("role_id = ?", role.ID).Count(&permissionCount).Error; err != nil {
			continue
		}
		
		fmt.Printf("      权限数量: %d\n", permissionCount)
		totalPermissions += int(permissionCount)
	}
	
	fmt.Printf("\n  🔐 用户总权限数量: %d\n", totalPermissions)
	
	// 检查角色管理权限
	fmt.Println("\n  🎯 角色管理相关权限:")
	var rolePermissions []Permission
	if err := db.Raw(`
		SELECT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND (p.resource LIKE '%role%' OR p.code LIKE '%role%')
	`, adminUser.ID).Scan(&rolePermissions).Error; err != nil {
		fmt.Printf("    ❌ 查询角色权限失败: %v\n", err)
	} else {
		for _, perm := range rolePermissions {
			fmt.Printf("    ✅ %s (%s) - 资源: %s, 动作: %s\n", 
				perm.Name, perm.Code, perm.Resource, perm.Action)
		}
	}

	fmt.Println("\n✅ admin账号权限配置完成!")
}