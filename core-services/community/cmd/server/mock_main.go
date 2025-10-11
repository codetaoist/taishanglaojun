package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 数据模型
type Post struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    string    `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	CategoryID  string    `json:"category_id"`
	Tags        []string  `json:"tags"`
	ViewCount   int       `json:"view_count"`
	LikeCount   int       `json:"like_count"`
	CommentCount int      `json:"comment_count"`
	IsSticky    bool      `json:"is_sticky"`
	IsHot       bool      `json:"is_hot"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	ParentID  *string   `json:"parent_id"`
	AuthorID  string    `json:"author_id"`
	AuthorName string   `json:"author_name"`
	Content   string    `json:"content"`
	LikeCount int       `json:"like_count"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	PostCount   int       `json:"post_count"`
	FollowerCount int     `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	Level       int       `json:"level"`
	Points      int       `json:"points"`
	Status      string    `json:"status"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChatRoom struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	CreatorID   string    `json:"creator_id"`
	MemberCount int       `json:"member_count"`
	MaxMembers  int       `json:"max_members"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChatMessage struct {
	ID       string    `json:"id"`
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Content  string    `json:"content"`
	Type     string    `json:"type"`
	SentAt   time.Time `json:"sent_at"`
}

// 请求结构�?
type CreatePostRequest struct {
	Title      string   `json:"title" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	CategoryID string   `json:"category_id"`
	Tags       []string `json:"tags"`
}

type CreateCommentRequest struct {
	PostID   string  `json:"post_id" binding:"required"`
	ParentID *string `json:"parent_id"`
	Content  string  `json:"content" binding:"required"`
}

type CreateChatRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsPrivate   bool   `json:"is_private"`
	MaxMembers  int    `json:"max_members"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
	Type    string `json:"type"`
}

// 模拟数据
var (
	mockUsers = []User{
		{
			ID: "user1", Username: "alice", Email: "alice@example.com", Nickname: "爱丽�?,
			Avatar: "https://example.com/avatar1.jpg", Bio: "热爱分享的技术爱好�?,
			PostCount: 15, FollowerCount: 120, FollowingCount: 80, Level: 3, Points: 1500,
			Status: "active", LastActiveAt: time.Now().Add(-time.Hour), CreatedAt: time.Now().AddDate(0, -6, 0),
		},
		{
			ID: "user2", Username: "bob", Email: "bob@example.com", Nickname: "鲍勃",
			Avatar: "https://example.com/avatar2.jpg", Bio: "程序员，喜欢探索新技�?,
			PostCount: 8, FollowerCount: 65, FollowingCount: 45, Level: 2, Points: 800,
			Status: "active", LastActiveAt: time.Now().Add(-30*time.Minute), CreatedAt: time.Now().AddDate(0, -3, 0),
		},
		{
			ID: "user3", Username: "charlie", Email: "charlie@example.com", Nickname: "查理",
			Avatar: "https://example.com/avatar3.jpg", Bio: "设计师，关注用户体验",
			PostCount: 22, FollowerCount: 200, FollowingCount: 150, Level: 4, Points: 2200,
			Status: "active", LastActiveAt: time.Now().Add(-10*time.Minute), CreatedAt: time.Now().AddDate(0, -8, 0),
		},
	}

	mockPosts = []Post{
		{
			ID: "post1", Title: "Go语言微服务架构实�?, Content: "本文分享了在实际项目中使用Go语言构建微服务架构的经验和最佳实�?..",
			AuthorID: "user1", AuthorName: "爱丽�?, CategoryID: "tech", Tags: []string{"Go", "微服�?, "架构"},
			ViewCount: 1250, LikeCount: 89, CommentCount: 23, IsSticky: true, IsHot: true, Status: "published",
			CreatedAt: time.Now().AddDate(0, 0, -2), UpdatedAt: time.Now().AddDate(0, 0, -1),
		},
		{
			ID: "post2", Title: "前端性能优化技�?, Content: "分享一些实用的前端性能优化技巧，包括代码分割、懒加载、缓存策略等...",
			AuthorID: "user3", AuthorName: "查理", CategoryID: "frontend", Tags: []string{"前端", "性能优化", "JavaScript"},
			ViewCount: 890, LikeCount: 67, CommentCount: 15, IsSticky: false, IsHot: true, Status: "published",
			CreatedAt: time.Now().AddDate(0, 0, -1), UpdatedAt: time.Now().AddDate(0, 0, -1),
		},
		{
			ID: "post3", Title: "数据库设计原�?, Content: "良好的数据库设计是系统性能的基础，本文介绍了数据库设计的基本原则...",
			AuthorID: "user2", AuthorName: "鲍勃", CategoryID: "database", Tags: []string{"数据�?, "设计", "SQL"},
			ViewCount: 650, LikeCount: 45, CommentCount: 12, IsSticky: false, IsHot: false, Status: "published",
			CreatedAt: time.Now().AddDate(0, 0, -3), UpdatedAt: time.Now().AddDate(0, 0, -3),
		},
	}

	mockComments = []Comment{
		{
			ID: "comment1", PostID: "post1", ParentID: nil, AuthorID: "user2", AuthorName: "鲍勃",
			Content: "很好的文章，学到了很多！", LikeCount: 12, Status: "approved",
			CreatedAt: time.Now().AddDate(0, 0, -1), UpdatedAt: time.Now().AddDate(0, 0, -1),
		},
		{
			ID: "comment2", PostID: "post1", ParentID: nil, AuthorID: "user3", AuthorName: "查理",
			Content: "微服务架构确实是个好方向，但也要注意复杂性管�?, LikeCount: 8, Status: "approved",
			CreatedAt: time.Now().AddDate(0, 0, -1), UpdatedAt: time.Now().AddDate(0, 0, -1),
		},
		{
			ID: "comment3", PostID: "post2", ParentID: nil, AuthorID: "user1", AuthorName: "爱丽�?,
			Content: "性能优化是个永恒的话题，感谢分享�?, LikeCount: 5, Status: "approved",
			CreatedAt: time.Now().Add(-12*time.Hour), UpdatedAt: time.Now().Add(-12*time.Hour),
		},
	}

	mockChatRooms = []ChatRoom{
		{
			ID: "room1", Name: "技术讨�?, Description: "讨论各种技术话�?, Type: "public",
			CreatorID: "user1", MemberCount: 25, MaxMembers: 100, IsPrivate: false,
			CreatedAt: time.Now().AddDate(0, 0, -10),
		},
		{
			ID: "room2", Name: "前端交流", Description: "前端开发经验分�?, Type: "public",
			CreatorID: "user3", MemberCount: 18, MaxMembers: 50, IsPrivate: false,
			CreatedAt: time.Now().AddDate(0, 0, -5),
		},
		{
			ID: "room3", Name: "项目协作", Description: "项目团队内部讨论", Type: "private",
			CreatorID: "user2", MemberCount: 8, MaxMembers: 20, IsPrivate: true,
			CreatedAt: time.Now().AddDate(0, 0, -3),
		},
	}

	mockChatMessages = []ChatMessage{
		{
			ID: "msg1", RoomID: "room1", UserID: "user1", Username: "爱丽�?,
			Content: "大家好，有人在用Go开发微服务吗？", Type: "text",
			SentAt: time.Now().Add(-2*time.Hour),
		},
		{
			ID: "msg2", RoomID: "room1", UserID: "user2", Username: "鲍勃",
			Content: "我在用，感觉还不�?, Type: "text",
			SentAt: time.Now().Add(-time.Hour),
		},
		{
			ID: "msg3", RoomID: "room2", UserID: "user3", Username: "查理",
			Content: "最近在研究React 18的新特�?, Type: "text",
			SentAt: time.Now().Add(-30*time.Minute),
		},
	}
)

func main() {
	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// API路由�?
	api := r.Group("/api/v1")

	// 帖子相关路由
	posts := api.Group("/posts")
	{
		posts.GET("", getPosts)
		posts.POST("", createPost)
		posts.GET("/stats", getPostStats)
		posts.GET("/search", searchPosts)
		posts.GET("/:id", getPost)
		posts.PUT("/:id", updatePost)
		posts.DELETE("/:id", deletePost)
		posts.PATCH("/:id/sticky", setPostSticky)
		posts.PATCH("/:id/hot", setPostHot)
	}

	// 帖子交互路由 - 使用不同的路�?
	api.POST("/post/:post_id/like", likePost)
	api.DELETE("/post/:post_id/like", unlikePost)
	api.POST("/post/:post_id/bookmark", bookmarkPost)
	api.DELETE("/post/:post_id/bookmark", unbookmarkPost)

	// 评论相关路由
	comments := api.Group("/comments")
	{
		comments.POST("", createComment)
		comments.GET("/post/:post_id", getComments)
		comments.GET("/:id", getComment)
		comments.PUT("/:id", updateComment)
		comments.DELETE("/:id", deleteComment)
		comments.GET("/stats", getCommentStats)
		comments.GET("/user/:user_id", getUserComments)
	}

	// 评论交互路由
	api.POST("/comment/:comment_id/like", likeComment)
	api.DELETE("/comment/:comment_id/like", unlikeComment)

	// 用户相关路由
	users := api.Group("/users")
	{
		users.GET("/profile", getMyProfile)
		users.PUT("/profile", updateUserProfile)
		users.GET("/:id", getUserProfile)
		users.GET("", getUsers)
		users.GET("/stats", getUserStats)
		users.GET("/search", searchUsers)
		users.POST("/:id/ban", banUser)
		users.DELETE("/:id/ban", unbanUser)
		users.GET("/:id/posts", getUserPosts)
		users.PUT("/:id/activity", updateUserActivity)
		users.POST("/:id/follow", followUser)
		users.DELETE("/:id/follow", unfollowUser)
		users.GET("/:id/followers", getUserFollowers)
		users.GET("/:id/following", getUserFollowing)
	}

	// 聊天相关路由
	chat := api.Group("/chat")
	{
		chat.POST("/rooms", createChatRoom)
		chat.GET("/rooms", getChatRooms)
		chat.POST("/rooms/:room_id/join", joinChatRoom)
		chat.POST("/rooms/:room_id/leave", leaveChatRoom)
		chat.GET("/rooms/:room_id/messages", getChatMessages)
		chat.POST("/rooms/:room_id/messages", sendMessage)
		chat.GET("/online-users", getOnlineUsers)
		chat.GET("/stats", getChatStats)
	}

	// 健康检�?
	r.GET("/health", healthCheck)

	log.Println("社区服务 (Mock版本) 启动在端�?8085")
	r.Run(":8085")
}

// 帖子相关处理函数
func getPosts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"posts": mockPosts,
			"total": len(mockPosts),
			"page": 1,
			"page_size": 10,
		},
	})
}

func createPost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	post := Post{
		ID: uuid.New().String(),
		Title: req.Title,
		Content: req.Content,
		AuthorID: "user1",
		AuthorName: "当前用户",
		CategoryID: req.CategoryID,
		Tags: req.Tags,
		ViewCount: 0,
		LikeCount: 0,
		CommentCount: 0,
		IsSticky: false,
		IsHot: false,
		Status: "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockPosts = append(mockPosts, post)

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "帖子创建成功",
		"data": post,
	})
}

func getPost(c *gin.Context) {
	id := c.Param("id")
	for _, post := range mockPosts {
		if post.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "success",
				"data": post,
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func updatePost(c *gin.Context) {
	id := c.Param("id")
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	for i, post := range mockPosts {
		if post.ID == id {
			mockPosts[i].Title = req.Title
			mockPosts[i].Content = req.Content
			mockPosts[i].CategoryID = req.CategoryID
			mockPosts[i].Tags = req.Tags
			mockPosts[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "帖子更新成功",
				"data": mockPosts[i],
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func deletePost(c *gin.Context) {
	id := c.Param("id")
	for i, post := range mockPosts {
		if post.ID == id {
			mockPosts = append(mockPosts[:i], mockPosts[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "帖子删除成功",
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func getPostStats(c *gin.Context) {
	totalPosts := len(mockPosts)
	totalViews := 0
	totalLikes := 0
	for _, post := range mockPosts {
		totalViews += post.ViewCount
		totalLikes += post.LikeCount
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"total_posts": totalPosts,
			"total_views": totalViews,
			"total_likes": totalLikes,
			"hot_posts": 2,
			"sticky_posts": 1,
		},
	})
}

func setPostSticky(c *gin.Context) {
	id := c.Param("id")
	sticky := c.Query("sticky") == "true"

	for i, post := range mockPosts {
		if post.ID == id {
			mockPosts[i].IsSticky = sticky
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "帖子置顶状态更新成�?,
				"data": mockPosts[i],
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func setPostHot(c *gin.Context) {
	id := c.Param("id")
	hot := c.Query("hot") == "true"

	for i, post := range mockPosts {
		if post.ID == id {
			mockPosts[i].IsHot = hot
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "帖子热门状态更新成�?,
				"data": mockPosts[i],
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func searchPosts(c *gin.Context) {
	keyword := c.Query("keyword")
	var results []Post

	for _, post := range mockPosts {
		if keyword == "" || 
		   post.Title == keyword || 
		   post.Content == keyword {
			results = append(results, post)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"posts": results,
			"total": len(results),
			"keyword": keyword,
		},
	})
}

func likePost(c *gin.Context) {
	postID := c.Param("post_id")
	for i, post := range mockPosts {
		if post.ID == postID {
			mockPosts[i].LikeCount++
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "点赞成功",
				"data": gin.H{"like_count": mockPosts[i].LikeCount},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func unlikePost(c *gin.Context) {
	postID := c.Param("post_id")
	for i, post := range mockPosts {
		if post.ID == postID && post.LikeCount > 0 {
			mockPosts[i].LikeCount--
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "取消点赞成功",
				"data": gin.H{"like_count": mockPosts[i].LikeCount},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "帖子不存�?})
}

func bookmarkPost(c *gin.Context) {
	postID := c.Param("post_id")
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "收藏成功",
		"data": gin.H{"post_id": postID},
	})
}

func unbookmarkPost(c *gin.Context) {
	postID := c.Param("post_id")
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "取消收藏成功",
		"data": gin.H{"post_id": postID},
	})
}

// 评论相关处理函数
func createComment(c *gin.Context) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	comment := Comment{
		ID: uuid.New().String(),
		PostID: req.PostID,
		ParentID: req.ParentID,
		AuthorID: "user1",
		AuthorName: "当前用户",
		Content: req.Content,
		LikeCount: 0,
		Status: "approved",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockComments = append(mockComments, comment)

	// 更新帖子评论�?
	for i, post := range mockPosts {
		if post.ID == req.PostID {
			mockPosts[i].CommentCount++
			break
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "评论创建成功",
		"data": comment,
	})
}

func getComments(c *gin.Context) {
	postID := c.Param("post_id")
	var comments []Comment

	for _, comment := range mockComments {
		if comment.PostID == postID {
			comments = append(comments, comment)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"comments": comments,
			"total": len(comments),
		},
	})
}

func getComment(c *gin.Context) {
	id := c.Param("id")
	for _, comment := range mockComments {
		if comment.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "success",
				"data": comment,
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "评论不存�?})
}

func updateComment(c *gin.Context) {
	id := c.Param("id")
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	for i, comment := range mockComments {
		if comment.ID == id {
			mockComments[i].Content = req.Content
			mockComments[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "评论更新成功",
				"data": mockComments[i],
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "评论不存�?})
}

func deleteComment(c *gin.Context) {
	id := c.Param("id")
	for i, comment := range mockComments {
		if comment.ID == id {
			// 更新帖子评论�?
			for j, post := range mockPosts {
				if post.ID == comment.PostID && post.CommentCount > 0 {
					mockPosts[j].CommentCount--
					break
				}
			}

			mockComments = append(mockComments[:i], mockComments[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "评论删除成功",
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "评论不存�?})
}

func getCommentStats(c *gin.Context) {
	totalComments := len(mockComments)
	totalLikes := 0
	for _, comment := range mockComments {
		totalLikes += comment.LikeCount
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"total_comments": totalComments,
			"total_likes": totalLikes,
			"approved_comments": totalComments,
			"pending_comments": 0,
		},
	})
}

func getUserComments(c *gin.Context) {
	userID := c.Param("user_id")
	var comments []Comment

	for _, comment := range mockComments {
		if comment.AuthorID == userID {
			comments = append(comments, comment)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"comments": comments,
			"total": len(comments),
		},
	})
}

func likeComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	for i, comment := range mockComments {
		if comment.ID == commentID {
			mockComments[i].LikeCount++
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "点赞成功",
				"data": gin.H{"like_count": mockComments[i].LikeCount},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "评论不存�?})
}

func unlikeComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	for i, comment := range mockComments {
		if comment.ID == commentID && comment.LikeCount > 0 {
			mockComments[i].LikeCount--
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "取消点赞成功",
				"data": gin.H{"like_count": mockComments[i].LikeCount},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "评论不存�?})
}

// 用户相关处理函数
func getMyProfile(c *gin.Context) {
	// 模拟当前用户为user1
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": mockUsers[0],
	})
}

func updateUserProfile(c *gin.Context) {
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 模拟更新当前用户资料
	mockUsers[0].Nickname = req.Nickname
	mockUsers[0].Bio = req.Bio
	mockUsers[0].Avatar = req.Avatar

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "用户资料更新成功",
		"data": mockUsers[0],
	})
}

func getUserProfile(c *gin.Context) {
	id := c.Param("id")
	for _, user := range mockUsers {
		if user.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "success",
				"data": user,
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存�?})
}

func getUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"users": mockUsers,
			"total": len(mockUsers),
			"page": page,
			"page_size": pageSize,
		},
	})
}

func getUserStats(c *gin.Context) {
	totalUsers := len(mockUsers)
	activeUsers := 0
	for _, user := range mockUsers {
		if user.Status == "active" {
			activeUsers++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"total_users": totalUsers,
			"active_users": activeUsers,
			"new_users_today": 2,
			"online_users": 15,
		},
	})
}

func searchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	var results []User

	for _, user := range mockUsers {
		if keyword == "" || 
		   user.Username == keyword || 
		   user.Nickname == keyword {
			results = append(results, user)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"users": results,
			"total": len(results),
			"keyword": keyword,
		},
	})
}

func banUser(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "用户封禁成功",
		"data": gin.H{"user_id": userID},
	})
}

func unbanUser(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "用户解封成功",
		"data": gin.H{"user_id": userID},
	})
}

func getUserPosts(c *gin.Context) {
	userID := c.Param("id")
	var posts []Post

	for _, post := range mockPosts {
		if post.AuthorID == userID {
			posts = append(posts, post)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"posts": posts,
			"total": len(posts),
		},
	})
}

func updateUserActivity(c *gin.Context) {
	userID := c.Param("id")
	for i, user := range mockUsers {
		if user.ID == userID {
			mockUsers[i].LastActiveAt = time.Now()
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "用户活跃度更新成�?,
				"data": mockUsers[i],
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存�?})
}

func followUser(c *gin.Context) {
	userID := c.Param("id")
	for i, user := range mockUsers {
		if user.ID == userID {
			mockUsers[i].FollowerCount++
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "关注成功",
				"data": gin.H{"follower_count": mockUsers[i].FollowerCount},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存�?})
}

func unfollowUser(c *gin.Context) {
	userID := c.Param("id")
	for i, user := range mockUsers {
		if user.ID == userID && user.FollowerCount > 0 {
			mockUsers[i].FollowerCount--
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "取消关注成功",
				"data": gin.H{"follower_count": mockUsers[i].FollowerCount},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存�?})
}

func getUserFollowers(c *gin.Context) {
	userID := c.Param("id")
	// 模拟返回关注者列�?
	followers := []User{mockUsers[1], mockUsers[2]}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"followers": followers,
			"total": len(followers),
			"user_id": userID,
		},
	})
}

func getUserFollowing(c *gin.Context) {
	userID := c.Param("id")
	// 模拟返回关注列表
	following := []User{mockUsers[0]}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"following": following,
			"total": len(following),
			"user_id": userID,
		},
	})
}

// 聊天相关处理函数
func createChatRoom(c *gin.Context) {
	var req CreateChatRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	room := ChatRoom{
		ID: uuid.New().String(),
		Name: req.Name,
		Description: req.Description,
		Type: req.Type,
		CreatorID: "user1",
		MemberCount: 1,
		MaxMembers: req.MaxMembers,
		IsPrivate: req.IsPrivate,
		CreatedAt: time.Now(),
	}

	mockChatRooms = append(mockChatRooms, room)

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "聊天室创建成�?,
		"data": room,
	})
}

func getChatRooms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"rooms": mockChatRooms,
			"total": len(mockChatRooms),
		},
	})
}

func joinChatRoom(c *gin.Context) {
	roomID := c.Param("room_id")
	for i, room := range mockChatRooms {
		if room.ID == roomID {
			mockChatRooms[i].MemberCount++
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "加入聊天室成�?,
				"data": gin.H{
					"room_id": roomID,
					"member_count": mockChatRooms[i].MemberCount,
				},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "聊天室不存在"})
}

func leaveChatRoom(c *gin.Context) {
	roomID := c.Param("room_id")
	for i, room := range mockChatRooms {
		if room.ID == roomID && room.MemberCount > 0 {
			mockChatRooms[i].MemberCount--
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "离开聊天室成�?,
				"data": gin.H{
					"room_id": roomID,
					"member_count": mockChatRooms[i].MemberCount,
				},
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "聊天室不存在"})
}

func getChatMessages(c *gin.Context) {
	roomID := c.Param("room_id")
	var messages []ChatMessage

	for _, message := range mockChatMessages {
		if message.RoomID == roomID {
			messages = append(messages, message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"messages": messages,
			"total": len(messages),
			"room_id": roomID,
		},
	})
}

func sendMessage(c *gin.Context) {
	roomID := c.Param("room_id")
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	message := ChatMessage{
		ID: uuid.New().String(),
		RoomID: roomID,
		UserID: "user1",
		Username: "当前用户",
		Content: req.Content,
		Type: req.Type,
		SentAt: time.Now(),
	}

	mockChatMessages = append(mockChatMessages, message)

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "消息发送成�?,
		"data": message,
	})
}

func getOnlineUsers(c *gin.Context) {
	// 模拟在线用户
	onlineUsers := []gin.H{
		{"user_id": "user1", "username": "爱丽�?, "status": "online"},
		{"user_id": "user2", "username": "鲍勃", "status": "online"},
		{"user_id": "user3", "username": "查理", "status": "away"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"online_users": onlineUsers,
			"total": len(onlineUsers),
		},
	})
}

func getChatStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"total_rooms": len(mockChatRooms),
			"total_messages": len(mockChatMessages),
			"online_users": 3,
			"active_rooms": 2,
		},
	})
}

// 健康检�?
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "community-service",
		"version": "1.0.0",
		"timestamp": time.Now().Unix(),
		"uptime": "running",
	})
}
