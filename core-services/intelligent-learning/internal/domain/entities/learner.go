package entities

import (
	"time"

	"github.com/google/uuid"
)

// LearningStyle 学习风格
type LearningStyle string

const (
	LearningStyleVisual     LearningStyle = "visual"     // 视觉�?
	LearningStyleAuditory   LearningStyle = "auditory"   // 听觉�?
	LearningStyleKinesthetic LearningStyle = "kinesthetic" // 动觉�?
	LearningStyleReading    LearningStyle = "reading"    // 阅读�?
)

// LearningPace 学习节奏
type LearningPace string

const (
	LearningPaceSlow   LearningPace = "slow"   // 慢节�?
	LearningPaceMedium LearningPace = "medium" // 中等节奏
	LearningPaceFast   LearningPace = "fast"   // 快节�?
)

// LearningGoal 学习目标
type LearningGoal struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TargetSkill string    `json:"target_skill"` // 目标技�?
	TargetDate  time.Time `json:"target_date"`
	TargetLevel int       `json:"target_level"` // 目标等级 1-10
	Priority    int       `json:"priority"`     // 1-10�?0为最高优先级
	IsActive    bool      `json:"is_active"`    // 是否激�?
	Achieved    bool      `json:"achieved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LearningPreference 学习偏好
type LearningPreference struct {
	Style               LearningStyle `json:"style"`
	Pace                LearningPace  `json:"pace"`
	PreferredTimeSlots  []TimeSlot    `json:"preferred_time_slots"`
	SessionDuration     int           `json:"session_duration"`     // 分钟
	BreakDuration       int           `json:"break_duration"`       // 分钟
	DifficultyTolerance float64       `json:"difficulty_tolerance"` // 0.0-1.0
	InteractiveContent  bool          `json:"interactive_content"`
	MultimediaContent   bool          `json:"multimedia_content"`
}

// TimeSlot 时间�?
type TimeSlot struct {
	StartHour int `json:"start_hour"` // 0-23
	EndHour   int `json:"end_hour"`   // 0-23
	DayOfWeek int `json:"day_of_week"` // 0-6�?为周�?
}

// LearningHistory 学习历史
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

// SkillLevel 技能水�?
type SkillLevel struct {
	SkillID     uuid.UUID `json:"skill_id"`
	SkillName   string    `json:"skill_name"`
	Level       int       `json:"level"`       // 1-10
	Experience  int       `json:"experience"`  // 经验�?
	Confidence  float64   `json:"confidence"`  // 0.0-1.0
	LastUpdated time.Time `json:"last_updated"`
}

// LearningStreak 学习连续�?
type LearningStreak struct {
	CurrentStreak int       `json:"current_streak"` // 当前连续天数
	LongestStreak int       `json:"longest_streak"` // 最长连续天�?
	LastStudyDate time.Time `json:"last_study_date"`
	TotalDays     int       `json:"total_days"` // 总学习天�?
}

// Learner 学习者实�?
type Learner struct {
	ID               uuid.UUID            `json:"id"`
	UserID           uuid.UUID            `json:"user_id"`
	Name             string               `json:"name"`
	Email            string               `json:"email"`
	Age              int                  `json:"age"`              // 年龄
	EducationLevel   string               `json:"education_level"`  // 教育水平
	LearningStyle    string               `json:"learning_style"`   // 学习风格
	AvatarURL        string               `json:"avatar_url"`
	Bio              string               `json:"bio"`
	Timezone         string               `json:"timezone"`
	Language         string               `json:"language"`
	Level            int                  `json:"level"`            // 学习者等�?
	Experience       int                  `json:"experience"`       // 总经验�?
	ExperienceLevel  DifficultyLevel      `json:"experience_level"` // 经验等级
	LearningGoals    []LearningGoal       `json:"learning_goals"`
	Preferences      LearningPreference   `json:"preferences"`
	Skills           []SkillLevel         `json:"skills"`
	LearningHistory  []LearningHistory    `json:"learning_history"`
	Streak           LearningStreak       `json:"streak"`
	Achievements     []uuid.UUID          `json:"achievements"`     // 成就ID列表
	CurrentPaths     []uuid.UUID          `json:"current_paths"`    // 当前学习路径ID列表
	CompletedPaths   []uuid.UUID          `json:"completed_paths"`  // 已完成学习路径ID列表
	WeeklyGoalHours  int                  `json:"weekly_goal_hours"` // 每周学习目标小时�?
	TotalStudyHours  int                  `json:"total_study_hours"` // 总学习小时数
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// NewLearner 创建新的学习�?
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
		WeeklyGoalHours:  10, // 默认每周10小时
		TotalStudyHours:  0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// getDefaultPreferences 获取默认学习偏好
func getDefaultPreferences() LearningPreference {
	return LearningPreference{
		Style:               LearningStyleVisual,
		Pace:                LearningPaceMedium,
		PreferredTimeSlots:  getDefaultTimeSlots(),
		SessionDuration:     45, // 45分钟
		BreakDuration:       15, // 15分钟
		DifficultyTolerance: 0.7,
		InteractiveContent:  true,
		MultimediaContent:   true,
	}
}

// getDefaultTimeSlots 获取默认时间�?
func getDefaultTimeSlots() []TimeSlot {
	return []TimeSlot{
		{StartHour: 9, EndHour: 11, DayOfWeek: 1}, // 周一 9-11�?
		{StartHour: 9, EndHour: 11, DayOfWeek: 3}, // 周三 9-11�?
		{StartHour: 9, EndHour: 11, DayOfWeek: 5}, // 周五 9-11�?
	}
}

// AddGoal 添加学习目标
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

// UpdatePreferences 更新学习偏好
func (l *Learner) UpdatePreferences(preferences LearningPreference) {
	l.Preferences = preferences
	l.UpdatedAt = time.Now()
}

// AddSkill 添加或更新技�?
func (l *Learner) AddSkill(skillID uuid.UUID, skillName string, level int, experience int, confidence float64) {
	// 查找是否已存在该技�?
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

	// 添加新技�?
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

// AddLearningHistory 添加学习历史记录
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
	l.TotalStudyHours += duration / 3600 // 转换为小�?
	l.updateStreak()
	l.UpdatedAt = time.Now()
}

// updateStreak 更新学习连续�?
func (l *Learner) updateStreak() {
	today := time.Now().Truncate(24 * time.Hour)
	lastStudy := l.Streak.LastStudyDate.Truncate(24 * time.Hour)

	if lastStudy.Equal(today) {
		// 今天已经学习过了，不需要更�?
		return
	}

	if lastStudy.Equal(today.Add(-24 * time.Hour)) {
		// 昨天学习过，连续天数+1
		l.Streak.CurrentStreak++
	} else if lastStudy.Before(today.Add(-24 * time.Hour)) {
		// 中断了，重新开�?
		l.Streak.CurrentStreak = 1
	}

	// 更新最长连续天�?
	if l.Streak.CurrentStreak > l.Streak.LongestStreak {
		l.Streak.LongestStreak = l.Streak.CurrentStreak
	}

	l.Streak.LastStudyDate = time.Now()
	l.Streak.TotalDays++
}

// AddExperience 增加经验�?
func (l *Learner) AddExperience(exp int) {
	l.Experience += exp
	
	// 检查是否升�?
	newLevel := l.calculateLevel(l.Experience)
	if newLevel > l.Level {
		l.Level = newLevel
		// 这里可以触发升级事件
	}
	
	l.UpdatedAt = time.Now()
}

// calculateLevel 根据经验值计算等�?
func (l *Learner) calculateLevel(experience int) int {
	// 简单的等级计算公式：每1000经验值升一�?
	return (experience / 1000) + 1
}

// GetSkillLevel 获取特定技能的等级
func (l *Learner) GetSkillLevel(skillID uuid.UUID) *SkillLevel {
	for _, skill := range l.Skills {
		if skill.SkillID == skillID {
			return &skill
		}
	}
	return nil
}

// GetActiveGoals 获取未完成的目标
func (l *Learner) GetActiveGoals() []LearningGoal {
	var activeGoals []LearningGoal
	for _, goal := range l.LearningGoals {
		if !goal.Achieved {
			activeGoals = append(activeGoals, goal)
		}
	}
	return activeGoals
}

// GetRecentHistory 获取最近的学习历史
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

// CalculateWeeklyProgress 计算本周学习进度
func (l *Learner) CalculateWeeklyProgress() float64 {
	if l.WeeklyGoalHours == 0 {
		return 0
	}

	// 获取本周开始时�?
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	// 计算本周学习时间
	var weeklyHours int
	for _, history := range l.LearningHistory {
		if history.CreatedAt.After(weekStart) {
			weeklyHours += int(history.Duration.Seconds()) / 3600
		}
	}

	return float64(weeklyHours) / float64(l.WeeklyGoalHours)
}

// Skill 技能定�?
type Skill struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`       // 1-10
	Category    string    `json:"category"`    // 技能分�?
	Description string    `json:"description"` // 技能描�?
	AcquiredAt  time.Time `json:"acquired_at"` // 获得时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}
