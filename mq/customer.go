package mq

import (
	"errors"
	"fmt"
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

	c.SetLogger(nil, 0)
	c.AddHandler(&MessageHandler{MessageChan: messageData})

	if err := c.ConnectToNSQLookupd(nsqLookUpURL); err != nil {
		log.Println(err)
		return
	}

}

type MessageHandler struct {
	MessageChan chan []byte
}

func (h *MessageHandler) HandleMessage(msg *nsq.Message) error {
	if msg.Body == nil {
		return errors.New("no message data")
	}
	fmt.Println("123", string(msg.Body))
	sendMQData(h.MessageChan, msg.Body)

	return nil
}

func sendMQData(bc chan []byte, data []byte) {
	if _, ok := <-bc; !ok {
		return
	}

	bc <- data
}

func GetMQData(bc chan []byte) []byte {
	if _, ok := <-bc; !ok {
		return nil
	}

	return <-bc
}