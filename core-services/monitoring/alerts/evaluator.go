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

// AlertEvaluator 澯?
type AlertEvaluator struct {
	storage interfaces.MetricStorage
	timeout time.Duration
}

// AlertEvaluationResult 澯
type AlertEvaluationResult struct {
	Firing bool                `json:"firing"`
	Value  float64             `json:"value"`
	Labels map[string]string   `json:"labels"`
}

// NewAlertEvaluator 澯?
func NewAlertEvaluator(storage interfaces.MetricStorage, timeout time.Duration) *AlertEvaluator {
	return &AlertEvaluator{
		storage: storage,
		timeout: timeout,
	}
}

// Evaluate 澯
func (ae *AlertEvaluator) Evaluate(ctx context.Context, rule *models.AlertRule) ([]*AlertEvaluationResult, error) {
	// ?
	evalCtx, cancel := context.WithTimeout(ctx, ae.timeout)
	defer cancel()
	
	// 
	result, err := ae.storage.QueryInstant(evalCtx, rule.Query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	
	// 
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

// evaluateVector 
func (ae *AlertEvaluator) evaluateVector(rule *models.AlertRule, result *models.QueryResult) []*AlertEvaluationResult {
	var results []*AlertEvaluationResult
	
	for _, sample := range result.Data.Vector {
		value := sample.Value
		
		// ?
		if math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		
		// 
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

// evaluateScalar 
func (ae *AlertEvaluator) evaluateScalar(rule *models.AlertRule, result *models.QueryResult) []*AlertEvaluationResult {
	value := result.Data.Scalar.Value
	
	// ?
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}
	
	// 
	firing := ae.evaluateCondition(rule.Condition, value)
	
	return []*AlertEvaluationResult{
		{
			Firing: firing,
			Value:  value,
			Labels: make(map[string]string),
		},
	}
}

// evaluateMatrix 
func (ae *AlertEvaluator) evaluateMatrix(rule *models.AlertRule, result *models.QueryResult) []*AlertEvaluationResult {
	var results []*AlertEvaluationResult
	
	for _, series := range result.Data.Matrix {
		if len(series.Values) == 0 {
			continue
		}
		
		// ?
		latestValue := series.Values[len(series.Values)-1]
		value := latestValue.Value
		
		// ?
		if math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		
		// 
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

// evaluateCondition 
func (ae *AlertEvaluator) evaluateCondition(condition string, value float64) bool {
	// ?
	// : >, <, >=, <=, ==, !=
	
	condition = strings.TrimSpace(condition)
	
	// ?
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
		// ?
		return value > 0
	}
	
	// ?
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		// false
		return false
	}
	
	// 
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
		return math.Abs(value-threshold) < 1e-9 // ?
	case "!=":
		return math.Abs(value-threshold) >= 1e-9 // ?
	default:
		return false
	}
}

// EvaluateExpression ?
func (ae *AlertEvaluator) EvaluateExpression(ctx context.Context, expression string, timestamp time.Time) (float64, error) {
	// 
	// ?
	
	// 
	if value, err := strconv.ParseFloat(expression, 64); err == nil {
		return value, nil
	}
	
	// ?
	result, err := ae.storage.QueryInstant(ctx, expression, timestamp)
	if err != nil {
		return 0, err
	}
	
	// ?
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

// ValidateRule 澯
func (ae *AlertEvaluator) ValidateRule(rule *models.AlertRule) error {
	// 
	if rule.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	
	// 
	if rule.Condition == "" {
		return fmt.Errorf("condition cannot be empty")
	}
	
	// 
	if rule.For < 0 {
		return fmt.Errorf("for duration cannot be negative")
	}
	
	// 
	switch rule.Severity {
	case models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow, models.SeverityInfo:
		// ?
	default:
		return fmt.Errorf("invalid severity level: %s", rule.Severity)
	}
	
	// ?
	if !ae.isValidCondition(rule.Condition) {
		return fmt.Errorf("invalid condition expression: %s", rule.Condition)
	}
	
	return nil
}

// isValidCondition ?
func (ae *AlertEvaluator) isValidCondition(condition string) bool {
	condition = strings.TrimSpace(condition)
	
	// ?
	operators := []string{">=", "<=", "==", "!=", ">", "<"}
	for _, op := range operators {
		if strings.Contains(condition, op) {
			parts := strings.Split(condition, op)
			if len(parts) == 2 {
				// ?
				thresholdStr := strings.TrimSpace(parts[1])
				if _, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
					return true
				}
			}
		}
	}
	
	return false
}

// TestRule 澯
func (ae *AlertEvaluator) TestRule(ctx context.Context, rule *models.AlertRule) (*TestResult, error) {
	// 
	if err := ae.ValidateRule(rule); err != nil {
		return &TestResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}
	
	// 
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

// TestResult 
type TestResult struct {
	Valid   bool                     `json:"valid"`
	Error   string                   `json:"error,omitempty"`
	Results []*AlertEvaluationResult `json:"results,omitempty"`
}

// EvaluateMultipleRules 
func (ae *AlertEvaluator) EvaluateMultipleRules(ctx context.Context, rules []*models.AlertRule) (map[string][]*AlertEvaluationResult, error) {
	results := make(map[string][]*AlertEvaluationResult)
	
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		
		ruleResults, err := ae.Evaluate(ctx, rule)
		if err != nil {
			// ?
			fmt.Printf("Failed to evaluate rule %s: %v\n", rule.Name, err)
			continue
		}
		
		results[rule.ID] = ruleResults
	}
	
	return results, nil
}

// GetQueryMetrics 
func (ae *AlertEvaluator) GetQueryMetrics(ctx context.Context, query string, start, end time.Time, step time.Duration) (*models.QueryResult, error) {
	return ae.storage.QueryRange(ctx, query, start, end, step)
}

// ExplainQuery 
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

// QueryExplanation 
type QueryExplanation struct {
	Query       string   `json:"query"`
	Type        string   `json:"type"`
	Complexity  string   `json:"complexity"`
	Description string   `json:"description"`
	Suggestions []string `json:"suggestions"`
}

// detectQueryType ?
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

// estimateComplexity 㸴?
func (ae *AlertEvaluator) estimateComplexity(query string) string {
	complexity := 0
	
	// 㺯
	functions := []string{"rate(", "irate(", "increase(", "sum(", "avg(", "max(", "min(", "count(", "histogram_quantile("}
	for _, fn := range functions {
		complexity += strings.Count(strings.ToLower(query), fn)
	}
	
	// ?
	operators := []string{"+", "-", "*", "/", "==", "!=", ">", "<", ">=", "<="}
	for _, op := range operators {
		complexity += strings.Count(query, op)
	}
	
	// 
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

// generateDescription 
func (ae *AlertEvaluator) generateDescription(query string) string {
	queryType := ae.detectQueryType(query)
	
	switch queryType {
	case "rate":
		return "仯"
	case "counter":
		return "仯"
	case "histogram":
		return "?
	case "aggregation":
		return ""
	case "grouping":
		return ""
	default:
		return ""
	}
}

// generateSuggestions 
func (ae *AlertEvaluator) generateSuggestions(query string) []string {
	var suggestions []string
	
	query = strings.ToLower(query)
	
	// rate䴰?
	if strings.Contains(query, "rate(") && !strings.Contains(query, "[") {
		suggestions = append(suggestions, "rate䴰rate(metric[5m])")
	}
	
	// ?
	if (strings.Contains(query, "sum(") || strings.Contains(query, "avg(")) && !strings.Contains(query, "by (") {
		suggestions = append(suggestions, "by?)
	}
	
	// ?
	if strings.Contains(query, "container_") || strings.Contains(query, "node_") {
		suggestions = append(suggestions, "?)
	}
	
	// 
	if strings.Contains(query, "=~") {
		suggestions = append(suggestions, "")
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "?)
	}
	
	return suggestions
}

