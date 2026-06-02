package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// Version is set at build time via -ldflags "-X main.Version=v1.2.3".
// Defaults to "dev" for local builds.
var Version = "dev"

func main() {
	app := NewApp(Version)

	err := wails.Run(&options.App{
		Title:            "smb-tools",
		Width:            1920,
		Height:           1080,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.logoAssetHandler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
