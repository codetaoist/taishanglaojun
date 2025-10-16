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

// ChatMessage 
type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatSession 
type ChatSession struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Title       string        `json:"title"`
	Messages    []ChatMessage `json:"messages"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	IsActive    bool          `json:"is_active"`
}

// AIRequest AI
type AIRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id"`
	Model     string `json:"model,omitempty"`
}

// AIResponse AI
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

// IntentRequest 
type IntentRequest struct {
	Text   string `json:"text"`
	UserID string `json:"user_id"`
}

// IntentResponse 
type IntentResponse struct {
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
	Entities   []struct {
		Entity string `json:"entity"`
		Value  string `json:"value"`
	} `json:"entities"`
}

// SentimentRequest 
type SentimentRequest struct {
	Text   string `json:"text"`
	UserID string `json:"user_id"`
}

// SentimentResponse 
type SentimentResponse struct {
	Sentiment  string  `json:"sentiment"` // positive, negative, neutral
	Confidence float64 `json:"confidence"`
	Score      float64 `json:"score"` // -1 to 1
}

// 
var sessions []ChatSession
var messages []ChatMessage

func initMockData() {
	// 
	sessionID1 := uuid.New().String()
	sessionID2 := uuid.New().String()

	sessions = []ChatSession{
		{
			ID:        sessionID1,
			UserID:    "user123",
			Title:     "",
			CreatedAt: time.Now().Add(-time.Hour * 2),
			UpdatedAt: time.Now().Add(-time.Minute * 30),
			IsActive:  true,
		},
		{
			ID:        sessionID2,
			UserID:    "user123",
			Title:     "",
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
			Content:   "\"\"",
			Timestamp: time.Now().Add(-time.Hour * 2),
		},
		{
			ID:        uuid.New().String(),
			SessionID: sessionID1,
			Role:      "assistant",
			Content:   "这是一个测试回复",
			Timestamp: time.Now().Add(-time.Hour * 2).Add(time.Minute * 2),
		},
	}
}

// 
func chat(c *gin.Context) {
	var req AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "",
			"error":   err.Error(),
		})
		return
	}

	// 
	var sessionID string
	if req.SessionID == "" {
		sessionID = uuid.New().String()
		// 
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
		// 
		for i, session := range sessions {
			if session.ID == sessionID {
				sessions[i].UpdatedAt = time.Now()
				break
			}
		}
	}

	// 
	userMessage := ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	messages = append(messages, userMessage)

	// AI
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
		"message": "",
		"data":    response,
	})
}

// 
func getSessions(c *gin.Context) {
	userID := c.Query("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filteredSessions := make([]ChatSession, 0)
	for _, session := range sessions {
		if userID == "" || session.UserID == userID {
			// 
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
		"message": "",
		"data": gin.H{
			"sessions":    filteredSessions[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 
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
		"message": "",
		"data": gin.H{
			"messages":    sessionMessages[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 
func intentRecognition(c *gin.Context) {
	var req IntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "",
			"error":   err.Error(),
		})
		return
	}

	// 
	intent := analyzeIntent(req.Text)
	
	response := IntentResponse{
		Intent:     intent.Intent,
		Confidence: intent.Confidence,
		Entities:   intent.Entities,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    response,
	})
}

// 
func sentimentAnalysis(c *gin.Context) {
	var req SentimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "",
			"error":   err.Error(),
		})
		return
	}

	// 
	sentiment := analyzeSentiment(req.Text)
	
	response := SentimentResponse{
		Sentiment:  sentiment.Sentiment,
		Confidence: sentiment.Confidence,
		Score:      sentiment.Score,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    response,
	})
}

// 
func getProviders(c *gin.Context) {
	providers := []string{"OpenAI", "Claude", "Local"}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data": gin.H{
			"providers": providers,
			"default":   "OpenAI",
		},
	})
}

// 
func getModels(c *gin.Context) {
	models := map[string][]string{
		"OpenAI": {"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"},
		"Claude": {"claude-3-haiku", "claude-3-sonnet", "claude-3-opus"},
		"Local":  {"llama2", "chatglm"},
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    models,
	})
}

// 
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

// AI
func generateAIResponse(userMessage string) string {
	responses := []string{
		"",
		"J",
		"",
		"",
		"漰",
	}
	
	// 
	if contains(userMessage, "") || contains(userMessage, "") {
		return "屦" + responses[0] + ""
	}
	if contains(userMessage, "") || contains(userMessage, "") {
		return "" + responses[1] + ""	
	}
	if contains(userMessage, "") || contains(userMessage, "") {
		return "" + responses[2] + ""
	}
	
	return responses[0] + ""
}

// 
func analyzeIntent(text string) IntentResponse {
	if contains(text, "") || contains(text, "") {
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
	if contains(text, "") || contains(text, "") {
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

// 
func analyzeSentiment(text string) SentimentResponse {
	positiveWords := []string{"", "", "", ""}
	negativeWords := []string{"", "", "", ""}
	
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

// 
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

// 
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func main() {
	// 
	initMockData()

	// Gin
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API
	api := r.Group("/api/v1")
	{
		// AI
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

	// 
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ai-integration",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Println("AI (Mock汾)  8083")
	log.Fatal(r.Run(":8083"))
}

