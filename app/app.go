package app

import (
	"d2tool/heroesGrid"
	"d2tool/startup"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log/slog"
	"slices"
	"time"
)

var appSize = fyne.NewSize(1000, 800)

type D2ToolApp struct {
	fApp       fyne.App
	mainWindow fyne.Window

	configBindings *ConfigBindings
}

// RunGUI starts the GUI application
func RunGUI(minimized bool) {
	fApp := app.NewWithID("com.github.millqk.d2tool")
	meta := fApp.Metadata()
	mainWindow := fApp.NewWindow(meta.Name)
	mainWindow.Resize(appSize)
	mainWindow.CenterOnScreen()

	configBindings := GetConfigBindings(fApp.Preferences())

	d2toolApp := &D2ToolApp{
		fApp:           fApp,
		mainWindow:     mainWindow,
		configBindings: configBindings,
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Home", d2toolApp.homeTabContent()),
		container.NewTabItem("Grid configs", d2toolApp.heroesGridConfigPathsTabContent()),
		container.NewTabItem("Positions order", d2toolApp.heroesGridPositionsOrderTabContent()),
		container.NewTabItem("Startup", d2toolApp.startupTabContent()),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	mainWindow.SetContent(tabs)
	mainWindow.SetPadded(true)

	if desk, ok := fApp.(desktop.App); ok {
		m := fyne.NewMenu(meta.Name,
			fyne.NewMenuItem("Show", func() {
				mainWindow.Show()
			}))
		desk.SetSystemTrayMenu(m)

		mainWindow.SetCloseIntercept(func() {
			mainWindow.Hide()
		})

		if !minimized {
			mainWindow.Show()
		}
	}

	fApp.Run()
}

func (app *D2ToolApp) homeTabContent() *fyne.Container {
	lastUpdateTimeLabel := widget.NewLabel(lastUpdateTimeText(time.UnixMilli(int64(app.configBindings.LastUpdateTimestampMillis.Get()))))
	lastUpdateTimeLabel.TextStyle = fyne.TextStyle{Bold: true}

	lastUpdateErrorLabel := widget.NewLabel("")
	updateLastUpdateErrorLabel(lastUpdateErrorLabel, app.configBindings.LastUpdateErrorMessage.Get())

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	forceUpdateChan := make(chan struct{})

	updateHeroesGridButton := widget.NewButton("Update heroes grid configs", func() {
		go func() {
			select {
			case forceUpdateChan <- struct{}{}:
				slog.Debug("Forcing update heroes grid configs")
			default:
				slog.Debug("Update is already in progress")
			}
		}()
	})

	go func() {
		updateStartedChan := make(chan struct{})
		updateResultChan := make(chan updateResult, 1)
		go app.heroesGridConfigUpdateLoop(forceUpdateChan, updateStartedChan, updateResultChan)
		for {
			select {
			case <-updateStartedChan:
				progressBar.Show()
				updateHeroesGridButton.Disable()
			case result := <-updateResultChan:
				app.configBindings.LastUpdateTimestampMillis.Set(int(result.updateTime.UnixMilli()))
				lastUpdateTimeLabel.SetText(lastUpdateTimeText(result.updateTime))

				errorMessage := ""
				if result.err != nil {
					errorMessage = result.err.Error()
				}
				app.configBindings.LastUpdateErrorMessage.Set(errorMessage)
				updateLastUpdateErrorLabel(lastUpdateErrorLabel, errorMessage)

				progressBar.Hide()
				updateHeroesGridButton.Enable()
			}
		}
	}()

	return container.NewVBox(
		lastUpdateTimeLabel,
		lastUpdateErrorLabel,
		updateHeroesGridButton,
		progressBar,
	)
}

func lastUpdateTimeText(lastUpdateTime time.Time) string {
	var lastUpdateTimeString string
	if lastUpdateTime.IsZero() {
		lastUpdateTimeString = "Never"
	} else {
		lastUpdateTimeString = lastUpdateTime.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("Last update time: %s", lastUpdateTimeString)
}

func updateLastUpdateErrorLabel(label *widget.Label, errorMessage string) {
	if errorMessage == "" {
		label.SetText("")
		label.Hide()
	} else {
		label.SetText(fmt.Sprintf("Error: %s", errorMessage))
		label.Show()
	}
}

func (app *D2ToolApp) heroesGridConfigPathsTabContent() *fyne.Container {
	heroesGridPathsList := widget.NewList(
		func() int {
			return len(app.configBindings.HeroesGridFilePaths.Get())
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				layout.NewSpacer(),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
			)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {

		},
	)

	heroesGridPathsList.UpdateItem = func(id widget.ListItemID, object fyne.CanvasObject) {
		item := app.configBindings.HeroesGridFilePaths.Get()[id]
		label := object.(*fyne.Container).Objects[0].(*widget.Label)
		label.SetText(item)

		deleteButton := object.(*fyne.Container).Objects[2].(*widget.Button)
		deleteButton.OnTapped = func() {
			currentPaths := app.configBindings.HeroesGridFilePaths.Get()
			app.configBindings.HeroesGridFilePaths.Set(append(currentPaths[:id], currentPaths[id+1:]...))
			heroesGridPathsList.Refresh()
		}
	}

	addConfigButton := widget.NewButton("Add Config", func() {
		dialog.ShowFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if reader == nil || err != nil {
					return
				}

				defer reader.Close()
				path := reader.URI().Path()
				currentPaths := app.configBindings.HeroesGridFilePaths.Get()
				if slices.Contains(currentPaths, path) {
					return
				}

				app.configBindings.HeroesGridFilePaths.Set(append(currentPaths, path))
				heroesGridPathsList.Refresh()
			},
			app.mainWindow,
		)
	})

	return container.NewBorder(
		widget.NewLabelWithStyle(
			"Heroes grid config paths",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		addConfigButton,
		nil,
		nil,
		heroesGridPathsList,
	)
}

func (app *D2ToolApp) heroesGridPositionsOrderTabContent() *fyne.Container {
	gridPositionsOrderList := widget.NewList(
		func() int {
			return len(app.configBindings.PositionsOrder.Get())
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				layout.NewSpacer(),
				container.NewVBox(
					widget.NewButtonWithIcon(
						"",
						theme.MoveUpIcon(),
						func() {
						},
					),
					widget.NewButtonWithIcon(
						"",
						theme.MoveDownIcon(),
						func() {
						},
					),
				),
			)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {

		},
	)

	gridPositionsOrderList.UpdateItem = func(id widget.ListItemID, object fyne.CanvasObject) {
		item := app.configBindings.PositionsOrder.Get()[id]
		listItemContainer := object.(*fyne.Container)
		listItemContainer.Objects[0].(*widget.Label).SetText(item)
		buttonsContainer := listItemContainer.Objects[2].(*fyne.Container)
		upButton := buttonsContainer.Objects[0].(*widget.Button)
		downButton := buttonsContainer.Objects[1].(*widget.Button)

		upButton.OnTapped = func() {
			if id > 0 {
				items := app.configBindings.PositionsOrder.Get()
				items[id], items[id-1] = items[id-1], items[id]
				app.configBindings.PositionsOrder.Set(items)
				gridPositionsOrderList.Refresh()
			}
		}

		downButton.OnTapped = func() {
			items := app.configBindings.PositionsOrder.Get()
			if id < len(items)-1 {
				items[id], items[id+1] = items[id+1], items[id]
				app.configBindings.PositionsOrder.Set(items)
				gridPositionsOrderList.Refresh()
			}
		}
	}

	return container.NewBorder(
		widget.NewLabelWithStyle(
			"Heroes grid positions order",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		nil,
		nil,
		nil,
		gridPositionsOrderList,
	)
}

func (app *D2ToolApp) startupTabContent() *fyne.Container {
	runOnStartupCheckBox := widget.NewCheck("Run on startup", func(value bool) {
		if value {
			startup.StartupRegister([]string{"-minimized"})
		} else {
			startup.StartupRemove()
		}
	})

	if startup.SupportsStartup() {
		startupRegistered, err := startup.IsStartupRegistered()
		if err != nil {
			slog.Warn("Unable to check startup registration", "error", err)
		}
		runOnStartupCheckBox.Checked = startupRegistered
	} else {
		runOnStartupCheckBox.Disable()
	}

	return container.NewVBox(
		runOnStartupCheckBox,
	)
}

type updateResult struct {
	updateTime time.Time
	err        error
}

func (app *D2ToolApp) heroesGridConfigUpdateLoop(
	forceUpdateChan chan struct{},
	updateStartedChan chan struct{},
	updateResultChan chan updateResult,
) {
	delay := time.Hour
	lastUpdateTime := time.UnixMilli(int64(app.configBindings.LastUpdateTimestampMillis.Get()))
	for {
		select {
		case <-forceUpdateChan:
			slog.Debug("Forcing update heroes grid configs")
		case <-time.After(lastUpdateTime.Add(delay).Sub(time.Now())):
			slog.Debug("Update heroes grid configs by time")
		}

		updateStartedChan <- struct{}{}

		err := heroesGrid.UpdateHeroesGrid(
			heroesGrid.UpdateHeroGridConfig{
				ConfigFilePaths: app.configBindings.HeroesGridFilePaths.Get(),
				Positions:       app.configBindings.PositionsOrder.Get(),
			},
		)

		lastUpdateTime = time.Now()
		updateResultChan <- updateResult{
			updateTime: lastUpdateTime,
			err:        err,
		}
	}
}
