package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 300 * time.Second
	pongWait       = 600 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type SocketEventStruct struct {
	EventName    string      `json:"eventName"`
	EventPayload interface{} `json:"eventPayload"`
}

type UserStruct struct {
	UserID   string `json:"userID"`
	Username string `json:"username"`
}

type JoinDisconnectPayload struct {
	UserID string       `json:"userID"`
	Users  []UserStruct `json:"users"`
}

func CreateNewSocketUser(pool *Pool, connection *websocket.Conn, username string) {
	uniqueID := uuid.New()
	client := &Client{
		userID:   uniqueID.String(),
		username: username,
		Conn:     connection,
		Pool:     pool,
		send:     make(chan SocketEventStruct),
	}

	go client.Write()
	go client.Read()
	log.Println("CreateNewSocketUser called")
	client.Pool.Register <- client
}

func HandleUserRegisterEvent(pool *Pool, client *Client) {
	pool.Clients[client] = true
	handleSocketPayloadEvents(client, SocketEventStruct{
		EventName:    "join",
		EventPayload: client.userID,
	})
}

// HandleUserDisconnectEvent will handle the Disconnect event for socket users
func HandleUserDisconnectEvent(pool *Pool, client *Client) {
	_, ok := pool.Clients[client]
	if ok {
		delete(pool.Clients, client)
		close(client.send)

		handleSocketPayloadEvents(client, SocketEventStruct{
			EventName:    "disconnect",
			EventPayload: client.userID,
		})
	}
}

func handleSocketPayloadEvents(client *Client, socketEventPayload SocketEventStruct) {
	switch socketEventPayload.EventName {
	case "join":
		log.Printf("Join Event triggered")
		BroadcastSocketEventToAllClient(client.Pool, SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: JoinDisconnectPayload{
				UserID: client.userID,
				Users:  getAllConnectedUsers(client.Pool),
			},
		})

	case "disconnect":
		log.Printf("Disconnect Event triggered")
		BroadcastSocketEventToAllClient(client.Pool, SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: JoinDisconnectPayload{
				UserID: client.userID,
				Users:  getAllConnectedUsers(client.Pool),
			},
		})

	case "message":
		// log.Printf("Message Event triggered")
		// selectedUserID := socketEventPayload.EventPayload.(map[string]interface{})["userID"].(string)
		// socketEventResponse.EventName = "message response"
		// socketEventResponse.EventPayload = map[string]interface{}{
		// 	"userID":   selectedUserID,
		// 	"username": getUsernameByUserID(client.Pool, selectedUserID),
		// 	"message":  socketEventPayload.EventPayload.(map[string]interface{})["message"],
		// }
		// EmitToSpecificClient(client.Pool, socketEventResponse, selectedUserID)

	default:
		log.Printf("Unknown event type: %s", socketEventPayload.EventName)
	}
}

func BroadcastSocketEventToAllClient(pool *Pool, payload SocketEventStruct) {
	for client := range pool.Clients {
		select {
		case client.send <- payload:
		default:
			close(client.send)
			delete(pool.Clients, client)
		}
	}
}

func getAllConnectedUsers(pool *Pool) []UserStruct {
	var users []UserStruct
	for singleClient := range pool.Clients {
		users = append(users, UserStruct{
			Username: singleClient.username,
			UserID:   singleClient.userID,
		})
	}
	return users
}

func getUsernameByUserID(pool *Pool, userID string) string {
	for singleClient := range pool.Clients {
		if singleClient.userID == userID {
			return singleClient.username
		}
	}
	return ""
}

func EmitToSpecificClient(pool *Pool, payload SocketEventStruct, userID string) {
	clientsToRemove := make([]*Client, 0)

	for client := range pool.Clients {
		if client.userID == userID {
			select {
			case client.send <- payload:
			default:
				clientsToRemove = append(clientsToRemove, client)
			}
		}
	}

	for _, client := range clientsToRemove {
		close(client.send)
		delete(pool.Clients, client)
	}
}

func handleRecivedMessage(pool *Pool, client *Client, message Message) {
	handleSocketPayloadEvents(client, SocketEventStruct{
		EventName:    "wordsubmission",
		EventPayload: message,
	})
}
