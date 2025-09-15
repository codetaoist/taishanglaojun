package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	Long: `退出当前登录状态，清除本地认证信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		performLogout()
	},
}

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "查看操作历史",
	Long: `查看与AI的对话历史和操作记录。

支持按时间、关键词筛选历史记录。`,
	Run: func(cmd *cobra.Command, args []string) {
		showHistory()
	},
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看状态",
	Long: `查看ct工具的当前状态，包括登录状态、项目绑定等信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看版本信息",
	Long: `显示ct工具的版本信息和构建详情。`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}

func performLogout() {
	if !isLoggedIn() {
		color.Yellow("您尚未登录")
		return
	}
	
	username := viper.GetString("auth.username")
	
	// 清除认证信息
	viper.Set("auth.token", "")
	viper.Set("auth.user_id", "")
	viper.Set("auth.username", "")
	viper.Set("auth.login_time", "")
	
	// 保存配置
	if err := viper.WriteConfig(); err != nil {
		color.Red("警告: 无法清除本地认证信息")
	}
	
	color.Green("✓ 已退出登录")
	if username != "" {
		fmt.Printf("用户: %s\n", username)
	}
	fmt.Println("感谢使用码道 (Code Taoist)！")
}

func showHistory() {
	if !isLoggedIn() {
		color.Red("错误: 请先登录 (ct login)")
		return
	}
	
	color.Cyan("操作历史:")
	fmt.Println()
	
	// 模拟历史数据
	history := [][]string{
		{"2024-01-15 10:30", "ai", "用Go写一个HTTP服务器", "成功"},
		{"2024-01-15 11:15", "ask", "项目的数据库配置在哪里？", "成功"},
		{"2024-01-15 14:20", "project create", "创建项目: Web应用开发", "成功"},
		{"2024-01-15 15:45", "ai", "优化当前代码的性能", "成功"},
		{"2024-01-15 16:30", "ask", "如何运行测试？", "成功"},
	}
	
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "命令", "内容", "状态"})
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	
	for _, record := range history {
		table.Append(record)
	}
	
	table.Render()
	
	fmt.Println()
	color.Cyan("使用 'ct history --filter <关键词>' 筛选历史记录")
}

func showStatus() {
	color.Cyan("📊 码道 (Code Taoist) 状态")
	fmt.Println()
	
	// 登录状态
	if isLoggedIn() {
		color.Green("✓ 已登录")
		fmt.Printf("  用户: %s\n", viper.GetString("auth.username"))
		fmt.Printf("  登录时间: %s\n", viper.GetString("auth.login_time"))
	} else {
		color.Red("✗ 未登录")
		fmt.Println("  使用 'ct login' 登录")
	}
	fmt.Println()
	
	// 项目状态
	projectID := viper.GetString("project.id")
	if projectID != "" {
		color.Green("✓ 已绑定项目")
		fmt.Printf("  项目ID: %s\n", projectID)
		fmt.Printf("  绑定目录: %s\n", viper.GetString("project.path"))
		fmt.Printf("  绑定时间: %s\n", viper.GetString("project.linked_time"))
	} else {
		color.Yellow("○ 未绑定项目")
		fmt.Println("  使用 'ct project link <项目ID>' 绑定项目")
	}
	fmt.Println()
	
	// 配置状态
	color.Cyan("⚙️  配置信息")
	fmt.Printf("  API端点: %s\n", viper.GetString("api-endpoint"))
	fmt.Printf("  详细输出: %v\n", viper.GetBool("verbose"))
	fmt.Printf("  配置文件: %s\n", viper.ConfigFileUsed())
	fmt.Println()
	
	// 系统信息
	color.Cyan("💻 系统信息")
	currentDir, _ := os.Getwd()
	fmt.Printf("  当前目录: %s\n", currentDir)
	fmt.Printf("  工作目录: %s\n", os.Getenv("PWD"))
}

func showVersion() {
	color.Cyan("📦 码道 (Code Taoist) 版本信息")
	fmt.Println()
	
	fmt.Printf("版本: %s\n", "1.0.0")
	fmt.Printf("构建时间: %s\n", "2024-01-15 12:00:00")
	fmt.Printf("Git提交: %s\n", "abc123def456")
	fmt.Printf("Go版本: %s\n", "go1.22.0")
	fmt.Println()
	
	color.Cyan("🌐 相关链接")
	fmt.Println("  官网: https://codetaoist.com")
	fmt.Println("  文档: https://docs.codetaoist.com")
	fmt.Println("  GitHub: https://github.com/codetaoist/ct")
	fmt.Println("  问题反馈: https://github.com/codetaoist/ct/issues")
	fmt.Println()
	
	color.Green("感谢使用码道 (Code Taoist)！")
}