// 太上老君AI平台CCPA合规性支持模?
package compliance

import (
	"context"
	"fmt"
	"time"
)

// CCPACompliance CCPA合规性管理器
type CCPACompliance struct {
	personalInfoManager PersonalInfoManager
	consumerRights      ConsumerRightsManager
	businessPurposes    BusinessPurposesManager
	thirdPartySharing   ThirdPartySharingManager
	optOutManager       OptOutManager
	disclosureManager   DisclosureManager
	auditLogger         AuditLogger
}

// PersonalInfoManager 个人信息管理器接?
type PersonalInfoManager interface {
	CollectPersonalInfo(ctx context.Context, info PersonalInformation) error
	GetPersonalInfo(ctx context.Context, consumerID string) (PersonalInformation, error)
	DeletePersonalInfo(ctx context.Context, consumerID string) error
	GetPersonalInfoCategories(ctx context.Context, consumerID string) ([]PersonalInfoCategory, error)
}

// ConsumerRightsManager 消费者权利管理器接口
type ConsumerRightsManager interface {
	HandleKnowRequest(ctx context.Context, request ConsumerRequest) error
	HandleDeleteRequest(ctx context.Context, request ConsumerRequest) error
	HandleOptOutRequest(ctx context.Context, request ConsumerRequest) error
	VerifyConsumerIdentity(ctx context.Context, verification IdentityVerification) (bool, error)
}

// BusinessPurposesManager 商业目的管理器接?
type BusinessPurposesManager interface {
	GetBusinessPurposes(ctx context.Context) ([]BusinessPurpose, error)
	ValidateBusinessPurpose(ctx context.Context, purpose string) (bool, error)
	GetPurposeDisclosure(ctx context.Context, purpose string) (PurposeDisclosure, error)
}

// ThirdPartySharingManager 第三方共享管理器接口
type ThirdPartySharingManager interface {
	RecordThirdPartySharing(ctx context.Context, sharing ThirdPartySharing) error
	GetThirdPartySharing(ctx context.Context, consumerID string) ([]ThirdPartySharing, error)
	GetThirdPartyCategories(ctx context.Context) ([]ThirdPartyCategory, error)
}

// OptOutManager 选择退出管理器接口
type OptOutManager interface {
	ProcessOptOut(ctx context.Context, optOut OptOutRequest) error
	GetOptOutStatus(ctx context.Context, consumerID string) (OptOutStatus, error)
	RespectOptOut(ctx context.Context, consumerID string) error
}

// DisclosureManager 披露管理器接?
type DisclosureManager interface {
	GeneratePrivacyPolicy(ctx context.Context, language string) (PrivacyPolicy, error)
	GenerateDataDisclosure(ctx context.Context, period TimePeriod) (DataDisclosure, error)
	UpdateDisclosures(ctx context.Context, updates DisclosureUpdates) error
}

// PersonalInformation 个人信息结构
type PersonalInformation struct {
	ConsumerID       string                 `json:"consumer_id"`
	Categories       []PersonalInfoCategory `json:"categories"`
	Sources          []InformationSource    `json:"sources"`
	BusinessPurposes []string               `json:"business_purposes"`
	ThirdParties     []ThirdPartyRecipient  `json:"third_parties"`
	SaleStatus       SaleStatus             `json:"sale_status"`
	CollectionDate   time.Time              `json:"collection_date"`
	LastUpdated      time.Time              `json:"last_updated"`
	RetentionPeriod  time.Duration          `json:"retention_period"`
	Metadata         PersonalInfoMetadata   `json:"metadata"`
}

// PersonalInfoCategory 个人信息类别
type PersonalInfoCategory struct {
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Examples    []string               `json:"examples"`
	Data        map[string]interface{} `json:"data"`
	Sensitive   bool                   `json:"sensitive"`
	Source      string                 `json:"source"`
	Purpose     []string               `json:"purpose"`
	Disclosed   bool                   `json:"disclosed"`
	Sold        bool                   `json:"sold"`
	CollectedAt time.Time              `json:"collected_at"`
}

// InformationSource 信息来源
type InformationSource struct {
	SourceType  string    `json:"source_type"` // direct, third_party, public_records, social_media
	SourceName  string    `json:"source_name"`
	Description string    `json:"description"`
	Categories  []string  `json:"categories"`
	CollectedAt time.Time `json:"collected_at"`
	Purpose     string    `json:"purpose"`
	Consent     bool      `json:"consent"`
}

// ThirdPartyRecipient 第三方接收方
type ThirdPartyRecipient struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"` // service_provider, third_party, affiliate
	Purpose      string    `json:"purpose"`
	Categories   []string  `json:"categories"`
	SharedAt     time.Time `json:"shared_at"`
	ContractType string    `json:"contract_type"`
	Location     string    `json:"location"`
}

// SaleStatus 销售状?
type SaleStatus struct {
	IsSold     bool       `json:"is_sold"`
	SoldTo     []string   `json:"sold_to"`
	SaleDate   *time.Time `json:"sale_date,omitempty"`
	SalePrice  *float64   `json:"sale_price,omitempty"`
	OptedOut   bool       `json:"opted_out"`
	OptOutDate *time.Time `json:"opt_out_date,omitempty"`
}

// PersonalInfoMetadata 个人信息元数?
type PersonalInfoMetadata struct {
	Version         string          `json:"version"`
	LastAudit       time.Time       `json:"last_audit"`
	ComplianceFlags map[string]bool `json:"compliance_flags"`
	Tags            []string        `json:"tags"`
	Notes           string          `json:"notes"`
}

// ConsumerRequest 消费者请?
type ConsumerRequest struct {
	ID                 string                 `json:"id"`
	ConsumerID         string                 `json:"consumer_id"`
	RequestType        string                 `json:"request_type"`   // know, delete, opt_out
	RequestMethod      string                 `json:"request_method"` // online, phone, email, mail
	RequestDate        time.Time              `json:"request_date"`
	VerificationMethod string                 `json:"verification_method"`
	VerificationData   map[string]interface{} `json:"verification_data"`
	Status             string                 `json:"status"` // pending, verified, processing, completed, denied
	ResponseDue        time.Time              `json:"response_due"`
	ResponseDate       *time.Time             `json:"response_date,omitempty"`
	Response           string                 `json:"response"`
	ContactInfo        ConsumerContact        `json:"contact_info"`
	SpecificRequests   []string               `json:"specific_requests"`
	Reason             string                 `json:"reason"`
	Priority           string                 `json:"priority"`
}

// ConsumerContact 消费者联系信?
type ConsumerContact struct {
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	MailingAddress  string `json:"mailing_address"`
	PreferredMethod string `json:"preferred_method"`
	Language        string `json:"language"`
}

// IdentityVerification 身份验证
type IdentityVerification struct {
	ConsumerID         string                 `json:"consumer_id"`
	RequestID          string                 `json:"request_id"`
	VerificationMethod string                 `json:"verification_method"`
	ProvidedData       map[string]interface{} `json:"provided_data"`
	RequiredData       []string               `json:"required_data"`
	VerificationDate   time.Time              `json:"verification_date"`
	Verified           bool                   `json:"verified"`
	FailureReason      string                 `json:"failure_reason,omitempty"`
	AttemptCount       int                    `json:"attempt_count"`
	MaxAttempts        int                    `json:"max_attempts"`
}

// BusinessPurpose 商业目的
type BusinessPurpose struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Categories      []string  `json:"categories"`
	LegalBasis      string    `json:"legal_basis"`
	IsCommercial    bool      `json:"is_commercial"`
	RequiresConsent bool      `json:"requires_consent"`
	RetentionPeriod string    `json:"retention_period"`
	Examples        []string  `json:"examples"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	IsActive        bool      `json:"is_active"`
}

// PurposeDisclosure 目的披露
type PurposeDisclosure struct {
	Purpose           string    `json:"purpose"`
	Categories        []string  `json:"categories"`
	Sources           []string  `json:"sources"`
	Recipients        []string  `json:"recipients"`
	RetentionPeriod   string    `json:"retention_period"`
	LastUpdated       time.Time `json:"last_updated"`
	PubliclyAvailable bool      `json:"publicly_available"`
}

// ThirdPartySharing 第三方共享记?
type ThirdPartySharing struct {
	ID           string    `json:"id"`
	ConsumerID   string    `json:"consumer_id"`
	ThirdParty   string    `json:"third_party"`
	Categories   []string  `json:"categories"`
	Purpose      string    `json:"purpose"`
	ShareDate    time.Time `json:"share_date"`
	ShareMethod  string    `json:"share_method"`
	IsSale       bool      `json:"is_sale"`
	Compensation *float64  `json:"compensation,omitempty"`
	ContractRef  string    `json:"contract_ref"`
	OptedOut     bool      `json:"opted_out"`
}

// ThirdPartyCategory 第三方分?
type ThirdPartyCategory struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	Purposes    []string `json:"purposes"`
	IsActive    bool     `json:"is_active"`
}

// OptOutRequest 选择退出请?
type OptOutRequest struct {
	ID                   string          `json:"id"`
	ConsumerID           string          `json:"consumer_id"`
	RequestDate          time.Time       `json:"request_date"`
	RequestMethod        string          `json:"request_method"`
	OptOutType           string          `json:"opt_out_type"` // sale, sharing, targeted_advertising
	Scope                []string        `json:"scope"`
	EffectiveDate        time.Time       `json:"effective_date"`
	ProcessedDate        *time.Time      `json:"processed_date,omitempty"`
	Status               string          `json:"status"`
	VerificationRequired bool            `json:"verification_required"`
	ContactInfo          ConsumerContact `json:"contact_info"`
}

// OptOutStatus 选择退出状?
type OptOutStatus struct {
	ConsumerID   string     `json:"consumer_id"`
	OptedOut     bool       `json:"opted_out"`
	OptOutDate   *time.Time `json:"opt_out_date,omitempty"`
	OptOutTypes  []string   `json:"opt_out_types"`
	LastUpdated  time.Time  `json:"last_updated"`
	GlobalOptOut bool       `json:"global_opt_out"`
	Exceptions   []string   `json:"exceptions"`
}

// PrivacyPolicy 隐私政策
type PrivacyPolicy struct {
	Version        string               `json:"version"`
	EffectiveDate  time.Time            `json:"effective_date"`
	Language       string               `json:"language"`
	Content        PrivacyPolicyContent `json:"content"`
	LastUpdated    time.Time            `json:"last_updated"`
	ApprovalStatus string               `json:"approval_status"`
	PublicURL      string               `json:"public_url"`
}

// PrivacyPolicyContent 隐私政策内容
type PrivacyPolicyContent struct {
	Introduction         string                     `json:"introduction"`
	InformationCollected []InformationCollection    `json:"information_collected"`
	BusinessPurposes     []BusinessPurposeSection   `json:"business_purposes"`
	ThirdPartySharing    []ThirdPartySharingSection `json:"third_party_sharing"`
	ConsumerRights       ConsumerRightsSection      `json:"consumer_rights"`
	DataSecurity         string                     `json:"data_security"`
	DataRetention        string                     `json:"data_retention"`
	ContactInformation   string                     `json:"contact_information"`
	Updates              string                     `json:"updates"`
}

// InformationCollection 信息收集部分
type InformationCollection struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	Sources     []string `json:"sources"`
	Purposes    []string `json:"purposes"`
}

// BusinessPurposeSection 商业目的部分
type BusinessPurposeSection struct {
	Purpose     string   `json:"purpose"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	Examples    []string `json:"examples"`
}

// ThirdPartySharingSection 第三方共享部?
type ThirdPartySharingSection struct {
	ThirdParty string   `json:"third_party"`
	Purpose    string   `json:"purpose"`
	Categories []string `json:"categories"`
	IsSale     bool     `json:"is_sale"`
}

// ConsumerRightsSection 消费者权利部?
type ConsumerRightsSection struct {
	RightToKnow       string `json:"right_to_know"`
	RightToDelete     string `json:"right_to_delete"`
	RightToOptOut     string `json:"right_to_opt_out"`
	NonDiscrimination string `json:"non_discrimination"`
	HowToExercise     string `json:"how_to_exercise"`
}

// DataDisclosure 数据披露
type DataDisclosure struct {
	Period              TimePeriod                  `json:"period"`
	GeneratedAt         time.Time                   `json:"generated_at"`
	CategoriesCollected []CategoryDisclosure        `json:"categories_collected"`
	CategoriesSold      []CategoryDisclosure        `json:"categories_sold"`
	CategoriesShared    []CategoryDisclosure        `json:"categories_shared"`
	BusinessPurposes    []BusinessPurposeDisclosure `json:"business_purposes"`
	ThirdParties        []ThirdPartyDisclosure      `json:"third_parties"`
	ConsumerRequests    RequestStatistics           `json:"consumer_requests"`
}

// CategoryDisclosure 类别披露
type CategoryDisclosure struct {
	Category   string   `json:"category"`
	Sources    []string `json:"sources"`
	Purposes   []string `json:"purposes"`
	Recipients []string `json:"recipients"`
	SoldTo     []string `json:"sold_to,omitempty"`
	SharedWith []string `json:"shared_with,omitempty"`
}

// BusinessPurposeDisclosure 商业目的披露
type BusinessPurposeDisclosure struct {
	Purpose    string   `json:"purpose"`
	Categories []string `json:"categories"`
	Frequency  string   `json:"frequency"`
}

// ThirdPartyDisclosure 第三方披?
type ThirdPartyDisclosure struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Categories []string `json:"categories"`
	Purpose    string   `json:"purpose"`
	IsSale     bool     `json:"is_sale"`
}

// DisclosureUpdates 披露更新
type DisclosureUpdates struct {
	PolicyUpdates     []PolicyUpdate     `json:"policy_updates"`
	CategoryUpdates   []CategoryUpdate   `json:"category_updates"`
	PurposeUpdates    []PurposeUpdate    `json:"purpose_updates"`
	ThirdPartyUpdates []ThirdPartyUpdate `json:"third_party_updates"`
	UpdateDate        time.Time          `json:"update_date"`
	UpdatedBy         string             `json:"updated_by"`
	Reason            string             `json:"reason"`
}

// PolicyUpdate 政策更新
type PolicyUpdate struct {
	Section    string `json:"section"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
	ChangeType string `json:"change_type"` // addition, modification, deletion
	Reason     string `json:"reason"`
}

// CategoryUpdate 类别更新
type CategoryUpdate struct {
	Category string   `json:"category"`
	Action   string   `json:"action"` // add, remove, modify
	Changes  []string `json:"changes"`
	Reason   string   `json:"reason"`
}

// PurposeUpdate 目的更新
type PurposeUpdate struct {
	Purpose string   `json:"purpose"`
	Action  string   `json:"action"`
	Changes []string `json:"changes"`
	Reason  string   `json:"reason"`
}

// ThirdPartyUpdate 第三方更?
type ThirdPartyUpdate struct {
	ThirdParty string   `json:"third_party"`
	Action     string   `json:"action"`
	Changes    []string `json:"changes"`
	Reason     string   `json:"reason"`
}

// NewCCPACompliance 创建新的CCPA合规性管理器
func NewCCPACompliance(
	personalInfoManager PersonalInfoManager,
	consumerRights ConsumerRightsManager,
	businessPurposes BusinessPurposesManager,
	thirdPartySharing ThirdPartySharingManager,
	optOutManager OptOutManager,
	disclosureManager DisclosureManager,
	auditLogger AuditLogger,
) *CCPACompliance {
	return &CCPACompliance{
		personalInfoManager: personalInfoManager,
		consumerRights:      consumerRights,
		businessPurposes:    businessPurposes,
		thirdPartySharing:   thirdPartySharing,
		optOutManager:       optOutManager,
		disclosureManager:   disclosureManager,
		auditLogger:         auditLogger,
	}
}

// CollectPersonalInformation 收集个人信息
func (c *CCPACompliance) CollectPersonalInformation(ctx context.Context, info PersonalInformation) error {
	// 验证收集的合?
	if err := c.validateCollection(ctx, info); err != nil {
		return fmt.Errorf("collection validation failed: %w", err)
	}

	// 记录收集
	if err := c.personalInfoManager.CollectPersonalInfo(ctx, info); err != nil {
		return fmt.Errorf("failed to collect personal info: %w", err)
	}

	// 记录审计日志
	accessLog := DataAccessLog{
		ID:         generateCCPAID(),
		UserID:     info.ConsumerID,
		AccessedBy: "system",
		DataType:   "personal_information",
		AccessType: "collect",
		Purpose:    fmt.Sprintf("business_purposes: %v", info.BusinessPurposes),
		Timestamp:  time.Now(),
		Success:    true,
	}

	return c.auditLogger.LogDataAccess(ctx, accessLog)
}

// HandleConsumerRequest 处理消费者请?
func (c *CCPACompliance) HandleConsumerRequest(ctx context.Context, request ConsumerRequest) error {
	// 验证消费者身?
	verification := IdentityVerification{
		ConsumerID:         request.ConsumerID,
		RequestID:          request.ID,
		VerificationMethod: request.VerificationMethod,
		ProvidedData:       request.VerificationData,
		VerificationDate:   time.Now(),
		MaxAttempts:        3,
	}

	verified, err := c.consumerRights.VerifyConsumerIdentity(ctx, verification)
	if err != nil {
		return fmt.Errorf("identity verification failed: %w", err)
	}

	if !verified {
		return fmt.Errorf("consumer identity could not be verified")
	}

	// 根据请求类型处理
	switch request.RequestType {
	case "know":
		return c.consumerRights.HandleKnowRequest(ctx, request)
	case "delete":
		return c.consumerRights.HandleDeleteRequest(ctx, request)
	case "opt_out":
		return c.consumerRights.HandleOptOutRequest(ctx, request)
	default:
		return fmt.Errorf("unsupported request type: %s", request.RequestType)
	}
}

// ProcessOptOut 处理选择退出请?
func (c *CCPACompliance) ProcessOptOut(ctx context.Context, optOut OptOutRequest) error {
	// 处理选择退出请?
	if err := c.optOutManager.ProcessOptOut(ctx, optOut); err != nil {
		return fmt.Errorf("failed to process opt-out: %w", err)
	}

	// 确保遵守选择退出协?
	if err := c.optOutManager.RespectOptOut(ctx, optOut.ConsumerID); err != nil {
		return fmt.Errorf("failed to respect opt-out: %w", err)
	}

	// 记录选择退出日?
	consentLog := ConsentChangeLog{
		ID:           generateCCPAID(),
		UserID:       optOut.ConsumerID,
		Purpose:      optOut.OptOutType,
		OldConsent:   true,
		NewConsent:   false,
		ChangeReason: "consumer_opt_out",
		ChangeMethod: "ccpa_opt_out",
		Timestamp:    time.Now(),
	}

	return c.auditLogger.LogConsentChange(ctx, consentLog)
}

// ValidateThirdPartySharing 验证第三方共?
func (c *CCPACompliance) ValidateThirdPartySharing(ctx context.Context, sharing ThirdPartySharing) error {
	// 检查消费者是否选择退?
	optOutStatus, err := c.optOutManager.GetOptOutStatus(ctx, sharing.ConsumerID)
	if err != nil {
		return fmt.Errorf("failed to get opt-out status: %w", err)
	}

	if optOutStatus.OptedOut {
		return fmt.Errorf("consumer has opted out of data sharing")
	}

	// 验证商业目的
	for _, category := range sharing.Categories {
		valid, err := c.businessPurposes.ValidateBusinessPurpose(ctx, sharing.Purpose)
		if err != nil {
			return fmt.Errorf("failed to validate business purpose: %w", err)
		}
		if !valid {
			return fmt.Errorf("invalid business purpose: %s", sharing.Purpose)
		}
	}

	// 记录第三方共?
	if err := c.thirdPartySharing.RecordThirdPartySharing(ctx, sharing); err != nil {
		return fmt.Errorf("failed to record third party sharing: %w", err)
	}

	return nil
}

// validateCollection 验证收集的合?
func (c *CCPACompliance) validateCollection(ctx context.Context, info PersonalInformation) error {
	// 验证商业目的
	for _, purpose := range info.BusinessPurposes {
		valid, err := c.businessPurposes.ValidateBusinessPurpose(ctx, purpose)
		if err != nil {
			return fmt.Errorf("failed to validate business purpose %s: %w", purpose, err)
		}
		if !valid {
			return fmt.Errorf("invalid business purpose: %s", purpose)
		}
	}

	// 验证数据最小化
	if err := c.validateDataMinimization(info); err != nil {
		return fmt.Errorf("data minimization validation failed: %w", err)
	}

	return nil
}

// validateDataMinimization 验证数据最小化
func (c *CCPACompliance) validateDataMinimization(info PersonalInformation) error {
	// 检查收集的信息是否与商业目的相?
	if len(info.Categories) == 0 {
		return fmt.Errorf("no categories specified for business purposes")
	}

	// 这里应该实现具体的验证逻辑
	return nil
}

// GenerateConsumerDisclosure 生成消费者披?
func (c *CCPACompliance) GenerateConsumerDisclosure(ctx context.Context, consumerID string) (string, error) {
	// 获取个人信息
	personalInfo, err := c.personalInfoManager.GetPersonalInfo(ctx, consumerID)
	if err != nil {
		return "", fmt.Errorf("failed to get personal info: %w", err)
	}

	// 获取第三方共享信?
	thirdPartySharing, err := c.thirdPartySharing.GetThirdPartySharing(ctx, consumerID)
	if err != nil {
		return "", fmt.Errorf("failed to get third party sharing: %w", err)
	}

	// 生成披露文档
	disclosure := fmt.Sprintf(`
CONSUMER DISCLOSURE UNDER CCPA

Consumer ID: %s
Disclosure Date: %s

PERSONAL INFORMATION COLLECTED:
`, consumerID, time.Now().Format("2006-01-02"))

	for _, category := range personalInfo.Categories {
		disclosure += fmt.Sprintf("- %s: %s\n", category.Category, category.Description)
	}

	disclosure += "\nBUSINESS PURPOSES:\n"
	for _, purpose := range personalInfo.BusinessPurposes {
		disclosure += fmt.Sprintf("- %s\n", purpose)
	}

	disclosure += "\nTHIRD PARTY SHARING:\n"
	for _, sharing := range thirdPartySharing {
		disclosure += fmt.Sprintf("- Shared with %s for %s\n", sharing.ThirdParty, sharing.Purpose)
	}

	disclosure += `
YOUR RIGHTS UNDER CCPA:
- Right to Know: You have the right to know what personal information we collect, use, disclose, and sell
- Right to Delete: You have the right to request deletion of your personal information
- Right to Opt-Out: You have the right to opt-out of the sale of your personal information
- Right to Non-Discrimination: We will not discriminate against you for exercising your rights

To exercise your rights, please contact us at [contact information].
`

	return disclosure, nil
}

// GetSupportedCategories 获取支持的个人信息类?
func (c *CCPACompliance) GetSupportedCategories() []PersonalInfoCategory {
	return []PersonalInfoCategory{
		{
			Category:    "identifiers",
			Description: "Real name, alias, postal address, unique personal identifier, online identifier, Internet Protocol address, email address, account name, social security number, driver's license number, passport number, or other similar identifiers",
			Examples:    []string{"name", "email", "phone", "address", "IP address"},
		},
		{
			Category:    "personal_info_records",
			Description: "Personal information as defined in the California Customer Records statute",
			Examples:    []string{"signature", "physical characteristics", "education", "employment"},
		},
		{
			Category:    "protected_characteristics",
			Description: "Characteristics of protected classifications under California or federal law",
			Examples:    []string{"age", "race", "religion", "sexual orientation", "disability status"},
			Sensitive:   true,
		},
		{
			Category:    "commercial_info",
			Description: "Commercial information including records of personal property, products or services purchased, obtained, or considered",
			Examples:    []string{"purchase history", "browsing history", "preferences"},
		},
		{
			Category:    "biometric_info",
			Description: "Biometric information",
			Examples:    []string{"fingerprints", "voiceprints", "facial recognition data"},
			Sensitive:   true,
		},
		{
			Category:    "internet_activity",
			Description: "Internet or other electronic network activity information",
			Examples:    []string{"browsing history", "search history", "interaction with websites"},
		},
		{
			Category:    "geolocation_data",
			Description: "Geolocation data",
			Examples:    []string{"precise location", "general location"},
		},
		{
			Category:    "sensory_data",
			Description: "Audio, electronic, visual, thermal, olfactory, or similar information",
			Examples:    []string{"call recordings", "photos", "videos"},
		},
		{
			Category:    "professional_info",
			Description: "Professional or employment-related information",
			Examples:    []string{"job title", "work history", "performance evaluations"},
		},
		{
			Category:    "education_info",
			Description: "Education information that is not publicly available personally identifiable information",
			Examples:    []string{"grades", "transcripts", "disciplinary records"},
		},
		{
			Category:    "inferences",
			Description: "Inferences drawn from any of the information to create a profile reflecting preferences, characteristics, psychological trends, predispositions, behavior, attitudes, intelligence, abilities, and aptitudes",
			Examples:    []string{"preferences", "predictions", "behavioral profiles"},
		},
	}
}

// generateCCPAID 生成CCPA相关的唯一ID
func generateCCPAID() string {
	return fmt.Sprintf("ccpa_%d", time.Now().UnixNano())
}

