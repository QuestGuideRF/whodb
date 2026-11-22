#!/bin/bash
set -e
REPO="clidey/whodb"
BINARY_NAME="whodb-cli"
INSTALL_DIR="${WHODB_INSTALL_DIR:-$HOME/.local/bin}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
print_step() {
    echo -e "${BLUE}==>${NC} $1"
}
print_success() {
    echo -e "${GREEN}==>${NC} $1"
}
print_warning() {
    echo -e "${YELLOW}Warning:${NC} $1"
}
print_error() {
    echo -e "${RED}Error:${NC} $1" >&2
}
detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)
            print_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
}
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7*|armhf) echo "armv7" ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}
get_latest_version() {
    local latest
    latest=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$latest" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi
    echo "$latest"
}
command_exists() {
    command -v "$1" >/dev/null 2>&1
}
main() {
    local version="${1:-}"
    local os
    local arch
    local binary_suffix
    local download_url
    local tmp_dir
    local install_path
    print_step "WhoDB CLI Installer"
    echo ""
    os=$(detect_os)
    arch=$(detect_arch)
    print_step "Detected system: ${os}/${arch}"
    if [ -z "$version" ] || [ "$version" = "latest" ]; then
        print_step "Fetching latest version..."
        version=$(get_latest_version)
    fi
    print_step "Installing version: ${version}"
    binary_suffix="${BINARY_NAME}-${os}-${arch}"
    download_url="https://github.com/${REPO}/releases/download/${version}/${binary_suffix}"
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT
    print_step "Downloading ${binary_suffix}..."
    if command_exists curl; then
        curl -fsSL -o "${tmp_dir}/${BINARY_NAME}" "$download_url"
    elif command_exists wget; then
        wget -q -O "${tmp_dir}/${BINARY_NAME}" "$download_url"
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    if [ ! -f "${tmp_dir}/${BINARY_NAME}" ] || [ ! -s "${tmp_dir}/${BINARY_NAME}" ]; then
        print_error "Download failed or file is empty"
        print_error "URL: ${download_url}"
        exit 1
    fi
    chmod +x "${tmp_dir}/${BINARY_NAME}"
    mkdir -p "$INSTALL_DIR"
    install_path="${INSTALL_DIR}/${BINARY_NAME}"
    print_step "Installing to ${install_path}..."
    mv "${tmp_dir}/${BINARY_NAME}" "$install_path"
    if [ ! -x "$install_path" ]; then
        print_error "Installation failed"
        exit 1
    fi
    print_success "WhoDB CLI ${version} installed successfully!"
    echo ""
    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
        print_warning "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "Add it to your shell profile:"
        echo ""
        local shell_config=""
        case "$SHELL" in
            */zsh)  shell_config="~/.zshrc" ;;
            */bash) shell_config="~/.bashrc" ;;
            */fish) shell_config="~/.config/fish/config.fish" ;;
            *)      shell_config="your shell config" ;;
        esac
        case "$SHELL" in
            */fish)
                echo "  fish_add_path ${INSTALL_DIR}"
                ;;
            *)
                echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
                ;;
        esac
        echo ""
        echo "Then restart your shell or run:"
        echo "  source ${shell_config}"
        echo ""
    fi
    echo "Get started:"
    echo "  ${BINARY_NAME}          # Launch interactive TUI"
    echo "  ${BINARY_NAME} mcp      # Run as MCP server"
    echo "  ${BINARY_NAME} --help   # Show help"
    echo ""
    echo "Documentation: https://github.com/clidey/whodb/tree/main/cli#readme"
}
main "$@"