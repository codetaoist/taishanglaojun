package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
	"strconv"
	"strings"
)



// ContentService 学习内容应用服务
type ContentService struct {
	contentRepo        repositories.LearningContentRepository
	learnerRepo        repositories.LearnerRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	analyticsService   LearningAnalyticsService
}

// NewContentService 创建新的学习内容应用服务
func NewContentService(
	contentRepo repositories.LearningContentRepository,
	learnerRepo repositories.LearnerRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	analyticsService LearningAnalyticsService,
) *ContentService {
	return &ContentService{
		contentRepo:        contentRepo,
		learnerRepo:        learnerRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		analyticsService:   analyticsService,
	}
}

// stringToDifficultyLevel 将字符串转换为DifficultyLevel
func stringToDifficultyLevel(s string) entities.DifficultyLevel {
	s = strings.ToLower(strings.TrimSpace(s))
	
	// 尝试按数字解析
	if num, err := strconv.Atoi(s); err == nil {
		switch num {
		case 1:
			return entities.DifficultyBeginner
		case 2:
			return entities.DifficultyElementary
		case 3:
			return entities.DifficultyIntermediate
		case 4:
			return entities.DifficultyAdvanced
		case 5:
			return entities.DifficultyExpert
		}
	}
	
	// 按字符串解析
	switch s {
	case "beginner", "初学者":
		return entities.DifficultyBeginner
	case "elementary", "基础":
		return entities.DifficultyElementary
	case "intermediate", "中级":
		return entities.DifficultyIntermediate
	case "advanced", "高级":
		return entities.DifficultyAdvanced
	case "expert", "专家":
		return entities.DifficultyExpert
	default:
		return entities.DifficultyBeginner // 默认为初学者
	}
}

// stringToContentType 将字符串转换为ContentType
func stringToContentType(s string) entities.ContentType {
	s = strings.ToLower(strings.TrimSpace(s))
	
	switch s {
	case "video", "视频":
		return entities.ContentTypeVideo
	case "text", "文本":
		return entities.ContentTypeText
	case "audio", "音频":
		return entities.ContentTypeAudio
	case "interactive", "交互式":
		return entities.ContentTypeInteractive
	case "quiz", "测验":
		return entities.ContentTypeQuiz
	case "exercise", "练习":
		return entities.ContentTypeExercise
	case "simulation", "模拟":
		return entities.ContentTypeSimulation
	case "game", "游戏":
		return entities.ContentTypeGame
	default:
		return entities.ContentTypeText // 默认为文本
	}
}

// CreateContentRequest 创建内容请求
type CreateContentRequest struct {
	Title              string                        `json:"title" validate:"required,min=5,max=200"`
	Description        string                        `json:"description" validate:"required,min=10,max=1000"`
	Type               string                        `json:"type" validate:"required"`
	Difficulty         string                        `json:"difficulty" validate:"required"`
	EstimatedDuration  int                           `json:"estimated_duration"` // minutes
	AuthorID           uuid.UUID                     `json:"author_id" validate:"required"`
	KnowledgeNodeIDs   []uuid.UUID                   `json:"knowledge_node_ids"`
	Tags               []string                      `json:"tags"`
	LearningObjectives []LearningObjectiveRequest    `json:"learning_objectives"`
	MediaResources     []MediaResourceRequest        `json:"media_resources"`
	QuizQuestions      []QuizQuestionRequest         `json:"quiz_questions"`
	Prerequisites      []uuid.UUID                   `json:"prerequisites"`
	AccessibilityInfo  AccessibilityInfoRequest      `json:"accessibility_info"`
}

// LearningObjectiveRequest 学习目标请求
type LearningObjectiveRequest struct {
	Description string `json:"description" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Priority    int    `json:"priority" validate:"min=1,max=5"`
}

// MediaResourceRequest 媒体资源请求
type MediaResourceRequest struct {
	Type        string `json:"type" validate:"required"`
	URL         string `json:"url" validate:"required,url"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Duration    int    `json:"duration"` // seconds
	FileSize    int64  `json:"file_size"`
	Format      string `json:"format"`
}

// QuizQuestionRequest 测验问题请求
type QuizQuestionRequest struct {
	Question      string   `json:"question" validate:"required"`
	Type          string   `json:"type" validate:"required"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer" validate:"required"`
	Explanation   string   `json:"explanation"`
	Points        int      `json:"points" validate:"min=1"`
	Difficulty    string   `json:"difficulty" validate:"required"`
}

// AccessibilityInfoRequest 可访问性信息请求
type AccessibilityInfoRequest struct {
	HasCaptions     bool     `json:"has_captions"`
	HasTranscript   bool     `json:"has_transcript"`
	HasAudioDesc    bool     `json:"has_audio_desc"`
	SupportedLangs  []string `json:"supported_langs"`
	ColorContrast   string   `json:"color_contrast"`
	FontSizeOptions []string `json:"font_size_options"`
}

// ContentResponse 内容响应
type ContentResponse struct {
	ID                 uuid.UUID                      `json:"id"`
	Title              string                         `json:"title"`
	Description        string                         `json:"description"`
	Type               string                         `json:"type"`
	Status             string                         `json:"status"`
	Difficulty         string                         `json:"difficulty"`
	EstimatedDuration  time.Duration                  `json:"estimated_duration"`
	AuthorID           uuid.UUID                      `json:"author_id"`
	KnowledgeNodeIDs   []uuid.UUID                    `json:"knowledge_node_ids"`
	Tags               []string                       `json:"tags"`
	LearningObjectives []*entities.LearningObjective  `json:"learning_objectives"`
	MediaResources     []*entities.MediaResource      `json:"media_resources"`
	QuizQuestions      []*entities.QuizQuestion       `json:"quiz_questions"`
	Prerequisites      []uuid.UUID                    `json:"prerequisites"`
	AccessibilityInfo  *entities.AccessibilityInfo    `json:"accessibility_info"`
	Analytics          *entities.ContentAnalytics     `json:"analytics"`
	CreatedAt          time.Time                      `json:"created_at"`
	UpdatedAt          time.Time                      `json:"updated_at"`
	PublishedAt        *time.Time                     `json:"published_at,omitempty"`
}

// UpdateContentRequest 更新内容请求
type UpdateContentRequest struct {
	Title              *string                       `json:"title,omitempty"`
	Description        *string                       `json:"description,omitempty"`
	Difficulty         *string                       `json:"difficulty,omitempty"`
	EstimatedDuration  *int                          `json:"estimated_duration,omitempty"`
	Tags               []string                      `json:"tags,omitempty"`
	LearningObjectives []LearningObjectiveRequest    `json:"learning_objectives,omitempty"`
	AccessibilityInfo  *AccessibilityInfoRequest     `json:"accessibility_info,omitempty"`
}

// ContentSearchRequest 内容搜索请求
type ContentSearchRequest struct {
	Query          string      `json:"query"`
	Type           string      `json:"type"`
	Difficulty     string      `json:"difficulty"`
	Tags           []string    `json:"tags"`
	AuthorID       *uuid.UUID  `json:"author_id,omitempty"`
	KnowledgeNodes []uuid.UUID `json:"knowledge_nodes"`
	MinDuration    *int        `json:"min_duration,omitempty"`
	MaxDuration    *int        `json:"max_duration,omitempty"`
	Status         string      `json:"status"`
	Limit          int         `json:"limit"`
	Offset         int         `json:"offset"`
	SortBy         string      `json:"sort_by"`
	SortOrder      string      `json:"sort_order"`
}

// ContentSearchResponse 内容搜索响应
type ContentSearchResponse struct {
	Contents   []*ContentResponse `json:"contents"`
	Total      int                `json:"total"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
}

// PersonalizedContentRequest 个性化内容请求
type PersonalizedContentRequest struct {
	LearnerID          uuid.UUID `json:"learner_id" validate:"required"`
	MaxRecommendations int       `json:"max_recommendations"`
	IncludeCompleted   bool      `json:"include_completed"`
	FocusAreas         []string  `json:"focus_areas"`
}

// SimpleContentRecommendation 简单内容推荐响应
type SimpleContentRecommendation struct {
	ContentID     uuid.UUID     `json:"content_id"`
	Title         string        `json:"title"`
	Type          string        `json:"type"`
	Difficulty    string        `json:"difficulty"`
	Relevance     float64       `json:"relevance"`
	Reason        string        `json:"reason"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// CreateContent 创建学习内容
func (s *ContentService) CreateContent(ctx context.Context, req *CreateContentRequest) (*ContentResponse, error) {
	// 创建内容实体
	content := entities.NewLearningContent(
		req.Title,
		req.Description,
		stringToContentType(req.Type),
		stringToDifficultyLevel(req.Difficulty),
		req.AuthorID,
	)
	
	// 设置预估学习时间
	content.EstimatedDuration = req.EstimatedDuration

	// 设置知识节点关联
	content.KnowledgeNodeIDs = req.KnowledgeNodeIDs
	content.Tags = req.Tags
	content.Prerequisites = req.Prerequisites

	// 添加学习目标
	for _, objReq := range req.LearningObjectives {
		content.AddLearningObjective(
			objReq.Description,
			objReq.Type,
			objReq.Priority > 3, // 优先级大于3视为可测量
			[]string{},          // 默认空的评估标准
		)
	}

	// 添加媒体资源
	for _, mediaReq := range req.MediaResources {
		content.AddMediaResource(
			mediaReq.Type,
			mediaReq.URL,
			mediaReq.Title,
			mediaReq.Duration,
			mediaReq.FileSize,
			mediaReq.Format,
		)
	}

	// 添加测验问题
	for _, quizReq := range req.QuizQuestions {
		// 转换难度字符串为整数
		difficultyInt := s.stringToDifficultyInt(quizReq.Difficulty)
		
		content.AddQuizQuestion(
			quizReq.Type,
			quizReq.Question,
			quizReq.Options,
			quizReq.CorrectAnswer,
			quizReq.Explanation,
			quizReq.Points,
			difficultyInt,
		)
	}

	// 设置可访问性信息
	if req.AccessibilityInfo.HasCaptions || req.AccessibilityInfo.HasTranscript {
		accessInfo := entities.AccessibilityInfo{
			HasCaptions:   req.AccessibilityInfo.HasCaptions,
			HasTranscript: req.AccessibilityInfo.HasTranscript,
			HasAudioDesc:  req.AccessibilityInfo.HasAudioDesc,
			HighContrast:  req.AccessibilityInfo.ColorContrast == "high",
			LargeText:     len(req.AccessibilityInfo.FontSizeOptions) > 0,
			KeyboardNav:   true, // 默认支持键盘导航
			ScreenReader:  true, // 默认支持屏幕阅读器
		}
		content.Metadata.Accessibility = accessInfo
	}

	// 保存到数据库
	if err := s.contentRepo.Create(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	return s.buildContentResponse(content), nil
}

// GetContent 获取内容详情
func (s *ContentService) GetContent(ctx context.Context, contentID uuid.UUID) (*ContentResponse, error) {
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	return s.buildContentResponse(content), nil
}

// ListContent 列出内容
func (s *ContentService) ListContent(ctx context.Context, limit, offset int, contentType, difficulty, status string) ([]*ContentResponse, error) {
	// 构建查询条件
	query := &repositories.ContentSearchQuery{}
	
	// 设置状态过滤
	if status != "" {
		contentStatus := entities.ContentStatus(status)
		query.Status = &contentStatus
	}
	
	// 设置类型过滤
	if contentType != "" {
		contentTypeEnum := entities.ContentType(contentType)
		query.ContentType = &contentTypeEnum
	}
	
	// 设置难度过滤
	if difficulty != "" {
		difficultyLevel := stringToDifficultyLevel(difficulty)
		query.DifficultyLevel = &difficultyLevel
	}
	
	// 执行查询
	contents, _, err := s.contentRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list content: %w", err)
	}
	
	// 应用分页
	start := offset
	end := offset + limit
	if start > len(contents) {
		start = len(contents)
	}
	if end > len(contents) {
		end = len(contents)
	}
	
	pagedContents := contents[start:end]
	
	// 构建响应
	responses := make([]*ContentResponse, len(pagedContents))
	for i, content := range pagedContents {
		responses[i] = s.buildContentResponse(content)
	}
	
	return responses, nil
}

// UpdateContent 更新内容
func (s *ContentService) UpdateContent(ctx context.Context, contentID uuid.UUID, req *UpdateContentRequest) (*ContentResponse, error) {
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// 更新字段
	if req.Title != nil {
		content.Title = *req.Title
	}
	if req.Description != nil {
		content.Description = *req.Description
	}
	if req.Difficulty != nil {
		content.Difficulty = stringToDifficultyLevel(*req.Difficulty)
	}
	if req.EstimatedDuration != nil {
		content.EstimatedDuration = *req.EstimatedDuration
	}
	if req.Tags != nil {
		content.Tags = req.Tags
	}

	// 更新学习目标
	if req.LearningObjectives != nil {
		content.LearningObjectives = []entities.LearningObjective{}
		for _, objReq := range req.LearningObjectives {
			content.AddLearningObjective(
				objReq.Description,
				objReq.Type,
				objReq.Priority > 3, // 优先级大于3视为可测量
				[]string{},          // 默认空的评估标准
			)
		}
	}

	// 更新可访问性信息
	if req.AccessibilityInfo != nil {
		content.Metadata.Accessibility.HasCaptions = req.AccessibilityInfo.HasCaptions
		content.Metadata.Accessibility.HasTranscript = req.AccessibilityInfo.HasTranscript
		content.Metadata.Accessibility.HasAudioDesc = req.AccessibilityInfo.HasAudioDesc
		content.Metadata.Accessibility.HighContrast = req.AccessibilityInfo.ColorContrast == "high"
		content.Metadata.Accessibility.LargeText = len(req.AccessibilityInfo.FontSizeOptions) > 0
		content.Metadata.Accessibility.KeyboardNav = true // 默认支持键盘导航
		content.Metadata.Accessibility.ScreenReader = true // 默认支持屏幕阅读器
	}

	content.UpdatedAt = time.Now()

	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to update content: %w", err)
	}

	return s.buildContentResponse(content), nil
}

// DeleteContent 删除内容
func (s *ContentService) DeleteContent(ctx context.Context, contentID uuid.UUID) error {
	if err := s.contentRepo.Delete(ctx, contentID); err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}
	return nil
}

// PublishContent 发布内容
func (s *ContentService) PublishContent(ctx context.Context, contentID uuid.UUID) (*ContentResponse, error) {
	if err := s.contentRepo.PublishContent(ctx, contentID); err != nil {
		return nil, fmt.Errorf("failed to publish content: %w", err)
	}
	
	// 获取更新后的内容
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content: %w", err)
	}
	
	return s.buildContentResponse(content), nil
}

// ArchiveContent 归档内容
func (s *ContentService) ArchiveContent(ctx context.Context, contentID uuid.UUID) (*ContentResponse, error) {
	if err := s.contentRepo.ArchiveContent(ctx, contentID); err != nil {
		return nil, fmt.Errorf("failed to archive content: %w", err)
	}
	
	// 获取更新后的内容
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content: %w", err)
	}
	
	return s.buildContentResponse(content), nil
}

// SearchContent 搜索内容
func (s *ContentService) SearchContent(ctx context.Context, req *ContentSearchRequest) (*ContentSearchResponse, error) {
	// 构建搜索查询
	query := &repositories.ContentSearchQuery{
		Keywords:       []string{req.Query},
		Tags:           req.Tags,
		AuthorID:       req.AuthorID,
	}

	// 转换状态字符串为枚举类型
	if req.Status != "" {
		status := entities.ContentStatus(req.Status)
		query.Status = &status
	}

	if req.MinDuration != nil {
		minDur := time.Duration(*req.MinDuration) * time.Minute
		query.MinDuration = &minDur
	}
	if req.MaxDuration != nil {
		maxDur := time.Duration(*req.MaxDuration) * time.Minute
		query.MaxDuration = &maxDur
	}

	contents, total, err := s.contentRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search content: %w", err)
	}

	// 构建响应
	contentResponses := make([]*ContentResponse, len(contents))
	for i, content := range contents {
		contentResponses[i] = s.buildContentResponse(content)
	}

	return &ContentSearchResponse{
		Contents: contentResponses,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
		HasMore:  req.Offset+len(contents) < total,
	}, nil
}

// GetPersonalizedContent 获取个性化内容推荐
func (s *ContentService) GetPersonalizedContent(ctx context.Context, req *PersonalizedContentRequest) ([]*SimpleContentRecommendation, error) {
	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 构建内容偏好
	preferences := &repositories.ContentPreferences{
		PreferredTypes:      []entities.ContentType{},
		PreferredDifficulty: entities.DifficultyIntermediate,
		PreferredDuration:   time.Hour,
		InterestAreas:       req.FocusAreas,
	}

	// 检查学习者偏好是否已设置（非零值）
	if learner.Preferences.DifficultyTolerance > 0 || learner.Preferences.SessionDuration > 0 {
		// LearningPreference结构体没有PreferredContentTypes和PreferredDifficulty字段
		// 根据DifficultyTolerance设置难度偏好
		if learner.Preferences.DifficultyTolerance <= 0.2 {
			preferences.PreferredDifficulty = entities.DifficultyBeginner
		} else if learner.Preferences.DifficultyTolerance <= 0.4 {
			preferences.PreferredDifficulty = entities.DifficultyElementary
		} else if learner.Preferences.DifficultyTolerance <= 0.6 {
			preferences.PreferredDifficulty = entities.DifficultyIntermediate
		} else if learner.Preferences.DifficultyTolerance <= 0.8 {
			preferences.PreferredDifficulty = entities.DifficultyAdvanced
		} else {
			preferences.PreferredDifficulty = entities.DifficultyExpert
		}
		
		if learner.Preferences.SessionDuration > 0 {
			preferences.PreferredDuration = time.Duration(learner.Preferences.SessionDuration) * time.Minute
		}
	}

	// 获取推荐内容
	recommendations, err := s.contentRepo.GetRecommendedContent(ctx, req.LearnerID, req.MaxRecommendations)
	if err != nil {
		return nil, fmt.Errorf("failed to get personalized recommendations: %w", err)
	}

	// 转换为响应格式
	result := make([]*SimpleContentRecommendation, len(recommendations))
	for i, rec := range recommendations {
		result[i] = &SimpleContentRecommendation{
			ContentID:     rec.Content.ID,
			Title:         rec.Content.Title,
			Type:          string(rec.Content.Type),
			Difficulty:    s.difficultyLevelToString(rec.Content.Difficulty),
			Relevance:     rec.RecommendationScore,
			Reason:        strings.Join(rec.Reasoning, "; "),
			EstimatedTime: time.Duration(rec.Content.EstimatedDuration) * time.Minute,
		}
	}

	return result, nil
}

// RecordContentInteraction 记录内容交互
func (s *ContentService) RecordContentInteraction(ctx context.Context, interaction *entities.InteractionRecord) error {
	if err := s.contentRepo.AddInteractionRecord(ctx, interaction); err != nil {
		return fmt.Errorf("failed to record interaction: %w", err)
	}

	// 异步更新内容分析
	go s.updateContentAnalytics(context.Background(), interaction.ContentID)

	return nil
}

// UpdateContentProgress 更新内容进度
func (s *ContentService) UpdateContentProgress(ctx context.Context, progress *entities.ContentProgress) error {
	if err := s.contentRepo.UpdateProgress(ctx, progress); err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	// 如果完成了内容，记录学习活动
	if progress.Progress >= 1.0 {
		activity := &entities.LearningHistory{
			ID:             uuid.New(),
			LearnerID:      progress.LearnerID,
			ContentID:      progress.ContentID,
			ContentType:    "learning_content",
			Progress:       progress.Progress,
			Duration:       time.Duration(progress.TimeSpent) * time.Second,
			Completed:      true,
			Timestamp:      time.Now(),
		}

		if err := s.learnerRepo.RecordLearningActivity(ctx, progress.LearnerID, activity); err != nil {
			fmt.Printf("Warning: failed to record learning activity: %v\n", err)
		}
	}

	return nil
}

// AddContentNote 添加内容笔记
func (s *ContentService) AddContentNote(ctx context.Context, note *entities.LearningNote) error {
	note.ID = uuid.New()
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()

	if err := s.contentRepo.AddLearningNote(ctx, note); err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}
	return nil
}

// AddContentBookmark 添加内容书签
func (s *ContentService) AddContentBookmark(ctx context.Context, bookmark *entities.Bookmark) error {
	bookmark.ID = uuid.New()
	bookmark.CreatedAt = time.Now()

	if err := s.contentRepo.AddBookmark(ctx, bookmark); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	return nil
}

// GetContentAnalytics 获取内容分析
func (s *ContentService) GetContentAnalytics(ctx context.Context, contentID uuid.UUID) (*entities.ContentAnalytics, error) {
	analytics, err := s.contentRepo.GetContentAnalytics(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content analytics: %w", err)
	}
	return analytics, nil
}

// GetContentsByKnowledgeNode 根据知识节点获取内容
func (s *ContentService) GetContentsByKnowledgeNode(ctx context.Context, nodeID uuid.UUID, offset, limit int) ([]*ContentResponse, error) {
	contents, err := s.contentRepo.GetByKnowledgeNode(ctx, nodeID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by knowledge node: %w", err)
	}

	responses := make([]*ContentResponse, len(contents))
	for i, content := range contents {
		responses[i] = s.buildContentResponse(content)
	}

	return responses, nil
}

// GetPrerequisiteContents 获取前置内容
func (s *ContentService) GetPrerequisiteContents(ctx context.Context, contentID uuid.UUID) ([]*ContentResponse, error) {
	contents, err := s.contentRepo.GetPrerequisites(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prerequisite contents: %w", err)
	}

	responses := make([]*ContentResponse, len(contents))
	for i, content := range contents {
		responses[i] = s.buildContentResponse(content)
	}

	return responses, nil
}

// GetFollowUpContents 获取后续内容
func (s *ContentService) GetFollowUpContents(ctx context.Context, contentID uuid.UUID) ([]*ContentResponse, error) {
	contents, err := s.contentRepo.GetFollowUpContent(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get follow-up contents: %w", err)
	}

	responses := make([]*ContentResponse, len(contents))
	for i, content := range contents {
		responses[i] = s.buildContentResponse(content)
	}

	return responses, nil
}

// ValidateContent 验证内容质量
func (s *ContentService) ValidateContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentValidation, error) {
	validation, err := s.contentRepo.ValidateContent(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate content: %w", err)
	}
	return validation, nil
}

// OptimizeContent 优化内容
func (s *ContentService) OptimizeContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentOptimization, error) {
	optimization, err := s.contentRepo.OptimizeContent(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize content: %w", err)
	}
	return optimization, nil
}

// 私有辅助方法

func (s *ContentService) buildContentResponse(content *entities.LearningContent) *ContentResponse {
	// 转换LearningObjectives为指针切片
	learningObjectives := make([]*entities.LearningObjective, len(content.LearningObjectives))
	for i := range content.LearningObjectives {
		learningObjectives[i] = &content.LearningObjectives[i]
	}
	
	// 转换MediaResources为指针切片
	mediaResources := make([]*entities.MediaResource, len(content.MediaResources))
	for i := range content.MediaResources {
		mediaResources[i] = &content.MediaResources[i]
	}
	
	// 转换QuizQuestions为指针切片
	quizQuestions := make([]*entities.QuizQuestion, len(content.QuizQuestions))
	for i := range content.QuizQuestions {
		quizQuestions[i] = &content.QuizQuestions[i]
	}

	return &ContentResponse{
		ID:                 content.ID,
		Title:              content.Title,
		Description:        content.Description,
		Type:               string(content.Type),
		Status:             string(content.Status),
		Difficulty:         s.difficultyLevelToString(content.Difficulty),
		EstimatedDuration:  time.Duration(content.EstimatedDuration) * time.Minute,
		AuthorID:           content.AuthorID,
		KnowledgeNodeIDs:   content.KnowledgeNodeIDs,
		Tags:               content.Tags,
		LearningObjectives: learningObjectives,
		MediaResources:     mediaResources,
		QuizQuestions:      quizQuestions,
		Prerequisites:      content.Prerequisites,
		AccessibilityInfo:  &content.Metadata.Accessibility,
		Analytics:          &content.Analytics,
		CreatedAt:          content.CreatedAt,
		UpdatedAt:          content.UpdatedAt,
		PublishedAt:        content.PublishedAt,
	}
}

func (s *ContentService) updateContentAnalytics(ctx context.Context, contentID uuid.UUID) {
	// 异步更新内容分析数据
	analytics, err := s.contentRepo.GetContentAnalytics(ctx, contentID)
	if err != nil {
		fmt.Printf("Warning: failed to get content analytics for update: %v\n", err)
		return
	}

	// 更新分析数据
	analytics.LastUpdated = time.Now()
	
	if err := s.contentRepo.UpdateContentAnalytics(ctx, contentID, analytics); err != nil {
		fmt.Printf("Warning: failed to update content analytics: %v\n", err)
	}
}

// difficultyLevelToString 将DifficultyLevel转换为字符串
func (s *ContentService) difficultyLevelToString(level entities.DifficultyLevel) string {
	switch level {
	case entities.DifficultyBeginner:
		return "beginner"
	case entities.DifficultyElementary:
		return "elementary"
	case entities.DifficultyIntermediate:
		return "intermediate"
	case entities.DifficultyAdvanced:
		return "advanced"
	case entities.DifficultyExpert:
		return "expert"
	default:
		return "intermediate"
	}
}

// stringToDifficultyInt 将难度字符串转换为整数
func (s *ContentService) stringToDifficultyInt(difficulty string) int {
	switch difficulty {
	case "Beginner", "beginner", "1":
		return 1
	case "Elementary", "elementary", "2":
		return 2
	case "Intermediate", "intermediate", "3":
		return 3
	case "Advanced", "advanced", "4":
		return 4
	case "Expert", "expert", "5":
		return 5
	default:
		return 1 // 默认为初级
	}
}