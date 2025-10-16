package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AIService AI服务
type AIService struct {
	db              *gorm.DB
	logger          *zap.Logger
	providerManager *providers.Manager
}

// NewAIService 创建AI服务
func NewAIService(db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) *AIService {
	return &AIService{
		db:              db,
		logger:          logger,
		providerManager: providerManager,
	}
}

// WisdomInterpretation 智慧解释
type WisdomInterpretation struct {
	WisdomID        string   `json:"wisdom_id"`
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	Interpretation  string   `json:"interpretation"`
	KeyPoints       []string `json:"key_points"`
	ModernRelevance string   `json:"modern_relevance"`
	PracticalAdvice string   `json:"practical_advice"`
	RelatedConcepts []string `json:"related_concepts"`
	//
	HistoricalContext    string `json:"historical_context"`
	EmotionalAnalysis    string `json:"emotional_analysis"`
	PhilosophicalDepth   string `json:"philosophical_depth"`
	CulturalSignificance string `json:"cultural_significance"`
}

// WisdomAnalysis 智慧分析
type WisdomAnalysis struct {
	WisdomID        string   `json:"wisdom_id"`
	Title           string   `json:"title"`
	AnalysisSummary string   `json:"analysis_summary"`
	KeyPoints       []string `json:"key_points"`
	ModernRelevance string   `json:"modern_relevance"`
	Recommendations []string `json:"recommendations"`
	//
	EmotionalTone     EmotionalAnalysis `json:"emotional_tone"`
	HistoricalContext HistoricalContext `json:"historical_context"`
	PhilosophicalCore PhilosophicalCore `json:"philosophical_core"`
	CulturalImpact    CulturalImpact    `json:"cultural_impact"`
}

// EmotionalAnalysis 情感分析
type EmotionalAnalysis struct {
	PrimaryEmotion string            `json:"primary_emotion"`
	EmotionalTone  string            `json:"emotional_tone"`
	Intensity      float64           `json:"intensity"`
	Keywords       []string          `json:"keywords"`
	Sentiment      string            `json:"sentiment"`
	Details        map[string]string `json:"details"`
}

// HistoricalContext 历史背景
type HistoricalContext struct {
	Period        string   `json:"period"`
	SocialContext string   `json:"social_context"`
	PoliticalBg   string   `json:"political_background"`
	CulturalEnv   string   `json:"cultural_environment"`
	KeyEvents     []string `json:"key_events"`
	Influences    []string `json:"influences"`
}

// PhilosophicalCore 哲学核心
type PhilosophicalCore struct {
	MainThought      string   `json:"main_thought"`
	PhilosophyType   string   `json:"philosophy_type"`
	CoreConcepts     []string `json:"core_concepts"`
	LogicalStructure string   `json:"logical_structure"`
	Methodology      string   `json:"methodology"`
}

// CulturalImpact 文化影响
type CulturalImpact struct {
	Influence        string   `json:"influence"`
	Legacy           string   `json:"legacy"`
	ModernAdaptation string   `json:"modern_adaptation"`
	GlobalRelevance  string   `json:"global_relevance"`
	RelatedWorks     []string `json:"related_works"`
}

// QARequest 问答请求
type QARequest struct {
	Question string `json:"question" binding:"required"`
	WisdomID string `json:"wisdom_id,omitempty"`
	Context  string `json:"context,omitempty"`
	Language string `json:"language,omitempty"`
}

// QAResponse 问答响应
type QAResponse struct {
	Question       string            `json:"question"`
	Answer         string            `json:"answer"`
	RelatedWisdoms []WisdomReference `json:"related_wisdoms"`
	Sources        []string          `json:"sources"`
	Confidence     float64           `json:"confidence"`
	Keywords       []string          `json:"keywords"`
	Category       string            `json:"category"`
}

// WisdomReference 相关智慧引用
type WisdomReference struct {
	WisdomID  string  `json:"wisdom_id"`
	Title     string  `json:"title"`
	Author    string  `json:"author"`
	School    string  `json:"school"`
	Excerpt   string  `json:"excerpt"`
	Relevance float64 `json:"relevance"`
}

// WisdomRecommendation 智慧推荐
type WisdomRecommendation struct {
	WisdomID  string  `json:"wisdom_id"`
	Title     string  `json:"title"`
	Author    string  `json:"author"`
	Category  string  `json:"category"`
	School    string  `json:"school"`
	Summary   string  `json:"summary"`
	Relevance float64 `json:"relevance"`
	Reason    string  `json:"reason"`
}

// InterpretWisdom 解释智慧
func (s *AIService) InterpretWisdom(ctx context.Context, wisdomID string) (*WisdomInterpretation, error) {
	//
	var wisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", wisdomID).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wisdom not found: %s", wisdomID)
		}
		return nil, fmt.Errorf("failed to get wisdom: %w", err)
	}

	//	构建解释提示
	prompt := s.buildInterpretationPrompt(wisdom)

	//	调用AI模型
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "你是一个专业的国学智慧分析师，请根据用户的问题提供深入的分析和见解。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.6,
		MaxTokens:   3000,
	}

	//	获取默认AI模型
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI interpretation failed: %w", err)
	}

	//	解析AI响应
	interpretation, err := s.parseInterpretationResponse(resp.Message.Content)
	if err != nil {
		s.logger.Warn("Failed to parse AI response, using raw content", zap.Error(err))
		//
		interpretation = &WisdomInterpretation{
			WisdomID:        wisdomID,
			Title:           wisdom.Title,
			Content:         wisdom.Content,
			Interpretation:  resp.Message.Content,
			KeyPoints:       []string{},
			ModernRelevance: "",
			PracticalAdvice: "",
			RelatedConcepts: []string{},
		}
	} else {
		interpretation.WisdomID = wisdomID
		interpretation.Title = wisdom.Title
		interpretation.Content = wisdom.Content
	}

	return interpretation, nil
}

// RecommendRelatedWisdom	推荐相关智慧
func (s *AIService) RecommendRelatedWisdom(ctx context.Context, wisdomID string, limit int) ([]WisdomRecommendation, error) {
	//	获取当前智慧
	var currentWisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", wisdomID).First(&currentWisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to get current wisdom: %w", err)
	}

	//	获取候选智慧
	var candidates []models.CulturalWisdom
	query := s.db.Where("id != ? AND status = ?", wisdomID, "published")

	//	根据当前智慧的分类和学校筛选候选智慧
	if currentWisdom.Category != "" || currentWisdom.School != "" {
		query = query.Where("category = ? OR school = ?", currentWisdom.Category, currentWisdom.School)
	}

	if err := query.Limit(limit * 3).Find(&candidates).Error; err != nil {
		return nil, fmt.Errorf("failed to get candidate wisdoms: %w", err)
	}

	if len(candidates) == 0 {
		return []WisdomRecommendation{}, nil
	}

	//	调用AI模型生成推荐
	recommendations, err := s.generateAIRecommendations(ctx, currentWisdom, candidates, limit)
	if err != nil {
		s.logger.Warn("AI recommendation failed, using fallback", zap.Error(err))
		//	使用基于规则的推荐
		return s.generateRuleBasedRecommendations(currentWisdom, candidates, limit), nil
	}

	return recommendations, nil
}

// buildInterpretationPrompt	构建解释提示
func (s *AIService) buildInterpretationPrompt(wisdom models.CulturalWisdom) string {
	var prompt strings.Builder

	prompt.WriteString("\n\n")

	//	添加智慧标题、内容、作者、学校
	prompt.WriteString("\n")
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Title))
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Content))
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Author))
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.School))

	if wisdom.Summary != "" {
		prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Summary))
	}

	//	添加推荐格式指令
	prompt.WriteString("\n\n")
	prompt.WriteString("1. 解释智慧的核心内容\n")
	prompt.WriteString("2. 分析智慧的历史背景和发展趋势\n")
	prompt.WriteString("3. 评估智慧的实用性和价值\n")
	prompt.WriteString("4. 推荐相关的智慧概念或主题\n")
	prompt.WriteString("5. 提供实际应用的建议或案例\n")
	prompt.WriteString("6. 总结智慧的核心思想和价值\n")

	prompt.WriteString("\nJSON\n")
	prompt.WriteString(`{
  "interpretation": "300?",
  "key_points": ["1", "2", "3", "4"],
  "modern_relevance": "50?",
  "practical_advice": "50?",
  "related_concepts": ["1", "2", "3"]
}`)

	return prompt.String()
}

// parseInterpretationResponse	解析解释响应
func (s *AIService) parseInterpretationResponse(content string) (*WisdomInterpretation, error) {
	// JSON 解析
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := content[start : end+1]

	var result struct {
		Interpretation  string   `json:"interpretation"`
		KeyPoints       []string `json:"key_points"`
		ModernRelevance string   `json:"modern_relevance"`
		PracticalAdvice string   `json:"practical_advice"`
		RelatedConcepts []string `json:"related_concepts"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &WisdomInterpretation{
		Interpretation:  result.Interpretation,
		KeyPoints:       result.KeyPoints,
		ModernRelevance: result.ModernRelevance,
		PracticalAdvice: result.PracticalAdvice,
		RelatedConcepts: result.RelatedConcepts,
	}, nil
}

// generateAIRecommendations AI 生成推荐
func (s *AIService) generateAIRecommendations(ctx context.Context, current models.CulturalWisdom, candidates []models.CulturalWisdom, limit int) ([]WisdomRecommendation, error) {
	//
	prompt := s.buildRecommendationPrompt(current, candidates, limit)

	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "JSON 解析\n",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.5,
		MaxTokens:   1500,
	}

	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, err
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	return s.parseRecommendationResponse(resp.Message.Content, candidates)
}

// buildRecommendationPrompt	构建推荐提示
func (s *AIService) buildRecommendationPrompt(current models.CulturalWisdom, candidates []models.CulturalWisdom, limit int) string {
	var prompt strings.Builder

	prompt.WriteString("\n\n")
	prompt.WriteString(fmt.Sprintf("%s\n", current.Title))
	prompt.WriteString(fmt.Sprintf("%s\n", current.Content))
	prompt.WriteString(fmt.Sprintf("%s\n", current.School))

	prompt.WriteString("\n\n")
	for i, candidate := range candidates {
		prompt.WriteString(fmt.Sprintf("%d. ID: %s, : %s, %s, : %s\n",
			i+1, candidate.ID, candidate.Title, candidate.Author, candidate.School))
	}

	prompt.WriteString(fmt.Sprintf("\n%dJSON\n", limit))
	prompt.WriteString(`[
  {
    "wisdom_id": "ID",
    "relevance": 0.95,
    "reason": ""
  }
]`)

	return prompt.String()
}

// parseRecommendationResponse	解析推荐响应
func (s *AIService) parseRecommendationResponse(content string, candidates []models.CulturalWisdom) ([]WisdomRecommendation, error) {
	//
	candidateMap := make(map[string]models.CulturalWisdom)
	for _, candidate := range candidates {
		candidateMap[candidate.ID] = candidate
	}

	// JSON 解析
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")

	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no valid JSON array found in response")
	}

	jsonStr := content[start : end+1]

	var aiRecommendations []struct {
		WisdomID  string  `json:"wisdom_id"`
		Relevance float64 `json:"relevance"`
		Reason    string  `json:"reason"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &aiRecommendations); err != nil {
		return nil, fmt.Errorf("failed to parse recommendation JSON: %w", err)
	}

	var recommendations []WisdomRecommendation
	for _, aiRec := range aiRecommendations {
		if wisdom, exists := candidateMap[aiRec.WisdomID]; exists {
			recommendations = append(recommendations, WisdomRecommendation{
				WisdomID:  wisdom.ID,
				Title:     wisdom.Title,
				Author:    wisdom.Author,
				Category:  wisdom.Category,
				School:    wisdom.School,
				Summary:   wisdom.Summary,
				Relevance: aiRec.Relevance,
				Reason:    aiRec.Reason,
			})
		}
	}

	return recommendations, nil
}

// generateRuleBasedRecommendations	基于规则生成推荐
func (s *AIService) generateRuleBasedRecommendations(current models.CulturalWisdom, candidates []models.CulturalWisdom, limit int) []WisdomRecommendation {
	var recommendations []WisdomRecommendation

	//
	for _, candidate := range candidates {
		relevance := s.calculateSimpleRelevance(current, candidate)
		if relevance > 0.3 { //
			reason := s.generateSimpleReason(current, candidate)
			recommendations = append(recommendations, WisdomRecommendation{
				WisdomID:  candidate.ID,
				Title:     candidate.Title,
				Author:    candidate.Author,
				Category:  candidate.Category,
				School:    candidate.School,
				Summary:   candidate.Summary,
				Relevance: relevance,
				Reason:    reason,
			})
		}
	}

	//
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations
}

// calculateSimpleRelevance	计算简单相关性
func (s *AIService) calculateSimpleRelevance(current, candidate models.CulturalWisdom) float64 {
	relevance := 0.0

	// 学校
	if current.School == candidate.School && current.School != "" {
		relevance += 0.4
	}

	// 分类
	if current.Category == candidate.Category && current.Category != "" {
		relevance += 0.3
	}

	// 作者
	if current.Author == candidate.Author && current.Author != "" {
		relevance += 0.2
	}

	// 标签
	if len(current.Tags) > 0 && len(candidate.Tags) > 0 {
		commonTags := 0
		for _, tag1 := range current.Tags {
			for _, tag2 := range candidate.Tags {
				if tag1 == tag2 {
					commonTags++
					break
				}
			}
		}
		if commonTags > 0 {
			relevance += float64(commonTags) / float64(len(current.Tags)) * 0.1
		}
	}

	return relevance
}

// generateSimpleReason	生成简单原因
func (s *AIService) generateSimpleReason(current, candidate models.CulturalWisdom) string {
	reasons := []string{}

	if current.School == candidate.School && current.School != "" {
		reasons = append(reasons, fmt.Sprintf("%s", current.School))
	}

	if current.Category == candidate.Category && current.Category != "" {
		reasons = append(reasons, fmt.Sprintf("%s", current.Category))
	}

	if current.Author == candidate.Author && current.Author != "" {
		reasons = append(reasons, fmt.Sprintf("%s", current.Author))
	}

	if len(reasons) == 0 {
		return ""
	}

	return strings.Join(reasons, "")
}

// AnalyzeWisdomInDepth	深度分析智慧
func (s *AIService) AnalyzeWisdomInDepth(ctx context.Context, wisdomID string) (*WisdomAnalysis, error) {
	// 获取智慧
	var wisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", wisdomID).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wisdom not found: %s", wisdomID)
		}
		return nil, fmt.Errorf("failed to get wisdom: %w", err)
	}

	// 构建分析提示
	prompt := s.buildDepthAnalysisPrompt(wisdom)

	// AI	调用
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.5,
		MaxTokens:   4000,
	}

	// AI	调用
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI depth analysis failed: %w", err)
	}

	// AI	解析响应
	analysis, err := s.parseDepthAnalysisResponse(resp.Message.Content)
	if err != nil {
		s.logger.Warn("Failed to parse depth analysis response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	analysis.WisdomID = wisdomID
	analysis.Title = wisdom.Title

	return analysis, nil
}

// buildDepthAnalysisPrompt	构建深度分析提示
func (s *AIService) buildDepthAnalysisPrompt(wisdom models.CulturalWisdom) string {
	var prompt strings.Builder

	prompt.WriteString("\n\n")

	// 智慧内容
	prompt.WriteString("\n")
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Title))
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Content))
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Author))
	prompt.WriteString(fmt.Sprintf("%s\n", wisdom.School))

	if wisdom.Summary != "" {
		prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Summary))
	}

	// 分析提示
	prompt.WriteString("\n\n")
	prompt.WriteString("\n")
	prompt.WriteString("1. \n")
	prompt.WriteString("2. \n")
	prompt.WriteString("3. \n")
	prompt.WriteString("4. \n")

	prompt.WriteString("\nJSON\n")
	prompt.WriteString(`{
  "analysis_summary": "",
  "key_points": ["1", "2", "3"],
  "modern_relevance": "0-1",
  "recommendations": ["1", "2", "3"],
  "emotional_tone": {
    "primary_emotion": "",
    "emotional_tone": "",
    "intensity": 0.8,
    "keywords": ["1", "2"],
    "sentiment": "positive/negative/neutral?,
    "details": {
      "1": "",
      "2": ""
    }
  },
  "historical_context": {
    "period": "",
    "social_context": "",
    "political_background": "",
    "cultural_environment": "",
    "key_events": ["1", "2"],
    "influences": ["1", "2"]
  },
  "philosophical_core": {
    "main_thought": "",
    "philosophy_type": "",
    "core_concepts": ["1", "2"],
    "logical_structure": "",
    "methodology": ""
  },
  "cultural_impact": {
    "influence": "",
    "legacy": "",
    "modern_adaptation": "",	
    "global_relevance": "",
    "related_works": ["1", "2"]
  }
}`)

	return prompt.String()
}

// parseDepthAnalysisResponse	解析深度分析响应
func (s *AIService) parseDepthAnalysisResponse(content string) (*WisdomAnalysis, error) {
	// JSON
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := content[start : end+1]

	var result WisdomAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// IntelligentQA	智能问答
func (s *AIService) IntelligentQA(ctx context.Context, request QARequest) (*QAResponse, error) {
	//
	prompt := s.buildQAPrompt(request)

	// AI	调用
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2500,
	}

	// AI	调用
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI QA failed: %w", err)
	}

	// AI	解析响应
	qaResponse, err := s.parseQAResponse(resp.Message.Content, request.Question)
	if err != nil {
		s.logger.Warn("Failed to parse QA response", zap.Error(err))
		//
		return &QAResponse{
			Question:       request.Question,
			Answer:         resp.Message.Content,
			RelatedWisdoms: []WisdomReference{},
			Sources:        []string{},
			Confidence:     0.7,
			Keywords:       []string{},
			Category:       "general",
		}, nil
	}

	// AI	解析响应	相关智慧
	relatedWisdoms, err := s.findRelatedWisdoms(ctx, request.Question, qaResponse.Keywords)
	if err != nil {
		s.logger.Warn("Failed to find related wisdoms", zap.Error(err))
	} else {
		qaResponse.RelatedWisdoms = relatedWisdoms
	}

	return qaResponse, nil
}

// buildQAPrompt	构建问答提示
func (s *AIService) buildQAPrompt(request QARequest) string {
	var prompt strings.Builder

	prompt.WriteString("\n\n")
	prompt.WriteString(fmt.Sprintf("%s\n", request.Question))

	if request.WisdomID != "" {
		// ID	智慧内容
		var wisdom models.CulturalWisdom
		if err := s.db.Where("id = ?", request.WisdomID).First(&wisdom).Error; err == nil {
			prompt.WriteString("\n\n")
			prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Title))
			prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Content))
			prompt.WriteString(fmt.Sprintf("%s\n", wisdom.Author))
			prompt.WriteString(fmt.Sprintf("%s\n", wisdom.School))
		}
	}

	if request.Context != "" {
		prompt.WriteString(fmt.Sprintf("\n%s\n", request.Context))
	}

	prompt.WriteString("\nJSON\n")
	prompt.WriteString(`{
  "answer": "1000",
  "sources": ["1", "2"],
  "confidence": 0.9,
  "keywords": ["1", "2", "3"],
  "category": ""
}`)

	return prompt.String()
}

// parseQAResponse	解析问答响应
func (s *AIService) parseQAResponse(content string, question string) (*QAResponse, error) {
	// JSON
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := content[start : end+1]

	var result struct {
		Answer     string   `json:"answer"`
		Sources    []string `json:"sources"`
		Confidence float64  `json:"confidence"`
		Keywords   []string `json:"keywords"`
		Category   string   `json:"category"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &QAResponse{
		Question:       question,
		Answer:         result.Answer,
		RelatedWisdoms: []WisdomReference{},
		Sources:        result.Sources,
		Confidence:     result.Confidence,
		Keywords:       result.Keywords,
		Category:       result.Category,
	}, nil
}

// findRelatedWisdoms	查找相关智慧
func (s *AIService) findRelatedWisdoms(ctx context.Context, question string, keywords []string) ([]WisdomReference, error) {
	var wisdoms []models.CulturalWisdom

	// 相关智慧	查询
	query := s.db.Model(&models.CulturalWisdom{})

	// 相关智慧	关键词
	if len(keywords) > 0 {
		var conditions []string
		var args []interface{}

		for _, keyword := range keywords {
			conditions = append(conditions, "(title LIKE ? OR content LIKE ? OR author LIKE ? OR school LIKE ?)")
			keywordPattern := "%" + keyword + "%"
			args = append(args, keywordPattern, keywordPattern, keywordPattern, keywordPattern)
		}

		if len(conditions) > 0 {
			query = query.Where(strings.Join(conditions, " OR "), args...)
		}
	}

	// 相关智慧	查询	限制5条
	if err := query.Limit(5).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("failed to find related wisdoms: %w", err)
	}

	var references []WisdomReference
	for _, wisdom := range wisdoms {
		// 相关智慧	计算相关性
		relevance := s.calculateQuestionRelevance(question, wisdom, keywords)

		// 相关智慧	截取内容
		excerpt := wisdom.Content
		if len(excerpt) > 100 {
			excerpt = excerpt[:100] + "..."
		}

		references = append(references, WisdomReference{
			WisdomID:  wisdom.ID,
			Title:     wisdom.Title,
			Author:    wisdom.Author,
			School:    wisdom.School,
			Excerpt:   excerpt,
			Relevance: relevance,
		})
	}

	return references, nil
}

// calculateQuestionRelevance	计算问题相关性
func (s *AIService) calculateQuestionRelevance(question string, wisdom models.CulturalWisdom, keywords []string) float64 {
	relevance := 0.0

	// 相关智慧	计算相关性	关键词
	for _, keyword := range keywords {
		if strings.Contains(wisdom.Title, keyword) {
			relevance += 0.3
		}
		if strings.Contains(wisdom.Content, keyword) {
			relevance += 0.2
		}
		if strings.Contains(wisdom.Author, keyword) {
			relevance += 0.1
		}
		if strings.Contains(wisdom.School, keyword) {
			relevance += 0.1
		}
	}

	// 相关智慧	计算相关性	标题
	if strings.Contains(wisdom.Title, question) {
		relevance += 0.2
	}

	// 相关智慧	计算相关性	限制1.0
	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance
}

// GetEmbedding	获取嵌入向量
func (s *AIService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("AI %w", err)
	}

	return provider.Embed(ctx, text)
}
