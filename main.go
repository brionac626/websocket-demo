package main

import (
	"log"
	"net/http"

	// _ "net/http/pprof"
	"websocket_demo/redis"
)

func main() {
	if err := redis.InitRedis(); err != nil {
		log.Fatalln(err)
	}
	http.HandleFunc("/ws", wsHandle)
	// go http.ListenAndServe(":12345", nil)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
