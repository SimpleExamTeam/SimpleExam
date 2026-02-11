# Simple Exam 构建指南

本文档说明如何构建 Simple Exam 项目，包括版本信息注入和启动横幅显示。

## 构建方法

### 方法 1: 使用 Makefile（推荐）

```bash
# 构建当前平台
make build

# 构建 Windows 版本
make build-windows

# 构建 Linux 版本
make build-linux

# 构建 macOS 版本
make build-darwin

# 构建所有平台
make build-all

# 构建并运行
make run

# 清理构建产物
make clean
```

### 方法 2: 使用构建脚本

**Linux/macOS:**
```bash
chmod +x build.sh
./build.sh
```

**Windows (PowerShell):**
```powershell
.\build.ps1
# 或指定输出文件名
.\build.ps1 -Output "exam-system.exe" -Version "v0.2.0"
```

### 方法 3: 手动构建

```bash
# 获取版本信息
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.1.2")
COMMIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')

# 构建
go build -ldflags "-X 'main.Version=${VERSION}' -X 'main.CommitHash=${COMMIT_HASH}' -X 'main.BuildTime=${BUILD_TIME}' -s -w" -o simpleexam .
```

## 版本信息

构建时会自动注入以下信息：

- **Version**: 版本号（从 Git 标签获取，或使用默认值 v0.1.2）
- **CommitHash**: Git 提交哈希（前 7 位）
- **BuildTime**: 构建时间
- **Go Version**: Go 运行时版本（自动获取）
- **OS/Arch**: 操作系统和架构（自动获取）

## 启动横幅

应用启动时会显示 ASCII 艺术横幅和版本信息：

```
   _____ _                 _        ______                     
  / ____(_)               | |      |  ____|                    
 | (___  _ _ __ ___  _ __ | | ___  | |__  __  ____ _ _ __ ___  
  \___ \| | '_ ' _ \| '_ \| |/ _ \ |  __| \ \/ / _' | '_ ' _ \ 
  ____) | | | | | | | |_) | |  __/ | |____ >  < (_| | | | | | |
 |_____/|_|_| |_| |_| .__/|_|\___| |______/_/\_\__,_|_| |_| |_|
                    | |                                         
                    |_|                                         

  Version:     v0.1.2
  Commit:      a1b2c3d
  Build Time:  2024-02-11 10:30:00
  Go Version:  go1.21.0
  OS/Arch:     windows/amd64
```

## 运行应用

```bash
# 使用默认配置
./simpleexam

# 指定配置文件
./simpleexam -c config/config.yaml

# 查看版本
./simpleexam --version

# 查看帮助
./simpleexam --help

# 重置用户密码
./simpleexam reset-password -u admin -p newpassword
```

## 交叉编译

构建其他平台的可执行文件：

```bash
# Windows (64位)
GOOS=windows GOARCH=amd64 go build -ldflags "..." -o simpleexam.exe .

# Linux (64位)
GOOS=linux GOARCH=amd64 go build -ldflags "..." -o simpleexam-linux .

# macOS (64位)
GOOS=darwin GOARCH=amd64 go build -ldflags "..." -o simpleexam-darwin .

# macOS (ARM64 - M1/M2)
GOOS=darwin GOARCH=arm64 go build -ldflags "..." -o simpleexam-darwin-arm64 .
```

## 发布版本

创建新版本时：

1. 更新版本号（如果需要）
2. 提交所有更改
3. 创建 Git 标签：
   ```bash
   git tag -a v0.2.0 -m "Release version 0.2.0"
   git push origin v0.2.0
   ```
4. 构建发布版本：
   ```bash
   make build-all
   ```

## 注意事项

- 确保已安装 Go 1.21 或更高版本
- 如果不在 Git 仓库中，版本信息将使用默认值
- `-s -w` 标志会减小二进制文件大小（去除调试信息和符号表）
- 生产环境建议使用 `make build` 或构建脚本，确保版本信息正确注入
