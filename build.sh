#!/bin/bash

# 检查是否存在app目录
if [ ! -d "app" ]; then
  mkdir -p public/user
  mkdir -p public/admin
  echo "创建了app目录，请将前端文件放入app/user和app/admin目录中"
fi

# 构建前端（如果有package.json）
if [ -f "public/user/package.json" ]; then
  echo "构建用户端前端..."
  cd public/user && npm install && npm run build && cd ../../
fi

if [ -f "public/admin/package.json" ]; then
  echo "构建管理端前端..."
  cd public/admin && npm install && npm run build && cd ../../
fi

# 执行go generate生成嵌入的静态资源
echo "生成嵌入的静态资源..."
go generate ./...

# 构建应用
echo "构建应用..."
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}')

# 设置ldflags以包含版本信息
LD_FLAGS="-X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GoVersion=${GO_VERSION}'"

go build -ldflags "${LD_FLAGS}" -o simpleexam main.go embed.go

echo "构建完成！应用程序已生成: SimpleExam"
echo "应用版本: ${VERSION}, 构建时间: ${BUILD_TIME}, Go版本: ${GO_VERSION}"
echo ""
echo "使用方法: "
echo "  ./simpleexam             - 使用默认配置文件 (config.yaml 或 config/config.yaml)"
echo "  ./simpleexam -c 配置文件路径  - 使用指定的配置文件"
echo "  ./simpleexam -h         - 显示帮助信息"
echo "  ./simpleexam -v         - 显示版本信息" 