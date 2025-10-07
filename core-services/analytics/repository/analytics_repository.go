package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/taishanglaojun/core-services/analytics"
)

// AnalyticsRepositoryImpl 数据分析仓储实现
type AnalyticsRepositoryImpl struct {
	db     *sqlx.DB
	config *AnalyticsRepositoryConfig
}

// AnalyticsRepositoryConfig 数据分析仓储配置
type AnalyticsRepositoryConfig struct {
	// 数据库配置
	DatabaseURL      string        `json:"database_url"`
	MaxOpenConns     int           `json:"max_open_conns"`
	MaxIdleConns     int           `json:"max_idle_conns"`
	ConnMaxLifetime  time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime  time.Duration `json:"conn_max_idle_time"`
	
	// 性能配置
	BatchSize        int           `json:"batch_size"`
	QueryTimeout     time.Duration `json:"query_timeout"`
	
	// 索引配置
	EnableIndexing   bool          `json:"enable_indexing"`
	IndexFields      []string      `json:"index_fields"`
	
	// 分区配置
	EnablePartitioning bool        `json:"enable_partitioning"`
	PartitionField     string      `json:"partition_field"`
	PartitionInterval  string      `json:"partition_interval"`
	
	// 审计配置
	EnableAudit      bool          `json:"enable_audit"`
	AuditTableName   string        `json:"audit_table_name"`
}

// JSONField JSON字段类型
type JSONField map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONField", value)
	}
	
	return json.Unmarshal(bytes, j)
}

// StringSliceField 字符串切片字段类型
type StringSliceField []string

// Value 实现 driver.Valuer 接口
func (s StringSliceField) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringSliceField) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringSliceField", value)
	}
	
	return json.Unmarshal(bytes, s)
}

// DataPointRow 数据点数据库行
type DataPointRow struct {
	ID          string           `db:"id"`
	Source      string           `db:"source"`
	Type        string           `db:"type"`
	Category    string           `db:"category"`
	Timestamp   time.Time        `db:"timestamp"`
	Value       JSONField        `db:"value"`
	Metadata    JSONField        `db:"metadata"`
	Tags        StringSliceField `db:"tags"`
	UserID      string           `db:"user_id"`
	TenantID    string           `db:"tenant_id"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
}

// AggregatedDataRow 聚合数据数据库行
type AggregatedDataRow struct {
	ID           string           `db:"id"`
	Source       string           `db:"source"`
	Type         string           `db:"type"`
	Category     string           `db:"category"`
	TimeStart    time.Time        `db:"time_start"`
	TimeEnd      time.Time        `db:"time_end"`
	Aggregation  string           `db:"aggregation"`
	Value        JSONField        `db:"value"`
	Count        int64            `db:"count"`
	Metadata     JSONField        `db:"metadata"`
	Tags         StringSliceField `db:"tags"`
	TenantID     string           `db:"tenant_id"`
	CreatedAt    time.Time        `db:"created_at"`
}

// ReportRow 报表数据库行
type ReportRow struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Type        string    `db:"type"`
	Format      string    `db:"format"`
	Status      string    `db:"status"`
	Config      JSONField `db:"config"`
	Data        JSONField `db:"data"`
	FilePath    string    `db:"file_path"`
	FileSize    int64     `db:"file_size"`
	UserID      string    `db:"user_id"`
	TenantID    string    `db:"tenant_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
}

// NewAnalyticsRepository 创建数据分析仓储
func NewAnalyticsRepository(config *AnalyticsRepositoryConfig) (*AnalyticsRepositoryImpl, error) {
	if config == nil {
		config = &AnalyticsRepositoryConfig{
			MaxOpenConns:       25,
			MaxIdleConns:       5,
			ConnMaxLifetime:    5 * time.Minute,
			ConnMaxIdleTime:    1 * time.Minute,
			BatchSize:          1000,
			QueryTimeout:       30 * time.Second,
			EnableIndexing:     true,
			EnablePartitioning: true,
			PartitionField:     "timestamp",
			PartitionInterval:  "month",
			EnableAudit:        true,
			AuditTableName:     "analytics_audit",
		}
	}

	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	repo := &AnalyticsRepositoryImpl{
		db:     db,
		config: config,
	}

	// 初始化数据库表
	if err := repo.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return repo, nil
}

// SaveDataPoint 保存数据点
func (r *AnalyticsRepositoryImpl) SaveDataPoint(ctx context.Context, dataPoint *analytics.DataPoint) error {
	if dataPoint == nil {
		return fmt.Errorf("data point cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	row := r.dataPointToRow(dataPoint)
	
	query := `
		INSERT INTO data_points (id, source, type, category, timestamp, value, metadata, tags, user_id, tenant_id, created_at, updated_at)
		VALUES (:id, :source, :type, :category, :timestamp, :value, :metadata, :tags, :user_id, :tenant_id, :created_at, :updated_at)
	`
	
	_, err := r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to save data point: %w", err)
	}

	// 审计日志
	if r.config.EnableAudit {
		r.auditDataPoint("INSERT", dataPoint)
	}

	return nil
}

// SaveDataPoints 批量保存数据点
func (r *AnalyticsRepositoryImpl) SaveDataPoints(ctx context.Context, dataPoints []*analytics.DataPoint) error {
	if len(dataPoints) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 分批处理
	batchSize := r.config.BatchSize
	for i := 0; i < len(dataPoints); i += batchSize {
		end := i + batchSize
		if end > len(dataPoints) {
			end = len(dataPoints)
		}

		batch := dataPoints[i:end]
		if err := r.saveBatchDataPoints(ctx, batch); err != nil {
			return fmt.Errorf("failed to save batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

// QueryDataPoints 查询数据点
func (r *AnalyticsRepositoryImpl) QueryDataPoints(ctx context.Context, filter *analytics.DataFilter) ([]*analytics.DataPoint, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query, args := r.buildDataPointQuery(filter)
	
	var rows []DataPointRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query data points: %w", err)
	}

	dataPoints := make([]*analytics.DataPoint, len(rows))
	for i, row := range rows {
		dataPoints[i] = r.rowToDataPoint(&row)
	}

	return dataPoints, nil
}

// QueryAggregatedData 查询聚合数据
func (r *AnalyticsRepositoryImpl) QueryAggregatedData(ctx context.Context, filter *analytics.AggregationFilter) ([]*analytics.AggregatedData, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query, args := r.buildAggregationQuery(filter)
	
	var rows []AggregatedDataRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated data: %w", err)
	}

	aggregatedData := make([]*analytics.AggregatedData, len(rows))
	for i, row := range rows {
		aggregatedData[i] = r.rowToAggregatedData(&row)
	}

	return aggregatedData, nil
}

// SaveReport 保存报表
func (r *AnalyticsRepositoryImpl) SaveReport(ctx context.Context, report *analytics.Report) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	row := r.reportToRow(report)
	
	query := `
		INSERT INTO reports (id, name, description, type, format, status, config, data, file_path, file_size, user_id, tenant_id, created_at, updated_at, expires_at)
		VALUES (:id, :name, :description, :type, :format, :status, :config, :data, :file_path, :file_size, :user_id, :tenant_id, :created_at, :updated_at, :expires_at)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			config = EXCLUDED.config,
			data = EXCLUDED.data,
			file_path = EXCLUDED.file_path,
			file_size = EXCLUDED.file_size,
			updated_at = EXCLUDED.updated_at,
			expires_at = EXCLUDED.expires_at
	`
	
	_, err := r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	return nil
}

// GetReport 获取报表
func (r *AnalyticsRepositoryImpl) GetReport(ctx context.Context, reportID string) (*analytics.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `SELECT * FROM reports WHERE id = $1`
	
	var row ReportRow
	err := r.db.GetContext(ctx, &row, query, reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	return r.rowToReport(&row), nil
}

// ListReports 列出报表
func (r *AnalyticsRepositoryImpl) ListReports(ctx context.Context, filter *analytics.ReportFilter) ([]*analytics.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query, args := r.buildReportQuery(filter)
	
	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	reports := make([]*analytics.Report, len(rows))
	for i, row := range rows {
		reports[i] = r.rowToReport(&row)
	}

	return reports, nil
}

// DeleteReport 删除报表
func (r *AnalyticsRepositoryImpl) DeleteReport(ctx context.Context, reportID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `DELETE FROM reports WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("report not found: %s", reportID)
	}

	return nil
}

// DeleteDataPoints 删除数据点
func (r *AnalyticsRepositoryImpl) DeleteDataPoints(ctx context.Context, filter *analytics.DataFilter) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	whereClause, args := r.buildDataPointWhereClause(filter)
	query := fmt.Sprintf("DELETE FROM data_points WHERE %s", whereClause)
	
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete data points: %w", err)
	}

	return nil
}

// HealthCheck 健康检查
func (r *AnalyticsRepositoryImpl) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result int
	err := r.db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// 私有方法

func (r *AnalyticsRepositoryImpl) initTables() error {
	// 创建数据点表
	dataPointsTable := `
		CREATE TABLE IF NOT EXISTS data_points (
			id VARCHAR(255) PRIMARY KEY,
			source VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(255),
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			value JSONB,
			metadata JSONB,
			tags JSONB,
			user_id VARCHAR(255),
			tenant_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`

	// 创建聚合数据表
	aggregatedDataTable := `
		CREATE TABLE IF NOT EXISTS aggregated_data (
			id VARCHAR(255) PRIMARY KEY,
			source VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(255),
			time_start TIMESTAMP WITH TIME ZONE NOT NULL,
			time_end TIMESTAMP WITH TIME ZONE NOT NULL,
			aggregation VARCHAR(50) NOT NULL,
			value JSONB,
			count BIGINT NOT NULL DEFAULT 0,
			metadata JSONB,
			tags JSONB,
			tenant_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`

	// 创建报表表
	reportsTable := `
		CREATE TABLE IF NOT EXISTS reports (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			format VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			config JSONB,
			data JSONB,
			file_path VARCHAR(1000),
			file_size BIGINT DEFAULT 0,
			user_id VARCHAR(255),
			tenant_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE
		)
	`

	tables := []string{dataPointsTable, aggregatedDataTable, reportsTable}
	
	for _, table := range tables {
		if _, err := r.db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// 创建索引
	if r.config.EnableIndexing {
		if err := r.createIndexes(); err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	return nil
}

func (r *AnalyticsRepositoryImpl) createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_data_points_timestamp ON data_points (timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_data_points_source ON data_points (source)",
		"CREATE INDEX IF NOT EXISTS idx_data_points_type ON data_points (type)",
		"CREATE INDEX IF NOT EXISTS idx_data_points_category ON data_points (category)",
		"CREATE INDEX IF NOT EXISTS idx_data_points_tenant_id ON data_points (tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_data_points_user_id ON data_points (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_aggregated_data_time_range ON aggregated_data (time_start, time_end)",
		"CREATE INDEX IF NOT EXISTS idx_aggregated_data_source ON aggregated_data (source)",
		"CREATE INDEX IF NOT EXISTS idx_aggregated_data_tenant_id ON aggregated_data (tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_status ON reports (status)",
		"CREATE INDEX IF NOT EXISTS idx_reports_type ON reports (type)",
		"CREATE INDEX IF NOT EXISTS idx_reports_tenant_id ON reports (tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_user_id ON reports (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_expires_at ON reports (expires_at)",
	}

	for _, index := range indexes {
		if _, err := r.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func (r *AnalyticsRepositoryImpl) saveBatchDataPoints(ctx context.Context, dataPoints []*analytics.DataPoint) error {
	rows := make([]DataPointRow, len(dataPoints))
	for i, dp := range dataPoints {
		rows[i] = *r.dataPointToRow(dp)
	}

	query := `
		INSERT INTO data_points (id, source, type, category, timestamp, value, metadata, tags, user_id, tenant_id, created_at, updated_at)
		VALUES (:id, :source, :type, :category, :timestamp, :value, :metadata, :tags, :user_id, :tenant_id, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, rows)
	return err
}

func (r *AnalyticsRepositoryImpl) buildDataPointQuery(filter *analytics.DataFilter) (string, []interface{}) {
	query := "SELECT * FROM data_points"
	whereClause, args := r.buildDataPointWhereClause(filter)
	
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	
	query += " ORDER BY timestamp DESC"
	
	return query, args
}

func (r *AnalyticsRepositoryImpl) buildDataPointWhereClause(filter *analytics.DataFilter) (string, []interface{}) {
	if filter == nil {
		return "1=1", []interface{}{}
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	if len(filter.Sources) > 0 {
		placeholders := make([]string, len(filter.Sources))
		for i, source := range filter.Sources {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, source)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("source IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, dataType := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, string(dataType))
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Categories) > 0 {
		placeholders := make([]string, len(filter.Categories))
		for i, category := range filter.Categories {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, category)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("category IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.TimeRange != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, filter.TimeRange.Start)
		argIndex++
		
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, filter.TimeRange.End)
		argIndex++
	}

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

	if len(conditions) == 0 {
		return "1=1", args
	}

	return strings.Join(conditions, " AND "), args
}

func (r *AnalyticsRepositoryImpl) buildAggregationQuery(filter *analytics.AggregationFilter) (string, []interface{}) {
	// 实现聚合查询构建逻辑
	query := "SELECT * FROM aggregated_data WHERE 1=1"
	args := []interface{}{}
	
	// 这里可以根据filter构建更复杂的聚合查询
	
	return query, args
}

func (r *AnalyticsRepositoryImpl) buildReportQuery(filter *analytics.ReportFilter) (string, []interface{}) {
	query := "SELECT * FROM reports"
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter != nil {
		if len(filter.Types) > 0 {
			placeholders := make([]string, len(filter.Types))
			for i, reportType := range filter.Types {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, string(reportType))
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
		}

		if len(filter.Statuses) > 0 {
			placeholders := make([]string, len(filter.Statuses))
			for i, status := range filter.Statuses {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, string(status))
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
		}

		if filter.UserID != "" {
			conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
			args = append(args, filter.UserID)
			argIndex++
		}

		if filter.TenantID != "" {
			conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
			args = append(args, filter.TenantID)
			argIndex++
		}

		if filter.TimeRange != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, filter.TimeRange.Start)
			argIndex++
			
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, filter.TimeRange.End)
			argIndex++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	return query, args
}

func (r *AnalyticsRepositoryImpl) dataPointToRow(dp *analytics.DataPoint) *DataPointRow {
	return &DataPointRow{
		ID:        dp.ID,
		Source:    dp.Source,
		Type:      string(dp.Type),
		Category:  dp.Category,
		Timestamp: dp.Timestamp,
		Value:     JSONField(map[string]interface{}{"value": dp.Value}),
		Metadata:  JSONField(dp.Metadata),
		Tags:      StringSliceField(dp.Tags),
		UserID:    dp.UserID,
		TenantID:  dp.TenantID,
		CreatedAt: dp.CreatedAt,
		UpdatedAt: dp.UpdatedAt,
	}
}

func (r *AnalyticsRepositoryImpl) rowToDataPoint(row *DataPointRow) *analytics.DataPoint {
	var value interface{}
	if row.Value != nil {
		if v, ok := row.Value["value"]; ok {
			value = v
		}
	}

	return &analytics.DataPoint{
		ID:        row.ID,
		Source:    row.Source,
		Type:      analytics.DataType(row.Type),
		Category:  row.Category,
		Timestamp: row.Timestamp,
		Value:     value,
		Metadata:  map[string]interface{}(row.Metadata),
		Tags:      []string(row.Tags),
		UserID:    row.UserID,
		TenantID:  row.TenantID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (r *AnalyticsRepositoryImpl) rowToAggregatedData(row *AggregatedDataRow) *analytics.AggregatedData {
	var value interface{}
	if row.Value != nil {
		if v, ok := row.Value["value"]; ok {
			value = v
		}
	}

	return &analytics.AggregatedData{
		ID:          row.ID,
		Source:      row.Source,
		Type:        analytics.DataType(row.Type),
		Category:    row.Category,
		TimeRange:   analytics.TimeRange{Start: row.TimeStart, End: row.TimeEnd},
		Aggregation: analytics.AggregationType(row.Aggregation),
		Value:       value,
		Count:       row.Count,
		Metadata:    map[string]interface{}(row.Metadata),
		Tags:        []string(row.Tags),
		TenantID:    row.TenantID,
		CreatedAt:   row.CreatedAt,
	}
}

func (r *AnalyticsRepositoryImpl) reportToRow(report *analytics.Report) *ReportRow {
	return &ReportRow{
		ID:          report.ID,
		Name:        report.Name,
		Description: report.Description,
		Type:        string(report.Type),
		Format:      string(report.Format),
		Status:      string(report.Status),
		Config:      JSONField(map[string]interface{}{"config": report.Config}),
		Data:        JSONField(map[string]interface{}{"data": report.Data}),
		FilePath:    report.FilePath,
		FileSize:    report.FileSize,
		UserID:      report.UserID,
		TenantID:    report.TenantID,
		CreatedAt:   report.CreatedAt,
		UpdatedAt:   report.UpdatedAt,
		ExpiresAt:   report.ExpiresAt,
	}
}

func (r *AnalyticsRepositoryImpl) rowToReport(row *ReportRow) *analytics.Report {
	var config analytics.ReportConfig
	var data interface{}

	if row.Config != nil {
		if c, ok := row.Config["config"]; ok {
			if configBytes, err := json.Marshal(c); err == nil {
				json.Unmarshal(configBytes, &config)
			}
		}
	}

	if row.Data != nil {
		if d, ok := row.Data["data"]; ok {
			data = d
		}
	}

	return &analytics.Report{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Type:        analytics.ReportType(row.Type),
		Format:      analytics.ReportFormat(row.Format),
		Status:      analytics.ReportStatus(row.Status),
		Config:      config,
		Data:        data,
		FilePath:    row.FilePath,
		FileSize:    row.FileSize,
		UserID:      row.UserID,
		TenantID:    row.TenantID,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		ExpiresAt:   row.ExpiresAt,
	}
}

func (r *AnalyticsRepositoryImpl) auditDataPoint(action string, dataPoint *analytics.DataPoint) {
	// 实现审计日志记录
	// 这里可以记录到专门的审计表或发送到审计服务
}