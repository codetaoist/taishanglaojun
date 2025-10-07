package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthRecommendationService 健康建议服务
type HealthRecommendationService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthRecommendationService 创建健康建议服务
func NewHealthRecommendationService(
	healthDataRepo domain.HealthDataRepository,
	healthProfileRepo domain.HealthProfileRepository,
	eventPublisher EventPublisher,
) *HealthRecommendationService {
	return &HealthRecommendationService{
		healthDataRepo:    healthDataRepo,
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// RecommendationType 建议类型
type RecommendationType string

const (
	RecommendationTypeExercise   RecommendationType = "exercise"   // 运动建议
	RecommendationTypeDiet       RecommendationType = "diet"       // 饮食建议
	RecommendationTypeSleep      RecommendationType = "sleep"      // 睡眠建议
	RecommendationTypeStress     RecommendationType = "stress"     // 压力管理
	RecommendationTypeMedical    RecommendationType = "medical"    // 医疗建议
	RecommendationTypeLifestyle  RecommendationType = "lifestyle"  // 生活方式
	RecommendationTypePrevention RecommendationType = "prevention" // 预防建议
)

// RecommendationPriority 建议优先级
type RecommendationPriority string

const (
	RecommendationPriorityHigh   RecommendationPriority = "high"   // 高优先级
	RecommendationPriorityMedium RecommendationPriority = "medium" // 中优先级
	RecommendationPriorityLow    RecommendationPriority = "low"    // 低优先级
)

// HealthRecommendation 健康建议
type HealthRecommendation struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Type        RecommendationType     `json:"type"`
	Priority    RecommendationPriority `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Actions     []string               `json:"actions"`
	Benefits    []string               `json:"benefits"`
	Duration    string                 `json:"duration,omitempty"`
	Frequency   string                 `json:"frequency,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateRecommendationsRequest 生成建议请求
type GenerateRecommendationsRequest struct {
	UserID uuid.UUID            `json:"user_id" binding:"required"`
	Types  []RecommendationType `json:"types,omitempty"`
	Days   int                  `json:"days,omitempty"`
	Limit  int                  `json:"limit,omitempty"`
}

// GenerateRecommendationsResponse 生成建议响应
type GenerateRecommendationsResponse struct {
	Recommendations []HealthRecommendation `json:"recommendations"`
	Summary         string                 `json:"summary"`
	Count           int                    `json:"count"`
	Timestamp       time.Time              `json:"timestamp"`
}

// GetPersonalizedTipsRequest 获取个性化提示请求
type GetPersonalizedTipsRequest struct {
	UserID   uuid.UUID            `json:"user_id" binding:"required"`
	Category RecommendationType   `json:"category,omitempty"`
	Limit    int                  `json:"limit,omitempty"`
}

// GetPersonalizedTipsResponse 获取个性化提示响应
type GetPersonalizedTipsResponse struct {
	Tips      []HealthTip `json:"tips"`
	Category  string      `json:"category"`
	Count     int         `json:"count"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthTip 健康提示
type HealthTip struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Difficulty  string    `json:"difficulty"`
	EstimatedTime string  `json:"estimated_time,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// GenerateRecommendations 生成健康建议
func (s *HealthRecommendationService) GenerateRecommendations(ctx context.Context, req *GenerateRecommendationsRequest) (*GenerateRecommendationsResponse, error) {
	// 设置默认值
	if req.Days == 0 {
		req.Days = 30
	}
	if req.Limit == 0 {
		req.Limit = 10
	}
	if len(req.Types) == 0 {
		req.Types = []RecommendationType{
			RecommendationTypeExercise,
			RecommendationTypeDiet,
			RecommendationTypeSleep,
			RecommendationTypeStress,
		}
	}

	// 获取用户健康档案
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 获取用户健康数据
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -req.Days)
	
	var recommendations []HealthRecommendation

	// 为每种类型生成建议
	for _, recType := range req.Types {
		typeRecommendations, err := s.generateRecommendationsByType(ctx, req.UserID, recType, startTime, endTime, profile)
		if err != nil {
			continue // 继续处理其他类型
		}
		recommendations = append(recommendations, typeRecommendations...)
	}

	// 按优先级排序
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[RecommendationPriority]int{
			RecommendationPriorityHigh:   3,
			RecommendationPriorityMedium: 2,
			RecommendationPriorityLow:    1,
		}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	// 限制数量
	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	// 生成摘要
	summary := s.generateRecommendationSummary(recommendations)

	return &GenerateRecommendationsResponse{
		Recommendations: recommendations,
		Summary:         summary,
		Count:           len(recommendations),
		Timestamp:       time.Now(),
	}, nil
}

// GetPersonalizedTips 获取个性化健康提示
func (s *HealthRecommendationService) GetPersonalizedTips(ctx context.Context, req *GetPersonalizedTipsRequest) (*GetPersonalizedTipsResponse, error) {
	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 5
	}

	// 获取用户健康档案
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 生成个性化提示
	tips := s.generatePersonalizedTips(req.UserID, req.Category, req.Limit, profile)

	category := "general"
	if req.Category != "" {
		category = string(req.Category)
	}

	return &GetPersonalizedTipsResponse{
		Tips:      tips,
		Category:  category,
		Count:     len(tips),
		Timestamp: time.Now(),
	}, nil
}

// generateRecommendationsByType 根据类型生成建议
func (s *HealthRecommendationService) generateRecommendationsByType(ctx context.Context, userID uuid.UUID, recType RecommendationType, startTime, endTime time.Time, profile *domain.HealthProfile) ([]HealthRecommendation, error) {
	var recommendations []HealthRecommendation

	switch recType {
	case RecommendationTypeExercise:
		recommendations = s.generateExerciseRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeDiet:
		recommendations = s.generateDietRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeSleep:
		recommendations = s.generateSleepRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeStress:
		recommendations = s.generateStressRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeMedical:
		recommendations = s.generateMedicalRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeLifestyle:
		recommendations = s.generateLifestyleRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypePrevention:
		recommendations = s.generatePreventionRecommendations(ctx, userID, startTime, endTime, profile)
	}

	return recommendations, nil
}

// generateExerciseRecommendations 生成运动建议
func (s *HealthRecommendationService) generateExerciseRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 获取步数数据
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	
	avgSteps := s.calculateAverageSteps(stepsData)
	age := 30
	if profile != nil {
		age = profile.GetAge()
	}

	// 基于步数生成建议
	if avgSteps < 5000 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeExercise,
			Priority:    RecommendationPriorityHigh,
			Title:       "增加日常活动量",
			Description: "您的日均步数较低，建议增加日常活动量以改善健康状况",
			Actions: []string{
				"每天至少步行30分钟",
				"使用楼梯代替电梯",
				"饭后散步15-20分钟",
				"设置每小时起身活动提醒",
			},
			Benefits: []string{
				"改善心血管健康",
				"增强肌肉力量",
				"提高新陈代谢",
				"改善睡眠质量",
			},
			Duration:  "持续进行",
			Frequency: "每天",
			Tags:      []string{"步行", "日常活动", "心血管"},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	} else if avgSteps < 8000 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeExercise,
			Priority:    RecommendationPriorityMedium,
			Title:       "提升运动强度",
			Description: "您的活动量不错，可以考虑增加一些中等强度的运动",
			Actions: []string{
				"每周进行3次30分钟的快走",
				"尝试游泳或骑自行车",
				"加入力量训练",
				"参加团体运动课程",
			},
			Benefits: []string{
				"提高心肺功能",
				"增强肌肉耐力",
				"改善身体协调性",
				"增强免疫力",
			},
			Duration:  "30-45分钟",
			Frequency: "每周3-4次",
			Tags:      []string{"有氧运动", "力量训练", "团体运动"},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	// 基于年龄生成建议
	if age > 50 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeExercise,
			Priority:    RecommendationPriorityMedium,
			Title:       "适合中老年的运动方案",
			Description: "针对您的年龄段，推荐低冲击、高效益的运动方式",
			Actions: []string{
				"太极拳或瑜伽",
				"水中有氧运动",
				"轻度力量训练",
				"平衡性训练",
			},
			Benefits: []string{
				"改善平衡能力",
				"增强骨密度",
				"减少跌倒风险",
				"提高生活质量",
			},
			Duration:  "30-45分钟",
			Frequency: "每周3-5次",
			Tags:      []string{"太极", "瑜伽", "平衡训练", "中老年"},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	return recommendations
}

// generateDietRecommendations 生成饮食建议
func (s *HealthRecommendationService) generateDietRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 获取血糖数据
	bloodSugarData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "blood_sugar", startTime, endTime)
	avgBloodSugar := s.calculateAverageValue(bloodSugarData)

	// 基于血糖水平生成建议
	if avgBloodSugar > 7.0 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeDiet,
			Priority:    RecommendationPriorityHigh,
			Title:       "血糖控制饮食方案",
			Description: "您的血糖水平偏高，建议调整饮食结构以控制血糖",
			Actions: []string{
				"减少精制糖和高GI食物摄入",
				"增加膳食纤维摄入",
				"控制碳水化合物总量",
				"规律进餐，避免暴饮暴食",
			},
			Benefits: []string{
				"稳定血糖水平",
				"降低糖尿病风险",
				"改善胰岛素敏感性",
				"控制体重",
			},
			Duration:  "长期坚持",
			Frequency: "每餐",
			Tags:      []string{"血糖控制", "低GI", "膳食纤维"},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	// 基于BMI生成建议
	if profile != nil {
		bmi := profile.GetBMI()
		if bmi > 25 {
			recommendations = append(recommendations, HealthRecommendation{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        RecommendationTypeDiet,
				Priority:    RecommendationPriorityHigh,
				Title:       "体重管理饮食计划",
				Description: "您的BMI偏高，建议通过合理饮食控制体重",
				Actions: []string{
					"控制每日总热量摄入",
					"增加蛋白质比例",
					"多吃蔬菜和水果",
					"减少高热量零食",
				},
				Benefits: []string{
					"健康减重",
					"降低慢性病风险",
					"提高代谢率",
					"改善身体机能",
				},
				Duration:  "3-6个月",
				Frequency: "每餐",
				Tags:      []string{"减重", "热量控制", "营养均衡"},
				CreatedAt: time.Now(),
				IsActive:  true,
			})
		}
	}

	return recommendations
}

// generateSleepRecommendations 生成睡眠建议
func (s *HealthRecommendationService) generateSleepRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 获取睡眠数据
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	avgSleep := s.calculateAverageValue(sleepData)

	// 基于睡眠时长生成建议
	if avgSleep < 7 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeSleep,
			Priority:    RecommendationPriorityHigh,
			Title:       "改善睡眠质量",
			Description: "您的睡眠时间不足，建议改善睡眠习惯以获得更好的休息",
			Actions: []string{
				"建立规律的睡眠时间",
				"睡前1小时避免电子设备",
				"创造舒适的睡眠环境",
				"避免睡前大量进食",
			},
			Benefits: []string{
				"提高免疫力",
				"改善记忆力",
				"调节情绪",
				"促进身体恢复",
			},
			Duration:  "持续进行",
			Frequency: "每晚",
			Tags:      []string{"睡眠质量", "作息规律", "睡眠环境"},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	return recommendations
}

// generateStressRecommendations 生成压力管理建议
func (s *HealthRecommendationService) generateStressRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 获取压力水平数据
	stressData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "stress_level", startTime, endTime)
	avgStress := s.calculateAverageValue(stressData)

	// 基于压力水平生成建议
	if avgStress > 7 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeStress,
			Priority:    RecommendationPriorityHigh,
			Title:       "压力管理方案",
			Description: "您的压力水平较高，建议采取措施来管理和缓解压力",
			Actions: []string{
				"练习深呼吸和冥想",
				"定期进行放松活动",
				"保持工作生活平衡",
				"寻求专业心理支持",
			},
			Benefits: []string{
				"降低焦虑水平",
				"改善睡眠质量",
				"提高工作效率",
				"增强心理健康",
			},
			Duration:  "每天15-30分钟",
			Frequency: "每天",
			Tags:      []string{"压力管理", "冥想", "放松", "心理健康"},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	return recommendations
}

// generateMedicalRecommendations 生成医疗建议
func (s *HealthRecommendationService) generateMedicalRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	if profile != nil {
		age := profile.GetAge()
		
		// 基于年龄生成体检建议
		if age > 40 {
			recommendations = append(recommendations, HealthRecommendation{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        RecommendationTypeMedical,
				Priority:    RecommendationPriorityMedium,
				Title:       "定期健康体检",
				Description: "建议定期进行健康体检，及早发现和预防疾病",
				Actions: []string{
					"每年进行一次全面体检",
					"定期检查血压、血糖、血脂",
					"进行癌症筛查",
					"关注心血管健康",
				},
				Benefits: []string{
					"早期发现疾病",
					"预防慢性病",
					"监控健康状况",
					"制定个性化健康计划",
				},
				Duration:  "1-2小时",
				Frequency: "每年",
				Tags:      []string{"体检", "预防", "筛查", "健康监控"},
				CreatedAt: time.Now(),
				IsActive:  true,
			})
		}
	}

	return recommendations
}

// generateLifestyleRecommendations 生成生活方式建议
func (s *HealthRecommendationService) generateLifestyleRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	recommendations = append(recommendations, HealthRecommendation{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        RecommendationTypeLifestyle,
		Priority:    RecommendationPriorityMedium,
		Title:       "健康生活方式",
		Description: "建立健康的生活习惯，提升整体生活质量",
		Actions: []string{
			"戒烟限酒",
			"保持积极心态",
			"培养兴趣爱好",
			"维护社交关系",
		},
		Benefits: []string{
			"提高生活质量",
			"增强幸福感",
			"延长健康寿命",
			"改善心理状态",
		},
		Duration:  "长期坚持",
		Frequency: "每天",
		Tags:      []string{"生活方式", "心理健康", "社交", "兴趣爱好"},
		CreatedAt: time.Now(),
		IsActive:  true,
	})

	return recommendations
}

// generatePreventionRecommendations 生成预防建议
func (s *HealthRecommendationService) generatePreventionRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	recommendations = append(recommendations, HealthRecommendation{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        RecommendationTypePrevention,
		Priority:    RecommendationPriorityMedium,
		Title:       "疾病预防措施",
		Description: "采取预防措施，降低疾病风险",
		Actions: []string{
			"接种疫苗",
			"注意个人卫生",
			"避免危险行为",
			"定期健康监测",
		},
		Benefits: []string{
			"降低感染风险",
			"预防慢性病",
			"提高免疫力",
			"保护家人健康",
		},
		Duration:  "持续进行",
		Frequency: "根据需要",
		Tags:      []string{"预防", "疫苗", "卫生", "健康监测"},
		CreatedAt: time.Now(),
		IsActive:  true,
	})

	return recommendations
}

// generatePersonalizedTips 生成个性化提示
func (s *HealthRecommendationService) generatePersonalizedTips(userID uuid.UUID, category RecommendationType, limit int, profile *domain.HealthProfile) []HealthTip {
	var tips []HealthTip

	// 根据类别生成不同的提示
	switch category {
	case RecommendationTypeExercise:
		tips = s.getExerciseTips(limit)
	case RecommendationTypeDiet:
		tips = s.getDietTips(limit)
	case RecommendationTypeSleep:
		tips = s.getSleepTips(limit)
	case RecommendationTypeStress:
		tips = s.getStressTips(limit)
	default:
		tips = s.getGeneralTips(limit)
	}

	return tips
}

// getExerciseTips 获取运动提示
func (s *HealthRecommendationService) getExerciseTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "每天走路10000步",
			Content:       "研究表明，每天走路10000步可以显著改善心血管健康，降低慢性病风险。",
			Category:      "exercise",
			Difficulty:    "easy",
			EstimatedTime: "60-90分钟",
			Tags:          []string{"步行", "心血管", "日常运动"},
		},
		{
			ID:            uuid.New(),
			Title:         "力量训练的重要性",
			Content:       "每周进行2-3次力量训练可以增强肌肉力量，提高骨密度，改善新陈代谢。",
			Category:      "exercise",
			Difficulty:    "medium",
			EstimatedTime: "30-45分钟",
			Tags:          []string{"力量训练", "肌肉", "骨密度"},
		},
		{
			ID:            uuid.New(),
			Title:         "拉伸运动的好处",
			Content:       "定期拉伸可以提高柔韧性，减少运动伤害，缓解肌肉紧张。",
			Category:      "exercise",
			Difficulty:    "easy",
			EstimatedTime: "10-15分钟",
			Tags:          []string{"拉伸", "柔韧性", "恢复"},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getDietTips 获取饮食提示
func (s *HealthRecommendationService) getDietTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "多吃蔬菜水果",
			Content:       "每天至少摄入5份蔬菜水果，可以提供丰富的维生素、矿物质和膳食纤维。",
			Category:      "diet",
			Difficulty:    "easy",
			EstimatedTime: "每餐",
			Tags:          []string{"蔬菜", "水果", "营养"},
		},
		{
			ID:            uuid.New(),
			Title:         "控制盐分摄入",
			Content:       "每日盐分摄入量应控制在6克以下，有助于预防高血压和心血管疾病。",
			Category:      "diet",
			Difficulty:    "medium",
			EstimatedTime: "每餐",
			Tags:          []string{"盐分", "血压", "心血管"},
		},
		{
			ID:            uuid.New(),
			Title:         "适量饮水",
			Content:       "每天饮水1.5-2升，有助于维持身体水分平衡，促进新陈代谢。",
			Category:      "diet",
			Difficulty:    "easy",
			EstimatedTime: "全天",
			Tags:          []string{"饮水", "水分", "代谢"},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getSleepTips 获取睡眠提示
func (s *HealthRecommendationService) getSleepTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "保持规律作息",
			Content:       "每天在相同时间睡觉和起床，有助于调节生物钟，提高睡眠质量。",
			Category:      "sleep",
			Difficulty:    "medium",
			EstimatedTime: "每天",
			Tags:          []string{"作息", "生物钟", "睡眠质量"},
		},
		{
			ID:            uuid.New(),
			Title:         "睡前放松",
			Content:       "睡前1小时避免使用电子设备，可以进行阅读、冥想等放松活动。",
			Category:      "sleep",
			Difficulty:    "easy",
			EstimatedTime: "1小时",
			Tags:          []string{"放松", "电子设备", "睡前准备"},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getStressTips 获取压力管理提示
func (s *HealthRecommendationService) getStressTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "深呼吸练习",
			Content:       "当感到压力时，进行深呼吸练习可以快速缓解紧张情绪，平静心情。",
			Category:      "stress",
			Difficulty:    "easy",
			EstimatedTime: "5-10分钟",
			Tags:          []string{"深呼吸", "放松", "情绪管理"},
		},
		{
			ID:            uuid.New(),
			Title:         "时间管理",
			Content:       "合理安排时间，设定优先级，避免过度承诺，可以有效减少压力。",
			Category:      "stress",
			Difficulty:    "medium",
			EstimatedTime: "持续进行",
			Tags:          []string{"时间管理", "优先级", "压力缓解"},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getGeneralTips 获取通用提示
func (s *HealthRecommendationService) getGeneralTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "保持积极心态",
			Content:       "积极的心态有助于提高免疫力，改善身心健康，增强生活幸福感。",
			Category:      "general",
			Difficulty:    "medium",
			EstimatedTime: "每天",
			Tags:          []string{"心态", "免疫力", "幸福感"},
		},
		{
			ID:            uuid.New(),
			Title:         "定期体检",
			Content:       "定期进行健康体检，可以及早发现健康问题，采取预防措施。",
			Category:      "general",
			Difficulty:    "easy",
			EstimatedTime: "每年",
			Tags:          []string{"体检", "预防", "健康监测"},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// 辅助函数
func (s *HealthRecommendationService) calculateAverageSteps(data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}

	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	return total / float64(len(data))
}

func (s *HealthRecommendationService) calculateAverageValue(data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}

	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	return total / float64(len(data))
}

func (s *HealthRecommendationService) generateRecommendationSummary(recommendations []HealthRecommendation) string {
	if len(recommendations) == 0 {
		return "暂无个性化建议"
	}

	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, rec := range recommendations {
		switch rec.Priority {
		case RecommendationPriorityHigh:
			highCount++
		case RecommendationPriorityMedium:
			mediumCount++
		case RecommendationPriorityLow:
			lowCount++
		}
	}

	summary := fmt.Sprintf("为您生成了%d条健康建议", len(recommendations))
	if highCount > 0 {
		summary += fmt.Sprintf("，其中%d条高优先级", highCount)
	}
	if mediumCount > 0 {
		summary += fmt.Sprintf("，%d条中优先级", mediumCount)
	}
	if lowCount > 0 {
		summary += fmt.Sprintf("，%d条低优先级", lowCount)
	}

	return summary
}