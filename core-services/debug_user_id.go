package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID       string `gorm:"type:char(36);primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
}



func main() {
	// 数据库连接
	dsn := "root:123456@tcp(localhost:3306)/taishanglaojun?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 1. 查找admin用户
	var user User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		log.Fatal("Failed to find admin user:", err)
	}

	fmt.Printf("数据库中admin用户信息:\n")
	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Email: %s\n", user.Email)

	// 2. 检查预期的用户ID
	fmt.Printf("\n预期的用户ID应该是: %s\n", user.ID)
	
	// 3. 检查是否有其他同名用户
	var users []User
	if err := db.Where("username = ?", "admin").Find(&users).Error; err != nil {
		log.Fatal("Failed to find users:", err)
	}
	
	fmt.Printf("\n所有名为admin的用户:\n")
	for i, u := range users {
		fmt.Printf("用户 %d: ID=%s, Username=%s, Email=%s\n", i+1, u.ID, u.Username, u.Email)
	}
	
	// 4. 检查是否有重复的用户名或邮箱
	var count int64
	db.Model(&User{}).Where("username = ? OR email LIKE ?", "admin", "%admin%").Count(&count)
	fmt.Printf("\n包含admin的用户总数: %d\n", count)
}