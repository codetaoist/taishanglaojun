package commands

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录到码道平台",
	Long: `通过设备码流程登录到码道平台。

登录后可以使用AI智能体服务、项目管理等功能。
登录信息将安全存储在本地配置文件中。`,
	Run: func(cmd *cobra.Command, args []string) {
		performLogin()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func performLogin() {
	// 创建加载动画
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 正在连接到码道平台..."
	s.Start()
	
	// 模拟登录过程
	time.Sleep(2 * time.Second)
	s.Stop()
	
	// 显示登录信息
	color.Green("✓ 成功连接到码道平台")
	fmt.Println()
	
	// 显示设备码登录信息
	color.Cyan("请在浏览器中访问以下链接完成登录：")
	color.Yellow("https://auth.codetaoist.com/device")
	fmt.Println()
	
	color.Cyan("设备码：")
	color.Yellow("ABCD-EFGH-IJKL")
	fmt.Println()
	
	// 模拟等待用户确认
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 等待用户确认登录..."
	s.Start()
	time.Sleep(3 * time.Second)
	s.Stop()
	
	// 登录成功
	color.Green("✓ 登录成功！")
	fmt.Println()
	
	// 保存登录信息到配置
	viper.Set("auth.token", "mock_jwt_token_12345")
	viper.Set("auth.user_id", "user_123")
	viper.Set("auth.username", "demo@codetaoist.com")
	viper.Set("auth.login_time", time.Now().Format(time.RFC3339))
	
	// 写入配置文件
	if err := viper.WriteConfig(); err != nil {
		if err := viper.SafeWriteConfig(); err != nil {
			color.Red("警告: 无法保存登录信息到配置文件")
		}
	}
	
	color.Green("欢迎使用码道 (Code Taoist)！")
	fmt.Println("现在您可以使用以下命令：")
	fmt.Println("  ct project create <name>  - 创建新项目")
	fmt.Println("  ct \"自然语言指令\"        - AI智能编程")
	fmt.Println("  ct ask \"问题\"            - 查询项目知识库")
}