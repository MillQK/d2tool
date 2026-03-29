//go:build windows

package update

import "os/exec"

func openDirectoryInFileManager(path string) error {
	return exec.Command("explorer", path).Start()
}
