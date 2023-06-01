package v1

import (
	"encoding/json"
	"errors"
	"log"

	"updater"
)

type AuthController struct {
	handler *updater.MessageHandler
}

func NewAuthController(handler *updater.MessageHandler) *AuthController {
	controller := &AuthController{
		handler: handler,
	}
	controller.registerHandlers()
	return controller
}

func (ac *AuthController) registerHandlers() {
	ac.handler.RegisterHandler("v1/Register", ac.Register)
}

type HeartBeatMsg struct {
	Time int64 `json:"time"`
}

func (ac *AuthController) Register(ctx *updater.Context) error {
	msg := ctx.Message

	if msg.Code != updater.CODE_SUCCESS {
		return errors.New("注册失败:" + msg.Msg)
	}

	var heartBeat HeartBeatMsg

	err := json.Unmarshal(msg.Data, &heartBeat)
	if err != nil {
		return err
	}

	log.Println("注册成功，服务器时间:", heartBeat.Time)
	ctx.Client.Registered = true
	return nil
}
