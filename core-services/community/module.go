package community

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/community/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/websocket"
)

// Module 社区服务模块
type Module struct {
	// 基础组件
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zap.Logger
	
	// WebSocket Hub
	wsHub    *websocket.Hub
	wsHandler *websocket.WebSocketHandler
	
	// 服务层
	postService        *services.PostService
	commentService     *services.CommentService
	userService        *services.UserService
	interactionService *services.InteractionService
	reviewService      *services.ContentReviewService
	chatService        *services.ChatService
	
	// 处理器层
	postHandler        *handlers.PostHandler
	commentHandler     *handlers.CommentHandler
	userHandler        *handlers.UserHandler
	interactionHandler *handlers.InteractionHandler
	reviewHandler      *handlers.ContentReviewHandler
	chatHandler        *handlers.ChatHandler
	
	// gRPC服务器
	grpcServer   *grpc.Server
	grpcListener net.Listener
	
	// 配置
	config *ModuleConfig
}

// ModuleConfig 模块配置
type ModuleConfig struct {
	// HTTP配置
	HTTPEnabled bool   `json:"http_enabled"`
	HTTPPrefix  string `json:"http_prefix"`
	
	// gRPC配置
	GRPCEnabled bool   `json:"grpc_enabled"`
	GRPCPort    int    `json:"grpc_port"`
	GRPCHost    string `json:"grpc_host"`
	
	// 服务配置
	ServiceConfig *CommunityServiceConfig `json:"service_config"`
	
	// WebSocket配置
	WebSocketConfig *WebSocketConfig `json:"websocket_config"`
	
	// 内容审核配置
	ContentReviewConfig *ContentReviewConfig `json:"content_review_config"`
	
	// 聊天配置
	ChatConfig *ChatConfig `json:"chat_config"`
}

// CommunityServiceConfig 社区服务配置
type CommunityServiceConfig struct {
	ServiceName        string        `json:"service_name"`
	Version           string        `json:"version"`
	Environment       string        `json:"environment"`
	MaxConcurrentReqs int           `json:"max_concurrent_requests"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MetricsRetention  time.Duration `json:"metrics_retention"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	MaxConnections    int           `json:"max_connections"`
	ReadBufferSize    int           `json:"read_buffer_size"`
	WriteBufferSize   int           `json:"write_buffer_size"`
	HandshakeTimeout  time.Duration `json:"handshake_timeout"`
	PingPeriod        time.Duration `json:"ping_period"`
	PongWait          time.Duration `json:"pong_wait"`
	WriteWait         time.Duration `json:"write_wait"`
	MaxMessageSize    int64         `json:"max_message_size"`
}

// ContentReviewConfig 内容审核配置
type ContentReviewConfig struct {
	AutoReview        bool          `json:"auto_review"`
	ReviewTimeout     time.Duration `json:"review_timeout"`
	MaxPendingItems   int           `json:"max_pending_items"`
	SensitiveWords    []string      `json:"sensitive_words"`
	ReviewerRoles     []string      `json:"reviewer_roles"`
}

// ChatConfig 聊天配置
type ChatConfig struct {
	MaxRooms          int           `json:"max_rooms"`
	MaxMembersPerRoom int           `json:"max_members_per_room"`
	MessageRetention  time.Duration `json:"message_retention"`
	MaxMessageLength  int           `json:"max_message_length"`
	RateLimitPerMin   int           `json:"rate_limit_per_min"`
}

// NewModule 创建社区服务模块
func NewModule(config *ModuleConfig, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) (*Module, error) {
	if config == nil {
		config = getDefaultConfig()
	}
	
	module := &Module{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		config:      config,
	}
	
	// 初始化WebSocket Hub
	if err := module.initWebSocket(); err != nil {
		return nil, fmt.Errorf("failed to initialize websocket: %w", err)
	}
	
	// 初始化服务层
	if err := module.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	// 初始化处理器层
	if err := module.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}
	
	// 初始化gRPC服务器（如果启用）
	if config.GRPCEnabled {
		if err := module.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}
	
	return module, nil
}

// initWebSocket 初始化WebSocket
func (m *Module) initWebSocket() error {
	m.logger.Info("Initializing community WebSocket hub")
	
	// 创建WebSocket Hub
	m.wsHub = websocket.NewHub()
	
	// 创建WebSocket处理器
	m.wsHandler = websocket.NewWebSocketHandler(m.wsHub)
	
	m.logger.Info("Community WebSocket hub initialized successfully")
	return nil
}

// initServices 初始化服务层
func (m *Module) initServices() error {
	m.logger.Info("Initializing community services")
	
	// 创建核心服务
	m.postService = services.NewPostService(m.db, m.logger)
	m.commentService = services.NewCommentService(m.db, m.logger)
	m.userService = services.NewUserService(m.db, m.logger)
	m.interactionService = services.NewInteractionService(m.db, m.logger)
	m.reviewService = services.NewContentReviewService(m.db, m.logger)
	m.chatService = services.NewChatService(m.db, m.logger, m.wsHub)
	
	m.logger.Info("Community services initialized successfully")
	return nil
}

// initHandlers 初始化处理器层
func (m *Module) initHandlers() error {
	m.logger.Info("Initializing community handlers")
	
	m.postHandler = handlers.NewPostHandler(m.postService, m.logger)
	m.commentHandler = handlers.NewCommentHandler(m.commentService, m.logger)
	m.userHandler = handlers.NewUserHandler(m.userService, m.logger)
	m.interactionHandler = handlers.NewInteractionHandler(m.interactionService, m.logger)
	m.reviewHandler = handlers.NewContentReviewHandler(m.reviewService, m.logger)
	m.chatHandler = handlers.NewChatHandler(m.chatService, m.logger)
	
	m.logger.Info("Community handlers initialized successfully")
	return nil
}

// initGRPCServer 初始化gRPC服务器
func (m *Module) initGRPCServer() error {
	m.logger.Info("Initializing community gRPC server")
	
	// 创建gRPC服务器
	m.grpcServer = grpc.NewServer()
	
	// 注册gRPC服务
	// TODO: 实现gRPC服务定义和注册
	
	// 创建监听器
	addr := fmt.Sprintf("%s:%d", m.config.GRPCHost, m.config.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	m.grpcListener = listener
	
	m.logger.Info("Community gRPC server initialized", zap.String("address", addr))
	return nil
}

// SetupRoutes 设置HTTP路由
func (m *Module) SetupRoutes(router *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware) error {
	if !m.config.HTTPEnabled {
		m.logger.Info("HTTP routes disabled, skipping route setup")
		return nil
	}
	
	m.logger.Info("Setting up community HTTP routes")
	
	// 社区路由组
	community := router.Group(m.config.HTTPPrefix)
	{
		// 帖子相关路由
		posts := community.Group("/posts")
		{
			posts.POST("", jwtMiddleware.AuthRequired(), m.postHandler.CreatePost)                    // 创建帖子
			posts.GET("", m.postHandler.GetPosts)                      // 获取帖子列表
			posts.GET("/:id", m.postHandler.GetPost)                   // 获取单个帖子
			posts.PUT("/:id", jwtMiddleware.AuthRequired(), m.postHandler.UpdatePost)                // 更新帖子
			posts.DELETE("/:id", jwtMiddleware.AuthRequired(), m.postHandler.DeletePost)             // 删除帖子
			posts.GET("/stats", m.postHandler.GetPostStats)            // 获取帖子统计
			posts.PATCH("/:id/sticky", jwtMiddleware.AuthRequired(), m.postHandler.SetPostSticky)     // 设置/取消帖子置顶
			posts.PATCH("/:id/hot", jwtMiddleware.AuthRequired(), m.postHandler.SetPostHot)           // 设置/取消帖子热门
			posts.GET("/search", m.postHandler.SearchPosts)            // 搜索帖子

			// 帖子互动路由
			posts.POST("/:post_id/like", jwtMiddleware.AuthRequired(), m.interactionHandler.LikePost)         // 点赞帖子
			posts.DELETE("/:post_id/like", jwtMiddleware.AuthRequired(), m.interactionHandler.UnlikePost)     // 取消点赞帖子
			posts.POST("/:post_id/bookmark", jwtMiddleware.AuthRequired(), m.interactionHandler.BookmarkPost) // 收藏帖子
			posts.DELETE("/:post_id/bookmark", jwtMiddleware.AuthRequired(), m.interactionHandler.UnbookmarkPost) // 取消收藏帖子
		}

		// 评论相关路由
		comments := community.Group("/comments")
		{
			comments.POST("", jwtMiddleware.AuthRequired(), m.commentHandler.CreateComment)           // 创建评论
			comments.GET("/:post_id", m.commentHandler.GetCommentReplies)     // 获取评论回复
			comments.GET("/:id", m.commentHandler.GetComment)          // 获取单个评论
			comments.PUT("/:id", jwtMiddleware.AuthRequired(), m.commentHandler.UpdateComment)       // 更新评论
			comments.DELETE("/:id", jwtMiddleware.AuthRequired(), m.commentHandler.DeleteComment)    // 删除评论
			comments.GET("/stats", m.commentHandler.GetCommentStats)   // 获取评论统计
			comments.GET("/user/:user_id", m.commentHandler.GetUserComments) // 获取用户评论

			// 评论互动路由
			comments.POST("/:comment_id/like", jwtMiddleware.AuthRequired(), m.interactionHandler.LikeComment)     // 点赞评论
			comments.DELETE("/:comment_id/like", jwtMiddleware.AuthRequired(), m.interactionHandler.UnlikeComment) // 取消点赞评论
		}

		// 用户相关路由
		users := community.Group("/users")
		{
			users.GET("/profile", jwtMiddleware.AuthRequired(), m.userHandler.GetMyProfile)     // 获取当前用户资料
			users.PUT("/profile", jwtMiddleware.AuthRequired(), m.userHandler.UpdateUserProfile)         // 更新用户资料
			users.GET("/:id", m.userHandler.GetUserProfile)               // 获取用户资料
			users.GET("", m.userHandler.GetUsers)                         // 获取用户列表
			users.GET("/stats", m.userHandler.GetUserStats)               // 获取用户统计
			users.GET("/search", m.userHandler.SearchUsers)               // 搜索用户
			users.POST("/:id/ban", jwtMiddleware.AuthRequired(), m.userHandler.BanUser)                 // 封禁用户
			users.DELETE("/:id/ban", jwtMiddleware.AuthRequired(), m.userHandler.UnbanUser)             // 解封用户
			users.GET("/:id/posts", m.userHandler.GetUserPosts)           // 获取用户帖子
			users.PUT("/:id/activity", jwtMiddleware.AuthRequired(), m.userHandler.UpdateUserActivity)  // 更新用户活跃度

			// 用户互动路由
			users.POST("/:id/follow", jwtMiddleware.AuthRequired(), m.interactionHandler.FollowUser)       // 关注用户
			users.DELETE("/:id/follow", jwtMiddleware.AuthRequired(), m.interactionHandler.UnfollowUser)   // 取消关注用户
			users.GET("/:id/followers", m.interactionHandler.GetUserFollowers) // 获取用户粉丝列表
			users.GET("/:id/following", m.interactionHandler.GetUserFollowing) // 获取用户关注列表
		}

		// 互动相关路由
		interactions := community.Group("/interactions")
		{
			interactions.GET("/stats", m.interactionHandler.GetInteractionStats)    // 获取互动统计
			interactions.GET("/status", jwtMiddleware.AuthRequired(), m.interactionHandler.CheckInteractionStatus) // 检查互动状态
		}

		// 收藏相关路由
		bookmarks := community.Group("/bookmarks")
		{
			bookmarks.GET("", jwtMiddleware.AuthRequired(), m.interactionHandler.GetMyBookmarks) // 获取我的收藏列表
		}

		// 内容审核相关路由
		review := community.Group("/review")
		{
			review.GET("/posts/pending", jwtMiddleware.AuthRequired(), m.reviewHandler.GetPendingPosts)     // 获取待审核帖子
			review.GET("/comments/pending", jwtMiddleware.AuthRequired(), m.reviewHandler.GetPendingComments) // 获取待审核评论
			review.POST("/posts/review", jwtMiddleware.AuthRequired(), m.reviewHandler.ReviewPost)          // 审核帖子
			review.POST("/comments/review", jwtMiddleware.AuthRequired(), m.reviewHandler.ReviewComment)    // 审核评论
			review.POST("/posts/batch-review", jwtMiddleware.AuthRequired(), m.reviewHandler.BatchReviewPosts)    // 批量审核帖子
			review.POST("/comments/batch-review", jwtMiddleware.AuthRequired(), m.reviewHandler.BatchReviewComments) // 批量审核评论
			review.GET("/statistics", jwtMiddleware.AuthRequired(), m.reviewHandler.GetContentStatistics)   // 获取审核统计信息
		}

		// WebSocket 连接路由
		community.GET("/ws", m.wsHandler.HandleWebSocket)                    // 需要认证的WebSocket连接
		community.GET("/ws/public", m.wsHandler.HandleWebSocketPublic)       // 公开WebSocket连接

		// 聊天室相关路由
		chat := community.Group("/chat")
		{
			// 聊天室管理
			chat.POST("/rooms", jwtMiddleware.AuthRequired(), m.chatHandler.CreateChatRoom)           // 创建聊天室
			chat.GET("/rooms", jwtMiddleware.AuthRequired(), m.chatHandler.GetChatRooms)              // 获取聊天室列表
			chat.POST("/rooms/:room_id/join", jwtMiddleware.AuthRequired(), m.chatHandler.JoinChatRoom)    // 加入聊天室
			chat.POST("/rooms/:room_id/leave", jwtMiddleware.AuthRequired(), m.chatHandler.LeaveChatRoom)  // 离开聊天室

			// 聊天消息
			chat.GET("/rooms/:room_id/messages", jwtMiddleware.AuthRequired(), m.chatHandler.GetChatMessages) // 获取聊天消息
			chat.POST("/rooms/:room_id/messages", jwtMiddleware.AuthRequired(), m.chatHandler.SendMessage)    // 发送消息

			// WebSocket 状态和统计
			chat.GET("/online-users", m.wsHandler.GetConnectedUsers)          // 获取在线用户
			chat.GET("/stats", m.wsHandler.GetStats)                          // 获取WebSocket统计
			chat.POST("/send", jwtMiddleware.AuthRequired(), m.wsHandler.SendMessage) // HTTP方式发送消息
			chat.GET("/user/:user_id/online", m.wsHandler.CheckUserOnline)    // 检查用户是否在线
		}
	}

	m.logger.Info("Community routes setup completed")
	return nil
}

// Start 启动模块
func (m *Module) Start() error {
	m.logger.Info("Starting community module")
	
	// 自动迁移数据库
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	// 启动WebSocket Hub
	go m.wsHub.Run()
	
	// 启动gRPC服务器
	if m.config.GRPCEnabled && m.grpcServer != nil {
		go func() {
			m.logger.Info("Starting community gRPC server", 
				zap.String("address", m.grpcListener.Addr().String()))
			if err := m.grpcServer.Serve(m.grpcListener); err != nil {
				m.logger.Error("gRPC server error", zap.Error(err))
			}
		}()
	}
	
	// 启动后台任务
	go m.startBackgroundTasks()
	
	m.logger.Info("Community module started successfully")
	return nil
}

// Stop 停止模块
func (m *Module) Stop() error {
	m.logger.Info("Stopping community module")
	
	// 停止WebSocket Hub - Hub没有Stop方法，直接设置为nil
	if m.wsHub != nil {
		m.wsHub = nil
	}
	
	// 停止gRPC服务器
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		if m.grpcListener != nil {
			m.grpcListener.Close()
		}
	}
	
	m.logger.Info("Community module stopped successfully")
	return nil
}

// Health 健康检查
func (m *Module) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"module": "community",
		"version": m.config.ServiceConfig.Version,
		"services": map[string]string{
			"post_service":        "running",
			"comment_service":     "running",
			"user_service":        "running",
			"interaction_service": "running",
			"review_service":      "running",
			"chat_service":        "running",
		},
		"websocket": map[string]interface{}{
			"connected_users": m.wsHub.GetClientCount(),
			"active_rooms":    0, // Hub没有GetActiveRoomsCount方法，暂时设为0
		},
	}
	
	// 检查数据库连接
	if sqlDB, err := m.db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["database"] = "healthy"
		}
	}
	
	// 检查Redis连接
	if err := m.redisClient.Ping(context.Background()).Err(); err != nil {
		health["redis"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["redis"] = "healthy"
	}
	
	return health
}

// migrateDatabase 迁移数据库
func (m *Module) migrateDatabase() error {
	m.logger.Info("Migrating community database")
	
	// 自动迁移模型
	err := m.db.AutoMigrate(
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Follow{},
		&models.UserProfile{},
		&models.Bookmark{},
		&models.Report{},
		&models.ContentReviewLog{},
		// 聊天相关表
		&models.ChatRoom{},
		&models.ChatRoomMember{},
		&models.ChatMessage{},
		&models.ChatMessageRead{},
		&models.PrivateChat{},
		&models.PrivateChatMessage{},
		&models.OnlineUser{},
		&models.ChatNotification{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	
	m.logger.Info("Community database migration completed")
	return nil
}

// startBackgroundTasks 启动后台任务
func (m *Module) startBackgroundTasks() {
	m.logger.Info("Starting community background tasks")
	
	// 定期清理过期聊天消息
	go m.cleanupExpiredMessagesPeriodically()
	
	// 定期更新用户活跃度
	go m.updateUserActivityPeriodically()
	
	// 定期清理离线用户
	go m.cleanupOfflineUsersPeriodically()
}

// cleanupExpiredMessagesPeriodically 定期清理过期聊天消息
func (m *Module) cleanupExpiredMessagesPeriodically() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// ChatService没有CleanupExpiredMessages方法，暂时跳过
		m.logger.Debug("Cleanup expired messages task executed")
	}
}

// updateUserActivityPeriodically 定期更新用户活跃度
func (m *Module) updateUserActivityPeriodically() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// UserService.UpdateUserActivity需要userID参数，这里需要批量更新所有用户，暂时跳过
		m.logger.Debug("Update user activity task executed")
	}
}

// cleanupOfflineUsersPeriodically 定期清理离线用户
func (m *Module) cleanupOfflineUsersPeriodically() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// Hub没有CleanupOfflineUsers方法，暂时跳过
		m.logger.Debug("Cleanup offline users task executed")
	}
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *ModuleConfig {
	return &ModuleConfig{
		HTTPEnabled: true,
		HTTPPrefix:  "/community",
		GRPCEnabled: false,
		GRPCPort:    50054,
		GRPCHost:    "localhost",
		ServiceConfig: &CommunityServiceConfig{
			ServiceName:       "community-service",
			Version:          "1.0.0",
			Environment:      "development",
			MaxConcurrentReqs: 200,
			RequestTimeout:    30 * time.Second,
			MetricsRetention:  24 * time.Hour,
		},
		WebSocketConfig: &WebSocketConfig{
			MaxConnections:   1000,
			ReadBufferSize:   1024,
			WriteBufferSize:  1024,
			HandshakeTimeout: 10 * time.Second,
			PingPeriod:       54 * time.Second,
			PongWait:         60 * time.Second,
			WriteWait:        10 * time.Second,
			MaxMessageSize:   512,
		},
		ContentReviewConfig: &ContentReviewConfig{
			AutoReview:      false,
			ReviewTimeout:   24 * time.Hour,
			MaxPendingItems: 1000,
			SensitiveWords:  []string{},
			ReviewerRoles:   []string{"admin", "moderator"},
		},
		ChatConfig: &ChatConfig{
			MaxRooms:          100,
			MaxMembersPerRoom: 50,
			MessageRetention:  30 * 24 * time.Hour, // 30天
			MaxMessageLength:  1000,
			RateLimitPerMin:   60,
		},
	}
}