package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/audit"
)

// AuditRepository 洢
type AuditRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
	config AuditRepositoryConfig
}

// AuditRepositoryConfig 洢
type AuditRepositoryConfig struct {
	TableName       string        `json:"table_name"`
	ArchiveTable    string        `json:"archive_table"`
	EnableSharding  bool          `json:"enable_sharding"`
	ShardingKey     string        `json:"sharding_key"`
	ShardCount      int           `json:"shard_count"`
	EnableIndexing  bool          `json:"enable_indexing"`
	IndexFields     []string      `json:"index_fields"`
	EnablePartition bool          `json:"enable_partition"`
	PartitionBy     string        `json:"partition_by"`
	BatchSize       int           `json:"batch_size"`
	QueryTimeout    time.Duration `json:"query_timeout"`
}

// NewAuditRepository 洢
func NewAuditRepository(db *sqlx.DB, config AuditRepositoryConfig, logger *zap.Logger) *AuditRepository {
	// 
	if config.TableName == "" {
		config.TableName = "audit_events"
	}
	if config.ArchiveTable == "" {
		config.ArchiveTable = "audit_events_archive"
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}
	if config.QueryTimeout == 0 {
		config.QueryTimeout = 30 * time.Second
	}

	return &AuditRepository{
		db:     db,
		logger: logger,
		config: config,
	}
}

// SaveEvent 浥
func (r *AuditRepository) SaveEvent(ctx context.Context, event *audit.AuditEvent) error {
	return r.SaveEvents(ctx, []*audit.AuditEvent{event})
}

// SaveEvents 
func (r *AuditRepository) SaveEvents(ctx context.Context, events []*audit.AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	// 
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 
	tableName := r.getTableName(events[0])
	query := r.buildInsertQuery(tableName)

	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 
	for _, event := range events {
		if err := r.insertEvent(ctx, stmt, event); err != nil {
			return fmt.Errorf("failed to insert event %s: %w", event.ID, err)
		}
	}

	// 
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Debug("Events saved",
		zap.Int("count", len(events)),
		zap.String("table", tableName))

	return nil
}

// QueryEvents 
func (r *AuditRepository) QueryEvents(ctx context.Context, query *audit.AuditQuery) ([]*audit.AuditEvent, int64, error) {
	// 
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 
	selectQuery, countQuery, args := r.buildSelectQuery(query)

	// 
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// 
	rows, err := r.db.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*audit.AuditEvent
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return events, total, nil
}

// GetStatistics 
func (r *AuditRepository) GetStatistics(ctx context.Context, filter *audit.StatisticsFilter) (*audit.AuditStatistics, error) {
	// 
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	stats := &audit.AuditStatistics{
		TimeRange: audit.TimeRange{
			StartTime: *filter.StartTime,
			EndTime:   *filter.EndTime,
		},
	}

	// 
	if err := r.getBasicStatistics(ctx, filter, stats); err != nil {
		return nil, fmt.Errorf("failed to get basic statistics: %w", err)
	}

	// 
	if err := r.getUserActivityStatistics(ctx, filter, stats); err != nil {
		return nil, fmt.Errorf("failed to get user activity statistics: %w", err)
	}

	// 
	if err := r.getResourceActivityStatistics(ctx, filter, stats); err != nil {
		return nil, fmt.Errorf("failed to get resource activity statistics: %w", err)
	}

	// 
	if err := r.getSecurityStatistics(ctx, filter, stats); err != nil {
		return nil, fmt.Errorf("failed to get security statistics: %w", err)
	}

	return stats, nil
}

// ArchiveEvents 鵵
func (r *AuditRepository) ArchiveEvents(ctx context.Context, cutoffTime time.Time, batchSize int) (int64, error) {
	// 
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	var totalArchived int64

	for {
		// 
		tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return totalArchived, fmt.Errorf("failed to begin transaction: %w", err)
		}

		// 鵵
		selectQuery := fmt.Sprintf(`
			SELECT * FROM %s 
			WHERE timestamp < $1 
			ORDER BY timestamp 
			LIMIT $2`,
			r.config.TableName)

		rows, err := tx.QueryxContext(ctx, selectQuery, cutoffTime, batchSize)
		if err != nil {
			tx.Rollback()
			return totalArchived, fmt.Errorf("failed to query events for archive: %w", err)
		}

		var events []*audit.AuditEvent
		var eventIDs []string

		for rows.Next() {
			event, err := r.scanEvent(rows)
			if err != nil {
				rows.Close()
				tx.Rollback()
				return totalArchived, fmt.Errorf("failed to scan event: %w", err)
			}
			events = append(events, event)
			eventIDs = append(eventIDs, event.ID)
		}
		rows.Close()

		if len(events) == 0 {
			tx.Rollback()
			break
		}

		// 鵵
		archiveQuery := r.buildInsertQuery(r.config.ArchiveTable)
		stmt, err := tx.PreparexContext(ctx, archiveQuery)
		if err != nil {
			tx.Rollback()
			return totalArchived, fmt.Errorf("failed to prepare archive statement: %w", err)
		}

		for _, event := range events {
			if err := r.insertEvent(ctx, stmt, event); err != nil {
				stmt.Close()
				tx.Rollback()
				return totalArchived, fmt.Errorf("failed to insert archived event: %w", err)
			}
		}
		stmt.Close()

		// 鵵
		deleteQuery := fmt.Sprintf(`DELETE FROM %s WHERE id = ANY($1)`, r.config.TableName)
		result, err := tx.ExecContext(ctx, deleteQuery, eventIDs)
		if err != nil {
			tx.Rollback()
			return totalArchived, fmt.Errorf("failed to delete events: %w", err)
		}

		// 
		if err := tx.Commit(); err != nil {
			return totalArchived, fmt.Errorf("failed to commit transaction: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		totalArchived += rowsAffected

		r.logger.Debug("Events archived",
			zap.Int64("count", rowsAffected),
			zap.Time("cutoff_time", cutoffTime))

		// 
		if len(events) < batchSize {
			break
		}
	}

	return totalArchived, nil
}

// DeleteExpiredEvents 
func (r *AuditRepository) DeleteExpiredEvents(ctx context.Context, cutoffTime time.Time, batchSize int) (int64, error) {
	// 
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	var totalDeleted int64

	for {
		// 
		deleteQuery := fmt.Sprintf(`
			DELETE FROM %s 
			WHERE id IN (
				SELECT id FROM %s 
				WHERE timestamp < $1 
				ORDER BY timestamp 
				LIMIT $2
			)`,
			r.config.TableName, r.config.TableName)

		result, err := r.db.ExecContext(ctx, deleteQuery, cutoffTime, batchSize)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete expired events: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += rowsAffected

		r.logger.Debug("Expired events deleted",
			zap.Int64("count", rowsAffected),
			zap.Time("cutoff_time", cutoffTime))

		// 
		if rowsAffected < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}

// HealthCheck 
func (r *AuditRepository) HealthCheck(ctx context.Context) error {
	// 
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = $1
		)`

	if err := r.db.GetContext(ctx, &exists, checkQuery, r.config.TableName); err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("audit table %s does not exist", r.config.TableName)
	}

	return nil
}

// 

// getTableName 
func (r *AuditRepository) getTableName(event *audit.AuditEvent) string {
	if !r.config.EnableSharding {
		return r.config.TableName
	}

	// 
	var shardKey string
	switch r.config.ShardingKey {
	case "tenant_id":
		shardKey = event.TenantID
	case "user_id":
		shardKey = event.UserID
	case "event_type":
		shardKey = string(event.EventType)
	default:
		shardKey = event.ID
	}

	shardIndex := r.calculateShard(shardKey)
	return fmt.Sprintf("%s_%d", r.config.TableName, shardIndex)
}

// calculateShard 
func (r *AuditRepository) calculateShard(key string) int {
	if r.config.ShardCount <= 1 {
		return 0
	}

	hash := 0
	for _, c := range key {
		hash = hash*31 + int(c)
	}

	if hash < 0 {
		hash = -hash
	}

	return hash % r.config.ShardCount
}

// buildInsertQuery 
func (r *AuditRepository) buildInsertQuery(tableName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s (
			id, timestamp, event_type, event_action, event_category,
			user_id, user_name, user_email, user_role,
			tenant_id, tenant_name,
			resource_id, resource_type, resource_name,
			request_id, session_id, correlation_id,
			ip_address, user_agent, request_method, request_url,
			response_status, response_size,
			changes, old_values, new_values,
			security_level, risk_score, threat_indicators,
			metadata, tags, custom_fields,
			source_system, source_component, source_version,
			compliance_tags, data_classification, retention_period,
			created_at, updated_at
		) VALUES (
			:id, :timestamp, :event_type, :event_action, :event_category,
			:user_id, :user_name, :user_email, :user_role,
			:tenant_id, :tenant_name,
			:resource_id, :resource_type, :resource_name,
			:request_id, :session_id, :correlation_id,
			:ip_address, :user_agent, :request_method, :request_url,
			:response_status, :response_size,
			:changes, :old_values, :new_values,
			:security_level, :risk_score, :threat_indicators,
			:metadata, :tags, :custom_fields,
			:source_system, :source_component, :source_version,
			:compliance_tags, :data_classification, :retention_period,
			:created_at, :updated_at
		)`, tableName)
}

// buildSelectQuery 
func (r *AuditRepository) buildSelectQuery(query *audit.AuditQuery) (string, string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// WHERE
	if query.EventTypes != nil && len(query.EventTypes) > 0 {
		placeholders := make([]string, len(query.EventTypes))
		for i, eventType := range query.EventTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, eventType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if query.EventActions != nil && len(query.EventActions) > 0 {
		placeholders := make([]string, len(query.EventActions))
		for i, action := range query.EventActions {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, action)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("event_action IN (%s)", strings.Join(placeholders, ",")))
	}

	if query.UserIDs != nil && len(query.UserIDs) > 0 {
		placeholders := make([]string, len(query.UserIDs))
		for i, userID := range query.UserIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, userID)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("user_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if query.TenantIDs != nil && len(query.TenantIDs) > 0 {
		placeholders := make([]string, len(query.TenantIDs))
		for i, tenantID := range query.TenantIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, tenantID)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("tenant_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if query.ResourceIDs != nil && len(query.ResourceIDs) > 0 {
		placeholders := make([]string, len(query.ResourceIDs))
		for i, resourceID := range query.ResourceIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, resourceID)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("resource_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if query.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *query.StartTime)
		argIndex++
	}

	if query.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, *query.EndTime)
		argIndex++
	}

	if query.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, query.IPAddress)
		argIndex++
	}

	if query.SecurityLevel != "" {
		conditions = append(conditions, fmt.Sprintf("security_level = $%d", argIndex))
		args = append(args, query.SecurityLevel)
		argIndex++
	}

	// WHERE
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// ORDER BY
	orderBy := "ORDER BY timestamp DESC"
	if query.SortBy != "" {
		direction := "ASC"
		if query.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", query.SortBy, direction)
	}

	// LIMITOFFSET
	limit := fmt.Sprintf("LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	// 
	selectQuery := fmt.Sprintf(`
		SELECT * FROM %s 
		%s 
		%s 
		%s`,
		r.config.TableName, whereClause, orderBy, limit)

	// 
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		%s`,
		r.config.TableName, whereClause)

	return selectQuery, countQuery, args
}

// insertEvent 
func (r *AuditRepository) insertEvent(ctx context.Context, stmt *sqlx.NamedStmt, event *audit.AuditEvent) error {
	// 
	row := r.eventToRow(event)
	_, err := stmt.ExecContext(ctx, row)
	return err
}

// scanEvent 
func (r *AuditRepository) scanEvent(rows *sqlx.Rows) (*audit.AuditEvent, error) {
	var row AuditEventRow
	if err := rows.StructScan(&row); err != nil {
		return nil, err
	}
	return r.rowToEvent(&row)
}

// eventToRow 
func (r *AuditRepository) eventToRow(event *audit.AuditEvent) *AuditEventRow {
	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = now
	}

	return &AuditEventRow{
		ID:                 event.ID,
		Timestamp:          event.Timestamp,
		EventType:          string(event.EventType),
		EventAction:        event.EventAction,
		EventCategory:      event.EventCategory,
		UserID:             event.UserID,
		UserName:           event.UserName,
		UserEmail:          event.UserEmail,
		UserRole:           event.UserRole,
		TenantID:           event.TenantID,
		TenantName:         event.TenantName,
		ResourceID:         event.ResourceID,
		ResourceType:       event.ResourceType,
		ResourceName:       event.ResourceName,
		RequestID:          event.RequestID,
		SessionID:          event.SessionID,
		CorrelationID:      event.CorrelationID,
		IPAddress:          event.IPAddress,
		UserAgent:          event.UserAgent,
		RequestMethod:      event.RequestMethod,
		RequestURL:         event.RequestURL,
		ResponseStatus:     event.ResponseStatus,
		ResponseSize:       event.ResponseSize,
		Changes:            JSONField(event.Changes),
		OldValues:          JSONField(event.OldValues),
		NewValues:          JSONField(event.NewValues),
		SecurityLevel:      string(event.SecurityLevel),
		RiskScore:          event.RiskScore,
		ThreatIndicators:   StringSliceField(event.ThreatIndicators),
		Metadata:           JSONField(event.Metadata),
		Tags:               StringSliceField(event.Tags),
		CustomFields:       JSONField(event.CustomFields),
		SourceSystem:       event.SourceSystem,
		SourceComponent:    event.SourceComponent,
		SourceVersion:      event.SourceVersion,
		ComplianceTags:     StringSliceField(event.ComplianceTags),
		DataClassification: event.DataClassification,
		RetentionPeriod:    event.RetentionPeriod,
		CreatedAt:          event.CreatedAt,
		UpdatedAt:          event.UpdatedAt,
	}
}

// rowToEvent 
func (r *AuditRepository) rowToEvent(row *AuditEventRow) (*audit.AuditEvent, error) {
	event := &audit.AuditEvent{
		ID:                 row.ID,
		Timestamp:          row.Timestamp,
		EventType:          audit.EventType(row.EventType),
		EventAction:        row.EventAction,
		EventCategory:      row.EventCategory,
		UserID:             row.UserID,
		UserName:           row.UserName,
		UserEmail:          row.UserEmail,
		UserRole:           row.UserRole,
		TenantID:           row.TenantID,
		TenantName:         row.TenantName,
		ResourceID:         row.ResourceID,
		ResourceType:       row.ResourceType,
		ResourceName:       row.ResourceName,
		RequestID:          row.RequestID,
		SessionID:          row.SessionID,
		CorrelationID:      row.CorrelationID,
		IPAddress:          row.IPAddress,
		UserAgent:          row.UserAgent,
		RequestMethod:      row.RequestMethod,
		RequestURL:         row.RequestURL,
		ResponseStatus:     row.ResponseStatus,
		ResponseSize:       row.ResponseSize,
		SecurityLevel:      audit.SecurityLevel(row.SecurityLevel),
		RiskScore:          row.RiskScore,
		ThreatIndicators:   []string(row.ThreatIndicators),
		Tags:               []string(row.Tags),
		SourceSystem:       row.SourceSystem,
		SourceComponent:    row.SourceComponent,
		SourceVersion:      row.SourceVersion,
		ComplianceTags:     []string(row.ComplianceTags),
		DataClassification: row.DataClassification,
		RetentionPeriod:    row.RetentionPeriod,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}

	// JSON
	if len(row.Changes) > 0 {
		if err := json.Unmarshal(row.Changes, &event.Changes); err != nil {
			r.logger.Warn("Failed to unmarshal changes", zap.Error(err))
		}
	}

	if len(row.OldValues) > 0 {
		if err := json.Unmarshal(row.OldValues, &event.OldValues); err != nil {
			r.logger.Warn("Failed to unmarshal old values", zap.Error(err))
		}
	}

	if len(row.NewValues) > 0 {
		if err := json.Unmarshal(row.NewValues, &event.NewValues); err != nil {
			r.logger.Warn("Failed to unmarshal new values", zap.Error(err))
		}
	}

	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &event.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal metadata", zap.Error(err))
		}
	}

	if len(row.CustomFields) > 0 {
		if err := json.Unmarshal(row.CustomFields, &event.CustomFields); err != nil {
			r.logger.Warn("Failed to unmarshal custom fields", zap.Error(err))
		}
	}

	return event, nil
}

// getBasicStatistics 
func (r *AuditRepository) getBasicStatistics(ctx context.Context, filter *audit.StatisticsFilter, stats *audit.AuditStatistics) error {
	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(DISTINCT tenant_id) as unique_tenants,
			COUNT(DISTINCT resource_id) as unique_resources
		FROM ` + r.config.TableName + `
		WHERE timestamp BETWEEN $1 AND $2`

	args := []interface{}{filter.StartTime, filter.EndTime}

	if filter.TenantID != "" {
		query += " AND tenant_id = $3"
		args = append(args, filter.TenantID)
	}

	row := r.db.QueryRowxContext(ctx, query, args...)
	return row.Scan(&stats.TotalEvents, &stats.UniqueUsers, &stats.UniqueTenants, &stats.UniqueResources)
}

// getUserActivityStatistics 
func (r *AuditRepository) getUserActivityStatistics(ctx context.Context, filter *audit.StatisticsFilter, stats *audit.AuditStatistics) error {
	query := `
		SELECT 
			user_id,
			user_name,
			COUNT(*) as event_count,
			COUNT(DISTINCT event_type) as unique_event_types,
			MIN(timestamp) as first_activity,
			MAX(timestamp) as last_activity
		FROM ` + r.config.TableName + `
		WHERE timestamp BETWEEN $1 AND $2`

	args := []interface{}{filter.StartTime, filter.EndTime}

	if filter.TenantID != "" {
		query += " AND tenant_id = $3"
		args = append(args, filter.TenantID)
	}

	query += `
		GROUP BY user_id, user_name
		ORDER BY event_count DESC
		LIMIT 10`

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userActivities []audit.UserActivity
	for rows.Next() {
		var activity audit.UserActivity
		if err := rows.Scan(
			&activity.UserID,
			&activity.UserName,
			&activity.EventCount,
			&activity.UniqueEventTypes,
			&activity.FirstActivity,
			&activity.LastActivity,
		); err != nil {
			return err
		}
		userActivities = append(userActivities, activity)
	}

	stats.UserActivities = userActivities
	return nil
}

// getResourceActivityStatistics 
func (r *AuditRepository) getResourceActivityStatistics(ctx context.Context, filter *audit.StatisticsFilter, stats *audit.AuditStatistics) error {
	query := `
		SELECT 
			resource_id,
			resource_type,
			resource_name,
			COUNT(*) as event_count,
			COUNT(DISTINCT user_id) as unique_users,
			MIN(timestamp) as first_access,
			MAX(timestamp) as last_access
		FROM ` + r.config.TableName + `
		WHERE timestamp BETWEEN $1 AND $2`

	args := []interface{}{filter.StartTime, filter.EndTime}

	if filter.TenantID != "" {
		query += " AND tenant_id = $3"
		args = append(args, filter.TenantID)
	}

	query += `
		GROUP BY resource_id, resource_type, resource_name
		ORDER BY event_count DESC
		LIMIT 10`

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var resourceActivities []audit.ResourceActivity
	for rows.Next() {
		var activity audit.ResourceActivity
		if err := rows.Scan(
			&activity.ResourceID,
			&activity.ResourceType,
			&activity.ResourceName,
			&activity.EventCount,
			&activity.UniqueUsers,
			&activity.FirstAccess,
			&activity.LastAccess,
		); err != nil {
			return err
		}
		resourceActivities = append(resourceActivities, activity)
	}

	stats.ResourceActivities = resourceActivities
	return nil
}

// getSecurityStatistics 
func (r *AuditRepository) getSecurityStatistics(ctx context.Context, filter *audit.StatisticsFilter, stats *audit.AuditStatistics) error {
	query := `
		SELECT 
			COUNT(CASE WHEN security_level = 'HIGH' THEN 1 END) as high_risk_events,
			COUNT(CASE WHEN security_level = 'MEDIUM' THEN 1 END) as medium_risk_events,
			COUNT(CASE WHEN security_level = 'LOW' THEN 1 END) as low_risk_events,
			COUNT(CASE WHEN event_action LIKE '%failed%' OR event_action LIKE '%error%' THEN 1 END) as failed_events,
			COUNT(DISTINCT ip_address) as unique_ip_addresses,
			AVG(risk_score) as average_risk_score
		FROM ` + r.config.TableName + `
		WHERE timestamp BETWEEN $1 AND $2`

	args := []interface{}{filter.StartTime, filter.EndTime}

	if filter.TenantID != "" {
		query += " AND tenant_id = $3"
		args = append(args, filter.TenantID)
	}

	row := r.db.QueryRowxContext(ctx, query, args...)

	securityStats := &audit.SecurityStatistics{}
	err := row.Scan(
		&securityStats.HighRiskEvents,
		&securityStats.MediumRiskEvents,
		&securityStats.LowRiskEvents,
		&securityStats.FailedEvents,
		&securityStats.UniqueIPAddresses,
		&securityStats.AverageRiskScore,
	)

	stats.SecurityStatistics = *securityStats
	return err
}

// AuditEventRow 
type AuditEventRow struct {
	ID                 string           `db:"id"`
	Timestamp          time.Time        `db:"timestamp"`
	EventType          string           `db:"event_type"`
	EventAction        string           `db:"event_action"`
	EventCategory      string           `db:"event_category"`
	UserID             string           `db:"user_id"`
	UserName           string           `db:"user_name"`
	UserEmail          string           `db:"user_email"`
	UserRole           string           `db:"user_role"`
	TenantID           string           `db:"tenant_id"`
	TenantName         string           `db:"tenant_name"`
	ResourceID         string           `db:"resource_id"`
	ResourceType       string           `db:"resource_type"`
	ResourceName       string           `db:"resource_name"`
	RequestID          string           `db:"request_id"`
	SessionID          string           `db:"session_id"`
	CorrelationID      string           `db:"correlation_id"`
	IPAddress          string           `db:"ip_address"`
	UserAgent          string           `db:"user_agent"`
	RequestMethod      string           `db:"request_method"`
	RequestURL         string           `db:"request_url"`
	ResponseStatus     int              `db:"response_status"`
	ResponseSize       int64            `db:"response_size"`
	Changes            JSONField        `db:"changes"`
	OldValues          JSONField        `db:"old_values"`
	NewValues          JSONField        `db:"new_values"`
	SecurityLevel      string           `db:"security_level"`
	RiskScore          float64          `db:"risk_score"`
	ThreatIndicators   StringSliceField `db:"threat_indicators"`
	Metadata           JSONField        `db:"metadata"`
	Tags               StringSliceField `db:"tags"`
	CustomFields       JSONField        `db:"custom_fields"`
	SourceSystem       string           `db:"source_system"`
	SourceComponent    string           `db:"source_component"`
	SourceVersion      string           `db:"source_version"`
	ComplianceTags     StringSliceField `db:"compliance_tags"`
	DataClassification string           `db:"data_classification"`
	RetentionPeriod    time.Duration    `db:"retention_period"`
	CreatedAt          time.Time        `db:"created_at"`
	UpdatedAt          time.Time        `db:"updated_at"`
}

// JSONField JSON
type JSONField []byte

// Value driver.Valuer
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan sql.Scanner
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*j = make(JSONField, len(v))
		copy(*j, v)
	case string:
		*j = JSONField(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONField", value)
	}

	return nil
}

// StringSliceField 
type StringSliceField []string

// Value driver.Valuer
func (s StringSliceField) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal([]string(s))
}

// Scan sql.Scanner
func (s *StringSliceField) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringSliceField", value)
	}

	var slice []string
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}

	*s = StringSliceField(slice)
	return nil
}

