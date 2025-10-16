package repositories

import (
	"context"
)

// Repository 通用数据访问接口
type Repository interface {
	// Create 创建记录
	Create(ctx context.Context, collection string, data interface{}) error
	
	// Get 获取记录
	Get(ctx context.Context, collection string, id string, result interface{}) error
	
	// Update 更新记录
	Update(ctx context.Context, collection string, id string, data interface{}) error
	
	// Delete 删除记录
	Delete(ctx context.Context, collection string, id string) error
	
	// List 列出记录
	List(ctx context.Context, collection string, filter map[string]interface{}, result interface{}) error
}

