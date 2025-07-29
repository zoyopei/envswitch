package main

import (
	"fmt"
	"os"

	"github.com/zoyopei/EnvSwitch/cmd"
	"github.com/zoyopei/EnvSwitch/internal/config"
)

func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// 执行命令
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
