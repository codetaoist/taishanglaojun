package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DefaultVideoService 默认视频服务实现
type DefaultVideoService struct {
	tempDir    string
	ffmpegPath string
	ffprobePath string
	logger     *zap.Logger
}

// NewDefaultVideoService 创建新的默认视频服务
func NewDefaultVideoService(tempDir, ffmpegPath, ffprobePath string, logger *zap.Logger) *DefaultVideoService {
	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Error("Failed to create temp directory", zap.Error(err))
	}

	return &DefaultVideoService{
		tempDir:     tempDir,
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
		logger:      logger,
	}
}

// GetDuration 获取视频时长
func (vs *DefaultVideoService) GetDuration(data []byte) (time.Duration, error) {
	if vs.ffprobePath == "" {
		return 0, fmt.Errorf("FFprobe not available")
	}

	// 检测视频格式
	format := vs.detectVideoFormat(data)
	if format == "unknown" {
		return 0, fmt.Errorf("unknown video format")
	}

	// 创建临时文件
	tempFile := filepath.Join(vs.tempDir, fmt.Sprintf("video_%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tempFile)

	// 写入临时文件
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write temp file: %w", err)
	}

	// 使用ffprobe获取时长
	cmd := exec.Command(vs.ffprobePath, "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", tempFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	// 解析时长
	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	duration := time.Duration(durationFloat * float64(time.Second))
	vs.logger.Info("Video duration extracted", zap.Duration("duration", duration))

	return duration, nil
}

// ProcessVideo 处理视频数据
func (vs *DefaultVideoService) ProcessVideo(data []byte, targetFormat string) ([]byte, error) {
	if vs.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// 检测输入格式
	inputFormat := vs.detectVideoFormat(data)
	if inputFormat == "unknown" {
		return nil, fmt.Errorf("unknown input format")
	}

	// 如果格式相同，直接返回
	if inputFormat == targetFormat {
		return data, nil
	}

	// 创建临时文件
	inputFile := filepath.Join(vs.tempDir, fmt.Sprintf("input_%d.%s", time.Now().UnixNano(), inputFormat))
	outputFile := filepath.Join(vs.tempDir, fmt.Sprintf("output_%d.%s", time.Now().UnixNano(), targetFormat))
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 执行转换
	cmd := exec.Command(vs.ffmpegPath, "-i", inputFile, "-c:v", "libx264", "-c:a", "aac", "-y", outputFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert video: %w", err)
	}

	// 读取输出文件
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	vs.logger.Info("Video converted successfully",
		zap.String("input_format", inputFormat),
		zap.String("target_format", targetFormat),
		zap.Int("input_size", len(data)),
		zap.Int("output_size", len(outputData)))

	return outputData, nil
}

// detectVideoFormat 检测视频格式
func (vs *DefaultVideoService) detectVideoFormat(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}

	// MP4格式检测
	if len(data) >= 8 && bytes.Equal(data[4:8], []byte("ftyp")) {
		return "mp4"
	}

	// AVI格式检测
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("AVI ")) {
		return "avi"
	}

	// MOV格式检测
	if len(data) >= 8 && (bytes.Equal(data[4:8], []byte("moov")) || bytes.Equal(data[4:8], []byte("mdat"))) {
		return "mov"
	}

	// WebM格式检测
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return "webm"
	}

	// MKV格式检测
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		// WebM和MKV都使用Matroska容器，需要进一步检测
		if vs.isWebM(data) {
			return "webm"
		}
		return "mkv"
	}

	// FLV格式检测
	if len(data) >= 3 && bytes.Equal(data[0:3], []byte("FLV")) {
		return "flv"
	}

	// WMV格式检测
	if len(data) >= 16 && bytes.Equal(data[0:16], []byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}) {
		return "wmv"
	}

	return "unknown"
}

// isWebM 检查是否为WebM格式
func (vs *DefaultVideoService) isWebM(data []byte) bool {
	// 简单的WebM检测，实际实现可能需要更复杂的逻辑
	return bytes.Contains(data[:1024], []byte("webm"))
}

// ExtractFrames 提取视频帧
func (vs *DefaultVideoService) ExtractFrames(data []byte, count int) ([][]byte, error) {
	if vs.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// 检测视频格式
	format := vs.detectVideoFormat(data)
	if format == "unknown" {
		return nil, fmt.Errorf("unknown video format")
	}

	// 创建临时文件
	inputFile := filepath.Join(vs.tempDir, fmt.Sprintf("video_%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(inputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	var frames [][]byte
	frameDir := filepath.Join(vs.tempDir, fmt.Sprintf("frames_%d", time.Now().UnixNano()))
	defer os.RemoveAll(frameDir)

	if err := os.MkdirAll(frameDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create frame directory: %w", err)
	}

	// 提取帧
	framePattern := filepath.Join(frameDir, "frame_%03d.jpg")
	cmd := exec.Command(vs.ffmpegPath, "-i", inputFile, "-vf", fmt.Sprintf("select='not(mod(n\\,%d))'", count), "-vsync", "vfr", framePattern)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract frames: %w", err)
	}

	// 读取提取的帧
	files, err := filepath.Glob(filepath.Join(frameDir, "*.jpg"))
	if err != nil {
		return nil, fmt.Errorf("failed to list frame files: %w", err)
	}

	for _, file := range files {
		frameData, err := os.ReadFile(file)
		if err != nil {
			vs.logger.Warn("Failed to read frame file", zap.String("file", file), zap.Error(err))
			continue
		}
		frames = append(frames, frameData)
	}

	vs.logger.Info("Frames extracted successfully", zap.Int("frame_count", len(frames)))

	return frames, nil
}

// GetVideoInfo 获取视频信息
func (vs *DefaultVideoService) GetVideoInfo(data []byte) (*VideoInfo, error) {
	if vs.ffprobePath == "" {
		return nil, fmt.Errorf("FFprobe not available")
	}

	// 检测视频格式
	format := vs.detectVideoFormat(data)
	if format == "unknown" {
		return nil, fmt.Errorf("unknown video format")
	}

	// 创建临时文件
	tempFile := filepath.Join(vs.tempDir, fmt.Sprintf("video_%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tempFile)

	// 写入临时文件
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// 使用ffprobe获取详细信息
	cmd := exec.Command(vs.ffprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", tempFile)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// 解析JSON输出
	var probeResult FFProbeResult
	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// 构建视频信息
	info := &VideoInfo{
		Format:   format,
		Size:     len(data),
		Filename: probeResult.Format.Filename,
	}

	// 解析时长
	if duration, err := strconv.ParseFloat(probeResult.Format.Duration, 64); err == nil {
		info.Duration = time.Duration(duration * float64(time.Second))
	}

	// 解析比特率
	if bitrate, err := strconv.Atoi(probeResult.Format.BitRate); err == nil {
		info.Bitrate = bitrate
	}

	// 查找视频流和音频流
	for _, stream := range probeResult.Streams {
		switch stream.CodecType {
		case "video":
			info.VideoCodec = stream.CodecName
			info.Width = stream.Width
			info.Height = stream.Height
			if fps, err := vs.parseFrameRate(stream.RFrameRate); err == nil {
				info.FrameRate = fps
			}
		case "audio":
			info.AudioCodec = stream.CodecName
			if sampleRate, err := strconv.Atoi(stream.SampleRate); err == nil {
				info.SampleRate = sampleRate
			}
			info.Channels = stream.Channels
		}
	}

	vs.logger.Info("Video info extracted successfully",
		zap.String("format", info.Format),
		zap.Duration("duration", info.Duration),
		zap.Int("width", info.Width),
		zap.Int("height", info.Height))

	return info, nil
}

// parseFrameRate 解析帧率
func (vs *DefaultVideoService) parseFrameRate(frameRateStr string) (float64, error) {
	// 帧率可能是分数形式，如 "30/1" 或 "29.97"
	if strings.Contains(frameRateStr, "/") {
		parts := strings.Split(frameRateStr, "/")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid frame rate format: %s", frameRateStr)
		}
		
		numerator, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		
		denominator, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		
		if denominator == 0 {
			return 0, fmt.Errorf("division by zero in frame rate")
		}
		
		return numerator / denominator, nil
	}
	
	return strconv.ParseFloat(frameRateStr, 64)
}

// CompressVideo 压缩视频
func (vs *DefaultVideoService) CompressVideo(data []byte, quality string) ([]byte, error) {
	if vs.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// 检测输入格式
	inputFormat := vs.detectVideoFormat(data)
	if inputFormat == "unknown" {
		return nil, fmt.Errorf("unknown input format")
	}

	// 创建临时文件
	inputFile := filepath.Join(vs.tempDir, fmt.Sprintf("input_%d.%s", time.Now().UnixNano(), inputFormat))
	outputFile := filepath.Join(vs.tempDir, fmt.Sprintf("output_%d.mp4", time.Now().UnixNano()))
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 根据质量设置CRF值
	var crf string
	switch quality {
	case "high":
		crf = "18"
	case "medium":
		crf = "23"
	case "low":
		crf = "28"
	default:
		crf = "23" // 默认中等质量
	}

	// 执行压缩
	cmd := exec.Command(vs.ffmpegPath, "-i", inputFile, "-c:v", "libx264", "-crf", crf, "-c:a", "aac", "-y", outputFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compress video: %w", err)
	}

	// 读取输出文件
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	vs.logger.Info("Video compressed successfully",
		zap.String("quality", quality),
		zap.String("crf", crf),
		zap.Int("input_size", len(data)),
		zap.Int("output_size", len(outputData)),
		zap.Float64("compression_ratio", float64(len(outputData))/float64(len(data))))

	return outputData, nil
}

// ExtractAudio 从视频中提取音频
func (vs *DefaultVideoService) ExtractAudio(data []byte, audioFormat string) ([]byte, error) {
	if vs.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// 检测视频格式
	videoFormat := vs.detectVideoFormat(data)
	if videoFormat == "unknown" {
		return nil, fmt.Errorf("unknown video format")
	}

	// 创建临时文件
	inputFile := filepath.Join(vs.tempDir, fmt.Sprintf("video_%d.%s", time.Now().UnixNano(), videoFormat))
	outputFile := filepath.Join(vs.tempDir, fmt.Sprintf("audio_%d.%s", time.Now().UnixNano(), audioFormat))
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 提取音频
	var codec string
	switch audioFormat {
	case "mp3":
		codec = "libmp3lame"
	case "aac":
		codec = "aac"
	case "wav":
		codec = "pcm_s16le"
	case "flac":
		codec = "flac"
	default:
		codec = "aac"
	}

	cmd := exec.Command(vs.ffmpegPath, "-i", inputFile, "-vn", "-c:a", codec, "-y", outputFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract audio: %w", err)
	}

	// 读取输出文件
	audioData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	vs.logger.Info("Audio extracted successfully",
		zap.String("audio_format", audioFormat),
		zap.String("codec", codec),
		zap.Int("video_size", len(data)),
		zap.Int("audio_size", len(audioData)))

	return audioData, nil
}

// VideoInfo 视频信息
type VideoInfo struct {
	Format      string        `json:"format"`
	Duration    time.Duration `json:"duration"`
	Size        int           `json:"size"`
	Width       int           `json:"width"`
	Height      int           `json:"height"`
	FrameRate   float64       `json:"frame_rate"`
	Bitrate     int           `json:"bitrate"`
	VideoCodec  string        `json:"video_codec"`
	AudioCodec  string        `json:"audio_codec"`
	SampleRate  int           `json:"sample_rate"`
	Channels    int           `json:"channels"`
	Filename    string        `json:"filename"`
}

// FFProbeResult FFprobe输出结构
type FFProbeResult struct {
	Streams []FFProbeStream `json:"streams"`
	Format  FFProbeFormat   `json:"format"`
}

// FFProbeStream 流信息
type FFProbeStream struct {
	Index         int    `json:"index"`
	CodecName     string `json:"codec_name"`
	CodecType     string `json:"codec_type"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	RFrameRate    string `json:"r_frame_rate"`
	SampleRate    string `json:"sample_rate"`
	Channels      int    `json:"channels"`
	Duration      string `json:"duration"`
	BitRate       string `json:"bit_rate"`
}

// FFProbeFormat 格式信息
type FFProbeFormat struct {
	Filename   string `json:"filename"`
	NbStreams  int    `json:"nb_streams"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
}