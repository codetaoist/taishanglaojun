package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
)

// BaseCollector ?
type BaseCollector struct {
	config CollectorConfig
	stats  *CollectorStats
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBaseCollector ?
func NewBaseCollector(config CollectorConfig) *BaseCollector {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &BaseCollector{
		config: config,
		stats:  &CollectorStats{},
		ctx:    ctx,
		cancel: cancel,
	}
}

// GetStats 
func (bc *BaseCollector) GetStats() *CollectorStats {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	stats := *bc.stats
	return &stats
}

// HealthCheck ?
func (bc *BaseCollector) HealthCheck() error {
	return nil
}

// updateStats 
func (bc *BaseCollector) updateStats(count int64, duration time.Duration) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.stats.CollectedMetrics += count
	bc.stats.LastCollection = time.Now()
	bc.stats.CollectionTime = duration
}

// recordError 
func (bc *BaseCollector) recordError() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.stats.Errors++
}

// CPUCollector CPU?
type CPUCollector struct {
	*BaseCollector
	lastCPUTimes map[string]uint64
}

// NewCPUCollector CPU?
func NewCPUCollector(config CollectorConfig) *CPUCollector {
	return &CPUCollector{
		BaseCollector: NewBaseCollector(config),
		lastCPUTimes:  make(map[string]uint64),
	}
}

// Start ?
func (cc *CPUCollector) Start() error {
	return nil
}

// Stop ?
func (cc *CPUCollector) Stop() error {
	cc.cancel()
	return nil
}

// Collect CPU
func (cc *CPUCollector) Collect() ([]interfaces.Metric, error) {
	start := time.Now()
	defer func() {
		cc.updateStats(1, time.Since(start))
	}()
	
	metrics := make([]interfaces.Metric, 0)
	timestamp := time.Now()
	
	// CPU?
	cpuUsage := cc.getCPUUsage()
	metrics = append(metrics, interfaces.Metric{
		Name:      "cpu_usage_percent",
		Value:     cpuUsage,
		Timestamp: timestamp,
		Labels:    map[string]string{"type": "total"},
	})
	
	// ?
	load1, load5, load15 := cc.getLoadAverage()
	metrics = append(metrics, 
		interfaces.Metric{
			Name:      "load_average",
			Value:     load1,
			Timestamp: timestamp,
			Labels:    map[string]string{"period": "1m"},
		},
		interfaces.Metric{
			Name:      "load_average",
			Value:     load5,
			Timestamp: timestamp,
			Labels:    map[string]string{"period": "5m"},
		},
		interfaces.Metric{
			Name:      "load_average",
			Value:     load15,
			Timestamp: timestamp,
			Labels:    map[string]string{"period": "15m"},
		},
	)
	
	// CPU?
	cores := runtime.NumCPU()
	metrics = append(metrics, interfaces.Metric{
		Name:      "cpu_cores",
		Value:     float64(cores),
		Timestamp: timestamp,
		Labels:    map[string]string{},
	})
	
	return metrics, nil
}

// GetMetrics CPU
func (cc *CPUCollector) GetMetrics() *CPUMetrics {
	return &CPUMetrics{
		Usage:    cc.getCPUUsage(),
		Cores:    runtime.NumCPU(),
		PerCore:  cc.getPerCoreUsage(),
	}
}

// getCPUUsage CPU?
func (cc *CPUCollector) getCPUUsage() float64 {
	//  /proc/stat API
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// CPU?
	return float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10.0
}

// getLoadAverage ?
func (cc *CPUCollector) getLoadAverage() (float64, float64, float64) {
	//  /proc/loadavg
	goroutines := float64(runtime.NumGoroutine())
	cores := float64(runtime.NumCPU())
	
	load := goroutines / cores
	return load, load * 0.9, load * 0.8
}

// getPerCoreUsage 
func (cc *CPUCollector) getPerCoreUsage() map[string]float64 {
	cores := runtime.NumCPU()
	usage := make(map[string]float64)
	
	for i := 0; i < cores; i++ {
		usage[fmt.Sprintf("cpu%d", i)] = cc.getCPUUsage() + float64(i)*2.0
	}
	
	return usage
}

// MemoryCollector ?
type MemoryCollector struct {
	*BaseCollector
}

// NewMemoryCollector ?
func NewMemoryCollector(config CollectorConfig) *MemoryCollector {
	return &MemoryCollector{
		BaseCollector: NewBaseCollector(config),
	}
}

// Start ?
func (mc *MemoryCollector) Start() error {
	return nil
}

// Stop ?
func (mc *MemoryCollector) Stop() error {
	mc.cancel()
	return nil
}

// Collect 
func (mc *MemoryCollector) Collect() ([]interfaces.Metric, error) {
	start := time.Now()
	defer func() {
		mc.updateStats(1, time.Since(start))
	}()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	metrics := make([]interfaces.Metric, 0)
	timestamp := time.Now()
	
	// ?
	metrics = append(metrics,
		interfaces.Metric{
			Name:      "memory_heap_bytes",
			Value:     float64(m.HeapAlloc),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "alloc"},
		},
		interfaces.Metric{
			Name:      "memory_heap_bytes",
			Value:     float64(m.HeapSys),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "sys"},
		},
		interfaces.Metric{
			Name:      "memory_heap_bytes",
			Value:     float64(m.HeapInuse),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "inuse"},
		},
		interfaces.Metric{
			Name:      "memory_heap_bytes",
			Value:     float64(m.HeapIdle),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "idle"},
		},
	)
	
	// GC
	metrics = append(metrics,
		interfaces.Metric{
			Name:      "memory_gc_runs_total",
			Value:     float64(m.NumGC),
			Timestamp: timestamp,
			Labels:    map[string]string{},
		},
		interfaces.Metric{
			Name:      "memory_gc_pause_seconds",
			Value:     float64(m.PauseTotalNs) / 1e9,
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "total"},
		},
	)
	
	// ?
	metrics = append(metrics,
		interfaces.Metric{
			Name:      "memory_stack_bytes",
			Value:     float64(m.StackInuse),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "inuse"},
		},
		interfaces.Metric{
			Name:      "memory_stack_bytes",
			Value:     float64(m.StackSys),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "sys"},
		},
	)
	
	return metrics, nil
}

// GetMetrics 
func (mc *MemoryCollector) GetMetrics() *MemoryMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 
	total := uint64(8 * 1024 * 1024 * 1024) // 8GB
	used := m.Sys
	free := total - used
	
	return &MemoryMetrics{
		Total:     total,
		Used:      used,
		Free:      free,
		Available: free + m.HeapIdle,
		Usage:     float64(used) / float64(total) * 100,
		Cached:    m.HeapIdle,
		Buffers:   m.StackSys,
		SwapTotal: 0,
		SwapUsed:  0,
		SwapFree:  0,
	}
}

// DiskCollector ?
type DiskCollector struct {
	*BaseCollector
	lastIOStats map[string]*DiskIOStats
}

// DiskIOStats IO
type DiskIOStats struct {
	ReadBytes  uint64
	WriteBytes uint64
	ReadOps    uint64
	WriteOps   uint64
	Timestamp  time.Time
}

// NewDiskCollector ?
func NewDiskCollector(config CollectorConfig) *DiskCollector {
	return &DiskCollector{
		BaseCollector: NewBaseCollector(config),
		lastIOStats:   make(map[string]*DiskIOStats),
	}
}

// Start ?
func (dc *DiskCollector) Start() error {
	return nil
}

// Stop ?
func (dc *DiskCollector) Stop() error {
	dc.cancel()
	return nil
}

// Collect 
func (dc *DiskCollector) Collect() ([]interfaces.Metric, error) {
	start := time.Now()
	defer func() {
		dc.updateStats(1, time.Since(start))
	}()
	
	metrics := make([]interfaces.Metric, 0)
	timestamp := time.Now()
	
	// 
	devices := []string{"sda", "sdb"}
	
	for _, device := range devices {
		// 
		total := uint64(100 * 1024 * 1024 * 1024) // 100GB
		used := uint64(60 * 1024 * 1024 * 1024)   // 60GB
		free := total - used
		usage := float64(used) / float64(total) * 100
		
		metrics = append(metrics,
			interfaces.Metric{
				Name:      "disk_space_bytes",
				Value:     float64(total),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "type": "total"},
			},
			interfaces.Metric{
				Name:      "disk_space_bytes",
				Value:     float64(used),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "type": "used"},
			},
			interfaces.Metric{
				Name:      "disk_space_bytes",
				Value:     float64(free),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "type": "free"},
			},
			interfaces.Metric{
				Name:      "disk_usage_percent",
				Value:     usage,
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device},
			},
		)
		
		// IO
		readBytes := uint64(1024 * 1024 * 100) // 100MB
		writeBytes := uint64(1024 * 1024 * 50) // 50MB
		readOps := uint64(1000)
		writeOps := uint64(500)
		
		metrics = append(metrics,
			interfaces.Metric{
				Name:      "disk_io_bytes",
				Value:     float64(readBytes),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "direction": "read"},
			},
			interfaces.Metric{
				Name:      "disk_io_bytes",
				Value:     float64(writeBytes),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "direction": "write"},
			},
			interfaces.Metric{
				Name:      "disk_io_ops",
				Value:     float64(readOps),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "direction": "read"},
			},
			interfaces.Metric{
				Name:      "disk_io_ops",
				Value:     float64(writeOps),
				Timestamp: timestamp,
				Labels:    map[string]string{"device": device, "direction": "write"},
			},
		)
	}
	
	return metrics, nil
}

// GetMetrics 
func (dc *DiskCollector) GetMetrics() *DiskMetrics {
	devices := make(map[string]*DiskDeviceMetrics)
	
	// 豸
	deviceNames := []string{"sda", "sdb"}
	
	for _, device := range deviceNames {
		total := uint64(100 * 1024 * 1024 * 1024) // 100GB
		used := uint64(60 * 1024 * 1024 * 1024)   // 60GB
		free := total - used
		
		devices[device] = &DiskDeviceMetrics{
			Total:      total,
			Used:       used,
			Free:       free,
			Usage:      float64(used) / float64(total) * 100,
			ReadBytes:  1024 * 1024 * 100, // 100MB
			WriteBytes: 1024 * 1024 * 50,  // 50MB
			ReadOps:    1000,
			WriteOps:   500,
			ReadTime:   100,  // 100ms
			WriteTime:  200,  // 200ms
			IOTime:     300,  // 300ms
			IOPS:       1500, // 1500 ops/s
			Throughput: 1024 * 1024 * 150, // 150MB/s
		}
	}
	
	return &DiskMetrics{
		Devices: devices,
	}
}

// NetworkCollector ?
type NetworkCollector struct {
	*BaseCollector
	lastNetStats map[string]*NetworkStats
}

// NetworkStats 
type NetworkStats struct {
	BytesReceived   uint64
	BytesSent       uint64
	PacketsReceived uint64
	PacketsSent     uint64
	Timestamp       time.Time
}

// NewNetworkCollector ?
func NewNetworkCollector(config CollectorConfig) *NetworkCollector {
	return &NetworkCollector{
		BaseCollector: NewBaseCollector(config),
		lastNetStats:  make(map[string]*NetworkStats),
	}
}

// Start ?
func (nc *NetworkCollector) Start() error {
	return nil
}

// Stop ?
func (nc *NetworkCollector) Stop() error {
	nc.cancel()
	return nil
}

// Collect 
func (nc *NetworkCollector) Collect() ([]interfaces.Metric, error) {
	start := time.Now()
	defer func() {
		nc.updateStats(1, time.Since(start))
	}()
	
	metrics := make([]interfaces.Metric, 0)
	timestamp := time.Now()
	
	// 
	interfaces := []string{"eth0", "lo"}
	
	for _, iface := range interfaces {
		// 
		bytesReceived := uint64(1024 * 1024 * 10)  // 10MB
		bytesSent := uint64(1024 * 1024 * 5)       // 5MB
		packetsReceived := uint64(10000)
		packetsSent := uint64(5000)
		
		metrics = append(metrics,
			interfaces.Metric{
				Name:      "network_bytes",
				Value:     float64(bytesReceived),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "received"},
			},
			interfaces.Metric{
				Name:      "network_bytes",
				Value:     float64(bytesSent),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "sent"},
			},
			interfaces.Metric{
				Name:      "network_packets",
				Value:     float64(packetsReceived),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "received"},
			},
			interfaces.Metric{
				Name:      "network_packets",
				Value:     float64(packetsSent),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "sent"},
			},
		)
		
		// ?
		errorsReceived := uint64(10)
		errorsSent := uint64(5)
		droppedReceived := uint64(2)
		droppedSent := uint64(1)
		
		metrics = append(metrics,
			interfaces.Metric{
				Name:      "network_errors",
				Value:     float64(errorsReceived),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "received"},
			},
			interfaces.Metric{
				Name:      "network_errors",
				Value:     float64(errorsSent),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "sent"},
			},
			interfaces.Metric{
				Name:      "network_dropped",
				Value:     float64(droppedReceived),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "received"},
			},
			interfaces.Metric{
				Name:      "network_dropped",
				Value:     float64(droppedSent),
				Timestamp: timestamp,
				Labels:    map[string]string{"interface": iface, "direction": "sent"},
			},
		)
	}
	
	return metrics, nil
}

// GetMetrics 
func (nc *NetworkCollector) GetMetrics() *NetworkMetrics {
	interfaces := make(map[string]*NetworkInterfaceMetrics)
	
	// 
	interfaceNames := []string{"eth0", "lo"}
	
	for _, iface := range interfaceNames {
		interfaces[iface] = &NetworkInterfaceMetrics{
			BytesReceived:   1024 * 1024 * 10, // 10MB
			BytesSent:       1024 * 1024 * 5,  // 5MB
			PacketsReceived: 10000,
			PacketsSent:     5000,
			ErrorsReceived:  10,
			ErrorsSent:      5,
			DroppedReceived: 2,
			DroppedSent:     1,
			Speed:           1000000000, // 1Gbps
			Duplex:          "full",
			MTU:             1500,
			RxRate:          1024 * 1024, // 1MB/s
			TxRate:          512 * 1024,  // 512KB/s
		}
	}
	
	return &NetworkMetrics{
		Interfaces: interfaces,
	}
}

// ProcessCollector ?
type ProcessCollector struct {
	*BaseCollector
}

// NewProcessCollector ?
func NewProcessCollector(config CollectorConfig) *ProcessCollector {
	return &ProcessCollector{
		BaseCollector: NewBaseCollector(config),
	}
}

// Start ?
func (pc *ProcessCollector) Start() error {
	return nil
}

// Stop ?
func (pc *ProcessCollector) Stop() error {
	pc.cancel()
	return nil
}

// Collect 
func (pc *ProcessCollector) Collect() ([]interfaces.Metric, error) {
	start := time.Now()
	defer func() {
		pc.updateStats(1, time.Since(start))
	}()
	
	metrics := make([]interfaces.Metric, 0)
	timestamp := time.Now()
	
	// 
	totalProcesses := runtime.NumGoroutine()
	
	metrics = append(metrics,
		interfaces.Metric{
			Name:      "process_count",
			Value:     float64(totalProcesses),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "total"},
		},
		interfaces.Metric{
			Name:      "process_count",
			Value:     float64(totalProcesses * 80 / 100),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "running"},
		},
		interfaces.Metric{
			Name:      "process_count",
			Value:     float64(totalProcesses * 15 / 100),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "sleeping"},
		},
		interfaces.Metric{
			Name:      "process_count",
			Value:     float64(totalProcesses * 5 / 100),
			Timestamp: timestamp,
			Labels:    map[string]string{"type": "stopped"},
		},
	)
	
	return metrics, nil
}

// GetMetrics 
func (pc *ProcessCollector) GetMetrics() *ProcessMetrics {
	totalProcesses := runtime.NumGoroutine()
	
	// 
	topCPU := []*ProcessInfo{
		{
			PID:           1234,
			Name:          "app",
			Command:       "/usr/bin/app",
			CPUUsage:      25.5,
			MemoryUsage:   1024 * 1024 * 100, // 100MB
			MemoryPercent: 1.2,
			Status:        "running",
			StartTime:     time.Now().Add(-time.Hour),
			User:          "root",
			Threads:       10,
			FDs:           50,
		},
	}
	
	topMemory := []*ProcessInfo{
		{
			PID:           5678,
			Name:          "database",
			Command:       "/usr/bin/database",
			CPUUsage:      15.2,
			MemoryUsage:   1024 * 1024 * 500, // 500MB
			MemoryPercent: 6.1,
			Status:        "running",
			StartTime:     time.Now().Add(-2 * time.Hour),
			User:          "postgres",
			Threads:       20,
			FDs:           100,
		},
	}
	
	return &ProcessMetrics{
		Count:     totalProcesses,
		Running:   totalProcesses * 80 / 100,
		Sleeping:  totalProcesses * 15 / 100,
		Stopped:   totalProcesses * 4 / 100,
		Zombie:    totalProcesses * 1 / 100,
		TopCPU:    topCPU,
		TopMemory: topMemory,
		Details:   make(map[string]*ProcessInfo),
	}
}

// CreateMetricCollector ?
func CreateMetricCollector(config CollectorConfig) (MetricCollector, error) {
	switch config.Type {
	case "cpu":
		return NewCPUCollector(config), nil
	case "memory":
		return NewMemoryCollector(config), nil
	case "disk":
		return NewDiskCollector(config), nil
	case "network":
		return NewNetworkCollector(config), nil
	case "process":
		return NewProcessCollector(config), nil
	default:
		return nil, fmt.Errorf("unknown collector type: %s", config.Type)
	}
}

