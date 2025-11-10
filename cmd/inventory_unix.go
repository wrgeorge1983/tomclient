//go:build unix

package cmd

import (
	"fmt"
	"os"
)

func getHostsFilePath() string {
	return "/etc/hosts"
}

func checkAdminPrivileges() error {
	// On Unix-like systems, check if the user is root
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (sudo) to update /etc/hosts")
	}
	return nil
}
