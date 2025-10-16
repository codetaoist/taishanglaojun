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

func main() {
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 模拟权限API的查询逻辑
	page := 1
	limit := 20
	search := ""

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var permissions []Permission
	var total int64

	query := db.Model(&Permission{})
	if search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		log.Fatal("Failed to count permissions:", err)
	}

	fmt.Printf("总权限数: %d\n", total)

	// 获取权限列表
	if err := query.Offset(offset).Limit(limit).Find(&permissions).Error; err != nil {
		log.Fatal("Failed to query permissions:", err)
	}

	fmt.Printf("查询到的权限数: %d\n", len(permissions))
	fmt.Printf("页码: %d, 限制: %d, 偏移: %d\n", page, limit, offset)
	fmt.Printf("总页数: %d\n", (total+int64(limit)-1)/int64(limit))

	// 显示前几个权限
	fmt.Println("\n查询到的权限:")
	for i, p := range permissions {
		if i >= 5 { // 只显示前5个
			break
		}
		fmt.Printf("ID: %d, Name: %s, Code: %s, Resource: %s, Action: %s\n", 
			p.ID, p.Name, p.Code, p.Resource, p.Action)
	}

	// 测试不同的查询条件
	fmt.Println("\n=== 测试直接查询所有权限 ===")
	var allPermissions []Permission
	if err := db.Find(&allPermissions).Error; err != nil {
		log.Fatal("Failed to query all permissions:", err)
	}
	fmt.Printf("直接查询所有权限数: %d\n", len(allPermissions))

	// 测试表是否存在
	fmt.Println("\n=== 检查表结构 ===")
	if db.Migrator().HasTable(&Permission{}) {
		fmt.Println("permissions表存在")
	} else {
		fmt.Println("permissions表不存在")
	}

	// 检查表名
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(&Permission{})
	fmt.Printf("表名: %s\n", stmt.Schema.Table)
}