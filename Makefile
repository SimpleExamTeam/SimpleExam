.PHONY: build run clean test help docker-build docker-build-local docker-run docker-stop docker-compose-up docker-compose-down

# 版本信息
VERSION ?= v0.1.2
COMMIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
OUTPUT := simpleexam
DOCKER_IMAGE := simpleexam
DOCKER_TAG := latest

# 编译标志
LDFLAGS := -X 'main.Version=$(VERSION)' \
           -X 'main.CommitHash=$(COMMIT_HASH)' \
           -X 'main.BuildTime=$(BUILD_TIME)' \
           -s -w

# 默认目标
all: build

# 构建应用
build:
	@echo "Building Simple Exam..."
	@echo "Version:     $(VERSION)"
	@echo "Commit:      $(shell echo $(COMMIT_HASH) | cut -c1-7)"
	@echo "Build Time:  $(BUILD_TIME)"
	@go generate ./...
	@go build -ldflags "$(LDFLAGS)" -o $(OUTPUT) .
	@echo "Build successful! Output: $(OUTPUT)"

# 构建 Windows 版本
build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT).exe .
	@echo "Build successful! Output: $(OUTPUT).exe"

# 构建 Linux 版本
build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT)-linux .
	@echo "Build successful! Output: $(OUTPUT)-linux"

# 构建 macOS 版本
build-darwin:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT)-darwin .
	@echo "Build successful! Output: $(OUTPUT)-darwin"

# 构建所有平台
build-all: build-windows build-linux build-darwin
	@echo "All platforms built successfully!"

# 运行应用
run: build
	@./$(OUTPUT)

# 运行应用（指定配置文件）
run-config: build
	@./$(OUTPUT) -c config/config.yaml

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -f $(OUTPUT) $(OUTPUT).exe $(OUTPUT)-linux $(OUTPUT)-darwin
	@echo "Clean complete!"

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 代码格式化
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 代码检查
vet:
	@echo "Vetting code..."
	@go vet ./...

# 显示帮助
help:
	@echo "Simple Exam Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - Build the application"
	@echo "  make build-windows  - Build for Windows"
	@echo "  make build-linux    - Build for Linux"
	@echo "  make build-darwin   - Build for macOS"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make run            - Build and run the application"
	@echo "  make run-config     - Build and run with config file"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Vet code"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-build        - Build Docker image (full build with frontend)"
	@echo "  make docker-build-local  - Build Docker image (using local frontend)"
	@echo "  make docker-run          - Run Docker container"
	@echo "  make docker-stop         - Stop and remove Docker container"
	@echo "  make docker-compose-up   - Start services with docker-compose"
	@echo "  make docker-compose-down - Stop services with docker-compose"
	@echo ""
	@echo "  make help           - Show this help message"

# Docker 命令

# 构建 Docker 镜像（完整构建，包含前端）
docker-build:
	@echo "Building Docker image (full build with frontend)..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# 构建 Docker 镜像（使用本地前端文件）
docker-build-local:
	@echo "Building Docker image (using local frontend)..."
	@if [ ! -d "public/user" ] || [ -z "$$(ls -A public/user 2>/dev/null)" ]; then \
		echo "Error: public/user directory not found or empty"; \
		exit 1; \
	fi
	@if [ ! -d "public/admin" ] || [ -z "$$(ls -A public/admin 2>/dev/null)" ]; then \
		echo "Error: public/admin directory not found or empty"; \
		exit 1; \
	fi
	@docker build -f Dockerfile.local -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# 运行 Docker 容器
docker-run:
	@echo "Running Docker container..."
	@docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 8080:8080 \
		-v $$(pwd)/config/config.yaml:/app/config/config.yaml:ro \
		-v $$(pwd)/certs:/app/certs:ro \
		-v $$(pwd)/logs:/app/logs \
		-e TZ=Asia/Shanghai \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started: $(DOCKER_IMAGE)"
	@echo "Access the application at http://localhost:8080"

# 停止并删除 Docker 容器
docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(DOCKER_IMAGE) 2>/dev/null || true
	@docker rm $(DOCKER_IMAGE) 2>/dev/null || true
	@echo "Container stopped and removed"

# 使用 docker-compose 启动服务
docker-compose-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d
	@echo "Services started"
	@echo "Access the application at http://localhost:8080"

# 使用 docker-compose 停止服务
docker-compose-down:
	@echo "Stopping services with docker-compose..."
	@docker-compose down
	@echo "Services stopped"

# 查看 Docker 日志
docker-logs:
	@docker logs -f $(DOCKER_IMAGE)

# 查看 docker-compose 日志
docker-compose-logs:
	@docker-compose logs -f app
