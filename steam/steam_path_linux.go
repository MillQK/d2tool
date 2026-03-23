//go:build linux

package steam

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindSteamPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}

	candidates := []string{
		filepath.Join(homeDir, ".steam", "steam"),
		filepath.Join(homeDir, ".local", "share", "Steam"),
	}

	for _, steamPath := range candidates {
		if _, err := os.Stat(steamPath); err == nil {
			return steamPath, nil
		}
	}

	return "", fmt.Errorf("could not find Steam installation path")
}
