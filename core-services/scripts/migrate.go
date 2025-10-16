package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

func main() {
	// 数据库连接配置
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	
	fmt.Println("MySQL数据库连接成功!")

	// 创建logger
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("创建logger失败: %v", err)
	}

	// 创建迁移服务
	migrationService := database.NewMigrationService(db, zapLogger)

	// 运行迁移
	fmt.Println("开始运行数据库迁移...")
	if err := migrationService.RunMigration(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("数据库迁移完成!")

	// 验证API文档相关表是否创建成功
	fmt.Println("验证API文档相关表...")
	
	// 检查api_categories表
	if db.Migrator().HasTable(&models.APICategory{}) {
		fmt.Println("✓ api_categories 表创建成功")
	} else {
		fmt.Println("✗ api_categories 表创建失败")
	}

	// 检查api_endpoints表
	if db.Migrator().HasTable(&models.APIEndpoint{}) {
		fmt.Println("✓ api_endpoints 表创建成功")
	} else {
		fmt.Println("✗ api_endpoints 表创建失败")
	}

	// 检查api_documentation_sources表
	if db.Migrator().HasTable(&models.APIDocumentationSource{}) {
		fmt.Println("✓ api_documentation_sources 表创建成功")
	} else {
		fmt.Println("✗ api_documentation_sources 表创建失败")
	}

	// 检查api_test_records表
	if db.Migrator().HasTable(&models.APITestRecord{}) {
		fmt.Println("✓ api_test_records 表创建成功")
	} else {
		fmt.Println("✗ api_test_records 表创建失败")
	}

	// 检查api_change_logs表
	if db.Migrator().HasTable(&models.APIChangeLog{}) {
		fmt.Println("✓ api_change_logs 表创建成功")
	} else {
		fmt.Println("✗ api_change_logs 表创建失败")
	}

	fmt.Println("数据库迁移验证完成!")
}