package mq

import (
	"errors"
	"log"
	"time"

	nsq "github.com/nsqio/go-nsq"
)

const (
	nsqLookUpURL = "localhost:4161"
)

func NewCustomer(messageData chan []byte) {
	config := nsq.NewConfig()
	config.LookupdPollInterval = 1 * time.Second
	c, err := nsq.NewConsumer("chatroom", "message", config)
	if err != nil {
		log.Println(err)
		return
	}

	c.SetLogger(nil, nsq.LogLevelError)
	c.ChangeMaxInFlight(100)
	c.AddConcurrentHandlers(&MessageHandler{MessageChan: messageData}, 100)

	if err := c.ConnectToNSQLookupd(nsqLookUpURL); err != nil {
		log.Println(err)
		return
	}

	for {
	}
}

type MessageHandler struct {
	MessageChan chan []byte
}

func (h *MessageHandler) HandleMessage(msg *nsq.Message) error {
	if msg.Body == nil {
		return errors.New("no message data")
	}
	h.MessageChan <- msg.Body

	return nil
}
