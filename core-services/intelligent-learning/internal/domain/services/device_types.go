package services

import "time"

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID     string    `json:"device_id"`     // 设备ID
	DeviceType   string    `json:"device_type"`   // 设备类型
	OS           string    `json:"os"`            // 操作系统
	OSVersion    string    `json:"os_version"`    // 操作系统版本
	AppVersion   string    `json:"app_version"`   // 应用版本
	ScreenSize   string    `json:"screen_size"`   // 屏幕尺寸
	Resolution   string    `json:"resolution"`    // 分辨率
	Battery      int       `json:"battery"`       // 电池电量
	Storage      string    `json:"storage"`       // 存储容量
	Memory       string    `json:"memory"`        // 内存容量
	Network      string    `json:"network"`       // 网络类型
	Location     string    `json:"location"`      // 位置服务状态
	Permissions  []string  `json:"permissions"`   // 权限列表
	Capabilities []string  `json:"capabilities"`  // 设备能力
	LastActive   time.Time `json:"last_active"`   // 最后活跃时间
	IsOnline     bool      `json:"is_online"`     // 是否在线
}

// DevicePerformance 设备性能
type DevicePerformance struct {
	CPUUsage     float64   `json:"cpu_usage"`     // CPU使用率
	MemoryUsage  float64   `json:"memory_usage"`  // 内存使用率
	BatteryLife  int       `json:"battery_life"`  // 电池寿命
	NetworkSpeed float64   `json:"network_speed"` // 网络速度
	StorageUsed  float64   `json:"storage_used"`  // 存储使用率
	Temperature  float64   `json:"temperature"`   // 设备温度
	Timestamp    time.Time `json:"timestamp"`     // 时间戳
}

// DeviceUsageRecord 设备使用记录
type DeviceUsageRecord struct {
	DeviceID    string                    `json:"device_id"`    // 设备ID
	UserID      string                    `json:"user_id"`      // 用户ID
	StartTime   time.Time                 `json:"start_time"`   // 开始时间
	EndTime     time.Time                 `json:"end_time"`     // 结束时间
	Duration    time.Duration             `json:"duration"`     // 使用时长
	Activity    string                    `json:"activity"`     // 活动类型
	AppUsage    map[string]time.Duration  `json:"app_usage"`    // 应用使用时长
	Performance *DevicePerformance        `json:"performance"`  // 性能数据
	Context     map[string]interface{}    `json:"context"`      // 上下文信息
}

// DeviceHealth 设备健康状态
type DeviceHealth struct {
	DeviceID         string    `json:"device_id"`         // 设备ID
	OverallHealth    string    `json:"overall_health"`    // 整体健康状态
	BatteryHealth    string    `json:"battery_health"`    // 电池健康状态
	PerformanceScore float64   `json:"performance_score"` // 性能评分
	Issues           []string  `json:"issues"`            // 问题列表
	Recommendations  []string  `json:"recommendations"`   // 建议列表
	LastCheck        time.Time `json:"last_check"`        // 最后检查时间
}