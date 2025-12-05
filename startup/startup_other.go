//go:build !windows
// +build !windows

package startup

import "fmt"

type startupServiceOtherImpl struct{}

func NewStartupService(runArgs []string) StartupService {
	return &startupServiceOtherImpl{}
}

func (s *startupServiceOtherImpl) StartupRegister() error {
	return fmt.Errorf("this functionality is only available on Windows")
}

func (s *startupServiceOtherImpl) StartupRemove() error {
	return fmt.Errorf("this functionality is only available on Windows")
}

func (s *startupServiceOtherImpl) IsStartupRegistered() (bool, error) {
	return false, fmt.Errorf("this functionality is only available on Windows")
}

func (s *startupServiceOtherImpl) SupportsStartup() bool {
	return false
}
