//go:build !windows
// +build !windows

package steam

import (
	"fmt"
)

// FindSteamPath is a stub for non-Windows platforms
func FindSteamPath() (string, error) {
	return "", fmt.Errorf("this functionality is only available on Windows")
}
