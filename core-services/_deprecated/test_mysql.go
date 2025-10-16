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
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	
	fmt.Println("MySQL数据库连接成功!")
	
	// 测试查询
	var result struct {
		Version string
	}
	
	err = db.Raw("SELECT VERSION() as version").Scan(&result).Error
	if err != nil {
		log.Fatalf("查询数据库版本失败: %v", err)
	}
	
	fmt.Printf("数据库版本: %s\n", result.Version)
	
	// 检查表是否存在
	var tableExists bool
	err = db.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = 'laojun' AND table_name = 'permissions'").Scan(&tableExists).Error
	if err != nil {
		log.Fatalf("检查表是否存在失败: %v", err)
	}
	
	fmt.Printf("permissions表是否存在: %v\n", tableExists)
	
	if tableExists {
		// 检查permissions表结构
		var columnInfo []struct {
			Field string `gorm:"column:Field"`
			Type  string `gorm:"column:Type"`
		}
		
		err = db.Raw("DESCRIBE permissions").Scan(&columnInfo).Error
		if err != nil {
			log.Fatalf("查询表结构失败: %v", err)
		}
		
		fmt.Println("permissions表结构:")
		for _, col := range columnInfo {
			fmt.Printf("  %s: %s\n", col.Field, col.Type)
		}
	}
	
	fmt.Println("MySQL数据库测试完成!")
}