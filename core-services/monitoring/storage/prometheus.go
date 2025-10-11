package storage

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// PrometheusStorage Prometheus存储实现
type PrometheusStorage struct {
	client   api.Client
	queryAPI v1.API
	config   *PrometheusConfig
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxSamples  int           `yaml:"max_samples"`
	QueryRange  time.Duration `yaml:"query_range"`
	Step        time.Duration `yaml:"step"`
	Compression bool          `yaml:"compression"`
	
	// 认证配置
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
	
	// TLS配置
	TLSConfig *TLSConfig `yaml:"tls_config"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	CAFile             string `yaml:"ca_file"`
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

// QueryOptions 查询选项
type QueryOptions struct {
	Start    time.Time
	End      time.Time
	Step     time.Duration
	Timeout  time.Duration
	MaxSamples int
}

// NewPrometheusStorage 创建Prometheus存储
func NewPrometheusStorage(config *PrometheusConfig) (*PrometheusStorage, error) {
	clientConfig := api.Config{
		Address: config.Address,
	}
	
	// 设置认证
	if config.Username != "" && config.Password != "" {
		clientConfig.RoundTripper = &BasicAuthRoundTripper{
			Username: config.Username,
			Password: config.Password,
		}
	} else if config.Token != "" {
		clientConfig.RoundTripper = &BearerTokenRoundTripper{
			Token: config.Token,
		}
	}
	
	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}
	
	return &PrometheusStorage{
		client:   client,
		queryAPI: v1.NewAPI(client),
		config:   config,
	}, nil
}

// Store 存储指标（Prometheus通常通过推送网关或拉取方式收集指标�?
func (p *PrometheusStorage) Store(ctx context.Context, metrics []models.Metric) error {
	// Prometheus通常不直接存储指标，而是通过拉取或推送网�?
	// 这里可以实现推送到Pushgateway的逻辑
	return fmt.Errorf("direct storage not supported, use push gateway or pull model")
}

// Query 查询指标
func (p *PrometheusStorage) Query(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	if query.Expression == "" {
		return nil, fmt.Errorf("query expression is required")
	}
	
	// 设置查询超时
	queryCtx := ctx
	if p.config.Timeout > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}
	
	// 执行查询
	if query.IsRange() {
		return p.queryRange(queryCtx, query)
	} else {
		return p.queryInstant(queryCtx, query)
	}
}

// queryInstant 即时查询
func (p *PrometheusStorage) queryInstant(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	timestamp := query.End
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	
	result, warnings, err := p.queryAPI.Query(ctx, query.Expression, timestamp)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	
	if len(warnings) > 0 {
		fmt.Printf("Prometheus query warnings: %v\n", warnings)
	}
	
	return p.convertResult(result, query), nil
}

// queryRange 范围查询
func (p *PrometheusStorage) queryRange(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	step := query.Step
	if step == 0 {
		step = p.config.Step
		if step == 0 {
			step = time.Minute // 默认步长
		}
	}
	
	r := v1.Range{
		Start: query.Start,
		End:   query.End,
		Step:  step,
	}
	
	result, warnings, err := p.queryAPI.QueryRange(ctx, query.Expression, r)
	if err != nil {
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}
	
	if len(warnings) > 0 {
		fmt.Printf("Prometheus range query warnings: %v\n", warnings)
	}
	
	return p.convertResult(result, query), nil
}

// convertResult 转换查询结果
func (p *PrometheusStorage) convertResult(value model.Value, query *models.MetricQuery) *models.MetricQueryResult {
	result := &models.MetricQueryResult{
		Query:     query,
		Timestamp: time.Now(),
		Series:    make([]models.MetricSeries, 0),
	}
	
	switch v := value.(type) {
	case model.Vector:
		for _, sample := range v {
			series := models.MetricSeries{
				Labels: make(map[string]string),
				Points: []models.MetricPoint{
					{
						Timestamp: sample.Timestamp.Time(),
						Value:     float64(sample.Value),
					},
				},
			}
			
			// 转换标签
			for k, v := range sample.Metric {
				series.Labels[string(k)] = string(v)
			}
			
			result.Series = append(result.Series, series)
		}
		
	case model.Matrix:
		for _, sampleStream := range v {
			series := models.MetricSeries{
				Labels: make(map[string]string),
				Points: make([]models.MetricPoint, 0, len(sampleStream.Values)),
			}
			
			// 转换标签
			for k, v := range sampleStream.Metric {
				series.Labels[string(k)] = string(v)
			}
			
			// 转换数据�?
			for _, pair := range sampleStream.Values {
				series.Points = append(series.Points, models.MetricPoint{
					Timestamp: pair.Timestamp.Time(),
					Value:     float64(pair.Value),
				})
			}
			
			result.Series = append(result.Series, series)
		}
		
	case *model.Scalar:
		series := models.MetricSeries{
			Labels: make(map[string]string),
			Points: []models.MetricPoint{
				{
					Timestamp: v.Timestamp.Time(),
					Value:     float64(v.Value),
				},
			},
		}
		result.Series = append(result.Series, series)
		
	case model.String:
		// 字符串结果，通常用于标签查询
		result.StringResult = string(v.Value)
	}
	
	return result
}

// QueryLabels 查询标签
func (p *PrometheusStorage) QueryLabels(ctx context.Context, matchers []string, start, end time.Time) ([]string, error) {
	var labelMatchers []string
	if len(matchers) > 0 {
		labelMatchers = matchers
	}
	
	labels, warnings, err := p.queryAPI.LabelNames(ctx, labelMatchers, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels: %w", err)
	}
	
	if len(warnings) > 0 {
		fmt.Printf("Label query warnings: %v\n", warnings)
	}
	
	return labels, nil
}

// QueryLabelValues 查询标签�?
func (p *PrometheusStorage) QueryLabelValues(ctx context.Context, label string, matchers []string, start, end time.Time) ([]string, error) {
	var labelMatchers []string
	if len(matchers) > 0 {
		labelMatchers = matchers
	}
	
	values, warnings, err := p.queryAPI.LabelValues(ctx, label, labelMatchers, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query label values: %w", err)
	}
	
	if len(warnings) > 0 {
		fmt.Printf("Label values query warnings: %v\n", warnings)
	}
	
	return values, nil
}

// QuerySeries 查询序列
func (p *PrometheusStorage) QuerySeries(ctx context.Context, matchers []string, start, end time.Time) ([]map[string]string, error) {
	series, warnings, err := p.queryAPI.Series(ctx, matchers, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query series: %w", err)
	}
	
	if len(warnings) > 0 {
		fmt.Printf("Series query warnings: %v\n", warnings)
	}
	
	result := make([]map[string]string, 0, len(series))
	for _, labelSet := range series {
		labels := make(map[string]string)
		for k, v := range labelSet {
			labels[string(k)] = string(v)
		}
		result = append(result, labels)
	}
	
	return result, nil
}

// GetMetricNames 获取指标名称
func (p *PrometheusStorage) GetMetricNames(ctx context.Context) ([]string, error) {
	labels, err := p.QueryLabels(ctx, nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	
	// 查找__name__标签的�?
	for _, label := range labels {
		if label == "__name__" {
			return p.QueryLabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
		}
	}
	
	return []string{}, nil
}

// Health 健康检�?
func (p *PrometheusStorage) Health(ctx context.Context) error {
	// 执行简单查询检查连�?
	_, err := p.queryAPI.Query(ctx, "up", time.Now())
	if err != nil {
		return fmt.Errorf("prometheus health check failed: %w", err)
	}
	return nil
}

// GetStats 获取统计信息
func (p *PrometheusStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 获取Prometheus配置信息
	config, err := p.queryAPI.Config(ctx)
	if err == nil {
		stats["config"] = config
	}
	
	// 获取运行时信�?
	runtimeInfo, err := p.queryAPI.Runtimeinfo(ctx)
	if err == nil {
		stats["runtime"] = runtimeInfo
	}
	
	// 获取构建信息
	buildInfo, err := p.queryAPI.Buildinfo(ctx)
	if err == nil {
		stats["build"] = buildInfo
	}
	
	// 获取TSDB状�?
	tsdbStatus, err := p.queryAPI.TSDB(ctx)
	if err == nil {
		stats["tsdb"] = tsdbStatus
	}
	
	return stats, nil
}

// GetTargets 获取目标信息
func (p *PrometheusStorage) GetTargets(ctx context.Context) (interface{}, error) {
	targets, err := p.queryAPI.Targets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %w", err)
	}
	return targets, nil
}

// GetRules 获取规则信息
func (p *PrometheusStorage) GetRules(ctx context.Context) (interface{}, error) {
	rules, err := p.queryAPI.Rules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	return rules, nil
}

// GetAlerts 获取告警信息
func (p *PrometheusStorage) GetAlerts(ctx context.Context) (interface{}, error) {
	alerts, err := p.queryAPI.Alerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	return alerts, nil
}

// GetAlertManagers 获取AlertManager信息
func (p *PrometheusStorage) GetAlertManagers(ctx context.Context) (interface{}, error) {
	alertManagers, err := p.queryAPI.AlertManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert managers: %w", err)
	}
	return alertManagers, nil
}

// QueryExemplars 查询示例
func (p *PrometheusStorage) QueryExemplars(ctx context.Context, query string, start, end time.Time) (interface{}, error) {
	exemplars, err := p.queryAPI.QueryExemplars(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query exemplars: %w", err)
	}
	return exemplars, nil
}

// Close 关闭连接
func (p *PrometheusStorage) Close() error {
	// Prometheus客户端不需要显式关�?
	return nil
}

// BasicAuthRoundTripper 基本认证
type BasicAuthRoundTripper struct {
	Username string
	Password string
	Next     http.RoundTripper
}

func (rt *BasicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.Username, rt.Password)
	next := rt.Next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}

// BearerTokenRoundTripper Bearer Token认证
type BearerTokenRoundTripper struct {
	Token string
	Next  http.RoundTripper
}

func (rt *BearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+rt.Token)
	next := rt.Next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}

// BuildQuery 构建查询表达�?
func BuildQuery(metric string, labels map[string]string, aggregation string, duration time.Duration) string {
	var query strings.Builder
	
	// 添加聚合函数
	if aggregation != "" {
		query.WriteString(aggregation)
		query.WriteString("(")
	}
	
	// 添加指标名称
	query.WriteString(metric)
	
	// 添加标签选择�?
	if len(labels) > 0 {
		query.WriteString("{")
		first := true
		for k, v := range labels {
			if !first {
				query.WriteString(",")
			}
			query.WriteString(k)
			query.WriteString("=\"")
			query.WriteString(v)
			query.WriteString("\"")
			first = false
		}
		query.WriteString("}")
	}
	
	// 添加时间范围
	if duration > 0 {
		query.WriteString("[")
		query.WriteString(duration.String())
		query.WriteString("]")
	}
	
	// 关闭聚合函数
	if aggregation != "" {
		query.WriteString(")")
	}
	
	return query.String()
}

// ParseLabels 解析标签字符�?
func ParseLabels(labelStr string) map[string]string {
	labels := make(map[string]string)
	if labelStr == "" {
		return labels
	}
	
	pairs := strings.Split(labelStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.Trim(strings.TrimSpace(kv[1]), "\"'")
			labels[key] = value
		}
	}
	
	return labels
}

// FormatLabels 格式化标签为字符�?
func FormatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	
	return "{" + strings.Join(parts, ",") + "}"
}

// ValidateQuery 验证查询表达�?
func ValidateQuery(query string) error {
	if query == "" {
		return fmt.Errorf("query expression cannot be empty")
	}
	
	// 基本的语法检�?
	if strings.Count(query, "(") != strings.Count(query, ")") {
		return fmt.Errorf("unmatched parentheses in query")
	}
	
	if strings.Count(query, "{") != strings.Count(query, "}") {
		return fmt.Errorf("unmatched braces in query")
	}
	
	if strings.Count(query, "[") != strings.Count(query, "]") {
		return fmt.Errorf("unmatched brackets in query")
	}
	
	return nil
}
