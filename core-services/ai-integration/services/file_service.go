package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DefaultFileService 默认文件服务实现
type DefaultFileService struct {
	basePath   string
	maxSize    int64
	allowedExt map[string]bool
	logger     *zap.Logger
}

// NewDefaultFileService 创建新的默认文件服务
func NewDefaultFileService(basePath string, maxSize int64, logger *zap.Logger) *DefaultFileService {
	// 创建基础目录
	if err := os.MkdirAll(basePath, 0755); err != nil {
		logger.Error("Failed to create base directory", zap.Error(err))
	}

	allowedExt := map[string]bool{
		".txt":  true, ".md": true, ".json": true, ".xml": true, ".csv": true,
		".jpg":  true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true,
		".mp3":  true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true,
		".mp4":  true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true,
		".pdf":  true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	}

	return &DefaultFileService{
		basePath:   basePath,
		maxSize:    maxSize,
		allowedExt: allowedExt,
		logger:     logger,
	}
}

// SaveFile 保存文件到本地存储
func (fs *DefaultFileService) SaveFile(ctx context.Context, data []byte, filename string) (string, error) {
	// 验证文件大小
	if int64(len(data)) > fs.maxSize {
		return "", fmt.Errorf("file size %d exceeds maximum allowed size %d", len(data), fs.maxSize)
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	if !fs.allowedExt[ext] {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	// 生成唯一文件名
	hash := md5.Sum(data)
	hashStr := fmt.Sprintf("%x", hash)
	timestamp := time.Now().Format("20060102150405")
	uniqueFilename := fmt.Sprintf("%s_%s_%s", timestamp, hashStr[:8], filename)

	// 创建子目录（按日期分组）
	dateDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(fs.basePath, dateDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		fs.logger.Error("Failed to create directory", zap.Error(err))
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// 保存文件
	fullPath := filepath.Join(fullDir, uniqueFilename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		fs.logger.Error("Failed to save file", zap.Error(err))
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// 返回相对路径作为URL
	relativeURL := filepath.Join(dateDir, uniqueFilename)
	relativeURL = strings.ReplaceAll(relativeURL, "\\", "/") // 确保使用正斜杠

	fs.logger.Info("File saved successfully",
		zap.String("filename", filename),
		zap.String("url", relativeURL),
		zap.Int("size", len(data)))

	return relativeURL, nil
}

// GetFile 从本地存储获取文件
func (fs *DefaultFileService) GetFile(ctx context.Context, url string) ([]byte, error) {
	// 清理URL路径
	cleanURL := strings.TrimPrefix(url, "/")
	cleanURL = strings.ReplaceAll(cleanURL, "/", string(filepath.Separator))

	// 构建完整路径
	fullPath := filepath.Join(fs.basePath, cleanURL)

	// 安全检查：确保路径在基础目录内
	absBasePath, err := filepath.Abs(fs.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute base path: %w", err)
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute file path: %w", err)
	}

	if !strings.HasPrefix(absFullPath, absBasePath) {
		return nil, fmt.Errorf("invalid file path: path traversal detected")
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", url)
	}

	// 读取文件
	data, err := os.ReadFile(fullPath)
	if err != nil {
		fs.logger.Error("Failed to read file", zap.Error(err), zap.String("path", fullPath))
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fs.logger.Info("File retrieved successfully",
		zap.String("url", url),
		zap.Int("size", len(data)))

	return data, nil
}

// GetFileFromURL 从远程URL获取文件
func (fs *DefaultFileService) GetFileFromURL(ctx context.Context, url string) ([]byte, error) {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置用户代理
	req.Header.Set("User-Agent", "TaishangLaojun-AI-Integration/1.0")

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file from URL: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// 检查内容长度
	if resp.ContentLength > fs.maxSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", resp.ContentLength, fs.maxSize)
	}

	// 读取响应体
	data, err := io.ReadAll(io.LimitReader(resp.Body, fs.maxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fs.logger.Info("File downloaded from URL",
		zap.String("url", url),
		zap.Int("size", len(data)))

	return data, nil
}

// DeleteFile 删除文件
func (fs *DefaultFileService) DeleteFile(ctx context.Context, url string) error {
	// 清理URL路径
	cleanURL := strings.TrimPrefix(url, "/")
	cleanURL = strings.ReplaceAll(cleanURL, "/", string(filepath.Separator))

	// 构建完整路径
	fullPath := filepath.Join(fs.basePath, cleanURL)

	// 安全检查
	absBasePath, err := filepath.Abs(fs.basePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute base path: %w", err)
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}

	if !strings.HasPrefix(absFullPath, absBasePath) {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", url)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fs.logger.Info("File deleted successfully", zap.String("url", url))
	return nil
}

// GetFileInfo 获取文件信息
func (fs *DefaultFileService) GetFileInfo(ctx context.Context, url string) (*FileInfo, error) {
	// 清理URL路径
	cleanURL := strings.TrimPrefix(url, "/")
	cleanURL = strings.ReplaceAll(cleanURL, "/", string(filepath.Separator))

	// 构建完整路径
	fullPath := filepath.Join(fs.basePath, cleanURL)

	// 获取文件信息
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", url)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		URL:     url,
	}, nil
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
	URL     string    `json:"url"`
}