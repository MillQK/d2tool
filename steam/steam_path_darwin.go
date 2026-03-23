//go:build darwin

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

	steamPath := filepath.Join(homeDir, "Library", "Application Support", "Steam")
	if _, err := os.Stat(steamPath); err == nil {
		return steamPath, nil
	}

	return "", fmt.Errorf("could not find Steam installation path")
}
