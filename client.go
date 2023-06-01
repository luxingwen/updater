package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	Url            string
	UUID           string
	Connected      bool
	Registered     bool // 是否已经注册
	HostIP         string
	LocalIPs       string
	HostName       string
	Vmuuid         string
	Token          string // 新增 Token 字段
	Server         *Server
	OS             string //
	Arch           string //
	Version        string
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
		send:           make(chan []byte, 4096),
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
	if c.UUID == "" {
		c.setUUID()
	}
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(c.Server.Url.String()+c.UUID, nil)
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
	log.Printf("Connected to %s, load: %d", c.Server.Url.String(), load)
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
	for {
		err := c.connect()
		if err != nil {
			log.Println("connect to:", c.Server.Url.String()+c.UUID, " error:", err)
			log.Println("retry after 5 seconds")
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	go c.readPump()
	go c.writePump()

	for {
		c.ClientRegister()
		log.Println("registering...")
		time.Sleep(time.Second * 5)
		if c.Registered {
			log.Println("register success")
			break
		}
		log.Println("register failed, retry after 5 seconds")
		time.Sleep(time.Second * 5)
	}

	return nil
}

func (c *Client) Stop() {
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	c.conn.Close()
}

func (c *Client) setInitClientInfo() {
	c.setVmUuid()
	c.setHostName()
	c.setUUID()
	c.OS = runtime.GOOS
	c.Arch = runtime.GOARCH
	c.Version = Version

	var err error
	c.LocalIPs, err = GetLocalIPs()

	if err != nil {
		log.Println("get local ips error:", err)
	}

	return
}

func (c *Client) ClientRegister() {
	c.setInitClientInfo()
	clientinfo := c.getClientInfo()
	clientinfo.LocalIPs = c.LocalIPs

	data, err := json.Marshal(clientinfo)
	if err != nil {
		log.Println("marshal client info error:", err)
		return
	}

	msg := &Message{
		Type:   "Register",
		Id:     uuid.New().String(),
		Data:   json.RawMessage(data),
		Method: METHOD_REQUEST,
	}

	c.SendMessage(msg)

}

func (c *Client) setVmUuid() {
	c.Vmuuid = getVmuuid()
	return
}

func (c *Client) setHostName() {
	c.HostName = getHostName()
	return
}

func (c *Client) setUUID() {
	if c.UUID == "" {
		// Check if the UUID file exists
		if _, err := os.Stat("uuid.txt"); err == nil {
			// Read the UUID from the file
			data, err := ioutil.ReadFile("uuid.txt")
			if err != nil {
				log.Println("Failed to read UUID from file:", err)
			} else {
				c.UUID = string(data)
				return
			}
		} else {
			// Generate a new UUID
			c.UUID = uuid.New().String()
			// Write the UUID to the file
			err := ioutil.WriteFile("uuid.txt", []byte(c.UUID), 0644)
			if err != nil {
				log.Println("Failed to write UUID to file:", err)
			}
		}
	}
	return
}

// 从websocket读取消息的goroutine
func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			c.Connected = false
			time.Sleep(5 * time.Second)
			log.Println("reconnecting to server...", c.Server.Url.String()+c.UUID)
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
				log.Println("send channel closed")
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
			log.Println("send heartbeat to server...")
			c.Heartbeat()
		}
	}
}

func (c *Client) getClientInfo() (r *ClientInfo) {
	r = &ClientInfo{
		UUID:      c.UUID,
		HostIP:    c.HostIP,
		HostName:  c.HostName,
		Vmuuid:    c.Vmuuid,
		OS:        c.OS,
		Arch:      c.Arch,
		Heartbeat: time.Now().Unix(),
	}
	return
}

func (c *Client) Heartbeat() {
	clientInfo := c.getClientInfo()
	b, _ := json.Marshal(clientInfo)
	msg := &Message{
		Id:     uuid.New().String(),
		Type:   "Heartbeat",
		Data:   json.RawMessage(b),
		Method: METHOD_REQUEST,
	}

	c.SendMessage(msg)

}

type ClientInfo struct {
	UUID      string `json:"uuid"`
	HostIP    string `json:"hostIp"`
	HostName  string `json:"hostName""`
	Vmuuid    string `json:"vmuuid"`
	Sn        string `json:"sn"`       // 序列号
	OS        string `json:"os"`       //
	Arch      string `json:"arch"`     //
	Heartbeat int64  `json:"hearbeat"` // 心跳时间
	LocalIPs  string `json:"localIps"` // 本地IP地址
}

// 向服务器发送消息
func (c *Client) SendMessage(msg *Message) {
	msg.From = c.UUID
	if msg.Id == "" {
		msg.Id = uuid.New().String()
	}

	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal message error:", err)
		return
	}
	c.Send(b)
}

func (c *Client) Send(msg []byte) {
	log.Println("send: ", string(msg))
	if c.Connected {
		log.Println("send message to server...")
		c.send <- msg
	}
}
