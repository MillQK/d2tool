package main

import (
	"context"
	"d2tool/config"
	"embed"
	"flag"
	"log/slog"
	"os"
	"path"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	// Parse command line flags
	minimized := flag.Bool("minimized", false, "start the application minimized")
	flag.Parse()

	// Setup file logging
	setupLogger()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:       "D2Tool",
		Width:       1000,
		Height:      800,
		StartHidden: *minimized,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		slog.Error("Error starting application", "error", err)
	}
}

// setupLogger configures file-based logging
func setupLogger() {
	executablePath, err := os.Executable()
	if err != nil {
		slog.Error("unable to get the executable path", "error", err)
		return
	}

	logFilePath := path.Join(path.Dir(executablePath), "d2tool.log")
	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		slog.Error("unable to open log file", "error", err, "path", logFilePath)
		return
	}

	textHandler := slog.NewTextHandler(file, nil)
	slog.SetDefault(slog.New(textHandler))
	slog.Info("Logger initialized", "path", logFilePath)
}

// App struct
type App struct {
	ctx    context.Context
	config *config.Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.config = config.LoadConfig()
	a.startBackgroundTasks()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.config.Save()
}
