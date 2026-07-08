---
name: mcphub
description: "通过 mcphub CLI 发现并调用已连接的 MCP 服务器上的工具。当需要调用 MCP 工具、查询已连接的 MCP 服务器、查看可用工具列表、或执行需要 MCP 服务器能力（如日历、待办、剪贴板、图片处理、文件搜索、视频处理等）的任务时使用。前提：用户已通过 `mcphub connect` 配置好服务器，工具能力取决于已连接的服务器。"
---

# MCPHub — MCP 工具调用

管理 HTTP MCP 服务器的连接、工具发现和调用。本 skill 只覆盖 CLI 用法（agent 集成场景）。

## 调用方式

`mcphub` 已全局安装，直接用命令名调用：

```bash
mcphub <command> [flags]
```

PowerShell 里传 JSON 参数用单引号包双引号：

```powershell
mcphub call <server> <tool> --args '{"key":"value"}'
```

## 命令

```bash
servers                                   # 服务器列表 +连接状态
servers --json=false                      # 可读文本
connect <名称> <URL>                     # 连接服务器
connect <名称> <URL> -H "K: V" -t streamable    # 带 header + 指定传输
disconnect <名称>                         # 断开并移除
tools                                     # 所有服务器工具摘要
tools <服务器名>                          # 指定服务器工具摘要
tools [<服务器名>] --search <keyword>     # 搜索工具名或描述（不指定=所有服务器）
tools <服务器名> <工具名...>              # 完整 schema（含 inputSchema，必须指定服务器）
call <服务器> <工具名> --args '{"k":"v"}' # 调用工具
call <服务器> <工具名>                     # 无参调用
call <服务器> <工具名> --json=false        # 文本模式输出
```

默认 JSON 输出。`-H` 可多次指定。`-t` 可选 `auto`(默认)/`streamable`。

## 工作流

### 流 A：服务维护

管理已配置的 MCP 服务器：查看、添加、删除。

```bash
mcphub servers                          # 查看所有服务器及连接状态
mcphub connect <名称> <URL> -H "K: V"   # 添加并连接服务器
mcphub disconnect <名称>                 # 断开并移除服务器
```

`mcphub servers` 返回数组，每项含 `name`/`url`/`transport`/`status`/`added_at`。只有 `status: "connected"` 的服务器可调用。

### 流 B：发现与调用工具

从已连接服务器发现工具，按需取 schema，然后执行。

#### 1. 浏览工具摘要

```bash
mcphub tools            # 所有已连接服务器
mcphub tools <服务器名>  # 指定服务器
```

返回摘要数组，每项含 `server`/`name`/`description`（不含 `inputSchema`）。

#### 2. 搜索工具

```bash
mcphub tools --search <keyword>            # 所有已连接服务器
mcphub tools <服务器名> --search <keyword>  # 指定服务器
```

按关键字搜索工具名或描述（大小写不敏感），返回匹配的摘要。

#### 3. 获取工具 schema

```bash
mcphub tools <服务器名> <工具名>
mcphub tools <服务器名> <工具1> <工具2> ...
```

返回完整 `ToolInfo` 数组（含 `inputSchema`）。看 `inputSchema.properties` 知参数：每项有 `type`/`description`，`inputSchema.required` 列必填项。支持一次传多个工具名批量获取。

#### 4. 调用

```bash
mcphub call utools utools.z6htitix.get_date_info --args '{"date":"2026-07-06"}'
```

#### 5. 解析结果

```json
{
  "server": "utools",
  "tool": "utools.z6htitix.search_next_festival",
  "isError": false,
  "content": [
    { "type": "text", "text": "{\"festival_name\":\"中秋节\",\"date\":\"2026-09-25\"}" }
  ]
}
```

**关键**：`content[].text` 通常是 **JSON 字符串**，需二次 `json.loads` 才能拿到结构化数据。

判断成功/失败：
- `isError: false` + exit 0 → `content[].text` 是结果（多半需二次解析）
- `isError: true` + exit 0 → 工具执行返回错误，`content[].text` 是错误信息
- 非 0 exit → 命令本身失败（网络/未连接/参数错），看 stderr 的 `Error: ...`

## 注意事项

- **输出默认 JSON**——用 JSON 解析器，别用正则
- **同步调用**——无流式，超时 120 秒
- **headers 自动持久化**——`connect` 时的 `-H` 后续 `call` 自动带上
- **Session 自动管理**——`Mcp-Session-Id` 由 mcphub 处理
- **多服务器工具名带前缀**——如 `utools.z6htitix.get_date_info`，`call` 时要完整传
- **配置持久化**——`connect` 写入配置文件，后续任意 CLI 调用自动加载并重连
