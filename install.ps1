# Jira Branch - PowerShell Auto-installer script
# Detects OS/architecture and downloads the appropriate binary

param(
    [string]$InstallPath = "$env:USERPROFILE\.jira-branch\bin"
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

Write-ColorOutput "Installing Jira Branch..." "Blue"

# Function to detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x64" }
        "ARM64" { return "arm64" }
        default { return "unknown" }
    }
}


# Function to get download URL and release info
function Get-ReleaseInfo {
    param(
        [string]$Arch
    )
    
    Write-ColorOutput "Fetching latest release info..." "White"
    
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
    
    return @{
        Version = $releaseInfo.tag_name
        DownloadUrl = $asset.browser_download_url
        Name = $asset.name
    }
}

# Function to add PowerShell alias
function Add-PowerShellAlias {
    param(
        [string]$InstallPath
    )
    
    Write-ColorOutput "" "White"
    $addAlias = Read-Host "Would you like to add an alias 'jb' for 'jira-branch' to your PowerShell profile? (y/N)"
    
    if ($addAlias -match "^[Yy]") {
        try {
            # Check if profile exists, create if not
            if (-not (Test-Path $PROFILE)) {
                New-Item -ItemType File -Path $PROFILE -Force | Out-Null
                Write-ColorOutput "Created PowerShell profile at $PROFILE" "Yellow"
            }
            
            # Check if alias already exists
            $profileContent = Get-Content $PROFILE -ErrorAction SilentlyContinue
            $aliasExists = $profileContent | Where-Object { $_ -like "*Set-Alias jb jira-branch*" -or $_ -like "*Set-Alias -Name jb*" }
            
            if ($aliasExists) {
                Write-ColorOutput "Alias 'jb' already exists in your PowerShell profile" "Yellow"
                return
            }
            
            # Add alias to profile
            $aliasLine = "Set-Alias jb jira-branch"
            Add-Content $PROFILE "`n# Jira Branch alias`n$aliasLine"
            
            Write-ColorOutput "✅ Added alias 'jb' to your PowerShell profile" "Green"
            Write-ColorOutput "Restart your PowerShell session or run: . `$PROFILE" "Blue"
        }
        catch {
            Write-ColorOutput "Warning: Could not add alias to PowerShell profile" "Yellow"
            Write-ColorOutput "You can manually add 'Set-Alias jb jira-branch' to your profile" "Yellow"
        }
    }
}

# Function to install binary
function Install-Binary {
    param(
        [string]$DownloadUrl,
        [string]$InstallPath,
        [string]$Filename
    )
    
    $binaryName = "jira-branch.exe"
    
    Write-ColorOutput "Downloading $Filename..." "White"
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    # Download binary
    $tempFile = Join-Path $env:TEMP $Filename
    $finalPath = Join-Path $InstallPath $binaryName
    
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $tempFile
        
        # If binary exists and is in use, try to stop any running processes
        if (Test-Path $finalPath) {
            try {
                Get-Process | Where-Object { $_.Path -eq $finalPath } | Stop-Process -Force -ErrorAction SilentlyContinue
                Start-Sleep -Milliseconds 500
            }
            catch {
                # Ignore errors from stopping processes
            }
        }
        
        Move-Item $tempFile $finalPath -Force
    }
    catch {
        Write-ColorOutput "Error: Failed to download or install binary" "Red"
        Write-ColorOutput $_.Exception.Message "Red"
        
        if ($_.Exception.Message -like "*being used by another process*") {
            Write-ColorOutput "The binary appears to be running. Please close jira-branch and try again." "Yellow"
        }
        
        exit 1
    }
    
    Write-ColorOutput "Binary installed to $finalPath" "Green"
    
    # Check if install path is in PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$InstallPath*") {
        Write-ColorOutput "Adding $InstallPath to user PATH..." "Yellow"
        
        try {
            $newPath = "$currentPath;$InstallPath"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-ColorOutput "Added to PATH. Restart your terminal or run: `$env:PATH += `";$InstallPath`"" "Green"
        }
        catch {
            Write-ColorOutput "Warning: Could not add to PATH automatically" "Yellow"
            Write-ColorOutput "Please add $InstallPath to your PATH manually" "Yellow"
        }
    }
    
    # Offer to add alias to PowerShell profile
    Add-PowerShellAlias -InstallPath $InstallPath
}

# Main installation process
function Main {
    Write-ColorOutput "Detecting platform..." "White"
    
    $arch = Get-Architecture
    $os = "windows"
    
    Write-ColorOutput "Detected: $os-$arch" "Blue"
    
    if ($arch -eq "unknown") {
        Write-ColorOutput "Error: Unsupported architecture: $arch" "Red"
        Write-ColorOutput "Supported architectures: x64, arm64" "Yellow"
        exit 1
    }
    
    # Get release info
    $release = Get-ReleaseInfo -Arch $arch
    Write-ColorOutput "Installing jira-branch $($release.Version)" "Green"
    
    # Install binary
    Install-Binary -DownloadUrl $release.DownloadUrl -InstallPath $InstallPath -Filename $release.Name
    
    Write-ColorOutput "Installation complete!" "Green"
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