package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/api/internal/models"
)

// Taishang Domain DAOs

type ModelDAO struct {
	db *sql.DB
}

func NewModelDAO(db *sql.DB) *ModelDAO {
	return &ModelDAO{db: db}
}

func (d *ModelDAO) List(ctx context.Context, tenantID, status, provider string, page, pageSize int) ([]*models.Model, int, error) {
	offset := (page - 1) * pageSize
	
	query := `
		SELECT id, tenant_id, name, provider, version, status, meta, created_at, updated_at
		FROM tai_models
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIndex := 2
	
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	
	if provider != "" {
		query += fmt.Sprintf(" AND provider ILIKE $%d", argIndex)
		args = append(args, "%"+provider+"%")
		argIndex++
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM tai_models WHERE " + query[len("FROM tai_models WHERE "):len(query)]
	var total int
	err := d.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var modelList []*models.Model
	for rows.Next() {
		var model models.Model
		err := rows.Scan(
			&model.ID, &model.TenantID, &model.Name, &model.Provider,
			&model.Version, &model.Status, &model.Meta, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		modelList = append(modelList, &model)
	}
	
	return modelList, total, nil
}

func (d *ModelDAO) GetByID(ctx context.Context, tenantID, id string) (*models.Model, error) {
	query := `
		SELECT id, tenant_id, name, provider, version, status, meta, created_at, updated_at
		FROM tai_models
		WHERE tenant_id = $1 AND id = $2
	`
	row := d.db.QueryRowContext(ctx, query, tenantID, id)
	
	var model models.Model
	err := row.Scan(
		&model.ID, &model.TenantID, &model.Name, &model.Provider,
		&model.Version, &model.Status, &model.Meta, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (d *ModelDAO) Create(ctx context.Context, model *models.Model) error {
	query := `
		INSERT INTO tai_models (id, tenant_id, name, provider, version, status, meta, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := d.db.ExecContext(ctx, query,
		model.ID, model.TenantID, model.Name, model.Provider,
		model.Version, model.Status, model.Meta, time.Now(), time.Now(),
	)
	return err
}

func (d *ModelDAO) Update(ctx context.Context, model *models.Model) error {
	query := `
		UPDATE tai_models
		SET name = $3, provider = $4, version = $5, status = $6, meta = $7, updated_at = $8
		WHERE tenant_id = $1 AND id = $2
	`
	_, err := d.db.ExecContext(ctx, query,
		model.TenantID, model.ID, model.Name, model.Provider,
		model.Version, model.Status, model.Meta, time.Now(),
	)
	return err
}

func (d *ModelDAO) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM tai_models WHERE tenant_id = $1 AND id = $2`
	_, err := d.db.ExecContext(ctx, query, tenantID, id)
	return err
}

type VectorCollectionDAO struct {
	db *sql.DB
}

func NewVectorCollectionDAO(db *sql.DB) *VectorCollectionDAO {
	return &VectorCollectionDAO{db: db}
}

func (d *VectorCollectionDAO) List(ctx context.Context, tenantID, modelID string, page, pageSize int) ([]*models.VectorCollection, int, error) {
	offset := (page - 1) * pageSize
	
	query := `
		SELECT id, tenant_id, name, model_id, dims, index_type, metric_type, extra_index_args, created_at
		FROM tai_vector_collections
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIndex := 2
	
	if modelID != "" {
		query += fmt.Sprintf(" AND model_id = $%d", argIndex)
		args = append(args, modelID)
		argIndex++
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM tai_vector_collections WHERE " + query[len("FROM tai_vector_collections WHERE "):len(query)]
	var total int
	err := d.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var collections []*models.VectorCollection
	for rows.Next() {
		var collection models.VectorCollection
		err := rows.Scan(
			&collection.ID, &collection.TenantID, &collection.Name, &collection.ModelID,
			&collection.Dims, &collection.IndexType, &collection.MetricType,
			&collection.ExtraIndexArgs, &collection.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		collections = append(collections, &collection)
	}
	
	return collections, total, nil
}

func (d *VectorCollectionDAO) GetByID(ctx context.Context, tenantID string, id int) (*models.VectorCollection, error) {
	query := `
		SELECT id, tenant_id, name, model_id, dims, index_type, metric_type, extra_index_args, created_at
		FROM tai_vector_collections
		WHERE tenant_id = $1 AND id = $2
	`
	row := d.db.QueryRowContext(ctx, query, tenantID, id)
	
	var collection models.VectorCollection
	err := row.Scan(
		&collection.ID, &collection.TenantID, &collection.Name, &collection.ModelID,
		&collection.Dims, &collection.IndexType, &collection.MetricType,
		&collection.ExtraIndexArgs, &collection.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (d *VectorCollectionDAO) Create(ctx context.Context, collection *models.VectorCollection) error {
	query := `
		INSERT INTO tai_vector_collections (tenant_id, name, model_id, dims, index_type, metric_type, extra_index_args, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := d.db.QueryRowContext(ctx, query,
		collection.TenantID, collection.Name, collection.ModelID,
		collection.Dims, collection.IndexType, collection.MetricType,
		collection.ExtraIndexArgs, time.Now(),
	).Scan(&collection.ID)
	return err
}

func (d *VectorCollectionDAO) Delete(ctx context.Context, tenantID string, id int) error {
	query := `DELETE FROM tai_vector_collections WHERE tenant_id = $1 AND id = $2`
	_, err := d.db.ExecContext(ctx, query, tenantID, id)
	return err
}

type TaskDAO struct {
	db *sql.DB
}

func NewTaskDAO(db *sql.DB) *TaskDAO {
	return &TaskDAO{db: db}
}

func (d *TaskDAO) List(ctx context.Context, tenantID string, status models.TaskStatus, page, pageSize int) ([]*models.Task, int, error) {
	offset := (page - 1) * pageSize
	
	query := `
		SELECT id, tenant_id, type, status, priority, payload, result, worker_id, started_at, finished_at, created_at, updated_at
		FROM tai_tasks
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIndex := 2
	
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM tai_tasks WHERE " + query[len("FROM tai_tasks WHERE "):len(query)]
	var total int
	err := d.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.TenantID, &task.Type, &task.Status, &task.Priority,
			&task.Payload, &task.Result, &task.WorkerID, &task.StartedAt,
			&task.FinishedAt, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, &task)
	}
	
	return tasks, total, nil
}

func (d *TaskDAO) GetByID(ctx context.Context, tenantID string, id int64) (*models.Task, error) {
	query := `
		SELECT id, tenant_id, type, status, priority, payload, result, worker_id, started_at, finished_at, created_at, updated_at
		FROM tai_tasks
		WHERE tenant_id = $1 AND id = $2
	`
	row := d.db.QueryRowContext(ctx, query, tenantID, id)
	
	var task models.Task
	err := row.Scan(
		&task.ID, &task.TenantID, &task.Type, &task.Status, &task.Priority,
		&task.Payload, &task.Result, &task.WorkerID, &task.StartedAt,
		&task.FinishedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (d *TaskDAO) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tai_tasks (tenant_id, type, status, priority, payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := d.db.QueryRowContext(ctx, query,
		task.TenantID, task.Type, task.Status, task.Priority,
		task.Payload, time.Now(), time.Now(),
	).Scan(&task.ID)
	return err
}

func (d *TaskDAO) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tai_tasks
		SET status = $3, priority = $4, payload = $5, result = $6, worker_id = $7, started_at = $8, finished_at = $9, updated_at = $10
		WHERE tenant_id = $1 AND id = $2
	`
	_, err := d.db.ExecContext(ctx, query,
		task.TenantID, task.ID, task.Status, task.Priority, task.Payload,
		task.Result, task.WorkerID, task.StartedAt, task.FinishedAt, time.Now(),
	)
	return err
}

func (d *TaskDAO) Delete(ctx context.Context, tenantID string, id int64) error {
	query := `DELETE FROM tai_tasks WHERE tenant_id = $1 AND id = $2`
	_, err := d.db.ExecContext(ctx, query, tenantID, id)
	return err
}