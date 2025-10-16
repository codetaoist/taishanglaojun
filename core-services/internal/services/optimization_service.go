package services

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// OptimizationService 优化服务
type OptimizationService struct {
	db           *sql.DB
	logger       *zap.Logger
	queryCache   *QueryCache
	indexAnalyzer *IndexAnalyzer
	slowQueries  *SlowQueryAnalyzer
}

// QueryCache 查询缓存
type QueryCache struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IndexAnalyzer 索引分析器
type IndexAnalyzer struct {
	db     *sql.DB
	logger *zap.Logger
}

// SlowQueryAnalyzer 慢查询分析器
type SlowQueryAnalyzer struct {
	queries   []SlowQuery
	mu        sync.RWMutex
	threshold time.Duration
}

// SlowQuery 慢查询
type SlowQuery struct {
	Query     string
	Duration  time.Duration
	Timestamp time.Time
	Count     int
}

// IndexSuggestion 索引建议
type IndexSuggestion struct {
	Table   string
	Columns []string
	Type    string // btree, hash, gin, gist
	Reason  string
}

// QueryOptimizationResult 查询优化结果
type QueryOptimizationResult struct {
	OriginalQuery    string
	OptimizedQuery   string
	IndexSuggestions []IndexSuggestion
	EstimatedSpeedup float64
	Explanation      string
}

// NewOptimizationService 创建优化服务
func NewOptimizationService(db *sql.DB, logger *zap.Logger) *OptimizationService {
	return &OptimizationService{
		db:     db,
		logger: logger,
		queryCache: &QueryCache{
			cache: make(map[string]*CacheEntry),
			ttl:   time.Minute * 15,
		},
		indexAnalyzer: &IndexAnalyzer{
			db:     db,
			logger: logger,
		},
		slowQueries: &SlowQueryAnalyzer{
			queries:   make([]SlowQuery, 0),
			threshold: time.Millisecond * 100,
		},
	}
}

// OptimizeQuery 优化查询
func (os *OptimizationService) OptimizeQuery(ctx context.Context, query string) (*QueryOptimizationResult, error) {
	result := &QueryOptimizationResult{
		OriginalQuery: query,
	}

	// 1. 分析查询结构
	queryInfo := os.analyzeQuery(query)
	
	// 2. 生成优化建议
	optimizedQuery := os.optimizeQueryStructure(query, queryInfo)
	result.OptimizedQuery = optimizedQuery

	// 3. 生成索引建议
	indexSuggestions, err := os.generateIndexSuggestions(ctx, queryInfo)
	if err != nil {
		os.logger.Error("Failed to generate index suggestions", zap.Error(err))
	} else {
		result.IndexSuggestions = indexSuggestions
	}

	// 4. 估算性能提升
	result.EstimatedSpeedup = os.estimateSpeedup(queryInfo, indexSuggestions)
	
	// 5. 生成解释
	result.Explanation = os.generateExplanation(queryInfo, indexSuggestions)

	return result, nil
}

// QueryInfo 查询信息
type QueryInfo struct {
	Type        string   // SELECT, INSERT, UPDATE, DELETE
	Tables      []string
	Columns     []string
	WhereClause string
	JoinClauses []string
	OrderBy     []string
	GroupBy     []string
	HasLimit    bool
	HasOffset   bool
}

// analyzeQuery 分析查询
func (os *OptimizationService) analyzeQuery(query string) *QueryInfo {
	info := &QueryInfo{}
	
	// 转换为大写进行分析
	upperQuery := strings.ToUpper(query)
	
	// 确定查询类型
	if strings.HasPrefix(upperQuery, "SELECT") {
		info.Type = "SELECT"
	} else if strings.HasPrefix(upperQuery, "INSERT") {
		info.Type = "INSERT"
	} else if strings.HasPrefix(upperQuery, "UPDATE") {
		info.Type = "UPDATE"
	} else if strings.HasPrefix(upperQuery, "DELETE") {
		info.Type = "DELETE"
	}

	// 提取表名
	info.Tables = os.extractTables(query)
	
	// 提取列名
	info.Columns = os.extractColumns(query)
	
	// 提取WHERE子句
	info.WhereClause = os.extractWhereClause(query)
	
	// 提取JOIN子句
	info.JoinClauses = os.extractJoinClauses(query)
	
	// 提取ORDER BY
	info.OrderBy = os.extractOrderBy(query)
	
	// 提取GROUP BY
	info.GroupBy = os.extractGroupBy(query)
	
	// 检查LIMIT和OFFSET
	info.HasLimit = strings.Contains(upperQuery, "LIMIT")
	info.HasOffset = strings.Contains(upperQuery, "OFFSET")

	return info
}

// extractTables 提取表名
func (os *OptimizationService) extractTables(query string) []string {
	// 简化的表名提取逻辑
	re := regexp.MustCompile(`(?i)FROM\s+(\w+)|JOIN\s+(\w+)`)
	matches := re.FindAllStringSubmatch(query, -1)
	
	var tables []string
	for _, match := range matches {
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				tables = append(tables, match[i])
			}
		}
	}
	
	return tables
}

// extractColumns 提取列名
func (os *OptimizationService) extractColumns(query string) []string {
	// 简化的列名提取逻辑
	re := regexp.MustCompile(`(?i)SELECT\s+(.*?)\s+FROM`)
	matches := re.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		return []string{}
	}
	
	columnsStr := matches[1]
	if columnsStr == "*" {
		return []string{"*"}
	}
	
	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}
	
	return columns
}

// extractWhereClause 提取WHERE子句
func (os *OptimizationService) extractWhereClause(query string) string {
	re := regexp.MustCompile(`(?i)WHERE\s+(.*?)(?:\s+ORDER\s+BY|\s+GROUP\s+BY|\s+LIMIT|$)`)
	matches := re.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		return ""
	}
	
	return strings.TrimSpace(matches[1])
}

// extractJoinClauses 提取JOIN子句
func (os *OptimizationService) extractJoinClauses(query string) []string {
	re := regexp.MustCompile(`(?i)((?:INNER|LEFT|RIGHT|FULL)\s+)?JOIN\s+\w+\s+ON\s+[^;]+`)
	matches := re.FindAllString(query, -1)
	
	return matches
}

// extractOrderBy 提取ORDER BY
func (os *OptimizationService) extractOrderBy(query string) []string {
	re := regexp.MustCompile(`(?i)ORDER\s+BY\s+(.*?)(?:\s+LIMIT|$)`)
	matches := re.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		return []string{}
	}
	
	orderBy := strings.Split(matches[1], ",")
	for i, col := range orderBy {
		orderBy[i] = strings.TrimSpace(col)
	}
	
	return orderBy
}

// extractGroupBy 提取GROUP BY
func (os *OptimizationService) extractGroupBy(query string) []string {
	re := regexp.MustCompile(`(?i)GROUP\s+BY\s+(.*?)(?:\s+ORDER\s+BY|\s+LIMIT|$)`)
	matches := re.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		return []string{}
	}
	
	groupBy := strings.Split(matches[1], ",")
	for i, col := range groupBy {
		groupBy[i] = strings.TrimSpace(col)
	}
	
	return groupBy
}

// optimizeQueryStructure 优化查询结构
func (os *OptimizationService) optimizeQueryStructure(query string, info *QueryInfo) string {
	optimized := query
	
	// 1. 添加LIMIT（如果没有）
	if info.Type == "SELECT" && !info.HasLimit {
		optimized += " LIMIT 1000"
	}
	
	// 2. 优化SELECT子句
	if info.Type == "SELECT" && len(info.Columns) == 1 && info.Columns[0] == "*" {
		// 建议指定具体列名而不是使用*
		os.logger.Warn("Query uses SELECT *, consider specifying exact columns")
	}
	
	// 3. 优化WHERE子句
	if info.WhereClause != "" {
		optimized = os.optimizeWhereClause(optimized, info.WhereClause)
	}
	
	return optimized
}

// optimizeWhereClause 优化WHERE子句
func (os *OptimizationService) optimizeWhereClause(query, whereClause string) string {
	// 检查是否使用了函数在WHERE子句中
	if strings.Contains(strings.ToUpper(whereClause), "UPPER(") ||
		strings.Contains(strings.ToUpper(whereClause), "LOWER(") {
		os.logger.Warn("WHERE clause uses functions, consider using functional indexes")
	}
	
	// 检查是否使用了LIKE '%pattern%'
	if strings.Contains(whereClause, "LIKE '%") {
		os.logger.Warn("WHERE clause uses leading wildcard LIKE, consider full-text search")
	}
	
	return query
}

// generateIndexSuggestions 生成索引建议
func (os *OptimizationService) generateIndexSuggestions(ctx context.Context, info *QueryInfo) ([]IndexSuggestion, error) {
	var suggestions []IndexSuggestion
	
	// 1. 为WHERE子句中的列建议索引
	if info.WhereClause != "" {
		whereColumns := os.extractColumnsFromWhere(info.WhereClause)
		for _, table := range info.Tables {
			for _, column := range whereColumns {
				suggestions = append(suggestions, IndexSuggestion{
					Table:   table,
					Columns: []string{column},
					Type:    "btree",
					Reason:  "Used in WHERE clause",
				})
			}
		}
	}
	
	// 2. 为ORDER BY列建议索引
	if len(info.OrderBy) > 0 {
		for _, table := range info.Tables {
			suggestions = append(suggestions, IndexSuggestion{
				Table:   table,
				Columns: info.OrderBy,
				Type:    "btree",
				Reason:  "Used in ORDER BY clause",
			})
		}
	}
	
	// 3. 为JOIN列建议索引
	for _, joinClause := range info.JoinClauses {
		joinColumns := os.extractJoinColumns(joinClause)
		for table, columns := range joinColumns {
			suggestions = append(suggestions, IndexSuggestion{
				Table:   table,
				Columns: columns,
				Type:    "btree",
				Reason:  "Used in JOIN condition",
			})
		}
	}
	
	// 4. 检查现有索引
	existingIndexes, err := os.getExistingIndexes(ctx, info.Tables)
	if err != nil {
		return suggestions, err
	}
	
	// 过滤已存在的索引
	suggestions = os.filterExistingIndexes(suggestions, existingIndexes)
	
	return suggestions, nil
}

// extractColumnsFromWhere 从WHERE子句提取列名
func (os *OptimizationService) extractColumnsFromWhere(whereClause string) []string {
	// 简化的列名提取逻辑
	re := regexp.MustCompile(`(\w+)\s*[=<>!]`)
	matches := re.FindAllStringSubmatch(whereClause, -1)
	
	var columns []string
	for _, match := range matches {
		if len(match) > 1 {
			columns = append(columns, match[1])
		}
	}
	
	return columns
}

// extractJoinColumns 从JOIN子句提取列名
func (os *OptimizationService) extractJoinColumns(joinClause string) map[string][]string {
	// 简化的JOIN列提取逻辑
	result := make(map[string][]string)
	
	re := regexp.MustCompile(`(\w+)\.(\w+)\s*=\s*(\w+)\.(\w+)`)
	matches := re.FindAllStringSubmatch(joinClause, -1)
	
	for _, match := range matches {
		if len(match) > 4 {
			table1, col1 := match[1], match[2]
			table2, col2 := match[3], match[4]
			
			result[table1] = append(result[table1], col1)
			result[table2] = append(result[table2], col2)
		}
	}
	
	return result
}

// getExistingIndexes 获取现有索引
func (os *OptimizationService) getExistingIndexes(ctx context.Context, tables []string) (map[string][]string, error) {
	indexes := make(map[string][]string)
	
	for _, table := range tables {
		query := `
			SELECT indexname, indexdef 
			FROM pg_indexes 
			WHERE tablename = $1
		`
		
		rows, err := os.db.QueryContext(ctx, query, table)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		var tableIndexes []string
		for rows.Next() {
			var indexName, indexDef string
			if err := rows.Scan(&indexName, &indexDef); err != nil {
				continue
			}
			tableIndexes = append(tableIndexes, indexName)
		}
		
		indexes[table] = tableIndexes
	}
	
	return indexes, nil
}

// filterExistingIndexes 过滤已存在的索引
func (os *OptimizationService) filterExistingIndexes(suggestions []IndexSuggestion, existing map[string][]string) []IndexSuggestion {
	var filtered []IndexSuggestion
	
	for _, suggestion := range suggestions {
		exists := false
		if tableIndexes, ok := existing[suggestion.Table]; ok {
			for _, indexName := range tableIndexes {
				// 简化的索引匹配逻辑
				if strings.Contains(indexName, suggestion.Columns[0]) {
					exists = true
					break
				}
			}
		}
		
		if !exists {
			filtered = append(filtered, suggestion)
		}
	}
	
	return filtered
}

// estimateSpeedup 估算性能提升
func (os *OptimizationService) estimateSpeedup(info *QueryInfo, suggestions []IndexSuggestion) float64 {
	speedup := 1.0
	
	// 基于索引建议估算提升
	for _, suggestion := range suggestions {
		switch suggestion.Reason {
		case "Used in WHERE clause":
			speedup *= 2.0 // WHERE索引通常能带来2倍提升
		case "Used in ORDER BY clause":
			speedup *= 1.5 // ORDER BY索引通常能带来1.5倍提升
		case "Used in JOIN condition":
			speedup *= 3.0 // JOIN索引通常能带来3倍提升
		}
	}
	
	// 基于查询结构估算
	if !info.HasLimit {
		speedup *= 1.2 // 添加LIMIT能带来20%提升
	}
	
	return speedup
}

// generateExplanation 生成解释
func (os *OptimizationService) generateExplanation(info *QueryInfo, suggestions []IndexSuggestion) string {
	var explanations []string
	
	if len(suggestions) > 0 {
		explanations = append(explanations, fmt.Sprintf("建议创建%d个索引以提升查询性能", len(suggestions)))
	}
	
	if !info.HasLimit && info.Type == "SELECT" {
		explanations = append(explanations, "建议添加LIMIT子句以限制返回结果数量")
	}
	
	if len(info.Columns) == 1 && info.Columns[0] == "*" {
		explanations = append(explanations, "建议指定具体列名而不是使用SELECT *")
	}
	
	if len(explanations) == 0 {
		return "查询已经相对优化，无明显改进建议"
	}
	
	return strings.Join(explanations, "; ")
}

// RecordSlowQuery 记录慢查询
func (os *OptimizationService) RecordSlowQuery(query string, duration time.Duration) {
	if duration < os.slowQueries.threshold {
		return
	}
	
	os.slowQueries.mu.Lock()
	defer os.slowQueries.mu.Unlock()
	
	// 查找是否已存在相同查询
	for i, sq := range os.slowQueries.queries {
		if sq.Query == query {
			os.slowQueries.queries[i].Count++
			if duration > sq.Duration {
				os.slowQueries.queries[i].Duration = duration
			}
			return
		}
	}
	
	// 添加新的慢查询记录
	os.slowQueries.queries = append(os.slowQueries.queries, SlowQuery{
		Query:     query,
		Duration:  duration,
		Timestamp: time.Now(),
		Count:     1,
	})
}

// GetSlowQueries 获取慢查询列表
func (os *OptimizationService) GetSlowQueries() []SlowQuery {
	os.slowQueries.mu.RLock()
	defer os.slowQueries.mu.RUnlock()
	
	// 复制切片以避免并发问题
	queries := make([]SlowQuery, len(os.slowQueries.queries))
	copy(queries, os.slowQueries.queries)
	
	return queries
}

// CacheGet 从缓存获取数据
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	
	entry, exists := qc.cache[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(entry.ExpiresAt) {
		delete(qc.cache, key)
		return nil, false
	}
	
	return entry.Data, true
}

// CacheSet 设置缓存数据
func (qc *QueryCache) Set(key string, data interface{}) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	qc.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(qc.ttl),
	}
}

// CacheClear 清理过期缓存
func (qc *QueryCache) Clear() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	now := time.Now()
	for key, entry := range qc.cache {
		if now.After(entry.ExpiresAt) {
			delete(qc.cache, key)
		}
	}
}