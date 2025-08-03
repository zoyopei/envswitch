package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/zoyopei/envswitch/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置设置",
	Long:  "查看和修改 envswitch 的配置设置，包括数据目录管理",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()

		fmt.Println("📋 当前配置:")
		fmt.Printf("  数据目录:     %s\n", cfg.DataDir)
		fmt.Printf("  备份目录:     %s\n", cfg.BackupDir)
		fmt.Printf("  Web端口:      %d\n", cfg.WebPort)
		fmt.Printf("  默认项目:     %s\n", cfg.DefaultProject)
		fmt.Printf("  数据目录检查: %t\n", cfg.EnableDataDirCheck)

		if cfg.OriginalDataDir != "" {
			fmt.Printf("  原始数据目录: %s\n", cfg.OriginalDataDir)
		}

		if len(cfg.DataDirHistory) > 0 {
			fmt.Printf("  历史数据目录:\n")
			for i, dir := range cfg.DataDirHistory {
				fmt.Printf("    %d. %s\n", i+1, dir)
			}
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项",
	Long: `设置配置项的值

支持的配置项:
  data_dir        - 数据目录路径
  backup_dir      - 备份目录路径  
  web_port        - Web服务端口
  default_project - 默认项目名称
  enable_data_dir_check - 是否启用数据目录检查 (true/false)`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		updates := make(map[string]interface{})

		switch key {
		case "data_dir":
			updates["data_dir"] = value
		case "backup_dir":
			updates["backup_dir"] = value
		case "web_port":
			var port int
			if _, err := fmt.Sscanf(value, "%d", &port); err != nil {
				fmt.Printf("❌ 错误: web_port 必须是数字\n")
				return
			}
			updates["web_port"] = port
		case "default_project":
			updates["default_project"] = value
		case "enable_data_dir_check":
			enable := strings.ToLower(value) == "true"
			updates["enable_data_dir_check"] = enable
		default:
			fmt.Printf("❌ 错误: 不支持的配置项 '%s'\n", key)
			fmt.Printf("支持的配置项: data_dir, backup_dir, web_port, default_project, enable_data_dir_check\n")
			return
		}

		if err := config.UpdateConfig(updates); err != nil {
			fmt.Printf("❌ 更新配置失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 配置项 '%s' 已更新为 '%s'\n", key, value)
	},
}

var configDataDirMigrateCmd = &cobra.Command{
	Use:   "migrate-datadir <new-directory>",
	Short: "迁移数据目录",
	Long:  "将数据从当前目录迁移到新目录",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		newDataDir := args[0]

		updates := map[string]interface{}{
			"data_dir": newDataDir,
		}

		if err := config.UpdateConfig(updates); err != nil {
			fmt.Printf("❌ 数据目录迁移失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configDataDirMigrateCmd)
}
