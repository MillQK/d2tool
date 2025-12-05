//go:build windows
// +build windows

package startup

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	appRegistryName     = "D2Tool"
	startupRegistryPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Run"
)

type startupServiceWindowsImpl struct {
	runArgs []string
}

func NewStartupService(runArgs []string) StartupService {
	return &startupServiceWindowsImpl{
		runArgs: runArgs,
	}
}

func (s *startupServiceWindowsImpl) StartupRegister() error {
	// Path to your application's executable
	appExecutable, err := os.Executable()
	if err != nil {
		slog.Warn(fmt.Sprintf("Error getting executable path: %v", err))
		return err
	}

	// Open "Run" registry key
	key, _, err := registry.CreateKey(registry.CURRENT_USER, startupRegistryPath, registry.ALL_ACCESS)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error opening registry key: %v", err))
		return err
	}
	defer key.Close()

	// Write the path of the executable to the registry
	err = key.SetStringValue(appRegistryName, s.executableRunCommand(appExecutable))
	if err != nil {
		slog.Warn(fmt.Sprintf("Error setting registry value: %v", err))
		return err
	}

	slog.Info("Application added to Windows Startup successfully.")
	return nil
}

func (s *startupServiceWindowsImpl) StartupRemove() error {
	// Open "Run" registry key
	key, _, err := registry.CreateKey(registry.CURRENT_USER, startupRegistryPath, registry.ALL_ACCESS)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error opening registry key: %v", err))
		return err
	}
	defer key.Close()

	// Delete the value
	err = key.DeleteValue(appRegistryName)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error deleting registry value: %v", err))
		return err
	}

	slog.Info("Application removed from Windows Startup successfully.")
	return nil
}

func (s *startupServiceWindowsImpl) IsStartupRegistered() (bool, error) {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, startupRegistryPath, registry.ALL_ACCESS)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error opening registry key: %v", err))
		return false, fmt.Errorf("error opening registry key: %w", err)
	}
	defer key.Close()

	value, _, err := key.GetStringValue(appRegistryName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("error getting registry value: %w", err)
	}

	appExecutable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("error getting executable path: %w", err)
	}

	if s.executableRunCommand(appExecutable) != value {
		return false, nil
	}

	return true, nil
}

func (s *startupServiceWindowsImpl) SupportsStartup() bool {
	return true
}

func (s *startupServiceWindowsImpl) executableRunCommand(appExecutable string) string {
	return fmt.Sprintf("\"%s\" %s", appExecutable, strings.Join(s.runArgs, " "))
}
