package models

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	
	// 带字段的日志方法
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	
	// 带错误的日志方法
	WithError(err error) Logger
}

