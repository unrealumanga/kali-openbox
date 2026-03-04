#!/bin/bash

PODSEC_DIR="${PODSEC_DIR:-$HOME/GitHub/kali-openbox/tools/pod-sec}"
PODSEC_BIN="$PODSEC_DIR/pod-sec"
MAPPING_FILE="$PODSEC_DIR/tool-mapping.json"

if [[ ! -x "$PODSEC_BIN" ]]; then
    if command -v go &>/dev/null && [[ -f "$PODSEC_DIR/main.go" ]]; then
        echo "Building pod-sec..."
        cd "$PODSEC_DIR" && go build -o pod-sec main.go
    else
        return 1
    fi
fi

if [[ ! -f "$MAPPING_FILE" ]]; then
    return 1
fi

podsec_wrapper() {
    local tool="$1"
    
    if [[ -z "$tool" ]]; then
        command not_found_handler "$@"
        return $?
    fi
    
    if grep -q "\"$tool\"" "$MAPPING_FILE" 2>/dev/null; then
        shift
        "$PODSEC_BIN" "$tool" "$@"
        return $?
    fi
    
    command not_found_handler "$@"
    return $?
}

if [[ -n "$BASH_VERSION" ]]; then
    not_found_handler() {
        local cmd="$1"
        printf "bash: %s: command not found\n" "$cmd" >&2
        return 127
    }
elif [[ -n "$ZSH_VERSION" ]]; then
    not_found_handler() {
        local cmd="$1"
        printf "zsh: command not found: %s\n" "$cmd" >&2
        return 127
    }
fi

command_not_found_handler() {
    local cmd="$1"
    shift
    
    if grep -q "\"$cmd\"" "$MAPPING_FILE" 2>/dev/null; then
        "$PODSEC_BIN" "$cmd" "$@"
        return $?
    fi
    
    if [[ -n "$BASH_VERSION" ]]; then
        printf "bash: %s: command not found\n" "$cmd" >&2
    elif [[ -n "$ZSH_VERSION" ]]; then
        printf "zsh: command not found: %s\n" "$cmd" >&2
    fi
    return 127
}

if [[ -n "$BASH_VERSION" ]]; then
    export COMMAND_NOT_FOUND_HANDLER
fi
