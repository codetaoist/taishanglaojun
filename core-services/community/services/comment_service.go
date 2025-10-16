package services

import (
	"fmt"
	"math"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/community/utils"
)

// CommentService 
type CommentService struct {
	db               *gorm.DB
	logger           *zap.Logger
	contentValidator *utils.ContentValidator
}

// NewCommentService 
func NewCommentService(db *gorm.DB, logger *zap.Logger) *CommentService {
	return &CommentService{
		db:               db,
		logger:           logger,
		contentValidator: utils.NewContentValidator(),
	}
}

// CreateComment 
func (s *CommentService) CreateComment(userID string, req *models.CommentCreateRequest) (*models.Comment, error) {
	// 
	var userProfile models.UserProfile
	if err := s.db.Where("user_id = ? AND status = ?", userID, models.UserStatusActive).First(&userProfile).Error; err != nil {
		s.logger.Error("User not found or inactive", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	validationResult := s.contentValidator.ValidateCommentContent(req.Content)
	if !validationResult.IsValid {
		s.logger.Warn("Comment content validation failed",
			zap.String("user_id", userID),
			zap.String("post_id", req.PostID),
			zap.Strings("errors", validationResult.Errors),
			zap.Int("risk_level", validationResult.RiskLevel))
		return nil, fmt.Errorf(": %s", validationResult.Errors[0])
	}

	// 
	var post models.Post
	if err := s.db.Where("id = ? AND status = ?", req.PostID, models.PostStatusPublished).First(&post).Error; err != nil {
		s.logger.Error("Post not found", zap.String("post_id", req.PostID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	if req.ParentID != nil {
		var parentComment models.Comment
		if err := s.db.Where("id = ? AND post_id = ? AND status = ?", *req.ParentID, req.PostID, models.CommentStatusPublished).First(&parentComment).Error; err != nil {
			s.logger.Error("Parent comment not found", zap.String("parent_id", *req.ParentID), zap.Error(err))
			return nil, fmt.Errorf("")
		}
	}

	// 
	commentStatus := models.CommentStatusPublished
	if validationResult.RiskLevel > 0 {
		commentStatus = models.CommentStatusPending // 
		s.logger.Info("Comment requires review due to risk level",
			zap.String("user_id", userID),
			zap.String("post_id", req.PostID),
			zap.Int("risk_level", validationResult.RiskLevel),
			zap.Strings("warnings", validationResult.Warnings))
	}

	// 
	comment := &models.Comment{
		ID:       uuid.New().String(),
		PostID:   req.PostID,
		AuthorID: userID,
		ParentID: req.ParentID,
		Content:  req.Content,
		Status:   commentStatus,
	}

	// 
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 
	if err := tx.Create(comment).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create comment", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	if err := tx.Model(&post).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update post comment count", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	userProfile.AddExperience(5) // 5
	if err := tx.Save(&userProfile).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update user experience", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	tx.Commit()

	// 
	comment.Author = &userProfile

	s.logger.Info("Comment created successfully", zap.String("comment_id", comment.ID), zap.String("user_id", userID))
	return comment, nil
}

// GetComment 
func (s *CommentService) GetComment(commentID string) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.Preload("Author").Preload("Replies.Author").Where("id = ? AND status != ?", commentID, models.CommentStatusDeleted).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to get comment", zap.String("comment_id", commentID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	return &comment, nil
}

// GetComments 
func (s *CommentService) GetComments(req *models.CommentListRequest, userID *string) (*models.CommentListResponse, error) {
	var comments []models.Comment
	var total int64

	//  - 
	query := s.db.Model(&models.Comment{}).
		Preload("Author").
		Preload("Replies", "status = ?", models.CommentStatusPublished).
		Preload("Replies.Author").
		Where("post_id = ? AND parent_id IS NULL AND status = ?", req.PostID, models.CommentStatusPublished)

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count comments", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	switch req.SortBy {
	case "oldest":
		query = query.Order("created_at ASC")
	case "likes":
		query = query.Order("like_count DESC, created_at DESC")
	default: // latest
		query = query.Order("created_at DESC")
	}

	// 
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&comments).Error; err != nil {
		s.logger.Error("Failed to get comments", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	commentResponses := make([]models.CommentResponse, len(comments))
	for i, comment := range comments {
		commentResponses[i] = comment.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &models.CommentListResponse{
		Comments:   commentResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetReplies 
func (s *CommentService) GetReplies(commentID string, page, pageSize int) (*models.CommentListResponse, error) {
	var replies []models.Comment
	var total int64

	// 
	query := s.db.Model(&models.Comment{}).
		Preload("Author").
		Where("parent_id = ? AND status = ?", commentID, models.CommentStatusPublished)

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count replies", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&replies).Error; err != nil {
		s.logger.Error("Failed to get replies", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	replyResponses := make([]models.CommentResponse, len(replies))
	for i, reply := range replies {
		replyResponses[i] = reply.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.CommentListResponse{
		Comments:   replyResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateComment 
func (s *CommentService) UpdateComment(commentID, userID string, req *models.CommentUpdateRequest) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.Where("id = ? AND author_id = ?", commentID, userID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("")
		}
		s.logger.Error("Failed to find comment for update", zap.String("comment_id", commentID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	updates := make(map[string]interface{})
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := s.db.Model(&comment).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update comment", zap.String("comment_id", commentID), zap.Error(err))
			return nil, fmt.Errorf("")
		}
	}

	// 
	if err := s.db.Preload("Author").First(&comment, "id = ?", commentID).Error; err != nil {
		s.logger.Error("Failed to reload comment", zap.String("comment_id", commentID), zap.Error(err))
		return nil, fmt.Errorf("")
	}

	s.logger.Info("Comment updated successfully", zap.String("comment_id", commentID), zap.String("user_id", userID))
	return &comment, nil
}

// DeleteComment 
func (s *CommentService) DeleteComment(commentID, userID string) error {
	var comment models.Comment
	if err := s.db.Where("id = ? AND author_id = ?", commentID, userID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("")
		}
		s.logger.Error("Failed to find comment for deletion", zap.String("comment_id", commentID), zap.Error(err))
		return fmt.Errorf("")
	}

	// 
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 
	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete comment", zap.String("comment_id", commentID), zap.Error(err))
		return fmt.Errorf("")
	}

	// 
	if err := tx.Where("parent_id = ?", commentID).Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete replies", zap.String("comment_id", commentID), zap.Error(err))
		return fmt.Errorf("")
	}

	// 
	var replyCount int64
	tx.Model(&models.Comment{}).Where("parent_id = ?", commentID).Count(&replyCount)
	totalDeleted := replyCount + 1 // 

	if err := tx.Model(&models.Post{}).Where("id = ?", comment.PostID).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", totalDeleted)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update post comment count", zap.Error(err))
		return fmt.Errorf("")
	}

	tx.Commit()

	s.logger.Info("Comment deleted successfully", zap.String("comment_id", commentID), zap.String("user_id", userID))
	return nil
}

// GetCommentStats 
func (s *CommentService) GetCommentStats() (*models.CommentStatsResponse, error) {
	var stats models.CommentStatsResponse

	// 
	s.db.Model(&models.Comment{}).Where("status = ?", models.CommentStatusPublished).Count(&stats.TotalComments)

	// 
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Comment{}).Where("status = ? AND created_at >= ?", models.CommentStatusPublished, today).Count(&stats.TodayComments)

	// 
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.Comment{}).Where("status = ? AND created_at >= ?", models.CommentStatusPublished, weekStart).Count(&stats.WeeklyComments)

	// 
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	s.db.Model(&models.Comment{}).Where("status = ? AND created_at >= ?", models.CommentStatusPublished, monthStart).Count(&stats.MonthlyComments)

	// 
	s.db.Model(&models.Comment{}).Where("status = ? AND created_at >= ?", models.CommentStatusPublished, weekStart).Distinct("author_id").Count(&stats.ActiveUsers)

	return &stats, nil
}

// GetUserComments 
func (s *CommentService) GetUserComments(userID string, page, pageSize int) (*models.CommentListResponse, error) {
	var comments []models.Comment
	var total int64

	// 
	query := s.db.Model(&models.Comment{}).
		Preload("Author").
		Preload("Post").
		Where("author_id = ? AND status = ?", userID, models.CommentStatusPublished)

	// 
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count user comments", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		s.logger.Error("Failed to get user comments", zap.Error(err))
		return nil, fmt.Errorf("")
	}

	// 
	commentResponses := make([]models.CommentResponse, len(comments))
	for i, comment := range comments {
		commentResponses[i] = comment.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.CommentListResponse{
		Comments:   commentResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPostCommentCount 
func (s *CommentService) GetPostCommentCount(postID string) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Comment{}).Where("post_id = ? AND status = ?", postID, models.CommentStatusPublished).Count(&count).Error; err != nil {
		s.logger.Error("Failed to count post comments", zap.String("post_id", postID), zap.Error(err))
		return 0, fmt.Errorf("")
	}
	return count, nil
}

