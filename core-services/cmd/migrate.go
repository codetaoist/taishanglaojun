package main

import (
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	cultural_models "github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
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

	// 迁移文化智慧服务相关表
	if err := db.AutoMigrate(
		&cultural_models.CulturalWisdom{},
		&cultural_models.Category{},
		&cultural_models.WisdomMetadata{},
		&cultural_models.UserBehavior{},
		&cultural_models.UserPreference{},
		&cultural_models.UserInteraction{},
		&cultural_models.RecommendationLog{},
		&cultural_models.UserSimilarity{},
		&cultural_models.WisdomFavorite{},
		&cultural_models.WisdomNote{},
	); err != nil {
		return err
	}

	// 迁移认证相关表
	if err := db.AutoMigrate(
		&middleware.User{},
	); err != nil {
		return err
	}

	return nil
}

