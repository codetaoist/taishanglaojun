package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/infrastructure/database-layer/internal/models"
)

// Repository 仓储接口
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
	SoftDelete(ctx context.Context, id uint) error
	List(ctx context.Context, opts *models.QueryOptions) ([]*T, error)
	Count(ctx context.Context, opts *models.QueryOptions) (int64, error)
	Paginate(ctx context.Context, opts *models.QueryOptions) (*models.PaginationResult, error)
	Exists(ctx context.Context, id uint) (bool, error)
	BatchCreate(ctx context.Context, entities []*T) error
	BatchUpdate(ctx context.Context, entities []*T) error
	BatchDelete(ctx context.Context, ids []uint) error
}

// BaseRepository 基础仓储实现
type BaseRepository[T any] struct {
	db     *gorm.DB
	logger *zap.Logger
	model  T
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](db *gorm.DB, logger *zap.Logger) *BaseRepository[T] {
	var model T
	return &BaseRepository[T]{
		db:     db,
		logger: logger,
		model:  model,
	}
}

// Create 创建实体
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		r.logger.Error("Failed to create entity",
			zap.String("type", r.getModelType()),
			zap.Error(err))
		return fmt.Errorf("failed to create entity: %w", err)
	}

	r.logger.Debug("Entity created successfully",
		zap.String("type", r.getModelType()))
	return nil
}

// GetByID 根据ID获取值实体
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("entity not found: %w", err)
		}
		r.logger.Error("Failed to get entity by ID",
			zap.String("type", r.getModelType()),
			zap.Uint("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return &entity, nil
}

// Update 更新实体
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		r.logger.Error("Failed to update entity",
			zap.String("type", r.getModelType()),
			zap.Error(err))
		return fmt.Errorf("failed to update entity: %w", err)
	}

	r.logger.Debug("Entity updated successfully",
		zap.String("type", r.getModelType()))
	return nil
}

// Delete 硬删除键实体
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	if err := r.db.WithContext(ctx).Unscoped().Delete(&entity, id).Error; err != nil {
		r.logger.Error("Failed to delete entity",
			zap.String("type", r.getModelType()),
			zap.Uint("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	r.logger.Debug("Entity deleted successfully",
		zap.String("type", r.getModelType()),
		zap.Uint("id", id))
	return nil
}

// SoftDelete 软删除键实体
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id uint) error {
	var entity T
	if err := r.db.WithContext(ctx).Delete(&entity, id).Error; err != nil {
		r.logger.Error("Failed to soft delete entity",
			zap.String("type", r.getModelType()),
			zap.Uint("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to soft delete entity: %w", err)
	}

	r.logger.Debug("Entity soft deleted successfully",
		zap.String("type", r.getModelType()),
		zap.Uint("id", id))
	return nil
}

// List 获取值实体列表
func (r *BaseRepository[T]) List(ctx context.Context, opts *models.QueryOptions) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx).Model(&r.model)

	if opts != nil {
		query = opts.ApplyToQuery(query)
	}

	if err := query.Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities",
			zap.String("type", r.getModelType()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}

	return entities, nil
}

// Count 统计实体数量
func (r *BaseRepository[T]) Count(ctx context.Context, opts *models.QueryOptions) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&r.model)

	if opts != nil {
		// 应用过滤条件，但不应用分页排序
		if len(opts.Filters) > 0 {
			for _, filter := range opts.Filters {
				query = r.applyFilter(query, filter)
			}
		}

		if opts.Search != nil && opts.Search.Keyword != "" {
			query = r.applySearch(query, opts.Search)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count entities",
			zap.String("type", r.getModelType()),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return count, nil
}

// Paginate 分页查询
func (r *BaseRepository[T]) Paginate(ctx context.Context, opts *models.QueryOptions) (*models.PaginationResult, error) {
	// 获取值总数
	total, err := r.Count(ctx, opts)
	if err != nil {
		return nil, err
	}

	// 获取值数据
	entities, err := r.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// 构建分页结果
	var pagination *models.PaginationQuery
	if opts != nil && opts.Pagination != nil {
		pagination = opts.Pagination
	} else {
		pagination = &models.PaginationQuery{Page: 1, PageSize: 10}
	}

	return models.NewPaginationResult(entities, total, pagination), nil
}

// Exists 检查实体是否存在
func (r *BaseRepository[T]) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&r.model).Where("id = ?", id).Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to check entity existence",
			zap.String("type", r.getModelType()),
			zap.Uint("id", id),
			zap.Error(err))
		return false, fmt.Errorf("failed to check entity existence: %w", err)
	}

	return count > 0, nil
}

// BatchCreate 批量创建实体
func (r *BaseRepository[T]) BatchCreate(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(entities, 100).Error; err != nil {
		r.logger.Error("Failed to batch create entities",
			zap.String("type", r.getModelType()),
			zap.Int("count", len(entities)),
			zap.Error(err))
		return fmt.Errorf("failed to batch create entities: %w", err)
	}

	r.logger.Debug("Entities batch created successfully",
		zap.String("type", r.getModelType()),
		zap.Int("count", len(entities)))
	return nil
}

// BatchUpdate 批量更新实体
func (r *BaseRepository[T]) BatchUpdate(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	// GORM不直接支持批量更新，需要在事务中逐个更新
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Save(entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchDelete 批量删除键实体
func (r *BaseRepository[T]) BatchDelete(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	var entity T
	if err := r.db.WithContext(ctx).Delete(&entity, ids).Error; err != nil {
		r.logger.Error("Failed to batch delete entities",
			zap.String("type", r.getModelType()),
			zap.Uints("ids", ids),
			zap.Error(err))
		return fmt.Errorf("failed to batch delete entities: %w", err)
	}

	r.logger.Debug("Entities batch deleted successfully",
		zap.String("type", r.getModelType()),
		zap.Int("count", len(ids)))
	return nil
}

// Transaction 执行事务
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// GetDB 获取值数据库实例
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// getModelType 获取值模型类型名称
func (r *BaseRepository[T]) getModelType() string {
	return reflect.TypeOf(r.model).String()
}

// applyFilter 应用过滤条件
func (r *BaseRepository[T]) applyFilter(db *gorm.DB, filter models.FilterQuery) *gorm.DB {
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
func (r *BaseRepository[T]) applySearch(db *gorm.DB, search *models.SearchQuery) *gorm.DB {
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

