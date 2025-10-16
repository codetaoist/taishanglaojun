package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ContextManager 
type ContextManager struct {
	db     *gorm.DB
	logger *zap.Logger
}

// ConversationContext 
type ConversationContext struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	SessionID     string                 `json:"session_id" gorm:"index"`
	UserID        uint                   `json:"user_id" gorm:"index"`
	Intent        string                 `json:"intent"`                // 
	Entities      map[string]interface{} `json:"entities" gorm:"type:json"` // 
	Sentiment     string                 `json:"sentiment"`             // 
	Topics        []string               `json:"topics" gorm:"type:json"`   // 
	Keywords      []string               `json:"keywords" gorm:"type:json"` // 
	Summary       string                 `json:"summary"`               // 
	LastUpdated   time.Time              `json:"last_updated"`
	MessageCount  int                    `json:"message_count"`
	ContextWindow []ContextMessage       `json:"context_window" gorm:"type:json"` // 
	// 
	PersonalityProfile map[string]interface{} `json:"personality_profile" gorm:"type:json"` // 
	ConversationFlow   []string               `json:"conversation_flow" gorm:"type:json"`   // 
	UserPreferences    map[string]interface{} `json:"user_preferences" gorm:"type:json"`    // 
	MemoryBank         []MemoryItem           `json:"memory_bank" gorm:"type:json"`         // 
	ContextScore       float64                `json:"context_score"`                        // 
}

// ContextMessage 
type ContextMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Intent    string    `json:"intent,omitempty"`
	Entities  map[string]interface{} `json:"entities,omitempty"`
	// 
	Importance float64 `json:"importance"` // 
	Reference  string  `json:"reference"`  // ID
}

// MemoryItem 
type MemoryItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // fact, preference, event, relationship
	Content     string                 `json:"content"`
	Entities    map[string]interface{} `json:"entities"`
	Timestamp   time.Time              `json:"timestamp"`
	Importance  float64                `json:"importance"`
	AccessCount int                    `json:"access_count"`
	LastAccess  time.Time              `json:"last_access"`
}

// IntentAnalysisResult 
type IntentAnalysisResult struct {
	Intent     string                 `json:"intent"`
	Confidence float64                `json:"confidence"`
	Entities   map[string]interface{} `json:"entities"`
	Sentiment  string                 `json:"sentiment"`
	Keywords   []string               `json:"keywords"`
	Topics     []string               `json:"topics"`
	// 
	EmotionalState string  `json:"emotional_state"` // 
	Urgency        float64 `json:"urgency"`         // 
	Complexity     float64 `json:"complexity"`      // 
}

// NewContextManager 
func NewContextManager(db *gorm.DB, logger *zap.Logger) *ContextManager {
	return &ContextManager{
		db:     db,
		logger: logger,
	}
}

// GetOrCreateContext 
func (cm *ContextManager) GetOrCreateContext(ctx context.Context, sessionID string, userID uint) (*ConversationContext, error) {
	var context ConversationContext
	
	err := cm.db.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&context).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 
			context = ConversationContext{
				SessionID:          sessionID,
				UserID:             userID,
				Intent:             "unknown",
				Entities:           make(map[string]interface{}),
				Sentiment:          "neutral",
				Topics:             []string{},
				Keywords:           []string{},
				Summary:            "",
				LastUpdated:        time.Now(),
				MessageCount:       0,
				ContextWindow:      []ContextMessage{},
				PersonalityProfile: make(map[string]interface{}),
				ConversationFlow:   []string{},
				UserPreferences:    make(map[string]interface{}),
				MemoryBank:         []MemoryItem{},
				ContextScore:       0.0,
			}
			
			if err := cm.db.Create(&context).Error; err != nil {
				return nil, fmt.Errorf("failed to create context: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get context: %w", err)
		}
	}
	
	return &context, nil
}

// UpdateContext 
func (cm *ContextManager) UpdateContext(ctx context.Context, sessionID string, userMessage, aiResponse string, analysis *IntentAnalysisResult) error {
	var context ConversationContext
	
	err := cm.db.Where("session_id = ?", sessionID).First(&context).Error
	if err != nil {
		return fmt.Errorf("failed to get context: %w", err)
	}

	// 
	context.Intent = analysis.Intent
	context.Sentiment = analysis.Sentiment
	context.Keywords = cm.mergeKeywords(context.Keywords, analysis.Keywords)
	context.Topics = cm.mergeTopics(context.Topics, analysis.Topics)
	context.MessageCount++
	context.LastUpdated = time.Now()

	// 
	userImportance := cm.calculateMessageImportance(userMessage, analysis)
	aiImportance := cm.calculateMessageImportance(aiResponse, nil)

	// 
	userContextMsg := ContextMessage{
		Role:       "user",
		Content:    userMessage,
		Timestamp:  time.Now(),
		Intent:     analysis.Intent,
		Entities:   analysis.Entities,
		Importance: userImportance,
	}

	aiContextMsg := ContextMessage{
		Role:       "assistant",
		Content:    aiResponse,
		Timestamp:  time.Now(),
		Importance: aiImportance,
	}

	context.ContextWindow = append(context.ContextWindow, userContextMsg, aiContextMsg)

	// 
	if len(context.ContextWindow) > 20 {
		context.ContextWindow = cm.pruneContextWindow(context.ContextWindow, 16)
	}

	// 
	cm.updateMemoryBank(&context, userMessage, analysis)

	// 
	cm.updatePersonalityProfile(&context, analysis)

	// 
	context.ConversationFlow = append(context.ConversationFlow, analysis.Intent)
	if len(context.ConversationFlow) > 50 {
		context.ConversationFlow = context.ConversationFlow[len(context.ConversationFlow)-50:]
	}

	// 
	context.Summary = cm.generateEnhancedSummary(context.ContextWindow, context.MemoryBank)

	// 
	context.ContextScore = cm.calculateContextScore(&context)

	// 
	if err := cm.db.Save(&context).Error; err != nil {
		return fmt.Errorf("failed to update context: %w", err)
	}

	return nil
}

// calculateMessageImportance 
func (cm *ContextManager) calculateMessageImportance(message string, analysis *IntentAnalysisResult) float64 {
	importance := 0.5 // 	

	// 
	if len(message) > 100 {
		importance += 0.1
	}

	// 
	if analysis != nil {
		importance += analysis.Confidence * 0.3
		importance += analysis.Urgency * 0.2
		importance += analysis.Complexity * 0.1
	}

	// 
	keywordCount := len(strings.Fields(message))
	if keywordCount > 10 {
		importance += 0.1
	}

	// 0-1
	if importance > 1.0 {
		importance = 1.0
	}
	if importance < 0.0 {
		importance = 0.0
	}

	return importance
}

// pruneContextWindow 
func (cm *ContextManager) pruneContextWindow(messages []ContextMessage, targetSize int) []ContextMessage {
	if len(messages) <= targetSize {
		return messages
	}

	// 
	sortedMessages := make([]ContextMessage, len(messages))
	copy(sortedMessages, messages)

	// 㷨
	for i := 0; i < len(sortedMessages)-1; i++ {
		for j := i + 1; j < len(sortedMessages); j++ {
			if sortedMessages[i].Importance < sortedMessages[j].Importance {
				sortedMessages[i], sortedMessages[j] = sortedMessages[j], sortedMessages[i]
			}
		}
	}

	// 
	result := make([]ContextMessage, 0, targetSize)
	
	// 
	recentCount := targetSize / 4
	if recentCount < 2 {
		recentCount = 2
	}
	
	// 
	startIdx := len(messages) - recentCount
	if startIdx < 0 {
		startIdx = 0
	}
	result = append(result, messages[startIdx:]...)

	// 
	remainingSlots := targetSize - len(result)
	for i := 0; i < len(sortedMessages) && remainingSlots > 0; i++ {
		msg := sortedMessages[i]
		// 
		if msg.Timestamp.Before(messages[startIdx].Timestamp) {
			result = append(result, msg)
			remainingSlots--
		}
	}

	return result
}

// updateMemoryBank 
func (cm *ContextManager) updateMemoryBank(context *ConversationContext, message string, analysis *IntentAnalysisResult) {
	// 
	if analysis.Confidence > 0.7 {
		memory := MemoryItem{
			ID:          fmt.Sprintf("mem_%d_%s", time.Now().Unix(), analysis.Intent),
			Type:        "fact",
			Content:     message,
			Entities:    analysis.Entities,
			Timestamp:   time.Now(),
			Importance:  analysis.Confidence,
			AccessCount: 0,
			LastAccess:  time.Now(),
		}

		context.MemoryBank = append(context.MemoryBank, memory)

		// 
		if len(context.MemoryBank) > 100 {
			// 
			context.MemoryBank = cm.pruneMemoryBank(context.MemoryBank, 80)
		}
	}
}

// pruneMemoryBank 
func (cm *ContextManager) pruneMemoryBank(memories []MemoryItem, targetSize int) []MemoryItem {
	if len(memories) <= targetSize {
		return memories
	}

	// 	
	for i := range memories {
		// 	
		timeFactor := 1.0 - float64(time.Since(memories[i].LastAccess).Hours())/8760 // 
		accessFactor := float64(memories[i].AccessCount) / 10.0
		if accessFactor > 1.0 {
			accessFactor = 1.0
		}
		
		memories[i].Importance = memories[i].Importance*0.5 + timeFactor*0.3 + accessFactor*0.2
	}

	// targetSize
	for i := 0; i < len(memories)-1; i++ {
		for j := i + 1; j < len(memories); j++ {
			if memories[i].Importance < memories[j].Importance {
				memories[i], memories[j] = memories[j], memories[i]
			}
		}
	}

	return memories[:targetSize]
}

// updatePersonalityProfile 
func (cm *ContextManager) updatePersonalityProfile(context *ConversationContext, analysis *IntentAnalysisResult) {
	// 
	if sentiment, exists := context.PersonalityProfile["sentiment_tendency"]; exists {
		if sentimentMap, ok := sentiment.(map[string]float64); ok {
			if count, exists := sentimentMap[analysis.Sentiment]; exists {
				sentimentMap[analysis.Sentiment] = count + 1
			} else {
				sentimentMap[analysis.Sentiment] = 1
			}
		}
	} else {
		context.PersonalityProfile["sentiment_tendency"] = map[string]float64{
			analysis.Sentiment: 1,
		}
	}

	// 
	if intent, exists := context.PersonalityProfile["intent_preference"]; exists {
		if intentMap, ok := intent.(map[string]float64); ok {
			if count, exists := intentMap[analysis.Intent]; exists {
				intentMap[analysis.Intent] = count + 1
			} else {
				intentMap[analysis.Intent] = 1
			}
		}
	} else {
		context.PersonalityProfile["intent_preference"] = map[string]float64{
			analysis.Intent: 1,
		}
	}
}

// calculateContextScore 
func (cm *ContextManager) calculateContextScore(context *ConversationContext) float64 {
	score := 0.0

	// 
	messageScore := float64(context.MessageCount) / 100.0
	if messageScore > 1.0 {
		messageScore = 1.0
	}
	score += messageScore * 0.2

	// 
	memoryScore := float64(len(context.MemoryBank)) / 50.0
	if memoryScore > 1.0 {
		memoryScore = 1.0
	}
	score += memoryScore * 0.3

	// 
	topicScore := float64(len(context.Topics)) / 20.0
	if topicScore > 1.0 {
		topicScore = 1.0
	}
	score += topicScore * 0.2

	// 
	timeSinceLastUpdate := time.Since(context.LastUpdated).Hours()
	timeScore := 1.0 - (timeSinceLastUpdate / 168.0) // 1.0
	if timeScore < 0 {
		timeScore = 0
	}
	score += timeScore * 0.3

	return score
}

// generateEnhancedSummary 
func (cm *ContextManager) generateEnhancedSummary(contextWindow []ContextMessage, memoryBank []MemoryItem) string {
	if len(contextWindow) == 0 {
		return ""
	}

	// 
	var topics []string
	var intents []string
	
	for _, msg := range contextWindow {
		if msg.Intent != "" && msg.Intent != "unknown" {
			intents = append(intents, msg.Intent)
		}
	}

	// 
	for _, memory := range memoryBank {
		if memory.Importance > 0.7 {
			topics = append(topics, memory.Type)
		}
	}

	// 
	summary := "漰"
	if len(intents) > 0 {
		uniqueIntents := cm.removeDuplicates(intents)
		summary += ": " + strings.Join(uniqueIntents[:min(3, len(uniqueIntents))], ", ")
	}
	
	if len(topics) > 0 {
		uniqueTopics := cm.removeDuplicates(topics)
		summary += "; : " + strings.Join(uniqueTopics[:min(3, len(uniqueTopics))], ", ")
	}

	return summary
}

// removeDuplicates 
func (cm *ContextManager) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// min 
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AnalyzeIntent 
func (cm *ContextManager) AnalyzeIntent(ctx context.Context, message string, conversationHistory []models.ChatMessage) (*IntentAnalysisResult, error) {
	// 
	intent := cm.detectIntent(message)
	entities := cm.extractEntities(message)
	sentiment := cm.analyzeSentiment(message)
	keywords := cm.extractKeywords(message)
	topics := cm.identifyTopics(message, conversationHistory)
	
	// 㸴
	complexity := cm.calculateComplexity(message, entities)
	urgency := cm.calculateUrgency(message, intent)
	emotionalState := cm.analyzeEmotionalState(message, sentiment)
	
	// 
	confidence := cm.calculateConfidence(intent, message, conversationHistory)
	
	return &IntentAnalysisResult{
		Intent:         intent,
		Confidence:     confidence,
		Entities:       entities,
		Sentiment:      sentiment,
		Keywords:       keywords,
		Topics:         topics,
		EmotionalState: emotionalState,
		Urgency:        urgency,
		Complexity:     complexity,
	}, nil
}

// GetContextualPrompt 
func (cm *ContextManager) GetContextualPrompt(ctx context.Context, sessionID string, currentMessage string) (string, error) {
	var context ConversationContext
	
	err := cm.db.Where("session_id = ?", sessionID).First(&context).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return currentMessage, nil // 
		}
		return "", fmt.Errorf("failed to get context: %w", err)
	}
	
	// 
	prompt := cm.buildEnhancedContextualPrompt(&context, currentMessage)
	return prompt, nil
}

// detectIntent 
func (cm *ContextManager) detectIntent(message string) string {
	message = strings.ToLower(message)
	
	// 
	intentKeywords := map[string][]string{
		"question":     {"", "", "", "", "", "when", "what", "how", "why", "where", "who", "?", },
		"request":      {"", "", "", "", "", "", "need", "please", "help", "can you", "could you", "would you"},
		"greeting":     {"", "hello", "hi", "", "", "", "goodbye", "bye"},
		"complaint":    {"", "", "", "", "bug", "", "complaint", "issue", "problem", "error", "wrong"},
		"praise":       {"", "", "", "", "", "great", "excellent", "thank", "awesome", "wonderful"},
		"cultural":     {"", "", "", "", "", "", "", "", "", "wisdom", "philosophy", "tradition"},
		"learning":     {"", "", "", "", "learn", "study", "knowledge", "understand", "explain"},
		"emotional":    {"", "", "", "", "", "", "feel", "emotion", "mood", "sad", "happy", "anxious"},
		"planning":     {"", "", "", "", "plan", "schedule", "prepare", "intend", "organize"},
		"comparison":   {"", "", "", "", "", "compare", "difference", "similar", "versus", "vs"},
	}
	
	// 
	intentScores := make(map[string]int)
	for intent, keywords := range intentKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				intentScores[intent]++
			}
		}
	}
	
	// 
	maxScore := 0
	bestIntent := "general"
	for intent, score := range intentScores {
		if score > maxScore {
			maxScore = score
			bestIntent = intent
		}
	}
	
	return bestIntent
}

// extractEntities 
func (cm *ContextManager) extractEntities(message string) map[string]interface{} {
	entities := make(map[string]interface{})
	
	// 
	timeKeywords := []string{"", "", "", "", "", "", "", "", "", "", "", ""}
	for _, keyword := range timeKeywords {
		if strings.Contains(message, keyword) {
			if entities["time"] == nil {
				entities["time"] = []string{}
			}
			entities["time"] = append(entities["time"].([]string), keyword)
		}
	}
	
	// 
	culturalEntities := []string{"", "", "", "", "", "", "", "", "", "", "", "", ""}
	for _, entity := range culturalEntities {
		if strings.Contains(message, entity) {
			if entities["cultural"] == nil {
				entities["cultural"] = []string{}
			}
			entities["cultural"] = append(entities["cultural"].([]string), entity)
		}
	}
	
	// 
	locationEntities := []string{"", "", "", "", "", "", "", "", "", "", ""}
	for _, entity := range locationEntities {
		if strings.Contains(message, entity) {
			if entities["location"] == nil {
				entities["location"] = []string{}
			}
			entities["location"] = append(entities["location"].([]string), entity)
		}
	}
	
	// 
	numbers := []string{"", "", "", "", "", "", "", "", "", "", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	for _, num := range numbers {
		if strings.Contains(message, num) {
			if entities["number"] == nil {
				entities["number"] = []string{}
			}
			entities["number"] = append(entities["number"].([]string), num)
		}
	}
	
	return entities
}

// analyzeSentiment 
func (cm *ContextManager) analyzeSentiment(message string) string {
	message = strings.ToLower(message)
	
	// 
	positiveWords := []string{
		"", "", "", "", "", "", "", "", "", "", "",
		"great", "good", "excellent", "happy", "satisfied", "love", "amazing", "wonderful", "fantastic",
	}
	negativeWords := []string{
		"", "", "", "", "", "", "", "", "", "", "",
		"bad", "terrible", "angry", "disappointed", "hate", "sad", "worried", "anxious", "frustrated",
	}
	
	positiveCount := 0
	negativeCount := 0
	
	// 
	for _, word := range positiveWords {
		positiveCount += strings.Count(message, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(message, word)
	}
	
	// 
	negationWords := []string{"", "", "not", "no", "never", "neither"}
	hasNegation := false
	for _, neg := range negationWords {
		if strings.Contains(message, neg) {
			hasNegation = true
			break
		}
	}
	
	// 
	if hasNegation {
		if positiveCount > negativeCount {
			return "negative"
		} else if negativeCount > positiveCount {
			return "positive"
		}
	} else {
		if positiveCount > negativeCount {
			return "positive"
		} else if negativeCount > positiveCount {
			return "negative"
		}
	}
	
	return "neutral"
}

// extractKeywords 
func (cm *ContextManager) extractKeywords(message string) []string {
	words := strings.Fields(message)
	keywords := []string{}
	
	// 停用词列表
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true, "有": true, "和": true, "就": true,
		"不": true, "人": true, "都": true, "一": true, "一个": true, "上": true, "也": true, "很": true,
		"and": true, "the": true, "is": true, "in": true, "to": true, "of": true, "a": true, "that": true,
		"it": true, "with": true, "for": true, "as": true, "was": true, "on": true, "are": true, "you": true,
	}
	
	// 重要词汇列表
	importantWords := map[string]bool{
		"问题": true, "解决": true, "方法": true, "技术": true, "系统": true, "功能": true, "开发": true,
		"设计": true, "实现": true, "优化": true, "性能": true, "安全": true, "数据": true,
	}
	
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		// 
		punctuation := `.,!?;:"'()[]{}""''`
		word = strings.Trim(word, punctuation)
		
		if len(word) > 1 && !stopWords[word] {
			keywords = append(keywords, word)
			// 
			if importantWords[word] {
				keywords = append(keywords, word)
			}
		}
	}
	
	// 
	uniqueKeywords := cm.removeDuplicates(keywords)
	if len(uniqueKeywords) > 15 {
		uniqueKeywords = uniqueKeywords[:15]
	}
	
	return uniqueKeywords
}

// identifyTopics 
func (cm *ContextManager) identifyTopics(message string, history []models.ChatMessage) []string {
	topics := []string{}
	message = strings.ToLower(message)
	
	// 
	topicKeywords := map[string][]string{
		"philosophy": {"", "", "", "", "", "philosophy", "wisdom", "thought", "concept", "idea"},
		"culture":    {"", "", "", "", "", "", "tradition", "custom", "civilization", "heritage"},
		"religion":   {"", "", "", "", "", "religion", "buddhism", "taoism", "confucianism", "belief"},
		"technology": {"", "", "AI", "", "", "technology", "artificial intelligence", "computer", "digital"},
		"life":       {"", "", "", "", "", "life", "living", "existence", "meaning", "purpose"},
		"learning":   {"", "", "", "", "", "learning", "education", "knowledge", "research", "study"},
		"emotion":    {"", "", "", "", "", "emotion", "feeling", "psychology", "experience", "mood"},
		"health":     {"", "", "", "", "", "health", "wellness", "medical", "physical", "mental health"},
		"art":        {"", "", "", "", "", "art", "music", "literature", "poetry", "painting"},
		"history":    {"", "", "", "", "", "history", "ancient", "modern", "past", "future"},
	}
	
	// 㻰
	topicScores := make(map[string]int)
	for topic, keywords := range topicKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				topicScores[topic]++
			}
		}
	}
	
	// 
	if len(history) > 0 {
		for _, msg := range history[max(0, len(history)-5):] { // 5
			msgContent := strings.ToLower(msg.Content)
			for topic, keywords := range topicKeywords {
				for _, keyword := range keywords {
					if strings.Contains(msgContent, keyword) {
						topicScores[topic]++ // 
					}
				}
			}
		}
	}
	
	// 
	for topic, score := range topicScores {
		if score > 0 {
			topics = append(topics, topic)
		}
	}
	
	// 
	if len(topics) > 5 {
		topics = topics[:5]
	}
	
	return topics
}

// calculateComplexity 
func (cm *ContextManager) calculateComplexity(message string, entities map[string]interface{}) float64 {
	complexity := 0.0
	
	// 
	wordCount := len(strings.Fields(message))
	if wordCount > 50 {
		complexity += 0.3
	} else if wordCount > 20 {
		complexity += 0.2
	} else if wordCount > 10 {
		complexity += 0.1
	}
	
	// 
	entityCount := len(entities)
	complexity += float64(entityCount) * 0.1
	
	// 
	sentenceCount := strings.Count(message, "") + strings.Count(message, ".") + strings.Count(message, "") + strings.Count(message, "?")
	if sentenceCount > 3 {
		complexity += 0.2
	}
	
	// 
	technicalWords := []string{"", "", "㷨", "", "", "philosophy", "technology", "algorithm", "system", "architecture"}
	for _, word := range technicalWords {
		if strings.Contains(strings.ToLower(message), word) {
			complexity += 0.1
		}
	}
	
	// -11
	if complexity > 1.0 {
		complexity = 1.0
	} else if complexity < -1.0 {
		complexity = -1.0
	}
	
	return complexity
}

// calculateUrgency 
func (cm *ContextManager) calculateUrgency(message string, intent string) float64 {
	urgency := 0.0
	message = strings.ToLower(message)
	
	// 
	urgentWords := []string{"o", "o", "", "", "", "urgent", "immediately", "now", "asap", "quickly"}
	for _, word := range urgentWords {
		if strings.Contains(message, word) {
			urgency += 0.3
		}
	}
	
	// 
	switch intent {
	case "complaint":
		urgency += 0.4
	case "request":
		urgency += 0.2
	case "question":
		urgency += 0.1
	}
	
	// 
	strongEmotions := []string{"", "", "", "o", "really", "very", "extremely", "quite"}
	for _, word := range strongEmotions {
		if strings.Contains(message, word) {
			urgency += 0.1
		}
	}
	
	// -11
	if urgency > 1.0 {
		urgency = 1.0
	} else if urgency < -1.0 {
		urgency = -1.0
	}
	
	return urgency
}

// analyzeEmotionalState 
func (cm *ContextManager) analyzeEmotionalState(message string, sentiment string) string {
	message = strings.ToLower(message)
	
	// 
	emotionKeywords := map[string][]string{
		"excited":    {"", "", "_", "d^", "excited", "thrilled", "happy", "joyful"},
		"anxious":    {"", "", "o", "", "anxious", "worried", "nervous", "concerned"},
		"confused":   {"", "", "", "_", "confused", "puzzled", "unclear", "lost"},
		"frustrated": {"", "", "", "", "frustrated", "disappointed", "annoyed", "upset"},
		"curious":    {"", "", "d", "", "curious", "interested", "wondering", "exploring"},
		"calm":       {"", "侲", "", "", "calm", "peaceful", "relaxed", "serene"},
	}
	
	for emotion, keywords := range emotionKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				return emotion
			}
		}
	}
	
	// 
	switch sentiment {
	case "positive":
		return "content"
	case "negative":
		return "concerned"
	default:
		return "neutral"
	}
}

// calculateConfidence 
func (cm *ContextManager) calculateConfidence(intent string, message string, history []models.ChatMessage) float64 {
	confidence := 0.5 // 
	
	// 
	wordCount := len(strings.Fields(message))
	if wordCount > 5 && wordCount < 100 {
		confidence += 0.2
	}
	
	// 
	clearIntents := []string{"question", "request", "greeting", "cultural", "learning"}
	for _, clearIntent := range clearIntents {
		if intent == clearIntent {
			confidence += 0.2
			break
		}
	}
	
	// 
	if len(history) > 0 {
		// 
		recentMessages := history[max(0, len(history)-3):]
		consistentTopics := 0
		for _, msg := range recentMessages {
			if strings.Contains(strings.ToLower(msg.Content), strings.ToLower(message[:min(20, len(message))])) {
				consistentTopics++
			}
		}
		if consistentTopics > 0 {
			confidence += 0.1
		}
	}
	
	// -11
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}
	
	return confidence
}

// buildEnhancedContextualPrompt 
func (cm *ContextManager) buildEnhancedContextualPrompt(context *ConversationContext, currentMessage string) string {
	prompt := ""
	
	// 
	if len(context.PersonalityProfile) > 0 {
		prompt += "\n"
		
		// 
		if sentiment, exists := context.PersonalityProfile["sentiment_tendency"]; exists {
			if sentimentMap, ok := sentiment.(map[string]float64); ok {
				var dominantSentiment string
				var maxCount float64
				for sent, count := range sentimentMap {
					if count > maxCount {
						maxCount = count
						dominantSentiment = sent
					}
				}
				prompt += fmt.Sprintf("- %s\n", dominantSentiment)
			}
		}
		
		// 
		if intent, exists := context.PersonalityProfile["intent_preference"]; exists {
			if intentMap, ok := intent.(map[string]float64); ok {
				var topIntents []string
				for int, count := range intentMap {
					if count >= 2 { // 2
						topIntents = append(topIntents, int)
					}
				}
				if len(topIntents) > 0 {
					prompt += fmt.Sprintf("- %s\n", strings.Join(topIntents, ""))
				}
			}
		}
	}
	
	// 
	if context.Summary != "" {
		prompt += fmt.Sprintf("\n%s\n", context.Summary)
	}
	
	// 
	if len(context.MemoryBank) > 0 {
		prompt += "\n\n"
		importantMemories := 0
		for _, memory := range context.MemoryBank {
			if memory.Importance > 0.7 && importantMemories < 3 {
				prompt += fmt.Sprintf("- %s\n", memory.Content)
				importantMemories++
			}
		}
	}
	
	// 
	if context.Intent != "unknown" && context.Intent != "general" {
		prompt += fmt.Sprintf("\n%s\n", context.Intent)
	}
	
	if context.Sentiment != "neutral" {
		prompt += fmt.Sprintf("%s\n", context.Sentiment)
	}
	
	if len(context.Topics) > 0 {
		prompt += fmt.Sprintf("%s\n", strings.Join(context.Topics, ""))
	}
	
	// 
	if len(context.ContextWindow) > 0 {
		prompt += "\n\n"
		
		// 
		importantMessages := make([]ContextMessage, 0)
		for _, msg := range context.ContextWindow {
			if msg.Importance > 0.6 {
				importantMessages = append(importantMessages, msg)
			}
		}
		
		// 
		if len(importantMessages) < 4 {
			recentStart := max(0, len(context.ContextWindow)-4)
			for i := recentStart; i < len(context.ContextWindow); i++ {
				found := false
				for _, imp := range importantMessages {
					if imp.Timestamp.Equal(context.ContextWindow[i].Timestamp) {
						found = true
						break
					}
				}
				if !found {
					importantMessages = append(importantMessages, context.ContextWindow[i])
				}
			}
		}
		
		// 
		displayCount := min(6, len(importantMessages))
		for i := 0; i < displayCount; i++ {
			msg := importantMessages[i]
			role := ""
			if msg.Role == "assistant" {
				role = ""
			}
			prompt += fmt.Sprintf("%s%s\n", role, msg.Content)
		}
	}
	
	// 
	if len(context.ConversationFlow) > 0 {
		recentFlow := context.ConversationFlow[max(0, len(context.ConversationFlow)-5):]
		prompt += fmt.Sprintf("\n%s\n", strings.Join(recentFlow, "  "))
	}
	
	prompt += fmt.Sprintf("\n%s\n", currentMessage)
	
	// 
	prompt += "\n\n"
	
	return prompt
}

// max 
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// mergeKeywords 
func (cm *ContextManager) mergeKeywords(existing, new []string) []string {
	keywordMap := make(map[string]bool)
	
	// 
	for _, keyword := range existing {
		keywordMap[keyword] = true
	}
	
	// 
	for _, keyword := range new {
		keywordMap[keyword] = true
	}
	
	// 	
	merged := make([]string, 0, len(keywordMap))
	for keyword := range keywordMap {
		merged = append(merged, keyword)
	}
	
	// 
	if len(merged) > 20 {
		merged = merged[:20]
	}
	
	return merged
}

// mergeTopics 
func (cm *ContextManager) mergeTopics(existing, new []string) []string {
	topicMap := make(map[string]bool)
	
	for _, topic := range existing {
		topicMap[topic] = true
	}
	
	for _, topic := range new {
		topicMap[topic] = true
	}
	
	merged := make([]string, 0, len(topicMap))
	for topic := range topicMap {
		merged = append(merged, topic)
	}
	
	return merged
}

