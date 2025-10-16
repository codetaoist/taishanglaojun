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
	
	// 检查roles表是否存在
	var tableExists bool
	err = db.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = 'laojun' AND table_name = 'roles'").Scan(&tableExists).Error
	if err != nil {
		log.Fatalf("检查表是否存在失败: %v", err)
	}
	
	fmt.Printf("roles表是否存在: %v\n", tableExists)
	
	if tableExists {
		// 检查roles表结构
		var columnInfo []struct {
			Field   string `gorm:"column:Field"`
			Type    string `gorm:"column:Type"`
			Null    string `gorm:"column:Null"`
			Key     string `gorm:"column:Key"`
			Default *string `gorm:"column:Default"`
			Extra   string `gorm:"column:Extra"`
		}
		
		err = db.Raw("DESCRIBE roles").Scan(&columnInfo).Error
		if err != nil {
			log.Fatalf("查询表结构失败: %v", err)
		}
		
		fmt.Println("roles表结构:")
		for _, col := range columnInfo {
			defaultVal := "NULL"
			if col.Default != nil {
				defaultVal = *col.Default
			}
			fmt.Printf("  %s: %s, Null: %s, Key: %s, Default: %s, Extra: %s\n", 
				col.Field, col.Type, col.Null, col.Key, defaultVal, col.Extra)
		}
	}
	
	// 检查permissions表结构
	err = db.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = 'laojun' AND table_name = 'permissions'").Scan(&tableExists).Error
	if err != nil {
		log.Fatalf("检查permissions表是否存在失败: %v", err)
	}
	
	fmt.Printf("permissions表是否存在: %v\n", tableExists)
	
	if tableExists {
		// 检查permissions表结构
		var columnInfo []struct {
			Field   string `gorm:"column:Field"`
			Type    string `gorm:"column:Type"`
			Null    string `gorm:"column:Null"`
			Key     string `gorm:"column:Key"`
			Default *string `gorm:"column:Default"`
			Extra   string `gorm:"column:Extra"`
		}
		
		err = db.Raw("DESCRIBE permissions").Scan(&columnInfo).Error
		if err != nil {
			log.Fatalf("查询表结构失败: %v", err)
		}
		
		fmt.Println("permissions表结构:")
		for _, col := range columnInfo {
			defaultVal := "NULL"
			if col.Default != nil {
				defaultVal = *col.Default
			}
			fmt.Printf("  %s: %s, Null: %s, Key: %s, Default: %s, Extra: %s\n", 
				col.Field, col.Type, col.Null, col.Key, defaultVal, col.Extra)
		}
	}
	
	fmt.Println("数据库表结构检查完成!")
}