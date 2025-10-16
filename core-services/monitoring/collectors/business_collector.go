package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// BusinessCollector ?
type BusinessCollector struct {
	name     string
	interval time.Duration
	enabled  bool
	labels   map[string]string
	
	// ?
	db *sql.DB
	
	// 
	collectUsers      bool
	collectOrders     bool
	collectPayments   bool
	collectContent    bool
	collectEngagement bool
	collectPerformance bool
	
	// 
	userMetrics       *UserMetrics
	orderMetrics      *OrderMetrics
	paymentMetrics    *PaymentMetrics
	contentMetrics    *ContentMetrics
	engagementMetrics *EngagementMetrics
	performanceMetrics *PerformanceMetrics
	
	// ?
	mutex sync.RWMutex
	
	// ?
	lastCollectTime time.Time
}

// BusinessCollectorConfig ?
type BusinessCollectorConfig struct {
	Interval           time.Duration     `yaml:"interval"`
	Enabled            bool              `yaml:"enabled"`
	Labels             map[string]string `yaml:"labels"`
	CollectUsers       bool              `yaml:"collect_users"`
	CollectOrders      bool              `yaml:"collect_orders"`
	CollectPayments    bool              `yaml:"collect_payments"`
	CollectContent     bool              `yaml:"collect_content"`
	CollectEngagement  bool              `yaml:"collect_engagement"`
	CollectPerformance bool              `yaml:"collect_performance"`
}

// UserMetrics 
type UserMetrics struct {
	TotalUsers       uint64            `json:"total_users"`
	ActiveUsers      uint64            `json:"active_users"`
	NewUsers         uint64            `json:"new_users"`
	RetentionRate    float64           `json:"retention_rate"`
	ChurnRate        float64           `json:"churn_rate"`
	UsersByRegion    map[string]uint64 `json:"users_by_region"`
	UsersByPlatform  map[string]uint64 `json:"users_by_platform"`
	AverageSessionTime time.Duration   `json:"average_session_time"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// OrderMetrics 
type OrderMetrics struct {
	TotalOrders      uint64            `json:"total_orders"`
	CompletedOrders  uint64            `json:"completed_orders"`
	CancelledOrders  uint64            `json:"cancelled_orders"`
	PendingOrders    uint64            `json:"pending_orders"`
	OrderValue       float64           `json:"order_value"`
	AverageOrderValue float64          `json:"average_order_value"`
	OrdersByStatus   map[string]uint64 `json:"orders_by_status"`
	OrdersByRegion   map[string]uint64 `json:"orders_by_region"`
	ConversionRate   float64           `json:"conversion_rate"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// PaymentMetrics 
type PaymentMetrics struct {
	TotalPayments     uint64            `json:"total_payments"`
	SuccessfulPayments uint64           `json:"successful_payments"`
	FailedPayments    uint64            `json:"failed_payments"`
	PaymentValue      float64           `json:"payment_value"`
	PaymentsByMethod  map[string]uint64 `json:"payments_by_method"`
	PaymentsByStatus  map[string]uint64 `json:"payments_by_status"`
	SuccessRate       float64           `json:"success_rate"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// ContentMetrics 
type ContentMetrics struct {
	TotalContent     uint64            `json:"total_content"`
	PublishedContent uint64            `json:"published_content"`
	DraftContent     uint64            `json:"draft_content"`
	ViewCount        uint64            `json:"view_count"`
	LikeCount        uint64            `json:"like_count"`
	ShareCount       uint64            `json:"share_count"`
	CommentCount     uint64            `json:"comment_count"`
	ContentByType    map[string]uint64 `json:"content_by_type"`
	ContentByAuthor  map[string]uint64 `json:"content_by_author"`
	EngagementRate   float64           `json:"engagement_rate"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// EngagementMetrics ?
type EngagementMetrics struct {
	PageViews        uint64            `json:"page_views"`
	UniqueVisitors   uint64            `json:"unique_visitors"`
	BounceRate       float64           `json:"bounce_rate"`
	SessionDuration  time.Duration     `json:"session_duration"`
	PagesPerSession  float64           `json:"pages_per_session"`
	ClickThroughRate float64           `json:"click_through_rate"`
	InteractionsByType map[string]uint64 `json:"interactions_by_type"`
	DeviceTypes      map[string]uint64 `json:"device_types"`
	TrafficSources   map[string]uint64 `json:"traffic_sources"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// PerformanceMetrics 
type PerformanceMetrics struct {
	Revenue          float64           `json:"revenue"`
	Profit           float64           `json:"profit"`
	ROI              float64           `json:"roi"`
	CustomerLifetimeValue float64      `json:"customer_lifetime_value"`
	CustomerAcquisitionCost float64    `json:"customer_acquisition_cost"`
	MonthlyRecurringRevenue float64    `json:"monthly_recurring_revenue"`
	ChurnRevenue     float64           `json:"churn_revenue"`
	KPIsByCategory   map[string]float64 `json:"kpis_by_category"`
	GrowthRate       float64           `json:"growth_rate"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// NewBusinessCollector ?
func NewBusinessCollector(config BusinessCollectorConfig, db *sql.DB) *BusinessCollector {
	labels := map[string]string{
		"collector": "business",
		"service":   "core-services",
	}
	
	// ?
	for k, v := range config.Labels {
		labels[k] = v
	}
	
	return &BusinessCollector{
		name:               "business",
		interval:           config.Interval,
		enabled:            config.Enabled,
		labels:             labels,
		db:                 db,
		collectUsers:       config.CollectUsers,
		collectOrders:      config.CollectOrders,
		collectPayments:    config.CollectPayments,
		collectContent:     config.CollectContent,
		collectEngagement:  config.CollectEngagement,
		collectPerformance: config.CollectPerformance,
		userMetrics:        &UserMetrics{UsersByRegion: make(map[string]uint64), UsersByPlatform: make(map[string]uint64)},
		orderMetrics:       &OrderMetrics{OrdersByStatus: make(map[string]uint64), OrdersByRegion: make(map[string]uint64)},
		paymentMetrics:     &PaymentMetrics{PaymentsByMethod: make(map[string]uint64), PaymentsByStatus: make(map[string]uint64)},
		contentMetrics:     &ContentMetrics{ContentByType: make(map[string]uint64), ContentByAuthor: make(map[string]uint64)},
		engagementMetrics:  &EngagementMetrics{InteractionsByType: make(map[string]uint64), DeviceTypes: make(map[string]uint64), TrafficSources: make(map[string]uint64)},
		performanceMetrics: &PerformanceMetrics{KPIsByCategory: make(map[string]float64)},
		lastCollectTime:    time.Now(),
	}
}

// GetName ?
func (c *BusinessCollector) GetName() string {
	return c.name
}

// GetCategory ?
func (c *BusinessCollector) GetCategory() models.MetricCategory {
	return models.CategoryBusiness
}

// GetInterval 
func (c *BusinessCollector) GetInterval() time.Duration {
	return c.interval
}

// IsEnabled ?
func (c *BusinessCollector) IsEnabled() bool {
	return c.enabled
}

// Start ?
func (c *BusinessCollector) Start(ctx context.Context) error {
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
				fmt.Printf("Business collector error: %v\n", err)
			}
		}
	}
}

// Stop ?
func (c *BusinessCollector) Stop() error {
	c.enabled = false
	return nil
}

// Health ?
func (c *BusinessCollector) Health() error {
	if !c.enabled {
		return fmt.Errorf("business collector is disabled")
	}
	
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// 
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}

// Collect 
func (c *BusinessCollector) Collect(ctx context.Context) ([]models.Metric, error) {
	if !c.enabled || c.db == nil {
		return nil, nil
	}
	
	var metrics []models.Metric
	now := time.Now()
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// 
	if c.collectUsers {
		userMetrics, err := c.collectUserMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect user metrics: %w", err)
		}
		metrics = append(metrics, userMetrics...)
	}
	
	// 
	if c.collectOrders {
		orderMetrics, err := c.collectOrderMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect order metrics: %w", err)
		}
		metrics = append(metrics, orderMetrics...)
	}
	
	// 
	if c.collectPayments {
		paymentMetrics, err := c.collectPaymentMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect payment metrics: %w", err)
		}
		metrics = append(metrics, paymentMetrics...)
	}
	
	// 
	if c.collectContent {
		contentMetrics, err := c.collectContentMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect content metrics: %w", err)
		}
		metrics = append(metrics, contentMetrics...)
	}
	
	// ?
	if c.collectEngagement {
		engagementMetrics, err := c.collectEngagementMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect engagement metrics: %w", err)
		}
		metrics = append(metrics, engagementMetrics...)
	}
	
	// 
	if c.collectPerformance {
		performanceMetrics, err := c.collectPerformanceMetrics(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect performance metrics: %w", err)
		}
		metrics = append(metrics, performanceMetrics...)
	}
	
	c.lastCollectTime = now
	return metrics, nil
}

// collectUserMetrics 
func (c *BusinessCollector) collectUserMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 
	var totalUsers uint64
	err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, err
	}
	c.userMetrics.TotalUsers = totalUsers
	
	metric := models.NewMetric("business_users_total", models.MetricTypeGauge, models.CategoryBusiness).
		WithLabels(c.labels).
		WithValue(float64(totalUsers)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "users"
	metric.Description = "Total number of users"
	metrics = append(metrics, *metric)
	
	// ?0
	var activeUsers uint64
	err = c.db.QueryRowContext(ctx, 
		"SELECT COUNT(DISTINCT user_id) FROM user_sessions WHERE created_at > NOW() - INTERVAL '30 days'").Scan(&activeUsers)
	if err == nil {
		c.userMetrics.ActiveUsers = activeUsers
		
		metric = models.NewMetric("business_users_active", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(activeUsers)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "users"
		metric.Description = "Number of active users (last 30 days)"
		metrics = append(metrics, *metric)
	}
	
	// ?4?
	var newUsers uint64
	err = c.db.QueryRowContext(ctx, 
		"SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&newUsers)
	if err == nil {
		c.userMetrics.NewUsers = newUsers
		
		metric = models.NewMetric("business_users_new", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(newUsers)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "users"
		metric.Description = "Number of new users (last 24 hours)"
		metrics = append(metrics, *metric)
	}
	
	// ?
	if totalUsers > 0 && activeUsers > 0 {
		retentionRate := float64(activeUsers) / float64(totalUsers) * 100
		c.userMetrics.RetentionRate = retentionRate
		
		metric = models.NewMetric("business_users_retention_rate", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(retentionRate).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "User retention rate"
		metrics = append(metrics, *metric)
	}
	
	// ?
	rows, err := c.db.QueryContext(ctx, 
		"SELECT region, COUNT(*) FROM users WHERE region IS NOT NULL GROUP BY region")
	if err == nil {
		defer rows.Close()
		
		for rows.Next() {
			var region string
			var count uint64
			if err := rows.Scan(&region, &count); err != nil {
				continue
			}
			
			c.userMetrics.UsersByRegion[region] = count
			
			labels := make(map[string]string)
			for k, v := range c.labels {
				labels[k] = v
			}
			labels["region"] = region
			
			metric = models.NewMetric("business_users_by_region", models.MetricTypeGauge, models.CategoryBusiness).
				WithLabels(labels).
				WithValue(float64(count)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "users"
			metric.Description = "Number of users by region"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	var avgSessionTime sql.NullFloat64
	err = c.db.QueryRowContext(ctx, 
		"SELECT AVG(EXTRACT(EPOCH FROM (ended_at - created_at))) FROM user_sessions WHERE ended_at IS NOT NULL AND created_at > NOW() - INTERVAL '7 days'").Scan(&avgSessionTime)
	if err == nil && avgSessionTime.Valid {
		sessionDuration := time.Duration(avgSessionTime.Float64) * time.Second
		c.userMetrics.AverageSessionTime = sessionDuration
		
		metric = models.NewMetric("business_users_avg_session_duration_seconds", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(sessionDuration.Seconds()).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Average user session duration"
		metrics = append(metrics, *metric)
	}
	
	c.userMetrics.LastUpdated = timestamp
	return metrics, nil
}

// collectOrderMetrics 
func (c *BusinessCollector) collectOrderMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 
	var totalOrders uint64
	err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&totalOrders)
	if err != nil {
		return nil, err
	}
	c.orderMetrics.TotalOrders = totalOrders
	
	metric := models.NewMetric("business_orders_total", models.MetricTypeGauge, models.CategoryBusiness).
		WithLabels(c.labels).
		WithValue(float64(totalOrders)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "orders"
	metric.Description = "Total number of orders"
	metrics = append(metrics, *metric)
	
	// 
	var completedOrders uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'completed'").Scan(&completedOrders)
	if err == nil {
		c.orderMetrics.CompletedOrders = completedOrders
		
		metric = models.NewMetric("business_orders_completed", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(completedOrders)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "orders"
		metric.Description = "Number of completed orders"
		metrics = append(metrics, *metric)
	}
	
	// 
	var cancelledOrders uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'cancelled'").Scan(&cancelledOrders)
	if err == nil {
		c.orderMetrics.CancelledOrders = cancelledOrders
		
		metric = models.NewMetric("business_orders_cancelled", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(cancelledOrders)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "orders"
		metric.Description = "Number of cancelled orders"
		metrics = append(metrics, *metric)
	}
	
	// 
	var pendingOrders uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'pending'").Scan(&pendingOrders)
	if err == nil {
		c.orderMetrics.PendingOrders = pendingOrders
		
		metric = models.NewMetric("business_orders_pending", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(pendingOrders)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "orders"
		metric.Description = "Number of pending orders"
		metrics = append(metrics, *metric)
	}
	
	// ?
	var orderValue sql.NullFloat64
	err = c.db.QueryRowContext(ctx, "SELECT SUM(total_amount) FROM orders WHERE status = 'completed'").Scan(&orderValue)
	if err == nil && orderValue.Valid {
		c.orderMetrics.OrderValue = orderValue.Float64
		
		metric = models.NewMetric("business_orders_total_value", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(orderValue.Float64).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "currency"
		metric.Description = "Total value of completed orders"
		metrics = append(metrics, *metric)
	}
	
	// ?
	if completedOrders > 0 && orderValue.Valid {
		avgOrderValue := orderValue.Float64 / float64(completedOrders)
		c.orderMetrics.AverageOrderValue = avgOrderValue
		
		metric = models.NewMetric("business_orders_avg_value", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(avgOrderValue).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "currency"
		metric.Description = "Average order value"
		metrics = append(metrics, *metric)
	}
	
	// ?
	rows, err := c.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM orders GROUP BY status")
	if err == nil {
		defer rows.Close()
		
		for rows.Next() {
			var status string
			var count uint64
			if err := rows.Scan(&status, &count); err != nil {
				continue
			}
			
			c.orderMetrics.OrdersByStatus[status] = count
			
			labels := make(map[string]string)
			for k, v := range c.labels {
				labels[k] = v
			}
			labels["status"] = status
			
			metric = models.NewMetric("business_orders_by_status", models.MetricTypeGauge, models.CategoryBusiness).
				WithLabels(labels).
				WithValue(float64(count)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "orders"
			metric.Description = "Number of orders by status"
			metrics = append(metrics, *metric)
		}
	}
	
	// ??
	if totalOrders > 0 {
		conversionRate := float64(completedOrders) / float64(totalOrders) * 100
		c.orderMetrics.ConversionRate = conversionRate
		
		metric = models.NewMetric("business_orders_conversion_rate", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(conversionRate).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Order conversion rate"
		metrics = append(metrics, *metric)
	}
	
	c.orderMetrics.LastUpdated = timestamp
	return metrics, nil
}

// collectPaymentMetrics 
func (c *BusinessCollector) collectPaymentMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 
	var totalPayments uint64
	err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments").Scan(&totalPayments)
	if err != nil {
		return nil, err
	}
	c.paymentMetrics.TotalPayments = totalPayments
	
	metric := models.NewMetric("business_payments_total", models.MetricTypeGauge, models.CategoryBusiness).
		WithLabels(c.labels).
		WithValue(float64(totalPayments)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "payments"
	metric.Description = "Total number of payments"
	metrics = append(metrics, *metric)
	
	// ?
	var successfulPayments uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments WHERE status = 'success'").Scan(&successfulPayments)
	if err == nil {
		c.paymentMetrics.SuccessfulPayments = successfulPayments
		
		metric = models.NewMetric("business_payments_successful", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(successfulPayments)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "payments"
		metric.Description = "Number of successful payments"
		metrics = append(metrics, *metric)
	}
	
	// ?
	var failedPayments uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments WHERE status = 'failed'").Scan(&failedPayments)
	if err == nil {
		c.paymentMetrics.FailedPayments = failedPayments
		
		metric = models.NewMetric("business_payments_failed", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(failedPayments)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "payments"
		metric.Description = "Number of failed payments"
		metrics = append(metrics, *metric)
	}
	
	// ?
	if totalPayments > 0 {
		successRate := float64(successfulPayments) / float64(totalPayments) * 100
		c.paymentMetrics.SuccessRate = successRate
		
		metric = models.NewMetric("business_payments_success_rate", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(successRate).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Payment success rate"
		metrics = append(metrics, *metric)
	}
	
	// ?
	var paymentValue sql.NullFloat64
	err = c.db.QueryRowContext(ctx, "SELECT SUM(amount) FROM payments WHERE status = 'success'").Scan(&paymentValue)
	if err == nil && paymentValue.Valid {
		c.paymentMetrics.PaymentValue = paymentValue.Float64
		
		metric = models.NewMetric("business_payments_total_value", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(paymentValue.Float64).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "currency"
		metric.Description = "Total value of successful payments"
		metrics = append(metrics, *metric)
	}
	
	// ?
	rows, err := c.db.QueryContext(ctx, "SELECT payment_method, COUNT(*) FROM payments GROUP BY payment_method")
	if err == nil {
		defer rows.Close()
		
		for rows.Next() {
			var method string
			var count uint64
			if err := rows.Scan(&method, &count); err != nil {
				continue
			}
			
			c.paymentMetrics.PaymentsByMethod[method] = count
			
			labels := make(map[string]string)
			for k, v := range c.labels {
				labels[k] = v
			}
			labels["method"] = method
			
			metric = models.NewMetric("business_payments_by_method", models.MetricTypeGauge, models.CategoryBusiness).
				WithLabels(labels).
				WithValue(float64(count)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "payments"
			metric.Description = "Number of payments by method"
			metrics = append(metrics, *metric)
		}
	}
	
	c.paymentMetrics.LastUpdated = timestamp
	return metrics, nil
}

// collectContentMetrics 
func (c *BusinessCollector) collectContentMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 
	var totalContent uint64
	err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM content").Scan(&totalContent)
	if err != nil {
		return nil, err
	}
	c.contentMetrics.TotalContent = totalContent
	
	metric := models.NewMetric("business_content_total", models.MetricTypeGauge, models.CategoryBusiness).
		WithLabels(c.labels).
		WithValue(float64(totalContent)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "content"
	metric.Description = "Total number of content items"
	metrics = append(metrics, *metric)
	
	// 
	var publishedContent uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM content WHERE status = 'published'").Scan(&publishedContent)
	if err == nil {
		c.contentMetrics.PublishedContent = publishedContent
		
		metric = models.NewMetric("business_content_published", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(publishedContent)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "content"
		metric.Description = "Number of published content items"
		metrics = append(metrics, *metric)
	}
	
	// ?
	var draftContent uint64
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM content WHERE status = 'draft'").Scan(&draftContent)
	if err == nil {
		c.contentMetrics.DraftContent = draftContent
		
		metric = models.NewMetric("business_content_draft", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(draftContent)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "content"
		metric.Description = "Number of draft content items"
		metrics = append(metrics, *metric)
	}
	
	// 
	var viewCount sql.NullInt64
	err = c.db.QueryRowContext(ctx, "SELECT SUM(view_count) FROM content").Scan(&viewCount)
	if err == nil && viewCount.Valid {
		c.contentMetrics.ViewCount = uint64(viewCount.Int64)
		
		metric = models.NewMetric("business_content_views_total", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(viewCount.Int64)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "views"
		metric.Description = "Total content views"
		metrics = append(metrics, *metric)
	}
	
	// 
	var likeCount sql.NullInt64
	err = c.db.QueryRowContext(ctx, "SELECT SUM(like_count) FROM content").Scan(&likeCount)
	if err == nil && likeCount.Valid {
		c.contentMetrics.LikeCount = uint64(likeCount.Int64)
		
		metric = models.NewMetric("business_content_likes_total", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(likeCount.Int64)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "likes"
		metric.Description = "Total content likes"
		metrics = append(metrics, *metric)
	}
	
	// ?
	rows, err := c.db.QueryContext(ctx, "SELECT content_type, COUNT(*) FROM content GROUP BY content_type")
	if err == nil {
		defer rows.Close()
		
		for rows.Next() {
			var contentType string
			var count uint64
			if err := rows.Scan(&contentType, &count); err != nil {
				continue
			}
			
			c.contentMetrics.ContentByType[contentType] = count
			
			labels := make(map[string]string)
			for k, v := range c.labels {
				labels[k] = v
			}
			labels["type"] = contentType
			
			metric = models.NewMetric("business_content_by_type", models.MetricTypeGauge, models.CategoryBusiness).
				WithLabels(labels).
				WithValue(float64(count)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "content"
			metric.Description = "Number of content items by type"
			metrics = append(metrics, *metric)
		}
	}
	
	c.contentMetrics.LastUpdated = timestamp
	return metrics, nil
}

// collectEngagementMetrics ?
func (c *BusinessCollector) collectEngagementMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// ?4?
	var pageViews uint64
	err := c.db.QueryRowContext(ctx, 
		"SELECT COUNT(*) FROM page_views WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&pageViews)
	if err == nil {
		c.engagementMetrics.PageViews = pageViews
		
		metric := models.NewMetric("business_engagement_page_views", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(pageViews)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "views"
		metric.Description = "Page views in last 24 hours"
		metrics = append(metrics, *metric)
	}
	
	// ?4?
	var uniqueVisitors uint64
	err = c.db.QueryRowContext(ctx, 
		"SELECT COUNT(DISTINCT user_id) FROM page_views WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&uniqueVisitors)
	if err == nil {
		c.engagementMetrics.UniqueVisitors = uniqueVisitors
		
		metric := models.NewMetric("business_engagement_unique_visitors", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(float64(uniqueVisitors)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "visitors"
		metric.Description = "Unique visitors in last 24 hours"
		metrics = append(metrics, *metric)
	}
	
	// ?
	if pageViews > 0 && uniqueVisitors > 0 {
		// 㵥 / 
		var singlePageSessions uint64
		err = c.db.QueryRowContext(ctx, 
			"SELECT COUNT(*) FROM (SELECT user_id FROM page_views WHERE created_at > NOW() - INTERVAL '24 hours' GROUP BY user_id HAVING COUNT(*) = 1) AS single_page").Scan(&singlePageSessions)
		if err == nil {
			bounceRate := float64(singlePageSessions) / float64(uniqueVisitors) * 100
			c.engagementMetrics.BounceRate = bounceRate
			
			metric := models.NewMetric("business_engagement_bounce_rate", models.MetricTypeGauge, models.CategoryBusiness).
				WithLabels(c.labels).
				WithValue(bounceRate).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "percent"
			metric.Description = "Bounce rate"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if uniqueVisitors > 0 {
		pagesPerSession := float64(pageViews) / float64(uniqueVisitors)
		c.engagementMetrics.PagesPerSession = pagesPerSession
		
		metric := models.NewMetric("business_engagement_pages_per_session", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(pagesPerSession).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "pages"
		metric.Description = "Pages per session"
		metrics = append(metrics, *metric)
	}
	
	c.engagementMetrics.LastUpdated = timestamp
	return metrics, nil
}

// collectPerformanceMetrics 
func (c *BusinessCollector) collectPerformanceMetrics(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// ?0
	var revenue sql.NullFloat64
	err := c.db.QueryRowContext(ctx, 
		"SELECT SUM(amount) FROM payments WHERE status = 'success' AND created_at > NOW() - INTERVAL '30 days'").Scan(&revenue)
	if err == nil && revenue.Valid {
		c.performanceMetrics.Revenue = revenue.Float64
		
		metric := models.NewMetric("business_performance_revenue", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(revenue.Float64).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "currency"
		metric.Description = "Revenue in last 30 days"
		metrics = append(metrics, *metric)
	}
	
	// MRR?
	var mrr sql.NullFloat64
	err = c.db.QueryRowContext(ctx, 
		"SELECT SUM(amount) FROM subscriptions WHERE status = 'active' AND billing_cycle = 'monthly'").Scan(&mrr)
	if err == nil && mrr.Valid {
		c.performanceMetrics.MonthlyRecurringRevenue = mrr.Float64
		
		metric := models.NewMetric("business_performance_mrr", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(mrr.Float64).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "currency"
		metric.Description = "Monthly recurring revenue"
		metrics = append(metrics, *metric)
	}
	
	// 
	if c.userMetrics.ActiveUsers > 0 && revenue.Valid {
		clv := revenue.Float64 / float64(c.userMetrics.ActiveUsers)
		c.performanceMetrics.CustomerLifetimeValue = clv
		
		metric := models.NewMetric("business_performance_customer_lifetime_value", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(clv).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "currency"
		metric.Description = "Customer lifetime value"
		metrics = append(metrics, *metric)
	}
	
	// 
	var lastMonthRevenue sql.NullFloat64
	err = c.db.QueryRowContext(ctx, 
		"SELECT SUM(amount) FROM payments WHERE status = 'success' AND created_at BETWEEN NOW() - INTERVAL '60 days' AND NOW() - INTERVAL '30 days'").Scan(&lastMonthRevenue)
	if err == nil && revenue.Valid && lastMonthRevenue.Valid && lastMonthRevenue.Float64 > 0 {
		growthRate := (revenue.Float64 - lastMonthRevenue.Float64) / lastMonthRevenue.Float64 * 100
		c.performanceMetrics.GrowthRate = growthRate
		
		metric := models.NewMetric("business_performance_growth_rate", models.MetricTypeGauge, models.CategoryBusiness).
			WithLabels(c.labels).
			WithValue(growthRate).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Month-over-month growth rate"
		metrics = append(metrics, *metric)
	}
	
	c.performanceMetrics.LastUpdated = timestamp
	return metrics, nil
}

// RecordUserAction 
func (c *BusinessCollector) RecordUserAction(userID string, action string, metadata map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// ?
	// 
}

// RecordBusinessEvent 
func (c *BusinessCollector) RecordBusinessEvent(eventType string, value float64, metadata map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// ?
	// 
}

// GetUserMetrics 
func (c *BusinessCollector) GetUserMetrics() *UserMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.userMetrics
}

// GetOrderMetrics 
func (c *BusinessCollector) GetOrderMetrics() *OrderMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.orderMetrics
}

// GetPaymentMetrics 
func (c *BusinessCollector) GetPaymentMetrics() *PaymentMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.paymentMetrics
}

// GetContentMetrics 
func (c *BusinessCollector) GetContentMetrics() *ContentMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.contentMetrics
}

// GetEngagementMetrics ?
func (c *BusinessCollector) GetEngagementMetrics() *EngagementMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.engagementMetrics
}

// GetPerformanceMetrics 
func (c *BusinessCollector) GetPerformanceMetrics() *PerformanceMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.performanceMetrics
}

// ?
var _ interfaces.MetricCollector = (*BusinessCollector)(nil)

