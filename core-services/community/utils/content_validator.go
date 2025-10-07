package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ContentValidationResult 内容验证结果
type ContentValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	RiskLevel    int      `json:"risk_level"`    // 0-低风险, 1-中风险, 2-高风险
	FilteredText string   `json:"filtered_text"` // 过滤后的文本
	RiskFactors  []string `json:"risk_factors"`  // 风险因素列表
	Score        int      `json:"score"`         // 风险评分 (0-100)
}

// ContentValidator 内容验证器
type ContentValidator struct {
	sensitiveWords []string
	bannedWords    []string
	spamPatterns   []*regexp.Regexp
}

// NewContentValidator 创建内容验证器实例
func NewContentValidator() *ContentValidator {
	validator := &ContentValidator{
		sensitiveWords: getDefaultSensitiveWords(),
		bannedWords:    getDefaultBannedWords(),
		spamPatterns:   getDefaultSpamPatterns(),
	}
	return validator
}

// ValidatePostContent 验证帖子内容
func (cv *ContentValidator) ValidatePostContent(title, content string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 验证标题
	if titleResult := cv.validateTitle(title); !titleResult.IsValid {
		result.IsValid = false
		result.Errors = append(result.Errors, titleResult.Errors...)
		result.Warnings = append(result.Warnings, titleResult.Warnings...)
		result.RiskFactors = append(result.RiskFactors, titleResult.RiskFactors...)
		result.Score += titleResult.Score
		if titleResult.RiskLevel > result.RiskLevel {
			result.RiskLevel = titleResult.RiskLevel
		}
	}

	// 验证内容
	if contentResult := cv.validateContent(content); !contentResult.IsValid {
		result.IsValid = false
		result.Errors = append(result.Errors, contentResult.Errors...)
		result.Warnings = append(result.Warnings, contentResult.Warnings...)
		result.RiskFactors = append(result.RiskFactors, contentResult.RiskFactors...)
		result.Score += contentResult.Score
		if contentResult.RiskLevel > result.RiskLevel {
			result.RiskLevel = contentResult.RiskLevel
		}
	}

	// 综合风险评估
	cv.assessOverallRisk(result)

	// 生成过滤后的文本
	result.FilteredText = cv.filterSensitiveWords(content)

	return result
}

// ValidateCommentContent 验证评论内容
func (cv *ContentValidator) ValidateCommentContent(content string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 验证评论长度
	if utf8.RuneCountInString(content) < 1 {
		result.IsValid = false
		result.Errors = append(result.Errors, "评论内容不能为空")
		result.RiskFactors = append(result.RiskFactors, "内容为空")
		result.Score += 50
	}

	if utf8.RuneCountInString(content) > 1000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "评论内容不能超过1000个字符")
		result.RiskFactors = append(result.RiskFactors, "内容过长")
		result.Score += 30
	}

	// 检查敏感词
	if cv.containsBannedWords(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "评论包含禁用词汇")
		result.RiskFactors = append(result.RiskFactors, "包含禁用词")
		result.Score += 80
		result.RiskLevel = 2
	}

	if cv.containsSensitiveWords(content) {
		result.Warnings = append(result.Warnings, "评论包含敏感词汇，需要人工审核")
		result.RiskFactors = append(result.RiskFactors, "包含敏感词")
		result.Score += 40
		result.RiskLevel = 1
	}

	// 检查垃圾内容
	if cv.isSpamContent(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到垃圾内容")
		result.RiskFactors = append(result.RiskFactors, "垃圾内容")
		result.Score += 70
		result.RiskLevel = 2
	}

	// 综合风险评估
	cv.assessOverallRisk(result)

	// 生成过滤后的文本
	result.FilteredText = cv.filterSensitiveWords(content)

	return result
}

// validateTitle 验证标题
func (cv *ContentValidator) validateTitle(title string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 长度验证
	titleLength := utf8.RuneCountInString(title)
	if titleLength < 5 {
		result.IsValid = false
		result.Errors = append(result.Errors, "标题至少需要5个字符")
		result.RiskFactors = append(result.RiskFactors, "标题过短")
		result.Score += 30
	}

	if titleLength > 100 {
		result.IsValid = false
		result.Errors = append(result.Errors, "标题不能超过100个字符")
		result.RiskFactors = append(result.RiskFactors, "标题过长")
		result.Score += 25
	}

	// 格式验证
	if strings.TrimSpace(title) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "标题不能为空或只包含空格")
		result.RiskFactors = append(result.RiskFactors, "标题为空")
		result.Score += 50
	}

	// 检查是否全是特殊字符或数字
	if matched, _ := regexp.MatchString(`^[^\p{L}]*$`, title); matched {
		result.IsValid = false
		result.Errors = append(result.Errors, "标题必须包含有意义的文字内容")
		result.RiskFactors = append(result.RiskFactors, "标题无意义")
		result.Score += 40
	}

	// 敏感词检查
	if cv.containsBannedWords(title) {
		result.IsValid = false
		result.Errors = append(result.Errors, "标题包含禁用词汇")
		result.RiskFactors = append(result.RiskFactors, "标题包含禁用词")
		result.Score += 80
		result.RiskLevel = 2
	}

	if cv.containsSensitiveWords(title) {
		result.Warnings = append(result.Warnings, "标题包含敏感词汇，需要人工审核")
		result.RiskFactors = append(result.RiskFactors, "标题包含敏感词")
		result.Score += 40
		result.RiskLevel = 1
	}

	return result
}

// validateContent 验证内容
func (cv *ContentValidator) validateContent(content string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 长度验证
	contentLength := utf8.RuneCountInString(content)
	if contentLength < 10 {
		result.IsValid = false
		result.Errors = append(result.Errors, "内容至少需要10个字符")
		result.RiskFactors = append(result.RiskFactors, "内容过短")
		result.Score += 25
	}

	if contentLength > 10000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "内容不能超过10000个字符")
		result.RiskFactors = append(result.RiskFactors, "内容过长")
		result.Score += 20
	}

	// 格式验证
	if strings.TrimSpace(content) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "内容不能为空或只包含空格")
		result.RiskFactors = append(result.RiskFactors, "内容为空")
		result.Score += 50
	}

	// 敏感词检查
	if cv.containsBannedWords(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "内容包含禁用词汇")
		result.RiskFactors = append(result.RiskFactors, "内容包含禁用词")
		result.Score += 80
		result.RiskLevel = 2
	}

	if cv.containsSensitiveWords(content) {
		result.Warnings = append(result.Warnings, "内容包含敏感词汇，需要人工审核")
		result.RiskFactors = append(result.RiskFactors, "内容包含敏感词")
		result.Score += 40
		result.RiskLevel = 1
	}

	// 垃圾内容检查
	if cv.isSpamContent(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到垃圾内容")
		result.RiskFactors = append(result.RiskFactors, "垃圾内容")
		result.Score += 70
		result.RiskLevel = 2
	}

	return result
}

// containsBannedWords 检查是否包含禁用词
func (cv *ContentValidator) containsBannedWords(text string) bool {
	lowerText := strings.ToLower(text)
	for _, word := range cv.bannedWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// containsSensitiveWords 检查是否包含敏感词
func (cv *ContentValidator) containsSensitiveWords(text string) bool {
	lowerText := strings.ToLower(text)
	for _, word := range cv.sensitiveWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// isSpamContent 检查是否为垃圾内容
func (cv *ContentValidator) isSpamContent(text string) bool {
	for _, pattern := range cv.spamPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}

	// 检查重复字符
	if cv.hasExcessiveRepetition(text) {
		return true
	}

	// 检查是否包含过多链接
	if cv.hasExcessiveLinks(text) {
		return true
	}

	return false
}

// hasExcessiveRepetition 检查是否有过度重复
func (cv *ContentValidator) hasExcessiveRepetition(text string) bool {
	// 检查连续重复字符（超过5个相同字符）
	pattern := regexp.MustCompile(`(.)\\1{5,}`)
	return pattern.MatchString(text)
}

// hasExcessiveLinks 检查是否包含过多链接
func (cv *ContentValidator) hasExcessiveLinks(text string) bool {
	// 简单的URL检测
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlPattern.FindAllString(text, -1)
	return len(matches) > 3 // 超过3个链接认为是垃圾内容
}

// filterSensitiveWords 过滤敏感词
func (cv *ContentValidator) filterSensitiveWords(text string) string {
	filtered := text
	for _, word := range cv.sensitiveWords {
		if strings.Contains(strings.ToLower(filtered), strings.ToLower(word)) {
			replacement := strings.Repeat("*", utf8.RuneCountInString(word))
			filtered = strings.ReplaceAll(filtered, word, replacement)
		}
	}
	return filtered
}

// getDefaultSensitiveWords 获取默认敏感词列表
func getDefaultSensitiveWords() []string {
	return []string{
		// 政治敏感词
		"政治敏感词1", "政治敏感词2",
		// 暴力相关
		"暴力", "血腥", "杀害",
		// 色情相关
		"色情", "黄色", "成人",
		// 赌博相关
		"赌博", "博彩", "彩票",
		// 其他敏感词
		"诈骗", "传销", "邪教",
	}
}

// getDefaultBannedWords 获取默认禁用词列表
func getDefaultBannedWords() []string {
	return []string{
		// 严重违法违规词汇
		"恐怖主义", "分裂国家", "颠覆政权",
		"制毒", "贩毒", "走私",
		"人体器官买卖", "儿童色情",
		// 极端仇恨言论
		"种族歧视", "宗教仇恨",
	}
}

// getDefaultSpamPatterns 获取默认垃圾内容模式
func getDefaultSpamPatterns() []*regexp.Regexp {
	patterns := []string{
		// QQ号码模式
		`\bQQ[:：]?\s*\d{5,12}\b`,
		// 微信号模式
		`\b微信[:：]?\s*[a-zA-Z0-9_-]{6,20}\b`,
		// 电话号码模式
		`\b1[3-9]\d{9}\b`,
		// 重复的广告词
		`(.{10,})\1{2,}`,
		// 常见垃圾词组合
		`(加我|联系我|私聊|代理|招聘|兼职|赚钱|月入|日赚).{0,20}(微信|QQ|\d{5,})`,
	}

	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		}
	}

	return compiledPatterns
}

// AddSensitiveWords 添加敏感词
func (cv *ContentValidator) AddSensitiveWords(words []string) {
	cv.sensitiveWords = append(cv.sensitiveWords, words...)
}

// AddBannedWords 添加禁用词
func (cv *ContentValidator) AddBannedWords(words []string) {
	cv.bannedWords = append(cv.bannedWords, words...)
}

// assessOverallRisk 综合风险评估
func (cv *ContentValidator) assessOverallRisk(result *ContentValidationResult) {
	// 根据评分确定风险等级
	if result.Score >= 80 {
		result.RiskLevel = 2 // 高风险
	} else if result.Score >= 40 {
		result.RiskLevel = 1 // 中风险
	} else if result.Score >= 20 {
		result.RiskLevel = 1 // 轻微风险，但仍需注意
	}
	// 分数低于20的保持为0（低风险）

	// 如果已经有错误，确保风险等级至少为1
	if len(result.Errors) > 0 && result.RiskLevel == 0 {
		result.RiskLevel = 1
	}

	// 如果有多个风险因素，增加风险等级
	if len(result.RiskFactors) >= 3 {
		result.RiskLevel = 2
	} else if len(result.RiskFactors) >= 2 && result.RiskLevel < 1 {
		result.RiskLevel = 1
	}
}

// GetRiskLevelDescription 获取风险等级描述
func GetRiskLevelDescription(level int) string {
	switch level {
	case 0:
		return "低风险"
	case 1:
		return "中风险"
	case 2:
		return "高风险"
	default:
		return "未知风险"
	}
}