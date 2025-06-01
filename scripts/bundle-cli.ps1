# Bundle CLI binary into Wails app for Windows
# This script runs after Wails build to embed the CLI binary

$ErrorActionPreference = "Stop"

Write-Host "Bundling CLI binary into Wails app..."

# Get the repository root directory (where this script's parent directory is)
$REPO_ROOT = Split-Path -Parent $PSScriptRoot
Write-Host "Repository root: $REPO_ROOT"

# Set paths relative to repository root
$CLI_BINARY_BASE = Join-Path $REPO_ROOT "insta"
$CLI_BINARY_EXE = Join-Path $REPO_ROOT "insta.exe"
$WAILS_APP_DIR = Join-Path $REPO_ROOT "cmd\instaui\build\bin"

Write-Host "Looking for CLI binary..."
Write-Host "Checking: $CLI_BINARY_BASE"
Write-Host "Checking: $CLI_BINARY_EXE"

# Check which binary exists
$CLI_BINARY = $null
if (Test-Path $CLI_BINARY_EXE) {
    $CLI_BINARY = $CLI_BINARY_EXE
    Write-Host "Found CLI binary: $CLI_BINARY_EXE"
} elseif (Test-Path $CLI_BINARY_BASE) {
    $CLI_BINARY = $CLI_BINARY_BASE
    Write-Host "Found CLI binary: $CLI_BINARY_BASE"
} else {
    Write-Host "Error: CLI binary not found at either:"
    Write-Host "  $CLI_BINARY_BASE"
    Write-Host "  $CLI_BINARY_EXE"
    Write-Host ""
    Write-Host "Current directory contents:"
    Get-ChildItem $REPO_ROOT | Format-Table Name, Length
    Write-Host ""
    Write-Host "Make sure to run 'make build' first"
    exit 1
}

# Ensure the destination directory exists
if (-Not (Test-Path $WAILS_APP_DIR)) {
    Write-Host "Error: Wails app directory not found at $WAILS_APP_DIR"
    Write-Host "Make sure to run 'wails build' first"
    exit 1
}

# Copy the CLI binary to the Wails app directory
$DEST_PATH = Join-Path $WAILS_APP_DIR "insta-cli.exe"
Copy-Item $CLI_BINARY $DEST_PATH

Write-Host "CLI binary bundled for Windows at $DEST_PATH"
Write-Host "CLI bundling complete!" 