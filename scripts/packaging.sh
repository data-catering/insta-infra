#!/bin/bash

# Exit on error
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Make sure we're in the root directory
cd "$(dirname "$0")/.."

# Version from git tag or default
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")}
BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}
BINARY_NAME=${BINARY_NAME:-insta}
BUILD_PACKAGES=${BUILD_PACKAGES:-false}

# Build directory
BUILD_DIR="build/packages"
RELEASE_DIR="release"
mkdir -p "$BUILD_DIR" "$RELEASE_DIR"

# Function to print colored messages
print_status() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if a command exists
check_command() {
    if ! command -v "$1" &> /dev/null; then
        print_error "$1 is not installed. Please install it first."
        exit 1
    fi
}

# Function to calculate SHA256
calculate_sha256() {
    local file="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        shasum -a 256 "$file" | cut -d' ' -f1
    else
        sha256sum "$file" | cut -d' ' -f1
    fi
}

# Function to update RPM spec file
update_rpm_spec() {
    local sha256="$1"
    local spec_file="packaging/rpm/insta.spec"
    if [ -f "$spec_file" ]; then
        print_status "Updating RPM spec file..."
        sed -i.bak "s/^Source0:.*/Source0:        insta-%{version}.tar.gz/" "$spec_file"
        sed -i.bak "s/^%global sha256.*/%global sha256 $sha256/" "$spec_file"
        rm -f "${spec_file}.bak"
    fi
}

# Function to update Arch PKGBUILD
update_pkgbuild() {
    local sha256="$1"
    local pkgbuild_file="packaging/arch/PKGBUILD"
    if [ -f "$pkgbuild_file" ]; then
        print_status "Updating Arch PKGBUILD..."
        sed -i.bak "s/^sha256sums=.*/sha256sums=('$sha256')/" "$pkgbuild_file"
        rm -f "${pkgbuild_file}.bak"
    fi
}

# Function to update Chocolatey nuspec
update_chocolatey_nuspec() {
    local sha256="$1"
    local nuspec_file="packaging/chocolatey/insta.nuspec"
    if [ -f "$nuspec_file" ]; then
        print_status "Updating Chocolatey nuspec..."
        sed -i.bak "s/<checksum type=\"sha256\">.*<\/checksum>/<checksum type=\"sha256\">$sha256<\/checksum>/" "$nuspec_file"
        rm -f "${nuspec_file}.bak"
    fi
}

# Function to create release archives
create_release_archives() {
    print_status "Creating release archives..."
    
    # Copy README and LICENSE to release directory
    cp README.md LICENSE "${RELEASE_DIR}/"
    
    # Create binary archives for each platform
    for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
        GOOS=${platform%/*}
        GOARCH=${platform##*/}
        BINARY_EXT="tar.gz"
        
        if [ "$GOOS" = "windows" ]; then
            BINARY_EXT="zip"
        fi
        
        # Create binary archive
        if [ "$GOOS" = "windows" ]; then
            (cd "${RELEASE_DIR}" && zip -q "${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.zip" "${BINARY_NAME}-${GOOS}-${GOARCH}" README.md LICENSE)
        else
            (cd "${RELEASE_DIR}" && tar -czf "${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.tar.gz" "${BINARY_NAME}-${GOOS}-${GOARCH}" README.md LICENSE)
        fi
    done
    
    # Create source tarball
    print_status "Creating source tarball..."
    git archive --format=tar.gz --prefix="insta-$VERSION/" HEAD > "${RELEASE_DIR}/insta-${VERSION}.tar.gz"
    
    # Calculate SHA256 checksums
    print_status "Creating checksums file..."
    {
        echo "Source tarball (insta-${VERSION}.tar.gz):"
        echo "SHA256: $(calculate_sha256 "${RELEASE_DIR}/insta-${VERSION}.tar.gz")"
        echo ""
        
        # Add checksums for each platform
        for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
            GOOS=${platform%/*}
            GOARCH=${platform##*/}
            BINARY_EXT="tar.gz"
            
            if [ "$GOOS" = "windows" ]; then
                BINARY_EXT="zip"
            fi
            
            echo "Binary archive (${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.${BINARY_EXT}):"
            echo "SHA256: $(calculate_sha256 "${RELEASE_DIR}/${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.${BINARY_EXT}")"
            echo ""
        done
    } > "${RELEASE_DIR}/checksums.txt"
}

# Function to build the Go binary
build_binary() {
    print_status "Building Go binary for ${GOOS}/${GOARCH}..."
    go build -ldflags "-s -w -X main.version=$VERSION -X main.buildTime=$BUILD_TIME" -o "${BINARY_NAME}-${GOOS}-${GOARCH}" ./cmd/insta
    
    # Apply UPX compression if available and supported
    if command -v upx >/dev/null 2>&1; then
        # Skip UPX on macOS ARM64
        if [ "$GOOS" != "darwin" ] || [ "$GOARCH" != "arm64" ]; then
            print_status "Compressing with UPX..."
            upx -q --best --lzma "${BINARY_NAME}-${GOOS}-${GOARCH}"
        else
            print_warning "Skipping UPX compression on macOS ARM64"
        fi
    else
        print_warning "UPX not found, skipping compression"
    fi
}

# Function to build Debian package
build_debian() {
    print_status "Building Debian package..."
    check_command "dpkg-buildpackage"
    
    # Create a temporary build directory
    TEMP_DIR=$(mktemp -d)
    cp -r packaging/debian/* "$TEMP_DIR/"
    cp "${BINARY_NAME}" "$TEMP_DIR/"
    
    # Update version in changelog
    sed -i "s/^insta (.*) unstable/insta ($VERSION) unstable/" "$TEMP_DIR/changelog"
    
    # Build the package
    cd "$TEMP_DIR"
    dpkg-buildpackage -us -uc
    
    # Move the package to build directory
    mv ../insta_${VERSION}_*.deb ../../$BUILD_DIR/
    cd ../..
    
    # Clean up
    rm -rf "$TEMP_DIR"
}

# Function to build RPM package
build_rpm() {
    print_status "Building RPM package..."
    check_command "rpmbuild"
    
    # Create RPM build directories
    mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
    
    # Copy spec file
    cp packaging/rpm/insta.spec ~/rpmbuild/SPECS/
    
    # Create source tarball
    tar -czf ~/rpmbuild/SOURCES/insta-${VERSION}.tar.gz \
        --transform "s,^,insta-${VERSION}/," \
        cmd/ docs/ LICENSE README.md
    
    # Build the package
    rpmbuild -ba ~/rpmbuild/SPECS/insta.spec
    
    # Copy the package to build directory
    cp ~/rpmbuild/RPMS/*/insta-${VERSION}*.rpm $BUILD_DIR/
}

# Function to build Arch package
build_arch() {
    print_status "Building Arch package..."
    check_command "makepkg"
    
    cd packaging/arch
    # Update version in PKGBUILD
    sed -i "s/^pkgver=.*/pkgver=$VERSION/" PKGBUILD
    
    # Build the package
    makepkg -f
    
    # Copy the package to build directory
    cp insta-${VERSION}*.pkg.tar.zst ../../$BUILD_DIR/
    cd ../..
}

# Function to build Chocolatey package
build_chocolatey() {
    print_status "Building Chocolatey package..."
    check_command "choco"
    
    cd packaging/chocolatey
    # Update version in nuspec
    sed -i "s/<version>.*<\/version>/<version>$VERSION<\/version>/" insta.nuspec
    
    # Build the package
    choco pack
    
    # Copy the package to build directory
    cp insta.${VERSION}.nupkg ../../$BUILD_DIR/
    cd ../..
}

# Function to publish Debian package
publish_debian() {
    print_status "Publishing Debian package..."
    check_command "dput"
    
    # Check for required environment variables
    if [ -z "$DEBIAN_REPOSITORY" ]; then
        print_error "DEBIAN_REPOSITORY environment variable not set"
        exit 1
    fi
    
    # Publish to Debian repository
    dput "$DEBIAN_REPOSITORY" "$BUILD_DIR/insta_${VERSION}_*.deb"
}

# Function to publish RPM package
publish_rpm() {
    print_status "Publishing RPM package..."
    check_command "rpm"
    
    # Check for required environment variables
    if [ -z "$RPM_REPOSITORY" ]; then
        print_error "RPM_REPOSITORY environment variable not set"
        exit 1
    fi
    
    # Sign the RPM package if GPG key is available
    if [ -n "$GPG_KEY_ID" ]; then
        print_status "Signing RPM package..."
        rpm --addsign "$BUILD_DIR/insta-${VERSION}*.rpm"
    fi
    
    # Copy to RPM repository
    cp "$BUILD_DIR/insta-${VERSION}*.rpm" "$RPM_REPOSITORY/"
    
    # Update repository metadata
    createrepo "$RPM_REPOSITORY"
}

# Function to publish Arch package
publish_arch() {
    print_status "Publishing Arch package..."
    check_command "aur"
    
    # Check for required environment variables
    if [ -z "$AUR_USERNAME" ]; then
        print_error "AUR_USERNAME environment variable not set"
        exit 1
    fi
    
    # Sign the package if GPG key is available
    if [ -n "$GPG_KEY_ID" ]; then
        print_status "Signing Arch package..."
        gpg --detach-sign "$BUILD_DIR/insta-${VERSION}*.pkg.tar.zst"
    fi
    
    # Update AUR package
    cd packaging/arch
    aur publish -u "$AUR_USERNAME" -p insta
    cd ../..
}

# Function to publish Chocolatey package
publish_chocolatey() {
    print_status "Publishing Chocolatey package..."
    check_command "choco"
    
    # Check for required environment variables
    if [ -z "$CHOCOLATEY_API_KEY" ]; then
        print_error "CHOCOLATEY_API_KEY environment variable not set"
        exit 1
    fi
    
    # Push to Chocolatey repository
    choco push "$BUILD_DIR/insta.${VERSION}.nupkg" --api-key "$CHOCOLATEY_API_KEY" --source https://push.chocolatey.org/
}

# Main build process
main() {
    print_status "Starting package build process..."
    
    # Build the Go binary for current platform
    build_binary
    
    # Create release archives if RELEASE=true
    if [ "${RELEASE:-false}" = "true" ]; then
        create_release_archives
    fi
    
    # Build packages based on platform if BUILD_PACKAGES=true
    if [ "$BUILD_PACKAGES" = "true" ]; then
        case "$(uname -s)" in
            Linux)
                if [ -f /etc/debian_version ]; then
                    build_debian
                    if [ "${PUBLISH:-false}" = "true" ]; then
                        publish_debian
                    fi
                elif [ -f /etc/redhat-release ]; then
                    build_rpm
                    if [ "${PUBLISH:-false}" = "true" ]; then
                        publish_rpm
                    fi
                elif [ -f /etc/arch-release ]; then
                    build_arch
                    if [ "${PUBLISH:-false}" = "true" ]; then
                        publish_arch
                    fi
                else
                    print_warning "Unsupported Linux distribution"
                fi
                ;;
            Darwin)
                print_warning "Package building on macOS is not fully supported"
                ;;
            *)
                print_warning "Unsupported operating system"
                ;;
        esac
        
        # Build Chocolatey package if on Windows
        if [[ "$(uname -s)" == "MINGW"* ]] || [[ "$(uname -s)" == "MSYS"* ]]; then
            build_chocolatey
            if [ "${PUBLISH:-false}" = "true" ]; then
                publish_chocolatey
            fi
        fi
        
        print_status "Build process completed. Packages are in $BUILD_DIR"
    fi
    
    if [ "${RELEASE:-false}" = "true" ]; then
        print_status "Release files are in $RELEASE_DIR"
    fi
}

# Run the main function
main 