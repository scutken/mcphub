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
		Short: "MCPHub - MCP (Model Context Protocol) 服务器管理器",
		Long: `MCPHub 是一个管理 HTTP MCP 服务器连接的 CLI 工具。

功能：
  - 连接 MCP 服务器（HTTP SSE / Streamable）
  - 列出已连接服务器的工具
  - 调用工具并传参

默认输出 JSON 格式，方便 AI 代理集成。
使用 --json=false 切换为可读文本输出。`,
	}

	root.PersistentFlags().BoolP("json", "j", true, "以 JSON 格式输出")

	root.AddCommand(newConnectCmd())
	root.AddCommand(newDisconnectCmd())
	root.AddCommand(newServersCmd())
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
	return true
}

// printOutput prints data as JSON or a formatted string.
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
		Use:   "connect <名称> <URL>",
		Short: "连接一个 MCP 服务器",
		Long: `通过 HTTP 连接 MCP 服务器并保存连接信息。

连接建立后立即进行握手。传输协议自动检测（优先 Streamable HTTP，其次 SSE）。

示例：
  mcphub connect my-server https://mcp.example.com
  mcphub connect github https://api.github.com/mcp --header "Authorization: Bearer ***"
  mcphub connect server3 https://example.com --transport sse`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]

			headerMap := make(map[string]string)
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("header 格式无效: %q（应为 'Key: Value'）", h)
				}
				headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}

			if err := h.Connect(name, url, headerMap, transport); err != nil {
				return err
			}

			servers, err := h.ListServers()
			if err != nil {
				return err
			}
			for _, s := range servers {
				if s.Name == name {
					return printOutput(cmd, s, func() string {
						return fmt.Sprintf("✓ 已连接 %s (%s) — 状态: %s", s.Name, s.URL, s.Status)
					})
				}
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "HTTP 请求头 (Key: Value)，可多次指定")
	cmd.Flags().StringVarP(&transport, "transport", "t", "auto", "传输协议: auto, sse, streamable")

	return cmd
}

// ==================== disconnect ====================

func newDisconnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect <名称>",
		Short: "断开并移除一个 MCP 服务器",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := h.Disconnect(name); err != nil {
				return err
			}

			return printOutput(cmd, map[string]string{"status": "disconnected", "server": name}, func() string {
				return fmt.Sprintf("✓ 已断开 %s", name)
			})
		},
	}

	return cmd
}

// ==================== servers ====================

func newServersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "servers",
		Short: "列出所有已配置的 MCP 服务器",
		RunE: func(cmd *cobra.Command, args []string) error {
			servers, err := h.ListServers()
			if err != nil {
				return err
			}

			return printOutput(cmd, servers, func() string {
				if len(servers) == 0 {
					return "暂无已配置的服务器。使用 'mcphub connect' 添加一个。"
				}
				var b strings.Builder
				fmt.Fprintf(&b, "服务器 (%d):\n", len(servers))
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
	var search string

	cmd := &cobra.Command{
		Use:   "tools [服务器名] [工具名...]",
		Short: "列出工具摘要，或获取指定工具的完整 schema",
		Long: `渐进式工具发现：
  - 默认输出工具摘要（server/name/description），不含 inputSchema。
  - --search/-s <keyword> 按关键字搜索工具名或描述（大小写不敏感）。
  - 获取完整 schema 时必须指定服务器，可一次传多个工具名。`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := ""
			var toolNames []string

			if len(args) > 0 {
				serverName = args[0]
				toolNames = args[1:]
			}

			// 获取 schema 必须指定服务器
			if len(toolNames) > 0 && serverName == "" {
				return fmt.Errorf("获取工具 schema 时必须指定服务器")
			}

			// 有工具名 → 返回完整 ToolInfo
			if len(toolNames) > 0 {
				tools, err := h.GetTools(serverName, toolNames)
				if err != nil {
					return err
				}
				return printOutput(cmd, tools, func() string {
					var b strings.Builder
					for _, t := range tools {
						fmt.Fprintf(&b, "服务器: %s\n", t.Server)
						fmt.Fprintf(&b, "  工具: %s\n", t.Name)
						if t.Description != "" {
							fmt.Fprintf(&b, "  描述: %s\n", t.Description)
						}
						if t.InputSchema.Properties != nil {
							fmt.Fprintf(&b, "  参数:\n")
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
			}

			// --search → 搜索摘要
			if search != "" {
				summaries, err := h.SearchToolSummaries(serverName, search)
				if err != nil {
					return err
				}
				return printOutput(cmd, summaries, func() string {
					if len(summaries) == 0 {
						return fmt.Sprintf("未找到包含 %q 的工具。", search)
					}
					var b strings.Builder
					fmt.Fprintf(&b, "搜索结果 (%d):\n", len(summaries))
					for _, s := range summaries {
						fmt.Fprintf(&b, "  %-20s %-30s %s\n", s.Server, s.Name, s.Description)
					}
					return b.String()
				})
			}

			// 默认 → 摘要列表
			summaries, err := h.ListToolSummaries(serverName)
			if err != nil {
				return err
			}
			return printOutput(cmd, summaries, func() string {
				if len(summaries) == 0 {
					return "暂无可用工具。请先连接一个服务器。"
				}
				var b strings.Builder
				currentServer := ""
				for _, s := range summaries {
					if s.Server != currentServer {
						if currentServer != "" {
							fmt.Fprintln(&b)
						}
						fmt.Fprintf(&b, "%s:\n", s.Server)
						currentServer = s.Server
					}
					fmt.Fprintf(&b, "  %-30s %s\n", s.Name, s.Description)
				}
				return b.String()
			})
		},
	}

	cmd.Flags().StringVarP(&search, "search", "s", "", "按关键字搜索工具名或描述")

	return cmd
}

// ==================== call ====================

func newCallCmd() *cobra.Command {
	var callArgs string

	cmd := &cobra.Command{
		Use:   "call <服务器名> <工具名>",
		Short: "调用 MCP 服务器上的工具",
		Long: `调用已连接 MCP 服务器上的工具，传入 JSON 参数。

示例：
  mcphub call github search_repos --args '{"query": "mcp"}'
  mcphub call time get_current_time`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			toolName := args[1]

			parsedArgs := make(map[string]interface{})
			if callArgs != "" {
				if err := json.Unmarshal([]byte(callArgs), &parsedArgs); err != nil {
					return fmt.Errorf("JSON 参数无效: %w", err)
				}
			}

			result, err := h.CallTool(serverName, toolName, parsedArgs)
			if err != nil {
				return err
			}

			return printOutput(cmd, result, func() string {
				var b strings.Builder
				if result.IsError {
					fmt.Fprintf(&b, "%s/%s 返回错误:\n", serverName, toolName)
				} else {
					fmt.Fprintf(&b, "%s/%s 返回结果:\n", serverName, toolName)
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

	cmd.Flags().StringVarP(&callArgs, "args", "a", "", "工具参数，JSON 字符串格式")

	return cmd
}
