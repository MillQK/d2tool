//go:build !windows

package systray

import "context"

func InitSystray(iconBytes []byte) {}
func StartSystray(context.Context) {}
func StopSystray()                 {}
func IsSupported() bool            { return false }
