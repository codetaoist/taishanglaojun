package interfaces

import (
	"context"

	domainservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// LearningAnalyticsService 学习分析服务接口
type LearningAnalyticsService interface {
	GenerateAnalyticsReport(ctx context.Context, req *domainservices.AnalyticsRequest) (*domainservices.LearningAnalyticsReport, error)
}

// KnowledgeGraphService 知识图谱服务接口
type KnowledgeGraphService interface {
	RecommendConcepts(ctx context.Context, req *domainservices.ConceptRecommendationRequest) ([]*domainservices.ConceptRecommendation, error)
	AnalyzeGraph(ctx context.Context, req *domainservices.GraphAnalysisRequest) (*domainservices.GraphAnalysisResult, error)
}

// LearningPathService 学习路径服务接口
type LearningPathService interface {
	GeneratePersonalizedPath(ctx context.Context, req interface{}) (interface{}, error)
	RecommendPaths(ctx context.Context, req interface{}) (interface{}, error)
}

