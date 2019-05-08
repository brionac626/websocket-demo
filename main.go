package main

import (
	"log"
	"net/http"

	"websocket_demo/redis"
)

func main() {
	if err := redis.InitRedis(); err != nil {
		log.Fatalln(err)
	}
	http.HandleFunc("/ws", wsHandle)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
