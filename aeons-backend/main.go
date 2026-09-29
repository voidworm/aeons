package main

import (
	"aeons/server"
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.Handle("/ws", server.New())

	log.Fatal(http.ListenAndServe(":8080", mux))

}
