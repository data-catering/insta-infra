$ErrorActionPreference = 'Stop'

$packageName = 'insta'
$url = 'https://github.com/data-catering/insta-infra/releases/download/v0.1.0/insta-0.1.0-windows-amd64.zip'
$checksum = 'SKIP' # Update with actual checksum

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$zipFile = Join-Path $toolsDir "$packageName.zip"

# Download the zip file
Invoke-WebRequest -Uri $url -OutFile $zipFile

# Verify checksum
$hash = Get-FileHash -Path $zipFile -Algorithm SHA256
if ($hash.Hash -ne $checksum) {
    throw "Checksum verification failed"
}

# Extract the zip file
Expand-Archive -Path $zipFile -DestinationPath $toolsDir -Force

# Clean up
Remove-Item $zipFile

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$toolsDir*") {
    [Environment]::SetEnvironmentVariable('Path', $userPath + ";$toolsDir", 'User')
} 