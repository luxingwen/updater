package v1

import (
	"encoding/json"
	"updater"
)

type ScriptController struct {
	handler *updater.MessageHandler
}

func NewScriptController(handler *updater.MessageHandler) *ScriptController {
	controller := &ScriptController{
		handler: handler,
	}
	controller.registerHandlers()
	return controller
}

func (sc *ScriptController) registerHandlers() {
	sc.handler.RegisterHandler("v1/ExecuteScript", sc.handleExecuteScript)
}

func (sc *ScriptController) handleExecuteScript(ctx *updater.Context) error {
	var req updater.ScriptTaskRequest
	if err := json.Unmarshal(ctx.Message.Data, &req); err != nil {
		return err
	}

	scriptTask := updater.NewScriptTask(&req)

	if err := scriptTask.Run(); err != nil {
		ctx.JSONError(updater.CODE_ERROR, err.Error())
		return err
	}
	ctx.JSONSuccess(scriptTask)
	return nil
}