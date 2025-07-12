# build-all.ps1
$version = "v0.1.0"
$appName = "jira-branch"

Write-Host "Building $appName $version for all platforms..."
Write-Host "=================================="

New-Item -ItemType Directory -Force -Path "builds" | Out-Null

# Windows builds
Write-Host "Building for Windows x64..."
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -o "builds/$appName-$version-windows-x64.exe"

Write-Host "Building for Windows ARM64..."
$env:GOOS="windows"; $env:GOARCH="arm64"
go build -o "builds/$appName-$version-windows-arm64.exe"

# Linux builds
Write-Host "Building for Linux x64..."
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o "builds/$appName-$version-linux-x64"

Write-Host "Building for Linux ARM64..."
$env:GOOS="linux"; $env:GOARCH="arm64"
go build -o "builds/$appName-$version-linux-arm64"

# macOS builds
Write-Host "Building for macOS x64 (Intel)..."
$env:GOOS="darwin"; $env:GOARCH="amd64"
go build -o "builds/$appName-$version-macos-x64"

Write-Host "Building for macOS ARM64 (Apple Silicon)..."
$env:GOOS="darwin"; $env:GOARCH="arm64"
go build -o "builds/$appName-$version-macos-arm64"

Write-Host ""
Write-Host "All builds completed! Files created:"
Write-Host "====================================="
Get-ChildItem -Path "builds" | ForEach-Object { Write-Host "  $($_.Name)" }