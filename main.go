package main

import (
	"embed"
	"log/slog"

	"github.com/chenyang-zz/boxify/internal/service"
	"github.com/chenyang-zz/boxify/internal/window"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	// 新的窗口事件
	application.RegisterEvent[map[string]interface{}]("window:opened")
	application.RegisterEvent[map[string]interface{}]("window:closed")
}

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	am := window.InitApplication(slog.LevelInfo, assets)

	// 注册服务
	services := []func(app *application.App) application.Service{
		func(app *application.App) application.Service {
			return application.NewService(service.NewService(app))
		},
		func(app *application.App) application.Service {
			return application.NewService(service.NewWindowService(am))
		},
	}

	am.RegisterService(services...)

	// 运行应用程序
	err := am.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
