package external

import (
	"context"
	"fmt"
	"time"

	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// MockLocationService 
type MockLocationService struct{}

// NewMockLocationService 
func NewMockLocationService() *MockLocationService {
	return &MockLocationService{}
}

// GetLocation 
func (s *MockLocationService) GetLocation(ctx context.Context, userID string) (*domainServices.Location, error) {
	// 
	return &domainServices.Location{
		Latitude:    39.9042,
		Longitude:   116.4074,
		Address:     "",
		City:        "",
		Country:     "",
		PlaceType:   "?,
		Familiarity: 0.8,
	}, nil
}

// GetLocationHistory 
func (s *MockLocationService) GetLocationHistory(ctx context.Context, userID string, limit int) ([]*domainServices.LocationRecord, error) {
	// 
	records := []*domainServices.LocationRecord{
		{
			UserID:    userID,
			Timestamp: time.Now().Add(-time.Hour),
			Location: &domainServices.Location{
				Latitude:    39.9042,
				Longitude:   116.4074,
				Address:     "",
				City:        "",
				Country:     "",
				PlaceType:   "?,
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

// GetLocationContext ?
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

// WeatherService 
type MockWeatherService struct{}

// NewMockWeatherService 
func NewMockWeatherService() *MockWeatherService {
	return &MockWeatherService{}
}

// GetWeather 
func (s *MockWeatherService) GetWeather(ctx context.Context, location *domainServices.Location) (*domainServices.WeatherInfo, error) {
	if location == nil {
		return nil, fmt.Errorf("")
	}
	
	// 
	return &domainServices.WeatherInfo{
		Temperature:    22.5,
		Humidity:       65,
		Pressure:       1013.25,
		WindSpeed:      5.2,
		WindDirection:  "",
		Visibility:     10.0,
		UVIndex:        5,
		Condition:      "",
		Description:    "",
		AirQuality:     "",
		Sunrise:        "06:30",
		Sunset:         "18:45",
		Comfort:        "?,
		LearningIndex:  8.5,
	}, nil
}

// GetWeatherForecast 
func (s *MockWeatherService) GetWeatherForecast(ctx context.Context, location *domainServices.Location, days int) ([]*domainServices.WeatherInfo, error) {
	if location == nil {
		return nil, fmt.Errorf("")
	}
	
	if days <= 0 {
		days = 1
	}
	
	// 
	forecast := make([]*domainServices.WeatherInfo, days)
	for i := 0; i < days; i++ {
		forecast[i] = &domainServices.WeatherInfo{
			Temperature:    22.5 + float64(i)*0.5,
			Humidity:       float64(65 - i*2),
			Pressure:       1013.25,
			WindSpeed:      5.2,
			WindDirection:  "",
			Visibility:     10.0,
			UVIndex:        5,
			Condition:      "",
			Description:    fmt.Sprintf("?d?, i+1),
			AirQuality:     "",
			Sunrise:        "06:30",
			Sunset:         "18:45",
			Comfort:        "?,
			LearningIndex:  8.5 - float64(i)*0.2,
		}
	}
	
	return forecast, nil
}

