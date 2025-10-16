package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB JSONB
type JSONB map[string]interface{}

// Value driver.Valuer
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan sql.Scanner
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, j)
}

// ThreatAlert 澯
type ThreatAlert struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Severity    string    `json:"severity" gorm:"not null"` // low, medium, high, critical
	Category    string    `json:"category" gorm:"not null"` // malware, intrusion, ddos, etc.
	SourceIP    string    `json:"source_ip"`
	TargetIP    string    `json:"target_ip"`
	UserID      *string   `json:"user_id" gorm:"type:uuid"`
	RuleID      string    `json:"rule_id" gorm:"type:uuid"`
	Status      string    `json:"status" gorm:"default:'open'"` // open, investigating, resolved, false_positive
	RawData     JSONB     `json:"raw_data" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DetectionRule ?
type DetectionRule struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Category    string    `json:"category" gorm:"not null"`
	Severity    string    `json:"severity" gorm:"not null"`
	Conditions  JSONB     `json:"conditions" gorm:"type:jsonb;not null"`
	Actions     JSONB     `json:"actions" gorm:"type:jsonb;not null"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedBy   string    `json:"created_by" gorm:"type:uuid"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Vulnerability 
type Vulnerability struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID       *string   `json:"project_id" gorm:"type:uuid"`
	Title           string    `json:"title" gorm:"not null"`
	Description     string    `json:"description"`
	Severity        string    `json:"severity" gorm:"not null"` // low, medium, high, critical
	CVSSScore       *float64  `json:"cvss_score"`
	CVEID           *string   `json:"cve_id"`
	AffectedSystems JSONB     `json:"affected_systems" gorm:"type:jsonb"`
	Remediation     string    `json:"remediation"`
	Status          string    `json:"status" gorm:"default:'open'"` // open, in_progress, resolved, wont_fix
	DiscoveredAt    time.Time `json:"discovered_at" gorm:"autoCreateTime"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ScanJob 
type ScanJob struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"` // vulnerability, port, web, network
	Target      string    `json:"target" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, running, completed, failed
	Progress    int       `json:"progress" gorm:"default:0"`
	Config      JSONB     `json:"config" gorm:"type:jsonb"`
	Results     JSONB     `json:"results" gorm:"type:jsonb"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedBy   string    `json:"created_by" gorm:"type:uuid"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PentestProject ?
type PentestProject struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"not null"`
	Description   string    `json:"description"`
	ClientID      string    `json:"client_id" gorm:"type:uuid;not null"`
	Status        string    `json:"status" gorm:"default:'created'"` // created, planning, testing, reporting, completed
	Scope         JSONB     `json:"scope" gorm:"type:jsonb;not null"`
	Authorization JSONB     `json:"authorization" gorm:"type:jsonb;not null"`
	TeamMembers   JSONB     `json:"team_members" gorm:"type:jsonb"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PentestResult ?
type PentestResult struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID   string    `json:"project_id" gorm:"type:uuid;not null"`
	TestType    string    `json:"test_type" gorm:"not null"` // reconnaissance, scanning, exploitation, post_exploitation
	Target      string    `json:"target" gorm:"not null"`
	Method      string    `json:"method"`
	Success     bool      `json:"success"`
	Evidence    JSONB     `json:"evidence" gorm:"type:jsonb"`
	Impact      string    `json:"impact"`
	Remediation string    `json:"remediation"`
	TestedBy    string    `json:"tested_by" gorm:"type:uuid"`
	TestedAt    time.Time `json:"tested_at" gorm:"autoCreateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SecurityCourse 
type SecurityCourse struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title        string    `json:"title" gorm:"not null"`
	Description  string    `json:"description"`
	Category     string    `json:"category" gorm:"not null"` // basic, intermediate, advanced
	Difficulty   string    `json:"difficulty" gorm:"not null"` // beginner, intermediate, expert
	Duration     int       `json:"duration"` // 
	Content      JSONB     `json:"content" gorm:"type:jsonb"`
	Prerequisites JSONB    `json:"prerequisites" gorm:"type:jsonb"`
	Tags         JSONB     `json:"tags" gorm:"type:jsonb"`
	Published    bool      `json:"published" gorm:"default:false"`
	CreatedBy    string    `json:"created_by" gorm:"type:uuid"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// LabEnvironment 黷
type LabEnvironment struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"not null"` // web, network, malware, forensics
	Difficulty  string    `json:"difficulty" gorm:"not null"`
	Config      JSONB     `json:"config" gorm:"type:jsonb"`
	Resources   JSONB     `json:"resources" gorm:"type:jsonb"`
	Status      string    `json:"status" gorm:"default:'available'"` // available, running, maintenance
	UserID      *string   `json:"user_id" gorm:"type:uuid"`
	StartedAt   *time.Time `json:"started_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SecurityCertification 
type SecurityCertification struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null"`
	CourseID    string    `json:"course_id" gorm:"type:uuid;not null"`
	Type        string    `json:"type" gorm:"not null"` // completion, achievement, certification
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Score       *int      `json:"score"`
	MaxScore    *int      `json:"max_score"`
	ValidUntil  *time.Time `json:"valid_until"`
	IssuedAt    time.Time `json:"issued_at" gorm:"autoCreateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// AuditLog 
type AuditLog struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       *string   `json:"user_id" gorm:"type:uuid"`
	Action       string    `json:"action" gorm:"not null"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Details      JSONB     `json:"details" gorm:"type:jsonb"`
	Success      bool      `json:"success" gorm:"default:true"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SecurityEvent 
type SecurityEvent struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EventType   string    `json:"event_type" gorm:"not null"`
	Severity    string    `json:"severity" gorm:"not null"`
	SourceIP    string    `json:"source_ip"`
	TargetIP    string    `json:"target_ip"`
	UserID      *string   `json:"user_id" gorm:"type:uuid"`
	Description string    `json:"description"`
	RawData     JSONB     `json:"raw_data" gorm:"type:jsonb"`
	Processed   bool      `json:"processed" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ComplianceReport 汨
type ComplianceReport struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"not null"`
	Standard    string    `json:"standard" gorm:"not null"` // iso27001, nist, gdpr, etc.
	Version     string    `json:"version"`
	Status      string    `json:"status" gorm:"default:'draft'"` // draft, in_review, approved, published
	Summary     string    `json:"summary"`
	Findings    JSONB     `json:"findings" gorm:"type:jsonb"`
	Recommendations JSONB `json:"recommendations" gorm:"type:jsonb"`
	Score       *float64  `json:"score"`
	MaxScore    *float64  `json:"max_score"`
	GeneratedBy string    `json:"generated_by" gorm:"type:uuid"`
	ReviewedBy  *string   `json:"reviewed_by" gorm:"type:uuid"`
	ApprovedBy  *string   `json:"approved_by" gorm:"type:uuid"`
	GeneratedAt time.Time `json:"generated_at" gorm:"autoCreateTime"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserSecurityProfile 
type UserSecurityProfile struct {
	ID                string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID            string    `json:"user_id" gorm:"type:uuid;not null;unique"`
	SecurityClearance string    `json:"security_clearance" gorm:"default:'basic'"` // basic, intermediate, advanced
	PentestPermission bool      `json:"pentest_permission" gorm:"default:false"`
	EducationLevel    string    `json:"education_level" gorm:"default:'beginner'"` // beginner, intermediate, expert
	CertificationIDs  JSONB     `json:"certification_ids" gorm:"type:jsonb"`
	LastSecurityCheck *time.Time `json:"last_security_check"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ThreatIntelligence 鱨
type ThreatIntelligence struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IOCType     string    `json:"ioc_type" gorm:"not null"` // ip, domain, hash, url
	Value       string    `json:"value" gorm:"not null"`
	ThreatType  string    `json:"threat_type" gorm:"not null"` // malware, phishing, c2, etc.
	Severity    string    `json:"severity" gorm:"not null"`
	Confidence  int       `json:"confidence"` // 0-100
	Source      string    `json:"source"`
	Description string    `json:"description"`
	Tags        JSONB     `json:"tags" gorm:"type:jsonb"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

