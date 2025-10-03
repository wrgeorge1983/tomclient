# Shell Autocomplete Setup

tomclient provides device name autocomplete for SSH and other commands using cached inventory from the Tom API.

## Quick Setup

### Bash

Add to `~/.bashrc`:

```bash
# Enable tomclient completion
eval "$(tomclient completion bash)"

# Enable SSH autocomplete with Tom inventory
source /path/to/tomclient/shell/ssh_complete.sh
```

Then reload:
```bash
source ~/.bashrc
```

### Zsh

Add to `~/.zshrc`:

```zsh
# Enable tomclient completion
eval "$(tomclient completion zsh)"

# Enable SSH autocomplete with Tom inventory
source /path/to/tomclient/shell/ssh_complete.zsh
```

Then reload:
```zsh
source ~/.zshrc
```

## Initial Cache Population

Before autocomplete works, populate the inventory cache:

```bash
# Fetch and cache all devices
tomclient inventory --refresh

# Verify cache
tomclient inventory
```

## Usage

Once configured, SSH and tomclient commands will autocomplete device names:

```bash
# SSH autocomplete (case-insensitive)
ssh router<TAB>
# Completes to: router1, router2, Router-Core-1, ROUTER2, etc.

ssh ROUTER<TAB>
# Same results - matching is case-insensitive

# tomclient device command autocomplete
tomclient device switch<TAB>
# Completes to: switch1, switch2, switch-access-1, SWITCH3, etc.
```

**Note:** Autocomplete is case-insensitive since hostnames are case-insensitive. You can type in any case and get matches.

## Cache Management

### View Cached Devices

```bash
# List all
tomclient inventory

# Filter by prefix
tomclient inventory --prefix=core
```

### Refresh Cache

The cache automatically refreshes when expired (default: 1 hour), but you can force refresh:

```bash
tomclient inventory --refresh
```

### Clear Cache

```bash
tomclient inventory clear
```

### Configure Cache TTL

Set custom TTL via environment variable:

```bash
# 30 minutes
export TOM_CACHE_TTL=30m

# 2 hours
export TOM_CACHE_TTL=2h

# 24 hours
export TOM_CACHE_TTL=24h
```

Add to your shell rc file to make permanent.

## How It Works

1. **Cache**: `tomclient inventory` fetches device list from Tom API and caches to `~/.tom/inventory_cache.json`
2. **TTL**: Cache expires after configured TTL (default: 1 hour)
3. **Autocomplete**: Shell completion scripts read from cache for instant results
4. **Auto-refresh**: Cache refreshes automatically when accessing expired cache

## Troubleshooting

### Autocomplete Not Working

1. **Check cache exists:**
   ```bash
   ls ~/.tom/inventory_cache.json
   ```

2. **Populate cache:**
   ```bash
   tomclient inventory --refresh
   ```

3. **Test manually:**
   ```bash
   tomclient inventory --prefix=test
   ```

4. **Verify shell rc is sourced:**
   ```bash
   # Re-source your shell rc
   source ~/.bashrc  # or ~/.zshrc
   ```

### Cache is Stale

Force refresh:
```bash
tomclient inventory --refresh
```

Or clear and let it refresh automatically:
```bash
tomclient inventory clear
tomclient inventory
```

### Permission Issues

Ensure config directory is writable:
```bash
ls -ld ~/.tom
chmod 700 ~/.tom
```

## Advanced Usage

### Custom Commands

You can extend the completion scripts to work with other commands besides SSH:

**Bash (`ssh_complete.sh`):**
```bash
# Also complete for scp and sftp
complete -F _ssh_tomclient_complete scp
complete -F _ssh_tomclient_complete sftp
```

**Zsh (`ssh_complete.zsh`):**
```zsh
# Also complete for scp and sftp
compdef _ssh_tomclient_complete scp
compdef _ssh_tomclient_complete sftp
```

### Different Config Directory

If using a custom config directory:

```bash
export TOM_CONFIG_DIR=/custom/path
tomclient inventory --refresh
```

The completion will automatically use the same config directory.
