# Simple Exam 考试系统

基于 Gin 框架开发的在线考试系统后端

## 技术栈

- Gin: Web 框架
- GORM: ORM 框架
- JWT: 用户认证
- MySQL: 数据存储
- 微信SDK: 微信登录和支付
- Zap: 高性能日志库
- Lumberjack: 日志轮转

## 项目结构

```
.
├── .github/               # GitHub Actions 工作流
├── certs/                 # 微信支付证书目录
├── config/                # 配置文件
│   ├── config.example.yaml
│   └── config.yaml
├── docs/                  # 文档目录
│   ├── BUILD.md          # 构建文档
│   └── CONFIG.md         # 配置文档
├── internal/              # 内部代码
│   ├── api/              # API 处理器
│   │   └── admin/        # 管理端 API
│   ├── config/           # 配置加载
│   ├── controller/       # 控制器层
│   ├── middleware/       # 中间件
│   ├── model/            # 数据模型
│   ├── pkg/              # 内部公共包
│   │   ├── banner/       # 启动横幅
│   │   ├── database/     # 数据库初始化
│   │   ├── logger/       # 日志工具
│   │   └── payment/      # 支付相关
│   ├── router/           # 路由配置
│   ├── service/          # 业务逻辑层
│   ├── types/            # 类型定义
│   └── utils/            # 工具函数
├── logs/                  # 日志文件目录
├── main.go                # 程序入口
├── embed.go               # 静态资源嵌入
├── Makefile               # 构建脚本
├── build.sh               # Linux/macOS 构建脚本
└── build.ps1              # Windows 构建脚本
```

## 快速开始

### 方式一：Docker 部署（推荐）

使用 Docker 是最简单的部署方式，无需安装 Go 和 Node.js 环境。

```bash
# 1. 准备配置文件
cp config/config.example.yaml config/config.yaml
# 编辑 config/config.yaml，修改数据库连接等配置

# 2. 使用 docker-compose 一键启动（包含数据库）
docker-compose up -d

# 3. 查看日志
docker-compose logs -f app
```

详细的 Docker 部署说明请参考 [Docker 部署文档](docs/DOCKER.md)

### 方式二：本地开发

### 1. 安装依赖
```bash
go mod download
```

### 2. 配置环境变量
```bash
# 复制配置模板
cp config/config.example.yaml config/config.yaml

# 编辑配置文件，修改数据库、微信等配置
# 详细配置说明请参考 docs/CONFIG.md
```

**配置要点：**
- 修改 `wechat.base_url` 为你的域名
- 配置数据库连接信息
- 设置 JWT 密钥（建议使用随机字符串）

详细配置说明请参考 [配置文档](docs/CONFIG.md)

### 3. 构建项目

**使用 Makefile（推荐）:**
```bash
make build        # 构建当前平台
make build-all    # 构建所有平台
make run          # 构建并运行
```

**使用构建脚本:**
```bash
# Linux/macOS（首次使用需要添加执行权限）
chmod +x docker-build.sh
./docker-build.sh

# Windows PowerShell
.\build.ps1
```

详细构建说明请参考 [构建文档](docs/BUILD.md)

### 4. 运行项目
```bash
# 使用默认配置
./simpleexam

# 指定配置文件
./simpleexam -c config/config.yaml

# 查看版本
./simpleexam --version

# 重置管理员密码
./simpleexam reset-password -u admin -p newpassword
```

启动时会显示 ASCII 艺术横幅和版本信息：
```
   _____ _                 _        ______                     
  / ____(_)               | |      |  ____|                    
 | (___  _ _ __ ___  _ __ | | ___  | |__  __  ____ _ _ __ ___  
  \___ \| | '_ ' _ \| '_ \| |/ _ \ |  __| \ \/ / _' | '_ ' _ \ 
  ____) | | | | | | | |_) | |  __/ | |____ >  < (_| | | | | | |
 |_____/|_|_| |_| |_| .__/|_|\___| |______/_/\_\__,_|_| |_| |_|
                    | |                                         
                    |_|                                         

  Version:     v0.1.0
  Commit:      a1b2c3d
  Build Time:  2026-02-11 10:30:00
  Go Version:  go1.21.0
  OS/Arch:     windows/amd64
```

## API 文档

### 健康检查接口

- `GET /api/v1/health` - 健康检查接口

```bash
curl http://localhost:8080/api/v1/health
```

响应示例：
```json
{
  "status": "ok"
}
```

### 其他 API

