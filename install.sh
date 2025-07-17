#!/bin/bash

# Jira Branch - Auto-installer script
# Detects OS/architecture and downloads the appropriate binary

set -e

REPO="joshwrn/jira-branch"
INSTALL_DIR="/usr/local/bin"

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

# Function to get download URL for platform
get_download_url() {
    local os=$1
    local arch=$2
    
    # Get latest release info
    echo -e "${YELLOW}Fetching latest release info...${NC}"
    local release_info=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to fetch release information${NC}"
        exit 1
    fi
    
    # Construct expected filename based on your build pattern
    local filename
    if [ "$os" = "windows" ]; then
        filename="jira-branch-.*-${os}-${arch}\.exe"
    else
        filename="jira-branch-.*-${os}-${arch}"
    fi
    
    # Extract download URL
    local download_url=$(echo "$release_info" | grep -E "browser_download_url.*$filename" | head -1 | cut -d '"' -f 4)
    
    if [ -z "$download_url" ]; then
        echo -e "${RED}Error: No release found for $os-$arch${NC}"
        echo "Available releases:"
        echo "$release_info" | grep -E '"name".*jira-branch' | cut -d '"' -f 4
        exit 1
    fi
    
    echo "$download_url"
}

# Function to install binary
install_binary() {
    local download_url=$1
    local os=$2
    
    # Get filename from URL
    local filename=$(basename "$download_url")
    local binary_name="jira-branch"
    
    echo -e "${YELLOW}Downloading $filename...${NC}"
    
    # Download to temp location
    local temp_file="/tmp/$filename"
    curl -L -o "$temp_file" "$download_url"
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to download binary${NC}"
        exit 1
    fi
    
    # Make executable
    chmod +x "$temp_file"
    
    # Install binary
    if [ "$os" = "windows" ]; then
        # For Windows, just move to current directory or suggest manual installation
        echo -e "${YELLOW}Moving binary to current directory...${NC}"
        mv "$temp_file" "./$binary_name.exe"
        echo -e "${GREEN}✓ Binary installed as ./$binary_name.exe${NC}"
        echo -e "${BLUE}Add the current directory to your PATH or move the binary to a directory in your PATH${NC}"
    else
        # For Unix-like systems, install to /usr/local/bin
        echo -e "${YELLOW}Installing to $INSTALL_DIR (may require sudo)...${NC}"
        
        # Check if we can write to install directory
        if [ -w "$INSTALL_DIR" ]; then
            mv "$temp_file" "$INSTALL_DIR/$binary_name"
        else
            sudo mv "$temp_file" "$INSTALL_DIR/$binary_name"
        fi
        
        echo -e "${GREEN}✓ Binary installed to $INSTALL_DIR/$binary_name${NC}"
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
    
    # Get download URL
    DOWNLOAD_URL=$(get_download_url "$OS" "$ARCH")
    echo -e "Download URL: ${BLUE}$DOWNLOAD_URL${NC}"
    
    # Install binary
    install_binary "$DOWNLOAD_URL" "$OS"
    
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