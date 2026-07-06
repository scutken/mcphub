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
	TransportSSE        TransportType = "sse"
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
	case TransportStreamable:
		return newStreamableTransport(baseURL, headers), nil
	case TransportSSE:
		return newSSETransport(baseURL, headers)
	case TransportAuto:
		// Try Streamable first, fall back to SSE
		t := newStreamableTransport(baseURL, headers)
		return t, nil
	default:
		return nil, fmt.Errorf("unknown transport type: %s", transportType)
	}
}

// ========== Streamable HTTP Transport (2025-11-25) ==========

type streamableTransport struct {
	baseURL string
	headers map[string]string
	client  *http.Client
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

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	// If the response is SSE, parse the first event
	if strings.Contains(contentType, "text/event-stream") {
		return t.parseSSEResponse(resp.Body, req.ID)
	}

	// Otherwise, parse as JSON-RPC response
	var rpcResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
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
					// Skip non-JSON events (e.g., endpoint events)
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

// ========== SSE Transport (2024-11-05) ==========

type sseTransport struct {
	baseURL   string
	headers   map[string]string
	client    *http.Client
	msgURL    string // The POST endpoint received via 'endpoint' event
	mu        sync.Mutex
	sseResp   *http.Response
	sseCancel context.CancelFunc
	respCh    chan *Response
	errCh     chan error
}

func newSSETransport(baseURL string, headers map[string]string) (*sseTransport, error) {
	// Build SSE endpoint URL
	sseURL := strings.TrimRight(baseURL, "/") + "/sse"

	t := &sseTransport{
		baseURL: baseURL,
		headers: headers,
		client: &http.Client{
			Timeout: 0, // No timeout for long-lived SSE connection
		},
		respCh: make(chan *Response, 100),
		errCh:  make(chan error, 1),
	}

	// Connect to SSE endpoint
	ctx, cancel := context.WithCancel(context.Background())
	t.sseCancel = cancel

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, sseURL, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create SSE request: %w", err)
	}
	httpReq.Header.Set("Accept", "text/event-stream")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("SSE connection failed: %w", err)
	}
	t.sseResp = resp

	// Start reading SSE events in background
	go t.readSSE(resp)

	// Wait for the 'endpoint' event to get the message URL
	// (it should arrive quickly as the first event)
	select {
	case rpcResp := <-t.respCh:
		// This shouldn't be a normal response before endpoint... but handle it
		_ = rpcResp
	case err := <-t.errCh:
		cancel()
		return nil, err
	case <-time.After(10 * time.Second):
		cancel()
		return nil, fmt.Errorf("timeout waiting for SSE endpoint event")
	}

	if t.msgURL == "" {
		cancel()
		return nil, fmt.Errorf("no endpoint event received from SSE server")
	}

	return t, nil
}

func (t *sseTransport) readSSE(resp *http.Response) {
	defer close(t.respCh)
	defer close(t.errCh)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<20)

	var eventType string
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// End of event
			if len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				dataLines = nil

				switch eventType {
				case "endpoint":
					t.mu.Lock()
					t.msgURL = strings.TrimSpace(data)
					t.mu.Unlock()
				case "message":
					var rpcResp Response
					if err := json.Unmarshal([]byte(data), &rpcResp); err != nil {
						continue
					}
					t.respCh <- &rpcResp
				}
				eventType = ""
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			d := strings.TrimPrefix(line, "data:")
			d = strings.TrimPrefix(d, " ")
			dataLines = append(dataLines, d)
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errCh <- err:
		default:
		}
	}
}

func (t *sseTransport) Send(ctx context.Context, req *Request) (*Response, error) {
	t.mu.Lock()
	msgURL := t.msgURL
	t.mu.Unlock()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, msgURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	// For SSE transport, the POST response may contain the result directly
	// or it may come through the SSE stream
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var rpcResp Response
		if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &rpcResp, nil
	}

	// Wait for response from SSE stream
	select {
	case rpcResp := <-t.respCh:
		return rpcResp, nil
	case err := <-t.errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (t *sseTransport) Close() error {
	if t.sseCancel != nil {
		t.sseCancel()
	}
	t.client.CloseIdleConnections()
	return nil
}
