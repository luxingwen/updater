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

func (ctx *Context) SendResponse(resp interface{}) {
	ctx.Message.Method = METHOD_RESPONSE
	ctx.Message.Data, _ = json.Marshal(resp)
	b, _ := json.Marshal(ctx.Message)
	ctx.Client.SendMessage(b)
}

func (ctx *Context) SendRequest(req interface{}) {
	ctx.Message.Method = METHOD_REQUEST
	ctx.Message.Data, _ = json.Marshal(req)
	b, _ := json.Marshal(ctx.Message)
	ctx.Client.SendMessage(b)
}
