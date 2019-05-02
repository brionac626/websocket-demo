package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Minute,
	WriteBufferSize:  4096,
	ReadBufferSize:   4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var allConn sync.Map

type WebsocketClient struct {
	lock   *sync.Mutex
	token  string
	wsConn *websocket.Conn
	data   chan []byte
}

func wsHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	c := NewWsClient(conn, token)
	go c.ReadMessage()
	go c.ProcessMessage()
	// go ShowServerStatus()
}

func NewWsClient(conn *websocket.Conn, token string) *WebsocketClient {
	client := WebsocketClient{
		lock:   &sync.Mutex{},
		token:  token,
		wsConn: conn,
		data:   make(chan []byte, 3000),
	}

	allConn.Store(token, client)

	return &client
}

func (c *WebsocketClient) WriteMessage(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.wsConn.WriteMessage(websocket.TextMessage, data)
}

func (c *WebsocketClient) ReadMessage() {
	defer allConn.Delete(c.token)
	for {
		_, data, err := c.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if err := c.CloseClient(); err != nil {
					log.Println(err)
				}
				break
			}
			log.Println(err)
			continue
		}

		c.data <- data
	}
}

type ChatroomData struct {
	Chatroom string `json:"chatroom"`
	Message  string `json:"message,omitempty"`
}

func (c *WebsocketClient) ProcessMessage() {
	for {
		select {
		case data, ok := <-c.data:
			if !ok {
				// log.Println("channel not ready")
				return
			}
			fmt.Println(string(data))

			results := gjson.GetManyBytes(data, "action", "data.chatroomID", "data.message")

			switch results[0].Str {
			case "create":
			case "join":
			case "chat":
			default:
			}

			if err := c.WriteMessage(data); err != nil {
				log.Println(err)
			}
		default:
		}
	}
}

func (c *WebsocketClient) CloseClient() error {
	close(c.data)
	return c.wsConn.Close()
}

func (c *WebsocketClient) CreateChatroom() error {
	return nil
}

func (c *WebsocketClient) PushChatroomMember() error {
	return nil
}

func (c *WebsocketClient) GetChatroomMemberToken() error {
	return nil
}

func SendMessage(token string, message []byte) error {
	if client, ok := allConn.Load(token); ok {
		if err := client.(*WebsocketClient).WriteMessage(message); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func ShowServerStatus() {
	t := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-t.C:
			fmt.Println(allConn)
		}
	}
}
