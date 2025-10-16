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

// PrometheusStorage Prometheus洢
type PrometheusStorage struct {
	client   api.Client
	queryAPI v1.API
	config   *PrometheusConfig
}

// PrometheusConfig Prometheus
type PrometheusConfig struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxSamples  int           `yaml:"max_samples"`
	QueryRange  time.Duration `yaml:"query_range"`
	Step        time.Duration `yaml:"step"`
	Compression bool          `yaml:"compression"`
	
	// 
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
	
	// TLS
	TLSConfig *TLSConfig `yaml:"tls_config"`
}

// TLSConfig TLS
type TLSConfig struct {
	CAFile             string `yaml:"ca_file"`
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

// QueryOptions 
type QueryOptions struct {
	Start    time.Time
	End      time.Time
	Step     time.Duration
	Timeout  time.Duration
	MaxSamples int
}

// NewPrometheusStorage Prometheus洢
func NewPrometheusStorage(config *PrometheusConfig) (*PrometheusStorage, error) {
	clientConfig := api.Config{
		Address: config.Address,
	}
	
	// 
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

// Store 洢Prometheus?
func (p *PrometheusStorage) Store(ctx context.Context, metrics []models.Metric) error {
	// Prometheus洢?
	// Pushgateway
	return fmt.Errorf("direct storage not supported, use push gateway or pull model")
}

// Query 
func (p *PrometheusStorage) Query(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	if query.Expression == "" {
		return nil, fmt.Errorf("query expression is required")
	}
	
	// 
	queryCtx := ctx
	if p.config.Timeout > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}
	
	// 
	if query.IsRange() {
		return p.queryRange(queryCtx, query)
	} else {
		return p.queryInstant(queryCtx, query)
	}
}

// queryInstant 
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

// queryRange 
func (p *PrometheusStorage) queryRange(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	step := query.Step
	if step == 0 {
		step = p.config.Step
		if step == 0 {
			step = time.Minute // 
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

// convertResult 
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
			
			// 
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
			
			// 
			for k, v := range sampleStream.Metric {
				series.Labels[string(k)] = string(v)
			}
			
			// ?
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
		// 
		result.StringResult = string(v.Value)
	}
	
	return result
}

// QueryLabels 
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

// QueryLabelValues ?
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

// QuerySeries 
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

// GetMetricNames 
func (p *PrometheusStorage) GetMetricNames(ctx context.Context) ([]string, error) {
	labels, err := p.QueryLabels(ctx, nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	
	// __name__?
	for _, label := range labels {
		if label == "__name__" {
			return p.QueryLabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
		}
	}
	
	return []string{}, nil
}

// Health ?
func (p *PrometheusStorage) Health(ctx context.Context) error {
	// ?
	_, err := p.queryAPI.Query(ctx, "up", time.Now())
	if err != nil {
		return fmt.Errorf("prometheus health check failed: %w", err)
	}
	return nil
}

// GetStats 
func (p *PrometheusStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Prometheus
	config, err := p.queryAPI.Config(ctx)
	if err == nil {
		stats["config"] = config
	}
	
	// ?
	runtimeInfo, err := p.queryAPI.Runtimeinfo(ctx)
	if err == nil {
		stats["runtime"] = runtimeInfo
	}
	
	// 
	buildInfo, err := p.queryAPI.Buildinfo(ctx)
	if err == nil {
		stats["build"] = buildInfo
	}
	
	// TSDB?
	tsdbStatus, err := p.queryAPI.TSDB(ctx)
	if err == nil {
		stats["tsdb"] = tsdbStatus
	}
	
	return stats, nil
}

// GetTargets 
func (p *PrometheusStorage) GetTargets(ctx context.Context) (interface{}, error) {
	targets, err := p.queryAPI.Targets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %w", err)
	}
	return targets, nil
}

// GetRules 
func (p *PrometheusStorage) GetRules(ctx context.Context) (interface{}, error) {
	rules, err := p.queryAPI.Rules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	return rules, nil
}

// GetAlerts 澯
func (p *PrometheusStorage) GetAlerts(ctx context.Context) (interface{}, error) {
	alerts, err := p.queryAPI.Alerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	return alerts, nil
}

// GetAlertManagers AlertManager
func (p *PrometheusStorage) GetAlertManagers(ctx context.Context) (interface{}, error) {
	alertManagers, err := p.queryAPI.AlertManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert managers: %w", err)
	}
	return alertManagers, nil
}

// QueryExemplars 
func (p *PrometheusStorage) QueryExemplars(ctx context.Context, query string, start, end time.Time) (interface{}, error) {
	exemplars, err := p.queryAPI.QueryExemplars(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query exemplars: %w", err)
	}
	return exemplars, nil
}

// Close 
func (p *PrometheusStorage) Close() error {
	// Prometheus?
	return nil
}

// BasicAuthRoundTripper 
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

// BearerTokenRoundTripper Bearer Token
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

// BuildQuery ?
func BuildQuery(metric string, labels map[string]string, aggregation string, duration time.Duration) string {
	var query strings.Builder
	
	// 
	if aggregation != "" {
		query.WriteString(aggregation)
		query.WriteString("(")
	}
	
	// 
	query.WriteString(metric)
	
	// ?
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
	
	// 
	if duration > 0 {
		query.WriteString("[")
		query.WriteString(duration.String())
		query.WriteString("]")
	}
	
	// 
	if aggregation != "" {
		query.WriteString(")")
	}
	
	return query.String()
}

// ParseLabels ?
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

// FormatLabels ?
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

// ValidateQuery ?
func ValidateQuery(query string) error {
	if query == "" {
		return fmt.Errorf("query expression cannot be empty")
	}
	
	// ?
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

