#!/bin/bash
# Go MCP CLI 项目初始化脚本

set -e

echo "🚀 Go MCP CLI - 项目初始化"
echo ""

# 创建项目目录结构
echo "📁 创建项目目录结构..."
mkdir -p cmd/mcp-cli
mkdir -p pkg/client
mkdir -p pkg/config
mkdir -p pkg/cli
mkdir -p pkg/session
mkdir -p examples
mkdir -p internal/testdata

echo "✓ 项目目录结构已创建"
echo ""

# 初始化 Go module
echo "📦 初始化 Go module..."
if [ ! -f "go.mod" ]; then
    go mod init go-mcp-cli
    echo "✓ go.mod 已创建"
else
    echo "✓ go.mod 已存在，跳过"
fi
echo ""

# 安装核心依赖
echo "📥 安装依赖包..."
go get github.com/modelcontextprotocol/go-sdk@latest
go get github.com/spf13/cobra@latest
go get github.com/spf13/pflag@latest

echo "✓ 依赖包已安装"
echo ""

# 安装开发工具
echo "🛠️ 安装开发工具..."
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo "✓ 开发工具已安装"
echo ""

# 整理依赖
echo "🧹 整理依赖..."
go mod tidy

echo "✓ 依赖已整理"
echo ""

# 构建项目
echo "🔨 构建 CLI 二进制文件..."
go build -o mcp-cli ./cmd/mcp-cli

echo "✓ 构建完成: mcp-cli"
echo ""

# 测试
echo "🧪 测试 CLI..."
./mcp-cli help

echo ""
echo "✅ 项目初始化完成！"
echo ""
echo "📝 下一步："
echo "  1. 运行: ./mcp-cli add time stdio --command uvx --args mcp-server-time"
echo "  2. 运行: ./mcp-cli list"
echo "  3. 运行: ./mcp-cli tools time"
echo ""
echo "📚 查看文档："
echo "  - GO_REFACTOR_PLAN.md - 完整重构方案"
echo "  - GO_EXAMPLES.md - Go 代码示例"
echo "  - MIGRATION_CHECKLIST.md - 迁移检查清单"
echo "  - README_GO.md - Go CLI 文档"
