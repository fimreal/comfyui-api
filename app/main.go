package main

import (
	"log"

	"github.com/fimreal/comfyui-api/src/serve"

	"github.com/spf13/cobra"
)

// version 是当前版本号
var version = "0.1.0"

// main 函数启动 Cobra CLI 应用
func main() {
	var rootCmd = &cobra.Command{
		Use:     "comfyui-cli",
		Version: version,
		Short:   "A CLI tool for interacting with ComfyUI.",
		Run: func(cmd *cobra.Command, args []string) {
			// 启动 HTTP 服务器
			if err := serve.StartServer(); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		},
	}

	// 添加更多命令，比如配置文件路径等
	rootCmd.Flags().StringP("config", "c", "", "Path to configuration file")
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}
