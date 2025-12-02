//go:build !windows

package systray

import "context"

func InitSystray()                 {}
func StartSystray(context.Context) {}
func StopSystray()                 {}
