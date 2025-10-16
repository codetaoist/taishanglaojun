package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProjectStatus ?
type ProjectStatus string

const (
	ProjectStatusPlanning   ProjectStatus = "planning"   // 滮?
	ProjectStatusActive     ProjectStatus = "active"     // ?
	ProjectStatusOnHold     ProjectStatus = "on_hold"    // 
	ProjectStatusCompleted  ProjectStatus = "completed"  // ?
	ProjectStatusCancelled  ProjectStatus = "cancelled"  // ?
	ProjectStatusArchived   ProjectStatus = "archived"   // ?
)

// ProjectPriority ?
type ProjectPriority string

const (
	ProjectPriorityLow      ProjectPriority = "low"      // 
	ProjectPriorityMedium   ProjectPriority = "medium"   // 
	ProjectPriorityHigh     ProjectPriority = "high"     // 
	ProjectPriorityCritical ProjectPriority = "critical" // ?
)

// ProjectType 
type ProjectType string

const (
	ProjectTypeDevelopment ProjectType = "development" // ?
	ProjectTypeResearch    ProjectType = "research"    // 
	ProjectTypeMaintenance ProjectType = "maintenance" // 
	ProjectTypeInternal    ProjectType = "internal"    // 
	ProjectTypeClient      ProjectType = "client"      // 
)

// ProjectMember 
type ProjectMember struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Role      string    `json:"role" gorm:"type:varchar(50);not null"` // owner, manager, developer, tester, etc.
	JoinedAt  time.Time `json:"joined_at" gorm:"autoCreateTime"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
}

// ProjectMilestone ?
type ProjectMilestone struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ProjectID   uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	DueDate     time.Time `json:"due_date" gorm:"not null"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	IsCompleted bool      `json:"is_completed" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ProjectStatistics 
type ProjectStatistics struct {
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	InProgressTasks int     `json:"in_progress_tasks"`
	PendingTasks    int     `json:"pending_tasks"`
	OverdueTasks    int     `json:"overdue_tasks"`
	CompletionRate  float64 `json:"completion_rate"`
	AverageTaskTime float64 `json:"average_task_time"` // 
}

// Project ?
type Project struct {
	// 
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string          `json:"name" gorm:"type:varchar(255);not null;index"`
	Description string          `json:"description" gorm:"type:text"`
	Status      ProjectStatus   `json:"status" gorm:"type:varchar(20);not null;index;default:'planning'"`
	Priority    ProjectPriority `json:"priority" gorm:"type:varchar(20);not null;index;default:'medium'"`
	Type        ProjectType     `json:"type" gorm:"type:varchar(30);not null;index"`

	// 
	OwnerID       uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null;index"`
	TeamID        *uuid.UUID `json:"team_id,omitempty" gorm:"type:uuid;index"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`

	// 
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// ?
	Budget           *float64 `json:"budget,omitempty" gorm:"type:decimal(12,2)"`
	EstimatedHours   *float64 `json:"estimated_hours,omitempty" gorm:"type:decimal(8,2)"`
	ActualHours      *float64 `json:"actual_hours,omitempty" gorm:"type:decimal(8,2)"`

	// 
	Tags     []string               `json:"tags" gorm:"type:text[]"`
	Labels   map[string]string      `json:"labels" gorm:"type:jsonb"`
	Metadata map[string]interface{} `json:"metadata" gorm:"type:jsonb"`

	// 
	Progress float64 `json:"progress" gorm:"type:decimal(5,2);default:0"` // 0-100

	// 
	Members    []ProjectMember    `json:"members,omitempty" gorm:"foreignKey:ProjectID"`
	Milestones []ProjectMilestone `json:"milestones,omitempty" gorm:"foreignKey:ProjectID"`

	// 
	domainEvents []DomainEvent `json:"-" gorm:"-"`
}

// NewProject ?
func NewProject(name, description string, projectType ProjectType, 
	priority ProjectPriority, ownerID, organizationID uuid.UUID) (*Project, error) {
	
	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}
	
	if ownerID == uuid.Nil {
		return nil, errors.New("owner ID cannot be empty")
	}
	
	if organizationID == uuid.Nil {
		return nil, errors.New("organization ID cannot be empty")
	}

	project := &Project{
		ID:             uuid.New(),
		Name:           name,
		Description:    description,
		Status:         ProjectStatusPlanning,
		Priority:       priority,
		Type:           projectType,
		OwnerID:        ownerID,
		OrganizationID: organizationID,
		Progress:       0.0,
		Tags:           make([]string, 0),
		Labels:         make(map[string]string),
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 
	ownerMember := ProjectMember{
		ID:        uuid.New(),
		ProjectID: project.ID,
		UserID:    ownerID,
		Role:      "owner",
		JoinedAt:  time.Now(),
		IsActive:  true,
	}
	project.Members = append(project.Members, ownerMember)

	// 
	event := &ProjectCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: project.ID,
			EventType:   "ProjectCreated",
			OccurredAt:  time.Now(),
		},
		ProjectID:      project.ID,
		Name:           project.Name,
		Type:           project.Type,
		Priority:       project.Priority,
		OwnerID:        project.OwnerID,
		OrganizationID: project.OrganizationID,
	}
	project.AddDomainEvent(event)

	return project, nil
}

// Start 
func (p *Project) Start(userID uuid.UUID) error {
	if p.Status != ProjectStatusPlanning {
		return errors.New("project can only be started from planning status")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to start project")
	}

	p.Status = ProjectStatusActive
	if p.StartDate == nil {
		now := time.Now()
		p.StartDate = &now
	}
	p.UpdatedAt = time.Now()

	// 
	event := &ProjectStartedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: p.ID,
			EventType:   "ProjectStarted",
			OccurredAt:  time.Now(),
		},
		ProjectID: p.ID,
		UserID:    userID,
	}
	p.AddDomainEvent(event)

	return nil
}

// Complete 
func (p *Project) Complete(userID uuid.UUID) error {
	if p.Status != ProjectStatusActive {
		return errors.New("project must be active to be completed")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to complete project")
	}

	p.Status = ProjectStatusCompleted
	now := time.Now()
	p.CompletedAt = &now
	p.Progress = 100.0
	p.UpdatedAt = time.Now()

	// 
	event := &ProjectCompletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: p.ID,
			EventType:   "ProjectCompleted",
			OccurredAt:  time.Now(),
		},
		ProjectID:   p.ID,
		UserID:      userID,
		CompletedAt: now,
	}
	p.AddDomainEvent(event)

	return nil
}

// Cancel 
func (p *Project) Cancel(userID uuid.UUID, reason string) error {
	if p.Status == ProjectStatusCompleted || p.Status == ProjectStatusCancelled {
		return errors.New("cannot cancel completed or already cancelled project")
	}

	if !p.isUserAuthorized(userID, []string{"owner"}) {
		return errors.New("only project owner can cancel project")
	}

	p.Status = ProjectStatusCancelled
	p.UpdatedAt = time.Now()

	// 
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata["cancellation_reason"] = reason
	p.Metadata["cancelled_by"] = userID
	p.Metadata["cancelled_at"] = time.Now()

	// 
	event := &ProjectCancelledEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: p.ID,
			EventType:   "ProjectCancelled",
			OccurredAt:  time.Now(),
		},
		ProjectID: p.ID,
		UserID:    userID,
		Reason:    reason,
	}
	p.AddDomainEvent(event)

	return nil
}

// Pause 
func (p *Project) Pause(userID uuid.UUID, reason string) error {
	if p.Status != ProjectStatusActive {
		return errors.New("only active projects can be paused")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to pause project")
	}

	p.Status = ProjectStatusOnHold
	p.UpdatedAt = time.Now()

	// 
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata["pause_reason"] = reason
	p.Metadata["paused_by"] = userID
	p.Metadata["paused_at"] = time.Now()

	return nil
}

// Resume 
func (p *Project) Resume(userID uuid.UUID) error {
	if p.Status != ProjectStatusOnHold {
		return errors.New("only paused projects can be resumed")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to resume project")
	}

	p.Status = ProjectStatusActive
	p.UpdatedAt = time.Now()

	// ?
	if p.Metadata != nil {
		delete(p.Metadata, "pause_reason")
		delete(p.Metadata, "paused_by")
		delete(p.Metadata, "paused_at")
	}

	return nil
}

// AddMember 
func (p *Project) AddMember(userID uuid.UUID, role string, addedBy uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if !p.isUserAuthorized(addedBy, []string{"owner", "manager"}) {
		return errors.New("user not authorized to add members")
	}

	// 
	for _, member := range p.Members {
		if member.UserID == userID && member.IsActive {
			return errors.New("user is already a member of this project")
		}
	}

	member := ProjectMember{
		ID:        uuid.New(),
		ProjectID: p.ID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
		IsActive:  true,
	}

	p.Members = append(p.Members, member)
	p.UpdatedAt = time.Now()

	// 
	event := &ProjectMemberAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: p.ID,
			EventType:   "ProjectMemberAdded",
			OccurredAt:  time.Now(),
		},
		ProjectID: p.ID,
		UserID:    userID,
		Role:      role,
		AddedBy:   addedBy,
	}
	p.AddDomainEvent(event)

	return nil
}

// RemoveMember 
func (p *Project) RemoveMember(userID uuid.UUID, removedBy uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if !p.isUserAuthorized(removedBy, []string{"owner", "manager"}) {
		return errors.New("user not authorized to remove members")
	}

	// ?
	if userID == p.OwnerID {
		return errors.New("cannot remove project owner")
	}

	// ?
	for i := range p.Members {
		if p.Members[i].UserID == userID && p.Members[i].IsActive {
			now := time.Now()
			p.Members[i].IsActive = false
			p.Members[i].LeftAt = &now
			p.UpdatedAt = time.Now()

			// 
			event := &ProjectMemberRemovedEvent{
				BaseDomainEvent: BaseDomainEvent{
					EventID:     uuid.New(),
					AggregateID: p.ID,
					EventType:   "ProjectMemberRemoved",
					OccurredAt:  time.Now(),
				},
				ProjectID: p.ID,
				UserID:    userID,
				RemovedBy: removedBy,
			}
			p.AddDomainEvent(event)

			return nil
		}
	}

	return errors.New("user is not a member of this project")
}

// UpdateMemberRole 
func (p *Project) UpdateMemberRole(userID uuid.UUID, newRole string, updatedBy uuid.UUID) error {
	if !p.isUserAuthorized(updatedBy, []string{"owner"}) {
		return errors.New("only project owner can update member roles")
	}

	// 
	if userID == p.OwnerID {
		return errors.New("cannot change project owner role")
	}

	// ?
	for i := range p.Members {
		if p.Members[i].UserID == userID && p.Members[i].IsActive {
			oldRole := p.Members[i].Role
			p.Members[i].Role = newRole
			p.UpdatedAt = time.Now()

			// 
			event := &ProjectMemberRoleUpdatedEvent{
				BaseDomainEvent: BaseDomainEvent{
					EventID:     uuid.New(),
					AggregateID: p.ID,
					EventType:   "ProjectMemberRoleUpdated",
					OccurredAt:  time.Now(),
				},
				ProjectID: p.ID,
				UserID:    userID,
				OldRole:   oldRole,
				NewRole:   newRole,
				UpdatedBy: updatedBy,
			}
			p.AddDomainEvent(event)

			return nil
		}
	}

	return errors.New("user is not a member of this project")
}

// AddMilestone ?
func (p *Project) AddMilestone(title, description string, dueDate time.Time, userID uuid.UUID) error {
	if title == "" {
		return errors.New("milestone title cannot be empty")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to add milestones")
	}

	milestone := ProjectMilestone{
		ID:          uuid.New(),
		ProjectID:   p.ID,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		IsCompleted: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	p.Milestones = append(p.Milestones, milestone)
	p.UpdatedAt = time.Now()

	return nil
}

// CompleteMilestone ?
func (p *Project) CompleteMilestone(milestoneID uuid.UUID, userID uuid.UUID) error {
	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to complete milestones")
	}

	for i := range p.Milestones {
		if p.Milestones[i].ID == milestoneID {
			if p.Milestones[i].IsCompleted {
				return errors.New("milestone is already completed")
			}

			now := time.Now()
			p.Milestones[i].IsCompleted = true
			p.Milestones[i].CompletedAt = &now
			p.Milestones[i].UpdatedAt = now
			p.UpdatedAt = time.Now()

			return nil
		}
	}

	return errors.New("milestone not found")
}

// UpdateProgress 
func (p *Project) UpdateProgress(progress float64, userID uuid.UUID) error {
	if progress < 0 || progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to update progress")
	}

	oldProgress := p.Progress
	p.Progress = progress
	p.UpdatedAt = time.Now()

	// 100%
	if progress == 100.0 && p.Status == ProjectStatusActive {
		// 
	}

	// 
	event := &ProjectProgressUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: p.ID,
			EventType:   "ProjectProgressUpdated",
			OccurredAt:  time.Now(),
		},
		ProjectID:   p.ID,
		UserID:      userID,
		OldProgress: oldProgress,
		NewProgress: progress,
	}
	p.AddDomainEvent(event)

	return nil
}

// SetBudget 
func (p *Project) SetBudget(budget float64, userID uuid.UUID) error {
	if budget < 0 {
		return errors.New("budget cannot be negative")
	}

	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to set budget")
	}

	p.Budget = &budget
	p.UpdatedAt = time.Now()

	return nil
}

// SetDueDate 
func (p *Project) SetDueDate(dueDate *time.Time, userID uuid.UUID) error {
	if !p.isUserAuthorized(userID, []string{"owner", "manager"}) {
		return errors.New("user not authorized to set due date")
	}

	p.DueDate = dueDate
	p.UpdatedAt = time.Now()

	return nil
}

// IsOverdue 
func (p *Project) IsOverdue() bool {
	if p.DueDate == nil || p.Status == ProjectStatusCompleted || p.Status == ProjectStatusCancelled {
		return false
	}
	return time.Now().After(*p.DueDate)
}

// GetActiveMemberCount 
func (p *Project) GetActiveMemberCount() int {
	count := 0
	for _, member := range p.Members {
		if member.IsActive {
			count++
		}
	}
	return count
}

// GetCompletedMilestoneCount 
func (p *Project) GetCompletedMilestoneCount() int {
	count := 0
	for _, milestone := range p.Milestones {
		if milestone.IsCompleted {
			count++
		}
	}
	return count
}

// isUserAuthorized 
func (p *Project) isUserAuthorized(userID uuid.UUID, allowedRoles []string) bool {
	for _, member := range p.Members {
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
func (p *Project) GetUserRole(userID uuid.UUID) (string, bool) {
	for _, member := range p.Members {
		if member.UserID == userID && member.IsActive {
			return member.Role, true
		}
	}
	return "", false
}

// AddTag 
func (p *Project) AddTag(tag string) {
	if tag == "" {
		return
	}

	// 
	for _, existingTag := range p.Tags {
		if existingTag == tag {
			return
		}
	}

	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now()
}

// RemoveTag 
func (p *Project) RemoveTag(tag string) {
	for i, existingTag := range p.Tags {
		if existingTag == tag {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			p.UpdatedAt = time.Now()
			break
		}
	}
}

// SetLabel 
func (p *Project) SetLabel(key, value string) {
	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	p.Labels[key] = value
	p.UpdatedAt = time.Now()
}

// RemoveLabel 
func (p *Project) RemoveLabel(key string) {
	if p.Labels != nil {
		delete(p.Labels, key)
		p.UpdatedAt = time.Now()
	}
}

// SetMetadata ?
func (p *Project) SetMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
	p.UpdatedAt = time.Now()
}

// GetMetadata ?
func (p *Project) GetMetadata(key string) (interface{}, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	value, exists := p.Metadata[key]
	return value, exists
}

// 
func (p *Project) AddDomainEvent(event DomainEvent) {
	p.domainEvents = append(p.domainEvents, event)
}

func (p *Project) GetDomainEvents() []DomainEvent {
	return p.domainEvents
}

func (p *Project) ClearDomainEvents() {
	p.domainEvents = nil
}

