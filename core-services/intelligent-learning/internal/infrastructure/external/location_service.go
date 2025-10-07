package external

import (
	"context"
	"fmt"
	"time"

	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// MockLocationService 模拟位置服务
type MockLocationService struct{}

// NewMockLocationService 创建模拟位置服务
func NewMockLocationService() *MockLocationService {
	return &MockLocationService{}
}

// GetLocation 获取用户位置
func (s *MockLocationService) GetLocation(ctx context.Context, userID string) (*domainServices.Location, error) {
	// 模拟位置数据
	return &domainServices.Location{
		Latitude:    39.9042,
		Longitude:   116.4074,
		Address:     "北京市朝阳区",
		City:        "北京",
		Country:     "中国",
		PlaceType:   "办公室",
		Familiarity: 0.8,
	}, nil
}

// GetLocationHistory 获取位置历史
func (s *MockLocationService) GetLocationHistory(ctx context.Context, userID string, limit int) ([]*domainServices.LocationRecord, error) {
	// 模拟历史位置数据
	records := []*domainServices.LocationRecord{
		{
			UserID:    userID,
			Timestamp: time.Now().Add(-time.Hour),
			Location: &domainServices.Location{
				Latitude:    39.9042,
				Longitude:   116.4074,
				Address:     "北京市朝阳区",
				City:        "北京",
				Country:     "中国",
				PlaceType:   "办公室",
				Familiarity: 0.8,
			},
			Duration: time.Hour,
		},
	}
	
	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}
	
	return records, nil
}

// GetLocationContext 获取位置上下文
func (s *MockLocationService) GetLocationContext(ctx context.Context, location *domainServices.Location) (*domainServices.LocationContext, error) {
	return &domainServices.LocationContext{
		PlaceType:     location.PlaceType,
		NoiseLevel:    0.3,
		CrowdLevel:    0.4,
		WiFiQuality:   0.9,
		Comfort:       0.8,
		Productivity:  0.7,
		Accessibility: true,
	}, nil
}

// WeatherService 天气服务接口实现
type MockWeatherService struct{}

// NewMockWeatherService 创建模拟天气服务
func NewMockWeatherService() *MockWeatherService {
	return &MockWeatherService{}
}

// GetWeather 获取天气信息
func (s *MockWeatherService) GetWeather(ctx context.Context, location *domainServices.Location) (*domainServices.WeatherInfo, error) {
	if location == nil {
		return nil, fmt.Errorf("位置信息不能为空")
	}
	
	// 模拟天气数据
	return &domainServices.WeatherInfo{
		Temperature:    22.5,
		Humidity:       65,
		Pressure:       1013.25,
		WindSpeed:      5.2,
		WindDirection:  "北风",
		Visibility:     10.0,
		UVIndex:        5,
		Condition:      "晴朗",
		Description:    "天气晴朗，适合学习",
		AirQuality:     "良好",
		Sunrise:        "06:30",
		Sunset:         "18:45",
		Comfort:        "舒适",
		LearningIndex:  8.5,
	}, nil
}

// GetWeatherForecast 获取天气预报
func (s *MockWeatherService) GetWeatherForecast(ctx context.Context, location *domainServices.Location, days int) ([]*domainServices.WeatherInfo, error) {
	if location == nil {
		return nil, fmt.Errorf("位置信息不能为空")
	}
	
	if days <= 0 {
		days = 1
	}
	
	// 模拟天气预报数据
	forecast := make([]*domainServices.WeatherInfo, days)
	for i := 0; i < days; i++ {
		forecast[i] = &domainServices.WeatherInfo{
			Temperature:    22.5 + float64(i)*0.5,
			Humidity:       float64(65 - i*2),
			Pressure:       1013.25,
			WindSpeed:      5.2,
			WindDirection:  "北风",
			Visibility:     10.0,
			UVIndex:        5,
			Condition:      "晴朗",
			Description:    fmt.Sprintf("第%d天天气晴朗", i+1),
			AirQuality:     "良好",
			Sunrise:        "06:30",
			Sunset:         "18:45",
			Comfort:        "舒适",
			LearningIndex:  8.5 - float64(i)*0.2,
		}
	}
	
	return forecast, nil
}