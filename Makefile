.PHONY: build run clean test help

# 版本信息
VERSION ?= v0.1.2
COMMIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
OUTPUT := simpleexam

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
	@echo "  make help           - Show this help message"
