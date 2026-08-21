package update

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"d2tool/github"
)

const (
	downloadFilesPrefix = ".download.d2tool-"
	stagedFilesPrefix   = ".new.d2tool-"
)

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func selectReleaseAsset(release *github.Release, goos, goarch string) (github.ReleaseAsset, error) {
	expected := fmt.Sprintf("d2tool-%s-%s.zip", goos, goarch)
	matches := make([]github.ReleaseAsset, 0, 1)
	for _, asset := range release.Assets {
		if asset.Name == expected {
			matches = append(matches, asset)
		}
	}
	if len(matches) != 1 {
		return github.ReleaseAsset{}, fmt.Errorf("expected exactly one asset %q, found %d", expected, len(matches))
	}
	return matches[0], nil
}

func downloadReleaseArchive(ctx context.Context, client httpDoer, asset github.ReleaseAsset, dir string) (result string, resultErr error) {
	if asset.Size <= 0 {
		return "", fmt.Errorf("invalid update asset size %d", asset.Size)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return "", fmt.Errorf("create asset request: %w", err)
	}
	request.Header.Set("Accept", "application/octet-stream")
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download asset: server returned %s", response.Status)
	}

	file, err := os.CreateTemp(dir, downloadFilesPrefix+"*")
	if err != nil {
		return "", fmt.Errorf("create download file: %w", err)
	}
	filePath := file.Name()
	keep := false
	defer func() {
		_ = file.Close()
		if !keep {
			_ = os.Remove(filePath)
		}
	}()

	hash := sha256.New()
	limited := &io.LimitedReader{R: response.Body, N: asset.Size}
	written, err := io.Copy(io.MultiWriter(file, hash), limited)
	if err != nil {
		return "", fmt.Errorf("write download: %w", err)
	}
	if written != asset.Size {
		return "", fmt.Errorf("downloaded asset size mismatch: expected %d, got %d", asset.Size, written)
	}
	var extra [1]byte
	extraBytes, extraErr := io.ReadFull(response.Body, extra[:])
	if extraBytes > 0 {
		return "", fmt.Errorf("downloaded asset size mismatch: response exceeds declared size %d", asset.Size)
	}
	if extraErr != io.EOF {
		return "", fmt.Errorf("check downloaded asset size: %w", extraErr)
	}
	if err := validateDigest(asset.Digest, hash.Sum(nil)); err != nil {
		return "", err
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync download: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close download: %w", err)
	}
	keep = true
	return filePath, nil
}

func validateDigest(value string, actual []byte) error {
	if value == "" {
		return nil
	}
	algorithm, expected, ok := strings.Cut(value, ":")
	if !ok || algorithm != "sha256" {
		return fmt.Errorf("unsupported asset digest %q", value)
	}
	if !strings.EqualFold(expected, hex.EncodeToString(actual)) {
		return fmt.Errorf("downloaded asset digest mismatch")
	}
	return nil
}

func expectedExecutableName(goos string) string {
	if goos == "windows" {
		return "d2tool.exe"
	}
	return "d2tool"
}

func extractReleaseExecutable(ctx context.Context, archivePath, dir, goos string) (result string, resultErr error) {
	archive, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open update archive: %w", err)
	}
	defer archive.Close()
	info, err := archive.Stat()
	if err != nil {
		return "", fmt.Errorf("stat update archive: %w", err)
	}
	reader, err := zip.NewReader(archive, info.Size())
	if err != nil {
		return "", fmt.Errorf("read update archive: %w", err)
	}

	expected := expectedExecutableName(goos)
	var executable *zip.File
	for _, entry := range reader.File {
		cleanName := path.Clean(entry.Name)
		if path.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, "../") {
			return "", fmt.Errorf("illegal archive path %q", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			continue
		}
		if entry.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("archive entry %q is a symlink", entry.Name)
		}
		if !entry.Mode().IsRegular() || cleanName != expected {
			return "", fmt.Errorf("unexpected archive entry %q", entry.Name)
		}
		if executable != nil {
			return "", fmt.Errorf("duplicate expected executable %q", entry.Name)
		}
		if goos != "windows" && entry.Mode().Perm()&0111 == 0 {
			return "", fmt.Errorf("archive executable %q is not executable", entry.Name)
		}
		executable = entry
	}
	if executable == nil {
		return "", fmt.Errorf("expected executable %q is missing", expected)
	}

	source, err := executable.Open()
	if err != nil {
		return "", fmt.Errorf("open archived executable: %w", err)
	}
	defer source.Close()
	staged, err := os.CreateTemp(dir, stagedFilesPrefix+"*")
	if err != nil {
		return "", fmt.Errorf("create staged executable: %w", err)
	}
	stagedPath := staged.Name()
	keep := false
	defer func() {
		_ = staged.Close()
		if !keep {
			_ = os.Remove(stagedPath)
		}
	}()
	if _, err := io.Copy(staged, contextReader{ctx: ctx, reader: source}); err != nil {
		return "", fmt.Errorf("write staged executable: %w", err)
	}
	if err := staged.Chmod(executable.Mode().Perm()); err != nil {
		return "", fmt.Errorf("set staged executable mode: %w", err)
	}
	if err := staged.Sync(); err != nil {
		return "", fmt.Errorf("sync staged executable: %w", err)
	}
	if err := staged.Close(); err != nil {
		return "", fmt.Errorf("close staged executable: %w", err)
	}
	keep = true
	return stagedPath, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}
