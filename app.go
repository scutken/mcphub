package main

import (
	"context"
	"fmt"

	"github.com/scutken/mcphub/pkg/hub"
)

// App represents the Wails application, exposing methods to the frontend.
type App struct {
	ctx context.Context
	hub *hub.Hub
}

// NewApp creates a new App instance.
func NewApp(h *hub.Hub) *App {
	return &App{hub: h}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	a.hub.Close()
}

// ==================== Server Management ====================

// ConnectServer adds and connects to an MCP server.
func (a *App) ConnectServer(name, url string, headers map[string]string, transport string) error {
	return a.hub.Connect(name, url, headers, transport)
}

// DisconnectServer disconnects and removes a server.
func (a *App) DisconnectServer(name string) error {
	return a.hub.Disconnect(name)
}

// ListServers returns all configured servers with status.
func (a *App) ListServers() ([]hub.ServerInfo, error) {
	return a.hub.ListServers()
}

// ==================== Tool Operations ====================

// ListTools returns tools from a server (empty name = all servers).
func (a *App) ListTools(serverName string) ([]hub.ToolInfo, error) {
	return a.hub.ListTools(serverName)
}

// CallTool invokes a tool on a server.
func (a *App) CallTool(serverName, toolName string, args map[string]interface{}) (*hub.CallResult, error) {
	return a.hub.CallTool(serverName, toolName, args)
}

// ==================== Utility ====================

// FormatResult formats a CallResult for display.
func (a *App) FormatResult(result *hub.CallResult) string {
	if result == nil {
		return ""
	}
	var s string
	for _, c := range result.Content {
		if c.Type == "text" {
			s += c.Text
		} else {
			s += fmt.Sprintf("[%s]", c.Type)
		}
	}
	return s
}
