package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/taishanglaojun/core-services/monitoring/models"
)

// DatabaseCollector 数据库指标收集器
type DatabaseCollector struct {
	name     string
	interval time.Duration
	enabled  bool
	labels   map[string]string
	
	// 数据库连接
	db       *sql.DB
	dbType   string // postgres, mysql, sqlite等
	dbName   string
	
	// 配置选项
	collectConnections bool
	collectQueries     bool
	collectTables      bool
	collectIndexes     bool
	collectLocks       bool
	collectReplication bool
	
	// 查询缓存
	queryCache map[string]*QueryStats
	
	// 统计信息
	lastCollectTime time.Time
}

// DatabaseCollectorConfig 数据库收集器配置
type DatabaseCollectorConfig struct {
	Interval           time.Duration     `yaml:"interval"`
	Enabled            bool              `yaml:"enabled"`
	Labels             map[string]string `yaml:"labels"`
	DBType             string            `yaml:"db_type"`
	DBName             string            `yaml:"db_name"`
	CollectConnections bool              `yaml:"collect_connections"`
	CollectQueries     bool              `yaml:"collect_queries"`
	CollectTables      bool              `yaml:"collect_tables"`
	CollectIndexes     bool              `yaml:"collect_indexes"`
	CollectLocks       bool              `yaml:"collect_locks"`
	CollectReplication bool              `yaml:"collect_replication"`
}

// QueryStats 查询统计
type QueryStats struct {
	Query         string
	Count         uint64
	TotalTime     time.Duration
	MinTime       time.Duration
	MaxTime       time.Duration
	AvgTime       time.Duration
	LastExecution time.Time
	Errors        uint64
}

// TableStats 表统计
type TableStats struct {
	TableName    string
	RowCount     uint64
	Size         uint64
	IndexSize    uint64
	LastVacuum   time.Time
	LastAnalyze  time.Time
	SeqScans     uint64
	IndexScans   uint64
	Inserts      uint64
	Updates      uint64
	Deletes      uint64
}

// IndexStats 索引统计
type IndexStats struct {
	IndexName  string
	TableName  string
	Size       uint64
	Scans      uint64
	TuplesRead uint64
	TuplesFetched uint64
}

// LockStats 锁统计
type LockStats struct {
	LockType     string
	Mode         string
	Count        uint64
	WaitingCount uint64
	MaxWaitTime  time.Duration
}

// ReplicationStats 复制统计
type ReplicationStats struct {
	Role           string // master/slave
	State          string
	Lag            time.Duration
	BytesReceived  uint64
	BytesSent      uint64
	LastReceived   time.Time
	ConnectedSlaves int
}

// NewDatabaseCollector 创建数据库指标收集器
func NewDatabaseCollector(config DatabaseCollectorConfig, db *sql.DB) *DatabaseCollector {
	labels := map[string]string{
		"collector": "database",
		"db_type":   config.DBType,
		"db_name":   config.DBName,
	}
	
	// 添加自定义标签
	for k, v := range config.Labels {
		labels[k] = v
	}
	
	return &DatabaseCollector{
		name:               "database",
		interval:           config.Interval,
		enabled:            config.Enabled,
		labels:             labels,
		db:                 db,
		dbType:             config.DBType,
		dbName:             config.DBName,
		collectConnections: config.CollectConnections,
		collectQueries:     config.CollectQueries,
		collectTables:      config.CollectTables,
		collectIndexes:     config.CollectIndexes,
		collectLocks:       config.CollectLocks,
		collectReplication: config.CollectReplication,
		queryCache:         make(map[string]*QueryStats),
		lastCollectTime:    time.Now(),
	}
}

// GetName 获取收集器名称
func (c *DatabaseCollector) GetName() string {
	return c.name
}

// GetCategory 获取收集器分类
func (c *DatabaseCollector) GetCategory() models.MetricCategory {
	return models.CategoryDatabase
}

// GetInterval 获取收集间隔
func (c *DatabaseCollector) GetInterval() time.Duration {
	return c.interval
}

// IsEnabled 检查是否启用
func (c *DatabaseCollector) IsEnabled() bool {
	return c.enabled
}

// Start 启动收集器
func (c *DatabaseCollector) Start(ctx context.Context) error {
	if !c.enabled {
		return nil
	}
	
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := c.Collect(ctx); err != nil {
				fmt.Printf("Database collector error: %v\n", err)
			}
		}
	}
}

// Stop 停止收集器
func (c *DatabaseCollector) Stop() error {
	c.enabled = false
	return nil
}

// Health 健康检查
func (c *DatabaseCollector) Health() error {
	if !c.enabled {
		return fmt.Errorf("database collector is disabled")
	}
	
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// 检查数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}

// Collect 收集指标
func (c *DatabaseCollector) Collect(ctx context.Context) ([]models.Metric, error) {
	if !c.enabled || c.db == nil {
		return nil, nil
	}
	
	var metrics []models.Metric
	now := time.Now()
	
	// 收集连接池指标
	if c.collectConnections {
		connMetrics, err := c.collectConnectionMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect connection metrics: %w", err)
		}
		metrics = append(metrics, connMetrics...)
	}
	
	// 收集查询统计
	if c.collectQueries {
		queryMetrics, err := c.collectQueryMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect query metrics: %w", err)
		}
		metrics = append(metrics, queryMetrics...)
	}
	
	// 收集表统计
	if c.collectTables {
		tableMetrics, err := c.collectTableMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect table metrics: %w", err)
		}
		metrics = append(metrics, tableMetrics...)
	}
	
	// 收集索引统计
	if c.collectIndexes {
		indexMetrics, err := c.collectIndexMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect index metrics: %w", err)
		}
		metrics = append(metrics, indexMetrics...)
	}
	
	// 收集锁统计
	if c.collectLocks {
		lockMetrics, err := c.collectLockMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect lock metrics: %w", err)
		}
		metrics = append(metrics, lockMetrics...)
	}
	
	// 收集复制统计
	if c.collectReplication {
		replMetrics, err := c.collectReplicationMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect replication metrics: %w", err)
		}
		metrics = append(metrics, replMetrics...)
	}
	
	c.lastCollectTime = now
	return metrics, nil
}

// collectConnectionMetrics 收集连接池指标
func (c *DatabaseCollector) collectConnectionMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	stats := c.db.Stats()
	
	// 最大打开连接数
	metric := models.NewMetric("database_connections_max_open", models.MetricTypeGauge, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.MaxOpenConnections)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Maximum number of open connections"
	metrics = append(metrics, *metric)
	
	// 当前打开连接数
	metric = models.NewMetric("database_connections_open", models.MetricTypeGauge, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.OpenConnections)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Current number of open connections"
	metrics = append(metrics, *metric)
	
	// 使用中连接数
	metric = models.NewMetric("database_connections_in_use", models.MetricTypeGauge, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.InUse)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Number of connections in use"
	metrics = append(metrics, *metric)
	
	// 空闲连接数
	metric = models.NewMetric("database_connections_idle", models.MetricTypeGauge, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.Idle)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Number of idle connections"
	metrics = append(metrics, *metric)
	
	// 等待连接数
	metric = models.NewMetric("database_connections_wait_count", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.WaitCount)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "count"
	metric.Description = "Total number of connections waited for"
	metrics = append(metrics, *metric)
	
	// 等待时间
	metric = models.NewMetric("database_connections_wait_duration_seconds", models.MetricTypeGauge, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(stats.WaitDuration.Seconds()).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "seconds"
	metric.Description = "Total time blocked waiting for connections"
	metrics = append(metrics, *metric)
	
	// 最大空闲关闭数
	metric = models.NewMetric("database_connections_max_idle_closed", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.MaxIdleClosed)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "count"
	metric.Description = "Total number of connections closed due to max idle"
	metrics = append(metrics, *metric)
	
	// 最大生命周期关闭数
	metric = models.NewMetric("database_connections_max_lifetime_closed", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(stats.MaxLifetimeClosed)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "count"
	metric.Description = "Total number of connections closed due to max lifetime"
	metrics = append(metrics, *metric)
	
	// 连接利用率
	if stats.MaxOpenConnections > 0 {
		utilization := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
		metric = models.NewMetric("database_connections_utilization_percent", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(c.labels).
			WithValue(utilization).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Connection pool utilization percentage"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectQueryMetrics 收集查询统计
func (c *DatabaseCollector) collectQueryMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 根据数据库类型执行不同的查询
	switch strings.ToLower(c.dbType) {
	case "postgres", "postgresql":
		return c.collectPostgreSQLQueryMetrics(ctx, timestamp)
	case "mysql":
		return c.collectMySQLQueryMetrics(ctx, timestamp)
	case "sqlite":
		return c.collectSQLiteQueryMetrics(ctx, timestamp)
	default:
		// 通用查询统计
		return c.collectGenericQueryMetrics(timestamp)
	}
}

// collectPostgreSQLQueryMetrics 收集PostgreSQL查询统计
func (c *DatabaseCollector) collectPostgreSQLQueryMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 查询pg_stat_statements获取查询统计
	query := `
		SELECT 
			query,
			calls,
			total_time,
			min_time,
			max_time,
			mean_time,
			rows
		FROM pg_stat_statements 
		ORDER BY total_time DESC 
		LIMIT 100
	`
	
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		// pg_stat_statements可能未启用，返回空指标
		return metrics, nil
	}
	defer rows.Close()
	
	var totalQueries uint64
	var totalTime float64
	
	for rows.Next() {
		var query string
		var calls uint64
		var totalTimeMs, minTimeMs, maxTimeMs, meanTimeMs float64
		var rowsAffected uint64
		
		if err := rows.Scan(&query, &calls, &totalTimeMs, &minTimeMs, &maxTimeMs, &meanTimeMs, &rowsAffected); err != nil {
			continue
		}
		
		totalQueries += calls
		totalTime += totalTimeMs
		
		// 为每个查询创建指标（限制查询长度）
		queryHash := fmt.Sprintf("%x", query)
		if len(queryHash) > 16 {
			queryHash = queryHash[:16]
		}
		
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["query_hash"] = queryHash
		
		// 查询调用次数
		metric := models.NewMetric("database_query_calls_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(calls)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "calls"
		metric.Description = "Total number of query calls"
		metrics = append(metrics, *metric)
		
		// 查询平均时间
		metric = models.NewMetric("database_query_duration_avg_seconds", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(meanTimeMs / 1000).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Average query duration"
		metrics = append(metrics, *metric)
		
		// 查询最大时间
		metric = models.NewMetric("database_query_duration_max_seconds", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(maxTimeMs / 1000).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Maximum query duration"
		metrics = append(metrics, *metric)
		
		// 影响行数
		metric = models.NewMetric("database_query_rows_affected", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(rowsAffected)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "rows"
		metric.Description = "Number of rows affected by query"
		metrics = append(metrics, *metric)
	}
	
	// 总查询数
	metric := models.NewMetric("database_queries_total", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(totalQueries)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "queries"
	metric.Description = "Total number of queries"
	metrics = append(metrics, *metric)
	
	// 总查询时间
	metric = models.NewMetric("database_query_time_total_seconds", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(totalTime / 1000).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "seconds"
	metric.Description = "Total query time"
	metrics = append(metrics, *metric)
	
	// 平均QPS
	timeDiff := timestamp.Sub(c.lastCollectTime).Seconds()
	if timeDiff > 0 {
		qps := float64(totalQueries) / timeDiff
		metric = models.NewMetric("database_queries_per_second", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(c.labels).
			WithValue(qps).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "queries_per_second"
		metric.Description = "Queries per second"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectMySQLQueryMetrics 收集MySQL查询统计
func (c *DatabaseCollector) collectMySQLQueryMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 查询performance_schema获取查询统计
	query := `
		SELECT 
			DIGEST_TEXT,
			COUNT_STAR,
			SUM_TIMER_WAIT,
			MIN_TIMER_WAIT,
			MAX_TIMER_WAIT,
			AVG_TIMER_WAIT,
			SUM_ROWS_EXAMINED,
			SUM_ROWS_SENT
		FROM performance_schema.events_statements_summary_by_digest 
		ORDER BY SUM_TIMER_WAIT DESC 
		LIMIT 100
	`
	
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		// performance_schema可能未启用，返回空指标
		return metrics, nil
	}
	defer rows.Close()
	
	var totalQueries uint64
	var totalTime uint64
	
	for rows.Next() {
		var digestText sql.NullString
		var countStar uint64
		var sumTimerWait, minTimerWait, maxTimerWait, avgTimerWait uint64
		var sumRowsExamined, sumRowsSent uint64
		
		if err := rows.Scan(&digestText, &countStar, &sumTimerWait, &minTimerWait, &maxTimerWait, &avgTimerWait, &sumRowsExamined, &sumRowsSent); err != nil {
			continue
		}
		
		totalQueries += countStar
		totalTime += sumTimerWait
		
		// 为每个查询创建指标
		queryHash := "unknown"
		if digestText.Valid && len(digestText.String) > 0 {
			queryHash = fmt.Sprintf("%x", digestText.String)
			if len(queryHash) > 16 {
				queryHash = queryHash[:16]
			}
		}
		
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["query_hash"] = queryHash
		
		// 查询调用次数
		metric := models.NewMetric("database_query_calls_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(countStar)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "calls"
		metric.Description = "Total number of query calls"
		metrics = append(metrics, *metric)
		
		// 查询平均时间（MySQL timer单位是皮秒）
		metric = models.NewMetric("database_query_duration_avg_seconds", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(avgTimerWait) / 1e12).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Average query duration"
		metrics = append(metrics, *metric)
		
		// 检查行数
		metric = models.NewMetric("database_query_rows_examined", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(sumRowsExamined)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "rows"
		metric.Description = "Number of rows examined by query"
		metrics = append(metrics, *metric)
		
		// 返回行数
		metric = models.NewMetric("database_query_rows_sent", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(sumRowsSent)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "rows"
		metric.Description = "Number of rows sent by query"
		metrics = append(metrics, *metric)
	}
	
	// 总查询数
	metric := models.NewMetric("database_queries_total", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(totalQueries)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "queries"
	metric.Description = "Total number of queries"
	metrics = append(metrics, *metric)
	
	// 总查询时间
	metric = models.NewMetric("database_query_time_total_seconds", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(totalTime) / 1e12).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "seconds"
	metric.Description = "Total query time"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// collectSQLiteQueryMetrics 收集SQLite查询统计
func (c *DatabaseCollector) collectSQLiteQueryMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// SQLite没有内置的查询统计，使用基本指标
	return c.collectGenericQueryMetrics(timestamp)
}

// collectGenericQueryMetrics 收集通用查询统计
func (c *DatabaseCollector) collectGenericQueryMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 使用缓存的查询统计
	var totalQueries uint64
	var totalTime time.Duration
	
	for _, stats := range c.queryCache {
		totalQueries += stats.Count
		totalTime += stats.TotalTime
	}
	
	// 总查询数
	metric := models.NewMetric("database_queries_total", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(totalQueries)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "queries"
	metric.Description = "Total number of queries"
	metrics = append(metrics, *metric)
	
	// 总查询时间
	metric = models.NewMetric("database_query_time_total_seconds", models.MetricTypeCounter, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(totalTime.Seconds()).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "seconds"
	metric.Description = "Total query time"
	metrics = append(metrics, *metric)
	
	// 平均查询时间
	if totalQueries > 0 {
		avgTime := totalTime / time.Duration(totalQueries)
		metric = models.NewMetric("database_query_duration_avg_seconds", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(c.labels).
			WithValue(avgTime.Seconds()).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Average query duration"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectTableMetrics 收集表统计
func (c *DatabaseCollector) collectTableMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 根据数据库类型执行不同的查询
	switch strings.ToLower(c.dbType) {
	case "postgres", "postgresql":
		return c.collectPostgreSQLTableMetrics(ctx, timestamp)
	case "mysql":
		return c.collectMySQLTableMetrics(ctx, timestamp)
	default:
		return metrics, nil
	}
}

// collectPostgreSQLTableMetrics 收集PostgreSQL表统计
func (c *DatabaseCollector) collectPostgreSQLTableMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	query := `
		SELECT 
			schemaname,
			tablename,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			n_live_tup,
			n_dead_tup,
			seq_scan,
			idx_scan
		FROM pg_stat_user_tables
	`
	
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var schemaName, tableName string
		var inserts, updates, deletes, liveTuples, deadTuples, seqScans, idxScans uint64
		
		if err := rows.Scan(&schemaName, &tableName, &inserts, &updates, &deletes, &liveTuples, &deadTuples, &seqScans, &idxScans); err != nil {
			continue
		}
		
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["schema"] = schemaName
		labels["table"] = tableName
		
		// 表行数
		metric := models.NewMetric("database_table_rows", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(liveTuples)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "rows"
		metric.Description = "Number of live rows in table"
		metrics = append(metrics, *metric)
		
		// 死行数
		metric = models.NewMetric("database_table_dead_rows", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(deadTuples)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "rows"
		metric.Description = "Number of dead rows in table"
		metrics = append(metrics, *metric)
		
		// 插入数
		metric = models.NewMetric("database_table_inserts_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(inserts)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "operations"
		metric.Description = "Total number of inserts"
		metrics = append(metrics, *metric)
		
		// 更新数
		metric = models.NewMetric("database_table_updates_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(updates)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "operations"
		metric.Description = "Total number of updates"
		metrics = append(metrics, *metric)
		
		// 删除数
		metric = models.NewMetric("database_table_deletes_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(deletes)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "operations"
		metric.Description = "Total number of deletes"
		metrics = append(metrics, *metric)
		
		// 顺序扫描数
		metric = models.NewMetric("database_table_seq_scans_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(seqScans)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "scans"
		metric.Description = "Total number of sequential scans"
		metrics = append(metrics, *metric)
		
		// 索引扫描数
		metric = models.NewMetric("database_table_index_scans_total", models.MetricTypeCounter, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(idxScans)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "scans"
		metric.Description = "Total number of index scans"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectMySQLTableMetrics 收集MySQL表统计
func (c *DatabaseCollector) collectMySQLTableMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	query := `
		SELECT 
			TABLE_SCHEMA,
			TABLE_NAME,
			TABLE_ROWS,
			DATA_LENGTH,
			INDEX_LENGTH
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ?
	`
	
	rows, err := c.db.QueryContext(ctx, query, c.dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var schemaName, tableName string
		var tableRows, dataLength, indexLength uint64
		
		if err := rows.Scan(&schemaName, &tableName, &tableRows, &dataLength, &indexLength); err != nil {
			continue
		}
		
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["schema"] = schemaName
		labels["table"] = tableName
		
		// 表行数
		metric := models.NewMetric("database_table_rows", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(tableRows)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "rows"
		metric.Description = "Number of rows in table"
		metrics = append(metrics, *metric)
		
		// 数据大小
		metric = models.NewMetric("database_table_data_size_bytes", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(dataLength)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Table data size in bytes"
		metrics = append(metrics, *metric)
		
		// 索引大小
		metric = models.NewMetric("database_table_index_size_bytes", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(indexLength)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Table index size in bytes"
		metrics = append(metrics, *metric)
		
		// 总大小
		totalSize := dataLength + indexLength
		metric = models.NewMetric("database_table_total_size_bytes", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(labels).
			WithValue(float64(totalSize)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Total table size in bytes"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectIndexMetrics 收集索引统计
func (c *DatabaseCollector) collectIndexMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 根据数据库类型实现索引统计收集
	// 这里简化实现，实际应该根据不同数据库类型执行相应查询
	
	return metrics, nil
}

// collectLockMetrics 收集锁统计
func (c *DatabaseCollector) collectLockMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 根据数据库类型实现锁统计收集
	// 这里简化实现，实际应该根据不同数据库类型执行相应查询
	
	return metrics, nil
}

// collectReplicationMetrics 收集复制统计
func (c *DatabaseCollector) collectReplicationMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 根据数据库类型实现复制统计收集
	// 这里简化实现，实际应该根据不同数据库类型执行相应查询
	
	return metrics, nil
}

// RecordQuery 记录查询统计
func (c *DatabaseCollector) RecordQuery(query string, duration time.Duration, err error) {
	// 简化查询文本作为key
	key := query
	if len(key) > 100 {
		key = key[:100]
	}
	
	stats, exists := c.queryCache[key]
	if !exists {
		stats = &QueryStats{
			Query:         query,
			MinTime:       duration,
			MaxTime:       duration,
			LastExecution: time.Now(),
		}
		c.queryCache[key] = stats
	}
	
	stats.Count++
	stats.TotalTime += duration
	stats.LastExecution = time.Now()
	
	if duration < stats.MinTime {
		stats.MinTime = duration
	}
	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}
	
	if stats.Count > 0 {
		stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)
	}
	
	if err != nil {
		stats.Errors++
	}
}

// 确保实现了接口
var _ interfaces.MetricCollector = (*DatabaseCollector)(nil)