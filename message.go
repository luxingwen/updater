package updater

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

const (
	METHOD_REQUEST  = "request"
	METHOD_RESPONSE = "response"
)

type Message struct {
	Id      string          `json:"id"`
	Type    string          `json:"type"`
	Method  string          `json:"method"`
	Data    json.RawMessage `json:"data"`
	Timeout time.Duration   // 添加 Timeout 字段
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
					Extra:   make(map[string]interface{}),
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
