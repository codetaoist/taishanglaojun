package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	pb "github.com/codetaoist/taishanglaojun/core-services/consciousness/proto"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/services"
	"go.uber.org/zap"
)

// ConsciousnessServer gRPC服务器实现
type ConsciousnessServer struct {
	pb.UnimplementedConsciousnessServiceServer
	service *services.ConsciousnessService
	logger  *zap.Logger
}

// NewConsciousnessServer 创建gRPC服务器实�?func NewConsciousnessServer(service *services.ConsciousnessService, logger *zap.Logger) *ConsciousnessServer {
	return &ConsciousnessServer{
		service: service,
		logger:  logger,
	}
}

// Health 健康检�?func (s *ConsciousnessServer) Health(ctx context.Context, req *emptypb.Empty) (*pb.HealthResponse, error) {
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
		Service:    health["service"].(string),
		Version:    health["version"].(string),
		Status:     health["status"].(string),
		Timestamp:  timestamppb.New(health["timestamp"].(time.Time)),
		Components: components,
	}, nil
}

// GetStats 获取服务统计信息
func (s *ConsciousnessServer) GetStats(ctx context.Context, req *emptypb.Empty) (*pb.StatsResponse, error) {
	stats := s.service.GetStats()
	
	return &pb.StatsResponse{
		StartTime:           timestamppb.New(stats.StartTime),
		TotalRequests:       stats.TotalRequests,
		SuccessfulRequests:  stats.SuccessfulRequests,
		FailedRequests:      stats.FailedRequests,
		ActiveSessions:      stats.ActiveSessions,
		TotalSessions:       stats.TotalSessions,
		CompletedSessions:   stats.CompletedSessions,
		FusionSessions:      stats.FusionSessions,
		EvolutionTracking:   stats.EvolutionTracking,
		GeneOperations:      stats.GeneOperations,
		CoordinationSessions: stats.CoordinationSessions,
		AverageResponseTime: durationpb.New(stats.AverageResponseTime),
		LastUpdateTime:      timestamppb.New(stats.LastUpdateTime),
	}, nil
}

// ProcessConsciousness 处理综合意识请求
func (s *ConsciousnessServer) ProcessConsciousness(ctx context.Context, req *pb.ConsciousnessRequest) (*pb.ConsciousnessResponse, error) {
	// 转换请求
	consciousnessReq := &models.ConsciousnessRequest{
		RequestID: req.RequestId,
		EntityID:  req.EntityId,
		Timestamp: req.Timestamp.AsTime(),
	}

	// 转换子请�?	if req.FusionRequest != nil {
		consciousnessReq.FusionRequest = &models.FusionRequest{
			SessionID:           req.FusionRequest.SessionId,
			EntityID:           req.FusionRequest.EntityId,
			Strategy:           req.FusionRequest.Strategy,
			CarbonCapabilities: req.FusionRequest.CarbonCapabilities,
			SiliconCapabilities: req.FusionRequest.SiliconCapabilities,
			Parameters:         req.FusionRequest.Parameters,
		}
	}

	if req.EvolutionRequest != nil {
		consciousnessReq.EvolutionRequest = &models.EvolutionRequest{
			Type:               req.EvolutionRequest.Type,
			EntityID:          req.EvolutionRequest.EntityId,
			InitialMetrics:    req.EvolutionRequest.InitialMetrics,
			Metrics:           req.EvolutionRequest.Metrics,
			PredictionHorizon: int(req.EvolutionRequest.PredictionHorizon),
		}
	}

	if req.GeneRequest != nil {
		consciousnessReq.GeneRequest = &models.GeneRequest{
			Type:              req.GeneRequest.Type,
			EntityID:         req.GeneRequest.EntityId,
			GeneID:           req.GeneRequest.GeneId,
			Intensity:        req.GeneRequest.Intensity,
			Duration:         req.GeneRequest.Duration.AsDuration(),
			MutationType:     req.GeneRequest.MutationType,
			MutationIntensity: req.GeneRequest.MutationIntensity,
		}

		// 转换基因数据
		if req.GeneRequest.Gene != nil {
			consciousnessReq.GeneRequest.Gene = convertProtoToGene(req.GeneRequest.Gene)
		}

		if len(req.GeneRequest.InitialGenes) > 0 {
			consciousnessReq.GeneRequest.InitialGenes = make([]*models.QuantumGene, len(req.GeneRequest.InitialGenes))
			for i, gene := range req.GeneRequest.InitialGenes {
				consciousnessReq.GeneRequest.InitialGenes[i] = convertProtoToGene(gene)
			}
		}
	}

	if req.CoordinationRequest != nil {
		consciousnessReq.CoordinationRequest = &models.CoordinationRequest{
			SessionID:     req.CoordinationRequest.SessionId,
			EntityID:     req.CoordinationRequest.EntityId,
			BalanceWeights: req.CoordinationRequest.BalanceWeights,
			Parameters:   req.CoordinationRequest.Parameters,
		}

		// 转换轴请�?		if req.CoordinationRequest.SAxis != nil {
			consciousnessReq.CoordinationRequest.SAxis = &models.SAxisRequest{
				Capabilities:    req.CoordinationRequest.SAxis.Capabilities,
				ProgressionType: req.CoordinationRequest.SAxis.ProgressionType,
				Parameters:     req.CoordinationRequest.SAxis.Parameters,
			}
		}

		if req.CoordinationRequest.CAxis != nil {
			consciousnessReq.CoordinationRequest.CAxis = &models.CAxisRequest{
				Elements:        req.CoordinationRequest.CAxis.Elements,
				CompositionType: req.CoordinationRequest.CAxis.CompositionType,
				Parameters:     req.CoordinationRequest.CAxis.Parameters,
			}
		}

		if req.CoordinationRequest.TAxis != nil {
			consciousnessReq.CoordinationRequest.TAxis = &models.TAxisRequest{
				ThoughtContent: req.CoordinationRequest.TAxis.ThoughtContent,
				RealmType:     req.CoordinationRequest.TAxis.RealmType,
				Parameters:    req.CoordinationRequest.TAxis.Parameters,
			}
		}
	}

	// 处理请求
	response, err := s.service.ProcessConsciousnessRequest(consciousnessReq)
	if err != nil {
		s.logger.Error("Failed to process consciousness request", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to process request: %v", err)
	}

	// 转换响应
	results := make(map[string]string)
	for k, v := range response.Results {
		results[k] = fmt.Sprintf("%v", v)
	}

	return &pb.ConsciousnessResponse{
		RequestId: response.RequestID,
		EntityId:  response.EntityID,
		Timestamp: timestamppb.New(response.Timestamp),
		Success:   response.Success,
		Error:     response.Error,
		Results:   results,
	}, nil
}

// 融合引擎接口实现
func (s *ConsciousnessServer) StartFusion(ctx context.Context, req *pb.FusionRequest) (*pb.FusionResponse, error) {
	fusionReq := &models.FusionRequest{
		SessionID:           req.SessionId,
		EntityID:           req.EntityId,
		Strategy:           req.Strategy,
		CarbonCapabilities: req.CarbonCapabilities,
		SiliconCapabilities: req.SiliconCapabilities,
		Parameters:         req.Parameters,
	}

	result, err := s.service.GetFusionEngine().StartFusion(fusionReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start fusion: %v", err)
	}

	fusionResult, ok := result.(*models.FusionResult)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid fusion result type")
	}

	return &pb.FusionResponse{
		SessionId: fusionResult.SessionID,
		Success:   true,
		Result:    convertFusionResultToProto(fusionResult),
	}, nil
}

func (s *ConsciousnessServer) GetFusionStatus(ctx context.Context, req *pb.FusionStatusRequest) (*pb.FusionStatusResponse, error) {
	status, err := s.service.GetFusionEngine().GetFusionStatus(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "fusion session not found: %v", err)
	}

	fusionStatus, ok := status.(*models.FusionStatus)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid fusion status type")
	}

	response := &pb.FusionStatusResponse{
		SessionId: fusionStatus.SessionID,
		Status:    fusionStatus.Status,
		Progress:  fusionStatus.Progress,
	}

	if fusionStatus.Result != nil {
		response.Result = convertFusionResultToProto(fusionStatus.Result)
	}

	return response, nil
}

func (s *ConsciousnessServer) CancelFusion(ctx context.Context, req *pb.FusionCancelRequest) (*pb.FusionCancelResponse, error) {
	err := s.service.GetFusionEngine().CancelFusion(req.SessionId)
	if err != nil {
		return &pb.FusionCancelResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.FusionCancelResponse{
		Success: true,
		Message: "Fusion session cancelled successfully",
	}, nil
}

func (s *ConsciousnessServer) GetFusionHistory(ctx context.Context, req *pb.FusionHistoryRequest) (*pb.FusionHistoryResponse, error) {
	history, err := s.service.GetFusionEngine().GetFusionHistory(req.EntityId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get fusion history: %v", err)
	}

	fusionHistory, ok := history.(*models.FusionHistory)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid fusion history type")
	}

	results := make([]*pb.FusionResult, len(fusionHistory.Results))
	for i, result := range fusionHistory.Results {
		results[i] = convertFusionResultToProto(result)
	}

	return &pb.FusionHistoryResponse{
		Results: results,
		Total:   int32(fusionHistory.Total),
	}, nil
}

func (s *ConsciousnessServer) GetFusionStrategies(ctx context.Context, req *emptypb.Empty) (*pb.FusionStrategiesResponse, error) {
	strategies := s.service.GetFusionEngine().GetAvailableStrategies()
	return &pb.FusionStrategiesResponse{
		Strategies: strategies,
	}, nil
}

func (s *ConsciousnessServer) GetFusionMetrics(ctx context.Context, req *pb.FusionMetricsRequest) (*pb.FusionMetricsResponse, error) {
	metrics, err := s.service.GetFusionEngine().GetFusionMetrics(req.EntityId, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get fusion metrics: %v", err)
	}

	fusionMetrics, ok := metrics.(map[string]float64)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid fusion metrics type")
	}

	return &pb.FusionMetricsResponse{
		Metrics: fusionMetrics,
	}, nil
}

// 进化追踪接口实现
func (s *ConsciousnessServer) GetEvolutionState(ctx context.Context, req *pb.EvolutionStateRequest) (*pb.EvolutionStateResponse, error) {
	state, err := s.service.GetEvolutionTracker().GetEvolutionState(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "evolution state not found: %v", err)
	}

	evolutionState, ok := state.(*models.EvolutionState)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution state type")
	}

	return &pb.EvolutionStateResponse{
		EntityId:          evolutionState.EntityID,
		SequenceLevel:     int32(evolutionState.SequenceLevel),
		EvolutionProgress: evolutionState.EvolutionProgress,
		CurrentMetrics:    evolutionState.CurrentMetrics,
		LastUpdate:        timestamppb.New(evolutionState.LastUpdate),
	}, nil
}

func (s *ConsciousnessServer) UpdateEvolutionState(ctx context.Context, req *pb.UpdateEvolutionStateRequest) (*pb.UpdateEvolutionStateResponse, error) {
	state, err := s.service.GetEvolutionTracker().UpdateEvolution(req.EntityId, req.Metrics)
	if err != nil {
		return &pb.UpdateEvolutionStateResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	evolutionState, ok := state.(*models.EvolutionState)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution state type")
	}

	return &pb.UpdateEvolutionStateResponse{
		Success: true,
		Message: "Evolution state updated successfully",
		State:   convertEvolutionStateToProto(evolutionState),
	}, nil
}

func (s *ConsciousnessServer) TrackEvolution(ctx context.Context, req *pb.TrackEvolutionRequest) (*pb.TrackEvolutionResponse, error) {
	state, err := s.service.GetEvolutionTracker().TrackEvolution(req.EntityId, req.InitialMetrics)
	if err != nil {
		return &pb.TrackEvolutionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	evolutionState, ok := state.(*models.EvolutionState)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution state type")
	}

	return &pb.TrackEvolutionResponse{
		Success: true,
		Message: "Evolution tracking started successfully",
		State:   convertEvolutionStateToProto(evolutionState),
	}, nil
}

func (s *ConsciousnessServer) GetEvolutionPrediction(ctx context.Context, req *pb.EvolutionPredictionRequest) (*pb.EvolutionPredictionResponse, error) {
	predictions, err := s.service.GetEvolutionTracker().PredictEvolution(req.EntityId, int(req.Horizon))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get evolution prediction: %v", err)
	}

	evolutionPredictions, ok := predictions.([]*models.EvolutionPrediction)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution predictions type")
	}

	protoPredictions := make([]*pb.EvolutionPrediction, len(evolutionPredictions))
	for i, pred := range evolutionPredictions {
		protoPredictions[i] = &pb.EvolutionPrediction{
			Step:             int32(pred.Step),
			PredictedMetrics: pred.PredictedMetrics,
			Confidence:       pred.Confidence,
		}
	}

	return &pb.EvolutionPredictionResponse{
		Predictions: protoPredictions,
	}, nil
}

func (s *ConsciousnessServer) GetEvolutionPath(ctx context.Context, req *pb.EvolutionPathRequest) (*pb.EvolutionPathResponse, error) {
	path, err := s.service.GetEvolutionTracker().GetEvolutionPath(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get evolution path: %v", err)
	}

	evolutionPath, ok := path.([]*models.EvolutionPathStep)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution path type")
	}

	protoSteps := make([]*pb.EvolutionPathStep, len(evolutionPath))
	for i, step := range evolutionPath {
		protoSteps[i] = &pb.EvolutionPathStep{
			SequenceLevel:  int32(step.SequenceLevel),
			Description:    step.Description,
			Requirements:   step.Requirements,
			EstimatedTime:  durationpb.New(step.EstimatedTime),
		}
	}

	return &pb.EvolutionPathResponse{
		Steps: protoSteps,
	}, nil
}

func (s *ConsciousnessServer) GetEvolutionMilestones(ctx context.Context, req *pb.EvolutionMilestonesRequest) (*pb.EvolutionMilestonesResponse, error) {
	milestones, err := s.service.GetEvolutionTracker().GetEvolutionMilestones(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get evolution milestones: %v", err)
	}

	evolutionMilestones, ok := milestones.([]*models.EvolutionMilestone)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution milestones type")
	}

	protoMilestones := make([]*pb.EvolutionMilestone, len(evolutionMilestones))
	for i, milestone := range evolutionMilestones {
		protoMilestones[i] = &pb.EvolutionMilestone{
			Name:          milestone.Name,
			Description:   milestone.Description,
			SequenceLevel: int32(milestone.SequenceLevel),
			Achieved:      milestone.Achieved,
		}
		if milestone.AchievedAt != nil {
			protoMilestones[i].AchievedAt = timestamppb.New(*milestone.AchievedAt)
		}
	}

	return &pb.EvolutionMilestonesResponse{
		Milestones: protoMilestones,
	}, nil
}

func (s *ConsciousnessServer) GetSequenceLevel(ctx context.Context, req *pb.SequenceLevelRequest) (*pb.SequenceLevelResponse, error) {
	levelInfo, err := s.service.GetEvolutionTracker().GetSequenceLevelInfo(int(req.Level))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "sequence level not found: %v", err)
	}

	sequenceLevel, ok := levelInfo.(*models.SequenceLevel)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid sequence level type")
	}

	return &pb.SequenceLevelResponse{
		Level:        int32(sequenceLevel.Level),
		Name:         sequenceLevel.Name,
		Description:  sequenceLevel.Description,
		Requirements: sequenceLevel.Requirements,
		Capabilities: sequenceLevel.Capabilities,
	}, nil
}

func (s *ConsciousnessServer) GetEvolutionStats(ctx context.Context, req *pb.EvolutionStatsRequest) (*pb.EvolutionStatsResponse, error) {
	stats, err := s.service.GetEvolutionTracker().GetEvolutionStats(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get evolution stats: %v", err)
	}

	evolutionStats, ok := stats.(map[string]float64)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid evolution stats type")
	}

	return &pb.EvolutionStatsResponse{
		Stats: evolutionStats,
	}, nil
}

// 量子基因接口实现（部分示例）
func (s *ConsciousnessServer) CreateGenePool(ctx context.Context, req *pb.CreateGenePoolRequest) (*pb.CreateGenePoolResponse, error) {
	initialGenes := make([]*models.QuantumGene, len(req.InitialGenes))
	for i, gene := range req.InitialGenes {
		initialGenes[i] = convertProtoToGene(gene)
	}

	pool, err := s.service.GetGeneManager().CreateGenePool(req.EntityId, initialGenes)
	if err != nil {
		return &pb.CreateGenePoolResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	genePool, ok := pool.(*models.GenePool)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid gene pool type")
	}

	return &pb.CreateGenePoolResponse{
		Success: true,
		Message: "Gene pool created successfully",
		Pool:    convertGenePoolToProto(genePool),
	}, nil
}

// 三轴协调接口实现（部分示例）
func (s *ConsciousnessServer) StartCoordination(ctx context.Context, req *pb.StartCoordinationRequest) (*pb.StartCoordinationResponse, error) {
	coordinationReq := &models.CoordinationRequest{
		SessionID:     req.Request.SessionId,
		EntityID:     req.EntityId,
		BalanceWeights: req.Request.BalanceWeights,
		Parameters:   req.Request.Parameters,
	}

	// 转换轴请�?	if req.Request.SAxis != nil {
		coordinationReq.SAxis = &models.SAxisRequest{
			Capabilities:    req.Request.SAxis.Capabilities,
			ProgressionType: req.Request.SAxis.ProgressionType,
			Parameters:     req.Request.SAxis.Parameters,
		}
	}

	if req.Request.CAxis != nil {
		coordinationReq.CAxis = &models.CAxisRequest{
			Elements:        req.Request.CAxis.Elements,
			CompositionType: req.Request.CAxis.CompositionType,
			Parameters:     req.Request.CAxis.Parameters,
		}
	}

	if req.Request.TAxis != nil {
		coordinationReq.TAxis = &models.TAxisRequest{
			ThoughtContent: req.Request.TAxis.ThoughtContent,
			RealmType:     req.Request.TAxis.RealmType,
			Parameters:    req.Request.TAxis.Parameters,
		}
	}

	result, err := s.service.GetCoordinator().StartCoordination(coordinationReq)
	if err != nil {
		return &pb.StartCoordinationResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	coordinationResult, ok := result.(*models.CoordinationResult)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid coordination result type")
	}

	return &pb.StartCoordinationResponse{
		SessionId: coordinationResult.SessionID,
		Success:   true,
		Message:   "Coordination started successfully",
		Result:    convertCoordinationResultToProto(coordinationResult),
	}, nil
}

// 辅助转换函数
func convertProtoToGene(protoGene *pb.QuantumGene) *models.QuantumGene {
	gene := &models.QuantumGene{
		ID:              protoGene.Id,
		Type:            protoGene.Type,
		Sequence:        protoGene.Sequence,
		ExpressionLevel: protoGene.ExpressionLevel,
		Stability:       protoGene.Stability,
		Properties:      protoGene.Properties,
		CreatedAt:       protoGene.CreatedAt.AsTime(),
	}
	
	if protoGene.LastExpressed != nil {
		lastExpressed := protoGene.LastExpressed.AsTime()
		gene.LastExpressed = &lastExpressed
	}
	
	return gene
}

func convertGeneToProto(gene *models.QuantumGene) *pb.QuantumGene {
	protoGene := &pb.QuantumGene{
		Id:              gene.ID,
		Type:            gene.Type,
		Sequence:        gene.Sequence,
		ExpressionLevel: gene.ExpressionLevel,
		Stability:       gene.Stability,
		Properties:      gene.Properties,
		CreatedAt:       timestamppb.New(gene.CreatedAt),
	}
	
	if gene.LastExpressed != nil {
		protoGene.LastExpressed = timestamppb.New(*gene.LastExpressed)
	}
	
	return protoGene
}

func convertFusionResultToProto(result *models.FusionResult) *pb.FusionResult {
	return &pb.FusionResult{
		SessionId:            result.SessionID,
		Strategy:             result.Strategy,
		QualityScore:         result.QualityScore,
		EnhancedCapabilities: result.EnhancedCapabilities,
		SynergyPoints:        result.SynergyPoints,
		Metrics:              result.Metrics,
		CreatedAt:            timestamppb.New(result.CreatedAt),
	}
}

func convertEvolutionStateToProto(state *models.EvolutionState) *pb.EvolutionState {
	return &pb.EvolutionState{
		EntityId:          state.EntityID,
		SequenceLevel:     int32(state.SequenceLevel),
		EvolutionProgress: state.EvolutionProgress,
		CurrentMetrics:    state.CurrentMetrics,
		LastUpdate:        timestamppb.New(state.LastUpdate),
	}
}

func convertGenePoolToProto(pool *models.GenePool) *pb.GenePool {
	protoGenes := make([]*pb.QuantumGene, len(pool.Genes))
	for i, gene := range pool.Genes {
		protoGenes[i] = convertGeneToProto(gene)
	}

	return &pb.GenePool{
		EntityId:     pool.EntityID,
		Genes:        protoGenes,
		PoolMetrics:  pool.PoolMetrics,
		CreatedAt:    timestamppb.New(pool.CreatedAt),
		LastUpdated:  timestamppb.New(pool.LastUpdated),
	}
}

func convertCoordinationResultToProto(result *models.CoordinationResult) *pb.CoordinationResult {
	protoResult := &pb.CoordinationResult{
		SessionId:            result.SessionID,
		CoordinationMetrics:  result.CoordinationMetrics,
		CreatedAt:            timestamppb.New(result.CreatedAt),
	}

	// 转换轴结�?	if result.SAxisResult != nil {
		protoResult.SAxisResult = &pb.SAxisResult{
			OptimizedSequence:  result.SAxisResult.OptimizedSequence,
			CapabilityScores:   result.SAxisResult.CapabilityScores,
			SequenceEfficiency: result.SAxisResult.SequenceEfficiency,
			Recommendations:    result.SAxisResult.Recommendations,
		}
	}

	if result.CAxisResult != nil {
		layers := make([]*pb.CompositionLayer, len(result.CAxisResult.Layers))
		for i, layer := range result.CAxisResult.Layers {
			layers[i] = &pb.CompositionLayer{
				Level:      int32(layer.Level),
				Elements:   layer.Elements,
				Properties: layer.Properties,
			}
		}

		protoResult.CAxisResult = &pb.CAxisResult{
			Layers:               layers,
			CompositionIntegrity: result.CAxisResult.CompositionIntegrity,
			LayerMetrics:         result.CAxisResult.LayerMetrics,
			Recommendations:      result.CAxisResult.Recommendations,
		}
	}

	if result.TAxisResult != nil {
		protoResult.TAxisResult = &pb.TAxisResult{
			ProcessedThought: result.TAxisResult.ProcessedThought,
			RealmLevel:       int32(result.TAxisResult.RealmLevel),
			ThoughtDepth:     result.TAxisResult.ThoughtDepth,
			Insights:         result.TAxisResult.Insights,
			RealmMetrics:     result.TAxisResult.RealmMetrics,
		}
	}

	if result.BalanceResult != nil {
		protoResult.BalanceResult = &pb.BalanceResult{
			OptimizedWeights: result.BalanceResult.OptimizedWeights,
			BalanceScore:     result.BalanceResult.BalanceScore,
			BalanceMetrics:   result.BalanceResult.BalanceMetrics,
			Recommendations:  result.BalanceResult.Recommendations,
		}
	}

	if result.SynergyResult != nil {
		protoResult.SynergyResult = &pb.SynergyResult{
			ActiveCatalysts:  result.SynergyResult.ActiveCatalysts,
			SynergyLevel:     result.SynergyResult.SynergyLevel,
			SynergyMetrics:   result.SynergyResult.SynergyMetrics,
			SynergyEffects:   result.SynergyResult.SynergyEffects,
		}
	}

	return protoResult
}

// 其他接口方法的实现可以按照类似的模式继续添加...
