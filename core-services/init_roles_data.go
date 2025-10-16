package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          string    `json:"id" gorm:"primaryKey;type:char(36)"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Description string    `json:"description" gorm:"type:text"`
	Type        string    `json:"type" gorm:"default:custom;type:varchar(50)"` // system, custom
	Level       int       `json:"level" gorm:"default:1"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Status      string    `json:"status" gorm:"default:active;type:varchar(50)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func main() {
	// MySQL 连接配置
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	
	fmt.Println("MySQL数据库连接成功!")
	
	// 自动迁移表结构
	fmt.Println("正在迁移表结构...")
	err = db.AutoMigrate(&Role{})
	if err != nil {
		log.Fatalf("迁移表结构失败: %v", err)
	}
	
	// 清空现有数据
	fmt.Println("正在清空现有角色数据...")
	db.Exec("DELETE FROM roles")
	
	// 初始化角色数据
	roles := []Role{
		{
			ID:          uuid.New().String(),
			Name:        "Super Administrator",
			Code:        "super_admin",
			Description: "系统超级管理员，拥有所有权限",
			Type:        "system",
			Level:       10,
			IsActive:    true,
			Status:      "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Administrator",
			Code:        "admin",
			Description: "系统管理员，拥有大部分管理权限",
			Type:        "system",
			Level:       8,
			IsActive:    true,
			Status:      "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Manager",
			Code:        "manager",
			Description: "部门经理，拥有部门管理权限",
			Type:        "custom",
			Level:       6,
			IsActive:    true,
			Status:      "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Editor",
			Code:        "editor",
			Description: "内容编辑员，拥有内容编辑权限",
			Type:        "custom",
			Level:       4,
			IsActive:    true,
			Status:      "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Viewer",
			Code:        "viewer",
			Description: "查看者，只有查看权限",
			Type:        "custom",
			Level:       2,
			IsActive:    true,
			Status:      "active",
		},
		{
			ID:          uuid.New().String(),
			Name:        "测试角色",
			Code:        "test_role",
			Description: "用于测试的角色",
			Type:        "custom",
			Level:       1,
			IsActive:    false,
			Status:      "inactive",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Guest",
			Code:        "guest",
			Description: "访客角色，最低权限",
			Type:        "system",
			Level:       1,
			IsActive:    true,
			Status:      "active",
		},
	}
	
	fmt.Println("正在插入角色数据...")
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			log.Printf("插入角色失败 %s: %v", role.Name, err)
		} else {
			fmt.Printf("成功插入角色: %s (%s)\n", role.Name, role.Code)
		}
	}
	
	// 验证数据
	var count int64
	db.Model(&Role{}).Count(&count)
	fmt.Printf("角色数据初始化完成，共插入 %d 条记录\n", count)
	
	// 显示所有角色
	var allRoles []Role
	db.Find(&allRoles)
	
	fmt.Println("\n当前角色列表:")
	fmt.Println("ID\t\t\t\t\t名称\t\t代码\t\t类型\t级别\t状态")
	fmt.Println("------------------------------------------------------------------------------------")
	for _, role := range allRoles {
		status := "启用"
		if !role.IsActive {
			status = "禁用"
		}
		fmt.Printf("%s\t%s\t\t%s\t\t%s\t%d\t%s\n", 
			role.ID[:8]+"...", role.Name, role.Code, role.Type, role.Level, status)
	}
}