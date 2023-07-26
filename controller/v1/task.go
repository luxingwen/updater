package v1

import (
	"updater"
	"updater/models"
	"updater/pkg/task"
)

type TaskController struct {
	handler *updater.MessageHandler
}

func NewTaskController(handler *updater.MessageHandler) *TaskController {
	controller := &TaskController{
		handler: handler,
	}
	controller.registerHandlers()
	return controller
}

func (tc *TaskController) registerHandlers() {
	tc.handler.RegisterHandler("v1/GetTaskInfo", tc.handleGetTaskInfo)
	tc.handler.RegisterHandler("v1/GetTaskInfo/Response", tc.handleGetTaskInfoResponse)
}

func (tc *TaskController) handleGetTaskInfo(ctx *updater.Context) error {
	var req models.ReqTaskInfo
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}

	tinfo, err := ctx.App().TaskManager.GetTask(req.TaskID)
	if err != nil {

		tinfo, err = ctx.App().TaskStore.GetTask(req.TaskID)
		if err != nil {
			return err
		}
	}

	if tinfo == nil {
		tinfo, err = ctx.App().TaskStore.GetTask(req.TaskID)
		if err != nil {
			return err
		}
	}

	ctx.JSONSuccess(tinfo.GetResult())

	return nil
}

func (tc *TaskController) handleGetTaskInfoResponse(ctx *updater.Context) error {
	var req models.ReqTaskInfo
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}

	tinfo, err := ctx.App().TaskManager.GetTask(req.TaskID)
	if err != nil {
		return err
	}

	if tinfo == nil {
		return nil
	}

	if tinfo.GetStatus() == task.TaskStatusCompleted || tinfo.GetStatus() == task.TaskStatusFailed {
		ctx.App().TaskManager.RemoveTask(req.TaskID)
		ctx.App().TaskStore.RemoveTask(req.TaskID)
	}

	return nil
}
