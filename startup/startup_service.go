package startup

type StartupService interface {
	StartupRegister() error
	StartupRemove() error
	IsStartupRegistered() (bool, error)
	SupportsStartup() bool
}
