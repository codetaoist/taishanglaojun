package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/config"
)

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// logrusLogger logrus实现
type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// New 创建新的日志实例
func New(cfg config.LogConfig) Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		// 确保日志目录存在
		logDir := filepath.Dir("logs/gateway.log")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logrus.Errorf("Failed to create log directory: %v", err)
			output = os.Stdout
		} else {
			output = &lumberjack.Logger{
				Filename:   "logs/gateway.log",
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}
		}
	default:
		output = os.Stdout
	}

	logger.SetOutput(output)

	return &logrusLogger{
		logger: logger,
		entry:  logger.WithFields(logrus.Fields{}),
	}
}

// Debug 调试日志
func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Debugf 格式化调试日志
func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Info 信息日志
func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Infof 格式化信息日志
func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warn 警告日志
func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Warnf 格式化警告日志
func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Error 错误日志
func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Errorf 格式化错误日志
func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatal 致命错误日志
func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// WithField 添加字段
func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

// WithFields 添加多个字段
func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}
	
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(logrusFields),
	}
}

// NewNop 创建空日志实例（用于测试）
func NewNop() Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	
	return &logrusLogger{
		logger: logger,
		entry:  logger.WithFields(logrus.Fields{}),
	}
}

// NewDevelopment 创建开发环境日志实例
func NewDevelopment() Logger {
	cfg := config.LogConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	return New(cfg)
}

// NewProduction 创建生产环境日志实例
func NewProduction() Logger {
	cfg := config.LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
	return New(cfg)
}