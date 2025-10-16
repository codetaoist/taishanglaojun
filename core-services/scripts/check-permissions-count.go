package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Permission struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:100;not null"`
	Code        string `gorm:"size:100;uniqueIndex;not null"`
	Description string `gorm:"size:255"`
	Resource    string `gorm:"size:100;not null"`
	Action      string `gorm:"size:50;not null"`
	CreatedAt   string
	UpdatedAt   string
}

func main() {
	// 数据库连接
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 检查权限数量
	var count int64
	if err := db.Model(&Permission{}).Count(&count).Error; err != nil {
		log.Fatal("查询权限数量失败:", err)
	}

	fmt.Printf("数据库中权限数量: %d\n", count)

	// 获取前5条权限数据
	var permissions []Permission
	if err := db.Limit(5).Find(&permissions).Error; err != nil {
		log.Fatal("查询权限数据失败:", err)
	}

	fmt.Printf("前5条权限数据:\n")
	for _, p := range permissions {
		fmt.Printf("ID: %d, Code: %s, Name: %s, Resource: %s, Action: %s\n", 
			p.ID, p.Code, p.Name, p.Resource, p.Action)
	}
}