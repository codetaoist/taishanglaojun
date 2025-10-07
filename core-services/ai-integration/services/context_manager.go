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

// ContextManager 上下文管理器
type ContextManager struct {
	db     *gorm.DB
	logger *zap.Logger
}

// ConversationContext 对话上下文
type ConversationContext struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	SessionID     string                 `json:"session_id" gorm:"index"`
	UserID        uint                   `json:"user_id" gorm:"index"`
	Intent        string                 `json:"intent"`                // 用户意图
	Entities      map[string]interface{} `json:"entities" gorm:"type:json"` // 实体信息
	Sentiment     string                 `json:"sentiment"`             // 情感分析结果
	Topics        []string               `json:"topics" gorm:"type:json"`   // 话题标签
	Keywords      []string               `json:"keywords" gorm:"type:json"` // 关键词
	Summary       string                 `json:"summary"`               // 对话摘要
	LastUpdated   time.Time              `json:"last_updated"`
	MessageCount  int                    `json:"message_count"`
	ContextWindow []ContextMessage       `json:"context_window" gorm:"type:json"` // 上下文窗口
	// 新增字段
	PersonalityProfile map[string]interface{} `json:"personality_profile" gorm:"type:json"` // 用户个性化档案
	ConversationFlow   []string               `json:"conversation_flow" gorm:"type:json"`   // 对话流程记录
	UserPreferences    map[string]interface{} `json:"user_preferences" gorm:"type:json"`    // 用户偏好
	MemoryBank         []MemoryItem           `json:"memory_bank" gorm:"type:json"`         // 长期记忆库
	ContextScore       float64                `json:"context_score"`                        // 上下文相关性评分
}

// ContextMessage 上下文消息
type ContextMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Intent    string    `json:"intent,omitempty"`
	Entities  map[string]interface{} `json:"entities,omitempty"`
	// 新增字段
	Importance float64 `json:"importance"` // 消息重要性评分
	Reference  string  `json:"reference"`  // 引用的历史消息ID
}

// MemoryItem 记忆项
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

// IntentAnalysisResult 意图分析结果
type IntentAnalysisResult struct {
	Intent     string                 `json:"intent"`
	Confidence float64                `json:"confidence"`
	Entities   map[string]interface{} `json:"entities"`
	Sentiment  string                 `json:"sentiment"`
	Keywords   []string               `json:"keywords"`
	Topics     []string               `json:"topics"`
	// 新增字段
	EmotionalState string  `json:"emotional_state"` // 情绪状态
	Urgency        float64 `json:"urgency"`         // 紧急程度
	Complexity     float64 `json:"complexity"`      // 复杂度
}

// NewContextManager 创建上下文管理器
func NewContextManager(db *gorm.DB, logger *zap.Logger) *ContextManager {
	return &ContextManager{
		db:     db,
		logger: logger,
	}
}

// GetOrCreateContext 获取或创建对话上下文
func (cm *ContextManager) GetOrCreateContext(ctx context.Context, sessionID string, userID uint) (*ConversationContext, error) {
	var context ConversationContext
	
	err := cm.db.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&context).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的上下文
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

// UpdateContext 更新对话上下文
func (cm *ContextManager) UpdateContext(ctx context.Context, sessionID string, userMessage, aiResponse string, analysis *IntentAnalysisResult) error {
	var context ConversationContext
	
	err := cm.db.Where("session_id = ?", sessionID).First(&context).Error
	if err != nil {
		return fmt.Errorf("failed to get context: %w", err)
	}

	// 更新基础信息
	context.Intent = analysis.Intent
	context.Sentiment = analysis.Sentiment
	context.Keywords = cm.mergeKeywords(context.Keywords, analysis.Keywords)
	context.Topics = cm.mergeTopics(context.Topics, analysis.Topics)
	context.MessageCount++
	context.LastUpdated = time.Now()

	// 计算消息重要性
	userImportance := cm.calculateMessageImportance(userMessage, analysis)
	aiImportance := cm.calculateMessageImportance(aiResponse, nil)

	// 添加到上下文窗口
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

	// 维护上下文窗口大小（保留最重要的消息）
	if len(context.ContextWindow) > 20 {
		context.ContextWindow = cm.pruneContextWindow(context.ContextWindow, 16)
	}

	// 更新记忆库
	cm.updateMemoryBank(&context, userMessage, analysis)

	// 更新用户个性化档案
	cm.updatePersonalityProfile(&context, analysis)

	// 更新对话流程
	context.ConversationFlow = append(context.ConversationFlow, analysis.Intent)
	if len(context.ConversationFlow) > 50 {
		context.ConversationFlow = context.ConversationFlow[len(context.ConversationFlow)-50:]
	}

	// 生成新的摘要
	context.Summary = cm.generateEnhancedSummary(context.ContextWindow, context.MemoryBank)

	// 计算上下文相关性评分
	context.ContextScore = cm.calculateContextScore(&context)

	// 保存更新
	if err := cm.db.Save(&context).Error; err != nil {
		return fmt.Errorf("failed to update context: %w", err)
	}

	return nil
}

// calculateMessageImportance 计算消息重要性
func (cm *ContextManager) calculateMessageImportance(message string, analysis *IntentAnalysisResult) float64 {
	importance := 0.5 // 基础重要性

	// 基于消息长度
	if len(message) > 100 {
		importance += 0.1
	}

	// 基于意图分析
	if analysis != nil {
		importance += analysis.Confidence * 0.3
		importance += analysis.Urgency * 0.2
		importance += analysis.Complexity * 0.1
	}

	// 基于关键词密度
	keywordCount := len(strings.Fields(message))
	if keywordCount > 10 {
		importance += 0.1
	}

	// 限制在0-1范围内
	if importance > 1.0 {
		importance = 1.0
	}
	if importance < 0.0 {
		importance = 0.0
	}

	return importance
}

// pruneContextWindow 修剪上下文窗口，保留最重要的消息
func (cm *ContextManager) pruneContextWindow(messages []ContextMessage, targetSize int) []ContextMessage {
	if len(messages) <= targetSize {
		return messages
	}

	// 按重要性排序
	sortedMessages := make([]ContextMessage, len(messages))
	copy(sortedMessages, messages)

	// 简单的重要性排序（实际应用中可以使用更复杂的算法）
	for i := 0; i < len(sortedMessages)-1; i++ {
		for j := i + 1; j < len(sortedMessages); j++ {
			if sortedMessages[i].Importance < sortedMessages[j].Importance {
				sortedMessages[i], sortedMessages[j] = sortedMessages[j], sortedMessages[i]
			}
		}
	}

	// 保留最重要的消息，但确保对话的连贯性
	result := make([]ContextMessage, 0, targetSize)
	
	// 总是保留最近的几条消息
	recentCount := targetSize / 4
	if recentCount < 2 {
		recentCount = 2
	}
	
	// 添加最近的消息
	startIdx := len(messages) - recentCount
	if startIdx < 0 {
		startIdx = 0
	}
	result = append(result, messages[startIdx:]...)

	// 添加重要的历史消息
	remainingSlots := targetSize - len(result)
	for i := 0; i < len(sortedMessages) && remainingSlots > 0; i++ {
		msg := sortedMessages[i]
		// 避免重复添加最近的消息
		if msg.Timestamp.Before(messages[startIdx].Timestamp) {
			result = append(result, msg)
			remainingSlots--
		}
	}

	return result
}

// updateMemoryBank 更新记忆库
func (cm *ContextManager) updateMemoryBank(context *ConversationContext, message string, analysis *IntentAnalysisResult) {
	// 提取重要信息作为记忆
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

		// 限制记忆库大小
		if len(context.MemoryBank) > 100 {
			// 移除最不重要且最少访问的记忆
			context.MemoryBank = cm.pruneMemoryBank(context.MemoryBank, 80)
		}
	}
}

// pruneMemoryBank 修剪记忆库
func (cm *ContextManager) pruneMemoryBank(memories []MemoryItem, targetSize int) []MemoryItem {
	if len(memories) <= targetSize {
		return memories
	}

	// 计算记忆的综合评分
	for i := range memories {
		// 基于重要性、访问次数和时间衰减的综合评分
		timeFactor := 1.0 - float64(time.Since(memories[i].LastAccess).Hours())/8760 // 一年衰减
		accessFactor := float64(memories[i].AccessCount) / 10.0
		if accessFactor > 1.0 {
			accessFactor = 1.0
		}
		
		memories[i].Importance = memories[i].Importance*0.5 + timeFactor*0.3 + accessFactor*0.2
	}

	// 按综合评分排序并保留前targetSize个
	for i := 0; i < len(memories)-1; i++ {
		for j := i + 1; j < len(memories); j++ {
			if memories[i].Importance < memories[j].Importance {
				memories[i], memories[j] = memories[j], memories[i]
			}
		}
	}

	return memories[:targetSize]
}

// updatePersonalityProfile 更新用户个性化档案
func (cm *ContextManager) updatePersonalityProfile(context *ConversationContext, analysis *IntentAnalysisResult) {
	// 更新情感倾向
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

	// 更新意图偏好
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

// calculateContextScore 计算上下文相关性评分
func (cm *ContextManager) calculateContextScore(context *ConversationContext) float64 {
	score := 0.0

	// 基于消息数量
	messageScore := float64(context.MessageCount) / 100.0
	if messageScore > 1.0 {
		messageScore = 1.0
	}
	score += messageScore * 0.2

	// 基于记忆库丰富度
	memoryScore := float64(len(context.MemoryBank)) / 50.0
	if memoryScore > 1.0 {
		memoryScore = 1.0
	}
	score += memoryScore * 0.3

	// 基于话题多样性
	topicScore := float64(len(context.Topics)) / 20.0
	if topicScore > 1.0 {
		topicScore = 1.0
	}
	score += topicScore * 0.2

	// 基于时间活跃度
	timeSinceLastUpdate := time.Since(context.LastUpdated).Hours()
	timeScore := 1.0 - (timeSinceLastUpdate / 168.0) // 一周内活跃度最高
	if timeScore < 0 {
		timeScore = 0
	}
	score += timeScore * 0.3

	return score
}

// generateEnhancedSummary 生成增强的对话摘要
func (cm *ContextManager) generateEnhancedSummary(contextWindow []ContextMessage, memoryBank []MemoryItem) string {
	if len(contextWindow) == 0 {
		return "新的对话会话"
	}

	// 提取关键信息
	var topics []string
	var intents []string
	
	for _, msg := range contextWindow {
		if msg.Intent != "" && msg.Intent != "unknown" {
			intents = append(intents, msg.Intent)
		}
	}

	// 从记忆库中提取重要主题
	for _, memory := range memoryBank {
		if memory.Importance > 0.7 {
			topics = append(topics, memory.Type)
		}
	}

	// 生成摘要
	summary := "对话涉及"
	if len(intents) > 0 {
		uniqueIntents := cm.removeDuplicates(intents)
		summary += "意图: " + strings.Join(uniqueIntents[:min(3, len(uniqueIntents))], ", ")
	}
	
	if len(topics) > 0 {
		uniqueTopics := cm.removeDuplicates(topics)
		summary += "; 主题: " + strings.Join(uniqueTopics[:min(3, len(uniqueTopics))], ", ")
	}

	return summary
}

// removeDuplicates 移除重复项
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

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AnalyzeIntent 分析用户意图
func (cm *ContextManager) AnalyzeIntent(ctx context.Context, message string, conversationHistory []models.ChatMessage) (*IntentAnalysisResult, error) {
	// 增强的意图识别逻辑
	intent := cm.detectIntent(message)
	entities := cm.extractEntities(message)
	sentiment := cm.analyzeSentiment(message)
	keywords := cm.extractKeywords(message)
	topics := cm.identifyTopics(message, conversationHistory)
	
	// 计算复杂度和紧急程度
	complexity := cm.calculateComplexity(message, entities)
	urgency := cm.calculateUrgency(message, intent)
	emotionalState := cm.analyzeEmotionalState(message, sentiment)
	
	// 基于历史对话调整置信度
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

// GetContextualPrompt 获取上下文化的提示词
func (cm *ContextManager) GetContextualPrompt(ctx context.Context, sessionID string, currentMessage string) (string, error) {
	var context ConversationContext
	
	err := cm.db.Where("session_id = ?", sessionID).First(&context).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return currentMessage, nil // 没有上下文，返回原消息
		}
		return "", fmt.Errorf("failed to get context: %w", err)
	}
	
	// 构建增强的上下文化提示词
	prompt := cm.buildEnhancedContextualPrompt(&context, currentMessage)
	return prompt, nil
}

// detectIntent 检测用户意图
func (cm *ContextManager) detectIntent(message string) string {
	message = strings.ToLower(message)
	
	// 增强的关键词匹配
	intentKeywords := map[string][]string{
		"question":     {"什么", "如何", "怎么", "为什么", "哪里", "谁", "什么时候", "when", "what", "how", "why", "where", "who", "?", "？"},
		"request":      {"请", "帮我", "能否", "可以", "希望", "想要", "需要", "please", "help", "can you", "could you", "would you", "need"},
		"greeting":     {"你好", "hello", "hi", "早上好", "晚上好", "再见", "goodbye", "bye", "嗨"},
		"complaint":    {"不满", "抱怨", "问题", "错误", "bug", "故障", "complaint", "issue", "problem", "error", "wrong"},
		"praise":       {"好", "棒", "优秀", "感谢", "谢谢", "赞", "great", "excellent", "thank", "awesome", "wonderful"},
		"cultural":     {"文化", "传统", "智慧", "哲学", "儒家", "道家", "佛家", "文学", "文化", "wisdom", "philosophy", "tradition"},
		"learning":     {"学习", "教", "学", "知识", "了解", "理解", "learn", "study", "knowledge", "understand", "explain"},
		"emotional":    {"感觉", "情绪", "心情", "难过", "开心", "焦虑", "feel", "emotion", "mood", "sad", "happy", "anxious"},
		"planning":     {"计划", "安排", "准备", "打算", "plan", "schedule", "prepare", "intend", "organize"},
		"comparison":   {"比较", "对比", "区别", "不同", "相同", "compare", "difference", "similar", "versus", "vs"},
	}
	
	// 计算每个意图的匹配分数
	intentScores := make(map[string]int)
	for intent, keywords := range intentKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				intentScores[intent]++
			}
		}
	}
	
	// 找到最高分的意图
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

// extractEntities 提取实体信息
func (cm *ContextManager) extractEntities(message string) map[string]interface{} {
	entities := make(map[string]interface{})
	
	// 时间实体提取
	timeKeywords := []string{"今天", "明天", "昨天", "现在", "今年", "去年", "明年", "这周", "下周", "上周", "今晚", "明早"}
	for _, keyword := range timeKeywords {
		if strings.Contains(message, keyword) {
			if entities["time"] == nil {
				entities["time"] = []string{}
			}
			entities["time"] = append(entities["time"].([]string), keyword)
		}
	}
	
	// 文化相关实体
	culturalEntities := []string{"孔子", "老子", "庄子", "佛陀", "儒家", "道家", "佛家", "法家", "墨家", "兵家", "诗经", "论语", "道德经"}
	for _, entity := range culturalEntities {
		if strings.Contains(message, entity) {
			if entities["cultural"] == nil {
				entities["cultural"] = []string{}
			}
			entities["cultural"] = append(entities["cultural"].([]string), entity)
		}
	}
	
	// 地点实体
	locationEntities := []string{"北京", "上海", "广州", "深圳", "杭州", "南京", "西安", "成都", "中国", "美国", "日本"}
	for _, entity := range locationEntities {
		if strings.Contains(message, entity) {
			if entities["location"] == nil {
				entities["location"] = []string{}
			}
			entities["location"] = append(entities["location"].([]string), entity)
		}
	}
	
	// 数字实体提取
	numbers := []string{"一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
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

// analyzeSentiment 分析情感
func (cm *ContextManager) analyzeSentiment(message string) string {
	message = strings.ToLower(message)
	
	// 扩展的情感词典
	positiveWords := []string{
		"好", "棒", "优秀", "喜欢", "满意", "开心", "高兴", "感谢", "赞", "爱", "美好", "完美", "惊喜",
		"great", "good", "excellent", "happy", "satisfied", "love", "amazing", "wonderful", "fantastic",
	}
	negativeWords := []string{
		"不好", "差", "糟糕", "不满", "生气", "失望", "讨厌", "难过", "痛苦", "烦恼", "焦虑", "担心",
		"bad", "terrible", "angry", "disappointed", "hate", "sad", "worried", "anxious", "frustrated",
	}
	
	positiveCount := 0
	negativeCount := 0
	
	// 计算情感词汇出现次数
	for _, word := range positiveWords {
		positiveCount += strings.Count(message, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(message, word)
	}
	
	// 考虑否定词的影响
	negationWords := []string{"不", "没", "非", "无", "not", "no", "never", "neither"}
	hasNegation := false
	for _, neg := range negationWords {
		if strings.Contains(message, neg) {
			hasNegation = true
			break
		}
	}
	
	// 如果有否定词，情感可能相反
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

// extractKeywords 提取关键词
func (cm *ContextManager) extractKeywords(message string) []string {
	words := strings.Fields(message)
	keywords := []string{}
	
	// 扩展的停用词列表
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true, "你": true, "他": true, "她": true,
		"这": true, "那": true, "有": true, "和": true, "与": true, "或": true, "但": true, "而": true,
		"and": true, "the": true, "is": true, "in": true, "to": true, "of": true, "a": true, "that": true,
		"it": true, "with": true, "for": true, "as": true, "was": true, "on": true, "are": true, "you": true,
	}
	
	// 重要词汇权重
	importantWords := map[string]bool{
		"文化": true, "智慧": true, "哲学": true, "传统": true, "学习": true, "知识": true, "理解": true,
		"思考": true, "探索": true, "发现": true, "创新": true, "成长": true, "进步": true,
	}
	
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		// 移除标点符号
		punctuation := `.,!?;:"'()[]{}，。！？；：""''（）【】`
		word = strings.Trim(word, punctuation)
		
		if len(word) > 1 && !stopWords[word] {
			keywords = append(keywords, word)
			// 重要词汇可以重复添加以提高权重
			if importantWords[word] {
				keywords = append(keywords, word)
			}
		}
	}
	
	// 去重并限制数量
	uniqueKeywords := cm.removeDuplicates(keywords)
	if len(uniqueKeywords) > 15 {
		uniqueKeywords = uniqueKeywords[:15]
	}
	
	return uniqueKeywords
}

// identifyTopics 识别话题
func (cm *ContextManager) identifyTopics(message string, history []models.ChatMessage) []string {
	topics := []string{}
	message = strings.ToLower(message)
	
	// 扩展的话题关键词
	topicKeywords := map[string][]string{
		"philosophy": {"哲学", "思想", "智慧", "理念", "观念", "philosophy", "wisdom", "thought", "concept", "idea"},
		"culture":    {"文化", "传统", "习俗", "民俗", "文明", "文化", "tradition", "custom", "civilization", "heritage"},
		"religion":   {"宗教", "佛教", "道教", "儒教", "信仰", "religion", "buddhism", "taoism", "confucianism", "belief"},
		"technology": {"技术", "科技", "AI", "人工智能", "计算机", "technology", "artificial intelligence", "computer", "digital"},
		"life":       {"生活", "人生", "生命", "存在", "意义", "life", "living", "existence", "meaning", "purpose"},
		"learning":   {"学习", "教育", "知识", "研究", "探索", "learning", "education", "knowledge", "research", "study"},
		"emotion":    {"情感", "情绪", "心理", "感受", "体验", "emotion", "feeling", "psychology", "experience", "mood"},
		"health":     {"健康", "养生", "医疗", "身体", "心理健康", "health", "wellness", "medical", "physical", "mental health"},
		"art":        {"艺术", "美术", "音乐", "文学", "诗歌", "art", "music", "literature", "poetry", "painting"},
		"history":    {"历史", "古代", "现代", "过去", "未来", "history", "ancient", "modern", "past", "future"},
	}
	
	// 计算话题匹配分数
	topicScores := make(map[string]int)
	for topic, keywords := range topicKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				topicScores[topic]++
			}
		}
	}
	
	// 从历史对话中提取话题上下文
	if len(history) > 0 {
		for _, msg := range history[max(0, len(history)-5):] { // 最近5条消息
			msgContent := strings.ToLower(msg.Content)
			for topic, keywords := range topicKeywords {
				for _, keyword := range keywords {
					if strings.Contains(msgContent, keyword) {
						topicScores[topic]++ // 历史话题权重较低
					}
				}
			}
		}
	}
	
	// 选择得分最高的话题
	for topic, score := range topicScores {
		if score > 0 {
			topics = append(topics, topic)
		}
	}
	
	// 按分数排序并限制数量
	if len(topics) > 5 {
		topics = topics[:5]
	}
	
	return topics
}

// calculateComplexity 计算消息复杂度
func (cm *ContextManager) calculateComplexity(message string, entities map[string]interface{}) float64 {
	complexity := 0.0
	
	// 基于消息长度
	wordCount := len(strings.Fields(message))
	if wordCount > 50 {
		complexity += 0.3
	} else if wordCount > 20 {
		complexity += 0.2
	} else if wordCount > 10 {
		complexity += 0.1
	}
	
	// 基于实体数量
	entityCount := len(entities)
	complexity += float64(entityCount) * 0.1
	
	// 基于句子结构复杂度（简单判断）
	sentenceCount := strings.Count(message, "。") + strings.Count(message, ".") + strings.Count(message, "?") + strings.Count(message, "？")
	if sentenceCount > 3 {
		complexity += 0.2
	}
	
	// 基于专业词汇
	technicalWords := []string{"哲学", "技术", "算法", "系统", "架构", "philosophy", "technology", "algorithm", "system", "architecture"}
	for _, word := range technicalWords {
		if strings.Contains(strings.ToLower(message), word) {
			complexity += 0.1
		}
	}
	
	// 限制在0-1范围内
	if complexity > 1.0 {
		complexity = 1.0
	}
	
	return complexity
}

// calculateUrgency 计算紧急程度
func (cm *ContextManager) calculateUrgency(message string, intent string) float64 {
	urgency := 0.0
	message = strings.ToLower(message)
	
	// 基于紧急词汇
	urgentWords := []string{"紧急", "急", "马上", "立即", "现在", "urgent", "immediately", "now", "asap", "quickly"}
	for _, word := range urgentWords {
		if strings.Contains(message, word) {
			urgency += 0.3
		}
	}
	
	// 基于意图类型
	switch intent {
	case "complaint":
		urgency += 0.4
	case "request":
		urgency += 0.2
	case "question":
		urgency += 0.1
	}
	
	// 基于情感强度
	strongEmotions := []string{"非常", "极其", "特别", "很", "really", "very", "extremely", "quite"}
	for _, word := range strongEmotions {
		if strings.Contains(message, word) {
			urgency += 0.1
		}
	}
	
	// 限制在0-1范围内
	if urgency > 1.0 {
		urgency = 1.0
	}
	
	return urgency
}

// analyzeEmotionalState 分析情绪状态
func (cm *ContextManager) analyzeEmotionalState(message string, sentiment string) string {
	message = strings.ToLower(message)
	
	// 具体情绪状态识别
	emotionKeywords := map[string][]string{
		"excited":    {"兴奋", "激动", "开心", "高兴", "excited", "thrilled", "happy", "joyful"},
		"anxious":    {"焦虑", "担心", "紧张", "不安", "anxious", "worried", "nervous", "concerned"},
		"confused":   {"困惑", "迷惑", "不懂", "不明白", "confused", "puzzled", "unclear", "lost"},
		"frustrated": {"沮丧", "失望", "烦躁", "不满", "frustrated", "disappointed", "annoyed", "upset"},
		"curious":    {"好奇", "想知道", "感兴趣", "探索", "curious", "interested", "wondering", "exploring"},
		"calm":       {"平静", "冷静", "安静", "放松", "calm", "peaceful", "relaxed", "serene"},
	}
	
	for emotion, keywords := range emotionKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				return emotion
			}
		}
	}
	
	// 如果没有找到具体情绪，基于情感倾向返回
	switch sentiment {
	case "positive":
		return "content"
	case "negative":
		return "concerned"
	default:
		return "neutral"
	}
}

// calculateConfidence 计算置信度
func (cm *ContextManager) calculateConfidence(intent string, message string, history []models.ChatMessage) float64 {
	confidence := 0.5 // 基础置信度
	
	// 基于消息长度和清晰度
	wordCount := len(strings.Fields(message))
	if wordCount > 5 && wordCount < 100 {
		confidence += 0.2
	}
	
	// 基于意图明确性
	clearIntents := []string{"question", "request", "greeting", "cultural", "learning"}
	for _, clearIntent := range clearIntents {
		if intent == clearIntent {
			confidence += 0.2
			break
		}
	}
	
	// 基于历史对话一致性
	if len(history) > 0 {
		// 简单的一致性检查
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
	
	// 限制在0-1范围内
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}
	
	return confidence
}

// buildEnhancedContextualPrompt 构建增强的上下文化提示词
func (cm *ContextManager) buildEnhancedContextualPrompt(context *ConversationContext, currentMessage string) string {
	prompt := ""
	
	// 添加用户个性化信息
	if len(context.PersonalityProfile) > 0 {
		prompt += "用户特征：\n"
		
		// 情感倾向
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
				prompt += fmt.Sprintf("- 情感倾向：%s\n", dominantSentiment)
			}
		}
		
		// 意图偏好
		if intent, exists := context.PersonalityProfile["intent_preference"]; exists {
			if intentMap, ok := intent.(map[string]float64); ok {
				var topIntents []string
				for int, count := range intentMap {
					if count >= 2 { // 出现2次以上的意图
						topIntents = append(topIntents, int)
					}
				}
				if len(topIntents) > 0 {
					prompt += fmt.Sprintf("- 常见意图：%s\n", strings.Join(topIntents, "、"))
				}
			}
		}
	}
	
	// 添加对话摘要
	if context.Summary != "" {
		prompt += fmt.Sprintf("\n对话背景：%s\n", context.Summary)
	}
	
	// 添加重要记忆
	if len(context.MemoryBank) > 0 {
		prompt += "\n重要记忆：\n"
		importantMemories := 0
		for _, memory := range context.MemoryBank {
			if memory.Importance > 0.7 && importantMemories < 3 {
				prompt += fmt.Sprintf("- %s\n", memory.Content)
				importantMemories++
			}
		}
	}
	
	// 添加当前上下文信息
	if context.Intent != "unknown" && context.Intent != "general" {
		prompt += fmt.Sprintf("\n当前对话意图：%s\n", context.Intent)
	}
	
	if context.Sentiment != "neutral" {
		prompt += fmt.Sprintf("当前情感状态：%s\n", context.Sentiment)
	}
	
	if len(context.Topics) > 0 {
		prompt += fmt.Sprintf("讨论话题：%s\n", strings.Join(context.Topics, "、"))
	}
	
	// 添加最近的重要对话
	if len(context.ContextWindow) > 0 {
		prompt += "\n最近重要对话：\n"
		
		// 选择重要性最高的几条消息
		importantMessages := make([]ContextMessage, 0)
		for _, msg := range context.ContextWindow {
			if msg.Importance > 0.6 {
				importantMessages = append(importantMessages, msg)
			}
		}
		
		// 如果重要消息不够，添加最近的消息
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
		
		// 限制显示数量
		displayCount := min(6, len(importantMessages))
		for i := 0; i < displayCount; i++ {
			msg := importantMessages[i]
			role := "用户"
			if msg.Role == "assistant" {
				role = "助手"
			}
			prompt += fmt.Sprintf("%s：%s\n", role, msg.Content)
		}
	}
	
	// 添加对话流程分析
	if len(context.ConversationFlow) > 0 {
		recentFlow := context.ConversationFlow[max(0, len(context.ConversationFlow)-5):]
		prompt += fmt.Sprintf("\n对话流程：%s\n", strings.Join(recentFlow, " → "))
	}
	
	prompt += fmt.Sprintf("\n当前问题：%s\n", currentMessage)
	
	// 添加响应指导
	prompt += "\n请基于以上上下文信息，提供个性化、连贯且有帮助的回复。"
	
	return prompt
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// mergeKeywords 合并关键词
func (cm *ContextManager) mergeKeywords(existing, new []string) []string {
	keywordMap := make(map[string]bool)
	
	// 添加现有关键词
	for _, keyword := range existing {
		keywordMap[keyword] = true
	}
	
	// 添加新关键词
	for _, keyword := range new {
		keywordMap[keyword] = true
	}
	
	// 转换回切片
	merged := make([]string, 0, len(keywordMap))
	for keyword := range keywordMap {
		merged = append(merged, keyword)
	}
	
	// 限制数量
	if len(merged) > 20 {
		merged = merged[:20]
	}
	
	return merged
}

// mergeTopics 合并话题
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