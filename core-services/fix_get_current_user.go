package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

// 数据库模型
type User struct {
	ID       string `gorm:"primaryKey"`
	Username string
	Email    string
	Role     string
}

type Role struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Code        string
	Description string
	Level       int
	Status      string
}

type Permission struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Code        string
	Description string
	Resource    string
	Action      string
	Status      string
}

type UserRole struct {
	UserID string `gorm:"primaryKey"`
	RoleID string `gorm:"primaryKey"`
}

type RolePermission struct {
	RoleID       string `gorm:"primaryKey"`
	PermissionID string `gorm:"primaryKey"`
}

func main() {
	fmt.Println("🔧 开始修复后端GetCurrentUser方法...")

	// 读取当前的handlers.go文件
	filePath := "D:\\work\\taishanglaojun\\core-services\\internal\\middleware\\handlers.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("读取文件失败:", err)
	}

	fileContent := string(content)

	// 查找GetCurrentUser方法的开始和结束位置
	startMarker := "// GetCurrentUser"
	endMarker := "// Logout"
	
	startIndex := strings.Index(fileContent, startMarker)
	endIndex := strings.Index(fileContent, endMarker)
	
	if startIndex == -1 || endIndex == -1 {
		log.Fatal("无法找到GetCurrentUser方法的位置")
	}

	fmt.Println("📍 找到GetCurrentUser方法位置")

	// 新的GetCurrentUser方法实现
	newGetCurrentUserMethod := `// GetCurrentUser 
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "USER_NOT_FOUND",
			"message": err.Error(),
		})
		return
	}

	// 从数据库获取用户的实际权限
	permissions, roles, err := h.getUserPermissionsFromDB(userID)
	if err != nil {
		h.logger.Error("Failed to get user permissions", zap.Error(err))
		// 如果获取权限失败，使用基础权限
		permissions = []string{"user:read"}
		roles = []string{"user"}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           userID,
			"user_id":      userID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.Username,
			"role":         string(user.Role),
			"roles":        roles,
			"permissions":  permissions,
			"level":        user.Level,
			"isAdmin":      contains(roles, "admin") || contains(roles, "super_admin") || contains(roles, "系统管理员"),
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		},
		"message": "",
	})
}

// getUserPermissionsFromDB 从数据库获取用户权限
func (h *AuthHandler) getUserPermissionsFromDB(userID string) ([]string, []string, error) {
	// 获取用户角色
	var userRoles []struct {
		RoleID string
		Role   struct {
			ID   string
			Name string
			Code string
		}
	}

	err := h.db.Table("user_roles").
		Select("user_roles.role_id, roles.id, roles.name, roles.code").
		Joins("LEFT JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = 'active'", userID).
		Scan(&userRoles).Error

	if err != nil {
		return nil, nil, fmt.Errorf("获取用户角色失败: %v", err)
	}

	var roleNames []string
	var roleIDs []string
	
	for _, ur := range userRoles {
		roleNames = append(roleNames, ur.Role.Name)
		roleIDs = append(roleIDs, ur.RoleID)
	}

	// 如果用户没有角色，返回基础权限
	if len(roleIDs) == 0 {
		return []string{"user:read"}, []string{"user"}, nil
	}

	// 获取角色权限
	var permissions []struct {
		Code string
	}

	err = h.db.Table("role_permissions").
		Select("permissions.code").
		Joins("LEFT JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id IN ? AND permissions.status = 'active'", roleIDs).
		Scan(&permissions).Error

	if err != nil {
		return nil, nil, fmt.Errorf("获取角色权限失败: %v", err)
	}

	var permissionCodes []string
	for _, p := range permissions {
		permissionCodes = append(permissionCodes, p.Code)
	}

	// 去重
	permissionCodes = removeDuplicates(permissionCodes)
	roleNames = removeDuplicates(roleNames)

	return permissionCodes, roleNames, nil
}

// contains 检查字符串数组是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// removeDuplicates 去除字符串数组中的重复项
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

`

	// 替换GetCurrentUser方法
	newContent := fileContent[:startIndex] + newGetCurrentUserMethod + fileContent[endIndex:]

	// 写回文件
	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		log.Fatal("写入文件失败:", err)
	}

	fmt.Println("✅ 成功修复GetCurrentUser方法")
	fmt.Println("📝 修改内容:")
	fmt.Println("   - 移除硬编码的权限分配逻辑")
	fmt.Println("   - 添加从数据库获取用户实际权限的方法")
	fmt.Println("   - 添加getUserPermissionsFromDB辅助方法")
	fmt.Println("   - 添加contains和removeDuplicates工具方法")
	fmt.Println("   - 根据实际角色判断isAdmin状态")

	fmt.Println("\n🔄 需要重启后端服务以应用更改")
}