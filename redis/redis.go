package redis

import (
	"github.com/go-redis/redis"
)

var uClient *RedisConnection

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
