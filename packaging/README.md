# Packaging

This directory contains package definitions for various package managers.

## Directory Structure

- `debian/` - Debian/Ubuntu package files
  - `control` - Package metadata and dependencies
  - `rules` - Build rules
  - `changelog` - Version history
  - `compat` - Debian compatibility level

- `rpm/` - RHEL/CentOS/Fedora package files
  - `insta.spec` - RPM spec file

- `arch/` - Arch Linux package files
  - `PKGBUILD` - Build script

- `chocolatey/` - Windows Chocolatey package files
  - `insta.nuspec` - Package metadata
  - `tools/` - Installation and uninstallation scripts

## Automated Build Script

A build script (`build.sh`) is provided to automate the package building process. This script will:

1. Build the Go binary
2. Detect the current platform
3. Build the appropriate package(s) for that platform
4. Place all built packages in the `build/packages` directory

### Usage

```bash
# Make the script executable
chmod +x packaging/build.sh

# Run the build script
./packaging/build.sh
```

### Requirements

The script will check for the following requirements based on the platform:

- Debian/Ubuntu:
  - `dpkg-buildpackage`
  - `golang-go`

- RHEL/CentOS/Fedora:
  - `rpmbuild`
  - `golang`

- Arch Linux:
  - `makepkg`
  - `go`

- Windows (Chocolatey):
  - `choco`
  - `go`

### Manual Building

If you prefer to build packages manually, here are the commands for each platform:

#### Debian/Ubuntu
```bash
cd packaging/debian
dpkg-buildpackage -us -uc
```

#### RHEL/CentOS/Fedora
```bash
cd packaging/rpm
rpmbuild -ba insta.spec
```

#### Arch Linux
```bash
cd packaging/arch
makepkg -si
```

#### Windows (Chocolatey)
```bash
cd packaging/chocolatey
choco pack
``` 