package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ContentValidationResult 
type ContentValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	RiskLevel    int      `json:"risk_level"`    // 0- 1- 2-
	FilteredText string   `json:"filtered_text"` // 
	RiskFactors  []string `json:"risk_factors"`  // 
	Score        int      `json:"score"`         //  (0-100)
}

// ContentValidator 
type ContentValidator struct {
	sensitiveWords []string
	bannedWords    []string
	spamPatterns   []*regexp.Regexp
}

// NewContentValidator 
func NewContentValidator() *ContentValidator {
	validator := &ContentValidator{
		sensitiveWords: getDefaultSensitiveWords(),
		bannedWords:    getDefaultBannedWords(),
		spamPatterns:   getDefaultSpamPatterns(),
	}
	return validator
}

// ValidatePostContent 
func (cv *ContentValidator) ValidatePostContent(title, content string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 
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

	// 
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

	// 
	cv.assessOverallRisk(result)

	// 
	result.FilteredText = cv.filterSensitiveWords(content)

	return result
}

// ValidateCommentContent 
func (cv *ContentValidator) ValidateCommentContent(content string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 
	if utf8.RuneCountInString(content) < 1 {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 50
	}

	if utf8.RuneCountInString(content) > 1000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "1000")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 30
	}

	// 
	if cv.containsBannedWords(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 80
		result.RiskLevel = 2
	}

	if cv.containsSensitiveWords(content) {
		result.Warnings = append(result.Warnings, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 40
		result.RiskLevel = 1
	}

	// 
	if cv.isSpamContent(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 70
		result.RiskLevel = 2
	}

	// 
	cv.assessOverallRisk(result)

	// 
	result.FilteredText = cv.filterSensitiveWords(content)

	return result
}

// validateTitle 
func (cv *ContentValidator) validateTitle(title string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 
	titleLength := utf8.RuneCountInString(title)
	if titleLength < 5 {
		result.IsValid = false
		result.Errors = append(result.Errors, "5")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 30
	}

	if titleLength > 100 {
		result.IsValid = false
		result.Errors = append(result.Errors, "100")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 25
	}

	// 
	if strings.TrimSpace(title) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 50
	}

	// 
	if matched, _ := regexp.MatchString(`^[^\p{L}]*$`, title); matched {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 40
	}

	// ?
	if cv.containsBannedWords(title) {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 80
		result.RiskLevel = 2
	}

	if cv.containsSensitiveWords(title) {
		result.Warnings = append(result.Warnings, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 40
		result.RiskLevel = 1
	}

	return result
}

// validateContent 
func (cv *ContentValidator) validateContent(content string) *ContentValidationResult {
	result := &ContentValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		RiskLevel:   0,
		RiskFactors: []string{},
		Score:       0,
	}

	// 
	contentLength := utf8.RuneCountInString(content)
	if contentLength < 10 {
		result.IsValid = false
		result.Errors = append(result.Errors, "10")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 25
	}

	if contentLength > 10000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "10000")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 20
	}

	// 
	if strings.TrimSpace(content) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 50
	}

	// ?
	if cv.containsBannedWords(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 80
		result.RiskLevel = 2
	}

	if cv.containsSensitiveWords(content) {
		result.Warnings = append(result.Warnings, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 40
		result.RiskLevel = 1
	}

	// 
	if cv.isSpamContent(content) {
		result.IsValid = false
		result.Errors = append(result.Errors, "")
		result.RiskFactors = append(result.RiskFactors, "")
		result.Score += 70
		result.RiskLevel = 2
	}

	return result
}

// containsBannedWords 
func (cv *ContentValidator) containsBannedWords(text string) bool {
	lowerText := strings.ToLower(text)
	for _, word := range cv.bannedWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// containsSensitiveWords 
func (cv *ContentValidator) containsSensitiveWords(text string) bool {
	lowerText := strings.ToLower(text)
	for _, word := range cv.sensitiveWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// isSpamContent 
func (cv *ContentValidator) isSpamContent(text string) bool {
	for _, pattern := range cv.spamPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}

	// 
	if cv.hasExcessiveRepetition(text) {
		return true
	}

	// 
	if cv.hasExcessiveLinks(text) {
		return true
	}

	return false
}

// hasExcessiveRepetition 
func (cv *ContentValidator) hasExcessiveRepetition(text string) bool {
	// 5
	pattern := regexp.MustCompile(`(.)\\1{5,}`)
	return pattern.MatchString(text)
}

// hasExcessiveLinks 
func (cv *ContentValidator) hasExcessiveLinks(text string) bool {
	// URL
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlPattern.FindAllString(text, -1)
	return len(matches) > 3 // 3
}

// filterSensitiveWords 
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

// getDefaultSensitiveWords 
func getDefaultSensitiveWords() []string {
	return []string{
		// 
		"", "",
		// 
		"", "", "",
		// 
		"", "", "",
		// 
		"", "", "",
		// 
		"", "", "",
	}
}

// getDefaultBannedWords 
func getDefaultBannedWords() []string {
	return []string{
		// 
		"", "", "",
		"", "", "",
		"", "",
		// 
		"", "",
	}
}

// getDefaultSpamPatterns 
func getDefaultSpamPatterns() []*regexp.Regexp {
	patterns := []string{
		// QQ
		`\bQQ[:]?\s*\d{5,12}\b`,
		// 
		`\b[:]?\s*[a-zA-Z0-9_-]{6,20}\b`,
		// 绰
		`\b1[3-9]\d{9}\b`,
		// 
		`(.{10,})\1{2,}`,
		// 
		`(||||||||).{0,20}(|QQ|\d{5,})`,
	}

	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		}
	}

	return compiledPatterns
}

// AddSensitiveWords 
func (cv *ContentValidator) AddSensitiveWords(words []string) {
	cv.sensitiveWords = append(cv.sensitiveWords, words...)
}

// AddBannedWords 
func (cv *ContentValidator) AddBannedWords(words []string) {
	cv.bannedWords = append(cv.bannedWords, words...)
}

// assessOverallRisk 
func (cv *ContentValidator) assessOverallRisk(result *ContentValidationResult) {
	// 
	if result.Score >= 80 {
		result.RiskLevel = 2 // 
	} else if result.Score >= 40 {
		result.RiskLevel = 1 // 
	} else if result.Score >= 20 {
		result.RiskLevel = 1 // 
	}
	// 200

	// 1
	if len(result.Errors) > 0 && result.RiskLevel == 0 {
		result.RiskLevel = 1
	}

	// 
	if len(result.RiskFactors) >= 3 {
		result.RiskLevel = 2
	} else if len(result.RiskFactors) >= 2 && result.RiskLevel < 1 {
		result.RiskLevel = 1
	}
}

// GetRiskLevelDescription 
func GetRiskLevelDescription(level int) string {
	switch level {
	case 0:
		return ""
	case 1:
		return ""
	case 2:
		return ""
	default:
		return ""
	}
}

