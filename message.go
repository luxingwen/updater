package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"time"
	log "updater/pkg/logger"

	"go.uber.org/zap"
)

const (
	METHOD_REQUEST  = "request"
	METHOD_RESPONSE = "response"

	CODE_SUCCESS = "success"
	CODE_ERROR   = "error"
	CODE_TIMEOUT = "timeout"
)

type Message struct {
	From    string          `json:"from"`   // 消息发送者
	To      string          `json:"to"`     // 消息接收者
	Id      string          `json:"id"`     // 消息 ID
	Type    string          `json:"type"`   // 消息类型
	Method  string          `json:"method"` // 消息方法
	Data    json.RawMessage `json:"data"`   // 消息数据
	Code    string          `json:"code"`   // 新增 Code 字段
	Msg     string          `json:"msg"`    // 新增 Msg 字段
	TraceId string          `json:"traceId"`
	Timeout time.Duration   // 添加 Timeout 字段
	TaskId  string          `json:"taskId"`
}

type HandlerFunc func(ctx *Context) error

type MessageHandler struct {
	handlers map[string]HandlerFunc
	in       chan *Message
}

func NewMessageHandler(bufferSize int) *MessageHandler {
	return &MessageHandler{
		handlers: make(map[string]HandlerFunc),
		in:       make(chan *Message, bufferSize),
	}
}

func (h *MessageHandler) RegisterHandler(messageType string, handler HandlerFunc) {
	if _, exists := h.handlers[messageType]; exists {
		log.Fatalf("Handler already registered for message type: %s", messageType)
	}

	h.handlers[messageType] = handler
}

func (h *MessageHandler) PrintRegisteredHandlers() {
	fmt.Println("Registered Handlers:")
	for messageType, handler := range h.handlers {
		// 使用反射获取处理程序的名称
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		fmt.Printf("Message Type:%s  %s\n", messageType, handlerName)

	}
	fmt.Println("------------------------")
}

func (h *MessageHandler) HandleMessages(client *Client, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func() {
			// 用于防止 panic 造成的程序崩溃
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in HandleMessages: %v", r)
				}
			}()

			for msg := range h.in {
				ctx := context.Background()
				if msg.Timeout > 0 {
					ctx, _ = context.WithTimeout(ctx, msg.Timeout)
				}

				ctxWithCancel, cancel := context.WithCancel(ctx)

				context := &Context{
					Client:  client,
					Message: msg,
					Ctx:     ctxWithCancel,
					Cancel:  cancel,
					app:     client.app,
					Extra:   make(map[string]interface{}),
					Logger:  client.app.Logger.With(zap.String("traceId", msg.TraceId), zap.String("taskId", msg.TaskId)),
				}

				if handler, ok := h.handlers[msg.Type]; ok {
					err := handler(context)
					if err != nil {
						log.Printf("Error handling message: %s", err)
					}
				} else {
					log.Printf("No handler registered for message type: %s", msg.Type)
				}
			}
		}()
	}
}

func (h *MessageHandler) SubmitMessage(msg *Message) {
	h.in <- msg
}
