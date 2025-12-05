package utils

import (
	"log/slog"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

// SlogAdapter adapts the default slog logger to the Wails logger interface.
type SlogAdapter struct{}

func NewSlogAdapter() *SlogAdapter {
	return &SlogAdapter{}
}

func (s *SlogAdapter) Print(message string) {
	slog.Info(message)
}

func (s *SlogAdapter) Trace(message string) {
	slog.Debug(message, "level", "trace")
}

func (s *SlogAdapter) Debug(message string) {
	slog.Debug(message)
}

func (s *SlogAdapter) Info(message string) {
	slog.Info(message)
}

func (s *SlogAdapter) Warning(message string) {
	slog.Warn(message)
}

func (s *SlogAdapter) Error(message string) {
	slog.Error(message)
}

func (s *SlogAdapter) Fatal(message string) {
	slog.Error(message, "level", "fatal")
}

var _ logger.Logger = (*SlogAdapter)(nil)

func IsStdoutAvailable() bool {
	// On Windows GUI apps, Stdout file descriptor is invalid
	if runtime.GOOS == "windows" {
		_, err := os.Stdout.Stat()
		return err == nil
	}
	return true
}
