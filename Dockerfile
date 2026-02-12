# 前端构建阶段 - 用户端
FROM node:22-alpine AS frontend-user

WORKDIR /frontend

# 克隆并构建用户端前端
RUN apk add --no-cache git && \
    git clone --depth 1 https://github.com/SimpleExamTeam/SimpleExam-Frontend.git . && \
    npm install && \
    npm run build

# 前端构建阶段 - 管理端
FROM node:22-alpine AS frontend-admin

WORKDIR /admin

# 克隆并构建管理端前端
RUN apk add --no-cache git && \
    git clone --depth 1 https://github.com/SimpleExamTeam/SimpleExam-Admin.git . && \
    cp .env.example .env && \
    npm install && \
    npm run build

# Go 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装构建依赖
RUN apk add --no-cache git make

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 从前端构建阶段复制编译好的静态文件
COPY --from=frontend-user /frontend/dist ./public/user
COPY --from=frontend-admin /admin/dist ./public/admin

# 构建应用（嵌入静态文件）
RUN BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') && \
    COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown') && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-s -w -X 'main.Version=v0.1.2' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.CommitHash=${COMMIT_HASH}'" \
    -o simpleexam .

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/simpleexam .

# 创建必要的目录并设置权限
RUN mkdir -p config logs certs && \
    chown -R appuser:appuser /app && \
    chmod -R 755 /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 启动应用
ENTRYPOINT ["./simpleexam"]
CMD ["-c", "config/config.yaml"]
