# 考试系统后端

基于 Gin 框架开发的在线考试系统后端

## 技术栈

- Gin: Web 框架
- GORM: ORM 框架
- JWT: 用户认证
- MySQL: 数据存储
- Redis: 缓存
- 微信SDK: 微信登录和支付

## 项目结构

```
.
├── cmd                     # 程序入口
├── config                  # 配置文件
├── internal               
│   ├── api                # API 处理器
│   ├── middleware         # 中间件
│   ├── model             # 数据模型
│   ├── repository        # 数据访问层
│   ├── service          # 业务逻辑层
│   └── pkg              # 内部公共包
├── pkg                    # 外部可用的公共包
└── scripts                # 脚本文件
```

## 快速开始

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
# Linux/macOS
./build.sh

# Windows PowerShell
.\build.ps1
```

详细构建说明请参考 [BUILD.md](BUILD.md)

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

  Version:     v0.1.2
  Commit:      a1b2c3d
  Build Time:  2024-02-11 10:30:00
  Go Version:  go1.21.0
  OS/Arch:     windows/amd64
```

## 数据库设计


## API 文档

