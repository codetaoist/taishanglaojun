package logger

import (
	"os"
	"path/filepath"

	"github.com/taishanglaojun/auth_system/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New 创建日志记录器
func New(cfg *config.Config) (*zap.Logger, error) {
	// 配置日志级别
	level, err := zapcore.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 配置编码器
	var encoderConfig zapcore.EncoderConfig
	if cfg.Log.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Log.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Log.Output {
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	case "file":
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Log.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		// 配置日志轮转
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.Log.Filename,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建日志记录器
	logger := zap.New(core)

	// 在开发环境添加调用者信息
	if cfg.IsDevelopment() {
		logger = logger.WithOptions(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 添加全局字段
	logger = logger.With(
		zap.String("service", "auth-system"),
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.Server.Mode),
	)

	return logger, nil
}

// NewNop 创建无操作日志记录器（用于测试）
func NewNop() *zap.Logger {
	return zap.NewNop()
}

// NewDevelopment 创建开发环境日志记录器
func NewDevelopment() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// NewProduction 创建生产环境日志记录器
func NewProduction() (*zap.Logger, error) {
	return zap.NewProduction()
}

// Sync 同步日志缓冲区
func Sync(logger *zap.Logger) {
	if logger != nil {
		_ = logger.Sync()
	}
}