package external

import (
	"context"
	"fmt"
	"time"

	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// MockDeviceService 模拟设备服务
type MockDeviceService struct{}

// NewMockDeviceService 创建模拟设备服务
func NewMockDeviceService() *MockDeviceService {
	return &MockDeviceService{}
}

// GetDeviceInfo 获取设备信息
func (s *MockDeviceService) GetDeviceInfo(ctx context.Context, userID string) (*domainServices.DeviceInfo, error) {
	// 模拟设备信息
	return &domainServices.DeviceInfo{
		DeviceID:     fmt.Sprintf("device_%s", userID),
		DeviceType:   "smartphone",
		OS:           "iOS",
		OSVersion:    "15.0",
		AppVersion:   "1.0.0",
		ScreenSize:   "6.1",
		Resolution:   "1170x2532",
		Battery:      85,
		Storage:      "128GB",
		Memory:       "6GB",
		Network:      "WiFi",
		Location:     "enabled",
		Permissions:  []string{"camera", "microphone", "location"},
		Capabilities: []string{"AR", "VR", "NFC"},
		LastActive:   time.Now(),
		IsOnline:     true,
	}, nil
}

// GetDeviceCapabilities 获取设备能力
func (s *MockDeviceService) GetDeviceCapabilities(ctx context.Context, deviceID string) (map[string]bool, error) {
	// 模拟设备能力
	return map[string]bool{
		"camera":           true,
		"microphone":       true,
		"gps":              true,
		"bluetooth":        true,
		"nfc":              true,
		"fingerprint":      true,
		"face_id":          false,
		"ar_support":       true,
		"vr_support":       false,
		"voice_recognition": true,
		"gesture_control":  true,
		"offline_mode":     true,
		"sync_capability":  true,
	}, nil
}

// GetDevicePerformance 获取设备性能
func (s *MockDeviceService) GetDevicePerformance(ctx context.Context, deviceID string) (*domainServices.DevicePerformance, error) {
	// 模拟设备性能数据
	return &domainServices.DevicePerformance{
		CPUUsage:    25.5,
		MemoryUsage: 60.2,
		BatteryLife: 85,
		NetworkSpeed: 50.0,
		StorageUsed: 45.8,
		Temperature: 35.2,
		Timestamp:   time.Now(),
	}, nil
}

// UpdateDeviceSettings 更新设备设置
func (s *MockDeviceService) UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]interface{}) error {
	// 模拟更新设备设置
	// 在实际实现中，这里会调用设备API来更新设置
	return nil
}

// GetDeviceHistory 获取设备使用历史
func (s *MockDeviceService) GetDeviceHistory(ctx context.Context, deviceID string, limit int) ([]*domainServices.DeviceUsageRecord, error) {
	// 模拟设备使用历史
	records := []*domainServices.DeviceUsageRecord{
		{
			DeviceID:  deviceID,
			UserID:    "user123",
			StartTime: time.Now().Add(-2 * time.Hour),
			EndTime:   time.Now().Add(-time.Hour),
			Duration:  time.Hour,
			Activity:  "learning",
			AppUsage: map[string]time.Duration{
				"learning_app": 45 * time.Minute,
				"browser":      15 * time.Minute,
			},
			Performance: &domainServices.DevicePerformance{
				CPUUsage:     30.0,
				MemoryUsage:  55.0,
				BatteryLife:  90,
				NetworkSpeed: 45.0,
				StorageUsed:  44.0,
				Temperature:  33.0,
				Timestamp:    time.Now().Add(-90 * time.Minute),
			},
			Context: map[string]interface{}{
				"location": "home",
				"network":  "WiFi",
			},
		},
	}
	
	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}
	
	return records, nil
}

// MonitorDeviceHealth 监控设备健康状态
func (s *MockDeviceService) MonitorDeviceHealth(ctx context.Context, deviceID string) (*domainServices.DeviceHealth, error) {
	// 模拟设备健康状态
	return &domainServices.DeviceHealth{
		DeviceID:        deviceID,
		OverallHealth:   "good",
		BatteryHealth:   "excellent",
		PerformanceScore: 85.5,
		Issues:          []string{},
		Recommendations: []string{"定期清理缓存", "更新应用版本"},
		LastCheck:       time.Now(),
	}, nil
}