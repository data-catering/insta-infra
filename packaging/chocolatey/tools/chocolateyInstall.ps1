$ErrorActionPreference = 'Stop'

$packageName = 'insta'
$url = 'https://github.com/data-catering/insta-infra/releases/download/v{{.Version}}/insta-{{.Version}}-windows-amd64.zip'
$checksum = '{{.Checksum}}'
$checksumType = 'sha256'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$zipFile = Join-Path $toolsDir "$packageName.zip"

# Download the zip file
Invoke-WebRequest -Uri $url -OutFile $zipFile

# Verify the checksum
$hash = Get-FileHash -Path $zipFile -Algorithm $checksumType
if ($hash.Hash -ne $checksum) {
    throw "Checksum verification failed. Expected $checksum but got $($hash.Hash)"
}

# Extract the zip file
Expand-Archive -Path $zipFile -DestinationPath $toolsDir -Force

# Remove the zip file
Remove-Item $zipFile

# Create shims
Install-BinFile -Name $packageName -Path (Join-Path $toolsDir "$packageName.exe")

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$toolsDir*") {
    [Environment]::SetEnvironmentVariable('Path', $userPath + ";$toolsDir", 'User')
} 