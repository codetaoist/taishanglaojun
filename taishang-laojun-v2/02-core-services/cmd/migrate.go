package main

import (
	"github.com/taishanglaojun/core-services/ai-integration/models"
	"gorm.io/gorm"
)

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	// 迁移AI集成服务相关表
	if err := db.AutoMigrate(
		&models.ChatSession{},
		&models.ChatMessage{},
	); err != nil {
		return err
	}

	// TODO: 迁移文化智慧服务相关表

	return nil
}
