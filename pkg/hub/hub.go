package hub

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/scutken/mcphub/pkg/config"
	"github.com/scutken/mcphub/pkg/mcp"
)

// ServerInfo combines config and runtime state for a server.
type ServerInfo struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Transport string `json:"transport"`
	Status    string `json:"status"` // "connected", "disconnected", "error"
	Error     string `json:"error,omitempty"`
	AddedAt   string `json:"added_at"`
}

// ToolInfo represents a tool from a server.
type ToolInfo struct {
	Server      string          `json:"server"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema mcp.InputSchema `json:"inputSchema"`
}

// ToolSummary 是工具的摘要信息，用于渐进式披露。
type ToolSummary struct {
	Server      string `json:"server"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CallResult is the result of a tool call.
type CallResult struct {
	Server  string        `json:"server"`
	Tool    string        `json:"tool"`
	IsError bool          `json:"isError"`
	Content []mcp.Content `json:"content"`
}

// Hub is the unified service layer for managing MCP connections.
// It is used by both the CLI and GUI.
type Hub struct {
	config  *config.Store
	clients map[string]*mcp.Client
	mu      sync.RWMutex
}

// New creates a new Hub with the given config store.
func New(store *config.Store) *Hub {
	return &Hub{
		config:  store,
		clients: make(map[string]*mcp.Client),
	}
}

// Connect adds and connects to an MCP server.
func (h *Hub) Connect(name, url string, headers map[string]string, transport string) error {
	// Validate URL scheme
	if url == "" {
		return fmt.Errorf("URL is required")
	}

	// Save to config first
	server := config.Server{
		Name:      name,
		URL:       url,
		Headers:   headers,
		Transport: transport,
		AddedAt:   time.Now(),
	}

	if err := h.config.AddServer(server); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Connect
	transportType := mcp.TransportAuto
	if transport == "streamable" {
		transportType = mcp.TransportStreamable
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mcp.Connect(ctx, url, headers, transportType)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}

	h.mu.Lock()
	// Close any existing connection for this server
	if old, ok := h.clients[name]; ok {
		old.Close()
	}
	h.clients[name] = client
	h.mu.Unlock()

	return nil
}

// Disconnect removes a server connection.
func (h *Hub) Disconnect(name string) error {
	h.mu.Lock()
	if client, ok := h.clients[name]; ok {
		client.Close()
		delete(h.clients, name)
	}
	h.mu.Unlock()

	return h.config.RemoveServer(name)
}

// getClient returns the connected client for a server, or an error.
func (h *Hub) getClient(name string) (*mcp.Client, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[name]
	if !ok {
		return nil, fmt.Errorf("server %q is not connected", name)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("server %q is disconnected", name)
	}

	return client, nil
}

// ListServers returns all configured servers with their runtime status.
func (h *Hub) ListServers() ([]ServerInfo, error) {
	servers, err := h.config.ListServers()
	if err != nil {
		return nil, err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]ServerInfo, 0, len(servers))
	for _, s := range servers {
		info := ServerInfo{
			Name:      s.Name,
			URL:       s.URL,
			Transport: s.Transport,
			AddedAt:   s.AddedAt.Format(time.RFC3339),
			Status:    "disconnected",
		}

		if client, ok := h.clients[s.Name]; ok && client.IsConnected() {
			info.Status = "connected"
		} else if _, ok := h.clients[s.Name]; ok {
			info.Status = "error"
			info.Error = "connection lost"
		}

		result = append(result, info)
	}

	return result, nil
}

// ListTools returns tools from the specified server, or all servers if server is empty.
func (h *Hub) ListTools(serverName string) ([]ToolInfo, error) {
	if serverName != "" {
		client, err := h.getClient(serverName)
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tools, err := client.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tools from %q: %w", serverName, err)
		}

		var result []ToolInfo
		for _, t := range tools {
			result = append(result, ToolInfo{
				Server:      serverName,
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
		return result, nil
	}

	// List all servers' tools
	servers, err := h.config.ListServers()
	if err != nil {
		return nil, err
	}

	var allTools []ToolInfo
	for _, s := range servers {
		client, err := h.getClient(s.Name)
		if err != nil {
			continue // Skip disconnected servers
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		tools, err := client.ListTools(ctx)
		cancel()
		if err != nil {
			continue
		}

		for _, t := range tools {
			allTools = append(allTools, ToolInfo{
				Server:      s.Name,
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
	}

	return allTools, nil
}

// GetTool returns the full ToolInfo (including InputSchema) for a single tool.
// 用于渐进式披露：tools 命令只展示名称+描述，确定工具后用本方法取完整 schema。
func (h *Hub) GetTool(serverName, toolName string) (*ToolInfo, error) {
	tools, err := h.ListTools(serverName)
	if err != nil {
		return nil, err
	}
	for i := range tools {
		if tools[i].Name == toolName {
			return &tools[i], nil
		}
	}
	return nil, fmt.Errorf("tool %q not found on server %q", toolName, serverName)
}

// ListToolSummaries 返回摘要列表（不含 inputSchema），用于渐进式披露。
// serverName 为空时返回所有已连接服务器的工具摘要。
func (h *Hub) ListToolSummaries(serverName string) ([]ToolSummary, error) {
	if serverName != "" {
		client, err := h.getClient(serverName)
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tools, err := client.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tool summaries from %q: %w", serverName, err)
		}

		result := make([]ToolSummary, 0, len(tools))
		for _, t := range tools {
			result = append(result, ToolSummary{
				Server:      serverName,
				Name:        t.Name,
				Description: t.Description,
			})
		}
		return result, nil
	}

	// 所有已连接服务器
	servers, err := h.config.ListServers()
	if err != nil {
		return nil, err
	}

	all := make([]ToolSummary, 0)
	for _, s := range servers {
		client, err := h.getClient(s.Name)
		if err != nil {
			continue // 跳过未连接的服务器
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		tools, err := client.ListTools(ctx)
		cancel()
		if err != nil {
			continue
		}

		for _, t := range tools {
			all = append(all, ToolSummary{
				Server:      s.Name,
				Name:        t.Name,
				Description: t.Description,
			})
		}
	}

	return all, nil
}

// SearchToolSummaries 按关键字搜索工具摘要（Name 或 Description 包含关键字，大小写不敏感）。
// serverName 为空时搜索所有已连接服务器。
func (h *Hub) SearchToolSummaries(serverName, keyword string) ([]ToolSummary, error) {
	summaries, err := h.ListToolSummaries(serverName)
	if err != nil {
		return nil, err
	}

	keyword = strings.ToLower(keyword)
	matched := make([]ToolSummary, 0)
	for _, s := range summaries {
		if strings.Contains(strings.ToLower(s.Name), keyword) ||
			strings.Contains(strings.ToLower(s.Description), keyword) {
			matched = append(matched, s)
		}
	}

	return matched, nil
}

// GetTools 返回指定工具名的完整 ToolInfo（含 InputSchema）。
// 保持 toolNames 的原始顺序。若任意工具未找到则返回错误。
func (h *Hub) GetTools(serverName string, toolNames []string) ([]ToolInfo, error) {
	client, err := h.getClient(serverName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tools from %q: %w", serverName, err)
	}

	// 建立 name → ToolInfo 索引
	toolMap := make(map[string]mcp.Tool, len(tools))
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	result := make([]ToolInfo, 0, len(toolNames))
	for _, name := range toolNames {
		t, ok := toolMap[name]
		if !ok {
			return nil, fmt.Errorf("tool %q not found on server %q", name, serverName)
		}
		result = append(result, ToolInfo{
			Server:      serverName,
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	return result, nil
}

// CallTool invokes a tool on the specified server.
func (h *Hub) CallTool(serverName, toolName string, args map[string]interface{}) (*CallResult, error) {
	client, err := h.getClient(serverName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := client.CallTool(ctx, toolName, args)
	if err != nil {
		return nil, fmt.Errorf("call tool %q on %q: %w", toolName, serverName, err)
	}

	return &CallResult{
		Server:  serverName,
		Tool:    toolName,
		IsError: result.IsError,
		Content: result.Content,
	}, nil
}

// ReconnectAll attempts to reconnect all configured servers at startup.
func (h *Hub) ReconnectAll() error {
	servers, err := h.config.ListServers()
	if err != nil {
		return err
	}

	for _, s := range servers {
		transportType := mcp.TransportAuto
		if s.Transport == "streamable" {
			transportType = mcp.TransportStreamable
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		client, err := mcp.Connect(ctx, s.URL, s.Headers, transportType)
		cancel()

		if err != nil {
			continue // Silently skip offline servers
		}

		h.mu.Lock()
		h.clients[s.Name] = client
		h.mu.Unlock()
	}

	return nil
}

// Close disconnects all servers.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, client := range h.clients {
		client.Close()
		delete(h.clients, name)
	}
}
