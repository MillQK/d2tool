//go:build windows
// +build windows

package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// Windows API constants and types
const (
	CSIDL_PROGRAM_FILES    = 0x0026
	CSIDL_PROGRAM_FILESX86 = 0x002a
	SHGFP_TYPE_CURRENT     = 0
	MAX_PATH               = 260
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	shGetFolderPathW = shell32.NewProc("SHGetFolderPathW")
)

// FindSteamPath attempts to find the Steam installation path on Windows
func FindSteamPath() (string, error) {
	// Try to get Steam path from registry
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		steamPath, _, err := k.GetStringValue("SteamPath")
		if err == nil && steamPath != "" {
			return steamPath, nil
		}
	}

	// If registry fails, try common installation paths
	programFiles, err := getSpecialFolderPath(CSIDL_PROGRAM_FILES)
	if err == nil {
		steamPath := filepath.Join(programFiles, "Steam")
		if _, err := os.Stat(steamPath); err == nil {
			return steamPath, nil
		}
	}

	programFilesX86, err := getSpecialFolderPath(CSIDL_PROGRAM_FILESX86)
	if err == nil {
		steamPath := filepath.Join(programFilesX86, "Steam")
		if _, err := os.Stat(steamPath); err == nil {
			return steamPath, nil
		}
	}

	// Check common drive letters
	for _, drive := range []string{"C:", "D:", "E:", "F:"} {
		steamPath := filepath.Join(drive, "Program Files (x86)", "Steam")
		if _, err := os.Stat(steamPath); err == nil {
			return steamPath, nil
		}

		steamPath = filepath.Join(drive, "Program Files", "Steam")
		if _, err := os.Stat(steamPath); err == nil {
			return steamPath, nil
		}
	}

	return "", fmt.Errorf("could not find Steam installation path")
}

// getSpecialFolderPath gets a special folder path using Windows API
func getSpecialFolderPath(folderID int) (string, error) {
	b := make([]uint16, MAX_PATH)
	ret, _, _ := shGetFolderPathW.Call(
		0,
		uintptr(folderID),
		0,
		uintptr(SHGFP_TYPE_CURRENT),
		uintptr(unsafe.Pointer(&b[0])),
	)
	if ret != 0 {
		return "", fmt.Errorf("failed to get folder path")
	}
	return syscall.UTF16ToString(b), nil
}
