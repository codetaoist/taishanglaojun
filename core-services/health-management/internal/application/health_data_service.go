package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthDataService 健康数据应用服务
type HealthDataService struct {
	healthDataRepo domain.HealthDataRepository
	eventPublisher EventPublisher
}

// NewHealthDataService 创建健康数据服务
func NewHealthDataService(healthDataRepo domain.HealthDataRepository, eventPublisher EventPublisher) *HealthDataService {
	return &HealthDataService{
		healthDataRepo: healthDataRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateHealthDataRequest 创建健康数据请求
type CreateHealthDataRequest struct {
	UserID     uuid.UUID                `json:"user_id" validate:"required"`
	DataType   domain.HealthDataType    `json:"data_type" validate:"required"`
	Value      float64                  `json:"value" validate:"required,gt=0"`
	Unit       string                   `json:"unit" validate:"required"`
	Source     domain.HealthDataSource  `json:"source" validate:"required"`
	DeviceID   *string                  `json:"device_id,omitempty"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
	RecordedAt *time.Time               `json:"recorded_at,omitempty"`
}

// CreateHealthDataResponse 创建健康数据响应
type CreateHealthDataResponse struct {
	ID         uuid.UUID                `json:"id"`
	UserID     uuid.UUID                `json:"user_id"`
	DataType   domain.HealthDataType    `json:"data_type"`
	Value      float64                  `json:"value"`
	Unit       string                   `json:"unit"`
	Source     domain.HealthDataSource  `json:"source"`
	DeviceID   *string                  `json:"device_id,omitempty"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
	RecordedAt time.Time                `json:"recorded_at"`
	CreatedAt  time.Time                `json:"created_at"`
	IsAbnormal bool                     `json:"is_abnormal"`
	RiskLevel  domain.RiskLevel         `json:"risk_level"`
}

// CreateHealthData 创建健康数据
func (s *HealthDataService) CreateHealthData(ctx context.Context, req *CreateHealthDataRequest) (*CreateHealthDataResponse, error) {
	// 创建健康数据聚合�?
	healthData := domain.NewHealthData(req.UserID, req.DataType, req.Value, req.Unit, req.Source)
	
	// 设置可选字�?
	if req.DeviceID != nil {
		healthData.SetDeviceID(*req.DeviceID)
	}
	
	if req.Metadata != nil {
		healthData.SetMetadata(req.Metadata)
	}
	
	if req.RecordedAt != nil {
		healthData.RecordedAt = *req.RecordedAt
	}
	
	// 保存到仓�?
	if err := s.healthDataRepo.Save(ctx, healthData); err != nil {
		return nil, fmt.Errorf("failed to save health data: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, healthData.GetEvents()); err != nil {
		// 记录日志但不影响主流�?
		// TODO: 添加日志记录
	}
	
	// 检查是否异常并发布警报事件
	if healthData.IsAbnormal() {
		abnormalEvent := domain.NewAbnormalHealthDataDetectedEvent(
			healthData.ID, 
			healthData.UserID, 
			healthData.DataType, 
			healthData.Value, 
			healthData.GetRiskLevel(),
		)
		if err := s.eventPublisher.Publish(ctx, abnormalEvent); err != nil {
			// 记录日志但不影响主流�?
			// TODO: 添加日志记录
		}
	}
	
	// 清除事件
	healthData.ClearEvents()
	
	return &CreateHealthDataResponse{
		ID:         healthData.ID,
		UserID:     healthData.UserID,
		DataType:   healthData.DataType,
		Value:      healthData.Value,
		Unit:       healthData.Unit,
		Source:     healthData.Source,
		DeviceID:   healthData.DeviceID,
		Metadata:   healthData.Metadata,
		RecordedAt: healthData.RecordedAt,
		CreatedAt:  healthData.CreatedAt,
		IsAbnormal: healthData.IsAbnormal(),
		RiskLevel:  healthData.GetRiskLevel(),
	}, nil
}

// GetHealthDataRequest 获取健康数据请求
type GetHealthDataRequest struct {
	UserID    uuid.UUID              `json:"user_id" validate:"required"`
	DataType  *domain.HealthDataType `json:"data_type,omitempty"`
	StartTime *time.Time             `json:"start_time,omitempty"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Limit     int                    `json:"limit" validate:"min=1,max=100"`
	Offset    int                    `json:"offset" validate:"min=0"`
}

// GetHealthDataResponse 获取健康数据响应
type GetHealthDataResponse struct {
	Data       []*CreateHealthDataResponse `json:"data"`
	Total      int64                       `json:"total"`
	Limit      int                         `json:"limit"`
	Offset     int                         `json:"offset"`
	HasMore    bool                        `json:"has_more"`
}

// GetHealthData 获取健康数据
func (s *HealthDataService) GetHealthData(ctx context.Context, req *GetHealthDataRequest) (*GetHealthDataResponse, error) {
	var healthDataList []*domain.HealthData
	var total int64
	var err error
	
	// 设置默认�?
	if req.Limit == 0 {
		req.Limit = 20
	}
	
	// 根据查询条件获取数据
	if req.StartTime != nil && req.EndTime != nil {
		if req.DataType != nil {
			healthDataList, err = s.healthDataRepo.FindByUserIDTypeAndTimeRange(
				ctx, req.UserID, *req.DataType, *req.StartTime, *req.EndTime,
			)
		} else {
			healthDataList, err = s.healthDataRepo.FindByUserIDAndTimeRange(
				ctx, req.UserID, *req.StartTime, *req.EndTime,
			)
		}
	} else if req.DataType != nil {
		healthDataList, err = s.healthDataRepo.FindByUserIDAndType(
			ctx, req.UserID, *req.DataType, req.Limit, req.Offset,
		)
		if err == nil {
			total, err = s.healthDataRepo.CountByUserIDAndType(ctx, req.UserID, *req.DataType)
		}
	} else {
		healthDataList, err = s.healthDataRepo.FindByUserID(
			ctx, req.UserID, req.Limit, req.Offset,
		)
		if err == nil {
			total, err = s.healthDataRepo.CountByUserID(ctx, req.UserID)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get health data: %w", err)
	}
	
	// 转换为响应格�?
	data := make([]*CreateHealthDataResponse, len(healthDataList))
	for i, healthData := range healthDataList {
		data[i] = &CreateHealthDataResponse{
			ID:         healthData.ID,
			UserID:     healthData.UserID,
			DataType:   healthData.DataType,
			Value:      healthData.Value,
			Unit:       healthData.Unit,
			Source:     healthData.Source,
			DeviceID:   healthData.DeviceID,
			Metadata:   healthData.Metadata,
			RecordedAt: healthData.RecordedAt,
			CreatedAt:  healthData.CreatedAt,
			IsAbnormal: healthData.IsAbnormal(),
			RiskLevel:  healthData.GetRiskLevel(),
		}
	}
	
	return &GetHealthDataResponse{
		Data:    data,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: int64(req.Offset+len(data)) < total,
	}, nil
}

// GetLatestHealthDataRequest 获取最新健康数据请�?
type GetLatestHealthDataRequest struct {
	UserID   uuid.UUID             `json:"user_id" validate:"required"`
	DataType domain.HealthDataType `json:"data_type" validate:"required"`
}

// GetLatestHealthData 获取最新健康数�?
func (s *HealthDataService) GetLatestHealthData(ctx context.Context, req *GetLatestHealthDataRequest) (*CreateHealthDataResponse, error) {
	healthData, err := s.healthDataRepo.GetLatestByUserIDAndType(ctx, req.UserID, req.DataType)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest health data: %w", err)
	}
	
	if healthData == nil {
		return nil, nil
	}
	
	return &CreateHealthDataResponse{
		ID:         healthData.ID,
		UserID:     healthData.UserID,
		DataType:   healthData.DataType,
		Value:      healthData.Value,
		Unit:       healthData.Unit,
		Source:     healthData.Source,
		DeviceID:   healthData.DeviceID,
		Metadata:   healthData.Metadata,
		RecordedAt: healthData.RecordedAt,
		CreatedAt:  healthData.CreatedAt,
		IsAbnormal: healthData.IsAbnormal(),
		RiskLevel:  healthData.GetRiskLevel(),
	}, nil
}

// GetHealthDataStatsRequest 获取健康数据统计请求
type GetHealthDataStatsRequest struct {
	UserID    uuid.UUID             `json:"user_id" validate:"required"`
	DataType  domain.HealthDataType `json:"data_type" validate:"required"`
	StartTime time.Time             `json:"start_time" validate:"required"`
	EndTime   time.Time             `json:"end_time" validate:"required"`
}

// GetHealthDataStatsResponse 获取健康数据统计响应
type GetHealthDataStatsResponse struct {
	UserID    uuid.UUID             `json:"user_id"`
	DataType  domain.HealthDataType `json:"data_type"`
	StartTime time.Time             `json:"start_time"`
	EndTime   time.Time             `json:"end_time"`
	Count     int64                 `json:"count"`
	Average   float64               `json:"average"`
	Min       float64               `json:"min"`
	Max       float64               `json:"max"`
	Unit      string                `json:"unit"`
}

// GetHealthDataStats 获取健康数据统计
func (s *HealthDataService) GetHealthDataStats(ctx context.Context, req *GetHealthDataStatsRequest) (*GetHealthDataStatsResponse, error) {
	// 获取数据列表以计算统计信�?
	healthDataList, err := s.healthDataRepo.FindByUserIDTypeAndTimeRange(
		ctx, req.UserID, req.DataType, req.StartTime, req.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get health data for stats: %w", err)
	}
	
	if len(healthDataList) == 0 {
		return &GetHealthDataStatsResponse{
			UserID:    req.UserID,
			DataType:  req.DataType,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			Count:     0,
		}, nil
	}
	
	// 计算统计信息
	var sum, min, max float64
	unit := healthDataList[0].Unit
	
	for i, data := range healthDataList {
		if i == 0 {
			min = data.Value
			max = data.Value
		} else {
			if data.Value < min {
				min = data.Value
			}
			if data.Value > max {
				max = data.Value
			}
		}
		sum += data.Value
	}
	
	average := sum / float64(len(healthDataList))
	
	return &GetHealthDataStatsResponse{
		UserID:    req.UserID,
		DataType:  req.DataType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Count:     int64(len(healthDataList)),
		Average:   average,
		Min:       min,
		Max:       max,
		Unit:      unit,
	}, nil
}

// GetAbnormalHealthDataRequest 获取异常健康数据请求
type GetAbnormalHealthDataRequest struct {
	UserID    uuid.UUID  `json:"user_id" validate:"required"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	Offset    int        `json:"offset" validate:"min=0"`
}

// GetAbnormalHealthData 获取异常健康数据
func (s *HealthDataService) GetAbnormalHealthData(ctx context.Context, req *GetAbnormalHealthDataRequest) (*GetHealthDataResponse, error) {
	var healthDataList []*domain.HealthData
	var err error
	
	// 设置默认�?
	if req.Limit == 0 {
		req.Limit = 20
	}
	
	// 根据查询条件获取异常数据
	if req.StartTime != nil && req.EndTime != nil {
		healthDataList, err = s.healthDataRepo.FindAbnormalDataByUserIDAndTimeRange(
			ctx, req.UserID, *req.StartTime, *req.EndTime,
		)
	} else {
		healthDataList, err = s.healthDataRepo.FindAbnormalDataByUserID(
			ctx, req.UserID, req.Limit, req.Offset,
		)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get abnormal health data: %w", err)
	}
	
	// 转换为响应格�?
	data := make([]*CreateHealthDataResponse, len(healthDataList))
	for i, healthData := range healthDataList {
		data[i] = &CreateHealthDataResponse{
			ID:         healthData.ID,
			UserID:     healthData.UserID,
			DataType:   healthData.DataType,
			Value:      healthData.Value,
			Unit:       healthData.Unit,
			Source:     healthData.Source,
			DeviceID:   healthData.DeviceID,
			Metadata:   healthData.Metadata,
			RecordedAt: healthData.RecordedAt,
			CreatedAt:  healthData.CreatedAt,
			IsAbnormal: healthData.IsAbnormal(),
			RiskLevel:  healthData.GetRiskLevel(),
		}
	}
	
	return &GetHealthDataResponse{
		Data:    data,
		Total:   int64(len(data)),
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: false, // 简化处理，实际应该计算总数
	}, nil
}

// publishEvents 发布领域事件
func (s *HealthDataService) publishEvents(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.GetEventType(), err)
		}
	}
	return nil
}

// EventPublisher 事件发布器接�?
type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
}
