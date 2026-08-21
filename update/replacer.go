package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const oldFilesPrefix = ".old.d2tool-"

type replacementOutcome uint8

const (
	replacementOutcomeUnchanged replacementOutcome = iota
	replacementOutcomeSucceeded
	replacementOutcomeRolledBack
	replacementOutcomeRecoveryRequired
)

type replacementResult struct {
	outcome    replacementOutcome
	backupPath string
}

type executableReplacer interface {
	Replace(stagedPath, executablePath string) (replacementResult, error)
	Cleanup(executablePath string) error
}

type osExecutableReplacer struct {
	rename  func(string, string) error
	remove  func(string) error
	readDir func(string) ([]os.DirEntry, error)
	stat    func(string) (os.FileInfo, error)
	now     func() time.Time
}

func newOSExecutableReplacer() *osExecutableReplacer {
	return &osExecutableReplacer{
		rename:  os.Rename,
		remove:  os.Remove,
		readDir: os.ReadDir,
		stat:    os.Stat,
		now:     time.Now,
	}
}

func (r *osExecutableReplacer) Replace(stagedPath, executablePath string) (replacementResult, error) {
	unchanged := replacementResult{outcome: replacementOutcomeUnchanged}
	if err := validateCanonicalExecutablePath(executablePath); err != nil {
		return unchanged, err
	}
	dir := filepath.Dir(executablePath)
	backupPath, err := r.nextBackupPath(dir, filepath.Base(executablePath))
	if err != nil {
		return unchanged, err
	}
	if err := r.rename(executablePath, backupPath); err != nil {
		return unchanged, fmt.Errorf("back up current executable %q: %w", executablePath, err)
	}
	if err := r.rename(stagedPath, executablePath); err != nil {
		if restoreErr := r.rename(backupPath, executablePath); restoreErr != nil {
			return replacementResult{
				outcome:    replacementOutcomeRecoveryRequired,
				backupPath: backupPath,
			}, fmt.Errorf("install staged executable %q at %q: %v; restore backup %q: %w", stagedPath, executablePath, err, backupPath, restoreErr)
		}
		return replacementResult{
			outcome:    replacementOutcomeRolledBack,
			backupPath: backupPath,
		}, fmt.Errorf("install staged executable %q at %q: %w; original restored", stagedPath, executablePath, err)
	}
	return replacementResult{
		outcome:    replacementOutcomeSucceeded,
		backupPath: backupPath,
	}, nil
}

func (r *osExecutableReplacer) nextBackupPath(dir, executableName string) (string, error) {
	base := fmt.Sprintf("%s%s.%d", oldFilesPrefix, executableName, r.now().UnixNano())
	for suffix := uint64(0); ; suffix++ {
		name := base
		if suffix > 0 {
			name = fmt.Sprintf("%s.%d", base, suffix)
		}
		path := filepath.Join(dir, name)
		_, err := r.stat(path)
		if os.IsNotExist(err) {
			return path, nil
		}
		if err != nil {
			return "", fmt.Errorf("inspect backup path %q: %w", path, err)
		}
	}
}

func (r *osExecutableReplacer) Cleanup(executablePath string) error {
	if err := validateCanonicalExecutablePath(executablePath); err != nil {
		return err
	}
	if _, err := r.stat(executablePath); err != nil {
		return fmt.Errorf("refusing update cleanup without executable %q: %w", executablePath, err)
	}
	dir := filepath.Dir(executablePath)
	entries, err := r.readDir(dir)
	if err != nil {
		return fmt.Errorf("read executable directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, oldFilesPrefix) && !strings.HasPrefix(name, stagedFilesPrefix) && !strings.HasPrefix(name, downloadFilesPrefix) {
			continue
		}
		if err := r.remove(filepath.Join(dir, name)); err != nil {
			return fmt.Errorf("remove stale update file %q: %w", name, err)
		}
	}
	return nil
}

func validateCanonicalExecutablePath(executablePath string) error {
	name := filepath.Base(executablePath)
	for _, prefix := range []string{oldFilesPrefix, stagedFilesPrefix, downloadFilesPrefix} {
		if strings.HasPrefix(name, prefix) {
			return fmt.Errorf("executable path %q uses reserved update prefix %q", executablePath, prefix)
		}
	}
	return nil
}
