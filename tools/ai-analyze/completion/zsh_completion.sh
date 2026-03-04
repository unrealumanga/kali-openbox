#compdef ai-analyze

_ai_analyze() {
    local -a opts
    opts=(
        '-h[Show help message]'
        '--help[Show help message]'
        '-o[Save report to file]:output file:_files'
        '--output[Save report to file]:output file:_files'
        '--no-color[Disable colored output]'
    )

    _arguments -s -w 800 ${opts} '*:file:_files'
}

_ai_analyze "$@"
