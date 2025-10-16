package community

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutes 鱣
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger) error {
	// 
	module, err := NewModule(nil, db, nil, logger)
	if err != nil {
		logger.Error("Failed to create community module", zap.Error(err))
		return err
	}

	// 
	if err := module.Start(); err != nil {
		logger.Error("Failed to start community module", zap.Error(err))
		return err
	}

	// JWT
	if err := module.SetupRoutes(router, nil); err != nil {
		logger.Error("Failed to setup community routes", zap.Error(err))
		return err
	}

	logger.Info("Community service routes setup completed")
	return nil
}

