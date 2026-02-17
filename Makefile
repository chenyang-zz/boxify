# Boxify Makefile
# 基于 Wails v3 的跨平台数据库管理应用

.PHONY: help dev build clean install frontend-install frontend-dev frontend-build test format tidy lint

# 默认目标
.DEFAULT_GOAL := help

# Wails 命令
WAILS := wails3

# 应用信息
APP_NAME := boxify
BUILD_DIR := build
FRONTEND_DIR := frontend

# 颜色输出
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

help: ## 显示帮助信息
	@echo "$(COLOR_BOLD)Boxify - 可用命令:$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

dev: ## 启动开发模式（热重载）
	@echo "$(COLOR_GREEN)🚀 启动开发模式...$(COLOR_RESET)"
	@$(WAILS) dev

build: ## 构建生产版本
	@echo "$(COLOR_GREEN)🔨 构建生产版本...$(COLOR_RESET)"
	@$(WAILS) build

build-release: ## 构建发布版本
	@echo "$(COLOR_GREEN)📦 构建发布版本...$(COLOR_RESET)"
	@if [ -f ./script/build-release.sh ]; then \
		./script/build-release.sh; \
	else \
		$(WAILS) build; \
	fi

clean: ## 清理构建文件
	@echo "$(COLOR_YELLOW)🧹 清理构建文件...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -rf $(FRONTEND_DIR)/wailsjs
	@echo "$(COLOR_GREEN)✓ 清理完成$(COLOR_RESET)"

install: ## 安装所有依赖（Go + 前端）
	@echo "$(COLOR_GREEN)📥 安装依赖...$(COLOR_RESET)"
	@make go-install
	@make frontend-install
	@echo "$(COLOR_GREEN)✓ 依赖安装完成$(COLOR_RESET)"

go-install: ## 安装 Go 依赖
	@echo "$(COLOR_BLUE)安装 Go 依赖...$(COLOR_RESET)"
	@go mod download
	@go mod verify

go-tidy: ## 整理 Go 依赖
	@echo "$(COLOR_BLUE)整理 Go 依赖...$(COLOR_RESET)"
	@go mod tidy

frontend-install: ## 安装前端依赖
	@echo "$(COLOR_BLUE)安装前端依赖...$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && pnpm install

frontend-dev: ## 仅启动前端开发服务器
	@echo "$(COLOR_GREEN)🎨 启动前端开发服务器...$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && pnpm run dev

frontend-build: ## 仅构建前端
	@echo "$(COLOR_GREEN)🎨 构建前端...$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && pnpm run build

test: ## 运行所有测试
	@echo "$(COLOR_GREEN)🧪 运行测试...$(COLOR_RESET)"
	@go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(COLOR_GREEN)📊 生成测试覆盖率...$(COLOR_RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ 覆盖率报告已生成: coverage.html$(COLOR_RESET)"

format: ## 格式化所有代码
	@echo "$(COLOR_BLUE)格式化代码...$(COLOR_RESET)"
	@echo "  → Go 代码"
	@go fmt ./...
	@echo "  → Goimports"
	@which goimports > /dev/null 2>&1 && goimports -w . || echo "    goimports 未安装，跳过"
	@echo "$(COLOR_GREEN)✓ 代码格式化完成$(COLOR_RESET)"

lint: ## 运行代码检查
	@echo "$(COLOR_BLUE)🔍 运行代码检查...$(COLOR_RESET)"
	@echo "  → Go 代码"
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "    golangci-lint 未安装，跳过"
	@echo "  → 前端代码"
	@cd $(FRONTEND_DIR) && which pnpm > /dev/null 2>&1 && pnpm run lint || echo "    lint 命令未配置"

tidy: go-tidy ## 整理依赖

deps-update: ## 更新所有依赖
	@echo "$(COLOR_BLUE)更新依赖...$(COLOR_RESET)"
	@echo "  → Go 依赖"
	@go get -u ./...
	@go mod tidy
	@echo "  → 前端依赖"
	@cd $(FRONTEND_DIR) && pnpm update

clean-cache: ## 清理缓存
	@echo "$(COLOR_YELLOW)🧹 清理缓存...$(COLOR_RESET)"
	@go clean -cache -testcache
	@cd $(FRONTEND_DIR) && rm -rf node_modules/.vite

check: ## 检查项目状态
	@echo "$(COLOR_BOLD)📋 项目状态:$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Go 模块:$(COLOR_RESET)"
	@go version
	@go mod verify
	@echo ""
	@echo "$(COLOR_BOLD)前端:$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && pnpm --version
	@echo ""
	@echo "$(COLOR_BOLD)Git 状态:$(COLOR_RESET)"
	@git status -sb

init: ## 初始化开发环境
	@echo "$(COLOR_GREEN)🎯 初始化开发环境...$(COLOR_RESET)"
	@make install
	@echo ""
	@echo "$(COLOR_GREEN)✓ 开发环境初始化完成$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)快速开始:$(COLOR_RESET)"
	@echo "  make dev    # 启动开发模式"
	@echo "  make build  # 构建应用"
	@echo ""
