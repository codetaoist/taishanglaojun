package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "项目管理",
	Long: `管理码道平台上的项目。

支持创建、查看、绑定项目等操作。
项目是团队协作和知识管理的基本单位。`,
}

// projectCreateCmd represents the project create command
var projectCreateCmd = &cobra.Command{
	Use:   "create [项目名称]",
	Short: "创建新项目",
	Long: `在码道平台创建新项目。

项目创建后会自动生成项目ID，可用于团队协作和知识管理。`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		createProject(projectName)
	},
}

// projectListCmd represents the project list command
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出项目",
	Long: `列出当前用户有权访问的所有项目。`,
	Run: func(cmd *cobra.Command, args []string) {
		listProjects()
	},
}

// projectLinkCmd represents the project link command
var projectLinkCmd = &cobra.Command{
	Use:   "link [项目ID]",
	Short: "绑定项目",
	Long: `将当前目录与指定项目绑定。

绑定后，在此目录下的所有ct命令都会关联到该项目。`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]
		linkProject(projectID)
	},
}

// projectInfoCmd represents the project info command
var projectInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "查看项目信息",
	Long: `查看当前绑定项目的详细信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		showProjectInfo()
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectLinkCmd)
	projectCmd.AddCommand(projectInfoCmd)
}

func createProject(name string) {
	if !isLoggedIn() {
		color.Red("错误: 请先登录 (ct login)")
		return
	}
	
	color.Cyan("正在创建项目: %s", name)
	
	// 模拟API调用
	time.Sleep(1 * time.Second)
	
	projectID := fmt.Sprintf("proj_%d", time.Now().Unix())
	
	color.Green("✓ 项目创建成功！")
	fmt.Printf("项目ID: %s\n", projectID)
	fmt.Printf("项目名称: %s\n", name)
	fmt.Println()
	fmt.Println("使用以下命令绑定到当前目录:")
	color.Yellow("ct project link %s", projectID)
}

func listProjects() {
	if !isLoggedIn() {
		color.Red("错误: 请先登录 (ct login)")
		return
	}
	
	// 模拟项目数据
	projects := [][]string{
		{"proj_1234567890", "我的第一个项目", "2024-01-15", "活跃"},
		{"proj_1234567891", "Web应用开发", "2024-02-20", "活跃"},
		{"proj_1234567892", "API服务", "2024-03-10", "暂停"},
	}
	
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"项目ID", "项目名称", "创建时间", "状态"})
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	
	for _, project := range projects {
		table.Append(project)
	}
	
	fmt.Println("您的项目列表:")
	table.Render()
}

func linkProject(projectID string) {
	if !isLoggedIn() {
		color.Red("错误: 请先登录 (ct login)")
		return
	}
	
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		color.Red("错误: 无法获取当前目录")
		return
	}
	
	// 保存项目绑定信息
	viper.Set("project.id", projectID)
	viper.Set("project.path", currentDir)
	viper.Set("project.linked_time", time.Now().Format(time.RFC3339))
	
	if err := viper.WriteConfig(); err != nil {
		color.Red("错误: 无法保存项目绑定信息")
		return
	}
	
	color.Green("✓ 项目绑定成功！")
	fmt.Printf("项目ID: %s\n", projectID)
	fmt.Printf("绑定目录: %s\n", currentDir)
}

func showProjectInfo() {
	projectID := viper.GetString("project.id")
	if projectID == "" {
		color.Yellow("当前目录未绑定任何项目")
		fmt.Println("使用 'ct project link <项目ID>' 绑定项目")
		return
	}
	
	color.Cyan("当前项目信息:")
	fmt.Printf("项目ID: %s\n", projectID)
	fmt.Printf("绑定目录: %s\n", viper.GetString("project.path"))
	fmt.Printf("绑定时间: %s\n", viper.GetString("project.linked_time"))
}

func isLoggedIn() bool {
	return viper.GetString("auth.token") != ""
}