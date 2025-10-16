package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DefaultAudioService 默认音频服务实现
type DefaultAudioService struct {
	tempDir    string
	ffmpegPath string
	logger     *zap.Logger
}

// NewDefaultAudioService 创建新的默认音频服务
func NewDefaultAudioService(tempDir, ffmpegPath string, logger *zap.Logger) *DefaultAudioService {
	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Error("Failed to create temp directory", zap.Error(err))
	}

	return &DefaultAudioService{
		tempDir:    tempDir,
		ffmpegPath: ffmpegPath,
		logger:     logger,
	}
}

// GetDuration 获取音频时长
func (as *DefaultAudioService) GetDuration(data []byte) (float64, error) {
	// 检测音频格式
	format := as.detectAudioFormat(data)
	
	switch format {
	case "wav":
		return as.getWAVDuration(data)
	case "mp3":
		return as.getMP3Duration(data)
	default:
		// 使用FFmpeg获取时长
		return as.getDurationWithFFmpeg(data, format)
	}
}

// ProcessAudio 处理音频数据
func (as *DefaultAudioService) ProcessAudio(data []byte, targetFormat string) ([]byte, error) {
	// 检测输入格式
	inputFormat := as.detectAudioFormat(data)
	
	// 如果格式相同且是支持的格式，直接返回
	if inputFormat == targetFormat && as.isSupportedFormat(targetFormat) {
		return data, nil
	}

	// 使用FFmpeg进行格式转换
	return as.convertAudioWithFFmpeg(data, inputFormat, targetFormat)
}

// detectAudioFormat 检测音频格式
func (as *DefaultAudioService) detectAudioFormat(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}

	// WAV格式检测
	if bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WAVE")) {
		return "wav"
	}

	// MP3格式检测
	if len(data) >= 3 && (bytes.Equal(data[0:3], []byte("ID3")) || 
		(data[0] == 0xFF && (data[1]&0xE0) == 0xE0)) {
		return "mp3"
	}

	// FLAC格式检测
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte("fLaC")) {
		return "flac"
	}

	// OGG格式检测
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte("OggS")) {
		return "ogg"
	}

	// AAC格式检测
	if len(data) >= 2 && ((data[0] == 0xFF && (data[1]&0xF0) == 0xF0) ||
		bytes.Equal(data[4:8], []byte("ftyp"))) {
		return "aac"
	}

	return "unknown"
}

// getWAVDuration 获取WAV文件时长
func (as *DefaultAudioService) getWAVDuration(data []byte) (float64, error) {
	if len(data) < 44 {
		return 0, fmt.Errorf("invalid WAV file: too short")
	}

	// 检查WAV头
	if !bytes.Equal(data[0:4], []byte("RIFF")) || !bytes.Equal(data[8:12], []byte("WAVE")) {
		return 0, fmt.Errorf("invalid WAV file: missing RIFF/WAVE header")
	}

	// 查找fmt chunk
	offset := 12
	for offset < len(data)-8 {
		chunkID := data[offset : offset+4]
		chunkSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		if bytes.Equal(chunkID, []byte("fmt ")) {
			if offset+8+int(chunkSize) > len(data) {
				return 0, fmt.Errorf("invalid WAV file: fmt chunk size exceeds file")
			}

			// 读取音频参数
			_ = binary.LittleEndian.Uint32(data[offset+12 : offset+16]) // sampleRate
			byteRate := binary.LittleEndian.Uint32(data[offset+16 : offset+20])

			if byteRate == 0 {
				return 0, fmt.Errorf("invalid WAV file: zero byte rate")
			}

			// 查找data chunk
			dataOffset := offset + 8 + int(chunkSize)
			for dataOffset < len(data)-8 {
				dataChunkID := data[dataOffset : dataOffset+4]
				dataChunkSize := binary.LittleEndian.Uint32(data[dataOffset+4 : dataOffset+8])

				if bytes.Equal(dataChunkID, []byte("data")) {
					duration := float64(dataChunkSize) / float64(byteRate)
					return duration, nil
				}

				dataOffset += 8 + int(dataChunkSize)
			}

			return 0, fmt.Errorf("invalid WAV file: data chunk not found")
		}

		offset += 8 + int(chunkSize)
	}

	return 0, fmt.Errorf("invalid WAV file: fmt chunk not found")
}

// getMP3Duration 获取MP3文件时长（简化实现）
func (as *DefaultAudioService) getMP3Duration(data []byte) (float64, error) {
	// MP3时长计算比较复杂，这里使用FFmpeg
	return as.getDurationWithFFmpeg(data, "mp3")
}

// getDurationWithFFmpeg 使用FFmpeg获取音频时长
func (as *DefaultAudioService) getDurationWithFFmpeg(data []byte, format string) (float64, error) {
	if as.ffmpegPath == "" {
		return 0, fmt.Errorf("FFmpeg not available")
	}

	// 创建临时文件
	tempFile := filepath.Join(as.tempDir, fmt.Sprintf("temp_audio_%d.%s", time.Now().UnixNano(), format))
	defer os.Remove(tempFile)

	// 写入临时文件
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write temp file: %w", err)
	}

	// 使用FFprobe获取时长
	cmd := exec.Command(as.ffmpegPath, "-i", tempFile, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get duration with FFmpeg: %w", err)
	}

	// 解析输出
	durationStr := strings.TrimSpace(string(output))
	var duration float64
	if _, err := fmt.Sscanf(durationStr, "%f", &duration); err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// convertAudioWithFFmpeg 使用FFmpeg转换音频格式
func (as *DefaultAudioService) convertAudioWithFFmpeg(data []byte, inputFormat, outputFormat string) ([]byte, error) {
	if as.ffmpegPath == "" {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// 创建临时文件
	inputFile := filepath.Join(as.tempDir, fmt.Sprintf("input_%d.%s", time.Now().UnixNano(), inputFormat))
	outputFile := filepath.Join(as.tempDir, fmt.Sprintf("output_%d.%s", time.Now().UnixNano(), outputFormat))
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// 写入输入文件
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 构建FFmpeg命令
	args := []string{"-i", inputFile}
	
	// 根据输出格式设置参数
	switch outputFormat {
	case "wav":
		args = append(args, "-acodec", "pcm_s16le", "-ar", "44100", "-ac", "2")
	case "mp3":
		args = append(args, "-acodec", "libmp3lame", "-ab", "128k", "-ar", "44100", "-ac", "2")
	case "flac":
		args = append(args, "-acodec", "flac", "-ar", "44100", "-ac", "2")
	case "aac":
		args = append(args, "-acodec", "aac", "-ab", "128k", "-ar", "44100", "-ac", "2")
	}

	args = append(args, "-y", outputFile)

	// 执行转换
	cmd := exec.Command(as.ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert audio with FFmpeg: %w", err)
	}

	// 读取输出文件
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	as.logger.Info("Audio converted successfully",
		zap.String("input_format", inputFormat),
		zap.String("output_format", outputFormat),
		zap.Int("input_size", len(data)),
		zap.Int("output_size", len(outputData)))

	return outputData, nil
}

// isSupportedFormat 检查是否为支持的格式
func (as *DefaultAudioService) isSupportedFormat(format string) bool {
	supportedFormats := map[string]bool{
		"wav":  true,
		"mp3":  true,
		"flac": true,
		"aac":  true,
		"ogg":  true,
	}
	return supportedFormats[format]
}

// NormalizeAudio 音频标准化
func (as *DefaultAudioService) NormalizeAudio(data []byte) ([]byte, error) {
	// 检测格式
	format := as.detectAudioFormat(data)
	if format != "wav" {
		// 先转换为WAV格式
		wavData, err := as.ProcessAudio(data, "wav")
		if err != nil {
			return nil, fmt.Errorf("failed to convert to WAV: %w", err)
		}
		data = wavData
	}

	// 对WAV数据进行标准化
	return as.normalizeWAVData(data)
}

// normalizeWAVData 标准化WAV数据
func (as *DefaultAudioService) normalizeWAVData(data []byte) ([]byte, error) {
	if len(data) < 44 {
		return nil, fmt.Errorf("invalid WAV file: too short")
	}

	// 复制头部
	result := make([]byte, len(data))
	copy(result, data)

	// 查找data chunk
	offset := 12
	for offset < len(data)-8 {
		chunkID := data[offset : offset+4]
		chunkSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		if bytes.Equal(chunkID, []byte("data")) {
			// 处理音频数据
			audioData := result[offset+8 : offset+8+int(chunkSize)]
			as.normalizeAudioSamples(audioData)
			break
		}

		offset += 8 + int(chunkSize)
	}

	return result, nil
}

// normalizeAudioSamples 标准化音频样本
func (as *DefaultAudioService) normalizeAudioSamples(data []byte) {
	if len(data)%2 != 0 {
		return // 确保是16位样本
	}

	// 找到最大振幅
	var maxAmplitude int16
	for i := 0; i < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		if sample < 0 {
			sample = -sample
		}
		if sample > maxAmplitude {
			maxAmplitude = sample
		}
	}

	if maxAmplitude == 0 {
		return // 避免除零
	}

	// 计算标准化因子
	targetAmplitude := int16(29490) // 90%的最大振幅 (32767 * 0.9)
	factor := float64(targetAmplitude) / float64(maxAmplitude)

	// 应用标准化
	for i := 0; i < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		normalizedSample := int16(float64(sample) * factor)
		binary.LittleEndian.PutUint16(data[i:i+2], uint16(normalizedSample))
	}
}

// ExtractFeatures 提取音频特征
func (as *DefaultAudioService) ExtractFeatures(data []byte) (*AudioFeatures, error) {
	// 确保是WAV格式
	format := as.detectAudioFormat(data)
	if format != "wav" {
		wavData, err := as.ProcessAudio(data, "wav")
		if err != nil {
			return nil, fmt.Errorf("failed to convert to WAV: %w", err)
		}
		data = wavData
	}

	duration, err := as.GetDuration(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// 提取基本特征
	features := &AudioFeatures{
		Duration:    duration,
		Format:      "wav",
		SampleRate:  44100, // 假设标准采样率
		Channels:    2,     // 假设立体声
		BitDepth:    16,    // 假设16位
		Energy:      as.calculateEnergy(data),
		ZeroCrossings: as.calculateZeroCrossings(data),
	}

	return features, nil
}

// calculateEnergy 计算音频能量
func (as *DefaultAudioService) calculateEnergy(data []byte) float64 {
	if len(data) < 44 {
		return 0
	}

	// 查找data chunk
	offset := 12
	for offset < len(data)-8 {
		chunkID := data[offset : offset+4]
		chunkSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		if bytes.Equal(chunkID, []byte("data")) {
			audioData := data[offset+8 : offset+8+int(chunkSize)]
			return as.calculateSampleEnergy(audioData)
		}

		offset += 8 + int(chunkSize)
	}

	return 0
}

// calculateSampleEnergy 计算样本能量
func (as *DefaultAudioService) calculateSampleEnergy(data []byte) float64 {
	if len(data)%2 != 0 {
		return 0
	}

	var totalEnergy float64
	sampleCount := len(data) / 2

	for i := 0; i < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		totalEnergy += float64(sample * sample)
	}

	return math.Sqrt(totalEnergy / float64(sampleCount))
}

// calculateZeroCrossings 计算过零率
func (as *DefaultAudioService) calculateZeroCrossings(data []byte) int {
	if len(data) < 44 {
		return 0
	}

	// 查找data chunk
	offset := 12
	for offset < len(data)-8 {
		chunkID := data[offset : offset+4]
		chunkSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		if bytes.Equal(chunkID, []byte("data")) {
			audioData := data[offset+8 : offset+8+int(chunkSize)]
			return as.calculateSampleZeroCrossings(audioData)
		}

		offset += 8 + int(chunkSize)
	}

	return 0
}

// calculateSampleZeroCrossings 计算样本过零率
func (as *DefaultAudioService) calculateSampleZeroCrossings(data []byte) int {
	if len(data) < 4 {
		return 0
	}

	var crossings int
	prevSample := int16(binary.LittleEndian.Uint16(data[0:2]))

	for i := 2; i < len(data); i += 2 {
		currentSample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		if (prevSample >= 0 && currentSample < 0) || (prevSample < 0 && currentSample >= 0) {
			crossings++
		}
		prevSample = currentSample
	}

	return crossings
}

// AudioFeatures 音频特征
type AudioFeatures struct {
	Duration      float64 `json:"duration"`
	Format        string  `json:"format"`
	SampleRate    int     `json:"sample_rate"`
	Channels      int     `json:"channels"`
	BitDepth      int     `json:"bit_depth"`
	Energy        float64 `json:"energy"`
	ZeroCrossings int     `json:"zero_crossings"`
}