//go:build windows

package systray

import (
	"context"
	"log/slog"

	"fyne.io/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	trayStop   func()
	systrayCtx chan context.Context
)

func InitSystray() {
	systrayCtx = make(chan context.Context, 1)

	var trayStart func()
	trayStart, trayStop = systray.RunWithExternalLoop(func() {
		ctx := <-systrayCtx
		onSystrayReady(ctx)
	}, nil)

	go trayStart()
}

func StartSystray(ctx context.Context) {
	systrayCtx <- ctx
}

func StopSystray() {
	if trayStop != nil {
		trayStop()
	}
}

func onSystrayReady(ctx context.Context) {
	systray.SetIcon(main.appIcon)
	systray.SetTooltip("D2Tool")

	mShow := systray.AddMenuItem("Show", "Show the app")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	go func() {
		for range mShow.ClickedCh {
			runtime.Show(ctx)
		}
	}()

	go func() {
		for range mQuit.ClickedCh {
			slog.Info("Quitting...")
			runtime.Quit(ctx)
		}
	}()
}
