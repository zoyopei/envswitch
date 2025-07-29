package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/zoyopei/EnvSwitch/internal/config"
	"github.com/zoyopei/EnvSwitch/internal/web"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server",
	Long:  "Start the HTTP web server for managing environments through a web interface",
	Run: func(cmd *cobra.Command, _ []string) {
		port, _ := cmd.Flags().GetInt("port")
		daemon, _ := cmd.Flags().GetBool("daemon")

		if port == 0 {
			port = config.GetWebPort()
		}

		// 创建web服务器
		server := web.NewServer()

		// 设置HTTP服务器
		httpServer := &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: server.SetupRoutes(),
		}

		if daemon {
			fmt.Printf("Starting web server in daemon mode on port %d\n", port)
			// 在实际生产环境中，这里应该实现真正的daemon模式
			// 这里简化处理，仅在后台运行
		} else {
			fmt.Printf("Starting web server on http://localhost:%d\n", port)
			fmt.Println("Press Ctrl+C to stop the server")
		}

		// 启动服务器
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("Failed to start server: %v\n", err)
				os.Exit(1)
			}
		}()

		// 等待中断信号
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c

		fmt.Println("\nShutting down server...")

		// 创建5秒超时的context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 优雅关闭服务器
		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Printf("Server forced to shutdown: %v\n", err)
		} else {
			fmt.Println("Server stopped gracefully")
		}
	},
}

func init() {
	serverCmd.Flags().IntP("port", "p", 0, "Port to run the server on (default from config)")
	serverCmd.Flags().BoolP("daemon", "d", false, "Run server in daemon mode")
}
