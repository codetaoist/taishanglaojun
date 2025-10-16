package models

import (
	"time"
)

// APICategory API分类
type APICategory struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null;index"`
	Code        string    `json:"code" gorm:"type:varchar(50);not null;unique;index"`
	Description string    `json:"description" gorm:"type:text"`
	Icon        string    `json:"icon" gorm:"type:varchar(100)"`
	Color       string    `json:"color" gorm:"type:varchar(20)"`
	SortOrder   int       `json:"sort_order" gorm:"default:0;index"`
	IsActive    bool      `json:"is_active" gorm:"default:true;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(36);not null;index"`
	
	// 关联的API接口
	APIs []APIEndpoint `json:"apis,omitempty" gorm:"foreignKey:CategoryID"`
}

// APIEndpoint API接口端点
type APIEndpoint struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	CategoryID  string    `json:"category_id" gorm:"type:varchar(36);not null;index"`
	Method      string    `json:"method" gorm:"type:varchar(10);not null;index"`
	Path        string    `json:"path" gorm:"type:varchar(500);not null;index"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Summary     string    `json:"summary" gorm:"type:varchar(500)"`
	Tags        string    `json:"tags" gorm:"type:text"` // JSON数组字符串
	
	// 文档来源信息
	SourceFile     string `json:"source_file" gorm:"type:varchar(500);not null"`
	SourcePath     string `json:"source_path" gorm:"type:varchar(1000);not null"`
	SourceLine     int    `json:"source_line" gorm:"default:0"`
	DocumentURL    string `json:"document_url" gorm:"type:varchar(1000)"`
	
	// 接口状态和版本
	Status      string `json:"status" gorm:"type:varchar(20);default:'active';index"` // active, deprecated, beta, alpha
	Version     string `json:"version" gorm:"type:varchar(20);default:'v1'"`
	IsPublic    bool   `json:"is_public" gorm:"default:true;index"`
	IsDeprecated bool  `json:"is_deprecated" gorm:"default:false;index"`
	
	// 请求和响应信息
	RequestExample  string `json:"request_example" gorm:"type:longtext"`
	ResponseExample string `json:"response_example" gorm:"type:longtext"`
	Parameters      string `json:"parameters" gorm:"type:longtext"` // JSON字符串
	Headers         string `json:"headers" gorm:"type:text"`        // JSON字符串
	
	// 统计信息
	ViewCount     int `json:"view_count" gorm:"default:0"`
	TestCount     int `json:"test_count" gorm:"default:0"`
	ErrorCount    int `json:"error_count" gorm:"default:0"`
	LastTestedAt  *time.Time `json:"last_tested_at"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(36);not null;index"`
	UpdatedBy string    `json:"updated_by" gorm:"type:varchar(36);index"`
	
	// 关联
	Category *APICategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

// APIDocumentationSource 文档来源
type APIDocumentationSource struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	FilePath    string    `json:"file_path" gorm:"type:varchar(500);not null;uniqueIndex:uni_api_documentation_sources_file_path"`
	FileType    string    `json:"file_type" gorm:"type:varchar(20);not null"` // markdown, yaml, json
	FileSize    int64     `json:"file_size" gorm:"default:0"`
	FileHash    string    `json:"file_hash" gorm:"type:varchar(64)"`
	LastScanned time.Time `json:"last_scanned"`
	ScanStatus  string    `json:"scan_status" gorm:"type:varchar(20);default:'pending'"` // pending, scanning, completed, failed
	APICount    int       `json:"api_count" gorm:"default:0"`
	ErrorMsg    string    `json:"error_msg" gorm:"type:text"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(36);not null;index"`
}

// APITestRecord API测试记录
type APITestRecord struct {
	ID           string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	EndpointID   string    `json:"endpoint_id" gorm:"type:varchar(36);not null;index"`
	TestType     string    `json:"test_type" gorm:"type:varchar(20);not null"` // manual, automated, integration
	RequestData  string    `json:"request_data" gorm:"type:longtext"`
	ResponseData string    `json:"response_data" gorm:"type:longtext"`
	StatusCode   int       `json:"status_code" gorm:"default:0"`
	ResponseTime int       `json:"response_time" gorm:"default:0"` // 毫秒
	IsSuccess    bool      `json:"is_success" gorm:"default:false;index"`
	ErrorMsg     string    `json:"error_msg" gorm:"type:text"`
	Environment  string    `json:"environment" gorm:"type:varchar(50)"` // dev, test, staging, prod
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(36);not null;index"`
	
	// 关联
	Endpoint *APIEndpoint `json:"endpoint,omitempty" gorm:"foreignKey:EndpointID"`
}

// APIChangeLog API变更日志
type APIChangeLog struct {
	ID         string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	EndpointID string    `json:"endpoint_id" gorm:"type:varchar(36);not null;index"`
	ChangeType string    `json:"change_type" gorm:"type:varchar(20);not null"` // created, updated, deprecated, deleted
	OldValue   string    `json:"old_value" gorm:"type:longtext"`
	NewValue   string    `json:"new_value" gorm:"type:longtext"`
	FieldName  string    `json:"field_name" gorm:"type:varchar(100)"`
	Reason     string    `json:"reason" gorm:"type:text"`
	Version    string    `json:"version" gorm:"type:varchar(20)"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(36);not null;index"`
	
	// 关联
	Endpoint *APIEndpoint `json:"endpoint,omitempty" gorm:"foreignKey:EndpointID"`
}

// TableName 设置表名
func (APICategory) TableName() string {
	return "api_categories"
}

func (APIEndpoint) TableName() string {
	return "api_endpoints"
}

func (APIDocumentationSource) TableName() string {
	return "api_documentation_sources"
}

func (APITestRecord) TableName() string {
	return "api_test_records"
}

func (APIChangeLog) TableName() string {
	return "api_change_logs"
}