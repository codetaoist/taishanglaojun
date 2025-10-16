package main

import (
	"fmt"
	"log"

	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 初始化数据库
	db, err := database.NewConnection(cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 查询所有菜单
	var menus []models.Menu
	err = db.Preload("Children").Where("parent_id IS NULL").Order("sort ASC").Find(&menus).Error
	if err != nil {
		log.Fatalf("Failed to query menus: %v", err)
	}

	fmt.Printf("=== 数据库中的菜单结构 ===\n")
	fmt.Printf("找到 %d 个顶级菜单\n\n", len(menus))

	for _, menu := range menus {
		printMenu(menu, 0)
	}

	// 检查超级管理员权限
	fmt.Printf("\n=== 超级管理员菜单权限检查 ===\n")
	
	// 查询所有菜单（包括子菜单）
	var allMenus []models.Menu
	err = db.Find(&allMenus).Error
	if err != nil {
		log.Fatalf("Failed to query all menus: %v", err)
	}

	fmt.Printf("数据库中总共有 %d 个菜单项\n", len(allMenus))
	
	// 按角色分类统计
	roleStats := make(map[models.UserRole]int)
	for _, menu := range allMenus {
		roleStats[menu.RequiredRole]++
	}

	fmt.Printf("\n按角色分类的菜单统计：\n")
	for role, count := range roleStats {
		fmt.Printf("- %s: %d 个菜单\n", role, count)
	}

	// 检查超级管理员可见的菜单
	var superAdminMenus []models.Menu
	err = db.Where("required_role IN (?, ?, ?)", models.RoleUser, models.RoleAdmin, models.RoleSuperAdmin).
		Where("is_visible = ? AND is_enabled = ?", true, true).
		Order("sort ASC").Find(&superAdminMenus).Error
	if err != nil {
		log.Fatalf("Failed to query super admin menus: %v", err)
	}

	fmt.Printf("\n超级管理员可见的菜单 (%d 个)：\n", len(superAdminMenus))
	for _, menu := range superAdminMenus {
		level := ""
		for i := 1; i < menu.Level; i++ {
			level += "  "
		}
		fmt.Printf("%s- %s (%s) - 角色要求: %s\n", level, menu.Title, menu.Path, menu.RequiredRole)
	}
}

func printMenu(menu models.Menu, level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	fmt.Printf("%s- %s (%s)\n", indent, menu.Title, menu.Path)
	fmt.Printf("%s  ID: %s\n", indent, menu.ID)
	fmt.Printf("%s  Name: %s\n", indent, menu.Name)
	fmt.Printf("%s  Icon: %s\n", indent, menu.Icon)
	fmt.Printf("%s  Sort: %d\n", indent, menu.Sort)
	fmt.Printf("%s  Level: %d\n", indent, menu.Level)
	fmt.Printf("%s  Visible: %t\n", indent, menu.IsVisible)
	fmt.Printf("%s  Enabled: %t\n", indent, menu.IsEnabled)
	fmt.Printf("%s  Required Role: %s\n", indent, menu.RequiredRole)

	if len(menu.Children) > 0 {
		fmt.Printf("%s  Children (%d):\n", indent, len(menu.Children))
		for _, child := range menu.Children {
			printMenu(child, level+1)
		}
	}
	fmt.Printf("\n")
}