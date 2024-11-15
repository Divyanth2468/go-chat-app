package websocket

import (
	"encoding/json"
	"go-chat-app/models"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type SocketMessage struct {
	ToFriendId uuid.UUID
	Mes        models.Message
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	id   uuid.UUID
	conn *websocket.Conn
	send chan SocketMessage // Message struct channel
}

var clients = make(map[uuid.UUID]*Client) // Use map to associate client id with websocket
var broadcast = make(chan SocketMessage)  // Broadcasting Message structs

// Broadcast to clients
func handleMessages() {
	for {
		msg := <-broadcast // Get the message to broadcast
		// Send message to the client with the matching `ToFriendId`
		for _, client := range clients {
			// log.Println("Broadcasting", client.id, msg.ToFriendId)
			select {
			case client.send <- msg: // Sending Message struct
			default:
				// If message can't be sent, clean up the client connection
				close(client.send)
				delete(clients, client.id)
			}

		}
	}
}

// Handle each client's WebSocket connection
func handleClientConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get `senderId` from URL params and convert it to uuid.UUID
	senderIdStr := r.URL.Query().Get("senderId")
	senderId, err := uuid.Parse(senderIdStr)
	if err != nil {
		log.Println(err)
		return
	}

	// Get `friendId` from URL params and convert it to uuid.UUID
	friendIDStr := r.URL.Query().Get("friendId")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		log.Println("Invalid friendId:", err)
		return
	}

	client := &Client{id: friendID, conn: conn, send: make(chan SocketMessage)}
	clients[friendID] = client
	log.Println("New client connected:", client.id.String())

	// Start reading messages from client in a separate goroutine
	go client.writeMessages()
	client.readMessages(senderId, friendID)
}

// Client message reading
func (c *Client) readMessages(senderId uuid.UUID, friendId uuid.UUID) {
	defer func() {
		delete(clients, c.id)
		c.conn.Close()
	}()
	for {
		// Read messages from WebSocket
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		message := models.Message{
			SenderId:   senderId,
			ReceiverId: c.id,
			Message:    string(msg),
			TimeStamp:  time.Now().UTC(),
		}
		message.SaveMessages() // Assuming this saves the message to a DB
		// Broadcast the message to the client
		broadcast <- SocketMessage{
			ToFriendId: friendId, // Sending the message to the specific friend
			Mes:        message,
		}
	}
}

// Client message writing
func (c *Client) writeMessages() {
	for msg := range c.send {
		// Marshal Message struct to []byte before sending
		msg, err := json.Marshal(msg)
		if err != nil {
			log.Println("Failed to marshal message", err)
			return
		}
		msgBytes := []byte(msg)
		err = c.conn.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func SocketConnection() {
	http.HandleFunc("/ws", handleClientConnection)
	go handleMessages()
}
