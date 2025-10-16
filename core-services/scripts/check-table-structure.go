package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 查询permissions表结构
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		IsNullable string `gorm:"column:is_nullable"`
	}

	err = db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'permissions' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error

	if err != nil {
		log.Fatal("查询表结构失败:", err)
	}

	fmt.Println("permissions表结构:")
	for _, col := range columns {
		fmt.Printf("列名: %s, 类型: %s, 可空: %s\n", col.ColumnName, col.DataType, col.IsNullable)
	}

	// 查询一条记录看看实际数据
	var result map[string]interface{}
	err = db.Raw("SELECT * FROM permissions LIMIT 1").Scan(&result).Error
	if err != nil {
		log.Fatal("查询数据失败:", err)
	}

	fmt.Println("\n第一条记录的数据:")
	for key, value := range result {
		fmt.Printf("%s: %v\n", key, value)
	}
}