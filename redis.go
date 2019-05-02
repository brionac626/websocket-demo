package main

import (
	"github.com/go-redis/redis"
)

var uClient *RedisConnection

type RedisConnection struct {
	Client *redis.Client
}

func InitRedis() error {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	uClient = &RedisConnection{Client: client}

	return nil
}
