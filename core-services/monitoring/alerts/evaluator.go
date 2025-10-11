package alerts

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// AlertEvaluator 告警评估�?
type AlertEvaluator struct {
	storage interfaces.MetricStorage
	timeout time.Duration
}

// AlertEvaluationResult 告警评估结果
type AlertEvaluationResult struct {
	Firing bool                `json:"firing"`
	Value  float64             `json:"value"`
	Labels map[string]string   `json:"labels"`
}

// NewAlertEvaluator 创建告警评估�?
func NewAlertEvaluator(storage interfaces.MetricStorage, timeout time.Duration) *AlertEvaluator {
	return &AlertEvaluator{
		storage: storage,
		timeout: timeout,
	}
}

// Evaluate 评估告警规则
func (ae *AlertEvaluator) Evaluate(ctx context.Context, rule *models.AlertRule) ([]*AlertEvaluationResult, error) {
	// 创建带超时的上下�?
	evalCtx, cancel := context.WithTimeout(ctx, ae.timeout)
	defer cancel()
	
	// 执行查询
	result, err := ae.storage.QueryInstant(evalCtx, rule.Query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	
	// 处理查询结果
	var evaluationResults []*AlertEvaluationResult
	
	switch result.Type {
	case "vector":
		evaluationResults = ae.evaluateVector(rule, result)
	case "scalar":
		evaluationResults = ae.evaluateScalar(rule, result)
	case "matrix":
		evaluationResults = ae.evaluateMatrix(rule, result)
	default:
		return nil, fmt.Errorf("unsupported result type: %s", result.Type)
	}
	
	return evaluationResults, nil
}

// evaluateVector 评估向量结果
func (ae *AlertEvaluator) evaluateVector(rule *models.AlertRule, result *models.QueryResult) []*AlertEvaluationResult {
	var results []*AlertEvaluationResult
	
	for _, sample := range result.Data.Vector {
		value := sample.Value
		
		// 检查值是否有�?
		if math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		
		// 评估条件
		firing := ae.evaluateCondition(rule.Condition, value)
		
		result := &AlertEvaluationResult{
			Firing: firing,
			Value:  value,
			Labels: sample.Labels,
		}
		
		results = append(results, result)
	}
	
	return results
}

// evaluateScalar 评估标量结果
func (ae *AlertEvaluator) evaluateScalar(rule *models.AlertRule, result *models.QueryResult) []*AlertEvaluationResult {
	value := result.Data.Scalar.Value
	
	// 检查值是否有�?
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}
	
	// 评估条件
	firing := ae.evaluateCondition(rule.Condition, value)
	
	return []*AlertEvaluationResult{
		{
			Firing: firing,
			Value:  value,
			Labels: make(map[string]string),
		},
	}
}

// evaluateMatrix 评估矩阵结果
func (ae *AlertEvaluator) evaluateMatrix(rule *models.AlertRule, result *models.QueryResult) []*AlertEvaluationResult {
	var results []*AlertEvaluationResult
	
	for _, series := range result.Data.Matrix {
		if len(series.Values) == 0 {
			continue
		}
		
		// 使用最新的�?
		latestValue := series.Values[len(series.Values)-1]
		value := latestValue.Value
		
		// 检查值是否有�?
		if math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		
		// 评估条件
		firing := ae.evaluateCondition(rule.Condition, value)
		
		result := &AlertEvaluationResult{
			Firing: firing,
			Value:  value,
			Labels: series.Labels,
		}
		
		results = append(results, result)
	}
	
	return results
}

// evaluateCondition 评估条件
func (ae *AlertEvaluator) evaluateCondition(condition string, value float64) bool {
	// 解析条件表达�?
	// 支持的操作符: >, <, >=, <=, ==, !=
	
	condition = strings.TrimSpace(condition)
	
	// 查找操作�?
	var operator string
	var thresholdStr string
	
	if strings.Contains(condition, ">=") {
		parts := strings.Split(condition, ">=")
		if len(parts) == 2 {
			operator = ">="
			thresholdStr = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(condition, "<=") {
		parts := strings.Split(condition, "<=")
		if len(parts) == 2 {
			operator = "<="
			thresholdStr = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(condition, "==") {
		parts := strings.Split(condition, "==")
		if len(parts) == 2 {
			operator = "=="
			thresholdStr = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(condition, "!=") {
		parts := strings.Split(condition, "!=")
		if len(parts) == 2 {
			operator = "!="
			thresholdStr = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(condition, ">") {
		parts := strings.Split(condition, ">")
		if len(parts) == 2 {
			operator = ">"
			thresholdStr = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(condition, "<") {
		parts := strings.Split(condition, "<")
		if len(parts) == 2 {
			operator = "<"
			thresholdStr = strings.TrimSpace(parts[1])
		}
	} else {
		// 默认条件：值大�?
		return value > 0
	}
	
	// 解析阈�?
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		// 如果无法解析阈值，默认返回false
		return false
	}
	
	// 执行比较
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return math.Abs(value-threshold) < 1e-9 // 浮点数相等比�?
	case "!=":
		return math.Abs(value-threshold) >= 1e-9 // 浮点数不等比�?
	default:
		return false
	}
}

// EvaluateExpression 评估表达�?
func (ae *AlertEvaluator) EvaluateExpression(ctx context.Context, expression string, timestamp time.Time) (float64, error) {
	// 这是一个简化的表达式评估器
	// 在实际实现中，可能需要更复杂的表达式解析�?
	
	// 如果表达式是一个简单的数字
	if value, err := strconv.ParseFloat(expression, 64); err == nil {
		return value, nil
	}
	
	// 如果表达式是一个查�?
	result, err := ae.storage.QueryInstant(ctx, expression, timestamp)
	if err != nil {
		return 0, err
	}
	
	// 提取单个�?
	switch result.Type {
	case "scalar":
		return result.Data.Scalar.Value, nil
	case "vector":
		if len(result.Data.Vector) > 0 {
			return result.Data.Vector[0].Value, nil
		}
		return 0, fmt.Errorf("empty vector result")
	default:
		return 0, fmt.Errorf("unsupported result type for expression: %s", result.Type)
	}
}

// ValidateRule 验证告警规则
func (ae *AlertEvaluator) ValidateRule(rule *models.AlertRule) error {
	// 验证查询语法
	if rule.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	
	// 验证条件
	if rule.Condition == "" {
		return fmt.Errorf("condition cannot be empty")
	}
	
	// 验证持续时间
	if rule.For < 0 {
		return fmt.Errorf("for duration cannot be negative")
	}
	
	// 验证严重级别
	switch rule.Severity {
	case models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow, models.SeverityInfo:
		// 有效的严重级�?
	default:
		return fmt.Errorf("invalid severity level: %s", rule.Severity)
	}
	
	// 尝试解析条件表达�?
	if !ae.isValidCondition(rule.Condition) {
		return fmt.Errorf("invalid condition expression: %s", rule.Condition)
	}
	
	return nil
}

// isValidCondition 检查条件是否有�?
func (ae *AlertEvaluator) isValidCondition(condition string) bool {
	condition = strings.TrimSpace(condition)
	
	// 检查是否包含有效的操作�?
	operators := []string{">=", "<=", "==", "!=", ">", "<"}
	for _, op := range operators {
		if strings.Contains(condition, op) {
			parts := strings.Split(condition, op)
			if len(parts) == 2 {
				// 检查右侧是否是有效的数�?
				thresholdStr := strings.TrimSpace(parts[1])
				if _, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
					return true
				}
			}
		}
	}
	
	return false
}

// TestRule 测试告警规则
func (ae *AlertEvaluator) TestRule(ctx context.Context, rule *models.AlertRule) (*TestResult, error) {
	// 验证规则
	if err := ae.ValidateRule(rule); err != nil {
		return &TestResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}
	
	// 执行评估
	results, err := ae.Evaluate(ctx, rule)
	if err != nil {
		return &TestResult{
			Valid:   false,
			Error:   err.Error(),
			Results: nil,
		}, nil
	}
	
	return &TestResult{
		Valid:   true,
		Error:   "",
		Results: results,
	}, nil
}

// TestResult 测试结果
type TestResult struct {
	Valid   bool                     `json:"valid"`
	Error   string                   `json:"error,omitempty"`
	Results []*AlertEvaluationResult `json:"results,omitempty"`
}

// EvaluateMultipleRules 批量评估多个规则
func (ae *AlertEvaluator) EvaluateMultipleRules(ctx context.Context, rules []*models.AlertRule) (map[string][]*AlertEvaluationResult, error) {
	results := make(map[string][]*AlertEvaluationResult)
	
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		
		ruleResults, err := ae.Evaluate(ctx, rule)
		if err != nil {
			// 记录错误但继续处理其他规�?
			fmt.Printf("Failed to evaluate rule %s: %v\n", rule.Name, err)
			continue
		}
		
		results[rule.ID] = ruleResults
	}
	
	return results, nil
}

// GetQueryMetrics 获取查询指标
func (ae *AlertEvaluator) GetQueryMetrics(ctx context.Context, query string, start, end time.Time, step time.Duration) (*models.QueryResult, error) {
	return ae.storage.QueryRange(ctx, query, start, end, step)
}

// ExplainQuery 解释查询
func (ae *AlertEvaluator) ExplainQuery(query string) *QueryExplanation {
	explanation := &QueryExplanation{
		Query:       query,
		Type:        ae.detectQueryType(query),
		Complexity:  ae.estimateComplexity(query),
		Description: ae.generateDescription(query),
		Suggestions: ae.generateSuggestions(query),
	}
	
	return explanation
}

// QueryExplanation 查询解释
type QueryExplanation struct {
	Query       string   `json:"query"`
	Type        string   `json:"type"`
	Complexity  string   `json:"complexity"`
	Description string   `json:"description"`
	Suggestions []string `json:"suggestions"`
}

// detectQueryType 检测查询类�?
func (ae *AlertEvaluator) detectQueryType(query string) string {
	query = strings.ToLower(query)
	
	if strings.Contains(query, "rate(") || strings.Contains(query, "irate(") {
		return "rate"
	}
	if strings.Contains(query, "increase(") {
		return "counter"
	}
	if strings.Contains(query, "histogram_quantile(") {
		return "histogram"
	}
	if strings.Contains(query, "avg(") || strings.Contains(query, "sum(") || strings.Contains(query, "max(") || strings.Contains(query, "min(") {
		return "aggregation"
	}
	if strings.Contains(query, "by (") || strings.Contains(query, "without (") {
		return "grouping"
	}
	
	return "simple"
}

// estimateComplexity 估算复杂�?
func (ae *AlertEvaluator) estimateComplexity(query string) string {
	complexity := 0
	
	// 计算函数数量
	functions := []string{"rate(", "irate(", "increase(", "sum(", "avg(", "max(", "min(", "count(", "histogram_quantile("}
	for _, fn := range functions {
		complexity += strings.Count(strings.ToLower(query), fn)
	}
	
	// 计算操作符数�?
	operators := []string{"+", "-", "*", "/", "==", "!=", ">", "<", ">=", "<="}
	for _, op := range operators {
		complexity += strings.Count(query, op)
	}
	
	// 计算分组数量
	complexity += strings.Count(strings.ToLower(query), "by (")
	complexity += strings.Count(strings.ToLower(query), "without (")
	
	if complexity <= 2 {
		return "low"
	} else if complexity <= 5 {
		return "medium"
	} else {
		return "high"
	}
}

// generateDescription 生成描述
func (ae *AlertEvaluator) generateDescription(query string) string {
	queryType := ae.detectQueryType(query)
	
	switch queryType {
	case "rate":
		return "计算指标的变化率，通常用于监控计数器类型的指标"
	case "counter":
		return "计算计数器的增长量，用于监控累积值的变化"
	case "histogram":
		return "计算直方图的分位数，用于监控延迟和响应时间分�?
	case "aggregation":
		return "对指标进行聚合计算，如求和、平均值、最大值等"
	case "grouping":
		return "按标签对指标进行分组聚合"
	default:
		return "简单的指标查询"
	}
}

// generateSuggestions 生成建议
func (ae *AlertEvaluator) generateSuggestions(query string) []string {
	var suggestions []string
	
	query = strings.ToLower(query)
	
	// 检查是否使用了rate函数但没有指定时间窗�?
	if strings.Contains(query, "rate(") && !strings.Contains(query, "[") {
		suggestions = append(suggestions, "建议为rate函数指定时间窗口，如rate(metric[5m])")
	}
	
	// 检查是否使用了聚合函数但没有分�?
	if (strings.Contains(query, "sum(") || strings.Contains(query, "avg(")) && !strings.Contains(query, "by (") {
		suggestions = append(suggestions, "考虑使用by子句对聚合结果进行分�?)
	}
	
	// 检查是否查询了高基数指�?
	if strings.Contains(query, "container_") || strings.Contains(query, "node_") {
		suggestions = append(suggestions, "注意高基数指标可能影响查询性能，考虑添加标签过滤�?)
	}
	
	// 检查是否使用了复杂的正则表达式
	if strings.Contains(query, "=~") {
		suggestions = append(suggestions, "正则表达式匹配可能影响性能，考虑使用精确匹配")
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "查询看起来不错，没有明显的优化建�?)
	}
	
	return suggestions
}
