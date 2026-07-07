---
name: mcphub
description: "使用 mcphub CLI 发现并调用 MCP 工具。按需访问，不占模型上下文。适用于已通过 mcphub connect 连接了 MCP 服务器的场景。"
---

# MCPHub — MCP 工具管理器

## 是什么

MCPHub 管理 HTTP MCP 服务器的连接、工具发现和调用。两种使用方式：

- **CLI**：`mcphub` 命令行，JSON 输出，适合 AI 代理集成
- **GUI**：Wails 桌面应用，`wails dev` 启动，可视化操作

## CLI 命令速查

```bash
# 查看已连接服务器
mcphub list

# 连接服务器（支持 Session 管理、Streamable/SSE 自动检测）
mcphub connect <名称> <URL> --header "Key: Value"

# 查工具列表
mcphub tools <服务器名>        # 指定服务器
mcphub tools                   # 全部服务器

# 调用工具
mcphub call <服务器> <工具名> --args '{"k":"v"}'
mcphub call <服务器> <工具名>                    # 无参调用

# 可读文本
mcphub list --json=false
```

## GUI 开发

```bash
# 启动开发模式（Windows）
wails dev

# 或者用 bat 脚本
dev.bat
```

启动后 WebView2 窗口自动弹出，Vite dev server 在 `localhost:5173`，Wails 代理在 `localhost:34115`。

## 工作流

### 1. 看有什么

```bash
mcphub list          # 服务器列表 + 连接状态
mcphub tools         # 所有服务器的工具总览
```

### 2. 看工具怎么用

查 `inputSchema.properties`，每项含 `type`、`description`、是否 `required`。

### 3. 调工具

```bash
mcphub call utools utools.z6htitix.get_date_info --args '{"date":"2026-07-06"}'
```

### 4. 解析结果

- `isError: false` → `content[].text` 是结果
- `isError: true` → `content[].text` 是错误信息
- 非零退出码 → 命令本身失败（网络、未连接等），看 stderr

## JSON 输出示例

连接服务器：
```json
{
  "name": "utools",
  "url": "http://127.0.0.1:3501/mcp",
  "transport": "auto",
  "status": "connected",
  "added_at": "2026-07-06T20:08:36+08:00"
}
```

调用工具：
```json
{
  "server": "utools",
  "tool": "utools.z6htitix.search_next_festival",
  "isError": false,
  "content": [
    {
      "type": "text",
      "text": "{\"festival_name\":\"中秋节\",\"date\":\"2026-09-25\",\"days_from_today\":81}"
    }
  ]
}
```

## 支持的传输协议

| 类型 | 说明 |
|------|------|
| `auto` | 自动检测（默认，先试 Streamable） |
| `streamable` | MCP Streamable HTTP (2025 规范) |
| `sse` | Server-Sent Events (2024 规范) |

Session 管理（`Mcp-Session-Id`）自动处理，无需手动干预。

## 注意事项

- **输出默认 JSON**——用 JSON 解析器，别用正则
- **服务器需预先连接**——AGENT 不能自动加服务器，用户需先 `mcphub connect`
- **同步调用**——无流式，超时 120 秒
- **请求头自动附带**——`connect` 时存的 headers 后续调用自动带上
- **GUI 模式会话独立**——CLI 和 GUI 各自维护连接，互不影响
