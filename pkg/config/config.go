package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultConfigDir returns the default config directory (%USERPROFILE%\.mcphub or ~/.mcphub).
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".mcphub"
	}
	return filepath.Join(home, ".mcphub")
}

// Server represents a configured MCP server connection.
type Server struct {
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Transport string            `json:"transport"` // "auto", "sse", "streamable"
	AddedAt   time.Time         `json:"added_at"`
}

// ServerConfig is the top-level configuration file.
type ServerConfig struct {
	Version int      `json:"version"`
	Servers []Server `json:"servers"`
}

// Store manages persistent server configuration.
type Store struct {
	path string
	mu   sync.RWMutex
}

// NewStore creates a new config store at the given path.
func NewStore(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	s := &Store{path: path}

	// Create default config if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultCfg := &ServerConfig{
			Version: 1,
			Servers: []Server{},
		}
		if err := s.write(defaultCfg); err != nil {
			return nil, fmt.Errorf("create default config: %w", err)
		}
	}

	return s, nil
}

// NewDefaultStore creates a store at the default config path.
func NewDefaultStore() (*Store, error) {
	path := filepath.Join(DefaultConfigDir(), "servers.json")
	return NewStore(path)
}

// Load reads the config from disk.
func (s *Store) Load() (*ServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.read()
}

func (s *Store) read() (*ServerConfig, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func (s *Store) write(cfg *ServerConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// AddServer adds a server connection to the config.
func (s *Store) AddServer(server Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil {
		return err
	}

	// Check for duplicate name
	for _, existing := range cfg.Servers {
		if existing.Name == server.Name {
			return fmt.Errorf("server %q already exists", server.Name)
		}
	}

	if server.Transport == "" {
		server.Transport = "auto"
	}
	if server.AddedAt.IsZero() {
		server.AddedAt = time.Now()
	}

	cfg.Servers = append(cfg.Servers, server)
	return s.write(cfg)
}

// RemoveServer removes a server connection by name.
func (s *Store) RemoveServer(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil {
		return err
	}

	for i, existing := range cfg.Servers {
		if existing.Name == name {
			cfg.Servers = append(cfg.Servers[:i], cfg.Servers[i+1:]...)
			return s.write(cfg)
		}
	}

	return fmt.Errorf("server %q not found", name)
}

// ListServers returns all configured servers.
func (s *Store) ListServers() ([]Server, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.read()
	if err != nil {
		return nil, err
	}

	return cfg.Servers, nil
}

// GetServer returns a specific server by name.
func (s *Store) GetServer(name string) (*Server, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.read()
	if err != nil {
		return nil, err
	}

	for i := range cfg.Servers {
		if cfg.Servers[i].Name == name {
			return &cfg.Servers[i], nil
		}
	}

	return nil, fmt.Errorf("server %q not found", name)
}

// Path returns the config file path.
func (s *Store) Path() string {
	return s.path
}
