package main

import (
	"d2tool/config"
	"d2tool/github"
	"d2tool/heroesLayout"
	"d2tool/startup"
	"d2tool/update"
	"log/slog"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	updateService        *update.UpdateService
	updateServiceOnce    sync.Once
	forceAppUpdateChan   chan struct{}
	forceAppUpdateChanMu sync.Mutex

	isUpdatingLayout     bool
	isUpdatingLayoutLock sync.Mutex
)

// AppUpdateState represents the state for the app update tab
type AppUpdateState struct {
	CurrentVersion      string `json:"currentVersion"`
	LatestVersion       string `json:"latestVersion"`
	LastCheckTime       string `json:"lastCheckTime"`
	UpdateAvailable     bool   `json:"updateAvailable"`
	IsCheckingForUpdate bool   `json:"isCheckingForUpdate"`
	IsDownloadingUpdate bool   `json:"isDownloadingUpdate"`
	AutoUpdateEnabled   bool   `json:"autoUpdateEnabled"`
}

// --- Heroes Layout Update ---

// GetIsUpdatingLayout returns whether an update is in progress
func (a *App) GetIsUpdatingLayout() bool {
	isUpdatingLayoutLock.Lock()
	defer isUpdatingLayoutLock.Unlock()
	return isUpdatingLayout
}

// UpdateHeroesLayout triggers the hero layout update
func (a *App) UpdateHeroesLayout() {
	go func() {
		isUpdatingLayoutLock.Lock()
		if isUpdatingLayout {
			isUpdatingLayoutLock.Unlock()
			return
		}
		isUpdatingLayout = true
		isUpdatingLayoutLock.Unlock()

		// Notify frontend that update started
		runtime.EventsEmit(a.ctx, "heroesLayoutUpdateStarted")

		// Get enabled files and positions
		enabledFilePaths := a.config.GetEnabledFilePaths()
		enabledPositions := a.config.GetEnabledPositionIDs()

		// Update each file
		now := time.Now()
		for _, filePath := range enabledFilePaths {
			err := heroesLayout.UpdateHeroesLayout(heroesLayout.UpdateHeroesLayoutConfig{
				ConfigFilePaths: []string{filePath},
				Positions:       enabledPositions,
			})

			errorMsg := ""
			if err != nil {
				errorMsg = err.Error()
				slog.Error("Error updating heroes layout", "file", filePath, "error", err)
			}

			a.config.UpdateHeroesLayoutFileStatus(filePath, now.UnixMilli(), errorMsg)
		}

		isUpdatingLayoutLock.Lock()
		isUpdatingLayout = false
		isUpdatingLayoutLock.Unlock()

		// Notify frontend that update finished with updated files
		runtime.EventsEmit(a.ctx, "heroesLayoutUpdateFinished", a.config.GetHeroesLayoutFiles())
	}()
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
func (a *App) OpenFileDialog() string {
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
		slog.Warn("Error opening file dialog", "error", err)
		return ""
	}
	return selection
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
func (a *App) GetStartupEnabled() bool {
	if !startup.SupportsStartup() {
		return false
	}
	registered, err := startup.IsStartupRegistered()
	if err != nil {
		slog.Warn("Error checking startup registration", "error", err)
		return false
	}
	return registered
}

// SetStartupEnabled enables or disables running on startup
func (a *App) SetStartupEnabled(enabled bool) error {
	if !startup.SupportsStartup() {
		return nil
	}
	if enabled {
		return startup.StartupRegister([]string{"-minimized"})
	}
	return startup.StartupRemove()
}

// IsStartupSupported returns whether startup registration is supported on this platform
func (a *App) IsStartupSupported() bool {
	return startup.SupportsStartup()
}

// --- App Update Tab Bindings ---

func (a *App) getUpdateService() *update.UpdateService {
	updateServiceOnce.Do(func() {
		updateService = update.NewUpdateService(
			github.NewHttpClient(),
			AppVersion,
		)
		// Cleanup old files on startup
		if err := updateService.CleanupOldFiles(); err != nil {
			slog.Warn("Unable to cleanup old files", "error", err)
		}
	})
	return updateService
}

// GetAppUpdateState returns the current state for the app update tab
func (a *App) GetAppUpdateState() AppUpdateState {
	svc := a.getUpdateService()

	lastCheckMillis := a.config.GetAppLastCheckTimestampMillis()
	var lastCheckTimeStr string
	if lastCheckMillis == 0 {
		lastCheckTimeStr = "Never"
	} else {
		lastCheckTimeStr = time.UnixMilli(lastCheckMillis).Format("2006-01-02 15:04:05")
	}

	return AppUpdateState{
		CurrentVersion:    AppVersion,
		LatestVersion:     svc.LatestAvailableVersion(),
		LastCheckTime:     lastCheckTimeStr,
		UpdateAvailable:   svc.UpdateAvailable(),
		AutoUpdateEnabled: a.config.GetAutoUpdateEnabled(),
	}
}

// GetAutoUpdateEnabled returns whether auto-update is enabled
func (a *App) GetAutoUpdateEnabled() bool {
	return a.config.GetAutoUpdateEnabled()
}

// SetAutoUpdateEnabled enables or disables auto-updates
func (a *App) SetAutoUpdateEnabled(enabled bool) {
	a.config.SetAutoUpdateEnabled(enabled)
}

// CheckForAppUpdate checks for application updates
func (a *App) CheckForAppUpdate() {
	forceAppUpdateChanMu.Lock()
	ch := forceAppUpdateChan
	forceAppUpdateChanMu.Unlock()

	if ch == nil {
		slog.Warn("Update service not initialized yet")
		return
	}

	// Trigger the check via the force channel (non-blocking)
	select {
	case ch <- struct{}{}:
		slog.Debug("Triggered app update check")
	default:
		slog.Debug("App update check already in progress")
	}
}

// DownloadAppUpdate downloads and installs the update
func (a *App) DownloadAppUpdate() {
	go func() {
		svc := a.getUpdateService()

		err := svc.UpdateApp()

		if err != nil {
			slog.Error("Error downloading update", "error", err)
			runtime.EventsEmit(a.ctx, "appUpdateDownloadFinished", map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		} else {
			runtime.EventsEmit(a.ctx, "appUpdateDownloadFinished", map[string]interface{}{
				"success": true,
				"error":   "",
			})
		}
	}()
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

			// Trigger update
			a.UpdateHeroesLayout()
		}
	}()

	// Start periodic app update check
	go func() {
		svc := a.getUpdateService()

		// Create and store the force update channel
		forceAppUpdateChanMu.Lock()
		forceAppUpdateChan = make(chan struct{}, 1)
		ch := forceAppUpdateChan
		forceAppUpdateChanMu.Unlock()

		// Start goroutine to listen for service events and emit Wails events
		go func() {
			for {
				select {
				case <-svc.OnCheckStarted:
					runtime.EventsEmit(a.ctx, "appUpdateCheckStarted")
				case <-svc.OnCheckFinished:
					now := time.Now()
					a.config.SetAppLastCheckTimestampMillis(now.UnixMilli())
					runtime.EventsEmit(a.ctx, "appUpdateCheckFinished", AppUpdateState{
						CurrentVersion:    AppVersion,
						LatestVersion:     svc.LatestAvailableVersion(),
						LastCheckTime:     now.Format("2006-01-02 15:04:05"),
						UpdateAvailable:   svc.UpdateAvailable(),
						AutoUpdateEnabled: a.config.GetAutoUpdateEnabled(),
					})
				case <-svc.OnUpdateStarted:
					runtime.EventsEmit(a.ctx, "appUpdateDownloadStarted")
				case <-svc.OnUpdateFinished:
					// This is handled in DownloadAppUpdate
				}
			}
		}()

		// Always check for updates on startup to show latest version
		slog.Info("Checking for updates on startup")
		go func() {
			time.Sleep(2 * time.Second)
			select {
			case ch <- struct{}{}:
				slog.Debug("Triggered startup app update check")
			default:
				slog.Debug("Startup app update check already in progress")
			}
		}()

		// Run the periodic update check loop
		svc.RunPeriodicUpdateCheck(ch)
	}()
}
