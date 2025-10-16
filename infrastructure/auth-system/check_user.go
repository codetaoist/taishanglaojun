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

	// 查询用户状态
	query := "SELECT id, username, email, status FROM users WHERE username = ?"
	row := db.QueryRow(query, "testuser5")

	var id, username, email, status string
	err = row.Scan(&id, &username, &email, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
		} else {
			log.Fatal("Failed to query user:", err)
		}
		return
	}

	fmt.Printf("User ID: %s\n", id)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Status: %s\n", status)
}