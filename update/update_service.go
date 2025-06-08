package update

import (
	"archive/zip"
	"d2tool/github"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	oldFilesPrefix = ".old."
)

type UpdateService struct {
	lock sync.Mutex

	githubClient  github.Client
	latestRelease *github.Release
	appVersion    string

	OnCheckStarted  chan struct{}
	OnCheckFinished chan struct{}

	OnUpdateStarted  chan struct{}
	OnUpdateFinished chan struct{}
}

func NewUpdateService(
	githubClient github.Client,
	appVersion string,
) *UpdateService {
	return &UpdateService{
		githubClient:     githubClient,
		appVersion:       appVersion,
		OnCheckStarted:   make(chan struct{}),
		OnCheckFinished:  make(chan struct{}),
		OnUpdateStarted:  make(chan struct{}),
		OnUpdateFinished: make(chan struct{}),
	}
}

func (s *UpdateService) UpdateAvailable() bool {
	latestVersion := s.LatestAvailableVersion()
	return latestVersion != "" && s.appVersion != latestVersion
}

func (s *UpdateService) UpdateApp() error {
	s.OnUpdateStarted <- struct{}{}
	defer func() {
		s.OnUpdateFinished <- struct{}{}
	}()

	err := downloadAndUnarchiveLatestReleaseVersion(s.latestRelease)
	if err != nil {
		return err
	}

	return nil
}

func (s *UpdateService) RunPeriodicUpdateCheck(
	forceUpdateChan chan struct{},
) {
	for {
		select {
		case <-time.After(1 * time.Hour):
			slog.Debug("Checking for updates after timeout")
		case <-forceUpdateChan:
			slog.Debug("Forcing update check")
		}

		err := s.checkForUpdate()
		if err != nil {
			slog.Error("Unable to check for updates", "error", err)
		}
	}
}

func (s *UpdateService) LatestAvailableVersion() string {
	latestRelease := s.latestRelease
	if latestRelease == nil {
		return ""
	}

	return latestRelease.TagName
}

func (s *UpdateService) CleanupOldFiles() error {
	return cleanupOldFiles()
}

func (s *UpdateService) checkForUpdate() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.OnCheckStarted <- struct{}{}
	defer func() {
		s.OnCheckFinished <- struct{}{}
	}()

	err := s.checkRelease()
	if err != nil {
		return err
	}

	return nil
}

func (s *UpdateService) checkRelease() error {
	release, err := s.githubClient.GetLatestRelease()
	if err != nil {
		return err
	}

	s.latestRelease = release
	return nil
}

func cleanupOldFiles() error {
	rootDir, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	return filepath.WalkDir(rootDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(entry.Name(), oldFilesPrefix) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("error removing old file: %w", err)
			}
		}

		return nil
	})
}

func downloadAndUnarchiveLatestReleaseVersion(
	release *github.Release,
) error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	archiveNamePrefix := constructArchiveNamePrefix()
	var appAsset *github.ReleaseAsset
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, archiveNamePrefix) {
			appAsset = &asset
			break
		}
	}

	if appAsset == nil {
		return fmt.Errorf("no asset with prefix %s found for release %s", archiveNamePrefix, release.TagName)
	}

	response, err := http.DefaultClient.Get(appAsset.URL)
	if err != nil {
		return fmt.Errorf("error downloading asset: %w", err)
	}

	defer response.Body.Close()

	rootDir := filepath.Dir(executablePath)
	file, err := os.Create(filepath.Join(rootDir, appAsset.Name))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()
	defer os.Remove(file.Name())

	written, err := io.Copy(file, response.Body)

	if err != nil {
		return fmt.Errorf("error copying asset: %w", err)
	}

	if written != appAsset.Size {
		return fmt.Errorf("downloaded asset size mismatch: expected %d, got %d", appAsset.Size, written)
	}

	zipReader, err := zip.NewReader(file, written)
	if err != nil {
		return fmt.Errorf("error creating zip reader: %w", err)
	}

	for _, f := range zipReader.File {
		err = extractFileFromArchive(f, rootDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractFileFromArchive(
	zipReaderFile *zip.File,
	rootDir string,
) error {
	filePath := filepath.Join(rootDir, zipReaderFile.Name)

	if zipReaderFile.FileInfo().IsDir() {
		if err := os.Mkdir(filePath, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("error creating parent directories: %w", err)
	}

	_, err := os.Stat(filePath)
	if err == nil {
		parentDir := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)
		err = os.Rename(filePath, filepath.Join(parentDir, oldFilesPrefix+fileName))
		if err != nil {
			return fmt.Errorf("error renaming file: %w", err)
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error getting file info: %w", err)
		}
	}

	dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipReaderFile.Mode())
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer dstFile.Close()

	fileInArchive, err := zipReaderFile.Open()
	if err != nil {
		return fmt.Errorf("error opening file in archive: %w", err)
	}

	defer fileInArchive.Close()

	_, err = io.Copy(dstFile, fileInArchive)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}

func constructArchiveNamePrefix() string {
	return fmt.Sprintf("d2tool-%s-%s", runtime.GOOS, runtime.GOARCH)
}
