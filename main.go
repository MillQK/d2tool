package main

import (
	"context"
	"d2tool/config"
	"d2tool/github"
	"d2tool/heroesLayout"
	"d2tool/startup"
	"d2tool/systray"
	"d2tool/update"
	"d2tool/utils"
	"embed"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.ico
var appIcon []byte

//go:embed wails.json
var wailsJSON []byte

func main() {
	// Early debug output (before logger is set up)
	fmt.Println("D2Tool starting...")

	// Parse command line flags
	minimizedFlagName := "minimized"
	minimized := flag.Bool(minimizedFlagName, false, "start the application minimized")
	flag.Parse()

	// Setup file logging
	setupLogger()

	wailsProjectConfig, err := config.ParseWailsProjectConfig(wailsJSON)
	if err != nil {
		slog.Error("Error parsing wails.json", "error", err)
		os.Exit(1)
	}

	appConfig := config.LoadConfig()

	// Create an instance of the app structure
	app := NewApp(
		appConfig,
		update.NewUpdateService(
			wailsProjectConfig.Info.ProductVersion,
			github.NewHttpClient(),
		),
		heroesLayout.NewHeroesLayoutService(appConfig),
		startup.NewStartupService([]string{fmt.Sprintf("-%s", minimizedFlagName)}),
	)

	// Initialize systray (Windows only, no-op on other platforms)
	systray.InitSystray(appIcon)

	// Create application with options
	err = wails.Run(&options.App{
		Title:             "D2Tool",
		Width:             1000,
		Height:            800,
		StartHidden:       *minimized,
		HideWindowOnClose: systray.IsSupported(), // Hide window when systray is supported
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 20, B: 25, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			systray.StartSystray(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			app.shutdown(ctx)
			systray.StopSystray()
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "d2tool-019ad9f4-1416-7b10-b8ce-2ab89c12279e",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				slog.Info("Second instance attempted to launch", "args", secondInstanceData.Args)
				fmt.Println("Another instance is already running!")
			},
		},
		Logger:             utils.NewSlogAdapter(),
		LogLevel:           logger.DEBUG,
		LogLevelProduction: logger.INFO,
	})

	if err != nil {
		slog.Error("Error starting application", "error", err)
		fmt.Printf("FATAL: Error starting application: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("D2Tool exited normally")
}

// setupLogger configures file-based logging
func setupLogger() {
	executablePath, err := os.Executable()
	if err != nil {
		slog.Error("unable to get the executable path", "error", err)
		return
	}

	logFilePath := filepath.Join(filepath.Dir(executablePath), "d2tool.log")

	fileLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    20, // megabytes
		MaxBackups: 3,
		MaxAge:     7,     //days
		Compress:   false, // disabled by default
	}

	multiWriter := io.MultiWriter(os.Stdout, fileLogger)
	textHandler := slog.NewTextHandler(multiWriter, nil)
	slog.SetDefault(slog.New(textHandler))
	slog.Info("Logger initialized", "path", logFilePath)
}

// App struct
type App struct {
	ctx                 context.Context
	config              *config.Config
	updateService       update.UpdateService
	heroesLayoutService heroesLayout.HeroesLayoutService
	startupService      startup.StartupService
}

// NewApp creates a new App application struct
func NewApp(
	config *config.Config,
	updateService update.UpdateService,
	heroesLayoutService heroesLayout.HeroesLayoutService,
	startupService startup.StartupService,
) *App {
	return &App{
		config:              config,
		updateService:       updateService,
		heroesLayoutService: heroesLayoutService,
		startupService:      startupService,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startBackgroundTasks()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.config.Save()
}
