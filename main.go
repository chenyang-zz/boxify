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

	// 数据同步事件
	application.RegisterEvent[service.DataSyncEvent]("data-sync:broadcast")
	application.RegisterEvent[service.DataSyncEvent]("data-sync:targeted")

	// 初始数据事件
	application.RegisterEvent[service.InitialDataEntry]("initial-data:received")
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
		func(app *application.App) application.Service {
			return application.NewService(service.NewDataSyncService(app, am.GetRegistry()))
		},
		func(app *application.App) application.Service {
			return application.NewService(service.NewInitialDataService(am))
		},
	}

	am.RegisterService(services...)

	// 运行应用程序
	err := am.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
