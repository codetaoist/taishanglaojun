package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningCommunityService 
type LearningCommunityService struct {
	communityRepo    repositories.CommunityRepository
	learnerRepo      repositories.LearnerRepository
	contentRepo      repositories.LearningContentRepository
	notificationService NotificationService
}

// NewLearningCommunityService 
func NewLearningCommunityService(
	communityRepo repositories.CommunityRepository,
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	notificationService NotificationService,
) *LearningCommunityService {
	return &LearningCommunityService{
		communityRepo:       communityRepo,
		learnerRepo:         learnerRepo,
		contentRepo:         contentRepo,
		notificationService: notificationService,
	}
}

// CommunityType 
type CommunityType string

const (
	CommunityTypePublic    CommunityType = "public"     // 
	CommunityTypePrivate   CommunityType = "private"    // 
	CommunityTypeCourse    CommunityType = "course"     // 
	CommunityTypeStudyGroup CommunityType = "study_group" // 
)

// PostType 
type PostType string

const (
	PostTypeDiscussion PostType = "discussion" // 
	PostTypeQuestion   PostType = "question"   // 
	PostTypeResource   PostType = "resource"   // 
	PostTypeAnnouncement PostType = "announcement" // 
)

// Community 
type Community struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        CommunityType `json:"type"`
	CreatorID   uuid.UUID     `json:"creator_id"`
	Avatar      string        `json:"avatar,omitempty"`
	Tags        []string      `json:"tags"`
	MemberCount int           `json:"member_count"`
	PostCount   int           `json:"post_count"`
	IsActive    bool          `json:"is_active"`
	Settings    CommunitySettings `json:"settings"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// CommunitySettings 
type CommunitySettings struct {
	AllowMemberPost     bool     `json:"allow_member_post"`
	RequireApproval     bool     `json:"require_approval"`
	AllowedPostTypes    []PostType `json:"allowed_post_types"`
	MaxMembersCount     int      `json:"max_members_count"`
	AutoJoin            bool     `json:"auto_join"`
	NotificationEnabled bool     `json:"notification_enabled"`
}

// CommunityMember 
type CommunityMember struct {
	ID          uuid.UUID `json:"id"`
	CommunityID uuid.UUID `json:"community_id"`
	LearnerID   uuid.UUID `json:"learner_id"`
	Role        MemberRole `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	IsActive    bool      `json:"is_active"`
}

// MemberRole 
type MemberRole string

const (
	MemberRoleOwner     MemberRole = "owner"     // ?
	MemberRoleModerator MemberRole = "moderator" // ?
	MemberRoleMember    MemberRole = "member"    // ?
)

// Post 
type Post struct {
	ID          uuid.UUID `json:"id"`
	CommunityID uuid.UUID `json:"community_id"`
	AuthorID    uuid.UUID `json:"author_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        PostType  `json:"type"`
	Tags        []string  `json:"tags"`
	Attachments []PostAttachment `json:"attachments,omitempty"`
	ViewCount   int       `json:"view_count"`
	LikeCount   int       `json:"like_count"`
	ReplyCount  int       `json:"reply_count"`
	IsPinned    bool      `json:"is_pinned"`
	IsLocked    bool      `json:"is_locked"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PostAttachment 
type PostAttachment struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"` // image, file, link
	URL      string    `json:"url"`
	Name     string    `json:"name"`
	Size     int64     `json:"size,omitempty"`
	MimeType string    `json:"mime_type,omitempty"`
}

// Reply 
type Reply struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"` // 
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StudyGroup 
type StudyGroup struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatorID   uuid.UUID   `json:"creator_id"`
	ContentID   *uuid.UUID  `json:"content_id,omitempty"` // ?
	MaxMembers  int         `json:"max_members"`
	CurrentMembers int      `json:"current_members"`
	Schedule    GroupSchedule `json:"schedule"`
	Status      GroupStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// GroupSchedule 䰲
type GroupSchedule struct {
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	MeetingDays []string  `json:"meeting_days"` // ["monday", "wednesday", "friday"]
	MeetingTime string    `json:"meeting_time"` // "19:00"
	Duration    int       `json:"duration"`     // 
	Timezone    string    `json:"timezone"`
}

// GroupStatus ?
type GroupStatus string

const (
	GroupStatusRecruiting GroupStatus = "recruiting" // ?
	GroupStatusActive     GroupStatus = "active"     // 
	GroupStatusCompleted  GroupStatus = "completed"  // ?
	GroupStatusCancelled  GroupStatus = "cancelled"  // ?
)

// 

// CreateCommunityRequest 
type CreateCommunityRequest struct {
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	Type        CommunityType `json:"type" binding:"required"`
	Tags        []string      `json:"tags"`
	Settings    CommunitySettings `json:"settings"`
}

// CreateCommunityResponse 
type CreateCommunityResponse struct {
	Community Community `json:"community"`
	Message   string    `json:"message"`
}

// JoinCommunityRequest 
type JoinCommunityRequest struct {
	CommunityID uuid.UUID `json:"community_id" binding:"required"`
	LearnerID   uuid.UUID `json:"learner_id" binding:"required"`
	Message     string    `json:"message,omitempty"` // 
}

// CreatePostRequest 
type CreatePostRequest struct {
	CommunityID uuid.UUID        `json:"community_id" binding:"required"`
	AuthorID    uuid.UUID        `json:"author_id" binding:"required"`
	Title       string           `json:"title" binding:"required"`
	Content     string           `json:"content" binding:"required"`
	Type        PostType         `json:"type" binding:"required"`
	Tags        []string         `json:"tags"`
	Attachments []PostAttachment `json:"attachments,omitempty"`
}

// CreateReplyRequest 
type CreateReplyRequest struct {
	PostID   uuid.UUID  `json:"post_id" binding:"required"`
	AuthorID uuid.UUID  `json:"author_id" binding:"required"`
	Content  string     `json:"content" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// GetCommunityPostsRequest 
type GetCommunityPostsRequest struct {
	CommunityID uuid.UUID `json:"community_id" binding:"required"`
	Type        *PostType `json:"type,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	SortBy      string    `json:"sort_by,omitempty"` // latest, popular, pinned
	Page        int       `json:"page,omitempty"`
	Limit       int       `json:"limit,omitempty"`
}

// GetCommunityPostsResponse 
type GetCommunityPostsResponse struct {
	Posts []Post `json:"posts"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// CreateStudyGroupRequest 
type CreateStudyGroupRequest struct {
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	CreatorID   uuid.UUID     `json:"creator_id" binding:"required"`
	ContentID   *uuid.UUID    `json:"content_id,omitempty"`
	MaxMembers  int           `json:"max_members" binding:"required"`
	Schedule    GroupSchedule `json:"schedule"`
}

// 

// CreateCommunity 
func (s *LearningCommunityService) CreateCommunity(ctx context.Context, req *CreateCommunityRequest, creatorID uuid.UUID) (*CreateCommunityResponse, error) {
	// ?
	_, err := s.learnerRepo.GetByID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// 
	newCommunity := &Community{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		CreatorID:   creatorID,
		Tags:        req.Tags,
		MemberCount: 1, // ?
		PostCount:   0,
		IsActive:    true,
		Settings:    req.Settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 
	if len(newCommunity.Settings.AllowedPostTypes) == 0 {
		newCommunity.Settings.AllowedPostTypes = []PostType{
			PostTypeDiscussion, PostTypeQuestion, PostTypeResource,
		}
	}

	repoCommunity := s.convertToRepositoryCommunity(newCommunity)
	err = s.communityRepo.CreateCommunity(ctx, repoCommunity)
	if err != nil {
		return nil, fmt.Errorf("failed to create community: %w", err)
	}

	// ?
	member := &CommunityMember{
		ID:          uuid.New(),
		CommunityID: newCommunity.ID,
		LearnerID:   creatorID,
		Role:        MemberRoleOwner,
		JoinedAt:    time.Now(),
		IsActive:    true,
	}

	repoMember := s.convertToRepositoryCommunityMember(member)
	err = s.communityRepo.AddMember(ctx, repoMember)
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	return &CreateCommunityResponse{
		Community: *newCommunity,
		Message:   "",
	}, nil
}

// JoinCommunity 
func (s *LearningCommunityService) JoinCommunity(ctx context.Context, req *JoinCommunityRequest) error {
	// ?
	repoCommunity, err := s.communityRepo.GetCommunityByID(ctx, req.CommunityID)
	if err != nil {
		return fmt.Errorf("failed to get community: %w", err)
	}

	// 
	community := s.convertFromRepositoryCommunity(repoCommunity)

	// 
	isMember, err := s.communityRepo.IsMember(ctx, req.CommunityID, req.LearnerID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	if isMember {
		return fmt.Errorf("already a member of this community")
	}

	// ?
	if community.Settings.MaxMembersCount > 0 && community.MemberCount >= community.Settings.MaxMembersCount {
		return fmt.Errorf("community has reached maximum member limit")
	}

	// 
	member := &CommunityMember{
		ID:          uuid.New(),
		CommunityID: req.CommunityID,
		LearnerID:   req.LearnerID,
		Role:        MemberRoleMember,
		JoinedAt:    time.Now(),
		IsActive:    true,
	}

	repoMember := s.convertToRepositoryCommunityMember(member)
	if err := s.communityRepo.AddMember(ctx, repoMember); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// 
	go s.sendWelcomeNotification(ctx, req.LearnerID, community)

	return nil
}

// CreatePost 
func (s *LearningCommunityService) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
	// ?
	isMember, err := s.communityRepo.IsMember(ctx, req.CommunityID, req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this community")
	}

	// ?
	repoCommunity, err := s.communityRepo.GetCommunityByID(ctx, req.CommunityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get community: %w", err)
	}

	// 
	community := s.convertFromRepositoryCommunity(repoCommunity)

	// ?
	allowedType := false
	for _, allowedPostType := range community.Settings.AllowedPostTypes {
		if allowedPostType == req.Type {
			allowedType = true
			break
		}
	}

	if !allowedType {
		return nil, fmt.Errorf("post type %s is not allowed in this community", req.Type)
	}

	// 
	post := &Post{
		ID:          uuid.New(),
		CommunityID: req.CommunityID,
		AuthorID:    req.AuthorID,
		Title:       req.Title,
		Content:     req.Content,
		Type:        req.Type,
		Tags:        req.Tags,
		Attachments: req.Attachments,
		ViewCount:   0,
		LikeCount:   0,
		ReplyCount:  0,
		IsPinned:    false,
		IsLocked:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repoPost := s.convertToRepositoryPost(post)
	if err := s.communityRepo.CreatePost(ctx, repoPost); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// ?
	go s.sendPostNotification(ctx, post, community)

	return post, nil
}

// CreateReply 
func (s *LearningCommunityService) CreateReply(ctx context.Context, req *CreateReplyRequest) (*Reply, error) {
	// 
	repoPost, err := s.communityRepo.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// 
	post := s.convertFromRepositoryPost(repoPost)

	// ?
	isMember, err := s.communityRepo.IsMember(ctx, post.CommunityID, req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this community")
	}

	// 
	if post.IsLocked {
		return nil, fmt.Errorf("post is locked and cannot accept new replies")
	}

	// 
	reply := &Reply{
		ID:        uuid.New(),
		PostID:    req.PostID,
		AuthorID:  req.AuthorID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		LikeCount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.communityRepo.CreateReply(ctx, s.convertToRepositoryReply(reply)); err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	// TODO: 
	// ?

	// ?
	go s.sendReplyNotification(ctx, reply, post)

	return reply, nil
}

// GetCommunityPosts 
func (s *LearningCommunityService) GetCommunityPosts(ctx context.Context, req *GetCommunityPostsRequest) (*GetCommunityPostsResponse, error) {
	// 
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "latest"
	}

	// ?
	offset := (req.Page - 1) * req.Limit

	// 
	repoPosts, err := s.communityRepo.GetCommunityPosts(ctx, req.CommunityID, offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get community posts: %w", err)
	}

	// 
	var posts []Post
	for _, repoPost := range repoPosts {
		posts = append(posts, *s.convertFromRepositoryPost(repoPost))
	}

	// TODO: ?
	total := len(posts)

	return &GetCommunityPostsResponse{
		Posts: posts,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// CreateStudyGroup 
func (s *LearningCommunityService) CreateStudyGroup(ctx context.Context, req *CreateStudyGroupRequest) (*StudyGroup, error) {
	// ?
	_, err := s.learnerRepo.GetByID(ctx, req.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// 
	if req.ContentID != nil {
		_, err := s.contentRepo.GetByID(ctx, *req.ContentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get learning content: %w", err)
		}
	}

	// 
	studyGroup := &StudyGroup{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		CreatorID:      req.CreatorID,
		ContentID:      req.ContentID,
		MaxMembers:     req.MaxMembers,
		CurrentMembers: 1, // ?
		Schedule:       req.Schedule,
		Status:         GroupStatusRecruiting,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.communityRepo.CreateStudyGroup(ctx, s.convertToRepositoryStudyGroup(studyGroup)); err != nil {
		return nil, fmt.Errorf("failed to create study group: %w", err)
	}

	return studyGroup, nil
}

// 

func (s *LearningCommunityService) sendWelcomeNotification(ctx context.Context, learnerID uuid.UUID, community *Community) {
	if s.notificationService != nil {
		notification := map[string]interface{}{
			"type":         "community_welcome",
			"learner_id":   learnerID,
			"community_id": community.ID,
			"title":        fmt.Sprintf(" %s", community.Name),
			"message":      fmt.Sprintf("?%s ?, community.Name),
		}

		if err := s.notificationService.SendNotification(ctx, notification); err != nil {
			fmt.Printf("Failed to send welcome notification: %v\n", err)
		}
	}
}

func (s *LearningCommunityService) sendPostNotification(ctx context.Context, post *Post, community *Community) {
	if s.notificationService != nil && community.Settings.NotificationEnabled {
		notification := map[string]interface{}{
			"type":         "new_post",
			"community_id": post.CommunityID,
			"post_id":      post.ID,
			"author_id":    post.AuthorID,
			"title":        fmt.Sprintf("%s", post.Title),
			"message":      fmt.Sprintf("?%s ", community.Name),
		}

		if err := s.notificationService.SendNotification(ctx, notification); err != nil {
			fmt.Printf("Failed to send post notification: %v\n", err)
		}
	}
}

func (s *LearningCommunityService) sendReplyNotification(ctx context.Context, reply *Reply, post *Post) {
	if s.notificationService != nil {
		notification := map[string]interface{}{
			"type":      "new_reply",
			"post_id":   reply.PostID,
			"reply_id":  reply.ID,
			"author_id": reply.AuthorID,
			"target_id": post.AuthorID, // ?
			"title":     "",
			"message":   fmt.Sprintf("?s?, post.Title),
		}

		if err := s.notificationService.SendNotification(ctx, notification); err != nil {
			fmt.Printf("Failed to send reply notification: %v\n", err)
		}
	}
}

// 
func (s *LearningCommunityService) convertFromRepositoryCommunity(repoCommunity *repositories.Community) *Community {
	var postTypes []PostType
	for _, ct := range repoCommunity.Settings.AllowedContentTypes {
		postTypes = append(postTypes, PostType(ct))
	}

	return &Community{
		ID:          repoCommunity.ID,
		Name:        repoCommunity.Name,
		Description: repoCommunity.Description,
		Type:        CommunityType(repoCommunity.Type),
		CreatorID:   repoCommunity.OwnerID,
		Tags:        repoCommunity.Settings.Tags,
		MemberCount: repoCommunity.MemberCount,
		PostCount:   0, // ?repositories ?
		IsActive:    repoCommunity.IsActive,
		Settings: CommunitySettings{
			AllowMemberPost:     repoCommunity.Settings.IsPublic,
			RequireApproval:     repoCommunity.Settings.RequireApproval,
			AllowedPostTypes:    postTypes,
			MaxMembersCount:     repoCommunity.Settings.MaxMembers,
			AutoJoin:            false, // ?
			NotificationEnabled: true,  // ?
		},
		CreatedAt: repoCommunity.CreatedAt,
		UpdatedAt: repoCommunity.UpdatedAt,
	}
}

func (s *LearningCommunityService) convertFromRepositoryPost(repoPost *repositories.Post) *Post {
	var attachments []PostAttachment
	for _, repoAttachment := range repoPost.Attachments {
		attachments = append(attachments, PostAttachment{
			ID:       repoAttachment.ID,
			Name:     repoAttachment.Title,
			Type:     repoAttachment.Type,
			URL:      repoAttachment.URL,
			Size:     repoAttachment.Size,
			MimeType: repoAttachment.MimeType,
		})
	}

	return &Post{
		ID:          repoPost.ID,
		CommunityID: repoPost.CommunityID,
		AuthorID:    repoPost.AuthorID,
		Title:       repoPost.Title,
		Content:     repoPost.Content,
		Type:        PostType(repoPost.Type),
		Tags:        repoPost.Tags,
		Attachments: attachments,
		LikeCount:   repoPost.LikeCount,
		ReplyCount:  repoPost.ReplyCount,
		ViewCount:   repoPost.ViewCount,
		IsPinned:    repoPost.IsPinned,
		IsLocked:    repoPost.IsLocked,
		CreatedAt:   repoPost.CreatedAt,
		UpdatedAt:   repoPost.UpdatedAt,
	}
}

func (s *LearningCommunityService) convertToRepositoryStudyGroup(studyGroup *StudyGroup) *repositories.StudyGroup {
	// ?GroupSchedule  GroupSchedule
	// ? StartDate, EndDate, MeetingDays, MeetingTime, Duration, Timezone
	// ? StartTime, EndTime, Frequency, DaysOfWeek, TimeZone, Description
	
	// ?MeetingDays ?DaysOfWeek (?
	daysOfWeek := make([]int, len(studyGroup.Schedule.MeetingDays))
	dayMap := map[string]int{
		"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3,
		"thursday": 4, "friday": 5, "saturday": 6,
	}
	for i, day := range studyGroup.Schedule.MeetingDays {
		if dayNum, ok := dayMap[strings.ToLower(day)]; ok {
			daysOfWeek[i] = dayNum
		}
	}
	
	repoSchedule := repositories.GroupSchedule{
		StartTime:   studyGroup.Schedule.StartDate,
		EndTime:     studyGroup.Schedule.EndDate,
		Frequency:   "weekly", // ?
		DaysOfWeek:  daysOfWeek,
		TimeZone:    studyGroup.Schedule.Timezone,
		Description: fmt.Sprintf("Meeting time: %s, Duration: %d minutes", studyGroup.Schedule.MeetingTime, studyGroup.Schedule.Duration),
	}

	return &repositories.StudyGroup{
		ID:          studyGroup.ID,
		Name:        studyGroup.Name,
		Description: studyGroup.Description,
		CommunityID: uuid.New(), // TODO:  CommunityID
		LeaderID:    studyGroup.CreatorID,
		MaxMembers:  studyGroup.MaxMembers,
		Schedule:    repoSchedule,
		Status:      string(studyGroup.Status),
		CreatedAt:   studyGroup.CreatedAt,
		UpdatedAt:   studyGroup.UpdatedAt,
	}
}

func (s *LearningCommunityService) convertToRepositoryReply(reply *Reply) *repositories.Reply {
	return &repositories.Reply{
		ID:        reply.ID,
		PostID:    reply.PostID,
		AuthorID:  reply.AuthorID,
		Content:   reply.Content,
		ParentID:  reply.ParentID,
		LikeCount: reply.LikeCount,
		CreatedAt: reply.CreatedAt,
		UpdatedAt: reply.UpdatedAt,
	}
}

func (s *LearningCommunityService) convertToRepositoryCommunity(community *Community) *repositories.Community {
	var postTypes []string
	for _, pt := range community.Settings.AllowedPostTypes {
		postTypes = append(postTypes, string(pt))
	}

	return &repositories.Community{
		ID:          community.ID,
		Name:        community.Name,
		Description: community.Description,
		Type:        string(community.Type),
		OwnerID:     community.CreatorID,
		Settings: repositories.CommunitySettings{
			IsPublic:           community.Settings.AllowMemberPost,
			RequireApproval:    community.Settings.RequireApproval,
			AllowedContentTypes: postTypes,
			MaxMembers:         community.Settings.MaxMembersCount,
			Tags:               community.Tags,
		},
		MemberCount: community.MemberCount,
		IsActive:    community.IsActive,
		CreatedAt:   community.CreatedAt,
		UpdatedAt:   community.UpdatedAt,
	}
}

func (s *LearningCommunityService) convertToRepositoryCommunityMember(member *CommunityMember) *repositories.CommunityMember {
	return &repositories.CommunityMember{
		ID:          member.ID,
		CommunityID: member.CommunityID,
		UserID:      member.LearnerID,
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt,
		IsActive:    member.IsActive,
	}
}

func (s *LearningCommunityService) convertToRepositoryPost(post *Post) *repositories.Post {
	// 
	var repoAttachments []repositories.PostAttachment
	for _, attachment := range post.Attachments {
		repoAttachments = append(repoAttachments, repositories.PostAttachment{
			ID:       attachment.ID,
			PostID:   post.ID,
			Type:     attachment.Type,
			URL:      attachment.URL,
			Title:    attachment.Name,
			Size:     attachment.Size,
			MimeType: attachment.MimeType,
		})
	}

	return &repositories.Post{
		ID:          post.ID,
		CommunityID: post.CommunityID,
		AuthorID:    post.AuthorID,
		Title:       post.Title,
		Content:     post.Content,
		Type:        string(post.Type),
		Tags:        post.Tags,
		Attachments: repoAttachments,
		LikeCount:   post.LikeCount,
		ReplyCount:  post.ReplyCount,
		ViewCount:   post.ViewCount,
		IsPinned:    post.IsPinned,
		IsLocked:    post.IsLocked,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}
}

