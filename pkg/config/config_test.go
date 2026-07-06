package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Verify it has default content
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestAddServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	err = store.AddServer(Server{
		Name: "github",
		URL:  "https://api.github.com/mcp",
	})
	if err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}

	s := cfg.Servers[0]
	if s.Name != "github" {
		t.Errorf("expected name 'github', got %q", s.Name)
	}
	if s.URL != "https://api.github.com/mcp" {
		t.Errorf("expected URL 'https://api.github.com/mcp', got %q", s.URL)
	}
	if s.Transport != "auto" {
		t.Errorf("expected transport 'auto', got %q", s.Transport)
	}
	if s.AddedAt.IsZero() {
		t.Error("expected AddedAt to be set")
	}
}

func TestAddDuplicateServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	err = store.AddServer(Server{Name: "test", URL: "http://example.com/mcp"})
	if err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	err = store.AddServer(Server{Name: "test", URL: "http://example.com/mcp"})
	if err == nil {
		t.Fatal("expected error for duplicate server name")
	}
}

func TestRemoveServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	store.AddServer(Server{Name: "server1", URL: "http://example.com/1"})
	store.AddServer(Server{Name: "server2", URL: "http://example.com/2"})

	err = store.RemoveServer("server1")
	if err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].Name != "server2" {
		t.Errorf("expected server2, got %q", cfg.Servers[0].Name)
	}
}

func TestRemoveNonExistentServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	err = store.RemoveServer("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}

func TestGetServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	store.AddServer(Server{
		Name: "github",
		URL:  "https://api.github.com/mcp",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
		},
	})

	s, err := store.GetServer("github")
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}

	if s.Name != "github" {
		t.Errorf("expected name 'github', got %q", s.Name)
	}
	if s.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("expected Authorization header 'Bearer token123', got %q", s.Headers["Authorization"])
	}

	_, err = store.GetServer("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}

func TestConfigPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	// Create store and add server
	store1, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	store1.AddServer(Server{Name: "test", URL: "http://example.com/mcp"})

	// Open a new store at the same path
	store2, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore (2) failed: %v", err)
	}

	cfg, err := store2.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].Name != "test" {
		t.Errorf("expected server 'test', got %q", cfg.Servers[0].Name)
	}
}
