package main

import (
	"fmt"
	"go-chat-app/database"
	"go-chat-app/handlers"
	"go-chat-app/websocket"
	"log"
	"net/http"
	"os"
)

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true // Allow connections from any origin
// 	},
// }

func routeHandling(port string) {
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/signup", handlers.SignUpHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/api/messages", handlers.GetMessagesHandler)
	http.HandleFunc("/api/userlist", handlers.UserListHandler)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	websocket.SocketConnection()
	fmt.Println("Server up and running")
	http.ListenAndServe(":"+port, nil)
}

func main() {

	// // Initialize databse connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	port := os.Getenv("PORT")
	routeHandling(port)
}
