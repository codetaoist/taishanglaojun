package main

import (
	"fmt"
	"log"
	"strings"

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
	ID     string `gorm:"column:id;primaryKey"`
	UserID string `gorm:"column:user_id"`
	RoleID string `gorm:"column:role_id"`
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
	ID           string `gorm:"column:id;primaryKey"`
	RoleID       string `gorm:"column:role_id"`
	PermissionID string `gorm:"column:permission_id"`
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
	var adminUsers []User
	if err := db.Table("users").Where("username LIKE ? OR email LIKE ?", "%admin%", "%admin%").Find(&adminUsers).Error; err != nil {
		log.Fatal("查询admin用户失败:", err)
	}

	fmt.Printf("\n👤 找到 %d 个admin相关用户:\n", len(adminUsers))
	for i, user := range adminUsers {
		fmt.Printf("  %d. ID: %s\n", i+1, user.ID)
		fmt.Printf("     用户名: %s\n", user.Username)
		fmt.Printf("     邮箱: %s\n", user.Email)
		fmt.Printf("     状态: %s, 激活: %t\n\n", user.Status, user.IsActive)
	}

	// 对每个admin用户检查其角色和权限
	for _, user := range adminUsers {
		fmt.Printf("🔍 检查用户 %s 的权限配置:\n", user.Username)
		
		// 查询用户角色
		var userRoles []UserRole
		if err := db.Table("user_roles").Where("user_id = ?", user.ID).Find(&userRoles).Error; err != nil {
			fmt.Printf("  ❌ 查询用户角色失败: %v\n", err)
			continue
		}

		fmt.Printf("  📋 用户角色 (%d个):\n", len(userRoles))
		
		var allPermissions []Permission
		
		for _, userRole := range userRoles {
			// 查询角色详情
			var role Role
			if err := db.Table("roles").Where("id = ?", userRole.RoleID).First(&role).Error; err != nil {
				fmt.Printf("    ❌ 查询角色详情失败: %v\n", err)
				continue
			}
			
			fmt.Printf("    - %s (%s) - 级别: %d, 状态: %s\n", 
				role.Name, role.Code, role.Level, role.Status)
			
			// 查询角色权限
			var rolePermissions []RolePermission
			if err := db.Table("role_permissions").Where("role_id = ?", role.ID).Find(&rolePermissions).Error; err != nil {
				fmt.Printf("      ❌ 查询角色权限失败: %v\n", err)
				continue
			}
			
			fmt.Printf("      权限数量: %d\n", len(rolePermissions))
			
			// 获取权限详情
			for _, rp := range rolePermissions {
				var permission Permission
				if err := db.Table("permissions").Where("id = ?", rp.PermissionID).First(&permission).Error; err != nil {
					continue
				}
				allPermissions = append(allPermissions, permission)
			}
		}
		
		// 显示所有权限
		fmt.Printf("\n  🔐 用户总权限 (%d个):\n", len(allPermissions))
		
		// 按资源分组显示权限
		resourcePermissions := make(map[string][]Permission)
		for _, perm := range allPermissions {
			resourcePermissions[perm.Resource] = append(resourcePermissions[perm.Resource], perm)
		}
		
		for resource, perms := range resourcePermissions {
			fmt.Printf("    📁 %s:\n", resource)
			for _, perm := range perms {
				fmt.Printf("      - %s (%s) - 动作: %s\n", 
					perm.Name, perm.Code, perm.Action)
			}
		}
		
		// 检查是否有角色管理相关权限
		fmt.Printf("\n  🎯 角色管理相关权限:\n")
		hasRolePermissions := false
		for _, perm := range allPermissions {
			if contains(perm.Resource, "role") || contains(perm.Code, "role") || 
			   contains(perm.Name, "角色") || contains(perm.Name, "role") {
				fmt.Printf("    ✅ %s (%s) - 资源: %s, 动作: %s\n", 
					perm.Name, perm.Code, perm.Resource, perm.Action)
				hasRolePermissions = true
			}
		}
		
		if !hasRolePermissions {
			fmt.Printf("    ❌ 未找到角色管理相关权限!\n")
		}
		
		fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
	}

	// 检查所有角色管理相关权限
	fmt.Println("🔍 系统中所有角色管理相关权限:")
	var allSystemPermissions []Permission
	if err := db.Table("permissions").Find(&allSystemPermissions).Error; err != nil {
		log.Fatal("查询系统权限失败:", err)
	}

	roleRelatedPerms := []Permission{}
	for _, perm := range allSystemPermissions {
		if contains(perm.Resource, "role") || contains(perm.Code, "role") || 
		   contains(perm.Name, "角色") || contains(perm.Name, "role") {
			roleRelatedPerms = append(roleRelatedPerms, perm)
		}
	}

	fmt.Printf("找到 %d 个角色相关权限:\n", len(roleRelatedPerms))
	for i, perm := range roleRelatedPerms {
		fmt.Printf("  %d. %s (%s)\n", i+1, perm.Name, perm.Code)
		fmt.Printf("     资源: %s, 动作: %s, 状态: %s\n", 
			perm.Resource, perm.Action, perm.Status)
	}
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}