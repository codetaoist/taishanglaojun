package main

import (
	"fmt"
	"os"
)

func debugConfig() {
	// 直接读取环境变量
	devSkipSignature := os.Getenv("DEV_SKIP_SIGNATURE")
	if devSkipSignature == "" {
		devSkipSignature = "true" // 默认值
	}
	
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production" // 默认值
	}
	
	fmt.Printf("DevSkipSignature: %s\n", devSkipSignature)
	fmt.Printf("JWTSecret: %s\n", jwtSecret)
}

func main() {
	debugConfig()
}