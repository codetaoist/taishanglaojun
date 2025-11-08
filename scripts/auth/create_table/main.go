package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 获取数据库URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost/taishanglaojun?sslmode=disable"
	}

	// 连接数据库
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 读取SQL文件
	sqlBytes, err := ioutil.ReadFile("../../services/auth/migrations/create_blacklist_table.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// 执行SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	log.Println("Successfully created token blacklist table")
}