package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// NotificationService 
type NotificationService struct {
	// 
	emailService EmailService
	smsService   SMSService
	pushService  PushService
}

// EmailService 
type EmailService interface {
	SendEmail(to, subject, body string) error
}

// SMSService 
type SMSService interface {
	SendSMS(to, message string) error
}

// PushService ?
type PushService interface {
	SendPush(userID uuid.UUID, title, message string) error
}

// NewNotificationService 
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

// SendTaskNotification 
func (s *NotificationService) SendTaskNotification(ctx context.Context, req *domain.TaskNotificationRequest) error {
	// 
	title, message := s.buildTaskNotificationContent(req)

	// 
	return s.sendNotificationByPreference(ctx, req.UserID, title, message, req.Preferences)
}

// SendProjectNotification 
func (s *NotificationService) SendProjectNotification(ctx context.Context, req *domain.ProjectNotificationRequest) error {
	// 
	title, message := s.buildProjectNotificationContent(req)

	// 
	for _, userID := range req.UserIDs {
		err := s.sendNotificationByPreference(ctx, userID, title, message, req.Preferences)
		if err != nil {
			// 
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
		}
	}

	return nil
}

// SendTeamNotification 
func (s *NotificationService) SendTeamNotification(ctx context.Context, req *domain.TeamNotificationRequest) error {
	// 
	title, message := s.buildTeamNotificationContent(req)

	// 
	for _, userID := range req.UserIDs {
		err := s.sendNotificationByPreference(ctx, userID, title, message, req.Preferences)
		if err != nil {
			// 
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
		}
	}

	return nil
}

// UpdateNotificationPreferences 
func (s *NotificationService) UpdateNotificationPreferences(ctx context.Context, req *domain.NotificationPreferencesRequest) error {
	// 浽?
	// 
	fmt.Printf("Updated notification preferences for user %s\n", req.UserID)
	return nil
}

// ==========  ==========

// buildTaskNotificationContent 
func (s *NotificationService) buildTaskNotificationContent(req *domain.TaskNotificationRequest) (string, string) {
	var title, message string

	switch req.Type {
	case domain.TaskNotificationTypeAssigned:
		title = "?
		message = fmt.Sprintf("?s", req.TaskTitle)

	case domain.TaskNotificationTypeStatusChanged:
		title = "?
		message = fmt.Sprintf(" \"%s\" %s", req.TaskTitle, req.NewStatus)

	case domain.TaskNotificationTypeDueSoon:
		title = ""
		message = fmt.Sprintf(" \"%s\"  %s ", req.TaskTitle, req.DueDate.Format("2006-01-02 15:04"))

	case domain.TaskNotificationTypeOverdue:
		title = ""
		message = fmt.Sprintf(" \"%s\"  %s ", req.TaskTitle, req.DueDate.Format("2006-01-02 15:04"))

	case domain.TaskNotificationTypeCompleted:
		title = "?
		message = fmt.Sprintf(" \"%s\" ?, req.TaskTitle)

	case domain.TaskNotificationTypeCommentAdded:
		title = "?
		message = fmt.Sprintf(" \"%s\" ", req.TaskTitle)

	case domain.TaskNotificationTypeDependencyResolved:
		title = "?
		message = fmt.Sprintf(" \"%s\" ?, req.TaskTitle)

	default:
		title = ""
		message = fmt.Sprintf(" \"%s\" ?, req.TaskTitle)
	}

	return title, message
}

// buildProjectNotificationContent 
func (s *NotificationService) buildProjectNotificationContent(req *domain.ProjectNotificationRequest) (string, string) {
	var title, message string

	switch req.Type {
	case domain.ProjectNotificationTypeCreated:
		title = "?
		message = fmt.Sprintf("?\"%s\" ?, req.ProjectName)

	case domain.ProjectNotificationTypeStatusChanged:
		title = "?
		message = fmt.Sprintf(" \"%s\" %s", req.ProjectName, req.NewStatus)

	case domain.ProjectNotificationTypeMemberAdded:
		title = ""
		message = fmt.Sprintf(" \"%s\"", req.ProjectName)

	case domain.ProjectNotificationTypeMemberRemoved:
		title = ""
		message = fmt.Sprintf("?\"%s\" ?, req.ProjectName)

	case domain.ProjectNotificationTypeMilestoneReached:
		title = "?
		message = fmt.Sprintf(" \"%s\" %s", req.ProjectName, req.MilestoneName)

	case domain.ProjectNotificationTypeDeadlineApproaching:
		title = ""
		message = fmt.Sprintf(" \"%s\"  %s ", req.ProjectName, req.Deadline.Format("2006-01-02"))

	case domain.ProjectNotificationTypeCompleted:
		title = "?
		message = fmt.Sprintf(" \"%s\" ?, req.ProjectName)

	default:
		title = ""
		message = fmt.Sprintf(" \"%s\" ?, req.ProjectName)
	}

	return title, message
}

// buildTeamNotificationContent 
func (s *NotificationService) buildTeamNotificationContent(req *domain.TeamNotificationRequest) (string, string) {
	var title, message string

	switch req.Type {
	case domain.TeamNotificationTypeMemberJoined:
		title = ""
		message = fmt.Sprintf(" \"%s\"", req.TeamName)

	case domain.TeamNotificationTypeMemberLeft:
		title = ""
		message = fmt.Sprintf(" \"%s\"", req.TeamName)

	case domain.TeamNotificationTypeRoleChanged:
		title = ""
		message = fmt.Sprintf(" \"%s\" ", req.TeamName)

	case domain.TeamNotificationTypeMeeting:
		title = ""
		message = fmt.Sprintf(" \"%s\"  %s ", req.TeamName, req.MeetingTime.Format("2006-01-02 15:04"))

	case domain.TeamNotificationTypeAnnouncement:
		title = ""
		message = fmt.Sprintf(" \"%s\" ?s", req.TeamName, req.AnnouncementTitle)

	case domain.TeamNotificationTypePerformanceReport:
		title = ""
		message = fmt.Sprintf(" \"%s\" ", req.TeamName)

	default:
		title = ""
		message = fmt.Sprintf(" \"%s\" ?, req.TeamName)
	}

	return title, message
}

// sendNotificationByPreference 
func (s *NotificationService) sendNotificationByPreference(ctx context.Context, userID uuid.UUID, title, message string, preferences *domain.NotificationPreferences) error {
	var errors []error

	// ?
	if preferences == nil {
		preferences = &domain.NotificationPreferences{
			PushEnabled: true,
		}
	}

	// 
	if preferences.PushEnabled {
		if err := s.pushService.SendPush(userID, title, message); err != nil {
			errors = append(errors, fmt.Errorf("push notification failed: %w", err))
		}
	}

	// 
	if preferences.EmailEnabled && preferences.Email != "" {
		emailBody := s.buildEmailBody(title, message)
		if err := s.emailService.SendEmail(preferences.Email, title, emailBody); err != nil {
			errors = append(errors, fmt.Errorf("email notification failed: %w", err))
		}
	}

	// 
	if preferences.SMSEnabled && preferences.Phone != "" {
		smsMessage := s.buildSMSMessage(title, message)
		if err := s.smsService.SendSMS(preferences.Phone, smsMessage); err != nil {
			errors = append(errors, fmt.Errorf("SMS notification failed: %w", err))
		}
	}

	// 
	if len(errors) > 0 && len(errors) == s.countEnabledChannels(preferences) {
		return fmt.Errorf("all notification channels failed: %v", errors)
	}

	return nil
}

// buildEmailBody 
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
            ?br>
            %s
        </p>
    </div>
</body>
</html>
`, title, title, message, time.Now().Format("2006-01-02 15:04:05"))
}

// buildSMSMessage 
func (s *NotificationService) buildSMSMessage(title, message string) string {
	// ?
	return fmt.Sprintf("[] %s: %s", title, message)
}

// countEnabledChannels 
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

// ==========  ==========

// MockEmailService 
type MockEmailService struct{}

func (s *MockEmailService) SendEmail(to, subject, body string) error {
	fmt.Printf(" Email sent to %s: %s\n", to, subject)
	return nil
}

// MockSMSService 
type MockSMSService struct{}

func (s *MockSMSService) SendSMS(to, message string) error {
	fmt.Printf(" SMS sent to %s: %s\n", to, message)
	return nil
}

// MockPushService ?
type MockPushService struct{}

func (s *MockPushService) SendPush(userID uuid.UUID, title, message string) error {
	fmt.Printf(" Push notification sent to %s: %s - %s\n", userID, title, message)
	return nil
}

// NewMockNotificationService 
func NewMockNotificationService() domain.NotificationService {
	return NewNotificationService(
		&MockEmailService{},
		&MockSMSService{},
		&MockPushService{},
	)
}

