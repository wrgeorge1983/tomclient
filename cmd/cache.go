package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"tomclient/auth"
)

var (
	cacheDevice string
	cacheAll    bool
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage command output cache",
	Long: `Manage the Tom API cache for device command outputs.
	
Cache entries store device command results to reduce load on network devices
and improve response times for frequently used commands.`,
	Example: `  tomclient cache stats
  tomclient cache list
  tomclient cache list --device router1
  tomclient cache invalidate router1
  tomclient cache clear --all`,
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics and configuration",
	Long:  `Display overall cache statistics including total entries, devices cached, and configuration settings.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true

		stats, err := client.GetCacheStats()
		handleError(err)

		fmt.Printf("Cache Statistics\n")
		fmt.Printf("================\n")
		fmt.Printf("Enabled:        %v\n", stats.Enabled)
		fmt.Printf("Total Entries:  %d\n", stats.TotalEntries)
		fmt.Printf("Devices Cached: %d\n", stats.DevicesCached)
		fmt.Printf("Default TTL:    %d seconds\n", stats.DefaultTTL)
		fmt.Printf("Max TTL:        %d seconds\n", stats.MaxTTL)
		fmt.Printf("Key Prefix:     %s\n", stats.KeyPrefix)

		if len(stats.EntriesPerDevice) > 0 {
			fmt.Printf("\nEntries Per Device:\n")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for device, count := range stats.EntriesPerDevice {
				fmt.Fprintf(w, "  %s\t%d\n", device, count)
			}
			w.Flush()
		}
	},
}

var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cache keys",
	Long:  `List all cache keys, optionally filtered by device name.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true

		keys, err := client.ListCacheKeys(cacheDevice)
		handleError(err)

		if keys.DeviceFilter != nil && *keys.DeviceFilter != "" {
			fmt.Printf("Cache Keys for Device: %s\n", *keys.DeviceFilter)
		} else {
			fmt.Printf("All Cache Keys\n")
		}
		fmt.Printf("Count: %d\n", keys.Count)

		if keys.Count > 0 {
			fmt.Printf("\nKeys:\n")
			for _, key := range keys.Keys {
				// Format keys nicely - they're in format device:command:hash
				parts := strings.SplitN(key, ":", 3)
				if len(parts) >= 2 {
					fmt.Printf("  %s: %s\n", parts[0], parts[1])
				} else {
					fmt.Printf("  %s\n", key)
				}
			}
		}
	},
}

var cacheInvalidateCmd = &cobra.Command{
	Use:   "invalidate <device-name>",
	Short: "Invalidate cache for a specific device",
	Long:  `Remove all cached command outputs for a specific device.`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Try to get device list from inventory cache for autocomplete
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
		result, err := client.InvalidateDeviceCache(deviceName)
		handleError(err)

		fmt.Println(result.Message)
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cache entries",
	Long: `Clear all cache entries across all devices.
Requires --all flag to confirm clearing all cache.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true

		if !cacheAll {
			fmt.Println("Error: --all flag required to clear all cache entries")
			fmt.Println("Usage: tomclient cache clear --all")
			os.Exit(1)
		}

		result, err := client.ClearAllCache()
		handleError(err)

		fmt.Println(result.Message)
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	// Add subcommands
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheInvalidateCmd)
	cacheCmd.AddCommand(cacheClearCmd)

	// Add flags
	cacheListCmd.Flags().StringVarP(&cacheDevice, "device", "d", "", "Filter by device name")
	cacheClearCmd.Flags().BoolVar(&cacheAll, "all", false, "Confirm clearing all cache entries")

	// Mark --all as required for clear command
	cacheClearCmd.MarkFlagRequired("all")
}
