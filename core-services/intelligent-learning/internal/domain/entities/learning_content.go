package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ContentType 
type ContentType string

const (
	ContentTypeVideo       ContentType = "video"       // 
	ContentTypeText        ContentType = "text"        // 
	ContentTypeAudio       ContentType = "audio"       // 
	ContentTypeInteractive ContentType = "interactive" // ?
	ContentTypeQuiz        ContentType = "quiz"        // 
	ContentTypeExercise    ContentType = "exercise"    // 
	ContentTypeSimulation  ContentType = "simulation"  // 
	ContentTypeGame        ContentType = "game"        // 
)

// ContentStatus ?
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"     // 
	ContentStatusReview    ContentStatus = "review"    // ?
	ContentStatusPublished ContentStatus = "published" // ?
	ContentStatusArchived  ContentStatus = "archived"  // ?
)

// MediaResource 
type MediaResource struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"`     // image, video, audio, document
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Duration int       `json:"duration"` // 
	Size     int64     `json:"size"`     // 
	Format   string    `json:"format"`   // mp4, pdf, jpg?
	Metadata map[string]interface{} `json:"metadata"`
}

// InteractiveElement 
type InteractiveElement struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`        // button, input, dropdown, drag_drop?
	Position    Position  `json:"position"`    // 
	Properties  map[string]interface{} `json:"properties"` // ?
	Actions     []Action  `json:"actions"`     // 
	Feedback    string    `json:"feedback"`    // 
	Points      int       `json:"points"`      // 
}

// Position 
type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Action 
type Action struct {
	Type       string                 `json:"type"`       // click, hover, input?
	Target     string                 `json:"target"`     // 
	Parameters map[string]interface{} `json:"parameters"` // 
	Response   string                 `json:"response"`   // 
}

// QuizQuestion 
type QuizQuestion struct {
	ID          uuid.UUID    `json:"id"`
	Type        string       `json:"type"`        // multiple_choice, true_false, fill_blank, essay
	Question    string       `json:"question"`
	Options     []string     `json:"options"`     // 
	CorrectAnswer interface{} `json:"correct_answer"` // 
	Explanation string       `json:"explanation"` // 
	Points      int          `json:"points"`      // ?
	Difficulty  int          `json:"difficulty"`  //  1-5
	Tags        []string     `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LearningObjective 
type LearningObjective struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Level       string    `json:"level"`       // remember, understand, apply, analyze, evaluate, create
	Measurable  bool      `json:"measurable"`  // ?
	Criteria    []string  `json:"criteria"`    // 
}

// ContentMetadata ?
type ContentMetadata struct {
	Author          string            `json:"author"`
	Version         string            `json:"version"`
	Language        string            `json:"language"`
	Keywords        []string          `json:"keywords"`
	Subject         string            `json:"subject"`
	Grade           string            `json:"grade"`
	AgeRange        string            `json:"age_range"`
	LearningStyle   []LearningStyle   `json:"learning_style"`
	Accessibility   AccessibilityInfo `json:"accessibility"`
	Copyright       string            `json:"copyright"`
	License         string            `json:"license"`
	LastReviewed    time.Time         `json:"last_reviewed"`
}

// AccessibilityInfo ?
type AccessibilityInfo struct {
	HasCaptions     bool `json:"has_captions"`     // ?
	HasTranscript   bool `json:"has_transcript"`   // 
	HasAudioDesc    bool `json:"has_audio_desc"`   // ?
	HighContrast    bool `json:"high_contrast"`    // 
	LargeText       bool `json:"large_text"`       // ?
	KeyboardNav     bool `json:"keyboard_nav"`     // 
	ScreenReader    bool `json:"screen_reader"`    // ?
}

// ContentAnalytics 
type ContentAnalytics struct {
	ViewCount       int                    `json:"view_count"`
	CompletionRate  float64                `json:"completion_rate"`
	AverageRating   float64                `json:"average_rating"`
	AverageTime     int                    `json:"average_time"`     // ?
	DropoffPoints   []int                  `json:"dropoff_points"`   // 
	InteractionData map[string]interface{} `json:"interaction_data"` // 
	FeedbackSummary FeedbackSummary        `json:"feedback_summary"`
	LastUpdated     time.Time              `json:"last_updated"`
}

// FeedbackSummary 
type FeedbackSummary struct {
	TotalFeedback   int                    `json:"total_feedback"`
	PositiveCount   int                    `json:"positive_count"`
	NegativeCount   int                    `json:"negative_count"`
	CommonIssues    []string               `json:"common_issues"`
	Suggestions     []string               `json:"suggestions"`
	SentimentScore  float64                `json:"sentiment_score"` // -1.0 ?1.0
}

// LearningContent 
type LearningContent struct {
	ID                uuid.UUID             `json:"id"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	Type              ContentType           `json:"type"`
	Status            ContentStatus         `json:"status"`
	Difficulty        DifficultyLevel       `json:"difficulty"`
	EstimatedDuration int                   `json:"estimated_duration"` // 
	Prerequisites     []uuid.UUID           `json:"prerequisites"`      // ID
	LearningObjectives []LearningObjective  `json:"learning_objectives"`
	Content           string                `json:"content"`            // HTML/Markdown?
	MediaResources    []MediaResource       `json:"media_resources"`
	InteractiveElements []InteractiveElement `json:"interactive_elements"`
	QuizQuestions     []QuizQuestion        `json:"quiz_questions"`
	Tags              []string              `json:"tags"`
	KnowledgeNodeIDs  []uuid.UUID           `json:"knowledge_node_ids"` // ID
	Metadata          ContentMetadata       `json:"metadata"`
	Analytics         ContentAnalytics      `json:"analytics"`
	AuthorID          uuid.UUID             `json:"author_id"`          // ID
	CreatedBy         uuid.UUID             `json:"created_by"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	PublishedAt       *time.Time            `json:"published_at"`
}

// ContentProgress 
type ContentProgress struct {
	ID              uuid.UUID              `json:"id"`
	LearnerID       uuid.UUID              `json:"learner_id"`
	ContentID       uuid.UUID              `json:"content_id"`
	Progress        float64                `json:"progress"`        // 0.0-1.0
	TimeSpent       int                    `json:"time_spent"`      // 
	LastPosition    int                    `json:"last_position"`   // 
	CompletedSections []string             `json:"completed_sections"` // 
	QuizScores      map[uuid.UUID]float64  `json:"quiz_scores"`     // 
	InteractionLog  []InteractionRecord    `json:"interaction_log"` // 
	Notes           []LearningNote         `json:"notes"`           // 
	Bookmarks       []Bookmark             `json:"bookmarks"`       // 
	IsCompleted     bool                   `json:"is_completed"`
	CompletedAt     *time.Time             `json:"completed_at"`
	StartedAt       time.Time              `json:"started_at"`
	LastAccessedAt  time.Time              `json:"last_accessed_at"`
}

// InteractionRecord 
type InteractionRecord struct {
	ID        uuid.UUID              `json:"id"`
	LearnerID uuid.UUID              `json:"learner_id"` // ID
	ContentID uuid.UUID              `json:"content_id"` // ID
	Type      string                 `json:"type"`       // click, scroll, pause, replay?
	Element   string                 `json:"element"`    // 
	Position  int                    `json:"position"`   // ?
	Data      map[string]interface{} `json:"data"`       // 
	Timestamp time.Time              `json:"timestamp"`
}

// LearningNote 
type LearningNote struct {
	ID        uuid.UUID `json:"id"`
	LearnerID uuid.UUID `json:"learner_id"`
	ContentID uuid.UUID `json:"content_id"`
	Content   string    `json:"content"`
	Position  int       `json:"position"`  // ?
	Tags      []string  `json:"tags"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Bookmark 
type Bookmark struct {
	ID        uuid.UUID `json:"id"`
	LearnerID uuid.UUID `json:"learner_id"`
	ContentID uuid.UUID `json:"content_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`  // ?
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLearningContent 
func NewLearningContent(title, description string, contentType ContentType, difficulty DifficultyLevel, createdBy uuid.UUID) *LearningContent {
	now := time.Now()
	return &LearningContent{
		ID:                  uuid.New(),
		Title:               title,
		Description:         description,
		Type:                contentType,
		Status:              ContentStatusDraft,
		Difficulty:          difficulty,
		EstimatedDuration:   0,
		Prerequisites:       make([]uuid.UUID, 0),
		LearningObjectives:  make([]LearningObjective, 0),
		Content:             "",
		MediaResources:      make([]MediaResource, 0),
		InteractiveElements: make([]InteractiveElement, 0),
		QuizQuestions:       make([]QuizQuestion, 0),
		Tags:                make([]string, 0),
		KnowledgeNodeIDs:    make([]uuid.UUID, 0),
		Metadata: ContentMetadata{
			Version:  "1.0.0",
			Language: "zh-CN",
			Keywords: make([]string, 0),
			LearningStyle: make([]LearningStyle, 0),
		},
		Analytics: ContentAnalytics{
			ViewCount:       0,
			CompletionRate:  0,
			AverageRating:   0,
			AverageTime:     0,
			DropoffPoints:   make([]int, 0),
			InteractionData: make(map[string]interface{}),
			FeedbackSummary: FeedbackSummary{
				CommonIssues: make([]string, 0),
				Suggestions:  make([]string, 0),
			},
			LastUpdated: now,
		},
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddLearningObjective 
func (lc *LearningContent) AddLearningObjective(description, level string, measurable bool, criteria []string) {
	objective := LearningObjective{
		ID:          uuid.New(),
		Description: description,
		Level:       level,
		Measurable:  measurable,
		Criteria:    criteria,
	}
	lc.LearningObjectives = append(lc.LearningObjectives, objective)
	lc.UpdatedAt = time.Now()
}

// AddMediaResource 
func (lc *LearningContent) AddMediaResource(resourceType, url, title string, duration int, size int64, format string) {
	resource := MediaResource{
		ID:       uuid.New(),
		Type:     resourceType,
		URL:      url,
		Title:    title,
		Duration: duration,
		Size:     size,
		Format:   format,
		Metadata: make(map[string]interface{}),
	}
	lc.MediaResources = append(lc.MediaResources, resource)
	lc.UpdatedAt = time.Now()
}

// AddQuizQuestion 
func (lc *LearningContent) AddQuizQuestion(questionType, question string, options []string, correctAnswer interface{}, explanation string, points, difficulty int) {
	quiz := QuizQuestion{
		ID:            uuid.New(),
		Type:          questionType,
		Question:      question,
		Options:       options,
		CorrectAnswer: correctAnswer,
		Explanation:   explanation,
		Points:        points,
		Difficulty:    difficulty,
		Tags:          make([]string, 0),
		Metadata:      make(map[string]interface{}),
	}
	lc.QuizQuestions = append(lc.QuizQuestions, quiz)
	lc.UpdatedAt = time.Now()
}

// Publish 
func (lc *LearningContent) Publish() error {
	if lc.Status != ContentStatusDraft && lc.Status != ContentStatusReview {
		return ErrInvalidStatusTransition
	}

	lc.Status = ContentStatusPublished
	now := time.Now()
	lc.PublishedAt = &now
	lc.UpdatedAt = now
	return nil
}

// Archive 鵵
func (lc *LearningContent) Archive() {
	lc.Status = ContentStatusArchived
	lc.UpdatedAt = time.Now()
}

// UpdateAnalytics 
func (lc *LearningContent) UpdateAnalytics(viewCount int, completionRate, averageRating float64, averageTime int) {
	lc.Analytics.ViewCount = viewCount
	lc.Analytics.CompletionRate = completionRate
	lc.Analytics.AverageRating = averageRating
	lc.Analytics.AverageTime = averageTime
	lc.Analytics.LastUpdated = time.Now()
	lc.UpdatedAt = time.Now()
}

// GetEstimatedDurationHours 
func (lc *LearningContent) GetEstimatedDurationHours() float64 {
	return float64(lc.EstimatedDuration) / 60.0
}

// IsAccessibleTo 
func (lc *LearningContent) IsAccessibleTo(learner *Learner) bool {
	if lc.Status != ContentStatusPublished {
		return false
	}

	// ?
	for _, prereqID := range lc.Prerequisites {
		hasPrereq := false
		for _, history := range learner.LearningHistory {
			if history.ContentID == prereqID && history.Completed {
				hasPrereq = true
				break
			}
		}
		if !hasPrereq {
			return false
		}
	}

	return true
}

// NewContentProgress 
func NewContentProgress(learnerID, contentID uuid.UUID) *ContentProgress {
	now := time.Now()
	return &ContentProgress{
		ID:                uuid.New(),
		LearnerID:         learnerID,
		ContentID:         contentID,
		Progress:          0,
		TimeSpent:         0,
		LastPosition:      0,
		CompletedSections: make([]string, 0),
		QuizScores:        make(map[uuid.UUID]float64),
		InteractionLog:    make([]InteractionRecord, 0),
		Notes:             make([]LearningNote, 0),
		Bookmarks:         make([]Bookmark, 0),
		IsCompleted:       false,
		StartedAt:         now,
		LastAccessedAt:    now,
	}
}

// UpdateProgress 
func (cp *ContentProgress) UpdateProgress(progress float64, position int, timeSpent int) {
	cp.Progress = progress
	cp.LastPosition = position
	cp.TimeSpent += timeSpent
	cp.LastAccessedAt = time.Now()

	// ?
	if progress >= 1.0 && !cp.IsCompleted {
		cp.IsCompleted = true
		now := time.Now()
		cp.CompletedAt = &now
	}
}

// AddNote 
func (cp *ContentProgress) AddNote(content string, position int, tags []string, isPublic bool) {
	note := LearningNote{
		ID:        uuid.New(),
		Content:   content,
		Position:  position,
		Tags:      tags,
		IsPublic:  isPublic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	cp.Notes = append(cp.Notes, note)
}

// AddBookmark 
func (cp *ContentProgress) AddBookmark(title string, position int, note string) {
	bookmark := Bookmark{
		ID:        uuid.New(),
		Title:     title,
		Position:  position,
		Note:      note,
		CreatedAt: time.Now(),
	}
	cp.Bookmarks = append(cp.Bookmarks, bookmark)
}

// RecordInteraction 
func (cp *ContentProgress) RecordInteraction(interactionType, element string, position int, data map[string]interface{}) {
	interaction := InteractionRecord{
		ID:        uuid.New(),
		Type:      interactionType,
		Element:   element,
		Position:  position,
		Data:      data,
		Timestamp: time.Now(),
	}
	cp.InteractionLog = append(cp.InteractionLog, interaction)
}

var (
	ErrInvalidStatusTransition = fmt.Errorf("invalid status transition")
)

