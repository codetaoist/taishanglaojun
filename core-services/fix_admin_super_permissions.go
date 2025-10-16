package main

import (
"fmt"
"log"
"time"
"github.com/google/uuid"
"gorm.io/driver/mysql"
"gorm.io/gorm"
)

type User struct {
ID        string    `gorm:"column:id;primaryKey"`
Username  string    `gorm:"column:username"`
Email     string    `gorm:"column:email"`
Role      string    `gorm:"column:role"`
Status    string    `gorm:"column:status"`
IsActive  bool      `gorm:"column:is_active"`
Level     int       `gorm:"column:level"`
CreatedAt time.Time `gorm:"column:created_at"`
UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Role struct {
ID          string    `gorm:"column:id;primaryKey"`
Name        string    `gorm:"column:name"`
Code        string    `gorm:"column:code"`
Description string    `gorm:"column:description"`
Level       int       `gorm:"column:level"`
Status      string    `gorm:"column:status"`
CreatedAt   time.Time `gorm:"column:created_at"`
UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type Permission struct {
ID          string    `gorm:"column:id;primaryKey"`
Name        string    `gorm:"column:name"`
Code        string    `gorm:"column:code"`
Description string    `gorm:"column:description"`
Resource    string    `gorm:"column:resource"`
Action      string    `gorm:"column:action"`
Status      string    `gorm:"column:status"`
CreatedAt   time.Time `gorm:"column:created_at"`
UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type UserRole struct {
ID        string    `gorm:"column:id;primaryKey"`
UserID    string    `gorm:"column:user_id"`
RoleID    string    `gorm:"column:role_id"`
CreatedAt time.Time `gorm:"column:created_at"`
UpdatedAt time.Time `gorm:"column:updated_at"`
}

type RolePermission struct {
RoleID       string `gorm:"column:role_id;primaryKey"`
PermissionID string `gorm:"column:permission_id;primaryKey"`
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
fmt.Printf(" 找到admin用户: %s (%s), 当前角色: %s\n", adminUser.Username, adminUser.Email, adminUser.Role)

// 查找或创建超级管理员角色
var superAdminRole Role
if err := db.Table("roles").Where("code = ?", "super_admin").First(&superAdminRole).Error; err != nil {
if err == gorm.ErrRecordNotFound {
// 创建超级管理员角色
superAdminRole = Role{
ID:          uuid.New().String(),
Name:        "超级管理员",
Code:        "super_admin",
Description: "系统超级管理员，拥有所有权限",
Level:       100,
Status:      "active",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}
if err := db.Table("roles").Create(&superAdminRole).Error; err != nil {
log.Fatal("创建超级管理员角色失败:", err)
}
fmt.Printf("✅ 创建超级管理员角色: %s (%s)\n", superAdminRole.Name, superAdminRole.Code)
} else {
log.Fatal("查询超级管理员角色失败:", err)
}
} else {
fmt.Printf(" 找到超级管理员角色: %s (%s)\n", superAdminRole.Name, superAdminRole.Code)
}

// 检查用户是否已经有超级管理员角色
var existingUserRole UserRole
if err := db.Table("user_roles").Where("user_id = ? AND role_id = ?", adminUser.ID, superAdminRole.ID).First(&existingUserRole).Error; err != nil {
if err == gorm.ErrRecordNotFound {
// 分配超级管理员角色给用户
userRole := UserRole{
ID:        uuid.New().String(),
UserID:    adminUser.ID,
RoleID:    superAdminRole.ID,
CreatedAt: time.Now(),
UpdatedAt: time.Now(),
}
if err := db.Table("user_roles").Create(&userRole).Error; err != nil {
log.Fatal("分配超级管理员角色失败:", err)
}
fmt.Printf(" 为admin用户分配超级管理员角色\n")
} else {
log.Fatal("检查用户角色失败:", err)
}
} else {
fmt.Printf(" admin用户已经拥有超级管理员角色\n")
}

// 获取所有权限
var allPermissions []Permission
if err := db.Table("permissions").Where("status = ?", "active").Find(&allPermissions).Error; err != nil {
log.Fatal("获取权限列表失败:", err)
}
fmt.Printf(" 找到 %d 个活跃权限\n", len(allPermissions))

// 为超级管理员角色分配所有权限
var existingRolePermissions []RolePermission
if err := db.Table("role_permissions").Where("role_id = ?", superAdminRole.ID).Find(&existingRolePermissions).Error; err != nil {
log.Fatal("查询角色权限失败:", err)
}

existingPermissionMap := make(map[string]bool)
for _, rp := range existingRolePermissions {
existingPermissionMap[rp.PermissionID] = true
}

addedCount := 0
for _, permission := range allPermissions {
if !existingPermissionMap[permission.ID] {
rolePermission := RolePermission{
RoleID:       superAdminRole.ID,
PermissionID: permission.ID,
}
if err := db.Table("role_permissions").Create(&rolePermission).Error; err != nil {
fmt.Printf(" 分配权限失败 %s: %v\n", permission.Code, err)
} else {
addedCount++
}
}
}
fmt.Printf(" 为超级管理员角色新增 %d 个权限\n", addedCount)

// 验证最终结果
fmt.Println("\n🔍 验证admin用户最终权限状态:")

// 查询用户角色
var userRoles []UserRole
if err := db.Table("user_roles").Where("user_id = ?", adminUser.ID).Find(&userRoles).Error; err != nil {
log.Fatal("查询用户角色失败:", err)
}

fmt.Printf("   用户角色数量: %d\n", len(userRoles))

totalPermissions := 0
for _, userRole := range userRoles {
var role Role
if err := db.Table("roles").Where("id = ?", userRole.RoleID).First(&role).Error; err != nil {
continue
}

var permissionCount int64
if err := db.Table("role_permissions").Where("role_id = ?", role.ID).Count(&permissionCount).Error; err != nil {
continue
}

fmt.Printf("    - %s (%s) - 级别: %d, 权限数: %d\n", role.Name, role.Code, role.Level, permissionCount)
totalPermissions += int(permissionCount)
}

fmt.Printf("   用户总权限数量: %d\n", totalPermissions)
fmt.Println("\n admin账号超级管理员权限配置完成!")
}
