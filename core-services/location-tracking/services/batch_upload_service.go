package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchUploadService 批量上传服务
type BatchUploadService struct {
	db           *gorm.DB
	logger       *zap.Logger
	batchSize    int
	retryLimit   int
	retryDelay   time.Duration
	uploadQueue  chan *UploadTask
	workers      int
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

// UploadTask 上传任务
type UploadTask struct {
	UserID       string
	TrajectoryID string
	Points       []models.LocationPoint
	RetryCount   int
	CreatedAt    time.Time
}

// BatchUploadConfig 批量上传配置
type BatchUploadConfig struct {
	BatchSize    int           `yaml:"batch_size" json:"batch_size"`
	Workers      int           `yaml:"workers" json:"workers"`
	RetryLimit   int           `yaml:"retry_limit" json:"retry_limit"`
	RetryDelay   time.Duration `yaml:"retry_delay" json:"retry_delay"`
	QueueSize    int           `yaml:"queue_size" json:"queue_size"`
}

// NewBatchUploadService 创建批量上传服务
func NewBatchUploadService(db *gorm.DB, logger *zap.Logger, config BatchUploadConfig) *BatchUploadService {
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.Workers == 0 {
		config.Workers = 5
	}
	if config.RetryLimit == 0 {
		config.RetryLimit = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second * 5
	}
	if config.QueueSize == 0 {
		config.QueueSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &BatchUploadService{
		db:          db,
		logger:      logger,
		batchSize:   config.BatchSize,
		retryLimit:  config.RetryLimit,
		retryDelay:  config.RetryDelay,
		uploadQueue: make(chan *UploadTask, config.QueueSize),
		workers:     config.Workers,
		ctx:         ctx,
		cancel:      cancel,
	}

	// 启动工作协程
	service.startWorkers()

	return service
}

// startWorkers 启动工作协程
func (s *BatchUploadService) startWorkers() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// worker 工作协程
func (s *BatchUploadService) worker(id int) {
	defer s.wg.Done()
	
	s.logger.Info("Upload worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Upload worker stopped", zap.Int("worker_id", id))
			return
		case task := <-s.uploadQueue:
			s.processUploadTask(task)
		}
	}
}

// processUploadTask 处理上传任务
func (s *BatchUploadService) processUploadTask(task *UploadTask) {
	start := time.Now()
	
	err := s.uploadLocationPoints(task)
	if err != nil {
		s.logger.Error("Failed to upload location points",
			zap.Error(err),
			zap.String("user_id", task.UserID),
			zap.String("trajectory_id", task.TrajectoryID),
			zap.Int("point_count", len(task.Points)),
			zap.Int("retry_count", task.RetryCount))

		// 重试逻辑
		if task.RetryCount < s.retryLimit {
			task.RetryCount++
			
			// 延迟重试
			go func() {
				time.Sleep(s.retryDelay * time.Duration(task.RetryCount))
				select {
				case s.uploadQueue <- task:
				case <-s.ctx.Done():
				}
			}()
			
			s.logger.Info("Scheduled retry for upload task",
				zap.String("user_id", task.UserID),
				zap.String("trajectory_id", task.TrajectoryID),
				zap.Int("retry_count", task.RetryCount))
		} else {
			s.logger.Error("Upload task failed after max retries",
				zap.String("user_id", task.UserID),
				zap.String("trajectory_id", task.TrajectoryID),
				zap.Int("point_count", len(task.Points)))
		}
	} else {
		duration := time.Since(start)
		s.logger.Info("Successfully uploaded location points",
			zap.String("user_id", task.UserID),
			zap.String("trajectory_id", task.TrajectoryID),
			zap.Int("point_count", len(task.Points)),
			zap.Duration("duration", duration))
	}
}

// uploadLocationPoints 上传位置点数�?
func (s *BatchUploadService) uploadLocationPoints(task *UploadTask) error {
	// 分批处理
	for i := 0; i < len(task.Points); i += s.batchSize {
		end := i + s.batchSize
		if end > len(task.Points) {
			end = len(task.Points)
		}

		batch := task.Points[i:end]
		
		// 使用事务确保数据一致�?
		err := s.db.Transaction(func(tx *gorm.DB) error {
			for _, point := range batch {
				// 验证数据
				if err := s.validateLocationPoint(&point); err != nil {
					return fmt.Errorf("invalid location point: %w", err)
				}

				// 插入数据
				if err := tx.Create(&point).Error; err != nil {
					return fmt.Errorf("failed to insert location point: %w", err)
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to upload batch %d-%d: %w", i, end-1, err)
		}

		s.logger.Debug("Uploaded batch",
			zap.String("user_id", task.UserID),
			zap.String("trajectory_id", task.TrajectoryID),
			zap.Int("batch_start", i),
			zap.Int("batch_end", end-1),
			zap.Int("batch_size", len(batch)))
	}

	return nil
}

// validateLocationPoint 验证位置点数�?
func (s *BatchUploadService) validateLocationPoint(point *models.LocationPoint) error {
	if point.TrajectoryID == "" {
		return fmt.Errorf("trajectory_id is required")
	}
	
	if point.Latitude < -90 || point.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %f", point.Latitude)
	}
	
	if point.Longitude < -180 || point.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %f", point.Longitude)
	}
	
	if point.Timestamp == 0 {
		return fmt.Errorf("timestamp is required")
	}
	
	if point.Accuracy != nil && *point.Accuracy < 0 {
		return fmt.Errorf("invalid accuracy: %f", *point.Accuracy)
	}
	
	if point.Speed != nil && *point.Speed < 0 {
		return fmt.Errorf("invalid speed: %f", *point.Speed)
	}
	
	if point.Altitude != nil && (*point.Altitude < -1000 || *point.Altitude > 10000) {
		return fmt.Errorf("invalid altitude: %f", *point.Altitude)
	}

	return nil
}

// SubmitUploadTask 提交上传任务
func (s *BatchUploadService) SubmitUploadTask(userID, trajectoryID string, points []models.LocationPoint) error {
	if len(points) == 0 {
		return fmt.Errorf("no points to upload")
	}

	task := &UploadTask{
		UserID:       userID,
		TrajectoryID: trajectoryID,
		Points:       points,
		RetryCount:   0,
		CreatedAt:    time.Now(),
	}

	select {
	case s.uploadQueue <- task:
		s.logger.Info("Upload task submitted",
			zap.String("user_id", userID),
			zap.String("trajectory_id", trajectoryID),
			zap.Int("point_count", len(points)))
		return nil
	case <-s.ctx.Done():
		return fmt.Errorf("service is shutting down")
	default:
		return fmt.Errorf("upload queue is full")
	}
}

// GetQueueStatus 获取队列状�?
func (s *BatchUploadService) GetQueueStatus() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": len(s.uploadQueue),
		"queue_capacity": cap(s.uploadQueue),
		"workers": s.workers,
		"batch_size": s.batchSize,
		"retry_limit": s.retryLimit,
	}
}

// Stop 停止服务
func (s *BatchUploadService) Stop() {
	s.logger.Info("Stopping batch upload service...")
	
	s.cancel()
	close(s.uploadQueue)
	s.wg.Wait()
	
	s.logger.Info("Batch upload service stopped")
}
