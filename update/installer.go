package update

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"d2tool/github"
)

type releaseInstaller interface {
	Install(context.Context, *github.Release) error
	Cleanup() error
}

type recoveryRequired interface {
	error
	recoveryRequired()
}

type recoveryRequiredError struct {
	cause error
}

func (e *recoveryRequiredError) Error() string     { return e.cause.Error() }
func (e *recoveryRequiredError) Unwrap() error     { return e.cause }
func (e *recoveryRequiredError) recoveryRequired() {}

func isRecoveryRequired(err error) bool {
	var target recoveryRequired
	return errors.As(err, &target)
}

type archiveInstaller struct {
	executablePath func() (string, error)
	client         httpDoer
	replacer       executableReplacer
	goos           string
	goarch         string
	selectAsset    func(*github.Release, string, string) (github.ReleaseAsset, error)
	download       func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error)
	extract        func(context.Context, string, string, string) (string, error)
}

func newArchiveInstaller(client httpDoer) *archiveInstaller {
	return &archiveInstaller{
		executablePath: os.Executable,
		client:         client,
		replacer:       newOSExecutableReplacer(),
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
		selectAsset:    selectReleaseAsset,
		download:       downloadReleaseArchive,
		extract:        extractReleaseExecutable,
	}
}

func newArchiveInstallerForTest(
	executablePath func() (string, error),
	replacer executableReplacer,
	selectAsset func(*github.Release, string, string) (github.ReleaseAsset, error),
	download func(context.Context, httpDoer, github.ReleaseAsset, string) (string, error),
	extract func(context.Context, string, string, string) (string, error),
) *archiveInstaller {
	return &archiveInstaller{
		executablePath: executablePath,
		client:         http.DefaultClient,
		replacer:       replacer,
		goos:           "linux",
		goarch:         "amd64",
		selectAsset:    selectAsset,
		download:       download,
		extract:        extract,
	}
}

func (i *archiveInstaller) Install(ctx context.Context, release *github.Release) error {
	executablePath, err := i.executablePath()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	if err := validateCanonicalExecutablePath(executablePath); err != nil {
		return fmt.Errorf("validate executable path: %w", err)
	}
	dir := filepath.Dir(executablePath)
	asset, err := i.selectAsset(release, i.goos, i.goarch)
	if err != nil {
		return fmt.Errorf("select update asset: %w", err)
	}
	archivePath, err := i.download(ctx, i.client, asset, dir)
	if err != nil {
		return fmt.Errorf("download update: %w", err)
	}
	defer os.Remove(archivePath)
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("extract update: %w", err)
	}
	stagedPath, err := i.extract(ctx, archivePath, dir, i.goos)
	if err != nil {
		return fmt.Errorf("extract update: %w", err)
	}
	if err := ctx.Err(); err != nil {
		if removeErr := os.Remove(stagedPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return fmt.Errorf("cancel update before replacement: %w; remove staged executable %q: %v", err, stagedPath, removeErr)
		}
		return fmt.Errorf("cancel update before replacement: %w", err)
	}
	result, replaceErr := i.replacer.Replace(stagedPath, executablePath)
	if result.outcome != replacementOutcomeRecoveryRequired {
		if removeErr := os.Remove(stagedPath); removeErr != nil && !os.IsNotExist(removeErr) {
			if replaceErr != nil {
				return fmt.Errorf("replace executable: %w; remove staged executable %q: %v", replaceErr, stagedPath, removeErr)
			}
			return fmt.Errorf("remove staged executable %q: %w", stagedPath, removeErr)
		}
	}
	if replaceErr != nil {
		wrapped := fmt.Errorf("replace executable: %w", replaceErr)
		if result.outcome == replacementOutcomeRecoveryRequired {
			return &recoveryRequiredError{cause: wrapped}
		}
		return wrapped
	}
	return nil
}

func (i *archiveInstaller) Cleanup() error {
	executablePath, err := i.executablePath()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	if err := validateCanonicalExecutablePath(executablePath); err != nil {
		return fmt.Errorf("validate executable path: %w", err)
	}
	if err := i.replacer.Cleanup(executablePath); err != nil {
		return fmt.Errorf("clean stale update files: %w", err)
	}
	return nil
}
