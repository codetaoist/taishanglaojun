package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	logger    *zap.Logger
	uploadDir string
	baseURL   string
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(logger *zap.Logger, uploadDir, baseURL string) *UploadHandler {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory", zap.Error(err))
	}

	return &UploadHandler{
		logger:    logger,
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadResponse 上传响应
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
	} `json:"data"`
}

// UploadErrorResponse 上传错误响应
type UploadErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// UploadAvatar 上传头像
// @Summary 上传用户头像
// @Description 上传用户头像图片
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像文件"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /upload/avatar [post]
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
c.JSON(http.StatusUnauthorized, UploadErrorResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "用户未认证",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, UploadErrorResponse{
			Success: false,
			Error:   "internal_error",
			Message: "内部服务器错误",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		h.logger.Error("Failed to get uploaded file", zap.Error(err))
		c.JSON(http.StatusBadRequest, UploadErrorResponse{
			Success: false,
			Error:   "invalid_file",
			Message: "无效的文件",
		})
		return
	}
	defer file.Close()

	// 验证文件类型
	if !h.isValidImageType(header) {
		c.JSON(http.StatusBadRequest, UploadErrorResponse{
			Success: false,
			Error:   "invalid_file_type",
			Message: "不支持的文件类型",
		})
		return
	}

	// 验证文件大小 (2MB限制)
	if header.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, UploadErrorResponse{
			Success: false,
			Error:   "file_too_large",
			Message: "文件大小超过限制",
		})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("avatar_%s_%d%s", uid.String(), time.Now().Unix(), ext)

	// 创建用户专属目录
	userDir := filepath.Join(h.uploadDir, "avatars", uid.String())
	if err := os.MkdirAll(userDir, 0755); err != nil {
		h.logger.Error("Failed to create user directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, UploadErrorResponse{
			Success: false,
			Error:   "internal_error",
			Message: "内部服务器错误",
		})
		return
	}

	// 保存文件
	filePath := filepath.Join(userDir, filename)
	if err := h.saveFile(file, filePath); err != nil {
		h.logger.Error("Failed to save file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, UploadErrorResponse{
			Success: false,
			Error:   "save_failed",
			Message: "保存文件失败",
		})
		return
	}

	// 生成访问URL
	fileURL := fmt.Sprintf("%s/uploads/avatars/%s/%s", h.baseURL, uid.String(), filename)

	h.logger.Info("Avatar uploaded successfully",
		zap.String("user_id", uid.String()),
		zap.String("filename", filename),
		zap.Int64("size", header.Size),
	)

	// 返回成功响应
	response := UploadResponse{
		Success: true,
		Message: "Avatar uploaded successfully",
	}
	response.Data.URL = fileURL
	response.Data.Filename = filename
	response.Data.Size = header.Size

	c.JSON(http.StatusOK, response)
}

// UploadFile 通用文件上传
// @Summary 上传文件
// @Description 上传通用文件
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param type query string false "文件类型" Enums(document,image,other)
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /upload/file [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, UploadErrorResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "用户未认证",
		})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, UploadErrorResponse{
			Success: false,
			Error:   "internal_error",
			Message: "内部服务器错误",
		})
		return
	}

	// 获取文件类型参数
	fileType := c.DefaultQuery("type", "other")

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get uploaded file", zap.Error(err))
		c.JSON(http.StatusBadRequest, UploadErrorResponse{
			Success: false,
			Error:   "invalid_file",
			Message: "无效的文件",
		})
		return
	}
	defer file.Close()

	// 根据类型验证文件
	if !h.isValidFileType(header, fileType) {
		c.JSON(http.StatusBadRequest, UploadErrorResponse{
			Success: false,
			Error:   "invalid_file_type",
			Message: "不支持的文件类型",
		})
		return
	}

	// 验证文件大小 (10MB限制)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, UploadErrorResponse{
			Success: false,
			Error:   "file_too_large",
			Message: "文件大小超过限制",
		})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s_%s_%d%s", fileType, uid.String(), time.Now().Unix(), ext)

	// 创建类型目录
	typeDir := filepath.Join(h.uploadDir, fileType, uid.String())
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		h.logger.Error("Failed to create type directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, UploadErrorResponse{
			Success: false,
			Error:   "internal_error",
			Message: "内部服务器错误",
		})
		return
	}

	// 保存文件
	filePath := filepath.Join(typeDir, filename)
	if err := h.saveFile(file, filePath); err != nil {
		h.logger.Error("Failed to save file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, UploadErrorResponse{
			Success: false,
			Error:   "save_failed",
			Message: "保存文件失败",
		})
		return
	}

	// 生成访问URL
	fileURL := fmt.Sprintf("%s/uploads/%s/%s/%s", h.baseURL, fileType, uid.String(), filename)

	h.logger.Info("File uploaded successfully",
		zap.String("user_id", uid.String()),
		zap.String("type", fileType),
		zap.String("filename", filename),
		zap.Int64("size", header.Size),
	)

	// 返回成功响应
	response := UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
	}
	response.Data.URL = fileURL
	response.Data.Filename = filename
	response.Data.Size = header.Size

	c.JSON(http.StatusOK, response)
}

// isValidImageType 验证是否为有效的图片类型
func (h *UploadHandler) isValidImageType(header *multipart.FileHeader) bool {
	contentType := header.Header.Get("Content-Type")
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}

	// 也检查文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// isValidFileType 验证文件类型
func (h *UploadHandler) isValidFileType(header *multipart.FileHeader, fileType string) bool {
	contentType := header.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(header.Filename))

	switch fileType {
	case "image":
		return h.isValidImageType(header)
	case "document":
		validTypes := []string{
			"application/pdf",
			"text/plain",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		}
		validExts := []string{".pdf", ".txt", ".doc", ".docx"}
		
		for _, validType := range validTypes {
			if contentType == validType {
				return true
			}
		}
		for _, validExt := range validExts {
			if ext == validExt {
				return true
			}
		}
		return false
	default:
		// 对于其他类型，允许大部分常见文件类型
		return true
	}
}

// saveFile 保存文件到指定路径
func (h *UploadHandler) saveFile(src multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}