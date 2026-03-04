#!/bin/bash

# Smart Recon Pipeline Wrapper
# Usage: ./recon.sh <target>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="auto-recon"

check_dependencies() {
    local missing=()
    
    for tool in nmap dirb dnsenum; do
        if ! command -v "$tool" &> /dev/null; then
            missing+=("$tool")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        echo "Missing dependencies: ${missing[*]}"
        echo "Please install missing tools before running."
        exit 1
    fi
    
    if [ ! -f "$SCRIPT_DIR/$BINARY_NAME" ]; then
        echo "Building auto-recon..."
        cd "$SCRIPT_DIR"
        if command -v go &> /dev/null; then
            go build -o "$BINARY_NAME" main.go config.yaml
        else
            echo "Go not found. Please install Go or use pre-built binary."
            exit 1
        fi
    fi
}

main() {
    if [ $# -eq 0 ]; then
        echo "Usage: $0 <target>"
        echo "Target: IP address or domain name"
        echo ""
        echo "Examples:"
        echo "  $0 192.168.1.1"
        echo "  $0 example.com"
        echo "  $0 10.0.0.1,example.com"
        exit 1
    fi
    
    check_dependencies
    
    TARGET="$1"
    
    if [[ "$TARGET" == *,* ]]; then
        IFS=',' read -ra TARGETS <<< "$TARGET"
        for t in "${TARGETS[@]}"; do
            echo "Running recon on: $t"
            "$SCRIPT_DIR/$BINARY_NAME" "$t"
        done
    else
        "$SCRIPT_DIR/$BINARY_NAME" "$TARGET"
    fi
}

main "$@"
