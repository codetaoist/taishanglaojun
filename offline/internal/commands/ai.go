package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// aiCmd represents the ai command (for natural language instructions)
var aiCmd = &cobra.Command{
	Use:   "ai [自然语言指令]",
	Short: "AI智能编程助手",
	Long: `使用自然语言与AI智能编程助手交互。

示例:
  ct ai "用Go写一个HTTP服务器"
  ct ai "解释这个函数的作用"
  ct ai "优化当前代码的性能"
  
支持中文和英文指令，AI会根据上下文智能理解并执行。`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		instruction := strings.Join(args, " ")
		processAIInstruction(instruction)
	},
}

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask [问题]",
	Short: "查询项目知识库",
	Long: `向项目知识库提问，获取相关信息。

示例:
  ct ask "项目的数据库配置在哪里？"
  ct ask "如何运行测试？"
  ct ask "API接口文档"
  
AI会从项目代码、文档、历史对话中检索相关信息。`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := strings.Join(args, " ")
		processAskQuestion(question)
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
	rootCmd.AddCommand(askCmd)
}

func processAIInstruction(instruction string) {
	if !isLoggedIn() {
		color.Red("错误: 请先登录 (ct login)")
		return
	}
	
	projectID := viper.GetString("project.id")
	if projectID == "" {
		color.Yellow("提示: 当前目录未绑定项目，将使用通用AI助手")
		fmt.Println("使用 'ct project link <项目ID>' 绑定项目以获得更好的上下文理解")
		fmt.Println()
	}
	
	color.Cyan("🤖 AI助手正在分析您的指令...")
	fmt.Printf("指令: %s\n\n", instruction)
	
	// 创建加载动画
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 正在生成执行计划..."
	s.Start()
	
	// 模拟AI分析过程
	time.Sleep(2 * time.Second)
	s.Stop()
	
	// 显示执行计划
	color.Green("✓ 执行计划生成完成")
	fmt.Println()
	
	color.Cyan("📋 执行计划:")
	plan := generateMockPlan(instruction)
	for i, step := range plan {
		fmt.Printf("  %d. %s\n", i+1, step)
	}
	fmt.Println()
	
	// 询问用户确认
	prompt := promptui.Prompt{
		Label:     "是否执行此计划",
		IsConfirm: true,
	}
	
	result, err := prompt.Run()
	if err != nil || strings.ToLower(result) != "y" {
		color.Yellow("操作已取消")
		return
	}
	
	// 执行计划
	color.Cyan("🚀 开始执行计划...")
	executePlan(plan)
}

func processAskQuestion(question string) {
	if !isLoggedIn() {
		color.Red("错误: 请先登录 (ct login)")
		return
	}
	
	projectID := viper.GetString("project.id")
	if projectID == "" {
		color.Yellow("提示: 当前目录未绑定项目，无法访问项目知识库")
		fmt.Println("使用 'ct project link <项目ID>' 绑定项目")
		return
	}
	
	color.Cyan("🔍 正在搜索项目知识库...")
	fmt.Printf("问题: %s\n\n", question)
	
	// 创建加载动画
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 正在检索相关信息..."
	s.Start()
	
	// 模拟搜索过程
	time.Sleep(1500 * time.Millisecond)
	s.Stop()
	
	// 显示搜索结果
	color.Green("✓ 找到相关信息")
	fmt.Println()
	
	answer := generateMockAnswer(question)
	color.Cyan("📖 答案:")
	fmt.Println(answer)
	fmt.Println()
	
	color.Cyan("📚 相关文件:")
	files := []string{
		"./config/database.go",
		"./docs/api.md",
		"./README.md",
	}
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}
}

func generateMockPlan(instruction string) []string {
	if strings.Contains(strings.ToLower(instruction), "http") || strings.Contains(strings.ToLower(instruction), "服务器") {
		return []string{
			"分析需求：创建HTTP服务器",
			"生成main.go文件，包含基本的HTTP服务器代码",
			"添加路由处理函数",
			"配置端口和中间件",
			"生成示例测试代码",
		}
	}
	
	if strings.Contains(strings.ToLower(instruction), "优化") {
		return []string{
			"扫描当前目录的代码文件",
			"分析性能瓶颈和改进点",
			"生成优化建议报告",
			"应用代码优化",
			"运行性能测试验证",
		}
	}
	
	return []string{
		"理解指令内容和上下文",
		"分析当前项目结构",
		"生成相应的代码或文档",
		"验证生成结果的正确性",
		"保存到适当的文件位置",
	}
}

func executePlan(plan []string) {
	for i, step := range plan {
		color.Cyan("执行步骤 %d: %s", i+1, step)
		
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Start()
		
		// 模拟执行时间
		time.Sleep(time.Duration(1000+i*500) * time.Millisecond)
		
		s.Stop()
		color.Green("  ✓ 完成")
	}
	
	fmt.Println()
	color.Green("🎉 所有步骤执行完成！")
	
	// 显示生成的文件
	color.Cyan("📁 生成的文件:")
	files := []string{
		"./main.go",
		"./handlers/api.go",
		"./config/server.go",
	}
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}
}

func generateMockAnswer(question string) string {
	if strings.Contains(strings.ToLower(question), "数据库") {
		return `项目使用PostgreSQL作为主数据库，配置文件位于 config/database.go。

主要配置参数：
- Host: localhost
- Port: 5432
- Database: codetaoist
- 连接池大小: 20

连接字符串格式：postgres://user:password@localhost:5432/codetaoist`
	}
	
	if strings.Contains(strings.ToLower(question), "测试") {
		return `项目使用Go的标准测试框架。

运行测试命令：
- 运行所有测试: go test ./...
- 运行特定包测试: go test ./pkg/auth
- 生成覆盖率报告: go test -cover ./...

测试文件位于各个包的 *_test.go 文件中。`
	}
	
	return `根据您的问题，我在项目知识库中找到了相关信息。

这是一个基于Go语言开发的智能编程助手项目，采用微服务架构。
主要包含CLI工具、API服务、数据库等组件。

如需更详细的信息，请查看相关文件或使用更具体的问题。`
}