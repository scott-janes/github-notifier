package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	isMock := false
	for _, arg := range os.Args[1:] {
		if arg == "--mock" || arg == "-mock" {
			isMock = true
			break
		}
	}

	app := NewApp(isMock)

	err := wails.Run(&options.App{
		Title:     "GitHub Notifications",
		Width:     60,
		Height:    60,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Frameless:       true,
		DisableResize:   true,
		AlwaysOnTop:     true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:       app.startup,
		OnShutdown:      app.shutdown,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
