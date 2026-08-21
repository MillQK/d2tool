package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"d2tool/config"
	"d2tool/update"
)

type fakeAutoUpdateSettings struct {
	mu      sync.Mutex
	enabled bool
}

func (f *fakeAutoUpdateSettings) GetAutoUpdateEnabled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.enabled
}

func (f *fakeAutoUpdateSettings) SetAutoUpdateEnabled(enabled bool) {
	f.mu.Lock()
	f.enabled = enabled
	f.mu.Unlock()
}

type fakeAppUpdateService struct {
	mu            sync.Mutex
	state         update.UpdateState
	changes       chan struct{}
	checkCalls    int
	prepareCalls  int
	cleanupCalls  int
	checkErr      error
	prepareErr    error
	checkSignal   chan struct{}
	stateSignal   chan struct{}
	prepareSignal chan struct{}
	prepareFinish <-chan struct{}
}

func newFakeAppUpdateService() *fakeAppUpdateService {
	return &fakeAppUpdateService{
		changes:       make(chan struct{}, 8),
		checkSignal:   make(chan struct{}, 8),
		stateSignal:   make(chan struct{}, 8),
		prepareSignal: make(chan struct{}, 8),
	}
}

func (f *fakeAppUpdateService) GetState() update.UpdateState {
	f.mu.Lock()
	state := f.state
	f.mu.Unlock()
	f.stateSignal <- struct{}{}
	return state
}
func (f *fakeAppUpdateService) StateChanges() <-chan struct{} { return f.changes }
func (f *fakeAppUpdateService) CheckForUpdate(context.Context) error {
	f.mu.Lock()
	f.checkCalls++
	err := f.checkErr
	f.mu.Unlock()
	f.checkSignal <- struct{}{}
	return err
}
func (f *fakeAppUpdateService) PrepareUpdate(context.Context) error {
	f.mu.Lock()
	f.prepareCalls++
	err := f.prepareErr
	finish := f.prepareFinish
	f.mu.Unlock()
	f.prepareSignal <- struct{}{}
	if finish != nil {
		<-finish
	}
	return err
}
func (f *fakeAppUpdateService) Cleanup() error {
	f.mu.Lock()
	f.cleanupCalls++
	f.mu.Unlock()
	return nil
}

func (f *fakeAppUpdateService) setUpdateAvailable(available bool) {
	f.mu.Lock()
	f.state.UpdateAvailable = available
	f.mu.Unlock()
}

func (f *fakeAppUpdateService) calls() (checks, preparations, cleanups int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.checkCalls, f.prepareCalls, f.cleanupCalls
}

func TestAutomaticAppUpdater_ChecksWhenDisabledWithoutPreparing(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: false}
	service := newFakeAppUpdateService()
	service.setUpdateAvailable(true)
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	updater.reconcile(context.Background())
	checks, preparations, _ := service.calls()
	if checks != 1 || preparations != 0 {
		t.Fatalf("check=%d prepare=%d", checks, preparations)
	}
}

func TestAutomaticAppUpdater_PreparesWhenEnabledAndAvailable(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.setUpdateAvailable(true)
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	updater.reconcile(context.Background())
	checks, preparations, _ := service.calls()
	if checks != 1 || preparations != 1 {
		t.Fatalf("check=%d prepare=%d", checks, preparations)
	}
}

func TestAutomaticAppUpdater_DoesNotPrepareTwice(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.state.UpdateAvailable = true
	service.state.PreparedVersion = "0.0.13"
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	updater.reconcile(context.Background())
	_, preparations, _ := service.calls()
	if preparations != 0 {
		t.Fatalf("prepare=%d, want 0", preparations)
	}
}

func TestAutomaticAppUpdater_CheckFailureSkipsPreparation(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.setUpdateAvailable(true)
	service.checkErr = errors.New("GitHub unavailable")
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	updater.reconcile(context.Background())
	checks, preparations, _ := service.calls()
	if checks != 1 || preparations != 0 {
		t.Fatalf("check=%d prepare=%d", checks, preparations)
	}
}

func TestAutomaticAppUpdater_PrepareFailureIsRetried(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.setUpdateAvailable(true)
	service.prepareErr = errors.New("disk full")
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	updater.reconcile(context.Background())
	updater.reconcile(context.Background())
	checks, preparations, _ := service.calls()
	if checks != 2 || preparations != 2 {
		t.Fatalf("check=%d prepare=%d, want 2 and 2", checks, preparations)
	}
}

func TestAutomaticAppUpdater_RecoveryRequiredContinuesChecksWithoutPreparing(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.state.UpdateAvailable = true
	service.state.RecoveryRequired = true
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	updater.reconcile(context.Background())
	updater.reconcile(context.Background())

	checks, preparations, _ := service.calls()
	if checks != 2 || preparations != 0 {
		t.Fatalf("check=%d prepare=%d, want 2 and 0", checks, preparations)
	}
}

func TestAutomaticAppUpdater_RunReactsToStartupTickAndEnablement(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: false}
	service := newFakeAppUpdateService()
	ticks := make(chan time.Time)
	tickerFactory := func(time.Duration) (<-chan time.Time, func()) { return ticks, func() {} }
	updater := newAutomaticAppUpdater(settings, service, time.Hour, tickerFactory, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		updater.Run(ctx)
		close(done)
	}()

	<-service.checkSignal
	<-service.stateSignal
	ticks <- time.Now()
	<-service.checkSignal
	<-service.stateSignal
	service.setUpdateAvailable(true)
	updater.SetAutoUpdateEnabled(true)
	<-service.checkSignal
	<-service.stateSignal
	<-service.prepareSignal

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("updater did not stop")
	}
	_, preparations, cleanups := service.calls()
	if cleanups != 1 || preparations != 1 {
		t.Fatalf("cleanup=%d prepare=%d", cleanups, preparations)
	}
}

func TestAutomaticAppUpdater_DisablingDoesNotCancelActivePreparation(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.setUpdateAvailable(true)
	finish := make(chan struct{})
	service.prepareFinish = finish
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	done := make(chan struct{})
	go func() {
		updater.reconcile(context.Background())
		close(done)
	}()
	<-service.prepareSignal
	updater.SetAutoUpdateEnabled(false)
	select {
	case <-done:
		t.Fatal("preparation finished before the installer was released")
	default:
	}
	close(finish)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("active preparation did not finish after disabling")
	}
	_, preparations, _ := service.calls()
	if preparations != 1 {
		t.Fatalf("prepare=%d, want 1", preparations)
	}
}

func TestAutomaticAppUpdater_ForwardsServiceStateChanges(t *testing.T) {
	settings := &fakeAutoUpdateSettings{}
	service := newFakeAppUpdateService()
	notified := make(chan struct{}, 1)
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, func(context.Context) { notified <- struct{}{} })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go updater.forwardStateChanges(ctx)
	service.changes <- struct{}{}
	select {
	case <-notified:
	case <-time.After(time.Second):
		t.Fatal("state change was not forwarded")
	}
}

func TestAutomaticAppUpdater_StateCombinesServiceAndSettings(t *testing.T) {
	settings := &fakeAutoUpdateSettings{enabled: true}
	service := newFakeAppUpdateService()
	service.state = update.UpdateState{
		CurrentAppVersion: "0.0.12",
		LatestAppVersion:  "0.0.13",
		UpdateAvailable:   true,
		Status:            update.UpdateStatusReady,
		PreparedVersion:   "0.0.13",
		ErrorMessage:      "visible error",
	}
	updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, nil)

	got := updater.State()
	want := AppUpdateState{
		CurrentVersion: "0.0.12", LatestVersion: "0.0.13", UpdateAvailable: true,
		AutoUpdateEnabled: true, Status: "ready", PreparedVersion: "0.0.13",
		ErrorMessage: "visible error",
	}
	if got != want {
		t.Fatalf("State() = %#v, want %#v", got, want)
	}
}

func TestAutomaticAppUpdater_SetAutoUpdateEnabledPersistsNotifiesAndWakesOnlyWhenEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "enabled", enabled: true},
		{name: "disabled", enabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &fakeAutoUpdateSettings{enabled: !tt.enabled}
			service := newFakeAppUpdateService()
			notified := make(chan struct{}, 1)
			updater := newAutomaticAppUpdater(settings, service, time.Hour, nil, func(context.Context) {
				notified <- struct{}{}
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go updater.forwardStateChanges(ctx)

			updater.SetAutoUpdateEnabled(tt.enabled)

			if got := settings.GetAutoUpdateEnabled(); got != tt.enabled {
				t.Fatalf("GetAutoUpdateEnabled() = %v, want %v", got, tt.enabled)
			}
			select {
			case <-notified:
			case <-time.After(time.Second):
				t.Fatal("setting change was not forwarded")
			}
			select {
			case <-updater.wake:
				if !tt.enabled {
					t.Fatal("disabling automatic updates woke reconciliation")
				}
			default:
				if tt.enabled {
					t.Fatal("enabling automatic updates did not wake reconciliation")
				}
			}
		})
	}
}

type fakeInjectedAppUpdater struct {
	state             AppUpdateState
	runContexts       chan context.Context
	cancelled         chan struct{}
	finishAfterCancel <-chan struct{}
}

func (f *fakeInjectedAppUpdater) Run(ctx context.Context) {
	if f.runContexts != nil {
		f.runContexts <- ctx
	}
	if f.finishAfterCancel != nil {
		<-ctx.Done()
		if f.cancelled != nil {
			close(f.cancelled)
		}
		<-f.finishAfterCancel
	}
}
func (f *fakeInjectedAppUpdater) State() AppUpdateState { return f.state }
func (f *fakeInjectedAppUpdater) SetAutoUpdateEnabled(enabled bool) {
	f.state.AutoUpdateEnabled = enabled
}

func TestAppUsesInjectedUpdaterForUpdateBindings(t *testing.T) {
	updater := &fakeInjectedAppUpdater{state: AppUpdateState{CurrentVersion: "0.0.12"}}
	app := NewApp(nil, updater, nil, nil, nil)

	if got := app.GetAppUpdateState().CurrentVersion; got != "0.0.12" {
		t.Fatalf("CurrentVersion = %q, want 0.0.12", got)
	}
	app.SetAutoUpdateEnabled(true)
	if !app.GetAppUpdateState().AutoUpdateEnabled {
		t.Fatal("App did not delegate the enabled setting to its updater")
	}
}

func TestAppStartsInjectedUpdaterWithApplicationContext(t *testing.T) {
	updater := &fakeInjectedAppUpdater{runContexts: make(chan context.Context, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app := NewApp(&config.Config{}, updater, nil, nil, nil)
	app.ctx = ctx

	app.startBackgroundTasks()

	select {
	case got := <-updater.runContexts:
		if got != ctx {
			t.Fatal("updater received a different application context")
		}
	case <-time.After(time.Second):
		t.Fatal("injected updater was not started")
	}
}

func TestAppShutdownCancelsAndJoinsUpdater(t *testing.T) {
	preserveTestConfigFile(t)
	releaseUpdater := make(chan struct{})
	updaterCancelled := make(chan struct{})
	updater := &fakeInjectedAppUpdater{
		runContexts:       make(chan context.Context, 1),
		cancelled:         updaterCancelled,
		finishAfterCancel: releaseUpdater,
	}
	parentCtx, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	app := NewApp(&config.Config{}, updater, nil, nil, nil)
	app.startup(parentCtx)
	<-updater.runContexts

	stopped := make(chan struct{})
	go func() {
		app.shutdown(context.Background())
		close(stopped)
	}()

	select {
	case <-updaterCancelled:
	case <-time.After(time.Second):
		close(releaseUpdater)
		t.Fatal("stopping background tasks did not cancel the updater context")
	}
	select {
	case <-stopped:
		close(releaseUpdater)
		t.Fatal("background shutdown returned before the updater finished")
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseUpdater)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("background shutdown did not return after the updater finished")
	}
}

func preserveTestConfigFile(t *testing.T) {
	t.Helper()
	executablePath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(filepath.Dir(executablePath), "d2tool_config.json")
	contents, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		t.Cleanup(func() {
			if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
				t.Errorf("remove test config: %v", err)
			}
		})
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(configPath, contents, info.Mode().Perm()); err != nil {
			t.Errorf("restore test config: %v", err)
		}
	})
}
