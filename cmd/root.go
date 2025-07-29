package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "envswitch",
	Short: "Environment management and switching tool",
	Long: `envswitch is a CLI tool for managing multiple projects and environments.
It allows you to quickly switch between different environment configurations
by replacing files in your system according to predefined configurations.

Complete documentation is available at https://github.com/zoyopei/envswitch`,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help()
	},
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is ./config.json or ~/.envswitch/config.json)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// 添加子命令
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(serverCmd)
}

// 通用函数
func checkError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
