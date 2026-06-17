package main

import (
	"fmt"
	"os"
)

func runCompletion(shell string) {
	switch shell {
	case "bash":
		fmt.Print(bashCompletionScript)
	case "zsh":
		fmt.Print(zshCompletionScript)
	default:
		fmt.Fprintf(os.Stderr, "usage: ft completion bash|zsh\n")
		os.Exit(1)
	}
}

const bashCompletionScript = `_ft_completions() {
    local cur prev words cword
    _init_completion || return

    if [ $cword -eq 1 ]; then
        COMPREPLY=($(compgen -W "migrate add fetch feeds list completion" -- "$cur"))
        return
    fi

    case ${words[1]} in
        list|fetch)
            case $prev in
                --feed-id)
                    return
                    ;;
                --limit)
                    return
                    ;;
                *)
                    COMPREPLY=($(compgen -W "$(ft feeds --names 2>/dev/null)" -- "$cur"))
                    ;;
            esac
            ;;
        completion)
            if [ $cword -eq 2 ]; then
                COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
            fi
            ;;
        add)
            ;;
    esac
} &&
complete -F _ft_completions ft
`

const zshCompletionScript = `#compdef ft

_ft() {
    local -a commands
    commands=(
        'migrate:Create or update the database schema'
        'add:Add a new feed by URL'
        'fetch:Fetch new entries from feed(s)'
        'feeds:List all tracked feeds'
        'list:List entries'
        'completion:Generate shell completion script'
    )

    _arguments -C \
        '--config[path to config file]:config file:_files' \
        '1:command:->subcommand' \
        '*::args:->args' && return

    case $state in
        subcommand)
            _describe 'command' commands
            ;;
        args)
            case $words[1] in
                list|fetch)
                    _alternative \
                        'feed-name::feed name:($(ft feeds --names 2>/dev/null))' \
                        'feed-id:feed id:'
                    ;;
                completion)
                    _arguments '2:shell:(bash zsh)'
                    ;;
            esac
            ;;
    esac
}

_ft "$@"
`
