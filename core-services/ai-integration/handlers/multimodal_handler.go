package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// MultimodalHandler 多模态处理
type MultimodalHandler struct {
	multimodalService *services.MultimodalService
	upgrader          websocket.Upgrader
}

// NewMultimodalHandler 多模态处理
func NewMultimodalHandler(multimodalService *services.MultimodalService) *MultimodalHandler {
	return &MultimodalHandler{
		multimodalService: multimodalService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
	}
}

// ProcessMultimodal 处理多模态请求
// @Summary 处理多模态请求
// @Description 处理多模态请求，包括文本、图像、语音等
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body models.MultimodalRequest true ""
// @Success 200 {object} models.MultimodalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/process [post]
func (h *MultimodalHandler) ProcessMultimodal(c *gin.Context) {
	var req models.MultimodalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// ID
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()
	req.Status = "processing"

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}
	req.UserID = userID.(string)

	// 处理多模态请求
	ctx := context.Background()
	response, err := h.multimodalService.ProcessMultimodalRequest(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Processing failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到多模态处理系统
// @Tags multimodal
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true ""
// @Param description formData string false ""
// @Success 200 {object} models.MultimodalInput
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/upload [post]
func (h *MultimodalHandler) UploadFile(c *gin.Context) {
	//
	input, err := h.multimodalService.ProcessFileUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "File upload failed",
			Message: err.Error(),
		})
		return
	}

	//
	if description := c.PostForm("description"); description != "" {
		switch content := input.Content.(type) {
		case models.ImageInput:
			content.Description = description
			input.Content = content
		}
	}

	c.JSON(http.StatusOK, input)
}

// StreamMultimodal WebSocket
// @Summary WebSocket
// @Description WebSocketAI
// @Tags multimodal
// @Param session_id query string false "ID"
// @Router /api/v1/multimodal/stream [get]
func (h *MultimodalHandler) StreamMultimodal(c *gin.Context) {
	// WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "WebSocket upgrade failed",
			Message: err.Error(),
		})
		return
	}
	defer conn.Close()

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		conn.WriteJSON(map[string]string{"error": "Unauthorized"})
		return
	}

	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	ctx := context.Background()

	for {
		// 读取客户端消息
		var req models.MultimodalRequest
		if err := conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// 生成请求ID
		req.ID = uuid.New().String()
		req.UserID = userID.(string)
		req.SessionID = sessionID
		req.CreatedAt = time.Now()
		req.Status = "processing"

		// 配置流式响应
		if req.Config.Provider == "" {
			req.Config = models.MultimodalConfig{
				Provider:    "openai",
				Model:       "gpt-4-vision",
				Temperature: 0.7,
				MaxTokens:   1000,
				Stream:      true,
			}
		} else {
			req.Config.Stream = true
		}

		// 处理流式响应
		responseChan := make(chan *models.MultimodalResponse, 10)

		// 启动流式处理 goroutine
		go func() {
			if err := h.multimodalService.StreamMultimodalResponse(ctx, &req, responseChan); err != nil {
				conn.WriteJSON(map[string]string{"error": err.Error()})
			}
		}()

		// 发送流式响应
		for response := range responseChan {
			if err := conn.WriteJSON(response); err != nil {
				fmt.Printf("WebSocket write error: %v\n", err)
				break
			}
		}

		// 发送完成消息
		conn.WriteJSON(map[string]string{"status": "completed"})
	}
}

// GetSessions
// @Summary 获取多模态会话列表
// @Description 获取用户的多模态会话列表
// @Tags multimodal
// @Produce json
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Param type query string false ""
// @Success 200 {object} PaginatedResponse{data=[]models.MultimodalSession}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions [get]
func (h *MultimodalHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// TODO:
	// userID
	_ = userID
	sessions := []models.MultimodalSession{}

	response := PaginatedResponse{
		Data:       sessions,
		Page:       page,
		Limit:      limit,
		Total:      int64(len(sessions)),
		TotalPages: (int64(len(sessions)) + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// CreateSession 创建多模态会话
// @Summary 创建多模态会话
// @Description 创建一个新的多模态会话
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true ""
// @Success 201 {object} models.MultimodalSession
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions [post]
func (h *MultimodalHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// 创建会话
	session := &models.MultimodalSession{
		ID:           uuid.New().String(),
		UserID:       userID.(string),
		Title:        req.Title,
		Type:         req.Type,
		Config:       req.Config,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		Status:       "active",
		Metadata:     req.Metadata,
	}

	// TODO: 保存会话到数据库

	c.JSON(http.StatusCreated, session)
}

// GetSession
// @Summary 获取多模态会话
// @Description 根据ID获取多模态会话详情
// @Tags multimodal
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.MultimodalSession
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions/{id} [get]
// @Description 根据ID获取多模态会话详情
// @Tags multimodal
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.MultimodalSession
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions/{id} [get]
func (h *MultimodalHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// TODO:
	session := &models.MultimodalSession{
		ID:     sessionID,
		UserID: userID.(string),
		Title:  "Sample Session",
		Status: "active",
	}

	c.JSON(http.StatusOK, session)
}

// UpdateSession
// @Summary 更新多模态会话
// @Description 更新多模态会话详情
// @Tags multimodal
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body UpdateSessionRequest true ""
// @Success 200 {object} models.MultimodalSession
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions/{id} [put]
func (h *MultimodalHandler) UpdateSession(c *gin.Context) {
	sessionID := c.Param("id")

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// TODO:
	session := &models.MultimodalSession{
		ID:        sessionID,
		UserID:    userID.(string),
		Title:     req.Title,
		Config:    req.Config,
		UpdatedAt: time.Now(),
		Metadata:  req.Metadata,
	}

	c.JSON(http.StatusOK, session)
}

// DeleteSession
// @Summary 删除多模态会话
// @Description 删除指定ID的多模态会话
// @Tags multimodal
// @Param id path string true "ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions/{id} [delete]
func (h *MultimodalHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// TODO: 删除会话
	_ = sessionID
	_ = userID

	c.Status(http.StatusNoContent)
}

// GetSessionMessages
// @Summary 获取多模态会话消息
// @Description 根据ID获取多模态会话的消息列表
// @Tags multimodal
// @Produce json
// @Param id path string true "ID"
// @Param page query int false "" default(1)
// @Param limit query int false "" default(50)
// @Success 200 {object} PaginatedResponse{data=[]models.MultimodalMessage}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/sessions/{id}/messages [get]
func (h *MultimodalHandler) GetSessionMessages(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	//
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// TODO:
	messages := []models.MultimodalMessage{}

	response := PaginatedResponse{
		Data:       messages,
		Page:       page,
		Limit:      limit,
		Total:      int64(len(messages)),
		TotalPages: (int64(len(messages)) + int64(limit) - 1) / int64(limit),
	}

	c.JSON(http.StatusOK, response)
}

// GenerateImage
// @Summary 生成多模态图像
// @Description 根据提示生成多模态图像
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body ImageGenerateRequest true ""
// @Success 200 {object} ImageGenerateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/image/generate [post]
func (h *MultimodalHandler) GenerateImage(c *gin.Context) {
	var req ImageGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// providers
	providerReq := &providers.ImageGenerateRequest{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		Count:          req.N,
		UserID:         req.UserID,
		Metadata:       req.Metadata,
	}

	response, err := h.multimodalService.GenerateImage(c.Request.Context(), providerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "generation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AnalyzeImage
// @Summary 分析多模态图像
// @Description 根据图像URL或Base64数据分析图像内容
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body ImageAnalyzeRequest true ""
// @Success 200 {object} ImageAnalyzeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/image/analyze [post]
func (h *MultimodalHandler) AnalyzeImage(c *gin.Context) {
	var req ImageAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// providers
	providerReq := &providers.ImageAnalyzeRequest{
		ImageURL:    req.ImageURL,
		ImageBase64: "", // handlerImageBase64
		ImageData:   req.ImageData,
		Prompt:      req.Prompt,
		Features:    req.Features,
		Language:    "", // handlerLanguage
		Detail:      req.Detail,
		UserID:      req.UserID,
		Metadata:    req.Metadata,
	}

	response, err := h.multimodalService.AnalyzeImage(c.Request.Context(), providerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// EditImage
// @Summary 编辑多模态图像
// @Description 根据提示编辑多模态图像
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body ImageEditRequest true ""
// @Success 200 {object} ImageEditResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/image/edit [post]
func (h *MultimodalHandler) EditImage(c *gin.Context) {
	var req ImageEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// providers
	providerReq := &providers.ImageEditRequest{
		ImageURL:    req.ImageURL,
		ImageBase64: "", // handlerImageBase64
		ImageData:   req.ImageData,
		MaskURL:     req.MaskURL,
		MaskBase64:  "", // handlerMaskBase64
		MaskData:    req.MaskData,
		Prompt:      req.Prompt,
		Size:        req.Size,
		Count:       req.N, // handlerNprovidersCount
		UserID:      req.UserID,
		Metadata:    req.Metadata,
	}

	response, err := h.multimodalService.EditImage(c.Request.Context(), providerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "edit_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UploadImageForAnalysis
// @Summary 上传多模态图像分析
// @Description 上传图像文件并根据提示分析图像内容
// @Tags multimodal
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true ""
// @Param prompt formData string false ""
// @Param features formData string false "()"
// @Success 200 {object} ImageAnalyzeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/image/upload-analyze [post]
func (h *MultimodalHandler) UploadImageForAnalysis(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_file",
			Message: "No file uploaded",
		})
		return
	}
	defer file.Close()

	//
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "file_read_error",
			Message: err.Error(),
		})
		return
	}

	//
	req := ImageAnalyzeRequest{
		ImageData: fileData,
		Prompt:    c.PostForm("prompt"),
		Detail:    "high",
	}

	//
	if featuresStr := c.PostForm("features"); featuresStr != "" {
		req.Features = strings.Split(featuresStr, ",")
	}

	// ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	//
	req.Metadata = map[string]string{
		"filename":     header.Filename,
		"content_type": header.Header.Get("Content-Type"),
		"size":         fmt.Sprintf("%d", header.Size),
	}

	// providers
	providerReq := &providers.ImageAnalyzeRequest{
		ImageData: req.ImageData,
		Prompt:    req.Prompt,
		Features:  req.Features,
		Detail:    req.Detail,
		UserID:    req.UserID,
		Metadata:  req.Metadata,
	}

	response, err := h.multimodalService.AnalyzeImage(c.Request.Context(), providerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

type CreateSessionRequest struct {
	Title    string                  `json:"title" binding:"required"`
	Type     models.MultimodalType   `json:"type" binding:"required"`
	Config   models.MultimodalConfig `json:"config"`
	Metadata map[string]interface{}  `json:"metadata"`
}

type UpdateSessionRequest struct {
	Title    string                  `json:"title"`
	Config   models.MultimodalConfig `json:"config"`
	Metadata map[string]interface{}  `json:"metadata"`
}

type ImageGenerateRequest struct {
	UserID         string            `json:"user_id,omitempty"`
	Prompt         string            `json:"prompt" binding:"required"`
	NegativePrompt string            `json:"negative_prompt,omitempty"`
	Size           string            `json:"size,omitempty"`
	Quality        string            `json:"quality,omitempty"`
	Style          string            `json:"style,omitempty"`
	N              int               `json:"n,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type ImageGenerateResponse struct {
	ID        string            `json:"id"`
	Images    []GeneratedImage  `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type GeneratedImage struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type ImageAnalyzeRequest struct {
	UserID    string            `json:"user_id,omitempty"`
	ImageURL  string            `json:"image_url,omitempty"`
	ImageData []byte            `json:"image_data,omitempty"`
	Prompt    string            `json:"prompt,omitempty"`
	Detail    string            `json:"detail,omitempty"`
	Features  []string          `json:"features,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ImageAnalyzeResponse struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Objects     []DetectedObject  `json:"objects,omitempty"`
	Text        string            `json:"text,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Colors      []string          `json:"colors,omitempty"`
	Confidence  float64           `json:"confidence"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type DetectedObject struct {
	Name        string       `json:"name"`
	Confidence  float64      `json:"confidence"`
	BoundingBox *BoundingBox `json:"bounding_box,omitempty"`
}

type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ImageEditRequest struct {
	UserID    string            `json:"user_id,omitempty"`
	ImageURL  string            `json:"image_url,omitempty"`
	ImageData []byte            `json:"image_data,omitempty"`
	MaskURL   string            `json:"mask_url,omitempty"`
	MaskData  []byte            `json:"mask_data,omitempty"`
	Prompt    string            `json:"prompt" binding:"required"`
	Size      string            `json:"size,omitempty"`
	N         int               `json:"n,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ImageEditResponse struct {
	ID        string            `json:"id"`
	Images    []GeneratedImage  `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}
