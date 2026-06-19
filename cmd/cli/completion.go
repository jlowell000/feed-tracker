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
        COMPREPLY=($(compgen -W "migrate add fetch feeds feed folder import export delete list search read unread open prune vacuum completion" -- "$cur"))
        return
    fi

    case ${words[1]} in
        list|fetch|delete)
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
        feed)
            if [ $cword -eq 2 ]; then
                COMPREPLY=($(compgen -W "update" -- "$cur"))
            fi
            if [ $cword -ge 3 ] && [ ${words[2]} = "update" ]; then
                COMPREPLY=($(compgen -W "--title --url" -- "$cur"))
            fi
            ;;
        folder)
            if [ $cword -eq 2 ]; then
                COMPREPLY=($(compgen -W "create delete rename move" -- "$cur"))
            fi
            ;;
        import)
            if [ $cword -eq 2 ]; then
                COMPREPLY=($(compgen -f -- "$cur"))
            fi
            if [[ "$prev" == "--dry-run" ]]; then
                COMPREPLY=($(compgen -f -- "$cur"))
            fi
            ;;
        export)
            case $prev in
                --output)
                    COMPREPLY=($(compgen -f -- "$cur"))
                    ;;
            esac
            ;;
        completion)
            if [ $cword -eq 2 ]; then
                COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
            fi
            ;;
        add|read|unread)
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
        'feed:Manage feed settings'
        'folder:Manage folders (create/delete/rename/move)'
        'import:Import feeds from OPML file'
        'export:Export feeds to OPML file'
        'delete:Delete a feed and all its entries'
        'list:List entries'
        'read:Mark entry as read'
        'unread:Mark entry as unread'
        'search:Search entries by keyword'
        'open:Open entry URL in system browser'
        'prune:Delete entries older than configured max_age'
        'vacuum:Reclaim database storage space'
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
                list|fetch|delete)
                    _alternative \
                        'feed-name::feed name:($(ft feeds --names 2>/dev/null))' \
                        'feed-id:feed id:'
                    ;;
                feed)
                    _arguments '2:subcommand:(update)' \
                        '--title[new feed title]' \
                        '--url[new feed URL]'
                    ;;
                folder)
                    _arguments '2:subcommand:(create delete rename move)'
                    ;;
                import)
                    _arguments '--dry-run[dry run preview]' '1:opml file:_files'
                    ;;
                export)
                    _arguments '--output[output file]:file:_files' \
                        '--folders-only[export only feeds in folders]' \
                        '--feeds-only[export only feeds without a folder]'
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
