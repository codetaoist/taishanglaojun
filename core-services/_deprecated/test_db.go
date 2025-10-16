package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// PostgreSQL 连接配置
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	
	fmt.Println("正在连接数据库...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	
	fmt.Println("数据库连接成功!")
	
	// 测试查询
	var result struct {
		Version string
	}
	
	err = db.Raw("SELECT version() as version").Scan(&result).Error
	if err != nil {
		log.Fatalf("查询数据库版本失败: %v", err)
	}
	
	fmt.Printf("数据库版本: %s\n", result.Version)
	
	// 检查表是否存在
	var tableExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'permissions')").Scan(&tableExists).Error
	if err != nil {
		log.Fatalf("检查表是否存在失败: %v", err)
	}
	
	fmt.Printf("permissions表是否存在: %v\n", tableExists)
	
	if tableExists {
		// 检查permissions表结构
		var columnInfo []struct {
			ColumnName string `gorm:"column:column_name"`
			DataType   string `gorm:"column:data_type"`
		}
		
		err = db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'permissions' ORDER BY ordinal_position").Scan(&columnInfo).Error
		if err != nil {
			log.Fatalf("查询表结构失败: %v", err)
		}
		
		fmt.Println("permissions表结构:")
		for _, col := range columnInfo {
			fmt.Printf("  %s: %s\n", col.ColumnName, col.DataType)
		}
	}
	
	fmt.Println("数据库测试完成!")
}