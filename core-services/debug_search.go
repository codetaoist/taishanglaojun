package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Role struct {
	ID          string       `json:"id" gorm:"primaryKey;type:char(36)"`
	Name        string       `json:"name" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Code        string       `json:"code" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Description string       `json:"description" gorm:"type:text"`
	Type        string       `json:"type" gorm:"default:custom;type:varchar(50)"`
	Level       int          `json:"level" gorm:"default:1"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	Status      string       `json:"status" gorm:"default:active;type:varchar(50)"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func main() {
	// 数据库连接字符串
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 启用SQL日志
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	fmt.Println("MySQL数据库连接成功!")

	// 测试搜索功能
	fmt.Println("\n🔍 测试搜索功能...")
	
	var roles []Role
	var total int64
	
	search := "admin"
	
	// 构建查询
	query := db.Model(&Role{})
	
	fmt.Printf("搜索关键词: %s\n", search)
	
	// 添加搜索条件
	if search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	
	// 获取总数
	fmt.Println("执行Count查询...")
	if err := query.Count(&total).Error; err != nil {
		fmt.Printf("❌ Count查询失败: %v\n", err)
		return
	}
	fmt.Printf("✅ Count查询成功，总数: %d\n", total)
	
	// 获取角色列表
	fmt.Println("执行Find查询...")
	if err := query.Limit(20).Find(&roles).Error; err != nil {
		fmt.Printf("❌ Find查询失败: %v\n", err)
		return
	}
	fmt.Printf("✅ Find查询成功，找到 %d 个角色\n", len(roles))
	
	// 显示结果
	if len(roles) > 0 {
		fmt.Println("\n角色列表:")
		for i, role := range roles {
			fmt.Printf("  %d. %s (%s) - %s - Active: %t\n", i+1, role.Name, role.Code, role.Type, role.IsActive)
		}
	}
	
	// 测试不同的搜索方式
	fmt.Println("\n🔍 测试LIKE查询...")
	var roles2 []Role
	if err := db.Model(&Role{}).Where("name LIKE ? OR code LIKE ?", "%admin%", "%admin%").Find(&roles2).Error; err != nil {
		fmt.Printf("❌ LIKE查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ LIKE查询成功，找到 %d 个角色\n", len(roles2))
	}
	
	fmt.Println("\n调试完成!")
}