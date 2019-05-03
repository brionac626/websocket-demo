package main

import (
	"log"
	"net/http"
)

func main() {
	if err := InitRedis(); err != nil {
		log.Fatalln(err)
	}
	http.HandleFunc("/ws", wsHandle)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
