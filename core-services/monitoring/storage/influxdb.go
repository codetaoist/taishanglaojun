package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/taishanglaojun/core-services/monitoring/models"
)

// InfluxDBStorage InfluxDB存储实现
type InfluxDBStorage struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	queryAPI api.QueryAPI
	config   *InfluxDBConfig
}

// InfluxDBConfig InfluxDB配置
type InfluxDBConfig struct {
	URL          string        `yaml:"url"`
	Token        string        `yaml:"token"`
	Organization string        `yaml:"organization"`
	Bucket       string        `yaml:"bucket"`
	Timeout      time.Duration `yaml:"timeout"`
	BatchSize    int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
	
	// 写入配置
	WriteOptions *WriteOptions `yaml:"write_options"`
	
	// 查询配置
	QueryOptions *QueryOptions `yaml:"query_options"`
	
	// 保留策略
	RetentionPolicy *RetentionPolicy `yaml:"retention_policy"`
}

// WriteOptions 写入选项
type WriteOptions struct {
	BatchSize        int           `yaml:"batch_size"`
	FlushInterval    time.Duration `yaml:"flush_interval"`
	RetryInterval    time.Duration `yaml:"retry_interval"`
	MaxRetries       int           `yaml:"max_retries"`
	MaxRetryInterval time.Duration `yaml:"max_retry_interval"`
	ExponentialBase  int           `yaml:"exponential_base"`
	UseGZip          bool          `yaml:"use_gzip"`
}

// QueryOptions 查询选项
type QueryOptions struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
	MaxQueryTime   time.Duration `yaml:"max_query_time"`
}

// RetentionPolicy 保留策略
type RetentionPolicy struct {
	Duration    time.Duration `yaml:"duration"`
	Replication int           `yaml:"replication"`
	ShardDuration time.Duration `yaml:"shard_duration"`
}

// NewInfluxDBStorage 创建InfluxDB存储
func NewInfluxDBStorage(config *InfluxDBConfig) (*InfluxDBStorage, error) {
	// 创建客户端选项
	options := influxdb2.DefaultOptions()
	
	if config.WriteOptions != nil {
		if config.WriteOptions.BatchSize > 0 {
			options = options.SetBatchSize(uint(config.WriteOptions.BatchSize))
		}
		if config.WriteOptions.FlushInterval > 0 {
			options = options.SetFlushInterval(uint(config.WriteOptions.FlushInterval.Milliseconds()))
		}
		if config.WriteOptions.RetryInterval > 0 {
			options = options.SetRetryInterval(uint(config.WriteOptions.RetryInterval.Milliseconds()))
		}
		if config.WriteOptions.MaxRetries > 0 {
			options = options.SetMaxRetries(uint(config.WriteOptions.MaxRetries))
		}
		if config.WriteOptions.UseGZip {
			options = options.SetUseGZip(true)
		}
	}
	
	// 创建客户端
	client := influxdb2.NewClientWithOptions(config.URL, config.Token, options)
	
	// 创建写入API
	writeAPI := client.WriteAPI(config.Organization, config.Bucket)
	
	// 创建查询API
	queryAPI := client.QueryAPI(config.Organization)
	
	storage := &InfluxDBStorage{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		config:   config,
	}
	
	// 启动错误处理协程
	go storage.handleWriteErrors()
	
	return storage, nil
}

// handleWriteErrors 处理写入错误
func (i *InfluxDBStorage) handleWriteErrors() {
	errorsCh := i.writeAPI.Errors()
	for err := range errorsCh {
		fmt.Printf("InfluxDB write error: %v\n", err)
	}
}

// Store 存储指标
func (i *InfluxDBStorage) Store(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		point, err := i.convertMetricToPoint(metric)
		if err != nil {
			return fmt.Errorf("failed to convert metric to point: %w", err)
		}
		
		// 写入数据点
		i.writeAPI.WritePoint(point)
	}
	
	// 强制刷新
	i.writeAPI.Flush()
	
	return nil
}

// convertMetricToPoint 转换指标为InfluxDB数据点
func (i *InfluxDBStorage) convertMetricToPoint(metric models.Metric) (*write.Point, error) {
	// 创建数据点
	point := influxdb2.NewPoint(
		metric.Name,
		metric.Labels,
		map[string]interface{}{
			"value": metric.Value,
		},
		metric.Timestamp,
	)
	
	// 添加额外字段
	if metric.Source != "" {
		point = point.AddTag("source", metric.Source)
	}
	
	if metric.Category != "" {
		point = point.AddTag("category", string(metric.Category))
	}
	
	if metric.Type != "" {
		point = point.AddTag("type", string(metric.Type))
	}
	
	// 根据指标类型添加特定字段
	switch metric.Type {
	case models.MetricTypeHistogram:
		if histogram, ok := metric.(*models.HistogramMetric); ok {
			fields := map[string]interface{}{
				"count": histogram.Count,
				"sum":   histogram.Sum,
			}
			
			// 添加桶数据
			for i, bucket := range histogram.Buckets {
				fields[fmt.Sprintf("bucket_%d", i)] = bucket.Count
				fields[fmt.Sprintf("bucket_%d_le", i)] = bucket.UpperBound
			}
			
			point = point.AddFields(fields)
		}
		
	case models.MetricTypeSummary:
		if summary, ok := metric.(*models.SummaryMetric); ok {
			fields := map[string]interface{}{
				"count": summary.Count,
				"sum":   summary.Sum,
			}
			
			// 添加分位数数据
			for quantile, value := range summary.Quantiles {
				fields[fmt.Sprintf("quantile_%s", strings.ReplaceAll(fmt.Sprintf("%.2f", quantile), ".", "_"))] = value
			}
			
			point = point.AddFields(fields)
		}
	}
	
	return point, nil
}

// Query 查询指标
func (i *InfluxDBStorage) Query(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	// 构建Flux查询
	fluxQuery, err := i.buildFluxQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build flux query: %w", err)
	}
	
	// 设置查询超时
	queryCtx := ctx
	if i.config.QueryOptions != nil && i.config.QueryOptions.DefaultTimeout > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, i.config.QueryOptions.DefaultTimeout)
		defer cancel()
	}
	
	// 执行查询
	result, err := i.queryAPI.Query(queryCtx, fluxQuery)
	if err != nil {
		return nil, fmt.Errorf("influxdb query failed: %w", err)
	}
	defer result.Close()
	
	return i.convertQueryResult(result, query)
}

// buildFluxQuery 构建Flux查询
func (i *InfluxDBStorage) buildFluxQuery(query *models.MetricQuery) (string, error) {
	var fluxQuery strings.Builder
	
	// 基本查询结构
	fluxQuery.WriteString(fmt.Sprintf(`from(bucket: "%s")`, i.config.Bucket))
	
	// 时间范围
	if !query.Start.IsZero() && !query.End.IsZero() {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> range(start: %s, stop: %s)`, 
			query.Start.Format(time.RFC3339), 
			query.End.Format(time.RFC3339)))
	} else if !query.Start.IsZero() {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> range(start: %s)`, query.Start.Format(time.RFC3339)))
	} else {
		// 默认查询最近1小时
		fluxQuery.WriteString(`
  |> range(start: -1h)`)
	}
	
	// 指标名称过滤
	if query.MetricName != "" {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> filter(fn: (r) => r._measurement == "%s")`, query.MetricName))
	}
	
	// 标签过滤
	for key, value := range query.Labels {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> filter(fn: (r) => r.%s == "%s")`, key, value))
	}
	
	// 字段过滤
	if query.Field != "" {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> filter(fn: (r) => r._field == "%s")`, query.Field))
	} else {
		fluxQuery.WriteString(`
  |> filter(fn: (r) => r._field == "value")`)
	}
	
	// 聚合
	if query.Aggregation != "" {
		switch query.Aggregation {
		case "mean", "avg":
			fluxQuery.WriteString(`
  |> mean()`)
		case "sum":
			fluxQuery.WriteString(`
  |> sum()`)
		case "max":
			fluxQuery.WriteString(`
  |> max()`)
		case "min":
			fluxQuery.WriteString(`
  |> min()`)
		case "count":
			fluxQuery.WriteString(`
  |> count()`)
		case "last":
			fluxQuery.WriteString(`
  |> last()`)
		case "first":
			fluxQuery.WriteString(`
  |> first()`)
		}
	}
	
	// 分组
	if len(query.GroupBy) > 0 {
		groupBy := strings.Join(query.GroupBy, `", "`)
		fluxQuery.WriteString(fmt.Sprintf(`
  |> group(columns: ["%s"])`, groupBy))
	}
	
	// 窗口聚合
	if query.Step > 0 {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> aggregateWindow(every: %s, fn: mean, createEmpty: false)`, query.Step.String()))
	}
	
	// 排序
	if query.OrderBy != "" {
		desc := ""
		if query.OrderDesc {
			desc = ", desc: true"
		}
		fluxQuery.WriteString(fmt.Sprintf(`
  |> sort(columns: ["%s"]%s)`, query.OrderBy, desc))
	}
	
	// 限制
	if query.Limit > 0 {
		fluxQuery.WriteString(fmt.Sprintf(`
  |> limit(n: %d)`, query.Limit))
	}
	
	return fluxQuery.String(), nil
}

// convertQueryResult 转换查询结果
func (i *InfluxDBStorage) convertQueryResult(result *api.QueryTableResult, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	queryResult := &models.MetricQueryResult{
		Query:     query,
		Timestamp: time.Now(),
		Series:    make([]models.MetricSeries, 0),
	}
	
	seriesMap := make(map[string]*models.MetricSeries)
	
	// 处理查询结果
	for result.Next() {
		record := result.Record()
		
		// 构建序列键
		seriesKey := i.buildSeriesKey(record)
		
		// 获取或创建序列
		series, exists := seriesMap[seriesKey]
		if !exists {
			series = &models.MetricSeries{
				Labels: make(map[string]string),
				Points: make([]models.MetricPoint, 0),
			}
			
			// 添加标签
			for key, value := range record.Values() {
				if keyStr, ok := key.(string); ok {
					if !strings.HasPrefix(keyStr, "_") && keyStr != "result" && keyStr != "table" {
						if valueStr, ok := value.(string); ok {
							series.Labels[keyStr] = valueStr
						}
					}
				}
			}
			
			seriesMap[seriesKey] = series
		}
		
		// 添加数据点
		if record.Time() != nil && record.Value() != nil {
			point := models.MetricPoint{
				Timestamp: *record.Time(),
			}
			
			// 转换值
			switch v := record.Value().(type) {
			case float64:
				point.Value = v
			case int64:
				point.Value = float64(v)
			case string:
				if val, err := strconv.ParseFloat(v, 64); err == nil {
					point.Value = val
				}
			}
			
			series.Points = append(series.Points, point)
		}
	}
	
	// 检查查询错误
	if result.Err() != nil {
		return nil, fmt.Errorf("query result error: %w", result.Err())
	}
	
	// 转换为切片
	for _, series := range seriesMap {
		queryResult.Series = append(queryResult.Series, *series)
	}
	
	return queryResult, nil
}

// buildSeriesKey 构建序列键
func (i *InfluxDBStorage) buildSeriesKey(record *api.FluxRecord) string {
	var keyParts []string
	
	// 添加测量名称
	if measurement := record.Measurement(); measurement != "" {
		keyParts = append(keyParts, "measurement:"+measurement)
	}
	
	// 添加字段名称
	if field := record.Field(); field != "" {
		keyParts = append(keyParts, "field:"+field)
	}
	
	// 添加标签
	for key, value := range record.Values() {
		if keyStr, ok := key.(string); ok {
			if !strings.HasPrefix(keyStr, "_") && keyStr != "result" && keyStr != "table" {
				if valueStr, ok := value.(string); ok {
					keyParts = append(keyParts, fmt.Sprintf("%s:%s", keyStr, valueStr))
				}
			}
		}
	}
	
	return strings.Join(keyParts, "|")
}

// Health 健康检查
func (i *InfluxDBStorage) Health(ctx context.Context) error {
	// 检查连接
	health, err := i.client.Health(ctx)
	if err != nil {
		return fmt.Errorf("influxdb health check failed: %w", err)
	}
	
	if health.Status != "pass" {
		return fmt.Errorf("influxdb health status: %s", health.Status)
	}
	
	return nil
}

// GetBuckets 获取存储桶列表
func (i *InfluxDBStorage) GetBuckets(ctx context.Context) ([]string, error) {
	bucketsAPI := i.client.BucketsAPI()
	buckets, err := bucketsAPI.GetBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get buckets: %w", err)
	}
	
	var bucketNames []string
	for _, bucket := range *buckets {
		bucketNames = append(bucketNames, *bucket.Name)
	}
	
	return bucketNames, nil
}

// GetMeasurements 获取测量名称
func (i *InfluxDBStorage) GetMeasurements(ctx context.Context) ([]string, error) {
	fluxQuery := fmt.Sprintf(`
		import "influxdata/influxdb/schema"
		schema.measurements(bucket: "%s")
	`, i.config.Bucket)
	
	result, err := i.queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get measurements: %w", err)
	}
	defer result.Close()
	
	var measurements []string
	for result.Next() {
		record := result.Record()
		if measurement := record.ValueByKey("_value"); measurement != nil {
			if measurementStr, ok := measurement.(string); ok {
				measurements = append(measurements, measurementStr)
			}
		}
	}
	
	if result.Err() != nil {
		return nil, fmt.Errorf("measurements query error: %w", result.Err())
	}
	
	return measurements, nil
}

// GetTagKeys 获取标签键
func (i *InfluxDBStorage) GetTagKeys(ctx context.Context, measurement string) ([]string, error) {
	fluxQuery := fmt.Sprintf(`
		import "influxdata/influxdb/schema"
		schema.tagKeys(bucket: "%s", predicate: (r) => r._measurement == "%s")
	`, i.config.Bucket, measurement)
	
	result, err := i.queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag keys: %w", err)
	}
	defer result.Close()
	
	var tagKeys []string
	for result.Next() {
		record := result.Record()
		if tagKey := record.ValueByKey("_value"); tagKey != nil {
			if tagKeyStr, ok := tagKey.(string); ok {
				tagKeys = append(tagKeys, tagKeyStr)
			}
		}
	}
	
	if result.Err() != nil {
		return nil, fmt.Errorf("tag keys query error: %w", result.Err())
	}
	
	return tagKeys, nil
}

// GetTagValues 获取标签值
func (i *InfluxDBStorage) GetTagValues(ctx context.Context, measurement, tagKey string) ([]string, error) {
	fluxQuery := fmt.Sprintf(`
		import "influxdata/influxdb/schema"
		schema.tagValues(bucket: "%s", tag: "%s", predicate: (r) => r._measurement == "%s")
	`, i.config.Bucket, tagKey, measurement)
	
	result, err := i.queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag values: %w", err)
	}
	defer result.Close()
	
	var tagValues []string
	for result.Next() {
		record := result.Record()
		if tagValue := record.ValueByKey("_value"); tagValue != nil {
			if tagValueStr, ok := tagValue.(string); ok {
				tagValues = append(tagValues, tagValueStr)
			}
		}
	}
	
	if result.Err() != nil {
		return nil, fmt.Errorf("tag values query error: %w", result.Err())
	}
	
	return tagValues, nil
}

// GetFieldKeys 获取字段键
func (i *InfluxDBStorage) GetFieldKeys(ctx context.Context, measurement string) ([]string, error) {
	fluxQuery := fmt.Sprintf(`
		import "influxdata/influxdb/schema"
		schema.fieldKeys(bucket: "%s", predicate: (r) => r._measurement == "%s")
	`, i.config.Bucket, measurement)
	
	result, err := i.queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get field keys: %w", err)
	}
	defer result.Close()
	
	var fieldKeys []string
	for result.Next() {
		record := result.Record()
		if fieldKey := record.ValueByKey("_value"); fieldKey != nil {
			if fieldKeyStr, ok := fieldKey.(string); ok {
				fieldKeys = append(fieldKeys, fieldKeyStr)
			}
		}
	}
	
	if result.Err() != nil {
		return nil, fmt.Errorf("field keys query error: %w", result.Err())
	}
	
	return fieldKeys, nil
}

// DeleteData 删除数据
func (i *InfluxDBStorage) DeleteData(ctx context.Context, start, end time.Time, predicate string) error {
	deleteAPI := i.client.DeleteAPI()
	
	err := deleteAPI.DeleteWithName(ctx, i.config.Organization, i.config.Bucket, start, end, predicate)
	if err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}
	
	return nil
}

// GetStats 获取统计信息
func (i *InfluxDBStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 获取组织信息
	orgsAPI := i.client.OrganizationsAPI()
	org, err := orgsAPI.FindOrganizationByName(ctx, i.config.Organization)
	if err == nil && org != nil {
		stats["organization"] = map[string]interface{}{
			"id":   *org.Id,
			"name": *org.Name,
		}
	}
	
	// 获取存储桶信息
	bucketsAPI := i.client.BucketsAPI()
	bucket, err := bucketsAPI.FindBucketByName(ctx, i.config.Bucket)
	if err == nil && bucket != nil {
		stats["bucket"] = map[string]interface{}{
			"id":   *bucket.Id,
			"name": *bucket.Name,
		}
	}
	
	// 获取测量数量
	measurements, err := i.GetMeasurements(ctx)
	if err == nil {
		stats["measurements_count"] = len(measurements)
		stats["measurements"] = measurements
	}
	
	return stats, nil
}

// Close 关闭连接
func (i *InfluxDBStorage) Close() error {
	// 关闭写入API
	i.writeAPI.Close()
	
	// 关闭客户端
	i.client.Close()
	
	return nil
}

// Flush 强制刷新缓冲区
func (i *InfluxDBStorage) Flush() {
	i.writeAPI.Flush()
}

// SetWriteOptions 设置写入选项
func (i *InfluxDBStorage) SetWriteOptions(options *WriteOptions) {
	if options == nil {
		return
	}
	
	i.config.WriteOptions = options
	
	// 重新创建写入API
	writeOptions := influxdb2.DefaultOptions()
	
	if options.BatchSize > 0 {
		writeOptions = writeOptions.SetBatchSize(uint(options.BatchSize))
	}
	if options.FlushInterval > 0 {
		writeOptions = writeOptions.SetFlushInterval(uint(options.FlushInterval.Milliseconds()))
	}
	if options.RetryInterval > 0 {
		writeOptions = writeOptions.SetRetryInterval(uint(options.RetryInterval.Milliseconds()))
	}
	if options.MaxRetries > 0 {
		writeOptions = writeOptions.SetMaxRetries(uint(options.MaxRetries))
	}
	if options.UseGZip {
		writeOptions = writeOptions.SetUseGZip(true)
	}
	
	// 关闭旧的写入API
	i.writeAPI.Close()
	
	// 创建新的客户端和写入API
	i.client = influxdb2.NewClientWithOptions(i.config.URL, i.config.Token, writeOptions)
	i.writeAPI = i.client.WriteAPI(i.config.Organization, i.config.Bucket)
	
	// 重新启动错误处理
	go i.handleWriteErrors()
}