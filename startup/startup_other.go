//go:build !windows
// +build !windows

package startup

import "fmt"

func StartupRegister(args []string) error {
	return fmt.Errorf("this functionality is only available on Windows")
}

func StartupRemove() error {
	return fmt.Errorf("this functionality is only available on Windows")
}

func IsStartupRegistered() (bool, error) {
	return false, fmt.Errorf("this functionality is only available on Windows")
}

func SupportsStartup() bool {
	return false
}
