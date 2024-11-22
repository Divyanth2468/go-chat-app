package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/joho/godotenv"
)

var DB *sql.DB

func InitDB() error {
	// if err := godotenv.Load(); err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	log.Println("fvdgbvgs", dbUser, dbPassword, dbHost, dbPort, dbName)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUser, dbPassword, dbHost, dbPort)
	// dsn := `root:chinnichotu@2702@tcp(127.0.0.1:3306)/`
	// Open connection to mysql
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Check if connection is working
	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}

	// if _, err := DB.Exec(`DROP DATABASE IF EXISTS chatapp;`); err != nil {
	// 	log.Println("Error deleting database")
	// }

	log.Println("Database Connection to Mysql server established securely")

	if !doesDBExist("chatapp") {
		err := createDatabase()
		if err != nil {
			log.Fatal("Error creating database ", err)
		}
		log.Println("Database 'chatapp' created successfully!")
	} else {
		log.Println("Database 'chatapp' already exists")
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	// dsn = `root:chinnichotu@2702@tcp(127.0.0.1:3306)/chatapp?parseTime=true`

	// Open connection to chatapp
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error connecting to database 'chatapp' ", err)
	}

	// Check if the connection to the 'chatapp' database is successful
	if err := DB.Ping(); err != nil {
		log.Fatal("Error pinging 'chatapp' database: ", err)
	}
	log.Println("Connected to 'chatapp' database successfully")

	// Create if not exists
	err = createTables()
	if err != nil {
		log.Fatal("Error creating tables ", err)
	}

	return nil
}

func doesDBExist(dbName string) bool {
	var name string
	query := fmt.Sprintf("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'", dbName)
	err := DB.QueryRow(query).Scan(&name)
	if err != nil {
		return err == nil // if query fails, assume the database does not exist
	}
	return true
}

// Create chatapp DB
func createDatabase() error {
	var err error
	// SQL Query to create if not exists
	createDBQuery := "CREATE DATABASE IF NOT EXISTS chatapp"

	_, err = DB.Exec(createDBQuery)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database created successfully")

	return err
}

// Create tables
func createTables() error {
	var err error
	// Create Tables
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id CHAR(36) NOT NULL PRIMARY KEY,   -- UUID as primary key
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			status ENUM('online', 'offline') DEFAULT 'offline',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	// SQL query to create 'messages' table
	createMessagesTable := `
		CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,         -- Auto-increment message ID
			sender_id CHAR(36) NOT NULL,               -- UUID foreign key for sender
			receiver_id CHAR(36) NOT NULL,             -- UUID foreign key for receiver
			message TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (sender_id) REFERENCES users(id),
			FOREIGN KEY (receiver_id) REFERENCES users(id)
		);
	`

	createFriendsTable := `
		CREATE TABLE IF NOT EXISTS friends (
			id INT PRIMARY KEY AUTO_INCREMENT,        -- Auto-increment ID for friend relationships
			user_id CHAR(36) NOT NULL,                -- UUID foreign key for user
			friend_id CHAR(36) NOT NULL,              -- UUID foreign key for friend
			status ENUM('pending', 'accepted', 'rejected') DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE (user_id, friend_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (friend_id) REFERENCES users(id)
		);
	`

	// Execute the queries
	_, err = DB.Exec(createUsersTable)
	if err != nil {
		return err
	}
	log.Println("Created user table")

	_, err = DB.Exec(createMessagesTable)
	if err != nil {
		return err
	}

	log.Println("Created message table")

	_, err = DB.Exec(createFriendsTable)
	if err != nil {
		return err
	}
	log.Println("Created friends table")

	return err
}
