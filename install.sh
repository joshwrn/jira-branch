#!/bin/bash

# Jira Branch - Auto-installer script
# Detects OS/architecture and downloads the appropriate binary

set -e

REPO="joshwrn/jira-branch"
INSTALL_DIR="$HOME/.jira-branch/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Jira Branch Installer${NC}"
echo "======================"

# Function to detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "macos";;
        CYGWIN*|MINGW*|MSYS*) echo "windows";;
        *)          echo "unknown";;
    esac
}

# Function to detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "x64";;
        arm64|aarch64)  echo "arm64";;
        *)              echo "unknown";;
    esac
}

# Function to get current version
get_current_version() {
    local binary_path
    
    # Check install directory first, then PATH, then current directory
    if [ -f "$INSTALL_DIR/jira-branch" ]; then
        binary_path="$INSTALL_DIR/jira-branch"
    elif command -v jira-branch >/dev/null 2>&1; then
        binary_path=$(which jira-branch)
    elif [ -f "./jira-branch" ]; then
        binary_path="./jira-branch"
    else
        echo ""
        return
    fi
    
    # Try to get version from the binary
    local version_output
    if version_output=$($binary_path --version 2>/dev/null); then
        echo "$version_output"
    else
        echo "unknown"
    fi
}

# Function to get release info
get_release_info() {
    local os=$1
    local arch=$2
    
    # Get latest release info
    echo -e "${YELLOW}Fetching latest release info...${NC}"
    local release_info=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to fetch release information${NC}"
        exit 1
    fi
    
    # Extract version
    local version=$(echo "$release_info" | grep '"tag_name"' | head -1 | cut -d '"' -f 4)
    
    # Construct expected filename based on your build pattern
    local filename
    if [ "$os" = "windows" ]; then
        filename="jira-branch-.*-${os}-${arch}\.exe"
    else
        filename="jira-branch-.*-${os}-${arch}"
    fi
    
    # Extract download URL
    local download_url=$(echo "$release_info" | grep -E "browser_download_url.*$filename" | head -1 | cut -d '"' -f 4)
    local asset_name=$(echo "$release_info" | grep -E "\"name\".*$filename" | head -1 | cut -d '"' -f 4)
    
    if [ -z "$download_url" ]; then
        echo -e "${RED}Error: No release found for $os-$arch${NC}"
        echo "Available releases:"
        echo "$release_info" | grep -E '"name".*jira-branch' | cut -d '"' -f 4
        exit 1
    fi
    
    # Return as space-separated values: version download_url asset_name
    echo "$version $download_url $asset_name"
}

# Function to install binary
install_binary() {
    local download_url=$1
    local os=$2
    local asset_name=$3
    
    local binary_name="jira-branch"
    
    echo -e "${YELLOW}Downloading $asset_name...${NC}"
    
    # Download to temp location
    local temp_file="/tmp/$asset_name"
    curl -L -o "$temp_file" "$download_url"
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to download binary${NC}"
        exit 1
    fi
    
    # Make executable
    chmod +x "$temp_file"
    
    # Install binary
    if [ "$os" = "windows" ]; then
        # For Windows in bash (like Git Bash), install to user directory
        echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"
        
        # Create install directory if it doesn't exist
        mkdir -p "$INSTALL_DIR"
        
        mv "$temp_file" "$INSTALL_DIR/$binary_name.exe"
        echo -e "${GREEN}✓ Binary installed to $INSTALL_DIR/$binary_name.exe${NC}"
        echo -e "${BLUE}Add $INSTALL_DIR to your PATH to use 'jira-branch' from anywhere${NC}"
    else
        # For Unix-like systems, install to user directory
        echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"
        
        # Create install directory if it doesn't exist
        mkdir -p "$INSTALL_DIR"
        
        mv "$temp_file" "$INSTALL_DIR/$binary_name"
        echo -e "${GREEN}✓ Binary installed to $INSTALL_DIR/$binary_name${NC}"
        
        # Check if install dir is in PATH
        if echo "$PATH" | grep -q "$INSTALL_DIR"; then
            echo -e "${GREEN}✓ $INSTALL_DIR is already in your PATH${NC}"
        else
            echo -e "${BLUE}Add $INSTALL_DIR to your PATH to use 'jira-branch' from anywhere:${NC}"
            echo -e "${BLUE}  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc && source ~/.bashrc${NC}"
        fi
    fi
}

# Main installation process
main() {
    echo -e "${YELLOW}Detecting platform...${NC}"
    
    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    echo -e "Detected: ${BLUE}$OS-$ARCH${NC}"
    
    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo -e "${RED}Error: Unsupported platform: $OS-$ARCH${NC}"
        echo "Supported platforms: linux-x64, linux-arm64, macos-x64, macos-arm64, windows-x64, windows-arm64"
        exit 1
    fi
    
    # Check current version
    CURRENT_VERSION=$(get_current_version)
    if [ -n "$CURRENT_VERSION" ]; then
        echo -e "Current version: ${BLUE}$CURRENT_VERSION${NC}"
    fi
    
    # Get release info (returns: version download_url asset_name)
    RELEASE_INFO=$(get_release_info "$OS" "$ARCH")
    LATEST_VERSION=$(echo "$RELEASE_INFO" | cut -d' ' -f1)
    DOWNLOAD_URL=$(echo "$RELEASE_INFO" | cut -d' ' -f2)
    ASSET_NAME=$(echo "$RELEASE_INFO" | cut -d' ' -f3)
    
    echo -e "Latest version: ${BLUE}$LATEST_VERSION${NC}"
    
    # Check if update is needed
    if [ -n "$CURRENT_VERSION" ] && [ "$CURRENT_VERSION" != "unknown" ]; then
        if echo "$CURRENT_VERSION" | grep -q "$LATEST_VERSION" || echo "$LATEST_VERSION" | grep -q "$CURRENT_VERSION"; then
            echo -e "${GREEN}You already have the latest version installed.${NC}"
            echo -n "Do you want to reinstall anyway? (y/N): "
            read -r CONTINUE
            if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
                echo -e "${YELLOW}Installation cancelled.${NC}"
                return
            fi
        else
            echo -e "${GREEN}Updating from $CURRENT_VERSION to $LATEST_VERSION${NC}"
        fi
    elif [ "$CURRENT_VERSION" = "unknown" ]; then
        echo -e "${YELLOW}Found existing installation (version unknown). Updating...${NC}"
    else
        echo -e "${GREEN}Installing jira-branch $LATEST_VERSION${NC}"
    fi
    
    # Install binary
    install_binary "$DOWNLOAD_URL" "$OS" "$ASSET_NAME"
    
    echo -e "${GREEN}✓ Installation complete!${NC}"
    echo -e "${BLUE}Run 'jira-branch' to get started${NC}"
}

# Check for required commands
for cmd in curl uname grep cut; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo -e "${RED}Error: Required command '$cmd' not found${NC}"
        exit 1
    fi
done

# Run main function
main "$@"