package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long: `管理ct工具的配置参数。

支持查看、设置、删除配置项。
配置文件默认保存在 ~/.ct.yaml`,
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:   "get [配置项]",
	Short: "获取配置值",
	Long: `获取指定配置项的值。

如果不指定配置项，则显示所有配置。`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			showAllConfig()
		} else {
			getConfig(args[0])
		}
	},
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set [配置项] [值]",
	Short: "设置配置值",
	Long: `设置指定配置项的值。

常用配置项：
  api-endpoint    API服务地址
  model          默认AI模型
  verbose        详细输出模式
  timeout        请求超时时间`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		setConfig(args[0], args[1])
	},
}

// configUnsetCmd represents the config unset command
var configUnsetCmd = &cobra.Command{
	Use:   "unset [配置项]",
	Short: "删除配置项",
	Long: `删除指定的配置项。`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		unsetConfig(args[0])
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
}

func showAllConfig() {
	color.Cyan("当前配置:")
	fmt.Println()
	
	// 获取所有配置
	allSettings := viper.AllSettings()
	
	if len(allSettings) == 0 {
		color.Yellow("暂无配置项")
		fmt.Println("使用 'ct config set <key> <value>' 添加配置")
		return
	}
	
	// 创建表格
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"配置项", "值"})
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	
	// 添加配置项到表格
	for key, value := range allSettings {
		// 隐藏敏感信息
		if key == "auth.token" {
			value = "[已设置]"
		}
		table.Append([]string{key, fmt.Sprintf("%v", value)})
	}
	
	table.Render()
	
	fmt.Println()
	color.Cyan("配置文件位置: %s", viper.ConfigFileUsed())
}

func getConfig(key string) {
	value := viper.Get(key)
	
	if value == nil {
		color.Yellow("配置项 '%s' 不存在", key)
		return
	}
	
	// 隐藏敏感信息
	if key == "auth.token" {
		color.Green("%s = [已设置]", key)
	} else {
		color.Green("%s = %v", key, value)
	}
}

func setConfig(key, value string) {
	// 设置配置值
	viper.Set(key, value)
	
	// 保存到配置文件
	if err := viper.WriteConfig(); err != nil {
		// 如果配置文件不存在，创建新的
		if err := viper.SafeWriteConfig(); err != nil {
			color.Red("错误: 无法保存配置 - %v", err)
			return
		}
	}
	
	color.Green("✓ 配置已保存: %s = %s", key, value)
	
	// 特殊配置项的提示
	switch key {
	case "api-endpoint":
		color.Cyan("提示: API端点已更新，下次请求将使用新地址")
	case "model":
		color.Cyan("提示: 默认AI模型已更新为 %s", value)
	case "verbose":
		if value == "true" {
			color.Cyan("提示: 已启用详细输出模式")
		} else {
			color.Cyan("提示: 已禁用详细输出模式")
		}
	}
}

func unsetConfig(key string) {
	// 检查配置项是否存在
	if !viper.IsSet(key) {
		color.Yellow("配置项 '%s' 不存在", key)
		return
	}
	
	// 删除配置项
	allSettings := viper.AllSettings()
	delete(allSettings, key)
	
	// 清空viper并重新设置
	viper.Reset()
	for k, v := range allSettings {
		viper.Set(k, v)
	}
	
	// 保存配置文件
	if err := viper.WriteConfig(); err != nil {
		color.Red("错误: 无法保存配置 - %v", err)
		return
	}
	
	color.Green("✓ 配置项已删除: %s", key)
	
	// 特殊配置项的警告
	switch key {
	case "auth.token":
		color.Yellow("警告: 认证令牌已删除，需要重新登录")
	case "api-endpoint":
		color.Cyan("提示: API端点已重置为默认值")
	}
}