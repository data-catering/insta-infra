# Bundle CLI binary into Wails app for Windows
# This script runs after Wails build to embed the CLI binary

$ErrorActionPreference = "Stop"

Write-Host "Bundling CLI binary into Wails app..."

# Set paths
$CLI_BINARY = "insta.exe"
$WAILS_APP_DIR = "cmd\instaui\build\bin"

# For Windows
if (-Not (Test-Path $CLI_BINARY)) {
    Write-Host "Error: CLI binary not found at $CLI_BINARY"
    Write-Host "Make sure to run 'make build' first"
    exit 1
}

$WINDOWS_BINARY_DIR = $WAILS_APP_DIR
Copy-Item $CLI_BINARY "$WINDOWS_BINARY_DIR\insta-cli.exe"

Write-Host "CLI binary bundled for Windows at $WINDOWS_BINARY_DIR\insta-cli.exe"
Write-Host "CLI bundling complete!" 