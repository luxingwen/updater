package updater

import (
	"context"
	"encoding/json"
)

type Context struct {
	Client  *Client
	Message *Message
	Extra   map[string]interface{}
	Ctx     context.Context
	Cancel  context.CancelFunc
}

func (ctx *Context) ShouldBindJSON(req interface{}) (err error) {
	b, err := ctx.Message.Data.MarshalJSON()
	if err != nil {
		return
	}
	err = json.Unmarshal(b, req)
	return
}

func (ctx *Context) Abort() {
	ctx.Cancel()
}

func (ctx *Context) JSONSuccess(req interface{}) {
	ctx.SendResponse(CODE_SUCCESS, "ok", req)
}

func (ctx *Context) JSONError(code string, msg string) {
	ctx.SendResponse(code, msg, nil)
}

func (ctx *Context) SendResponse(code string, msg string, resp interface{}) {
	ctx.Message.Method = METHOD_RESPONSE
	ctx.Message.Code = code
	ctx.Message.Msg = msg
	if resp != nil {
		ctx.Message.Data, _ = json.Marshal(resp)
	}

	ctx.Client.SendMessage(ctx.Message)
}

func (ctx *Context) SendRequest(req interface{}) {
	ctx.Message.Method = METHOD_REQUEST
	ctx.Message.Data, _ = json.Marshal(req)
	ctx.Client.SendMessage(ctx.Message)
}
