package content

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContentAnalyzer 内容分析器实现
type HelperContentAnalyzer struct {
	config *ContentAnalysisSettings
}

// AnalyzeContent 分析内容
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, contentID string, contentData map[string]interface{}) (*ContentAnalysis, error) {
	analysis := &ContentAnalysis{
		ContentID:  contentID,
		AnalyzedAt: time.Now(),
	}
	
	// 语义特征分析
	if ca.Config.EnableSemanticAnalysis {
		semanticFeatures, err := ca.extractSemanticFeatures(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to extract semantic features: %w", err)
		}
		analysis.SemanticFeatures = semanticFeatures
	}
	
	// 难度分析
	if ca.Config.EnableDifficultyAnalysis {
		difficultyAnalysis, err := ca.analyzeDifficulty(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze difficulty: %w", err)
		}
		analysis.DifficultyAnalysis = difficultyAnalysis
	}
	
	// 主题提取
	if ca.Config.EnableTopicExtraction {
		topicExtraction, err := ca.extractTopics(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to extract topics: %w", err)
		}
		analysis.TopicExtraction = topicExtraction
	}
	
	// 先决条件分析
	if ca.Config.EnablePrerequisiteAnalysis {
		prerequisiteAnalysis, err := ca.analyzePrerequisites(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze prerequisites: %w", err)
		}
		analysis.PrerequisiteAnalysis = prerequisiteAnalysis
	}
	
	// 质量评估
	qualityAssessment, err := ca.assessQuality(contentData)
	if err != nil {
		return nil, fmt.Errorf("failed to assess quality: %w", err)
	}
	analysis.QualityAssessment = qualityAssessment
	
	return analysis, nil
}

// extractSemanticFeatures 提取语义特征
func (ca *ContentAnalyzer) extractSemanticFeatures(contentData map[string]interface{}) (*SemanticFeatures, error) {
	features := &SemanticFeatures{
		Embeddings:         make([]float64, 768), // 假设使用768维嵌入
		Keywords:           make([]string, 0),
		Concepts:           make([]string, 0),
		Entities:           make([]string, 0),
		SemanticSimilarity: make(map[string]float64),
	}
	
	// 简化实现：从内容数据中提取文本
	if text, ok := contentData["text"].(string); ok {
		// 提取关键词
		features.Keywords = ca.extractKeywords(text)
		
		// 提取概念
		features.Concepts = ca.extractConcepts(text)
		
		// 提取实体
		features.Entities = ca.extractEntities(text)
		
		// 生成嵌入向量（简化实现）
		features.Embeddings = ca.generateEmbeddings(text)
	}
	
	return features, nil
}

// extractKeywords 提取关键词
func (ca *ContentAnalyzer) extractKeywords(text string) []string {
	// 简化实现：基于词频的关键词提取
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)
	
	for _, word := range words {
		if len(word) > 3 { // 过滤短词
			wordCount[word]++
		}
	}
	
	// 按频率排序
	type wordFreq struct {
		word  string
		count int
	}
	
	var wordFreqs []wordFreq
	for word, count := range wordCount {
		wordFreqs = append(wordFreqs, wordFreq{word, count})
	}
	
	sort.Slice(wordFreqs, func(i, j int) bool {
		return wordFreqs[i].count > wordFreqs[j].count
	})
	
	// 返回前10个关键词
	keywords := make([]string, 0, 10)
	for i, wf := range wordFreqs {
		if i >= 10 {
			break
		}
		keywords = append(keywords, wf.word)
	}
	
	return keywords
}

// extractConcepts 提取概念
func (ca *ContentAnalyzer) extractConcepts(text string) []string {
	// 简化实现：基于预定义概念列表
	predefinedConcepts := []string{
		"machine learning", "artificial intelligence", "deep learning",
		"neural network", "algorithm", "data science", "programming",
		"mathematics", "statistics", "computer science",
	}
	
	concepts := make([]string, 0)
	textLower := strings.ToLower(text)
	
	for _, concept := range predefinedConcepts {
		if strings.Contains(textLower, concept) {
			concepts = append(concepts, concept)
		}
	}
	
	return concepts
}

// extractEntities 提取实体
func (ca *ContentAnalyzer) extractEntities(text string) []string {
	// 简化实现：基于大写字母开头的词
	words := strings.Fields(text)
	entities := make([]string, 0)
	
	for _, word := range words {
		if len(word) > 1 && word[0] >= 'A' && word[0] <= 'Z' {
			entities = append(entities, word)
		}
	}
	
	return entities
}

// generateEmbeddings 生成嵌入向量
func (ca *ContentAnalyzer) generateEmbeddings(text string) []float64 {
	// 简化实现：生成随机嵌入向量
	embeddings := make([]float64, 768)
	for i := range embeddings {
		embeddings[i] = math.Sin(float64(i) * 0.1) // 简单的确定性生成
	}
	return embeddings
}

// analyzeDifficulty 分析难度
func (ca *ContentAnalyzer) analyzeDifficulty(contentData map[string]interface{}) (*DifficultyAnalysis, error) {
	analysis := &DifficultyAnalysis{
		DifficultyFactors: make(map[string]float64),
	}
	
	if text, ok := contentData["text"].(string); ok {
		// 计算认知难度
		analysis.CognitiveDifficulty = ca.calculateCognitiveDifficulty(text)
		
		// 计算语言难度
		analysis.LinguisticDifficulty = ca.calculateLinguisticDifficulty(text)
		
		// 计算概念难度
		analysis.ConceptualDifficulty = ca.calculateConceptualDifficulty(text)
		
		// 计算总体难度
		analysis.OverallDifficulty = ca.determineOverallDifficulty(
			analysis.CognitiveDifficulty,
			analysis.LinguisticDifficulty,
			analysis.ConceptualDifficulty,
		)
		
		// 设置难度因子
		analysis.DifficultyFactors["cognitive"] = analysis.CognitiveDifficulty
		analysis.DifficultyFactors["linguistic"] = analysis.LinguisticDifficulty
		analysis.DifficultyFactors["conceptual"] = analysis.ConceptualDifficulty
	}
	
	return analysis, nil
}

// calculateCognitiveDifficulty 计算认知难度
func (ca *ContentAnalyzer) calculateCognitiveDifficulty(text string) float64 {
	// 简化实现：基于文本长度和复杂度
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	
	// 认知难度与平均句长相关
	difficulty := math.Min(1.0, avgWordsPerSentence/20.0)
	
	return difficulty
}

// calculateLinguisticDifficulty 计算语言难度
func (ca *ContentAnalyzer) calculateLinguisticDifficulty(text string) float64 {
	// 简化实现：基于词汇复杂度
	words := strings.Fields(text)
	complexWords := 0
	
	for _, word := range words {
		if len(word) > 7 { // 长词被认为更复杂
			complexWords++
		}
	}
	
	difficulty := float64(complexWords) / float64(len(words))
	return math.Min(1.0, difficulty*2.0)
}

// calculateConceptualDifficulty 计算概念难度
func (ca *ContentAnalyzer) calculateConceptualDifficulty(text string) float64 {
	// 简化实现：基于技术术语密度
	technicalTerms := []string{
		"algorithm", "function", "variable", "parameter", "optimization",
		"neural", "network", "learning", "training", "model",
	}
	
	textLower := strings.ToLower(text)
	termCount := 0
	
	for _, term := range technicalTerms {
		if strings.Contains(textLower, term) {
			termCount++
		}
	}
	
	difficulty := float64(termCount) / float64(len(technicalTerms))
	return math.Min(1.0, difficulty*1.5)
}

// determineOverallDifficulty 确定总体难度
func (ca *ContentAnalyzer) determineOverallDifficulty(cognitive, linguistic, conceptual float64) DifficultyLevel {
	overall := (cognitive + linguistic + conceptual) / 3.0
	
	if overall < 0.3 {
		return BeginnerLevel
	} else if overall < 0.6 {
		return IntermediateLevel
	} else if overall < 0.8 {
		return AdvancedLevel
	} else {
		return ExpertLevel
	}
}

// extractTopics 提取主题
func (ca *ContentAnalyzer) extractTopics(contentData map[string]interface{}) (*TopicExtraction, error) {
	extraction := &TopicExtraction{
		MainTopics:     make([]string, 0),
		SubTopics:      make([]string, 0),
		TopicWeights:   make(map[string]float64),
		TopicHierarchy: make(map[string][]string),
	}
	
	if text, ok := contentData["text"].(string); ok {
		// 简化实现：基于关键词聚类
		keywords := ca.extractKeywords(text)
		
		// 主题分类
		topicCategories := map[string][]string{
			"programming": {"code", "function", "variable", "algorithm"},
			"mathematics": {"equation", "formula", "calculation", "number"},
			"science":     {"theory", "experiment", "hypothesis", "research"},
			"technology":  {"computer", "software", "system", "network"},
		}
		
		for topic, relatedWords := range topicCategories {
			weight := 0.0
			for _, keyword := range keywords {
				for _, relatedWord := range relatedWords {
					if strings.Contains(keyword, relatedWord) {
						weight += 1.0
					}
				}
			}
			
			if weight > 0 {
				extraction.MainTopics = append(extraction.MainTopics, topic)
				extraction.TopicWeights[topic] = weight / float64(len(keywords))
			}
		}
	}
	
	return extraction, nil
}

// analyzePrerequisites 分析先决条件
func (ca *ContentAnalyzer) analyzePrerequisites(contentData map[string]interface{}) (*PrerequisiteAnalysis, error) {
	analysis := &PrerequisiteAnalysis{
		RequiredKnowledge:    make([]string, 0),
		RecommendedSkills:    make([]string, 0),
		PrerequisiteConcepts: make([]string, 0),
		DependencyGraph:      make(map[string][]string),
	}
	
	if text, ok := contentData["text"].(string); ok {
		// 简化实现：基于内容复杂度推断先决条件
		difficulty := ca.calculateCognitiveDifficulty(text)
		
		if difficulty > 0.7 {
			analysis.RequiredKnowledge = []string{"advanced mathematics", "programming experience"}
			analysis.RecommendedSkills = []string{"analytical thinking", "problem solving"}
		} else if difficulty > 0.4 {
			analysis.RequiredKnowledge = []string{"basic mathematics", "basic programming"}
			analysis.RecommendedSkills = []string{"logical thinking"}
		} else {
			analysis.RequiredKnowledge = []string{"basic literacy"}
			analysis.RecommendedSkills = []string{"curiosity"}
		}
	}
	
	return analysis, nil
}

// assessQuality 评估质量
func (ca *ContentAnalyzer) assessQuality(contentData map[string]interface{}) (*ContentQualityAssessment, error) {
	assessment := &ContentQualityAssessment{}
	
	if text, ok := contentData["text"].(string); ok {
		// 内容准确性（简化实现）
		assessment.ContentAccuracy = ca.assessAccuracy(text)
		
		// 清晰度
		assessment.Clarity = ca.assessClarity(text)
		
		// 完整性
		assessment.Completeness = ca.assessCompleteness(text)
		
		// 参与度
		assessment.Engagement = ca.assessEngagement(text)
		
		// 新鲜度
		assessment.Freshness = ca.assessFreshness(contentData)
		
		// 总体质量
		assessment.OverallQuality = (assessment.ContentAccuracy + assessment.Clarity + 
			assessment.Completeness + assessment.Engagement + assessment.Freshness) / 5.0
	}
	
	return assessment, nil
}

// assessAccuracy 评估准确性
func (ca *ContentAnalyzer) assessAccuracy(text string) float64 {
	// 简化实现：基于文本结构和一致性
	sentences := strings.Split(text, ".")
	if len(sentences) < 3 {
		return 0.5 // 内容太短，难以评估
	}
	
	// 检查是否有矛盾表述（简化检查）
	contradictionWords := []string{"but", "however", "although", "despite"}
	contradictions := 0
	
	for _, sentence := range sentences {
		for _, word := range contradictionWords {
			if strings.Contains(strings.ToLower(sentence), word) {
				contradictions++
			}
		}
	}
	
	// 适度的矛盾表述是正常的，过多可能表示不准确
	accuracy := 1.0 - math.Min(0.5, float64(contradictions)/float64(len(sentences)))
	return accuracy
}

// assessClarity 评估清晰度
func (ca *ContentAnalyzer) assessClarity(text string) float64 {
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	
	if len(sentences) == 0 {
		return 0.0
	}
	
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	
	// 理想的句长是10-20词
	clarity := 1.0 - math.Abs(avgWordsPerSentence-15.0)/15.0
	return math.Max(0.0, math.Min(1.0, clarity))
}

// assessCompleteness 评估完整性
func (ca *ContentAnalyzer) assessCompleteness(text string) float64 {
	// 简化实现：基于文本长度和结构
	words := strings.Fields(text)
	
	if len(words) < 50 {
		return 0.3 // 太短
	} else if len(words) < 200 {
		return 0.6 // 中等
	} else if len(words) < 500 {
		return 0.8 // 较完整
	} else {
		return 1.0 // 完整
	}
}

// assessEngagement 评估参与度
func (ca *ContentAnalyzer) assessEngagement(text string) float64 {
	// 简化实现：基于问句和互动元素
	questions := strings.Count(text, "?")
	exclamations := strings.Count(text, "!")
	sentences := len(strings.Split(text, "."))
	
	if sentences == 0 {
		return 0.0
	}
	
	interactiveRatio := float64(questions+exclamations) / float64(sentences)
	engagement := math.Min(1.0, interactiveRatio*3.0)
	
	return engagement
}

// assessFreshness 评估新鲜度
func (ca *ContentAnalyzer) assessFreshness(contentData map[string]interface{}) float64 {
	// 简化实现：基于创建时间
	if createdAt, ok := contentData["created_at"].(time.Time); ok {
		age := time.Since(createdAt)
		
		if age < time.Hour*24*30 { // 30天内
			return 1.0
		} else if age < time.Hour*24*90 { // 90天内
			return 0.8
		} else if age < time.Hour*24*365 { // 1年内
			return 0.6
		} else {
			return 0.3
		}
	}
	
	return 0.5 // 未知创建时间
}

// LearnerProfiler 学习者画像器实现
type HelperLearnerProfiler struct {
	Config *LearnerProfilingSettings
}

// BuildProfile 构建学习者画像
func (lp *LearnerProfiler) BuildProfile(ctx context.Context, learnerID string, learningHistory []map[string]interface{}) (*LearnerProfile, error) {
	profile := &LearnerProfile{
		LearnerID: learnerID,
		UpdatedAt: time.Now(),
	}
	
	// 分析学习偏好
	if lp.Config.EnablePreferenceAnalysis {
		preferences, err := lp.analyzeLearningPreferences(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze preferences: %w", err)
		}
		profile.LearningPreferences = preferences
	}
	
	// 分析学习行为
	if lp.Config.EnableBehaviorAnalysis {
		behavior, err := lp.analyzeLearningBehavior(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze behavior: %w", err)
		}
		profile.LearningBehavior = behavior
	}
	
	// 分析性能表现
	if lp.Config.EnablePerformanceAnalysis {
		performance, err := lp.analyzePerformanceProfile(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze performance: %w", err)
		}
		profile.PerformanceProfile = performance
	}
	
	// 分析学习风格
	if lp.Config.EnableLearningStyleAnalysis {
		style, err := lp.analyzeLearningStyle(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze learning style: %w", err)
		}
		profile.LearningStyle = style
	}
	
	// 构建兴趣画像
	interests, err := lp.buildInterestProfile(learningHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to build interest profile: %w", err)
	}
	profile.InterestProfile = interests
	
	// 构建知识状态
	knowledge, err := lp.buildKnowledgeState(learningHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to build knowledge state: %w", err)
	}
	profile.KnowledgeState = knowledge
	
	return profile, nil
}

// analyzeLearningPreferences 分析学习偏好
func (lp *LearnerProfiler) analyzeLearningPreferences(history []map[string]interface{}) (*LearningPreferences, error) {
	preferences := &LearningPreferences{
		PreferredContentTypes: make([]ContentType, 0),
		PreferredSubjects:     make([]string, 0),
		PreferredLanguages:    make([]string, 0),
		PreferredFormats:      make([]string, 0),
	}
	
	// 统计内容类型偏好
	contentTypeCount := make(map[ContentType]int)
	subjectCount := make(map[string]int)
	var totalDuration time.Duration
	var sessionCount int
	
	for _, record := range history {
		if contentType, ok := record["content_type"].(string); ok {
			contentTypeCount[ContentType(contentType)]++
		}
		
		if subject, ok := record["subject"].(string); ok {
			subjectCount[subject]++
		}
		
		if duration, ok := record["duration"].(time.Duration); ok {
			totalDuration += duration
			sessionCount++
		}
	}
	
	// 确定偏好的内容类型
	for contentType, count := range contentTypeCount {
		if count > len(history)/4 { // 超过25%的使用率
			preferences.PreferredContentTypes = append(preferences.PreferredContentTypes, contentType)
		}
	}
	
	// 确定偏好的学科
	for subject, count := range subjectCount {
		if count > len(history)/5 { // 超过20%的使用率
			preferences.PreferredSubjects = append(preferences.PreferredSubjects, subject)
		}
	}
	
	// 计算偏好的学习时长
	if sessionCount > 0 {
		preferences.PreferredDuration = totalDuration / time.Duration(sessionCount)
	}
	
	// 默认偏好设置
	if len(preferences.PreferredContentTypes) == 0 {
		preferences.PreferredContentTypes = []ContentType{VideoContent, TextContent}
	}
	
	preferences.PreferredDifficulty = IntermediateLevel
	preferences.PreferredLanguages = []string{"zh", "en"}
	preferences.PreferredFormats = []string{"interactive", "multimedia"}
	
	return preferences, nil
}

// analyzeLearningBehavior 分析学习行为
func (lp *LearnerProfiler) analyzeLearningBehavior(history []map[string]interface{}) (*LearningBehavior, error) {
	behavior := &LearningBehavior{}
	
	// 分析学习模式
	studyPatterns, err := lp.analyzeStudyPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.StudyPatterns = studyPatterns
	
	// 分析参与模式
	engagementPatterns, err := lp.analyzeEngagementPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.EngagementPatterns = engagementPatterns
	
	// 分析进度模式
	progressPatterns, err := lp.analyzeProgressPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.ProgressPatterns = progressPatterns
	
	// 分析交互模式
	interactionPatterns, err := lp.analyzeInteractionPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.InteractionPatterns = interactionPatterns
	
	return behavior, nil
}

// analyzeStudyPatterns 分析学习模式
func (lp *LearnerProfiler) analyzeStudyPatterns(history []map[string]interface{}) (*StudyPatterns, error) {
	patterns := &StudyPatterns{
		PreferredStudyTimes: make([]time.Time, 0),
		BreakPatterns:       make([]time.Duration, 0),
	}
	
	var totalDuration time.Duration
	var sessionCount int
	studyTimes := make(map[int]int) // 小时 -> 次数
	
	for _, record := range history {
		if startTime, ok := record["start_time"].(time.Time); ok {
			hour := startTime.Hour()
			studyTimes[hour]++
		}
		
		if duration, ok := record["duration"].(time.Duration); ok {
			totalDuration += duration
			sessionCount++
		}
	}
	
	// 计算平均学习时长
	if sessionCount > 0 {
		patterns.AverageSessionDuration = totalDuration / time.Duration(sessionCount)
	}
	
	// 计算学习频率（每天平均学习次数）
	if len(history) > 0 {
		patterns.StudyFrequency = float64(len(history)) / 30.0 // 假设30天的历史
	}
	
	// 找出偏好的学习时间
	maxCount := 0
	preferredHour := 0
	for hour, count := range studyTimes {
		if count > maxCount {
			maxCount = count
			preferredHour = hour
		}
	}
	
	// 设置偏好学习时间
	preferredTime := time.Date(2024, 1, 1, preferredHour, 0, 0, 0, time.UTC)
	patterns.PreferredStudyTimes = []time.Time{preferredTime}
	
	return patterns, nil
}

// analyzeEngagementPatterns 分析参与模式
func (lp *LearnerProfiler) analyzeEngagementPatterns(history []map[string]interface{}) (*EngagementPatterns, error) {
	patterns := &EngagementPatterns{}
	
	var totalEngagement float64
	var totalAttention time.Duration
	var totalInteractions int
	var recordCount int
	
	for _, record := range history {
		if engagement, ok := record["engagement_score"].(float64); ok {
			totalEngagement += engagement
			recordCount++
		}
		
		if attention, ok := record["attention_span"].(time.Duration); ok {
			totalAttention += attention
		}
		
		if interactions, ok := record["interaction_count"].(int); ok {
			totalInteractions += interactions
		}
	}
	
	// 计算平均参与度
	if recordCount > 0 {
		patterns.EngagementLevel = totalEngagement / float64(recordCount)
	}
	
	// 计算平均注意力持续时间
	if recordCount > 0 {
		patterns.AttentionSpan = totalAttention / time.Duration(recordCount)
	}
	
	// 计算交互频率
	if recordCount > 0 {
		patterns.InteractionFrequency = float64(totalInteractions) / float64(recordCount)
	}
	
	// 设置默认值
	if patterns.EngagementLevel == 0 {
		patterns.EngagementLevel = 0.7
	}
	if patterns.AttentionSpan == 0 {
		patterns.AttentionSpan = time.Minute * 15
	}
	if patterns.InteractionFrequency == 0 {
		patterns.InteractionFrequency = 5.0
	}
	
	patterns.FeedbackResponsiveness = 0.8 // 默认值
	
	return patterns, nil
}

// analyzeProgressPatterns 分析进度模式
func (lp *LearnerProfiler) analyzeProgressPatterns(history []map[string]interface{}) (*ProgressPatterns, error) {
	patterns := &ProgressPatterns{}
	
	var completedCount int
	var totalCount int
	var totalProgress float64
	var progressCount int
	
	for _, record := range history {
		totalCount++
		
		if completed, ok := record["completed"].(bool); ok && completed {
			completedCount++
		}
		
		if progress, ok := record["progress"].(float64); ok {
			totalProgress += progress
			progressCount++
		}
	}
	
	// 计算完成率
	if totalCount > 0 {
		patterns.CompletionRate = float64(completedCount) / float64(totalCount)
	}
	
	// 计算学习速度（简化实现）
	if progressCount > 0 {
		avgProgress := totalProgress / float64(progressCount)
		patterns.LearningVelocity = avgProgress
	}
	
	// 设置默认值
	if patterns.LearningVelocity == 0 {
		patterns.LearningVelocity = 0.6
	}
	if patterns.CompletionRate == 0 {
		patterns.CompletionRate = 0.7
	}
	
	patterns.RetentionRate = 0.8  // 默认保持率
	patterns.MasteryRate = 0.75   // 默认掌握率
	
	return patterns, nil
}

// analyzeInteractionPatterns 分析交互模式
func (lp *LearnerProfiler) analyzeInteractionPatterns(history []map[string]interface{}) (*InteractionPatterns, error) {
	patterns := &InteractionPatterns{
		PreferredInteractionTypes: make([]string, 0),
	}
	
	interactionTypes := make(map[string]int)
	var totalResponseTime time.Duration
	var responseCount int
	
	for _, record := range history {
		if interactionType, ok := record["interaction_type"].(string); ok {
			interactionTypes[interactionType]++
		}
		
		if responseTime, ok := record["response_time"].(time.Duration); ok {
			totalResponseTime += responseTime
			responseCount++
		}
	}
	
	// 确定偏好的交互类型
	for interactionType, count := range interactionTypes {
		if count > len(history)/5 { // 超过20%的使用率
			patterns.PreferredInteractionTypes = append(patterns.PreferredInteractionTypes, interactionType)
		}
	}
	
	// 计算平均响应时间
	if responseCount > 0 {
		patterns.ResponseTime = totalResponseTime / time.Duration(responseCount)
	}
	
	// 设置默认值
	if len(patterns.PreferredInteractionTypes) == 0 {
		patterns.PreferredInteractionTypes = []string{"click", "scroll", "type"}
	}
	if patterns.ResponseTime == 0 {
		patterns.ResponseTime = time.Second * 3
	}
	
	patterns.HelpSeekingBehavior = 0.6      // 默认求助行为
	patterns.CollaborationPreference = 0.5  // 默认协作偏好
	
	return patterns, nil
}

// analyzePerformanceProfile 分析性能表现
func (lp *LearnerProfiler) analyzePerformanceProfile(history []map[string]interface{}) (*PerformanceProfile, error) {
	profile := &PerformanceProfile{
		SubjectPerformance: make(map[string]float64),
		SkillLevels:        make(map[string]float64),
		StrengthAreas:      make([]string, 0),
		ImprovementAreas:   make([]string, 0),
	}
	
	subjectScores := make(map[string][]float64)
	var totalScore float64
	var scoreCount int
	
	for _, record := range history {
		if subject, ok := record["subject"].(string); ok {
			if score, ok := record["score"].(float64); ok {
				subjectScores[subject] = append(subjectScores[subject], score)
				totalScore += score
				scoreCount++
			}
		}
	}
	
	// 计算总体性能
	if scoreCount > 0 {
		profile.OverallPerformance = totalScore / float64(scoreCount)
	}
	
	// 计算各学科性能
	for subject, scores := range subjectScores {
		var sum float64
		for _, score := range scores {
			sum += score
		}
		avgScore := sum / float64(len(scores))
		profile.SubjectPerformance[subject] = avgScore
		
		// 确定优势和改进领域
		if avgScore > 0.8 {
			profile.StrengthAreas = append(profile.StrengthAreas, subject)
		} else if avgScore < 0.6 {
			profile.ImprovementAreas = append(profile.ImprovementAreas, subject)
		}
	}
	
	// 设置默认技能水平
	profile.SkillLevels["problem_solving"] = 0.7
	profile.SkillLevels["critical_thinking"] = 0.6
	profile.SkillLevels["creativity"] = 0.8
	
	// 计算学习效率
	profile.LearningEfficiency = profile.OverallPerformance * 0.8 // 简化计算
	
	return profile, nil
}

// analyzeLearningStyle 分析学习风格
func (lp *LearnerProfiler) analyzeLearningStyle(history []map[string]interface{}) (*LearningStyle, error) {
	style := &LearningStyle{}
	
	contentTypePreferences := make(map[string]int)
	
	for _, record := range history {
		if contentType, ok := record["content_type"].(string); ok {
			contentTypePreferences[contentType]++
		}
	}
	
	total := len(history)
	if total == 0 {
		// 设置默认学习风格
		style.VisualLearning = 0.7
		style.AuditoryLearning = 0.5
		style.KinestheticLearning = 0.6
		style.ReadingWritingLearning = 0.8
		style.SequentialLearning = 0.6
		style.GlobalLearning = 0.4
		return style, nil
	}
	
	// 基于内容类型偏好推断学习风格
	style.VisualLearning = float64(contentTypePreferences["video"]+contentTypePreferences["image"]) / float64(total)
	style.AuditoryLearning = float64(contentTypePreferences["audio"]+contentTypePreferences["podcast"]) / float64(total)
	style.KinestheticLearning = float64(contentTypePreferences["interactive"]+contentTypePreferences["exercise"]) / float64(total)
	style.ReadingWritingLearning = float64(contentTypePreferences["text"]+contentTypePreferences["document"]) / float64(total)
	
	// 设置默认值
	if style.VisualLearning == 0 {
		style.VisualLearning = 0.5
	}
	if style.AuditoryLearning == 0 {
		style.AuditoryLearning = 0.3
	}
	if style.KinestheticLearning == 0 {
		style.KinestheticLearning = 0.4
	}
	if style.ReadingWritingLearning == 0 {
		style.ReadingWritingLearning = 0.6
	}
	
	style.SequentialLearning = 0.6 // 默认值
	style.GlobalLearning = 0.4     // 默认值
	
	return style, nil
}

// buildInterestProfile 构建兴趣画像
func (lp *LearnerProfiler) buildInterestProfile(history []map[string]interface{}) (*InterestProfile, error) {
	profile := &InterestProfile{
		TopicInterests:    make(map[string]float64),
		SubjectInterests:  make(map[string]float64),
		ActivityInterests: make(map[string]float64),
		InterestTrends:    make(map[string][]float64),
	}
	
	topicCount := make(map[string]int)
	subjectCount := make(map[string]int)
	activityCount := make(map[string]int)
	
	for _, record := range history {
		if topic, ok := record["topic"].(string); ok {
			topicCount[topic]++
		}
		
		if subject, ok := record["subject"].(string); ok {
			subjectCount[subject]++
		}
		
		if activity, ok := record["activity_type"].(string); ok {
			activityCount[activity]++
		}
	}
	
	total := len(history)
	if total == 0 {
		return profile, nil
	}
	
	// 计算兴趣分数
	for topic, count := range topicCount {
		profile.TopicInterests[topic] = float64(count) / float64(total)
	}
	
	for subject, count := range subjectCount {
		profile.SubjectInterests[subject] = float64(count) / float64(total)
	}
	
	for activity, count := range activityCount {
		profile.ActivityInterests[activity] = float64(count) / float64(total)
	}
	
	return profile, nil
}

// buildKnowledgeState 构建知识状态
func (lp *LearnerProfiler) buildKnowledgeState(history []map[string]interface{}) (*KnowledgeState, error) {
	state := &KnowledgeState{
		MasteredConcepts:  make([]string, 0),
		LearningConcepts:  make([]string, 0),
		ConceptMastery:    make(map[string]float64),
		KnowledgeGaps:     make([]string, 0),
		LearningGoals:     make([]string, 0),
	}
	
	conceptProgress := make(map[string][]float64)
	
	for _, record := range history {
		if concept, ok := record["concept"].(string); ok {
			if progress, ok := record["mastery_level"].(float64); ok {
				conceptProgress[concept] = append(conceptProgress[concept], progress)
			}
		}
	}
	
	// 分析概念掌握情况
	for concept, progressList := range conceptProgress {
		if len(progressList) == 0 {
			continue
		}
		
		// 计算平均掌握度
		var sum float64
		for _, progress := range progressList {
			sum += progress
		}
		avgMastery := sum / float64(len(progressList))
		state.ConceptMastery[concept] = avgMastery
		
		// 分类概念
		if avgMastery >= 0.8 {
			state.MasteredConcepts = append(state.MasteredConcepts, concept)
		} else if avgMastery >= 0.4 {
			state.LearningConcepts = append(state.LearningConcepts, concept)
		} else {
			state.KnowledgeGaps = append(state.KnowledgeGaps, concept)
		}
	}
	
	// 设置学习目标（简化实现）
	state.LearningGoals = []string{
		"improve weak areas",
		"master current concepts",
		"explore new topics",
	}
	
	return state, nil
}

// RecommendationAlgorithm 推荐算法接口
type RecommendationAlgorithm interface {
	GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error)
	GetAlgorithmName() string
	GetAlgorithmType() RecommendationStrategy
}

// HelperRecommendationAlgorithm 推荐算法接口
type HelperRecommendationAlgorithm interface {
	GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error)
	GetAlgorithmName() string
	GetAlgorithmType() RecommendationStrategy
}

// CollaborativeFilteringAlgorithm 协同过滤算法
type CollaborativeFilteringAlgorithm struct {
	config *HelperCollaborativeFilteringConfig
}

// GenerateRecommendations 生成推荐
func (cf *CollaborativeFilteringAlgorithm) GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error) {
	recommendations := make([]*ContentRecommendation, 0)
	
	// 简化的协同过滤实现
	for _, content := range availableContent {
		score := cf.calculateCollaborativeScore(profile, &content)
		
		if score > config.MinConfidenceScore {
			recommendation := &ContentRecommendation{
				RecommendationID: uuid.New().String(),
				ContentID:        content.ContentID,
				LearnerID:        profile.LearnerID,
				ConfidenceScore:  score,
				RecommendationStrategy: CollaborativeFiltering,
				GeneratedAt:      time.Now(),
				Reasons: []*RecommendationReason{
					{
						Type:        "collaborative_similarity",
						Description: "Similar learners also engaged with this content",
						Weight:      score,
						Evidence: []*RecommendationEvidence{
							{
								Type:        "user_similarity",
								Value:       score,
								Description: "Based on similar learning patterns",
							},
						},
					},
				},
			}
			recommendations = append(recommendations, recommendation)
		}
	}
	
	// 按置信度排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].ConfidenceScore > recommendations[j].ConfidenceScore
	})
	
	// 限制返回数量
	if len(recommendations) > config.MaxRecommendations {
		recommendations = recommendations[:config.MaxRecommendations]
	}
	
	return recommendations, nil
}

// calculateCollaborativeScore 计算协同过滤分数
func (cf *CollaborativeFilteringAlgorithm) calculateCollaborativeScore(profile *LearnerProfile, content *ContentAnalysis) float64 {
	// 简化实现：基于学习偏好匹配
	score := 0.0
	
	// 检查内容类型匹配
	for _, preferredType := range profile.LearningPreferences.PreferredContentTypes {
		if string(preferredType) == content.ContentID { // 简化匹配
			score += 0.3
		}
	}
	
	// 检查难度匹配
	if content.DifficultyAnalysis != nil {
		difficultyMatch := 1.0 - math.Abs(DifficultyLevelToFloat64(profile.LearningPreferences.PreferredDifficulty)-DifficultyLevelToFloat64(content.DifficultyAnalysis.OverallDifficulty))
		score += difficultyMatch * 0.4
	}
	
	// 检查主题匹配
	if content.TopicExtraction != nil {
		for _, topic := range content.TopicExtraction.MainTopics {
			if interest, exists := profile.InterestProfile.TopicInterests[topic]; exists {
				score += interest * 0.3
			}
		}
	}
	
	return math.Min(1.0, score)
}

// GetAlgorithmName 获取算法名称
func (cf *CollaborativeFilteringAlgorithm) GetAlgorithmName() string {
	return "Collaborative Filtering"
}

// GetAlgorithmType 获取算法类型
func (cf *CollaborativeFilteringAlgorithm) GetAlgorithmType() RecommendationStrategy {
	return CollaborativeFiltering
}

// ContentBasedAlgorithm 基于内容的推荐算法
type ContentBasedAlgorithm struct {
	config *ContentBasedConfig
}

// GenerateRecommendations 生成推荐
func (cb *ContentBasedAlgorithm) GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error) {
	recommendations := make([]*ContentRecommendation, 0)
	
	for _, content := range availableContent {
		score := cb.calculateContentBasedScore(profile, &content)
		
		if score > config.MinConfidenceScore {
			recommendation := &ContentRecommendation{
				RecommendationID: uuid.New().String(),
				ContentID:        content.ContentID,
				LearnerID:        profile.LearnerID,
				ConfidenceScore:  score,
				RecommendationStrategy: ContentBased,
				GeneratedAt:      time.Now(),
				Reasons: []*RecommendationReason{
					{
						Type:        "content_similarity",
						Description: "Content matches your learning preferences and interests",
						Weight:      score,
						Evidence: []*RecommendationEvidence{
							{
								Type:        "content_match",
								Value:       score,
								Description: "Based on content analysis and your profile",
							},
						},
					},
				},
			}
			recommendations = append(recommendations, recommendation)
		}
	}
	
	// 按置信度排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].ConfidenceScore > recommendations[j].ConfidenceScore
	})
	
	// 限制返回数量
	if len(recommendations) > config.MaxRecommendations {
		recommendations = recommendations[:config.MaxRecommendations]
	}
	
	return recommendations, nil
}

// calculateContentBasedScore 计算基于内容的分数
func (cb *ContentBasedAlgorithm) calculateContentBasedScore(profile *LearnerProfile, content *ContentAnalysis) float64 {
	score := 0.0
	
	// 语义相似度
	if content.SemanticFeatures != nil {
		semanticScore := cb.calculateSemanticSimilarity(profile, content.SemanticFeatures)
		score += semanticScore * 0.4
	}
	
	// 难度匹配
	if content.DifficultyAnalysis != nil {
		difficultyScore := cb.calculateDifficultyMatch(profile, content.DifficultyAnalysis)
		score += difficultyScore * 0.3
	}
	
	// 主题兴趣匹配
	if content.TopicExtraction != nil {
		topicScore := cb.calculateTopicMatch(profile, content.TopicExtraction)
		score += topicScore * 0.3
	}
	
	return math.Min(1.0, score)
}

// calculateSemanticSimilarity 计算语义相似度
func (cb *ContentBasedAlgorithm) calculateSemanticSimilarity(profile *LearnerProfile, features *SemanticFeatures) float64 {
	// 简化实现：基于关键词匹配
	matchCount := 0
	totalKeywords := len(features.Keywords)
	
	if totalKeywords == 0 {
		return 0.5 // 默认中等相似度
	}
	
	// 检查关键词是否与学习者兴趣匹配
	for _, keyword := range features.Keywords {
		for topic := range profile.InterestProfile.TopicInterests {
			if strings.Contains(strings.ToLower(topic), strings.ToLower(keyword)) {
				matchCount++
				break
			}
		}
	}
	
	return float64(matchCount) / float64(totalKeywords)
}

// calculateDifficultyMatch 计算难度匹配度
func (cb *ContentBasedAlgorithm) calculateDifficultyMatch(profile *LearnerProfile, difficulty *DifficultyAnalysis) float64 {
	preferredDifficulty := DifficultyLevelToFloat64(profile.LearningPreferences.PreferredDifficulty)
	contentDifficulty := DifficultyLevelToFloat64(difficulty.OverallDifficulty)
	
	// 计算难度差异
	difficultyDiff := math.Abs(preferredDifficulty - contentDifficulty)
	
	// 转换为匹配分数（差异越小，匹配度越高）
	matchScore := 1.0 - (difficultyDiff / 3.0) // 假设难度范围是0-3
	
	return math.Max(0.0, matchScore)
}

// calculateTopicMatch 计算主题匹配度
func (cb *ContentBasedAlgorithm) calculateTopicMatch(profile *LearnerProfile, topics *TopicExtraction) float64 {
	if len(topics.MainTopics) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, topic := range topics.MainTopics {
		if interest, exists := profile.InterestProfile.TopicInterests[topic]; exists {
			weight := topics.TopicWeights[topic]
			totalScore += interest * weight
		}
	}
	
	return totalScore / float64(len(topics.MainTopics))
}

// GetAlgorithmName 获取算法名称
func (cb *ContentBasedAlgorithm) GetAlgorithmName() string {
	return "Content-Based Filtering"
}

// GetAlgorithmType 获取算法类型
func (cb *ContentBasedAlgorithm) GetAlgorithmType() RecommendationStrategy {
	return ContentBased
}

// HybridAlgorithm 混合推荐算法
type HybridAlgorithm struct {
	collaborativeAlgorithm *CollaborativeFilteringAlgorithm
	contentBasedAlgorithm  *ContentBasedAlgorithm
	config                 *HybridConfig
}

// GenerateRecommendations 生成推荐
func (h *HybridAlgorithm) GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error) {
	// 获取协同过滤推荐
	collaborativeRecs, err := h.collaborativeAlgorithm.GenerateRecommendations(ctx, profile, availableContent, config)
	if err != nil {
		return nil, fmt.Errorf("collaborative filtering failed: %w", err)
	}
	
	// 获取基于内容的推荐
	contentBasedRecs, err := h.contentBasedAlgorithm.GenerateRecommendations(ctx, profile, availableContent, config)
	if err != nil {
		return nil, fmt.Errorf("content-based filtering failed: %w", err)
	}
	
	// 合并和重新评分
	hybridRecs := h.combineRecommendations(collaborativeRecs, contentBasedRecs)
	
	// 按置信度排序
	sort.Slice(hybridRecs, func(i, j int) bool {
		return hybridRecs[i].ConfidenceScore > hybridRecs[j].ConfidenceScore
	})
	
	// 限制返回数量
	if len(hybridRecs) > config.MaxRecommendations {
		hybridRecs = hybridRecs[:config.MaxRecommendations]
	}
	
	return hybridRecs, nil
}

// combineRecommendations 合并推荐结果
func (h *HybridAlgorithm) combineRecommendations(collaborativeRecs, contentBasedRecs []*ContentRecommendation) []*ContentRecommendation {
	recommendationMap := make(map[string]*ContentRecommendation)
	
	// 添加协同过滤推荐
	for _, rec := range collaborativeRecs {
		rec.ConfidenceScore *= h.config.CollaborativeWeight
		rec.RecommendationStrategy = HybridApproach
		recommendationMap[rec.ContentID] = rec
	}
	
	// 添加或合并基于内容的推荐
	for _, rec := range contentBasedRecs {
		rec.ConfidenceScore *= h.config.ContentBasedWeight
		
		if existingRec, exists := recommendationMap[rec.ContentID]; exists {
			// 合并分数
			existingRec.ConfidenceScore += rec.ConfidenceScore
			existingRec.ConfidenceScore = math.Min(1.0, existingRec.ConfidenceScore)
			
			// 合并原因
			existingRec.Reasons = append(existingRec.Reasons, rec.Reasons...)
		} else {
			rec.RecommendationStrategy = HybridApproach
			recommendationMap[rec.ContentID] = rec
		}
	}
	
	// 转换为切片
	result := make([]*ContentRecommendation, 0, len(recommendationMap))
	for _, rec := range recommendationMap {
		result = append(result, rec)
	}
	
	return result
}

// GetAlgorithmName 获取算法名称
func (h *HybridAlgorithm) GetAlgorithmName() string {
	return "Hybrid Recommendation"
}

// GetAlgorithmType 获取算法类型
func (h *HybridAlgorithm) GetAlgorithmType() RecommendationStrategy {
	return HybridApproach
}

// RecommendationEngine 推荐引擎
type HelperRecommendationEngine struct {
	algorithms map[RecommendationStrategy]HelperRecommendationAlgorithm
	config     *RecommendationEngineConfig
}

// NewRecommendationEngine 创建推荐引擎
func NewHelperRecommendationEngine(config *RecommendationEngineConfig) *HelperRecommendationEngine {
	engine := &HelperRecommendationEngine{
		algorithms: make(map[RecommendationStrategy]HelperRecommendationAlgorithm),
		config:     config,
	}
	
	// 注册算法
	engine.registerAlgorithms()
	
	return engine
}

// registerAlgorithms 注册算法
func (re *HelperRecommendationEngine) registerAlgorithms() {
	// 协同过滤算法
	collaborativeAlgorithm := &CollaborativeFilteringAlgorithm{
		config: &HelperCollaborativeFilteringConfig{
			MinSimilarUsers:    5,
			SimilarityThreshold: 0.7,
			UserSimilarityMethod: "cosine",
		},
	}
	re.algorithms[CollaborativeFiltering] = collaborativeAlgorithm
	
	// 基于内容的算法
	contentBasedAlgorithm := &ContentBasedAlgorithm{
		config: &ContentBasedConfig{
			FeatureWeights: map[string]float64{
				"semantic":    0.4,
				"difficulty":  0.3,
				"topic":       0.3,
			},
			SimilarityThreshold: 0.6,
		},
	}
	re.algorithms[ContentBased] = contentBasedAlgorithm
	
	// 混合算法
	hybridAlgorithm := &HybridAlgorithm{
		collaborativeAlgorithm: collaborativeAlgorithm,
		contentBasedAlgorithm:  contentBasedAlgorithm,
		config: &HybridConfig{
			CollaborativeWeight: 0.6,
			ContentBasedWeight:  0.4,
			CombinationMethod:   "weighted_sum",
		},
	}
	re.algorithms[HybridApproach] = hybridAlgorithm
}

// registerAlgorithms 注册算法 (原有的RecommendationEngine方法)
func (re *RecommendationEngine) registerAlgorithms() {
	// 协同过滤算法
	collaborativeAlgorithm := &CollaborativeFilteringAlgorithm{
		config: &HelperCollaborativeFilteringConfig{
			MinSimilarUsers:    5,
			SimilarityThreshold: 0.7,
			UserSimilarityMethod: "cosine",
		},
	}
	re.algorithms[CollaborativeFiltering] = collaborativeAlgorithm
	
	// 基于内容的算法
	contentBasedAlgorithm := &ContentBasedAlgorithm{
		config: &ContentBasedConfig{
			FeatureWeights: map[string]float64{
				"semantic":    0.4,
				"difficulty":  0.3,
				"topic":       0.3,
			},
			SimilarityThreshold: 0.6,
		},
	}
	re.algorithms[ContentBased] = contentBasedAlgorithm
	
	// 混合算法
	hybridAlgorithm := &HybridAlgorithm{
		collaborativeAlgorithm: collaborativeAlgorithm,
		contentBasedAlgorithm:  contentBasedAlgorithm,
		config: &HybridConfig{
			CollaborativeWeight: 0.6,
			ContentBasedWeight:  0.4,
			CombinationMethod:   "weighted_sum",
		},
	}
	re.algorithms[HybridApproach] = hybridAlgorithm
}

// GenerateRecommendations 生成推荐
func (re *RecommendationEngine) GenerateRecommendations(ctx context.Context, strategy RecommendationStrategy, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error) {
	algorithmInterface, exists := re.algorithms[strategy]
	if !exists {
		return nil, fmt.Errorf("unsupported recommendation strategy: %v", strategy)
	}
	
	algorithm, ok := algorithmInterface.(RecommendationAlgorithm)
	if !ok {
		return nil, fmt.Errorf("algorithm does not implement RecommendationAlgorithm interface")
	}
	
	return algorithm.GenerateRecommendations(ctx, profile, availableContent, config)
}

// GetAvailableStrategies 获取可用策略
func (re *RecommendationEngine) GetAvailableStrategies() []RecommendationStrategy {
	strategies := make([]RecommendationStrategy, 0, len(re.algorithms))
	for strategy := range re.algorithms {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// 配置结构体
type HelperCollaborativeFilteringConfig struct {
	MinSimilarUsers      int
	SimilarityThreshold  float64
	UserSimilarityMethod string
}

type HelperContentBasedConfig struct {
	FeatureWeights      map[string]float64
	SimilarityThreshold float64
}

type HelperHybridConfig struct {
	CollaborativeWeight float64
	ContentBasedWeight  float64
	CombinationMethod   string
}

type CollaborativeFilteringConfig struct {
	MinSimilarUsers      int
	SimilarityThreshold  float64
	UserSimilarityMethod string
}

type ContentBasedConfig struct {
	FeatureWeights      map[string]float64
	SimilarityThreshold float64
}

type HybridConfig struct {
	CollaborativeWeight float64
	ContentBasedWeight  float64
	CombinationMethod   string
}

type RecommendationEngineConfig struct {
	DefaultStrategy     RecommendationStrategy
	FallbackStrategy    RecommendationStrategy
	MaxRecommendations  int
	MinConfidenceScore  float64
	EnableCaching       bool
	CacheExpiration     time.Duration
}

// RecommendationPersonalizationEngine 推荐个性化引擎
type RecommendationPersonalizationEngine struct {
	config *PersonalizationEngineConfig
}

// NewRecommendationPersonalizationEngine 创建推荐个性化引擎
func NewRecommendationPersonalizationEngine(config *PersonalizationEngineConfig) *RecommendationPersonalizationEngine {
	return &RecommendationPersonalizationEngine{
		config: config,
	}
}

// PersonalizeRecommendations 个性化推荐
func (pe *RecommendationPersonalizationEngine) PersonalizeRecommendations(ctx context.Context, recommendations []*ContentRecommendation, profile *LearnerProfile) ([]*ContentRecommendation, error) {
	personalizedRecs := make([]*ContentRecommendation, len(recommendations))
	copy(personalizedRecs, recommendations)
	
	for _, rec := range personalizedRecs {
		// 应用个性化调整
		pe.applyPersonalizationFactors(rec, profile)
		
		// 添加个性化原因
		pe.addPersonalizationReasons(rec, profile)
		
		// 调整置信度
		pe.adjustConfidenceScore(rec, profile)
	}
	
	// 重新排序
	sort.Slice(personalizedRecs, func(i, j int) bool {
		return personalizedRecs[i].ConfidenceScore > personalizedRecs[j].ConfidenceScore
	})
	
	return personalizedRecs, nil
}

// applyPersonalizationFactors 应用个性化因子
func (pe *RecommendationPersonalizationEngine) applyPersonalizationFactors(rec *ContentRecommendation, profile *LearnerProfile) {
	// 基于学习风格调整
	if pe.config.EnableLearningStylePersonalization {
		pe.adjustForLearningStyle(rec, profile.LearningStyle)
	}
	
	// 基于性能表现调整
	if pe.config.EnablePerformancePersonalization {
		pe.adjustForPerformance(rec, profile.PerformanceProfile)
	}
	
	// 基于学习行为调整
	if pe.config.EnableBehaviorPersonalization {
		pe.adjustForBehavior(rec, profile.LearningBehavior)
	}
}

// adjustForLearningStyle 基于学习风格调整
func (pe *RecommendationPersonalizationEngine) adjustForLearningStyle(rec *ContentRecommendation, style *LearningStyle) {
	// 简化实现：基于学习风格偏好调整分数
	styleBonus := 0.0
	
	// 假设内容类型信息可以从ContentID推断
	contentType := pe.inferContentType(rec.ContentID)
	
	switch contentType {
	case "visual":
		styleBonus = style.VisualLearning * 0.2
	case "audio":
		styleBonus = style.AuditoryLearning * 0.2
	case "interactive":
		styleBonus = style.KinestheticLearning * 0.2
	case "text":
		styleBonus = style.ReadingWritingLearning * 0.2
	}
	
	rec.ConfidenceScore = math.Min(1.0, rec.ConfidenceScore+styleBonus)
}

// adjustForPerformance 基于性能调整
func (pe *RecommendationPersonalizationEngine) adjustForPerformance(rec *ContentRecommendation, performance *PerformanceProfile) {
	// 基于总体性能调整推荐置信度
	performanceBonus := (performance.OverallPerformance - 0.5) * 0.1
	rec.ConfidenceScore = math.Max(0.0, math.Min(1.0, rec.ConfidenceScore+performanceBonus))
}

// adjustForBehavior 基于行为调整
func (pe *RecommendationPersonalizationEngine) adjustForBehavior(rec *ContentRecommendation, behavior *LearningBehavior) {
	// 基于参与度调整
	if behavior.EngagementPatterns != nil {
		engagementBonus := (behavior.EngagementPatterns.EngagementLevel - 0.5) * 0.1
		rec.ConfidenceScore = math.Max(0.0, math.Min(1.0, rec.ConfidenceScore+engagementBonus))
	}
}

// inferContentType 推断内容类型
func (pe *RecommendationPersonalizationEngine) inferContentType(contentID string) string {
	// 简化实现：基于ID模式推断
	contentIDLower := strings.ToLower(contentID)
	
	if strings.Contains(contentIDLower, "video") {
		return "visual"
	} else if strings.Contains(contentIDLower, "audio") {
		return "audio"
	} else if strings.Contains(contentIDLower, "interactive") {
		return "interactive"
	} else {
		return "text"
	}
}

// addPersonalizationReasons 添加个性化原因
func (pe *RecommendationPersonalizationEngine) addPersonalizationReasons(rec *ContentRecommendation, profile *LearnerProfile) {
	// 添加基于学习风格的原因
	if pe.config.EnableLearningStylePersonalization {
		reason := &RecommendationReason{
			Type:        "learning_style_match",
			Description: "Matches your preferred learning style",
			Weight:      0.2,
			Evidence: []*RecommendationEvidence{
				{
					Type:        "style_preference",
					Value:       pe.calculateStyleMatch(rec, profile.LearningStyle),
					Description: "Based on your learning style preferences",
				},
			},
		}
		rec.Reasons = append(rec.Reasons, reason)
	}
}

// calculateStyleMatch 计算风格匹配度
func (pe *RecommendationPersonalizationEngine) calculateStyleMatch(rec *ContentRecommendation, style *LearningStyle) float64 {
	contentType := pe.inferContentType(rec.ContentID)
	
	switch contentType {
	case "visual":
		return style.VisualLearning
	case "audio":
		return style.AuditoryLearning
	case "interactive":
		return style.KinestheticLearning
	case "text":
		return style.ReadingWritingLearning
	default:
		return 0.5
	}
}

// adjustConfidenceScore 调整置信度分数
func (pe *RecommendationPersonalizationEngine) adjustConfidenceScore(rec *ContentRecommendation, profile *LearnerProfile) {
	// 确保分数在有效范围内
	rec.ConfidenceScore = math.Max(0.0, math.Min(1.0, rec.ConfidenceScore))
	
	// 应用最小置信度阈值
	if rec.ConfidenceScore < pe.config.MinConfidenceThreshold {
		rec.ConfidenceScore = pe.config.MinConfidenceThreshold
	}
}

// PersonalizationEngineConfig 个性化引擎配置
type PersonalizationEngineConfig struct {
	EnableLearningStylePersonalization bool
	EnablePerformancePersonalization   bool
	EnableBehaviorPersonalization      bool
	MinConfidenceThreshold             float64
	PersonalizationWeight              float64
}