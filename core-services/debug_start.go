package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("=== 调试启动脚本 ===")
	
	// 1. 加载配置
	fmt.Println("1. 加载配置...")
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Printf("配置加载成功: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	
	// 2. 初始化日志
	fmt.Println("2. 初始化日志...")
	logConfig := logger.LogConfig{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		Filename:   cfg.Logger.Filename,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	
	log, err := logger.New(logConfig)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()
	fmt.Println("日志初始化成功")
	
	// 3. 连接数据库
	fmt.Println("3. 连接数据库...")
	dbConfig := database.Config{
		Driver:          "postgres",
		Host:            "localhost",
		Port:            5432,
		Database:        "taishanglaojun",
		Username:        "postgres",
		Password:        "password",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		SSLMode:         "disable",
		ConnectTimeout:  30 * time.Second,
	}
	
	db, err := database.New(dbConfig, log)
	if err != nil {
		log.Fatal("连接数据库失败", zap.Error(err))
	}
	fmt.Println("数据库连接成功")
	
	// 4. 运行数据库迁移
	fmt.Println("4. 运行数据库迁移...")
	migrationService := database.NewMigrationService(db.GetDB(), log)
	if err := migrationService.RunMigration(); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}
	fmt.Println("数据库迁移完成")
	
	fmt.Println("=== 所有初始化步骤完成 ===")
	fmt.Println("现在可以启动完整服务...")
}