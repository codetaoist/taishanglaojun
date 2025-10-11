package domain

import (
	"time"

	"github.com/google/uuid"
)

// Gender 性别枚举
type Gender string

const (
	GenderMale   Gender = "male"   // 男�?
	GenderFemale Gender = "female" // 女�?
	GenderOther  Gender = "other"  // 其他
)

// BloodType 血型枚�?
type BloodType string

const (
	BloodTypeA  BloodType = "A"  // A�?
	BloodTypeB  BloodType = "B"  // B�?
	BloodTypeAB BloodType = "AB" // AB�?
	BloodTypeO  BloodType = "O"  // O�?
)

// HealthProfile 健康档案聚合�?
type HealthProfile struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;unique"`
	Gender            Gender     `json:"gender" gorm:"type:varchar(10)"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	Height            *float64   `json:"height,omitempty" gorm:"comment:身高(cm)"`
	BloodType         *BloodType `json:"blood_type,omitempty" gorm:"type:varchar(5)"`
	EmergencyContact  string     `json:"emergency_contact" gorm:"type:varchar(20)"`
	EmergencyName     string     `json:"emergency_name" gorm:"type:varchar(100)"`
	MedicalHistory    []string   `json:"medical_history,omitempty" gorm:"type:jsonb;comment:病史"`
	Allergies         []string   `json:"allergies,omitempty" gorm:"type:jsonb;comment:过敏�?`
	Medications       []string   `json:"medications,omitempty" gorm:"type:jsonb;comment:用药�?`
	HealthGoals       []string   `json:"health_goals,omitempty" gorm:"type:jsonb;comment:健康目标"`
	PreferredUnits    map[string]string `json:"preferred_units,omitempty" gorm:"type:jsonb;comment:偏好单位"`
	NotificationPrefs map[string]bool   `json:"notification_prefs,omitempty" gorm:"type:jsonb;comment:通知偏好"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 领域事件
	events []DomainEvent
}

// NewHealthProfile 创建新的健康档案
func NewHealthProfile(userID uuid.UUID, gender Gender) *HealthProfile {
	id := uuid.New()
	now := time.Now()
	
	profile := &HealthProfile{
		ID:                id,
		UserID:            userID,
		Gender:            gender,
		PreferredUnits:    getDefaultUnits(),
		NotificationPrefs: getDefaultNotificationPrefs(),
		CreatedAt:         now,
		UpdatedAt:         now,
		events:            make([]DomainEvent, 0),
	}
	
	// 发布健康档案创建事件
	profile.publishEvent(NewHealthProfileCreatedEvent(id, userID, gender))
	
	return profile
}

// UpdateBasicInfo 更新基本信息
func (h *HealthProfile) UpdateBasicInfo(gender Gender, dateOfBirth *time.Time, height *float64, bloodType *BloodType) {
	h.Gender = gender
	h.DateOfBirth = dateOfBirth
	h.Height = height
	h.BloodType = bloodType
	h.UpdatedAt = time.Now()
	
	// 发布基本信息更新事件
	h.publishEvent(NewHealthProfileUpdatedEvent(h.ID, h.UserID, "basic_info"))
}

// UpdateEmergencyContact 更新紧急联系人
func (h *HealthProfile) UpdateEmergencyContact(name, contact string) {
	h.EmergencyName = name
	h.EmergencyContact = contact
	h.UpdatedAt = time.Now()
	
	// 发布紧急联系人更新事件
	h.publishEvent(NewHealthProfileUpdatedEvent(h.ID, h.UserID, "emergency_contact"))
}

// AddMedicalHistory 添加病史
func (h *HealthProfile) AddMedicalHistory(condition string) {
	if h.MedicalHistory == nil {
		h.MedicalHistory = make([]string, 0)
	}
	
	// 检查是否已存在
	for _, existing := range h.MedicalHistory {
		if existing == condition {
			return
		}
	}
	
	h.MedicalHistory = append(h.MedicalHistory, condition)
	h.UpdatedAt = time.Now()
	
	// 发布病史添加事件
	h.publishEvent(NewMedicalHistoryAddedEvent(h.ID, h.UserID, condition))
}

// RemoveMedicalHistory 移除病史
func (h *HealthProfile) RemoveMedicalHistory(condition string) {
	if h.MedicalHistory == nil {
		return
	}
	
	for i, existing := range h.MedicalHistory {
		if existing == condition {
			h.MedicalHistory = append(h.MedicalHistory[:i], h.MedicalHistory[i+1:]...)
			h.UpdatedAt = time.Now()
			
			// 发布病史移除事件
			h.publishEvent(NewMedicalHistoryRemovedEvent(h.ID, h.UserID, condition))
			break
		}
	}
}

// AddAllergy 添加过敏�?
func (h *HealthProfile) AddAllergy(allergen string) {
	if h.Allergies == nil {
		h.Allergies = make([]string, 0)
	}
	
	// 检查是否已存在
	for _, existing := range h.Allergies {
		if existing == allergen {
			return
		}
	}
	
	h.Allergies = append(h.Allergies, allergen)
	h.UpdatedAt = time.Now()
	
	// 发布过敏史添加事�?
	h.publishEvent(NewAllergyAddedEvent(h.ID, h.UserID, allergen))
}

// RemoveAllergy 移除过敏�?
func (h *HealthProfile) RemoveAllergy(allergen string) {
	if h.Allergies == nil {
		return
	}
	
	for i, existing := range h.Allergies {
		if existing == allergen {
			h.Allergies = append(h.Allergies[:i], h.Allergies[i+1:]...)
			h.UpdatedAt = time.Now()
			
			// 发布过敏史移除事�?
			h.publishEvent(NewAllergyRemovedEvent(h.ID, h.UserID, allergen))
			break
		}
	}
}

// AddMedication 添加用药�?
func (h *HealthProfile) AddMedication(medication string) {
	if h.Medications == nil {
		h.Medications = make([]string, 0)
	}
	
	// 检查是否已存在
	for _, existing := range h.Medications {
		if existing == medication {
			return
		}
	}
	
	h.Medications = append(h.Medications, medication)
	h.UpdatedAt = time.Time{}
	
	// 发布用药史添加事�?
	h.publishEvent(NewMedicationAddedEvent(h.ID, h.UserID, medication))
}

// RemoveMedication 移除用药�?
func (h *HealthProfile) RemoveMedication(medication string) {
	if h.Medications == nil {
		return
	}
	
	for i, existing := range h.Medications {
		if existing == medication {
			h.Medications = append(h.Medications[:i], h.Medications[i+1:]...)
			h.UpdatedAt = time.Now()
			
			// 发布用药史移除事�?
			h.publishEvent(NewMedicationRemovedEvent(h.ID, h.UserID, medication))
			break
		}
	}
}

// SetHealthGoals 设置健康目标
func (h *HealthProfile) SetHealthGoals(goals []string) {
	h.HealthGoals = goals
	h.UpdatedAt = time.Now()
	
	// 发布健康目标设置事件
	h.publishEvent(NewHealthGoalsSetEvent(h.ID, h.UserID, goals))
}

// UpdatePreferredUnits 更新偏好单位
func (h *HealthProfile) UpdatePreferredUnits(units map[string]string) {
	if h.PreferredUnits == nil {
		h.PreferredUnits = make(map[string]string)
	}
	
	for key, value := range units {
		h.PreferredUnits[key] = value
	}
	h.UpdatedAt = time.Now()
}

// UpdateNotificationPrefs 更新通知偏好
func (h *HealthProfile) UpdateNotificationPrefs(prefs map[string]bool) {
	if h.NotificationPrefs == nil {
		h.NotificationPrefs = make(map[string]bool)
	}
	
	for key, value := range prefs {
		h.NotificationPrefs[key] = value
	}
	h.UpdatedAt = time.Now()
}

// GetAge 计算年龄
func (h *HealthProfile) GetAge() *int {
	if h.DateOfBirth == nil {
		return nil
	}
	
	now := time.Now()
	age := now.Year() - h.DateOfBirth.Year()
	
	// 检查是否还没过生日
	if now.YearDay() < h.DateOfBirth.YearDay() {
		age--
	}
	
	return &age
}

// GetBMI 计算BMI
func (h *HealthProfile) GetBMI(weight float64) *float64 {
	if h.Height == nil || *h.Height <= 0 || weight <= 0 {
		return nil
	}
	
	heightInMeters := *h.Height / 100 // 转换为米
	bmi := weight / (heightInMeters * heightInMeters)
	
	return &bmi
}

// publishEvent 发布领域事件
func (h *HealthProfile) publishEvent(event DomainEvent) {
	h.events = append(h.events, event)
}

// GetEvents 获取领域事件
func (h *HealthProfile) GetEvents() []DomainEvent {
	return h.events
}

// ClearEvents 清除领域事件
func (h *HealthProfile) ClearEvents() {
	h.events = make([]DomainEvent, 0)
}

// getDefaultUnits 获取默认单位设置
func getDefaultUnits() map[string]string {
	return map[string]string{
		"weight":      "kg",
		"height":      "cm",
		"temperature": "celsius",
		"distance":    "km",
		"blood_sugar": "mmol/L",
	}
}

// getDefaultNotificationPrefs 获取默认通知偏好
func getDefaultNotificationPrefs() map[string]bool {
	return map[string]bool{
		"abnormal_data":    true,
		"health_reminders": true,
		"goal_progress":    true,
		"weekly_summary":   true,
	}
}
