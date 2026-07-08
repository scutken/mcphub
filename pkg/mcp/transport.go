package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Transport abstracts the MCP transport mechanism.
type Transport interface {
	// Send sends a JSON-RPC request and returns the response.
	Send(ctx context.Context, req *Request) (*Response, error)
	// Close closes the transport.
	Close() error
}

// TransportType represents the type of transport to use.
type TransportType string

const (
	TransportAuto       TransportType = "auto"
	TransportStreamable TransportType = "streamable"
)

// NewTransport creates a transport based on auto-detection or explicit type.
func NewTransport(baseURL string, headers map[string]string, transportType TransportType) (Transport, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", baseURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q: only http/https are supported", u.Scheme)
	}

	switch transportType {
	case TransportStreamable, TransportAuto:
		return newStreamableTransport(baseURL, headers), nil
	default:
		return nil, fmt.Errorf("unknown transport type: %s", transportType)
	}
}

// ========== Streamable HTTP Transport (2025-11-25) ==========

type streamableTransport struct {
	baseURL   string
	headers   map[string]string
	client    *http.Client
	sessionID string // Mcp-Session-Id from server
	mu        sync.Mutex
}

func newStreamableTransport(baseURL string, headers map[string]string) *streamableTransport {
	return &streamableTransport{
		baseURL: baseURL,
		headers: headers,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *streamableTransport) Send(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	// 携带之前服务器下发的 session ID
	t.mu.Lock()
	if t.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", t.sessionID)
	}
	t.mu.Unlock()

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 捕获服务器返回的 session ID
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		t.mu.Lock()
		t.sessionID = sid
		t.mu.Unlock()
	}

	contentType := resp.Header.Get("Content-Type")

	// 202 Accepted — 通知无需响应体，直接返回空响应
	if resp.StatusCode == http.StatusAccepted {
		return &Response{}, nil
	}

	// If the response is SSE, parse the first event
	if strings.Contains(contentType, "text/event-stream") {
		return t.parseSSEResponse(resp.Body, req.ID)
	}

	// Otherwise, parse as JSON-RPC response
	var rpcResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		// 空 body 对通知是正常的
		if err == io.EOF {
			return &Response{}, nil
		}
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &rpcResp, nil
}

func (t *streamableTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

// parseSSEResponse reads SSE events and returns the first message event as a Response.
func (t *streamableTransport) parseSSEResponse(r io.Reader, expectedID *int64) (*Response, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<20) // 1MB buffer

	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				dataLines = nil

				var rpcResp Response
				if err := json.Unmarshal([]byte(data), &rpcResp); err != nil {
					// Skip non-JSON events
					continue
				}
				// Return first valid JSON-RPC response
				return &rpcResp, nil
			}
			continue
		}

		// Fields: "data: xxx", "event: xxx", "id: xxx", "retry: xxx"
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data:"))
			// Trim single leading space per SSE spec
			dataLines[len(dataLines)-1] = strings.TrimPrefix(dataLines[len(dataLines)-1], " ")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read SSE stream: %w", err)
	}

	return nil, fmt.Errorf("no JSON-RPC response received in SSE stream")
}
