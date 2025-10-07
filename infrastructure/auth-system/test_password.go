package main

import (
	"fmt"
	"log"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/utils"
)

func main() {
	fmt.Println("Testing password validation...")

	password := "Abcd1234!@#$"
	userInfo := []string{"debuguser", "debuguser@example.com", "", ""}

	fmt.Printf("Testing password: %s\n", password)
	fmt.Printf("User info: %v\n", userInfo)

	// 创建自定义规则
	customRules := []func(string) (bool, string){
		utils.CommonCustomRules.NoUserInfo(userInfo),
		utils.CommonCustomRules.MinUniqueChars(6),
	}

	fmt.Println("Created custom rules")

	// 测试密码验证
	result := utils.ValidatePasswordWithCustomRules(password, customRules)
	
	fmt.Printf("Validation result:\n")
	fmt.Printf("  Valid: %t\n", result.Valid)
	fmt.Printf("  Score: %d\n", result.Score)
	fmt.Printf("  Strength: %s\n", utils.GetStrengthText(result.Strength))
	fmt.Printf("  Errors: %v\n", result.Errors)
	fmt.Printf("  Warnings: %v\n", result.Warnings)

	if !result.Valid {
		log.Printf("Password validation failed: %s", result.Errors[0])
	} else {
		fmt.Println("Password validation passed!")
	}
}