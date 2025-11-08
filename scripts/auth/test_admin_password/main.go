package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 获取数据库URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost/taishanglaojun?sslmode=disable"
	}

	// 连接数据库
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 获取admin用户的密码哈希
	var passwordHash string
	err = db.QueryRow("SELECT password_hash FROM lao_users WHERE username = 'admin'").Scan(&passwordHash)
	if err != nil {
		log.Fatalf("Failed to get admin user: %v", err)
	}

	// 验证密码
	password := "admin123"
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
	} else {
		fmt.Println("Password verification succeeded")
	}

	// 重新生成密码哈希
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 更新admin用户的密码哈希
	_, err = db.Exec("UPDATE lao_users SET password_hash = $1 WHERE username = 'admin'", string(newHashedPassword))
	if err != nil {
		log.Fatalf("Failed to update admin password: %v", err)
	}

	fmt.Println("Admin password updated successfully")

	// 再次验证密码
	err = bcrypt.CompareHashAndPassword([]byte(newHashedPassword), []byte(password))
	if err != nil {
		fmt.Printf("Password verification failed after update: %v\n", err)
	} else {
		fmt.Println("Password verification succeeded after update")
	}
}