package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
)

// LearnerRepository 定义学习者数据访问接�?
type LearnerRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, learner *entities.Learner) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Learner, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Learner, error)
	Update(ctx context.Context, learner *entities.Learner) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*entities.Learner, error)
	ListByLearningStyle(ctx context.Context, style entities.LearningStyle, offset, limit int) ([]*entities.Learner, error)
	ListByLearningPace(ctx context.Context, pace entities.LearningPace, offset, limit int) ([]*entities.Learner, error)

	// 学习目标相关
	GetLearnerGoals(ctx context.Context, learnerID uuid.UUID) ([]*entities.LearningGoal, error)
	AddLearnerGoal(ctx context.Context, learnerID uuid.UUID, goal *entities.LearningGoal) error
	UpdateLearnerGoal(ctx context.Context, learnerID uuid.UUID, goal *entities.LearningGoal) error
	RemoveLearnerGoal(ctx context.Context, learnerID uuid.UUID, goalID uuid.UUID) error
	GetActiveGoals(ctx context.Context, learnerID uuid.UUID) ([]*entities.LearningGoal, error)

	// 学习偏好相关
	GetLearnerPreferences(ctx context.Context, learnerID uuid.UUID) (*entities.LearningPreference, error)
	UpdateLearnerPreferences(ctx context.Context, learnerID uuid.UUID, preferences *entities.LearningPreference) error

	// 技能相�?
	GetLearnerSkills(ctx context.Context, learnerID uuid.UUID) (map[string]*entities.SkillLevel, error)
	AddOrUpdateSkill(ctx context.Context, learnerID uuid.UUID, skill string, level *entities.SkillLevel) error
	RemoveSkill(ctx context.Context, learnerID uuid.UUID, skill string) error
	GetSkillsByLevel(ctx context.Context, learnerID uuid.UUID, minLevel, maxLevel int) (map[string]*entities.SkillLevel, error)

	// 学习历史相关
	GetLearningHistory(ctx context.Context, learnerID uuid.UUID, limit int) ([]*entities.LearningHistory, error)
	AddLearningHistory(ctx context.Context, learnerID uuid.UUID, history *entities.LearningHistory) error
	RecordLearningActivity(ctx context.Context, learnerID uuid.UUID, history *entities.LearningHistory) error
	GetLearningHistoryByDateRange(ctx context.Context, learnerID uuid.UUID, startDate, endDate time.Time) ([]*entities.LearningHistory, error)
	GetLearningHistoryByContent(ctx context.Context, learnerID, contentID uuid.UUID) ([]*entities.LearningHistory, error)

	// 学习连续性相�?
	GetLearningStreaks(ctx context.Context, learnerID uuid.UUID) ([]*entities.LearningStreak, error)
	UpdateLearningStreak(ctx context.Context, learnerID uuid.UUID, streak *entities.LearningStreak) error
	GetCurrentStreak(ctx context.Context, learnerID uuid.UUID, streakType string) (*entities.LearningStreak, error)

	// 统计和分�?
	GetLearnerStatistics(ctx context.Context, learnerID uuid.UUID) (*LearnerStatistics, error)
	GetWeeklyProgress(ctx context.Context, learnerID uuid.UUID, weekStart time.Time) (*WeeklyProgress, error)
	GetLearningTrends(ctx context.Context, learnerID uuid.UUID, days int) (*LearningTrends, error)

	// 批量操作
	BatchCreate(ctx context.Context, learners []*entities.Learner) error
	BatchUpdate(ctx context.Context, learners []*entities.Learner) error

	// 搜索和过�?
	SearchLearners(ctx context.Context, query *LearnerSearchQuery) ([]*entities.Learner, int, error)
	GetSimilarLearners(ctx context.Context, learnerID uuid.UUID, limit int) ([]*entities.Learner, error)
}

// LearnerStatistics 学习者统计信�?
type LearnerStatistics struct {
	TotalLearningTime   time.Duration       `json:"total_learning_time"`
	CompletedContent    int                 `json:"completed_content"`
	ActiveGoals         int                 `json:"active_goals"`
	CompletedGoals      int                 `json:"completed_goals"`
	SkillCount          int                 `json:"skill_count"`
	AverageSkillLevel   float64             `json:"average_skill_level"`
	LongestStreak       int                 `json:"longest_streak"`
	CurrentStreaks      map[string]int      `json:"current_streaks"`
	LearningFrequency   float64             `json:"learning_frequency"` // 每周学习天数
	PreferredTimeSlots  []entities.TimeSlot `json:"preferred_time_slots"`
	SkillDistribution   map[string]int      `json:"skill_distribution"` // 按技能级别分�?
	ContentTypeProgress map[string]float64  `json:"content_type_progress"`
}

// WeeklyProgress 周学习进�?
type WeeklyProgress struct {
	WeekStart         time.Time                 `json:"week_start"`
	WeekEnd           time.Time                 `json:"week_end"`
	TotalLearningTime time.Duration             `json:"total_learning_time"`
	DaysActive        int                       `json:"days_active"`
	ContentCompleted  int                       `json:"content_completed"`
	SkillsImproved    []string                  `json:"skills_improved"`
	GoalsProgress     map[uuid.UUID]float64     `json:"goals_progress"`
	DailyProgress     map[time.Time]DayProgress `json:"daily_progress"`
}

// DayProgress 日学习进�?
type DayProgress struct {
	Date             time.Time     `json:"date"`
	LearningTime     time.Duration `json:"learning_time"`
	ContentCompleted int           `json:"content_completed"`
	SkillsWorkedOn   []string      `json:"skills_worked_on"`
	GoalsWorkedOn    []uuid.UUID   `json:"goals_worked_on"`
}

// LearningTrends 学习趋势
type LearningTrends struct {
	Period                 int                   `json:"period"` // 天数
	LearningTimetrend      TimeTrend             `json:"learning_time_trend"`
	ContentCompletionTrend CompletionTrend       `json:"content_completion_trend"`
	SkillProgressTrend     map[string]SkillTrend `json:"skill_progress_trend"`
	EngagementTrend        EngagementTrend       `json:"engagement_trend"`
	Predictions            TrendPredictions      `json:"predictions"`
}

// TimeTrend 时间趋势
type TimeTrend struct {
	DailyAverage    time.Duration   `json:"daily_average"`
	WeeklyAverage   time.Duration   `json:"weekly_average"`
	Trend           string          `json:"trend"` // "increasing", "decreasing", "stable"
	TrendPercentage float64         `json:"trend_percentage"`
	DailyData       []time.Duration `json:"daily_data"`
}

// CompletionTrend 完成度趋�?
type CompletionTrend struct {
	DailyAverage    float64   `json:"daily_average"`
	WeeklyAverage   float64   `json:"weekly_average"`
	Trend           string    `json:"trend"`
	TrendPercentage float64   `json:"trend_percentage"`
	DailyData       []float64 `json:"daily_data"`
}

// SkillTrend 技能趋�?
type SkillTrend struct {
	CurrentLevel      int     `json:"current_level"`
	PreviousLevel     int     `json:"previous_level"`
	Improvement       float64 `json:"improvement"`
	Trend             string  `json:"trend"`
	PracticeFrequency float64 `json:"practice_frequency"`
}

// EngagementTrend 参与度趋�?
type EngagementTrend struct {
	DailyEngagement    []float64 `json:"daily_engagement"`
	AverageEngagement  float64   `json:"average_engagement"`
	Trend              string    `json:"trend"`
	TrendPercentage    float64   `json:"trend_percentage"`
	PeakEngagementTime time.Time `json:"peak_engagement_time"`
}

// TrendPredictions 趋势预测
type TrendPredictions struct {
	NextWeekLearningTime time.Duration           `json:"next_week_learning_time"`
	NextWeekCompletion   float64                 `json:"next_week_completion"`
	GoalCompletionDates  map[uuid.UUID]time.Time `json:"goal_completion_dates"`
	SkillLevelUpDates    map[string]time.Time    `json:"skill_level_up_dates"`
	RiskFactors          []string                `json:"risk_factors"`
	Recommendations      []string                `json:"recommendations"`
}

// LearnerSearchQuery 学习者搜索查�?
type LearnerSearchQuery struct {
	// 搜索查询
	Query    string `json:"query,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	Language string `json:"language,omitempty"`
	
	// 基本过滤
	LearningStyle   *entities.LearningStyle `json:"learning_style,omitempty"`
	LearningPace    *entities.LearningPace  `json:"learning_pace,omitempty"`
	ExperienceLevel *int                    `json:"experience_level,omitempty"`

	// 技能过�?
	RequiredSkills []string `json:"required_skills,omitempty"`
	MinSkillLevel  *int     `json:"min_skill_level,omitempty"`
	MaxSkillLevel  *int     `json:"max_skill_level,omitempty"`

	// 活跃度过�?
	MinLearningTime  *time.Duration `json:"min_learning_time,omitempty"`
	MaxLearningTime  *time.Duration `json:"max_learning_time,omitempty"`
	LastActiveAfter  *time.Time     `json:"last_active_after,omitempty"`
	LastActiveBefore *time.Time     `json:"last_active_before,omitempty"`

	// 目标过滤
	HasActiveGoals *bool    `json:"has_active_goals,omitempty"`
	GoalCategories []string `json:"goal_categories,omitempty"`

	// 排序
	SortBy    string `json:"sort_by,omitempty"`    // "experience", "learning_time", "last_active", "skill_level"
	SortOrder string `json:"sort_order,omitempty"` // "asc", "desc"

	// 分页
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
