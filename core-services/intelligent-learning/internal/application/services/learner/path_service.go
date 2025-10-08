package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningPathService 学习路径应用服务
type LearningPathService struct {
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	pathRepo           repositories.LearningPathRepository
	analyticsService   LearningAnalyticsService
}

// NewLearningPathService 创建新的学习路径应用服务
func NewLearningPathService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	pathRepo repositories.LearningPathRepository,
	analyticsService LearningAnalyticsService,
) *LearningPathService {
	return &LearningPathService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		pathRepo:           pathRepo,
		analyticsService:   analyticsService,
	}
}

// GeneratePersonalizedPathRequest 生成个性化路径请求
type GeneratePersonalizedPathRequest struct {
	LearnerID       uuid.UUID `json:"learner_id" binding:"required"`
	TargetSkills    []string  `json:"target_skills" binding:"required"`
	TimeConstraint  int       `json:"time_constraint,omitempty"` // 学习时间限制（小时）
	DifficultyLevel string    `json:"difficulty_level,omitempty"`
	LearningStyle   string    `json:"learning_style,omitempty"`
	Prerequisites   []string  `json:"prerequisites,omitempty"`
}

// LearningPathResponse 学习路径响应
type LearningPathResponse struct {
	ID                uuid.UUID                    `json:"id"`
	LearnerID         uuid.UUID                    `json:"learner_id"`
	Title             string                       `json:"title"`
	Description       string                       `json:"description"`
	EstimatedDuration int                          `json:"estimated_duration"` // 预计学习时间（小时）
	DifficultyLevel   string                       `json:"difficulty_level"`
	Steps             []LearningPathStepResponse   `json:"steps"`
	Milestones        []LearningMilestoneResponse  `json:"milestones"`
	Progress          float64                      `json:"progress"`
	Status            string                       `json:"status"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

// LearningPathStepResponse 学习路径步骤响应
type LearningPathStepResponse struct {
	ID              uuid.UUID `json:"id"`
	Order           int       `json:"order"`
	ContentID       uuid.UUID `json:"content_id"`
	ContentTitle    string    `json:"content_title"`
	ContentType     string    `json:"content_type"`
	EstimatedTime   int       `json:"estimated_time"` // 预计学习时间（分钟）
	Prerequisites   []string  `json:"prerequisites"`
	LearningGoals   []string  `json:"learning_goals"`
	IsCompleted     bool      `json:"is_completed"`
	CompletionRate  float64   `json:"completion_rate"`
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty"`
}

// LearningMilestoneResponse 学习里程碑响应
type LearningMilestoneResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TargetStep  int       `json:"target_step"`
	IsAchieved  bool      `json:"is_achieved"`
	AchievedAt  *time.Time `json:"achieved_at,omitempty"`
	Reward      string    `json:"reward,omitempty"`
}

// UpdatePathProgressRequest 更新路径进度请求
type UpdatePathProgressRequest struct {
	PathID    uuid.UUID `json:"path_id" binding:"required"`
	StepID    uuid.UUID `json:"step_id" binding:"required"`
	Progress  float64   `json:"progress" binding:"min=0,max=1"`
	TimeSpent int       `json:"time_spent"` // 学习时间（分钟）
}

// PathRecommendationRequest 路径推荐请求
type PathRecommendationRequest struct {
	LearnerID      uuid.UUID `json:"learner_id" binding:"required"`
	CurrentSkills  []string  `json:"current_skills"`
	InterestAreas  []string  `json:"interest_areas"`
	AvailableTime  int       `json:"available_time,omitempty"` // 可用学习时间（小时/周）
	LearningGoals  []string  `json:"learning_goals"`
}

// PathRecommendationResponse 路径推荐响应
type PathRecommendationResponse struct {
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
	Reasoning        string            `json:"reasoning"`
	Confidence       float64           `json:"confidence"`
}

// RecommendedPath 推荐路径
type RecommendedPath struct {
	PathID          uuid.UUID `json:"path_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	MatchScore      float64   `json:"match_score"`
	EstimatedTime   int       `json:"estimated_time"`
	DifficultyLevel string    `json:"difficulty_level"`
	SkillsGained    []string  `json:"skills_gained"`
	Reasons         []string  `json:"reasons"`
}

// GeneratePersonalizedPath 生成个性化学习路径
func (s *LearningPathService) GeneratePersonalizedPath(ctx context.Context, req *GeneratePersonalizedPathRequest) (*LearningPathResponse, error) {
	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 分析学习者当前技能水平
	currentSkills := s.extractCurrentSkills(learner)
	
	// 基于目标技能构建知识图谱路径
	knowledgePath, err := s.buildKnowledgePath(ctx, req.TargetSkills, currentSkills)
	if err != nil {
		return nil, fmt.Errorf("failed to build knowledge path: %w", err)
	}

	// 根据学习者偏好选择合适的内容
	pathSteps, err := s.selectOptimalContent(ctx, knowledgePath, learner, req)
	if err != nil {
		return nil, fmt.Errorf("failed to select content: %w", err)
	}

	// 创建学习路径
	path := &entities.LearningPath{
		ID:                uuid.New(),
		LearnerID:         req.LearnerID,
		Title:             s.generatePathTitle(req.TargetSkills),
		Description:       s.generatePathDescription(req.TargetSkills, len(pathSteps)),
		EstimatedDuration: s.calculateEstimatedDuration(pathSteps),
		DifficultyLevel:   s.determineDifficultyLevel(req.DifficultyLevel, learner),
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// 保存学习路径
	if err := s.pathRepo.Create(ctx, path); err != nil {
		return nil, fmt.Errorf("failed to create learning path: %w", err)
	}

	// 生成里程碑
	milestones := s.generateMilestones(pathSteps)

	return &LearningPathResponse{
		ID:                path.ID,
		LearnerID:         path.LearnerID,
		Title:             path.Title,
		Description:       path.Description,
		EstimatedDuration: path.EstimatedDuration,
		DifficultyLevel:   path.DifficultyLevel,
		Steps:             pathSteps,
		Milestones:        milestones,
		Progress:          0.0,
		Status:            path.Status,
		CreatedAt:         path.CreatedAt,
		UpdatedAt:         path.UpdatedAt,
	}, nil
}

// UpdatePathProgress 更新学习路径进度
func (s *LearningPathService) UpdatePathProgress(ctx context.Context, req *UpdatePathProgressRequest) error {
	// 获取学习路径
	path, err := s.pathRepo.GetByID(ctx, req.PathID)
	if err != nil {
		return fmt.Errorf("failed to get learning path: %w", err)
	}

	// 更新步骤进度
	// 这里需要实现具体的进度更新逻辑
	
	// 记录学习分析数据
	if s.analyticsService != nil {
		analyticsData := map[string]interface{}{
			"path_id":    req.PathID,
			"step_id":    req.StepID,
			"progress":   req.Progress,
			"time_spent": req.TimeSpent,
			"timestamp":  time.Now(),
		}
		
		// 异步记录分析数据
		go func() {
			if err := s.analyticsService.RecordLearningActivity(ctx, path.LearnerID, "path_progress", analyticsData); err != nil {
				// 记录日志但不影响主流程
				fmt.Printf("Failed to record analytics: %v\n", err)
			}
		}()
	}

	return nil
}

// GetRecommendedPaths 获取推荐学习路径
func (s *LearningPathService) GetRecommendedPaths(ctx context.Context, req *PathRecommendationRequest) (*PathRecommendationResponse, error) {
	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 分析学习者特征
	learnerProfile := s.analyzeLearnerProfile(learner, req)
	
	// 获取候选路径
	candidatePaths, err := s.getCandidatePaths(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate paths: %w", err)
	}

	// 计算匹配分数
	recommendedPaths := s.calculateMatchScores(candidatePaths, learnerProfile)
	
	// 排序并选择前N个
	sort.Slice(recommendedPaths, func(i, j int) bool {
		return recommendedPaths[i].MatchScore > recommendedPaths[j].MatchScore
	})

	if len(recommendedPaths) > 5 {
		recommendedPaths = recommendedPaths[:5]
	}

	return &PathRecommendationResponse{
		RecommendedPaths: recommendedPaths,
		Reasoning:        s.generateRecommendationReasoning(learnerProfile),
		Confidence:       s.calculateRecommendationConfidence(recommendedPaths),
	}, nil
}

// 辅助方法

func (s *LearningPathService) extractCurrentSkills(learner *entities.Learner) []string {
	var skills []string
	for _, skill := range learner.Skills {
		if skill.Level >= entities.SkillLevelIntermediate {
			skills = append(skills, skill.Name)
		}
	}
	return skills
}

func (s *LearningPathService) buildKnowledgePath(ctx context.Context, targetSkills, currentSkills []string) ([]string, error) {
	// 这里应该调用知识图谱服务来构建最优学习路径
	// 简化实现，返回目标技能列表
	return targetSkills, nil
}

func (s *LearningPathService) selectOptimalContent(ctx context.Context, knowledgePath []string, learner *entities.Learner, req *GeneratePersonalizedPathRequest) ([]LearningPathStepResponse, error) {
	var steps []LearningPathStepResponse
	
	for i, skill := range knowledgePath {
		// 根据技能查找相关内容
		contents, err := s.contentRepo.FindBySkill(ctx, skill)
		if err != nil {
			continue
		}

		// 选择最适合的内容
		if len(contents) > 0 {
			content := s.selectBestContent(contents, learner)
			step := LearningPathStepResponse{
				ID:             uuid.New(),
				Order:          i + 1,
				ContentID:      content.ID,
				ContentTitle:   content.Title,
				ContentType:    string(content.Type),
				EstimatedTime:  content.EstimatedDuration,
				Prerequisites:  []string{},
				LearningGoals:  []string{skill},
				IsCompleted:    false,
				CompletionRate: 0.0,
			}
			steps = append(steps, step)
		}
	}

	return steps, nil
}

func (s *LearningPathService) selectBestContent(contents []*entities.LearningContent, learner *entities.Learner) *entities.LearningContent {
	// 简化实现，选择第一个内容
	// 实际应该根据学习者偏好、难度等因素选择
	if len(contents) > 0 {
		return contents[0]
	}
	return nil
}

func (s *LearningPathService) generatePathTitle(targetSkills []string) string {
	if len(targetSkills) == 1 {
		return fmt.Sprintf("%s 学习路径", targetSkills[0])
	}
	return fmt.Sprintf("综合技能学习路径 (%d项技能)", len(targetSkills))
}

func (s *LearningPathService) generatePathDescription(targetSkills []string, stepCount int) string {
	return fmt.Sprintf("这是一个包含%d个学习步骤的个性化学习路径，旨在帮助您掌握以下技能：%v", stepCount, targetSkills)
}

func (s *LearningPathService) calculateEstimatedDuration(steps []LearningPathStepResponse) int {
	total := 0
	for _, step := range steps {
		total += step.EstimatedTime
	}
	return total / 60 // 转换为小时
}

func (s *LearningPathService) determineDifficultyLevel(requested string, learner *entities.Learner) string {
	if requested != "" {
		return requested
	}
	
	// 根据学习者技能水平确定难度
	avgLevel := s.calculateAverageSkillLevel(learner)
	if avgLevel >= 4 {
		return "advanced"
	} else if avgLevel >= 3 {
		return "intermediate"
	}
	return "beginner"
}

func (s *LearningPathService) calculateAverageSkillLevel(learner *entities.Learner) float64 {
	if len(learner.Skills) == 0 {
		return 1.0
	}
	
	total := 0.0
	for _, skill := range learner.Skills {
		total += float64(skill.Level)
	}
	return total / float64(len(learner.Skills))
}

func (s *LearningPathService) generateMilestones(steps []LearningPathStepResponse) []LearningMilestoneResponse {
	var milestones []LearningMilestoneResponse
	
	// 每完成25%的步骤设置一个里程碑
	stepCount := len(steps)
	milestonePoints := []int{stepCount / 4, stepCount / 2, stepCount * 3 / 4, stepCount}
	
	for i, point := range milestonePoints {
		if point > 0 && point <= stepCount {
			milestone := LearningMilestoneResponse{
				ID:          uuid.New(),
				Title:       fmt.Sprintf("里程碑 %d", i+1),
				Description: fmt.Sprintf("完成前 %d 个学习步骤", point),
				TargetStep:  point,
				IsAchieved:  false,
			}
			milestones = append(milestones, milestone)
		}
	}
	
	return milestones
}

func (s *LearningPathService) analyzeLearnerProfile(learner *entities.Learner, req *PathRecommendationRequest) map[string]interface{} {
	return map[string]interface{}{
		"learning_style":   learner.LearningStyle,
		"current_skills":   req.CurrentSkills,
		"interest_areas":   req.InterestAreas,
		"available_time":   req.AvailableTime,
		"learning_goals":   req.LearningGoals,
		"skill_level":      s.calculateAverageSkillLevel(learner),
	}
}

func (s *LearningPathService) getCandidatePaths(ctx context.Context, req *PathRecommendationRequest) ([]RecommendedPath, error) {
	// 简化实现，返回模拟数据
	// 实际应该从数据库查询相关路径
	return []RecommendedPath{}, nil
}

func (s *LearningPathService) calculateMatchScores(paths []RecommendedPath, profile map[string]interface{}) []RecommendedPath {
	// 简化实现，为每个路径计算匹配分数
	for i := range paths {
		paths[i].MatchScore = 0.8 // 模拟分数
	}
	return paths
}

func (s *LearningPathService) generateRecommendationReasoning(profile map[string]interface{}) string {
	return "基于您的学习风格、当前技能水平和学习目标，我们为您推荐了以下学习路径。"
}

func (s *LearningPathService) calculateRecommendationConfidence(paths []RecommendedPath) float64 {
	if len(paths) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, path := range paths {
		totalScore += path.MatchScore
	}
	return totalScore / float64(len(paths))
}