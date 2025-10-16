package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
	"strconv"
	"strings"
)



// ContentService 
type ContentService struct {
	contentRepo        repositories.LearningContentRepository
	learnerRepo        repositories.LearnerRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	analyticsService   interfaces.LearningAnalyticsService
}

// NewContentService 
func NewContentService(
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	learnerRepo repositories.LearnerRepository,
	analyticsService interfaces.LearningAnalyticsService,
) *ContentService { {
	return &ContentService{
		contentRepo:        contentRepo,
		learnerRepo:        learnerRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		analyticsService:   analyticsService,
	}
}

// stringToDifficultyLevel DifficultyLevel
func stringToDifficultyLevel(s string) entities.DifficultyLevel {
	s = strings.ToLower(strings.TrimSpace(s))
	
	// ?
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
	
	// 
	switch s {
	case "beginner", "?:
		return entities.DifficultyBeginner
	case "elementary", "":
		return entities.DifficultyElementary
	case "intermediate", "":
		return entities.DifficultyIntermediate
	case "advanced", "":
		return entities.DifficultyAdvanced
	case "expert", "":
		return entities.DifficultyExpert
	default:
		return entities.DifficultyBeginner // ?
	}
}

// stringToContentType ContentType
func stringToContentType(s string) entities.ContentType {
	s = strings.ToLower(strings.TrimSpace(s))
	
	switch s {
	case "video", "":
		return entities.ContentTypeVideo
	case "text", "":
		return entities.ContentTypeText
	case "audio", "":
		return entities.ContentTypeAudio
	case "interactive", "?:
		return entities.ContentTypeInteractive
	case "quiz", "":
		return entities.ContentTypeQuiz
	case "exercise", "":
		return entities.ContentTypeExercise
	case "simulation", "":
		return entities.ContentTypeSimulation
	case "game", "":
		return entities.ContentTypeGame
	default:
		return entities.ContentTypeText // ?
	}
}

// CreateContentRequest 
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

// LearningObjectiveRequest 
type LearningObjectiveRequest struct {
	Description string `json:"description" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Priority    int    `json:"priority" validate:"min=1,max=5"`
}

// MediaResourceRequest 
type MediaResourceRequest struct {
	Type        string `json:"type" validate:"required"`
	URL         string `json:"url" validate:"required,url"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Duration    int    `json:"duration"` // seconds
	FileSize    int64  `json:"file_size"`
	Format      string `json:"format"`
}

// QuizQuestionRequest 
type QuizQuestionRequest struct {
	Question      string   `json:"question" validate:"required"`
	Type          string   `json:"type" validate:"required"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer" validate:"required"`
	Explanation   string   `json:"explanation"`
	Points        int      `json:"points" validate:"min=1"`
	Difficulty    string   `json:"difficulty" validate:"required"`
}

// AccessibilityInfoRequest ?
type AccessibilityInfoRequest struct {
	HasCaptions     bool     `json:"has_captions"`
	HasTranscript   bool     `json:"has_transcript"`
	HasAudioDesc    bool     `json:"has_audio_desc"`
	SupportedLangs  []string `json:"supported_langs"`
	ColorContrast   string   `json:"color_contrast"`
	FontSizeOptions []string `json:"font_size_options"`
}

// ContentResponse 
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

// UpdateContentRequest 
type UpdateContentRequest struct {
	Title              *string                       `json:"title,omitempty"`
	Description        *string                       `json:"description,omitempty"`
	Difficulty         *string                       `json:"difficulty,omitempty"`
	EstimatedDuration  *int                          `json:"estimated_duration,omitempty"`
	Tags               []string                      `json:"tags,omitempty"`
	LearningObjectives []LearningObjectiveRequest    `json:"learning_objectives,omitempty"`
	AccessibilityInfo  *AccessibilityInfoRequest     `json:"accessibility_info,omitempty"`
}

// ContentSearchRequest 
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

// ContentSearchResponse 
type ContentSearchResponse struct {
	Contents   []*ContentResponse `json:"contents"`
	Total      int                `json:"total"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
}

// PersonalizedContentRequest 
type PersonalizedContentRequest struct {
	LearnerID          uuid.UUID `json:"learner_id" validate:"required"`
	MaxRecommendations int       `json:"max_recommendations"`
	IncludeCompleted   bool      `json:"include_completed"`
	FocusAreas         []string  `json:"focus_areas"`
}

// SimpleContentRecommendation ?
type SimpleContentRecommendation struct {
	ContentID     uuid.UUID     `json:"content_id"`
	Title         string        `json:"title"`
	Type          string        `json:"type"`
	Difficulty    string        `json:"difficulty"`
	Relevance     float64       `json:"relevance"`
	Reason        string        `json:"reason"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// CreateContent 
func (s *ContentService) CreateContent(ctx context.Context, req *CreateContentRequest) (*ContentResponse, error) {
	// 
	content := entities.NewLearningContent(
		req.Title,
		req.Description,
		stringToContentType(req.Type),
		stringToDifficultyLevel(req.Difficulty),
		req.AuthorID,
	)
	
	// 
	content.EstimatedDuration = req.EstimatedDuration

	// 
	content.KnowledgeNodeIDs = req.KnowledgeNodeIDs
	content.Tags = req.Tags
	content.Prerequisites = req.Prerequisites

	// 
	for _, objReq := range req.LearningObjectives {
		content.AddLearningObjective(
			objReq.Description,
			objReq.Type,
			objReq.Priority > 3, // ??
			[]string{},          // 
		)
	}

	// 
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

	// 
	for _, quizReq := range req.QuizQuestions {
		// 
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

	// ?
	if req.AccessibilityInfo.HasCaptions || req.AccessibilityInfo.HasTranscript {
		accessInfo := entities.AccessibilityInfo{
			HasCaptions:   req.AccessibilityInfo.HasCaptions,
			HasTranscript: req.AccessibilityInfo.HasTranscript,
			HasAudioDesc:  req.AccessibilityInfo.HasAudioDesc,
			HighContrast:  req.AccessibilityInfo.ColorContrast == "high",
			LargeText:     len(req.AccessibilityInfo.FontSizeOptions) > 0,
			KeyboardNav:   true, // 
			ScreenReader:  true, // ?
		}
		content.Metadata.Accessibility = accessInfo
	}

	// 浽
	if err := s.contentRepo.Create(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	return s.buildContentResponse(content), nil
}

// GetContent 
func (s *ContentService) GetContent(ctx context.Context, contentID uuid.UUID) (*ContentResponse, error) {
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	return s.buildContentResponse(content), nil
}

// ListContent 
func (s *ContentService) ListContent(ctx context.Context, limit, offset int, contentType, difficulty, status string) ([]*ContentResponse, error) {
	// 
	query := &repositories.ContentSearchQuery{}
	
	// ?
	if status != "" {
		contentStatus := entities.ContentStatus(status)
		query.Status = &contentStatus
	}
	
	// 
	if contentType != "" {
		contentTypeEnum := entities.ContentType(contentType)
		query.ContentType = &contentTypeEnum
	}
	
	// 
	if difficulty != "" {
		difficultyLevel := stringToDifficultyLevel(difficulty)
		query.DifficultyLevel = &difficultyLevel
	}
	
	// 
	contents, _, err := s.contentRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list content: %w", err)
	}
	
	// 
	start := offset
	end := offset + limit
	if start > len(contents) {
		start = len(contents)
	}
	if end > len(contents) {
		end = len(contents)
	}
	
	pagedContents := contents[start:end]
	
	// 
	responses := make([]*ContentResponse, len(pagedContents))
	for i, content := range pagedContents {
		responses[i] = s.buildContentResponse(content)
	}
	
	return responses, nil
}

// UpdateContent 
func (s *ContentService) UpdateContent(ctx context.Context, contentID uuid.UUID, req *UpdateContentRequest) (*ContentResponse, error) {
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// 
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

	// 
	if req.LearningObjectives != nil {
		content.LearningObjectives = []entities.LearningObjective{}
		for _, objReq := range req.LearningObjectives {
			content.AddLearningObjective(
				objReq.Description,
				objReq.Type,
				objReq.Priority > 3, // ??
				[]string{},          // 
			)
		}
	}

	// ?
	if req.AccessibilityInfo != nil {
		content.Metadata.Accessibility.HasCaptions = req.AccessibilityInfo.HasCaptions
		content.Metadata.Accessibility.HasTranscript = req.AccessibilityInfo.HasTranscript
		content.Metadata.Accessibility.HasAudioDesc = req.AccessibilityInfo.HasAudioDesc
		content.Metadata.Accessibility.HighContrast = req.AccessibilityInfo.ColorContrast == "high"
		content.Metadata.Accessibility.LargeText = len(req.AccessibilityInfo.FontSizeOptions) > 0
		content.Metadata.Accessibility.KeyboardNav = true // 
		content.Metadata.Accessibility.ScreenReader = true // ?
	}

	content.UpdatedAt = time.Now()

	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to update content: %w", err)
	}

	return s.buildContentResponse(content), nil
}

// DeleteContent 
func (s *ContentService) DeleteContent(ctx context.Context, contentID uuid.UUID) error {
	if err := s.contentRepo.Delete(ctx, contentID); err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}
	return nil
}

// PublishContent 
func (s *ContentService) PublishContent(ctx context.Context, contentID uuid.UUID) (*ContentResponse, error) {
	if err := s.contentRepo.PublishContent(ctx, contentID); err != nil {
		return nil, fmt.Errorf("failed to publish content: %w", err)
	}
	
	// 
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content: %w", err)
	}
	
	return s.buildContentResponse(content), nil
}

// ArchiveContent 鵵
func (s *ContentService) ArchiveContent(ctx context.Context, contentID uuid.UUID) (*ContentResponse, error) {
	if err := s.contentRepo.ArchiveContent(ctx, contentID); err != nil {
		return nil, fmt.Errorf("failed to archive content: %w", err)
	}
	
	// 
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content: %w", err)
	}
	
	return s.buildContentResponse(content), nil
}

// SearchContent 
func (s *ContentService) SearchContent(ctx context.Context, req *ContentSearchRequest) (*ContentSearchResponse, error) {
	// 
	query := &repositories.ContentSearchQuery{
		Keywords:       []string{req.Query},
		Tags:           req.Tags,
		AuthorID:       req.AuthorID,
	}

	// ?
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

	// 
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

// GetPersonalizedContent 
func (s *ContentService) GetPersonalizedContent(ctx context.Context, req *PersonalizedContentRequest) ([]*SimpleContentRecommendation, error) {
	// ?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 
	preferences := &repositories.ContentPreferences{
		PreferredTypes:      []entities.ContentType{},
		PreferredDifficulty: entities.DifficultyIntermediate,
		PreferredDuration:   time.Hour,
		InterestAreas:       req.FocusAreas,
	}

	// 
	if learner.Preferences.DifficultyTolerance > 0 || learner.Preferences.SessionDuration > 0 {
		// LearningPreferencePreferredContentTypesPreferredDifficulty
		// DifficultyTolerance
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

	// 
	recommendations, err := s.contentRepo.GetRecommendedContent(ctx, req.LearnerID, req.MaxRecommendations)
	if err != nil {
		return nil, fmt.Errorf("failed to get personalized recommendations: %w", err)
	}

	// ?
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

// RecordContentInteraction 
func (s *ContentService) RecordContentInteraction(ctx context.Context, interaction *entities.InteractionRecord) error {
	if err := s.contentRepo.AddInteractionRecord(ctx, interaction); err != nil {
		return fmt.Errorf("failed to record interaction: %w", err)
	}

	// 
	go s.updateContentAnalytics(context.Background(), interaction.ContentID)

	return nil
}

// UpdateContentProgress 
func (s *ContentService) UpdateContentProgress(ctx context.Context, progress *entities.ContentProgress) error {
	if err := s.contentRepo.UpdateProgress(ctx, progress); err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	// 
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

// AddContentNote 
func (s *ContentService) AddContentNote(ctx context.Context, note *entities.LearningNote) error {
	note.ID = uuid.New()
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()

	if err := s.contentRepo.AddLearningNote(ctx, note); err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}
	return nil
}

// AddContentBookmark 
func (s *ContentService) AddContentBookmark(ctx context.Context, bookmark *entities.Bookmark) error {
	bookmark.ID = uuid.New()
	bookmark.CreatedAt = time.Now()

	if err := s.contentRepo.AddBookmark(ctx, bookmark); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	return nil
}

// GetContentAnalytics 
func (s *ContentService) GetContentAnalytics(ctx context.Context, contentID uuid.UUID) (*entities.ContentAnalytics, error) {
	analytics, err := s.contentRepo.GetContentAnalytics(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content analytics: %w", err)
	}
	return analytics, nil
}

// GetContentsByKnowledgeNode 
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

// GetPrerequisiteContents 
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

// GetFollowUpContents 
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

// ValidateContent 
func (s *ContentService) ValidateContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentValidation, error) {
	validation, err := s.contentRepo.ValidateContent(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate content: %w", err)
	}
	return validation, nil
}

// OptimizeContent 
func (s *ContentService) OptimizeContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentOptimization, error) {
	optimization, err := s.contentRepo.OptimizeContent(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize content: %w", err)
	}
	return optimization, nil
}

// 

func (s *ContentService) buildContentResponse(content *entities.LearningContent) *ContentResponse {
	// LearningObjectives?
	learningObjectives := make([]*entities.LearningObjective, len(content.LearningObjectives))
	for i := range content.LearningObjectives {
		learningObjectives[i] = &content.LearningObjectives[i]
	}
	
	// MediaResources?
	mediaResources := make([]*entities.MediaResource, len(content.MediaResources))
	for i := range content.MediaResources {
		mediaResources[i] = &content.MediaResources[i]
	}
	
	// QuizQuestions?
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
	// 
	analytics, err := s.contentRepo.GetContentAnalytics(ctx, contentID)
	if err != nil {
		fmt.Printf("Warning: failed to get content analytics for update: %v\n", err)
		return
	}

	// 
	analytics.LastUpdated = time.Now()
	
	if err := s.contentRepo.UpdateContentAnalytics(ctx, contentID, analytics); err != nil {
		fmt.Printf("Warning: failed to update content analytics: %v\n", err)
	}
}

// difficultyLevelToString DifficultyLevel
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

// stringToDifficultyInt ?
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
		return 1 // ?
	}
}

