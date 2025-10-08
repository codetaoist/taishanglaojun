package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Community 社区实体
type Community struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // public, private, course, study_group
	OwnerID     uuid.UUID `json:"owner_id"`
	Settings    CommunitySettings `json:"settings"`
	MemberCount int       `json:"member_count"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CommunitySettings 社区设置
type CommunitySettings struct {
	IsPublic           bool     `json:"is_public"`
	RequireApproval    bool     `json:"require_approval"`
	AllowedContentTypes []string `json:"allowed_content_types"`
	MaxMembers         int      `json:"max_members"`
	Tags               []string `json:"tags"`
}

// CommunityMember 社区成员
type CommunityMember struct {
	ID          uuid.UUID `json:"id"`
	CommunityID uuid.UUID `json:"community_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"` // owner, moderator, member
	JoinedAt    time.Time `json:"joined_at"`
	IsActive    bool      `json:"is_active"`
}

// Post 帖子
type Post struct {
	ID          uuid.UUID `json:"id"`
	CommunityID uuid.UUID `json:"community_id"`
	AuthorID    uuid.UUID `json:"author_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        string    `json:"type"` // discussion, question, resource, announcement
	Tags        []string  `json:"tags"`
	Attachments []PostAttachment `json:"attachments"`
	LikeCount   int       `json:"like_count"`
	ReplyCount  int       `json:"reply_count"`
	ViewCount   int       `json:"view_count"`
	IsPinned    bool      `json:"is_pinned"`
	IsLocked    bool      `json:"is_locked"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PostAttachment 帖子附件
type PostAttachment struct {
	ID       uuid.UUID `json:"id"`
	PostID   uuid.UUID `json:"post_id"`
	Type     string    `json:"type"` // image, file, link, video
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mime_type"`
}

// Reply 回复
type Reply struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StudyGroup 学习小组
type StudyGroup struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CommunityID uuid.UUID `json:"community_id"`
	LeaderID    uuid.UUID `json:"leader_id"`
	MaxMembers  int       `json:"max_members"`
	Schedule    GroupSchedule `json:"schedule"`
	Status      string    `json:"status"` // active, inactive, completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupSchedule 小组日程
type GroupSchedule struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Frequency   string    `json:"frequency"` // daily, weekly, monthly
	DaysOfWeek  []int     `json:"days_of_week"`
	TimeZone    string    `json:"time_zone"`
	Description string    `json:"description"`
}

// CommunityRepository 社区数据访问接口
type CommunityRepository interface {
	// 社区管理
	CreateCommunity(ctx context.Context, community *Community) error
	GetCommunityByID(ctx context.Context, id uuid.UUID) (*Community, error)
	UpdateCommunity(ctx context.Context, community *Community) error
	DeleteCommunity(ctx context.Context, id uuid.UUID) error
	ListCommunities(ctx context.Context, offset, limit int) ([]*Community, error)
	GetCommunitiesByType(ctx context.Context, communityType string) ([]*Community, error)
	SearchCommunities(ctx context.Context, query string, offset, limit int) ([]*Community, error)

	// 成员管理
	AddMember(ctx context.Context, member *CommunityMember) error
	RemoveMember(ctx context.Context, communityID, userID uuid.UUID) error
	GetCommunityMembers(ctx context.Context, communityID uuid.UUID) ([]*CommunityMember, error)
	GetMemberRole(ctx context.Context, communityID, userID uuid.UUID) (string, error)
	UpdateMemberRole(ctx context.Context, communityID, userID uuid.UUID, role string) error
	IsMember(ctx context.Context, communityID, userID uuid.UUID) (bool, error)

	// 帖子管理
	CreatePost(ctx context.Context, post *Post) error
	GetPostByID(ctx context.Context, id uuid.UUID) (*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id uuid.UUID) error
	GetCommunityPosts(ctx context.Context, communityID uuid.UUID, offset, limit int) ([]*Post, error)
	GetPostsByAuthor(ctx context.Context, authorID uuid.UUID) ([]*Post, error)
	GetPostsByType(ctx context.Context, communityID uuid.UUID, postType string) ([]*Post, error)

	// 回复管理
	CreateReply(ctx context.Context, reply *Reply) error
	GetRepliesByPost(ctx context.Context, postID uuid.UUID) ([]*Reply, error)
	UpdateReply(ctx context.Context, reply *Reply) error
	DeleteReply(ctx context.Context, id uuid.UUID) error

	// 学习小组管理
	CreateStudyGroup(ctx context.Context, group *StudyGroup) error
	GetStudyGroupByID(ctx context.Context, id uuid.UUID) (*StudyGroup, error)
	UpdateStudyGroup(ctx context.Context, group *StudyGroup) error
	DeleteStudyGroup(ctx context.Context, id uuid.UUID) error
	GetCommunityStudyGroups(ctx context.Context, communityID uuid.UUID) ([]*StudyGroup, error)
	GetStudyGroupsByLeader(ctx context.Context, leaderID uuid.UUID) ([]*StudyGroup, error)
}