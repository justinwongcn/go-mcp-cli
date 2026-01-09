# 发布说明

## 文件权限问题修复

### 问题描述
之前通过 GitHub Actions 发布的二进制文件下载后显示为"文稿"而不是可执行文件，这是因为：

1. **GitHub Actions Artifacts 不保留文件权限**：`actions/upload-artifact@v4` 和 `actions/download-artifact@v4` 在上传/下载过程中会丢失 Unix 文件权限信息
2. **Release 上传也可能丢失权限**：某些上传方式不会保留可执行权限标志

### 解决方案

现在的发布流程会：

1. **为每个二进制文件创建压缩包**：
   - `*.tar.gz` - 保留 Unix 文件权限，适用于 Linux 和 macOS
   - `*.zip` - 适用于 Windows 用户

2. **同时提供原始二进制文件**：
   - 直接下载的二进制文件（如果系统支持）
   - 压缩包内的二进制文件（推荐）

3. **生成 SHA256 校验和**：
   - `checksums.txt` 包含所有文件的校验和

### 下载和使用

#### Linux / macOS 用户（推荐使用 tar.gz）

```bash
# 下载 tar.gz 压缩包
wget https://github.com/justinwongcn/go-mcp-cli/releases/latest/download/mcp-cli-linux-amd64.tar.gz

# 解压（会自动保留执行权限）
tar -xzf mcp-cli-linux-amd64.tar.gz

# 直接运行
./mcp-cli-linux-amd64 --version

# 可选：移动到系统路径
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
```

#### macOS 用户（Apple Silicon）

```bash
# 下载 ARM64 版本
wget https://github.com/justinwongcn/go-mcp-cli/releases/latest/download/mcp-cli-darwin-arm64.tar.gz

# 解压
tar -xzf mcp-cli-darwin-arm64.tar.gz

# 运行
./mcp-cli-darwin-arm64 --version
```

#### Windows 用户

```powershell
# 下载 zip 文件
# 使用浏览器或 PowerShell 下载

# 解压后直接运行
.\mcp-cli-windows-amd64.exe --version
```

#### 如果下载了原始二进制文件（没有压缩包）

如果你直接下载了二进制文件而不是压缩包，可能需要手动添加执行权限：

```bash
# Linux / macOS
chmod +x mcp-cli-linux-amd64
./mcp-cli-linux-amd64 --version
```

### 验证下载文件

```bash
# 下载校验和文件
wget https://github.com/justinwongcn/go-mcp-cli/releases/latest/download/checksums.txt

# 验证文件完整性
shasum -a 256 -c checksums.txt
```

## 可用的平台

- `mcp-cli-windows-amd64.exe` - Windows 64位
- `mcp-cli-linux-amd64` - Linux 64位
- `mcp-cli-linux-arm64` - Linux ARM64
- `mcp-cli-darwin-amd64` - macOS Intel
- `mcp-cli-darwin-arm64` - macOS Apple Silicon

每个平台都提供：
- 原始二进制文件
- `.tar.gz` 压缩包
- `.zip` 压缩包
