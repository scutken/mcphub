package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/scutken/mcphub/cmd/cli"
	"github.com/scutken/mcphub/pkg/config"
	"github.com/scutken/mcphub/pkg/hub"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

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
		// Launch Wails GUI
		app := NewApp(h)

		err := wails.Run(&options.App{
			Title:  "MCPHub - MCP Server Manager",
			Width:  1024,
			Height: 768,
			MinWidth:  800,
			MinHeight: 600,
			AssetServer: &assetserver.Options{
				// 通过 Handler 校验，dev 模式下 Wails 自动替换为 Vite dev server
				// 生产 build 时 Wails 会替换为 embedded assets
				Handler: http.FileServer(http.Dir("frontend/dist")),
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
