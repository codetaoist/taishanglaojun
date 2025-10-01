package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/jwt"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUserSuspended      = errors.New("user is suspended")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

// AuthService 认证服务接口
type AuthService interface {
	// 用户注册和登�?
	Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, req *models.LogoutRequest) (*models.LogoutResponse, error)
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	
	// 令牌管理
	RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error)
	ValidateToken(ctx context.Context, req *models.ValidateTokenRequest) (*models.ValidateTokenResponse, error)
	RevokeToken(ctx context.Context, tokenID uuid.UUID) error
	
	// 密码管理
	ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) (*models.ResetPasswordResponse, error)
	
	// 邮箱验证
	VerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) (*models.VerifyEmailResponse, error)
	ResendVerification(ctx context.Context, req *models.ResendVerificationRequest) (*models.ResendVerificationResponse, error)
	
	// 会话管理
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
	
	// 用户信息
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.PublicUser, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest) (*models.PublicUser, error)
}

// authService 认证服务实现
type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	tokenRepo   repository.TokenRepository
	jwtManager  *jwt.Manager
	logger      *zap.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	tokenRepo repository.TokenRepository,
	jwtManager *jwt.Manager,
	logger *zap.Logger,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokenRepo:   tokenRepo,
		jwtManager:  jwtManager,
		logger:      logger,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// 检查用户名是否已存�?
	if exists, err := s.userRepo.Exists(ctx, "username", req.Username); err != nil {
		s.logger.Error("Failed to check username existence", 
			zap.String("username", req.Username),
			zap.Error(err),
		)
		return nil, err
	} else if exists {
		return nil, ErrUserExists
	}
	
	// 检查邮箱是否已存在
	if exists, err := s.userRepo.Exists(ctx, "email", req.Email); err != nil {
		s.logger.Error("Failed to check email existence", 
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, err
	} else if exists {
		return nil, ErrUserExists
	}
	
	// 创建用户
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password, // 将在BeforeCreate钩子中哈�?
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    models.UserStatusInactive, // 需要邮箱验�?
		Role:      models.RoleUser,
	}
	
	// 哈希密码
	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", 
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 生成邮箱验证令牌
	verificationToken := &models.Token{
		UserID:    user.ID,
		Type:      models.TokenTypeVerification,
		Status:    models.TokenStatusActive,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时有效�?
	}
	
	if err := s.tokenRepo.Create(ctx, verificationToken); err != nil {
		s.logger.Error("Failed to create verification token", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		// 不返回错误，用户已创建成�?
	}
	
	s.logger.Info("User registered successfully", 
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)
	
	return &models.RegisterResponse{
		User:    user.ToPublic(),
		Token:   verificationToken.Token,
		Message: "Registration successful. Please verify your email.",
	}, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// 验证用户凭据
	user, err := s.userRepo.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, repository.ErrInvalidCredentials) {
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("Failed to authenticate user", 
			zap.String("username", req.Username),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 检查用户状�?
	if user.Status == models.UserStatusInactive {
		return nil, ErrUserInactive
	}
	if user.Status == models.UserStatusSuspended {
		return nil, ErrUserSuspended
	}
	
	// 创建会话
	session := &models.Session{
		UserID:    user.ID,
		Status:    models.SessionStatusActive,
		UserAgent: req.UserAgent,
		IPAddress: req.IPAddress,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30天有效期
	}
	
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("Failed to create session", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 生成JWT令牌�?
	accessToken, _, err := s.jwtManager.GenerateAccessToken(
		user.ID, 
		user.Username, 
		user.Email, 
		string(user.Role), 
		session.ID, 
		[]string{}, // TODO: 实现权限系统
	)
	if err != nil {
		s.logger.Error("Failed to generate access token", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID, session.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 更新最后登录时�?
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Warn("Failed to update last login time", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}
	
	s.logger.Info("User logged in successfully", 
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("session_id", session.ID.String()),
	)
	
	return &models.LoginResponse{
		User:         user.ToPublic(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetConfig().AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
		SessionID:    session.ID,
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, req *models.LogoutRequest) (*models.LogoutResponse, error) {
	// 撤销会话
	if err := s.sessionRepo.RevokeSession(ctx, req.SessionID); err != nil {
		if !errors.Is(err, repository.ErrSessionNotFound) {
			s.logger.Error("Failed to revoke session", 
				zap.String("session_id", req.SessionID.String()),
				zap.Error(err),
			)
			return nil, err
		}
	}
	
	// 撤销相关令牌
	if req.RefreshToken != "" {
		// 这里可以实现令牌黑名单机�?
		s.logger.Info("Refresh token should be blacklisted", 
			zap.String("session_id", req.SessionID.String()),
		)
	}
	
	s.logger.Info("User logged out successfully", 
		zap.String("session_id", req.SessionID.String()),
	)
	
	return &models.LogoutResponse{
		Message: "Logout successful",
	}, nil
}

// LogoutAll 登出所有会�?
func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		s.logger.Error("Failed to revoke all user sessions", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}
	
	s.logger.Info("All user sessions revoked successfully", 
		zap.String("user_id", userID.String()),
	)
	
	return nil
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	
	// 检查令牌类�?
	if claims.TokenType != "refresh" {
		return nil, ErrInvalidToken
	}
	
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	// 检查用户状�?
	if user.Status != models.UserStatusActive {
		return nil, ErrUserInactive
	}
	
	// 验证会话
	session, err := s.sessionRepo.GetByID(ctx, claims.SessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	
	if !session.IsActive() {
		return nil, ErrSessionExpired
	}
	
	// 生成新的访问令牌
	accessToken, _, err := s.jwtManager.GenerateAccessToken(
		user.ID, 
		user.Username, 
		user.Email, 
		string(user.Role), 
		session.ID, 
		[]string{}, // TODO: 实现权限系统
	)
	if err != nil {
		s.logger.Error("Failed to generate new access token", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 如果刷新令牌即将过期，生成新的刷新令�?
	var newRefreshToken string
	refreshClaims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err == nil && s.jwtManager.IsTokenExpiringSoon(refreshClaims) {
		newRefreshToken, _, err = s.jwtManager.GenerateRefreshToken(user.ID, session.ID)
		if err != nil {
			s.logger.Error("Failed to generate new refresh token", 
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
			return nil, err
		}
	}
	
	// 更新会话活动时间
	if err := s.sessionRepo.RefreshSession(ctx, session.ID, 24*time.Hour); err != nil {
		s.logger.Warn("Failed to refresh session", 
			zap.String("session_id", session.ID.String()),
			zap.Error(err),
		)
	}
	
	s.logger.Info("Token refreshed successfully", 
		zap.String("user_id", user.ID.String()),
		zap.String("session_id", session.ID.String()),
	)
	
	response := &models.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.jwtManager.GetConfig().AccessTokenTTL.Seconds()),
	}
	
	if newRefreshToken != "" {
		response.RefreshToken = newRefreshToken
	}
	
	return response, nil
}

// ValidateToken 验证令牌
func (s *authService) ValidateToken(ctx context.Context, req *models.ValidateTokenRequest) (*models.ValidateTokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(req.Token)
	if err != nil {
		return &models.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid token",
		}, nil
	}
	
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return &models.ValidateTokenResponse{
			Valid:   false,
			Message: "User not found",
		}, nil
	}
	
	// 检查用户状�?
	if user.Status != models.UserStatusActive {
		return &models.ValidateTokenResponse{
			Valid:   false,
			Message: "User inactive",
		}, nil
	}
	
	// 验证会话（如果是访问令牌�?
	if claims.TokenType == "access" {
		session, err := s.sessionRepo.GetByID(ctx, claims.SessionID)
		if err != nil || !session.IsActive() {
			return &models.ValidateTokenResponse{
				Valid:   false,
				Message: "Session expired",
			}, nil
		}
	}
	
	return &models.ValidateTokenResponse{
		Valid:   true,
		Claims:  claims,
		Message: "Token is valid",
	}, nil
}

// RevokeToken 撤销令牌
func (s *authService) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	if err := s.tokenRepo.RevokeToken(ctx, tokenID); err != nil {
		s.logger.Error("Failed to revoke token", 
			zap.String("token_id", tokenID.String()),
			zap.Error(err),
		)
		return err
	}
	
	s.logger.Info("Token revoked successfully", 
		zap.String("token_id", tokenID.String()),
	)
	
	return nil
}

// ChangePassword 修改密码
func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest) error {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	
	// 验证当前密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return ErrInvalidCredentials
	}
	
	// 更新密码
	if err := s.userRepo.ChangePassword(ctx, userID, req.NewPassword); err != nil {
		s.logger.Error("Failed to change password", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}
	
	// 撤销所有会话（强制重新登录�?
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		s.logger.Warn("Failed to revoke sessions after password change", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
	}
	
	s.logger.Info("Password changed successfully", 
		zap.String("user_id", userID.String()),
	)
	
	return nil
}

// ForgotPassword 忘记密码
func (s *authService) ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// 为了安全，不透露用户是否存在
			return &models.ForgotPasswordResponse{
				Message: "If the email exists, a reset link has been sent.",
			}, nil
		}
		return nil, err
	}
	
	// 撤销之前的重置令�?
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, user.ID, models.TokenTypeReset); err != nil {
		s.logger.Warn("Failed to revoke previous reset tokens", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}
	
	// 生成密码重置令牌
	resetToken := &models.Token{
		UserID:    user.ID,
		Type:      models.TokenTypeReset,
		Status:    models.TokenStatusActive,
		Purpose:   "password_reset",
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1小时有效�?
	}
	
	if err := s.tokenRepo.Create(ctx, resetToken); err != nil {
		s.logger.Error("Failed to create reset token", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	s.logger.Info("Password reset token created", 
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)
	
	return &models.ForgotPasswordResponse{
		Token: resetToken.Token,
		Message:    "Password reset token generated successfully.",
	}, nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	// 验证重置令牌
	token, err := s.tokenRepo.ValidateToken(ctx, req.Token, models.TokenTypeReset)
	if err != nil {
		if errors.Is(err, repository.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	
	// 使用令牌
	if err := s.tokenRepo.UseToken(ctx, token.ID); err != nil {
		s.logger.Error("Failed to use reset token", 
			zap.String("token_id", token.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 更新密码
	if err := s.userRepo.ChangePassword(ctx, token.UserID, req.Password); err != nil {
		s.logger.Error("Failed to reset password", 
			zap.String("user_id", token.UserID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 撤销所有会�?
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, token.UserID); err != nil {
		s.logger.Warn("Failed to revoke sessions after password reset", 
			zap.String("user_id", token.UserID.String()),
			zap.Error(err),
		)
	}
	
	s.logger.Info("Password reset successfully", 
		zap.String("user_id", token.UserID.String()),
	)
	
	return &models.ResetPasswordResponse{
		Message: "Password reset successfully.",
	}, nil
}

// VerifyEmail 验证邮箱
func (s *authService) VerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) (*models.VerifyEmailResponse, error) {
	// 验证邮箱验证令牌
	token, err := s.tokenRepo.ValidateToken(ctx, req.Token, models.TokenTypeVerification)
	if err != nil {
		if errors.Is(err, repository.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	
	// 使用令牌
	if err := s.tokenRepo.UseToken(ctx, token.ID); err != nil {
		s.logger.Error("Failed to use verification token", 
			zap.String("token_id", token.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	// 激活用�?
	if err := s.userRepo.UpdateStatus(ctx, token.UserID, models.UserStatusActive); err != nil {
		s.logger.Error("Failed to activate user", 
			zap.String("user_id", token.UserID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	s.logger.Info("Email verified successfully", 
		zap.String("user_id", token.UserID.String()),
	)
	
	return &models.VerifyEmailResponse{
		Message: "Email verified successfully.",
	}, nil
}

// ResendVerification 重新发送验证邮�?
func (s *authService) ResendVerification(ctx context.Context, req *models.ResendVerificationRequest) (*models.ResendVerificationResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &models.ResendVerificationResponse{
				Message: "If the email exists, a verification link has been sent.",
			}, nil
		}
		return nil, err
	}
	
	// 检查用户是否已激�?
	if user.Status == models.UserStatusActive {
		return &models.ResendVerificationResponse{
			Message: "Email is already verified.",
		}, nil
	}
	
	// 撤销之前的验证令�?
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, user.ID, models.TokenTypeVerification); err != nil {
		s.logger.Warn("Failed to revoke previous verification tokens", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}
	
	// 生成新的验证令牌
	verificationToken := &models.Token{
		UserID:    user.ID,
		Type:      models.TokenTypeVerification,
		Status:    models.TokenStatusActive,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时有效�?
	}
	
	if err := s.tokenRepo.Create(ctx, verificationToken); err != nil {
		s.logger.Error("Failed to create verification token", 
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	s.logger.Info("Verification token resent", 
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)
	
	return &models.ResendVerificationResponse{
		Token: verificationToken.Token,
		Message:           "Verification email sent successfully.",
	}, nil
}

// GetUserSessions 获取用户会话
func (s *authService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user sessions", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	return sessions, nil
}

// RevokeSession 撤销会话
func (s *authService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.RevokeSession(ctx, sessionID); err != nil {
		s.logger.Error("Failed to revoke session", 
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		return err
	}
	
	s.logger.Info("Session revoked successfully", 
		zap.String("session_id", sessionID.String()),
	)
	
	return nil
}

// RevokeAllSessions 撤销所有会�?
func (s *authService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		s.logger.Error("Failed to revoke all sessions", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}
	
	s.logger.Info("All sessions revoked successfully", 
		zap.String("user_id", userID.String()),
	)
	
	return nil
}

// GetUserProfile 获取用户资料
func (s *authService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.PublicUser, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return user.ToPublic(), nil
}

// UpdateUserProfile 更新用户资料
func (s *authService) UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest) (*models.PublicUser, error) {
	// 获取当前用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// 更新其他字段
	if req.FirstName != nil && *req.FirstName != "" {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil && *req.LastName != "" {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	
	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user profile", 
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	
	s.logger.Info("User profile updated successfully", 
		zap.String("user_id", userID.String()),
	)
	
	return user.ToPublic(), nil
}
