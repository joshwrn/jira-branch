# Jira Branch - PowerShell Auto-installer script
# Detects OS/architecture and downloads the appropriate binary

param(
    [string]$InstallPath = "$env:LOCALAPPDATA\jira-branch"
)

$ErrorActionPreference = "Stop"

$REPO = "joshwrn/jira-branch"

# Colors for output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

Write-ColorOutput "Jira Branch Installer" "Blue"
Write-ColorOutput "======================" "Blue"

# Function to detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x64" }
        "ARM64" { return "arm64" }
        default { return "unknown" }
    }
}

# Function to get download URL
function Get-DownloadUrl {
    param(
        [string]$Arch
    )
    
    Write-ColorOutput "Fetching latest release info..." "Yellow"
    
    try {
        $releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
    }
    catch {
        Write-ColorOutput "Error: Failed to fetch release information" "Red"
        Write-ColorOutput $_.Exception.Message "Red"
        exit 1
    }
    
    # Find the Windows asset for the detected architecture
    $expectedName = "jira-branch-*-windows-$Arch.exe"
    $asset = $releaseInfo.assets | Where-Object { $_.name -like $expectedName }
    
    if (-not $asset) {
        Write-ColorOutput "Error: No release found for windows-$Arch" "Red"
        Write-ColorOutput "Available releases:" "Yellow"
        $releaseInfo.assets | ForEach-Object { Write-Host "  $($_.name)" }
        exit 1
    }
    
    return $asset.browser_download_url
}

# Function to install binary
function Install-Binary {
    param(
        [string]$DownloadUrl,
        [string]$InstallPath
    )
    
    # Get filename from URL
    $filename = Split-Path $DownloadUrl -Leaf
    $binaryName = "jira-branch.exe"
    
    Write-ColorOutput "Downloading $filename..." "Yellow"
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    # Download binary
    $tempFile = Join-Path $env:TEMP $filename
    $finalPath = Join-Path $InstallPath $binaryName
    
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $tempFile
        Move-Item $tempFile $finalPath -Force
    }
    catch {
        Write-ColorOutput "Error: Failed to download or install binary" "Red"
        Write-ColorOutput $_.Exception.Message "Red"
        exit 1
    }
    
    Write-ColorOutput "✓ Binary installed to $finalPath" "Green"
    
    # Check if install path is in PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$InstallPath*") {
        Write-ColorOutput "Adding $InstallPath to user PATH..." "Yellow"
        
        try {
            $newPath = "$currentPath;$InstallPath"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-ColorOutput "✓ Added to PATH. Restart your terminal or run: `$env:PATH += `";$InstallPath`"" "Green"
        }
        catch {
            Write-ColorOutput "Warning: Could not add to PATH automatically" "Yellow"
            Write-ColorOutput "Please add $InstallPath to your PATH manually" "Yellow"
        }
    }
}

# Main installation process
function Main {
    Write-ColorOutput "Detecting platform..." "Yellow"
    
    $arch = Get-Architecture
    $os = "windows"
    
    Write-ColorOutput "Detected: $os-$arch" "Blue"
    
    if ($arch -eq "unknown") {
        Write-ColorOutput "Error: Unsupported architecture: $arch" "Red"
        Write-ColorOutput "Supported architectures: x64, arm64" "Yellow"
        exit 1
    }
    
    # Get download URL
    $downloadUrl = Get-DownloadUrl -Arch $arch
    Write-ColorOutput "Download URL: $downloadUrl" "Blue"
    
    # Install binary
    Install-Binary -DownloadUrl $downloadUrl -InstallPath $InstallPath
    
    Write-ColorOutput "✓ Installation complete!" "Green"
    Write-ColorOutput "Run 'jira-branch' to get started" "Blue"
}

# Run main function
try {
    Main
}
catch {
    Write-ColorOutput "Installation failed: $($_.Exception.Message)" "Red"
    exit 1
}