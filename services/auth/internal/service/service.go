package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/codetaoist/taishanglaojun/auth/internal/model"
	"github.com/codetaoist/taishanglaojun/auth/internal/repository"
)

// AuthService interface defines authentication service operations
type AuthService interface {
	Login(req *model.LoginRequest) (*model.LoginResponse, error)
	Register(req *model.RegisterRequest) error
	RefreshToken(req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
	Logout(token string) error
	ChangePassword(userID int, req *model.ChangePasswordRequest) error
	GetUser(userID int) (*model.User, error)
	ValidateToken(token string) (*model.User, error)
	RevokeToken(token string, reason string) error
}

// authService implements AuthService
type authService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	blacklistRepo repository.BlacklistRepository
	jwtSecret    string
	jwtExp       int
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	blacklistRepo repository.BlacklistRepository,
	jwtSecret string,
	jwtExp int,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		blacklistRepo: blacklistRepo,
		jwtSecret:    jwtSecret,
		jwtExp:       jwtExp,
	}
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Check if user is active (since lao_users doesn't have status field, we assume all users are active)
	// if user.Status != "active" {
	// 	return nil, fmt.Errorf("account is not active")
	// }

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Generate JWT token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	session := &model.Session{
		UserID:       user.ID,
		RefreshToken: token, // Use token as refresh token since table has refresh_token column
		ExpiresAt:    expiresAt,
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Remove password from response
	user.Password = ""

	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

// Register creates a new user
func (s *authService) Register(req *model.RegisterRequest) error {
	// Check if username already exists
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		Status:   req.Role, // Use role as status since lao_users doesn't have status field
	}

	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// RefreshToken refreshes a JWT token
func (s *authService) RefreshToken(req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	// Validate token
	user, err := s.ValidateToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Generate new token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update session
	session, err := s.sessionRepo.GetByRefreshToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	session.RefreshToken = token
	session.ExpiresAt = expiresAt
	if err := s.sessionRepo.Update(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &model.RefreshTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// Logout invalidates a JWT token
func (s *authService) Logout(token string) error {
	// Validate token first
	_, err := s.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Revoke the token
	return s.RevokeToken(token, "logout")
}

// ChangePassword changes a user's password
func (s *authService) ChangePassword(userID int, req *model.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetUser returns a user by ID
func (s *authService) GetUser(userID int) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Remove password from response
	user.Password = ""

	return user, nil
}

// ValidateToken validates a JWT token and returns the associated user
func (s *authService) ValidateToken(token string) (*model.User, error) {
	// Parse and validate token
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is blacklisted
	tokenHash := hashToken(token)
	_, err = s.blacklistRepo.GetByTokenHash(tokenHash)
	if err == nil {
		return nil, fmt.Errorf("token is blacklisted")
	}

	// Check if session exists
	session, err := s.sessionRepo.GetByRefreshToken(token)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Get user
	userID := 0
	sub := claims.Subject
	if sub != "" {
		_, err := fmt.Sscanf(sub, "%d", &userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID in token: %w", err)
		}
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Remove password from response
	user.Password = ""

	return user, nil
}

// generateToken generates a JWT token for a user
func (s *authService) generateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.jwtExp) * time.Second)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", user.ID),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	token, err := claims.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return token, expiresAt, nil
}

// hashToken hashes a token for storage
func hashToken(token string) string {
	// In a real implementation, you would use a proper hash function like SHA-256
	// For simplicity, we'll use base64 encoding
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// RevokeToken revokes a JWT token by adding it to the blacklist
func (s *authService) RevokeToken(token string, reason string) error {
	// Get user from token
	user, err := s.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Parse token to get expiration time
	tokenObj, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	// Get claims
	claims, ok := tokenObj.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Get expiration time from claims
	expiresAtFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("invalid expiration time in token")
	}

	expiresAt := time.Unix(int64(expiresAtFloat), 0)

	// Create blacklist entry
	blacklist := &model.TokenBlacklist{
		TokenHash: hashToken(token),
		UserID:    user.ID,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Add to blacklist
	if err := s.blacklistRepo.Create(blacklist); err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	// Also delete the session if it exists
	session, err := s.sessionRepo.GetByRefreshToken(token)
	if err == nil {
		s.sessionRepo.Delete(session.ID)
	}

	return nil
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}