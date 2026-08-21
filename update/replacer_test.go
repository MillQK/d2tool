package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestExecutableReplacer_Replace(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	staged := filepath.Join(dir, stagedFilesPrefix+"candidate")
	writeFile(t, target, "old")
	writeFile(t, staged, "new")

	replacer := newOSExecutableReplacer()
	result, err := replacer.Replace(staged, target)
	if err != nil {
		t.Fatal(err)
	}
	if result.outcome != replacementOutcomeSucceeded {
		t.Fatalf("replacement outcome = %v, want succeeded", result.outcome)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Fatalf("target = %q", got)
	}
	if got, _ := os.ReadFile(result.backupPath); string(got) != "old" {
		t.Fatalf("backup = %q", got)
	}
}

func TestExecutableReplacer_ReplaceKeepsRetainedBackupsDistinctWithFixedClock(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	firstStaged := filepath.Join(dir, stagedFilesPrefix+"first")
	secondStaged := filepath.Join(dir, stagedFilesPrefix+"second")
	writeFile(t, target, "old")
	writeFile(t, firstStaged, "first")

	replacer := newOSExecutableReplacer()
	replacer.now = func() time.Time { return time.Unix(123, 0) }
	firstResult, err := replacer.Replace(firstStaged, target)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, secondStaged, "second")
	secondResult, err := replacer.Replace(secondStaged, target)
	if err != nil {
		t.Fatal(err)
	}

	if firstResult.backupPath == secondResult.backupPath {
		t.Fatalf("backup paths must be distinct: %q", firstResult.backupPath)
	}
	if got, err := os.ReadFile(firstResult.backupPath); err != nil || string(got) != "old" {
		t.Fatalf("first backup = %q, error = %v", got, err)
	}
	if got, err := os.ReadFile(secondResult.backupPath); err != nil || string(got) != "first" {
		t.Fatalf("second backup = %q, error = %v", got, err)
	}
}

func TestExecutableReplacer_ReportsUnchangedWhenBackupRenameFails(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	staged := filepath.Join(dir, stagedFilesPrefix+"candidate")
	writeFile(t, target, "old")
	writeFile(t, staged, "new")

	replacer := newOSExecutableReplacer()
	replacer.rename = func(string, string) error { return errors.New("backup rename failed") }

	result, err := replacer.Replace(staged, target)
	if err == nil {
		t.Fatal("replacement must fail")
	}
	if result.outcome != replacementOutcomeUnchanged {
		t.Fatalf("replacement outcome = %v, want unchanged", result.outcome)
	}
	if got, readErr := os.ReadFile(target); readErr != nil || string(got) != "old" {
		t.Fatalf("target = %q, error = %v", got, readErr)
	}
	if got, readErr := os.ReadFile(staged); readErr != nil || string(got) != "new" {
		t.Fatalf("staged = %q, error = %v", got, readErr)
	}
}

func TestExecutableReplacer_RestoresBackupWhenInstallRenameFails(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	staged := filepath.Join(dir, stagedFilesPrefix+"candidate")
	writeFile(t, target, "old")
	writeFile(t, staged, "new")

	realRename := os.Rename
	calls := 0
	replacer := newOSExecutableReplacer()
	replacer.rename = func(oldPath, newPath string) error {
		calls++
		if calls == 2 {
			return errors.New("install rename failed")
		}
		return realRename(oldPath, newPath)
	}

	result, err := replacer.Replace(staged, target)
	if err == nil {
		t.Fatal("replacement must fail")
	}
	if result.outcome != replacementOutcomeRolledBack {
		t.Fatalf("replacement outcome = %v, want rolled back", result.outcome)
	}
	if got, _ := os.ReadFile(target); string(got) != "old" {
		t.Fatalf("target was not restored: %q", got)
	}
}

func TestExecutableReplacer_ReportsRestorationFailurePaths(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	staged := filepath.Join(dir, stagedFilesPrefix+"candidate")
	writeFile(t, target, "old")
	writeFile(t, staged, "new")

	realRename := os.Rename
	calls := 0
	replacer := newOSExecutableReplacer()
	replacer.rename = func(oldPath, newPath string) error {
		calls++
		if calls >= 2 {
			return errors.New("rename failed")
		}
		return realRename(oldPath, newPath)
	}

	result, err := replacer.Replace(staged, target)
	if err == nil {
		t.Fatal("replacement must fail")
	}
	if result.outcome != replacementOutcomeRecoveryRequired {
		t.Fatalf("replacement outcome = %v, want recovery required", result.outcome)
	}
	if !strings.Contains(err.Error(), target) || !strings.Contains(err.Error(), result.backupPath) {
		t.Fatalf("error must contain recovery paths: %v", err)
	}
	if result.backupPath == staged {
		t.Fatalf("backup must differ from staged executable: %q", result.backupPath)
	}
	if !strings.HasPrefix(filepath.Base(result.backupPath), oldFilesPrefix) {
		t.Fatalf("backup must use reserved namespace: %q", result.backupPath)
	}
	if got, statErr := os.ReadFile(result.backupPath); statErr != nil || string(got) != "old" {
		t.Fatalf("backup must preserve original executable: contents=%q error=%v", got, statErr)
	}
}

func TestExecutableReplacer_ReplaceRefusesReservedExecutablePath(t *testing.T) {
	for _, prefix := range []string{oldFilesPrefix, stagedFilesPrefix, downloadFilesPrefix} {
		t.Run(prefix, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, prefix+"recovery")
			staged := filepath.Join(dir, "candidate")
			writeFile(t, target, "recovery")
			writeFile(t, staged, "candidate")

			replacer := newOSExecutableReplacer()
			if _, err := replacer.Replace(staged, target); err == nil || !strings.Contains(err.Error(), "reserved") {
				t.Fatalf("Replace() error = %v, want reserved-path rejection", err)
			}
			if got, err := os.ReadFile(target); err != nil || string(got) != "recovery" {
				t.Fatalf("recovery executable changed: contents=%q error=%v", got, err)
			}
			if got, err := os.ReadFile(staged); err != nil || string(got) != "candidate" {
				t.Fatalf("staged executable changed: contents=%q error=%v", got, err)
			}
		})
	}
}

func TestExecutableReplacer_CleanupRemovesOnlyReservedFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	writeFile(t, target, "current")
	reserved := []string{
		filepath.Join(dir, oldFilesPrefix+"d2tool.1"),
		filepath.Join(dir, stagedFilesPrefix+"candidate"),
		filepath.Join(dir, downloadFilesPrefix+"archive"),
	}
	for _, path := range reserved {
		writeFile(t, path, "temporary")
	}
	unrelated := []string{
		filepath.Join(dir, "notes.txt"),
		filepath.Join(dir, ".env"),
		filepath.Join(dir, ".old.other"),
		filepath.Join(dir, "x.old.d2tool-candidate"),
	}
	for _, path := range unrelated {
		writeFile(t, path, "keep")
	}

	if err := newOSExecutableReplacer().Cleanup(target); err != nil {
		t.Fatal(err)
	}
	for _, path := range reserved {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("reserved file still exists: %s", path)
		}
	}
	for _, path := range unrelated {
		if got, err := os.ReadFile(path); err != nil || string(got) != "keep" {
			t.Fatalf("unrelated file changed: path=%q contents=%q error=%v", path, got, err)
		}
	}
}

func TestExecutableReplacer_CleanupRefusesWhenTargetIsMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d2tool")
	recovery := filepath.Join(dir, oldFilesPrefix+"d2tool.1")
	writeFile(t, recovery, "old")

	if err := newOSExecutableReplacer().Cleanup(target); err == nil {
		t.Fatal("cleanup must fail without the normal executable")
	}
	if got, err := os.ReadFile(recovery); err != nil || string(got) != "old" {
		t.Fatalf("recovery file changed: contents=%q error=%v", got, err)
	}
}

func TestExecutableReplacer_CleanupRefusesReservedExecutablePath(t *testing.T) {
	for _, prefix := range []string{oldFilesPrefix, stagedFilesPrefix, downloadFilesPrefix} {
		t.Run(prefix, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, prefix+"recovery")
			otherRecovery := filepath.Join(dir, oldFilesPrefix+"d2tool.1")
			writeFile(t, target, "launched recovery")
			writeFile(t, otherRecovery, "known good backup")

			if err := newOSExecutableReplacer().Cleanup(target); err == nil || !strings.Contains(err.Error(), "reserved") {
				t.Fatalf("Cleanup() error = %v, want reserved-path rejection", err)
			}
			if got, err := os.ReadFile(target); err != nil || string(got) != "launched recovery" {
				t.Fatalf("launched recovery changed: contents=%q error=%v", got, err)
			}
			if got, err := os.ReadFile(otherRecovery); err != nil || string(got) != "known good backup" {
				t.Fatalf("other recovery changed: contents=%q error=%v", got, err)
			}
		})
	}
}
