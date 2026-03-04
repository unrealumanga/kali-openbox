#!/usr/bin/env zsh

nl2cmd() {
    local prompt="$*"
    local api_url="http://localhost:8080/api/command"
    
    if [[ -z "$prompt" ]]; then
        echo "Usage: nl2cmd <natural language command>"
        return 1
    fi
    
    echo "🤖 NL2Cmd: Converting to command..."
    echo "   Prompt: $prompt"
    echo ""
    
    local response
    response=$(curl -s -X POST "$api_url" \
        -H "Content-Type: application/json" \
        -d "{\"prompt\": \"$prompt\", \"context\": \"linux\"}" 2>/dev/null)
    
    if [[ -z "$response" ]]; then
        echo "❌ Error: Could not connect to ai-core daemon (localhost:8080)"
        echo "   Make sure the daemon is running."
        return 1
    fi
    
    local command
    command=$(echo "$response" | grep -oP '"command"\s*:\s*"\K[^"]+' 2>/dev/null)
    
    if [[ -z "$command" ]]; then
        echo "❌ Error: Invalid response from ai-core"
        echo "   Response: $response"
        return 1
    fi
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📝 Suggested command:"
    echo ""
    echo "   $command"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    read -q "exec?Execute this command? [y/N] "
    echo ""
    
    if [[ "$exec" =~ ^[Yy]$ ]]; then
        echo "▶ Executing..."
        eval "$command"
    else
        echo "⏸ Command not executed. Copy it from above."
    fi
}

ai() {
    nl2cmd "$@"
}

_nl2cmd_complete() {
    return 0
}

compdef _nl2cmd_complete nl2cmd
compdef _nl2cmd_complete ai
