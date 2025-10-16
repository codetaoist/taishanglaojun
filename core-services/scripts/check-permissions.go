package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Permission struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Code        string `json:"code" gorm:"uniqueIndex"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

type Role struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Code        string `json:"code" gorm:"uniqueIndex"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	Status      string `json:"status"`
}

func main() {
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 检查权限数量
	var permissionCount int64
	db.Model(&Permission{}).Count(&permissionCount)
	fmt.Printf("权限表中的记录数: %d\n", permissionCount)

	// 检查角色数量
	var roleCount int64
	db.Model(&Role{}).Count(&roleCount)
	fmt.Printf("角色表中的记录数: %d\n", roleCount)

	// 显示前几个权限
	var permissions []Permission
	db.Limit(10).Find(&permissions)
	fmt.Println("\n前10个权限:")
	for _, p := range permissions {
		fmt.Printf("ID: %d, Name: %s, Code: %s, Resource: %s, Action: %s\n", 
			p.ID, p.Name, p.Code, p.Resource, p.Action)
	}

	// 显示前几个角色
	var roles []Role
	db.Limit(10).Find(&roles)
	fmt.Println("\n前10个角色:")
	for _, r := range roles {
		fmt.Printf("ID: %d, Name: %s, Code: %s, Level: %d, Status: %s\n", 
			r.ID, r.Name, r.Code, r.Level, r.Status)
	}
}