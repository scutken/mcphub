# MCPHub Agent 指南

本文件面向协助 MCPHub 开发的 AI Agent，说明项目规范与常用流程。

## 项目简介

MCPHub 是一个 Go 实现的 MCP（Model Context Protocol）客户端工具，同时提供：

- **CLI**：基于 Cobra，默认 JSON 输出，方便 Agent 集成
- **桌面 GUI**：基于 Wails v2 + Svelte 5

核心能力：连接 HTTP MCP 服务器（SSE / Streamable HTTP），渐进式发现工具，调用工具。

## 代码规范

- **注释用中文**（解释为什么这么做）
- **日志、错误信息、返回值、label 等面向用户/机器的文本一律英文**
- 示例：
  - ✅ `# 跳过已处理的文件` + `log.info("Skipping processed file")`
  - ❌ `# Skip processed files` + `log.info("Skipping processed file")`

## 发版流程

发布新版本请按以下步骤执行：

### 1. 更新变更日志

在 `CHANGELOG.md` 的 `[Unreleased]` 下方新增版本小节：

```markdown
## [x.y.z] - YYYY-MM-DD

### Added

- 新增功能说明

### Changed

- 变更说明

### Fixed

- 修复说明
```

### 2. 更新版本号

修改以下文件中嘅版本号：

- `wails.json`：`info.productVersion`
- `versioninfo.json`：`FixedFileInfo.FileVersion`、`FixedFileInfo.ProductVersion`、`StringFileInfo.FileVersion`、`StringFileInfo.ProductVersion`

### 3. 提交变更

```bash
git add CHANGELOG.md wails.json versioninfo.json
git commit -m "chore: bump version to x.y.z"
git push origin main
```

### 4. 打 tag 并推送

```bash
git tag vx.y.z
git push origin vx.y.z
```

### 5. 等待 CI 构建

GitHub Actions 嘅 `Build Windows` workflow 会自动：

- 编译 Windows 版 `mcphub.exe`（Wails production build）
- 创建 / 更新 Release
- 将 `mcphub.exe` 上传到 Release assets
- 用 `CHANGELOG.md` 中对应版本嘅内容作为 Release body

访问 `https://github.com/scutken/mcphub/actions/workflows/build-windows.yml` 查看构建状态。

### 6. 验证 Release

- 打开 `https://github.com/scutken/mcphub/releases/tag/vx.y.z`
- 确认 Release body 包含 changelog 内容
- 确认 Assets 中有 `mcphub.exe`
- 可用以下命令快速验证 asset 是否存在：

```bash
curl -I https://github.com/scutken/mcphub/releases/download/vx.y.z/mcphub.exe
# 期望返回 302 Found
```

## 本地构建

### CLI + GUI 完整构建

```bash
wails build -platform windows/amd64 -o mcphub.exe
```

输出：`build/bin/mcphub.exe`

注意：纯 `go build` 不会嵌入 Wails frontend，运行 GUI 模式会报错，生产发布请用 `wails build`。

### 运行测试

```bash
go test ./...
```

## 常用命令

```bash
# 启动 GUI 开发模式
wails dev

# 运行 CLI
mcphub servers
mcphub tools <server>
mcphub tools <server> --search <keyword>
mcphub tools <server> <tool>
mcphub call <server> <tool> --args '{"key":"value"}'
```

## Skill 安装路径

Agent skill 安装在用户目录：

```
~/.agents/skills/mcphub
```

这样 OpenCode 同其他 Agent 都能读取。
