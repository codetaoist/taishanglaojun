package services

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
)

// ContentReviewService 内容审核服务
type ContentReviewService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewContentReviewService 创建内容审核服务实例
func NewContentReviewService(db *gorm.DB, logger *zap.Logger) *ContentReviewService {
	return &ContentReviewService{
		db:     db,
		logger: logger,
	}
}

// ReviewPostRequest 审核帖子请求
type ReviewPostRequest struct {
	PostID       string `json:"post_id" binding:"required"`
	ReviewerID   string `json:"reviewer_id" binding:"required"`
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	ReviewReason string `json:"review_reason,omitempty"`
}

// ReviewCommentRequest 审核评论请求
type ReviewCommentRequest struct {
	CommentID    string `json:"comment_id" binding:"required"`
	ReviewerID   string `json:"reviewer_id" binding:"required"`
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	ReviewReason string `json:"review_reason,omitempty"`
}

// ReviewPost 审核帖子
func (s *ContentReviewService) ReviewPost(req *ReviewPostRequest) error {
	// 验证帖子是否存在且处于待审核状�?
	var post models.Post
	if err := s.db.Where("id = ? AND status = ?", req.PostID, models.PostStatusPending).First(&post).Error; err != nil {
		s.logger.Error("Post not found or not pending review", zap.String("post_id", req.PostID), zap.Error(err))
		return fmt.Errorf("帖子不存在或不在待审核状�?)
	}

	// 验证审核员是否存�?
	var reviewer models.UserProfile
	if err := s.db.Where("user_id = ? AND status = ?", req.ReviewerID, models.UserStatusActive).First(&reviewer).Error; err != nil {
		s.logger.Error("Reviewer not found", zap.String("reviewer_id", req.ReviewerID), zap.Error(err))
		return fmt.Errorf("审核员不存在")
	}

	// 开启事�?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新帖子状�?
	var newStatus models.PostStatus
	if req.Action == "approve" {
		newStatus = models.PostStatusPublished
	} else {
		newStatus = models.PostStatusRejected
	}

	if err := tx.Model(&post).Updates(map[string]interface{}{
		"status":     newStatus,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update post status", zap.String("post_id", req.PostID), zap.Error(err))
		return fmt.Errorf("更新帖子状态失�?)
	}

	// 记录审核日志
	reviewLog := &models.ContentReviewLog{
		ContentType:  "post",
		ContentID:    req.PostID,
		ReviewerID:   req.ReviewerID,
		Action:       req.Action,
		ReviewReason: req.ReviewReason,
		ReviewedAt:   time.Now(),
	}

	if err := tx.Create(reviewLog).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create review log", zap.String("post_id", req.PostID), zap.Error(err))
		return fmt.Errorf("创建审核日志失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.String("post_id", req.PostID), zap.Error(err))
		return fmt.Errorf("提交事务失败")
	}

	s.logger.Info("Post reviewed successfully", 
		zap.String("post_id", req.PostID),
		zap.String("reviewer_id", req.ReviewerID),
		zap.String("action", req.Action),
		zap.String("new_status", string(newStatus)))

	return nil
}

// ReviewComment 审核评论
func (s *ContentReviewService) ReviewComment(req *ReviewCommentRequest) error {
	// 验证评论是否存在且处于待审核状�?
	var comment models.Comment
	if err := s.db.Where("id = ? AND status = ?", req.CommentID, models.CommentStatusPending).First(&comment).Error; err != nil {
		s.logger.Error("Comment not found or not pending review", zap.String("comment_id", req.CommentID), zap.Error(err))
		return fmt.Errorf("评论不存在或不在待审核状�?)
	}

	// 验证审核员是否存�?
	var reviewer models.UserProfile
	if err := s.db.Where("user_id = ? AND status = ?", req.ReviewerID, models.UserStatusActive).First(&reviewer).Error; err != nil {
		s.logger.Error("Reviewer not found", zap.String("reviewer_id", req.ReviewerID), zap.Error(err))
		return fmt.Errorf("审核员不存在")
	}

	// 开启事�?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新评论状�?
	var newStatus models.CommentStatus
	if req.Action == "approve" {
		newStatus = models.CommentStatusPublished
	} else {
		newStatus = models.CommentStatusRejected
	}

	if err := tx.Model(&comment).Updates(map[string]interface{}{
		"status":     newStatus,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update comment status", zap.String("comment_id", req.CommentID), zap.Error(err))
		return fmt.Errorf("更新评论状态失�?)
	}

	// 记录审核日志
	reviewLog := &models.ContentReviewLog{
		ContentType:  "comment",
		ContentID:    req.CommentID,
		ReviewerID:   req.ReviewerID,
		Action:       req.Action,
		ReviewReason: req.ReviewReason,
		ReviewedAt:   time.Now(),
	}

	if err := tx.Create(reviewLog).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create review log", zap.String("comment_id", req.CommentID), zap.Error(err))
		return fmt.Errorf("创建审核日志失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.String("comment_id", req.CommentID), zap.Error(err))
		return fmt.Errorf("提交事务失败")
	}

	s.logger.Info("Comment reviewed successfully", 
		zap.String("comment_id", req.CommentID),
		zap.String("reviewer_id", req.ReviewerID),
		zap.String("action", req.Action),
		zap.String("new_status", string(newStatus)))

	return nil
}

// GetPendingPosts 获取待审核的帖子列表
func (s *ContentReviewService) GetPendingPosts(page, pageSize int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPending).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count pending posts", zap.Error(err))
		return nil, 0, fmt.Errorf("获取待审核帖子数量失�?)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := s.db.Where("status = ?", models.PostStatusPending).
		Preload("Author").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		s.logger.Error("Failed to get pending posts", zap.Error(err))
		return nil, 0, fmt.Errorf("获取待审核帖子失�?)
	}

	return posts, total, nil
}

// GetPendingComments 获取待审核的评论列表
func (s *ContentReviewService) GetPendingComments(page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Comment{}).Where("status = ?", models.CommentStatusPending).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count pending comments", zap.Error(err))
		return nil, 0, fmt.Errorf("获取待审核评论数量失�?)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := s.db.Where("status = ?", models.CommentStatusPending).
		Preload("Author").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		s.logger.Error("Failed to get pending comments", zap.Error(err))
		return nil, 0, fmt.Errorf("获取待审核评论失�?)
	}

	return comments, total, nil
}

// BatchReviewRequest 批量审核请求
type BatchReviewRequest struct {
	ContentIDs   []string `json:"content_ids" binding:"required"`
	ReviewerID   string   `json:"reviewer_id" binding:"required"`
	Action       string   `json:"action" binding:"required,oneof=approve reject"`
	ReviewReason string   `json:"review_reason,omitempty"`
}

// BatchReviewResult 批量审核结果
type BatchReviewResult struct {
	ContentID string `json:"content_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// BatchReviewPosts 批量审核帖子
func (s *ContentReviewService) BatchReviewPosts(req *BatchReviewRequest) ([]BatchReviewResult, error) {
	var results []BatchReviewResult

	// 验证审核员是否存�?
	var reviewer models.UserProfile
	if err := s.db.Where("id = ?", req.ReviewerID).First(&reviewer).Error; err != nil {
		s.logger.Error("Reviewer not found", zap.String("reviewer_id", req.ReviewerID), zap.Error(err))
		return nil, fmt.Errorf("审核员不存在")
	}

	// 逐个处理每个帖子
	for _, postID := range req.ContentIDs {
		result := BatchReviewResult{
			ContentID: postID,
			Success:   false,
		}

		// 创建单个审核请求
		singleReq := &ReviewPostRequest{
			PostID:       postID,
			ReviewerID:   req.ReviewerID,
			Action:       req.Action,
			ReviewReason: req.ReviewReason,
		}

		// 执行审核
		if err := s.ReviewPost(singleReq); err != nil {
			result.Error = err.Error()
			s.logger.Warn("Failed to review post in batch", 
				zap.String("post_id", postID), 
				zap.Error(err))
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results, nil
}

// BatchReviewComments 批量审核评论
func (s *ContentReviewService) BatchReviewComments(req *BatchReviewRequest) ([]BatchReviewResult, error) {
	var results []BatchReviewResult

	// 验证审核员是否存�?
	var reviewer models.UserProfile
	if err := s.db.Where("id = ?", req.ReviewerID).First(&reviewer).Error; err != nil {
		s.logger.Error("Reviewer not found", zap.String("reviewer_id", req.ReviewerID), zap.Error(err))
		return nil, fmt.Errorf("审核员不存在")
	}

	// 逐个处理每个评论
	for _, commentID := range req.ContentIDs {
		result := BatchReviewResult{
			ContentID: commentID,
			Success:   false,
		}

		// 创建单个审核请求
		singleReq := &ReviewCommentRequest{
			CommentID:    commentID,
			ReviewerID:   req.ReviewerID,
			Action:       req.Action,
			ReviewReason: req.ReviewReason,
		}

		// 执行审核
		if err := s.ReviewComment(singleReq); err != nil {
			result.Error = err.Error()
			s.logger.Warn("Failed to review comment in batch", 
				zap.String("comment_id", commentID), 
				zap.Error(err))
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results, nil
}

// ContentStatistics 内容统计信息
type ContentStatistics struct {
	PendingPosts    int64 `json:"pending_posts"`
	PendingComments int64 `json:"pending_comments"`
	TotalPosts      int64 `json:"total_posts"`
	TotalComments   int64 `json:"total_comments"`
	PublishedPosts  int64 `json:"published_posts"`
	PublishedComments int64 `json:"published_comments"`
	RejectedPosts   int64 `json:"rejected_posts"`
	RejectedComments int64 `json:"rejected_comments"`
}

// GetContentStatistics 获取内容审核统计信息
func (s *ContentReviewService) GetContentStatistics() (*ContentStatistics, error) {
	stats := &ContentStatistics{}

	// 统计帖子数据
	if err := s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPending).Count(&stats.PendingPosts).Error; err != nil {
		return nil, fmt.Errorf("统计待审核帖子失�? %v", err)
	}

	if err := s.db.Model(&models.Post{}).Count(&stats.TotalPosts).Error; err != nil {
		return nil, fmt.Errorf("统计总帖子数失败: %v", err)
	}

	if err := s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusPublished).Count(&stats.PublishedPosts).Error; err != nil {
		return nil, fmt.Errorf("统计已发布帖子失�? %v", err)
	}

	if err := s.db.Model(&models.Post{}).Where("status = ?", models.PostStatusRejected).Count(&stats.RejectedPosts).Error; err != nil {
		return nil, fmt.Errorf("统计被拒绝帖子失�? %v", err)
	}

	// 统计评论数据
	if err := s.db.Model(&models.Comment{}).Where("status = ?", models.CommentStatusPending).Count(&stats.PendingComments).Error; err != nil {
		return nil, fmt.Errorf("统计待审核评论失�? %v", err)
	}

	if err := s.db.Model(&models.Comment{}).Count(&stats.TotalComments).Error; err != nil {
		return nil, fmt.Errorf("统计总评论数失败: %v", err)
	}

	if err := s.db.Model(&models.Comment{}).Where("status = ?", models.CommentStatusPublished).Count(&stats.PublishedComments).Error; err != nil {
		return nil, fmt.Errorf("统计已发布评论失�? %v", err)
	}

	if err := s.db.Model(&models.Comment{}).Where("status = ?", models.CommentStatusRejected).Count(&stats.RejectedComments).Error; err != nil {
		return nil, fmt.Errorf("统计被拒绝评论失�? %v", err)
	}

	return stats, nil
}
