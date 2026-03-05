#!/bin/bash
#
# Kali Openbox ISO Build Script
# Requires: live-build, debootstrap, and ~40GB disk space
#

set -e

KALI_VERSION="kali-rolling"
BUILD_DIR="build"
OUTPUT_DIR="output"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root"
    exit 1
fi

# Check dependencies
check_deps() {
    log_info "Checking dependencies..."
    
    MISSING=""
    for cmd in lb debootstrap; do
        if ! command -v $cmd &> /dev/null; then
            MISSING="$MISSING $cmd"
        fi
    done
    
    if [ -n "$MISSING" ]; then
        log_error "Missing dependencies:$MISSING"
        log_info "Install with: apt-get install live-build"
        exit 1
    fi
}

# Setup build directory
setup_build() {
    log_info "Setting up build directory..."
    
    rm -rf $BUILD_DIR
    mkdir -p $BUILD_DIR
    cd $BUILD_DIR
    
    # Initialize live-build
    if [ ! -f /usr/share/keyrings/kali-archive-keyring.gpg ]; then
        log_info "Downloading Kali keyring..."
        mkdir -p /usr/share/keyrings
        wget -q https://archive.kali.org/archive-key.asc -O - | gpg --dearmor > /usr/share/keyrings/kali-archive-keyring.gpg
        chmod 644 /usr/share/keyrings/kali-archive-keyring.gpg
    fi

    lb config \
        --distribution $KALI_VERSION \
        --architectures amd64 \
        --binary-image iso-hybrid \
        --bootappend-live "boot=live components hostname=kali-openbox username=kali" \
        --mirror-bootstrap http://http.kali.org/kali \
        --mirror-chroot http://http.kali.org/kali \
        --mirror-chroot-security http://http.kali.org/kali \
        --mirror-binary http://http.kali.org/kali \
        --mirror-binary-security http://http.kali.org/kali \
        --parent-mirror-bootstrap http://http.kali.org/kali \
        --parent-mirror-chroot http://http.kali.org/kali \
        --parent-mirror-chroot-security http://http.kali.org/kali \
        --parent-mirror-binary http://http.kali.org/kali \
        --parent-mirror-binary-security http://http.kali.org/kali \
        --linux-flavours amd64 \
        --source false \
        --debian-installer false
}

# Copy configs
copy_configs() {
    log_info "Copying configuration files..."
    
    # Package lists
    cp -r ../config/package-lists/* config/package-lists/
    
    # Hooks
    mkdir -p config/hooks
    cp -r ../config/hooks/* config/hooks/ 2>/dev/null || true
    
    # Preseeds
    mkdir -p config/preseed
    cp -r ../config/preseed/* config/preseed/ 2>/dev/null || true
}

# Build ISO
build_iso() {
    log_info "Building ISO (this may take a while)..."
    
    # Build
    lb build 2>&1 | tee $OUTPUT_DIR/build.log
    
    if [ -f binary.iso ]; then
        mv binary.iso $OUTPUT_DIR/kali-openbox-$(date +%Y%m%d).iso
        log_info "ISO built successfully!"
    else
        log_error "Build failed. Check log."
        exit 1
    fi
}

# Main
main() {
    echo "==================================="
    echo "  Kali Openbox ISO Builder"
    echo "==================================="
    
    check_deps
    setup_build
    copy_configs
    build_iso
    
    echo ""
    echo "==================================="
    echo "  Build complete!"
    echo "  ISO location: $OUTPUT_DIR/"
    echo "==================================="
}

main "$@"
