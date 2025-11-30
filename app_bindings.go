package main

import (
	"d2tool/github"
	"d2tool/heroesGrid"
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

	isUpdatingGrid     bool
	isUpdatingGridLock sync.Mutex
)

// HomeState represents the state for the home tab
type HomeState struct {
	LastUpdateTime  string `json:"lastUpdateTime"`
	LastUpdateError string `json:"lastUpdateError"`
	IsUpdating      bool   `json:"isUpdating"`
}

// AppUpdateState represents the state for the app update tab
type AppUpdateState struct {
	CurrentVersion      string `json:"currentVersion"`
	LatestVersion       string `json:"latestVersion"`
	LastCheckTime       string `json:"lastCheckTime"`
	UpdateAvailable     bool   `json:"updateAvailable"`
	IsCheckingForUpdate bool   `json:"isCheckingForUpdate"`
	IsDownloadingUpdate bool   `json:"isDownloadingUpdate"`
}

// --- Home Tab Bindings ---

// GetHomeState returns the current state for the home tab
func (a *App) GetHomeState() HomeState {
	lastUpdateMillis := a.config.GetLastUpdateTimestampMillis()
	var lastUpdateTimeStr string
	if lastUpdateMillis == 0 {
		lastUpdateTimeStr = "Never"
	} else {
		lastUpdateTimeStr = time.UnixMilli(lastUpdateMillis).Format("2006-01-02 15:04:05")
	}

	isUpdatingGridLock.Lock()
	updating := isUpdatingGrid
	isUpdatingGridLock.Unlock()

	return HomeState{
		LastUpdateTime:  lastUpdateTimeStr,
		LastUpdateError: a.config.GetLastUpdateErrorMessage(),
		IsUpdating:      updating,
	}
}

// UpdateHeroesGrid triggers the hero grid update
func (a *App) UpdateHeroesGrid() {
	go func() {
		isUpdatingGridLock.Lock()
		if isUpdatingGrid {
			isUpdatingGridLock.Unlock()
			return
		}
		isUpdatingGrid = true
		isUpdatingGridLock.Unlock()

		// Notify frontend that update started
		runtime.EventsEmit(a.ctx, "heroesGridUpdateStarted")

		err := heroesGrid.UpdateHeroesGrid(heroesGrid.UpdateHeroGridConfig{
			ConfigFilePaths: a.config.GetHeroesGridFilePaths(),
			Positions:       a.config.GetPositionsOrder(),
		})

		now := time.Now()
		a.config.SetLastUpdateTimestampMillis(now.UnixMilli())

		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
			slog.Error("Error updating heroes grid", "error", err)
		}
		a.config.SetLastUpdateErrorMessage(errorMsg)

		isUpdatingGridLock.Lock()
		isUpdatingGrid = false
		isUpdatingGridLock.Unlock()

		// Notify frontend that update finished
		runtime.EventsEmit(a.ctx, "heroesGridUpdateFinished", HomeState{
			LastUpdateTime:  now.Format("2006-01-02 15:04:05"),
			LastUpdateError: errorMsg,
			IsUpdating:      false,
		})
	}()
}

// --- Grid Configs Tab Bindings ---

// GetGridConfigPaths returns the list of hero grid config file paths
func (a *App) GetGridConfigPaths() []string {
	return a.config.GetHeroesGridFilePaths()
}

// AddGridConfigPath adds a new config path
func (a *App) AddGridConfigPath(path string) {
	paths := a.config.GetHeroesGridFilePaths()
	// Check for duplicates
	for _, p := range paths {
		if p == path {
			return
		}
	}
	paths = append(paths, path)
	a.config.SetHeroesGridFilePaths(paths)
}

// RemoveGridConfigPath removes a config path by index
func (a *App) RemoveGridConfigPath(index int) {
	paths := a.config.GetHeroesGridFilePaths()
	if index < 0 || index >= len(paths) {
		return
	}
	paths = append(paths[:index], paths[index+1:]...)
	a.config.SetHeroesGridFilePaths(paths)
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

// --- Positions Order Tab Bindings ---

// GetPositionsOrder returns the current positions order
func (a *App) GetPositionsOrder() []string {
	return a.config.GetPositionsOrder()
}

// SetPositionsOrder updates the positions order
func (a *App) SetPositionsOrder(positions []string) {
	a.config.SetPositionsOrder(positions)
}

// MovePositionUp moves a position up in the list
func (a *App) MovePositionUp(index int) []string {
	positions := a.config.GetPositionsOrder()
	if index <= 0 || index >= len(positions) {
		return positions
	}
	positions[index], positions[index-1] = positions[index-1], positions[index]
	a.config.SetPositionsOrder(positions)
	return positions
}

// MovePositionDown moves a position down in the list
func (a *App) MovePositionDown(index int) []string {
	positions := a.config.GetPositionsOrder()
	if index < 0 || index >= len(positions)-1 {
		return positions
	}
	positions[index], positions[index+1] = positions[index+1], positions[index]
	a.config.SetPositionsOrder(positions)
	return positions
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
		return startup.StartupRegister([]string{})
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

	lastCheckMillis := a.config.GetAppLastUpdateCheckTimestampMillis()
	var lastCheckTimeStr string
	if lastCheckMillis == 0 {
		lastCheckTimeStr = "Never"
	} else {
		lastCheckTimeStr = time.UnixMilli(lastCheckMillis).Format("2006-01-02 15:04:05")
	}

	return AppUpdateState{
		CurrentVersion:  AppVersion,
		LatestVersion:   svc.LatestAvailableVersion(),
		LastCheckTime:   lastCheckTimeStr,
		UpdateAvailable: svc.UpdateAvailable(),
	}
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

		// Note: appUpdateDownloadStarted is emitted by the background listener
		// when it receives OnUpdateStarted from the service

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
	// Start periodic hero grid update
	go func() {
		delay := time.Hour
		lastUpdateTime := time.UnixMilli(a.config.GetLastUpdateTimestampMillis())

		for {
			nextUpdate := lastUpdateTime.Add(delay)
			waitDuration := time.Until(nextUpdate)
			if waitDuration < 0 {
				waitDuration = 0
			}

			<-time.After(waitDuration)

			// Trigger update
			a.UpdateHeroesGrid()
			lastUpdateTime = time.Now()
		}
	}()

	// Start periodic app update check
	go func() {
		svc := a.getUpdateService()

		// Create and store the force update channel
		forceAppUpdateChanMu.Lock()
		forceAppUpdateChan = make(chan struct{}, 1) // Buffered to prevent blocking
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
					a.config.SetAppLastUpdateCheckTimestampMillis(now.UnixMilli())
					runtime.EventsEmit(a.ctx, "appUpdateCheckFinished", AppUpdateState{
						CurrentVersion:  AppVersion,
						LatestVersion:   svc.LatestAvailableVersion(),
						LastCheckTime:   now.Format("2006-01-02 15:04:05"),
						UpdateAvailable: svc.UpdateAvailable(),
					})
				case <-svc.OnUpdateStarted:
					runtime.EventsEmit(a.ctx, "appUpdateDownloadStarted")
				case <-svc.OnUpdateFinished:
					// This is handled in DownloadAppUpdate
				}
			}
		}()

		// Run the periodic update check loop
		svc.RunPeriodicUpdateCheck(ch)
	}()
}
