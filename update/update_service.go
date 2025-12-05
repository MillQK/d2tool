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

type UpdateState struct {
	UpdateAvailable   bool
	CurrentAppVersion string
	LatestAppVersion  string
	LastCheckTime     time.Time
}

type UpdateService interface {
	GetState() UpdateState
	CheckForUpdate() error
	UpdateApp() error
}

type UpdateServiceImpl struct {
	lock sync.Mutex

	currentAppVersion string
	githubClient      github.Client
	downloadClient    *http.Client

	latestRelease *github.Release
	lastCheckTime time.Time
}

func NewUpdateService(
	currentAppVersion string,
	githubClient github.Client,
) *UpdateServiceImpl {
	return &UpdateServiceImpl{
		currentAppVersion: currentAppVersion,
		githubClient:      githubClient,
		downloadClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
		lastCheckTime: time.UnixMilli(0),
	}
}

func (s *UpdateServiceImpl) GetState() UpdateState {
	s.lock.Lock()
	defer s.lock.Unlock()

	latestVersion := s.latestAvailableVersionLocked()

	return UpdateState{
		UpdateAvailable:   isUpdateAvailable(latestVersion, s.currentAppVersion),
		CurrentAppVersion: s.currentAppVersion,
		LatestAppVersion:  latestVersion,
		LastCheckTime:     s.lastCheckTime,
	}
}

func (s *UpdateServiceImpl) CheckForUpdate() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := cleanupOldFiles(); err != nil {
		slog.Warn("Error cleaning up old files", "error", err)
	}

	release, err := s.githubClient.GetLatestRelease()
	if err != nil {
		return err
	}

	s.latestRelease = release
	s.lastCheckTime = time.Now()
	return nil
}

func (s *UpdateServiceImpl) UpdateApp() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.latestRelease == nil {
		return fmt.Errorf("no release available to update to")
	}

	latestVersion := s.latestAvailableVersionLocked()
	if !isUpdateAvailable(latestVersion, s.currentAppVersion) {
		return fmt.Errorf("no update available for current version %s and latest version %s", s.currentAppVersion, latestVersion)
	}

	return s.downloadAndUnarchiveLatestReleaseVersion()
}

func (s *UpdateServiceImpl) latestAvailableVersionLocked() string {
	if s.latestRelease == nil {
		return ""
	}
	return s.latestRelease.TagName
}

func isUpdateAvailable(latestVersion string, currentVersion string) bool {
	return latestVersion != "" && currentVersion != latestVersion
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

func (s *UpdateServiceImpl) downloadAndUnarchiveLatestReleaseVersion() error {
	if err := cleanupOldFiles(); err != nil {
		return fmt.Errorf("error cleaning up old files: %w", err)
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	archiveNamePrefix := constructArchiveNamePrefix()
	var appAsset *github.ReleaseAsset
	for _, asset := range s.latestRelease.Assets {
		if strings.HasPrefix(asset.Name, archiveNamePrefix) {
			appAsset = &asset
			break
		}
	}

	if appAsset == nil {
		return fmt.Errorf("no asset with prefix %s found for release %s", archiveNamePrefix, s.latestRelease.TagName)
	}

	slog.Info("Downloading and unarchiving latest release version", "asset", appAsset)

	request, err := http.NewRequest(http.MethodGet, appAsset.URL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	request.Header.Set("Accept", "application/octet-stream")

	response, err := s.downloadClient.Do(request)
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
