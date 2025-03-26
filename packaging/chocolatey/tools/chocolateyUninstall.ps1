$ErrorActionPreference = 'Stop'

$packageName = 'insta'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Remove from PATH
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$newPath = ($userPath.Split(';') | Where-Object { $_ -ne $toolsDir }) -join ';'
[Environment]::SetEnvironmentVariable('Path', $newPath, 'User')

# Remove the installation directory
Remove-Item -Path $toolsDir -Recurse -Force 