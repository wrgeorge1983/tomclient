#!/bin/zsh
# SSH autocomplete helper for zsh using tomclient inventory cache
# 
# Usage: Source this file in your ~/.zshrc:
#   source /path/to/ssh_complete.zsh
#
# Or add to ~/.zshrc:
#   eval "$(tomclient completion zsh)"
#   source /path/to/ssh_complete.zsh

_ssh_tomclient_complete() {
    local devices
    # Convert input to lowercase for case-insensitive matching
    local prefix_lower="${PREFIX:l}"
    devices=($(tomclient inventory --prefix="$prefix_lower" 2>/dev/null))
    
    if [ $? -eq 0 ] && [ ${#devices[@]} -gt 0 ]; then
        # Enable case-insensitive matching for zsh completion
        _describe -V 'devices' devices
    else
        # Fall back to default completion
        _ssh
    fi
}

# Enable case-insensitive matching globally for this completion
zstyle ':completion:*:ssh_tomclient_complete:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'

# Register completion for ssh command
compdef _ssh_tomclient_complete ssh

# Optional: Also complete for other commands
# compdef _ssh_tomclient_complete scp
# compdef _ssh_tomclient_complete sftp
# compdef _ssh_tomclient_complete ping
