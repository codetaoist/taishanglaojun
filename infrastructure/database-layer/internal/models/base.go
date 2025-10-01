package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Model 模型接口
type Model interface {
	TableName() string
	BeforeCreate(tx *gorm.DB) error
	BeforeUpdate(tx *gorm.DB) error
}

// SoftDeleteModel 软删除模型接�?
type SoftDeleteModel interface {
	Model
	IsDeleted() bool
	SoftDelete() error
	Restore() error
}

// TimestampModel 时间戳模型接�?
type TimestampModel interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

// GetCreatedAt 获取创建时间
func (m *BaseModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (m *BaseModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

// SetCreatedAt 设置创建时间
func (m *BaseModel) SetCreatedAt(t time.Time) {
	m.CreatedAt = t
}

// SetUpdatedAt 设置更新时间
func (m *BaseModel) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}

// IsDeleted 检查是否已软删�?
func (m *BaseModel) IsDeleted() bool {
	return m.DeletedAt.Valid
}

// BeforeCreate GORM钩子：创建前
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// PaginationQuery 分页查询参数
type PaginationQuery struct {
	Page     int    `json:"page" form:"page" binding:"min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	OrderBy  string `json:"order_by" form:"order_by"`
	Sort     string `json:"sort" form:"sort" binding:"oneof=asc desc"`
}

// GetOffset 获取偏移�?
func (p *PaginationQuery) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

// GetLimit 获取限制数量
func (p *PaginationQuery) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}

// GetOrderBy 获取排序字段
func (p *PaginationQuery) GetOrderBy() string {
	if p.OrderBy == "" {
		return "id"
	}
	return p.OrderBy
}

// GetSort 获取排序方向
func (p *PaginationQuery) GetSort() string {
	if p.Sort == "" {
		return "desc"
	}
	return p.Sort
}

// PaginationResult 分页结果
type PaginationResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginationResult 创建分页结果
func NewPaginationResult(data interface{}, total int64, query *PaginationQuery) *PaginationResult {
	totalPages := int(total) / query.GetLimit()
	if int(total)%query.GetLimit() > 0 {
		totalPages++
	}

	return &PaginationResult{
		Data:       data,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.GetLimit(),
		TotalPages: totalPages,
	}
}

// FilterQuery 过滤查询参数
type FilterQuery struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, gte, lt, lte, like, in, not_in
	Value    interface{} `json:"value"`
}

// SearchQuery 搜索查询参数
type SearchQuery struct {
	Keyword string   `json:"keyword" form:"keyword"`
	Fields  []string `json:"fields" form:"fields"`
}

// QueryOptions 查询选项
type QueryOptions struct {
	Pagination *PaginationQuery `json:"pagination"`
	Filters    []FilterQuery    `json:"filters"`
	Search     *SearchQuery     `json:"search"`
	Preload    []string         `json:"preload"`
	Select     []string         `json:"select"`
	Omit       []string         `json:"omit"`
}

// ApplyToQuery 将查询选项应用到GORM查询
func (opts *QueryOptions) ApplyToQuery(db *gorm.DB) *gorm.DB {
	query := db

	// 应用预加�?
	if len(opts.Preload) > 0 {
		for _, preload := range opts.Preload {
			query = query.Preload(preload)
		}
	}

	// 应用字段选择
	if len(opts.Select) > 0 {
		query = query.Select(opts.Select)
	}

	// 应用字段忽略
	if len(opts.Omit) > 0 {
		query = query.Omit(opts.Omit...)
	}

	// 应用过滤条件
	if len(opts.Filters) > 0 {
		for _, filter := range opts.Filters {
			query = applyFilter(query, filter)
		}
	}

	// 应用搜索条件
	if opts.Search != nil && opts.Search.Keyword != "" {
		query = applySearch(query, opts.Search)
	}

	// 应用分页和排�?
	if opts.Pagination != nil {
		orderBy := opts.Pagination.GetOrderBy() + " " + opts.Pagination.GetSort()
		query = query.Order(orderBy).
			Offset(opts.Pagination.GetOffset()).
			Limit(opts.Pagination.GetLimit())
	}

	return query
}

// applyFilter 应用过滤条件
func applyFilter(db *gorm.DB, filter FilterQuery) *gorm.DB {
	switch filter.Operator {
	case "eq":
		return db.Where(filter.Field+" = ?", filter.Value)
	case "ne":
		return db.Where(filter.Field+" != ?", filter.Value)
	case "gt":
		return db.Where(filter.Field+" > ?", filter.Value)
	case "gte":
		return db.Where(filter.Field+" >= ?", filter.Value)
	case "lt":
		return db.Where(filter.Field+" < ?", filter.Value)
	case "lte":
		return db.Where(filter.Field+" <= ?", filter.Value)
	case "like":
		return db.Where(filter.Field+" LIKE ?", "%"+filter.Value.(string)+"%")
	case "in":
		return db.Where(filter.Field+" IN ?", filter.Value)
	case "not_in":
		return db.Where(filter.Field+" NOT IN ?", filter.Value)
	default:
		return db
	}
}

// applySearch 应用搜索条件
func applySearch(db *gorm.DB, search *SearchQuery) *gorm.DB {
	if len(search.Fields) == 0 {
		return db
	}

	query := db
	for i, field := range search.Fields {
		if i == 0 {
			query = query.Where(field+" LIKE ?", "%"+search.Keyword+"%")
		} else {
			query = query.Or(field+" LIKE ?", "%"+search.Keyword+"%")
		}
	}
	return query
}
