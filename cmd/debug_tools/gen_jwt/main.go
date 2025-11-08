package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "1",
		"username": "testuser",
		"role":     "user",
		"exp":      1736399400, // 2025-01-09 13:30:00
	})
	
	tokenString, err := token.SignedString([]byte("your-secret-key-change-in-production"))
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		return
	}
	
	fmt.Println(tokenString)
}