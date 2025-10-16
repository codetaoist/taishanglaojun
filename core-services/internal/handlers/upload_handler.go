package handlers

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// UploadHandler 处理文件上传
type UploadHandler struct {
    logger *zap.Logger
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(logger *zap.Logger) *UploadHandler {
    return &UploadHandler{logger: logger}
}

// HandleUpload 处理上传请求，接收 form 字段名为 "file" 的文件
func (h *UploadHandler) HandleUpload(c *gin.Context) {
    // 读取上传的文件
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        h.logger.Warn("upload: no file in request", zapError(err))
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "未检测到上传文件", "error": err.Error()})
        return
    }
    defer file.Close()

    // 基本的类型检查（允许常见图片和图标类型）
    contentType := header.Header.Get("Content-Type")
    allowedTypes := map[string]bool{
        "image/png":            true,
        "image/jpeg":           true,
        "image/jpg":            true,
        "image/gif":            true,
        "image/webp":           true,
        "image/svg+xml":        true,
        "image/x-icon":         true,
        "image/vnd.microsoft.icon": true,
    }
    if contentType != "" && !allowedTypes[strings.ToLower(contentType)] {
        h.logger.Warn("upload: disallowed content type", zap.String("contentType", contentType))
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "不支持的文件类型", "contentType": contentType})
        return
    }

    // 组织保存路径 uploads/YYYY/MM/DD/{timestamp}_{filename}
    datePath := time.Now().Format("2006/01/02")
    // 真实磁盘路径使用 filepath.Join（Windows 兼容）
    saveDir := filepath.Join("uploads", datePath)
    if err := os.MkdirAll(saveDir, 0o755); err != nil {
        h.logger.Error("upload: mkdir failed", zapError(err))
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建上传目录失败"})
        return
    }

    // 清理文件名，仅保留基名
    baseName := filepath.Base(header.Filename)
    if baseName == "" {
        baseName = fmt.Sprintf("file_%d", time.Now().UnixNano())
    }
    fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), baseName)
    savePath := filepath.Join(saveDir, fileName)

    out, err := os.Create(savePath)
    if err != nil {
        h.logger.Error("upload: create file failed", zapError(err), zap.String("path", savePath))
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存文件失败"})
        return
    }
    defer out.Close()

    if _, err := io.Copy(out, file); err != nil {
        h.logger.Error("upload: write file failed", zapError(err))
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "写入文件失败"})
        return
    }

    // URL 路径使用正斜杠，确保前端可直接访问
    urlPath := path.Join("/uploads", datePath, fileName)
    h.logger.Info("upload: success", zap.String("url", urlPath), zap.String("contentType", contentType))

    c.JSON(http.StatusOK, gin.H{
        "success":   true,
        "url":       urlPath,
        "filename":  header.Filename,
        "size":      header.Size,
        "mime":      contentType,
        "timestamp": time.Now().Unix(),
    })
}

// zapError 是小工具，避免引入 zap.NamedError 依赖差异
func zapError(err error) zap.Field { return zap.Error(err) }