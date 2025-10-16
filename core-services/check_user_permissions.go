package main

import (
"fmt"
"log"
"gorm.io/driver/mysql"
"gorm.io/gorm"
)

type User struct {
ID       string `gorm:"column:id;primaryKey"`
Username string `gorm:"column:username"`
Email    string `gorm:"column:email"`
Role     string `gorm:"column:role"`
}

type UserRole struct {
ID     string `gorm:"column:id;primaryKey"`
UserID string `gorm:"column:user_id"`
RoleID string `gorm:"column:role_id"`
}

type Role struct {
ID   string `gorm:"column:id;primaryKey"`
Name string `gorm:"column:name"`
Code string `gorm:"column:code"`
}

type Permission struct {
ID   string `gorm:"column:id;primaryKey"`
Name string `gorm:"column:name"`
Code string `gorm:"column:code"`
}

func main() {
dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"

fmt.Println("正在连接MySQL数据库...")
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
log.Fatal("连接数据库失败:", err)
}
fmt.Println(" MySQL数据库连接成功!")

// 查找admin用户
var adminUser User
if err := db.Table("users").Where("username = ?", "admin").First(&adminUser).Error; err != nil {
log.Fatal("查找admin用户失败:", err)
}
fmt.Printf(" 找到admin用户: %s (ID: %s)\n", adminUser.Username, adminUser.ID)

// 查询用户角色
var userRoles []UserRole
if err := db.Table("user_roles").Where("user_id = ?", adminUser.ID).Find(&userRoles).Error; err != nil {
log.Fatal("查询用户角色失败:", err)
}
fmt.Printf(" 用户拥有 %d 个角色\n", len(userRoles))

for _, userRole := range userRoles {
var role Role
if err := db.Table("roles").Where("id = ?", userRole.RoleID).First(&role).Error; err != nil {
continue
}
fmt.Printf("  - 角色: %s (%s) - ID: %s\n", role.Name, role.Code, role.ID)

// 查询角色权限
var permissions []Permission
if err := db.Raw(`
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = ?
`, role.ID).Scan(&permissions).Error; err != nil {
fmt.Printf("     查询角色权限失败: %v\n", err)
continue
}

fmt.Printf("    权限数量: %d\n", len(permissions))
for i, perm := range permissions {
if i < 5 { // 只显示前5个权限
fmt.Printf("      - %s (%s)\n", perm.Name, perm.Code)
} else if i == 5 {
fmt.Printf("      - ... 还有 %d 个权限\n", len(permissions)-5)
break
}
}
}

// 检查API中的用户ID是否匹配
fmt.Printf("\n API返回的用户ID: 1c54d101-7840-46bb-81d9-4f430b0a41de\n")
fmt.Printf(" 数据库中的用户ID: %s\n", adminUser.ID)

if adminUser.ID == "1c54d101-7840-46bb-81d9-4f430b0a41de" {
fmt.Println(" 用户ID匹配")
} else {
fmt.Println(" 用户ID不匹配！这可能是权限问题的原因")

// 查找API返回的用户ID
var apiUser User
if err := db.Table("users").Where("id = ?", "1c54d101-7840-46bb-81d9-4f430b0a41de").First(&apiUser).Error; err != nil {
fmt.Printf(" 未找到API返回的用户ID: %v\n", err)
} else {
fmt.Printf(" 找到API用户: %s (%s)\n", apiUser.Username, apiUser.Email)
}
}
}
