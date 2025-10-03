#!/bin/bash
# SSH autocomplete helper using tomclient inventory cache
# 
# Usage: Source this file in your shell rc file:
#   source /path/to/ssh_complete.sh
#
# Or add to ~/.bashrc or ~/.zshrc:
#   eval "$(tomclient completion bash)"  # or zsh
#   source /path/to/ssh_complete.sh

_ssh_tomclient_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    
    # Get device names from tomclient inventory cache
    # Convert input to lowercase for case-insensitive matching
    local devices=$(tomclient inventory --prefix="${cur,,}" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$devices" ]; then
        # Use compgen with case-insensitive matching
        COMPREPLY=($(compgen -W "$devices" -- "$cur"))
        
        # If no matches with case-sensitive compgen, try case-insensitive
        if [ ${#COMPREPLY[@]} -eq 0 ]; then
            local IFS=$'\n'
            COMPREPLY=($(echo "$devices" | grep -i "^$cur"))
        fi
    else
        # Fall back to default completion if cache unavailable
        COMPREPLY=()
    fi
}

# Register completion for ssh command
complete -F _ssh_tomclient_complete ssh

# Optional: Also complete for other commands
# complete -F _ssh_tomclient_complete scp
# complete -F _ssh_tomclient_complete sftp
# complete -F _ssh_tomclient_complete ping
