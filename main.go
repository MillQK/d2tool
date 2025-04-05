package main

import (
	"d2tool/app"
	"flag"
	"log/slog"
	"os"
	"path"
)

func main() {
	setupLogger()

	minimized := flag.Bool("minimized", false, "run the application minimized")
	flag.Parse()

	// Run the GUI application
	app.RunGUI(*minimized)
}

func setupLogger() {
	executablePath, err := os.Executable()
	if err != nil {
		slog.Error("unable to get the executable path", "error", err)
		return
	}

	logFilePath := path.Join(path.Dir(executablePath), "d2tool.log")
	slog.Info("setting up logger", "path", logFilePath)
	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		slog.Error("unable to log file", "error", err, "path", logFilePath)
		return
	}
	textHandler := slog.NewTextHandler(file, nil)
	slog.SetDefault(slog.New(textHandler))
}
