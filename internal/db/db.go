package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("[DB] Подключение к PostgreSQL успешно")
	return db, nil
}

func CreateTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            first_name TEXT,
            last_name TEXT,
            email TEXT UNIQUE,
            password_hash TEXT,
            refresh_token TEXT,
            refresh_expires TIMESTAMP,
            created_at TIMESTAMP DEFAULT now()
        )`,
		`CREATE TABLE IF NOT EXISTS badges (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS user_badges (
            user_id INT,
            badge_id INT,
            progress INT DEFAULT 0,
            PRIMARY KEY(user_id, badge_id)
        )`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
