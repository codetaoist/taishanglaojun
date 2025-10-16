package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 使用与权限处理器完全相同的模型定义
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func main() {
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 完全模拟权限API的逻辑
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

	// 显示API响应格式
	fmt.Printf("API响应:\n")
	fmt.Printf("{\n")
	fmt.Printf("  \"permissions\": [\n")
	for i, p := range permissions {
		if i > 0 {
			fmt.Printf(",\n")
		}
		fmt.Printf("    {\n")
		fmt.Printf("      \"id\": %d,\n", p.ID)
		fmt.Printf("      \"name\": \"%s\",\n", p.Name)
		fmt.Printf("      \"code\": \"%s\",\n", p.Code)
		fmt.Printf("      \"description\": \"%s\",\n", p.Description)
		fmt.Printf("      \"resource\": \"%s\",\n", p.Resource)
		fmt.Printf("      \"action\": \"%s\",\n", p.Action)
		fmt.Printf("      \"created_at\": \"%s\",\n", p.CreatedAt.Format(time.RFC3339))
		fmt.Printf("      \"updated_at\": \"%s\"\n", p.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("    }")
		if i >= 2 { // 只显示前3个
			fmt.Printf(",\n    ...")
			break
		}
	}
	fmt.Printf("\n  ],\n")
	fmt.Printf("  \"total\": %d,\n", total)
	fmt.Printf("  \"page\": %d,\n", page)
	fmt.Printf("  \"limit\": %d,\n", limit)
	fmt.Printf("  \"pages\": %d\n", (total+int64(limit)-1)/int64(limit))
	fmt.Printf("}\n")

	// 检查表名映射
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(&Permission{})
	fmt.Printf("\n表名映射: %s\n", stmt.Schema.Table)

	// 检查是否有数据库连接问题
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取SQL DB失败:", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("数据库ping失败:", err)
	}
	
	fmt.Println("数据库连接正常")
}