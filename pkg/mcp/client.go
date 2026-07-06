package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// Client is an MCP client that communicates with an MCP server over HTTP.
type Client struct {
	transport Transport
	serverInfo *Implementation
	mu         sync.Mutex
	requestID  atomic.Int64
	connected  bool
}

// Connect creates a new MCP client and performs the initialization handshake.
func Connect(ctx context.Context, serverURL string, headers map[string]string, transportType TransportType) (*Client, error) {
	t, err := NewTransport(serverURL, headers, transportType)
	if err != nil {
		return nil, fmt.Errorf("create transport: %w", err)
	}

	c := &Client{
		transport: t,
	}

	// Perform initialization handshake
	if err := c.initialize(ctx); err != nil {
		t.Close()
		return nil, fmt.Errorf("initialize: %w", err)
	}

	c.connected = true
	return c, nil
}

// initialize performs the MCP initialization handshake.
func (c *Client) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    ClientCapabilities{},
		ClientInfo: Implementation{
			Name:    "mcphub",
			Version: "1.0.0",
		},
	}

	resp, err := c.send(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("parse initialize result: %w", err)
	}
	c.serverInfo = &result.ServerInfo

	// Send the 'initialized' notification (no response expected)
	_, err = c.sendNotification(ctx, "notifications/initialized", nil)
	if err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	return nil
}

// send sends a JSON-RPC request and returns the response.
func (c *Client) send(ctx context.Context, method string, params interface{}) (*Response, error) {
	id := c.requestID.Add(1)
	req, err := NewRequest(id, method, params)
	if err != nil {
		return nil, err
	}
	return c.transport.Send(ctx, req)
}

// sendNotification sends a JSON-RPC notification (no response expected).
func (c *Client) sendNotification(ctx context.Context, method string, params interface{}) (*Response, error) {
	req, err := NewNotification(method, params)
	if err != nil {
		return nil, err
	}
	return c.transport.Send(ctx, req)
}

// ServerInfo returns the server's implementation info from the handshake.
func (c *Client) ServerInfo() *Implementation {
	return c.serverInfo
}

// IsConnected returns true if the client completed initialization.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// ListTools retrieves the list of tools from the server.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is disconnected")
	}
	resp, err := c.send(ctx, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list failed: %w", err)
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var result ListToolsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse tools/list result: %w", err)
	}

	return result.Tools, nil
}

// CallTool invokes a tool on the server with the given arguments.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is disconnected")
	}
	params := CallToolParams{
		Name:      name,
		Arguments: args,
	}

	resp, err := c.send(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call failed: %w", err)
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var result CallToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse tools/call result: %w", err)
	}

	return &result, nil
}

// Close disconnects from the server.
func (c *Client) Close() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	return c.transport.Close()
}
