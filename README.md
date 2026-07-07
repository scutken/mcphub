# MCPHub

Go 实现的 MCP（Model Context Protocol）客户端工具，支持桌面 GUI 和 CLI。管理 MCP 服务器连接，按需发现和调用工具——让 AI Agent 无需在上下文中配置 MCP。

## 特性

- **零外部 MCP 依赖** — 自实现 JSON-RPC 2.0 + SSE + Streamable HTTP 客户端，~800 行 Go
- **双模式** — 同一个 .exe 同时提供 CLI 和桌面 GUI（Wails v2 + Svelte 5）
- **HTTP 传输** — 支持 MCP 2024-11-05 (SSE) 和 2025-11-25 (Streamable HTTP)
- **AI 友好** — 所有 CLI 输出默认 JSON，可直接被 grep/jq/LLM 解析
- **国风设计** — 墨韵暗色系 + IBM Plex Sans/Mono 字体
- **Windows 原生** — 纯 Go 编译，无运行时依赖（WebView2 除外）

## 快速开始

### 下载

从 [Releases](https://github.com/scutken/mcphub/releases) 下载 `mcphub.exe`。

### CLI 使用

```bash
# 添加 MCP 服务器
mcphub connect github https://api.github.com/mcp -H "Authorization: Bearer ghp_xxx"

# 查看所有服务器
mcphub list

# 浏览可用工具（JSON 输出）
mcphub tools github

# 调用工具
mcphub call github search_repos --args '{"query":"mcp server"}'

# 人类可读输出
mcphub list --json=false

# 断开连接
mcphub disconnect github

# 启动桌面 GUI
mcphub serve
```

### 安装 Agent Skill

将 MCPHub 注册为 OpenCode 的 skill，让 AI Agent 自动发现和调用 MCP 工具：

```powershell
# 方式一：软链接（推荐，跟随项目更新）
New-Item -ItemType SymbolicLink `
  -Path "$env:USERPROFILE\.config\opencode\skills\mcphub" `
  -Target (Resolve-Path skills\mcphub)

# 方式二：直接复制
Copy-Item -Recurse skills\mcphub "$env:USERPROFILE\.config\opencode\skills\mcphub"
```

安装后 Agent 即可通过 `mcphub tools` / `mcphub call` 按需访问 MCP 工具，无需在上下文中配置 MCP 服务器。

Skill 结构：
```
skills/mcphub/
└── SKILL.md    ← Agent 使用说明（CLI 命令参考 + 工作流）
```

## 从源码构建

### 本地开发

```bash
# 1. 安装 pnpm 和 Wails CLI
npm install -g pnpm
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 2. 克隆项目
git clone https://github.com/scutken/mcphub.git
cd mcphub

# 3. 安装前端依赖
cd frontend
pnpm install --ignore-scripts
pnpm rebuild esbuild
cd ..

# 4. 启动开发模式（热重载）
wails dev

# 5. 或直接构建 .exe
wails build -platform windows/amd64
```

## 项目结构

```
mcphub/
├── main.go                   # 入口（CLI / GUI 双模式）
├── app.go                    # Wails GUI Binding
├── wails.json                # Wails 构建配置
│
├── cmd/cli/                  # Cobra CLI
│   ├── root.go               # connect / disconnect / list / tools / call
│   └── cli_test.go
│
├── pkg/
│   ├── mcp/                  # MCP 协议客户端（零外部依赖）
│   │   ├── protocol.go       # JSON-RPC 2.0 + MCP 类型
│   │   ├── transport.go      # SSE + Streamable HTTP
│   │   └── client.go         # Connect / ListTools / CallTool
│   ├── config/               # JSON 配置持久化
│   └── hub/                  # 统一服务层（CLI & GUI 共享）
│
├── frontend/                 # Svelte 5 + Tailwind
│   └── src/
│       ├── app.css           # 墨韵国风色系
│       ├── routes/           # 页面路由
│       └── lib/components/   # UI 组件
│
├── skills/
│   └── mcphub/
│       └── SKILL.md            # AI Agent skill（CLI 命令 + 工作流）
└── .github/workflows/        # Windows 构建 CI
```

## 设计

| 维度 | 选择 |
|------|------|
| 语言 | Go 1.24 |
| 桌面框架 | Wails v2 |
| 前端 | Svelte 5 + Tailwind CSS |
| 色系 | 墨韵国风暗色系（墨/玄青/黛蓝 + 鎏金/朱砂/石青） |
| 字体 | IBM Plex Sans（UI）+ IBM Plex Mono（代码） |
| 图标 | Lucide Icons |

## License

MIT
