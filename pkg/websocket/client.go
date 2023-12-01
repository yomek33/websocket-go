package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	userID   string
	username string
	Conn     *websocket.Conn
	Pool     *Pool
	mu       sync.Mutex
	send     chan SocketEventStruct
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type ReceivedMessage struct {
	Username string `json:"username"`
	Hiragana string `json:"hiragana"`
}

func (c *Client) Read() {
	// クライアントが切断されたら、登録を解除する
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// クライアントからのメッセージを読み込む
	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Message from %s, : %s", c.username, string(p))

		message := Message{Type: messageType, Body: string(p)}
		c.Pool.Broadcast <- message

	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			// サーバーが切断されたら、クライアントも切断する
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("error getting next writer: %v", err)
				return
			}

			if err := json.NewEncoder(w).Encode(message); err != nil {
				log.Printf("error writing message: %v", err)
				w.Close()
				return
			}

			// バッファに溜まっているメッセージを送信する
			n := len(c.send)
			for i := 0; i < n; i++ {
				msg := <-c.send
				if err := json.NewEncoder(w).Encode(msg); err != nil {
					log.Printf("error writing additional message: %v", err)
					break
				}
			}

			if err := w.Close(); err != nil {
				log.Printf("error closing writer: %v", err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("error sending ping: %v", err)
				return
			}
		}
	}
}
