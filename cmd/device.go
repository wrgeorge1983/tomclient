package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"tomclient/auth"
)

var (
	deviceTimeout int
	deviceWait    bool
	deviceRaw     bool
	deviceUser    string
	devicePass    string
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

		var result string
		var err error

		if deviceUser != "" || devicePass != "" {
			result, err = client.SendDeviceCommandWithAuth(
				deviceName, command, deviceUser, devicePass,
				deviceWait, deviceRaw, deviceTimeout,
			)
		} else {
			result, err = client.SendDeviceCommand(deviceName, command, deviceWait, deviceRaw)
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
	deviceCmd.Flags().StringVarP(&deviceUser, "username", "u", "", "Override username for authentication")
	deviceCmd.Flags().StringVarP(&devicePass, "password", "p", "", "Override password for authentication")

	// Mark password flag as sensitive (won't show in help examples)
	deviceCmd.Flags().MarkHidden("password")
}
