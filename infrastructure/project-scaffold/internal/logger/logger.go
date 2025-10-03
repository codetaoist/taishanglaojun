package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Sync() error
}

// zapLogger zap日志实现
type zapLogger struct {
	logger *zap.Logger
}

// New 创建新的日志实例
func New(level, format string) Logger {
	// 解析日志级别
	logLevel := parseLevel(level)

	// 创建编码器器配置�?
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建编码器�?
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// 创建logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &zapLogger{logger: logger}
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug 调试日志
func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, l.parseFields(fields...)...)
}

// Info 信息日志
func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, l.parseFields(fields...)...)
}

// Warn 警告日志
func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, l.parseFields(fields...)...)
}

// Error 错误日志
func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, l.parseFields(fields...)...)
}

// Fatal 致命错误日志
func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal(msg, l.parseFields(fields...)...)
}

// Sync 同步日志
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// parseFields 解析字段
func (l *zapLogger) parseFields(fields ...interface{}) []zap.Field {
	if len(fields)%2 != 0 {
		return []zap.Field{}
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}

