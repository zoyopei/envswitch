#!/bin/bash

set -e

# EnvSwitch installer script

# Default settings
GITHUB_REPO="zoyopei/EnvSwitch"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="envswitch"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_color() {
    printf "${1}${2}${NC}\n"
}

# Print info message
info() {
    print_color $BLUE "INFO: $1"
}

# Print success message
success() {
    print_color $GREEN "SUCCESS: $1"
}

# Print warning message
warning() {
    print_color $YELLOW "WARNING: $1"
}

# Print error message
error() {
    print_color $RED "ERROR: $1"
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    info "Fetching latest release version..."
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
        exit 1
    fi
    
    info "Latest version: $VERSION"
}

# Download and install binary
install_binary() {
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}-${VERSION}-${PLATFORM}.tar.gz"
    local temp_dir=$(mktemp -d)
    local temp_file="$temp_dir/${BINARY_NAME}.tar.gz"
    
    info "Downloading $BINARY_NAME $VERSION for $PLATFORM..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -sL "$download_url" -o "$temp_file"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "$temp_file"
    fi
    
    if [ ! -f "$temp_file" ]; then
        error "Failed to download $BINARY_NAME"
        exit 1
    fi
    
    info "Extracting archive..."
    tar -xzf "$temp_file" -C "$temp_dir"
    
    local binary_path="$temp_dir/${BINARY_NAME}-${PLATFORM}"
    if [ ! -f "$binary_path" ]; then
        error "Binary not found in archive"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Install binary
    if [ -w "$INSTALL_DIR" ]; then
        info "Installing $BINARY_NAME to $INSTALL_DIR..."
        mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    else
        info "Installing $BINARY_NAME to $INSTALL_DIR (requires sudo)..."
        sudo mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    success "$BINARY_NAME $VERSION installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v $BINARY_NAME >/dev/null 2>&1; then
        local installed_version=$($BINARY_NAME --version 2>/dev/null | head -n1 || echo "unknown")
        success "Installation verified: $installed_version"
        info "Run '$BINARY_NAME --help' to get started."
    else
        warning "Installation completed, but $BINARY_NAME is not in PATH."
        info "You may need to restart your terminal or add $INSTALL_DIR to your PATH."
    fi
}

# Main installation process
main() {
    echo "EnvSwitch Installer"
    echo "==================="
    echo
    
    # Check if already installed
    if command -v $BINARY_NAME >/dev/null 2>&1; then
        local current_version=$($BINARY_NAME --version 2>/dev/null | head -n1 || echo "unknown")
        warning "$BINARY_NAME is already installed: $current_version"
        read -p "Do you want to reinstall? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Installation cancelled."
            exit 0
        fi
    fi
    
    detect_platform
    get_latest_version
    install_binary
    verify_installation
    
    echo
    success "Installation completed!"
    echo
    echo "Quick start:"
    echo "  $BINARY_NAME --help                     # Show help"
    echo "  $BINARY_NAME project create my-project  # Create a project"
    echo "  $BINARY_NAME server                     # Start web server"
    echo
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "EnvSwitch Installer"
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version      Install specific version"
        echo "  --dir DIR      Install to specific directory (default: $INSTALL_DIR)"
        echo
        echo "Environment variables:"
        echo "  INSTALL_DIR    Installation directory (default: $INSTALL_DIR)"
        echo
        exit 0
        ;;
    --version)
        if [ -z "${2:-}" ]; then
            error "--version requires a version argument"
            exit 1
        fi
        VERSION="$2"
        ;;
    --dir)
        if [ -z "${2:-}" ]; then
            error "--dir requires a directory argument"
            exit 1
        fi
        INSTALL_DIR="$2"
        ;;
esac

# Override with environment variable
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Run main installation
main 