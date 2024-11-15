package models

import (
	"database/sql"
	"errors"
	"go-chat-app/database"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	Id        uuid.UUID
	Name      string
	Email     string
	Password  string
	Status    string
	CreatedAt time.Time
}

// Save new Users
func (u *Users) SaveUsers() error {

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error hashing password: ", err)
		return err
	}

	// query to insert users
	query := "INSERT INTO users (id, name, email, password, status) VALUES (?, ?, ?, ?, ?)"

	// preparing statemnt
	stmt, err := database.DB.Prepare(query)
	if err != nil {
		log.Fatal("Error preparing statement: ", err)
	}

	//Executing statemtn
	if _, err := stmt.Exec(uuid.New().String(), strings.ToLower(u.Name), strings.ToLower(u.Email), hashPassword, "online"); err != nil {
		log.Println(err)
		return errors.New("ALREADY EXISTS")
	}

	return nil
}

// User Auth func
func (u *Users) Login(email, password string) (*Users, error) {
	query := "SELECT id, name, email, password, status FROM users WHERE email = ?"
	row := database.DB.QueryRow(query, strings.ToLower(email))
	var user Users
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Status)
	if err != nil {
		log.Println("Error preparing statement: ", err)
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	// update statsus if needed
	query = "UPDATE users SET status = 'online' WHERE id = ?"
	_, err = database.DB.Exec(query, user.Id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Filter by Id
func GetUserById(id uuid.UUID) *Users {
	query := "SELECT id, name, email, password, status FROM users WHERE id = ?"
	row := database.DB.QueryRow(query, id)
	var user Users
	if err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Status); err != nil {
		if err == sql.ErrNoRows {
			// Handle case where no rows are found
			log.Println("User not found.")
			return nil // or return an appropriate error
		}
	}
	return &user
}

func GetFriendsByUserId(userId uuid.UUID) []Users {
	query := "SELECT id, name, status FROM users WHERE id != ?"
	rows, err := database.DB.Query(query, userId)
	if err != nil {
		log.Println("Error fetching friends", err)
		return nil
	}
	defer rows.Close()

	var users []Users
	for rows.Next() {
		var user Users
		err := rows.Scan(&user.Id, &user.Name, &user.Status)
		if err != nil {
			log.Println(err)
		}
		users = append(users, user)
	}
	return users
}

// Filter by email
func GetUserByEmail(email string) *Users {
	query := "SELECT id, name, email, password, status FROM users WHERE email = ?"
	row := database.DB.QueryRow(query, strings.ToLower(email))
	var user Users
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Status)
	if err != nil {
		log.Println("Error getting user: ", err)
		return nil
	}
	return &user
}

// Update Credentials
func (u *Users) UpdateUser() error {
	query := "UPDATE users SET name = ?, email = ?, password = ? WHERE id = ?"
	_, err := database.DB.Exec(query, strings.ToLower(u.Name), strings.ToLower(u.Email), u.Password, u.Id)

	return err
}

// Delete users
func (u *Users) DeleteUser() error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := database.DB.Exec(query, u.Id)
	return err
}

func Logout(userId uuid.UUID) error {
	query := "UPDATE users SET status = 'offline' WHERE id = ?"
	_, err := database.DB.Exec(query, userId)
	return err
}
