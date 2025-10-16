package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Achievement 成就实体
type Achievement struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Level       string    `json:"level"`
	Icon        string    `json:"icon"`
	Points      int       `json:"points"`
	Criteria    map[string]interface{} `json:"criteria"`
	Rewards     []map[string]interface{} `json:"rewards"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LearnerAchievement 学习者成就实?
type LearnerAchievement struct {
	ID            uuid.UUID   `json:"id"`
	LearnerID     uuid.UUID   `json:"learner_id"`
	AchievementID uuid.UUID   `json:"achievement_id"`
	Achievement   *Achievement `json:"achievement,omitempty"`
	Progress      float64     `json:"progress"`
	IsUnlocked    bool        `json:"is_unlocked"`
	UnlockedAt    *time.Time  `json:"unlocked_at,omitempty"`
	CurrentValue  float64     `json:"current_value"`
	TargetValue   float64     `json:"target_value"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// AchievementRepository 成就数据访问接口
type AchievementRepository interface {
	// 成就管理
	CreateAchievement(ctx context.Context, achievement *Achievement) error
	GetAchievementByID(ctx context.Context, id uuid.UUID) (*Achievement, error)
	UpdateAchievement(ctx context.Context, achievement *Achievement) error
	DeleteAchievement(ctx context.Context, id uuid.UUID) error
	ListAchievements(ctx context.Context, offset, limit int) ([]*Achievement, error)
	GetAchievementsByType(ctx context.Context, achievementType string) ([]*Achievement, error)
	GetActiveAchievements(ctx context.Context) ([]*Achievement, error)

	// 学习者成就管?
	CreateLearnerAchievement(ctx context.Context, learnerAchievement *LearnerAchievement) error
	GetLearnerAchievement(ctx context.Context, learnerID, achievementID uuid.UUID) (*LearnerAchievement, error)
	UpdateLearnerAchievement(ctx context.Context, learnerAchievement *LearnerAchievement) error
	GetLearnerAchievements(ctx context.Context, learnerID uuid.UUID, offset, limit int) ([]*LearnerAchievement, error)
	GetLearnerAchievementsByType(ctx context.Context, learnerID uuid.UUID, achievementType string) ([]*LearnerAchievement, error)
	GetLearnerAchievementsByStatus(ctx context.Context, learnerID uuid.UUID, isUnlocked bool) ([]*LearnerAchievement, error)

	// 成就统计
	GetLearnerAchievementCount(ctx context.Context, learnerID uuid.UUID) (int, error)
	GetLearnerUnlockedAchievementCount(ctx context.Context, learnerID uuid.UUID) (int, error)
	GetLearnerTotalPoints(ctx context.Context, learnerID uuid.UUID) (int, error)
	GetLearnerRecentAchievements(ctx context.Context, learnerID uuid.UUID, limit int) ([]*LearnerAchievement, error)
	GetLearnerNextAchievements(ctx context.Context, learnerID uuid.UUID, limit int) ([]*LearnerAchievement, error)

	// 排行?
	GetLeaderboard(ctx context.Context, achievementType string, period string, limit int) ([]*LeaderboardEntry, error)
	GetLearnerRank(ctx context.Context, learnerID uuid.UUID, achievementType string) (int, error)

	// 成就进度
	UpdateAchievementProgress(ctx context.Context, learnerID, achievementID uuid.UUID, progress float64, currentValue float64) error
	UnlockAchievement(ctx context.Context, learnerID, achievementID uuid.UUID) error
	CheckAchievementCriteria(ctx context.Context, learnerID uuid.UUID, eventType string, eventData map[string]interface{}) ([]*Achievement, error)
}

// LeaderboardEntry 排行榜条?
type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	LearnerID    uuid.UUID `json:"learner_id"`
	LearnerName  string    `json:"learner_name"`
	Score        int       `json:"score"`
	Achievements int       `json:"achievements"`
	Avatar       string    `json:"avatar,omitempty"`
}

