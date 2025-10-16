package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"go.uber.org/zap"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// DefaultImageService 默认图像服务实现
type DefaultImageService struct {
	tempDir    string
	ffmpegPath string
	logger     *zap.Logger
}

// NewDefaultImageService 创建新的默认图像服务
func NewDefaultImageService(tempDir, ffmpegPath string, logger *zap.Logger) *DefaultImageService {
	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Error("Failed to create temp directory", zap.Error(err))
	}

	return &DefaultImageService{
		tempDir:    tempDir,
		ffmpegPath: ffmpegPath,
		logger:     logger,
	}
}

// GetDimensions 获取图像尺寸
func (is *DefaultImageService) GetDimensions(data []byte) (models.ImageDimensions, error) {
	// 检测图像格式
	format := is.detectImageFormat(data)
	
	switch format {
	case "jpeg", "jpg":
		return is.getJPEGDimensions(data)
	case "png":
		return is.getPNGDimensions(data)
	case "gif":
		return is.getGIFDimensions(data)
	case "bmp":
		return is.getBMPDimensions(data)
	case "webp":
		return is.getWebPDimensions(data)
	default:
		// 使用Go标准库解析
		return is.getDimensionsWithStandardLib(data)
	}
}

// ProcessImage 处理图像数据
func (is *DefaultImageService) ProcessImage(data []byte, targetFormat string) ([]byte, error) {
	// 检测输入格式
	inputFormat := is.detectImageFormat(data)
	
	// 如果格式相同且是支持的格式，直接返回
	if inputFormat == targetFormat && is.isSupportedFormat(targetFormat) {
		return data, nil
	}

	// 使用Go标准库进行格式转换
	return is.convertImageFormat(data, targetFormat)
}

// detectImageFormat 检测图像格式
func (is *DefaultImageService) detectImageFormat(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}

	// JPEG格式检测
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return "jpeg"
	}

	// PNG格式检测
	if len(data) >= 8 && bytes.Equal(data[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "png"
	}

	// GIF格式检测
	if len(data) >= 6 && (bytes.Equal(data[0:6], []byte("GIF87a")) || bytes.Equal(data[0:6], []byte("GIF89a"))) {
		return "gif"
	}

	// BMP格式检测
	if len(data) >= 2 && bytes.Equal(data[0:2], []byte("BM")) {
		return "bmp"
	}

	// WebP格式检测
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "webp"
	}

	// TIFF格式检测
	if len(data) >= 4 && (bytes.Equal(data[0:4], []byte("II*\x00")) || bytes.Equal(data[0:4], []byte("MM\x00*"))) {
		return "tiff"
	}

	return "unknown"
}

// getJPEGDimensions 获取JPEG图像尺寸
func (is *DefaultImageService) getJPEGDimensions(data []byte) (models.ImageDimensions, error) {
	reader := bytes.NewReader(data)
	
	// 跳过SOI标记
	if _, err := reader.Seek(2, io.SeekStart); err != nil {
		return models.ImageDimensions{}, err
	}

	for {
		// 读取标记
		var marker [2]byte
		if _, err := reader.Read(marker[:]); err != nil {
			return models.ImageDimensions{}, err
		}

		if marker[0] != 0xFF {
			return models.ImageDimensions{}, fmt.Errorf("invalid JPEG marker")
		}

		// 检查是否是SOF标记
		if (marker[1] >= 0xC0 && marker[1] <= 0xC3) || (marker[1] >= 0xC5 && marker[1] <= 0xC7) ||
		   (marker[1] >= 0xC9 && marker[1] <= 0xCB) || (marker[1] >= 0xCD && marker[1] <= 0xCF) {
			
			// 跳过长度字段
			if _, err := reader.Seek(3, io.SeekCurrent); err != nil {
				return models.ImageDimensions{}, err
			}

			// 读取高度和宽度
			var height, width uint16
			if err := binary.Read(reader, binary.BigEndian, &height); err != nil {
				return models.ImageDimensions{}, err
			}
			if err := binary.Read(reader, binary.BigEndian, &width); err != nil {
				return models.ImageDimensions{}, err
			}

			return models.ImageDimensions{Width: int(width), Height: int(height)}, nil
		}

		// 读取段长度并跳过
		var length uint16
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			return models.ImageDimensions{}, err
		}
		if _, err := reader.Seek(int64(length-2), io.SeekCurrent); err != nil {
			return models.ImageDimensions{}, err
		}
	}
}

// getPNGDimensions 获取PNG图像尺寸
func (is *DefaultImageService) getPNGDimensions(data []byte) (models.ImageDimensions, error) {
	if len(data) < 24 {
		return models.ImageDimensions{}, fmt.Errorf("invalid PNG file: too short")
	}

	// PNG IHDR chunk在偏移16处
	width := binary.BigEndian.Uint32(data[16:20])
	height := binary.BigEndian.Uint32(data[20:24])

	return models.ImageDimensions{Width: int(width), Height: int(height)}, nil
}

// getGIFDimensions 获取GIF图像尺寸
func (is *DefaultImageService) getGIFDimensions(data []byte) (models.ImageDimensions, error) {
	if len(data) < 10 {
		return models.ImageDimensions{}, fmt.Errorf("invalid GIF file: too short")
	}

	// GIF尺寸在偏移6处，小端序
	width := binary.LittleEndian.Uint16(data[6:8])
	height := binary.LittleEndian.Uint16(data[8:10])

	return models.ImageDimensions{Width: int(width), Height: int(height)}, nil
}

// getBMPDimensions 获取BMP图像尺寸
func (is *DefaultImageService) getBMPDimensions(data []byte) (models.ImageDimensions, error) {
	if len(data) < 26 {
		return models.ImageDimensions{}, fmt.Errorf("invalid BMP file: too short")
	}

	// BMP尺寸在偏移18处，小端序
	width := binary.LittleEndian.Uint32(data[18:22])
	height := binary.LittleEndian.Uint32(data[22:26])

	return models.ImageDimensions{Width: int(width), Height: int(height)}, nil
}

// getWebPDimensions 获取WebP图像尺寸
func (is *DefaultImageService) getWebPDimensions(data []byte) (models.ImageDimensions, error) {
	if len(data) < 30 {
		return models.ImageDimensions{}, fmt.Errorf("invalid WebP file: too short")
	}

	// 检查VP8格式
	if bytes.Equal(data[12:16], []byte("VP8 ")) {
		// VP8格式
		width := binary.LittleEndian.Uint16(data[26:28]) & 0x3FFF
		height := binary.LittleEndian.Uint16(data[28:30]) & 0x3FFF
		return models.ImageDimensions{Width: int(width), Height: int(height)}, nil
	} else if bytes.Equal(data[12:16], []byte("VP8L")) {
		// VP8L格式
		if len(data) < 25 {
			return models.ImageDimensions{}, fmt.Errorf("invalid WebP VP8L file: too short")
		}
		// VP8L的尺寸编码比较复杂，使用标准库解析
		return is.getDimensionsWithStandardLib(data)
	}

	return models.ImageDimensions{}, fmt.Errorf("unsupported WebP format")
}

// getDimensionsWithStandardLib 使用Go标准库获取图像尺寸
func (is *DefaultImageService) getDimensionsWithStandardLib(data []byte) (models.ImageDimensions, error) {
	reader := bytes.NewReader(data)
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return models.ImageDimensions{}, fmt.Errorf("failed to decode image config: %w", err)
	}

	return models.ImageDimensions{Width: config.Width, Height: config.Height}, nil
}

// convertImageFormat 转换图像格式
func (is *DefaultImageService) convertImageFormat(data []byte, targetFormat string) ([]byte, error) {
	// 解码图像
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 编码为目标格式
	var buf bytes.Buffer
	switch strings.ToLower(targetFormat) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	case "bmp":
		err = bmp.Encode(&buf, img)
	case "tiff":
		err = tiff.Encode(&buf, img, nil)
	case "webp":
		// WebP编码需要第三方库，这里使用FFmpeg
		return is.convertImageWithFFmpeg(data, targetFormat)
	default:
		return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	is.logger.Info("Image converted successfully",
		zap.String("target_format", targetFormat),
		zap.Int("input_size", len(data)),
		zap.Int("output_size", buf.Len()))

	return buf.Bytes(), nil
}

// convertImageWithFFmpeg 使用FFmpeg转换图像格式
func (is *DefaultImageService) convertImageWithFFmpeg(data []byte, targetFormat string) ([]byte, error) {
	if is.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// 检测输入格式
	inputFormat := is.detectImageFormat(data)
	if inputFormat == "unknown" {
		return nil, fmt.Errorf("unknown input format")
	}

	// 创建临时文件
	inputFile := filepath.Join(is.tempDir, fmt.Sprintf("input_%d.%s", time.Now().UnixNano(), inputFormat))
	outputFile := filepath.Join(is.tempDir, fmt.Sprintf("output_%d.%s", time.Now().UnixNano(), targetFormat))
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 执行转换
	cmd := exec.Command(is.ffmpegPath, "-i", inputFile, "-y", outputFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert image with FFmpeg: %w", err)
	}

	// 读取输出文件
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	return outputData, nil
}

// isSupportedFormat 检查是否为支持的格式
func (is *DefaultImageService) isSupportedFormat(format string) bool {
	supportedFormats := map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
		"gif":  true,
		"bmp":  true,
		"webp": true,
		"tiff": true,
	}
	return supportedFormats[format]
}

// ResizeImage 调整图像大小
func (is *DefaultImageService) ResizeImage(data []byte, width, height int) ([]byte, error) {
	if is.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available for image resizing")
	}

	// 检测输入格式
	inputFormat := is.detectImageFormat(data)
	if inputFormat == "unknown" {
		return nil, fmt.Errorf("unknown input format")
	}

	// 创建临时文件
	inputFile := filepath.Join(is.tempDir, fmt.Sprintf("input_%d.%s", time.Now().UnixNano(), inputFormat))
	outputFile := filepath.Join(is.tempDir, fmt.Sprintf("output_%d.%s", time.Now().UnixNano(), inputFormat))
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 执行调整大小
	cmd := exec.Command(is.ffmpegPath, "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", width, height), "-y", outputFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to resize image with FFmpeg: %w", err)
	}

	// 读取输出文件
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	is.logger.Info("Image resized successfully",
		zap.Int("width", width),
		zap.Int("height", height),
		zap.Int("input_size", len(data)),
		zap.Int("output_size", len(outputData)))

	return outputData, nil
}

// ExtractFeatures 提取图像特征
func (is *DefaultImageService) ExtractFeatures(data []byte) (*ImageFeatures, error) {
	// 获取基本信息
	dimensions, err := is.GetDimensions(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get dimensions: %w", err)
	}

	format := is.detectImageFormat(data)
	
	// 解码图像以提取更多特征
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 计算特征
	features := &ImageFeatures{
		Width:      dimensions.Width,
		Height:     dimensions.Height,
		Format:     format,
		Size:       len(data),
		AspectRatio: float64(dimensions.Width) / float64(dimensions.Height),
		Brightness: is.calculateBrightness(img),
		Contrast:   is.calculateContrast(img),
	}

	return features, nil
}

// calculateBrightness 计算图像亮度
func (is *DefaultImageService) calculateBrightness(img image.Image) float64 {
	bounds := img.Bounds()
	var totalBrightness float64
	var pixelCount int

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// 转换为0-255范围
			r, g, b = r>>8, g>>8, b>>8
			// 计算亮度（使用标准公式）
			brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			totalBrightness += brightness
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0
	}

	return totalBrightness / float64(pixelCount)
}

// calculateContrast 计算图像对比度
func (is *DefaultImageService) calculateContrast(img image.Image) float64 {
	bounds := img.Bounds()
	var brightnesses []float64

	// 收集所有像素的亮度值
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8
			brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			brightnesses = append(brightnesses, brightness)
		}
	}

	if len(brightnesses) == 0 {
		return 0
	}

	// 计算平均值
	var sum float64
	for _, brightness := range brightnesses {
		sum += brightness
	}
	mean := sum / float64(len(brightnesses))

	// 计算标准差（对比度的度量）
	var variance float64
	for _, brightness := range brightnesses {
		variance += math.Pow(brightness-mean, 2)
	}
	variance /= float64(len(brightnesses))

	return math.Sqrt(variance)
}

// ImageFeatures 图像特征
type ImageFeatures struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Format      string  `json:"format"`
	Size        int     `json:"size"`
	AspectRatio float64 `json:"aspect_ratio"`
	Brightness  float64 `json:"brightness"`
	Contrast    float64 `json:"contrast"`
}