package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"

	"go.uber.org/zap"

	"auth-system/internal/config"
	"auth-system/internal/models"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendVerificationEmail(ctx context.Context, user *models.User, token string) error
	SendPasswordResetEmail(ctx context.Context, user *models.User, token string) error
	SendWelcomeEmail(ctx context.Context, user *models.User) error
}

// emailService 邮件服务实现
type emailService struct {
	config    config.EmailConfig
	logger    *zap.Logger
	templates map[string]*template.Template
}

// NewEmailService 创建邮件服务
func NewEmailService(cfg config.EmailConfig, logger *zap.Logger) (EmailService, error) {
	service := &emailService{
		config:    cfg,
		logger:    logger,
		templates: make(map[string]*template.Template),
	}

	// 加载邮件模板
	if err := service.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return service, nil
}

// loadTemplates 加载邮件模板
func (s *emailService) loadTemplates() error {
	templateNames := []string{
		"verification",
		"password_reset",
		"welcome",
	}

	for _, name := range templateNames {
		// 使用默认模板，因为模板目录可能不存在
		s.logger.Info("Loading default email template", zap.String("template", name))
		tmpl := s.getDefaultTemplate(name)
		s.templates[name] = tmpl
	}

	return nil
}

// getDefaultTemplate 获取默认模板
func (s *emailService) getDefaultTemplate(name string) *template.Template {
	var templateContent string

	switch name {
	case "verification":
		templateContent = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>邮箱验证</title>
</head>
<body>
    <h2>欢迎注册太上老君系统！</h2>
    <p>亲爱的 {{.Username}}，</p>
    <p>感谢您注册我们的系统。请点击下面的链接验证您的邮箱地址：</p>
    <p><a href="{{.VerificationURL}}" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; display: inline-block;">验证邮箱</a></p>
    <p>或者复制以下链接到浏览器地址栏：</p>
    <p>{{.VerificationURL}}</p>
    <p>此链接将在24小时后过期。</p>
    <p>如果您没有注册我们的系统，请忽略此邮件。</p>
    <br>
    <p>太上老君系统团队</p>
</body>
</html>`
	case "password_reset":
		templateContent = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>密码重置</title>
</head>
<body>
    <h2>密码重置请求</h2>
    <p>亲爱的 {{.Username}}，</p>
    <p>我们收到了您的密码重置请求。请点击下面的链接重置您的密码：</p>
    <p><a href="{{.ResetURL}}" style="background-color: #f44336; color: white; padding: 14px 20px; text-decoration: none; display: inline-block;">重置密码</a></p>
    <p>或者复制以下链接到浏览器地址栏：</p>
    <p>{{.ResetURL}}</p>
    <p>此链接将在1小时后过期。</p>
    <p>如果您没有请求密码重置，请忽略此邮件。</p>
    <br>
    <p>太上老君系统团队</p>
</body>
</html>`
	case "welcome":
		templateContent = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>欢迎加入</title>
</head>
<body>
    <h2>欢迎加入太上老君系统！</h2>
    <p>亲爱的 {{.Username}}，</p>
    <p>恭喜您成功注册并验证了邮箱！</p>
    <p>您现在可以开始使用我们的系统了。</p>
    <p><a href="{{.LoginURL}}" style="background-color: #2196F3; color: white; padding: 14px 20px; text-decoration: none; display: inline-block;">立即登录</a></p>
    <p>如果您有任何问题，请随时联系我们的客服团队。</p>
    <br>
    <p>太上老君系统团队</p>
</body>
</html>`
	default:
		templateContent = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>系统通知</title>
</head>
<body>
    <h2>系统通知</h2>
    <p>{{.Content}}</p>
    <br>
    <p>太上老君系统团队</p>
</body>
</html>`
	}

	tmpl, _ := template.New(name).Parse(templateContent)
	return tmpl
}

// SendVerificationEmail 发送验证邮件
func (s *emailService) SendVerificationEmail(ctx context.Context, user *models.User, token string) error {
	subject := "验证您的邮箱地址"
	verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)

	data := map[string]interface{}{
		"Username":        user.Username,
		"Email":          user.Email,
		"VerificationURL": verificationURL,
		"Token":          token,
	}

	return s.sendEmail(ctx, user.Email, subject, "verification", data)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *emailService) SendPasswordResetEmail(ctx context.Context, user *models.User, token string) error {
	subject := "重置您的密码"
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)

	data := map[string]interface{}{
		"Username": user.Username,
		"Email":    user.Email,
		"ResetURL": resetURL,
		"Token":    token,
	}

	return s.sendEmail(ctx, user.Email, subject, "password_reset", data)
}

// SendWelcomeEmail 发送欢迎邮件
func (s *emailService) SendWelcomeEmail(ctx context.Context, user *models.User) error {
	subject := "欢迎加入太上老君系统"
	loginURL := "http://localhost:3000/login"

	data := map[string]interface{}{
		"Username": user.Username,
		"Email":    user.Email,
		"LoginURL": loginURL,
	}

	return s.sendEmail(ctx, user.Email, subject, "welcome", data)
}

// sendEmail 发送邮件
func (s *emailService) sendEmail(ctx context.Context, to, subject, templateName string, data map[string]interface{}) error {
	// 渲染模板
	tmpl, exists := s.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		s.logger.Error("Failed to execute email template",
			zap.String("template", templateName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 构建邮件内容
	message := s.buildMessage(to, subject, body.String())

	// 发送邮件
	if err := s.sendSMTP(message, []string{to}); err != nil {
		s.logger.Error("Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.String("template", templateName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("template", templateName),
	)

	return nil
}

// buildMessage 构建邮件消息
func (s *emailService) buildMessage(to, subject, body string) string {
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	return message
}

// sendSMTP 通过SMTP发送邮件
func (s *emailService) sendSMTP(message string, recipients []string) error {
	// 如果没有配置SMTP，则跳过发送（开发环境）
	if s.config.SMTPUsername == "" || s.config.SMTPPassword == "" {
		s.logger.Warn("SMTP credentials not configured, skipping email send")
		return nil
	}

	// 建立连接
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	
	var client *smtp.Client
	var err error

	// 普通连接
	client, err = smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP: %w", err)
	}
	defer client.Close()

	// STARTTLS
	if s.config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: s.config.SMTPHost,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// 认证
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// 设置发件人
	if err := client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// 设置收件人
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer writer.Close()

	if _, err := writer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}