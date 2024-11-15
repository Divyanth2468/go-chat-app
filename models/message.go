package models

import (
	"fmt"
	"go-chat-app/database"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
)

// Message struct as per message table

type Message struct {
	Id         int
	SenderId   uuid.UUID
	ReceiverId uuid.UUID
	Message    string
	TimeStamp  time.Time
}

type Mes struct {
	SenderId   uuid.UUID
	ReceiverId uuid.UUID
	Message    string
	TimeStamp  time.Time
}

func (m *Message) SaveMessages() error {

	if database.DB == nil {
		return fmt.Errorf("databse connection is nil")
	}

	query := "INSERT INTO messages (sender_id, receiver_id, message, created_at) VALUES (?, ?, ?, ?)"

	// Preparing the query statement for exec
	stmt, err := database.DB.Prepare(query)
	if err != nil {
		log.Fatal("Error preparing the query: ", err)
		return err
	}

	if _, err = stmt.Exec(m.SenderId, m.ReceiverId, m.Message, m.TimeStamp); err != nil {
		log.Fatal("Error executing the query: ", err)
		return err
	}

	return nil
}

func GetMessagesByUserId(userId uuid.UUID) ([]Message, error) {
	query := "SELECT id, sender_id, receiver_id, message, created_at FROM messages WHERE sender_id= ? OR receiver_id= ?"
	rows, err := database.DB.Query(query, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Id, &msg.SenderId, &msg.ReceiverId, &msg.Message, &msg.TimeStamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func GetMessagesBetweenUsers(senderId, receiverId uuid.UUID) ([]Mes, error) {
	query := "SELECT sender_id, receiver_id, message, created_at FROM messages WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)"
	rows, err := database.DB.Query(query, senderId, receiverId, receiverId, senderId)
	if err != nil {
		return nil, err
	}

	var messages []Mes
	for rows.Next() {
		var msg Mes
		err := rows.Scan(&msg.SenderId, &msg.ReceiverId, &msg.Message, &msg.TimeStamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Sort the messages by Timestamp using sort.Slice
	sort.Slice(messages, func(i, j int) bool {
		// Sort in ascending order by Timestamp
		return messages[i].TimeStamp.Before(messages[j].TimeStamp)
	})
	return messages, nil
}

func (m *Message) DeleteMessages() error {
	query := "DELETE FROM messages WHERE id = ?"
	_, err := database.DB.Exec(query, m.Id)
	return err
}
