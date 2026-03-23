//go:build !windows && !darwin && !linux

package steam

import "fmt"

func FindSteamPath() (string, error) {
	return "", fmt.Errorf("steam path auto-discovery is not supported on this platform")
}
