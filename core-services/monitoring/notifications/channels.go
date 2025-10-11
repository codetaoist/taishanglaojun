package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// EmailChannel 邮件通知渠道
type EmailChannel struct {
	name     string
	config   *EmailChannelConfig
	client   *http.Client
}

// EmailChannelConfig 邮件渠道配置
type EmailChannelConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	From         string `json:"from"`
	To           []string `json:"to"`
	Subject      string `json:"subject"`
	Template     string `json:"template"`
	TLS          bool   `json:"tls"`
}

// NewEmailChannel 创建邮件渠道
func NewEmailChannel(name string, config map[string]interface{}) (*EmailChannel, error) {
	emailConfig := &EmailChannelConfig{}
	
	// 解析配置
	if host, ok := config["smtp_host"].(string); ok {
		emailConfig.SMTPHost = host
	}
	if port, ok := config["smtp_port"].(float64); ok {
		emailConfig.SMTPPort = int(port)
	}
	if username, ok := config["username"].(string); ok {
		emailConfig.Username = username
	}
	if password, ok := config["password"].(string); ok {
		emailConfig.Password = password
	}
	if from, ok := config["from"].(string); ok {
		emailConfig.From = from
	}
	if to, ok := config["to"].([]interface{}); ok {
		for _, t := range to {
			if email, ok := t.(string); ok {
				emailConfig.To = append(emailConfig.To, email)
			}
		}
	}
	if subject, ok := config["subject"].(string); ok {
		emailConfig.Subject = subject
	}
	if template, ok := config["template"].(string); ok {
		emailConfig.Template = template
	}
	if tls, ok := config["tls"].(bool); ok {
		emailConfig.TLS = tls
	}
	
	return &EmailChannel{
		name:   name,
		config: emailConfig,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetName 获取渠道名称
func (ec *EmailChannel) GetName() string {
	return ec.name
}

// GetType 获取渠道类型
func (ec *EmailChannel) GetType() string {
	return "email"
}

// Send 发送邮件通知
func (ec *EmailChannel) Send(ctx context.Context, notification *models.Notification) error {
	// 构建邮件内容
	subject := ec.config.Subject
	if subject == "" {
		subject = notification.Title
	}
	
	body := notification.Message
	if ec.config.Template != "" {
		// 这里可以使用模板引擎渲染邮件内容
		body = ec.config.Template
	}
	
	// 构建邮件
	msg := fmt.Sprintf("To: %s\r\n", strings.Join(ec.config.To, ","))
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += body
	
	// 发送邮�?
	auth := smtp.PlainAuth("", ec.config.Username, ec.config.Password, ec.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", ec.config.SMTPHost, ec.config.SMTPPort)
	
	return smtp.SendMail(addr, auth, ec.config.From, ec.config.To, []byte(msg))
}

// WebhookChannel Webhook通知渠道
type WebhookChannel struct {
	name   string
	config *WebhookChannelConfig
	client *http.Client
}

// WebhookChannelConfig Webhook渠道配置
type WebhookChannelConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Template    string            `json:"template"`
	Timeout     time.Duration     `json:"timeout"`
	MaxRetries  int               `json:"max_retries"`
}

// NewWebhookChannel 创建Webhook渠道
func NewWebhookChannel(name string, config map[string]interface{}) (*WebhookChannel, error) {
	webhookConfig := &WebhookChannelConfig{
		Method:  "POST",
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}
	
	// 解析配置
	if url, ok := config["url"].(string); ok {
		webhookConfig.URL = url
	}
	if method, ok := config["method"].(string); ok {
		webhookConfig.Method = method
	}
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if str, ok := v.(string); ok {
				webhookConfig.Headers[k] = str
			}
		}
	}
	if template, ok := config["template"].(string); ok {
		webhookConfig.Template = template
	}
	if timeout, ok := config["timeout"].(float64); ok {
		webhookConfig.Timeout = time.Duration(timeout) * time.Second
	}
	if maxRetries, ok := config["max_retries"].(float64); ok {
		webhookConfig.MaxRetries = int(maxRetries)
	}
	
	return &WebhookChannel{
		name:   name,
		config: webhookConfig,
		client: &http.Client{Timeout: webhookConfig.Timeout},
	}, nil
}

// GetName 获取渠道名称
func (wc *WebhookChannel) GetName() string {
	return wc.name
}

// GetType 获取渠道类型
func (wc *WebhookChannel) GetType() string {
	return "webhook"
}

// Send 发送Webhook通知
func (wc *WebhookChannel) Send(ctx context.Context, notification *models.Notification) error {
	// 构建请求�?
	var body []byte
	var err error
	
	if wc.config.Template != "" {
		// 使用模板
		body = []byte(wc.config.Template)
	} else {
		// 使用默认JSON格式
		body, err = json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("failed to marshal notification: %w", err)
		}
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, wc.config.Method, wc.config.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置头部
	req.Header.Set("Content-Type", "application/json")
	for k, v := range wc.config.Headers {
		req.Header.Set(k, v)
	}
	
	// 发送请�?
	resp, err := wc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状�?
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// SlackChannel Slack通知渠道
type SlackChannel struct {
	name   string
	config *SlackChannelConfig
	client *http.Client
}

// SlackChannelConfig Slack渠道配置
type SlackChannelConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
	IconURL    string `json:"icon_url"`
}

// SlackMessage Slack消息格式
type SlackMessage struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment Slack附件
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField Slack字段
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackChannel 创建Slack渠道
func NewSlackChannel(name string, config map[string]interface{}) (*SlackChannel, error) {
	slackConfig := &SlackChannelConfig{}
	
	// 解析配置
	if webhookURL, ok := config["webhook_url"].(string); ok {
		slackConfig.WebhookURL = webhookURL
	}
	if channel, ok := config["channel"].(string); ok {
		slackConfig.Channel = channel
	}
	if username, ok := config["username"].(string); ok {
		slackConfig.Username = username
	}
	if iconEmoji, ok := config["icon_emoji"].(string); ok {
		slackConfig.IconEmoji = iconEmoji
	}
	if iconURL, ok := config["icon_url"].(string); ok {
		slackConfig.IconURL = iconURL
	}
	
	return &SlackChannel{
		name:   name,
		config: slackConfig,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetName 获取渠道名称
func (sc *SlackChannel) GetName() string {
	return sc.name
}

// GetType 获取渠道类型
func (sc *SlackChannel) GetType() string {
	return "slack"
}

// Send 发送Slack通知
func (sc *SlackChannel) Send(ctx context.Context, notification *models.Notification) error {
	// 构建Slack消息
	message := SlackMessage{
		Channel:   sc.config.Channel,
		Username:  sc.config.Username,
		IconEmoji: sc.config.IconEmoji,
		IconURL:   sc.config.IconURL,
		Text:      notification.Title,
	}
	
	// 添加附件
	color := sc.getSeverityColor(notification.Severity)
	attachment := SlackAttachment{
		Color:     color,
		Title:     notification.Title,
		Text:      notification.Message,
		Timestamp: notification.CreatedAt.Unix(),
	}
	
	// 添加标签字段
	if len(notification.Labels) > 0 {
		for k, v := range notification.Labels {
			attachment.Fields = append(attachment.Fields, SlackField{
				Title: k,
				Value: v,
				Short: true,
			})
		}
	}
	
	message.Attachments = []SlackAttachment{attachment}
	
	// 序列化消�?
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}
	
	// 发送请�?
	req, err := http.NewRequestWithContext(ctx, "POST", sc.config.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// getSeverityColor 获取严重级别对应的颜�?
func (sc *SlackChannel) getSeverityColor(severity models.Severity) string {
	switch severity {
	case models.SeverityCritical:
		return "danger"
	case models.SeverityHigh:
		return "warning"
	case models.SeverityMedium:
		return "good"
	case models.SeverityLow:
		return "#439FE0"
	case models.SeverityInfo:
		return "#36a64f"
	default:
		return "good"
	}
}

// DingTalkChannel 钉钉通知渠道
type DingTalkChannel struct {
	name   string
	config *DingTalkChannelConfig
	client *http.Client
}

// DingTalkChannelConfig 钉钉渠道配置
type DingTalkChannelConfig struct {
	WebhookURL string   `json:"webhook_url"`
	Secret     string   `json:"secret"`
	AtMobiles  []string `json:"at_mobiles"`
	AtAll      bool     `json:"at_all"`
}

// DingTalkMessage 钉钉消息格式
type DingTalkMessage struct {
	MsgType  string                 `json:"msgtype"`
	Text     *DingTalkText          `json:"text,omitempty"`
	Markdown *DingTalkMarkdown      `json:"markdown,omitempty"`
	At       *DingTalkAt            `json:"at,omitempty"`
}

// DingTalkText 钉钉文本消息
type DingTalkText struct {
	Content string `json:"content"`
}

// DingTalkMarkdown 钉钉Markdown消息
type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingTalkAt 钉钉@配置
type DingTalkAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// NewDingTalkChannel 创建钉钉渠道
func NewDingTalkChannel(name string, config map[string]interface{}) (*DingTalkChannel, error) {
	dingTalkConfig := &DingTalkChannelConfig{}
	
	// 解析配置
	if webhookURL, ok := config["webhook_url"].(string); ok {
		dingTalkConfig.WebhookURL = webhookURL
	}
	if secret, ok := config["secret"].(string); ok {
		dingTalkConfig.Secret = secret
	}
	if atMobiles, ok := config["at_mobiles"].([]interface{}); ok {
		for _, mobile := range atMobiles {
			if m, ok := mobile.(string); ok {
				dingTalkConfig.AtMobiles = append(dingTalkConfig.AtMobiles, m)
			}
		}
	}
	if atAll, ok := config["at_all"].(bool); ok {
		dingTalkConfig.AtAll = atAll
	}
	
	return &DingTalkChannel{
		name:   name,
		config: dingTalkConfig,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetName 获取渠道名称
func (dtc *DingTalkChannel) GetName() string {
	return dtc.name
}

// GetType 获取渠道类型
func (dtc *DingTalkChannel) GetType() string {
	return "dingtalk"
}

// Send 发送钉钉通知
func (dtc *DingTalkChannel) Send(ctx context.Context, notification *models.Notification) error {
	// 构建钉钉消息
	message := DingTalkMessage{
		MsgType: "markdown",
		Markdown: &DingTalkMarkdown{
			Title: notification.Title,
			Text:  dtc.buildMarkdownText(notification),
		},
		At: &DingTalkAt{
			AtMobiles: dtc.config.AtMobiles,
			IsAtAll:   dtc.config.AtAll,
		},
	}
	
	// 序列化消�?
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal dingtalk message: %w", err)
	}
	
	// 构建URL（包含签名）
	webhookURL := dtc.buildWebhookURL()
	
	// 发送请�?
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := dtc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dingtalk returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// buildMarkdownText 构建Markdown文本
func (dtc *DingTalkChannel) buildMarkdownText(notification *models.Notification) string {
	text := fmt.Sprintf("## %s\n\n", notification.Title)
	text += fmt.Sprintf("**消息**: %s\n\n", notification.Message)
	text += fmt.Sprintf("**严重级别**: %s\n\n", notification.Severity)
	text += fmt.Sprintf("**时间**: %s\n\n", notification.CreatedAt.Format("2006-01-02 15:04:05"))
	
	if len(notification.Labels) > 0 {
		text += "**标签**:\n\n"
		for k, v := range notification.Labels {
			text += fmt.Sprintf("- %s: %s\n", k, v)
		}
	}
	
	return text
}

// buildWebhookURL 构建带签名的Webhook URL
func (dtc *DingTalkChannel) buildWebhookURL() string {
	if dtc.config.Secret == "" {
		return dtc.config.WebhookURL
	}
	
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, dtc.config.Secret)
	
	h := hmac.New(sha256.New, []byte(dtc.config.Secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	return fmt.Sprintf("%s&timestamp=%d&sign=%s", dtc.config.WebhookURL, timestamp, url.QueryEscape(signature))
}

// WeChatChannel 微信通知渠道
type WeChatChannel struct {
	name   string
	config *WeChatChannelConfig
	client *http.Client
}

// WeChatChannelConfig 微信渠道配置
type WeChatChannelConfig struct {
	WebhookURL string   `json:"webhook_url"`
	MsgType    string   `json:"msg_type"`
	AtUsers    []string `json:"at_users"`
	AtAll      bool     `json:"at_all"`
}

// WeChatMessage 微信消息格式
type WeChatMessage struct {
	MsgType  string                `json:"msgtype"`
	Text     *WeChatText           `json:"text,omitempty"`
	Markdown *WeChatMarkdown       `json:"markdown,omitempty"`
}

// WeChatText 微信文本消息
type WeChatText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// WeChatMarkdown 微信Markdown消息
type WeChatMarkdown struct {
	Content string `json:"content"`
}

// NewWeChatChannel 创建微信渠道
func NewWeChatChannel(name string, config map[string]interface{}) (*WeChatChannel, error) {
	weChatConfig := &WeChatChannelConfig{
		MsgType: "text",
	}
	
	// 解析配置
	if webhookURL, ok := config["webhook_url"].(string); ok {
		weChatConfig.WebhookURL = webhookURL
	}
	if msgType, ok := config["msg_type"].(string); ok {
		weChatConfig.MsgType = msgType
	}
	if atUsers, ok := config["at_users"].([]interface{}); ok {
		for _, user := range atUsers {
			if u, ok := user.(string); ok {
				weChatConfig.AtUsers = append(weChatConfig.AtUsers, u)
			}
		}
	}
	if atAll, ok := config["at_all"].(bool); ok {
		weChatConfig.AtAll = atAll
	}
	
	return &WeChatChannel{
		name:   name,
		config: weChatConfig,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetName 获取渠道名称
func (wcc *WeChatChannel) GetName() string {
	return wcc.name
}

// GetType 获取渠道类型
func (wcc *WeChatChannel) GetType() string {
	return "wechat"
}

// Send 发送微信通知
func (wcc *WeChatChannel) Send(ctx context.Context, notification *models.Notification) error {
	var message WeChatMessage
	
	content := fmt.Sprintf("%s\n%s", notification.Title, notification.Message)
	
	switch wcc.config.MsgType {
	case "markdown":
		message = WeChatMessage{
			MsgType: "markdown",
			Markdown: &WeChatMarkdown{
				Content: wcc.buildMarkdownContent(notification),
			},
		}
	default:
		mentionedList := wcc.config.AtUsers
		if wcc.config.AtAll {
			mentionedList = append(mentionedList, "@all")
		}
		
		message = WeChatMessage{
			MsgType: "text",
			Text: &WeChatText{
				Content:       content,
				MentionedList: mentionedList,
			},
		}
	}
	
	// 序列化消�?
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal wechat message: %w", err)
	}
	
	// 发送请�?
	req, err := http.NewRequestWithContext(ctx, "POST", wcc.config.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := wcc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wechat returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// buildMarkdownContent 构建Markdown内容
func (wcc *WeChatChannel) buildMarkdownContent(notification *models.Notification) string {
	content := fmt.Sprintf("## %s\n", notification.Title)
	content += fmt.Sprintf("**消息**: %s\n", notification.Message)
	content += fmt.Sprintf("**严重级别**: %s\n", notification.Severity)
	content += fmt.Sprintf("**时间**: %s\n", notification.CreatedAt.Format("2006-01-02 15:04:05"))
	
	if len(notification.Labels) > 0 {
		content += "**标签**:\n"
		for k, v := range notification.Labels {
			content += fmt.Sprintf("- %s: %s\n", k, v)
		}
	}
	
	return content
}

// SMSChannel 短信通知渠道
type SMSChannel struct {
	name   string
	config *SMSChannelConfig
	client *http.Client
}

// SMSChannelConfig 短信渠道配置
type SMSChannelConfig struct {
	Provider     string   `json:"provider"`
	AccessKey    string   `json:"access_key"`
	SecretKey    string   `json:"secret_key"`
	SignName     string   `json:"sign_name"`
	TemplateCode string   `json:"template_code"`
	PhoneNumbers []string `json:"phone_numbers"`
}

// NewSMSChannel 创建短信渠道
func NewSMSChannel(name string, config map[string]interface{}) (*SMSChannel, error) {
	smsConfig := &SMSChannelConfig{}
	
	// 解析配置
	if provider, ok := config["provider"].(string); ok {
		smsConfig.Provider = provider
	}
	if accessKey, ok := config["access_key"].(string); ok {
		smsConfig.AccessKey = accessKey
	}
	if secretKey, ok := config["secret_key"].(string); ok {
		smsConfig.SecretKey = secretKey
	}
	if signName, ok := config["sign_name"].(string); ok {
		smsConfig.SignName = signName
	}
	if templateCode, ok := config["template_code"].(string); ok {
		smsConfig.TemplateCode = templateCode
	}
	if phoneNumbers, ok := config["phone_numbers"].([]interface{}); ok {
		for _, phone := range phoneNumbers {
			if p, ok := phone.(string); ok {
				smsConfig.PhoneNumbers = append(smsConfig.PhoneNumbers, p)
			}
		}
	}
	
	return &SMSChannel{
		name:   name,
		config: smsConfig,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetName 获取渠道名称
func (sc *SMSChannel) GetName() string {
	return sc.name
}

// GetType 获取渠道类型
func (sc *SMSChannel) GetType() string {
	return "sms"
}

// Send 发送短信通知
func (sc *SMSChannel) Send(ctx context.Context, notification *models.Notification) error {
	// 这里需要根据不同的短信服务提供商实现具体的发送逻辑
	// 例如阿里云短信、腾讯云短信�?
	
	switch sc.config.Provider {
	case "aliyun":
		return sc.sendAliyunSMS(ctx, notification)
	case "tencent":
		return sc.sendTencentSMS(ctx, notification)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", sc.config.Provider)
	}
}

// sendAliyunSMS 发送阿里云短信
func (sc *SMSChannel) sendAliyunSMS(ctx context.Context, notification *models.Notification) error {
	// 这里是阿里云短信的实现示�?
	// 实际实现需要使用阿里云SDK
	
	// 构建短信内容
	content := fmt.Sprintf("告警: %s, 详情: %s", notification.Title, notification.Message)
	
	// 发送到所有手机号
	for _, phone := range sc.config.PhoneNumbers {
		// 这里应该调用阿里云短信API
		fmt.Printf("Sending SMS to %s: %s\n", phone, content)
	}
	
	return nil
}

// sendTencentSMS 发送腾讯云短信
func (sc *SMSChannel) sendTencentSMS(ctx context.Context, notification *models.Notification) error {
	// 这里是腾讯云短信的实现示�?
	// 实际实现需要使用腾讯云SDK
	
	// 构建短信内容
	content := fmt.Sprintf("告警: %s, 详情: %s", notification.Title, notification.Message)
	
	// 发送到所有手机号
	for _, phone := range sc.config.PhoneNumbers {
		// 这里应该调用腾讯云短信API
		fmt.Printf("Sending SMS to %s: %s\n", phone, content)
	}
	
	return nil
}
