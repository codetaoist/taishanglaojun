package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TaskStatus ?
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"     // ?
	TaskStatusAssigned   TaskStatus = "assigned"    // ?
	TaskStatusInProgress TaskStatus = "in_progress" // ?
	TaskStatusCompleted  TaskStatus = "completed"   // ?
	TaskStatusCancelled  TaskStatus = "cancelled"   // ?
	TaskStatusOnHold     TaskStatus = "on_hold"     // 
	TaskStatusOverdue    TaskStatus = "overdue"     // 
)

// TaskPriority ?
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"      // 
	TaskPriorityMedium   TaskPriority = "medium"   // 
	TaskPriorityHigh     TaskPriority = "high"     // 
	TaskPriorityCritical TaskPriority = "critical" // ?
)

// TaskType 
type TaskType string

const (
	TaskTypeDevelopment TaskType = "development" // ?
	TaskTypeBug         TaskType = "bug"         // 
	TaskTypeFeature     TaskType = "feature"     // ?
	TaskTypeResearch    TaskType = "research"    // 
	TaskTypeMaintenance TaskType = "maintenance" // 
	TaskTypeReview      TaskType = "review"      // 
	TaskTypeTesting     TaskType = "testing"     // 
	TaskTypeDocumentation TaskType = "documentation" // 
)

// TaskComplexity ?
type TaskComplexity string

const (
	TaskComplexitySimple   TaskComplexity = "simple"   // ?
	TaskComplexityModerate TaskComplexity = "moderate" // 
	TaskComplexityComplex  TaskComplexity = "complex"  // 
	TaskComplexityExpert   TaskComplexity = "expert"   // ?
)

// TaskDependency 
type TaskDependency struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TaskID       uuid.UUID `json:"task_id" gorm:"type:uuid;not null;index"`
	DependsOnID  uuid.UUID `json:"depends_on_id" gorm:"type:uuid;not null;index"`
	DependencyType string  `json:"dependency_type" gorm:"type:varchar(50);not null"` // finish_to_start, start_to_start, etc.
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TaskAssignment 
type TaskAssignment struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TaskID     uuid.UUID  `json:"task_id" gorm:"type:uuid;not null;index"`
	AssigneeID uuid.UUID  `json:"assignee_id" gorm:"type:uuid;not null;index"`
	AssignerID uuid.UUID  `json:"assigner_id" gorm:"type:uuid;not null"`
	AssignedAt time.Time  `json:"assigned_at" gorm:"autoCreateTime"`
	UnassignedAt *time.Time `json:"unassigned_at,omitempty"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
}

// TaskComment 
type TaskComment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TaskID    uuid.UUID `json:"task_id" gorm:"type:uuid;not null;index"`
	AuthorID  uuid.UUID `json:"author_id" gorm:"type:uuid;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TaskAttachment 
type TaskAttachment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TaskID    uuid.UUID `json:"task_id" gorm:"type:uuid;not null;index"`
	FileName  string    `json:"file_name" gorm:"type:varchar(255);not null"`
	FileURL   string    `json:"file_url" gorm:"type:varchar(500);not null"`
	FileSize  int64     `json:"file_size" gorm:"not null"`
	MimeType  string    `json:"mime_type" gorm:"type:varchar(100)"`
	UploadedBy uuid.UUID `json:"uploaded_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TaskTimeLog 
type TaskTimeLog struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TaskID      uuid.UUID  `json:"task_id" gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	StartTime   time.Time  `json:"start_time" gorm:"not null"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Duration    int64      `json:"duration" gorm:"default:0"` // ?
	Description string     `json:"description" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// Task ?
type Task struct {
	// 
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Title       string       `json:"title" gorm:"type:varchar(255);not null;index"`
	Description string       `json:"description" gorm:"type:text"`
	Status      TaskStatus   `json:"status" gorm:"type:varchar(20);not null;index;default:'pending'"`
	Priority    TaskPriority `json:"priority" gorm:"type:varchar(20);not null;index;default:'medium'"`
	Type        TaskType     `json:"type" gorm:"type:varchar(30);not null;index"`
	Complexity  TaskComplexity `json:"complexity" gorm:"type:varchar(20);not null;default:'moderate'"`

	// ?
	ProjectID    uuid.UUID  `json:"project_id" gorm:"type:uuid;not null;index"`
	TeamID       *uuid.UUID `json:"team_id,omitempty" gorm:"type:uuid;index"`
	CreatorID    uuid.UUID  `json:"creator_id" gorm:"type:uuid;not null;index"`
	AssigneeID   *uuid.UUID `json:"assignee_id,omitempty" gorm:"type:uuid;index"`
	ReviewerID   *uuid.UUID `json:"reviewer_id,omitempty" gorm:"type:uuid;index"`

	// 
	EstimatedHours   *float64   `json:"estimated_hours,omitempty" gorm:"type:decimal(8,2)"`
	ActualHours      *float64   `json:"actual_hours,omitempty" gorm:"type:decimal(8,2)"`
	StartDate        *time.Time `json:"start_date,omitempty"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 
	Tags     []string               `json:"tags" gorm:"type:text[]"`
	Labels   map[string]string      `json:"labels" gorm:"type:jsonb"`
	Metadata map[string]interface{} `json:"metadata" gorm:"type:jsonb"`

	// ?
	Progress      float64 `json:"progress" gorm:"type:decimal(5,2);default:0"` // 0-100
	QualityScore  *float64 `json:"quality_score,omitempty" gorm:"type:decimal(3,2)"` // 0-10
	
	// 
	Dependencies []TaskDependency `json:"dependencies,omitempty" gorm:"foreignKey:TaskID"`
	Assignments  []TaskAssignment `json:"assignments,omitempty" gorm:"foreignKey:TaskID"`
	Comments     []TaskComment    `json:"comments,omitempty" gorm:"foreignKey:TaskID"`
	Attachments  []TaskAttachment `json:"attachments,omitempty" gorm:"foreignKey:TaskID"`
	TimeLogs     []TaskTimeLog    `json:"time_logs,omitempty" gorm:"foreignKey:TaskID"`

	// 
	domainEvents []DomainEvent `json:"-" gorm:"-"`
}

// NewTask ?
func NewTask(title, description string, taskType TaskType, priority TaskPriority, 
	complexity TaskComplexity, projectID, creatorID uuid.UUID) (*Task, error) {
	
	if title == "" {
		return nil, errors.New("task title cannot be empty")
	}
	
	if projectID == uuid.Nil {
		return nil, errors.New("project ID cannot be empty")
	}
	
	if creatorID == uuid.Nil {
		return nil, errors.New("creator ID cannot be empty")
	}

	task := &Task{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Status:      TaskStatusPending,
		Priority:    priority,
		Type:        taskType,
		Complexity:  complexity,
		ProjectID:   projectID,
		CreatorID:   creatorID,
		Progress:    0.0,
		Tags:        make([]string, 0),
		Labels:      make(map[string]string),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 
	event := &TaskCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: task.ID,
			EventType:   "TaskCreated",
			OccurredAt:  time.Now(),
		},
		TaskID:      task.ID,
		Title:       task.Title,
		Type:        task.Type,
		Priority:    task.Priority,
		ProjectID:   task.ProjectID,
		CreatorID:   task.CreatorID,
	}
	task.AddDomainEvent(event)

	return task, nil
}

// AssignTo ?
func (t *Task) AssignTo(assigneeID, assignerID uuid.UUID) error {
	if assigneeID == uuid.Nil {
		return errors.New("assignee ID cannot be empty")
	}
	
	if assignerID == uuid.Nil {
		return errors.New("assigner ID cannot be empty")
	}

	// ?
	if t.AssigneeID != nil && *t.AssigneeID == assigneeID {
		return nil
	}

	// ?
	if t.AssigneeID != nil {
		t.unassignCurrent()
	}

	t.AssigneeID = &assigneeID
	t.Status = TaskStatusAssigned
	t.UpdatedAt = time.Now()

	// 
	assignment := TaskAssignment{
		ID:         uuid.New(),
		TaskID:     t.ID,
		AssigneeID: assigneeID,
		AssignerID: assignerID,
		AssignedAt: time.Now(),
		IsActive:   true,
	}
	t.Assignments = append(t.Assignments, assignment)

	// 
	event := &TaskAssignedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskAssigned",
			OccurredAt:  time.Now(),
		},
		TaskID:     t.ID,
		AssigneeID: assigneeID,
		AssignerID: assignerID,
	}
	t.AddDomainEvent(event)

	return nil
}

// Unassign 
func (t *Task) Unassign() {
	if t.AssigneeID == nil {
		return
	}

	t.unassignCurrent()
	t.AssigneeID = nil
	t.Status = TaskStatusPending
	t.UpdatedAt = time.Now()

	// 
	event := &TaskUnassignedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskUnassigned",
			OccurredAt:  time.Now(),
		},
		TaskID: t.ID,
	}
	t.AddDomainEvent(event)
}

// unassignCurrent 
func (t *Task) unassignCurrent() {
	now := time.Now()
	for i := range t.Assignments {
		if t.Assignments[i].IsActive {
			t.Assignments[i].IsActive = false
			t.Assignments[i].UnassignedAt = &now
		}
	}
}

// Start ?
func (t *Task) Start(userID uuid.UUID) error {
	if t.Status != TaskStatusAssigned && t.Status != TaskStatusPending {
		return errors.New("task cannot be started in current status")
	}

	if t.AssigneeID == nil || *t.AssigneeID != userID {
		return errors.New("only assigned user can start the task")
	}

	t.Status = TaskStatusInProgress
	if t.StartDate == nil {
		now := time.Now()
		t.StartDate = &now
	}
	t.UpdatedAt = time.Now()

	// ?
	event := &TaskStartedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskStarted",
			OccurredAt:  time.Now(),
		},
		TaskID: t.ID,
		UserID: userID,
	}
	t.AddDomainEvent(event)

	return nil
}

// Complete 
func (t *Task) Complete(userID uuid.UUID) error {
	if t.Status != TaskStatusInProgress {
		return errors.New("task must be in progress to be completed")
	}

	if t.AssigneeID == nil || *t.AssigneeID != userID {
		return errors.New("only assigned user can complete the task")
	}

	t.Status = TaskStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
	t.Progress = 100.0
	t.UpdatedAt = time.Now()

	// 
	event := &TaskCompletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskCompleted",
			OccurredAt:  time.Now(),
		},
		TaskID:      t.ID,
		UserID:      userID,
		CompletedAt: now,
	}
	t.AddDomainEvent(event)

	return nil
}

// Cancel 
func (t *Task) Cancel(userID uuid.UUID, reason string) error {
	if t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return errors.New("cannot cancel completed or already cancelled task")
	}

	t.Status = TaskStatusCancelled
	t.UpdatedAt = time.Now()

	// 
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata["cancellation_reason"] = reason
	t.Metadata["cancelled_by"] = userID
	t.Metadata["cancelled_at"] = time.Now()

	// 
	event := &TaskCancelledEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskCancelled",
			OccurredAt:  time.Now(),
		},
		TaskID: t.ID,
		UserID: userID,
		Reason: reason,
	}
	t.AddDomainEvent(event)

	return nil
}

// UpdateProgress 
func (t *Task) UpdateProgress(progress float64, userID uuid.UUID) error {
	if progress < 0 || progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}

	if t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return errors.New("cannot update progress of completed or cancelled task")
	}

	oldProgress := t.Progress
	t.Progress = progress
	t.UpdatedAt = time.Now()

	// 100%?
	if progress == 100.0 && t.Status == TaskStatusInProgress {
		t.Complete(userID)
	}

	// 
	event := &TaskProgressUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskProgressUpdated",
			OccurredAt:  time.Now(),
		},
		TaskID:      t.ID,
		UserID:      userID,
		OldProgress: oldProgress,
		NewProgress: progress,
	}
	t.AddDomainEvent(event)

	return nil
}

// SetPriority ?
func (t *Task) SetPriority(priority TaskPriority, userID uuid.UUID) error {
	if t.Priority == priority {
		return nil
	}

	oldPriority := t.Priority
	t.Priority = priority
	t.UpdatedAt = time.Now()

	// ?
	event := &TaskPriorityUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskPriorityUpdated",
			OccurredAt:  time.Now(),
		},
		TaskID:      t.ID,
		UserID:      userID,
		OldPriority: oldPriority,
		NewPriority: priority,
	}
	t.AddDomainEvent(event)

	return nil
}

// SetDueDate 
func (t *Task) SetDueDate(dueDate *time.Time, userID uuid.UUID) error {
	t.DueDate = dueDate
	t.UpdatedAt = time.Now()

	// 
	event := &TaskDueDateUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskDueDateUpdated",
			OccurredAt:  time.Now(),
		},
		TaskID:  t.ID,
		UserID:  userID,
		DueDate: dueDate,
	}
	t.AddDomainEvent(event)

	return nil
}

// AddComment 
func (t *Task) AddComment(content string, authorID uuid.UUID) error {
	if content == "" {
		return errors.New("comment content cannot be empty")
	}

	comment := TaskComment{
		ID:        uuid.New(),
		TaskID:    t.ID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Comments = append(t.Comments, comment)
	t.UpdatedAt = time.Now()

	// 
	event := &TaskCommentAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventID:     uuid.New(),
			AggregateID: t.ID,
			EventType:   "TaskCommentAdded",
			OccurredAt:  time.Now(),
		},
		TaskID:    t.ID,
		CommentID: comment.ID,
		AuthorID:  authorID,
		Content:   content,
	}
	t.AddDomainEvent(event)

	return nil
}

// AddAttachment 
func (t *Task) AddAttachment(fileName, fileURL string, fileSize int64, 
	mimeType string, uploadedBy uuid.UUID) error {
	
	if fileName == "" || fileURL == "" {
		return errors.New("file name and URL cannot be empty")
	}

	attachment := TaskAttachment{
		ID:         uuid.New(),
		TaskID:     t.ID,
		FileName:   fileName,
		FileURL:    fileURL,
		FileSize:   fileSize,
		MimeType:   mimeType,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}

	t.Attachments = append(t.Attachments, attachment)
	t.UpdatedAt = time.Now()

	return nil
}

// AddTimeLog 
func (t *Task) AddTimeLog(userID uuid.UUID, startTime time.Time, 
	endTime *time.Time, description string) error {
	
	timeLog := TaskTimeLog{
		ID:          uuid.New(),
		TaskID:      t.ID,
		UserID:      userID,
		StartTime:   startTime,
		EndTime:     endTime,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 
	if endTime != nil {
		duration := endTime.Sub(startTime)
		timeLog.Duration = int64(duration.Seconds())
	}

	t.TimeLogs = append(t.TimeLogs, timeLog)
	t.UpdatedAt = time.Now()

	return nil
}

// AddDependency 
func (t *Task) AddDependency(dependsOnID uuid.UUID, dependencyType string) error {
	if dependsOnID == uuid.Nil {
		return errors.New("dependency task ID cannot be empty")
	}

	if dependsOnID == t.ID {
		return errors.New("task cannot depend on itself")
	}

	// 
	for _, dep := range t.Dependencies {
		if dep.DependsOnID == dependsOnID && dep.DependencyType == dependencyType {
			return nil // ?
		}
	}

	dependency := TaskDependency{
		ID:             uuid.New(),
		TaskID:         t.ID,
		DependsOnID:    dependsOnID,
		DependencyType: dependencyType,
		CreatedAt:      time.Now(),
	}

	t.Dependencies = append(t.Dependencies, dependency)
	t.UpdatedAt = time.Now()

	return nil
}

// IsOverdue 
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// GetEstimatedDuration 
func (t *Task) GetEstimatedDuration() time.Duration {
	if t.EstimatedHours == nil {
		return 0
	}
	return time.Duration(*t.EstimatedHours * float64(time.Hour))
}

// GetActualDuration 
func (t *Task) GetActualDuration() time.Duration {
	if t.ActualHours == nil {
		return 0
	}
	return time.Duration(*t.ActualHours * float64(time.Hour))
}

// CalculateActualHours ?
func (t *Task) CalculateActualHours() float64 {
	var totalSeconds int64
	for _, timeLog := range t.TimeLogs {
		totalSeconds += timeLog.Duration
	}
	return float64(totalSeconds) / 3600.0 // ?
}

// UpdateActualHours ?
func (t *Task) UpdateActualHours() {
	actualHours := t.CalculateActualHours()
	t.ActualHours = &actualHours
	t.UpdatedAt = time.Now()
}

// AddTag 
func (t *Task) AddTag(tag string) {
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
func (t *Task) RemoveTag(tag string) {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// SetLabel 
func (t *Task) SetLabel(key, value string) {
	if t.Labels == nil {
		t.Labels = make(map[string]string)
	}
	t.Labels[key] = value
	t.UpdatedAt = time.Now()
}

// RemoveLabel 
func (t *Task) RemoveLabel(key string) {
	if t.Labels != nil {
		delete(t.Labels, key)
		t.UpdatedAt = time.Now()
	}
}

// SetMetadata ?
func (t *Task) SetMetadata(key string, value interface{}) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata[key] = value
	t.UpdatedAt = time.Now()
}

// GetMetadata ?
func (t *Task) GetMetadata(key string) (interface{}, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	value, exists := t.Metadata[key]
	return value, exists
}

// 
func (t *Task) AddDomainEvent(event DomainEvent) {
	t.domainEvents = append(t.domainEvents, event)
}

func (t *Task) GetDomainEvents() []DomainEvent {
	return t.domainEvents
}

func (t *Task) ClearDomainEvents() {
	t.domainEvents = nil
}

