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

	// 查询所有表名
	var tables []string
	if err := db.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		log.Fatal("查询表名失败:", err)
	}

	fmt.Println("\n📋 数据库中的所有表:")
	for i, table := range tables {
		fmt.Printf("  %d. %s\n", i+1, table)
	}

	// 检查角色相关表
	roleRelatedTables := []string{}
	for _, table := range tables {
		if contains(table, "role") || contains(table, "permission") || contains(table, "user") {
			roleRelatedTables = append(roleRelatedTables, table)
		}
	}

	fmt.Println("\n🔍 角色/权限相关表:")
	for _, table := range roleRelatedTables {
		fmt.Printf("  - %s\n", table)
		
		// 查询表结构
		var columns []struct {
			Field   string `gorm:"column:Field"`
			Type    string `gorm:"column:Type"`
			Null    string `gorm:"column:Null"`
			Key     string `gorm:"column:Key"`
			Default *string `gorm:"column:Default"`
			Extra   string `gorm:"column:Extra"`
		}
		
		if err := db.Raw(fmt.Sprintf("DESCRIBE %s", table)).Scan(&columns).Error; err != nil {
			fmt.Printf("    查询表结构失败: %v\n", err)
			continue
		}
		
		fmt.Printf("    字段结构:\n")
		for _, col := range columns {
			defaultVal := "NULL"
			if col.Default != nil {
				defaultVal = *col.Default
			}
			fmt.Printf("      %s (%s) - Null:%s, Key:%s, Default:%s\n", 
				col.Field, col.Type, col.Null, col.Key, defaultVal)
		}
		fmt.Println()
	}

	// 查看roles表的具体数据
	fmt.Println("🎯 roles表的当前数据:")
	var roles []map[string]interface{}
	if err := db.Raw("SELECT * FROM roles LIMIT 5").Scan(&roles).Error; err != nil {
		fmt.Printf("查询roles表数据失败: %v\n", err)
	} else {
		for i, role := range roles {
			fmt.Printf("  角色 %d:\n", i+1)
			for key, value := range role {
				fmt.Printf("    %s: %v\n", key, value)
			}
			fmt.Println()
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}