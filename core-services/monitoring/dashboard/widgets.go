package dashboard

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/models"
)

// Widget 组件接口
type Widget interface {
	GetID() string
	GetType() string
	GetTitle() string
	GetConfig() map[string]interface{}
	Render(data interface{}) (map[string]interface{}, error)
	Validate() error
}

// BaseWidget 基础组件
type BaseWidget struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    WidgetPosition         `json:"position"`
	Config      map[string]interface{} `json:"config"`
	DataSource  DataSourceConfig       `json:"data_source"`
}

// WidgetPosition 组件位置
type WidgetPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	Type     string                 `json:"type"`     // prometheus, influxdb, static
	Query    string                 `json:"query"`    // 查询语句
	Interval time.Duration          `json:"interval"` // 刷新间隔
	Params   map[string]interface{} `json:"params"`   // 额外参数
}

// GetID 获取组件ID
func (bw *BaseWidget) GetID() string {
	return bw.ID
}

// GetType 获取组件类型
func (bw *BaseWidget) GetType() string {
	return bw.Type
}

// GetTitle 获取组件标题
func (bw *BaseWidget) GetTitle() string {
	return bw.Title
}

// GetConfig 获取组件配置
func (bw *BaseWidget) GetConfig() map[string]interface{} {
	return bw.Config
}

// Validate 验证组件配置
func (bw *BaseWidget) Validate() error {
	if bw.ID == "" {
		return fmt.Errorf("widget ID is required")
	}
	if bw.Type == "" {
		return fmt.Errorf("widget type is required")
	}
	if bw.Title == "" {
		return fmt.Errorf("widget title is required")
	}
	return nil
}

// MetricWidget 指标组件
type MetricWidget struct {
	BaseWidget
	MetricName string            `json:"metric_name"`
	Unit       string            `json:"unit"`
	Format     string            `json:"format"`
	Thresholds []MetricThreshold `json:"thresholds"`
}

// MetricThreshold 指标阈值
type MetricThreshold struct {
	Value     float64 `json:"value"`
	Color     string  `json:"color"`
	Condition string  `json:"condition"` // >, <, >=, <=, ==, !=
}

// NewMetricWidget 创建指标组件
func NewMetricWidget(id, title, metricName string) *MetricWidget {
	return &MetricWidget{
		BaseWidget: BaseWidget{
			ID:    id,
			Type:  "metric",
			Title: title,
			Config: map[string]interface{}{
				"show_trend": true,
				"show_sparkline": false,
			},
		},
		MetricName: metricName,
		Unit:       "",
		Format:     "number",
	}
}

// Render 渲染指标组件
func (mw *MetricWidget) Render(data interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":          mw.ID,
		"type":        mw.Type,
		"title":       mw.Title,
		"description": mw.Description,
		"position":    mw.Position,
		"config":      mw.Config,
	}
	
	// 处理数据
	if metrics, ok := data.([]*models.Metric); ok && len(metrics) > 0 {
		metric := metrics[0]
		
		// 获取当前值
		var currentValue float64
		if len(metric.Samples) > 0 {
			currentValue = metric.Samples[len(metric.Samples)-1].Value
		}
		
		// 格式化值
		formattedValue := mw.formatValue(currentValue)
		
		// 确定颜色
		color := mw.getThresholdColor(currentValue)
		
		result["data"] = map[string]interface{}{
			"value":           currentValue,
			"formatted_value": formattedValue,
			"unit":           mw.Unit,
			"color":          color,
			"timestamp":      time.Now().Unix(),
		}
		
		// 添加趋势数据
		if mw.Config["show_trend"].(bool) && len(metric.Samples) > 1 {
			trend := mw.calculateTrend(metric.Samples)
			result["data"].(map[string]interface{})["trend"] = trend
		}
		
		// 添加迷你图数据
		if mw.Config["show_sparkline"].(bool) {
			sparkline := mw.generateSparkline(metric.Samples)
			result["data"].(map[string]interface{})["sparkline"] = sparkline
		}
	}
	
	return result, nil
}

// formatValue 格式化值
func (mw *MetricWidget) formatValue(value float64) string {
	switch mw.Format {
	case "bytes":
		return formatBytes(value)
	case "percent":
		return fmt.Sprintf("%.2f%%", value)
	case "duration":
		return formatDuration(value)
	default:
		return fmt.Sprintf("%.2f", value)
	}
}

// getThresholdColor 获取阈值颜色
func (mw *MetricWidget) getThresholdColor(value float64) string {
	for _, threshold := range mw.Thresholds {
		if mw.checkThreshold(value, threshold) {
			return threshold.Color
		}
	}
	return "#28a745" // 默认绿色
}

// checkThreshold 检查阈值
func (mw *MetricWidget) checkThreshold(value float64, threshold MetricThreshold) bool {
	switch threshold.Condition {
	case ">":
		return value > threshold.Value
	case "<":
		return value < threshold.Value
	case ">=":
		return value >= threshold.Value
	case "<=":
		return value <= threshold.Value
	case "==":
		return value == threshold.Value
	case "!=":
		return value != threshold.Value
	default:
		return false
	}
}

// calculateTrend 计算趋势
func (mw *MetricWidget) calculateTrend(samples []*models.Sample) map[string]interface{} {
	if len(samples) < 2 {
		return map[string]interface{}{
			"direction": "stable",
			"percentage": 0.0,
		}
	}
	
	current := samples[len(samples)-1].Value
	previous := samples[len(samples)-2].Value
	
	if previous == 0 {
		return map[string]interface{}{
			"direction": "stable",
			"percentage": 0.0,
		}
	}
	
	change := (current - previous) / previous * 100
	
	direction := "stable"
	if change > 0.1 {
		direction = "up"
	} else if change < -0.1 {
		direction = "down"
	}
	
	return map[string]interface{}{
		"direction":  direction,
		"percentage": change,
	}
}

// generateSparkline 生成迷你图数据
func (mw *MetricWidget) generateSparkline(samples []*models.Sample) []float64 {
	var values []float64
	for _, sample := range samples {
		values = append(values, sample.Value)
	}
	return values
}

// ChartWidget 图表组件
type ChartWidget struct {
	BaseWidget
	ChartType   string              `json:"chart_type"`   // line, bar, pie, area
	Series      []ChartSeries       `json:"series"`
	XAxis       ChartAxis           `json:"x_axis"`
	YAxis       ChartAxis           `json:"y_axis"`
	Legend      ChartLegend         `json:"legend"`
	Colors      []string            `json:"colors"`
	Annotations []ChartAnnotation   `json:"annotations"`
}

// ChartSeries 图表系列
type ChartSeries struct {
	Name   string `json:"name"`
	Query  string `json:"query"`
	Color  string `json:"color"`
	Type   string `json:"type"`   // line, bar, area
	YAxis  int    `json:"y_axis"` // 0 for left, 1 for right
}

// ChartAxis 图表轴
type ChartAxis struct {
	Title    string  `json:"title"`
	Min      *float64 `json:"min"`
	Max      *float64 `json:"max"`
	Unit     string  `json:"unit"`
	Format   string  `json:"format"`
	LogScale bool    `json:"log_scale"`
}

// ChartLegend 图表图例
type ChartLegend struct {
	Show     bool   `json:"show"`
	Position string `json:"position"` // top, bottom, left, right
}

// ChartAnnotation 图表注释
type ChartAnnotation struct {
	Type        string    `json:"type"`        // line, region
	Value       float64   `json:"value"`
	StartValue  float64   `json:"start_value"`
	EndValue    float64   `json:"end_value"`
	Color       string    `json:"color"`
	Label       string    `json:"label"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewChartWidget 创建图表组件
func NewChartWidget(id, title, chartType string) *ChartWidget {
	return &ChartWidget{
		BaseWidget: BaseWidget{
			ID:    id,
			Type:  "chart",
			Title: title,
			Config: map[string]interface{}{
				"show_grid":   true,
				"show_points": false,
				"smooth":      true,
			},
		},
		ChartType: chartType,
		XAxis: ChartAxis{
			Title:  "时间",
			Format: "time",
		},
		YAxis: ChartAxis{
			Title:  "值",
			Format: "number",
		},
		Legend: ChartLegend{
			Show:     true,
			Position: "bottom",
		},
		Colors: []string{"#007bff", "#28a745", "#ffc107", "#dc3545", "#6f42c1"},
	}
}

// Render 渲染图表组件
func (cw *ChartWidget) Render(data interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":          cw.ID,
		"type":        cw.Type,
		"title":       cw.Title,
		"description": cw.Description,
		"position":    cw.Position,
		"config":      cw.Config,
		"chart_type":  cw.ChartType,
		"x_axis":      cw.XAxis,
		"y_axis":      cw.YAxis,
		"legend":      cw.Legend,
		"colors":      cw.Colors,
	}
	
	// 处理数据
	if seriesData, ok := data.(map[string][]*models.Metric); ok {
		chartData := make(map[string]interface{})
		
		for seriesName, metrics := range seriesData {
			if len(metrics) > 0 {
				points := make([]map[string]interface{}, 0)
				
				for _, metric := range metrics {
					for _, sample := range metric.Samples {
						points = append(points, map[string]interface{}{
							"x": sample.Timestamp.Unix() * 1000, // JavaScript时间戳
							"y": sample.Value,
						})
					}
				}
				
				chartData[seriesName] = points
			}
		}
		
		result["data"] = chartData
	}
	
	return result, nil
}

// TableWidget 表格组件
type TableWidget struct {
	BaseWidget
	Columns     []TableColumn `json:"columns"`
	Pagination  bool          `json:"pagination"`
	PageSize    int           `json:"page_size"`
	Sortable    bool          `json:"sortable"`
	Searchable  bool          `json:"searchable"`
}

// TableColumn 表格列
type TableColumn struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	Width     int    `json:"width"`
	Sortable  bool   `json:"sortable"`
	Format    string `json:"format"`
	Alignment string `json:"alignment"` // left, center, right
}

// NewTableWidget 创建表格组件
func NewTableWidget(id, title string) *TableWidget {
	return &TableWidget{
		BaseWidget: BaseWidget{
			ID:    id,
			Type:  "table",
			Title: title,
			Config: map[string]interface{}{
				"striped": true,
				"bordered": true,
				"hover": true,
			},
		},
		Pagination: true,
		PageSize:   10,
		Sortable:   true,
		Searchable: true,
	}
}

// Render 渲染表格组件
func (tw *TableWidget) Render(data interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":          tw.ID,
		"type":        tw.Type,
		"title":       tw.Title,
		"description": tw.Description,
		"position":    tw.Position,
		"config":      tw.Config,
		"columns":     tw.Columns,
		"pagination":  tw.Pagination,
		"page_size":   tw.PageSize,
		"sortable":    tw.Sortable,
		"searchable":  tw.Searchable,
	}
	
	// 处理数据
	if rows, ok := data.([]map[string]interface{}); ok {
		result["data"] = rows
		result["total"] = len(rows)
	}
	
	return result, nil
}

// StatWidget 统计组件
type StatWidget struct {
	BaseWidget
	Stats []StatItem `json:"stats"`
}

// StatItem 统计项
type StatItem struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Color string  `json:"color"`
	Icon  string  `json:"icon"`
}

// NewStatWidget 创建统计组件
func NewStatWidget(id, title string) *StatWidget {
	return &StatWidget{
		BaseWidget: BaseWidget{
			ID:    id,
			Type:  "stat",
			Title: title,
			Config: map[string]interface{}{
				"layout": "horizontal",
			},
		},
	}
}

// Render 渲染统计组件
func (sw *StatWidget) Render(data interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":          sw.ID,
		"type":        sw.Type,
		"title":       sw.Title,
		"description": sw.Description,
		"position":    sw.Position,
		"config":      sw.Config,
		"stats":       sw.Stats,
	}
	
	// 处理数据
	if stats, ok := data.([]StatItem); ok {
		result["data"] = stats
	}
	
	return result, nil
}

// AlertWidget 告警组件
type AlertWidget struct {
	BaseWidget
	Severity    []string `json:"severity"`    // 显示的严重级别
	MaxAlerts   int      `json:"max_alerts"`  // 最大显示数量
	ShowResolved bool    `json:"show_resolved"` // 是否显示已解决的告警
}

// NewAlertWidget 创建告警组件
func NewAlertWidget(id, title string) *AlertWidget {
	return &AlertWidget{
		BaseWidget: BaseWidget{
			ID:    id,
			Type:  "alert",
			Title: title,
			Config: map[string]interface{}{
				"auto_refresh": true,
				"refresh_interval": 30,
			},
		},
		Severity:     []string{"critical", "high", "medium", "low"},
		MaxAlerts:    20,
		ShowResolved: false,
	}
}

// Render 渲染告警组件
func (aw *AlertWidget) Render(data interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":            aw.ID,
		"type":          aw.Type,
		"title":         aw.Title,
		"description":   aw.Description,
		"position":      aw.Position,
		"config":        aw.Config,
		"severity":      aw.Severity,
		"max_alerts":    aw.MaxAlerts,
		"show_resolved": aw.ShowResolved,
	}
	
	// 处理数据
	if alerts, ok := data.([]*models.Alert); ok {
		// 过滤告警
		filteredAlerts := make([]*models.Alert, 0)
		for _, alert := range alerts {
			// 检查严重级别
			severityMatch := false
			for _, severity := range aw.Severity {
				if string(alert.Severity) == severity {
					severityMatch = true
					break
				}
			}
			
			if !severityMatch {
				continue
			}
			
			// 检查是否显示已解决的告警
			if !aw.ShowResolved && alert.Status == models.AlertStatusResolved {
				continue
			}
			
			filteredAlerts = append(filteredAlerts, alert)
			
			// 检查最大数量
			if len(filteredAlerts) >= aw.MaxAlerts {
				break
			}
		}
		
		result["data"] = filteredAlerts
		result["total"] = len(filteredAlerts)
	}
	
	return result, nil
}

// 工具函数

// formatBytes 格式化字节数
func formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", bytes/float64(div), "KMGTPE"[exp])
}

// formatDuration 格式化持续时间
func formatDuration(seconds float64) string {
	duration := time.Duration(seconds) * time.Second
	
	if duration < time.Minute {
		return fmt.Sprintf("%.1fs", duration.Seconds())
	} else if duration < time.Hour {
		return fmt.Sprintf("%.1fm", duration.Minutes())
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%.1fh", duration.Hours())
	} else {
		return fmt.Sprintf("%.1fd", duration.Hours()/24)
	}
}

// WidgetFactory 组件工厂
type WidgetFactory struct{}

// CreateWidget 创建组件
func (wf *WidgetFactory) CreateWidget(widgetType, id, title string) (Widget, error) {
	switch widgetType {
	case "metric":
		return NewMetricWidget(id, title, ""), nil
	case "chart":
		return NewChartWidget(id, title, "line"), nil
	case "table":
		return NewTableWidget(id, title), nil
	case "stat":
		return NewStatWidget(id, title), nil
	case "alert":
		return NewAlertWidget(id, title), nil
	default:
		return nil, fmt.Errorf("unknown widget type: %s", widgetType)
	}
}

// CreateWidgetFromJSON 从JSON创建组件
func (wf *WidgetFactory) CreateWidgetFromJSON(data []byte) (Widget, error) {
	var base BaseWidget
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to parse base widget: %w", err)
	}
	
	switch base.Type {
	case "metric":
		var widget MetricWidget
		if err := json.Unmarshal(data, &widget); err != nil {
			return nil, fmt.Errorf("failed to parse metric widget: %w", err)
		}
		return &widget, nil
	case "chart":
		var widget ChartWidget
		if err := json.Unmarshal(data, &widget); err != nil {
			return nil, fmt.Errorf("failed to parse chart widget: %w", err)
		}
		return &widget, nil
	case "table":
		var widget TableWidget
		if err := json.Unmarshal(data, &widget); err != nil {
			return nil, fmt.Errorf("failed to parse table widget: %w", err)
		}
		return &widget, nil
	case "stat":
		var widget StatWidget
		if err := json.Unmarshal(data, &widget); err != nil {
			return nil, fmt.Errorf("failed to parse stat widget: %w", err)
		}
		return &widget, nil
	case "alert":
		var widget AlertWidget
		if err := json.Unmarshal(data, &widget); err != nil {
			return nil, fmt.Errorf("failed to parse alert widget: %w", err)
		}
		return &widget, nil
	default:
		return nil, fmt.Errorf("unknown widget type: %s", base.Type)
	}
}