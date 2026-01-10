# Go MCP CLI

基于 Go 和官方 [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) 实现的 Model Context Protocol 客户端 CLI 工具。

## 功能特性

- **多传输支持**：Stdio、SSE、Streamable HTTP
- **类型安全**：完整的 Go 类型系统
- **CLI 工具**：管理 MCP 服务器的命令行界面
- **配置持久化**：服务器配置保存在 `./.mcp-cli/config.json`
- **临时调用**：无需预配置即可直接调用 MCP 工具
- **Claude Desktop 兼容**：支持导入 Claude Desktop JSON 配置文件
- **智能传输检测**：URL 自动选择 HTTP 或 SSE 传输

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/justinwongcn/go-mcp-cli
cd go-mcp-cli

# 安装依赖
go mod download

# 构建 CLI
go build -o mcp-cli ./cmd/mcp-cli
```

### 预编译二进制

从 [GitHub Releases](https://github.com/justinwongcn/go-mcp-cli/releases) 下载预编译的二进制文件。

#### macOS 用户注意事项

由于二进制文件未经过 Apple 代码签名，首次运行时 macOS 会显示安全警告。解决方法：

**方法 1：使用命令行移除隔离属性**
```bash
# 下载二进制文件后
# 移除隔离属性
xattr -d com.apple.quarantine mcp-cli-darwin-arm64
# 添加执行权限
chmod +x mcp-cli-darwin-arm64
# 移动到 PATH 目录（可选）
sudo mv mcp-cli-darwin-arm64 /usr/local/bin/mcp-cli
```

**方法 2：通过系统设置允许**
1. 双击运行二进制文件，会弹出安全警告
2. 打开"系统设置" → "隐私与安全性"
3. 在底部找到"仍要打开"按钮并点击
4. 再次确认打开

## 快速开始

### 添加服务器

```bash
# Stdio 传输（本地进程）
mcp-cli add time stdio --command uvx --args mcp-server-time

# SSE 传输（服务器发送事件）
mcp-cli add gitmcp sse --url https://gitmcp.io/modelcontextprotocol/typescript-sdk

# HTTP 传输（自动检测）
mcp-cli add context7 http --url https://mcp.context7.com/mcp

# 带头部认证
mcp-cli add myserver http --url https://example.com/mcp --header "Authorization=Bearer token123"

# 带环境变量
mcp-cli add myserver stdio --command python --args "server.py,--port,8080" --env "PYTHONPATH=/app"
```

### 列出服务器

```bash
mcp-cli list
```

### 导入 Claude Desktop 配置

```bash
# 从 Claude Desktop 配置文件导入
mcp-cli import ~/.config/claude/claude_desktop_config.json

# 从指定路径导入
mcp-cli import /path/to/config.json
```

### 列出工具

```bash
# 预配置模式下列出
mcp-cli tools time

# 临时模式下列出（无需预配置）
mcp-cli exec stdio --list --command uvx --args mcp-server-time
mcp-cli exec sse --list --url https://gitmcp.io/modelcontextprotocol/typescript-sdk
mcp-cli exec http --list --url https://mcp.context7.com/mcp
```

### 调用工具

```bash
# 预配置模式下调用
mcp-cli call time get_current_time --arg timezone=Asia/Shanghai

# 临时调用（无需预配置）
mcp-cli exec stdio get_current_time --command uvx --args mcp-server-time --arg timezone=Asia/Shanghai
```

### 删除服务器

```bash
mcp-cli remove time
```

## 临时调用

无需预先配置服务器，直接通过命令行调用 MCP 工具：

```bash
# Stdio 传输 - 列出工具
mcp-cli exec stdio --list --command uvx --args mcp-server-time

# Stdio 传输 - 调用工具
mcp-cli exec stdio <工具名> --command <命令> --args <参数> --arg <工具参数>

# SSE 传输 - 列出工具
mcp-cli exec sse --list --url <SSE端点>

# SSE 传输 - 调用工具
mcp-cli exec sse <工具名> --url <SSE端点> --header <请求头> --arg <工具参数>

# HTTP 传输 - 列出工具
mcp-cli exec http --list --url <HTTP端点>

# HTTP 传输 - 调用工具
mcp-cli exec http <工具名> --url <HTTP端点> --header <请求头> --arg <工具参数> --retries 3
```

示例：

```bash
# 时间服务器（Stdio）- 列出工具
mcp-cli exec stdio --list --command uvx --args mcp-server-time

# 时间服务器（Stdio）- 调用工具
mcp-cli exec stdio get_current_time --command uvx --args mcp-server-time --arg timezone=Asia/Shanghai

# GitMCP（SSE）- 列出工具
mcp-cli exec sse --list --url https://gitmcp.io/modelcontextprotocol/typescript-sdk

# GitMCP（SSE）- 调用工具
mcp-cli exec sse search_typescript_sdk_code --url https://gitmcp.io/modelcontextprotocol/typescript-sdk --arg query=Client

# Context7（HTTP）- 列出工具
mcp-cli exec http --list --url https://mcp.context7.com/mcp

# Context7（HTTP）- 调用工具
mcp-cli exec http resolve-library-id --url https://mcp.context7.com/mcp --arg query=React --arg libraryName=express

# 带头部认证的 HTTP 调用
mcp-cli exec http query-docs --url https://mcp.context7.com/mcp --header "CONTEXT7_API_KEY=your-key" --arg libraryId=/expressjs/express --arg query="How to use middleware"
```

## 支持的传输类型

| 传输类型 | 使用场景 | 配置项 |
|---------|---------|--------|
| Stdio | 本地 MCP 服务器 | `--command`, `--args`, `--env` |
| SSE | 服务器发送事件 | `--url` (包含 `/sse`), `--header` |
| HTTP | Streamable HTTP | `--url`, `--header`, `--retries` |

## Claude Desktop 配置兼容

支持直接导入 Claude Desktop JSON 配置文件：

```json
{
  "mcpServers": {
    "context7": {
      "url": "https://mcp.context7.com/mcp",
      "headers": {
        "CONTEXT7_API_KEY": "your-api-key"
      }
    },
    "time-server": {
      "command": "uvx",
      "args": ["mcp-server-time"]
    },
    "deepwiki": {
      "url": "https://mcp.deepwiki.com/sse"
    }
  }
}
```

**自动检测规则：**
- 有 `command` → Stdio 传输
- URL 包含 `/sse` → SSE 传输
- 其他 URL → HTTP (Streamable) 传输

## 配置文件

配置文件默认保存在 `./.mcp-cli/config.json`：

```json
{
  "version": "1.0.0",
  "servers": {
    "context7": {
      "name": "context7",
      "transport": "http",
      "url": "https://mcp.context7.com/mcp",
      "headers": {
        "CONTEXT7_API_KEY": "your-api-key"
      }
    }
  }
}
```

指定自定义配置文件路径：

```bash
mcp-cli --config /path/to/config.json list
mcp-cli --config /path/to/config.json add server-name stdio --command echo
```

## 构建多平台二进制

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o mcp-cli.exe ./cmd/mcp-cli

# Linux
GOOS=linux GOARCH=amd64 go build -o mcp-cli-linux-amd64 ./cmd/mcp-cli
GOOS=linux GOARCH=arm64 go build -o mcp-cli-linux-arm64 ./cmd/mcp-cli

# macOS
GOOS=darwin GOARCH=amd64 go build -o mcp-cli-darwin-amd64 ./cmd/mcp-cli
GOOS=darwin GOARCH=arm64 go build -o mcp-cli-darwin-arm64 ./cmd/mcp-cli
```

## 许可证

MIT License

## 参考

- [Model Context Protocol](https://modelcontextprotocol.io)
- [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
