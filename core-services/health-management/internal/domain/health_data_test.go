package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthData(t *testing.T) {
	userID := uuid.New()
	recordedAt := time.Now()

	tests := []struct {
		name        string
		userID      uuid.UUID
		dataType    HealthDataType
		value       float64
		unit        string
		source      HealthDataSource
		recordedAt  time.Time
		expectError bool
	}{
		{
			name:        "valid heart rate data",
			userID:      userID,
			dataType:    HeartRate,
			value:       72.5,
			unit:        "bpm",
			source:      SmartWatch,
			recordedAt:  recordedAt,
			expectError: false,
		},
		{
			name:        "valid blood pressure data",
			userID:      userID,
			dataType:    BloodPressure,
			value:       120.0,
			unit:        "mmHg",
			source:      MedicalDevice,
			recordedAt:  recordedAt,
			expectError: false,
		},
		{
			name:        "invalid negative value",
			userID:      userID,
			dataType:    HeartRate,
			value:       -10.0,
			unit:        "bpm",
			source:      SmartWatch,
			recordedAt:  recordedAt,
			expectError: true,
		},
		{
			name:        "empty unit",
			userID:      userID,
			dataType:    HeartRate,
			value:       72.5,
			unit:        "",
			source:      SmartWatch,
			recordedAt:  recordedAt,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthData, err := NewHealthData(
				tt.userID,
				tt.dataType,
				tt.value,
				tt.unit,
				tt.source,
				tt.recordedAt,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, healthData)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, healthData)
				assert.Equal(t, tt.userID, healthData.UserID)
				assert.Equal(t, tt.dataType, healthData.DataType)
				assert.Equal(t, tt.value, healthData.Value)
				assert.Equal(t, tt.unit, healthData.Unit)
				assert.Equal(t, tt.source, healthData.Source)
				assert.Equal(t, tt.recordedAt, healthData.RecordedAt)
				assert.NotEqual(t, uuid.Nil, healthData.ID)
			}
		})
	}
}

func TestHealthData_SetDeviceID(t *testing.T) {
	healthData := createTestHealthData(t)

	deviceID := "apple_watch_001"
	healthData.SetDeviceID(deviceID)

	assert.Equal(t, deviceID, *healthData.DeviceID)
}

func TestHealthData_SetMetadata(t *testing.T) {
	healthData := createTestHealthData(t)

	metadata := map[string]interface{}{
		"activity": "running",
		"location": "gym",
	}
	healthData.SetMetadata(metadata)

	assert.Equal(t, metadata, healthData.Metadata)
}

func TestHealthData_UpdateValue(t *testing.T) {
	healthData := createTestHealthData(t)
	originalValue := healthData.Value

	newValue := 80.0
	err := healthData.UpdateValue(newValue)

	assert.NoError(t, err)
	assert.Equal(t, newValue, healthData.Value)
	assert.NotEqual(t, originalValue, healthData.Value)

	// Test invalid value
	err = healthData.UpdateValue(-10.0)
	assert.Error(t, err)
	assert.Equal(t, newValue, healthData.Value) // Value should remain unchanged
}

func TestHealthData_IsAbnormal(t *testing.T) {
	tests := []struct {
		name     string
		dataType HealthDataType
		value    float64
		expected bool
	}{
		{
			name:     "normal heart rate",
			dataType: HeartRate,
			value:    75.0,
			expected: false,
		},
		{
			name:     "high heart rate",
			dataType: HeartRate,
			value:    120.0,
			expected: true,
		},
		{
			name:     "low heart rate",
			dataType: HeartRate,
			value:    45.0,
			expected: true,
		},
		{
			name:     "normal blood pressure systolic",
			dataType: BloodPressure,
			value:    120.0,
			expected: false,
		},
		{
			name:     "high blood pressure systolic",
			dataType: BloodPressure,
			value:    160.0,
			expected: true,
		},
		{
			name:     "normal body temperature",
			dataType: BodyTemperature,
			value:    36.8,
			expected: false,
		},
		{
			name:     "high body temperature",
			dataType: BodyTemperature,
			value:    39.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthData := &HealthData{
				DataType: tt.dataType,
				Value:    tt.value,
			}

			result := healthData.IsAbnormal()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthData_GetRiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		dataType HealthDataType
		value    float64
		expected string
	}{
		{
			name:     "normal heart rate - low risk",
			dataType: HeartRate,
			value:    75.0,
			expected: "low",
		},
		{
			name:     "slightly high heart rate - medium risk",
			dataType: HeartRate,
			value:    105.0,
			expected: "medium",
		},
		{
			name:     "very high heart rate - high risk",
			dataType: HeartRate,
			value:    140.0,
			expected: "high",
		},
		{
			name:     "extremely high heart rate - critical risk",
			dataType: HeartRate,
			value:    180.0,
			expected: "critical",
		},
		{
			name:     "normal blood pressure - low risk",
			dataType: BloodPressure,
			value:    120.0,
			expected: "low",
		},
		{
			name:     "high blood pressure - high risk",
			dataType: BloodPressure,
			value:    160.0,
			expected: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthData := &HealthData{
				DataType: tt.dataType,
				Value:    tt.value,
			}

			result := healthData.GetRiskLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthData_AddDomainEvent(t *testing.T) {
	healthData := createTestHealthData(t)

	event := &HealthDataCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: healthData.ID,
			EventType:   "HealthDataCreated",
			OccurredAt:  time.Now(),
		},
		UserID:     healthData.UserID,
		DataType:   healthData.DataType,
		Value:      healthData.Value,
		RecordedAt: healthData.RecordedAt,
	}

	healthData.AddDomainEvent(event)

	events := healthData.GetDomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, event, events[0])
}

func TestHealthData_ClearDomainEvents(t *testing.T) {
	healthData := createTestHealthData(t)

	// Add some events
	event1 := &HealthDataCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: healthData.ID,
			EventType:   "HealthDataCreated",
			OccurredAt:  time.Now(),
		},
	}
	event2 := &HealthDataUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: healthData.ID,
			EventType:   "HealthDataUpdated",
			OccurredAt:  time.Now(),
		},
	}

	healthData.AddDomainEvent(event1)
	healthData.AddDomainEvent(event2)

	assert.Len(t, healthData.GetDomainEvents(), 2)

	healthData.ClearDomainEvents()
	assert.Len(t, healthData.GetDomainEvents(), 0)
}

func TestHealthDataType_String(t *testing.T) {
	tests := []struct {
		dataType HealthDataType
		expected string
	}{
		{HeartRate, "heart_rate"},
		{BloodPressure, "blood_pressure"},
		{BloodSugar, "blood_sugar"},
		{BodyTemperature, "body_temperature"},
		{Weight, "weight"},
		{Height, "height"},
		{BMI, "bmi"},
		{Steps, "steps"},
		{SleepDuration, "sleep_duration"},
		{StressLevel, "stress_level"},
	}

	for _, tt := range tests {
		t.Run(string(tt.dataType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.dataType))
		})
	}
}

func TestHealthDataSource_String(t *testing.T) {
	tests := []struct {
		source   HealthDataSource
		expected string
	}{
		{ManualInput, "manual_input"},
		{SmartWatch, "smart_watch"},
		{FitnessTracker, "fitness_tracker"},
		{MedicalDevice, "medical_device"},
		{MobileApp, "mobile_app"},
		{IoTSensor, "iot_sensor"},
		{HospitalSystem, "hospital_system"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.source))
		})
	}
}

// Helper function to create test health data
func createTestHealthData(t *testing.T) *HealthData {
	userID := uuid.New()
	recordedAt := time.Now()

	healthData, err := NewHealthData(
		userID,
		HeartRate,
		72.5,
		"bpm",
		SmartWatch,
		recordedAt,
	)

	require.NoError(t, err)
	require.NotNil(t, healthData)

	return healthData
}

// Benchmark tests
func BenchmarkNewHealthData(b *testing.B) {
	userID := uuid.New()
	recordedAt := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewHealthData(
			userID,
			HeartRate,
			72.5,
			"bpm",
			SmartWatch,
			recordedAt,
		)
	}
}

func BenchmarkHealthData_IsAbnormal(b *testing.B) {
	healthData := &HealthData{
		DataType: HeartRate,
		Value:    120.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = healthData.IsAbnormal()
	}
}

func BenchmarkHealthData_GetRiskLevel(b *testing.B) {
	healthData := &HealthData{
		DataType: HeartRate,
		Value:    120.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = healthData.GetRiskLevel()
	}
}