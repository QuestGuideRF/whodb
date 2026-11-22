package common
import (
	"context"
	"fmt"
	"os"
	goruntime "runtime"
	"strings"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/clidey/whodb/core/src/analytics"
	"github.com/clidey/whodb/core/src/common/config"
	"github.com/clidey/whodb/core/src/common/datadir"
	"github.com/clidey/whodb/core/src/log"
)
const (
	filePermissionUserRW = 0644
)
type App struct {
	ctx            context.Context
	edition        string
	windowSettings WindowSettings
}
type WindowSettings struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}
func NewApp(edition string) *App {
	return &App{
		edition: edition,
	}
}
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.RestoreWindowState()
	a.SetupApplicationMenu()
	a.SetupSystemTray()
	analytics.Capture(ctx, "desktop_app_launched", map[string]any{
		"edition":  a.edition,
		"platform": goruntime.GOOS,
		"arch":     goruntime.GOARCH,
	})
}
func (a *App) Shutdown(ctx context.Context) {
	analytics.Capture(ctx, "desktop_app_closed", map[string]any{
		"edition": a.edition,
	})
}
func (a *App) beforeClose(ctx context.Context) bool {
	if err := a.SaveWindowState(); err != nil {
		log.Warnf("Failed to save window state: %v", err)
	}
	return false
}
func (a *App) DomReady(ctx context.Context) {
}
func (a *App) OpenURL(url string) error {
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}
func (a *App) SaveFile(data string, defaultName string) (string, error) {
	options := runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Save File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "CSV Files (*.csv)",
				Pattern:     "*.csv",
			},
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "SQL Files (*.sql)",
				Pattern:     "*.sql",
			},
			{
				DisplayName: "Excel Files (*.xlsx)",
				Pattern:     "*.xlsx",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	}
	filepath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	if filepath == "" {
		return "", nil
	}
	err = os.WriteFile(filepath, []byte(data), filePermissionUserRW)
	if err != nil {
		return "", err
	}
	return filepath, nil
}
func (a *App) SaveBinaryFile(data []byte, defaultName string) (string, error) {
	options := runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Save File",
	}
	filepath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	if filepath == "" {
		return "", nil
	}
	err = os.WriteFile(filepath, data, filePermissionUserRW)
	if err != nil {
		return "", err
	}
	return filepath, nil
}
func (a *App) SelectDirectory() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Directory",
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return dir, nil
}
type databaseFileConfig struct {
	Title       string
	DisplayName string
	Pattern     string
	Extensions  map[string]bool
}
var databaseFileConfigs = map[string]databaseFileConfig{
	"Sqlite3": {
		Title:       "Select SQLite Database",
		DisplayName: "SQLite Database Files (*.db,*.sqlite,*.sqlite3,*.db3)",
		Pattern:     "*.db;*.sqlite;*.sqlite3;*.db3",
		Extensions:  map[string]bool{"db": true, "sqlite": true, "sqlite3": true, "db3": true},
	},
	"DuckDB": {
		Title:       "Select DuckDB Database",
		DisplayName: "DuckDB Database Files (*.duckdb,*.db)",
		Pattern:     "*.duckdb;*.db",
		Extensions:  map[string]bool{"duckdb": true, "db": true},
	},
}
func (a *App) SelectDatabaseFile(dbType string) (string, error) {
	cfg, ok := databaseFileConfigs[dbType]
	if !ok {
		return "", fmt.Errorf("unsupported file-based database type: %s", dbType)
	}
	options := runtime.OpenDialogOptions{
		Title: cfg.Title,
		Filters: []runtime.FileFilter{
			{
				DisplayName: cfg.DisplayName,
				Pattern:     cfg.Pattern,
			},
		},
	}
	filepath, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	if filepath == "" {
		return "", nil
	}
	ext := strings.ToLower(filepath)
	dotIndex := strings.LastIndex(ext, ".")
	if dotIndex == -1 || dotIndex == len(ext)-1 {
		return "", fmt.Errorf("invalid file type for %s", dbType)
	}
	ext = ext[dotIndex+1:]
	if !cfg.Extensions[ext] {
		return "", fmt.Errorf("invalid file type for %s", dbType)
	}
	return filepath, nil
}
func (a *App) CopyToClipboard(text string) error {
	return runtime.ClipboardSetText(a.ctx, text)
}
func (a *App) GetFromClipboard() (string, error) {
	return runtime.ClipboardGetText(a.ctx)
}
func (a *App) getConfigOptions() datadir.Options {
	return datadir.Options{
		AppName:           "whodb",
		EnterpriseEdition: a.edition == "ee",
		Development:       false,
	}
}
func (a *App) SaveWindowState() error {
	x, y := runtime.WindowGetPosition(a.ctx)
	width, height := runtime.WindowGetSize(a.ctx)
	maximized := runtime.WindowIsMaximised(a.ctx)
	a.windowSettings = WindowSettings{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Maximized: maximized,
	}
	opts := a.getConfigOptions()
	return config.WriteSection(config.SectionDesktop, a.windowSettings, opts)
}
func (a *App) RestoreWindowState() error {
	opts := a.getConfigOptions()
	if err := config.ReadSection(config.SectionDesktop, &a.windowSettings, opts); err != nil {
		return err
	}
	if a.windowSettings.X != 0 || a.windowSettings.Y != 0 {
		runtime.WindowSetPosition(a.ctx, a.windowSettings.X, a.windowSettings.Y)
	}
	if a.windowSettings.Width > 0 && a.windowSettings.Height > 0 {
		runtime.WindowSetSize(a.ctx, a.windowSettings.Width, a.windowSettings.Height)
	}
	if a.windowSettings.Maximized {
		runtime.WindowMaximise(a.ctx)
	}
	return nil
}
func (a *App) MinimizeWindow() {
	runtime.WindowMinimise(a.ctx)
}
func (a *App) MaximizeWindow() {
	runtime.WindowToggleMaximise(a.ctx)
}
func (a *App) ShowAboutDialog() {
	version := "Community Edition"
	if a.edition == "ee" {
		version = "Enterprise Edition"
	}
	message := fmt.Sprintf(`WhoDB %s
The AI-First Database Management System
Build: Desktop Application
Platform: %s/%s
© 2025 Clidey, Inc.
All rights reserved.
Website: https://whodb.com
GitHub: https://github.com/clidey/whodb
Documentation: https://docs.whodb.com`,
		version,
		goruntime.GOOS,
		goruntime.GOARCH)
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "About WhoDB",
		Message: message,
	})
}
func (a *App) SetupApplicationMenu() {
	appMenu := menu.NewMenu()
	if goruntime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
	}
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("New Connection", keys.CmdOrCtrl("n"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:toggle-sidebar-new-connection")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Export Data", keys.Combo("e", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:export-data")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})
	appMenu.Append(menu.EditMenu())
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Refresh", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:refresh")
	})
	viewMenu.AddText("Toggle Sidebar", keys.CmdOrCtrl("b"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:toggle-sidebar")
	})
	dbMenu := appMenu.AddSubmenu("Database")
	dbMenu.AddText("Execute Query", keys.CmdOrCtrl("return"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:execute-query")
	})
	dbMenu.AddText("New Scratchpad Page", keys.CmdOrCtrl("t"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:new-scratchpad-page")
	})
	dbMenu.AddSeparator()
	dbMenu.AddText("Disconnect", keys.CmdOrCtrl("d"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:disconnect")
	})
	windowMenu := appMenu.AddSubmenu("Window")
	windowMenu.AddText("Minimize", keys.CmdOrCtrl("m"), func(_ *menu.CallbackData) {
		a.MinimizeWindow()
	})
	windowMenu.AddText("Maximize", nil, func(_ *menu.CallbackData) {
		a.MaximizeWindow()
	})
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Documentation", nil, func(_ *menu.CallbackData) {
		a.OpenURL("https://docs.whodb.com")
	})
	helpMenu.AddText("Report Issue", nil, func(_ *menu.CallbackData) {
		a.OpenURL("https://github.com/clidey/whodb/issues")
	})
	helpMenu.AddSeparator()
	aboutText := "About WhoDB"
	if a.edition == "ee" {
		aboutText = "About WhoDB Enterprise"
	}
	helpMenu.AddText(aboutText, nil, func(_ *menu.CallbackData) {
		a.ShowAboutDialog()
	})
	runtime.MenuSetApplicationMenu(a.ctx, appMenu)
}
func (a *App) ShowMessageDialog(title, message string, dialogType string) (string, error) {
	options := runtime.MessageDialogOptions{
		Title:   title,
		Message: message,
	}
	switch dialogType {
	case "info":
		options.Type = runtime.InfoDialog
	case "warning":
		options.Type = runtime.WarningDialog
	case "error":
		options.Type = runtime.ErrorDialog
	case "question":
		options.Type = runtime.QuestionDialog
		options.Buttons = []string{"Yes", "No"}
		options.DefaultButton = "Yes"
		options.CancelButton = "No"
	}
	result, err := runtime.MessageDialog(a.ctx, options)
	return result, err
}
func (a *App) ShowConfirmDialog(title, message string) (bool, error) {
	result, err := a.ShowMessageDialog(title, message, "question")
	return result == "Yes", err
}
func (a *App) SetupSystemTray() {
}
func (a *App) ShowNotification(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Title:   title,
		Message: message,
		Type:    runtime.InfoDialog,
	})
}