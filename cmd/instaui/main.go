package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// createMenu creates the application menu
func createMenu(ctx context.Context) *menu.Menu {
	appMenu := menu.NewMenu()

	// File menu
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		// Quit the application
	})

	// View menu
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Reload", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		// Reload the application
	})

	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("About", nil, func(_ *menu.CallbackData) {
		// Show about dialog
	})

	return appMenu
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "Insta-Infra UI",
		Width:             1200,
		Height:            800,
		MinWidth:          800,
		MinHeight:         600,
		MaxWidth:          1920,
		MaxHeight:         1080,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Menu:             createMenu(context.Background()),
		OnStartup:        app.startup,
		LogLevel:         logger.ERROR,
		Bind: []interface{}{
			app,
		},
		// Windows specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// Hide console window
			DisableFramelessWindowDecorations: false,
			ResizeDebounceMS:                  10,
		},
		// macOS specific options
		Mac: &mac.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Insta-Infra UI",
				Message: "A modern UI for managing data infrastructure services",
			},
		},
		// Linux specific options
		Linux: &linux.Options{
			Icon: nil, // Can add an icon here
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
