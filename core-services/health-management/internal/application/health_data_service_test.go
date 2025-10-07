package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"health-management/internal/domain"
)

// Mock repository
type MockHealthDataRepository struct {
	mock.Mock
}

func (m *MockHealthDataRepository) Save(ctx context.Context, healthData *domain.HealthData) error {
	args := m.Called(ctx, healthData)
	return args.Error(0)
}

func (m *MockHealthDataRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.HealthData, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthData), args.Error(1)
}

func (m *MockHealthDataRepository) Update(ctx context.Context, healthData *domain.HealthData) error {
	args := m.Called(ctx, healthData)
	return args.Error(0)
}

func (m *MockHealthDataRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHealthDataRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HealthData), args.Error(1)
}

func (m *MockHealthDataRepository) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, limit, offset int) ([]*domain.HealthData, error) {
	args := m.Called(ctx, userID, dataType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HealthData), args.Error(1)
}

func (m *MockHealthDataRepository) FindByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, limit, offset int) ([]*domain.HealthData, error) {
	args := m.Called(ctx, userID, startTime, endTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HealthData), args.Error(1)
}

func (m *MockHealthDataRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockHealthDataRepository) GetLatestByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType) (*domain.HealthData, error) {
	args := m.Called(ctx, userID, dataType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthData), args.Error(1)
}

func (m *MockHealthDataRepository) GetAverageByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, startTime, endTime time.Time) (float64, error) {
	args := m.Called(ctx, userID, dataType, startTime, endTime)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockHealthDataRepository) GetMinMaxByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, startTime, endTime time.Time) (float64, float64, error) {
	args := m.Called(ctx, userID, dataType, startTime, endTime)
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

func (m *MockHealthDataRepository) FindAbnormalData(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.HealthData), args.Error(1)
}

// Mock event publisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestHealthDataService_CreateHealthData(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &CreateHealthDataRequest{
		UserID:     userID,
		DataType:   domain.HeartRate,
		Value:      72.5,
		Unit:       "bpm",
		Source:     domain.SmartWatch,
		DeviceID:   stringPtr("apple_watch_001"),
		RecordedAt: time.Now(),
	}

	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.HealthData")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("*domain.HealthDataCreatedEvent")).Return(nil)

	response, err := service.CreateHealthData(context.Background(), request)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, domain.HeartRate, response.DataType)
	assert.Equal(t, 72.5, response.Value)
	assert.Equal(t, "bpm", response.Unit)
	assert.Equal(t, domain.SmartWatch, response.Source)
	assert.Equal(t, "apple_watch_001", *response.DeviceID)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestHealthDataService_CreateHealthData_ValidationError(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	tests := []struct {
		name    string
		request *CreateHealthDataRequest
	}{
		{
			name: "empty user ID",
			request: &CreateHealthDataRequest{
				UserID:     uuid.Nil,
				DataType:   domain.HeartRate,
				Value:      72.5,
				Unit:       "bpm",
				Source:     domain.SmartWatch,
				RecordedAt: time.Now(),
			},
		},
		{
			name: "negative value",
			request: &CreateHealthDataRequest{
				UserID:     uuid.New(),
				DataType:   domain.HeartRate,
				Value:      -10.0,
				Unit:       "bpm",
				Source:     domain.SmartWatch,
				RecordedAt: time.Now(),
			},
		},
		{
			name: "empty unit",
			request: &CreateHealthDataRequest{
				UserID:     uuid.New(),
				DataType:   domain.HeartRate,
				Value:      72.5,
				Unit:       "",
				Source:     domain.SmartWatch,
				RecordedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.CreateHealthData(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Nil(t, response)
		})
	}
}

func TestHealthDataService_GetHealthDataByID(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	healthDataID := uuid.New()
	userID := uuid.New()
	recordedAt := time.Now()

	healthData := &domain.HealthData{
		ID:         healthDataID,
		UserID:     userID,
		DataType:   domain.HeartRate,
		Value:      72.5,
		Unit:       "bpm",
		Source:     domain.SmartWatch,
		RecordedAt: recordedAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, healthDataID).Return(healthData, nil)

	response, err := service.GetHealthDataByID(context.Background(), healthDataID)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, healthDataID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, domain.HeartRate, response.DataType)
	assert.Equal(t, 72.5, response.Value)

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_GetHealthDataByID_NotFound(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	healthDataID := uuid.New()

	mockRepo.On("FindByID", mock.Anything, healthDataID).Return(nil, errors.New("not found"))

	response, err := service.GetHealthDataByID(context.Background(), healthDataID)

	assert.Error(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_GetHealthDataByUser(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &GetHealthDataByUserRequest{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	}

	healthDataList := []*domain.HealthData{
		{
			ID:         uuid.New(),
			UserID:     userID,
			DataType:   domain.HeartRate,
			Value:      72.5,
			Unit:       "bpm",
			Source:     domain.SmartWatch,
			RecordedAt: time.Now(),
		},
		{
			ID:         uuid.New(),
			UserID:     userID,
			DataType:   domain.BloodPressure,
			Value:      120.0,
			Unit:       "mmHg",
			Source:     domain.MedicalDevice,
			RecordedAt: time.Now(),
		},
	}

	mockRepo.On("FindByUserID", mock.Anything, userID, 10, 0).Return(healthDataList, nil)
	mockRepo.On("Count", mock.Anything, userID).Return(int64(2), nil)

	response, err := service.GetHealthDataByUser(context.Background(), request)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_GetLatestHealthData(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &GetLatestHealthDataRequest{
		UserID:   userID,
		DataType: domain.HeartRate,
	}

	healthData := &domain.HealthData{
		ID:         uuid.New(),
		UserID:     userID,
		DataType:   domain.HeartRate,
		Value:      72.5,
		Unit:       "bpm",
		Source:     domain.SmartWatch,
		RecordedAt: time.Now(),
	}

	mockRepo.On("GetLatestByUserIDAndType", mock.Anything, userID, domain.HeartRate).Return(healthData, nil)

	response, err := service.GetLatestHealthData(context.Background(), request)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, domain.HeartRate, response.DataType)
	assert.Equal(t, 72.5, response.Value)

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_GetHealthDataStatistics(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	request := &GetHealthDataStatisticsRequest{
		UserID:    userID,
		DataType:  domain.HeartRate,
		StartTime: startTime,
		EndTime:   endTime,
	}

	mockRepo.On("GetAverageByUserIDAndType", mock.Anything, userID, domain.HeartRate, startTime, endTime).Return(75.5, nil)
	mockRepo.On("GetMinMaxByUserIDAndType", mock.Anything, userID, domain.HeartRate, startTime, endTime).Return(65.0, 85.0, nil)

	response, err := service.GetHealthDataStatistics(context.Background(), request)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, domain.HeartRate, response.DataType)
	assert.Equal(t, 75.5, response.Average)
	assert.Equal(t, 65.0, response.Min)
	assert.Equal(t, 85.0, response.Max)

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_GetAbnormalHealthData(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &GetAbnormalHealthDataRequest{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	}

	abnormalData := []*domain.HealthData{
		{
			ID:         uuid.New(),
			UserID:     userID,
			DataType:   domain.HeartRate,
			Value:      120.0, // Abnormally high
			Unit:       "bpm",
			Source:     domain.SmartWatch,
			RecordedAt: time.Now(),
		},
	}

	mockRepo.On("FindAbnormalData", mock.Anything, userID, 10, 0).Return(abnormalData, nil)

	response, err := service.GetAbnormalHealthData(context.Background(), request)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, 120.0, response.Data[0].Value)

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_Repository_Error(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &CreateHealthDataRequest{
		UserID:     userID,
		DataType:   domain.HeartRate,
		Value:      72.5,
		Unit:       "bpm",
		Source:     domain.SmartWatch,
		RecordedAt: time.Now(),
	}

	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.HealthData")).Return(errors.New("database error"))

	response, err := service.CreateHealthData(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "database error")

	mockRepo.AssertExpectations(t)
}

func TestHealthDataService_EventPublisher_Error(t *testing.T) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &CreateHealthDataRequest{
		UserID:     userID,
		DataType:   domain.HeartRate,
		Value:      72.5,
		Unit:       "bpm",
		Source:     domain.SmartWatch,
		RecordedAt: time.Now(),
	}

	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.HealthData")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("*domain.HealthDataCreatedEvent")).Return(errors.New("event publish error"))

	// Event publishing error should not fail the operation
	response, err := service.CreateHealthData(context.Background(), request)

	assert.NoError(t, err)
	require.NotNil(t, response)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// Benchmark tests
func BenchmarkHealthDataService_CreateHealthData(b *testing.B) {
	mockRepo := new(MockHealthDataRepository)
	mockPublisher := new(MockEventPublisher)
	service := NewHealthDataService(mockRepo, mockPublisher)

	userID := uuid.New()
	request := &CreateHealthDataRequest{
		UserID:     userID,
		DataType:   domain.HeartRate,
		Value:      72.5,
		Unit:       "bpm",
		Source:     domain.SmartWatch,
		RecordedAt: time.Now(),
	}

	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.HealthData")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("*domain.HealthDataCreatedEvent")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateHealthData(context.Background(), request)
	}
}