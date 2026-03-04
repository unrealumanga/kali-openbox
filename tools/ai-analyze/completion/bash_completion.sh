#!/bin/bash

_ai_analyze_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="-h --help -o --output --no-color"

    case "${prev}" in
        -o|--output)
            _filedir
            return 0
            ;;
        *)
            ;;
    esac

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    _filedir
}

complete -F _ai_analyze_completion ai-analyze
