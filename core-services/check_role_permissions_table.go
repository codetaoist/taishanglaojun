package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接字符串
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	fmt.Println("MySQL数据库连接成功!")

	// 检查role_permissions表结构
	fmt.Println("\n🔍 检查role_permissions表结构:")
	var columns []struct {
		Field   string `gorm:"column:Field"`
		Type    string `gorm:"column:Type"`
		Null    string `gorm:"column:Null"`
		Key     string `gorm:"column:Key"`
		Default *string `gorm:"column:Default"`
		Extra   string `gorm:"column:Extra"`
	}
	
	if err := db.Raw("DESCRIBE role_permissions").Scan(&columns).Error; err != nil {
		fmt.Printf("查询role_permissions表结构失败: %v\n", err)
		
		// 检查表是否存在
		var tableExists int
		if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'laojun' AND table_name = 'role_permissions'").Scan(&tableExists).Error; err != nil {
			fmt.Printf("检查表存在性失败: %v\n", err)
		} else if tableExists == 0 {
			fmt.Println("❌ role_permissions表不存在!")
			
			// 创建role_permissions表
			fmt.Println("🔧 正在创建role_permissions表...")
			createTableSQL := `
			CREATE TABLE role_permissions (
				role_id VARCHAR(36) NOT NULL,
				permission_id VARCHAR(36) NOT NULL,
				created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
				updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
				PRIMARY KEY (role_id, permission_id),
				INDEX idx_role_permissions_role_id (role_id),
				INDEX idx_role_permissions_permission_id (permission_id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
			`
			if err := db.Exec(createTableSQL).Error; err != nil {
				log.Fatal("创建role_permissions表失败:", err)
			}
			fmt.Println("✅ role_permissions表创建成功!")
		}
	} else {
		fmt.Printf("字段结构:\n")
		for _, col := range columns {
			defaultVal := "NULL"
			if col.Default != nil {
				defaultVal = *col.Default
			}
			fmt.Printf("  %s (%s) - Null:%s, Key:%s, Default:%s\n", 
				col.Field, col.Type, col.Null, col.Key, defaultVal)
		}
	}

	// 检查user_roles表结构
	fmt.Println("\n🔍 检查user_roles表结构:")
	if err := db.Raw("DESCRIBE user_roles").Scan(&columns).Error; err != nil {
		fmt.Printf("查询user_roles表结构失败: %v\n", err)
		
		// 检查表是否存在
		var tableExists int
		if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'laojun' AND table_name = 'user_roles'").Scan(&tableExists).Error; err != nil {
			fmt.Printf("检查表存在性失败: %v\n", err)
		} else if tableExists == 0 {
			fmt.Println("❌ user_roles表不存在!")
			
			// 创建user_roles表
			fmt.Println("🔧 正在创建user_roles表...")
			createTableSQL := `
			CREATE TABLE user_roles (
				id VARCHAR(36) NOT NULL PRIMARY KEY,
				user_id VARCHAR(36) NOT NULL,
				role_id VARCHAR(36) NOT NULL,
				created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
				updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
				INDEX idx_user_roles_user_id (user_id),
				INDEX idx_user_roles_role_id (role_id),
				UNIQUE KEY unique_user_role (user_id, role_id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
			`
			if err := db.Exec(createTableSQL).Error; err != nil {
				log.Fatal("创建user_roles表失败:", err)
			}
			fmt.Println("✅ user_roles表创建成功!")
		}
	} else {
		fmt.Printf("字段结构:\n")
		for _, col := range columns {
			defaultVal := "NULL"
			if col.Default != nil {
				defaultVal = *col.Default
			}
			fmt.Printf("  %s (%s) - Null:%s, Key:%s, Default:%s\n", 
				col.Field, col.Type, col.Null, col.Key, defaultVal)
		}
	}

	// 查看现有数据
	fmt.Println("\n📋 检查现有数据:")
	
	var userRoleCount int64
	if err := db.Raw("SELECT COUNT(*) FROM user_roles").Scan(&userRoleCount).Error; err != nil {
		fmt.Printf("查询user_roles数据失败: %v\n", err)
	} else {
		fmt.Printf("user_roles表记录数: %d\n", userRoleCount)
	}
	
	var rolePermissionCount int64
	if err := db.Raw("SELECT COUNT(*) FROM role_permissions").Scan(&rolePermissionCount).Error; err != nil {
		fmt.Printf("查询role_permissions数据失败: %v\n", err)
	} else {
		fmt.Printf("role_permissions表记录数: %d\n", rolePermissionCount)
	}
}