package community

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutes 设置社区模块路由（全局函数，用于与其他模块保持一致的接口）
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger) error {
	// 创建默认配置的模块
	module, err := NewModule(nil, db, nil, logger)
	if err != nil {
		logger.Error("Failed to create community module", zap.Error(err))
		return err
	}

	// 启动模块
	if err := module.Start(); err != nil {
		logger.Error("Failed to start community module", zap.Error(err))
		return err
	}

	// 设置路由（暂时不使用JWT中间件）
	if err := module.SetupRoutes(router, nil); err != nil {
		logger.Error("Failed to setup community routes", zap.Error(err))
		return err
	}

	logger.Info("Community service routes setup completed")
	return nil
}