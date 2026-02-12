package main

import (
	"embed"

	"github.com/chenyang-zz/boxify/internal/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	// Create application with options
	app := application.New(application.Options{
		Name: "Boxify",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Boxify",
		Width:  1024,
		Height: 768,
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				AppearsTransparent: true,
				FullSizeContent:    true,
			},
		},
	})

	// 注册服务
	app.RegisterService(application.NewService(service.NewService(app)))

	err := app.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
