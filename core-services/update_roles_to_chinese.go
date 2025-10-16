package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Role struct {
	ID          string `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	Code        string `gorm:"column:code"`
	Description string `gorm:"column:description"`
	Type        string `gorm:"column:type"`
	Level       int    `gorm:"column:level"`
	Status      string `gorm:"column:status"`
	IsActive    bool   `gorm:"column:is_active"`
}

func main() {
	// 数据库连接字符串
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	fmt.Println("MySQL数据库连接成功!")

	// 定义角色的中文映射
	roleTranslations := map[string]map[string]string{
		"qa_engineer": {
			"name":        "质量保证工程师",
			"description": "负责软件质量保证和测试工作",
		},
		"viewer": {
			"name":        "查看者",
			"description": "只拥有查看权限的用户角色",
		},
		"manager": {
			"name":        "部门经理",
			"description": "拥有部门管理权限的管理者角色",
		},
		"content_editor": {
			"name":        "内容编辑员",
			"description": "负责内容编辑和审核工作",
		},
		"test_role": {
			"name":        "测试角色",
			"description": "用于系统测试的角色",
		},
		"admin": {
			"name":        "系统管理员",
			"description": "拥有系统最高权限的管理员角色",
		},
		"user": {
			"name":        "普通用户",
			"description": "系统的普通用户角色",
		},
		"editor": {
			"name":        "编辑员",
			"description": "负责内容编辑的用户角色",
		},
		"developer": {
			"name":        "开发工程师",
			"description": "负责系统开发和维护的技术人员",
		},
		"analyst": {
			"name":        "数据分析师",
			"description": "负责数据分析和报告的专业人员",
		},
	}

	// 查询所有角色
	var roles []Role
	if err := db.Table("roles").Find(&roles).Error; err != nil {
		log.Fatal("查询角色失败:", err)
	}

	fmt.Printf("\n📋 找到 %d 个角色，开始更新为中文:\n", len(roles))

	// 更新每个角色
	for _, role := range roles {
		fmt.Printf("\n🔄 处理角色: %s (code: %s)\n", role.Name, role.Code)
		
		// 查找对应的中文翻译
		if translation, exists := roleTranslations[role.Code]; exists {
			// 更新角色名称和描述
			updateData := map[string]interface{}{
				"name":        translation["name"],
				"description": translation["description"],
			}
			
			if err := db.Table("roles").Where("id = ?", role.ID).Updates(updateData).Error; err != nil {
				fmt.Printf("  ❌ 更新失败: %v\n", err)
				continue
			}
			
			fmt.Printf("  ✅ 更新成功:\n")
			fmt.Printf("     名称: %s -> %s\n", role.Name, translation["name"])
			fmt.Printf("     描述: %s -> %s\n", role.Description, translation["description"])
		} else {
			fmt.Printf("  ⚠️  未找到对应的中文翻译，跳过\n")
		}
	}

	// 验证更新结果
	fmt.Println("\n🎯 更新后的角色列表:")
	var updatedRoles []Role
	if err := db.Table("roles").Find(&updatedRoles).Error; err != nil {
		log.Fatal("查询更新后的角色失败:", err)
	}

	for i, role := range updatedRoles {
		fmt.Printf("  %d. %s (%s)\n", i+1, role.Name, role.Code)
		fmt.Printf("     描述: %s\n", role.Description)
		fmt.Printf("     状态: %s, 级别: %d\n\n", role.Status, role.Level)
	}

	fmt.Println("✅ 角色中文化更新完成!")
}