package config

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDatabase initializes the SQLite database
func InitDatabase() error {
	var err error
	DB, err = sql.Open("sqlite", "./website.db")
	if err != nil {
		return err
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS blogs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS user_playtime (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_name TEXT NOT NULL,
		date DATE NOT NULL DEFAULT CURRENT_DATE,
		last_login DATETIME,
		playtime INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS public_ip_updates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		new_public_ip_address TEXT NOT NULL,
		old_public_ip_address TEXT,
		changed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS log_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message TEXT NOT NULL,
		level TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() {
	if DB != nil {
		DB.Close()
	}
}

func LogMessage(level string, message string) {
	_, err := DB.Exec("INSERT INTO log_messages (level, message) VALUES (?, ?)", level, message)
	if err != nil {
		log.Println("Error logging message:", err)
	}
}