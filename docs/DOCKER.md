# Docker 部署指南

本文档介绍如何使用 Docker 部署 Simple Exam 考试系统。

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+

## 构建说明

项目提供两种 Docker 构建方式：

### 1. 完整构建（推荐）
使用 `Dockerfile`，会自动从 GitHub 克隆并编译前端代码，然后嵌入到 Go 二进制文件中。

**优点**: 一键构建，无需手动准备前端文件
**缺点**: 首次构建时间较长（约 5-10 分钟）

### 2. 本地构建
使用 `Dockerfile.local`，使用本地已构建的 `public/` 目录。

**优点**: 构建速度快（约 1-2 分钟）
**缺点**: 需要先手动构建前端

## 快速开始

### 1. 使用 Docker Compose（推荐）

这是最简单的部署方式，会自动启动应用和 MySQL 数据库。

```bash
# 1. 准备配置文件
cp config/config.example.yaml config/config.yaml

# 2. 编辑配置文件，修改数据库连接为：
#    host: mysql
#    port: "3306"
#    username: admin
#    password: 123456
#    dbname: simple_exam

# 3. 启动服务（首次构建需要 5-10 分钟）
docker-compose up -d

# 4. 查看日志
docker-compose logs -f app

# 5. 停止服务
docker-compose down

# 6. 停止并删除数据
docker-compose down -v
```

### 2. 使用构建脚本

项目提供了便捷的构建脚本：

**Linux/macOS:**
```bash
# 完整构建（包含前端编译）
./docker-build.sh

# 使用本地前端文件快速构建
./docker-build.sh --local

# 指定标签
./docker-build.sh --tag v1.0.0

# 构建并推送到镜像仓库
./docker-build.sh --push --registry docker.io/username
```

**Windows PowerShell:**
```powershell
# 完整构建（包含前端编译）
.\docker-build.ps1

# 使用本地前端文件快速构建
.\docker-build.ps1 -Local

# 指定标签
.\docker-build.ps1 -Tag v1.0.0

# 构建并推送到镜像仓库
.\docker-build.ps1 -Push -Registry docker.io/username
```

### 3. 手动构建

**完整构建（包含前端）:**
```bash
# 构建镜像（需要 5-10 分钟）
docker build -t simpleexam:latest .

# 运行容器
docker run -d \
  --name simpleexam \
  -p 8080:8080 \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v $(pwd)/certs:/app/certs:ro \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  simpleexam:latest
```

**本地构建（使用已有前端）:**
```bash
# 确保 public/user 和 public/admin 目录存在且包含前端文件

# 构建镜像（快速，约 1-2 分钟）
docker build -f Dockerfile.local -t simpleexam:latest .

# 运行容器
docker run -d \
  --name simpleexam \
  -p 8080:8080 \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v $(pwd)/certs:/app/certs:ro \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  simpleexam:latest
```

## 前端构建说明

如果使用 `Dockerfile.local` 进行本地构建，需要先准备前端文件：

### 方法 1: 使用 GitHub Actions 产物
从 GitHub Actions 构建产物中下载 `frontend-dist` 并解压到 `public/` 目录。

### 方法 2: 手动构建前端

**构建用户端:**
```bash
git clone https://github.com/SimpleExamTeam/SimpleExam-Frontend.git frontend-temp
cd frontend-temp
npm install
npm run build
cd ..
mkdir -p public/user
cp -r frontend-temp/dist/* public/user/
rm -rf frontend-temp
```

**构建管理端:**
```bash
git clone https://github.com/SimpleExamTeam/SimpleExam-Admin.git admin-temp
cd admin-temp
cp .env.example .env
npm install
npm run build
cd ..
mkdir -p public/admin
cp -r admin-temp/dist/* public/admin/
rm -rf admin-temp
```

### 方法 3: 使用 Docker 多阶段构建
直接使用 `Dockerfile`，它会自动完成前端构建。

## 配置说明

### 数据库配置

使用 Docker Compose 时，数据库配置应为：

```yaml
database:
  driver: "mysql"
  host: "mysql"          # 使用服务名
  port: "3306"
  username: "admin"
  password: "123456"
  dbname: "simple_exam"
```

### 端口映射

默认映射：
- 应用端口：8080:8080
- MySQL 端口：3306:3306

修改端口映射：
```yaml
services:
  app:
    ports:
      - "9090:8080"  # 将应用映射到主机的 9090 端口
```

### 数据持久化

Docker Compose 会自动创建以下卷：
- `mysql_data`: MySQL 数据目录
- `./logs`: 应用日志目录（挂载到主机）

### 证书配置

将微信支付证书放在 `certs/` 目录下：
```
certs/
└── apiclient_cert.p12
```

## 常用命令

### 查看运行状态
```bash
docker-compose ps
```

### 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 只查看应用日志
docker-compose logs -f app

# 只查看数据库日志
docker-compose logs -f mysql
```

### 重启服务
```bash
# 重启所有服务
docker-compose restart

# 只重启应用
docker-compose restart app
```

### 进入容器
```bash
# 进入应用容器
docker-compose exec app sh

# 进入数据库容器
docker-compose exec mysql bash
```

### 重置管理员密码
```bash
docker-compose exec app ./simpleexam reset-password -u admin -p newpassword
```

### 数据库备份
```bash
# 备份数据库
docker-compose exec mysql mysqldump -uadmin -p123456 simple_exam > backup.sql

# 恢复数据库
docker-compose exec -T mysql mysql -uadmin -p123456 simple_exam < backup.sql
```

## 生产环境建议

### 1. 使用环境变量

创建 `.env` 文件：
```env
MYSQL_ROOT_PASSWORD=your_secure_root_password
MYSQL_PASSWORD=your_secure_password
JWT_SECRET=your_secure_jwt_secret
```

修改 `docker-compose.yml` 使用环境变量：
```yaml
environment:
  MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
  MYSQL_PASSWORD: ${MYSQL_PASSWORD}
```

### 2. 使用外部数据库

如果使用云数据库或外部 MySQL，只需启动应用服务：
```bash
docker-compose up -d app
```

### 3. 配置反向代理

使用 Nginx 作为反向代理：
```nginx
server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 4. 健康检查

应用内置健康检查端点：
```bash
curl http://localhost:8080/health
```

### 5. 日志管理

配置日志轮转，避免日志文件过大：
```yaml
log:
  max_size: 100      # 单个文件最大 100MB
  max_backups: 3     # 保留 3 个备份
  max_age: 28        # 保留 28 天
  compress: true     # 压缩旧日志
```

## 故障排查

### 应用无法连接数据库

1. 检查数据库是否启动：
```bash
docker-compose ps mysql
```

2. 检查数据库健康状态：
```bash
docker-compose exec mysql mysqladmin ping -h localhost
```

3. 检查配置文件中的数据库地址是否为 `mysql`

### 端口冲突

如果端口被占用，修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "8081:8080"  # 使用其他端口
```

### 查看详细错误

```bash
# 查看应用日志
docker-compose logs app

# 查看容器内日志文件
docker-compose exec app cat logs/app.log
```

## 镜像优化

当前 Dockerfile 使用多阶段构建，最终镜像大小约 20-30MB（不含前端资源）。

### 构建时间对比

| 构建方式 | 首次构建 | 增量构建 | 镜像大小 |
|---------|---------|---------|---------|
| 完整构建 (Dockerfile) | 5-10 分钟 | 2-5 分钟 | ~50-80MB |
| 本地构建 (Dockerfile.local) | 1-2 分钟 | 30-60 秒 | ~50-80MB |

### 优化建议

1. **使用 Docker 构建缓存**: Docker 会缓存各个构建层，修改代码后重新构建会更快
2. **使用本地构建**: 如果频繁修改后端代码，建议先构建前端，然后使用 `Dockerfile.local`
3. **使用 .dockerignore**: 已配置排除不必要的文件，减小构建上下文

查看镜像大小：
```bash
docker images simpleexam
```

## 更新部署

```bash
# 1. 拉取最新代码
git pull

# 2. 重新构建并启动
docker-compose up -d --build

# 3. 查看日志确认启动成功
docker-compose logs -f app
```
