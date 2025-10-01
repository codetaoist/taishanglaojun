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
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionRevoked  = errors.New("session revoked")
)

// SessionRepository 会话仓储接口
type SessionRepository interface {
	// 基础CRUD
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// 查询方法
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	List(ctx context.Context, query *models.SessionQuery) ([]*models.Session, int64, error)
	
	// 会话管理
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	RevokeExpiredSessions(ctx context.Context) (int64, error)
	RefreshSession(ctx context.Context, sessionID uuid.UUID, duration time.Duration) error
	
	// 验证方法
	ValidateSession(ctx context.Context, token string) (*models.Session, error)
	IsSessionActive(ctx context.Context, sessionID uuid.UUID) (bool, error)
	
	// 清理方法
	CleanupExpiredSessions(ctx context.Context) (int64, error)
	CleanupRevokedSessions(ctx context.Context, olderThan time.Duration) (int64, error)
	
	// 统计方法
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

// sessionRepository 会话仓储实现
type sessionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSessionRepository 创建会话仓储
func NewSessionRepository(db *gorm.DB, logger *zap.Logger) SessionRepository {
	return &sessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建会话
func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		r.logger.Error("Failed to create session", 
			zap.String("user_id", session.UserID.String()),
			zap.String("ip_address", session.IPAddress),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("Session created successfully", 
		zap.String("session_id", session.ID.String()),
		zap.String("user_id", session.UserID.String()),
	)
	
	return nil
}

// GetByID 根据ID获取会话
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Preload("User").First(&session, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		r.logger.Error("Failed to get session by ID", 
			zap.String("session_id", id.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return &session, nil
}

// GetByToken 根据令牌获取会话
func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Preload("User").First(&session, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		r.logger.Error("Failed to get session by token", zap.Error(err))
		return nil, err
	}
	
	return &session, nil
}

// Update 更新会话
func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		r.logger.Error("Failed to update session", 
			zap.String("session_id", session.ID.String()),
			zap.Error(err),
		)
		return err
	}
	
	r.logger.Info("Session updated successfully", 
		zap.String("session_id", session.ID.String()),
	)
	
	return nil
}

// Delete 删除会话
func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to delete session", 
			zap.String("session_id", id.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	
	r.logger.Info("Session deleted successfully", 
		zap.String("session_id", id.String()),
	)
	
	return nil
}

// GetByUserID 获取用户的所有会话
func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		r.logger.Error("Failed to get sessions by user ID", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return sessions, nil
}

// GetActiveSessions 获取用户的活跃会话
func (r *sessionRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND expires_at > ?", userID, models.SessionStatusActive, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		r.logger.Error("Failed to get active sessions", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return sessions, nil
}

// List 获取会话列表
func (r *sessionRepository) List(ctx context.Context, query *models.SessionQuery) ([]*models.Session, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.Session{}).Preload("User")
	
	// 应用过滤条件
	if query.UserID != uuid.Nil {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.IPAddress != "" {
		db = db.Where("ip_address = ?", query.IPAddress)
	}
	
	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count sessions", zap.Error(err))
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
	
	var sessions []*models.Session
	if err := db.Find(&sessions).Error; err != nil {
		r.logger.Error("Failed to list sessions", zap.Error(err))
		return nil, 0, err
	}
	
	return sessions, total, nil
}

// RevokeSession 撤销会话
func (r *sessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("status", models.SessionStatusRevoked)
	
	if result.Error != nil {
		r.logger.Error("Failed to revoke session", 
			zap.String("session_id", sessionID.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	
	r.logger.Info("Session revoked successfully", 
		zap.String("session_id", sessionID.String()),
	)
	
	return nil
}

// RevokeAllUserSessions 撤销用户的所有会话
func (r *sessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND status = ?", userID, models.SessionStatusActive).
		Update("status", models.SessionStatusRevoked)
	
	if result.Error != nil {
		r.logger.Error("Failed to revoke all user sessions", 
			zap.String("user_id", userID.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	r.logger.Info("All user sessions revoked successfully", 
		zap.String("user_id", userID.String()),
		zap.Int64("affected", result.RowsAffected),
	)
	
	return nil
}

// RevokeExpiredSessions 撤销过期会话
func (r *sessionRepository) RevokeExpiredSessions(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("status = ? AND expires_at <= ?", models.SessionStatusActive, time.Now()).
		Update("status", models.SessionStatusExpired)
	
	if result.Error != nil {
		r.logger.Error("Failed to revoke expired sessions", zap.Error(result.Error))
		return 0, result.Error
	}
	
	r.logger.Info("Expired sessions revoked successfully", 
		zap.Int64("affected", result.RowsAffected),
	)
	
	return result.RowsAffected, nil
}

// RefreshSession 刷新会话
func (r *sessionRepository) RefreshSession(ctx context.Context, sessionID uuid.UUID, duration time.Duration) error {
	newExpiresAt := time.Now().Add(duration)
	
	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ? AND status = ?", sessionID, models.SessionStatusActive).
		Updates(map[string]interface{}{
			"expires_at": newExpiresAt,
			"status":     models.SessionStatusActive,
		})
	
	if result.Error != nil {
		r.logger.Error("Failed to refresh session", 
			zap.String("session_id", sessionID.String()),
			zap.Error(result.Error),
		)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	
	r.logger.Info("Session refreshed successfully", 
		zap.String("session_id", sessionID.String()),
		zap.Time("new_expires_at", newExpiresAt),
	)
	
	return nil
}

// ValidateSession 验证会话
func (r *sessionRepository) ValidateSession(ctx context.Context, token string) (*models.Session, error) {
	session, err := r.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	
	// 检查会话状态
	if session.Status != models.SessionStatusActive {
		if session.Status == models.SessionStatusExpired {
			return nil, ErrSessionExpired
		}
		return nil, ErrSessionRevoked
	}
	
	// 检查是否过期
	if session.IsExpired() {
		// 自动标记为过期
		session.Status = models.SessionStatusExpired
		r.Update(ctx, session)
		return nil, ErrSessionExpired
	}
	
	return session, nil
}

// IsSessionActive 检查会话是否活跃
func (r *sessionRepository) IsSessionActive(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ? AND status = ? AND expires_at > ?", sessionID, models.SessionStatusActive, time.Now()).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to check session active status", 
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		return false, err
	}
	
	return count > 0, nil
}

// CleanupExpiredSessions 清理过期会话
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&models.Session{})
	
	if result.Error != nil {
		r.logger.Error("Failed to cleanup expired sessions", zap.Error(result.Error))
		return 0, result.Error
	}
	
	r.logger.Info("Expired sessions cleaned up successfully", 
		zap.Int64("deleted", result.RowsAffected),
	)
	
	return result.RowsAffected, nil
}

// CleanupRevokedSessions 清理撤销的会话
func (r *sessionRepository) CleanupRevokedSessions(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	
	result := r.db.WithContext(ctx).
		Where("status = ? AND updated_at <= ?", models.SessionStatusRevoked, cutoffTime).
		Delete(&models.Session{})
	
	if result.Error != nil {
		r.logger.Error("Failed to cleanup revoked sessions", zap.Error(result.Error))
		return 0, result.Error
	}
	
	r.logger.Info("Revoked sessions cleaned up successfully", 
		zap.Int64("deleted", result.RowsAffected),
		zap.Duration("older_than", olderThan),
	)
	
	return result.RowsAffected, nil
}

// Count 获取会话总数
func (r *sessionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count sessions", zap.Error(err))
		return 0, err
	}
	
	return count, nil
}

// CountByUser 获取用户会话数
func (r *sessionRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count sessions by user", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}

// CountActiveByUser 获取用户活跃会话数
func (r *sessionRepository) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND status = ? AND expires_at > ?", userID, models.SessionStatusActive, time.Now()).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count active sessions by user", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return 0, err
	}
	
	return count, nil
}