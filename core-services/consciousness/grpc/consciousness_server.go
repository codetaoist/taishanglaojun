package grpc

import (
	"context"
	"time"

	// "google.golang.org/protobuf/types/known/durationpb"
	// "google.golang.org/protobuf/types/known/emptypb"
	// "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	// pb "github.com/codetaoist/taishanglaojun/core-services/consciousness/proto"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/services"
	"go.uber.org/zap"
)

// ConsciousnessServer gRPC
type ConsciousnessServer struct {
	service *services.ConsciousnessService
	logger  *zap.Logger
}

// NewConsciousnessServer gRPC
func NewConsciousnessServer(service *services.ConsciousnessService, logger *zap.Logger) *ConsciousnessServer {
	return &ConsciousnessServer{
		service: service,
		logger:  logger,
	}
}

// Health 
func (s *ConsciousnessServer) Health(ctx context.Context, req *emptypb.Empty) (*pb.HealthResponse, error) {
	health := s.service.Health()

	components := make(map[string]string)
	if comp, ok := health["components"].(map[string]interface{}); ok {
		for k, v := range comp {
			if str, ok := v.(string); ok {
				components[k] = str
			}
		}
	}

	return &pb.HealthResponse{
		Status:    health["status"].(string),
		Timestamp: timestamppb.New(health["timestamp"].(time.Time)),
		Version:   health["version"].(string),
	}, nil
}

// GetStats 
func (s *ConsciousnessServer) GetStats(ctx context.Context, req *emptypb.Empty) (*pb.StatsResponse, error) {
	stats := s.service.GetStats()

	return &pb.StatsResponse{
		ActiveSessions:    stats.ActiveSessions,
		TotalRequests:     stats.TotalRequests,
		AverageLatency:    durationpb.New(stats.AverageResponseTime),
		LastRequestTime:   timestamppb.New(stats.LastUpdateTime),
		ServiceStartTime:  timestamppb.New(stats.StartTime),
	}, nil
}

// ProcessConsciousness 
func (s *ConsciousnessServer) ProcessConsciousness(ctx context.Context, req *pb.ConsciousnessRequest) (*pb.ConsciousnessResponse, error) {
	// 简化的意识处理请求
	consciousnessReq := &models.ConsciousnessRequest{
		Type:       "general",
		EntityID:   req.SessionId, // 使用 SessionId 作为 EntityID
		Parameters: make(map[string]interface{}), // 初始化空的参数映射
	}

	// 调用服务
	result, err := s.service.ProcessConsciousnessRequest(consciousnessReq)
	if err != nil {
		s.logger.Error("ProcessConsciousness failed", zap.Error(err))
		return &pb.ConsciousnessResponse{
			SessionId: req.SessionId,
			Success:   false,
			Error:     err.Error(),
		}, nil
	}

	// 转换结果
	results := make(map[string]string)
	if result.Metadata != nil {
		for k, v := range result.Metadata {
			if str, ok := v.(string); ok {
				results[k] = str
			}
		}
	}

	// 从结果中提取输出
	output := ""
	if result.Result != nil {
		if resultMap, ok := result.Result.(map[string]interface{}); ok {
			if status, exists := resultMap["status"]; exists {
				if statusStr, ok := status.(string); ok {
					output = statusStr
				}
			}
		}
	}

	return &pb.ConsciousnessResponse{
		SessionId: req.SessionId,
		Output:    output,
		Metadata:  results,
		Success:   result.Success,
		Error:     result.Error,
	}, nil
}

// 融合相关方法暂时移除 - 等待 protobuf 文件更新

// 进化相关方法暂时移除 - 等待 protobuf 文件更新

// 基因相关方法暂时移除 - 等待 protobuf 文件更新

// 协调相关方法暂时移除 - 等待 protobuf 文件更新

// 转换函数暂时移除 - 等待 protobuf 文件更新

