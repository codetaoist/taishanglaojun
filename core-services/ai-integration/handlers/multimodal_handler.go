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

// MultimodalHandler 多模态AI处理�?
type MultimodalHandler struct {
	multimodalService *services.MultimodalService
	upgrader          websocket.Upgrader
}

// NewMultimodalHandler 创建多模态AI处理�?
func NewMultimodalHandler(multimodalService *services.MultimodalService) *MultimodalHandler {
	return &MultimodalHandler{
		multimodalService: multimodalService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该进行适当的来源检�?
			},
		},
	}
}

// ProcessMultimodal 处理多模态请�?
// @Summary 处理多模态AI请求
// @Description 支持文本、语音、图像等多种输入类型的AI处理
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body models.MultimodalRequest true "多模态请�?
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

	// 设置请求ID和时间戳
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()
	req.Status = "processing"

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}
	req.UserID = userID.(string)

	// 处理请求
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
// @Summary 上传多媒体文�?
// @Description 支持上传图像、音频、视频等文件用于多模态处�?
// @Tags multimodal
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Param description formData string false "文件描述"
// @Success 200 {object} models.MultimodalInput
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/multimodal/upload [post]
func (h *MultimodalHandler) UploadFile(c *gin.Context) {
	// 处理文件上传
	input, err := h.multimodalService.ProcessFileUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "File upload failed",
			Message: err.Error(),
		})
		return
	}

	// 添加描述信息
	if description := c.PostForm("description"); description != "" {
		switch content := input.Content.(type) {
		case models.ImageInput:
			content.Description = description
			input.Content = content
		}
	}

	c.JSON(http.StatusOK, input)
}

// StreamMultimodal WebSocket流式处理
// @Summary WebSocket流式多模态处�?
// @Description 通过WebSocket进行实时多模态AI交互
// @Tags multimodal
// @Param session_id query string false "会话ID"
// @Router /api/v1/multimodal/stream [get]
func (h *MultimodalHandler) StreamMultimodal(c *gin.Context) {
	// 升级到WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "WebSocket upgrade failed",
			Message: err.Error(),
		})
		return
	}
	defer conn.Close()

	// 获取用户ID
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
		// 读取客户端消�?
		var req models.MultimodalRequest
		if err := conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// 设置请求信息
		req.ID = uuid.New().String()
		req.UserID = userID.(string)
		req.SessionID = sessionID
		req.CreatedAt = time.Now()
		req.Status = "processing"

		// 启用流式处理
		req.Config.Stream = true

		// 创建输出通道
		outputChan := make(chan *models.MultimodalOutput, 10)

		// 启动流式处理
		go func() {
			if err := h.multimodalService.StreamMultimodalResponse(ctx, &req, outputChan); err != nil {
				conn.WriteJSON(map[string]string{"error": err.Error()})
			}
		}()

		// 发送流式响�?
		for output := range outputChan {
			if err := conn.WriteJSON(output); err != nil {
				fmt.Printf("WebSocket write error: %v\n", err)
				break
			}
		}

		// 发送完成信�?
		conn.WriteJSON(map[string]string{"status": "completed"})
	}
}

// GetSessions 获取用户的多模态会话列�?
// @Summary 获取多模态会话列�?
// @Description 获取当前用户的所有多模态会�?
// @Tags multimodal
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param type query string false "会话类型"
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

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// TODO: 实现会话查询逻辑
	// 暂时忽略userID变量，等待实际实�?
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

// CreateSession 创建新的多模态会�?
// @Summary 创建多模态会�?
// @Description 创建新的多模态AI会话
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true "会话创建请求"
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

	// TODO: 保存到数据库

	c.JSON(http.StatusCreated, session)
}

// GetSession 获取特定会话详情
// @Summary 获取会话详情
// @Description 获取指定ID的多模态会话详�?
// @Tags multimodal
// @Produce json
// @Param id path string true "会话ID"
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

	// TODO: 从数据库获取会话
	session := &models.MultimodalSession{
		ID:     sessionID,
		UserID: userID.(string),
		Title:  "Sample Session",
		Status: "active",
	}

	c.JSON(http.StatusOK, session)
}

// UpdateSession 更新会话信息
// @Summary 更新会话信息
// @Description 更新多模态会话的标题、配置等信息
// @Tags multimodal
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Param request body UpdateSessionRequest true "会话更新请求"
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

	// TODO: 更新数据库中的会�?
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

// DeleteSession 删除会话
// @Summary 删除会话
// @Description 删除指定的多模态会�?
// @Tags multimodal
// @Param id path string true "会话ID"
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

	// TODO: 从数据库删除会话
	_ = sessionID
	_ = userID

	c.Status(http.StatusNoContent)
}

// GetSessionMessages 获取会话消息
// @Summary 获取会话消息
// @Description 获取指定会话的所有消�?
// @Tags multimodal
// @Produce json
// @Param id path string true "会话ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(50)
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

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// TODO: 从数据库获取消息
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

// GenerateImage 图像生成端点
// @Summary 生成图像
// @Description 根据文本提示生成图像
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body ImageGenerateRequest true "图像生成请求"
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

	// 设置用户ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// 转换为providers请求结构�?
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

// AnalyzeImage 图像分析端点
// @Summary 分析图像
// @Description 分析图像内容并返回详细信�?
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body ImageAnalyzeRequest true "图像分析请求"
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

	// 设置用户ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// 转换为providers请求结构�?
	providerReq := &providers.ImageAnalyzeRequest{
		ImageURL:    req.ImageURL,
		ImageBase64: "", // handler中没有ImageBase64字段
		ImageData:   req.ImageData,
		Prompt:      req.Prompt,
		Features:    req.Features,
		Language:    "", // handler中没有Language字段
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

// EditImage 图像编辑端点
// @Summary 编辑图像
// @Description 根据提示编辑图像
// @Tags multimodal
// @Accept json
// @Produce json
// @Param request body ImageEditRequest true "图像编辑请求"
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

	// 设置用户ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// 转换为providers请求结构�?
	providerReq := &providers.ImageEditRequest{
		ImageURL:    req.ImageURL,
		ImageBase64: "", // handler中没有ImageBase64字段
		ImageData:   req.ImageData,
		MaskURL:     req.MaskURL,
		MaskBase64:  "", // handler中没有MaskBase64字段
		MaskData:    req.MaskData,
		Prompt:      req.Prompt,
		Size:        req.Size,
		Count:       req.N, // handler中使用N字段，providers中使用Count字段
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

// UploadImageForAnalysis 上传图像进行分析
// @Summary 上传图像进行分析
// @Description 上传图像文件并进行分�?
// @Tags multimodal
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图像文件"
// @Param prompt formData string false "分析提示"
// @Param features formData string false "分析特征(逗号分隔)"
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

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "file_read_error",
			Message: err.Error(),
		})
		return
	}

	// 构建分析请求
	req := ImageAnalyzeRequest{
		ImageData: fileData,
		Prompt:    c.PostForm("prompt"),
		Detail:    "high",
	}

	// 解析特征
	if featuresStr := c.PostForm("features"); featuresStr != "" {
		req.Features = strings.Split(featuresStr, ",")
	}

	// 设置用户ID
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// 设置元数�?
	req.Metadata = map[string]string{
		"filename":    header.Filename,
		"content_type": header.Header.Get("Content-Type"),
		"size":        fmt.Sprintf("%d", header.Size),
	}

	// 转换为providers请求结构�?
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

// 请求和响应结构体
type CreateSessionRequest struct {
	Title    string                 `json:"title" binding:"required"`
	Type     models.MultimodalType  `json:"type" binding:"required"`
	Config   models.MultimodalConfig `json:"config"`
	Metadata map[string]interface{} `json:"metadata"`
}

type UpdateSessionRequest struct {
	Title    string                 `json:"title"`
	Config   models.MultimodalConfig `json:"config"`
	Metadata map[string]interface{} `json:"metadata"`
}



// 图像处理相关请求和响应结构体
type ImageGenerateRequest struct {
	UserID     string            `json:"user_id,omitempty"`
	Prompt     string            `json:"prompt" binding:"required"`
	NegativePrompt string        `json:"negative_prompt,omitempty"`
	Size       string            `json:"size,omitempty"`
	Quality    string            `json:"quality,omitempty"`
	Style      string            `json:"style,omitempty"`
	N          int               `json:"n,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type ImageGenerateResponse struct {
	ID        string            `json:"id"`
	Images    []GeneratedImage  `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type GeneratedImage struct {
	URL       string `json:"url,omitempty"`
	B64JSON   string `json:"b64_json,omitempty"`
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
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	BoundingBox *BoundingBox `json:"bounding_box,omitempty"`
}

type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ImageEditRequest struct {
	UserID      string            `json:"user_id,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	ImageData   []byte            `json:"image_data,omitempty"`
	MaskURL     string            `json:"mask_url,omitempty"`
	MaskData    []byte            `json:"mask_data,omitempty"`
	Prompt      string            `json:"prompt" binding:"required"`
	Size        string            `json:"size,omitempty"`
	N           int               `json:"n,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ImageEditResponse struct {
	ID        string            `json:"id"`
	Images    []GeneratedImage  `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}
