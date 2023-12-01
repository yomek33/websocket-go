package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Client connected to endpoint")
	return conn, nil
}
