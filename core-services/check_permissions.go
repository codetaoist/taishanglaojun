package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Permission struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func main() {
	// 数据库连接
	dsn := "root:123456@tcp(localhost:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 检查permissions表结构
	var permissions []Permission
	result := db.Limit(10).Find(&permissions)
	if result.Error != nil {
		log.Fatal("Failed to query permissions:", result.Error)
	}

	fmt.Printf("Found %d permissions:\n", len(permissions))
	for _, p := range permissions {
		fmt.Printf("ID: %d, Name: %s\n", p.ID, p.Name)
	}

	// 检查表结构
	var tableInfo []map[string]interface{}
	db.Raw("DESCRIBE permissions").Scan(&tableInfo)
	
	fmt.Println("\nTable structure:")
	for _, info := range tableInfo {
		fmt.Printf("%+v\n", info)
	}
}