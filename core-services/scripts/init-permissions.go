package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Permission 权限模型
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role 角色模型
type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"uniqueIndex;not null"`
	Code        string       `json:"code" gorm:"uniqueIndex;not null"`
	Description string       `json:"description"`
	Level       int          `json:"level" gorm:"default:1"`
	Status      string       `json:"status" gorm:"default:active"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func main() {
	// 数据库连接
	dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&Permission{}, &Role{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化权限数据
	if err := initPermissions(db); err != nil {
		log.Fatal("Failed to initialize permissions:", err)
	}

	// 初始化角色数据
	if err := initRoles(db); err != nil {
		log.Fatal("Failed to initialize roles:", err)
	}

	fmt.Println("权限和角色数据初始化完成！")
}

func initPermissions(db *gorm.DB) error {
	permissions := []Permission{
		// 菜单管理权限
		{Name: "菜单查看", Code: "menu:read", Description: "查看菜单列表和详情", Resource: "menu", Action: "read"},
		{Name: "菜单创建", Code: "menu:create", Description: "创建新菜单", Resource: "menu", Action: "create"},
		{Name: "菜单编辑", Code: "menu:update", Description: "编辑菜单信息", Resource: "menu", Action: "update"},
		{Name: "菜单删除", Code: "menu:delete", Description: "删除菜单", Resource: "menu", Action: "delete"},
		{Name: "菜单树查看", Code: "menu:tree", Description: "查看菜单树结构", Resource: "menu", Action: "tree"},

		// 用户管理权限
		{Name: "用户查看", Code: "user:read", Description: "查看用户列表和详情", Resource: "user", Action: "read"},
		{Name: "用户创建", Code: "user:create", Description: "创建新用户", Resource: "user", Action: "create"},
		{Name: "用户编辑", Code: "user:update", Description: "编辑用户信息", Resource: "user", Action: "update"},
		{Name: "用户删除", Code: "user:delete", Description: "删除用户", Resource: "user", Action: "delete"},
		{Name: "用户状态管理", Code: "user:status", Description: "启用/禁用用户", Resource: "user", Action: "status"},

		// 角色管理权限
		{Name: "角色查看", Code: "role:read", Description: "查看角色列表和详情", Resource: "role", Action: "read"},
		{Name: "角色创建", Code: "role:create", Description: "创建新角色", Resource: "role", Action: "create"},
		{Name: "角色编辑", Code: "role:update", Description: "编辑角色信息", Resource: "role", Action: "update"},
		{Name: "角色删除", Code: "role:delete", Description: "删除角色", Resource: "role", Action: "delete"},
		{Name: "角色权限分配", Code: "role:assign", Description: "为角色分配权限", Resource: "role", Action: "assign"},

		// 权限管理权限
		{Name: "权限查看", Code: "permission:read", Description: "查看权限列表和详情", Resource: "permission", Action: "read"},
		{Name: "权限创建", Code: "permission:create", Description: "创建新权限", Resource: "permission", Action: "create"},
		{Name: "权限编辑", Code: "permission:update", Description: "编辑权限信息", Resource: "permission", Action: "update"},
		{Name: "权限删除", Code: "permission:delete", Description: "删除权限", Resource: "permission", Action: "delete"},

		// 系统管理权限
		{Name: "系统配置查看", Code: "system:config:read", Description: "查看系统配置", Resource: "system", Action: "config:read"},
		{Name: "系统配置编辑", Code: "system:config:update", Description: "编辑系统配置", Resource: "system", Action: "config:update"},
		{Name: "系统日志查看", Code: "system:log:read", Description: "查看系统日志", Resource: "system", Action: "log:read"},
		{Name: "系统监控", Code: "system:monitor", Description: "系统监控和状态查看", Resource: "system", Action: "monitor"},

		// 文件管理权限
		{Name: "文件上传", Code: "file:upload", Description: "上传文件", Resource: "file", Action: "upload"},
		{Name: "文件下载", Code: "file:download", Description: "下载文件", Resource: "file", Action: "download"},
		{Name: "文件删除", Code: "file:delete", Description: "删除文件", Resource: "file", Action: "delete"},
		{Name: "文件查看", Code: "file:read", Description: "查看文件列表", Resource: "file", Action: "read"},

		// 数据管理权限
		{Name: "数据导出", Code: "data:export", Description: "导出数据", Resource: "data", Action: "export"},
		{Name: "数据导入", Code: "data:import", Description: "导入数据", Resource: "data", Action: "import"},
		{Name: "数据备份", Code: "data:backup", Description: "数据备份", Resource: "data", Action: "backup"},
		{Name: "数据恢复", Code: "data:restore", Description: "数据恢复", Resource: "data", Action: "restore"},
	}

	for _, perm := range permissions {
		var existingPerm Permission
		result := db.Where("code = ?", perm.Code).First(&existingPerm)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to create permission %s: %v", perm.Code, err)
			}
			fmt.Printf("Created permission: %s\n", perm.Name)
		} else {
			fmt.Printf("Permission already exists: %s\n", perm.Name)
		}
	}

	return nil
}

func initRoles(db *gorm.DB) error {
	roles := []struct {
		Role        Role
		Permissions []string // 权限代码列表
	}{
		{
			Role: Role{
				Name:        "超级管理员",
				Code:        "super_admin",
				Description: "系统超级管理员，拥有所有权限",
				Level:       1,
				Status:      "active",
			},
			Permissions: []string{
				"menu:read", "menu:create", "menu:update", "menu:delete", "menu:tree",
				"user:read", "user:create", "user:update", "user:delete", "user:status",
				"role:read", "role:create", "role:update", "role:delete", "role:assign",
				"permission:read", "permission:create", "permission:update", "permission:delete",
				"system:config:read", "system:config:update", "system:log:read", "system:monitor",
				"file:upload", "file:download", "file:delete", "file:read",
				"data:export", "data:import", "data:backup", "data:restore",
			},
		},
		{
			Role: Role{
				Name:        "管理员",
				Code:        "admin",
				Description: "系统管理员，拥有大部分管理权限",
				Level:       2,
				Status:      "active",
			},
			Permissions: []string{
				"menu:read", "menu:create", "menu:update", "menu:delete", "menu:tree",
				"user:read", "user:create", "user:update", "user:status",
				"role:read", "role:create", "role:update",
				"permission:read",
				"system:config:read", "system:log:read", "system:monitor",
				"file:upload", "file:download", "file:read",
				"data:export", "data:import",
			},
		},
		{
			Role: Role{
				Name:        "普通用户",
				Code:        "user",
				Description: "普通用户，拥有基本的查看权限",
				Level:       3,
				Status:      "active",
			},
			Permissions: []string{
				"menu:read", "menu:tree",
				"user:read",
				"file:read", "file:download",
			},
		},
		{
			Role: Role{
				Name:        "访客",
				Code:        "guest",
				Description: "访客用户，只有最基本的查看权限",
				Level:       4,
				Status:      "active",
			},
			Permissions: []string{
				"menu:read",
			},
		},
	}

	for _, roleData := range roles {
		var existingRole Role
		result := db.Where("code = ?", roleData.Role.Code).First(&existingRole)
		if result.Error == gorm.ErrRecordNotFound {
			// 创建角色
			if err := db.Create(&roleData.Role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %v", roleData.Role.Code, err)
			}

			// 获取权限并关联到角色
			var permissions []Permission
			if err := db.Where("code IN ?", roleData.Permissions).Find(&permissions).Error; err != nil {
				return fmt.Errorf("failed to find permissions for role %s: %v", roleData.Role.Code, err)
			}

			// 关联权限到角色
			if err := db.Model(&roleData.Role).Association("Permissions").Append(permissions); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %v", roleData.Role.Code, err)
			}

			fmt.Printf("Created role: %s with %d permissions\n", roleData.Role.Name, len(permissions))
		} else {
			fmt.Printf("Role already exists: %s\n", roleData.Role.Name)
		}
	}

	return nil
}