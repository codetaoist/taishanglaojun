package models

import (
	"time"
)

// ConversationExport 单个对话导出结构
type ConversationExport struct {
	Conversation Conversation `json:"conversation"`
	Messages     []Message    `json:"messages"`
	ExportedAt   time.Time    `json:"exported_at"`
}

// ConversationsExport 多个对话导出结构
type ConversationsExport struct {
	UserID       string               `json:"user_id"`
	Conversations []ConversationExport `json:"conversations"`
	ExportedAt   time.Time            `json:"exported_at"`
}

// ExportFormat 导出格式枚举
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatTXT  ExportFormat = "txt"
)

// ExportRequest 导出请求
type ExportRequest struct {
	ConversationIDs []string     `json:"conversation_ids,omitempty"` // 可选，如果为空则导出所有对话
	Format         ExportFormat `json:"format"`                      // 导出格式
	IncludeMetadata bool        `json:"include_metadata"`           // 是否包含元数据
}

// ExportResponse 导出响应
type ExportResponse struct {
	Data        []byte    `json:"data"`         // 导出数据
	Format      ExportFormat `json:"format"`     // 导出格式
	Filename    string    `json:"filename"`     // 建议的文件名
	ExportedAt  time.Time `json:"exported_at"`  // 导出时间
	Size        int64     `json:"size"`         // 数据大小(字节)
}

// ImportRequest 导入请求
type ImportRequest struct {
	Data        []byte      `json:"data"`         // 导入数据
	Format      ExportFormat `json:"format"`      // 导入格式
	ReplaceExisting bool     `json:"replace_existing"` // 是否替换已存在的对话
}

// ImportResponse 导入响应
type ImportResponse struct {
	ImportedCount int        `json:"imported_count"` // 导入的对话数量
	SkippedCount  int        `json:"skipped_count"`  // 跳过的对话数量
	Errors        []string   `json:"errors"`         // 错误信息
	ImportedIDs   []string   `json:"imported_ids"`   // 导入的对话ID列表
	ImportedAt    time.Time  `json:"imported_at"`    // 导入时间
}

// ImportConflict 导入冲突处理策略
type ImportConflict string

const (
	ImportConflictSkip      ImportConflict = "skip"       // 跳过已存在的对话
	ImportConflictReplace   ImportConflict = "replace"    // 替换已存在的对话
	ImportConflictRename    ImportConflict = "rename"     // 重命名已存在的对话
	ImportConflictMerge     ImportConflict = "merge"      // 合并到已存在的对话
)

// ImportOptions 导入选项
type ImportOptions struct {
	ConflictStrategy ImportConflict `json:"conflict_strategy"` // 冲突处理策略
	PreserveIDs      bool           `json:"preserve_ids"`       // 是否保留原始ID
	AssignToUser     string         `json:"assign_to_user"`     // 分配给指定用户ID
}