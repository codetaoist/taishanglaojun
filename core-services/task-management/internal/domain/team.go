package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TeamStatus ?
type TeamStatus string

const (
	TeamStatusActive    TeamStatus = "active"    // 
	TeamStatusInactive  TeamStatus = "inactive"  // ?
	TeamStatusDisbanded TeamStatus = "disbanded" // ?
)

// TeamMemberRole 
type TeamMemberRole string

const (
	TeamMemberRoleLeader    TeamMemberRole = "leader"    // ?
	TeamMemberRoleMember    TeamMemberRole = "member"    // ?
	TeamMemberRoleMentor    TeamMemberRole = "mentor"    // 
	TeamMemberRoleIntern    TeamMemberRole = "intern"    // ?
	TeamMemberRoleConsultant TeamMemberRole = "consultant" // 
)

// TeamMember 
type TeamMember struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TeamID    uuid.UUID      `json:"team_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Role      TeamMemberRole `json:"role" gorm:"type:varchar(20);not null"`
	JoinedAt  time.Time      `json:"joined_at" gorm:"autoCreateTime"`
	LeftAt    *time.Time     `json:"left_at,omitempty"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	
	// 
	Skills       []string               `json:"skills" gorm:"type:text[]"`
	Specialties  []string               `json:"specialties" gorm:"type:text[]"`
	Availability map[string]interface{} `json:"availability" gorm:"type:jsonb"` // ?
	
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TeamSkill ?
type TeamSkill struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TeamID      uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	SkillName   string    `json:"skill_name" gorm:"type:varchar(100);not null"`
	Level       string    `json:"level" gorm:"type:varchar(20);not null"` // beginner, intermediate, advanced, expert
	MemberCount int       `json:"member_count" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TeamMetrics 
type TeamMetrics struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TeamID            uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	Period            string    `json:"period" gorm:"type:varchar(20);not null"` // daily, weekly, monthly
	Date              time.Time `json:"date" gorm:"not null;index"`
	
	// 
	TasksCompleted    int     `json:"tasks_completed" gorm:"default:0"`
	TasksInProgress   int     `json:"tasks_in_progress" gorm:"default:0"`
	TasksOverdue      int     `json:"tasks_overdue" gorm:"default:0"`
	AverageTaskTime   float64 `json:"average_task_time" gorm:"type:decimal(8,2);default:0"`
	
	// 
	QualityScore      float64 `json:"quality_score" gorm:"type:decimal(3,2);default:0"`
	BugRate           float64 `json:"bug_rate" gorm:"type:decimal(5,4);default:0"`
	ReworkRate        float64 `json:"rework_rate" gorm:"type:decimal(5,4);default:0"`
	
	// 
	CollaborationScore float64 `json:"collaboration_score" gorm:"type:decimal(3,2);default:0"`
	CommunicationScore float64 `json:"communication_score" gorm:"type:decimal(3,2);default:0"`
	
	// 
	Productivity      float64 `json:"productivity" gorm:"type:decimal(5,2);default:0"`
	Velocity          float64 `json:"velocity" gorm:"type:decimal(5,2);default:0"`
	
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Team ?
type Team struct {
	// 
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null;index"`
	Description string     `json:"description" gorm:"type:text"`
	Status      TeamStatus `json:"status" gorm:"type:varchar(20);not null;index;default:'active'"`
	
	// 
	LeaderID       uuid.UUID `json:"leader_id" gorm:"type:uuid;not null;index"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
	ParentTeamID   *uuid.UUID `json:"parent_team_id,omitempty" gorm:"type:uuid;index"`
	
	// 
	MaxMembers     *int    `json:"max_members,omitempty" gorm:"default:null"`
	TimeZone       string  `json:"time_zone" gorm:"type:varchar(50);default:'UTC'"`
	WorkingHours   string  `json:"working_hours" gorm:"type:varchar(100)"` // JSON?
	
	// 
	Tags     []string               `json:"tags" gorm:"type:text[]"`
	Labels   map[string]string      `json:"labels" gorm:"type:jsonb"`
	Metadata map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	
	// 
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	
	// 
	Members []TeamMember  `json:"members,omitempty" gorm:"foreignKey:TeamID"`
	Skills  []TeamSkill   `json:"skills,omitempty" gorm:"foreignKey:TeamID"`
	Metrics []TeamMetrics `json:"metrics,omitempty" gorm:"foreignKey:TeamID"`
	
	// 
	domainEvents []DomainEvent `json:"-" gorm:"-"`
}

// NewTeam ?
func NewTeam(name, description string, leaderID, organizationID uuid.UUID) (*Team, error) {
	if name == "" {
		return nil, errors.New("team name cannot be empty")
	}
	
	if leaderID == uuid.Nil {
		return nil, errors.New("leader ID cannot be empty")
	}
	
	if organizationID == uuid.Nil {
		return nil, errors.New("organization ID cannot be empty")
	}

	team := &Team{
		ID:             uuid.New(),
		Name:           name,
		Description:    description,
		Status:         TeamStatusActive,
		LeaderID:       leaderID,
		OrganizationID: organizationID,
		TimeZone:       "UTC",
		Tags:           make([]string, 0),
		Labels:         make(map[string]string),
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 
	leaderMember := TeamMember{
		ID:       uuid.New(),
		TeamID:   team.ID,
		UserID:   leaderID,
		Role:     TeamMemberRoleLeader,
		JoinedAt: time.Now(),
		IsActive: true,
		Skills:   make([]string, 0),
		Specialties: make([]string, 0),
		Availability: make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	team.Members = append(team.Members, leaderMember)

	// 
	event := &TeamCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: team.ID,
			EventType:   "TeamCreated",
			OccurredAt:  time.Now(),
		},
		TeamID:         team.ID,
		Name:           team.Name,
		LeaderID:       team.LeaderID,
		OrganizationID: team.OrganizationID,
	}
	team.AddDomainEvent(event)

	return team, nil
}

// AddMember 
func (t *Team) AddMember(userID uuid.UUID, role TeamMemberRole, addedBy uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	// 
	if !t.isUserAuthorized(addedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("user not authorized to add members")
	}

	// 
	for _, member := range t.Members {
		if member.UserID == userID && member.IsActive {
			return errors.New("user is already a member of this team")
		}
	}

	// ?
	if t.MaxMembers != nil && t.GetActiveMemberCount() >= *t.MaxMembers {
		return errors.New("team has reached maximum member limit")
	}

	member := TeamMember{
		ID:           uuid.New(),
		TeamID:       t.ID,
		UserID:       userID,
		Role:         role,
		JoinedAt:     time.Now(),
		IsActive:     true,
		Skills:       make([]string, 0),
		Specialties:  make([]string, 0),
		Availability: make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Members = append(t.Members, member)
	t.UpdatedAt = time.Now()

	// 
	event := &TeamMemberAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TeamMemberAdded",
			OccurredAt:  time.Now(),
		},
		TeamID:  t.ID,
		UserID:  userID,
		Role:    string(role),
		AddedBy: addedBy,
	}
	t.AddDomainEvent(event)

	return nil
}

// RemoveMember 
func (t *Team) RemoveMember(userID uuid.UUID, removedBy uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	// 
	if !t.isUserAuthorized(removedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("user not authorized to remove members")
	}

	// ?
	if userID == t.LeaderID {
		return errors.New("cannot remove team leader")
	}

	// ?
	for i := range t.Members {
		if t.Members[i].UserID == userID && t.Members[i].IsActive {
			now := time.Now()
			t.Members[i].IsActive = false
			t.Members[i].LeftAt = &now
			t.Members[i].UpdatedAt = now
			t.UpdatedAt = time.Now()

			// 
			event := &TeamMemberRemovedEvent{
				BaseDomainEvent: BaseDomainEvent{
					EventID:     uuid.New(),
					AggregateID: t.ID,
					EventType:   "TeamMemberRemoved",
					OccurredAt:  time.Now(),
				},
				TeamID:    t.ID,
				UserID:    userID,
				RemovedBy: removedBy,
			}
			t.AddDomainEvent(event)

			return nil
		}
	}

	return errors.New("user is not a member of this team")
}

// UpdateMemberRole 
func (t *Team) UpdateMemberRole(userID uuid.UUID, newRole TeamMemberRole, updatedBy uuid.UUID) error {
	// 
	if !t.isUserAuthorized(updatedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("only team leader can update member roles")
	}

	// 
	if userID == t.LeaderID {
		return errors.New("cannot change team leader role")
	}

	// ?
	for i := range t.Members {
		if t.Members[i].UserID == userID && t.Members[i].IsActive {
			t.Members[i].Role = newRole
			t.Members[i].UpdatedAt = time.Now()
			t.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("user is not a member of this team")
}

// ChangeLeader ?
func (t *Team) ChangeLeader(newLeaderID uuid.UUID, changedBy uuid.UUID) error {
	if newLeaderID == uuid.Nil {
		return errors.New("new leader ID cannot be empty")
	}

	// 
	if changedBy != t.LeaderID {
		// ?
		return errors.New("only current leader can change team leader")
	}

	// 
	var newLeaderMember *TeamMember
	for i := range t.Members {
		if t.Members[i].UserID == newLeaderID && t.Members[i].IsActive {
			newLeaderMember = &t.Members[i]
			break
		}
	}

	if newLeaderMember == nil {
		return errors.New("new leader must be a team member")
	}

	// 
	for i := range t.Members {
		if t.Members[i].UserID == t.LeaderID && t.Members[i].IsActive {
			t.Members[i].Role = TeamMemberRoleMember
			t.Members[i].UpdatedAt = time.Now()
			break
		}
	}

	// 
	newLeaderMember.Role = TeamMemberRoleLeader
	newLeaderMember.UpdatedAt = time.Now()

	// ?
	t.LeaderID = newLeaderID
	t.UpdatedAt = time.Now()

	return nil
}

// UpdateMemberSkills ?
func (t *Team) UpdateMemberSkills(userID uuid.UUID, skills []string, updatedBy uuid.UUID) error {
	// ?
	if updatedBy != userID && !t.isUserAuthorized(updatedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("not authorized to update member skills")
	}

	// ?
	for i := range t.Members {
		if t.Members[i].UserID == userID && t.Members[i].IsActive {
			t.Members[i].Skills = skills
			t.Members[i].UpdatedAt = time.Now()
			t.UpdatedAt = time.Now()

			// ?
			t.updateTeamSkills()
			return nil
		}
	}

	return errors.New("user is not a member of this team")
}

// UpdateMemberAvailability ?
func (t *Team) UpdateMemberAvailability(userID uuid.UUID, availability map[string]interface{}, updatedBy uuid.UUID) error {
	// ?
	if updatedBy != userID && !t.isUserAuthorized(updatedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("not authorized to update member availability")
	}

	// ?
	for i := range t.Members {
		if t.Members[i].UserID == userID && t.Members[i].IsActive {
			t.Members[i].Availability = availability
			t.Members[i].UpdatedAt = time.Now()
			t.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("user is not a member of this team")
}

// Disband 
func (t *Team) Disband(disbandedBy uuid.UUID, reason string) error {
	// ?
	if !t.isUserAuthorized(disbandedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("only team leader can disband team")
	}

	if t.Status == TeamStatusDisbanded {
		return errors.New("team is already disbanded")
	}

	t.Status = TeamStatusDisbanded
	now := time.Now()
	t.ArchivedAt = &now
	t.UpdatedAt = time.Now()

	// 
	for i := range t.Members {
		if t.Members[i].IsActive {
			t.Members[i].IsActive = false
			t.Members[i].LeftAt = &now
			t.Members[i].UpdatedAt = now
		}
	}

	// 
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata["disbandment_reason"] = reason
	t.Metadata["disbanded_by"] = disbandedBy
	t.Metadata["disbanded_at"] = now

	// 
	event := &TeamDisbandedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TeamDisbanded",
			OccurredAt:  time.Now(),
		},
		TeamID:      t.ID,
		DisbandedBy: disbandedBy,
		Reason:      reason,
	}
	t.AddDomainEvent(event)

	return nil
}

// SetMaxMembers 
func (t *Team) SetMaxMembers(maxMembers int, updatedBy uuid.UUID) error {
	if maxMembers < 1 {
		return errors.New("max members must be at least 1")
	}

	if !t.isUserAuthorized(updatedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("only team leader can set max members")
	}

	// 鵱
	currentMemberCount := t.GetActiveMemberCount()
	if currentMemberCount > maxMembers {
		return errors.New("current member count exceeds new limit")
	}

	t.MaxMembers = &maxMembers
	t.UpdatedAt = time.Now()

	return nil
}

// SetWorkingHours 
func (t *Team) SetWorkingHours(workingHours string, updatedBy uuid.UUID) error {
	if !t.isUserAuthorized(updatedBy, []TeamMemberRole{TeamMemberRoleLeader}) {
		return errors.New("only team leader can set working hours")
	}

	t.WorkingHours = workingHours
	t.UpdatedAt = time.Now()

	return nil
}

// GetActiveMemberCount 
func (t *Team) GetActiveMemberCount() int {
	count := 0
	for _, member := range t.Members {
		if member.IsActive {
			count++
		}
	}
	return count
}

// GetMembersByRole ?
func (t *Team) GetMembersByRole(role TeamMemberRole) []TeamMember {
	var members []TeamMember
	for _, member := range t.Members {
		if member.IsActive && member.Role == role {
			members = append(members, member)
		}
	}
	return members
}

// GetMemberSkills ?
func (t *Team) GetMemberSkills(userID uuid.UUID) ([]string, error) {
	for _, member := range t.Members {
		if member.UserID == userID && member.IsActive {
			return member.Skills, nil
		}
	}
	return nil, errors.New("user is not a member of this team")
}

// GetTeamSkillCoverage ?
func (t *Team) GetTeamSkillCoverage() map[string]int {
	skillCount := make(map[string]int)
	for _, member := range t.Members {
		if member.IsActive {
			for _, skill := range member.Skills {
				skillCount[skill]++
			}
		}
	}
	return skillCount
}

// HasSkill ?
func (t *Team) HasSkill(skill string) bool {
	for _, member := range t.Members {
		if member.IsActive {
			for _, memberSkill := range member.Skills {
				if memberSkill == skill {
					return true
				}
			}
		}
	}
	return false
}

// GetAvailableMembers 
func (t *Team) GetAvailableMembers(timeSlot string) []TeamMember {
	var availableMembers []TeamMember
	for _, member := range t.Members {
		if member.IsActive {
			// ?
			// 
			availableMembers = append(availableMembers, member)
		}
	}
	return availableMembers
}

// updateTeamSkills ?
func (t *Team) updateTeamSkills() {
	skillCount := t.GetTeamSkillCoverage()
	
	// ?
	t.Skills = nil
	
	// ?
	for skill, count := range skillCount {
		teamSkill := TeamSkill{
			ID:          uuid.New(),
			TeamID:      t.ID,
			SkillName:   skill,
			Level:       "intermediate", // ?
			MemberCount: count,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		t.Skills = append(t.Skills, teamSkill)
	}
}

// isUserAuthorized 
func (t *Team) isUserAuthorized(userID uuid.UUID, allowedRoles []TeamMemberRole) bool {
	for _, member := range t.Members {
		if member.UserID == userID && member.IsActive {
			for _, role := range allowedRoles {
				if member.Role == role {
					return true
				}
			}
		}
	}
	return false
}

// GetUserRole ?
func (t *Team) GetUserRole(userID uuid.UUID) (TeamMemberRole, bool) {
	for _, member := range t.Members {
		if member.UserID == userID && member.IsActive {
			return member.Role, true
		}
	}
	return "", false
}

// IsUserMember 
func (t *Team) IsUserMember(userID uuid.UUID) bool {
	for _, member := range t.Members {
		if member.UserID == userID && member.IsActive {
			return true
		}
	}
	return false
}

// AddTag 
func (t *Team) AddTag(tag string) {
	if tag == "" {
		return
	}

	// 
	for _, existingTag := range t.Tags {
		if existingTag == tag {
			return
		}
	}

	t.Tags = append(t.Tags, tag)
	t.UpdatedAt = time.Now()
}

// RemoveTag 
func (t *Team) RemoveTag(tag string) {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// SetLabel 
func (t *Team) SetLabel(key, value string) {
	if t.Labels == nil {
		t.Labels = make(map[string]string)
	}
	t.Labels[key] = value
	t.UpdatedAt = time.Now()
}

// RemoveLabel 
func (t *Team) RemoveLabel(key string) {
	if t.Labels != nil {
		delete(t.Labels, key)
		t.UpdatedAt = time.Now()
	}
}

// SetMetadata ?
func (t *Team) SetMetadata(key string, value interface{}) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata[key] = value
	t.UpdatedAt = time.Now()
}

// GetMetadata ?
func (t *Team) GetMetadata(key string) (interface{}, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	value, exists := t.Metadata[key]
	return value, exists
}

// 
func (t *Team) AddDomainEvent(event DomainEvent) {
	t.domainEvents = append(t.domainEvents, event)
}

func (t *Team) GetDomainEvents() []DomainEvent {
	return t.domainEvents
}

func (t *Team) ClearDomainEvents() {
	t.domainEvents = nil
}

