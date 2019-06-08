package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"websocket-demo/mq"
	"websocket-demo/redis"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

const secretKey = "demoService"

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Minute,
	WriteBufferSize:  4096,
	ReadBufferSize:   4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	allConn      sync.Map
	customerData chan []byte
	producer     *mq.NsqProducer
)

type WebsocketClient struct {
	lock   *sync.Mutex
	token  string
	wsConn *websocket.Conn
	data   chan []byte
	rc     *redis.RedisConnection
}

func wsHandle(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if !tokenVerify(tokenString) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := NewWsClient(conn, tokenString)
	var initCustomer sync.Once
	initCustomer.Do(func() {
		customerData = make(chan []byte, 3000)
		producer = mq.NewProducer()
		go mq.NewCustomer(customerData)
	})
	go c.ReadMessage()
	go c.ProcessMessage()
	go c.ReadChatroomMessage()
	// go ShowServerStatus()
}

func NewWsClient(conn *websocket.Conn, token string) *WebsocketClient {
	client := &WebsocketClient{
		lock:   &sync.Mutex{},
		token:  token,
		wsConn: conn,
		data:   make(chan []byte, 3000),
		rc:     redis.NewRedisClient(),
	}

	allConn.Store(token, client)

	return client
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
	Action   string `json:"action"`
	Chatroom string `json:"chatroom,omitempty"`
	Message  string `json:"message,omitempty"`
}

func (c *WebsocketClient) ProcessMessage() {
	for {
		select {
		case data := <-c.data:
			// fmt.Println(string(data))
			results := gjson.GetManyBytes(data, "action", "data.chatroom", "data.message")

			switch results[0].Str {
			case "create":
				resp := make([]byte, 0)
				chatroomID := c.CreateChatroom()

				if chatroomID != "" {
					resp = RespSuccessCreateChatroom(chatroomID)
				} else {
					resp = RespInternalError(999)
				}

				if resp != nil {
					if err := c.WriteMessage(resp); err != nil {
						log.Println(err)
						break
					}
				}
			case "join":
				err := c.PushChatroomMember(results[1].Str, []string{c.token})
				if err != nil {
					log.Println(err)
					continue
				}
				if err := c.WriteMessage(RespChatroomMessage(results[1].Str, "join chatroom succed")); err != nil {
					log.Println(err)
				}
			case "chat":
				err := producer.SendMessageTopic(data)
				if err != nil {
					log.Println(err)
				}
			default:
				if err := c.WriteMessage(RespInternalError(19999)); err != nil {
					log.Println(err)
				}
			}
		}

	}
}

func (c *WebsocketClient) CloseClient() error {
	close(c.data)
	return c.wsConn.Close()
}

func (c *WebsocketClient) CreateChatroom() string {
	if c.rc == nil {
		log.Println("redis client didn't init")
		return ""
	}

	chatroomID, err := c.rc.SetChatroom("", c.token)
	if err != nil {
		log.Println(err)
		return ""
	}

	return chatroomID
}

func (c *WebsocketClient) PushChatroomMember(chatroomID string, token []string) error {
	_, err := c.rc.SetChatroom(chatroomID, token...)
	if err != nil {
		return err
	}

	return nil
}

func (c *WebsocketClient) GetChatroomMemberToken(chatroomID string) ([]string, error) {
	return c.rc.GetMember(chatroomID)
}

func (c *WebsocketClient) ReadChatroomMessage() {
	for {
		select {
		case data := <-customerData:
			// fmt.Println(string(data))
			if data != nil {
				results := gjson.GetManyBytes(data, "action", "data.chatroom", "data.message")

				tokens, err := c.GetChatroomMemberToken(results[1].Str)
				if err != nil {
					log.Println(err)
					if err := c.WriteMessage(RespInternalError(2000)); err != nil {
						log.Println(err)
					}
				}

				for _, token := range tokens {
					client, ok := allConn.Load(token)
					if !ok {
						continue
					}
					go func() {
						err := client.(*WebsocketClient).WriteMessage(RespChatroomMessage(results[1].Str, results[2].Str))
						if err != nil {
							log.Println(err)
							return
						}
					}()
				}

				if err := redis.NewRedisClient().RenewExpireTime(results[1].Str); err != nil {
					log.Println(err)
				}
			}
		}
	}
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

func tokenVerify(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Println(err)
		return false
	}

	return token.Valid
}

func RespSuccessCreateChatroom(chatroomID string) []byte {
	resp, err := json.Marshal(&ChatroomData{Action: "1", Chatroom: chatroomID})
	if err != nil {
		log.Println(err)
		return RespInternalError(5000)
	}

	return resp
}

func RespChatroomMessage(chatroomID, message string) []byte {
	resp, err := json.Marshal(&ChatroomData{Action: "3", Chatroom: chatroomID, Message: message})
	if err != nil {
		log.Println(err)
		return RespInternalError(5000)
	}

	return resp
}

func RespInternalError(code int) []byte {
	resp, err := json.Marshal(&ChatroomData{Action: strconv.Itoa(code)})
	if err != nil {
		log.Println(err)
		return nil
	}

	return resp
}
