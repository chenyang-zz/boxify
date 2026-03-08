# Boxify Makefile
# 基于 Wails v3 的跨平台数据库管理应用

.PHONY: help dev build sync-build-assets build-macos-app build-macos-app-universal refresh-icons package-macos package-macos-universal run-macos-app clean install frontend-install frontend-dev frontend-build test format tidy lint release-tag release-auto-tag release-undo-version

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

release-auto-tag: ## 自动递增版本并推送标签
	@PART="$(PART)"; \
	if [ -z "$$PART" ]; then PART="patch"; fi; \
	if ! (git diff --quiet && git diff --cached --quiet); then \
		echo "$(COLOR_YELLOW)工作区有未提交改动，请先提交后再发布$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	command -v node >/dev/null 2>&1 || (echo "$(COLOR_YELLOW)未安装 Node.js$(COLOR_RESET)" && exit 1); \
	\
	MAX_REMOTE_VERSION=$$(git tag -l "v*.*.*" | sed 's/^v//' | sort -V | tail -n1); \
	if [ -z "$$MAX_REMOTE_VERSION" ]; then MAX_REMOTE_VERSION="0.0.0"; fi; \
	\
	NEW_VERSION=$$(PART="$$PART" MAX_REMOTE_VERSION="$$MAX_REMOTE_VERSION" node scripts/bump-version.js); \
	\
	echo "$(COLOR_GREEN)版本已更新为: $$NEW_VERSION$(COLOR_RESET)"; \
	\
	git add frontend/package.json; \
	git commit -m "chore(release): bump version to v$$NEW_VERSION"; \
	git push origin HEAD; \
	\
	$(MAKE) release-tag VERSION="$$NEW_VERSION"

release-undo-version: ## 撤销最近一次版本号修改（按目标版本批量删除更高版本的 Release/Tag，回退 package.json，清空 RELEASE_NOTES.md，移除 CHANGELOG 更高版本条目）
	@command -v node >/dev/null 2>&1 || (echo "$(COLOR_YELLOW)未安装 Node.js$(COLOR_RESET)" && exit 1); \
	CURRENT_VERSION=$$(node -e 'const fs=require("fs");const p=JSON.parse(fs.readFileSync("frontend/package.json","utf8"));process.stdout.write(p.version||"")'); \
	\
	REMOTE_VERSION=""; \
	if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then \
		REMOTE_VERSION=$$(gh release list --limit 1 --json tagName --jq '.[0].tagName' 2>/dev/null | sed 's/^v//'); \
	fi; \
	\
	VERSION_TO_UNDO="$$CURRENT_VERSION"; \
	if [ -n "$$REMOTE_VERSION" ]; then \
		REMOTE_MAJOR=$$(echo "$$REMOTE_VERSION" | cut -d. -f1); \
		REMOTE_MINOR=$$(echo "$$REMOTE_VERSION" | cut -d. -f2); \
		REMOTE_PATCH=$$(echo "$$REMOTE_VERSION" | cut -d. -f3); \
		CURRENT_MAJOR=$$(echo "$$CURRENT_VERSION" | cut -d. -f1); \
		CURRENT_MINOR=$$(echo "$$CURRENT_VERSION" | cut -d. -f2); \
		CURRENT_PATCH=$$(echo "$$CURRENT_VERSION" | cut -d. -f3); \
		IS_REMOTE_NEWER=0; \
		if [ "$$REMOTE_MAJOR" -gt "$$CURRENT_MAJOR" ] 2>/dev/null; then \
			IS_REMOTE_NEWER=1; \
		elif [ "$$REMOTE_MAJOR" -eq "$$CURRENT_MAJOR" ] 2>/dev/null; then \
			if [ "$$REMOTE_MINOR" -gt "$$CURRENT_MINOR" ] 2>/dev/null; then \
				IS_REMOTE_NEWER=1; \
			elif [ "$$REMOTE_MINOR" -eq "$$CURRENT_MINOR" ] 2>/dev/null; then \
				if [ "$$REMOTE_PATCH" -gt "$$CURRENT_PATCH" ] 2>/dev/null; then \
					IS_REMOTE_NEWER=1; \
				fi; \
			fi; \
		fi; \
		if [ "$$IS_REMOTE_NEWER" -eq 1 ]; then \
			echo "$(COLOR_YELLOW)检测到远程版本 $$REMOTE_VERSION 高于本地 $$CURRENT_VERSION$(COLOR_RESET)"; \
			VERSION_TO_UNDO="$$REMOTE_VERSION"; \
		elif [ "$$REMOTE_VERSION" = "$$CURRENT_VERSION" ]; then \
			echo "$(COLOR_BLUE)检测到远程版本 $$REMOTE_VERSION 与本地一致$(COLOR_RESET)"; \
		fi; \
	fi; \
	\
	TARGET_COMMIT=""; \
	SEEN_TARGET=0; \
	for COMMIT in $$(git log --format=%H -- frontend/package.json); do \
		CANDIDATE_VERSION=$$(git show "$${COMMIT}:frontend/package.json" | node -e 'let s="";process.stdin.on("data",d=>s+=d);process.stdin.on("end",()=>{const p=JSON.parse(s);process.stdout.write(p.version||"")})'); \
		if [ "$$SEEN_TARGET" -eq 0 ]; then \
			if [ "$$CANDIDATE_VERSION" = "$$VERSION_TO_UNDO" ]; then \
				SEEN_TARGET=1; \
			fi; \
			continue; \
		fi; \
		if [ "$$CANDIDATE_VERSION" != "$$VERSION_TO_UNDO" ]; then \
			TARGET_COMMIT="$$COMMIT"; \
			break; \
		fi; \
	done; \
	if [ "$$SEEN_TARGET" -eq 0 ]; then \
		echo "$(COLOR_YELLOW)历史中未找到要撤销的版本: $$VERSION_TO_UNDO$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	if [ -z "$$TARGET_COMMIT" ]; then \
		echo "$(COLOR_YELLOW)未找到可回退的更早版本（当前: $$VERSION_TO_UNDO）$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	git show "$${TARGET_COMMIT}:frontend/package.json" > frontend/package.json; \
	NEW_VERSION=$$(node -e 'const fs=require("fs");const p=JSON.parse(fs.readFileSync("frontend/package.json","utf8"));process.stdout.write(p.version||"")'); \
	: > RELEASE_NOTES.md; \
	CHANGELOG_REMOVED=0; \
	if [ -f CHANGELOG.md ]; then \
		if NEW_VERSION="$$NEW_VERSION" node -e 'const fs=require("fs");const path="CHANGELOG.md";const target=(process.env.NEW_VERSION||"").trim();const raw=fs.readFileSync(path,"utf8");const lines=raw.split(/\r?\n/);const parse=(v)=>{const m=v.match(/^(\d+)\.(\d+)\.(\d+)$$/);if(!m)return null;return [+m[1],+m[2],+m[3]]};const cmp=(a,b)=>{for(let i=0;i<3;i++){if(a[i]!==b[i])return a[i]-b[i]}return 0};const t=parse(target);if(!t){process.exit(3)};let removed=0;let skip=false;const out=[];for(const line of lines){const m=line.match(/^## \[v(\d+\.\d+\.\d+)\]/);if(m){const cur=parse(m[1]);if(cur&&cmp(cur,t)>0){skip=true;removed=1;continue}skip=false}if(!skip)out.push(line)}fs.writeFileSync(path+".tmp",out.join("\n"));process.stdout.write(String(removed));' > .changelog_removed.tmp; then \
			mv CHANGELOG.md.tmp CHANGELOG.md; \
			CHANGELOG_REMOVED=$$(cat .changelog_removed.tmp 2>/dev/null || echo 0); \
		else \
			STATUS=$$?; \
			rm -f CHANGELOG.md.tmp; \
			rm -f .changelog_removed.tmp; \
			if [ "$$STATUS" -eq 3 ]; then \
				echo "$(COLOR_YELLOW)目标版本无效，处理 CHANGELOG.md 失败: $$NEW_VERSION$(COLOR_RESET)"; \
			else \
				echo "$(COLOR_YELLOW)处理 CHANGELOG.md 失败$(COLOR_RESET)"; \
			fi; \
			exit 1; \
		fi; \
		rm -f .changelog_removed.tmp; \
	fi; \
	CLEANUP_VERSIONS=$$( \
		{ \
			git tag --list 'v*' 2>/dev/null || true; \
			if git remote get-url origin >/dev/null 2>&1; then \
				git ls-remote --tags origin 'refs/tags/v*' 2>/dev/null | awk '{print $$2}' | sed 's#refs/tags/##;s/\^{}$$//'; \
			fi; \
		} | sed '/^$$/d' | sort -u | NEW_VERSION="$$NEW_VERSION" node -e 'let s="";process.stdin.on("data",d=>s+=d);process.stdin.on("end",()=>{const target=(process.env.NEW_VERSION||"").trim();const parse=(x)=>{const c=x.replace(/^v/,"").trim();const m=c.match(/^(\d+)\.(\d+)\.(\d+)$$/);if(!m)return null;return {raw:c,val:[+m[1],+m[2],+m[3]]}};const cmp=(a,b)=>{for(let i=0;i<3;i++){if(a[i]!==b[i])return a[i]-b[i]}return 0};const t=parse(target);if(!t){process.exit(0)};const versions=[...new Set(s.split(/\r?\n/).map(v=>v.trim()).filter(Boolean))].map(parse).filter(Boolean).filter(v=>cmp(v.val,t.val)>0).sort((a,b)=>cmp(b.val,a.val));process.stdout.write(versions.map(v=>v.raw).join("\n"));});' \
	); \
	CLEANUP_COUNT=$$(printf "%s\n" "$$CLEANUP_VERSIONS" | sed '/^$$/d' | wc -l | tr -d ' '); \
	echo "$(COLOR_GREEN)已回退版本: $$VERSION_TO_UNDO -> $$NEW_VERSION$(COLOR_RESET)"; \
	echo "$(COLOR_GREEN)已清空 RELEASE_NOTES.md$(COLOR_RESET)"; \
	if [ "$$CHANGELOG_REMOVED" -eq 1 ]; then \
		echo "$(COLOR_GREEN)已移除 CHANGELOG.md 中高于 v$$NEW_VERSION 的条目$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)未找到高于 v$$NEW_VERSION 的 CHANGELOG 条目，已跳过$(COLOR_RESET)"; \
	fi; \
	if [ "$$CLEANUP_COUNT" -eq 0 ]; then \
		echo "$(COLOR_YELLOW)未发现高于 v$$NEW_VERSION 的发布产物$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_BLUE)将删除 $$CLEANUP_COUNT 个高于 v$$NEW_VERSION 的版本产物$(COLOR_RESET)"; \
		for VERSION in $$CLEANUP_VERSIONS; do \
			if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then \
				if gh release view "v$$VERSION" >/dev/null 2>&1; then \
					echo "$(COLOR_BLUE)正在删除远程 Release: v$$VERSION$(COLOR_RESET)"; \
					gh release delete "v$$VERSION" --yes >/dev/null || (echo "$(COLOR_YELLOW)删除远程 Release 失败: v$$VERSION$(COLOR_RESET)" && exit 1); \
					echo "$(COLOR_GREEN)已删除远程 Release: v$$VERSION$(COLOR_RESET)"; \
				fi; \
			fi; \
			if git remote get-url origin >/dev/null 2>&1; then \
				if git ls-remote --exit-code --tags origin "refs/tags/v$$VERSION" >/dev/null 2>&1; then \
					echo "$(COLOR_BLUE)正在删除远程 Tag: v$$VERSION$(COLOR_RESET)"; \
					git push origin ":refs/tags/v$$VERSION" >/dev/null || (echo "$(COLOR_YELLOW)删除远程 Tag 失败: v$$VERSION$(COLOR_RESET)" && exit 1); \
					echo "$(COLOR_GREEN)已删除远程 Tag: v$$VERSION$(COLOR_RESET)"; \
				fi; \
			fi; \
			if git rev-parse -q --verify "refs/tags/v$$VERSION" >/dev/null 2>&1; then \
				echo "$(COLOR_BLUE)正在删除本地 Tag: v$$VERSION$(COLOR_RESET)"; \
				git tag -d "v$$VERSION" >/dev/null || (echo "$(COLOR_YELLOW)删除本地 Tag 失败: v$$VERSION$(COLOR_RESET)" && exit 1); \
				echo "$(COLOR_GREEN)已删除本地 Tag: v$$VERSION$(COLOR_RESET)"; \
			fi; \
		done; \
	fi; \
	echo "$(COLOR_BLUE)提示: 请检查后手动提交变更$(COLOR_RESET)"

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
