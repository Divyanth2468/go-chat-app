package handlers

import (
	"encoding/json"
	"go-chat-app/models"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type alert struct {
	Alert string
}

type home_data struct {
	UserId   uuid.UUID
	UserName string
	Friends  []models.Users
	Users    []models.Users
	Requests []models.Users
}

var store = sessions.NewCookieStore([]byte("Secret Keys"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./public/login.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	log.Println(email, password)

	// Login User
	user, err := (&models.Users{}).Login(email, password)
	if err != nil {
		al := alert{
			Alert: "Invalid Credentials please try again",
		}
		tmpl.Execute(w, al)
		return
	}

	if saveSession(user, w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
	}
}

func saveSession(user *models.Users, w http.ResponseWriter, r *http.Request) bool {

	session, err := store.Get(r, "sessions")
	if err != nil {
		return false
	}
	session.Values["user_id"] = user.Id.String()
	session.Values["status"] = "online"
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./public/signup.html")
	if err != nil {
		http.Error(w, "Error Rendering Template", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user := &models.Users{
		Name:     name,
		Email:    email,
		Password: password,
	}

	err = user.SaveUsers()
	if err != nil && err.Error() == "ALREADY EXISTS" {
		al := alert{
			Alert: "ALREADY EXISTS",
		}
		if err := tmpl.Execute(w, al); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		return
	}
	user = models.GetUserByEmail(user.Email)

	if saveSession(user, w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get Session and remove user_id
	session, err := store.Get(r, "sessions")
	if err != nil {
		http.Error(w, "Cannot logout", http.StatusInternalServerError)
	}
	if session.Values["user_id"] != nil {
		id := session.Values["user_id"].(string)
		if err := models.Logout(uuid.MustParse(id)); err != nil {
			log.Println(err)
		}
		session.Values["user_id"] = nil
		session.Values = make(map[interface{}]interface{})
	}

	// Optionally, destroy the session
	session.Options.MaxAge = -1

	session.Save(r, w)
	if err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func IsAuthenticated(r *http.Request) (uuid.UUID, bool) {
	session, err := store.Get(r, "sessions")
	if err != nil {
		return uuid.Nil, false
	}
	id, ok := session.Values["user_id"].(string) // Interface to string
	if ok {
		userid := uuid.MustParse(id) // UUID string to uuid
		return userid, ok
	}
	return uuid.Nil, false
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := IsAuthenticated(r)
	// log.Println(id, ok, "Authenticated")
	if ok {
		tmpl, err := template.ParseFiles("./public/home.html")
		if err != nil {
			http.Error(w, "Error rendering home page", http.StatusInternalServerError)
		}
		tmpl.Execute(w, nil)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func UserListHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := IsAuthenticated(r)
	user := models.GetUserById(id)
	friends, err := models.GetFriendsByUserId(id)
	if err != nil {
		log.Println("Error getting friends", err)
	}
	requests, err := models.GetPendingRequestsByUser(id)
	allusers := models.GetAllUsers(id, requests, friends)
	if err != nil {
		log.Println("Error getting pending requests", err)
	}

	// Convert friends data to JSON format
	// log.Println(friends)
	data := home_data{
		UserId:   user.Id,
		UserName: user.Name,
		Friends:  friends,
		Users:    allusers,
		Requests: requests,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetMessagesHandler handles the request to get messages between two users
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	senderId := r.URL.Query().Get("senderId")
	friendId := r.URL.Query().Get("friendId")

	// Convert to UUID
	senderUUID, err := uuid.Parse(senderId)
	if err != nil {
		http.Error(w, "Invalid sender ID", http.StatusBadRequest)
		return
	}

	friendUUID, err := uuid.Parse(friendId)
	if err != nil {
		http.Error(w, "Invalid friend ID", http.StatusBadRequest)
		return
	}

	// Fetch messages between the two users
	messages, err := models.GetMessagesBetweenUsers(senderUUID, friendUUID)
	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}
	// for _, message := range messages {
	// 	log.Println(message.SenderId)
	// }

	// Convert messages to JSON and send as response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Error encoding messages", http.StatusInternalServerError)
	}
}
