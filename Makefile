# Boxify Makefile
# 基于 Wails v3 的跨平台数据库管理应用

.PHONY: help dev build sync-build-assets build-macos-app build-macos-app-universal refresh-icons package-macos package-macos-universal run-macos-app clean install frontend-install frontend-dev frontend-build test format tidy lint release-tag release-auto-tag

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

refresh-icons: ## 根据 build/appicon.png 重新生成图标资源
	@echo "$(COLOR_GREEN)🖼️ 重新生成应用图标...$(COLOR_RESET)"
	@$(WAILS) task common:generate:icons

sync-build-assets: ## 同步构建资产（修正 Info.plist 可执行名等元数据）
	@echo "$(COLOR_GREEN)🧩 同步构建资产...$(COLOR_RESET)"
	@$(WAILS) task common:update:build-assets APP_NAME=$(APP_NAME)

package-macos: sync-build-assets ## 打包 macOS .app（包含应用图标）
	@echo "$(COLOR_GREEN)📦 打包 macOS 应用...$(COLOR_RESET)"
	@$(WAILS) task darwin:package

package-macos-universal: sync-build-assets ## 打包 macOS Universal .app（arm64 + amd64）
	@echo "$(COLOR_GREEN)📦 打包 macOS Universal 应用...$(COLOR_RESET)"
	@$(WAILS) task darwin:package:universal

build-macos-app: refresh-icons package-macos ## 重新生成图标并打包 macOS .app

build-macos-app-universal: refresh-icons package-macos-universal ## 重新生成图标并打包 Universal .app

run-macos-app: ## 启动已打包的 macOS .app（自动移除隔离属性）
	@echo "$(COLOR_GREEN)▶️ 启动 macOS 应用...$(COLOR_RESET)"
	@test -d bin/boxify.app || (echo "$(COLOR_YELLOW)未找到 bin/boxify.app，请先执行 make build-macos-app$(COLOR_RESET)" && exit 1)
	@xattr -dr com.apple.quarantine bin/boxify.app 2>/dev/null || true
	@open bin/boxify.app

release-tag: ## 仅创建并推送版本标签（触发 GitHub Workflow 发布，示例: make release-tag VERSION=0.0.6）
	@if [ -z "$(VERSION)" ]; then \
		echo "$(COLOR_YELLOW)请提供版本号: make release-tag VERSION=0.0.6$(COLOR_RESET)"; \
		exit 1; \
	fi
	@git diff --quiet && git diff --cached --quiet || (echo "$(COLOR_YELLOW)工作区有未提交改动，请先提交后再发布$(COLOR_RESET)" && exit 1)
	@git rev-parse -q --verify "refs/tags/v$(VERSION)" >/dev/null && (echo "$(COLOR_YELLOW)本地已存在 tag v$(VERSION)$(COLOR_RESET)" && exit 1) || true
	@git ls-remote --tags origin "refs/tags/v$(VERSION)" | grep -q "refs/tags/v$(VERSION)$$" && (echo "$(COLOR_YELLOW)远端已存在 tag v$(VERSION)$(COLOR_RESET)" && exit 1) || true
	@git tag -a "v$(VERSION)" -m "release v$(VERSION)"
	@git push origin "v$(VERSION)"
	@echo "$(COLOR_GREEN)✅ 已推送 tag: v$(VERSION)（GitHub Actions 将自动发布）$(COLOR_RESET)"

release-auto-tag: ## 自动递增版本并推送标签（由 GitHub Workflow 负责构建与发布）
	@PART="$(PART)"; \
	if [ -z "$$PART" ]; then PART="patch"; fi; \
	if ! (git diff --quiet && git diff --cached --quiet); then \
		echo "$(COLOR_YELLOW)工作区有未提交改动，请先提交后再发布$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	command -v node >/dev/null 2>&1 || (echo "$(COLOR_YELLOW)未安装 Node.js，无法自动递增 frontend/package.json 版本$(COLOR_RESET)" && exit 1); \
	MAX_REMOTE_VERSION=$$(git ls-remote --tags --refs origin "v*.*.*" | awk -F'/' '{print $$3}' | sed 's/^v//' | sort -V | tail -n1); \
	if [ -z "$$MAX_REMOTE_VERSION" ]; then MAX_REMOTE_VERSION="0.0.0"; fi; \
	NEW_VERSION=$$(PART="$$PART" MAX_REMOTE_VERSION="$$MAX_REMOTE_VERSION" node -e 'const fs=require("fs"); const file="frontend/package.json"; const part=process.env.PART||"patch"; const maxRemote=process.env.MAX_REMOTE_VERSION||"0.0.0"; const pkg=JSON.parse(fs.readFileSync(file,"utf8")); const parse=v=>(v||"0.0.0").split(".").map(n=>parseInt(n,10)||0).concat([0,0,0]).slice(0,3); const cmp=(a,b)=>a[0]-b[0]||a[1]-b[1]||a[2]-b[2]; const seg=cmp(parse(pkg.version),parse(maxRemote))>=0?parse(pkg.version):parse(maxRemote); if(part==="major"){seg[0]++; seg[1]=0; seg[2]=0;} else if(part==="minor"){seg[1]++; seg[2]=0;} else {seg[2]++;} const next=seg.join("."); pkg.version=next; fs.writeFileSync(file, JSON.stringify(pkg,null,2)+"\n"); process.stdout.write(next);'); \
	echo "$(COLOR_GREEN)版本已更新为: $$NEW_VERSION$(COLOR_RESET)"; \
	git add frontend/package.json; \
	git commit -m "🔧 chore(release): bump version to v$$NEW_VERSION"; \
	git push origin HEAD; \
	$(MAKE) release-tag VERSION="$$NEW_VERSION"

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
