$ErrorActionPreference = 'Stop'

$packageName = 'insta'

# Remove the shim
Uninstall-BinFile -Name $packageName

# Remove the installation directory
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
Remove-Item -Path $toolsDir -Recurse -Force 