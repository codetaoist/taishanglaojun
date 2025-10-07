package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// PasswordUtils 密码工具类
type PasswordUtils struct{}

// HashPassword 使用bcrypt哈希密码
func (p *PasswordUtils) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func (p *PasswordUtils) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword 生成随机密码
func (p *PasswordUtils) GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return "", err
		}
		b[i] = charset[int(randomByte[0])%len(charset)]
	}
	return string(b), nil
}

// ValidatePasswordStrength 验证密码强度
func (p *PasswordUtils) ValidatePasswordStrength(password string) (bool, []string) {
	var errors []string
	
	if len(password) < 8 {
		errors = append(errors, "密码长度至少8位")
	}
	
	if len(password) > 128 {
		errors = append(errors, "密码长度不能超过128位")
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		errors = append(errors, "密码必须包含大写字母")
	}
	
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		errors = append(errors, "密码必须包含小写字母")
	}
	
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		errors = append(errors, "密码必须包含数字")
	}
	
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		errors = append(errors, "密码必须包含特殊字符")
	}
	
	return len(errors) == 0, errors
}

// EncryptionUtils 加密工具类
type EncryptionUtils struct{}

// AESEncrypt AES加密
func (e *EncryptionUtils) AESEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES解密
func (e *EncryptionUtils) AESDecrypt(ciphertext, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateAESKey 生成AES密钥
func (e *EncryptionUtils) GenerateAESKey(keySize int) (string, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// HashUtils 哈希工具类
type HashUtils struct{}

// MD5Hash MD5哈希
func (h *HashUtils) MD5Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash SHA256哈希
func (h *HashUtils) SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA512Hash SHA512哈希
func (h *HashUtils) SHA512Hash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ScryptHash Scrypt哈希（用于密码）
func (h *HashUtils) ScryptHash(password, salt string) (string, error) {
	dk, err := scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(dk), nil
}

// GenerateSalt 生成盐值
func (h *HashUtils) GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// ValidationUtils 验证工具类
type ValidationUtils struct{}

// ValidateEmail 验证邮箱格式
func (v *ValidationUtils) ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePhone 验证手机号格式
func (v *ValidationUtils) ValidatePhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// ValidateURL 验证URL格式
func (v *ValidationUtils) ValidateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// ValidateIP 验证IP地址格式
func (v *ValidationUtils) ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ValidateIPv4 验证IPv4地址格式
func (v *ValidationUtils) ValidateIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// ValidateIPv6 验证IPv6地址格式
func (v *ValidationUtils) ValidateIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

// SanitizeInput 清理输入数据
func (v *ValidationUtils) SanitizeInput(input string) string {
	// 移除HTML标签
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	input = htmlRegex.ReplaceAllString(input, "")
	
	// 移除JavaScript代码
	jsRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = jsRegex.ReplaceAllString(input, "")
	
	// 移除危险字符
	dangerousChars := []string{"<", ">", "\"", "'", "&", "javascript:", "vbscript:", "onload=", "onerror="}
	for _, char := range dangerousChars {
		input = strings.ReplaceAll(input, char, "")
	}
	
	return strings.TrimSpace(input)
}

// DetectionUtils 检测工具类
type DetectionUtils struct{}

// DetectSQLInjection 检测SQL注入
func (d *DetectionUtils) DetectSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(select\s+.*\s+from)`,
		`(?i)(insert\s+into)`,
		`(?i)(delete\s+from)`,
		`(?i)(update\s+.*\s+set)`,
		`(?i)(drop\s+table)`,
		`(?i)(create\s+table)`,
		`(?i)(alter\s+table)`,
		`(?i)(\'\s*or\s*\'\s*=\s*\')`,
		`(?i)(\'\s*or\s*1\s*=\s*1)`,
		`(?i)(--\s*)`,
		`(?i)(/\*.*\*/)`,
	}
	
	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	
	return false
}

// DetectXSS 检测XSS攻击
func (d *DetectionUtils) DetectXSS(input string) bool {
	xssPatterns := []string{
		`(?i)<script[^>]*>`,
		`(?i)</script>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload\s*=`,
		`(?i)onerror\s*=`,
		`(?i)onclick\s*=`,
		`(?i)onmouseover\s*=`,
		`(?i)onfocus\s*=`,
		`(?i)onblur\s*=`,
		`(?i)eval\s*\(`,
		`(?i)document\.cookie`,
		`(?i)document\.write`,
		`(?i)window\.location`,
	}
	
	for _, pattern := range xssPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	
	return false
}

// DetectPathTraversal 检测路径遍历攻击
func (d *DetectionUtils) DetectPathTraversal(input string) bool {
	pathPatterns := []string{
		`\.\.\/`,
		`\.\.\\`,
		`%2e%2e%2f`,
		`%2e%2e%5c`,
		`%252e%252e%252f`,
		`%252e%252e%255c`,
	}
	
	for _, pattern := range pathPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	
	return false
}

// DetectCommandInjection 检测命令注入
func (d *DetectionUtils) DetectCommandInjection(input string) bool {
	cmdPatterns := []string{
		`(?i)(;|\||\&\&|\|\|)\s*(ls|dir|cat|type|echo|ping|wget|curl|nc|netcat)`,
		`(?i)\$\(.*\)`,
		`(?i)\`.*\``,
		`(?i)(rm\s+-rf)`,
		`(?i)(chmod\s+777)`,
		`(?i)(sudo\s+)`,
		`(?i)(passwd\s+)`,
	}
	
	for _, pattern := range cmdPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	
	return false
}

// TokenUtils 令牌工具类
type TokenUtils struct{}

// GenerateRandomToken 生成随机令牌
func (t *TokenUtils) GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateCSRFToken 生成CSRF令牌
func (t *TokenUtils) GenerateCSRFToken() (string, error) {
	return t.GenerateRandomToken(32)
}

// GenerateSessionID 生成会话ID
func (t *TokenUtils) GenerateSessionID() (string, error) {
	return t.GenerateRandomToken(32)
}

// SecurityUtils 安全工具集合
type SecurityUtils struct {
	Password   *PasswordUtils
	Encryption *EncryptionUtils
	Hash       *HashUtils
	Validation *ValidationUtils
	Detection  *DetectionUtils
	Token      *TokenUtils
}

// NewSecurityUtils 创建安全工具实例
func NewSecurityUtils() *SecurityUtils {
	return &SecurityUtils{
		Password:   &PasswordUtils{},
		Encryption: &EncryptionUtils{},
		Hash:       &HashUtils{},
		Validation: &ValidationUtils{},
		Detection:  &DetectionUtils{},
		Token:      &TokenUtils{},
	}
}

// TimeUtils 时间工具类
type TimeUtils struct{}

// IsExpired 检查时间是否过期
func (t *TimeUtils) IsExpired(timestamp time.Time, duration time.Duration) bool {
	return time.Since(timestamp) > duration
}

// GetExpirationTime 获取过期时间
func (t *TimeUtils) GetExpirationTime(duration time.Duration) time.Time {
	return time.Now().Add(duration)
}

// FormatSecurityTimestamp 格式化安全时间戳
func (t *TimeUtils) FormatSecurityTimestamp(t_time time.Time) string {
	return t_time.Format("2006-01-02 15:04:05 MST")
}

// ParseSecurityTimestamp 解析安全时间戳
func (t *TimeUtils) ParseSecurityTimestamp(timestamp string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05 MST", timestamp)
}

// NetworkUtils 网络工具类
type NetworkUtils struct{}

// IsPrivateIP 检查是否为私有IP
func (n *NetworkUtils) IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// 检查IPv4私有地址范围
	if parsedIP.To4() != nil {
		// 10.0.0.0/8
		if parsedIP[12] == 10 {
			return true
		}
		// 172.16.0.0/12
		if parsedIP[12] == 172 && parsedIP[13] >= 16 && parsedIP[13] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if parsedIP[12] == 192 && parsedIP[13] == 168 {
			return true
		}
		// 127.0.0.0/8 (localhost)
		if parsedIP[12] == 127 {
			return true
		}
	}
	
	return false
}

// IsLoopbackIP 检查是否为回环IP
func (n *NetworkUtils) IsLoopbackIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.IsLoopback()
}

// GetIPRange 获取IP范围
func (n *NetworkUtils) GetIPRange(cidr string) (net.IP, net.IP, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, err
	}
	
	// 计算网络地址和广播地址
	networkAddr := ipNet.IP
	broadcastAddr := make(net.IP, len(networkAddr))
	copy(broadcastAddr, networkAddr)
	
	for i := 0; i < len(ipNet.Mask); i++ {
		broadcastAddr[i] |= ^ipNet.Mask[i]
	}
	
	return networkAddr, broadcastAddr, nil
}

// IsIPInRange 检查IP是否在指定范围内
func (n *NetworkUtils) IsIPInRange(ip, cidr string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	
	return ipNet.Contains(parsedIP)
}