package update

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2tool/github"
)

type fakeReplacer struct {
	replaceStaged string
	replaceTarget string
	cleanupTarget string
	replaceCalls  int
	cleanupCalls  int
	replaceResult replacementResult
	replaceErr    error
	cleanupErr    error
}

func (f *fakeReplacer) Replace(stagedPath, executablePath string) (replacementResult, error) {
	f.replaceStaged = stagedPath
	f.replaceTarget = executablePath
	f.replaceCalls++
	return f.replaceResult, f.replaceErr
}

func (f *fakeReplacer) Cleanup(executablePath string) error {
	f.cleanupTarget = executablePath
	f.cleanupCalls++
	return f.cleanupErr
}

func TestArchiveInstaller_InstallsAndOwnsTemporaryFiles(t *testing.T) {
	dir := t.TempDir()
	executablePath := filepath.Join(dir, "d2tool")
	archivePath := filepath.Join(dir, downloadFilesPrefix+"archive")
	stagedPath := filepath.Join(dir, stagedFilesPrefix+"candidate")
	if err := os.WriteFile(archivePath, []byte("archive"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stagedPath, []byte("new"), 0700); err != nil {
		t.Fatal(err)
	}

	replacer := &fakeReplacer{}
	installer := newArchiveInstallerForTest(
		func() (string, error) { return executablePath, nil },
		replacer,
		func(*github.Release, string, string) (github.ReleaseAsset, error) { return github.ReleaseAsset{}, nil },
		func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error) { return archivePath, nil },
		func(context.Context, string, string, string) (string, error) { return stagedPath, nil },
	)

	if err := installer.Install(context.Background(), &github.Release{}); err != nil {
		t.Fatal(err)
	}
	if replacer.replaceStaged != stagedPath || replacer.replaceTarget != executablePath {
		t.Fatalf("replace called with %q -> %q", replacer.replaceStaged, replacer.replaceTarget)
	}
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Fatalf("archive was not removed: %v", err)
	}
	if _, err := os.Stat(stagedPath); !os.IsNotExist(err) {
		t.Fatalf("staged file was not removed after ownership transfer: %v", err)
	}
}

func TestArchiveInstaller_CancellationAfterDownloadSkipsExtraction(t *testing.T) {
	dir := t.TempDir()
	executablePath := filepath.Join(dir, "d2tool")
	archivePath := filepath.Join(dir, downloadFilesPrefix+"archive")
	if err := os.WriteFile(archivePath, []byte("archive"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	extractCalls := 0
	replacer := &fakeReplacer{}
	installer := newArchiveInstallerForTest(
		func() (string, error) { return executablePath, nil },
		replacer,
		func(*github.Release, string, string) (github.ReleaseAsset, error) { return github.ReleaseAsset{}, nil },
		func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error) {
			cancel()
			return archivePath, nil
		},
		func(context.Context, string, string, string) (string, error) {
			extractCalls++
			return "", errors.New("unexpected extraction")
		},
	)

	err := installer.Install(ctx, &github.Release{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Install() error = %v, want context.Canceled", err)
	}
	if extractCalls != 0 || replacer.replaceCalls != 0 {
		t.Errorf("work continued after cancellation: extract=%d replace=%d", extractCalls, replacer.replaceCalls)
	}
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Errorf("archive was not removed: %v", err)
	}
}

func TestArchiveInstaller_CancellationBeforeReplacementRemovesStagedCandidate(t *testing.T) {
	dir := t.TempDir()
	executablePath := filepath.Join(dir, "d2tool")
	archivePath := filepath.Join(dir, downloadFilesPrefix+"archive")
	stagedPath := filepath.Join(dir, stagedFilesPrefix+"candidate")
	if err := os.WriteFile(archivePath, []byte("archive"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	replacer := &fakeReplacer{}
	installer := newArchiveInstallerForTest(
		func() (string, error) { return executablePath, nil },
		replacer,
		func(*github.Release, string, string) (github.ReleaseAsset, error) { return github.ReleaseAsset{}, nil },
		func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error) { return archivePath, nil },
		func(context.Context, string, string, string) (string, error) {
			if err := os.WriteFile(stagedPath, []byte("new"), 0700); err != nil {
				t.Fatal(err)
			}
			cancel()
			return stagedPath, nil
		},
	)

	err := installer.Install(ctx, &github.Release{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Install() error = %v, want context.Canceled", err)
	}
	if replacer.replaceCalls != 0 {
		t.Errorf("replacement entered after cancellation: calls=%d", replacer.replaceCalls)
	}
	if _, err := os.Stat(stagedPath); !os.IsNotExist(err) {
		t.Errorf("cancelled staged candidate was not removed: %v", err)
	}
}

func TestArchiveInstaller_RemovesStagedCandidateAfterRecoverableReplacementErrors(t *testing.T) {
	tests := []struct {
		name             string
		failedRenameCall int
	}{
		{name: "backup rename fails", failedRenameCall: 1},
		{name: "install rename fails and rollback succeeds", failedRenameCall: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			executablePath := filepath.Join(dir, "d2tool")
			archivePath := filepath.Join(dir, downloadFilesPrefix+"archive")
			stagedPath := filepath.Join(dir, stagedFilesPrefix+"candidate")
			if err := os.WriteFile(executablePath, []byte("old"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(archivePath, []byte("archive"), 0600); err != nil {
				t.Fatal(err)
			}

			realRename := os.Rename
			renameCalls := 0
			replacer := newOSExecutableReplacer()
			replacer.rename = func(oldPath, newPath string) error {
				renameCalls++
				if renameCalls == tt.failedRenameCall {
					return errors.New("injected rename failure")
				}
				return realRename(oldPath, newPath)
			}
			installer := newArchiveInstallerForTest(
				func() (string, error) { return executablePath, nil },
				replacer,
				func(*github.Release, string, string) (github.ReleaseAsset, error) { return github.ReleaseAsset{}, nil },
				func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error) { return archivePath, nil },
				func(context.Context, string, string, string) (string, error) {
					if err := os.WriteFile(stagedPath, []byte("new"), 0700); err != nil {
						t.Fatal(err)
					}
					return stagedPath, nil
				},
			)

			if err := installer.Install(context.Background(), &github.Release{}); err == nil {
				t.Fatal("Install() must report the replacement failure")
			}
			if got, err := os.ReadFile(executablePath); err != nil || string(got) != "old" {
				t.Fatalf("canonical executable = %q, error = %v", got, err)
			}
			if _, err := os.Stat(stagedPath); !os.IsNotExist(err) {
				t.Fatalf("recoverable replacement left staged candidate: %v", err)
			}
		})
	}
}

func TestArchiveInstaller_CleanupUsesInstalledExecutablePath(t *testing.T) {
	replacer := &fakeReplacer{}
	installer := newArchiveInstallerForTest(
		func() (string, error) { return "/install/d2tool", nil },
		replacer,
		nil,
		nil,
		nil,
	)
	if err := installer.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if replacer.cleanupTarget != "/install/d2tool" {
		t.Fatalf("cleanup target = %q", replacer.cleanupTarget)
	}
}

func TestArchiveInstaller_RefusesReservedExecutablePathBeforePreparation(t *testing.T) {
	for _, prefix := range []string{oldFilesPrefix, stagedFilesPrefix, downloadFilesPrefix} {
		t.Run(prefix, func(t *testing.T) {
			dir := t.TempDir()
			executablePath := filepath.Join(dir, prefix+"recovery")
			selectCalls := 0
			downloadCalls := 0
			replacer := &fakeReplacer{}
			installer := newArchiveInstallerForTest(
				func() (string, error) { return executablePath, nil },
				replacer,
				func(*github.Release, string, string) (github.ReleaseAsset, error) {
					selectCalls++
					return github.ReleaseAsset{}, nil
				},
				func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error) {
					downloadCalls++
					return "", errors.New("unexpected network work")
				},
				func(context.Context, string, string, string) (string, error) {
					return "", errors.New("unexpected extraction")
				},
			)

			if err := installer.Install(context.Background(), &github.Release{}); err == nil || !strings.Contains(err.Error(), "reserved") {
				t.Fatalf("Install() error = %v, want reserved-path rejection", err)
			}
			if selectCalls != 0 || downloadCalls != 0 || replacer.replaceCalls != 0 {
				t.Fatalf("work started for reserved path: select=%d download=%d replace=%d", selectCalls, downloadCalls, replacer.replaceCalls)
			}
		})
	}
}

func TestArchiveInstaller_CleanupRefusesReservedExecutablePath(t *testing.T) {
	for _, prefix := range []string{oldFilesPrefix, stagedFilesPrefix, downloadFilesPrefix} {
		t.Run(prefix, func(t *testing.T) {
			replacer := &fakeReplacer{}
			installer := newArchiveInstallerForTest(
				func() (string, error) { return filepath.Join("/install", prefix+"recovery"), nil },
				replacer,
				nil,
				nil,
				nil,
			)

			if err := installer.Cleanup(); err == nil || !strings.Contains(err.Error(), "reserved") {
				t.Fatalf("Cleanup() error = %v, want reserved-path rejection", err)
			}
			if replacer.cleanupCalls != 0 {
				t.Fatalf("replacer cleanup calls = %d, want 0", replacer.cleanupCalls)
			}
		})
	}
}

func TestArchiveInstaller_StopsAndCleansUpAfterErrors(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name          string
		executableErr error
		selectErr     error
		downloadErr   error
		extractErr    error
		replaceResult replacementResult
		replaceErr    error
		wantMessage   string
		wantCalls     [4]int
		wantStaged    bool
		wantRecovery  bool
	}{
		{name: "executable path", executableErr: failure, wantMessage: "get executable path", wantCalls: [4]int{0, 0, 0, 0}},
		{name: "asset selection", selectErr: failure, wantMessage: "select update asset", wantCalls: [4]int{1, 0, 0, 0}},
		{name: "download", downloadErr: failure, wantMessage: "download update", wantCalls: [4]int{1, 1, 0, 0}},
		{name: "extraction", extractErr: failure, wantMessage: "extract update", wantCalls: [4]int{1, 1, 1, 0}},
		{
			name: "recoverable replacement", replaceResult: replacementResult{outcome: replacementOutcomeRolledBack},
			replaceErr: failure, wantMessage: "replace executable", wantCalls: [4]int{1, 1, 1, 1},
		},
		{
			name: "replacement requiring recovery", replaceResult: replacementResult{outcome: replacementOutcomeRecoveryRequired},
			replaceErr: failure, wantMessage: "replace executable", wantCalls: [4]int{1, 1, 1, 1},
			wantStaged: true, wantRecovery: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			executablePath := filepath.Join(dir, "d2tool")
			archivePath := filepath.Join(dir, downloadFilesPrefix+"archive")
			stagedPath := filepath.Join(dir, stagedFilesPrefix+"candidate")
			selectCalls, downloadCalls, extractCalls := 0, 0, 0
			replacer := &fakeReplacer{replaceResult: tt.replaceResult, replaceErr: tt.replaceErr}

			installer := newArchiveInstallerForTest(
				func() (string, error) { return executablePath, tt.executableErr },
				replacer,
				func(*github.Release, string, string) (github.ReleaseAsset, error) {
					selectCalls++
					return github.ReleaseAsset{}, tt.selectErr
				},
				func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error) {
					downloadCalls++
					if tt.downloadErr != nil {
						return "", tt.downloadErr
					}
					if err := os.WriteFile(archivePath, []byte("archive"), 0600); err != nil {
						t.Fatal(err)
					}
					return archivePath, nil
				},
				func(context.Context, string, string, string) (string, error) {
					extractCalls++
					if tt.extractErr != nil {
						return "", tt.extractErr
					}
					if err := os.WriteFile(stagedPath, []byte("new"), 0700); err != nil {
						t.Fatal(err)
					}
					return stagedPath, nil
				},
			)

			err := installer.Install(context.Background(), &github.Release{})
			if err == nil || !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("Install() error = %v, want substring %q", err, tt.wantMessage)
			}
			if got := isRecoveryRequired(err); got != tt.wantRecovery {
				t.Fatalf("isRecoveryRequired(Install error) = %v, want %v; error = %v", got, tt.wantRecovery, err)
			}
			gotCalls := [4]int{selectCalls, downloadCalls, extractCalls, replacer.replaceCalls}
			if gotCalls != tt.wantCalls {
				t.Fatalf("calls = %v, want %v", gotCalls, tt.wantCalls)
			}
			if downloadCalls > 0 && tt.downloadErr == nil {
				if _, statErr := os.Stat(archivePath); !os.IsNotExist(statErr) {
					t.Fatalf("archive was not removed: %v", statErr)
				}
			}
			if extractCalls > 0 && tt.extractErr == nil {
				contents, statErr := os.ReadFile(stagedPath)
				if tt.wantStaged {
					if statErr != nil || string(contents) != "new" {
						t.Fatalf("staged recovery file = %q, error = %v", contents, statErr)
					}
				} else if !os.IsNotExist(statErr) {
					t.Fatalf("disposable staged file was not removed: %v", statErr)
				}
			}
		})
	}
}

func TestArchiveInstaller_CleanupWrapsReplacerError(t *testing.T) {
	replacer := &fakeReplacer{cleanupErr: errors.New("remove failed")}
	installer := newArchiveInstallerForTest(
		func() (string, error) { return "/install/d2tool", nil },
		replacer,
		nil,
		nil,
		nil,
	)
	if err := installer.Cleanup(); err == nil || !strings.Contains(err.Error(), "clean stale update files") {
		t.Fatalf("Cleanup() error = %v", err)
	}
}
