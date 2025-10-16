﻿package collectors

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// RedisCollector Redis?
type RedisCollector struct {
	name     string
	interval time.Duration
	enabled  bool
	labels   map[string]string
	
	// Redis?
	client redis.UniversalClient
	
	// 
	collectInfo       bool
	collectMemory     bool
	collectStats      bool
	collectReplication bool
	collectCluster    bool
	collectKeyspace   bool
	collectSlowlog    bool
	
	// Redis
	serverInfo    *RedisServerInfo
	memoryInfo    *RedisMemoryInfo
	statsInfo     *RedisStatsInfo
	replicationInfo *RedisReplicationInfo
	clusterInfo   *RedisClusterInfo
	keyspaceInfo  *RedisKeyspaceInfo
	slowlogInfo   *RedisSlowlogInfo
	
	// ?
	mutex sync.RWMutex
	
	// ?
	lastCollectTime time.Time
}

// RedisCollectorConfig Redis?
type RedisCollectorConfig struct {
	Interval           time.Duration     `yaml:"interval"`
	Enabled            bool              `yaml:"enabled"`
	Labels             map[string]string `yaml:"labels"`
	CollectInfo        bool              `yaml:"collect_info"`
	CollectMemory      bool              `yaml:"collect_memory"`
	CollectStats       bool              `yaml:"collect_stats"`
	CollectReplication bool              `yaml:"collect_replication"`
	CollectCluster     bool              `yaml:"collect_cluster"`
	CollectKeyspace    bool              `yaml:"collect_keyspace"`
	CollectSlowlog     bool              `yaml:"collect_slowlog"`
}

// RedisServerInfo Redis?
type RedisServerInfo struct {
	Version         string    `json:"version"`
	Mode            string    `json:"mode"`
	Role            string    `json:"role"`
	UptimeInSeconds uint64    `json:"uptime_in_seconds"`
	UptimeInDays    uint64    `json:"uptime_in_days"`
	ConnectedClients uint64   `json:"connected_clients"`
	BlockedClients  uint64    `json:"blocked_clients"`
	MaxClients      uint64    `json:"maxclients"`
	LastUpdated     time.Time `json:"last_updated"`
}

// RedisMemoryInfo Redis
type RedisMemoryInfo struct {
	UsedMemory         uint64    `json:"used_memory"`
	UsedMemoryHuman    string    `json:"used_memory_human"`
	UsedMemoryRss      uint64    `json:"used_memory_rss"`
	UsedMemoryPeak     uint64    `json:"used_memory_peak"`
	UsedMemoryPeakHuman string   `json:"used_memory_peak_human"`
	MaxMemory          uint64    `json:"maxmemory"`
	MaxMemoryHuman     string    `json:"maxmemory_human"`
	MemoryFragmentationRatio float64 `json:"mem_fragmentation_ratio"`
	UsedMemoryDataset  uint64    `json:"used_memory_dataset"`
	TotalSystemMemory  uint64    `json:"total_system_memory"`
	LastUpdated        time.Time `json:"last_updated"`
}

// RedisStatsInfo Redis
type RedisStatsInfo struct {
	TotalConnectionsReceived uint64    `json:"total_connections_received"`
	TotalCommandsProcessed   uint64    `json:"total_commands_processed"`
	InstantaneousOpsPerSec   uint64    `json:"instantaneous_ops_per_sec"`
	TotalNetInputBytes       uint64    `json:"total_net_input_bytes"`
	TotalNetOutputBytes      uint64    `json:"total_net_output_bytes"`
	InstantaneousInputKbps   float64   `json:"instantaneous_input_kbps"`
	InstantaneousOutputKbps  float64   `json:"instantaneous_output_kbps"`
	RejectedConnections      uint64    `json:"rejected_connections"`
	SyncFull                 uint64    `json:"sync_full"`
	SyncPartialOk            uint64    `json:"sync_partial_ok"`
	SyncPartialErr           uint64    `json:"sync_partial_err"`
	ExpiredKeys              uint64    `json:"expired_keys"`
	EvictedKeys              uint64    `json:"evicted_keys"`
	KeyspaceHits             uint64    `json:"keyspace_hits"`
	KeyspaceMisses           uint64    `json:"keyspace_misses"`
	PubsubChannels           uint64    `json:"pubsub_channels"`
	PubsubPatterns           uint64    `json:"pubsub_patterns"`
	LastUpdated              time.Time `json:"last_updated"`
}

// RedisReplicationInfo Redis
type RedisReplicationInfo struct {
	Role                string              `json:"role"`
	ConnectedSlaves     uint64              `json:"connected_slaves"`
	MasterReplOffset    uint64              `json:"master_repl_offset"`
	ReplBacklogActive   bool                `json:"repl_backlog_active"`
	ReplBacklogSize     uint64              `json:"repl_backlog_size"`
	ReplBacklogFirstByteOffset uint64       `json:"repl_backlog_first_byte_offset"`
	ReplBacklogHistlen  uint64              `json:"repl_backlog_histlen"`
	SlaveInfo           map[string]SlaveInfo `json:"slave_info"`
	LastUpdated         time.Time           `json:"last_updated"`
}

// SlaveInfo ?
type SlaveInfo struct {
	IP     string `json:"ip"`
	Port   string `json:"port"`
	State  string `json:"state"`
	Offset uint64 `json:"offset"`
	Lag    uint64 `json:"lag"`
}

// RedisClusterInfo Redis
type RedisClusterInfo struct {
	ClusterEnabled      bool      `json:"cluster_enabled"`
	ClusterState        string    `json:"cluster_state"`
	ClusterSlotsAssigned uint64   `json:"cluster_slots_assigned"`
	ClusterSlotsOk      uint64    `json:"cluster_slots_ok"`
	ClusterSlotsPfail   uint64    `json:"cluster_slots_pfail"`
	ClusterSlotsFail    uint64    `json:"cluster_slots_fail"`
	ClusterKnownNodes   uint64    `json:"cluster_known_nodes"`
	ClusterSize         uint64    `json:"cluster_size"`
	LastUpdated         time.Time `json:"last_updated"`
}

// RedisKeyspaceInfo Redis?
type RedisKeyspaceInfo struct {
	Databases   map[string]DatabaseInfo `json:"databases"`
	LastUpdated time.Time               `json:"last_updated"`
}

// DatabaseInfo ?
type DatabaseInfo struct {
	Keys    uint64 `json:"keys"`
	Expires uint64 `json:"expires"`
	AvgTTL  uint64 `json:"avg_ttl"`
}

// RedisSlowlogInfo Redis?
type RedisSlowlogInfo struct {
	SlowlogLen    uint64        `json:"slowlog_len"`
	SlowlogEntries []SlowlogEntry `json:"slowlog_entries"`
	LastUpdated   time.Time     `json:"last_updated"`
}

// SlowlogEntry ?
type SlowlogEntry struct {
	ID        uint64        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Command   []string      `json:"command"`
	ClientIP  string        `json:"client_ip"`
	ClientName string       `json:"client_name"`
}

// NewRedisCollector Redis?
func NewRedisCollector(config RedisCollectorConfig, client redis.UniversalClient) *RedisCollector {
	labels := map[string]string{
		"collector": "redis",
		"service":   "core-services",
	}
	
	// ?
	for k, v := range config.Labels {
		labels[k] = v
	}
	
	return &RedisCollector{
		name:               "redis",
		interval:           config.Interval,
		enabled:            config.Enabled,
		labels:             labels,
		client:             client,
		collectInfo:        config.CollectInfo,
		collectMemory:      config.CollectMemory,
		collectStats:       config.CollectStats,
		collectReplication: config.CollectReplication,
		collectCluster:     config.CollectCluster,
		collectKeyspace:    config.CollectKeyspace,
		collectSlowlog:     config.CollectSlowlog,
		serverInfo:         &RedisServerInfo{},
		memoryInfo:         &RedisMemoryInfo{},
		statsInfo:          &RedisStatsInfo{},
		replicationInfo:    &RedisReplicationInfo{SlaveInfo: make(map[string]SlaveInfo)},
		clusterInfo:        &RedisClusterInfo{},
		keyspaceInfo:       &RedisKeyspaceInfo{Databases: make(map[string]DatabaseInfo)},
		slowlogInfo:        &RedisSlowlogInfo{},
		lastCollectTime:    time.Now(),
	}
}

// GetName ?
func (c *RedisCollector) GetName() string {
	return c.name
}

// GetCategory ?
func (c *RedisCollector) GetCategory() models.MetricCategory {
	return models.CategoryDatabase
}

// GetInterval 
func (c *RedisCollector) GetInterval() time.Duration {
	return c.interval
}

// IsEnabled ?
func (c *RedisCollector) IsEnabled() bool {
	return c.enabled
}

// Start ?
func (c *RedisCollector) Start(ctx context.Context) error {
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
				fmt.Printf("Redis collector error: %v\n", err)
			}
		}
	}
}

// Stop ?
func (c *RedisCollector) Stop() error {
	c.enabled = false
	return nil
}

// Health ?
func (c *RedisCollector) Health() error {
	if !c.enabled {
		return fmt.Errorf("redis collector is disabled")
	}
	
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	
	// Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	
	return nil
}

// Collect 
func (c *RedisCollector) Collect(ctx context.Context) ([]models.Metric, error) {
	if !c.enabled || c.client == nil {
		return nil, nil
	}
	
	var metrics []models.Metric
	now := time.Now()
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Redis INFO
	infoResult := c.client.Info(ctx)
	if infoResult.Err() != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", infoResult.Err())
	}
	
	infoData := infoResult.Val()
	infoMap := c.parseInfo(infoData)
	
	// ?
	if c.collectInfo {
		serverMetrics, err := c.collectServerInfo(infoMap, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect server info: %w", err)
		}
		metrics = append(metrics, serverMetrics...)
	}
	
	// 
	if c.collectMemory {
		memoryMetrics, err := c.collectMemoryInfo(infoMap, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect memory info: %w", err)
		}
		metrics = append(metrics, memoryMetrics...)
	}
	
	// 
	if c.collectStats {
		statsMetrics, err := c.collectStatsInfo(infoMap, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect stats info: %w", err)
		}
		metrics = append(metrics, statsMetrics...)
	}
	
	// 
	if c.collectReplication {
		replicationMetrics, err := c.collectReplicationInfo(infoMap, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect replication info: %w", err)
		}
		metrics = append(metrics, replicationMetrics...)
	}
	
	// 
	if c.collectCluster {
		clusterMetrics, err := c.collectClusterInfo(ctx, infoMap, now)
		if err != nil {
			// ?
			fmt.Printf("Failed to collect cluster info: %v\n", err)
		} else {
			metrics = append(metrics, clusterMetrics...)
		}
	}
	
	// ?
	if c.collectKeyspace {
		keyspaceMetrics, err := c.collectKeyspaceInfo(infoMap, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect keyspace info: %w", err)
		}
		metrics = append(metrics, keyspaceMetrics...)
	}
	
	// ?
	if c.collectSlowlog {
		slowlogMetrics, err := c.collectSlowlogInfo(ctx, now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect slowlog info: %w", err)
		}
		metrics = append(metrics, slowlogMetrics...)
	}
	
	c.lastCollectTime = now
	return metrics, nil
}

// parseInfo Redis INFO
func (c *RedisCollector) parseInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	
	return result
}

// collectServerInfo ?
func (c *RedisCollector) collectServerInfo(infoMap map[string]string, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// Redis汾
	if version, ok := infoMap["redis_version"]; ok {
		c.serverInfo.Version = version
	}
	
	// Redis
	if mode, ok := infoMap["redis_mode"]; ok {
		c.serverInfo.Mode = mode
	}
	
	// 
	if role, ok := infoMap["role"]; ok {
		c.serverInfo.Role = role
	}
	
	// ?
	if uptimeStr, ok := infoMap["uptime_in_seconds"]; ok {
		if uptime, err := strconv.ParseUint(uptimeStr, 10, 64); err == nil {
			c.serverInfo.UptimeInSeconds = uptime
			
			metric := models.NewMetric("redis_uptime_seconds", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(uptime)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "seconds"
			metric.Description = "Redis server uptime in seconds"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if clientsStr, ok := infoMap["connected_clients"]; ok {
		if clients, err := strconv.ParseUint(clientsStr, 10, 64); err == nil {
			c.serverInfo.ConnectedClients = clients
			
			metric := models.NewMetric("redis_connected_clients", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(clients)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "clients"
			metric.Description = "Number of connected clients"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if blockedStr, ok := infoMap["blocked_clients"]; ok {
		if blocked, err := strconv.ParseUint(blockedStr, 10, 64); err == nil {
			c.serverInfo.BlockedClients = blocked
			
			metric := models.NewMetric("redis_blocked_clients", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(blocked)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "clients"
			metric.Description = "Number of blocked clients"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if maxClientsStr, ok := infoMap["maxclients"]; ok {
		if maxClients, err := strconv.ParseUint(maxClientsStr, 10, 64); err == nil {
			c.serverInfo.MaxClients = maxClients
			
			metric := models.NewMetric("redis_max_clients", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(maxClients)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "clients"
			metric.Description = "Maximum number of clients"
			metrics = append(metrics, *metric)
		}
	}
	
	c.serverInfo.LastUpdated = timestamp
	return metrics, nil
}

// collectMemoryInfo 
func (c *RedisCollector) collectMemoryInfo(infoMap map[string]string, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// ?
	if usedMemoryStr, ok := infoMap["used_memory"]; ok {
		if usedMemory, err := strconv.ParseUint(usedMemoryStr, 10, 64); err == nil {
			c.memoryInfo.UsedMemory = usedMemory
			
			metric := models.NewMetric("redis_memory_used_bytes", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(usedMemory)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "Used memory in bytes"
			metrics = append(metrics, *metric)
		}
	}
	
	// RSS
	if usedMemoryRssStr, ok := infoMap["used_memory_rss"]; ok {
		if usedMemoryRss, err := strconv.ParseUint(usedMemoryRssStr, 10, 64); err == nil {
			c.memoryInfo.UsedMemoryRss = usedMemoryRss
			
			metric := models.NewMetric("redis_memory_rss_bytes", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(usedMemoryRss)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "RSS memory in bytes"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if usedMemoryPeakStr, ok := infoMap["used_memory_peak"]; ok {
		if usedMemoryPeak, err := strconv.ParseUint(usedMemoryPeakStr, 10, 64); err == nil {
			c.memoryInfo.UsedMemoryPeak = usedMemoryPeak
			
			metric := models.NewMetric("redis_memory_peak_bytes", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(usedMemoryPeak)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "Peak memory usage in bytes"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if maxMemoryStr, ok := infoMap["maxmemory"]; ok {
		if maxMemory, err := strconv.ParseUint(maxMemoryStr, 10, 64); err == nil {
			c.memoryInfo.MaxMemory = maxMemory
			
			metric := models.NewMetric("redis_memory_max_bytes", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(maxMemory)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "Maximum memory limit in bytes"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if fragRatioStr, ok := infoMap["mem_fragmentation_ratio"]; ok {
		if fragRatio, err := strconv.ParseFloat(fragRatioStr, 64); err == nil {
			c.memoryInfo.MemoryFragmentationRatio = fragRatio
			
			metric := models.NewMetric("redis_memory_fragmentation_ratio", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(fragRatio).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "ratio"
			metric.Description = "Memory fragmentation ratio"
			metrics = append(metrics, *metric)
		}
	}
	
	c.memoryInfo.LastUpdated = timestamp
	return metrics, nil
}

// collectStatsInfo 
func (c *RedisCollector) collectStatsInfo(infoMap map[string]string, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 
	if totalConnectionsStr, ok := infoMap["total_connections_received"]; ok {
		if totalConnections, err := strconv.ParseUint(totalConnectionsStr, 10, 64); err == nil {
			c.statsInfo.TotalConnectionsReceived = totalConnections
			
			metric := models.NewMetric("redis_connections_received_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(totalConnections)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "connections"
			metric.Description = "Total connections received"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if totalCommandsStr, ok := infoMap["total_commands_processed"]; ok {
		if totalCommands, err := strconv.ParseUint(totalCommandsStr, 10, 64); err == nil {
			c.statsInfo.TotalCommandsProcessed = totalCommands
			
			metric := models.NewMetric("redis_commands_processed_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(totalCommands)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "commands"
			metric.Description = "Total commands processed"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if opsPerSecStr, ok := infoMap["instantaneous_ops_per_sec"]; ok {
		if opsPerSec, err := strconv.ParseUint(opsPerSecStr, 10, 64); err == nil {
			c.statsInfo.InstantaneousOpsPerSec = opsPerSec
			
			metric := models.NewMetric("redis_ops_per_sec", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(opsPerSec)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "ops/sec"
			metric.Description = "Operations per second"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if netInputStr, ok := infoMap["total_net_input_bytes"]; ok {
		if netInput, err := strconv.ParseUint(netInputStr, 10, 64); err == nil {
			c.statsInfo.TotalNetInputBytes = netInput
			
			metric := models.NewMetric("redis_net_input_bytes_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(netInput)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "Total network input bytes"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if netOutputStr, ok := infoMap["total_net_output_bytes"]; ok {
		if netOutput, err := strconv.ParseUint(netOutputStr, 10, 64); err == nil {
			c.statsInfo.TotalNetOutputBytes = netOutput
			
			metric := models.NewMetric("redis_net_output_bytes_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(netOutput)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "Total network output bytes"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if rejectedStr, ok := infoMap["rejected_connections"]; ok {
		if rejected, err := strconv.ParseUint(rejectedStr, 10, 64); err == nil {
			c.statsInfo.RejectedConnections = rejected
			
			metric := models.NewMetric("redis_connections_rejected_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(rejected)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "connections"
			metric.Description = "Total rejected connections"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if expiredStr, ok := infoMap["expired_keys"]; ok {
		if expired, err := strconv.ParseUint(expiredStr, 10, 64); err == nil {
			c.statsInfo.ExpiredKeys = expired
			
			metric := models.NewMetric("redis_keys_expired_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(expired)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "keys"
			metric.Description = "Total expired keys"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if evictedStr, ok := infoMap["evicted_keys"]; ok {
		if evicted, err := strconv.ParseUint(evictedStr, 10, 64); err == nil {
			c.statsInfo.EvictedKeys = evicted
			
			metric := models.NewMetric("redis_keys_evicted_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(evicted)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "keys"
			metric.Description = "Total evicted keys"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if hitsStr, ok := infoMap["keyspace_hits"]; ok {
		if hits, err := strconv.ParseUint(hitsStr, 10, 64); err == nil {
			c.statsInfo.KeyspaceHits = hits
			
			metric := models.NewMetric("redis_keyspace_hits_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(hits)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "hits"
			metric.Description = "Total keyspace hits"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if missesStr, ok := infoMap["keyspace_misses"]; ok {
		if misses, err := strconv.ParseUint(missesStr, 10, 64); err == nil {
			c.statsInfo.KeyspaceMisses = misses
			
			metric := models.NewMetric("redis_keyspace_misses_total", models.MetricTypeCounter, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(misses)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "misses"
			metric.Description = "Total keyspace misses"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if c.statsInfo.KeyspaceHits > 0 || c.statsInfo.KeyspaceMisses > 0 {
		total := c.statsInfo.KeyspaceHits + c.statsInfo.KeyspaceMisses
		hitRate := float64(c.statsInfo.KeyspaceHits) / float64(total) * 100
		
		metric := models.NewMetric("redis_keyspace_hit_rate", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(c.labels).
			WithValue(hitRate).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Keyspace hit rate"
		metrics = append(metrics, *metric)
	}
	
	c.statsInfo.LastUpdated = timestamp
	return metrics, nil
}

// collectReplicationInfo 
func (c *RedisCollector) collectReplicationInfo(infoMap map[string]string, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 
	if role, ok := infoMap["role"]; ok {
		c.replicationInfo.Role = role
	}
	
	// ?
	if connectedSlavesStr, ok := infoMap["connected_slaves"]; ok {
		if connectedSlaves, err := strconv.ParseUint(connectedSlavesStr, 10, 64); err == nil {
			c.replicationInfo.ConnectedSlaves = connectedSlaves
			
			metric := models.NewMetric("redis_connected_slaves", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(connectedSlaves)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "slaves"
			metric.Description = "Number of connected slaves"
			metrics = append(metrics, *metric)
		}
	}
	
	// 㸴
	if masterReplOffsetStr, ok := infoMap["master_repl_offset"]; ok {
		if masterReplOffset, err := strconv.ParseUint(masterReplOffsetStr, 10, 64); err == nil {
			c.replicationInfo.MasterReplOffset = masterReplOffset
			
			metric := models.NewMetric("redis_master_repl_offset", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(masterReplOffset)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "offset"
			metric.Description = "Master replication offset"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if replBacklogSizeStr, ok := infoMap["repl_backlog_size"]; ok {
		if replBacklogSize, err := strconv.ParseUint(replBacklogSizeStr, 10, 64); err == nil {
			c.replicationInfo.ReplBacklogSize = replBacklogSize
			
			metric := models.NewMetric("redis_repl_backlog_size", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(replBacklogSize)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes"
			metric.Description = "Replication backlog size"
			metrics = append(metrics, *metric)
		}
	}
	
	c.replicationInfo.LastUpdated = timestamp
	return metrics, nil
}

// collectClusterInfo 
func (c *RedisCollector) collectClusterInfo(ctx context.Context, infoMap map[string]string, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// ?
	clusterEnabledStr, ok := infoMap["cluster_enabled"]
	if !ok {
		return metrics, nil
	}
	
	clusterEnabled := clusterEnabledStr == "1"
	c.clusterInfo.ClusterEnabled = clusterEnabled
	
	if !clusterEnabled {
		return metrics, nil
	}
	
	// 
	clusterInfoResult := c.client.ClusterInfo(ctx)
	if clusterInfoResult.Err() != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", clusterInfoResult.Err())
	}
	
	clusterInfoData := clusterInfoResult.Val()
	clusterInfoMap := c.parseInfo(clusterInfoData)
	
	// ?
	if clusterState, ok := clusterInfoMap["cluster_state"]; ok {
		c.clusterInfo.ClusterState = clusterState
		
		// ok=1, fail=0?
		stateValue := 0.0
		if clusterState == "ok" {
			stateValue = 1.0
		}
		
		metric := models.NewMetric("redis_cluster_state", models.MetricTypeGauge, models.CategoryDatabase).
			WithLabels(c.labels).
			WithValue(stateValue).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "state"
		metric.Description = "Cluster state (1=ok, 0=fail)"
		metrics = append(metrics, *metric)
	}
	
	// 
	if slotsAssignedStr, ok := clusterInfoMap["cluster_slots_assigned"]; ok {
		if slotsAssigned, err := strconv.ParseUint(slotsAssignedStr, 10, 64); err == nil {
			c.clusterInfo.ClusterSlotsAssigned = slotsAssigned
			
			metric := models.NewMetric("redis_cluster_slots_assigned", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(slotsAssigned)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "slots"
			metric.Description = "Number of assigned cluster slots"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if slotsOkStr, ok := clusterInfoMap["cluster_slots_ok"]; ok {
		if slotsOk, err := strconv.ParseUint(slotsOkStr, 10, 64); err == nil {
			c.clusterInfo.ClusterSlotsOk = slotsOk
			
			metric := models.NewMetric("redis_cluster_slots_ok", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(slotsOk)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "slots"
			metric.Description = "Number of ok cluster slots"
			metrics = append(metrics, *metric)
		}
	}
	
	// ?
	if knownNodesStr, ok := clusterInfoMap["cluster_known_nodes"]; ok {
		if knownNodes, err := strconv.ParseUint(knownNodesStr, 10, 64); err == nil {
			c.clusterInfo.ClusterKnownNodes = knownNodes
			
			metric := models.NewMetric("redis_cluster_known_nodes", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(knownNodes)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "nodes"
			metric.Description = "Number of known cluster nodes"
			metrics = append(metrics, *metric)
		}
	}
	
	// 
	if clusterSizeStr, ok := clusterInfoMap["cluster_size"]; ok {
		if clusterSize, err := strconv.ParseUint(clusterSizeStr, 10, 64); err == nil {
			c.clusterInfo.ClusterSize = clusterSize
			
			metric := models.NewMetric("redis_cluster_size", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(c.labels).
				WithValue(float64(clusterSize)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "nodes"
			metric.Description = "Cluster size"
			metrics = append(metrics, *metric)
		}
	}
	
	c.clusterInfo.LastUpdated = timestamp
	return metrics, nil
}

// collectKeyspaceInfo ?
func (c *RedisCollector) collectKeyspaceInfo(infoMap map[string]string, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// ?
	for key, value := range infoMap {
		if strings.HasPrefix(key, "db") {
			dbName := key
			
			// ?
			// : keys=123,expires=456,avg_ttl=789
			parts := strings.Split(value, ",")
			dbInfo := DatabaseInfo{}
			
			for _, part := range parts {
				kv := strings.Split(part, "=")
				if len(kv) != 2 {
					continue
				}
				
				switch kv[0] {
				case "keys":
					if keys, err := strconv.ParseUint(kv[1], 10, 64); err == nil {
						dbInfo.Keys = keys
					}
				case "expires":
					if expires, err := strconv.ParseUint(kv[1], 10, 64); err == nil {
						dbInfo.Expires = expires
					}
				case "avg_ttl":
					if avgTTL, err := strconv.ParseUint(kv[1], 10, 64); err == nil {
						dbInfo.AvgTTL = avgTTL
					}
				}
			}
			
			c.keyspaceInfo.Databases[dbName] = dbInfo
			
			// 
			labels := make(map[string]string)
			for k, v := range c.labels {
				labels[k] = v
			}
			labels["database"] = dbName
			
			// ?
			metric := models.NewMetric("redis_db_keys", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(labels).
				WithValue(float64(dbInfo.Keys)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "keys"
			metric.Description = "Number of keys in database"
			metrics = append(metrics, *metric)
			
			// ?
			metric = models.NewMetric("redis_db_expires", models.MetricTypeGauge, models.CategoryDatabase).
				WithLabels(labels).
				WithValue(float64(dbInfo.Expires)).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "keys"
			metric.Description = "Number of keys with expiration in database"
			metrics = append(metrics, *metric)
			
			// TTL
			if dbInfo.AvgTTL > 0 {
				metric = models.NewMetric("redis_db_avg_ttl_seconds", models.MetricTypeGauge, models.CategoryDatabase).
					WithLabels(labels).
					WithValue(float64(dbInfo.AvgTTL) / 1000). // 
					WithSource(c.name)
				metric.Timestamp = timestamp
				metric.Unit = "seconds"
				metric.Description = "Average TTL of keys in database"
				metrics = append(metrics, *metric)
			}
		}
	}
	
	c.keyspaceInfo.LastUpdated = timestamp
	return metrics, nil
}

// collectSlowlogInfo ?
func (c *RedisCollector) collectSlowlogInfo(ctx context.Context, timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// ?
	slowlogLenResult := c.client.SlowLogLen(ctx)
	if slowlogLenResult.Err() != nil {
		return nil, fmt.Errorf("failed to get slowlog length: %w", slowlogLenResult.Err())
	}
	
	slowlogLen := uint64(slowlogLenResult.Val())
	c.slowlogInfo.SlowlogLen = slowlogLen
	
	metric := models.NewMetric("redis_slowlog_length", models.MetricTypeGauge, models.CategoryDatabase).
		WithLabels(c.labels).
		WithValue(float64(slowlogLen)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "entries"
	metric.Description = "Number of entries in slowlog"
	metrics = append(metrics, *metric)
	
	// ?0
	slowlogResult := c.client.SlowLogGet(ctx, 10)
	if slowlogResult.Err() != nil {
		return nil, fmt.Errorf("failed to get slowlog entries: %w", slowlogResult.Err())
	}
	
	slowlogEntries := slowlogResult.Val()
	c.slowlogInfo.SlowlogEntries = make([]SlowlogEntry, 0, len(slowlogEntries))
	
	for _, entry := range slowlogEntries {
		slowlogEntry := SlowlogEntry{
			ID:        uint64(entry.ID),
			Timestamp: entry.Time,
			Duration:  entry.Duration,
			Command:   entry.Args,
		}
		c.slowlogInfo.SlowlogEntries = append(c.slowlogInfo.SlowlogEntries, slowlogEntry)
	}
	
	c.slowlogInfo.LastUpdated = timestamp
	return metrics, nil
}

// GetServerInfo ?
func (c *RedisCollector) GetServerInfo() *RedisServerInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.serverInfo
}

// GetMemoryInfo 
func (c *RedisCollector) GetMemoryInfo() *RedisMemoryInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.memoryInfo
}

// GetStatsInfo 
func (c *RedisCollector) GetStatsInfo() *RedisStatsInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.statsInfo
}

// GetReplicationInfo 
func (c *RedisCollector) GetReplicationInfo() *RedisReplicationInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.replicationInfo
}

// GetClusterInfo 
func (c *RedisCollector) GetClusterInfo() *RedisClusterInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.clusterInfo
}

// GetKeyspaceInfo ?
func (c *RedisCollector) GetKeyspaceInfo() *RedisKeyspaceInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.keyspaceInfo
}

// GetSlowlogInfo ?
func (c *RedisCollector) GetSlowlogInfo() *RedisSlowlogInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.slowlogInfo
}

// ?
var _ interfaces.MetricCollector = (*RedisCollector)(nil)

