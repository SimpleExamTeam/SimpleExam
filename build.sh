#!/bin/bash

# 检查并构建前端
build_frontend() {
  local dir=$1
  local repo=$2
  local name=$3
  
  # 检查目录是否为空或不存在
  if [ ! -d "$dir" ] || [ -z "$(ls -A $dir 2>/dev/null)" ]; then
    echo "检测到 $dir 目录为空或不存在，开始从源码构建 $name..."
    
    # 克隆仓库
    echo "克隆仓库: $repo"
    git clone --depth 1 "$repo" "${name}-temp"
    
    # 进入目录并构建
    cd "${name}-temp"
    echo "安装依赖..."
    npm install
    echo "构建 $name..."
    npm run build
    cd ..
    
    # 复制构建产物
    mkdir -p "$dir"
    echo "复制构建产物到 $dir"
    cp -r "${name}-temp/dist/"* "$dir/"
    
    # 清理临时目录
    rm -rf "${name}-temp"
    echo "$name 构建完成！"
  else
    echo "$dir 目录已存在且不为空，跳过构建"
  fi
}

# 构建用户端前端
build_frontend "public/user" "https://github.com/SimpleExamTeam/SimpleExam-Frontend.git" "用户端"

# 构建管理端前端
build_frontend "public/admin" "https://github.com/SimpleExamTeam/SimpleExam-Admin.git" "管理端"

# 执行go generate生成嵌入的静态资源
echo "生成嵌入的静态资源..."
go generate ./...

# 构建应用
echo "构建应用..."
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.1.2")
COMMIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')

# 设置ldflags以包含版本信息
LD_FLAGS="-X 'main.Version=${VERSION}' -X 'main.CommitHash=${COMMIT_HASH}' -X 'main.BuildTime=${BUILD_TIME}' -s -w"

go build -ldflags "${LD_FLAGS}" -o simpleexam main.go embed.go

echo "构建完成！应用程序已生成: simpleexam"
echo "应用版本: ${VERSION}"
echo "提交哈希: ${COMMIT_HASH:0:7}"
echo "构建时间: ${BUILD_TIME}"
echo ""
echo "使用方法: "
echo "  ./simpleexam             - 使用默认配置文件 (config.yaml 或 config/config.yaml)"
echo "  ./simpleexam -c 配置文件路径  - 使用指定的配置文件"
echo "  ./simpleexam -h         - 显示帮助信息"
echo "  ./simpleexam -v         - 显示版本信息" 