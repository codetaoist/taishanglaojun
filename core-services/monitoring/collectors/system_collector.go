package collectors

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/taishanglaojun/core-services/monitoring/models"
)

// SystemCollector 系统指标收集器
type SystemCollector struct {
	name     string
	interval time.Duration
	enabled  bool
	labels   map[string]string
	
	// 配置选项
	collectCPU     bool
	collectMemory  bool
	collectDisk    bool
	collectNetwork bool
	collectLoad    bool
	collectProcess bool
	
	// 缓存上次的网络统计
	lastNetStats map[string]net.IOCountersStat
	lastTime     time.Time
}

// NewSystemCollector 创建系统指标收集器
func NewSystemCollector(config SystemCollectorConfig) *SystemCollector {
	hostname, _ := host.Info()
	labels := map[string]string{
		"collector": "system",
		"hostname":  hostname.Hostname,
		"os":        hostname.OS,
		"platform":  hostname.Platform,
	}
	
	// 添加自定义标签
	for k, v := range config.Labels {
		labels[k] = v
	}
	
	return &SystemCollector{
		name:           "system",
		interval:       config.Interval,
		enabled:        config.Enabled,
		labels:         labels,
		collectCPU:     config.CollectCPU,
		collectMemory:  config.CollectMemory,
		collectDisk:    config.CollectDisk,
		collectNetwork: config.CollectNetwork,
		collectLoad:    config.CollectLoad,
		collectProcess: config.CollectProcess,
		lastNetStats:   make(map[string]net.IOCountersStat),
		lastTime:       time.Now(),
	}
}

// SystemCollectorConfig 系统收集器配置
type SystemCollectorConfig struct {
	Interval       time.Duration     `yaml:"interval"`
	Enabled        bool              `yaml:"enabled"`
	Labels         map[string]string `yaml:"labels"`
	CollectCPU     bool              `yaml:"collect_cpu"`
	CollectMemory  bool              `yaml:"collect_memory"`
	CollectDisk    bool              `yaml:"collect_disk"`
	CollectNetwork bool              `yaml:"collect_network"`
	CollectLoad    bool              `yaml:"collect_load"`
	CollectProcess bool              `yaml:"collect_process"`
}

// GetName 获取收集器名称
func (c *SystemCollector) GetName() string {
	return c.name
}

// GetCategory 获取收集器分类
func (c *SystemCollector) GetCategory() models.MetricCategory {
	return models.CategorySystem
}

// GetInterval 获取收集间隔
func (c *SystemCollector) GetInterval() time.Duration {
	return c.interval
}

// IsEnabled 检查是否启用
func (c *SystemCollector) IsEnabled() bool {
	return c.enabled
}

// Start 启动收集器
func (c *SystemCollector) Start(ctx context.Context) error {
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
				// 记录错误但继续运行
				fmt.Printf("System collector error: %v\n", err)
			}
		}
	}
}

// Stop 停止收集器
func (c *SystemCollector) Stop() error {
	c.enabled = false
	return nil
}

// Health 健康检查
func (c *SystemCollector) Health() error {
	if !c.enabled {
		return fmt.Errorf("system collector is disabled")
	}
	
	// 尝试收集一个简单的指标来验证健康状态
	_, err := cpu.Percent(0, false)
	if err != nil {
		return fmt.Errorf("failed to collect CPU metrics: %w", err)
	}
	
	return nil
}

// Collect 收集指标
func (c *SystemCollector) Collect(ctx context.Context) ([]models.Metric, error) {
	if !c.enabled {
		return nil, nil
	}
	
	var metrics []models.Metric
	now := time.Now()
	
	// 收集CPU指标
	if c.collectCPU {
		cpuMetrics, err := c.collectCPUMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
		}
		metrics = append(metrics, cpuMetrics...)
	}
	
	// 收集内存指标
	if c.collectMemory {
		memMetrics, err := c.collectMemoryMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
		}
		metrics = append(metrics, memMetrics...)
	}
	
	// 收集磁盘指标
	if c.collectDisk {
		diskMetrics, err := c.collectDiskMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
		}
		metrics = append(metrics, diskMetrics...)
	}
	
	// 收集网络指标
	if c.collectNetwork {
		netMetrics, err := c.collectNetworkMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect network metrics: %w", err)
		}
		metrics = append(metrics, netMetrics...)
	}
	
	// 收集负载指标
	if c.collectLoad {
		loadMetrics, err := c.collectLoadMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect load metrics: %w", err)
		}
		metrics = append(metrics, loadMetrics...)
	}
	
	// 收集进程指标
	if c.collectProcess {
		processMetrics, err := c.collectProcessMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect process metrics: %w", err)
		}
		metrics = append(metrics, processMetrics...)
	}
	
	return metrics, nil
}

// collectCPUMetrics 收集CPU指标
func (c *SystemCollector) collectCPUMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// CPU使用率
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}
	
	// 总体CPU使用率
	totalPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	
	if len(totalPercent) > 0 {
		metric := models.NewMetric("system_cpu_usage_percent", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(c.labels).
			WithValue(totalPercent[0]).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Total CPU usage percentage"
		metrics = append(metrics, *metric)
	}
	
	// 每个CPU核心的使用率
	for i, percent := range cpuPercents {
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["cpu"] = fmt.Sprintf("cpu%d", i)
		
		metric := models.NewMetric("system_cpu_usage_percent_per_core", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(percent).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "CPU usage percentage per core"
		metrics = append(metrics, *metric)
	}
	
	// CPU时间统计
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return nil, err
	}
	
	if len(cpuTimes) > 0 {
		cpuTime := cpuTimes[0]
		
		// 用户时间
		metric := models.NewMetric("system_cpu_time_user_seconds", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(c.labels).
			WithValue(cpuTime.User).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "CPU time spent in user mode"
		metrics = append(metrics, *metric)
		
		// 系统时间
		metric = models.NewMetric("system_cpu_time_system_seconds", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(c.labels).
			WithValue(cpuTime.System).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "CPU time spent in system mode"
		metrics = append(metrics, *metric)
		
		// 空闲时间
		metric = models.NewMetric("system_cpu_time_idle_seconds", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(c.labels).
			WithValue(cpuTime.Idle).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "CPU time spent in idle mode"
		metrics = append(metrics, *metric)
		
		// IO等待时间
		metric = models.NewMetric("system_cpu_time_iowait_seconds", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(c.labels).
			WithValue(cpuTime.Iowait).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "CPU time spent waiting for I/O"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectMemoryMetrics 收集内存指标
func (c *SystemCollector) collectMemoryMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 虚拟内存
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	
	// 总内存
	metric := models.NewMetric("system_memory_total_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(vmStat.Total)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Total system memory"
	metrics = append(metrics, *metric)
	
	// 可用内存
	metric = models.NewMetric("system_memory_available_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(vmStat.Available)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Available system memory"
	metrics = append(metrics, *metric)
	
	// 已用内存
	metric = models.NewMetric("system_memory_used_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(vmStat.Used)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Used system memory"
	metrics = append(metrics, *metric)
	
	// 内存使用率
	metric = models.NewMetric("system_memory_usage_percent", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(vmStat.UsedPercent).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "percent"
	metric.Description = "Memory usage percentage"
	metrics = append(metrics, *metric)
	
	// 缓存内存
	metric = models.NewMetric("system_memory_cached_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(vmStat.Cached)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Cached memory"
	metrics = append(metrics, *metric)
	
	// 缓冲区内存
	metric = models.NewMetric("system_memory_buffers_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(vmStat.Buffers)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Buffer memory"
	metrics = append(metrics, *metric)
	
	// 交换内存
	swapStat, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	
	// 总交换内存
	metric = models.NewMetric("system_swap_total_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(swapStat.Total)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Total swap memory"
	metrics = append(metrics, *metric)
	
	// 已用交换内存
	metric = models.NewMetric("system_swap_used_bytes", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(swapStat.Used)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Used swap memory"
	metrics = append(metrics, *metric)
	
	// 交换内存使用率
	metric = models.NewMetric("system_swap_usage_percent", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(swapStat.UsedPercent).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "percent"
	metric.Description = "Swap usage percentage"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// collectDiskMetrics 收集磁盘指标
func (c *SystemCollector) collectDiskMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 磁盘分区
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}
		
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["device"] = partition.Device
		labels["mountpoint"] = partition.Mountpoint
		labels["fstype"] = partition.Fstype
		
		// 总空间
		metric := models.NewMetric("system_disk_total_bytes", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(usage.Total)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Total disk space"
		metrics = append(metrics, *metric)
		
		// 已用空间
		metric = models.NewMetric("system_disk_used_bytes", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(usage.Used)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Used disk space"
		metrics = append(metrics, *metric)
		
		// 可用空间
		metric = models.NewMetric("system_disk_free_bytes", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(usage.Free)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Free disk space"
		metrics = append(metrics, *metric)
		
		// 使用率
		metric = models.NewMetric("system_disk_usage_percent", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(usage.UsedPercent).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Disk usage percentage"
		metrics = append(metrics, *metric)
		
		// Inode信息
		metric = models.NewMetric("system_disk_inodes_total", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(usage.InodesTotal)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "count"
		metric.Description = "Total inodes"
		metrics = append(metrics, *metric)
		
		metric = models.NewMetric("system_disk_inodes_used", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(usage.InodesUsed)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "count"
		metric.Description = "Used inodes"
		metrics = append(metrics, *metric)
		
		metric = models.NewMetric("system_disk_inodes_usage_percent", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(usage.InodesUsedPercent).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "percent"
		metric.Description = "Inode usage percentage"
		metrics = append(metrics, *metric)
	}
	
	// 磁盘IO统计
	ioStats, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}
	
	for device, stat := range ioStats {
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["device"] = device
		
		// 读取次数
		metric := models.NewMetric("system_disk_reads_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.ReadCount)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "count"
		metric.Description = "Total disk reads"
		metrics = append(metrics, *metric)
		
		// 写入次数
		metric = models.NewMetric("system_disk_writes_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.WriteCount)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "count"
		metric.Description = "Total disk writes"
		metrics = append(metrics, *metric)
		
		// 读取字节数
		metric = models.NewMetric("system_disk_read_bytes_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.ReadBytes)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Total bytes read from disk"
		metrics = append(metrics, *metric)
		
		// 写入字节数
		metric = models.NewMetric("system_disk_write_bytes_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.WriteBytes)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Total bytes written to disk"
		metrics = append(metrics, *metric)
		
		// 读取时间
		metric = models.NewMetric("system_disk_read_time_seconds_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.ReadTime) / 1000.0).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Total time spent reading from disk"
		metrics = append(metrics, *metric)
		
		// 写入时间
		metric = models.NewMetric("system_disk_write_time_seconds_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.WriteTime) / 1000.0).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Total time spent writing to disk"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectNetworkMetrics 收集网络指标
func (c *SystemCollector) collectNetworkMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 网络接口统计
	netStats, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	
	currentTime := timestamp
	timeDelta := currentTime.Sub(c.lastTime).Seconds()
	
	for _, stat := range netStats {
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["interface"] = stat.Name
		
		// 接收字节数
		metric := models.NewMetric("system_network_receive_bytes_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.BytesRecv)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Total bytes received"
		metrics = append(metrics, *metric)
		
		// 发送字节数
		metric = models.NewMetric("system_network_transmit_bytes_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.BytesSent)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "bytes"
		metric.Description = "Total bytes transmitted"
		metrics = append(metrics, *metric)
		
		// 接收包数
		metric = models.NewMetric("system_network_receive_packets_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.PacketsRecv)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "packets"
		metric.Description = "Total packets received"
		metrics = append(metrics, *metric)
		
		// 发送包数
		metric = models.NewMetric("system_network_transmit_packets_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.PacketsSent)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "packets"
		metric.Description = "Total packets transmitted"
		metrics = append(metrics, *metric)
		
		// 接收错误数
		metric = models.NewMetric("system_network_receive_errors_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.Errin)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "errors"
		metric.Description = "Total receive errors"
		metrics = append(metrics, *metric)
		
		// 发送错误数
		metric = models.NewMetric("system_network_transmit_errors_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.Errout)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "errors"
		metric.Description = "Total transmit errors"
		metrics = append(metrics, *metric)
		
		// 丢包数
		metric = models.NewMetric("system_network_receive_dropped_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.Dropin)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "packets"
		metric.Description = "Total dropped received packets"
		metrics = append(metrics, *metric)
		
		metric = models.NewMetric("system_network_transmit_dropped_total", models.MetricTypeCounter, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(stat.Dropout)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "packets"
		metric.Description = "Total dropped transmitted packets"
		metrics = append(metrics, *metric)
		
		// 计算速率（如果有上次的数据）
		if lastStat, exists := c.lastNetStats[stat.Name]; exists && timeDelta > 0 {
			// 接收速率
			recvRate := float64(stat.BytesRecv-lastStat.BytesRecv) / timeDelta
			metric = models.NewMetric("system_network_receive_bytes_per_second", models.MetricTypeGauge, models.CategorySystem).
				WithLabels(labels).
				WithValue(recvRate).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes_per_second"
			metric.Description = "Network receive rate"
			metrics = append(metrics, *metric)
			
			// 发送速率
			sendRate := float64(stat.BytesSent-lastStat.BytesSent) / timeDelta
			metric = models.NewMetric("system_network_transmit_bytes_per_second", models.MetricTypeGauge, models.CategorySystem).
				WithLabels(labels).
				WithValue(sendRate).
				WithSource(c.name)
			metric.Timestamp = timestamp
			metric.Unit = "bytes_per_second"
			metric.Description = "Network transmit rate"
			metrics = append(metrics, *metric)
		}
		
		// 更新缓存
		c.lastNetStats[stat.Name] = stat
	}
	
	c.lastTime = currentTime
	return metrics, nil
}

// collectLoadMetrics 收集负载指标
func (c *SystemCollector) collectLoadMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 系统负载
	loadStat, err := load.Avg()
	if err != nil {
		return nil, err
	}
	
	// 1分钟负载
	metric := models.NewMetric("system_load_average_1m", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(loadStat.Load1).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "load"
	metric.Description = "1 minute load average"
	metrics = append(metrics, *metric)
	
	// 5分钟负载
	metric = models.NewMetric("system_load_average_5m", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(loadStat.Load5).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "load"
	metric.Description = "5 minute load average"
	metrics = append(metrics, *metric)
	
	// 15分钟负载
	metric = models.NewMetric("system_load_average_15m", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(loadStat.Load15).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "load"
	metric.Description = "15 minute load average"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// collectProcessMetrics 收集进程指标
func (c *SystemCollector) collectProcessMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 进程数量
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	
	// 总进程数
	metric := models.NewMetric("system_processes_total", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(len(processes))).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "count"
	metric.Description = "Total number of processes"
	metrics = append(metrics, *metric)
	
	// 按状态统计进程
	statusCount := make(map[string]int)
	for _, p := range processes {
		status, err := p.Status()
		if err != nil {
			continue
		}
		statusCount[status[0]]++
	}
	
	for status, count := range statusCount {
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		labels["status"] = status
		
		metric := models.NewMetric("system_processes_by_status", models.MetricTypeGauge, models.CategorySystem).
			WithLabels(labels).
			WithValue(float64(count)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "count"
		metric.Description = "Number of processes by status"
		metrics = append(metrics, *metric)
	}
	
	// 文件描述符
	metric = models.NewMetric("system_file_descriptors_open", models.MetricTypeGauge, models.CategorySystem).
		WithLabels(c.labels).
		WithValue(float64(runtime.NumGoroutine())).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "count"
	metric.Description = "Number of open file descriptors"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// 确保实现了接口
var _ interfaces.MetricCollector = (*SystemCollector)(nil)