package utils

import (
	"fmt"
	"strings"
	"unicode"
)

// PasswordStrength 密码强度等级
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordFair
	PasswordGood
	PasswordStrong
	PasswordVeryStrong
)

// PasswordValidationResult 密码验证结果
type PasswordValidationResult struct {
	Valid    bool     `json:"valid"`
	Strength PasswordStrength `json:"strength"`
	Score    int      `json:"score"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// PasswordPolicy 密码策略配置
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSpecial   bool `json:"require_special"`
	MaxLength        int  `json:"max_length"`
	MinScore         int  `json:"min_score"`
}

// DefaultPasswordPolicy 默认密码策略
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:        8,
	RequireUppercase: true,
	RequireLowercase: true,
	RequireNumbers:   true,
	RequireSpecial:   true,
	MaxLength:        128,
	MinScore:         3,
}

// 常见弱密码列表
var commonWeakPasswords = []string{
	"password", "123456", "123456789", "qwerty", "abc123",
	"password123", "admin", "letmein", "welcome", "monkey",
	"1234567890", "qwertyuiop", "asdfghjkl", "zxcvbnm",
	"password1", "123123", "000000", "iloveyou", "1234567",
	"princess", "dragon", "sunshine", "master", "123321",
	"666666", "654321", "7777777", "123", "D1lakiss",
	"555555", "lovely", "888888", "charlie", "donald",
	"freedom", "111111", "121212", "696969", "12345678",
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string, policy *PasswordPolicy) *PasswordValidationResult {
	if policy == nil {
		policy = &DefaultPasswordPolicy
	}

	result := &PasswordValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Score:    0,
	}

	// 基本长度检查
	if len(password) < policy.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("密码长度至少需要%d个字符", policy.MinLength))
	} else {
		result.Score++
	}

	if len(password) > policy.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("密码长度不能超过%d个字符", policy.MaxLength))
	}

	// 字符类型检查
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasNumber = true
		} else if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	// 检查必需的字符类型
	if policy.RequireUppercase && !hasUpper {
		result.Valid = false
		result.Errors = append(result.Errors, "密码必须包含大写字母")
	} else if hasUpper {
		result.Score++
	}

	if policy.RequireLowercase && !hasLower {
		result.Valid = false
		result.Errors = append(result.Errors, "密码必须包含小写字母")
	} else if hasLower {
		result.Score++
	}

	if policy.RequireNumbers && !hasNumber {
		result.Valid = false
		result.Errors = append(result.Errors, "密码必须包含数字")
	} else if hasNumber {
		result.Score++
	}

	if policy.RequireSpecial && !hasSpecial {
		result.Valid = false
		result.Errors = append(result.Errors, "密码必须包含特殊字符")
	} else if hasSpecial {
		result.Score++
	}

	// 额外的强度检查
	result.Score += calculateAdditionalScore(password)

	// 检查常见弱密码
	if isCommonWeakPassword(password) {
		result.Valid = false
		result.Errors = append(result.Errors, "请避免使用常见的弱密码")
		result.Score = max(0, result.Score-2)
	}

	// 检查重复字符
	if hasRepeatingChars(password) {
		result.Warnings = append(result.Warnings, "避免使用连续重复的字符")
		result.Score = max(0, result.Score-1)
	}

	// 检查键盘模式
	if hasKeyboardPattern(password) {
		result.Warnings = append(result.Warnings, "避免使用键盘上的连续字符")
		result.Score = max(0, result.Score-1)
	}

	// 确定强度等级
	result.Strength = determineStrength(result.Score)

	// 检查是否满足最低分数要求
	if result.Score < policy.MinScore {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("密码强度不足，当前分数：%d，最低要求：%d", result.Score, policy.MinScore))
	}

	return result
}

// calculateAdditionalScore 计算额外分数
func calculateAdditionalScore(password string) int {
	score := 0

	// 长度奖励
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// 字符多样性奖励
	uniqueChars := make(map[rune]bool)
	for _, char := range password {
		uniqueChars[char] = true
	}
	
	if len(uniqueChars) >= len(password)*3/4 {
		score++
	}

	return score
}

// isCommonWeakPassword 检查是否为常见弱密码
func isCommonWeakPassword(password string) bool {
	lowerPassword := strings.ToLower(password)
	for _, weak := range commonWeakPasswords {
		// 精确匹配或者密码主要由弱密码组成
		if lowerPassword == weak || 
		   (len(weak) >= 6 && strings.Contains(lowerPassword, weak) && len(lowerPassword) <= len(weak)+3) {
			return true
		}
	}
	return false
}

// hasRepeatingChars 检查是否有连续重复字符
func hasRepeatingChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}
	return false
}

// hasKeyboardPattern 检查是否有键盘模式
func hasKeyboardPattern(password string) bool {
	keyboardPatterns := []string{
		"123", "234", "345", "456", "567", "678", "789", "890",
		"qwe", "wer", "ert", "rty", "tyu", "yui", "uio", "iop",
		"asd", "sdf", "dfg", "fgh", "ghj", "hjk", "jkl",
		"zxc", "xcv", "cvb", "vbn", "bnm",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij", "ijk",
	}

	lowerPassword := strings.ToLower(password)
	for _, pattern := range keyboardPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
		// 检查反向模式
		reversed := reverseString(pattern)
		if strings.Contains(lowerPassword, reversed) {
			return true
		}
	}
	return false
}

// determineStrength 确定密码强度等级
func determineStrength(score int) PasswordStrength {
	switch {
	case score >= 7:
		return PasswordVeryStrong
	case score >= 5:
		return PasswordStrong
	case score >= 4:
		return PasswordGood
	case score >= 2:
		return PasswordFair
	default:
		return PasswordWeak
	}
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetStrengthText 获取强度等级文本
func GetStrengthText(strength PasswordStrength) string {
	switch strength {
	case PasswordWeak:
		return "弱"
	case PasswordFair:
		return "一般"
	case PasswordGood:
		return "良好"
	case PasswordStrong:
		return "强"
	case PasswordVeryStrong:
		return "很强"
	default:
		return "未知"
	}
}

// ValidatePasswordWithCustomRules 使用自定义规则验证密码
func ValidatePasswordWithCustomRules(password string, customRules []func(string) (bool, string)) *PasswordValidationResult {
	result := ValidatePassword(password, nil)

	// 应用自定义规则
	for _, rule := range customRules {
		if valid, message := rule(password); !valid {
			result.Valid = false
			result.Errors = append(result.Errors, message)
		}
	}

	return result
}

// CommonCustomRules 常用的自定义规则
var CommonCustomRules = struct {
	NoUserInfo     func(userInfo []string) func(string) (bool, string)
	NoRecentPasswords func(recentPasswords []string) func(string) (bool, string)
	MinUniqueChars func(minUnique int) func(string) (bool, string)
}{
	// 不能包含用户信息
	NoUserInfo: func(userInfo []string) func(string) (bool, string) {
		return func(password string) (bool, string) {
			lowerPassword := strings.ToLower(password)
			for _, info := range userInfo {
				if info != "" && strings.Contains(lowerPassword, strings.ToLower(info)) {
					return false, "密码不能包含用户名、邮箱或个人信息"
				}
			}
			return true, ""
		}
	},
	
	// 不能与最近使用的密码相同
	NoRecentPasswords: func(recentPasswords []string) func(string) (bool, string) {
		return func(password string) (bool, string) {
			for _, recent := range recentPasswords {
				if password == recent {
					return false, "不能使用最近使用过的密码"
				}
			}
			return true, ""
		}
	},
	
	// 最少唯一字符数
	MinUniqueChars: func(minUnique int) func(string) (bool, string) {
		return func(password string) (bool, string) {
			uniqueChars := make(map[rune]bool)
			for _, char := range password {
				uniqueChars[char] = true
			}
			if len(uniqueChars) < minUnique {
				return false, fmt.Sprintf("密码至少需要%d个不同的字符", minUnique)
			}
			return true, ""
		}
	},
}