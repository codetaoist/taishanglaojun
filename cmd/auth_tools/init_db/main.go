package main

import (
	"database/sql"
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

	// 创建用户表
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS lao_users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL DEFAULT 'user',
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`

	if _, err := db.Exec(createUserTableSQL); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// 创建会话表
	createSessionTableSQL := `
	CREATE TABLE IF NOT EXISTS lao_sessions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		token_hash VARCHAR(255) NOT NULL UNIQUE,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		FOREIGN KEY (user_id) REFERENCES lao_users(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(createSessionTableSQL); err != nil {
		log.Fatalf("Failed to create sessions table: %v", err)
	}

	// 检查admin用户是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM lao_users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check admin user: %v", err)
	}

	// 如果admin用户不存在，则创建
	if count == 0 {
		// 生成密码哈希
		password := "admin123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		// 插入admin用户
		_, err = db.Exec(`
			INSERT INTO lao_users (username, email, password_hash, role, status)
			VALUES ('admin', 'admin@example.com', $1, 'admin', 'active')
		`, string(hashedPassword))
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		log.Println("Admin user created successfully")
	} else {
		log.Println("Admin user already exists")
	}

	// 创建黑名单表
	createBlacklistTableSQL := `
	CREATE TABLE IF NOT EXISTS lao_token_blacklist (
		id SERIAL PRIMARY KEY,
		token_hash VARCHAR(255) NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		reason VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		FOREIGN KEY (user_id) REFERENCES lao_users(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(createBlacklistTableSQL); err != nil {
		log.Fatalf("Failed to create blacklist table: %v", err)
	}

	log.Println("Database initialization completed successfully")
}