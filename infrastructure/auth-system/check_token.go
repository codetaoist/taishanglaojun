package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	dsn := "root:123456@tcp(localhost:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 查询用户ID
	var userID string
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser5").Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
		} else {
			log.Fatal("Failed to query user:", err)
		}
		return
	}

	fmt.Printf("User ID: %s\n", userID)

	// 查询验证token
	query := "SELECT token, type, expires_at, used_at FROM tokens WHERE user_id = ? AND type = 'verification' ORDER BY created_at DESC LIMIT 1"
	row := db.QueryRow(query, userID)

	var token, tokenType, expiresAt string
	var usedAt sql.NullString
	err = row.Scan(&token, &tokenType, &expiresAt, &usedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No verification token found")
		} else {
			log.Fatal("Failed to query token:", err)
		}
		return
	}

	fmt.Printf("Token: %s\n", token)
	fmt.Printf("Type: %s\n", tokenType)
	fmt.Printf("Expires At: %s\n", expiresAt)
	if usedAt.Valid {
		fmt.Printf("Used At: %s\n", usedAt.String)
	} else {
		fmt.Println("Used At: NULL (not used)")
	}
}