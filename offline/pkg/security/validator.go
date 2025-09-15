package security

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Validator 安全验证器
type Validator struct {
	allowedPaths []string
	blockedPaths []string
}

// NewValidator 创建新的安全验证器
func NewValidator() *Validator {
	return &Validator{
		allowedPaths: []string{
			".",
			"./",
			"../",
		},
		blockedPaths: []string{
			"/etc",
			"/sys",
			"/proc",
			"/dev",
			"/root",
			"C:\\Windows",
			"C:\\System32",
		},
	}
}

// ValidateFilePath 验证文件路径安全性
func (v *Validator) ValidateFilePath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)
	
	// 检查是否包含危险字符
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("路径包含危险的上级目录引用: %s", path)
	}
	
	// 检查是否在阻止列表中
	for _, blocked := range v.blockedPaths {
		if strings.HasPrefix(cleanPath, blocked) {
			return fmt.Errorf("路径被阻止访问: %s", path)
		}
	}
	
	return nil
}

// ValidateCommand 验证命令安全性
func (v *Validator) ValidateCommand(command string, args []string) error {
	// 危险命令列表
	dangerousCommands := []string{
		"rm", "del", "format", "fdisk",
		"sudo", "su", "chmod", "chown",
		"wget", "curl", "nc", "netcat",
		"python", "node", "ruby", "perl",
	}
	
	for _, dangerous := range dangerousCommands {
		if strings.EqualFold(command, dangerous) {
			return fmt.Errorf("命令被阻止执行: %s", command)
		}
	}
	
	// 检查参数中的危险模式
	for _, arg := range args {
		if strings.Contains(arg, "--force") || strings.Contains(arg, "-f") {
			return fmt.Errorf("检测到危险参数: %s", arg)
		}
	}
	
	return nil
}

// ValidateInput 验证用户输入
func (v *Validator) ValidateInput(input string) error {
	// 检查输入长度
	if len(input) > 10000 {
		return fmt.Errorf("输入内容过长，最大允许10000字符")
	}
	
	// 检查是否包含恶意脚本
	maliciousPatterns := []string{
		`<script`,
		`javascript:`,
		`eval\(`,
		`exec\(`,
		`system\(`,
		`shell_exec`,
	}
	
	for _, pattern := range maliciousPatterns {
		matched, _ := regexp.MatchString(pattern, strings.ToLower(input))
		if matched {
			return fmt.Errorf("输入包含潜在恶意内容")
		}
	}
	
	return nil
}

// SanitizeFilename 清理文件名
func (v *Validator) SanitizeFilename(filename string) string {
	// 移除危险字符
	dangerousChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := dangerousChars.ReplaceAllString(filename, "_")
	
	// 限制长度
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}
	
	return sanitized
}

// IsAllowedFileType 检查文件类型是否允许
func (v *Validator) IsAllowedFileType(filename string) bool {
	allowedExtensions := []string{
		".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h",
		".md", ".txt", ".json", ".yaml", ".yml", ".xml",
		".html", ".css", ".scss", ".less",
		".sql", ".sh", ".bat", ".ps1",
	}
	
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	
	return false
}

// ValidateProjectName 验证项目名称
func (v *Validator) ValidateProjectName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("项目名称不能为空")
	}
	
	if len(name) > 100 {
		return fmt.Errorf("项目名称过长，最大允许100字符")
	}
	
	// 检查是否包含有效字符
	validName := regexp.MustCompile(`^[a-zA-Z0-9\u4e00-\u9fa5_\-\s]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("项目名称包含无效字符")
	}
	
	return nil
}