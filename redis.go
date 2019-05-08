package main

import (
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

var uClient *RedisConnection

const (
	redisMemberKey     = "member:"
	chatroomExpireTime = 43200
)

type RedisConnection struct {
	client *redis.Client
}

func InitRedis() error {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
		// Password: "",
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	uClient = &RedisConnection{client: client}

	return nil
}

func NewRedisClient() *RedisConnection {
	return uClient
}

func (conn *RedisConnection) SetChatroom(chatroomID string, token ...string) (string, error) {
	if chatroomID == "" {
		chatroomID = uuid.NewV4().String()
	}

	p := conn.client.Pipeline()
	p.SAdd(redisMemberKey+chatroomID, token)
	p.Expire(redisMemberKey+chatroomID, chatroomExpireTime)
	_, err := p.Exec()

	if err != nil {
		return "", err
	}

	return chatroomID, nil
}

func (conn *RedisConnection) GetMember(chatroomID string) ([]string, error) {
	return conn.client.SMembers(redisMemberKey + chatroomID).Result()
}

func (conn *RedisConnection) RemoveMember(chatroomID string, tokens ...string) error {
	_, err := conn.client.SRem(redisMemberKey+chatroomID, tokens).Result()
	if err != nil {
		return err
	}

	return nil
}

func (conn *RedisConnection) RenewExpireTime(chatroomID string) error {
	_, err := conn.client.Expire(redisMemberKey+chatroomID, chatroomExpireTime).Result()
	if err != nil {
		return err
	}

	return nil
}
