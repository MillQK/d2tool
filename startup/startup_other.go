//go:build !windows
// +build !windows

package startup

import "fmt"

func StartupRegister(rawArgs []string) error {
	return fmt.Errorf("this functionality is only available on Windows")
}

func StartupRemove() error {
	return fmt.Errorf("this functionality is only available on Windows")
}
