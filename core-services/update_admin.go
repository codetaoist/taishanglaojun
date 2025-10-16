package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接信息
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun"
	
	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// 更新用户角色
	result, err := db.Exec("UPDATE users SET role = 'SUPER_ADMIN' WHERE username = 'superadmin'")
	if err != nil {
		log.Fatal("Failed to update user role:", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal("Failed to get rows affected:", err)
	}

	fmt.Printf("Updated %d rows\n", rowsAffected)

	// 查询更新后的用户信息
	var id, username, email, role string
	err = db.QueryRow("SELECT id, username, email, role FROM users WHERE username = 'superadmin'").Scan(&id, &username, &email, &role)
	if err != nil {
		log.Fatal("Failed to query user:", err)
	}

	fmt.Printf("User updated successfully:\n")
	fmt.Printf("ID: %s\n", id)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Role: %s\n", role)
}