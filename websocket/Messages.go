package websocket

import (
	"encoding/json"
	"go-chat-app/database"
	"go-chat-app/models"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type SocketMessage struct {
	ToFriendId  uuid.UUID
	Mes         models.Message
	UserList    []models.Users
	FriendList  []models.Users
	RequestList []models.Users
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type FriendRequestMessage struct {
	UserID    string `json:"user_id"`
	ReqUserID string `json:"req_user_id"`
	Action    string `json:"action"`
}

type Client struct {
	id   uuid.UUID
	conn *websocket.Conn
	send chan SocketMessage // Message struct channel
}

var clients = make(map[uuid.UUID]*Client)         // Use map to associate client id with websocket
var broadcast = make(chan SocketMessage)          // Broadcasting Message structs
var updates = make(map[uuid.UUID]*websocket.Conn) // Update Users

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

func FriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Extract user IDs from query params
	userID := r.URL.Query().Get("user_id")
	log.Println("WebSocket connection established for user:", userID)

	id, err := uuid.Parse(userID)
	if err != nil {
		log.Println("Error pasrsing user id", err)
	}
	updates[id] = conn

	defer func() {
		delete(updates, id)
	}()

	for {
		// Reading incoming requests
		_, message, err := conn.ReadMessage()
		if err != nil {
			// Handle WebSocket close error gracefully
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("WebSocket closed for user %s: %v", userID, err)
			} else {
				log.Printf("Error reading WebSocket message for user %s: %v", userID, err)
			}
			break // Exit the loop on error
		}
		if len(message) == 0 {
			log.Println("Received empty WebSocket message; skipping.")
			continue
		}
		var req FriendRequestMessage
		if err := json.Unmarshal(message, &req); err != nil {
			log.Println("Failed to parse WebSocket message:", err)
			continue
		}

		// Log received request
		log.Println("Received Friend Request:", req)

		// Process the friend request
		handleFriendRequest(req.UserID, req.ReqUserID, req.Action)

		// Notify users in real-time after processing the request
		notifyFriendUpdate(req.UserID, req.ReqUserID)
	}

}

func handleFriendRequest(user_id, req_user_id string, action string) {

	// log.Println("User ID:", user_id, "Requested User ID:", req_user_id, "Action:", action)

	var query string
	if action == "" {
		// Send friend request
		query = "INSERT INTO friends (user_id, friend_id) VALUES (?, ?)"
		if _, err := database.DB.Exec(query, req_user_id, user_id); err != nil {
			log.Println("failed to update friends list", err)
		}
		return
	} else if action == "accept" {
		// Accept friend request
		query = `UPDATE friends SET status = 'accepted' WHERE user_id = ? AND friend_id = ?`
	} else {
		// Reject friend request
		query = `DELETE FROM friends WHERE user_id = ? AND friend_id = ?`
	}

	// Execute the query
	if _, err := database.DB.Exec(query, user_id, req_user_id); err != nil {
		log.Println("failed to update friends list", err)
	}
}

// Notify connected users about updates
func notifyFriendUpdate(user1ID, user2ID string) {
	// Simulate fetching the updated lists for both users
	var pendingList1, pendingList2, friendList1, friendList2, userList1, userList2 []models.Users
	var err error
	id1, err := uuid.Parse(user1ID)
	if err != nil {
		log.Println("Error parsing UUID for user1:", err)
	}
	id2, err := uuid.Parse(user2ID)
	if err != nil {
		log.Println("Error parsing UUID for user2:", err)
	}

	pendingList1, err = models.GetPendingRequestsByUser(id1)
	if err != nil {
		log.Println("Error getting pending requests", err)
	}
	pendingList2, err = models.GetPendingRequestsByUser(id2)
	if err != nil {
		log.Println("Error getting pending requests", err)
	}

	friendList1, err = models.GetFriendsByUserId(id1)
	if err != nil {
		log.Println("Error getting friends", err)
	}
	friendList2, err = models.GetFriendsByUserId(id2)
	if err != nil {
		log.Println("Error getting friends", err)
	}

	userList1 = models.GetAllUsers(id1, pendingList1, friendList1)
	userList2 = models.GetAllUsers(id2, pendingList2, friendList2)

	// Notify user1 if connected
	if conn, exists := updates[id1]; exists {
		notifyUser(conn, user1ID, pendingList1, friendList1, userList1)
	}

	// Notify user2 if connected
	if conn, exists := updates[id2]; exists {
		notifyUser(conn, user2ID, pendingList2, friendList2, userList2)
	}
}

// Helper function to notify a user
func notifyUser(conn *websocket.Conn, userID string, pendingList, friendList, userList []models.Users) {

	// Send updates lists to the user
	message := SocketMessage{
		ToFriendId:  uuid.MustParse(userID),
		RequestList: pendingList,
		FriendList:  friendList,
		UserList:    userList,
	}

	// Send notification via WebSocket
	if err := conn.WriteJSON(message); err != nil {
		log.Println("Failed to notify user:", userID, err)
	}
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
	http.HandleFunc("/ws-friend-request", FriendRequestHandler)
	go handleMessages()
}
