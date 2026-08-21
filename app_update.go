package main

import (
	"context"
	"log/slog"
	"time"

	"d2tool/update"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type appUpdater interface {
	Run(context.Context)
	State() AppUpdateState
	SetAutoUpdateEnabled(bool)
}

type autoUpdateSettings interface {
	GetAutoUpdateEnabled() bool
	SetAutoUpdateEnabled(bool)
}

type tickerFactory func(time.Duration) (<-chan time.Time, func())

type appUpdateNotifier func(context.Context)

type automaticAppUpdater struct {
	settings       autoUpdateSettings
	service        update.UpdateService
	interval       time.Duration
	newTicker      tickerFactory
	wake           chan struct{}
	settingChanges chan struct{}
	notifyChanged  appUpdateNotifier
}

func newAutomaticAppUpdater(
	settings autoUpdateSettings,
	service update.UpdateService,
	interval time.Duration,
	factory tickerFactory,
	notify appUpdateNotifier,
) *automaticAppUpdater {
	if factory == nil {
		factory = func(interval time.Duration) (<-chan time.Time, func()) {
			ticker := time.NewTicker(interval)
			return ticker.C, ticker.Stop
		}
	}
	if notify == nil {
		notify = func(context.Context) {}
	}
	return &automaticAppUpdater{
		settings: settings, service: service, interval: interval,
		newTicker: factory, wake: make(chan struct{}, 1),
		settingChanges: make(chan struct{}, 1), notifyChanged: notify,
	}
}

func (u *automaticAppUpdater) wakeReconciliation() {
	select {
	case u.wake <- struct{}{}:
	default:
	}
}

func (u *automaticAppUpdater) signalSettingChanged() {
	select {
	case u.settingChanges <- struct{}{}:
	default:
	}
}

func (u *automaticAppUpdater) Run(ctx context.Context) {
	go u.forwardStateChanges(ctx)
	if err := u.service.Cleanup(); err != nil {
		slog.Warn("Unable to clean stale update files", "error", err)
	}
	u.reconcile(ctx)
	ticks, stop := u.newTicker(u.interval)
	defer stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			u.reconcile(ctx)
		case <-u.wake:
			u.reconcile(ctx)
		}
	}
}

func (u *automaticAppUpdater) reconcile(ctx context.Context) {
	if err := u.service.CheckForUpdate(ctx); err != nil {
		slog.Warn("Unable to check for application updates", "error", err)
		return
	}
	state := u.service.GetState()
	if !u.settings.GetAutoUpdateEnabled() || !state.UpdateAvailable || state.PreparedVersion != "" || state.RecoveryRequired {
		return
	}
	if err := u.service.PrepareUpdate(ctx); err != nil {
		slog.Warn("Unable to prepare application update", "error", err)
	}
}

func (u *automaticAppUpdater) forwardStateChanges(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-u.service.StateChanges():
			u.notifyChanged(ctx)
		case <-u.settingChanges:
			u.notifyChanged(ctx)
		}
	}
}

type AppUpdateState struct {
	CurrentVersion    string `json:"currentVersion"`
	LatestVersion     string `json:"latestVersion"`
	UpdateAvailable   bool   `json:"updateAvailable"`
	AutoUpdateEnabled bool   `json:"autoUpdateEnabled"`
	Status            string `json:"status"`
	PreparedVersion   string `json:"preparedVersion"`
	ErrorMessage      string `json:"errorMessage"`
}

func (u *automaticAppUpdater) State() AppUpdateState {
	state := u.service.GetState()
	return AppUpdateState{
		CurrentVersion: state.CurrentAppVersion, LatestVersion: state.LatestAppVersion,
		UpdateAvailable: state.UpdateAvailable, AutoUpdateEnabled: u.settings.GetAutoUpdateEnabled(),
		Status: string(state.Status), PreparedVersion: state.PreparedVersion, ErrorMessage: state.ErrorMessage,
	}
}

func (u *automaticAppUpdater) SetAutoUpdateEnabled(enabled bool) {
	u.settings.SetAutoUpdateEnabled(enabled)
	u.signalSettingChanged()
	if enabled {
		u.wakeReconciliation()
	}
}

func (a *App) GetAppUpdateState() AppUpdateState {
	return a.appUpdater.State()
}

func (a *App) SetAutoUpdateEnabled(enabled bool) {
	a.appUpdater.SetAutoUpdateEnabled(enabled)
}

func emitAppUpdateChanged(ctx context.Context) {
	if ctx != nil {
		runtime.EventsEmit(ctx, EventAppUpdateDataChanged)
	}
}
