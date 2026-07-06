---
name: mcphub
description: "Use the mcphub CLI to discover and call MCP (Model Context Protocol) tools from configured servers. On-demand MCP access without context overhead."
version: 1.0.0
platforms: [windows]
metadata:
  hermes:
    tags: [mcp, cli, tools, tool-calling]
---

# MCPHub CLI — On-Demand MCP Tool Access

## Overview

This skill provides instructions for using the `mcphub` CLI to interact with MCP (Model Context Protocol) servers. Instead of configuring MCP servers directly in the model's context, use `mcphub` CLI commands on demand to discover tools and call them. Output is JSON by default for easy parsing.

## Prerequisites

- `mcphub.exe` must be available in PATH
- Servers must be pre-connected by the user via `mcphub connect`

## Commands

### List connected servers

```bash
mcphub list
# or: mcphub servers, mcphub ls
```

Returns JSON array of server info with connection status:
```json
[
  {
    "name": "github",
    "url": "https://api.github.com/mcp",
    "transport": "auto",
    "status": "connected",
    "added_at": "2026-07-06T10:00:00Z"
  }
]
```

### Discover tools from a server

```bash
mcphub tools <server-name>
# Example: mcphub tools github
```

Returns JSON array of tool definitions:
```json
[
  {
    "server": "github",
    "name": "search_repositories",
    "description": "Search GitHub repositories",
    "inputSchema": {
      "type": "object",
      "properties": {
        "query": {"type": "string", "description": "Search query keywords"},
        "page": {"type": "integer", "description": "Page number (default: 1)"}
      },
      "required": ["query"]
    }
  }
]
```

To list tools from ALL connected servers (omit server name):
```bash
mcphub tools
```

### Call a tool

```bash
mcphub call <server-name> <tool-name> --args '<JSON-args>'
# Example: mcphub call github search_repositories --args '{"query":"mcp server"}'
```

Returns JSON result:
```json
{
  "server": "github",
  "tool": "search_repositories",
  "isError": false,
  "content": [
    {
      "type": "text",
      "text": "Found 42 repositories matching 'mcp server'..."
    }
  ]
}
```

### Human-readable output

Add `--json=false` to any command for formatted text output:
```bash
mcphub tools github --json=false
```

## Workflow Pattern

When you need external data or actions, follow these steps:

### Step 1: Check what's available

```bash
mcphub list      # see all connected servers
mcphub tools     # see all available tools across servers
```

### Step 2: Read the tool schema

Look at `inputSchema.properties` to understand required and optional parameters. Each property has `type` and `description` fields.

### Step 3: Call the tool

Construct a JSON string for `--args` matching the schema:
```bash
mcphub call <server> <tool> --args '{"param1":"value1","param2":42}'
```

### Step 4: Use the result

Parse the JSON output. `content[].text` contains human-readable text results. `isError: true` means the tool reported an error.

## Error Handling

- If `mcphub` returns non-zero exit code → the command itself failed (e.g., server not connected, network error). Read stderr.
- If exit code is 0 but `isError: true` in the JSON → the tool call succeeded but the tool reported a logical error (e.g., invalid arguments).
- If exit code is 0 and `isError: false` → success, use the content.

## Important Notes

- **Default output is JSON** — always parse with a JSON parser, not regex
- **Connection management** — servers must be pre-configured by the user via `mcphub connect`. You cannot add servers at runtime.
- **No streaming** — tool calls are synchronous request/response. Set reasonable timeouts.
- **Headers/Secrets** — authentication headers are stored in config and included automatically. You do not need to (and cannot) pass them on each call.
