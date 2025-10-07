package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ContentType 内容类型
type ContentType string

const (
	ContentTypeVideo       ContentType = "video"       // 视频
	ContentTypeText        ContentType = "text"        // 文本
	ContentTypeAudio       ContentType = "audio"       // 音频
	ContentTypeInteractive ContentType = "interactive" // 交互式内容
	ContentTypeQuiz        ContentType = "quiz"        // 测验
	ContentTypeExercise    ContentType = "exercise"    // 练习
	ContentTypeSimulation  ContentType = "simulation"  // 模拟
	ContentTypeGame        ContentType = "game"        // 游戏
)

// ContentStatus 内容状态
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"     // 草稿
	ContentStatusReview    ContentStatus = "review"    // 审核中
	ContentStatusPublished ContentStatus = "published" // 已发布
	ContentStatusArchived  ContentStatus = "archived"  // 已归档
)

// MediaResource 媒体资源
type MediaResource struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"`     // image, video, audio, document
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Duration int       `json:"duration"` // 秒，适用于视频和音频
	Size     int64     `json:"size"`     // 字节
	Format   string    `json:"format"`   // mp4, pdf, jpg等
	Metadata map[string]interface{} `json:"metadata"`
}

// InteractiveElement 交互元素
type InteractiveElement struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`        // button, input, dropdown, drag_drop等
	Position    Position  `json:"position"`    // 位置信息
	Properties  map[string]interface{} `json:"properties"` // 属性
	Actions     []Action  `json:"actions"`     // 交互动作
	Feedback    string    `json:"feedback"`    // 反馈信息
	Points      int       `json:"points"`      // 得分
}

// Position 位置信息
type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Action 交互动作
type Action struct {
	Type       string                 `json:"type"`       // click, hover, input等
	Target     string                 `json:"target"`     // 目标元素
	Parameters map[string]interface{} `json:"parameters"` // 参数
	Response   string                 `json:"response"`   // 响应
}

// QuizQuestion 测验问题
type QuizQuestion struct {
	ID          uuid.UUID    `json:"id"`
	Type        string       `json:"type"`        // multiple_choice, true_false, fill_blank, essay
	Question    string       `json:"question"`
	Options     []string     `json:"options"`     // 选项（适用于选择题）
	CorrectAnswer interface{} `json:"correct_answer"` // 正确答案
	Explanation string       `json:"explanation"` // 解释
	Points      int          `json:"points"`      // 分值
	Difficulty  int          `json:"difficulty"`  // 难度 1-5
	Tags        []string     `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LearningObjective 学习目标
type LearningObjective struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Level       string    `json:"level"`       // remember, understand, apply, analyze, evaluate, create
	Measurable  bool      `json:"measurable"`  // 是否可测量
	Criteria    []string  `json:"criteria"`    // 评估标准
}

// ContentMetadata 内容元数据
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

// AccessibilityInfo 无障碍信息
type AccessibilityInfo struct {
	HasCaptions     bool `json:"has_captions"`     // 有字幕
	HasTranscript   bool `json:"has_transcript"`   // 有文字稿
	HasAudioDesc    bool `json:"has_audio_desc"`   // 有音频描述
	HighContrast    bool `json:"high_contrast"`    // 高对比度
	LargeText       bool `json:"large_text"`       // 大字体
	KeyboardNav     bool `json:"keyboard_nav"`     // 键盘导航
	ScreenReader    bool `json:"screen_reader"`    // 屏幕阅读器兼容
}

// ContentAnalytics 内容分析数据
type ContentAnalytics struct {
	ViewCount       int                    `json:"view_count"`
	CompletionRate  float64                `json:"completion_rate"`
	AverageRating   float64                `json:"average_rating"`
	AverageTime     int                    `json:"average_time"`     // 平均学习时间（秒）
	DropoffPoints   []int                  `json:"dropoff_points"`   // 流失点（百分比）
	InteractionData map[string]interface{} `json:"interaction_data"` // 交互数据
	FeedbackSummary FeedbackSummary        `json:"feedback_summary"`
	LastUpdated     time.Time              `json:"last_updated"`
}

// FeedbackSummary 反馈摘要
type FeedbackSummary struct {
	TotalFeedback   int                    `json:"total_feedback"`
	PositiveCount   int                    `json:"positive_count"`
	NegativeCount   int                    `json:"negative_count"`
	CommonIssues    []string               `json:"common_issues"`
	Suggestions     []string               `json:"suggestions"`
	SentimentScore  float64                `json:"sentiment_score"` // -1.0 到 1.0
}

// LearningContent 学习内容实体
type LearningContent struct {
	ID                uuid.UUID             `json:"id"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	Type              ContentType           `json:"type"`
	Status            ContentStatus         `json:"status"`
	Difficulty        DifficultyLevel       `json:"difficulty"`
	EstimatedDuration int                   `json:"estimated_duration"` // 预估学习时间（分钟）
	Prerequisites     []uuid.UUID           `json:"prerequisites"`      // 前置内容ID
	LearningObjectives []LearningObjective  `json:"learning_objectives"`
	Content           string                `json:"content"`            // 主要内容（HTML/Markdown）
	MediaResources    []MediaResource       `json:"media_resources"`
	InteractiveElements []InteractiveElement `json:"interactive_elements"`
	QuizQuestions     []QuizQuestion        `json:"quiz_questions"`
	Tags              []string              `json:"tags"`
	KnowledgeNodeIDs  []uuid.UUID           `json:"knowledge_node_ids"` // 关联的知识点ID
	Metadata          ContentMetadata       `json:"metadata"`
	Analytics         ContentAnalytics      `json:"analytics"`
	AuthorID          uuid.UUID             `json:"author_id"`          // 作者ID
	CreatedBy         uuid.UUID             `json:"created_by"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	PublishedAt       *time.Time            `json:"published_at"`
}

// ContentProgress 内容学习进度
type ContentProgress struct {
	ID              uuid.UUID              `json:"id"`
	LearnerID       uuid.UUID              `json:"learner_id"`
	ContentID       uuid.UUID              `json:"content_id"`
	Progress        float64                `json:"progress"`        // 0.0-1.0
	TimeSpent       int                    `json:"time_spent"`      // 已花费时间（秒）
	LastPosition    int                    `json:"last_position"`   // 最后位置（百分比）
	CompletedSections []string             `json:"completed_sections"` // 已完成的章节
	QuizScores      map[uuid.UUID]float64  `json:"quiz_scores"`     // 测验得分
	InteractionLog  []InteractionRecord    `json:"interaction_log"` // 交互记录
	Notes           []LearningNote         `json:"notes"`           // 学习笔记
	Bookmarks       []Bookmark             `json:"bookmarks"`       // 书签
	IsCompleted     bool                   `json:"is_completed"`
	CompletedAt     *time.Time             `json:"completed_at"`
	StartedAt       time.Time              `json:"started_at"`
	LastAccessedAt  time.Time              `json:"last_accessed_at"`
}

// InteractionRecord 交互记录
type InteractionRecord struct {
	ID        uuid.UUID              `json:"id"`
	LearnerID uuid.UUID              `json:"learner_id"` // 学习者ID
	ContentID uuid.UUID              `json:"content_id"` // 内容ID
	Type      string                 `json:"type"`       // click, scroll, pause, replay等
	Element   string                 `json:"element"`    // 交互元素
	Position  int                    `json:"position"`   // 位置（秒或百分比）
	Data      map[string]interface{} `json:"data"`       // 交互数据
	Timestamp time.Time              `json:"timestamp"`
}

// LearningNote 学习笔记
type LearningNote struct {
	ID        uuid.UUID `json:"id"`
	LearnerID uuid.UUID `json:"learner_id"`
	ContentID uuid.UUID `json:"content_id"`
	Content   string    `json:"content"`
	Position  int       `json:"position"`  // 位置（秒或百分比）
	Tags      []string  `json:"tags"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Bookmark 书签
type Bookmark struct {
	ID        uuid.UUID `json:"id"`
	LearnerID uuid.UUID `json:"learner_id"`
	ContentID uuid.UUID `json:"content_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`  // 位置（秒或百分比）
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLearningContent 创建新的学习内容
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

// AddLearningObjective 添加学习目标
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

// AddMediaResource 添加媒体资源
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

// AddQuizQuestion 添加测验问题
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

// Publish 发布内容
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

// Archive 归档内容
func (lc *LearningContent) Archive() {
	lc.Status = ContentStatusArchived
	lc.UpdatedAt = time.Now()
}

// UpdateAnalytics 更新分析数据
func (lc *LearningContent) UpdateAnalytics(viewCount int, completionRate, averageRating float64, averageTime int) {
	lc.Analytics.ViewCount = viewCount
	lc.Analytics.CompletionRate = completionRate
	lc.Analytics.AverageRating = averageRating
	lc.Analytics.AverageTime = averageTime
	lc.Analytics.LastUpdated = time.Now()
	lc.UpdatedAt = time.Now()
}

// GetEstimatedDurationHours 获取预估学习时间（小时）
func (lc *LearningContent) GetEstimatedDurationHours() float64 {
	return float64(lc.EstimatedDuration) / 60.0
}

// IsAccessibleTo 检查内容是否对学习者可访问
func (lc *LearningContent) IsAccessibleTo(learner *Learner) bool {
	if lc.Status != ContentStatusPublished {
		return false
	}

	// 检查前置条件
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

// NewContentProgress 创建新的内容进度
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

// UpdateProgress 更新学习进度
func (cp *ContentProgress) UpdateProgress(progress float64, position int, timeSpent int) {
	cp.Progress = progress
	cp.LastPosition = position
	cp.TimeSpent += timeSpent
	cp.LastAccessedAt = time.Now()

	// 检查是否完成
	if progress >= 1.0 && !cp.IsCompleted {
		cp.IsCompleted = true
		now := time.Now()
		cp.CompletedAt = &now
	}
}

// AddNote 添加学习笔记
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

// AddBookmark 添加书签
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

// RecordInteraction 记录交互
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