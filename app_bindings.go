package main

import (
	"d2tool/config"
	"d2tool/steam"
	"fmt"
	"log/slog"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppUpdateState represents the state for the app update tab
type AppUpdateState struct {
	CurrentVersion      string `json:"currentVersion"`
	LatestVersion       string `json:"latestVersion"`
	LastCheckTimeMillis int64  `json:"lastCheckTimeMillis"`
	UpdateAvailable     bool   `json:"updateAvailable"`
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

// RemoveHeroesLayoutFile removes a config file by path
func (a *App) RemoveHeroesLayoutFile(filePath string) {
	a.config.RemoveHeroesLayoutFile(filePath)
}

// SetHeroesLayoutFileEnabled enables or disables a file by path
func (a *App) SetHeroesLayoutFileEnabled(filePath string, enabled bool) {
	a.config.SetHeroesLayoutFileEnabled(filePath, enabled)
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

// --- D2PT Provider Bindings ---

// GetD2PTConfig returns the D2PT provider configuration
func (a *App) GetD2PTConfig() config.D2PTConfig {
	return a.config.GetD2PTConfig()
}

// SetD2PTPeriod sets the D2PT period parameter
func (a *App) SetD2PTPeriod(period string) {
	a.config.SetD2PTPeriod(period)
}

// --- Heroes Layout Settings Bindings ---

// GetHeroesPerRow returns the configured heroes per row
func (a *App) GetHeroesPerRow() int {
	return a.config.GetHeroesPerRow()
}

// SetHeroesPerRow sets the heroes per row value
func (a *App) SetHeroesPerRow(heroesPerRow int) error {
	return a.config.SetHeroesPerRow(heroesPerRow)
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

	var lastCheckTimeMillis int64
	if !updateState.LastCheckTime.IsZero() {
		lastCheckTimeMillis = updateState.LastCheckTime.UnixMilli()
	}

	return AppUpdateState{
		CurrentVersion:      updateState.CurrentAppVersion,
		LatestVersion:       updateState.LatestAppVersion,
		LastCheckTimeMillis: lastCheckTimeMillis,
		UpdateAvailable:     updateState.UpdateAvailable,
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

// --- Steam Bindings ---

func (a *App) GetSteamConfig() config.SteamConfig {
	return a.config.GetSteamConfig()
}

func (a *App) SetSteamPath(path string) error {
	a.config.SetSteamPath(path)
	if err := a.steamService.Scan(); err != nil {
		return fmt.Errorf("error scanning steam accounts: %w", err)
	}
	runtime.EventsEmit(a.ctx, EventSteamPathChanged)
	runtime.EventsEmit(a.ctx, EventSteamAccountsChanged)
	return nil
}

func (a *App) SetAutoEnableNewAccounts(enabled bool) {
	a.config.SetAutoEnableNewAccounts(enabled)
}

func (a *App) GetSteamAccounts() []steam.SteamAccountView {
	return a.steamService.GetAccounts()
}

func (a *App) SetSteamAccountEnabled(steamId64 string, enabled bool) {
	a.steamService.SetAccountEnabled(steamId64, enabled)
}

func (a *App) RescanSteamAccounts() error {
	if err := a.steamService.Scan(); err != nil {
		return fmt.Errorf("error scanning steam accounts: %w", err)
	}
	runtime.EventsEmit(a.ctx, EventSteamAccountsChanged)
	return nil
}

func (a *App) OpenDirectoryDialog() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Steam Directory",
	})
	if err != nil {
		return "", fmt.Errorf("error opening directory dialog: %w", err)
	}
	return selection, nil
}

func (a *App) IsSteamPathValid() bool {
	return a.steamService.IsPathValid()
}

// --- Background Tasks ---

func (a *App) startBackgroundTasks() {
	// Start periodic hero layout update
	go func() {
		delay := time.Hour

		for {
			var lastUpdateTime time.Time
			updateIfNewer := func(millis int64) {
				if millis > 0 {
					if t := time.UnixMilli(millis); t.After(lastUpdateTime) {
						lastUpdateTime = t
					}
				}
			}
			for _, f := range a.config.GetHeroesLayoutFiles() {
				updateIfNewer(f.LastUpdateTimestampMillis)
			}
			for _, acc := range a.config.GetSteamAccounts() {
				updateIfNewer(acc.LastUpdateTimestampMillis)
			}

			var waitDuration time.Duration
			if lastUpdateTime.IsZero() {
				waitDuration = delay
			} else {
				nextUpdate := lastUpdateTime.Add(delay)
				waitDuration = time.Until(nextUpdate)
				if waitDuration < 0 {
					waitDuration = 0
				}
			}

			select {
			case <-a.ctx.Done():
				slog.Info("Stopping background hero layout update task")
				return
			case <-time.After(waitDuration):
				// continue
			}

			// Rescan Steam accounts
			if err := a.steamService.Scan(); err != nil {
				slog.Warn("Error scanning steam accounts", "error", err)
			}

			slog.Info("Performing hero layout update after timeout")
			if err := a.heroesLayoutService.UpdateHeroesLayout(); err != nil {
				slog.Warn("Error updating hero layout", "error", err)
			}

			runtime.EventsEmit(a.ctx, EventHeroesLayoutDataChanged)
			runtime.EventsEmit(a.ctx, EventSteamAccountsChanged)
		}
	}()

	// Start periodic app update check
	go func() {
		// Check for updates on startup after a short delay
		slog.Info("Checking for updates on startup")
		if err := a.updateService.CheckForUpdate(); err != nil {
			slog.Warn("Error checking for updates on startup", "error", err)
		}

		runtime.EventsEmit(a.ctx, EventAppUpdateDataChanged)

		// Periodic check loop
		for {
			select {
			case <-a.ctx.Done():
				slog.Info("Stopping background app update check task")
				return
			case <-time.After(1 * time.Hour):
				// continue
			}

			slog.Info("Checking for app updates after timeout")
			if err := a.updateService.CheckForUpdate(); err != nil {
				slog.Warn("Error checking for updates after timeout", "error", err)
			}
			runtime.EventsEmit(a.ctx, EventAppUpdateDataChanged)
		}
	}()
}
