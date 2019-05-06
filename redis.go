package main

import (
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

var uClient *RedisConnection

const redisMemberKey = "member:"

type RedisConnection struct {
	Client *redis.Client
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

	uClient = &RedisConnection{Client: client}

	return nil
}

func NewRedisClient() *RedisConnection {
	return uClient
}

func (conn *RedisConnection) SetChatroom(chatroomID string, token ...string) (string, error) {
	if chatroomID == "" {
		chatroomID = uuid.NewV4().String()
	}
	_, err := conn.Client.SAdd(redisMemberKey+chatroomID, token).Result()
	if err != nil {
		return "", err
	}

	return chatroomID, nil
}

func (conn *RedisConnection) GetMember(chatroomID string) ([]string, error) {
	return conn.Client.SMembers(redisMemberKey + chatroomID).Result()
}
