package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Role struct {
	ID          string `gorm:"type:char(36);primaryKey"`
	Name        string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Code        string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description string `gorm:"type:text"`
	Type        string `gorm:"type:varchar(50);default:custom"`
	Level       int    `gorm:"default:1"`
	IsActive    bool   `gorm:"default:true"`
	Status      string `gorm:"type:varchar(20);default:active"`
	CreatedAt   string `gorm:"type:datetime(3)"`
	UpdatedAt   string `gorm:"type:datetime(3)"`
}

func main() {
	// 数据库连接字符串 - 使用配置文件中的远程数据库
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	fmt.Println("MySQL数据库连接成功!")

	// 查询所有角色数据
	var roles []Role
	result := db.Find(&roles)
	if result.Error != nil {
		log.Fatal("查询角色数据失败:", result.Error)
	}

	fmt.Printf("找到 %d 个角色:\n", len(roles))
	fmt.Println("ID\t\t\t\t\tName\t\t\tCode\t\t\tType\t\tIsActive\tStatus")
	fmt.Println("--------------------------------------------------------------------------------------------")
	
	for _, role := range roles {
		fmt.Printf("%s\t%s\t\t%s\t\t%s\t\t%t\t\t%s\n", 
			role.ID[:8]+"...", 
			role.Name, 
			role.Code, 
			role.Type, 
			role.IsActive, 
			role.Status)
	}
	
	fmt.Println("\n数据查询完成!")
}