package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/repositories"
)

// OAuthService OAuth
type OAuthService struct {
	repo       *repositories.OAuthRepository
	httpClient *http.Client
}

// NewOAuthService OAuth
func NewOAuthService(repo *repositories.OAuthRepository) *OAuthService {
	return &OAuthService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateOAuthApp OAuth
func (s *OAuthService) CreateOAuthApp(userID int64, name, provider, clientID, clientSecret, redirectURI string, scopes []string) (*models.OAuth, error) {
	oauth := &models.OAuth{
		UserID:       userID,
		Name:         name,
		Provider:     provider,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
		Status:       models.OAuthStatusActive,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// OAuth
	if err := s.validateOAuthConfig(oauth); err != nil {
		return nil, fmt.Errorf("invalid OAuth configuration: %w", err)
	}

	// 浽
	id, err := s.repo.Create(oauth)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth app: %w", err)
	}

	oauth.ID = id
	return oauth, nil
}

// GetOAuthApp OAuth
func (s *OAuthService) GetOAuthApp(id int64) (*models.OAuth, error) {
	return s.repo.GetByID(id)
}

// ListOAuthApps OAuth
func (s *OAuthService) ListOAuthApps(userID int64, provider string, limit, offset int) ([]*models.OAuth, int64, error) {
	return s.repo.ListByUserID(userID, provider, limit, offset)
}

// UpdateOAuthApp OAuth
func (s *OAuthService) UpdateOAuthApp(id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.repo.Update(id, updates)
}

// DeleteOAuthApp OAuth
func (s *OAuthService) DeleteOAuthApp(id int64) error {
	// 
	if err := s.RevokeAllTokens(id); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	return s.repo.Delete(id)
}

// GetAuthorizationURL URL
func (s *OAuthService) GetAuthorizationURL(id int64, state string) (string, error) {
	oauth, err := s.repo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("OAuth app not found: %w", err)
	}

	if !oauth.IsActive {
		return "", fmt.Errorf("OAuth app is not active")
	}

	// state
	if state == "" {
		state = s.generateState()
	}

	// URL
	authURL, err := s.buildAuthorizationURL(oauth, state)
	if err != nil {
		return "", fmt.Errorf("failed to build authorization URL: %w", err)
	}

	// state
	s.repo.Update(id, map[string]interface{}{
		"state":      state,
		"updated_at": time.Now(),
	})

	return authURL, nil
}

// ExchangeCodeForToken 
func (s *OAuthService) ExchangeCodeForToken(id int64, code, state string) (*TokenResponse, error) {
	oauth, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("OAuth app not found: %w", err)
	}

	// state
	if oauth.State != state {
		return nil, fmt.Errorf("invalid state parameter")
	}

	// 
	tokenResp, err := s.exchangeToken(oauth, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// 
	updates := map[string]interface{}{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"token_type":    tokenResp.TokenType,
		"expires_at":    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		"status":        models.OAuthStatusAuthorized,
		"updated_at":    time.Now(),
	}

	if err := s.repo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return tokenResp, nil
}

// RefreshToken 
func (s *OAuthService) RefreshToken(id int64) (*TokenResponse, error) {
	oauth, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("OAuth app not found: %w", err)
	}

	if oauth.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// 
	tokenResp, err := s.refreshAccessToken(oauth)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// 
	updates := map[string]interface{}{
		"access_token": tokenResp.AccessToken,
		"token_type":   tokenResp.TokenType,
		"expires_at":   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		"updated_at":   time.Now(),
	}

	if tokenResp.RefreshToken != "" {
		updates["refresh_token"] = tokenResp.RefreshToken
	}

	if err := s.repo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return tokenResp, nil
}

// RevokeToken 
func (s *OAuthService) RevokeToken(id int64) error {
	oauth, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("OAuth app not found: %w", err)
	}

	if oauth.AccessToken == "" {
		return fmt.Errorf("no access token to revoke")
	}

	// 
	if err := s.revokeAccessToken(oauth); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	// 
	updates := map[string]interface{}{
		"access_token":  "",
		"refresh_token": "",
		"token_type":    "",
		"expires_at":    nil,
		"status":        models.OAuthStatusRevoked,
		"updated_at":    time.Now(),
	}

	return s.repo.Update(id, updates)
}

// RevokeAllTokens ?func (s *OAuthService) RevokeAllTokens(id int64) error {
	return s.RevokeToken(id)
}

// GetValidToken ?func (s *OAuthService) GetValidToken(id int64) (string, error) {
	oauth, err := s.repo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("OAuth app not found: %w", err)
	}

	if oauth.AccessToken == "" {
		return "", fmt.Errorf("no access token available")
	}

	// ?	if oauth.ExpiresAt != nil && time.Now().After(*oauth.ExpiresAt) {
		// 
		if oauth.RefreshToken != "" {
			_, err := s.RefreshToken(id)
			if err != nil {
				return "", fmt.Errorf("token expired and refresh failed: %w", err)
			}
			// OAuth
			oauth, err = s.repo.GetByID(id)
			if err != nil {
				return "", fmt.Errorf("failed to get updated OAuth info: %w", err)
			}
		} else {
			return "", fmt.Errorf("token expired and no refresh token available")
		}
	}

	return oauth.AccessToken, nil
}

// MakeAuthenticatedRequest OAuth
func (s *OAuthService) MakeAuthenticatedRequest(id int64, method, url string, body []byte) (*http.Response, error) {
	token, err := s.GetValidToken(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return s.httpClient.Do(req)
}

// TokenResponse 
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// validateOAuthConfig OAuth
func (s *OAuthService) validateOAuthConfig(oauth *models.OAuth) error {
	if oauth.Name == "" {
		return fmt.Errorf("OAuth app name is required")
	}

	if oauth.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if oauth.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}

	if oauth.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	if oauth.RedirectURI == "" {
		return fmt.Errorf("redirect URI is required")
	}

	if !strings.HasPrefix(oauth.RedirectURI, "http://") && !strings.HasPrefix(oauth.RedirectURI, "https://") {
		return fmt.Errorf("redirect URI must start with http:// or https://")
	}

	return nil
}

// generateState state
func (s *OAuthService) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// buildAuthorizationURL URL
func (s *OAuthService) buildAuthorizationURL(oauth *models.OAuth, state string) (string, error) {
	var authURL string

	switch oauth.Provider {
	case "github":
		authURL = "https://github.com/login/oauth/authorize"
	case "google":
		authURL = "https://accounts.google.com/o/oauth2/v2/auth"
	case "microsoft":
		authURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	case "slack":
		authURL = "https://slack.com/oauth/v2/authorize"
	default:
		return "", fmt.Errorf("unsupported provider: %s", oauth.Provider)
	}

	params := url.Values{}
	params.Set("client_id", oauth.ClientID)
	params.Set("redirect_uri", oauth.RedirectURI)
	params.Set("state", state)
	params.Set("response_type", "code")

	if len(oauth.Scopes) > 0 {
		params.Set("scope", strings.Join(oauth.Scopes, " "))
	}

	// 
	switch oauth.Provider {
	case "google":
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	case "microsoft":
		params.Set("response_mode", "query")
	}

	return authURL + "?" + params.Encode(), nil
}

// exchangeToken 
func (s *OAuthService) exchangeToken(oauth *models.OAuth, code string) (*TokenResponse, error) {
	var tokenURL string

	switch oauth.Provider {
	case "github":
		tokenURL = "https://github.com/login/oauth/access_token"
	case "google":
		tokenURL = "https://oauth2.googleapis.com/token"
	case "microsoft":
		tokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	case "slack":
		tokenURL = "https://slack.com/api/oauth.v2.access"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", oauth.Provider)
	}

	params := url.Values{}
	params.Set("client_id", oauth.ClientID)
	params.Set("client_secret", oauth.ClientSecret)
	params.Set("code", code)
	params.Set("redirect_uri", oauth.RedirectURI)
	params.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// refreshAccessToken 
func (s *OAuthService) refreshAccessToken(oauth *models.OAuth) (*TokenResponse, error) {
	var tokenURL string

	switch oauth.Provider {
	case "github":
		return nil, fmt.Errorf("GitHub does not support token refresh")
	case "google":
		tokenURL = "https://oauth2.googleapis.com/token"
	case "microsoft":
		tokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	case "slack":
		tokenURL = "https://slack.com/api/oauth.v2.access"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", oauth.Provider)
	}

	params := url.Values{}
	params.Set("client_id", oauth.ClientID)
	params.Set("client_secret", oauth.ClientSecret)
	params.Set("refresh_token", oauth.RefreshToken)
	params.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh request failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &tokenResp, nil
}

// revokeAccessToken 
func (s *OAuthService) revokeAccessToken(oauth *models.OAuth) error {
	var revokeURL string

	switch oauth.Provider {
	case "github":
		// GitHub
		return nil
	case "google":
		revokeURL = "https://oauth2.googleapis.com/revoke"
	case "microsoft":
		// Microsoft
		return nil
	case "slack":
		revokeURL = "https://slack.com/api/auth.revoke"
	default:
		return fmt.Errorf("unsupported provider: %s", oauth.Provider)
	}

	if revokeURL == "" {
		return nil // 
	}

	params := url.Values{}
	params.Set("token", oauth.AccessToken)

	req, err := http.NewRequest("POST", revokeURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetOAuthStats OAuth
func (s *OAuthService) GetOAuthStats(id int64) (map[string]interface{}, error) {
	oauth, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("OAuth app not found: %w", err)
	}

	stats := map[string]interface{}{
		"name":         oauth.Name,
		"provider":     oauth.Provider,
		"client_id":    oauth.ClientID,
		"redirect_uri": oauth.RedirectURI,
		"scopes":       oauth.Scopes,
		"status":       oauth.Status,
		"is_active":    oauth.IsActive,
		"expires_at":   oauth.ExpiresAt,
		"created_at":   oauth.CreatedAt,
		"updated_at":   oauth.UpdatedAt,
	}

	// ?	if oauth.AccessToken != "" {
		stats["has_access_token"] = true
		if oauth.ExpiresAt != nil {
			stats["token_expired"] = time.Now().After(*oauth.ExpiresAt)
		}
	} else {
		stats["has_access_token"] = false
	}

	stats["has_refresh_token"] = oauth.RefreshToken != ""

	return stats, nil
}

