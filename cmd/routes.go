package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yomek33/websocket-go/pkg/websocket"
)

func SetStaticFolder(route *mux.Router) {
	fs := http.FileServer(http.Dir("./public/"))
	route.PathPrefix("/").Handler(http.StripPrefix("/", fs))
}

func Routes(route *mux.Router) {

	log.Println("Loading routes...")
	pool := websocket.NewPool()
	go pool.Start()

	route.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(pool, w, r)
	})
	log.Println("Routes loaded!")

}

func ServeWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Not a websocket request", http.StatusBadRequest)
		return
	}
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
	}

	username := r.URL.Query().Get("username")
	log.Println("username: ", username)
	websocket.CreateNewSocketUser(pool, conn, username)
}
