package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"tomclient/auth"
)

var (
	deviceTimeout      int
	deviceWait         bool
	deviceRaw          bool
	deviceUser         string
	devicePass         string
	deviceCache        bool
	deviceCacheTTL     int
	deviceCacheRefresh bool
)

var deviceCmd = &cobra.Command{
	Use:   "device <device-name> <command>",
	Short: "Run command on a network device",
	Long: `Execute a command on a specific network device through the Tom API.
Supports credential override and timeout configuration.`,
	Example: `  tomclient device router1 "show version" --timeout=30
  tomclient device switch2 "show interface" -t 60 --raw
  tomclient device -u admin -p secret router3 "show running-config"`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cache, err := auth.LoadInventoryCache(cfg.ConfigDir)
		if err != nil || cache == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return cache.Devices, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true

		deviceName := args[0]
		command := args[1]

		// Load config to get cache defaults
		cfg, _ := auth.LoadConfig(configDir)

		// Determine cache settings
		useCache := cfg.CacheEnabled // default from config
		// Command-line flag overrides config if explicitly set
		if cmd.Flags().Changed("cache") {
			useCache = deviceCache
		}

		// Set cache TTL
		var cacheTTL *int
		if cmd.Flags().Changed("cache-ttl") && deviceCacheTTL > 0 {
			cacheTTL = &deviceCacheTTL
		} else if cfg.CacheTTL > 0 && useCache {
			cacheTTL = &cfg.CacheTTL
		}

		var result string
		var err error

		if deviceUser != "" || devicePass != "" {
			result, err = client.SendDeviceCommandWithAuth(
				deviceName, command, deviceUser, devicePass,
				deviceWait, deviceRaw, deviceTimeout,
				useCache, cacheTTL, deviceCacheRefresh,
			)
		} else {
			result, err = client.SendDeviceCommand(deviceName, command, deviceWait, deviceRaw,
				useCache, cacheTTL, deviceCacheRefresh)
		}

		handleError(err)
		fmt.Print(result)
	},
}

func init() {
	rootCmd.AddCommand(deviceCmd)

	// POSIX-style flags with both long and short versions
	deviceCmd.Flags().IntVarP(&deviceTimeout, "timeout", "t", 10, "Command timeout in seconds")
	deviceCmd.Flags().BoolVarP(&deviceWait, "wait", "w", true, "Wait for command completion")
	deviceCmd.Flags().BoolVarP(&deviceRaw, "raw", "r", true, "Return raw command output")

	// Cache control flags (grouped together)
	// Note: default here doesn't matter since we check if it was changed
	deviceCmd.Flags().BoolVarP(&deviceCache, "cache", "c", false, "Enable/disable caching (use --cache=false or -c=false to disable)")
	deviceCmd.Flags().IntVarP(&deviceCacheTTL, "cache-ttl", "T", 0, "Cache TTL in seconds (0 uses server default)")
	deviceCmd.Flags().BoolVarP(&deviceCacheRefresh, "cache-refresh", "R", false, "Force refresh cached result")

	// Authentication flags
	deviceCmd.Flags().StringVarP(&deviceUser, "username", "u", "", "Override username for authentication")
	deviceCmd.Flags().StringVarP(&devicePass, "password", "p", "", "Override password for authentication")

	// Mark password flag as sensitive (won't show in help examples)
	deviceCmd.Flags().MarkHidden("password")
}
