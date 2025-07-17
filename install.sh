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

# Function to detect shell config file
detect_shell_config() {
    # Check current shell
    local current_shell=$(basename "$SHELL" 2>/dev/null || echo "bash")
    
    case "$current_shell" in
        zsh)
            if [ -f "$HOME/.zshrc" ]; then
                echo "$HOME/.zshrc"
            else
                echo "$HOME/.zshrc"  # Will be created if needed
            fi
            ;;
        bash)
            if [ -f "$HOME/.bashrc" ]; then
                echo "$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                echo "$HOME/.bash_profile"
            else
                echo "$HOME/.bashrc"  # Will be created if needed
            fi
            ;;
        *)
            # Default to .bashrc for unknown shells
            echo "$HOME/.bashrc"
            ;;
    esac
}

# Function to configure shell alias
configure_alias() {
    local config_file=$(detect_shell_config)
    local shell_name=$(basename "$SHELL" 2>/dev/null || echo "bash")
    local alias_line="alias jb='jira-branch'"
    
    echo -e "${YELLOW}Setting up 'jb' alias...${NC}"
    
    # Check if alias already exists
    if [ -f "$config_file" ] && grep -q "alias jb=" "$config_file"; then
        echo -e "${YELLOW}Alias 'jb' already exists in $config_file${NC}"
        echo -e "${BLUE}You can run 'jb' instead of 'jira-branch'${NC}"
        return 0
    fi
    
    # Create config file if it doesn't exist
    if [ ! -f "$config_file" ]; then
        touch "$config_file"
        echo -e "${YELLOW}Created $config_file${NC}"
    fi
    
    # Add alias to config file
    echo "" >> "$config_file"
    echo "# Jira Branch alias" >> "$config_file"
    echo "$alias_line" >> "$config_file"
    
    echo -e "${GREEN}✓ Added alias 'jb' to $config_file${NC}"
    echo -e "${BLUE}Restart your terminal or run 'source $config_file' to use the alias${NC}"
    echo -e "${BLUE}You can now run 'jb' instead of 'jira-branch'${NC}"
    
    return 0
}

# Function to prompt for alias setup
prompt_alias_setup() {
    echo ""
    echo -e "${BLUE}Would you like to set up a shell alias 'jb' for jira-branch?${NC}"
    echo -e "${BLUE}This will add 'alias jb=jira-branch' to your shell configuration.${NC}"
    
    while true; do
        read -p "Set up alias? (y/n): " yn
        case $yn in
            [Yy]* ) 
                configure_alias
                break
                ;;
            [Nn]* ) 
                echo -e "${BLUE}Skipping alias setup. You can run 'jira-branch' directly.${NC}"
                break
                ;;
            * ) 
                echo "Please answer yes or no."
                ;;
        esac
    done
}


# Function to get release info
get_release_info() {
    local os=$1
    local arch=$2
    
    # Get latest release info (redirect to stderr so it doesn't interfere with return value)
    echo -e "${YELLOW}Fetching latest release info...${NC}" >&2
    local release_info=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to fetch release information${NC}" >&2
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
    local download_url=$(echo "$release_info" | grep -E "browser_download_url.*$filename" | head -1 | cut -d '"' -f 4 | xargs)
    local asset_name=$(echo "$release_info" | grep -E "\"name\".*$filename" | head -1 | cut -d '"' -f 4 | xargs)
    
    if [ -z "$download_url" ]; then
        echo -e "${RED}Error: No release found for $os-$arch${NC}" >&2
        echo "Available releases:" >&2
        echo "$release_info" | grep -E '"name".*jira-branch' | cut -d '"' -f 4 >&2
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
    
    # Get release info (returns: version download_url asset_name)
    RELEASE_INFO=$(get_release_info "$OS" "$ARCH")
    LATEST_VERSION=$(echo "$RELEASE_INFO" | cut -d' ' -f1)
    DOWNLOAD_URL=$(echo "$RELEASE_INFO" | cut -d' ' -f2)
    ASSET_NAME=$(echo "$RELEASE_INFO" | cut -d' ' -f3)
    
    echo -e "Installing jira-branch ${BLUE}$LATEST_VERSION${NC}"
    
    # Install binary
    install_binary "$DOWNLOAD_URL" "$OS" "$ASSET_NAME"
    
    # Prompt for alias setup
    prompt_alias_setup
    
    echo ""
    echo -e "${GREEN}✓ Installation complete!${NC}"
    echo -e "${BLUE}Run 'jira-branch' (or 'jb' if you set up the alias) to get started${NC}"
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