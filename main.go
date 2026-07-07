package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/scutken/mcphub/cmd/cli"
	"github.com/scutken/mcphub/pkg/config"
	"github.com/scutken/mcphub/pkg/hub"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check mode early for console attachment on Windows
	args := os.Args[1:]
	// 无参数（wails dev / 直接双击）视为 GUI 模式，显式 "serve" 也进 GUI
	isServe := len(args) == 0 || args[0] == "serve"

	// On Windows GUI subsystem, attach to parent console for CLI mode
	if !isServe {
		attachConsole()
	}

	// Initialize config store
	store, err := config.NewDefaultStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Initialize hub
	h := hub.New(store)
	defer h.Close()

	// Try to reconnect to previously connected servers (non-fatal)
	if err := h.ReconnectAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: reconnect failed: %v\n", err)
	}

	// Build CLI root command
	rootCmd := cli.NewRootCmd(h)

	if isServe {
		// 进入 GUI 模式，隐藏控制台窗口（console 子系统构建会弹控制台）
		hideConsole()

		// Launch Wails GUI
		app := NewApp(h)

		err := wails.Run(&options.App{
			Title:  "MCPHub - MCP Server Manager",
			Width:  1024,
			Height: 768,
			MinWidth:  800,
			MinHeight: 600,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			OnStartup:  app.startup,
			OnShutdown: app.shutdown,
			Bind: []interface{}{
				app,
			},
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run as CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
