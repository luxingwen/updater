package updater

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	Url            string
	Connected      bool
	HostIP         string
	HostName       string
	UUID           string
	Token          string // 新增 Token 字段
	Server         *Server
	OS             string //
	Arch           string //
	messageHandler *MessageHandler
}

type Server struct {
	Url     *url.URL
	Load    int
	Checked bool
	mu      sync.Mutex
}

func NewClient(server *Server, messageHandler *MessageHandler) *Client {
	return &Client{
		Server:         server,
		send:           make(chan []byte),
		messageHandler: messageHandler,
	}
}

func NewServer(urlAddress string) *Server {
	u, err := url.Parse(urlAddress)
	if err != nil {
		log.Fatal(err)
	}
	return &Server{
		Url: u,
	}
}

func (c *Client) connect() error {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(c.Server.Url.String(), nil)
	if err != nil {
		return err
	}
	// 读取服务器的负载信息
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	load, err := strconv.Atoi(string(message))
	if err != nil {
		return err
	}
	// 保存连接和负载信息
	c.conn = conn
	c.Connected = true
	c.Server.mu.Lock()
	c.Server.Load = load
	c.Server.Checked = true
	c.Server.mu.Unlock()
	return nil
}

// 尝试连接到每个服务器，返回连接成功并且负载最低的客户端
func ConnectToServers(servers []*Server, messageHandler *MessageHandler) (*Client, error) {
	var minLoad int
	var minClient *Client
	for _, server := range servers {
		client := NewClient(server, messageHandler)
		err := client.connect()
		if err == nil {
			if minClient == nil || client.Server.Load < minLoad {
				minLoad = client.Server.Load
				minClient = client
			}
		}
	}
	if minClient == nil {
		return nil, fmt.Errorf("failed to connect to any server")
	}
	return minClient, nil
}

// 启动客户端，包括连接到服务器以及启动读写goroutines
func (c *Client) Start() error {
	err := c.connect()
	if err != nil {
		return err
	}
	go c.readPump()
	go c.writePump()
	return nil
}

func (c *Client) Stop() {
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	c.conn.Close()
}

// 从websocket读取消息的goroutine
func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			c.Connected = false
			c.connect()
			continue
		}
		log.Println("recv: ", string(message))
		// 这里你可以添加处理消息的逻辑

		msg := new(Message)
		err = json.Unmarshal(message, msg)
		if err != nil {
			log.Println("json unmarshal error:", err)
			continue
		}
		if c.messageHandler != nil {
			c.messageHandler.SubmitMessage(msg)
		}
	}
}

// 写消息到websocket的goroutine
func (c *Client) writePump() {
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// 如果通道关闭，关闭websocket连接
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-ticker.C:
			// 定时ping服务器以保持连接
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 向服务器发送消息
func (c *Client) SendMessage(msg []byte) {
	if c.Connected {
		c.send <- msg
	}
}
