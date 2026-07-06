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
					},
				}),
			}

		case req.Method == "tools/call":
			var params mcp.CallToolParams
			json.Unmarshal(req.Params, &params)
			resp = mcp.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mustMarshal(mcp.CallToolResult{
					Content: []mcp.Content{
						{Type: "text", Text: fmt.Sprintf("Hello, %v!", params.Arguments["name"])},
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

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "hello" {
		t.Errorf("expected tool 'hello', got %q", tools[0].Name)
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

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools total, got %d", len(tools))
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
