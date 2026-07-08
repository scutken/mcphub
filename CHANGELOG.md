# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/lang/zh-CN/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-07-08

### Changed

- 仅支持 Streamable HTTP 传输，移除 SSE 支持
- CLI `--transport` 选项移除 `sse`，保留 `auto`/`streamable`
- GUI 添加服务器时不再显示 SSE 选项

### Fixed

- 清理 transport 类型判断逻辑，减少协议分支

## [0.0.1] - 2026-07-07

### Added

- 初始版本发布
- 支持连接 Streamable HTTP MCP 服务器
- 支持桌面 GUI 与 CLI 双模式
- CLI 渐进式工具发现：摘要列表、搜索、批量 schema 获取
- `mcphub servers` 命令查看已配置服务器状态
- `mcphub tools` 命令支持按服务器/关键字搜索工具
- `mcphub call` 命令调用 MCP 工具
- OpenCode / Agent skill，安装到 `~\.agents\skills`
- GitHub Actions 自动构建 Windows exe 并发布 Release

### Changed

- 将 `mcphub list` 改为 `mcphub servers`
- 工具列表默认仅展示名称与描述，schema 需指定工具名获取

[Unreleased]: https://github.com/scutken/mcphub/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/scutken/mcphub/releases/tag/v0.1.0
[0.0.1]: https://github.com/scutken/mcphub/releases/tag/v0.0.1
