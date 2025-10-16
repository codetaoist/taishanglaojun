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

// PasswordUtils ?
type PasswordUtils struct{}

// HashPassword bcrypt
func (p *PasswordUtils) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 
func (p *PasswordUtils) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword 
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

// ValidatePasswordStrength 
func (p *PasswordUtils) ValidatePasswordStrength(password string) (bool, []string) {
	var errors []string
	
	if len(password) < 8 {
		errors = append(errors, "8?)
	}
	
	if len(password) > 128 {
		errors = append(errors, "128?)
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		errors = append(errors, "")
	}
	
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		errors = append(errors, "")
	}
	
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		errors = append(errors, "")
	}
	
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		errors = append(errors, "")
	}
	
	return len(errors) == 0, errors
}

// EncryptionUtils ?
type EncryptionUtils struct{}

// AESEncrypt AES
func (e *EncryptionUtils) AESEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES
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

// GenerateAESKey AES
func (e *EncryptionUtils) GenerateAESKey(keySize int) (string, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// HashUtils ?
type HashUtils struct{}

// MD5Hash MD5
func (h *HashUtils) MD5Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash SHA256
func (h *HashUtils) SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA512Hash SHA512
func (h *HashUtils) SHA512Hash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ScryptHash Scrypt
func (h *HashUtils) ScryptHash(password, salt string) (string, error) {
	dk, err := scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(dk), nil
}

// GenerateSalt ?
func (h *HashUtils) GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// ValidationUtils ?
type ValidationUtils struct{}

// ValidateEmail 
func (v *ValidationUtils) ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePhone ?
func (v *ValidationUtils) ValidatePhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// ValidateURL URL
func (v *ValidationUtils) ValidateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// ValidateIP IP
func (v *ValidationUtils) ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ValidateIPv4 IPv4
func (v *ValidationUtils) ValidateIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// ValidateIPv6 IPv6
func (v *ValidationUtils) ValidateIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

// SanitizeInput 
func (v *ValidationUtils) SanitizeInput(input string) string {
	// HTML
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	input = htmlRegex.ReplaceAllString(input, "")
	
	// JavaScript
	jsRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = jsRegex.ReplaceAllString(input, "")
	
	// 
	dangerousChars := []string{"<", ">", "\"", "'", "&", "javascript:", "vbscript:", "onload=", "onerror="}
	for _, char := range dangerousChars {
		input = strings.ReplaceAll(input, char, "")
	}
	
	return strings.TrimSpace(input)
}

// DetectionUtils 
type DetectionUtils struct{}

// DetectSQLInjection SQL
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

// DetectXSS XSS
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

// DetectPathTraversal ?
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

// DetectCommandInjection ?
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

// TokenUtils ?
type TokenUtils struct{}

// GenerateRandomToken 
func (t *TokenUtils) GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateCSRFToken CSRF
func (t *TokenUtils) GenerateCSRFToken() (string, error) {
	return t.GenerateRandomToken(32)
}

// GenerateSessionID ID
func (t *TokenUtils) GenerateSessionID() (string, error) {
	return t.GenerateRandomToken(32)
}

// SecurityUtils 
type SecurityUtils struct {
	Password   *PasswordUtils
	Encryption *EncryptionUtils
	Hash       *HashUtils
	Validation *ValidationUtils
	Detection  *DetectionUtils
	Token      *TokenUtils
}

// NewSecurityUtils 
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

// TimeUtils 乤?
type TimeUtils struct{}

// IsExpired ?
func (t *TimeUtils) IsExpired(timestamp time.Time, duration time.Duration) bool {
	return time.Since(timestamp) > duration
}

// GetExpirationTime 
func (t *TimeUtils) GetExpirationTime(duration time.Duration) time.Time {
	return time.Now().Add(duration)
}

// FormatSecurityTimestamp 
func (t *TimeUtils) FormatSecurityTimestamp(t_time time.Time) string {
	return t_time.Format("2006-01-02 15:04:05 MST")
}

// ParseSecurityTimestamp ?
func (t *TimeUtils) ParseSecurityTimestamp(timestamp string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05 MST", timestamp)
}

// NetworkUtils 繤?
type NetworkUtils struct{}

// IsPrivateIP IP
func (n *NetworkUtils) IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// IPv4
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

// IsLoopbackIP IP
func (n *NetworkUtils) IsLoopbackIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.IsLoopback()
}

// GetIPRange IP
func (n *NetworkUtils) GetIPRange(cidr string) (net.IP, net.IP, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, err
	}
	
	// 㲥
	networkAddr := ipNet.IP
	broadcastAddr := make(net.IP, len(networkAddr))
	copy(broadcastAddr, networkAddr)
	
	for i := 0; i < len(ipNet.Mask); i++ {
		broadcastAddr[i] |= ^ipNet.Mask[i]
	}
	
	return networkAddr, broadcastAddr, nil
}

// IsIPInRange IP
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

