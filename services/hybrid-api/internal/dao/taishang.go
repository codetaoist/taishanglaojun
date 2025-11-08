package dao

import (
	"context"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"gorm.io/gorm"
)

// Taishang Domain DAOs

type TaishangModelDAO struct {
	db *gorm.DB
}

func NewTaishangModelDAO(db *gorm.DB) *TaishangModelDAO {
	return &TaishangModelDAO{db: db}
}

func (d *TaishangModelDAO) List(ctx context.Context, tenantID, status, provider string, page, pageSize int) ([]*models.Model, int, error) {
	offset := (page - 1) * pageSize
	
	var modelList []*models.Model
	query := d.db.WithContext(ctx).Table("tai_models").Where("tenant_id = ?", tenantID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if provider != "" {
		query = query.Where("provider ILIKE ?", "%"+provider+"%")
	}
	
	// Get total count
	var total int64
	countQuery := d.db.WithContext(ctx).Table("tai_models").Where("tenant_id = ?", tenantID)
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if provider != "" {
		countQuery = countQuery.Where("provider ILIKE ?", "%"+provider+"%")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&modelList).Error; err != nil {
		return nil, 0, err
	}
	
	return modelList, int(total), nil
}

func (d *TaishangModelDAO) GetByID(ctx context.Context, tenantID, id string) (*models.Model, error) {
	var model models.Model
	err := d.db.WithContext(ctx).Table("tai_models").Where("tenant_id = ? AND id = ?", tenantID, id).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (d *TaishangModelDAO) Create(ctx context.Context, model *models.Model) error {
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	return d.db.WithContext(ctx).Table("tai_models").Create(model).Error
}

func (d *TaishangModelDAO) Update(ctx context.Context, model *models.Model) error {
	model.UpdatedAt = time.Now()
	return d.db.WithContext(ctx).Table("tai_models").Save(model).Error
}

func (d *TaishangModelDAO) Delete(ctx context.Context, tenantID, id string) error {
	return d.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.Model{}).Error
}

type VectorCollectionDAO struct {
	db *gorm.DB
}

func NewVectorCollectionDAO(db *gorm.DB) *VectorCollectionDAO {
	return &VectorCollectionDAO{db: db}
}

func (d *VectorCollectionDAO) List(ctx context.Context, tenantID, modelID string, page, pageSize int) ([]*models.VectorCollection, int, error) {
	offset := (page - 1) * pageSize
	
	var collections []models.VectorCollection
	query := d.db.WithContext(ctx).Table("tai_vector_collections").Where("tenant_id = ?", tenantID)
	
	if modelID != "" {
		query = query.Where("model_id = ?", modelID)
	}
	
	// Get total count
	var total int64
	countQuery := d.db.WithContext(ctx).Table("tai_vector_collections").Where("tenant_id = ?", tenantID)
	if modelID != "" {
		countQuery = countQuery.Where("model_id = ?", modelID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&collections).Error; err != nil {
		return nil, 0, err
	}
	
	// Convert to pointer slice
	var collectionList []*models.VectorCollection
	for i := range collections {
		collectionList = append(collectionList, &collections[i])
	}
	
	return collectionList, int(total), nil
}

func (d *VectorCollectionDAO) GetByID(ctx context.Context, tenantID string, id int) (*models.VectorCollection, error) {
	var collection models.VectorCollection
	err := d.db.WithContext(ctx).Table("tai_vector_collections").Where("tenant_id = ? AND id = ?", tenantID, id).First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (d *VectorCollectionDAO) Create(ctx context.Context, collection *models.VectorCollection) error {
	collection.CreatedAt = time.Now()
	return d.db.WithContext(ctx).Table("tai_vector_collections").Create(collection).Error
}

func (d *VectorCollectionDAO) Delete(ctx context.Context, tenantID string, id int) error {
	return d.db.WithContext(ctx).Table("tai_vector_collections").Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.VectorCollection{}).Error
}

type TaskDAO struct {
	db *gorm.DB
}

func NewTaskDAO(db *gorm.DB) *TaskDAO {
	return &TaskDAO{db: db}
}

func (d *TaskDAO) List(ctx context.Context, tenantID string, status models.TaskStatus, page, pageSize int) ([]*models.Task, int, error) {
	offset := (page - 1) * pageSize
	
	var taskList []*models.Task
	query := d.db.WithContext(ctx).Table("tai_tasks").Where("tenant_id = ?", tenantID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// Get total count
	var total int64
	countQuery := d.db.WithContext(ctx).Table("tai_tasks").Where("tenant_id = ?", tenantID)
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&taskList).Error; err != nil {
		return nil, 0, err
	}
	
	return taskList, int(total), nil
}

func (d *TaskDAO) GetByID(ctx context.Context, tenantID string, id int64) (*models.Task, error) {
	var task models.Task
	err := d.db.WithContext(ctx).Table("tai_tasks").Where("tenant_id = ? AND id = ?", tenantID, id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (d *TaskDAO) Create(ctx context.Context, task *models.Task) error {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	return d.db.WithContext(ctx).Table("tai_tasks").Create(task).Error
}

func (d *TaskDAO) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()
	return d.db.WithContext(ctx).Table("tai_tasks").Save(task).Error
}

func (d *TaskDAO) Delete(ctx context.Context, tenantID string, id int64) error {
	return d.db.WithContext(ctx).Table("tai_tasks").Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.Task{}).Error
}