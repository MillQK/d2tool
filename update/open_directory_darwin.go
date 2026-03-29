//go:build darwin

package update

import "os/exec"

func openDirectoryInFileManager(path string) error {
	return exec.Command("open", path).Start()
}
