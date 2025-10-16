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

	// 检查重复的name值
	var duplicateNames []struct {
		Name  string
		Count int64
	}
	
	if err := db.Model(&Permission{}).
		Select("name, COUNT(*) as count").
		Group("name").
		Having("COUNT(*) > 1").
		Find(&duplicateNames).Error; err != nil {
		log.Fatal("查询重复name失败:", err)
	}

	fmt.Printf("重复的name值:\n")
	for _, dup := range duplicateNames {
		fmt.Printf("Name: %s, Count: %d\n", dup.Name, dup.Count)
	}

	// 检查空的name值
	var emptyNameCount int64
	if err := db.Model(&Permission{}).Where("name = '' OR name IS NULL").Count(&emptyNameCount).Error; err != nil {
		log.Fatal("查询空name失败:", err)
	}
	fmt.Printf("空name值数量: %d\n", emptyNameCount)

	// 检查重复的code值
	var duplicateCodes []struct {
		Code  string
		Count int64
	}
	
	if err := db.Model(&Permission{}).
		Select("code, COUNT(*) as count").
		Group("code").
		Having("COUNT(*) > 1").
		Find(&duplicateCodes).Error; err != nil {
		log.Fatal("查询重复code失败:", err)
	}

	fmt.Printf("重复的code值:\n")
	for _, dup := range duplicateCodes {
		fmt.Printf("Code: %s, Count: %d\n", dup.Code, dup.Count)
	}

	// 检查空的code值
	var emptyCodeCount int64
	if err := db.Model(&Permission{}).Where("code = '' OR code IS NULL").Count(&emptyCodeCount).Error; err != nil {
		log.Fatal("查询空code失败:", err)
	}
	fmt.Printf("空code值数量: %d\n", emptyCodeCount)

	// 显示所有权限的name和code
	var permissions []Permission
	if err := db.Find(&permissions).Error; err != nil {
		log.Fatal("查询所有权限失败:", err)
	}

	fmt.Printf("\n所有权限的name和code:\n")
	for _, p := range permissions {
		fmt.Printf("ID: %d, Name: '%s', Code: '%s'\n", p.ID, p.Name, p.Code)
	}
}