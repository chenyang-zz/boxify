#!/usr/bin/env bash
# ============================================================
# Boxify 一键安装脚本 (Linux/macOS)
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/chenyang-zz/boxify/main/scripts/install.sh ｜ sudo bash install.sh
#   wget -qO- https://raw.githubusercontent.com/chenyang-zz/boxify/main/scripts/install.sh | sudo bash
# ============================================================

set -e

APP_NAME="Boxify"
BINARY_NAME="boxify"
INSTALL_DIR="/opt/boxify"
REPO="chenyang-zz/boxify"
DEFAULT_VERSION="0.0.0"
MAC_MOUNT_POINT="/Volumes/BoxifyInstaller"
MAC_TMP_DMG="/tmp/boxify-installer.dmg"

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
CYAN='\033[36m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${GREEN}[Boxify]${NC} $1"; }
info() { echo -e "${CYAN}[Boxify]${NC} $1"; }
warn() { echo -e "${YELLOW}[Boxify]${NC} $1"; }
err()  { echo -e "${RED}[Boxify]${NC} $1"; exit 1; }

get_latest_version() {
    local tag=""
    local ver=""
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"

    # 方案1: 使用 GitHub API
    if command -v curl >/dev/null 2>&1; then
        tag=$(curl -fsSL --connect-timeout 10 "$api_url" 2>/dev/null | grep -m1 '"tag_name"' | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        tag=$(wget -qO- --timeout=10 "$api_url" 2>/dev/null | grep -m1 '"tag_name"' | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    fi

    # 方案2: 从 releases 页面解析 (API 限流时的备用方案)
    if [ -z "$tag" ]; then
        info "GitHub API 不可用，尝试从 releases 页面获取版本..." >&2
        local releases_url="https://github.com/${REPO}/releases/latest"
        if command -v curl >/dev/null 2>&1; then
            tag=$(curl -fsSL --connect-timeout 10 -L "$releases_url" 2>/dev/null | grep -oE '/releases/tag/v[0-9]+\.[0-9]+\.[0-9]+[a-zA-Z0-9._-]*' | head -1 | sed 's|.*/v||')
        elif command -v wget >/dev/null 2>&1; then
            tag=$(wget -qO- --timeout=10 -L "$releases_url" 2>/dev/null | grep -oE '/releases/tag/v[0-9]+\.[0-9]+\.[0-9]+[a-zA-Z0-9._-]*' | head -1 | sed 's|.*/v||')
        fi
    fi

    # 移除 'v' 前缀
    ver="${tag#v}"

    # 验证版本号格式 (支持语义化版本如 0.0.18, 1.0.0-beta 等)
    if [[ ! "$ver" =~ ^[0-9]+\.[0-9]+\.[0-9]+[a-zA-Z0-9._-]*$ ]]; then
        warn "无法获取最新版本，使用默认版本: ${DEFAULT_VERSION}" >&2
        ver=""
    fi

    echo "${ver:-$DEFAULT_VERSION}"
}

detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux) echo "linux" ;;
        darwin) echo "darwin" ;;
        *) err "不支持的操作系统: ${os}（仅支持 Linux/macOS）" ;;
    esac
}

detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) err "不支持的 CPU 架构: ${arch}（仅支持 amd64/arm64）" ;;
    esac
}

get_macos_dmg_arch_suffix() {
    local arch="$1"
    case "$arch" in
        amd64) echo "x64" ;;
        arm64) echo "arm64" ;;
        *) echo "$arch" ;;
    esac
}

get_linux_appimage_arch_suffix() {
    local arch="$1"
    case "$arch" in
        amd64) echo "x86_64" ;;
        arm64) echo "arm64" ;;
        *) echo "$arch" ;;
    esac
}

download_file() {
    local url="$1"
    local output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fL --progress-bar -o "$output" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget --show-progress -q -O "$output" "$url"
    else
        err "系统缺少 curl 或 wget，请先安装后重试。"
    fi
}

download_from_candidates() {
    local output="$1"
    shift
    local url

    for url in "$@"; do
        info "尝试下载: ${url}"
        if download_file "${url}" "${output}"; then
            log "下载成功: ${url}"
            return 0
        fi
        warn "下载失败，继续尝试其他候选资源。"
        rm -f "${output}"
    done

    return 1
}

install_linux() {
    local version="$1"
    local arch="$2"
    local appimage_arch
    local tmp_file="/tmp/boxify-installer-linux-${arch}"
    appimage_arch=$(get_linux_appimage_arch_suffix "$arch")
    local download_urls=(
        "https://github.com/${REPO}/releases/download/v${version}/${APP_NAME}-${version}-linux-${appimage_arch}.AppImage"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-linux-${appimage_arch}.AppImage"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-v${version}-linux-${arch}"
        "https://github.com/${REPO}/releases/download/v${version}/${APP_NAME}-${version}-linux-${arch}.tar.gz"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-linux-${arch}.tar.gz"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-linux-${arch}"
    )

    download_from_candidates "${tmp_file}" "${download_urls[@]}" || err "下载失败，请检查版本或网络。"

    mkdir -p "${INSTALL_DIR}"

    if tar -tzf "${tmp_file}" >/dev/null 2>&1; then
        tar -xzf "${tmp_file}" -C "${INSTALL_DIR}"
    else
        install -m 755 "${tmp_file}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    if [ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        local extracted_bin
        extracted_bin=$(find "${INSTALL_DIR}" -maxdepth 3 -type f -name "${BINARY_NAME}" | head -n 1)
        if [ -n "${extracted_bin}" ]; then
            mv "${extracted_bin}" "${INSTALL_DIR}/${BINARY_NAME}"
        else
            rm -f "${tmp_file}"
            err "安装失败：未找到可执行文件 ${BINARY_NAME}。"
        fi
    fi

    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    ln -sf "${INSTALL_DIR}/${BINARY_NAME}" /usr/local/bin/${BINARY_NAME}
    rm -f "${tmp_file}"

    log "安装完成: ${INSTALL_DIR}/${BINARY_NAME}"
    echo ""
    echo -e "  ${BOLD}启动命令${NC}: ${CYAN}${BINARY_NAME}${NC}"
    echo -e "  ${BOLD}卸载命令${NC}: sudo rm -f /usr/local/bin/${BINARY_NAME} && sudo rm -rf ${INSTALL_DIR}"
    echo ""
}

install_macos() {
    local version="$1"
    local arch="$2"
    local dmg_arch
    local tmp_file="/tmp/boxify-installer-macos-${arch}"
    local app_path="${MAC_MOUNT_POINT}/${APP_NAME}.app"
    dmg_arch=$(get_macos_dmg_arch_suffix "$arch")
    local download_urls=(
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-macos-${dmg_arch}.dmg"
        "https://github.com/${REPO}/releases/download/v${version}/${APP_NAME}-${version}-macos-${dmg_arch}.dmg"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-macos-universal.dmg"
        "https://github.com/${REPO}/releases/download/v${version}/${APP_NAME}-${version}-macos-universal.dmg"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-v${version}-darwin-${arch}"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-mac-${arch}.dmg"
        "https://github.com/${REPO}/releases/download/v${version}/${APP_NAME}-${version}-mac-${arch}.dmg"
        "https://github.com/${REPO}/releases/download/v${version}/${BINARY_NAME}-${version}-darwin-${arch}"
    )

    download_from_candidates "${tmp_file}" "${download_urls[@]}" || err "下载失败，请检查版本或网络。"

    if hdiutil imageinfo "${tmp_file}" >/dev/null 2>&1; then
        cp "${tmp_file}" "${MAC_TMP_DMG}"
        rm -f "${tmp_file}"
        hdiutil detach "${MAC_MOUNT_POINT}" >/dev/null 2>&1 || true
        hdiutil attach "${MAC_TMP_DMG}" -mountpoint "${MAC_MOUNT_POINT}" -nobrowse -quiet || err "挂载 DMG 失败。"

        if [ ! -d "${app_path}" ]; then
            hdiutil detach "${MAC_MOUNT_POINT}" >/dev/null 2>&1 || true
            err "未在 DMG 中找到 ${APP_NAME}.app。"
        fi

        rm -rf "/Applications/${APP_NAME}.app"
        cp -R "${app_path}" "/Applications/${APP_NAME}.app"
        hdiutil detach "${MAC_MOUNT_POINT}" -quiet || true
        rm -f "${MAC_TMP_DMG}"

        log "安装完成: /Applications/${APP_NAME}.app"
        echo ""
        echo -e "  ${BOLD}启动方式${NC}: 在 Launchpad 或 /Applications 中打开 ${APP_NAME}"
        echo -e "  ${BOLD}卸载命令${NC}: sudo rm -rf /Applications/${APP_NAME}.app ${INSTALL_DIR}"
        echo ""
        return
    fi

    mkdir -p "${INSTALL_DIR}"
    install -m 755 "${tmp_file}" "${INSTALL_DIR}/${BINARY_NAME}"
    ln -sf "${INSTALL_DIR}/${BINARY_NAME}" /usr/local/bin/${BINARY_NAME}
    rm -f "${tmp_file}"

    log "安装完成: ${INSTALL_DIR}/${BINARY_NAME}"
    echo ""
    echo -e "  ${BOLD}启动命令${NC}: ${CYAN}${BINARY_NAME}${NC}"
    echo -e "  ${BOLD}卸载命令${NC}: sudo rm -f /usr/local/bin/${BINARY_NAME} && sudo rm -rf ${INSTALL_DIR}"
    echo ""
}

main() {
    local os
    local arch
    local version

    if [ "$(id -u)" -ne 0 ]; then
        err "请使用 root 或 sudo 运行此脚本。示例: sudo bash install.sh"
    fi

    os=$(detect_os)
    arch=$(detect_arch)
    version=$(get_latest_version)

    echo ""
    echo -e "${GREEN}=============================================================${NC}"
    echo -e "${GREEN} ${APP_NAME} 一键安装脚本${NC}"
    echo -e "${GREEN} 版本: v${version}  系统: ${os}/${arch}${NC}"
    echo -e "${GREEN}=============================================================${NC}"
    echo ""

    if [ "$os" = "linux" ]; then
        install_linux "$version" "$arch"
    else
        install_macos "$version" "$arch"
    fi

    echo -e "${GREEN}安装完成。${NC}"
}

main "$@"
