//go:build linux

package update

import "os/exec"

func openDirectoryInFileManager(path string) error {
	return exec.Command("xdg-open", path).Start()
}
