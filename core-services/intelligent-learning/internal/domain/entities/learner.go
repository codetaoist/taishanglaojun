package entities

import (
	"time"

	"github.com/google/uuid"
)

// LearningStyle 
type LearningStyle string

const (
	LearningStyleVisual     LearningStyle = "visual"     // ?
	LearningStyleAuditory   LearningStyle = "auditory"   // ?
	LearningStyleKinesthetic LearningStyle = "kinesthetic" // ?
	LearningStyleReading    LearningStyle = "reading"    // ?
)

// LearningPace 
type LearningPace string

const (
	LearningPaceSlow   LearningPace = "slow"   // ?
	LearningPaceMedium LearningPace = "medium" // 
	LearningPaceFast   LearningPace = "fast"   // ?
)

// LearningGoal 
type LearningGoal struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TargetSkill string    `json:"target_skill"` // ?
	TargetDate  time.Time `json:"target_date"`
	TargetLevel int       `json:"target_level"` //  1-10
	Priority    int       `json:"priority"`     // 1-10?0
	IsActive    bool      `json:"is_active"`    // ?
	Achieved    bool      `json:"achieved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LearningPreference 
type LearningPreference struct {
	Style               LearningStyle `json:"style"`
	Pace                LearningPace  `json:"pace"`
	PreferredTimeSlots  []TimeSlot    `json:"preferred_time_slots"`
	SessionDuration     int           `json:"session_duration"`     // 
	BreakDuration       int           `json:"break_duration"`       // 
	DifficultyTolerance float64       `json:"difficulty_tolerance"` // 0.0-1.0
	InteractiveContent  bool          `json:"interactive_content"`
	MultimediaContent   bool          `json:"multimedia_content"`
}

// TimeSlot ?
type TimeSlot struct {
	StartHour int `json:"start_hour"` // 0-23
	EndHour   int `json:"end_hour"`   // 0-23
	DayOfWeek int `json:"day_of_week"` // 0-6??
}

// LearningHistory 
type LearningHistory struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	LearnerID      uuid.UUID              `json:"learner_id" gorm:"type:uuid;not null;index"`
	ContentID      uuid.UUID              `json:"content_id" gorm:"type:uuid;not null;index"`
	ContentType    string                 `json:"content_type" gorm:"type:varchar(50);not null"`
	ContentTitle   string                 `json:"content_title" gorm:"type:varchar(200)"`
	SkillName      string                 `json:"skill_name" gorm:"type:varchar(100);index"`
	DifficultyLevel DifficultyLevel       `json:"difficulty_level" gorm:"type:varchar(20);default:'beginner'"`
	StartTime      time.Time              `json:"start_time" gorm:"not null"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Duration       time.Duration          `json:"duration" gorm:"type:bigint"`
	Progress       float64                `json:"progress" gorm:"type:decimal(5,2);default:0"`
	Score          *float64               `json:"score,omitempty" gorm:"type:decimal(5,2)"`
	Completed      bool                   `json:"completed" gorm:"default:false"`
	Interactions   map[string]interface{} `json:"interactions" gorm:"type:jsonb"`
	Timestamp      time.Time              `json:"timestamp" gorm:"not null;index"`
	CreatedAt      time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// SkillLevel ?
type SkillLevel struct {
	SkillID     uuid.UUID `json:"skill_id"`
	SkillName   string    `json:"skill_name"`
	Level       int       `json:"level"`       // 1-10
	Experience  int       `json:"experience"`  // ?
	Confidence  float64   `json:"confidence"`  // 0.0-1.0
	LastUpdated time.Time `json:"last_updated"`
}

// LearningStreak ?
type LearningStreak struct {
	CurrentStreak int       `json:"current_streak"` // 
	LongestStreak int       `json:"longest_streak"` // ?
	LastStudyDate time.Time `json:"last_study_date"`
	TotalDays     int       `json:"total_days"` // ?
}

// Learner ?
type Learner struct {
	ID               uuid.UUID            `json:"id"`
	UserID           uuid.UUID            `json:"user_id"`
	Name             string               `json:"name"`
	Email            string               `json:"email"`
	Age              int                  `json:"age"`              // 
	EducationLevel   string               `json:"education_level"`  // 
	LearningStyle    string               `json:"learning_style"`   // 
	AvatarURL        string               `json:"avatar_url"`
	Bio              string               `json:"bio"`
	Timezone         string               `json:"timezone"`
	Language         string               `json:"language"`
	Level            int                  `json:"level"`            // ?
	Experience       int                  `json:"experience"`       // ?
	ExperienceLevel  DifficultyLevel      `json:"experience_level"` // 
	LearningGoals    []LearningGoal       `json:"learning_goals"`
	Preferences      LearningPreference   `json:"preferences"`
	Skills           []SkillLevel         `json:"skills"`
	LearningHistory  []LearningHistory    `json:"learning_history"`
	Streak           LearningStreak       `json:"streak"`
	Achievements     []uuid.UUID          `json:"achievements"`     // ID
	CurrentPaths     []uuid.UUID          `json:"current_paths"`    // ID
	CompletedPaths   []uuid.UUID          `json:"completed_paths"`  // ID
	WeeklyGoalHours  int                  `json:"weekly_goal_hours"` // ?
	TotalStudyHours  int                  `json:"total_study_hours"` // 
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// NewLearner ?
func NewLearner(userID uuid.UUID, name, email string) *Learner {
	now := time.Now()
	return &Learner{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            name,
		Email:           email,
		Level:           1,
		Experience:      0,
		ExperienceLevel: DifficultyBeginner,
		LearningGoals:   make([]LearningGoal, 0),
		Preferences:     getDefaultPreferences(),
		Skills:          make([]SkillLevel, 0),
		LearningHistory: make([]LearningHistory, 0),
		Streak: LearningStreak{
			CurrentStreak: 0,
			LongestStreak: 0,
			TotalDays:     0,
		},
		Achievements:     make([]uuid.UUID, 0),
		CurrentPaths:     make([]uuid.UUID, 0),
		CompletedPaths:   make([]uuid.UUID, 0),
		WeeklyGoalHours:  10, // 10
		TotalStudyHours:  0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// getDefaultPreferences 
func getDefaultPreferences() LearningPreference {
	return LearningPreference{
		Style:               LearningStyleVisual,
		Pace:                LearningPaceMedium,
		PreferredTimeSlots:  getDefaultTimeSlots(),
		SessionDuration:     45, // 45
		BreakDuration:       15, // 15
		DifficultyTolerance: 0.7,
		InteractiveContent:  true,
		MultimediaContent:   true,
	}
}

// getDefaultTimeSlots ?
func getDefaultTimeSlots() []TimeSlot {
	return []TimeSlot{
		{StartHour: 9, EndHour: 11, DayOfWeek: 1}, //  9-11?
		{StartHour: 9, EndHour: 11, DayOfWeek: 3}, //  9-11?
		{StartHour: 9, EndHour: 11, DayOfWeek: 5}, //  9-11?
	}
}

// AddGoal 
func (l *Learner) AddGoal(title, description string, targetDate time.Time, priority int) {
	goal := LearningGoal{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		TargetDate:  targetDate,
		Priority:    priority,
		Achieved:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	l.LearningGoals = append(l.LearningGoals, goal)
	l.UpdatedAt = time.Now()
}

// UpdatePreferences 
func (l *Learner) UpdatePreferences(preferences LearningPreference) {
	l.Preferences = preferences
	l.UpdatedAt = time.Now()
}

// AddSkill ?
func (l *Learner) AddSkill(skillID uuid.UUID, skillName string, level int, experience int, confidence float64) {
	// ?
	for i, skill := range l.Skills {
		if skill.SkillID == skillID {
			l.Skills[i].Level = level
			l.Skills[i].Experience = experience
			l.Skills[i].Confidence = confidence
			l.Skills[i].LastUpdated = time.Now()
			l.UpdatedAt = time.Now()
			return
		}
	}

	// ?
	skill := SkillLevel{
		SkillID:     skillID,
		SkillName:   skillName,
		Level:       level,
		Experience:  experience,
		Confidence:  confidence,
		LastUpdated: time.Now(),
	}
	l.Skills = append(l.Skills, skill)
	l.UpdatedAt = time.Now()
}

// AddLearningHistory 
func (l *Learner) AddLearningHistory(contentID uuid.UUID, contentType string, duration int, progress float64, score *float64, completed bool, interactions int) {
	history := LearningHistory{
		ID:           uuid.New(),
		LearnerID:    l.ID,
		ContentID:    contentID,
		ContentType:  contentType,
		StartTime:    time.Now(),
		Duration:     time.Duration(duration) * time.Second,
		Progress:     progress,
		Score:        score,
		Completed:    completed,
		Interactions: map[string]interface{}{"count": interactions},
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}

	if completed {
		endTime := time.Now()
		history.EndTime = &endTime
	}

	l.LearningHistory = append(l.LearningHistory, history)
	l.TotalStudyHours += duration / 3600 // ?
	l.updateStreak()
	l.UpdatedAt = time.Now()
}

// updateStreak ?
func (l *Learner) updateStreak() {
	today := time.Now().Truncate(24 * time.Hour)
	lastStudy := l.Streak.LastStudyDate.Truncate(24 * time.Hour)

	if lastStudy.Equal(today) {
		// ?
		return
	}

	if lastStudy.Equal(today.Add(-24 * time.Hour)) {
		// +1
		l.Streak.CurrentStreak++
	} else if lastStudy.Before(today.Add(-24 * time.Hour)) {
		// ?
		l.Streak.CurrentStreak = 1
	}

	// ?
	if l.Streak.CurrentStreak > l.Streak.LongestStreak {
		l.Streak.LongestStreak = l.Streak.CurrentStreak
	}

	l.Streak.LastStudyDate = time.Now()
	l.Streak.TotalDays++
}

// AddExperience ?
func (l *Learner) AddExperience(exp int) {
	l.Experience += exp
	
	// ?
	newLevel := l.calculateLevel(l.Experience)
	if newLevel > l.Level {
		l.Level = newLevel
		// 
	}
	
	l.UpdatedAt = time.Now()
}

// calculateLevel ?
func (l *Learner) calculateLevel(experience int) int {
	// 㹫1000?
	return (experience / 1000) + 1
}

// GetSkillLevel 
func (l *Learner) GetSkillLevel(skillID uuid.UUID) *SkillLevel {
	for _, skill := range l.Skills {
		if skill.SkillID == skillID {
			return &skill
		}
	}
	return nil
}

// GetActiveGoals 
func (l *Learner) GetActiveGoals() []LearningGoal {
	var activeGoals []LearningGoal
	for _, goal := range l.LearningGoals {
		if !goal.Achieved {
			activeGoals = append(activeGoals, goal)
		}
	}
	return activeGoals
}

// GetRecentHistory 
func (l *Learner) GetRecentHistory(days int) []LearningHistory {
	cutoff := time.Now().AddDate(0, 0, -days)
	var recentHistory []LearningHistory
	
	for _, history := range l.LearningHistory {
		if history.CreatedAt.After(cutoff) {
			recentHistory = append(recentHistory, history)
		}
	}
	
	return recentHistory
}

// CalculateWeeklyProgress 㱾
func (l *Learner) CalculateWeeklyProgress() float64 {
	if l.WeeklyGoalHours == 0 {
		return 0
	}

	// ?
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	// 㱾
	var weeklyHours int
	for _, history := range l.LearningHistory {
		if history.CreatedAt.After(weekStart) {
			weeklyHours += int(history.Duration.Seconds()) / 3600
		}
	}

	return float64(weeklyHours) / float64(l.WeeklyGoalHours)
}

// Skill ?
type Skill struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`       // 1-10
	Category    string    `json:"category"`    // ?
	Description string    `json:"description"` // ?
	AcquiredAt  time.Time `json:"acquired_at"` // 
	UpdatedAt   time.Time `json:"updated_at"`  // 
}

