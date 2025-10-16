package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接配置
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 删除有问题的表
	tables := []string{
		"role_permissions",
		"user_roles", 
		"permissions",
		"roles",
	}

	for _, table := range tables {
		// 删除表（PostgreSQL语法）
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("Successfully dropped table: %s", table)
		}
	}

	log.Println("All permission-related tables have been dropped successfully")
}