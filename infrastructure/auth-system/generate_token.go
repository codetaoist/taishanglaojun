package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims JWT声明 - 匹配认证服务的结构
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	SessionID   uuid.UUID `json:"session_id"`
	TokenType   string    `json:"token_type"` // access, refresh
	Permissions []string  `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func main() {
	now := time.Now().UTC()
	userID := uuid.MustParse("12345678-90ab-cdef-1234-567890abcdef")
	sessionID := uuid.New()

	claims := &Claims{
		UserID:      userID,
		Username:    "admin",
		Email:       "admin@example.com",
		Role:        "admin",
		SessionID:   sessionID,
		TokenType:   "access",
		Permissions: []string{"read", "write"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    "taishang-auth-system",
			Subject:   userID.String(),
			Audience:  []string{"taishang-system", "taishang-web", "taishang-mobile"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("your-super-secret-jwt-key-change-in-production-laojun-2024"))
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		return
	}

	fmt.Println(tokenString)
}