package mq

import (
	"log"

	nsq "github.com/nsqio/go-nsq"
)

const (
	nsqDURL = "localhost:4150"
)

type NsqProducer struct {
	Producer *nsq.Producer
}

func NewProducer() *NsqProducer {
	p, err := nsq.NewProducer(nsqDURL, nsq.NewConfig())
	if err != nil {
		log.Println(err)
		return nil
	}

	p.SetLogger(nil, nsq.LogLevelError)
	// p.Stop()

	return &NsqProducer{Producer: p}
}

func (np *NsqProducer) SendMessageTopic(message []byte) error {
	err := np.Producer.Publish("chatroom", message)
	if err != nil {
		return err
	}

	// np.Producer.Stop()

	return nil
}
