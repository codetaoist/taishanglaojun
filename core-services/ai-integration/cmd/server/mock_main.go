package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatSession 聊天会话模型
type ChatSession struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Title       string        `json:"title"`
	Messages    []ChatMessage `json:"messages"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	IsActive    bool          `json:"is_active"`
}

// AIRequest AI请求模型
type AIRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id"`
	Model     string `json:"model,omitempty"`
}

// AIResponse AI响应模型
type AIResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	MessageID string `json:"message_id"`
	Model     string `json:"model"`
	Usage     struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// IntentRequest 意图识别请求
type IntentRequest struct {
	Text   string `json:"text"`
	UserID string `json:"user_id"`
}

// IntentResponse 意图识别响应
type IntentResponse struct {
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
	Entities   []struct {
		Entity string `json:"entity"`
		Value  string `json:"value"`
	} `json:"entities"`
}

// SentimentRequest 情感分析请求
type SentimentRequest struct {
	Text   string `json:"text"`
	UserID string `json:"user_id"`
}

// SentimentResponse 情感分析响应
type SentimentResponse struct {
	Sentiment  string  `json:"sentiment"` // positive, negative, neutral
	Confidence float64 `json:"confidence"`
	Score      float64 `json:"score"` // -1 to 1
}

// 模拟数据
var sessions []ChatSession
var messages []ChatMessage

func initMockData() {
	// 初始化一些示例会话
	sessionID1 := uuid.New().String()
	sessionID2 := uuid.New().String()

	sessions = []ChatSession{
		{
			ID:        sessionID1,
			UserID:    "user123",
			Title:     "关于道德经的讨论",
			CreatedAt: time.Now().Add(-time.Hour * 2),
			UpdatedAt: time.Now().Add(-time.Minute * 30),
			IsActive:  true,
		},
		{
			ID:        sessionID2,
			UserID:    "user123",
			Title:     "儒家思想学习",
			CreatedAt: time.Now().Add(-time.Hour * 24),
			UpdatedAt: time.Now().Add(-time.Hour * 23),
			IsActive:  false,
		},
	}

	messages = []ChatMessage{
		{
			ID:        uuid.New().String(),
			SessionID: sessionID1,
			Role:      "user",
			Content:   "请解释一下道德经中的\"道可道，非常道\"",
			Timestamp: time.Now().Add(-time.Hour * 2),
		},
		{
			ID:        uuid.New().String(),
			SessionID: sessionID1,
			Role:      "assistant",
			Content:   "这句话是道德经的开篇，表达了老子对\"道\"的深刻理解。\"道可道，非常道\"意思是：能够用语言表达出来的道，就不是永恒不变的道。老子认为真正的\"道\"是超越语言和概念的，是宇宙万物的根本规律。",
			Timestamp: time.Now().Add(-time.Hour * 2).Add(time.Minute * 2),
		},
	}
}

// 智能对话
func chat(c *gin.Context) {
	var req AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 创建或获取会话
	var sessionID string
	if req.SessionID == "" {
		sessionID = uuid.New().String()
		// 创建新会话
		newSession := ChatSession{
			ID:        sessionID,
			UserID:    req.UserID,
			Title:     truncateString(req.Message, 20),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}
		sessions = append(sessions, newSession)
	} else {
		sessionID = req.SessionID
		// 更新会话时间
		for i, session := range sessions {
			if session.ID == sessionID {
				sessions[i].UpdatedAt = time.Now()
				break
			}
		}
	}

	// 添加用户消息
	userMessage := ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	messages = append(messages, userMessage)

	// 生成AI回复（简单的模拟逻辑）
	aiResponse := generateAIResponse(req.Message)
	assistantMessage := ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now().Add(time.Second * 2),
	}
	messages = append(messages, assistantMessage)

	response := AIResponse{
		Message:   aiResponse,
		SessionID: sessionID,
		MessageID: assistantMessage.ID,
		Model:     "mock-gpt-3.5",
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     len(req.Message) / 4,
			CompletionTokens: len(aiResponse) / 4,
			TotalTokens:      (len(req.Message) + len(aiResponse)) / 4,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "对话成功",
		"data":    response,
	})
}

// 获取会话列表
func getSessions(c *gin.Context) {
	userID := c.Query("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filteredSessions := make([]ChatSession, 0)
	for _, session := range sessions {
		if userID == "" || session.UserID == userID {
			// 获取会话的消息
			sessionMessages := make([]ChatMessage, 0)
			for _, msg := range messages {
				if msg.SessionID == session.ID {
					sessionMessages = append(sessionMessages, msg)
				}
			}
			session.Messages = sessionMessages
			filteredSessions = append(filteredSessions, session)
		}
	}

	total := len(filteredSessions)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取会话列表成功",
		"data": gin.H{
			"sessions":    filteredSessions[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 获取会话消息
func getMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	sessionMessages := make([]ChatMessage, 0)
	for _, msg := range messages {
		if msg.SessionID == sessionID {
			sessionMessages = append(sessionMessages, msg)
		}
	}

	total := len(sessionMessages)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取消息列表成功",
		"data": gin.H{
			"messages":    sessionMessages[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 意图识别
func intentRecognition(c *gin.Context) {
	var req IntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 简单的意图识别逻辑
	intent := analyzeIntent(req.Text)
	
	response := IntentResponse{
		Intent:     intent.Intent,
		Confidence: intent.Confidence,
		Entities:   intent.Entities,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "意图识别成功",
		"data":    response,
	})
}

// 情感分析
func sentimentAnalysis(c *gin.Context) {
	var req SentimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 简单的情感分析逻辑
	sentiment := analyzeSentiment(req.Text)
	
	response := SentimentResponse{
		Sentiment:  sentiment.Sentiment,
		Confidence: sentiment.Confidence,
		Score:      sentiment.Score,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "情感分析成功",
		"data":    response,
	})
}

// 获取提供商信息
func getProviders(c *gin.Context) {
	providers := []string{"OpenAI", "Claude", "Local"}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取提供商信息成功",
		"data": gin.H{
			"providers": providers,
			"default":   "OpenAI",
		},
	})
}

// 获取模型信息
func getModels(c *gin.Context) {
	models := map[string][]string{
		"OpenAI": {"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"},
		"Claude": {"claude-3-haiku", "claude-3-sonnet", "claude-3-opus"},
		"Local":  {"llama2", "chatglm"},
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取模型信息成功",
		"data":    models,
	})
}

// 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "ai-integration",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
		"providers": gin.H{
			"OpenAI": "available",
			"Claude": "available",
			"Local":  "available",
		},
	})
}

// 辅助函数：生成AI回复
func generateAIResponse(userMessage string) string {
	responses := []string{
		"这是一个很有深度的问题。从传统文化的角度来看，",
		"根据古代智慧的指导，我认为",
		"这让我想起了古代圣贤的教诲：",
		"从哲学的角度分析，",
		"这个问题涉及到深层的人生智慧，",
	}
	
	// 简单的关键词匹配回复
	if contains(userMessage, "道德经") || contains(userMessage, "老子") {
		return "道德经是中华文化的瑰宝，老子的智慧至今仍有重要指导意义。" + responses[0] + "道法自然是其核心思想之一。"
	}
	if contains(userMessage, "孔子") || contains(userMessage, "儒家") {
		return "孔子的思想强调仁爱和礼制，" + responses[1] + "修身齐家治国平天下是儒家的理想境界。"
	}
	if contains(userMessage, "佛") || contains(userMessage, "禅") {
		return "佛家智慧注重内心的觉悟和解脱，" + responses[2] + "苦集灭道四谛是佛陀的根本教法。"
	}
	
	return responses[0] + "这需要我们深入思考和实践。您还有什么想了解的吗？"
}

// 辅助函数：分析意图
func analyzeIntent(text string) IntentResponse {
	if contains(text, "学习") || contains(text, "了解") {
		return IntentResponse{
			Intent:     "learning",
			Confidence: 0.85,
			Entities: []struct {
				Entity string `json:"entity"`
				Value  string `json:"value"`
			}{
				{Entity: "action", Value: "learn"},
			},
		}
	}
	if contains(text, "问题") || contains(text, "帮助") {
		return IntentResponse{
			Intent:     "help",
			Confidence: 0.90,
			Entities: []struct {
				Entity string `json:"entity"`
				Value  string `json:"value"`
			}{
				{Entity: "action", Value: "help"},
			},
		}
	}
	return IntentResponse{
		Intent:     "general",
		Confidence: 0.70,
		Entities:   []struct {
			Entity string `json:"entity"`
			Value  string `json:"value"`
		}{},
	}
}

// 辅助函数：分析情感
func analyzeSentiment(text string) SentimentResponse {
	positiveWords := []string{"好", "棒", "优秀", "喜欢", "满意", "开心"}
	negativeWords := []string{"坏", "差", "糟糕", "讨厌", "不满", "难过"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if contains(text, word) {
			positiveCount++
		}
	}
	for _, word := range negativeWords {
		if contains(text, word) {
			negativeCount++
		}
	}
	
	if positiveCount > negativeCount {
		return SentimentResponse{
			Sentiment:  "positive",
			Confidence: 0.80,
			Score:      0.6,
		}
	} else if negativeCount > positiveCount {
		return SentimentResponse{
			Sentiment:  "negative",
			Confidence: 0.80,
			Score:      -0.6,
		}
	}
	
	return SentimentResponse{
		Sentiment:  "neutral",
		Confidence: 0.75,
		Score:      0.0,
	}
}

// 辅助函数：检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func main() {
	// 初始化模拟数据
	initMockData()

	// 创建Gin路由器
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API路由组
	api := r.Group("/api/v1")
	{
		// AI对话相关路由
		ai := api.Group("/ai")
		{
			ai.POST("/chat", chat)
			ai.GET("/sessions", getSessions)
			ai.GET("/sessions/:session_id/messages", getMessages)
			ai.POST("/intent", intentRecognition)
			ai.POST("/sentiment", sentimentAnalysis)
			ai.GET("/providers", getProviders)
			ai.GET("/models", getModels)
			ai.GET("/health", healthCheck)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ai-integration",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Println("AI集成服务 (Mock版本) 启动在端口 8083")
	log.Fatal(r.Run(":8083"))
}