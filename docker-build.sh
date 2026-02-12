#!/bin/bash

# Docker 构建脚本
# 用法: ./docker-build.sh [选项]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认值
IMAGE_NAME="simpleexam"
IMAGE_TAG="latest"
BUILD_TYPE="full"  # full 或 local
PUSH_IMAGE=false
REGISTRY=""

# 显示帮助信息
show_help() {
    cat << EOF
用法: $0 [选项]

选项:
    -h, --help              显示此帮助信息
    -t, --tag TAG           设置镜像标签 (默认: latest)
    -n, --name NAME         设置镜像名称 (默认: simpleexam)
    -l, --local             使用本地已构建的前端文件 (Dockerfile.local)
    -p, --push              构建后推送到镜像仓库
    -r, --registry URL      镜像仓库地址 (例如: docker.io/username)

示例:
    $0                                      # 完整构建（包含前端）
    $0 --local                              # 使用本地前端文件构建
    $0 --tag v1.0.0                         # 指定标签
    $0 --push --registry docker.io/myuser   # 构建并推送到 Docker Hub

EOF
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -l|--local)
            BUILD_TYPE="local"
            shift
            ;;
        -p|--push)
            PUSH_IMAGE=true
            shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}错误: 未知选项 $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# 构建完整镜像名称
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
else
    FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"
fi

echo -e "${GREEN}=== SimpleExam Docker 构建 ===${NC}"
echo "镜像名称: $FULL_IMAGE_NAME"
echo "构建类型: $BUILD_TYPE"
echo ""

# 检查本地构建时前端文件是否存在
if [ "$BUILD_TYPE" = "local" ]; then
    echo -e "${YELLOW}检查本地前端文件...${NC}"
    
    if [ ! -d "public/user" ] || [ -z "$(ls -A public/user 2>/dev/null)" ]; then
        echo -e "${RED}错误: public/user 目录不存在或为空${NC}"
        echo "请先构建前端或使用完整构建模式（不加 --local 参数）"
        exit 1
    fi
    
    if [ ! -d "public/admin" ] || [ -z "$(ls -A public/admin 2>/dev/null)" ]; then
        echo -e "${RED}错误: public/admin 目录不存在或为空${NC}"
        echo "请先构建前端或使用完整构建模式（不加 --local 参数）"
        exit 1
    fi
    
    echo -e "${GREEN}✓ 前端文件检查通过${NC}"
    DOCKERFILE="Dockerfile.local"
else
    echo -e "${YELLOW}将从源码构建前端（需要较长时间）...${NC}"
    DOCKERFILE="Dockerfile"
fi

# 开始构建
echo -e "${GREEN}开始构建镜像...${NC}"
docker build -f "$DOCKERFILE" -t "$FULL_IMAGE_NAME" .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 镜像构建成功: $FULL_IMAGE_NAME${NC}"
    
    # 显示镜像信息
    echo ""
    echo "镜像信息:"
    docker images "$FULL_IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
    
    # 推送镜像
    if [ "$PUSH_IMAGE" = true ]; then
        echo ""
        echo -e "${GREEN}推送镜像到仓库...${NC}"
        docker push "$FULL_IMAGE_NAME"
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ 镜像推送成功${NC}"
        else
            echo -e "${RED}✗ 镜像推送失败${NC}"
            exit 1
        fi
    fi
    
    echo ""
    echo -e "${GREEN}构建完成！${NC}"
    echo ""
    echo "运行容器:"
    echo "  docker run -d -p 8080:8080 -v \$(pwd)/config/config.yaml:/app/config/config.yaml:ro $FULL_IMAGE_NAME"
    echo ""
    echo "或使用 docker-compose:"
    echo "  docker-compose up -d"
    
else
    echo -e "${RED}✗ 镜像构建失败${NC}"
    exit 1
fi
