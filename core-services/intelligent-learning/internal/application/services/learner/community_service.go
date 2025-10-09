package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningCommunityService 学习社区应用服务
type LearningCommunityService struct {
	communityRepo    repositories.CommunityRepository
	learnerRepo      repositories.LearnerRepository
	contentRepo      repositories.LearningContentRepository
	notificationService NotificationService
}

// NewLearningCommunityService 创建新的学习社区应用服务
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

// CommunityType 社区类型
type CommunityType string

const (
	CommunityTypePublic    CommunityType = "public"     // 公开社区
	CommunityTypePrivate   CommunityType = "private"    // 私有社区
	CommunityTypeCourse    CommunityType = "course"     // 课程社区
	CommunityTypeStudyGroup CommunityType = "study_group" // 学习小组
)

// PostType 帖子类型
type PostType string

const (
	PostTypeDiscussion PostType = "discussion" // 讨论
	PostTypeQuestion   PostType = "question"   // 问题
	PostTypeResource   PostType = "resource"   // 资源分享
	PostTypeAnnouncement PostType = "announcement" // 公告
)

// Community 学习社区
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

// CommunitySettings 社区设置
type CommunitySettings struct {
	AllowMemberPost     bool     `json:"allow_member_post"`
	RequireApproval     bool     `json:"require_approval"`
	AllowedPostTypes    []PostType `json:"allowed_post_types"`
	MaxMembersCount     int      `json:"max_members_count"`
	AutoJoin            bool     `json:"auto_join"`
	NotificationEnabled bool     `json:"notification_enabled"`
}

// CommunityMember 社区成员
type CommunityMember struct {
	ID          uuid.UUID `json:"id"`
	CommunityID uuid.UUID `json:"community_id"`
	LearnerID   uuid.UUID `json:"learner_id"`
	Role        MemberRole `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	IsActive    bool      `json:"is_active"`
}

// MemberRole 成员角色
type MemberRole string

const (
	MemberRoleOwner     MemberRole = "owner"     // 所有者
	MemberRoleModerator MemberRole = "moderator" // 管理员
	MemberRoleMember    MemberRole = "member"    // 普通成员
)

// Post 帖子
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

// PostAttachment 帖子附件
type PostAttachment struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"` // image, file, link
	URL      string    `json:"url"`
	Name     string    `json:"name"`
	Size     int64     `json:"size,omitempty"`
	MimeType string    `json:"mime_type,omitempty"`
}

// Reply 回复
type Reply struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"` // 用于嵌套回复
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StudyGroup 学习小组
type StudyGroup struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatorID   uuid.UUID   `json:"creator_id"`
	ContentID   *uuid.UUID  `json:"content_id,omitempty"` // 关联的学习内容
	MaxMembers  int         `json:"max_members"`
	CurrentMembers int      `json:"current_members"`
	Schedule    GroupSchedule `json:"schedule"`
	Status      GroupStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// GroupSchedule 小组时间安排
type GroupSchedule struct {
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	MeetingDays []string  `json:"meeting_days"` // ["monday", "wednesday", "friday"]
	MeetingTime string    `json:"meeting_time"` // "19:00"
	Duration    int       `json:"duration"`     // 分钟
	Timezone    string    `json:"timezone"`
}

// GroupStatus 小组状态
type GroupStatus string

const (
	GroupStatusRecruiting GroupStatus = "recruiting" // 招募中
	GroupStatusActive     GroupStatus = "active"     // 活跃
	GroupStatusCompleted  GroupStatus = "completed"  // 已完成
	GroupStatusCancelled  GroupStatus = "cancelled"  // 已取消
)

// 请求和响应结构体

// CreateCommunityRequest 创建社区请求
type CreateCommunityRequest struct {
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	Type        CommunityType `json:"type" binding:"required"`
	Tags        []string      `json:"tags"`
	Settings    CommunitySettings `json:"settings"`
}

// CreateCommunityResponse 创建社区响应
type CreateCommunityResponse struct {
	Community Community `json:"community"`
	Message   string    `json:"message"`
}

// JoinCommunityRequest 加入社区请求
type JoinCommunityRequest struct {
	CommunityID uuid.UUID `json:"community_id" binding:"required"`
	LearnerID   uuid.UUID `json:"learner_id" binding:"required"`
	Message     string    `json:"message,omitempty"` // 申请消息
}

// CreatePostRequest 创建帖子请求
type CreatePostRequest struct {
	CommunityID uuid.UUID        `json:"community_id" binding:"required"`
	AuthorID    uuid.UUID        `json:"author_id" binding:"required"`
	Title       string           `json:"title" binding:"required"`
	Content     string           `json:"content" binding:"required"`
	Type        PostType         `json:"type" binding:"required"`
	Tags        []string         `json:"tags"`
	Attachments []PostAttachment `json:"attachments,omitempty"`
}

// CreateReplyRequest 创建回复请求
type CreateReplyRequest struct {
	PostID   uuid.UUID  `json:"post_id" binding:"required"`
	AuthorID uuid.UUID  `json:"author_id" binding:"required"`
	Content  string     `json:"content" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// GetCommunityPostsRequest 获取社区帖子请求
type GetCommunityPostsRequest struct {
	CommunityID uuid.UUID `json:"community_id" binding:"required"`
	Type        *PostType `json:"type,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	SortBy      string    `json:"sort_by,omitempty"` // latest, popular, pinned
	Page        int       `json:"page,omitempty"`
	Limit       int       `json:"limit,omitempty"`
}

// GetCommunityPostsResponse 获取社区帖子响应
type GetCommunityPostsResponse struct {
	Posts []Post `json:"posts"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// CreateStudyGroupRequest 创建学习小组请求
type CreateStudyGroupRequest struct {
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	CreatorID   uuid.UUID     `json:"creator_id" binding:"required"`
	ContentID   *uuid.UUID    `json:"content_id,omitempty"`
	MaxMembers  int           `json:"max_members" binding:"required"`
	Schedule    GroupSchedule `json:"schedule"`
}

// 服务方法实现

// CreateCommunity 创建学习社区
func (s *LearningCommunityService) CreateCommunity(ctx context.Context, req *CreateCommunityRequest, creatorID uuid.UUID) (*CreateCommunityResponse, error) {
	// 验证创建者
	_, err := s.learnerRepo.GetByID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// 创建社区
	newCommunity := &Community{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		CreatorID:   creatorID,
		Tags:        req.Tags,
		MemberCount: 1, // 创建者自动成为成员
		PostCount:   0,
		IsActive:    true,
		Settings:    req.Settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 设置默认设置
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

	// 创建者自动加入社区
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
		Message:   "社区创建成功",
	}, nil
}

// JoinCommunity 加入社区
func (s *LearningCommunityService) JoinCommunity(ctx context.Context, req *JoinCommunityRequest) error {
	// 检查社区是否存在
	repoCommunity, err := s.communityRepo.GetCommunityByID(ctx, req.CommunityID)
	if err != nil {
		return fmt.Errorf("failed to get community: %w", err)
	}

	// 转换为应用层类型
	community := s.convertFromRepositoryCommunity(repoCommunity)

	// 检查是否已经是成员
	isMember, err := s.communityRepo.IsMember(ctx, req.CommunityID, req.LearnerID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	if isMember {
		return fmt.Errorf("already a member of this community")
	}

	// 检查成员数量限制
	if community.Settings.MaxMembersCount > 0 && community.MemberCount >= community.Settings.MaxMembersCount {
		return fmt.Errorf("community has reached maximum member limit")
	}

	// 创建成员记录
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

	// 发送欢迎通知
	go s.sendWelcomeNotification(ctx, req.LearnerID, community)

	return nil
}

// CreatePost 创建帖子
func (s *LearningCommunityService) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
	// 验证用户是否为社区成员
	isMember, err := s.communityRepo.IsMember(ctx, req.CommunityID, req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this community")
	}

	// 检查社区设置
	repoCommunity, err := s.communityRepo.GetCommunityByID(ctx, req.CommunityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get community: %w", err)
	}

	// 转换为应用层类型
	community := s.convertFromRepositoryCommunity(repoCommunity)

	// 检查是否允许该类型的帖子
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

	// 创建帖子
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

	// 发送通知给社区成员
	go s.sendPostNotification(ctx, post, community)

	return post, nil
}

// CreateReply 创建回复
func (s *LearningCommunityService) CreateReply(ctx context.Context, req *CreateReplyRequest) (*Reply, error) {
	// 获取帖子信息
	repoPost, err := s.communityRepo.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// 转换为应用层类型
	post := s.convertFromRepositoryPost(repoPost)

	// 验证用户是否为社区成员
	isMember, err := s.communityRepo.IsMember(ctx, post.CommunityID, req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this community")
	}

	// 检查帖子是否被锁定
	if post.IsLocked {
		return nil, fmt.Errorf("post is locked and cannot accept new replies")
	}

	// 创建回复
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

	// TODO: 实现更新帖子回复数量的逻辑
	// 可以通过获取帖子，更新回复数量，然后保存帖子来实现

	// 发送通知给帖子作者
	go s.sendReplyNotification(ctx, reply, post)

	return reply, nil
}

// GetCommunityPosts 获取社区帖子列表
func (s *LearningCommunityService) GetCommunityPosts(ctx context.Context, req *GetCommunityPostsRequest) (*GetCommunityPostsResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "latest"
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.Limit

	// 获取社区帖子
	repoPosts, err := s.communityRepo.GetCommunityPosts(ctx, req.CommunityID, offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get community posts: %w", err)
	}

	// 转换为应用层类型
	var posts []Post
	for _, repoPost := range repoPosts {
		posts = append(posts, *s.convertFromRepositoryPost(repoPost))
	}

	// TODO: 实现总数统计，目前返回当前页的数量
	total := len(posts)

	return &GetCommunityPostsResponse{
		Posts: posts,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// CreateStudyGroup 创建学习小组
func (s *LearningCommunityService) CreateStudyGroup(ctx context.Context, req *CreateStudyGroupRequest) (*StudyGroup, error) {
	// 验证创建者
	_, err := s.learnerRepo.GetByID(ctx, req.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// 如果指定了学习内容，验证内容是否存在
	if req.ContentID != nil {
		_, err := s.contentRepo.GetByID(ctx, *req.ContentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get learning content: %w", err)
		}
	}

	// 创建学习小组
	studyGroup := &StudyGroup{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		CreatorID:      req.CreatorID,
		ContentID:      req.ContentID,
		MaxMembers:     req.MaxMembers,
		CurrentMembers: 1, // 创建者自动加入
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

// 辅助方法

func (s *LearningCommunityService) sendWelcomeNotification(ctx context.Context, learnerID uuid.UUID, community *Community) {
	if s.notificationService != nil {
		notification := map[string]interface{}{
			"type":         "community_welcome",
			"learner_id":   learnerID,
			"community_id": community.ID,
			"title":        fmt.Sprintf("欢迎加入 %s", community.Name),
			"message":      fmt.Sprintf("欢迎您加入 %s 学习社区！", community.Name),
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
			"title":        fmt.Sprintf("新帖子：%s", post.Title),
			"message":      fmt.Sprintf("在 %s 社区发布了新帖子", community.Name),
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
			"target_id": post.AuthorID, // 通知帖子作者
			"title":     "您的帖子有新回复",
			"message":   fmt.Sprintf("您的帖子《%s》收到了新回复", post.Title),
		}

		if err := s.notificationService.SendNotification(ctx, notification); err != nil {
			fmt.Printf("Failed to send reply notification: %v\n", err)
		}
	}
}

// 类型转换函数
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
		PostCount:   0, // 这个字段在 repositories 层没有对应
		IsActive:    repoCommunity.IsActive,
		Settings: CommunitySettings{
			AllowMemberPost:     repoCommunity.Settings.IsPublic,
			RequireApproval:     repoCommunity.Settings.RequireApproval,
			AllowedPostTypes:    postTypes,
			MaxMembersCount:     repoCommunity.Settings.MaxMembers,
			AutoJoin:            false, // 默认值
			NotificationEnabled: true,  // 默认值
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
	// 转换应用层 GroupSchedule 到仓储层 GroupSchedule
	// 应用层字段: StartDate, EndDate, MeetingDays, MeetingTime, Duration, Timezone
	// 仓储层字段: StartTime, EndTime, Frequency, DaysOfWeek, TimeZone, Description
	
	// 将 MeetingDays 转换为 DaysOfWeek (字符串数组转为整数数组)
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
		Frequency:   "weekly", // 默认为每周
		DaysOfWeek:  daysOfWeek,
		TimeZone:    studyGroup.Schedule.Timezone,
		Description: fmt.Sprintf("Meeting time: %s, Duration: %d minutes", studyGroup.Schedule.MeetingTime, studyGroup.Schedule.Duration),
	}

	return &repositories.StudyGroup{
		ID:          studyGroup.ID,
		Name:        studyGroup.Name,
		Description: studyGroup.Description,
		CommunityID: uuid.New(), // TODO: 需要从上下文获取正确的 CommunityID
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
	// 转换附件
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