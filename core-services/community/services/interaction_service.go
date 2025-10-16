package services

import (
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InteractionService 互动服务
type InteractionService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewInteractionService 创建互动服务实例
func NewInteractionService(db *gorm.DB, logger *zap.Logger) *InteractionService {
	return &InteractionService{
		db:     db,
		logger: logger,
	}
}

// LikePost 点赞帖子
func (s *InteractionService) LikePost(userID, postID string) (*models.Like, error) {
	// 检查是否已经点赞过?
	var existingLike models.Like
	err := s.db.Where("user_id = ? AND post_id = ? AND type = ?", userID, postID, models.LikeTypePost).First(&existingLike).Error
	if err == nil {
		return nil, fmt.Errorf("已经点赞过了")
	}
	if err != gorm.ErrRecordNotFound {
		s.logger.Error("Failed to check existing like", zap.Error(err))
		return nil, fmt.Errorf("检查点赞状态失? %v", err)
	}

	// 验证帖子是否存在
	var post models.Post
	if err := s.db.Where("id = ? AND status = ?", postID, models.PostStatusPublished).First(&post).Error; err != nil {
		s.logger.Error("Post not found", zap.String("post_id", postID), zap.Error(err))
		return nil, fmt.Errorf("帖子不存在或已被删除")
	}

	// 创建点赞记录
	like := &models.Like{
		ID:     uuid.New().String(),
		UserID: userID,
		PostID: &postID,
		Type:   models.LikeTypePost,
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存点赞记录
	if err := tx.Create(like).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create like", zap.Error(err))
		return nil, fmt.Errorf("点赞失败: %v", err)
	}

	// 更新帖子点赞?
	if err := tx.Model(&post).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update post like count", zap.Error(err))
		return nil, fmt.Errorf("更新帖子统计失败: %v", err)
	}

	// 更新作者获赞数
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", post.AuthorID).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update author like count", zap.Error(err))
		return nil, fmt.Errorf("更新作者统计失? %v", err)
	}

	tx.Commit()

	s.logger.Info("Post liked successfully", zap.String("user_id", userID), zap.String("post_id", postID))
	return like, nil
}

// UnlikePost 取消点赞帖子
func (s *InteractionService) UnlikePost(userID, postID string) error {
	// 查找点赞记录
	var like models.Like
	if err := s.db.Where("user_id = ? AND post_id = ? AND type = ?", userID, postID, models.LikeTypePost).First(&like).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("未找到点赞记录")
		}
		s.logger.Error("Failed to find like record", zap.Error(err))
		return fmt.Errorf("查找点赞记录失败: %v", err)
	}

	// 获取帖子信息
	var post models.Post
	if err := s.db.Where("id = ?", postID).First(&post).Error; err != nil {
		s.logger.Error("Post not found", zap.String("post_id", postID), zap.Error(err))
		return fmt.Errorf("帖子不存在或已被删除")
	}

	// 开启事?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除点赞记录
	if err := tx.Delete(&like).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete like", zap.Error(err))
		return fmt.Errorf("取消点赞失败: %v", err)
	}

	// 更新帖子点赞?
	if err := tx.Model(&post).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update post like count", zap.Error(err))
		return fmt.Errorf("更新帖子统计失败: %v", err)
	}

	// 更新作者获赞数
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", post.AuthorID).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update author like count", zap.Error(err))
		return fmt.Errorf("更新作者统计失? %v", err)
	}

	tx.Commit()

	s.logger.Info("Post unliked successfully", zap.String("user_id", userID), zap.String("post_id", postID))
	return nil
}

// LikeComment 点赞评论
func (s *InteractionService) LikeComment(userID, commentID string) (*models.Like, error) {
	// 检查是否已经点赞过?
	var existingLike models.Like
	err := s.db.Where("user_id = ? AND comment_id = ? AND type = ?", userID, commentID, models.LikeTypeComment).First(&existingLike).Error
	if err == nil {
		return nil, fmt.Errorf("已经点赞过了")
	}
	if err != gorm.ErrRecordNotFound {
		s.logger.Error("Failed to check existing comment like", zap.Error(err))
		return nil, fmt.Errorf("检查点赞状态失? %v", err)
	}

	// 验证评论是否存在
	var comment models.Comment
	if err := s.db.Where("id = ? AND status = ?", commentID, models.CommentStatusPublished).First(&comment).Error; err != nil {
		s.logger.Error("Comment not found", zap.String("comment_id", commentID), zap.Error(err))
		return nil, fmt.Errorf("评论不存在或已被删除")
	}

	// 创建点赞记录
	like := &models.Like{
		ID:        uuid.New().String(),
		UserID:    userID,
		CommentID: &commentID,
		Type:      models.LikeTypeComment,
	}

	// 开启事?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存点赞记录
	if err := tx.Create(like).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create comment like", zap.Error(err))
		return nil, fmt.Errorf("点赞失败: %v", err)
	}

	// 更新评论点赞?
	if err := tx.Model(&comment).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update comment like count", zap.Error(err))
		return nil, fmt.Errorf("更新评论统计失败")
	}

	tx.Commit()

	s.logger.Info("Comment liked successfully", zap.String("user_id", userID), zap.String("comment_id", commentID))
	return like, nil
}

// UnlikeComment 取消点赞评论
func (s *InteractionService) UnlikeComment(userID, commentID string) error {
	// 查找点赞记录
	var like models.Like
	if err := s.db.Where("user_id = ? AND comment_id = ? AND type = ?", userID, commentID, models.LikeTypeComment).First(&like).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("未找到点赞记录")
		}
		s.logger.Error("Failed to find comment like record", zap.Error(err))
		return fmt.Errorf("查找点赞记录失败: %v", err)
	}

	// 开启事?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除点赞记录
	if err := tx.Delete(&like).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete comment like", zap.Error(err))
		return fmt.Errorf("取消点赞失败")
	}

	// 更新评论点赞?
	if err := tx.Model(&models.Comment{}).Where("id = ?", commentID).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update comment like count", zap.Error(err))
		return fmt.Errorf("更新评论统计失败: %v", err)
	}

	tx.Commit()

	s.logger.Info("Comment unliked successfully", zap.String("user_id", userID), zap.String("comment_id", commentID))
	return nil
}

// FollowUser 关注用户
func (s *InteractionService) FollowUser(followerID, followingID string) (*models.Follow, error) {
	// 不能关注自己
	if followerID == followingID {
		return nil, fmt.Errorf("不能关注自己")
	}

	// 检查是否已经关注过?
	var existingFollow models.Follow
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&existingFollow).Error
	if err == nil {
		return nil, fmt.Errorf("已经关注过了")
	}
	if err != gorm.ErrRecordNotFound {
		s.logger.Error("Failed to check existing follow", zap.Error(err))
		return nil, fmt.Errorf("检查关注状态失? %v", err)
	}

	// 验证被关注用户是否存在且活跃
	var followingUser models.UserProfile
	if err := s.db.Where("user_id = ? AND status = ?", followingID, models.UserStatusActive).First(&followingUser).Error; err != nil {
		s.logger.Error("Following user not found", zap.String("following_id", followingID), zap.Error(err))
		return nil, fmt.Errorf("用户不存在或已被禁用")
	}

	// 创建关注记录
	follow := &models.Follow{
		ID:          uuid.New().String(),
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	// 开启事?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存关注记录
	if err := tx.Create(follow).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create follow", zap.Error(err))
		return nil, fmt.Errorf("关注失败: %v", err)
	}

	// 更新关注者的关注?
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update follower following count", zap.Error(err))
		return nil, fmt.Errorf("更新关注统计失败: %v", err)
	}

	// 更新被关注者的粉丝?
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update following follower count", zap.Error(err))
		return nil, fmt.Errorf("更新粉丝统计失败: %v", err)
	}

	tx.Commit()

	// 加载被关注用户信?
	follow.Following = &followingUser

	s.logger.Info("User followed successfully", zap.String("follower_id", followerID), zap.String("following_id", followingID))
	return follow, nil
}

// UnfollowUser 取消关注用户
func (s *InteractionService) UnfollowUser(followerID, followingID string) error {
	// 查找关注记录
	var follow models.Follow
	if err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("未找到关注记录")
		}
		s.logger.Error("Failed to find follow record", zap.Error(err))
		return fmt.Errorf("查找关注记录失败: %v", err)
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除关注记录
	if err := tx.Delete(&follow).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete follow", zap.Error(err))
		return fmt.Errorf("取消关注失败")
	}

	// 更新关注者的关注?
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update follower following count", zap.Error(err))
		return fmt.Errorf("更新关注统计失败")
	}

	// 更新被关注者的粉丝?
	if err := tx.Model(&models.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update following follower count", zap.Error(err))
		return fmt.Errorf("更新粉丝统计失败: %v", err)
	}

	tx.Commit()

	s.logger.Info("User unfollowed successfully", zap.String("follower_id", followerID), zap.String("following_id", followingID))
	return nil
}

// BookmarkPost 收藏帖子
func (s *InteractionService) BookmarkPost(userID, postID string) (*models.Bookmark, error) {
	// 检查是否已经收藏过?
	var existingBookmark models.Bookmark
	err := s.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingBookmark).Error
	if err == nil {
		return nil, fmt.Errorf("已经收藏过了")
	}
	if err != gorm.ErrRecordNotFound {
		s.logger.Error("Failed to check existing bookmark", zap.Error(err))
		return nil, fmt.Errorf("检查收藏状态失? %v", err)
	}

	// 验证帖子是否存在
	var post models.Post
	if err := s.db.Where("id = ? AND status = ?", postID, models.PostStatusPublished).First(&post).Error; err != nil {
		s.logger.Error("Post not found", zap.String("post_id", postID), zap.Error(err))
		return nil, fmt.Errorf("帖子不存在或已被禁用")
	}

	// 创建收藏记录
	bookmark := &models.Bookmark{
		ID:     uuid.New().String(),
		UserID: userID,
		PostID: postID,
	}

	if err := s.db.Create(bookmark).Error; err != nil {
		s.logger.Error("Failed to create bookmark", zap.Error(err))
		return nil, fmt.Errorf("收藏失败: %v", err)
	}

	// 加载帖子信息
	bookmark.Post = &post

	s.logger.Info("Post bookmarked successfully", zap.String("user_id", userID), zap.String("post_id", postID))
	return bookmark, nil
}

// UnbookmarkPost 取消收藏帖子
func (s *InteractionService) UnbookmarkPost(userID, postID string) error {
	// 查找收藏记录
	var bookmark models.Bookmark
	if err := s.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&bookmark).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("未找到收藏记录")
		}
		s.logger.Error("Failed to find bookmark record", zap.Error(err))
		return fmt.Errorf("查找收藏记录失败: %v", err)
	}

	// 删除收藏记录
	if err := s.db.Delete(&bookmark).Error; err != nil {
		s.logger.Error("Failed to delete bookmark", zap.Error(err))
		return fmt.Errorf("取消收藏失败: %v", err)
	}

	s.logger.Info("Post unbookmarked successfully", zap.String("user_id", userID), zap.String("post_id", postID))
	return nil
}

// GetUserBookmarks 获取用户收藏列表
func (s *InteractionService) GetUserBookmarks(userID string, page, pageSize int) ([]models.BookmarkResponse, int64, error) {
	var bookmarks []models.Bookmark
	var total int64

	// 构建查询
	query := s.db.Model(&models.Bookmark{}).
		Preload("Post").
		Preload("Post.Author").
		Where("user_id = ?", userID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count bookmarks", zap.Error(err))
		return nil, 0, fmt.Errorf("获取收藏数量失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&bookmarks).Error; err != nil {
		s.logger.Error("Failed to get bookmarks", zap.Error(err))
		return nil, 0, fmt.Errorf("获取收藏列表失败")
	}

	// 转换为响应格?
	bookmarkResponses := make([]models.BookmarkResponse, len(bookmarks))
	for i, bookmark := range bookmarks {
		bookmarkResponses[i] = bookmark.ToResponse()
	}

	return bookmarkResponses, total, nil
}

// GetUserFollowers 获取用户粉丝列表
func (s *InteractionService) GetUserFollowers(userID string, page, pageSize int) ([]models.FollowResponse, int64, error) {
	var follows []models.Follow
	var total int64

	// 构建查询
	query := s.db.Model(&models.Follow{}).
		Preload("Follower").
		Where("following_id = ?", userID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count followers", zap.Error(err))
		return nil, 0, fmt.Errorf("获取粉丝数量失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		s.logger.Error("Failed to get followers", zap.Error(err))
		return nil, 0, fmt.Errorf("获取粉丝列表失败")
	}

	// 转换为响应格?
	followResponses := make([]models.FollowResponse, len(follows))
	for i, follow := range follows {
		followResponses[i] = follow.ToResponse()
	}

	return followResponses, total, nil
}

// GetUserFollowing 获取用户关注列表
func (s *InteractionService) GetUserFollowing(userID string, page, pageSize int) ([]models.FollowResponse, int64, error) {
	var follows []models.Follow
	var total int64

	// 构建查询
	query := s.db.Model(&models.Follow{}).
		Preload("Following").
		Where("follower_id = ?", userID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count following", zap.Error(err))
		return nil, 0, fmt.Errorf("获取关注数量失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		s.logger.Error("Failed to get following", zap.Error(err))
		return nil, 0, fmt.Errorf("获取关注列表失败")
	}

	// 转换为响应格?
	followResponses := make([]models.FollowResponse, len(follows))
	for i, follow := range follows {
		followResponses[i] = follow.ToResponse()
	}

	return followResponses, total, nil
}

// GetInteractionStats 获取互动统计
func (s *InteractionService) GetInteractionStats() (*models.InteractionStatsResponse, error) {
	var stats models.InteractionStatsResponse

	// 总点赞数
	s.db.Model(&models.Like{}).Count(&stats.TotalLikes)

	// 总关注数
	s.db.Model(&models.Follow{}).Count(&stats.TotalFollows)

	// 总收藏数
	s.db.Model(&models.Bookmark{}).Count(&stats.TotalBookmarks)

	// 总举报数
	s.db.Model(&models.Report{}).Count(&stats.TotalReports)

	// 今日数据
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Like{}).Where("created_at >= ?", today).Count(&stats.TodayLikes)
	s.db.Model(&models.Follow{}).Where("created_at >= ?", today).Count(&stats.TodayFollows)
	s.db.Model(&models.Bookmark{}).Where("created_at >= ?", today).Count(&stats.TodayBookmarks)
	s.db.Model(&models.Report{}).Where("created_at >= ?", today).Count(&stats.TodayReports)

	return &stats, nil
}

// IsPostLiked 检查用户是否点赞了帖子
func (s *InteractionService) IsPostLiked(userID, postID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Like{}).
		Where("user_id = ? AND post_id = ? AND type = ?", userID, postID, models.LikeTypePost).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to check post like status", zap.Error(err))
		return false, fmt.Errorf("检查点赞状态失? %v", err)
	}
	return count > 0, nil
}

// IsCommentLiked 检查用户是否点赞了评论
func (s *InteractionService) IsCommentLiked(userID, commentID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Like{}).
		Where("user_id = ? AND comment_id = ? AND type = ?", userID, commentID, models.LikeTypeComment).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to check comment like status", zap.Error(err))
		return false, fmt.Errorf("检查点赞状态失? %v", err)
	}
	return count > 0, nil
}

// IsUserFollowed 检查是否关注了用户
func (s *InteractionService) IsUserFollowed(followerID, followingID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to check follow status", zap.Error(err))
		return false, fmt.Errorf("检查关注状态失? %v", err)
	}
	return count > 0, nil
}

// IsPostBookmarked 检查用户是否收藏了帖子
func (s *InteractionService) IsPostBookmarked(userID, postID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Bookmark{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to check bookmark status", zap.Error(err))
		return false, fmt.Errorf("检查收藏状态失? %v", err)
	}
	return count > 0, nil
}

