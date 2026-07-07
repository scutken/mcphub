package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/scutken/mcphub/pkg/config"
	"github.com/scutken/mcphub/pkg/mcp"
)

// startMockServer starts a mock MCP server and returns its URL.
func startMockServer(t *testing.T) string {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req mcp.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var resp mcp.Response

		switch {
		case req.Method == "initialize":
			resp = mcp.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mustMarshal(mcp.InitializeResult{
					ProtocolVersion: "2024-11-05",
					Capabilities:    mcp.ServerCapabilities{Tools: &struct{}{}},
					ServerInfo: mcp.Implementation{
						Name:    "mock-server",
						Version: "1.0.0",
					},
				}),
			}

		case req.Method == "notifications/initialized":
			resp = mcp.Response{
				JSONRPC: "2.0",
				Result:  json.RawMessage("{}"),
			}

		case req.Method == "tools/list":
			resp = mcp.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mustMarshal(mcp.ListToolsResult{
				Tools: []mcp.Tool{
					{
						Name:        "hello",
						Description: "Says hello",
						InputSchema: mcp.InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"name": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
					{
						Name:        "goodbye",
						Description: "Says goodbye",
						InputSchema: mcp.InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"name": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				}),
			}

		case req.Method == "tools/call":
			var params mcp.CallToolParams
			json.Unmarshal(req.Params, &params)
			text := fmt.Sprintf("Hello, %v!", params.Arguments["name"])
			if params.Name == "goodbye" {
				text = fmt.Sprintf("Goodbye, %v!", params.Arguments["name"])
			}
			resp = mcp.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mustMarshal(mcp.CallToolResult{
					Content: []mcp.Content{
						{Type: "text", Text: text},
					},
				}),
			}

		default:
			resp = mcp.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &mcp.Error{
					Code: -32601, Message: "Method not found",
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}

func TestHubConnectAndListServers(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	hub := New(store)
	defer hub.Close()

	err = hub.Connect("test-server", serverURL, nil, "auto")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	servers, err := hub.ListServers()
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}

	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}

	if servers[0].Status != "connected" {
		t.Errorf("expected status 'connected', got %q", servers[0].Status)
	}
}

func TestHubListTools(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	hub := New(store)
	defer hub.Close()

	err = hub.Connect("test-server", serverURL, nil, "auto")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	tools, err := hub.ListTools("test-server")
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	if tools[0].Name != "hello" {
		t.Errorf("expected first tool 'hello', got %q", tools[0].Name)
	}
	if tools[0].Server != "test-server" {
		t.Errorf("expected server 'test-server', got %q", tools[0].Server)
	}
}

func TestHubCallTool(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	hub := New(store)
	defer hub.Close()

	err = hub.Connect("test-server", serverURL, nil, "auto")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	result, err := hub.CallTool("test-server", "hello", map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if result.IsError {
		t.Fatal("expected IsError=false")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Text != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", result.Content[0].Text)
	}
}

func TestHubDisconnect(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	hub := New(store)
	defer hub.Close()

	err = hub.Connect("test-server", serverURL, nil, "auto")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = hub.Disconnect("test-server")
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	servers, err := hub.ListServers()
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}

	if len(servers) != 0 {
		t.Fatalf("expected 0 servers after disconnect, got %d", len(servers))
	}
}

func TestHubCallToolDisconnected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	hub := New(store)
	defer hub.Close()

	_, err = hub.CallTool("nonexistent", "tool", nil)
	if err == nil {
		t.Fatal("expected error calling tool on nonexistent server")
	}
}

func TestHubListAllTools(t *testing.T) {
	serverURL1 := startMockServer(t)
	serverURL2 := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	hub := New(store)
	defer hub.Close()

	hub.Connect("server1", serverURL1, nil, "auto")
	hub.Connect("server2", serverURL2, nil, "auto")

	// List all tools (empty server name = all servers)
	tools, err := hub.ListTools("")
	if err != nil {
		t.Fatalf("ListTools(all) failed: %v", err)
	}

	if len(tools) != 4 {
		t.Fatalf("expected 4 tools total, got %d", len(tools))
	}
}

func TestHubListToolSummaries(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	h := New(store)
	defer h.Close()

	if err := h.Connect("test-server", serverURL, nil, "auto"); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	summaries, err := h.ListToolSummaries("test-server")
	if err != nil {
		t.Fatalf("ListToolSummaries failed: %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].Name != "hello" {
		t.Errorf("expected first name 'hello', got %q", summaries[0].Name)
	}
	if summaries[0].Server != "test-server" {
		t.Errorf("expected server 'test-server', got %q", summaries[0].Server)
	}
	if summaries[0].Description != "Says hello" {
		t.Errorf("expected description 'Says hello', got %q", summaries[0].Description)
	}
}

func TestHubSearchToolSummaries(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	h := New(store)
	defer h.Close()

	if err := h.Connect("test-server", serverURL, nil, "auto"); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// 搜索匹配的关键字
	matched, err := h.SearchToolSummaries("test-server", "hello")
	if err != nil {
		t.Fatalf("SearchToolSummaries failed: %v", err)
	}
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
	if matched[0].Name != "hello" {
		t.Errorf("expected name 'hello', got %q", matched[0].Name)
	}

	// 搜索不匹配的关键字
	matched, err = h.SearchToolSummaries("test-server", "nonexistent")
	if err != nil {
		t.Fatalf("SearchToolSummaries failed: %v", err)
	}
	if len(matched) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matched))
	}

	// 大小写不敏感搜索
	matched, err = h.SearchToolSummaries("test-server", "HELLO")
	if err != nil {
		t.Fatalf("SearchToolSummaries failed: %v", err)
	}
	if len(matched) != 1 {
		t.Errorf("expected 1 case-insensitive match, got %d", len(matched))
	}

	// 搜索描述（只匹配 goodbye）
	matched, err = h.SearchToolSummaries("test-server", "goodbye")
	if err != nil {
		t.Fatalf("SearchToolSummaries failed: %v", err)
	}
	if len(matched) != 1 {
		t.Errorf("expected 1 description match, got %d", len(matched))
	}
	if matched[0].Name != "goodbye" {
		t.Errorf("expected match 'goodbye', got %q", matched[0].Name)
	}
}

func TestHubGetTools(t *testing.T) {
	serverURL := startMockServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	h := New(store)
	defer h.Close()

	if err := h.Connect("test-server", serverURL, nil, "auto"); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// 批量获取存在的工具
	tools, err := h.GetTools("test-server", []string{"hello", "goodbye"})
	if err != nil {
		t.Fatalf("GetTools failed: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	if tools[0].Name != "hello" {
		t.Errorf("expected first name 'hello', got %q", tools[0].Name)
	}
	if tools[1].Name != "goodbye" {
		t.Errorf("expected second name 'goodbye', got %q", tools[1].Name)
	}
	if tools[0].InputSchema.Type != "object" {
		t.Errorf("expected inputSchema.type 'object', got %q", tools[0].InputSchema.Type)
	}

	// 获取单个工具
	tools, err = h.GetTools("test-server", []string{"hello"})
	if err != nil {
		t.Fatalf("GetTools single failed: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "hello" {
		t.Errorf("expected single tool 'hello', got %+v", tools)
	}

	// 获取不存在的工具
	_, err = h.GetTools("test-server", []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent tool")
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
