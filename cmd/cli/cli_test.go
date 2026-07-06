package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/scutken/mcphub/pkg/config"
	"github.com/scutken/mcphub/pkg/hub"
	"github.com/scutken/mcphub/pkg/mcp"
)

func TestCLIIntegration(t *testing.T) {
	// Start mock MCP server
	mockURL := startMockMCPServer(t)

	// Setup config with temp dir
	dir := t.TempDir()
	storePath := filepath.Join(dir, "servers.json")
	store, err := config.NewStore(storePath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	h := hub.New(store)
	defer h.Close()

	// Test 1: List servers (empty, JSON)
	output := runCommand(t, h, "list")
	var servers []hub.ServerInfo
	if err := json.Unmarshal([]byte(output), &servers); err != nil {
		t.Fatalf("parse list JSON: %v\nOutput: %s", err, output)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}

	// Test 2: Connect
	output = runCommand(t, h, "connect", "test", mockURL)
	var srv hub.ServerInfo
	if err := json.Unmarshal([]byte(output), &srv); err != nil {
		t.Fatalf("parse connect JSON: %v\nOutput: %s", err, output)
	}
	if srv.Status != "connected" {
		t.Errorf("expected connected, got %q", srv.Status)
	}

	// Test 3: List servers (should show 1)
	output = runCommand(t, h, "list")
	if err := json.Unmarshal([]byte(output), &servers); err != nil {
		t.Fatalf("parse list JSON: %v\nOutput: %s", err, output)
	}
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}

	// Test 4: List tools
	output = runCommand(t, h, "tools", "test")
	var tools []hub.ToolInfo
	if err := json.Unmarshal([]byte(output), &tools); err != nil {
		t.Fatalf("parse tools JSON: %v\nOutput: %s", err, output)
	}
	if len(tools) != 1 || tools[0].Name != "hello" {
		t.Errorf("expected 1 tool 'hello', got %+v", tools)
	}

	// Test 5: Call tool
	output = runCommand(t, h, "call", "test", "hello", "--args", `{"name":"World"}`)
	var result hub.CallResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("parse call JSON: %v\nOutput: %s", err, output)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got: %+v", result.Content)
	}

	// Test 6: Disconnect
	output = runCommand(t, h, "disconnect", "test")
	if output == "" {
		t.Fatal("expected disconnect output")
	}

	// Test 7: List after disconnect (should be empty)
	output = runCommand(t, h, "list")
	if err := json.Unmarshal([]byte(output), &servers); err != nil {
		t.Fatalf("parse list JSON: %v\nOutput: %s", err, output)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}

// runCommand creates a fresh root command and executes it, returning stdout as string.
func runCommand(t *testing.T, hub *hub.Hub, args ...string) string {
	t.Helper()

	rootCmd := NewRootCmd(hub)
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("command %v failed: %v", args, err)
	}

	return buf.String()
}

// startMockMCPServer starts a mock MCP server and returns its URL.
func startMockMCPServer(t *testing.T) string {
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
				Result: mustMarshalJSON(mcp.InitializeResult{
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
				Result: mustMarshalJSON(mcp.ListToolsResult{
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
				Result: mustMarshalJSON(mcp.CallToolResult{
					Content: []mcp.Content{
						{Type: "text", Text: fmt.Sprintf("Hello, %v!", params.Arguments["name"])},
					},
				}),
			}

		default:
			resp = mcp.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcp.Error{Code: -32601, Message: "Method not found"},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
