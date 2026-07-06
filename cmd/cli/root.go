package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/scutken/mcphub/pkg/hub"

	"github.com/spf13/cobra"
)

var h *hub.Hub

// NewRootCmd creates the root CLI command.
func NewRootCmd(hubInstance *hub.Hub) *cobra.Command {
	h = hubInstance

	root := &cobra.Command{
		Use:   "mcphub",
		Short: "MCPHub - Manage and call MCP (Model Context Protocol) servers",
		Long: `MCPHub is a CLI tool for managing connections to MCP servers over HTTP.

It allows you to:
  - Connect to MCP servers (HTTP SSE / Streamable)
  - List available tools from connected servers
  - Call tools with arguments

All commands output JSON by default for easy integration with AI agents.
Use --json=false for human-readable output.`,
	}

	root.PersistentFlags().BoolP("json", "j", true, "Output in JSON format")

	root.AddCommand(newConnectCmd())
	root.AddCommand(newDisconnectCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newToolsCmd())
	root.AddCommand(newCallCmd())

	return root
}

// isJSON returns whether JSON output is requested.
func isJSON(cmd *cobra.Command) bool {
	f := cmd.Flag("json")
	if f != nil {
		return f.Value.String() == "true"
	}
	return true // default
}

// printOutput prints data as JSON or a formatted string to the command's output.
func printOutput(cmd *cobra.Command, v interface{}, formatter func() string) error {
	if isJSON(cmd) {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	fmt.Fprintln(cmd.OutOrStdout(), formatter())
	return nil
}

// ==================== connect ====================

func newConnectCmd() *cobra.Command {
	var headers []string
	var transport string

	cmd := &cobra.Command{
		Use:   "connect <name> <url>",
		Short: "Connect to an MCP server",
		Long: `Connect to an MCP server over HTTP and save the connection.

The server is immediately connected and a handshake is performed.
Transport is auto-detected (Streamable HTTP first, then SSE).

Examples:
  mcphub connect my-server https://mcp.example.com
  mcphub connect github https://api.github.com/mcp --header "Authorization: Bearer ghp_xxx"
  mcphub connect server3 https://example.com --transport sse`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]

			headerMap := make(map[string]string)
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header format: %q (expected 'Key: Value')", h)
				}
				headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}

			if err := h.Connect(name, url, headerMap, transport); err != nil {
				return err
			}

			// Show server info
			servers, err := h.ListServers()
			if err != nil {
				return err
			}
			for _, s := range servers {
				if s.Name == name {
					return printOutput(cmd, s, func() string {
						return fmt.Sprintf("✓ Connected to %s (%s) — status: %s", s.Name, s.URL, s.Status)
					})
				}
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "HTTP header (Key: Value). Can be specified multiple times.")
	cmd.Flags().StringVarP(&transport, "transport", "t", "auto", "Transport type: auto, sse, streamable")

	return cmd
}

// ==================== disconnect ====================

func newDisconnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect <name>",
		Short: "Disconnect and remove an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := h.Disconnect(name); err != nil {
				return err
			}

			return printOutput(cmd, map[string]string{"status": "disconnected", "server": name}, func() string {
				return fmt.Sprintf("✓ Disconnected %s", name)
			})
		},
	}

	return cmd
}

// ==================== list / servers ====================

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"servers", "ls"},
		Short:   "List all configured MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			servers, err := h.ListServers()
			if err != nil {
				return err
			}

			return printOutput(cmd, servers, func() string {
				if len(servers) == 0 {
					return "No servers configured. Use 'mcphub connect' to add one."
				}
				var b strings.Builder
				fmt.Fprintf(&b, "Servers (%d):\n", len(servers))
				for _, s := range servers {
					statusIcon := "○"
					if s.Status == "connected" {
						statusIcon = "●"
					}
					fmt.Fprintf(&b, "  %s %-20s %s  [%s]\n", statusIcon, s.Name, s.Status, s.URL)
				}
				return b.String()
			})
		},
	}

	return cmd
}

// ==================== tools ====================

func newToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools [server]",
		Short: "List tools from a server (or all servers)",
		Long: `List available tools. If a server name is provided, lists tools from that server only.
Without a server name, lists tools from all connected servers.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := ""
			if len(args) > 0 {
				serverName = args[0]
			}

			tools, err := h.ListTools(serverName)
			if err != nil {
				return err
			}

			return printOutput(cmd, tools, func() string {
				if len(tools) == 0 {
					return "No tools available. Connect to a server first."
				}
				var b strings.Builder
				for _, t := range tools {
					fmt.Fprintf(&b, "Server: %s\n", t.Server)
					fmt.Fprintf(&b, "  Tool: %s\n", t.Name)
					if t.Description != "" {
						fmt.Fprintf(&b, "  Description: %s\n", t.Description)
					}
					if t.InputSchema.Properties != nil {
						fmt.Fprintf(&b, "  Parameters:\n")
						for name, prop := range t.InputSchema.Properties {
							p, _ := prop.(map[string]interface{})
							desc, _ := p["description"].(string)
							typ, _ := p["type"].(string)
							fmt.Fprintf(&b, "    - %s (%s): %s\n", name, typ, desc)
						}
					}
					fmt.Fprintln(&b)
				}
				return b.String()
			})
		},
	}

	return cmd
}

// ==================== call ====================

func newCallCmd() *cobra.Command {
	var callArgs string

	cmd := &cobra.Command{
		Use:   "call <server> <tool>",
		Short: "Call a tool on an MCP server",
		Long: `Call a tool on a connected MCP server with JSON arguments.

Examples:
  mcphub call github search_repos --args '{"query": "mcp"}'
  mcphub call time get_current_time`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			toolName := args[1]

			var parsedArgs map[string]interface{}
			if callArgs != "" {
				if err := json.Unmarshal([]byte(callArgs), &parsedArgs); err != nil {
					return fmt.Errorf("invalid JSON args: %w", err)
				}
			}

			result, err := h.CallTool(serverName, toolName, parsedArgs)
			if err != nil {
				return err
			}

			return printOutput(cmd, result, func() string {
				var b strings.Builder
				if result.IsError {
					fmt.Fprintf(&b, "Error from %s/%s:\n", serverName, toolName)
				} else {
					fmt.Fprintf(&b, "Result from %s/%s:\n", serverName, toolName)
				}
				for _, c := range result.Content {
					if c.Type == "text" {
						fmt.Fprintln(&b, c.Text)
					} else {
						fmt.Fprintf(&b, "[%s: %s]\n", c.Type, c.Data)
					}
				}
				return b.String()
			})
		},
	}

	cmd.Flags().StringVarP(&callArgs, "args", "a", "", "Tool arguments as JSON string")

	return cmd
}
