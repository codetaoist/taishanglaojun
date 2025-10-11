// 太上老君AI平台GDPR合规性支持模�?package compliance

import (
	"context"
	"fmt"
	"time"
)

// GDPRCompliance GDPR合规性管理器
type GDPRCompliance struct {
	dataProcessor    DataProcessor
	consentManager   ConsentManager
	auditLogger      AuditLogger
	dataRetention    DataRetentionManager
	dataPortability  DataPortabilityManager
	privacyOfficer   PrivacyOfficer
}

// DataProcessor 数据处理器接�?type DataProcessor interface {
	ProcessPersonalData(ctx context.Context, data PersonalData, purpose string) error
	DeletePersonalData(ctx context.Context, userID string) error
	AnonymizePersonalData(ctx context.Context, userID string) error
	ExportPersonalData(ctx context.Context, userID string) (PersonalDataExport, error)
}

// ConsentManager 同意管理器接�?type ConsentManager interface {
	RecordConsent(ctx context.Context, consent UserConsent) error
	WithdrawConsent(ctx context.Context, userID string, purpose string) error
	GetConsent(ctx context.Context, userID string) ([]UserConsent, error)
	IsConsentValid(ctx context.Context, userID string, purpose string) (bool, error)
}

// AuditLogger 审计日志接口
type AuditLogger interface {
	LogDataAccess(ctx context.Context, log DataAccessLog) error
	LogDataModification(ctx context.Context, log DataModificationLog) error
	LogConsentChange(ctx context.Context, log ConsentChangeLog) error
	GetAuditTrail(ctx context.Context, userID string) ([]AuditLog, error)
}

// DataRetentionManager 数据保留管理器接�?type DataRetentionManager interface {
	SetRetentionPolicy(ctx context.Context, policy RetentionPolicy) error
	GetRetentionPolicy(ctx context.Context, dataType string) (RetentionPolicy, error)
	ScheduleDataDeletion(ctx context.Context, userID string) error
	ExecuteDataDeletion(ctx context.Context) error
}

// DataPortabilityManager 数据可移植性管理器接口
type DataPortabilityManager interface {
	ExportUserData(ctx context.Context, userID string, format string) (DataExport, error)
	ImportUserData(ctx context.Context, userID string, data DataImport) error
	ValidateDataFormat(ctx context.Context, format string) error
}

// PrivacyOfficer 隐私官接�?type PrivacyOfficer interface {
	HandleDataSubjectRequest(ctx context.Context, request DataSubjectRequest) error
	GeneratePrivacyReport(ctx context.Context, period TimePeriod) (PrivacyReport, error)
	ConductPrivacyImpactAssessment(ctx context.Context, assessment PIARequest) (PIAResult, error)
}

// PersonalData 个人数据结构
type PersonalData struct {
	UserID       string                 `json:"user_id"`
	DataType     string                 `json:"data_type"`
	Data         map[string]interface{} `json:"data"`
	Source       string                 `json:"source"`
	Purpose      []string               `json:"purpose"`
	LegalBasis   string                 `json:"legal_basis"`
	Sensitivity  string                 `json:"sensitivity"` // normal, sensitive, special
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	Encrypted    bool                   `json:"encrypted"`
	Pseudonymized bool                  `json:"pseudonymized"`
}

// UserConsent 用户同意记录
type UserConsent struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Purpose         string    `json:"purpose"`
	ConsentGiven    bool      `json:"consent_given"`
	ConsentMethod   string    `json:"consent_method"` // explicit, implicit, opt_in, opt_out
	ConsentText     string    `json:"consent_text"`
	ConsentVersion  string    `json:"consent_version"`
	ConsentDate     time.Time `json:"consent_date"`
	WithdrawalDate  *time.Time `json:"withdrawal_date,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	ConsentProof    string    `json:"consent_proof"`
	IsActive        bool      `json:"is_active"`
	LegalBasis      string    `json:"legal_basis"`
	ProcessingScope []string  `json:"processing_scope"`
}

// DataAccessLog 数据访问日志
type DataAccessLog struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	AccessedBy   string    `json:"accessed_by"`
	DataType     string    `json:"data_type"`
	AccessType   string    `json:"access_type"` // read, write, delete, export
	Purpose      string    `json:"purpose"`
	LegalBasis   string    `json:"legal_basis"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Timestamp    time.Time `json:"timestamp"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
	DataFields   []string  `json:"data_fields"`
}

// DataModificationLog 数据修改日志
type DataModificationLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	ModifiedBy   string                 `json:"modified_by"`
	DataType     string                 `json:"data_type"`
	Operation    string                 `json:"operation"` // create, update, delete, anonymize
	OldValues    map[string]interface{} `json:"old_values,omitempty"`
	NewValues    map[string]interface{} `json:"new_values,omitempty"`
	Reason       string                 `json:"reason"`
	LegalBasis   string                 `json:"legal_basis"`
	Timestamp    time.Time              `json:"timestamp"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
}

// ConsentChangeLog 同意变更日志
type ConsentChangeLog struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Purpose        string    `json:"purpose"`
	OldConsent     bool      `json:"old_consent"`
	NewConsent     bool      `json:"new_consent"`
	ChangeReason   string    `json:"change_reason"`
	ChangeMethod   string    `json:"change_method"`
	Timestamp      time.Time `json:"timestamp"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	ConsentVersion string    `json:"consent_version"`
}

// AuditLog 审计日志基础结构
type AuditLog struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // access, modification, consent
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// RetentionPolicy 数据保留策略
type RetentionPolicy struct {
	ID                string        `json:"id"`
	DataType          string        `json:"data_type"`
	RetentionPeriod   time.Duration `json:"retention_period"`
	DeletionMethod    string        `json:"deletion_method"` // hard_delete, soft_delete, anonymize
	LegalBasis        string        `json:"legal_basis"`
	Exceptions        []string      `json:"exceptions"`
	AutoDeletion      bool          `json:"auto_deletion"`
	NotificationDays  int           `json:"notification_days"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	IsActive          bool          `json:"is_active"`
}

// PersonalDataExport 个人数据导出结构
type PersonalDataExport struct {
	UserID      string                 `json:"user_id"`
	ExportDate  time.Time              `json:"export_date"`
	Format      string                 `json:"format"`
	Data        map[string]interface{} `json:"data"`
	Metadata    ExportMetadata         `json:"metadata"`
	Checksum    string                 `json:"checksum"`
	DownloadURL string                 `json:"download_url"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

// ExportMetadata 导出元数�?type ExportMetadata struct {
	TotalRecords    int       `json:"total_records"`
	DataTypes       []string  `json:"data_types"`
	ExportVersion   string    `json:"export_version"`
	IncludesDeleted bool      `json:"includes_deleted"`
	DateRange       DateRange `json:"date_range"`
}

// DateRange 日期范围
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// DataExport 数据导出结构
type DataExport struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Format      string    `json:"format"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	DownloadURL string    `json:"download_url"`
}

// DataImport 数据导入结构
type DataImport struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Format     string                 `json:"format"`
	Data       map[string]interface{} `json:"data"`
	Validation ImportValidation       `json:"validation"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	ImportedAt *time.Time             `json:"imported_at,omitempty"`
}

// ImportValidation 导入验证
type ImportValidation struct {
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	RecordCount  int      `json:"record_count"`
	ValidRecords int      `json:"valid_records"`
}

// DataSubjectRequest 数据主体请求
type DataSubjectRequest struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	RequestType  string                 `json:"request_type"` // access, rectification, erasure, portability, restriction, objection
	Description  string                 `json:"description"`
	RequestData  map[string]interface{} `json:"request_data"`
	Status       string                 `json:"status"` // pending, in_progress, completed, rejected
	Priority     string                 `json:"priority"`
	CreatedAt    time.Time              `json:"created_at"`
	DueDate      time.Time              `json:"due_date"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	AssignedTo   string                 `json:"assigned_to"`
	Response     string                 `json:"response"`
	Documents    []string               `json:"documents"`
	ContactInfo  ContactInfo            `json:"contact_info"`
}

// ContactInfo 联系信息
type ContactInfo struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	PreferredMethod string `json:"preferred_method"`
}

// PrivacyReport 隐私报告
type PrivacyReport struct {
	ID                    string                 `json:"id"`
	Period                TimePeriod             `json:"period"`
	GeneratedAt           time.Time              `json:"generated_at"`
	DataSubjectRequests   RequestStatistics      `json:"data_subject_requests"`
	DataBreaches          []DataBreach           `json:"data_breaches"`
	ConsentStatistics     ConsentStatistics      `json:"consent_statistics"`
	DataProcessingActivities []ProcessingActivity `json:"data_processing_activities"`
	ComplianceScore       float64                `json:"compliance_score"`
	Recommendations       []string               `json:"recommendations"`
	RiskAssessment        RiskAssessment         `json:"risk_assessment"`
}

// TimePeriod 时间周期
type TimePeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// RequestStatistics 请求统计
type RequestStatistics struct {
	TotalRequests     int                    `json:"total_requests"`
	CompletedRequests int                    `json:"completed_requests"`
	PendingRequests   int                    `json:"pending_requests"`
	RequestsByType    map[string]int         `json:"requests_by_type"`
	AverageResponseTime time.Duration        `json:"average_response_time"`
	ComplianceRate    float64                `json:"compliance_rate"`
}

// DataBreach 数据泄露记录
type DataBreach struct {
	ID              string    `json:"id"`
	IncidentDate    time.Time `json:"incident_date"`
	DetectedDate    time.Time `json:"detected_date"`
	ReportedDate    *time.Time `json:"reported_date,omitempty"`
	Severity        string    `json:"severity"` // low, medium, high, critical
	AffectedUsers   int       `json:"affected_users"`
	DataTypes       []string  `json:"data_types"`
	Cause           string    `json:"cause"`
	Description     string    `json:"description"`
	ContainmentActions []string `json:"containment_actions"`
	NotificationRequired bool   `json:"notification_required"`
	AuthorityNotified   bool    `json:"authority_notified"`
	Status          string    `json:"status"` // investigating, contained, resolved
}

// ConsentStatistics 同意统计
type ConsentStatistics struct {
	TotalUsers        int                `json:"total_users"`
	ConsentedUsers    int                `json:"consented_users"`
	WithdrawnConsents int                `json:"withdrawn_consents"`
	ConsentRate       float64            `json:"consent_rate"`
	ConsentByPurpose  map[string]int     `json:"consent_by_purpose"`
	ConsentTrends     []ConsentTrend     `json:"consent_trends"`
}

// ConsentTrend 同意趋势
type ConsentTrend struct {
	Date        time.Time `json:"date"`
	NewConsents int       `json:"new_consents"`
	Withdrawals int       `json:"withdrawals"`
	NetChange   int       `json:"net_change"`
}

// ProcessingActivity 处理活动
type ProcessingActivity struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Purpose         string   `json:"purpose"`
	LegalBasis      string   `json:"legal_basis"`
	DataTypes       []string `json:"data_types"`
	DataSubjects    []string `json:"data_subjects"`
	Recipients      []string `json:"recipients"`
	Transfers       []string `json:"transfers"`
	RetentionPeriod string   `json:"retention_period"`
	SecurityMeasures []string `json:"security_measures"`
	LastReviewed    time.Time `json:"last_reviewed"`
}

// RiskAssessment 风险评估
type RiskAssessment struct {
	OverallRisk     string              `json:"overall_risk"` // low, medium, high
	RiskFactors     []RiskFactor        `json:"risk_factors"`
	Mitigations     []string            `json:"mitigations"`
	LastAssessment  time.Time           `json:"last_assessment"`
	NextAssessment  time.Time           `json:"next_assessment"`
}

// RiskFactor 风险因素
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Impact      string  `json:"impact"`
	Likelihood  string  `json:"likelihood"`
	RiskLevel   string  `json:"risk_level"`
	Mitigation  string  `json:"mitigation"`
}

// PIARequest 隐私影响评估请求
type PIARequest struct {
	ID                string                 `json:"id"`
	ProjectName       string                 `json:"project_name"`
	Description       string                 `json:"description"`
	DataTypes         []string               `json:"data_types"`
	ProcessingPurpose string                 `json:"processing_purpose"`
	LegalBasis        string                 `json:"legal_basis"`
	DataSources       []string               `json:"data_sources"`
	Recipients        []string               `json:"recipients"`
	Transfers         []string               `json:"transfers"`
	SecurityMeasures  []string               `json:"security_measures"`
	RequestedBy       string                 `json:"requested_by"`
	RequestDate       time.Time              `json:"request_date"`
	Urgency           string                 `json:"urgency"`
	AdditionalInfo    map[string]interface{} `json:"additional_info"`
}

// PIAResult 隐私影响评估结果
type PIAResult struct {
	ID              string        `json:"id"`
	RequestID       string        `json:"request_id"`
	AssessmentDate  time.Time     `json:"assessment_date"`
	Assessor        string        `json:"assessor"`
	RiskLevel       string        `json:"risk_level"`
	Findings        []PIAFinding  `json:"findings"`
	Recommendations []string      `json:"recommendations"`
	Approval        PIAApproval   `json:"approval"`
	ReviewDate      time.Time     `json:"review_date"`
	Status          string        `json:"status"`
}

// PIAFinding 隐私影响评估发现
type PIAFinding struct {
	Category    string `json:"category"`
	Risk        string `json:"risk"`
	Impact      string `json:"impact"`
	Likelihood  string `json:"likelihood"`
	Mitigation  string `json:"mitigation"`
	Responsible string `json:"responsible"`
	Deadline    time.Time `json:"deadline"`
}

// PIAApproval 隐私影响评估批准
type PIAApproval struct {
	Approved    bool      `json:"approved"`
	ApprovedBy  string    `json:"approved_by"`
	ApprovalDate time.Time `json:"approval_date"`
	Conditions  []string  `json:"conditions"`
	Comments    string    `json:"comments"`
}

// NewGDPRCompliance 创建新的GDPR合规性管理器
func NewGDPRCompliance(
	dataProcessor DataProcessor,
	consentManager ConsentManager,
	auditLogger AuditLogger,
	dataRetention DataRetentionManager,
	dataPortability DataPortabilityManager,
	privacyOfficer PrivacyOfficer,
) *GDPRCompliance {
	return &GDPRCompliance{
		dataProcessor:    dataProcessor,
		consentManager:   consentManager,
		auditLogger:      auditLogger,
		dataRetention:    dataRetention,
		dataPortability:  dataPortability,
		privacyOfficer:   privacyOfficer,
	}
}

// ProcessPersonalDataWithConsent 在获得同意的情况下处理个人数�?func (g *GDPRCompliance) ProcessPersonalDataWithConsent(ctx context.Context, data PersonalData, purpose string) error {
	// 检查用户同�?	hasConsent, err := g.consentManager.IsConsentValid(ctx, data.UserID, purpose)
	if err != nil {
		return fmt.Errorf("failed to check consent: %w", err)
	}

	if !hasConsent {
		return fmt.Errorf("no valid consent for purpose: %s", purpose)
	}

	// 记录数据访问
	accessLog := DataAccessLog{
		ID:         generateID(),
		UserID:     data.UserID,
		AccessedBy: "system",
		DataType:   data.DataType,
		AccessType: "write",
		Purpose:    purpose,
		LegalBasis: data.LegalBasis,
		Timestamp:  time.Now(),
		Success:    true,
	}

	if err := g.auditLogger.LogDataAccess(ctx, accessLog); err != nil {
		return fmt.Errorf("failed to log data access: %w", err)
	}

	// 处理数据
	return g.dataProcessor.ProcessPersonalData(ctx, data, purpose)
}

// HandleDataSubjectRights 处理数据主体权利请求
func (g *GDPRCompliance) HandleDataSubjectRights(ctx context.Context, request DataSubjectRequest) error {
	switch request.RequestType {
	case "access":
		return g.handleAccessRequest(ctx, request)
	case "rectification":
		return g.handleRectificationRequest(ctx, request)
	case "erasure":
		return g.handleErasureRequest(ctx, request)
	case "portability":
		return g.handlePortabilityRequest(ctx, request)
	case "restriction":
		return g.handleRestrictionRequest(ctx, request)
	case "objection":
		return g.handleObjectionRequest(ctx, request)
	default:
		return fmt.Errorf("unsupported request type: %s", request.RequestType)
	}
}

// handleAccessRequest 处理访问请求
func (g *GDPRCompliance) handleAccessRequest(ctx context.Context, request DataSubjectRequest) error {
	// 导出用户数据
	export, err := g.dataProcessor.ExportPersonalData(ctx, request.UserID)
	if err != nil {
		return fmt.Errorf("failed to export personal data: %w", err)
	}

	// 记录访问
	accessLog := DataAccessLog{
		ID:         generateID(),
		UserID:     request.UserID,
		AccessedBy: request.AssignedTo,
		DataType:   "all",
		AccessType: "export",
		Purpose:    "data_subject_access",
		LegalBasis: "article_15_gdpr",
		Timestamp:  time.Now(),
		Success:    true,
	}

	return g.auditLogger.LogDataAccess(ctx, accessLog)
}

// handleErasureRequest 处理删除请求
func (g *GDPRCompliance) handleErasureRequest(ctx context.Context, request DataSubjectRequest) error {
	// 删除个人数据
	if err := g.dataProcessor.DeletePersonalData(ctx, request.UserID); err != nil {
		return fmt.Errorf("failed to delete personal data: %w", err)
	}

	// 记录删除
	modLog := DataModificationLog{
		ID:         generateID(),
		UserID:     request.UserID,
		ModifiedBy: request.AssignedTo,
		DataType:   "all",
		Operation:  "delete",
		Reason:     "data_subject_erasure",
		LegalBasis: "article_17_gdpr",
		Timestamp:  time.Now(),
	}

	return g.auditLogger.LogDataModification(ctx, modLog)
}

// handlePortabilityRequest 处理可移植性请�?func (g *GDPRCompliance) handlePortabilityRequest(ctx context.Context, request DataSubjectRequest) error {
	format := "json"
	if formatReq, ok := request.RequestData["format"].(string); ok {
		format = formatReq
	}

	// 导出数据
	export, err := g.dataPortability.ExportUserData(ctx, request.UserID, format)
	if err != nil {
		return fmt.Errorf("failed to export user data: %w", err)
	}

	// 记录导出
	accessLog := DataAccessLog{
		ID:         generateID(),
		UserID:     request.UserID,
		AccessedBy: request.AssignedTo,
		DataType:   "all",
		AccessType: "export",
		Purpose:    "data_portability",
		LegalBasis: "article_20_gdpr",
		Timestamp:  time.Now(),
		Success:    true,
	}

	return g.auditLogger.LogDataAccess(ctx, accessLog)
}

// handleRectificationRequest 处理更正请求
func (g *GDPRCompliance) handleRectificationRequest(ctx context.Context, request DataSubjectRequest) error {
	// 这里需要实现数据更正逻辑
	// 暂时记录修改日志
	modLog := DataModificationLog{
		ID:         generateID(),
		UserID:     request.UserID,
		ModifiedBy: request.AssignedTo,
		DataType:   "personal_data",
		Operation:  "update",
		NewValues:  request.RequestData,
		Reason:     "data_subject_rectification",
		LegalBasis: "article_16_gdpr",
		Timestamp:  time.Now(),
	}

	return g.auditLogger.LogDataModification(ctx, modLog)
}

// handleRestrictionRequest 处理限制处理请求
func (g *GDPRCompliance) handleRestrictionRequest(ctx context.Context, request DataSubjectRequest) error {
	// 这里需要实现处理限制逻辑
	modLog := DataModificationLog{
		ID:         generateID(),
		UserID:     request.UserID,
		ModifiedBy: request.AssignedTo,
		DataType:   "processing_status",
		Operation:  "restrict",
		Reason:     "data_subject_restriction",
		LegalBasis: "article_18_gdpr",
		Timestamp:  time.Now(),
	}

	return g.auditLogger.LogDataModification(ctx, modLog)
}

// handleObjectionRequest 处理反对请求
func (g *GDPRCompliance) handleObjectionRequest(ctx context.Context, request DataSubjectRequest) error {
	// 撤回相关同意
	if purpose, ok := request.RequestData["purpose"].(string); ok {
		if err := g.consentManager.WithdrawConsent(ctx, request.UserID, purpose); err != nil {
			return fmt.Errorf("failed to withdraw consent: %w", err)
		}
	}

	// 记录反对
	consentLog := ConsentChangeLog{
		ID:           generateID(),
		UserID:       request.UserID,
		Purpose:      "processing_objection",
		OldConsent:   true,
		NewConsent:   false,
		ChangeReason: "data_subject_objection",
		ChangeMethod: "article_21_gdpr",
		Timestamp:    time.Now(),
	}

	return g.auditLogger.LogConsentChange(ctx, consentLog)
}

// ValidateDataProcessing 验证数据处理的合法�?func (g *GDPRCompliance) ValidateDataProcessing(ctx context.Context, data PersonalData) error {
	// 检查法律依�?	if data.LegalBasis == "" {
		return fmt.Errorf("legal basis is required for data processing")
	}

	// 检查目的限�?	if len(data.Purpose) == 0 {
		return fmt.Errorf("processing purpose is required")
	}

	// 检查数据最小化
	if err := g.validateDataMinimization(data); err != nil {
		return fmt.Errorf("data minimization validation failed: %w", err)
	}

	// 检查保留期�?	policy, err := g.dataRetention.GetRetentionPolicy(ctx, data.DataType)
	if err != nil {
		return fmt.Errorf("failed to get retention policy: %w", err)
	}

	if data.ExpiresAt == nil {
		expiryDate := time.Now().Add(policy.RetentionPeriod)
		data.ExpiresAt = &expiryDate
	}

	return nil
}

// validateDataMinimization 验证数据最小化原则
func (g *GDPRCompliance) validateDataMinimization(data PersonalData) error {
	// 这里应该实现数据最小化验证逻辑
	// 检查收集的数据是否与处理目的相关且必要
	return nil
}

// GeneratePrivacyNotice 生成隐私通知
func (g *GDPRCompliance) GeneratePrivacyNotice(ctx context.Context, language string) (string, error) {
	// 这里应该根据语言生成隐私通知
	// 暂时返回英文模板
	notice := `
PRIVACY NOTICE

This privacy notice explains how we collect, use, and protect your personal data in accordance with the General Data Protection Regulation (GDPR).

1. DATA CONTROLLER
[Company Name and Contact Information]

2. DATA PROTECTION OFFICER
[DPO Contact Information]

3. PERSONAL DATA WE COLLECT
- Identity data (name, email, phone)
- Usage data (interaction with our services)
- Technical data (IP address, browser type)

4. LEGAL BASIS FOR PROCESSING
- Consent (Article 6(1)(a) GDPR)
- Contract performance (Article 6(1)(b) GDPR)
- Legitimate interests (Article 6(1)(f) GDPR)

5. YOUR RIGHTS
- Right of access (Article 15)
- Right to rectification (Article 16)
- Right to erasure (Article 17)
- Right to restrict processing (Article 18)
- Right to data portability (Article 20)
- Right to object (Article 21)

6. DATA RETENTION
Personal data will be retained only for as long as necessary for the purposes for which it was collected.

7. CONTACT US
If you have any questions about this privacy notice or wish to exercise your rights, please contact us at [contact information].
`
	return notice, nil
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("gdpr_%d", time.Now().UnixNano())
}
