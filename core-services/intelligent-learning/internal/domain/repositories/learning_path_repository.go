package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LearningPath 学习路径实体
type LearningPath struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	LearnerID   uuid.UUID `json:"learner_id"`
	GraphID     uuid.UUID `json:"graph_id"`
	Nodes       []PathNode `json:"nodes"`
	Edges       []PathEdge `json:"edges"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PathNode 路径节点
type PathNode struct {
	ID          uuid.UUID `json:"id"`
	ContentID   uuid.UUID `json:"content_id"`
	Position    int       `json:"position"`
	IsCompleted bool      `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// PathEdge 路径?
type PathEdge struct {
	ID       uuid.UUID `json:"id"`
	FromNode uuid.UUID `json:"from_node"`
	ToNode   uuid.UUID `json:"to_node"`
	Weight   float64   `json:"weight"`
}

// PersonalizedPath 个性化路径
type PersonalizedPath struct {
	ID              uuid.UUID `json:"id"`
	LearnerID       uuid.UUID `json:"learner_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Difficulty      string    `json:"difficulty"`
	EstimatedTime   int       `json:"estimated_time"`
	CompletionRate  float64   `json:"completion_rate"`
	Nodes           []PathNode `json:"nodes"`
	Recommendations []string  `json:"recommendations"`
	CreatedAt       time.Time `json:"created_at"`
}

// LearningPathRepository 学习路径数据访问接口
type LearningPathRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, path *LearningPath) error
	GetByID(ctx context.Context, id uuid.UUID) (*LearningPath, error)
	Update(ctx context.Context, path *LearningPath) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 查询操作
	GetByLearnerID(ctx context.Context, learnerID uuid.UUID) ([]*LearningPath, error)
	GetByGraphID(ctx context.Context, graphID uuid.UUID) ([]*LearningPath, error)
	GetActivePaths(ctx context.Context, learnerID uuid.UUID) ([]*LearningPath, error)

	// 个性化路径
	CreatePersonalizedPath(ctx context.Context, path *PersonalizedPath) error
	GetPersonalizedPaths(ctx context.Context, learnerID uuid.UUID) ([]*PersonalizedPath, error)
	UpdatePersonalizedPath(ctx context.Context, path *PersonalizedPath) error

	// 路径推荐
	GetRecommendedPaths(ctx context.Context, learnerID uuid.UUID, limit int) ([]*PersonalizedPath, error)
	GetPathsByDifficulty(ctx context.Context, difficulty string) ([]*LearningPath, error)

	// 路径进度
	UpdatePathProgress(ctx context.Context, pathID uuid.UUID, nodeID uuid.UUID, completed bool) error
	GetPathProgress(ctx context.Context, pathID uuid.UUID) (float64, error)
}

