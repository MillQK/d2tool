package update

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"d2tool/github"
)

type UpdateStatus string

const (
	UpdateStatusIdle     UpdateStatus = "idle"
	UpdateStatusUpdating UpdateStatus = "updating"
	UpdateStatusReady    UpdateStatus = "ready"
	UpdateStatusError    UpdateStatus = "error"
)

type updateErrorKind string

const (
	updateErrorCheck            updateErrorKind = "check"
	updateErrorPrepare          updateErrorKind = "prepare"
	updateErrorRecoveryRequired updateErrorKind = "recovery-required"
)

type UpdateState struct {
	UpdateAvailable   bool
	CurrentAppVersion string
	LatestAppVersion  string
	LastCheckTime     time.Time
	Status            UpdateStatus
	PreparedVersion   string
	ErrorMessage      string
	RecoveryRequired  bool
}

type UpdateService interface {
	GetState() UpdateState
	StateChanges() <-chan struct{}
	CheckForUpdate(ctx context.Context) error
	PrepareUpdate(ctx context.Context) error
	Cleanup() error
}

type UpdateServiceImpl struct {
	stateLock sync.RWMutex // protects reads/writes of state fields (never held during I/O)
	opLock    sync.Mutex   // serializes CheckForUpdate / PrepareUpdate (held during I/O)

	currentAppVersion string
	githubClient      github.Client
	installer         releaseInstaller
	now               func() time.Time
	stateChanges      chan struct{}

	latestRelease   *github.Release
	lastCheckTime   time.Time
	status          UpdateStatus
	preparedVersion string
	errorMessage    string
	errorKind       updateErrorKind
}

func NewUpdateService(
	currentAppVersion string,
	githubClient github.Client,
) *UpdateServiceImpl {
	downloadClient := &http.Client{Timeout: 10 * time.Minute}
	return newUpdateService(
		currentAppVersion,
		githubClient,
		newArchiveInstaller(downloadClient),
		time.Now,
	)
}

func newUpdateService(
	currentAppVersion string,
	githubClient github.Client,
	installer releaseInstaller,
	now func() time.Time,
) *UpdateServiceImpl {
	return &UpdateServiceImpl{
		currentAppVersion: currentAppVersion,
		githubClient:      githubClient,
		installer:         installer,
		now:               now,
		stateChanges:      make(chan struct{}, 8),
		status:            UpdateStatusIdle,
		lastCheckTime:     time.UnixMilli(0),
	}
}

func (s *UpdateServiceImpl) Cleanup() error {
	return s.installer.Cleanup()
}

func (s *UpdateServiceImpl) notifyStateChanged() {
	select {
	case s.stateChanges <- struct{}{}:
	default:
	}
}

func (s *UpdateServiceImpl) StateChanges() <-chan struct{} {
	return s.stateChanges
}

func (s *UpdateServiceImpl) GetState() UpdateState {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	latestVersion := s.latestAvailableVersionLocked()

	return UpdateState{
		UpdateAvailable:   isUpdateAvailable(latestVersion, s.currentAppVersion),
		CurrentAppVersion: s.currentAppVersion,
		LatestAppVersion:  latestVersion,
		LastCheckTime:     s.lastCheckTime,
		Status:            s.status,
		PreparedVersion:   s.preparedVersion,
		ErrorMessage:      s.errorMessage,
		RecoveryRequired:  s.errorKind == updateErrorRecoveryRequired,
	}
}

func (s *UpdateServiceImpl) CheckForUpdate(ctx context.Context) error {
	s.opLock.Lock()
	defer s.opLock.Unlock()

	release, err := s.githubClient.GetLatestRelease(ctx)
	if err != nil {
		wrapped := fmt.Errorf("check latest release: %w", err)
		s.setError(updateErrorCheck, wrapped)
		return wrapped
	}
	if release == nil {
		err := fmt.Errorf("check latest release: GitHub returned no release")
		s.setError(updateErrorCheck, err)
		return err
	}
	if _, err := compareVersions(release.Name, s.currentAppVersion); err != nil {
		wrapped := fmt.Errorf("validate latest release %q: %w", release.Name, err)
		s.setError(updateErrorCheck, wrapped)
		return wrapped
	}

	s.stateLock.Lock()
	s.latestRelease = release
	s.lastCheckTime = s.now()
	if s.errorKind == updateErrorCheck {
		s.status = UpdateStatusIdle
		s.errorKind = ""
		s.errorMessage = ""
	}
	s.stateLock.Unlock()
	s.notifyStateChanged()

	return nil
}

func (s *UpdateServiceImpl) PrepareUpdate(ctx context.Context) error {
	s.opLock.Lock()
	defer s.opLock.Unlock()

	s.stateLock.RLock()
	if s.preparedVersion != "" {
		s.stateLock.RUnlock()
		return nil
	}
	if s.errorKind == updateErrorRecoveryRequired {
		s.stateLock.RUnlock()
		return fmt.Errorf("prepare update blocked: recovery required")
	}
	release := s.latestRelease
	s.stateLock.RUnlock()

	if release == nil {
		err := fmt.Errorf("prepare update: no release is available")
		s.setError(updateErrorPrepare, err)
		return err
	}
	comparison, err := compareVersions(release.Name, s.currentAppVersion)
	if err != nil {
		wrapped := fmt.Errorf("prepare update version %q: %w", release.Name, err)
		s.setError(updateErrorPrepare, wrapped)
		return wrapped
	}
	if comparison <= 0 {
		err := fmt.Errorf("prepare update: release %q is not newer than current version %q", release.Name, s.currentAppVersion)
		s.setError(updateErrorPrepare, err)
		return err
	}

	s.stateLock.Lock()
	s.status = UpdateStatusUpdating
	s.errorKind = ""
	s.errorMessage = ""
	s.stateLock.Unlock()
	s.notifyStateChanged()

	if err := s.installer.Install(ctx, release); err != nil {
		wrapped := fmt.Errorf("prepare update version %q: %w", release.Name, err)
		kind := updateErrorPrepare
		if isRecoveryRequired(err) {
			kind = updateErrorRecoveryRequired
		}
		s.setError(kind, wrapped)
		return wrapped
	}

	s.stateLock.Lock()
	s.status = UpdateStatusReady
	s.preparedVersion = release.Name
	s.errorKind = ""
	s.errorMessage = ""
	s.stateLock.Unlock()
	s.notifyStateChanged()
	return nil
}

func (s *UpdateServiceImpl) setError(kind updateErrorKind, err error) {
	changed := false
	s.stateLock.Lock()
	if s.status != UpdateStatusReady && !(kind == updateErrorCheck && (s.errorKind == updateErrorPrepare || s.errorKind == updateErrorRecoveryRequired)) {
		s.status = UpdateStatusError
		s.errorKind = kind
		s.errorMessage = err.Error()
		changed = true
	}
	s.stateLock.Unlock()
	if changed {
		s.notifyStateChanged()
	}
}

func (s *UpdateServiceImpl) latestAvailableVersionLocked() string {
	if s.latestRelease == nil {
		return ""
	}
	return s.latestRelease.Name
}
