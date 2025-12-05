package main

import (
	"d2tool/config"
	"fmt"
	"log/slog"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppUpdateState represents the state for the app update tab
type AppUpdateState struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	LastCheckTime   string `json:"lastCheckTime"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

// --- Heroes Layout Update ---

// UpdateHeroesLayout performs the hero layout update synchronously
func (a *App) UpdateHeroesLayout() error {
	if err := a.heroesLayoutService.UpdateHeroesLayout(); err != nil {
		return fmt.Errorf("error updating hero layout: %w", err)
	}

	return nil
}

// --- Heroes Layout Files Bindings ---

// GetHeroesLayoutFiles returns the list of hero layout config files
func (a *App) GetHeroesLayoutFiles() []config.FileConfig {
	return a.config.GetHeroesLayoutFiles()
}

// AddHeroesLayoutFile adds a new config file
func (a *App) AddHeroesLayoutFile(path string) {
	a.config.AddHeroesLayoutFile(path)
}

// RemoveHeroesLayoutFile removes a config file by index
func (a *App) RemoveHeroesLayoutFile(index int) {
	a.config.RemoveHeroesLayoutFile(index)
}

// SetHeroesLayoutFileEnabled enables or disables a file by index
func (a *App) SetHeroesLayoutFileEnabled(index int, enabled bool) {
	a.config.SetHeroesLayoutFileEnabled(index, enabled)
}

// OpenFileDialog opens a file dialog and returns the selected path
func (a *App) OpenFileDialog() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select hero_grid_config.json",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("error opening file dialog: %w", err)
	}

	return selection, nil
}

// --- Positions Bindings ---

// GetPositions returns the current positions
func (a *App) GetPositions() []config.PositionConfig {
	return a.config.GetPositions()
}

// SetPositions updates the positions (order and enabled state)
func (a *App) SetPositions(positions []config.PositionConfig) {
	a.config.SetPositions(positions)
}

// SetPositionEnabled enables or disables a position by ID
func (a *App) SetPositionEnabled(id string, enabled bool) {
	a.config.SetPositionEnabled(id, enabled)
}

// --- Startup Tab Bindings ---

// GetStartupEnabled returns whether the app is set to run on startup
func (a *App) GetStartupEnabled() (bool, error) {
	if !a.startupService.SupportsStartup() {
		return false, nil
	}
	registered, err := a.startupService.IsStartupRegistered()
	if err != nil {
		slog.Warn("Error checking startup registration", "error", err)
		return false, fmt.Errorf("error checking startup registration: %w", err)
	}
	return registered, nil
}

// SetStartupEnabled enables or disables running on startup
func (a *App) SetStartupEnabled(enabled bool) error {
	if !a.startupService.SupportsStartup() {
		return nil
	}
	if enabled {
		return a.startupService.StartupRegister()
	}
	return a.startupService.StartupRemove()
}

// IsStartupSupported returns whether startup registration is supported on this platform
func (a *App) IsStartupSupported() bool {
	return a.startupService.SupportsStartup()
}

// --- App Update Tab Bindings ---

// GetAppUpdateState returns the current state for the app update tab
func (a *App) GetAppUpdateState() AppUpdateState {
	updateState := a.updateService.GetState()

	var lastCheckTimeStr string
	if updateState.LastCheckTime.IsZero() {
		lastCheckTimeStr = "Never"
	} else {
		lastCheckTimeStr = updateState.LastCheckTime.Format("2006-01-02 15:04:05")
	}

	return AppUpdateState{
		CurrentVersion:  updateState.CurrentAppVersion,
		LatestVersion:   updateState.LatestAppVersion,
		LastCheckTime:   lastCheckTimeStr,
		UpdateAvailable: updateState.UpdateAvailable,
	}
}

// CheckForAppUpdate checks for application updates synchronously
func (a *App) CheckForAppUpdate() error {
	if err := a.updateService.CheckForUpdate(); err != nil {
		return fmt.Errorf("error checking for updates: %w", err)
	}

	return nil
}

// DownloadAppUpdate downloads and installs the update synchronously
func (a *App) DownloadAppUpdate() error {
	if err := a.updateService.UpdateApp(); err != nil {
		return fmt.Errorf("error downloading update: %w", err)
	}

	return nil
}

// --- Background Tasks ---

func (a *App) startBackgroundTasks() {
	// Start periodic hero layout update
	go func() {
		delay := time.Hour

		for {
			// Find the most recent update time across all files
			files := a.config.GetHeroesLayoutFiles()
			var lastUpdateTime time.Time
			for _, f := range files {
				if f.LastUpdateTimestampMillis > 0 {
					t := time.UnixMilli(f.LastUpdateTimestampMillis)
					if t.After(lastUpdateTime) {
						lastUpdateTime = t
					}
				}
			}

			var waitDuration time.Duration
			if lastUpdateTime.IsZero() {
				// No files have been updated yet, wait for the full delay
				waitDuration = delay
			} else {
				nextUpdate := lastUpdateTime.Add(delay)
				waitDuration = time.Until(nextUpdate)
				if waitDuration < 0 {
					waitDuration = 0
				}
			}

			<-time.After(waitDuration)

			// Perform background update
			a.UpdateHeroesLayout()

			// Notify frontend that data was updated
			runtime.EventsEmit(a.ctx, "heroesLayoutDataChanged")
		}
	}()

	// Start periodic app update check
	go func() {
		// Check for updates on startup after a short delay
		slog.Info("Checking for updates on startup")
		a.CheckForAppUpdate()
		runtime.EventsEmit(a.ctx, "appUpdateDataChanged")

		// Periodic check loop
		for {
			<-time.After(1 * time.Hour)
			slog.Debug("Checking for updates after timeout")
			a.CheckForAppUpdate()
			runtime.EventsEmit(a.ctx, "appUpdateDataChanged")
		}
	}()
}
