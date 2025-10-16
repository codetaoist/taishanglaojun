package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// MySQL 连接配置 (从config.yaml获取)
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 删除有问题的表 (注意顺序，先删除有外键的表)
	tables := []string{
		"role_permissions",
		"user_roles", 
		"permissions",
		"roles",
	}

	for _, table := range tables {
		// 删除表（MySQL语法）
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			log.Printf("Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("Successfully dropped table: %s", table)
		}
	}

	log.Println("All permission-related tables have been dropped successfully")
}