package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// NotificationService 通知服务实现
type NotificationService struct {
	// 这里可以注入邮件服务、短信服务、推送服务等
	emailService EmailService
	smsService   SMSService
	pushService  PushService
}

// EmailService 邮件服务接口
type EmailService interface {
	SendEmail(to, subject, body string) error
}

// SMSService 短信服务接口
type SMSService interface {
	SendSMS(to, message string) error
}

// PushService 推送服务接口
type PushService interface {
	SendPush(userID uuid.UUID, title, message string) error
}

// NewNotificationService 创建通知服务
func NewNotificationService(
	emailService EmailService,
	smsService SMSService,
	pushService PushService,
) domain.NotificationService {
	return &NotificationService{
		emailService: emailService,
		smsService:   smsService,
		pushService:  pushService,
	}
}

// SendTaskNotification 发送任务通知
func (s *NotificationService) SendTaskNotification(ctx context.Context, req *domain.TaskNotificationRequest) error {
	// 根据通知类型构建消息内容
	title, message := s.buildTaskNotificationContent(req)

	// 根据用户偏好发送通知
	return s.sendNotificationByPreference(ctx, req.UserID, title, message, req.Preferences)
}

// SendProjectNotification 发送项目通知
func (s *NotificationService) SendProjectNotification(ctx context.Context, req *domain.ProjectNotificationRequest) error {
	// 根据通知类型构建消息内容
	title, message := s.buildProjectNotificationContent(req)

	// 发送给项目相关用户
	for _, userID := range req.UserIDs {
		err := s.sendNotificationByPreference(ctx, userID, title, message, req.Preferences)
		if err != nil {
			// 记录错误但继续发送给其他用户
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
		}
	}

	return nil
}

// SendTeamNotification 发送团队通知
func (s *NotificationService) SendTeamNotification(ctx context.Context, req *domain.TeamNotificationRequest) error {
	// 根据通知类型构建消息内容
	title, message := s.buildTeamNotificationContent(req)

	// 发送给团队成员
	for _, userID := range req.UserIDs {
		err := s.sendNotificationByPreference(ctx, userID, title, message, req.Preferences)
		if err != nil {
			// 记录错误但继续发送给其他用户
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
		}
	}

	return nil
}

// UpdateNotificationPreferences 更新通知偏好
func (s *NotificationService) UpdateNotificationPreferences(ctx context.Context, req *domain.NotificationPreferencesRequest) error {
	// 这里应该将偏好设置保存到数据库
	// 简化处理，直接返回成功
	fmt.Printf("Updated notification preferences for user %s\n", req.UserID)
	return nil
}

// ========== 私有辅助方法 ==========

// buildTaskNotificationContent 构建任务通知内容
func (s *NotificationService) buildTaskNotificationContent(req *domain.TaskNotificationRequest) (string, string) {
	var title, message string

	switch req.Type {
	case domain.TaskNotificationTypeAssigned:
		title = "新任务分配"
		message = fmt.Sprintf("您被分配了新任务：%s", req.TaskTitle)

	case domain.TaskNotificationTypeStatusChanged:
		title = "任务状态更新"
		message = fmt.Sprintf("任务 \"%s\" 的状态已更新为：%s", req.TaskTitle, req.NewStatus)

	case domain.TaskNotificationTypeDueSoon:
		title = "任务即将到期"
		message = fmt.Sprintf("任务 \"%s\" 将在 %s 到期，请及时处理", req.TaskTitle, req.DueDate.Format("2006-01-02 15:04"))

	case domain.TaskNotificationTypeOverdue:
		title = "任务已逾期"
		message = fmt.Sprintf("任务 \"%s\" 已于 %s 逾期，请尽快处理", req.TaskTitle, req.DueDate.Format("2006-01-02 15:04"))

	case domain.TaskNotificationTypeCompleted:
		title = "任务已完成"
		message = fmt.Sprintf("任务 \"%s\" 已完成", req.TaskTitle)

	case domain.TaskNotificationTypeCommentAdded:
		title = "任务新评论"
		message = fmt.Sprintf("任务 \"%s\" 有新评论", req.TaskTitle)

	case domain.TaskNotificationTypeDependencyResolved:
		title = "任务依赖已解决"
		message = fmt.Sprintf("任务 \"%s\" 的依赖已解决，可以开始处理", req.TaskTitle)

	default:
		title = "任务通知"
		message = fmt.Sprintf("任务 \"%s\" 有更新", req.TaskTitle)
	}

	return title, message
}

// buildProjectNotificationContent 构建项目通知内容
func (s *NotificationService) buildProjectNotificationContent(req *domain.ProjectNotificationRequest) (string, string) {
	var title, message string

	switch req.Type {
	case domain.ProjectNotificationTypeCreated:
		title = "新项目创建"
		message = fmt.Sprintf("新项目 \"%s\" 已创建", req.ProjectName)

	case domain.ProjectNotificationTypeStatusChanged:
		title = "项目状态更新"
		message = fmt.Sprintf("项目 \"%s\" 的状态已更新为：%s", req.ProjectName, req.NewStatus)

	case domain.ProjectNotificationTypeMemberAdded:
		title = "项目成员变更"
		message = fmt.Sprintf("您已被添加到项目 \"%s\"", req.ProjectName)

	case domain.ProjectNotificationTypeMemberRemoved:
		title = "项目成员变更"
		message = fmt.Sprintf("您已从项目 \"%s\" 中移除", req.ProjectName)

	case domain.ProjectNotificationTypeMilestoneReached:
		title = "项目里程碑达成"
		message = fmt.Sprintf("项目 \"%s\" 达成里程碑：%s", req.ProjectName, req.MilestoneName)

	case domain.ProjectNotificationTypeDeadlineApproaching:
		title = "项目截止日期临近"
		message = fmt.Sprintf("项目 \"%s\" 将在 %s 截止", req.ProjectName, req.Deadline.Format("2006-01-02"))

	case domain.ProjectNotificationTypeCompleted:
		title = "项目已完成"
		message = fmt.Sprintf("项目 \"%s\" 已完成", req.ProjectName)

	default:
		title = "项目通知"
		message = fmt.Sprintf("项目 \"%s\" 有更新", req.ProjectName)
	}

	return title, message
}

// buildTeamNotificationContent 构建团队通知内容
func (s *NotificationService) buildTeamNotificationContent(req *domain.TeamNotificationRequest) (string, string) {
	var title, message string

	switch req.Type {
	case domain.TeamNotificationTypeMemberJoined:
		title = "团队成员加入"
		message = fmt.Sprintf("新成员已加入团队 \"%s\"", req.TeamName)

	case domain.TeamNotificationTypeMemberLeft:
		title = "团队成员离开"
		message = fmt.Sprintf("成员已离开团队 \"%s\"", req.TeamName)

	case domain.TeamNotificationTypeRoleChanged:
		title = "团队角色变更"
		message = fmt.Sprintf("您在团队 \"%s\" 的角色已更新", req.TeamName)

	case domain.TeamNotificationTypeMeeting:
		title = "团队会议通知"
		message = fmt.Sprintf("团队 \"%s\" 将在 %s 举行会议", req.TeamName, req.MeetingTime.Format("2006-01-02 15:04"))

	case domain.TeamNotificationTypeAnnouncement:
		title = "团队公告"
		message = fmt.Sprintf("团队 \"%s\" 发布了新公告：%s", req.TeamName, req.AnnouncementTitle)

	case domain.TeamNotificationTypePerformanceReport:
		title = "团队绩效报告"
		message = fmt.Sprintf("团队 \"%s\" 的绩效报告已生成", req.TeamName)

	default:
		title = "团队通知"
		message = fmt.Sprintf("团队 \"%s\" 有更新", req.TeamName)
	}

	return title, message
}

// sendNotificationByPreference 根据用户偏好发送通知
func (s *NotificationService) sendNotificationByPreference(ctx context.Context, userID uuid.UUID, title, message string, preferences *domain.NotificationPreferences) error {
	var errors []error

	// 如果没有偏好设置，使用默认设置（推送通知）
	if preferences == nil {
		preferences = &domain.NotificationPreferences{
			PushEnabled: true,
		}
	}

	// 发送推送通知
	if preferences.PushEnabled {
		if err := s.pushService.SendPush(userID, title, message); err != nil {
			errors = append(errors, fmt.Errorf("push notification failed: %w", err))
		}
	}

	// 发送邮件通知
	if preferences.EmailEnabled && preferences.Email != "" {
		emailBody := s.buildEmailBody(title, message)
		if err := s.emailService.SendEmail(preferences.Email, title, emailBody); err != nil {
			errors = append(errors, fmt.Errorf("email notification failed: %w", err))
		}
	}

	// 发送短信通知
	if preferences.SMSEnabled && preferences.Phone != "" {
		smsMessage := s.buildSMSMessage(title, message)
		if err := s.smsService.SendSMS(preferences.Phone, smsMessage); err != nil {
			errors = append(errors, fmt.Errorf("SMS notification failed: %w", err))
		}
	}

	// 如果所有通知方式都失败，返回错误
	if len(errors) > 0 && len(errors) == s.countEnabledChannels(preferences) {
		return fmt.Errorf("all notification channels failed: %v", errors)
	}

	return nil
}

// buildEmailBody 构建邮件正文
func (s *NotificationService) buildEmailBody(title, message string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
</head>
<body>
    <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
        <h2 style="color: #333;">%s</h2>
        <p style="color: #666; line-height: 1.6;">%s</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">
            此邮件由太上老君任务管理系统自动发送，请勿回复。<br>
            发送时间：%s
        </p>
    </div>
</body>
</html>
`, title, title, message, time.Now().Format("2006-01-02 15:04:05"))
}

// buildSMSMessage 构建短信内容
func (s *NotificationService) buildSMSMessage(title, message string) string {
	// 短信内容需要简洁
	return fmt.Sprintf("[太上老君] %s: %s", title, message)
}

// countEnabledChannels 计算启用的通知渠道数量
func (s *NotificationService) countEnabledChannels(preferences *domain.NotificationPreferences) int {
	count := 0
	if preferences.PushEnabled {
		count++
	}
	if preferences.EmailEnabled {
		count++
	}
	if preferences.SMSEnabled {
		count++
	}
	return count
}

// ========== 模拟服务实现 ==========

// MockEmailService 模拟邮件服务
type MockEmailService struct{}

func (s *MockEmailService) SendEmail(to, subject, body string) error {
	fmt.Printf("📧 Email sent to %s: %s\n", to, subject)
	return nil
}

// MockSMSService 模拟短信服务
type MockSMSService struct{}

func (s *MockSMSService) SendSMS(to, message string) error {
	fmt.Printf("📱 SMS sent to %s: %s\n", to, message)
	return nil
}

// MockPushService 模拟推送服务
type MockPushService struct{}

func (s *MockPushService) SendPush(userID uuid.UUID, title, message string) error {
	fmt.Printf("🔔 Push notification sent to %s: %s - %s\n", userID, title, message)
	return nil
}

// NewMockNotificationService 创建模拟通知服务
func NewMockNotificationService() domain.NotificationService {
	return NewNotificationService(
		&MockEmailService{},
		&MockSMSService{},
		&MockPushService{},
	)
}