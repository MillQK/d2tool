package update

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"d2tool/github"
)

type fakeGitHubClient struct {
	mu      sync.Mutex
	release *github.Release
	err     error
	calls   int
}

func (f *fakeGitHubClient) GetLatestRelease(context.Context) (*github.Release, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.release, f.err
}

type fakeReleaseInstaller struct {
	started     chan struct{}
	finish      chan struct{}
	err         error
	startedOnce sync.Once
	mu          sync.Mutex
	calls       int
}

func (f *fakeReleaseInstaller) Install(context.Context, *github.Release) error {
	f.mu.Lock()
	f.calls++
	err := f.err
	f.mu.Unlock()
	if f.started != nil {
		f.startedOnce.Do(func() { close(f.started) })
	}
	if f.finish != nil {
		<-f.finish
	}
	return err
}

func (f *fakeReleaseInstaller) Cleanup() error { return nil }

type testRecoveryRequiredError struct {
	message string
}

func (e *testRecoveryRequiredError) Error() string     { return e.message }
func (e *testRecoveryRequiredError) recoveryRequired() {}

func TestUpdateService_CheckForUpdateStoresNewerRelease(t *testing.T) {
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	service := newUpdateService("0.0.12", client, &fakeReleaseInstaller{}, func() time.Time { return now })

	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := service.GetState()
	if state.LatestAppVersion != "0.0.13" || !state.UpdateAvailable || !state.LastCheckTime.Equal(now) {
		t.Fatalf("unexpected state: %+v", state)
	}
	if state.Status != UpdateStatusIdle || state.ErrorMessage != "" {
		t.Fatalf("unexpected status: %+v", state)
	}
}

func TestUpdateService_CheckErrorPreservesLastKnownRelease(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	service := newUpdateService("0.0.12", client, &fakeReleaseInstaller{}, time.Now)
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	client.mu.Lock()
	client.err = errors.New("GitHub unavailable")
	client.mu.Unlock()
	if err := service.CheckForUpdate(context.Background()); err == nil {
		t.Fatal("check must fail")
	}
	state := service.GetState()
	if state.LatestAppVersion != "0.0.13" || state.Status != UpdateStatusError {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestUpdateService_CheckForUpdateRejectsNilAndMalformedReleases(t *testing.T) {
	tests := []struct {
		name    string
		release *github.Release
	}{
		{name: "nil release"},
		{name: "malformed release name", release: &github.Release{Name: "Release 2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newUpdateService(
				"0.0.12",
				&fakeGitHubClient{release: tt.release},
				&fakeReleaseInstaller{},
				time.Now,
			)

			if err := service.CheckForUpdate(context.Background()); err == nil {
				t.Fatal("CheckForUpdate() must reject invalid release metadata")
			}
			state := service.GetState()
			if state.Status != UpdateStatusError || state.ErrorMessage == "" {
				t.Fatalf("invalid release did not produce visible error state: %+v", state)
			}
			if state.LatestAppVersion != "" || state.UpdateAvailable {
				t.Fatalf("invalid release was published: %+v", state)
			}
		})
	}
}

func TestUpdateService_PrepareUpdateTransitionsUpdatingToReady(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	installer := &fakeReleaseInstaller{started: make(chan struct{}), finish: make(chan struct{})}
	service := newUpdateService("0.0.12", client, installer, time.Now)
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- service.PrepareUpdate(context.Background()) }()
	<-installer.started
	if state := service.GetState(); state.Status != UpdateStatusUpdating {
		t.Fatalf("status while blocked = %q", state.Status)
	}
	close(installer.finish)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	state := service.GetState()
	if state.Status != UpdateStatusReady || state.PreparedVersion != "0.0.13" || state.ErrorMessage != "" {
		t.Fatalf("unexpected ready state: %+v", state)
	}

	if err := service.PrepareUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	installer.mu.Lock()
	defer installer.mu.Unlock()
	if installer.calls != 1 {
		t.Fatalf("Install calls = %d, want 1", installer.calls)
	}
}

func TestUpdateService_ConcurrentPrepareUpdateInstallsOnce(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	installer := &fakeReleaseInstaller{started: make(chan struct{}), finish: make(chan struct{})}
	service := newUpdateService("0.0.12", client, installer, time.Now)
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}

	firstDone := make(chan error, 1)
	go func() { firstDone <- service.PrepareUpdate(context.Background()) }()
	<-installer.started

	secondEntered := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		close(secondEntered)
		secondDone <- service.PrepareUpdate(context.Background())
	}()
	<-secondEntered
	time.Sleep(100 * time.Millisecond)
	close(installer.finish)

	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	if err := <-secondDone; err != nil {
		t.Fatal(err)
	}
	installer.mu.Lock()
	defer installer.mu.Unlock()
	if installer.calls != 1 {
		t.Fatalf("Install calls = %d, want 1", installer.calls)
	}
}

func TestUpdateService_PrepareFailureBecomesRetryableError(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	installer := &fakeReleaseInstaller{err: errors.New("disk full")}
	service := newUpdateService("0.0.12", client, installer, time.Now)
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.PrepareUpdate(context.Background()); err == nil {
		t.Fatal("prepare must fail")
	}
	state := service.GetState()
	if state.Status != UpdateStatusError || state.PreparedVersion != "" || state.ErrorMessage == "" {
		t.Fatalf("unexpected error state: %+v", state)
	}

	installer.mu.Lock()
	installer.err = nil
	installer.mu.Unlock()
	if err := service.PrepareUpdate(context.Background()); err != nil {
		t.Fatalf("retry PrepareUpdate() error = %v", err)
	}
	state = service.GetState()
	if state.Status != UpdateStatusReady || state.PreparedVersion != "0.0.13" || state.ErrorMessage != "" {
		t.Fatalf("unexpected retry state: %+v", state)
	}
	installer.mu.Lock()
	defer installer.mu.Unlock()
	if installer.calls != 2 {
		t.Fatalf("Install calls = %d, want 2", installer.calls)
	}
}

func TestUpdateService_RecoveryRequiredFailureSurvivesChecksAndIsNotRetried(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	installer := &fakeReleaseInstaller{err: &testRecoveryRequiredError{
		message: "restore backup /install/.old.d2tool-d2tool.1 to /install/d2tool failed",
	}}
	service := newUpdateService("0.0.12", client, installer, time.Now)
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.PrepareUpdate(context.Background()); err == nil {
		t.Fatal("prepare must fail")
	}
	recoveryState := service.GetState()
	recoveryMessage := recoveryState.ErrorMessage
	if !strings.Contains(recoveryMessage, "/install/.old.d2tool-d2tool.1") {
		t.Fatalf("recovery paths missing from error: %q", recoveryMessage)
	}
	if !recoveryState.RecoveryRequired {
		t.Fatalf("recovery-required failure was not represented in state: %+v", recoveryState)
	}

	client.mu.Lock()
	client.err = errors.New("GitHub unavailable")
	client.mu.Unlock()
	if err := service.CheckForUpdate(context.Background()); err == nil {
		t.Fatal("check must fail")
	}
	stateAfterFailedCheck := service.GetState()
	if !stateAfterFailedCheck.RecoveryRequired || stateAfterFailedCheck.ErrorMessage != recoveryMessage {
		t.Fatalf("failed check overwrote recovery state: %+v", stateAfterFailedCheck)
	}

	client.mu.Lock()
	client.release = &github.Release{Name: "0.0.14"}
	client.err = nil
	client.mu.Unlock()
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	stateAfterSuccessfulCheck := service.GetState()
	if !stateAfterSuccessfulCheck.RecoveryRequired || stateAfterSuccessfulCheck.ErrorMessage != recoveryMessage || stateAfterSuccessfulCheck.LatestAppVersion != "0.0.14" {
		t.Fatalf("successful check did not refresh release while preserving recovery state: %+v", stateAfterSuccessfulCheck)
	}

	installer.mu.Lock()
	installer.err = errors.New("canonical executable is missing")
	installer.mu.Unlock()
	if err := service.PrepareUpdate(context.Background()); err == nil {
		t.Fatal("blocked retry must still report an error")
	}

	state := service.GetState()
	if state.Status != UpdateStatusError || !state.RecoveryRequired || state.PreparedVersion != "" || state.ErrorMessage != recoveryMessage {
		t.Fatalf("recovery state was overwritten: %+v", state)
	}
	installer.mu.Lock()
	defer installer.mu.Unlock()
	if installer.calls != 1 {
		t.Fatalf("Install calls = %d, want 1", installer.calls)
	}
}

func TestUpdateService_PrepareRejectsMissingOrNonNewRelease(t *testing.T) {
	tests := []struct {
		name    string
		release *github.Release
	}{
		{name: "missing release"},
		{name: "equal release", release: &github.Release{Name: "0.0.12"}},
		{name: "older release", release: &github.Release{Name: "0.0.11"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeGitHubClient{release: tt.release}
			installer := &fakeReleaseInstaller{}
			service := newUpdateService("0.0.12", client, installer, time.Now)
			if tt.release != nil {
				if err := service.CheckForUpdate(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			if err := service.PrepareUpdate(context.Background()); err == nil {
				t.Fatal("PrepareUpdate() must reject a missing or non-new release")
			}
			installer.mu.Lock()
			defer installer.mu.Unlock()
			if installer.calls != 0 {
				t.Fatalf("Install calls = %d, want 0", installer.calls)
			}
		})
	}
}

func requireStateChange(t *testing.T, changes <-chan struct{}) {
	t.Helper()
	select {
	case <-changes:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for state change")
	}
}

func drainStateChanges(changes <-chan struct{}) {
	for {
		select {
		case <-changes:
		default:
			return
		}
	}
}

func TestUpdateService_StateChangesReportsCheckUpdatingAndReady(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	installer := &fakeReleaseInstaller{started: make(chan struct{}), finish: make(chan struct{})}
	service := newUpdateService("0.0.12", client, installer, time.Now)

	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	requireStateChange(t, service.StateChanges())

	done := make(chan error, 1)
	go func() { done <- service.PrepareUpdate(context.Background()) }()
	<-installer.started
	requireStateChange(t, service.StateChanges())
	if state := service.GetState(); state.Status != UpdateStatusUpdating {
		t.Fatalf("state after updating notification: %+v", state)
	}

	close(installer.finish)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	requireStateChange(t, service.StateChanges())
	if state := service.GetState(); state.Status != UpdateStatusReady {
		t.Fatalf("state after ready notification: %+v", state)
	}
}

func TestUpdateService_ReadyTakesPrecedenceOverLaterCheckError(t *testing.T) {
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	service := newUpdateService("0.0.12", client, &fakeReleaseInstaller{}, time.Now)
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.PrepareUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	drainStateChanges(service.StateChanges())

	client.mu.Lock()
	client.err = errors.New("GitHub unavailable")
	client.mu.Unlock()
	if err := service.CheckForUpdate(context.Background()); err == nil {
		t.Fatal("check must fail")
	}
	state := service.GetState()
	if state.Status != UpdateStatusReady || state.PreparedVersion != "0.0.13" || state.ErrorMessage != "" {
		t.Fatalf("later check error replaced ready state: %+v", state)
	}
	select {
	case <-service.StateChanges():
		t.Fatal("unchanged ready state must not emit a state change")
	default:
	}
}

func TestUpdateService_PrepareErrorSurvivesFailedAndSuccessfulChecks(t *testing.T) {
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	nextCheckTime := now.Add(time.Hour)
	client := &fakeGitHubClient{release: &github.Release{Name: "0.0.13"}}
	installer := &fakeReleaseInstaller{err: errors.New("disk full")}
	service := newUpdateService("0.0.12", client, installer, func() time.Time { return now })
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.PrepareUpdate(context.Background()); err == nil {
		t.Fatal("prepare must fail")
	}
	prepareError := service.GetState().ErrorMessage
	drainStateChanges(service.StateChanges())

	client.mu.Lock()
	client.err = errors.New("GitHub unavailable")
	client.mu.Unlock()
	if err := service.CheckForUpdate(context.Background()); err == nil {
		t.Fatal("check must fail")
	}
	state := service.GetState()
	if state.Status != UpdateStatusError || state.ErrorMessage != prepareError {
		t.Fatalf("failed check replaced preparation error: %+v", state)
	}
	select {
	case <-service.StateChanges():
		t.Fatal("failed check with preserved preparation error must not emit a state change")
	default:
	}

	client.mu.Lock()
	client.release = &github.Release{Name: "0.0.14"}
	client.err = nil
	client.mu.Unlock()
	now = nextCheckTime
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	requireStateChange(t, service.StateChanges())
	state = service.GetState()
	if state.Status != UpdateStatusError || state.ErrorMessage != prepareError || state.LatestAppVersion != "0.0.14" || !state.LastCheckTime.Equal(nextCheckTime) {
		t.Fatalf("successful check did not refresh state while preserving preparation error: %+v", state)
	}
}

func TestUpdateService_SuccessfulCheckClearsCheckError(t *testing.T) {
	now := time.Date(2026, 8, 20, 13, 0, 0, 0, time.UTC)
	client := &fakeGitHubClient{err: errors.New("GitHub unavailable")}
	service := newUpdateService("0.0.12", client, &fakeReleaseInstaller{}, func() time.Time { return now })
	if err := service.CheckForUpdate(context.Background()); err == nil {
		t.Fatal("check must fail")
	}
	if state := service.GetState(); state.Status != UpdateStatusError || state.ErrorMessage == "" {
		t.Fatalf("unexpected check error state: %+v", state)
	}
	drainStateChanges(service.StateChanges())

	client.mu.Lock()
	client.release = &github.Release{Name: "0.0.13"}
	client.err = nil
	client.mu.Unlock()
	if err := service.CheckForUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	requireStateChange(t, service.StateChanges())
	state := service.GetState()
	if state.Status != UpdateStatusIdle || state.ErrorMessage != "" || state.LatestAppVersion != "0.0.13" || !state.LastCheckTime.Equal(now) {
		t.Fatalf("successful check did not clear check error: %+v", state)
	}
}
