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

// AIService 文化智慧AI服务
type AIService struct {
	db              *gorm.DB
	logger          *zap.Logger
	providerManager *providers.Manager
}

// NewAIService 创建AI服务实例
func NewAIService(db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) *AIService {
	return &AIService{
		db:              db,
		logger:          logger,
		providerManager: providerManager,
	}
}

// WisdomInterpretation 智慧解读结果
type WisdomInterpretation struct {
	WisdomID         string   `json:"wisdom_id"`
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	Interpretation   string   `json:"interpretation"`
	KeyPoints        []string `json:"key_points"`
	ModernRelevance  string   `json:"modern_relevance"`
	PracticalAdvice  string   `json:"practical_advice"`
	RelatedConcepts  []string `json:"related_concepts"`
	// 新增分析维度
	HistoricalContext string   `json:"historical_context"`
	EmotionalAnalysis string   `json:"emotional_analysis"`
	PhilosophicalDepth string  `json:"philosophical_depth"`
	CulturalSignificance string `json:"cultural_significance"`
}

// WisdomAnalysis 智慧深度分析结果
type WisdomAnalysis struct {
	WisdomID          string            `json:"wisdom_id"`
	Title             string            `json:"title"`
	AnalysisSummary   string            `json:"analysis_summary"`
	KeyPoints         []string          `json:"key_points"`
	ModernRelevance   string            `json:"modern_relevance"`
	Recommendations   []string          `json:"recommendations"`
	// 新增详细分析维度
	EmotionalTone     EmotionalAnalysis `json:"emotional_tone"`
	HistoricalContext HistoricalContext `json:"historical_context"`
	PhilosophicalCore PhilosophicalCore `json:"philosophical_core"`
	CulturalImpact    CulturalImpact    `json:"cultural_impact"`
}

// EmotionalAnalysis 情感分析结果
type EmotionalAnalysis struct {
	PrimaryEmotion string            `json:"primary_emotion"`
	EmotionalTone  string            `json:"emotional_tone"`
	Intensity      float64           `json:"intensity"`
	Keywords       []string          `json:"keywords"`
	Sentiment      string            `json:"sentiment"`
	Details        map[string]string `json:"details"`
}

// HistoricalContext 历史背景分析
type HistoricalContext struct {
	Period         string   `json:"period"`
	SocialContext  string   `json:"social_context"`
	PoliticalBg    string   `json:"political_background"`
	CulturalEnv    string   `json:"cultural_environment"`
	KeyEvents      []string `json:"key_events"`
	Influences     []string `json:"influences"`
}

// PhilosophicalCore 哲学内核分析
type PhilosophicalCore struct {
	MainThought    string   `json:"main_thought"`
	PhilosophyType string   `json:"philosophy_type"`
	CoreConcepts   []string `json:"core_concepts"`
	LogicalStructure string `json:"logical_structure"`
	Methodology    string   `json:"methodology"`
}

// CulturalImpact 文化影响分析
type CulturalImpact struct {
	Influence      string   `json:"influence"`
	Legacy         string   `json:"legacy"`
	ModernAdaptation string `json:"modern_adaptation"`
	GlobalRelevance string  `json:"global_relevance"`
	RelatedWorks   []string `json:"related_works"`
}

// QARequest 问答请求
type QARequest struct {
	Question   string `json:"question" binding:"required"`
	WisdomID   string `json:"wisdom_id,omitempty"`
	Context    string `json:"context,omitempty"`
	Language   string `json:"language,omitempty"`
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

// WisdomReference 智慧引用
type WisdomReference struct {
	WisdomID string  `json:"wisdom_id"`
	Title    string  `json:"title"`
	Author   string  `json:"author"`
	School   string  `json:"school"`
	Excerpt  string  `json:"excerpt"`
	Relevance float64 `json:"relevance"`
}

// WisdomRecommendation 智慧推荐结果
type WisdomRecommendation struct {
	WisdomID    string  `json:"wisdom_id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Category    string  `json:"category"`
	School      string  `json:"school"`
	Summary     string  `json:"summary"`
	Relevance   float64 `json:"relevance"`
	Reason      string  `json:"reason"`
}

// InterpretWisdom 解读文化智慧
func (s *AIService) InterpretWisdom(ctx context.Context, wisdomID string) (*WisdomInterpretation, error) {
	// 获取智慧内容
	var wisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", wisdomID).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wisdom not found: %s", wisdomID)
		}
		return nil, fmt.Errorf("failed to get wisdom: %w", err)
	}

	// 构建解读提示
	prompt := s.buildInterpretationPrompt(wisdom)

	// 调用AI进行解读
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "你是一位精通中华传统文化的资深学者，具有深厚的古典文献功底和现代文化视野。你擅长从历史、哲学、文化、社会等多个维度深入解读古代智慧，并能够将传统智慧与现代生活紧密结合，为当代人提供有价值的启发和指导。请用学者的严谨态度和深入浅出的表达方式进行解读�?,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.6,
		MaxTokens:   3000,
	}

	// 使用默认AI提供�?
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI interpretation failed: %w", err)
	}

	// 解析AI响应
	interpretation, err := s.parseInterpretationResponse(resp.Message.Content)
	if err != nil {
		s.logger.Warn("Failed to parse AI response, using raw content", zap.Error(err))
		// 如果解析失败，使用原始内�?
		interpretation = &WisdomInterpretation{
			WisdomID:       wisdomID,
			Title:          wisdom.Title,
			Content:        wisdom.Content,
			Interpretation: resp.Message.Content,
			KeyPoints:      []string{},
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

// RecommendRelatedWisdom 推荐相关智慧
func (s *AIService) RecommendRelatedWisdom(ctx context.Context, wisdomID string, limit int) ([]WisdomRecommendation, error) {
	// 获取当前智慧内容
	var currentWisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", wisdomID).First(&currentWisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to get current wisdom: %w", err)
	}

	// 获取候选智慧列表（排除当前智慧�?
	var candidates []models.CulturalWisdom
	query := s.db.Where("id != ? AND status = ?", wisdomID, "published")
	
	// 优先推荐同类别或同学派的智慧
	if currentWisdom.Category != "" || currentWisdom.School != "" {
		query = query.Where("category = ? OR school = ?", currentWisdom.Category, currentWisdom.School)
	}
	
	if err := query.Limit(limit * 3).Find(&candidates).Error; err != nil {
		return nil, fmt.Errorf("failed to get candidate wisdoms: %w", err)
	}

	if len(candidates) == 0 {
		return []WisdomRecommendation{}, nil
	}

	// 使用AI进行智能推荐
	recommendations, err := s.generateAIRecommendations(ctx, currentWisdom, candidates, limit)
	if err != nil {
		s.logger.Warn("AI recommendation failed, using fallback", zap.Error(err))
		// 如果AI推荐失败，使用简单的规则推荐
		return s.generateRuleBasedRecommendations(currentWisdom, candidates, limit), nil
	}

	return recommendations, nil
}

// buildInterpretationPrompt 构建解读提示
func (s *AIService) buildInterpretationPrompt(wisdom models.CulturalWisdom) string {
	var prompt strings.Builder
	
	prompt.WriteString("作为一位深谙中华传统文化的学者，请对以下文化智慧进行深度解读。请从历史背景、哲学内涵、文化价值、现代应用等多个维度进行分析：\n\n")
	
	// 基本信息
	prompt.WriteString("【智慧信息】\n")
	prompt.WriteString(fmt.Sprintf("标题�?s\n", wisdom.Title))
	prompt.WriteString(fmt.Sprintf("内容�?s\n", wisdom.Content))
	prompt.WriteString(fmt.Sprintf("作者：%s\n", wisdom.Author))
	prompt.WriteString(fmt.Sprintf("学派�?s\n", wisdom.School))
	
	if wisdom.Summary != "" {
		prompt.WriteString(fmt.Sprintf("概要�?s\n", wisdom.Summary))
	}
	
	// 解读要求
	prompt.WriteString("\n【解读要求】\n")
	prompt.WriteString("1. 历史背景：请分析这段智慧产生的历史时代背景、社会环境和文化语境\n")
	prompt.WriteString("2. 深层含义：请深入阐释其哲学思想、精神内核和文化价值\n")
	prompt.WriteString("3. 关键要点：提炼出3-5个核心观点或重要概念\n")
	prompt.WriteString("4. 现代意义：分析其在当代社会的相关性、启发性和指导价值\n")
	prompt.WriteString("5. 实践应用：提供具体的现代生活应用建议和实践方法\n")
	prompt.WriteString("6. 相关概念：列出与此智慧相关的重要概念、思想或其他经典\n")
	
	prompt.WriteString("\n请严格按照以下JSON格式返回解读结果：\n")
	prompt.WriteString(`{
  "interpretation": "详细的智慧解读，包含历史背景分析、哲学内涵阐释、文化价值探讨等内容，字数不少于300�?,
  "key_points": ["核心要点1：具体阐�?, "核心要点2：具体阐�?, "核心要点3：具体阐�?, "核心要点4：具体阐�?],
  "modern_relevance": "详细分析这个智慧在现代社会的相关性、对当代人的启发意义以及在解决现代问题中的价值，字数不少�?50�?,
  "practical_advice": "基于这个智慧的具体实用建议，包括在工作、生活、人际关系、个人成长等方面的应用方法，字数不少�?50�?,
  "related_concepts": ["相关概念1：简要说�?, "相关概念2：简要说�?, "相关概念3：简要说�?]
}`)

	return prompt.String()
}

// parseInterpretationResponse 解析解读响应
func (s *AIService) parseInterpretationResponse(content string) (*WisdomInterpretation, error) {
	// 尝试提取JSON部分
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

// generateAIRecommendations 使用AI生成推荐
func (s *AIService) generateAIRecommendations(ctx context.Context, current models.CulturalWisdom, candidates []models.CulturalWisdom, limit int) ([]WisdomRecommendation, error) {
	// 构建推荐提示
	prompt := s.buildRecommendationPrompt(current, candidates, limit)
	
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "你是一位文化智慧推荐专家，能够根据用户当前阅读的智慧内容，推荐相关度高、有价值的其他智慧内容�?,
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

// buildRecommendationPrompt 构建推荐提示
func (s *AIService) buildRecommendationPrompt(current models.CulturalWisdom, candidates []models.CulturalWisdom, limit int) string {
	var prompt strings.Builder
	
	prompt.WriteString("用户正在阅读以下文化智慧：\n\n")
	prompt.WriteString(fmt.Sprintf("标题�?s\n", current.Title))
	prompt.WriteString(fmt.Sprintf("内容�?s\n", current.Content))
	prompt.WriteString(fmt.Sprintf("学派�?s\n", current.School))
	
	prompt.WriteString("\n候选推荐列表：\n")
	for i, candidate := range candidates {
		prompt.WriteString(fmt.Sprintf("%d. ID: %s, 标题: %s, 作�? %s, 学派: %s\n", 
			i+1, candidate.ID, candidate.Title, candidate.Author, candidate.School))
	}
	
	prompt.WriteString(fmt.Sprintf("\n请从候选列表中选择%d个最相关的智慧进行推荐，按照以下JSON格式返回：\n", limit))
	prompt.WriteString(`[
  {
    "wisdom_id": "候选智慧的ID",
    "relevance": 0.95,
    "reason": "推荐理由"
  }
]`)

	return prompt.String()
}

// parseRecommendationResponse 解析推荐响应
func (s *AIService) parseRecommendationResponse(content string, candidates []models.CulturalWisdom) ([]WisdomRecommendation, error) {
	// 创建候选智慧的映射
	candidateMap := make(map[string]models.CulturalWisdom)
	for _, candidate := range candidates {
		candidateMap[candidate.ID] = candidate
	}
	
	// 尝试提取JSON部分
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

// generateRuleBasedRecommendations 基于规则的推荐（备用方案�?
func (s *AIService) generateRuleBasedRecommendations(current models.CulturalWisdom, candidates []models.CulturalWisdom, limit int) []WisdomRecommendation {
	var recommendations []WisdomRecommendation
	
	// 简单的相似度计�?
	for _, candidate := range candidates {
		relevance := s.calculateSimpleRelevance(current, candidate)
		if relevance > 0.3 { // 设置最低相关度阈�?
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
	
	// 按相关度排序并限制数�?
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}
	
	return recommendations
}

// calculateSimpleRelevance 计算简单相关度
func (s *AIService) calculateSimpleRelevance(current, candidate models.CulturalWisdom) float64 {
	relevance := 0.0
	
	// 同学派加�?
	if current.School == candidate.School && current.School != "" {
		relevance += 0.4
	}
	
	// 同类别加�?
	if current.Category == candidate.Category && current.Category != "" {
		relevance += 0.3
	}
	
	// 同作者加�?
	if current.Author == candidate.Author && current.Author != "" {
		relevance += 0.2
	}
	
	// 标签相似度（简单实现）
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

// generateSimpleReason 生成简单推荐理�?
func (s *AIService) generateSimpleReason(current, candidate models.CulturalWisdom) string {
	reasons := []string{}
	
	if current.School == candidate.School && current.School != "" {
		reasons = append(reasons, fmt.Sprintf("同属%s学派", current.School))
	}
	
	if current.Category == candidate.Category && current.Category != "" {
		reasons = append(reasons, "同类别智�?)
	}
	
	if current.Author == candidate.Author && current.Author != "" {
		reasons = append(reasons, fmt.Sprintf("同为%s的智�?, current.Author))
	}
	
	if len(reasons) == 0 {
		return "相关智慧推荐"
	}
	
	return strings.Join(reasons, "�?)
}

// AnalyzeWisdomInDepth 深度分析文化智慧
func (s *AIService) AnalyzeWisdomInDepth(ctx context.Context, wisdomID string) (*WisdomAnalysis, error) {
	// 获取智慧内容
	var wisdom models.CulturalWisdom
	if err := s.db.Where("id = ?", wisdomID).First(&wisdom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wisdom not found: %s", wisdomID)
		}
		return nil, fmt.Errorf("failed to get wisdom: %w", err)
	}

	// 构建深度分析提示
	prompt := s.buildDepthAnalysisPrompt(wisdom)

	// 调用AI进行深度分析
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "你是一位具有深厚学术功底的文化学者和哲学家，精通中华传统文化的历史脉络、哲学思想、情感内涵和文化价值。请从多个专业维度对文化智慧进行深度分析，包括情感分析、历史背景、哲学内核和文化影响等方面�?,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.5,
		MaxTokens:   4000,
	}

	// 使用默认AI提供�?
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI depth analysis failed: %w", err)
	}

	// 解析AI响应
	analysis, err := s.parseDepthAnalysisResponse(resp.Message.Content)
	if err != nil {
		s.logger.Warn("Failed to parse depth analysis response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	analysis.WisdomID = wisdomID
	analysis.Title = wisdom.Title

	return analysis, nil
}

// buildDepthAnalysisPrompt 构建深度分析提示
func (s *AIService) buildDepthAnalysisPrompt(wisdom models.CulturalWisdom) string {
	var prompt strings.Builder
	
	prompt.WriteString("请对以下文化智慧进行全面深度分析，从情感、历史、哲学、文化等多个专业维度进行详细解读：\n\n")
	
	// 基本信息
	prompt.WriteString("【智慧信息】\n")
	prompt.WriteString(fmt.Sprintf("标题�?s\n", wisdom.Title))
	prompt.WriteString(fmt.Sprintf("内容�?s\n", wisdom.Content))
	prompt.WriteString(fmt.Sprintf("作者：%s\n", wisdom.Author))
	prompt.WriteString(fmt.Sprintf("学派�?s\n", wisdom.School))
	
	if wisdom.Summary != "" {
		prompt.WriteString(fmt.Sprintf("概要�?s\n", wisdom.Summary))
	}
	
	// 分析要求
	prompt.WriteString("\n【分析要求】\n")
	prompt.WriteString("请从以下四个专业维度进行深度分析：\n")
	prompt.WriteString("1. 情感分析：分析文本的情感色彩、情感强度、主要情感倾向\n")
	prompt.WriteString("2. 历史背景：分析产生的历史时期、社会环境、政治背景、文化环境\n")
	prompt.WriteString("3. 哲学内核：分析主要哲学思想、核心概念、逻辑结构、思维方法\n")
	prompt.WriteString("4. 文化影响：分析文化影响力、历史传承、现代适应性、全球相关性\n")
	
	prompt.WriteString("\n请严格按照以下JSON格式返回分析结果：\n")
	prompt.WriteString(`{
  "analysis_summary": "整体分析总结，概括这个智慧的核心价值和意义",
  "key_points": ["关键�?", "关键�?", "关键�?"],
  "modern_relevance": "现代相关性分�?,
  "recommendations": ["建议1", "建议2", "建议3"],
  "emotional_tone": {
    "primary_emotion": "主要情感（如：平静、激昂、忧虑、喜悦等�?,
    "emotional_tone": "情感基调（如：积极、消极、中性）",
    "intensity": 0.8,
    "keywords": ["情感关键�?", "情感关键�?"],
    "sentiment": "情感倾向（positive/negative/neutral�?,
    "details": {
      "情感特征1": "具体描述",
      "情感特征2": "具体描述"
    }
  },
  "historical_context": {
    "period": "历史时期",
    "social_context": "社会背景描述",
    "political_background": "政治背景描述",
    "cultural_environment": "文化环境描述",
    "key_events": ["重要历史事件1", "重要历史事件2"],
    "influences": ["影响因素1", "影响因素2"]
  },
  "philosophical_core": {
    "main_thought": "主要哲学思想",
    "philosophy_type": "哲学类型（如：儒家、道家、法家等�?,
    "core_concepts": ["核心概念1", "核心概念2"],
    "logical_structure": "逻辑结构分析",
    "methodology": "思维方法�?
  },
  "cultural_impact": {
    "influence": "文化影响力描�?,
    "legacy": "历史传承情况",
    "modern_adaptation": "现代适应�?,
    "global_relevance": "全球相关�?,
    "related_works": ["相关作品1", "相关作品2"]
  }
}`)

	return prompt.String()
}

// parseDepthAnalysisResponse 解析深度分析响应
func (s *AIService) parseDepthAnalysisResponse(content string) (*WisdomAnalysis, error) {
	// 尝试提取JSON部分
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

// IntelligentQA 智能问答
func (s *AIService) IntelligentQA(ctx context.Context, request QARequest) (*QAResponse, error) {
	// 构建问答提示
	prompt := s.buildQAPrompt(request)

	// 调用AI进行问答
	chatReq := providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "system",
				Content: "你是一位博学的中华传统文化专家和智慧导师，精通各种古典文献、哲学思想和文化典故。你能够准确回答关于中华传统文化的各种问题，并能引用相关的经典智慧进行说明。请用专业易于懂的语言回答问题，并尽可能引用相关的文化智慧作为支撑�?,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2500,
	}

	// 使用默认AI提供�?
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	resp, err := provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI QA failed: %w", err)
	}

	// 解析AI响应
	qaResponse, err := s.parseQAResponse(resp.Message.Content, request.Question)
	if err != nil {
		s.logger.Warn("Failed to parse QA response", zap.Error(err))
		// 如果解析失败，返回基本响�?
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

	// 查找相关智慧
	relatedWisdoms, err := s.findRelatedWisdoms(ctx, request.Question, qaResponse.Keywords)
	if err != nil {
		s.logger.Warn("Failed to find related wisdoms", zap.Error(err))
	} else {
		qaResponse.RelatedWisdoms = relatedWisdoms
	}

	return qaResponse, nil
}

// buildQAPrompt 构建问答提示
func (s *AIService) buildQAPrompt(request QARequest) string {
	var prompt strings.Builder
	
	prompt.WriteString("请回答以下关于中华传统文化的问题：\n\n")
	prompt.WriteString(fmt.Sprintf("问题�?s\n", request.Question))
	
	if request.WisdomID != "" {
		// 如果指定了智慧ID，获取相关智慧内容作为上下文
		var wisdom models.CulturalWisdom
		if err := s.db.Where("id = ?", request.WisdomID).First(&wisdom).Error; err == nil {
			prompt.WriteString("\n相关智慧内容：\n")
			prompt.WriteString(fmt.Sprintf("标题�?s\n", wisdom.Title))
			prompt.WriteString(fmt.Sprintf("内容�?s\n", wisdom.Content))
			prompt.WriteString(fmt.Sprintf("作者：%s\n", wisdom.Author))
			prompt.WriteString(fmt.Sprintf("学派�?s\n", wisdom.School))
		}
	}
	
	if request.Context != "" {
		prompt.WriteString(fmt.Sprintf("\n补充上下文：%s\n", request.Context))
	}
	
	prompt.WriteString("\n请按照以下JSON格式回答：\n")
	prompt.WriteString(`{
  "answer": "详细的回答内容，要求准确、专业、易懂，字数不少�?00�?,
  "sources": ["引用的经典或文献1", "引用的经典或文献2"],
  "confidence": 0.9,
  "keywords": ["关键�?", "关键�?", "关键�?"],
  "category": "问题分类（如：哲学思想、历史文化、道德修养、人生智慧等�?
}`)

	return prompt.String()
}

// parseQAResponse 解析问答响应
func (s *AIService) parseQAResponse(content string, question string) (*QAResponse, error) {
	// 尝试提取JSON部分
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

// findRelatedWisdoms 查找相关智慧
func (s *AIService) findRelatedWisdoms(ctx context.Context, question string, keywords []string) ([]WisdomReference, error) {
	var wisdoms []models.CulturalWisdom
	
	// 构建搜索查询
	query := s.db.Model(&models.CulturalWisdom{})
	
	// 基于关键词搜�?
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
	
	// 限制结果数量
	if err := query.Limit(5).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("failed to find related wisdoms: %w", err)
	}
	
	var references []WisdomReference
	for _, wisdom := range wisdoms {
		// 计算相关度（简单实现）
		relevance := s.calculateQuestionRelevance(question, wisdom, keywords)
		
		// 生成摘录
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

// calculateQuestionRelevance 计算问题相关�?
func (s *AIService) calculateQuestionRelevance(question string, wisdom models.CulturalWisdom, keywords []string) float64 {
	relevance := 0.0
	
	// 基于关键词匹�?
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
	
	// 限制最大相关度
	if relevance > 1.0 {
		relevance = 1.0
	}
	
	return relevance
}

// GetEmbedding 获取文本的向量表�?
func (s *AIService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	provider, err := s.providerManager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("没有可用的AI提供�? %w", err)
	}

	return provider.Embed(ctx, text)
}
