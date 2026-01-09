@echo off
REM Go MCP CLI - 项目初始化和构建脚本 (Windows)

echo 🚀 Go MCP CLI - 项目初始化
echo.

REM 1. 创建项目目录结构
echo 📁 创建项目目录结构...
if not exist cmd\mcp-cli mkdir cmd\mcp-cli
if not exist pkg\client mkdir pkg\client
if not exist pkg\config mkdir pkg\config
if not exist pkg\cli mkdir pkg\cli
if not exist pkg\session mkdir pkg\session
if not exist examples mkdir examples

echo ✓ 项目目录结构已创建
echo.

REM 2. 初始化 Go module
echo 📦 初始化 Go module...
if not exist go.mod (
    go mod init go-mcp-cli
    echo ✓ go.mod 已创建
) else (
    echo ✓ go.mod 已存在，跳过
)
echo.

REM 3. 安装依赖
echo 📥 安装依赖包...
go get github.com/modelcontextprotocol/go-sdk@latest
go get github.com/spf13/cobra@latest
go get github.com/spf13/pflag@latest

echo ✓ 依赖包已安装
echo.

REM 4. 整理依赖
echo 🧹 整理依赖...
go mod tidy

echo ✓ 依赖已整理
echo.

REM 5. 构建项目
echo 🔨 构建 CLI 二进制文件...
go build -o mcp-cli.exe ./cmd/mcp-cli

echo ✓ 构建完成: mcp-cli.exe
echo.

REM 6. 测试
echo 🧪 测试 CLI...
mcp-cli.exe help

echo.
echo ✅ 项目初始化完成！
echo.
echo 📝 下一步：
echo   1. 运行: mcp-cli.exe add time stdio --command uvx --args mcp-server-time
echo   2. 运行: mcp-cli.exe list
echo   3. 运行: mcp-cli.exe tools time
echo.
echo 📚 查看文档：
echo   - GO_REFACTOR_PLAN.md - 完整重构方案
echo   - GO_EXAMPLES.md - Go 代码示例
echo   - MIGRATION_CHECKLIST.md - 迁移检查清单
echo   - README_GO.md - Go CLI 文档
pause
