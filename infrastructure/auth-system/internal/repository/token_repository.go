package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/taishanglaojun/auth_system/internal/models"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenUsed     = errors.New("token already used")
	ErrTokenRevoked  = errors.New("token revoked")
)

// TokenRepository 令牌仓储接口
type TokenRepository interface {
	// 基础CRUD
	Create(ctx context.Context, token *models.Token) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Token, error)
	GetByToken(ctx context.Context, token string) (*models.Token, error)
	Update(ctx context.Context, token *models.Token) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// 查询方法
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Token, error)
	GetByUserAndType(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) ([]*models.Token, error)
	List(ctx context.Context, query *models.TokenQuery) ([]*models.Token, int64, error)
	
	// 令牌管理
	UseToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) error
	ExpireToken(ctx context.Context, tokenID uuid.UUID) error
	
	// 验证方法
	ValidateToken(ctx context.Context, token string, tokenType models.TokenType) (*models.Token, error)
	IsTokenValid(ctx context.Context, tokenID uuid.UUID) (bool, error)
	
	// 清理方法
	CleanupExpiredTokens(ctx context.Context) (int64, error)
	CleanupUsedTokens(ctx context.Context, olderThan time.Duration) (int64, error)
	CleanupRevokedTokens(ctx context.Context, olderThan time.Duration) (int64, error)
	
	// 统计方法
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByType(ctx context.Context, tokenType models.TokenType) (int64, error)
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

// tokenRepository 令牌仓储实现
type tokenRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTokenRepository 创建令牌仓储
func NewTokenRepository(db *gorm.DB, logger *zap.Logger) TokenRepository {
	return &tokenRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建令牌
func (r *tokenRepository) Create(ctx context.Context, token *models.Token) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		r.logger.Error("Failed to create token", 
			zap.String("user_id", token.UserID.String()),
			zap.String("type", string(token.Type)),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("Token created successfully", 
		zap.String("token_id", token.ID.String()),
		zap.String("user_id", token.UserID.String()),
		zap.String("type", string(token.Type)),
	)
	
	return nil
}

// GetByID 根据ID获取令牌
func (r *tokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Token, error) {
	var token models.Token
	if err := r.db.WithContext(ctx).Preload("User").First(&token, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		r.logger.Error("Failed to get token by ID", 
			zap.String("token_id", id.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return &token, nil
}

// GetByToken 根据令牌字符串获取令牌
func (r *tokenRepository) GetByToken(ctx context.Context, token string) (*models.Token, error) {
	var tokenModel models.Token
	if err := r.db.WithContext(ctx).Preload("User").First(&tokenModel, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		r.logger.Error("Failed to get token by token string", zap.Error(err))
		return nil, err
	}
	
	return &tokenModel, nil
}

// Update 更新令牌
func (r *tokenRepository) Update(ctx context.Context, token *models.Token) error {
	if err := r.db.WithContext(ctx).Save(token).Error; err != nil {
		r.logger.Error("Failed to update token", 
			zap.String("token_id", token.ID.String()),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("Token updated successfully", 
		zap.String("token_id", token.ID.String()),
	)
	
	return nil
}

// Delete 删除令牌
func (r *tokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Token{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to delete token", 
			zap.String("token_id", id.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	
	r.logger.Info("Token deleted successfully", 
		zap.String("token_id", id.String()),
	)
	
	return nil
}

// GetByUserID 获取用户的所有令牌
func (r *tokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Token, error) {
	var tokens []*models.Token
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error; err != nil {
		r.logger.Error("Failed to get tokens by user ID", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return tokens, nil
}

// GetByUserAndType 获取用户指定类型的令牌
func (r *tokenRepository) GetByUserAndType(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) ([]*models.Token, error) {
	var tokens []*models.Token
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, tokenType).
		Order("created_at DESC").
		Find(&tokens).Error; err != nil {
		r.logger.Error("Failed to get tokens by user and type", 
			zap.String("user_id", userID.String()),
			zap.String("type", string(tokenType)),
			zap.Error(err),
		)
		return nil, err
	}
	
	return tokens, nil
}

// List 获取令牌列表
func (r *tokenRepository) List(ctx context.Context, query *models.TokenQuery) ([]*models.Token, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.Token{}).Preload("User")
	
	// 应用过滤条件
	if query.UserID != uuid.Nil {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Purpose != "" {
		db = db.Where("purpose ILIKE ?", "%"+query.Purpose+"%")
	}
	
	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count tokens", zap.Error(err))
		return nil, 0, err
	}
	
	// 应用排序
	orderBy := "created_at"
	if query.OrderBy != "" {
		orderBy = query.OrderBy
	}
	order := "desc"
	if query.Order != "" {
		order = query.Order
	}
	db = db.Order(fmt.Sprintf("%s %s", orderBy, order))
	
	// 应用分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}
	
	var tokens []*models.Token
	if err := db.Find(&tokens).Error; err != nil {
		r.logger.Error("Failed to list tokens", zap.Error(err))
		return nil, 0, err
	}
	
	return tokens, total, nil
}

// UseToken 使用令牌
func (r *tokenRepository) UseToken(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("id = ? AND status = ?", tokenID, models.TokenStatusActive).
		Updates(map[string]interface{}{
			"status":  models.TokenStatusUsed,
			"used_at": &now,
		})
	
	if result.Error != nil {
		r.logger.Error("Failed to use token", 
			zap.String("token_id", tokenID.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	
	r.logger.Info("Token used successfully", 
		zap.String("token_id", tokenID.String()),
	)
	
	return nil
}

// RevokeToken 撤销令牌
func (r *tokenRepository) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("status", models.TokenStatusRevoked)
	
	if result.Error != nil {
		r.logger.Error("Failed to revoke token", 
			zap.String("token_id", tokenID.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	
	r.logger.Info("Token revoked successfully", 
		zap.String("token_id", tokenID.String()),
	)
	
	return nil
}

// RevokeAllUserTokens 撤销用户的所有指定类型令牌
func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) error {
	query := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("user_id = ? AND status = ?", userID, models.TokenStatusActive)
	
	if tokenType != "" {
		query = query.Where("type = ?", tokenType)
	}
	
	result := query.Update("status", models.TokenStatusRevoked)
	
	if result.Error != nil {
		r.logger.Error("Failed to revoke all user tokens", 
			zap.String("user_id", userID.String()),
			zap.String("type", string(tokenType)),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	r.logger.Info("All user tokens revoked successfully", 
		zap.String("user_id", userID.String()),
		zap.String("type", string(tokenType)),
		zap.Int64("affected", result.RowsAffected),
	)
	
	return nil
}

// ExpireToken 使令牌过期
func (r *tokenRepository) ExpireToken(ctx context.Context, tokenID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("status", models.TokenStatusExpired)
	
	if result.Error != nil {
		r.logger.Error("Failed to expire token", 
			zap.String("token_id", tokenID.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	
	r.logger.Info("Token expired successfully", 
		zap.String("token_id", tokenID.String()),
	)
	
	return nil
}

// ValidateToken 验证令牌
func (r *tokenRepository) ValidateToken(ctx context.Context, token string, tokenType models.TokenType) (*models.Token, error) {
	tokenModel, err := r.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	
	// 检查令牌类型
	if tokenType != "" && tokenModel.Type != tokenType {
		return nil, ErrTokenNotFound
	}
	
	// 检查令牌状态
	if tokenModel.Status != models.TokenStatusActive {
		switch tokenModel.Status {
		case models.TokenStatusExpired:
			return nil, ErrTokenExpired
		case models.TokenStatusUsed:
			return nil, ErrTokenUsed
		case models.TokenStatusRevoked:
			return nil, ErrTokenRevoked
		default:
			return nil, ErrTokenNotFound
		}
	}
	
	// 检查是否过期
	if tokenModel.IsExpired() {
		// 自动标记为过期
		tokenModel.Status = models.TokenStatusExpired
		r.Update(ctx, tokenModel)
		return nil, ErrTokenExpired
	}
	
	return tokenModel, nil
}

// IsTokenValid 检查令牌是否有效
func (r *tokenRepository) IsTokenValid(ctx context.Context, tokenID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("id = ? AND status = ? AND expires_at > ?", tokenID, models.TokenStatusActive, time.Now()).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to check token validity", 
			zap.String("token_id", tokenID.String()),
			zap.Error(err),
		)
		return false, err
	}
	
	return count > 0, nil
}

// CleanupExpiredTokens 清理过期令牌
func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&models.Token{})
	
	if result.Error != nil {
		r.logger.Error("Failed to cleanup expired tokens", zap.Error(result.Error))
		return 0, result.Error
	}
	
	r.logger.Info("Expired tokens cleaned up successfully", 
		zap.Int64("deleted", result.RowsAffected),
	)
	
	return result.RowsAffected, nil
}

// CleanupUsedTokens 清理已使用的令牌
func (r *tokenRepository) CleanupUsedTokens(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	
	result := r.db.WithContext(ctx).
		Where("status = ? AND used_at <= ?", models.TokenStatusUsed, cutoffTime).
		Delete(&models.Token{})
	
	if result.Error != nil {
		r.logger.Error("Failed to cleanup used tokens", zap.Error(result.Error))
		return 0, result.Error
	}
	
	r.logger.Info("Used tokens cleaned up successfully", 
		zap.Int64("deleted", result.RowsAffected),
		zap.Duration("older_than", olderThan),
	)
	
	return result.RowsAffected, nil
}

// CleanupRevokedTokens 清理撤销的令牌
func (r *tokenRepository) CleanupRevokedTokens(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	
	result := r.db.WithContext(ctx).
		Where("status = ? AND updated_at <= ?", models.TokenStatusRevoked, cutoffTime).
		Delete(&models.Token{})
	
	if result.Error != nil {
		r.logger.Error("Failed to cleanup revoked tokens", zap.Error(result.Error))
		return 0, result.Error
	}
	
	r.logger.Info("Revoked tokens cleaned up successfully", 
		zap.Int64("deleted", result.RowsAffected),
		zap.Duration("older_than", olderThan),
	)
	
	return result.RowsAffected, nil
}

// Count 获取令牌总数
func (r *tokenRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Token{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count tokens", zap.Error(err))
		return 0, err
	}
	
	return count, nil
}

// CountByUser 获取用户令牌数
func (r *tokenRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count tokens by user", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}

// CountByType 根据类型统计令牌数
func (r *tokenRepository) CountByType(ctx context.Context, tokenType models.TokenType) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("type = ?", tokenType).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count tokens by type", 
			zap.String("type", string(tokenType)),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}

// CountActiveByUser 获取用户活跃令牌数
func (r *tokenRepository) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Token{}).
		Where("user_id = ? AND status = ? AND expires_at > ?", userID, models.TokenStatusActive, time.Now()).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count active tokens by user", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}