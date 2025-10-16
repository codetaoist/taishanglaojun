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

// ContentAnalyzer ?
type HelperContentAnalyzer struct {
	config *ContentAnalysisSettings
}

// AnalyzeContent 
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, contentID string, contentData map[string]interface{}) (*ContentAnalysis, error) {
	analysis := &ContentAnalysis{
		ContentID:  contentID,
		AnalyzedAt: time.Now(),
	}
	
	// 
	if ca.Config.EnableSemanticAnalysis {
		semanticFeatures, err := ca.extractSemanticFeatures(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to extract semantic features: %w", err)
		}
		analysis.SemanticFeatures = semanticFeatures
	}
	
	// 
	if ca.Config.EnableDifficultyAnalysis {
		difficultyAnalysis, err := ca.analyzeDifficulty(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze difficulty: %w", err)
		}
		analysis.DifficultyAnalysis = difficultyAnalysis
	}
	
	// 
	if ca.Config.EnableTopicExtraction {
		topicExtraction, err := ca.extractTopics(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to extract topics: %w", err)
		}
		analysis.TopicExtraction = topicExtraction
	}
	
	// 
	if ca.Config.EnablePrerequisiteAnalysis {
		prerequisiteAnalysis, err := ca.analyzePrerequisites(contentData)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze prerequisites: %w", err)
		}
		analysis.PrerequisiteAnalysis = prerequisiteAnalysis
	}
	
	// 
	qualityAssessment, err := ca.assessQuality(contentData)
	if err != nil {
		return nil, fmt.Errorf("failed to assess quality: %w", err)
	}
	analysis.QualityAssessment = qualityAssessment
	
	return analysis, nil
}

// extractSemanticFeatures 
func (ca *ContentAnalyzer) extractSemanticFeatures(contentData map[string]interface{}) (*SemanticFeatures, error) {
	features := &SemanticFeatures{
		Embeddings:         make([]float64, 768), // 768?
		Keywords:           make([]string, 0),
		Concepts:           make([]string, 0),
		Entities:           make([]string, 0),
		SemanticSimilarity: make(map[string]float64),
	}
	
	// 
	if text, ok := contentData["text"].(string); ok {
		// ?
		features.Keywords = ca.extractKeywords(text)
		
		// 
		features.Concepts = ca.extractConcepts(text)
		
		// 
		features.Entities = ca.extractEntities(text)
		
		// 
		features.Embeddings = ca.generateEmbeddings(text)
	}
	
	return features, nil
}

// extractKeywords ?
func (ca *ContentAnalyzer) extractKeywords(text string) []string {
	// 
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)
	
	for _, word := range words {
		if len(word) > 3 { // 
			wordCount[word]++
		}
	}
	
	// ?
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
	
	// ?0
	keywords := make([]string, 0, 10)
	for i, wf := range wordFreqs {
		if i >= 10 {
			break
		}
		keywords = append(keywords, wf.word)
	}
	
	return keywords
}

// extractConcepts 
func (ca *ContentAnalyzer) extractConcepts(text string) []string {
	// ?
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

// extractEntities 
func (ca *ContentAnalyzer) extractEntities(text string) []string {
	// ?
	words := strings.Fields(text)
	entities := make([]string, 0)
	
	for _, word := range words {
		if len(word) > 1 && word[0] >= 'A' && word[0] <= 'Z' {
			entities = append(entities, word)
		}
	}
	
	return entities
}

// generateEmbeddings 
func (ca *ContentAnalyzer) generateEmbeddings(text string) []float64 {
	// 
	embeddings := make([]float64, 768)
	for i := range embeddings {
		embeddings[i] = math.Sin(float64(i) * 0.1) // ?
	}
	return embeddings
}

// analyzeDifficulty 
func (ca *ContentAnalyzer) analyzeDifficulty(contentData map[string]interface{}) (*DifficultyAnalysis, error) {
	analysis := &DifficultyAnalysis{
		DifficultyFactors: make(map[string]float64),
	}
	
	if text, ok := contentData["text"].(string); ok {
		// 
		analysis.CognitiveDifficulty = ca.calculateCognitiveDifficulty(text)
		
		// 
		analysis.LinguisticDifficulty = ca.calculateLinguisticDifficulty(text)
		
		// 
		analysis.ConceptualDifficulty = ca.calculateConceptualDifficulty(text)
		
		// 
		analysis.OverallDifficulty = ca.determineOverallDifficulty(
			analysis.CognitiveDifficulty,
			analysis.LinguisticDifficulty,
			analysis.ConceptualDifficulty,
		)
		
		// 
		analysis.DifficultyFactors["cognitive"] = analysis.CognitiveDifficulty
		analysis.DifficultyFactors["linguistic"] = analysis.LinguisticDifficulty
		analysis.DifficultyFactors["conceptual"] = analysis.ConceptualDifficulty
	}
	
	return analysis, nil
}

// calculateCognitiveDifficulty 
func (ca *ContentAnalyzer) calculateCognitiveDifficulty(text string) float64 {
	// 
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	
	// 䳤?
	difficulty := math.Min(1.0, avgWordsPerSentence/20.0)
	
	return difficulty
}

// calculateLinguisticDifficulty 
func (ca *ContentAnalyzer) calculateLinguisticDifficulty(text string) float64 {
	// 㸴?
	words := strings.Fields(text)
	complexWords := 0
	
	for _, word := range words {
		if len(word) > 7 { // 
			complexWords++
		}
	}
	
	difficulty := float64(complexWords) / float64(len(words))
	return math.Min(1.0, difficulty*2.0)
}

// calculateConceptualDifficulty 
func (ca *ContentAnalyzer) calculateConceptualDifficulty(text string) float64 {
	// ?
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

// determineOverallDifficulty 
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

// extractTopics 
func (ca *ContentAnalyzer) extractTopics(contentData map[string]interface{}) (*TopicExtraction, error) {
	extraction := &TopicExtraction{
		MainTopics:     make([]string, 0),
		SubTopics:      make([]string, 0),
		TopicWeights:   make(map[string]float64),
		TopicHierarchy: make(map[string][]string),
	}
	
	if text, ok := contentData["text"].(string); ok {
		// ?
		keywords := ca.extractKeywords(text)
		
		// 
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

// analyzePrerequisites 
func (ca *ContentAnalyzer) analyzePrerequisites(contentData map[string]interface{}) (*PrerequisiteAnalysis, error) {
	analysis := &PrerequisiteAnalysis{
		RequiredKnowledge:    make([]string, 0),
		RecommendedSkills:    make([]string, 0),
		PrerequisiteConcepts: make([]string, 0),
		DependencyGraph:      make(map[string][]string),
	}
	
	if text, ok := contentData["text"].(string); ok {
		// ?
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

// assessQuality 
func (ca *ContentAnalyzer) assessQuality(contentData map[string]interface{}) (*ContentQualityAssessment, error) {
	assessment := &ContentQualityAssessment{}
	
	if text, ok := contentData["text"].(string); ok {
		// 
		assessment.ContentAccuracy = ca.assessAccuracy(text)
		
		// ?
		assessment.Clarity = ca.assessClarity(text)
		
		// ?
		assessment.Completeness = ca.assessCompleteness(text)
		
		// ?
		assessment.Engagement = ca.assessEngagement(text)
		
		// ?
		assessment.Freshness = ca.assessFreshness(contentData)
		
		// 
		assessment.OverallQuality = (assessment.ContentAccuracy + assessment.Clarity + 
			assessment.Completeness + assessment.Engagement + assessment.Freshness) / 5.0
	}
	
	return assessment, nil
}

// assessAccuracy ?
func (ca *ContentAnalyzer) assessAccuracy(text string) float64 {
	// ?
	sentences := strings.Split(text, ".")
	if len(sentences) < 3 {
		return 0.5 // ?
	}
	
	// 
	contradictionWords := []string{"but", "however", "although", "despite"}
	contradictions := 0
	
	for _, sentence := range sentences {
		for _, word := range contradictionWords {
			if strings.Contains(strings.ToLower(sentence), word) {
				contradictions++
			}
		}
	}
	
	// ?
	accuracy := 1.0 - math.Min(0.5, float64(contradictions)/float64(len(sentences)))
	return accuracy
}

// assessClarity ?
func (ca *ContentAnalyzer) assessClarity(text string) float64 {
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	
	if len(sentences) == 0 {
		return 0.0
	}
	
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	
	// 䳤10-20?
	clarity := 1.0 - math.Abs(avgWordsPerSentence-15.0)/15.0
	return math.Max(0.0, math.Min(1.0, clarity))
}

// assessCompleteness ?
func (ca *ContentAnalyzer) assessCompleteness(text string) float64 {
	// ?
	words := strings.Fields(text)
	
	if len(words) < 50 {
		return 0.3 // 
	} else if len(words) < 200 {
		return 0.6 // 
	} else if len(words) < 500 {
		return 0.8 // ?
	} else {
		return 1.0 // 
	}
}

// assessEngagement ?
func (ca *ContentAnalyzer) assessEngagement(text string) float64 {
	// ?
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

// assessFreshness ?
func (ca *ContentAnalyzer) assessFreshness(contentData map[string]interface{}) float64 {
	// 
	if createdAt, ok := contentData["created_at"].(time.Time); ok {
		age := time.Since(createdAt)
		
		if age < time.Hour*24*30 { // 30
			return 1.0
		} else if age < time.Hour*24*90 { // 90
			return 0.8
		} else if age < time.Hour*24*365 { // 1
			return 0.6
		} else {
			return 0.3
		}
	}
	
	return 0.5 // 
}

// LearnerProfiler 
type HelperLearnerProfiler struct {
	Config *LearnerProfilingSettings
}

// BuildProfile ?
func (lp *LearnerProfiler) BuildProfile(ctx context.Context, learnerID string, learningHistory []map[string]interface{}) (*LearnerProfile, error) {
	profile := &LearnerProfile{
		LearnerID: learnerID,
		UpdatedAt: time.Now(),
	}
	
	// 
	if lp.Config.EnablePreferenceAnalysis {
		preferences, err := lp.analyzeLearningPreferences(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze preferences: %w", err)
		}
		profile.LearningPreferences = preferences
	}
	
	// 
	if lp.Config.EnableBehaviorAnalysis {
		behavior, err := lp.analyzeLearningBehavior(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze behavior: %w", err)
		}
		profile.LearningBehavior = behavior
	}
	
	// 
	if lp.Config.EnablePerformanceAnalysis {
		performance, err := lp.analyzePerformanceProfile(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze performance: %w", err)
		}
		profile.PerformanceProfile = performance
	}
	
	// 
	if lp.Config.EnableLearningStyleAnalysis {
		style, err := lp.analyzeLearningStyle(learningHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze learning style: %w", err)
		}
		profile.LearningStyle = style
	}
	
	// 
	interests, err := lp.buildInterestProfile(learningHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to build interest profile: %w", err)
	}
	profile.InterestProfile = interests
	
	// ?
	knowledge, err := lp.buildKnowledgeState(learningHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to build knowledge state: %w", err)
	}
	profile.KnowledgeState = knowledge
	
	return profile, nil
}

// analyzeLearningPreferences 
func (lp *LearnerProfiler) analyzeLearningPreferences(history []map[string]interface{}) (*LearningPreferences, error) {
	preferences := &LearningPreferences{
		PreferredContentTypes: make([]ContentType, 0),
		PreferredSubjects:     make([]string, 0),
		PreferredLanguages:    make([]string, 0),
		PreferredFormats:      make([]string, 0),
	}
	
	// 
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
	
	// ?
	for contentType, count := range contentTypeCount {
		if count > len(history)/4 { // 25%
			preferences.PreferredContentTypes = append(preferences.PreferredContentTypes, contentType)
		}
	}
	
	// ?
	for subject, count := range subjectCount {
		if count > len(history)/5 { // 20%
			preferences.PreferredSubjects = append(preferences.PreferredSubjects, subject)
		}
	}
	
	// ?
	if sessionCount > 0 {
		preferences.PreferredDuration = totalDuration / time.Duration(sessionCount)
	}
	
	// 
	if len(preferences.PreferredContentTypes) == 0 {
		preferences.PreferredContentTypes = []ContentType{VideoContent, TextContent}
	}
	
	preferences.PreferredDifficulty = IntermediateLevel
	preferences.PreferredLanguages = []string{"zh", "en"}
	preferences.PreferredFormats = []string{"interactive", "multimedia"}
	
	return preferences, nil
}

// analyzeLearningBehavior 
func (lp *LearnerProfiler) analyzeLearningBehavior(history []map[string]interface{}) (*LearningBehavior, error) {
	behavior := &LearningBehavior{}
	
	// 
	studyPatterns, err := lp.analyzeStudyPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.StudyPatterns = studyPatterns
	
	// 
	engagementPatterns, err := lp.analyzeEngagementPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.EngagementPatterns = engagementPatterns
	
	// 
	progressPatterns, err := lp.analyzeProgressPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.ProgressPatterns = progressPatterns
	
	// 
	interactionPatterns, err := lp.analyzeInteractionPatterns(history)
	if err != nil {
		return nil, err
	}
	behavior.InteractionPatterns = interactionPatterns
	
	return behavior, nil
}

// analyzeStudyPatterns 
func (lp *LearnerProfiler) analyzeStudyPatterns(history []map[string]interface{}) (*StudyPatterns, error) {
	patterns := &StudyPatterns{
		PreferredStudyTimes: make([]time.Time, 0),
		BreakPatterns:       make([]time.Duration, 0),
	}
	
	var totalDuration time.Duration
	var sessionCount int
	studyTimes := make(map[int]int) //  -> 
	
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
	
	// 
	if sessionCount > 0 {
		patterns.AverageSessionDuration = totalDuration / time.Duration(sessionCount)
	}
	
	// 
	if len(history) > 0 {
		patterns.StudyFrequency = float64(len(history)) / 30.0 // 30
	}
	
	// ?
	maxCount := 0
	preferredHour := 0
	for hour, count := range studyTimes {
		if count > maxCount {
			maxCount = count
			preferredHour = hour
		}
	}
	
	// 
	preferredTime := time.Date(2024, 1, 1, preferredHour, 0, 0, 0, time.UTC)
	patterns.PreferredStudyTimes = []time.Time{preferredTime}
	
	return patterns, nil
}

// analyzeEngagementPatterns 
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
	
	// ?
	if recordCount > 0 {
		patterns.EngagementLevel = totalEngagement / float64(recordCount)
	}
	
	// ?
	if recordCount > 0 {
		patterns.AttentionSpan = totalAttention / time.Duration(recordCount)
	}
	
	// 㽻
	if recordCount > 0 {
		patterns.InteractionFrequency = float64(totalInteractions) / float64(recordCount)
	}
	
	// ?
	if patterns.EngagementLevel == 0 {
		patterns.EngagementLevel = 0.7
	}
	if patterns.AttentionSpan == 0 {
		patterns.AttentionSpan = time.Minute * 15
	}
	if patterns.InteractionFrequency == 0 {
		patterns.InteractionFrequency = 5.0
	}
	
	patterns.FeedbackResponsiveness = 0.8 // ?
	
	return patterns, nil
}

// analyzeProgressPatterns 
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
	
	// ?
	if totalCount > 0 {
		patterns.CompletionRate = float64(completedCount) / float64(totalCount)
	}
	
	// 
	if progressCount > 0 {
		avgProgress := totalProgress / float64(progressCount)
		patterns.LearningVelocity = avgProgress
	}
	
	// ?
	if patterns.LearningVelocity == 0 {
		patterns.LearningVelocity = 0.6
	}
	if patterns.CompletionRate == 0 {
		patterns.CompletionRate = 0.7
	}
	
	patterns.RetentionRate = 0.8  // ?
	patterns.MasteryRate = 0.75   // ?
	
	return patterns, nil
}

// analyzeInteractionPatterns 
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
	
	// ?
	for interactionType, count := range interactionTypes {
		if count > len(history)/5 { // 20%
			patterns.PreferredInteractionTypes = append(patterns.PreferredInteractionTypes, interactionType)
		}
	}
	
	// 
	if responseCount > 0 {
		patterns.ResponseTime = totalResponseTime / time.Duration(responseCount)
	}
	
	// ?
	if len(patterns.PreferredInteractionTypes) == 0 {
		patterns.PreferredInteractionTypes = []string{"click", "scroll", "type"}
	}
	if patterns.ResponseTime == 0 {
		patterns.ResponseTime = time.Second * 3
	}
	
	patterns.HelpSeekingBehavior = 0.6      // 
	patterns.CollaborationPreference = 0.5  // 
	
	return patterns, nil
}

// analyzePerformanceProfile 
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
	
	// 
	if scoreCount > 0 {
		profile.OverallPerformance = totalScore / float64(scoreCount)
	}
	
	// 
	for subject, scores := range subjectScores {
		var sum float64
		for _, score := range scores {
			sum += score
		}
		avgScore := sum / float64(len(scores))
		profile.SubjectPerformance[subject] = avgScore
		
		// ?
		if avgScore > 0.8 {
			profile.StrengthAreas = append(profile.StrengthAreas, subject)
		} else if avgScore < 0.6 {
			profile.ImprovementAreas = append(profile.ImprovementAreas, subject)
		}
	}
	
	// ?
	profile.SkillLevels["problem_solving"] = 0.7
	profile.SkillLevels["critical_thinking"] = 0.6
	profile.SkillLevels["creativity"] = 0.8
	
	// 
	profile.LearningEfficiency = profile.OverallPerformance * 0.8 // ?
	
	return profile, nil
}

// analyzeLearningStyle 
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
		// 
		style.VisualLearning = 0.7
		style.AuditoryLearning = 0.5
		style.KinestheticLearning = 0.6
		style.ReadingWritingLearning = 0.8
		style.SequentialLearning = 0.6
		style.GlobalLearning = 0.4
		return style, nil
	}
	
	// 
	style.VisualLearning = float64(contentTypePreferences["video"]+contentTypePreferences["image"]) / float64(total)
	style.AuditoryLearning = float64(contentTypePreferences["audio"]+contentTypePreferences["podcast"]) / float64(total)
	style.KinestheticLearning = float64(contentTypePreferences["interactive"]+contentTypePreferences["exercise"]) / float64(total)
	style.ReadingWritingLearning = float64(contentTypePreferences["text"]+contentTypePreferences["document"]) / float64(total)
	
	// ?
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
	
	style.SequentialLearning = 0.6 // ?
	style.GlobalLearning = 0.4     // ?
	
	return style, nil
}

// buildInterestProfile 
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
	
	// 
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

// buildKnowledgeState ?
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
	
	// 
	for concept, progressList := range conceptProgress {
		if len(progressList) == 0 {
			continue
		}
		
		// ?
		var sum float64
		for _, progress := range progressList {
			sum += progress
		}
		avgMastery := sum / float64(len(progressList))
		state.ConceptMastery[concept] = avgMastery
		
		// 
		if avgMastery >= 0.8 {
			state.MasteredConcepts = append(state.MasteredConcepts, concept)
		} else if avgMastery >= 0.4 {
			state.LearningConcepts = append(state.LearningConcepts, concept)
		} else {
			state.KnowledgeGaps = append(state.KnowledgeGaps, concept)
		}
	}
	
	// 
	state.LearningGoals = []string{
		"improve weak areas",
		"master current concepts",
		"explore new topics",
	}
	
	return state, nil
}

// RecommendationAlgorithm 㷨
type RecommendationAlgorithm interface {
	GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error)
	GetAlgorithmName() string
	GetAlgorithmType() RecommendationStrategy
}

// HelperRecommendationAlgorithm 㷨
type HelperRecommendationAlgorithm interface {
	GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error)
	GetAlgorithmName() string
	GetAlgorithmType() RecommendationStrategy
}

// CollaborativeFilteringAlgorithm 㷨
type CollaborativeFilteringAlgorithm struct {
	config *HelperCollaborativeFilteringConfig
}

// GenerateRecommendations 
func (cf *CollaborativeFilteringAlgorithm) GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error) {
	recommendations := make([]*ContentRecommendation, 0)
	
	// 
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
	
	// 
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].ConfidenceScore > recommendations[j].ConfidenceScore
	})
	
	// 
	if len(recommendations) > config.MaxRecommendations {
		recommendations = recommendations[:config.MaxRecommendations]
	}
	
	return recommendations, nil
}

// calculateCollaborativeScore 
func (cf *CollaborativeFilteringAlgorithm) calculateCollaborativeScore(profile *LearnerProfile, content *ContentAnalysis) float64 {
	// 
	score := 0.0
	
	// ?
	for _, preferredType := range profile.LearningPreferences.PreferredContentTypes {
		if string(preferredType) == content.ContentID { // ?
			score += 0.3
		}
	}
	
	// ?
	if content.DifficultyAnalysis != nil {
		difficultyMatch := 1.0 - math.Abs(DifficultyLevelToFloat64(profile.LearningPreferences.PreferredDifficulty)-DifficultyLevelToFloat64(content.DifficultyAnalysis.OverallDifficulty))
		score += difficultyMatch * 0.4
	}
	
	// ?
	if content.TopicExtraction != nil {
		for _, topic := range content.TopicExtraction.MainTopics {
			if interest, exists := profile.InterestProfile.TopicInterests[topic]; exists {
				score += interest * 0.3
			}
		}
	}
	
	return math.Min(1.0, score)
}

// GetAlgorithmName 㷨
func (cf *CollaborativeFilteringAlgorithm) GetAlgorithmName() string {
	return "Collaborative Filtering"
}

// GetAlgorithmType 㷨
func (cf *CollaborativeFilteringAlgorithm) GetAlgorithmType() RecommendationStrategy {
	return CollaborativeFiltering
}

// ContentBasedAlgorithm ?
type ContentBasedAlgorithm struct {
	config *ContentBasedConfig
}

// GenerateRecommendations 
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
	
	// 
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].ConfidenceScore > recommendations[j].ConfidenceScore
	})
	
	// 
	if len(recommendations) > config.MaxRecommendations {
		recommendations = recommendations[:config.MaxRecommendations]
	}
	
	return recommendations, nil
}

// calculateContentBasedScore ?
func (cb *ContentBasedAlgorithm) calculateContentBasedScore(profile *LearnerProfile, content *ContentAnalysis) float64 {
	score := 0.0
	
	// ?
	if content.SemanticFeatures != nil {
		semanticScore := cb.calculateSemanticSimilarity(profile, content.SemanticFeatures)
		score += semanticScore * 0.4
	}
	
	// 
	if content.DifficultyAnalysis != nil {
		difficultyScore := cb.calculateDifficultyMatch(profile, content.DifficultyAnalysis)
		score += difficultyScore * 0.3
	}
	
	// 
	if content.TopicExtraction != nil {
		topicScore := cb.calculateTopicMatch(profile, content.TopicExtraction)
		score += topicScore * 0.3
	}
	
	return math.Min(1.0, score)
}

// calculateSemanticSimilarity ?
func (cb *ContentBasedAlgorithm) calculateSemanticSimilarity(profile *LearnerProfile, features *SemanticFeatures) float64 {
	// ?
	matchCount := 0
	totalKeywords := len(features.Keywords)
	
	if totalKeywords == 0 {
		return 0.5 // ?
	}
	
	// ?
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

// calculateDifficultyMatch ?
func (cb *ContentBasedAlgorithm) calculateDifficultyMatch(profile *LearnerProfile, difficulty *DifficultyAnalysis) float64 {
	preferredDifficulty := DifficultyLevelToFloat64(profile.LearningPreferences.PreferredDifficulty)
	contentDifficulty := DifficultyLevelToFloat64(difficulty.OverallDifficulty)
	
	// 
	difficultyDiff := math.Abs(preferredDifficulty - contentDifficulty)
	
	// ?
	matchScore := 1.0 - (difficultyDiff / 3.0) // ?-3
	
	return math.Max(0.0, matchScore)
}

// calculateTopicMatch ?
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

// GetAlgorithmName 㷨
func (cb *ContentBasedAlgorithm) GetAlgorithmName() string {
	return "Content-Based Filtering"
}

// GetAlgorithmType 㷨
func (cb *ContentBasedAlgorithm) GetAlgorithmType() RecommendationStrategy {
	return ContentBased
}

// HybridAlgorithm 㷨
type HybridAlgorithm struct {
	collaborativeAlgorithm *CollaborativeFilteringAlgorithm
	contentBasedAlgorithm  *ContentBasedAlgorithm
	config                 *HybridConfig
}

// GenerateRecommendations 
func (h *HybridAlgorithm) GenerateRecommendations(ctx context.Context, profile *LearnerProfile, availableContent []ContentAnalysis, config *AlgorithmSettings) ([]*ContentRecommendation, error) {
	// 
	collaborativeRecs, err := h.collaborativeAlgorithm.GenerateRecommendations(ctx, profile, availableContent, config)
	if err != nil {
		return nil, fmt.Errorf("collaborative filtering failed: %w", err)
	}
	
	// ?
	contentBasedRecs, err := h.contentBasedAlgorithm.GenerateRecommendations(ctx, profile, availableContent, config)
	if err != nil {
		return nil, fmt.Errorf("content-based filtering failed: %w", err)
	}
	
	// ?
	hybridRecs := h.combineRecommendations(collaborativeRecs, contentBasedRecs)
	
	// 
	sort.Slice(hybridRecs, func(i, j int) bool {
		return hybridRecs[i].ConfidenceScore > hybridRecs[j].ConfidenceScore
	})
	
	// 
	if len(hybridRecs) > config.MaxRecommendations {
		hybridRecs = hybridRecs[:config.MaxRecommendations]
	}
	
	return hybridRecs, nil
}

// combineRecommendations 
func (h *HybridAlgorithm) combineRecommendations(collaborativeRecs, contentBasedRecs []*ContentRecommendation) []*ContentRecommendation {
	recommendationMap := make(map[string]*ContentRecommendation)
	
	// 
	for _, rec := range collaborativeRecs {
		rec.ConfidenceScore *= h.config.CollaborativeWeight
		rec.RecommendationStrategy = HybridApproach
		recommendationMap[rec.ContentID] = rec
	}
	
	// 
	for _, rec := range contentBasedRecs {
		rec.ConfidenceScore *= h.config.ContentBasedWeight
		
		if existingRec, exists := recommendationMap[rec.ContentID]; exists {
			// 
			existingRec.ConfidenceScore += rec.ConfidenceScore
			existingRec.ConfidenceScore = math.Min(1.0, existingRec.ConfidenceScore)
			
			// 
			existingRec.Reasons = append(existingRec.Reasons, rec.Reasons...)
		} else {
			rec.RecommendationStrategy = HybridApproach
			recommendationMap[rec.ContentID] = rec
		}
	}
	
	// ?
	result := make([]*ContentRecommendation, 0, len(recommendationMap))
	for _, rec := range recommendationMap {
		result = append(result, rec)
	}
	
	return result
}

// GetAlgorithmName 㷨
func (h *HybridAlgorithm) GetAlgorithmName() string {
	return "Hybrid Recommendation"
}

// GetAlgorithmType 㷨
func (h *HybridAlgorithm) GetAlgorithmType() RecommendationStrategy {
	return HybridApproach
}

// RecommendationEngine 
type HelperRecommendationEngine struct {
	algorithms map[RecommendationStrategy]HelperRecommendationAlgorithm
	config     *RecommendationEngineConfig
}

// NewRecommendationEngine 
func NewHelperRecommendationEngine(config *RecommendationEngineConfig) *HelperRecommendationEngine {
	engine := &HelperRecommendationEngine{
		algorithms: make(map[RecommendationStrategy]HelperRecommendationAlgorithm),
		config:     config,
	}
	
	// 㷨
	engine.registerAlgorithms()
	
	return engine
}

// registerAlgorithms 㷨
func (re *HelperRecommendationEngine) registerAlgorithms() {
	// 㷨
	collaborativeAlgorithm := &CollaborativeFilteringAlgorithm{
		config: &HelperCollaborativeFilteringConfig{
			MinSimilarUsers:    5,
			SimilarityThreshold: 0.7,
			UserSimilarityMethod: "cosine",
		},
	}
	re.algorithms[CollaborativeFiltering] = collaborativeAlgorithm
	
	// ?
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
	
	// 㷨
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

// registerAlgorithms 㷨 (RecommendationEngine)
func (re *RecommendationEngine) registerAlgorithms() {
	// 㷨
	collaborativeAlgorithm := &CollaborativeFilteringAlgorithm{
		config: &HelperCollaborativeFilteringConfig{
			MinSimilarUsers:    5,
			SimilarityThreshold: 0.7,
			UserSimilarityMethod: "cosine",
		},
	}
	re.algorithms[CollaborativeFiltering] = collaborativeAlgorithm
	
	// ?
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
	
	// 㷨
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

// GenerateRecommendations 
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

// GetAvailableStrategies 
func (re *RecommendationEngine) GetAvailableStrategies() []RecommendationStrategy {
	strategies := make([]RecommendationStrategy, 0, len(re.algorithms))
	for strategy := range re.algorithms {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// ?
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

// RecommendationPersonalizationEngine 
type RecommendationPersonalizationEngine struct {
	config *PersonalizationEngineConfig
}

// NewRecommendationPersonalizationEngine 
func NewRecommendationPersonalizationEngine(config *PersonalizationEngineConfig) *RecommendationPersonalizationEngine {
	return &RecommendationPersonalizationEngine{
		config: config,
	}
}

// PersonalizeRecommendations 
func (pe *RecommendationPersonalizationEngine) PersonalizeRecommendations(ctx context.Context, recommendations []*ContentRecommendation, profile *LearnerProfile) ([]*ContentRecommendation, error) {
	personalizedRecs := make([]*ContentRecommendation, len(recommendations))
	copy(personalizedRecs, recommendations)
	
	for _, rec := range personalizedRecs {
		// 
		pe.applyPersonalizationFactors(rec, profile)
		
		// 
		pe.addPersonalizationReasons(rec, profile)
		
		// ?
		pe.adjustConfidenceScore(rec, profile)
	}
	
	// 
	sort.Slice(personalizedRecs, func(i, j int) bool {
		return personalizedRecs[i].ConfidenceScore > personalizedRecs[j].ConfidenceScore
	})
	
	return personalizedRecs, nil
}

// applyPersonalizationFactors 
func (pe *RecommendationPersonalizationEngine) applyPersonalizationFactors(rec *ContentRecommendation, profile *LearnerProfile) {
	// 
	if pe.config.EnableLearningStylePersonalization {
		pe.adjustForLearningStyle(rec, profile.LearningStyle)
	}
	
	// 
	if pe.config.EnablePerformancePersonalization {
		pe.adjustForPerformance(rec, profile.PerformanceProfile)
	}
	
	// 
	if pe.config.EnableBehaviorPersonalization {
		pe.adjustForBehavior(rec, profile.LearningBehavior)
	}
}

// adjustForLearningStyle 
func (pe *RecommendationPersonalizationEngine) adjustForLearningStyle(rec *ContentRecommendation, style *LearningStyle) {
	// 
	styleBonus := 0.0
	
	// ContentID
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

// adjustForPerformance 
func (pe *RecommendationPersonalizationEngine) adjustForPerformance(rec *ContentRecommendation, performance *PerformanceProfile) {
	// ?
	performanceBonus := (performance.OverallPerformance - 0.5) * 0.1
	rec.ConfidenceScore = math.Max(0.0, math.Min(1.0, rec.ConfidenceScore+performanceBonus))
}

// adjustForBehavior 
func (pe *RecommendationPersonalizationEngine) adjustForBehavior(rec *ContentRecommendation, behavior *LearningBehavior) {
	// ?
	if behavior.EngagementPatterns != nil {
		engagementBonus := (behavior.EngagementPatterns.EngagementLevel - 0.5) * 0.1
		rec.ConfidenceScore = math.Max(0.0, math.Min(1.0, rec.ConfidenceScore+engagementBonus))
	}
}

// inferContentType 
func (pe *RecommendationPersonalizationEngine) inferContentType(contentID string) string {
	// ID
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

// addPersonalizationReasons 
func (pe *RecommendationPersonalizationEngine) addPersonalizationReasons(rec *ContentRecommendation, profile *LearnerProfile) {
	// ?
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

// calculateStyleMatch ?
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

// adjustConfidenceScore ?
func (pe *RecommendationPersonalizationEngine) adjustConfidenceScore(rec *ContentRecommendation, profile *LearnerProfile) {
	// 
	rec.ConfidenceScore = math.Max(0.0, math.Min(1.0, rec.ConfidenceScore))
	
	// ?
	if rec.ConfidenceScore < pe.config.MinConfidenceThreshold {
		rec.ConfidenceScore = pe.config.MinConfidenceThreshold
	}
}

// PersonalizationEngineConfig 
type PersonalizationEngineConfig struct {
	EnableLearningStylePersonalization bool
	EnablePerformancePersonalization   bool
	EnableBehaviorPersonalization      bool
	MinConfidenceThreshold             float64
	PersonalizationWeight              float64
}

