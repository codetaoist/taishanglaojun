package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// FavoritesService 收藏服务
type FavoritesService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewFavoritesService 创建收藏服务
func NewFavoritesService(db *gorm.DB, logger *zap.Logger) *FavoritesService {
	return &FavoritesService{
		db:     db,
		logger: logger,
	}
}

// AddFavorite 添加收藏
func (s *FavoritesService) AddFavorite(ctx context.Context, userID, wisdomID string) (*models.FavoriteResponse, error) {
	s.logger.Info("Adding favorite", zap.String("user_id", userID), zap.String("wisdom_id", wisdomID))

	// ?
	var existingFavorite models.WisdomFavorite
	err := s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, wisdomID).First(&existingFavorite).Error
	if err == nil {
		return nil, errors.New("")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Failed to check existing favorite", zap.Error(err))
		return nil, fmt.Errorf("? %w", err)
	}

	//
	favorite := &models.WisdomFavorite{
		UserID:   userID,
		WisdomID: wisdomID,
	}

	if err := s.db.WithContext(ctx).Create(favorite).Error; err != nil {
		s.logger.Error("Failed to create favorite", zap.Error(err))
		return nil, fmt.Errorf("查找笔记失败: %w", err)
	}

	response := &models.FavoriteResponse{
		ID:        favorite.ID,
		UserID:    favorite.UserID,
		WisdomID:  favorite.WisdomID,
		CreatedAt: favorite.CreatedAt,
	}

	s.logger.Info("Favorite added successfully", zap.Uint("favorite_id", favorite.ID))
	return response, nil
}

// RemoveFavorite 删除收藏
func (s *FavoritesService) RemoveFavorite(ctx context.Context, userID, wisdomID string) error {
	s.logger.Info("Removing favorite", zap.String("user_id", userID), zap.String("wisdom_id", wisdomID))

	result := s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, wisdomID).Delete(&models.WisdomFavorite{})
	if result.Error != nil {
		s.logger.Error("Failed to remove favorite", zap.Error(result.Error))
		return fmt.Errorf("删除收藏失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("收藏记录不存在")
	}

	s.logger.Info("Favorite removed successfully")
	return nil
}

// GetUserFavorites 获取用户收藏
func (s *FavoritesService) GetUserFavorites(ctx context.Context, userID string, page, pageSize int) ([]*models.FavoriteResponse, int64, error) {
	s.logger.Info("Getting user favorites", zap.String("user_id", userID), zap.Int("page", page), zap.Int("page_size", pageSize))

	var favorites []models.WisdomFavorite
	var total int64

	// 统计用户收藏总数
	if err := s.db.WithContext(ctx).Model(&models.WisdomFavorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count favorites", zap.Error(err))
		return nil, 0, fmt.Errorf("获取收藏总数失败: %w", err)
	}

	// 分页查询用户收藏
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&favorites).Error; err != nil {
		s.logger.Error("Failed to get favorites", zap.Error(err))
		return nil, 0, fmt.Errorf(": %w", err)
	}

	// 转换为响应模型
	responses := make([]*models.FavoriteResponse, len(favorites))
	for i, favorite := range favorites {
		responses[i] = &models.FavoriteResponse{
			ID:        favorite.ID,
			UserID:    favorite.UserID,
			WisdomID:  favorite.WisdomID,
			CreatedAt: favorite.CreatedAt,
		}
	}

	s.logger.Info("User favorites retrieved successfully", zap.Int("count", len(responses)), zap.Int64("total", total))
	return responses, total, nil
}

// IsFavorited 检查是否收藏
func (s *FavoritesService) IsFavorited(ctx context.Context, userID, wisdomID string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.WisdomFavorite{}).
		Where("user_id = ? AND wisdom_id = ?", userID, wisdomID).
		Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to check favorite status", zap.Error(err))
		return false, fmt.Errorf("? %w", err)
	}
	return count > 0, nil
}

// CreateNote 创建笔记
func (s *FavoritesService) CreateNote(ctx context.Context, userID string, req *models.NoteRequest) (*models.NoteResponse, error) {
	s.logger.Info("Creating note", zap.String("user_id", userID), zap.String("wisdom_id", req.WisdomID))

	// 检查是否已存在笔记
	var existingNote models.WisdomNote
	err := s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, req.WisdomID).First(&existingNote).Error
	if err == nil {
		return nil, errors.New("")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Failed to check existing note", zap.Error(err))
		return nil, fmt.Errorf("? %w", err)
	}

	// 检查是否已收藏
	var existingFavorite models.WisdomFavorite
	err = s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, req.WisdomID).First(&existingFavorite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("未收藏该智慧")
		}
		s.logger.Error("Failed to check existing favorite", zap.Error(err))
		return nil, fmt.Errorf("? %w", err)
	}

	// 创建笔记
	tagsJSON, _ := json.Marshal(req.Tags)
	note := &models.WisdomNote{
		UserID:    userID,
		WisdomID:  req.WisdomID,
		Title:     req.Title,
		Content:   req.Content,
		IsPrivate: req.IsPrivate,
		Tags:      string(tagsJSON),
	}

	if err := s.db.WithContext(ctx).Create(note).Error; err != nil {
		s.logger.Error("Failed to create note", zap.Error(err))
		return nil, fmt.Errorf("创建笔记失败: %w", err)
	}

	// 解析标签
	var tags []string
	json.Unmarshal([]byte(note.Tags), &tags)

	response := &models.NoteResponse{
		ID:        note.ID,
		UserID:    note.UserID,
		WisdomID:  note.WisdomID,
		Title:     note.Title,
		Content:   note.Content,
		IsPrivate: note.IsPrivate,
		Tags:      tags,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}

	s.logger.Info("Note created successfully", zap.Uint("note_id", note.ID))
	return response, nil
}

// UpdateNote 更新笔记
func (s *FavoritesService) UpdateNote(ctx context.Context, userID string, wisdomID string, req *models.NoteUpdateRequest) (*models.NoteResponse, error) {
	s.logger.Info("Updating note", zap.String("user_id", userID), zap.String("wisdom_id", wisdomID))

	var note models.WisdomNote
	if err := s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, wisdomID).First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		s.logger.Error("Failed to find note", zap.Error(err))
		return nil, fmt.Errorf(": %w", err)
	}

	// 更新笔记字段
	tagsJSON, _ := json.Marshal(req.Tags)
	note.Title = req.Title
	note.Content = req.Content
	note.IsPrivate = req.IsPrivate
	note.Tags = string(tagsJSON)

	if err := s.db.WithContext(ctx).Save(&note).Error; err != nil {
		s.logger.Error("Failed to update note", zap.Error(err))
		return nil, fmt.Errorf("更新笔记失败: %w", err)
	}

	// 解析标签
	var tags []string
	json.Unmarshal([]byte(note.Tags), &tags)

	response := &models.NoteResponse{
		ID:        note.ID,
		UserID:    note.UserID,
		WisdomID:  note.WisdomID,
		Title:     note.Title,
		Content:   note.Content,
		IsPrivate: note.IsPrivate,
		Tags:      tags,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}

	s.logger.Info("Note updated successfully", zap.Uint("note_id", note.ID))
	return response, nil
}

// GetNote 获取笔记
func (s *FavoritesService) GetNote(ctx context.Context, userID, wisdomID string) (*models.NoteResponse, error) {
	s.logger.Info("Getting note", zap.String("user_id", userID), zap.String("wisdom_id", wisdomID))

	var note models.WisdomNote
	if err := s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, wisdomID).First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		s.logger.Error("Failed to get note", zap.Error(err))
		return nil, fmt.Errorf(": %w", err)
	}

	var tags []string
	json.Unmarshal([]byte(note.Tags), &tags)

	response := &models.NoteResponse{
		ID:        note.ID,
		UserID:    note.UserID,
		WisdomID:  note.WisdomID,
		Title:     note.Title,
		Content:   note.Content,
		IsPrivate: note.IsPrivate,
		Tags:      tags,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}

	s.logger.Info("Note retrieved successfully", zap.Uint("note_id", note.ID))
	return response, nil
}

// GetUserNotes 获取用户所有笔记
func (s *FavoritesService) GetUserNotes(ctx context.Context, userID string, page, pageSize int) ([]*models.NoteResponse, int64, error) {
	s.logger.Info("Getting user notes", zap.String("user_id", userID), zap.Int("page", page), zap.Int("page_size", pageSize))

	var notes []models.WisdomNote
	var total int64

	// 统计用户笔记总数
	if err := s.db.WithContext(ctx).Model(&models.WisdomNote{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count notes", zap.Error(err))
		return nil, 0, fmt.Errorf("统计用户笔记总数失败: %w", err)
	}

	// 获取分页笔记
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notes).Error; err != nil {
		s.logger.Error("Failed to get notes", zap.Error(err))
		return nil, 0, fmt.Errorf("获取用户笔记失败: %w", err)
	}

	// 解析标签并创建响应
	responses := make([]*models.NoteResponse, len(notes))
	for i, note := range notes {
		var tags []string
		json.Unmarshal([]byte(note.Tags), &tags)

		responses[i] = &models.NoteResponse{
			ID:        note.ID,
			UserID:    note.UserID,
			WisdomID:  note.WisdomID,
			Title:     note.Title,
			Content:   note.Content,
			IsPrivate: note.IsPrivate,
			Tags:      tags,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		}
	}

	s.logger.Info("User notes retrieved successfully", zap.Int("count", len(responses)), zap.Int64("total", total))
	return responses, total, nil
}

// DeleteNote 删除笔记
func (s *FavoritesService) DeleteNote(ctx context.Context, userID, wisdomID string) error {
	s.logger.Info("Deleting note", zap.String("user_id", userID), zap.String("wisdom_id", wisdomID))

	result := s.db.WithContext(ctx).Where("user_id = ? AND wisdom_id = ?", userID, wisdomID).Delete(&models.WisdomNote{})
	if result.Error != nil {
		s.logger.Error("Failed to delete note", zap.Error(result.Error))
		return fmt.Errorf("删除笔记失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("笔记不存在")
	}

	s.logger.Info("Note deleted successfully")
	return nil
}
