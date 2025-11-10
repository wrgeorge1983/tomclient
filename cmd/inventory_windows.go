//go:build windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func getHostsFilePath() string {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}
	return filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")
}

func checkAdminPrivileges() error {
	// On Windows, check if the user has administrative privileges
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(&windows.SECURITY_NT_AUTHORITY, 2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return fmt.Errorf("failed to allocate SID (to check admin privileges): %w", err)
	}
	defer windows.FreeSid(sid)
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return fmt.Errorf("failed to check admin privileges: %w", err)
	}
	if !member {
		return fmt.Errorf("must run as administrator to update %s", getHostsFilePath())
	}
	return nil
}
