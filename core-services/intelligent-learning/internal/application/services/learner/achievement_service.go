package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningAchievementService 学习成就应用服务
type LearningAchievementService struct {
	learnerRepo      repositories.LearnerRepository
	achievementRepo  repositories.AchievementRepository
	analyticsService interfaces.LearningAnalyticsService
	notificationService NotificationService
}

// NewLearningAchievementService 创建新的学习成就应用服务
func NewLearningAchievementService(
	learnerRepo repositories.LearnerRepository,
	achievementRepo repositories.AchievementRepository,
	analyticsService interfaces.LearningAnalyticsService,
	notificationService NotificationService,
) *LearningAchievementService {
	return &LearningAchievementService{
		learnerRepo:         learnerRepo,
		achievementRepo:     achievementRepo,
		analyticsService:    analyticsService,
		notificationService: notificationService,
	}
}

// AchievementType 成就类型
type AchievementType string

const (
	AchievementTypeProgress    AchievementType = "progress"     // 进度成就
	AchievementTypeStreak      AchievementType = "streak"       // 连续学习成就
	AchievementTypeSkill       AchievementType = "skill"        // 技能掌握成�?
	AchievementTypeMilestone   AchievementType = "milestone"    // 里程碑成�?
	AchievementTypeTime        AchievementType = "time"         // 时间成就
	AchievementTypeQuality     AchievementType = "quality"      // 质量成就
	AchievementTypeSocial      AchievementType = "social"       // 社交成就
	AchievementTypeChallenge   AchievementType = "challenge"    // 挑战成就
)

// AchievementLevel 成就等级
type AchievementLevel string

const (
	AchievementLevelBronze   AchievementLevel = "bronze"
	AchievementLevelSilver   AchievementLevel = "silver"
	AchievementLevelGold     AchievementLevel = "gold"
	AchievementLevelPlatinum AchievementLevel = "platinum"
	AchievementLevelDiamond  AchievementLevel = "diamond"
)

// Achievement 成就
type Achievement struct {
	ID          uuid.UUID        `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Type        AchievementType  `json:"type"`
	Level       AchievementLevel `json:"level"`
	Icon        string           `json:"icon"`
	Points      int              `json:"points"`
	Criteria    AchievementCriteria `json:"criteria"`
	Rewards     []AchievementReward `json:"rewards"`
	IsActive    bool             `json:"is_active"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// AchievementCriteria 成就标准
type AchievementCriteria struct {
	Type           string                 `json:"type"`
	TargetValue    float64                `json:"target_value"`
	TimeFrame      *time.Duration         `json:"time_frame,omitempty"`
	Conditions     map[string]interface{} `json:"conditions"`
	Dependencies   []uuid.UUID            `json:"dependencies,omitempty"`
}

// AchievementReward 成就奖励
type AchievementReward struct {
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
}

// LearnerAchievement 学习者成�?
type LearnerAchievement struct {
	ID            uuid.UUID   `json:"id"`
	LearnerID     uuid.UUID   `json:"learner_id"`
	AchievementID uuid.UUID   `json:"achievement_id"`
	Achievement   Achievement `json:"achievement"`
	Progress      float64     `json:"progress"`
	IsUnlocked    bool        `json:"is_unlocked"`
	UnlockedAt    *time.Time  `json:"unlocked_at,omitempty"`
	CurrentValue  float64     `json:"current_value"`
	TargetValue   float64     `json:"target_value"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// AchievementSummary 成就摘要
type AchievementSummary struct {
	LearnerID        uuid.UUID `json:"learner_id"`
	TotalPoints      int       `json:"total_points"`
	UnlockedCount    int       `json:"unlocked_count"`
	InProgressCount  int       `json:"in_progress_count"`
	CompletionRate   float64   `json:"completion_rate"`
	RecentAchievements []LearnerAchievement `json:"recent_achievements"`
	NextAchievements   []LearnerAchievement `json:"next_achievements"`
	LevelDistribution  map[AchievementLevel]int `json:"level_distribution"`
	TypeDistribution   map[AchievementType]int  `json:"type_distribution"`
}

// CheckAchievementsRequest 检查成就请�?
type CheckAchievementsRequest struct {
	LearnerID uuid.UUID              `json:"learner_id" binding:"required"`
	EventType string                 `json:"event_type" binding:"required"`
	EventData map[string]interface{} `json:"event_data"`
}

// CheckAchievementsResponse 检查成就响�?
type CheckAchievementsResponse struct {
	NewAchievements     []LearnerAchievement `json:"new_achievements"`
	UpdatedAchievements []LearnerAchievement `json:"updated_achievements"`
	TotalPoints         int                  `json:"total_points"`
	Message             string               `json:"message"`
}

// GetAchievementsRequest 获取成就请求
type GetAchievementsRequest struct {
	LearnerID uuid.UUID        `json:"learner_id" binding:"required"`
	Type      *AchievementType `json:"type,omitempty"`
	Level     *AchievementLevel `json:"level,omitempty"`
	Status    string           `json:"status,omitempty"` // unlocked, in_progress, locked
	Page      int              `json:"page,omitempty"`
	Limit     int              `json:"limit,omitempty"`
}

// GetAchievementsResponse 获取成就响应
type GetAchievementsResponse struct {
	Achievements []LearnerAchievement `json:"achievements"`
	Summary      AchievementSummary   `json:"summary"`
	Total        int                  `json:"total"`
	Page         int                  `json:"page"`
	Limit        int                  `json:"limit"`
}

// CheckAchievements 检查并更新学习者成�?
func (s *LearningAchievementService) CheckAchievements(ctx context.Context, req *CheckAchievementsRequest) (*CheckAchievementsResponse, error) {
	// 获取学习者信�?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 获取学习者当前成就状�?
	currentAchievements, err := s.achievementRepo.GetLearnerAchievements(ctx, req.LearnerID, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get current achievements: %w", err)
	}

	// 获取所有可用成�?
	availableAchievements, err := s.achievementRepo.GetActiveAchievements(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available achievements: %w", err)
	}

	var newAchievements []LearnerAchievement
	var updatedAchievements []LearnerAchievement

	// 转换领域层类型到应用层类�?
	convertedCurrentAchievements := s.convertDomainLearnerAchievements(currentAchievements)

	// 检查每个成�?
	for _, domainAchievement := range availableAchievements {
		learnerAchievement := s.findLearnerAchievement(convertedCurrentAchievements, domainAchievement.ID)
		
		if learnerAchievement == nil {
			// 创建新的学习者成就记�?
			achievement := s.convertDomainAchievement(domainAchievement)
			learnerAchievement = &LearnerAchievement{
				ID:            uuid.New(),
				LearnerID:     req.LearnerID,
				AchievementID: domainAchievement.ID,
				Achievement:   achievement,
				Progress:      0.0,
				IsUnlocked:    false,
				CurrentValue:  0.0,
				TargetValue:   s.getTargetValueFromCriteria(domainAchievement.Criteria),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
		}

		// 检查成就进�?
		updated, err := s.checkAchievementProgress(ctx, learnerAchievement, learner, req.EventType, req.EventData)
		if err != nil {
			continue // 记录错误但继续处理其他成�?
		}

		if updated {
			if learnerAchievement.IsUnlocked && learnerAchievement.UnlockedAt == nil {
				// 新解锁的成就
				now := time.Now()
				learnerAchievement.UnlockedAt = &now
				newAchievements = append(newAchievements, *learnerAchievement)
				
				// 发送通知
				go s.sendAchievementNotification(ctx, learnerAchievement)
			} else {
				// 更新的成�?
				updatedAchievements = append(updatedAchievements, *learnerAchievement)
			}

			// 保存更新
			domainLearnerAchievement := s.convertAppLearnerAchievementToDomain(learnerAchievement)
			var err error
			if learnerAchievement.CreatedAt.IsZero() {
				err = s.achievementRepo.CreateLearnerAchievement(ctx, domainLearnerAchievement)
			} else {
				err = s.achievementRepo.UpdateLearnerAchievement(ctx, domainLearnerAchievement)
			}
			if err != nil {
				// 记录错误但继�?
				fmt.Printf("Failed to save achievement: %v\n", err)
			}
		}
	}

	// 计算总积�?
	totalPoints := s.calculateTotalPoints(convertedCurrentAchievements)

	return &CheckAchievementsResponse{
		NewAchievements:     newAchievements,
		UpdatedAchievements: updatedAchievements,
		TotalPoints:         totalPoints,
		Message:             s.generateAchievementMessage(newAchievements, updatedAchievements),
	}, nil
}

// GetLearnerAchievements 获取学习者成�?
func (s *LearningAchievementService) GetLearnerAchievements(ctx context.Context, req *GetAchievementsRequest) (*GetAchievementsResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// 获取学习者成�?
	domainAchievements, err := s.achievementRepo.GetLearnerAchievements(ctx, req.LearnerID, (req.Page-1)*req.Limit, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner achievements: %w", err)
	}

	// 转换为应用层类型
	achievements := s.convertDomainLearnerAchievements(domainAchievements)
	total := len(achievements)

	// 生成成就摘要
	summary, err := s.generateAchievementSummary(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate achievement summary: %w", err)
	}

	return &GetAchievementsResponse{
		Achievements: achievements,
		Summary:      *summary,
		Total:        total,
		Page:         req.Page,
		Limit:        req.Limit,
	}, nil
}

// CreateAchievement 创建新成�?
func (s *LearningAchievementService) CreateAchievement(ctx context.Context, achievement *Achievement) error {
	achievement.ID = uuid.New()
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	// 转换应用层Achievement到领域层
	domainAchievement := s.convertAppAchievementToDomain(achievement)
	return s.achievementRepo.CreateAchievement(ctx, domainAchievement)
}

// 辅助方法

func (s *LearningAchievementService) findLearnerAchievement(achievements []LearnerAchievement, achievementID uuid.UUID) *LearnerAchievement {
	for i := range achievements {
		if achievements[i].AchievementID == achievementID {
			return &achievements[i]
		}
	}
	return nil
}

// convertDomainLearnerAchievements 转换领域层学习者成就到应用�?
func (s *LearningAchievementService) convertDomainLearnerAchievements(domainAchievements []*repositories.LearnerAchievement) []LearnerAchievement {
	result := make([]LearnerAchievement, len(domainAchievements))
	for i, da := range domainAchievements {
		achievement := Achievement{}
		if da.Achievement != nil {
			achievement = s.convertDomainAchievement(da.Achievement)
		}
		result[i] = LearnerAchievement{
			ID:            da.ID,
			LearnerID:     da.LearnerID,
			AchievementID: da.AchievementID,
			Achievement:   achievement,
			Progress:      da.Progress,
			IsUnlocked:    da.IsUnlocked,
			UnlockedAt:    da.UnlockedAt,
			CurrentValue:  da.CurrentValue,
			TargetValue:   da.TargetValue,
			CreatedAt:     da.CreatedAt,
			UpdatedAt:     da.UpdatedAt,
		}
	}
	return result
}

// convertDomainAchievement 转换领域层成就到应用�?
func (s *LearningAchievementService) convertDomainAchievement(domainAchievement *repositories.Achievement) Achievement {
	return Achievement{
		ID:          domainAchievement.ID,
		Title:       domainAchievement.Title,
		Description: domainAchievement.Description,
		Type:        AchievementType(domainAchievement.Type),
		Level:       AchievementLevel(domainAchievement.Level),
		Points:      domainAchievement.Points,
		Icon:        domainAchievement.Icon,
		IsActive:    domainAchievement.IsActive,
		CreatedAt:   domainAchievement.CreatedAt,
		UpdatedAt:   domainAchievement.UpdatedAt,
	}
}

// convertAppLearnerAchievementToDomain 转换应用层学习者成就到领域�?
func (s *LearningAchievementService) convertAppLearnerAchievementToDomain(appAchievement *LearnerAchievement) *repositories.LearnerAchievement {
	var achievement *repositories.Achievement
	if appAchievement.Achievement.ID != uuid.Nil {
		achievement = s.convertAppAchievementToDomain(&appAchievement.Achievement)
	}
	
	return &repositories.LearnerAchievement{
		ID:            appAchievement.ID,
		LearnerID:     appAchievement.LearnerID,
		AchievementID: appAchievement.AchievementID,
		Achievement:   achievement,
		Progress:      appAchievement.Progress,
		IsUnlocked:    appAchievement.IsUnlocked,
		UnlockedAt:    appAchievement.UnlockedAt,
		CurrentValue:  appAchievement.CurrentValue,
		TargetValue:   appAchievement.TargetValue,
		CreatedAt:     appAchievement.CreatedAt,
		UpdatedAt:     appAchievement.UpdatedAt,
	}
}

// convertAppAchievementToDomain 转换应用层成就到领域�?
func (s *LearningAchievementService) convertAppAchievementToDomain(appAchievement *Achievement) *repositories.Achievement {
	return &repositories.Achievement{
		ID:          appAchievement.ID,
		Title:       appAchievement.Title,
		Description: appAchievement.Description,
		Type:        string(appAchievement.Type),
		Level:       string(appAchievement.Level),
		Points:      appAchievement.Points,
		Icon:        appAchievement.Icon,
		IsActive:    appAchievement.IsActive,
		CreatedAt:   appAchievement.CreatedAt,
		UpdatedAt:   appAchievement.UpdatedAt,
	}
}

func (s *LearningAchievementService) checkAchievementProgress(
	ctx context.Context,
	learnerAchievement *LearnerAchievement,
	learner *entities.Learner,
	eventType string,
	eventData map[string]interface{},
) (bool, error) {
	achievement := learnerAchievement.Achievement
	
	// 如果已经解锁，无需再检�?
	if learnerAchievement.IsUnlocked {
		return false, nil
	}

	var currentValue float64
	var err error

	// 根据成就类型计算当前�?
	switch achievement.Type {
	case AchievementTypeProgress:
		currentValue, err = s.calculateProgressValue(ctx, learner, achievement.Criteria)
	case AchievementTypeStreak:
		currentValue, err = s.calculateStreakValue(ctx, learner, achievement.Criteria)
	case AchievementTypeSkill:
		currentValue, err = s.calculateSkillValue(ctx, learner, achievement.Criteria)
	case AchievementTypeTime:
		currentValue, err = s.calculateTimeValue(ctx, learner, achievement.Criteria)
	default:
		return false, fmt.Errorf("unsupported achievement type: %s", achievement.Type)
	}

	if err != nil {
		return false, err
	}

	// 更新当前值和进度
	oldValue := learnerAchievement.CurrentValue
	learnerAchievement.CurrentValue = currentValue
	learnerAchievement.Progress = currentValue / achievement.Criteria.TargetValue
	learnerAchievement.UpdatedAt = time.Now()

	// 检查是否解�?
	if currentValue >= achievement.Criteria.TargetValue && !learnerAchievement.IsUnlocked {
		learnerAchievement.IsUnlocked = true
		learnerAchievement.Progress = 1.0
		return true, nil
	}

	// 检查是否有进度更新
	return currentValue != oldValue, nil
}

func (s *LearningAchievementService) calculateProgressValue(ctx context.Context, learner *entities.Learner, criteria AchievementCriteria) (float64, error) {
	// 简化实现：计算完成的学习活动数�?
	if s.analyticsService != nil {
		// 这里应该调用分析服务获取实际数据
		return 10.0, nil // 模拟�?
	}
	return 0.0, nil
}

func (s *LearningAchievementService) calculateStreakValue(ctx context.Context, learner *entities.Learner, criteria AchievementCriteria) (float64, error) {
	// 计算连续学习天数
	return float64(learner.Streak.CurrentStreak), nil
}

func (s *LearningAchievementService) calculateSkillValue(ctx context.Context, learner *entities.Learner, criteria AchievementCriteria) (float64, error) {
	// 计算掌握的技能数�?
	masteredSkills := 0
	for _, skill := range learner.Skills {
		if entities.DifficultyLevel(skill.Level) >= entities.DifficultyAdvanced {
			masteredSkills++
		}
	}
	return float64(masteredSkills), nil
}

func (s *LearningAchievementService) calculateTimeValue(ctx context.Context, learner *entities.Learner, criteria AchievementCriteria) (float64, error) {
	// 计算总学习时间（小时�?
	if s.analyticsService != nil {
		// 这里应该调用分析服务获取实际数据
		return 50.0, nil // 模拟�?
	}
	return 0.0, nil
}

func (s *LearningAchievementService) calculateTotalPoints(achievements []LearnerAchievement) int {
	total := 0
	for _, achievement := range achievements {
		if achievement.IsUnlocked {
			total += achievement.Achievement.Points
		}
	}
	return total
}

func (s *LearningAchievementService) generateAchievementMessage(newAchievements, updatedAchievements []LearnerAchievement) string {
	if len(newAchievements) > 0 {
		if len(newAchievements) == 1 {
			return fmt.Sprintf("恭喜！您解锁了新成就�?s", newAchievements[0].Achievement.Title)
		}
		return fmt.Sprintf("恭喜！您解锁�?%d 个新成就", len(newAchievements))
	}
	
	if len(updatedAchievements) > 0 {
		return fmt.Sprintf("您在 %d 个成就上取得了进�?, len(updatedAchievements))
	}
	
	return "继续努力学习，更多成就等待您解锁�?
}

func (s *LearningAchievementService) generateAchievementSummary(ctx context.Context, learnerID uuid.UUID) (*AchievementSummary, error) {
	achievements, err := s.achievementRepo.GetLearnerAchievements(ctx, learnerID, 0, 100)
	if err != nil {
		return nil, err
	}

	// 转换为应用层类型
	appAchievements := s.convertDomainLearnerAchievements(achievements)

	summary := &AchievementSummary{
		LearnerID:         learnerID,
		LevelDistribution: make(map[AchievementLevel]int),
		TypeDistribution:  make(map[AchievementType]int),
	}

	unlockedCount := 0
	inProgressCount := 0
	totalPoints := 0

	for _, achievement := range appAchievements {
		if achievement.IsUnlocked {
			unlockedCount++
			totalPoints += achievement.Achievement.Points
		} else if achievement.Progress > 0 {
			inProgressCount++
		}

		summary.LevelDistribution[achievement.Achievement.Level]++
		summary.TypeDistribution[achievement.Achievement.Type]++
	}

	summary.UnlockedCount = unlockedCount
	summary.InProgressCount = inProgressCount
	summary.TotalPoints = totalPoints
	
	if len(appAchievements) > 0 {
		summary.CompletionRate = float64(unlockedCount) / float64(len(appAchievements))
	}
	
	// 获取最近解锁的成就
	summary.RecentAchievements = s.getRecentAchievements(appAchievements, 5)
	
	// 获取即将解锁的成�?
	summary.NextAchievements = s.getNextAchievements(appAchievements, 3)

	return summary, nil
}

func (s *LearningAchievementService) getRecentAchievements(achievements []LearnerAchievement, limit int) []LearnerAchievement {
	var recent []LearnerAchievement
	for _, achievement := range achievements {
		if achievement.IsUnlocked && achievement.UnlockedAt != nil {
			recent = append(recent, achievement)
		}
	}
	
	// 按解锁时间排�?
	// 这里简化处理，实际应该排序
	if len(recent) > limit {
		recent = recent[:limit]
	}
	
	return recent
}

func (s *LearningAchievementService) getNextAchievements(achievements []LearnerAchievement, limit int) []LearnerAchievement {
	var next []LearnerAchievement
	for _, achievement := range achievements {
		if !achievement.IsUnlocked && achievement.Progress > 0 {
			next = append(next, achievement)
		}
	}
	
	// 按进度排序，选择最接近完成�?
	// 这里简化处理，实际应该排序
	if len(next) > limit {
		next = next[:limit]
	}
	
	return next
}

func (s *LearningAchievementService) sendAchievementNotification(ctx context.Context, achievement *LearnerAchievement) {
	if s.notificationService != nil {
		notification := map[string]interface{}{
			"type":           "achievement_unlocked",
			"learner_id":     achievement.LearnerID,
			"achievement_id": achievement.AchievementID,
			"title":          achievement.Achievement.Title,
			"description":    achievement.Achievement.Description,
			"points":         achievement.Achievement.Points,
		}
		
		// 异步发送通知
		if err := s.notificationService.SendNotification(ctx, notification); err != nil {
			fmt.Printf("Failed to send achievement notification: %v\n", err)
		}
	}
}

// NotificationService 通知服务接口
type NotificationService interface {
	SendNotification(ctx context.Context, notification map[string]interface{}) error
}

// getTargetValueFromCriteria 从成就标准中提取目标�?
func (s *LearningAchievementService) getTargetValueFromCriteria(criteria map[string]interface{}) float64 {
	if targetValue, ok := criteria["target_value"]; ok {
		switch v := targetValue.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			// 尝试解析字符串为数字
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				return val
			}
		}
	}
	// 默认返回0
	return 0.0
}
