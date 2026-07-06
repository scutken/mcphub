package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestMCPClientIntegration tests the full client lifecycle with a mock MCP server.
func TestMCPClientIntegration(t *testing.T) {
	server := newMockMCPServer(t)
	defer server.Close()

	client, err := Connect(context.Background(), server.URL, nil, TransportAuto)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Fatal("client should be connected after Connect")
	}

	if client.ServerInfo() == nil {
		t.Fatal("server info should not be nil")
	}
	if client.ServerInfo().Name != "test-mcp-server" {
		t.Errorf("expected server name 'test-mcp-server', got %q", client.ServerInfo().Name)
	}
}

func TestListTools(t *testing.T) {
	server := newMockMCPServer(t)
	defer server.Close()

	client, err := Connect(context.Background(), server.URL, nil, TransportAuto)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	tools, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	if tools[0].Name != "echo" {
		t.Errorf("expected tool 'echo', got %q", tools[0].Name)
	}
	if tools[1].Name != "add" {
		t.Errorf("expected tool 'add', got %q", tools[1].Name)
	}
}

func TestCallTool(t *testing.T) {
	server := newMockMCPServer(t)
	defer server.Close()

	client, err := Connect(context.Background(), server.URL, nil, TransportAuto)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	result, err := client.CallTool(context.Background(), "echo", map[string]interface{}{
		"message": "hello world",
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
	if result.Content[0].Text != "Echo: hello world" {
		t.Errorf("expected 'Echo: hello world', got %q", result.Content[0].Text)
	}
}

func TestCallToolWithError(t *testing.T) {
	server := newMockMCPServer(t)
	defer server.Close()

	client, err := Connect(context.Background(), server.URL, nil, TransportAuto)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	result, err := client.CallTool(context.Background(), "add", map[string]interface{}{
		"a": 1,
		"b": "not_a_number",
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for invalid input")
	}
}

func TestDisconnect(t *testing.T) {
	server := newMockMCPServer(t)
	defer server.Close()

	client, err := Connect(context.Background(), server.URL, nil, TransportAuto)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if client.IsConnected() {
		t.Fatal("client should be disconnected after Close")
	}

	// Operations after close should fail
	_, err = client.ListTools(context.Background())
	if err == nil {
		t.Fatal("expected error after close")
	}
}

func TestConnectTimeout(t *testing.T) {
	// Connect to a non-routable address to test timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Connect(ctx, "http://10.255.255.1:12345/mcp", nil, TransportAuto)
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestTransportAutoDetection(t *testing.T) {
	// Test that TransportAuto works with a simple JSON-RPC endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			resp := handleJSONRPC(req)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client, err := Connect(context.Background(), server.URL, nil, TransportAuto)
	if err != nil {
		t.Fatalf("Connect with auto transport failed: %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Fatal("client should be connected")
	}
}

// ========== Mock MCP Server ==========

type mockMCPServer struct {
	*httptest.Server
}

func newMockMCPServer(t *testing.T) *mockMCPServer {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if r.Method == http.MethodPost && contentType == "application/json" {
			var req Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			resp := handleJSONRPC(req)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	})

	// Use httptest.NewUnstartedServer so we can set up listener properly
	server := httptest.NewUnstartedServer(handler)
	// Use a real listener instead of the default loopback
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	server.Listener = listener
	server.Start()

	return &mockMCPServer{Server: server}
}

// handleJSONRPC dispatches JSON-RPC requests to mock handlers.
func handleJSONRPC(req Request) *Response {
	switch {
	case req.Method == "initialize":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  mustMarshal(InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: ServerCapabilities{
					Tools: &struct{}{},
				},
				ServerInfo: Implementation{
					Name:    "test-mcp-server",
					Version: "1.0.0",
				},
			}),
		}

	case req.Method == "notifications/initialized":
		return &Response{
			JSONRPC: "2.0",
			Result:  json.RawMessage("{}"),
		}

	case req.Method == "tools/list":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: mustMarshal(ListToolsResult{
				Tools: []Tool{
					{
						Name:        "echo",
						Description: "Echoes back the message",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"message": map[string]interface{}{
									"type":        "string",
									"description": "The message to echo",
								},
							},
							Required: []string{"message"},
						},
					},
					{
						Name:        "add",
						Description: "Adds two numbers",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"a": map[string]interface{}{
									"type":        "number",
									"description": "First number",
								},
								"b": map[string]interface{}{
									"type":        "number",
									"description": "Second number",
								},
							},
							Required: []string{"a", "b"},
						},
					},
				},
			}),
		}

	case req.Method == "tools/call":
		var params CallToolParams
		json.Unmarshal(req.Params, &params)

		switch params.Name {
		case "echo":
			msg, _ := params.Arguments["message"].(string)
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mustMarshal(CallToolResult{
					Content: []Content{
						{Type: "text", Text: fmt.Sprintf("Echo: %s", msg)},
					},
				}),
			}

		case "add":
			a, aOk := toFloat(params.Arguments["a"])
			b, bOk := toFloat(params.Arguments["b"])
			if !aOk || !bOk {
				return &Response{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: mustMarshal(CallToolResult{
						Content: []Content{
							{Type: "text", Text: "Error: arguments 'a' and 'b' must be numbers"},
						},
						IsError: true,
					}),
				}
			}
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: mustMarshal(CallToolResult{
					Content: []Content{
						{Type: "text", Text: fmt.Sprintf("Result: %v", a+b)},
					},
				}),
			}

		default:
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &Error{
					Code:    -32601,
					Message: fmt.Sprintf("Unknown tool: %s", params.Name),
				},
			}
		}

	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
