package app

import (
	"updater/pkg/config"
	"updater/pkg/logger"
	"updater/pkg/task"
)

type App struct {
	Logger      *logger.Logger
	Config      *config.Config
	TaskManager *task.TaskManager
	TaskStore   *task.TaskStore
}

func NewApp() *App {
	var err error
	app := &App{}
	config.InitConfig()
	app.Config = config.GetConfig()
	app.Logger = logger.NewLogger(app.Config.LogConfig)
	app.TaskManager = task.NewTaskManager()
	app.TaskStore, err = task.NewTaskStore(app.Config.TaskStorePath)
	if err != nil {
		app.Logger.Fatalf("初始化任务存储失败: %s", err)
	}
	return app
}
