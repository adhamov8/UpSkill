package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/bcrypt"

	"upskill-backend/internal/config"
	"upskill-backend/internal/events"
)

func StartAuthService(db *sql.DB, writer *kafka.Writer, cfg *config.Config) {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Валидация
		if !ValidateName(body.FirstName) || !ValidateName(body.LastName) {
			http.Error(w, "Invalid name", http.StatusBadRequest)
			return
		}
		if !ValidateEmail(body.Email) {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}
		if !ValidatePassword(body.Password) {
			http.Error(w, "Password does not meet requirements", http.StatusBadRequest)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		var newID int
		err = db.QueryRow(`INSERT INTO users (first_name, last_name, email, password_hash)
                           VALUES ($1,$2,$3,$4) RETURNING id`,
			body.FirstName, body.LastName, body.Email, string(hashed),
		).Scan(&newID)
		if err != nil {
			http.Error(w, fmt.Sprintf("DB error: %v", err), http.StatusConflict)
			return
		}

		go events.ProduceEvent(writer, "UserCreated", fmt.Sprintf("UserID=%d", newID))
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Registered user ID=%d\n", newID)
	})

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var (
			userID     int
			storedHash string
		)
		err := db.QueryRow(`SELECT id, password_hash FROM users WHERE email=$1`, body.Email).Scan(&userID, &storedHash)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(body.Password)); err != nil {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		accessToken, err := GenerateJWT(userID, cfg.JwtSecret)
		if err != nil {
			http.Error(w, "JWT error", http.StatusInternalServerError)
			return
		}

		refreshToken := fmt.Sprintf("rf_%d", time.Now().UnixNano())
		refreshExpires := time.Now().Add(7 * 24 * time.Hour)

		_, err = db.Exec(`UPDATE users SET refresh_token=$1, refresh_expires=$2 WHERE id=$3`,
			refreshToken, refreshExpires, userID)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		resp := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		var (
			userID  int
			expires time.Time
		)
		err := db.QueryRow(`SELECT id, refresh_expires FROM users WHERE refresh_token=$1`, body.RefreshToken).Scan(&userID, &expires)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if time.Now().After(expires) {
			http.Error(w, "Refresh token expired", http.StatusUnauthorized)
			return
		}
		newAccess, err := GenerateJWT(userID, cfg.JwtSecret)
		if err != nil {
			http.Error(w, "JWT error", http.StatusInternalServerError)
			return
		}
		resp := map[string]string{
			"access_token": newAccess,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("[AuthService] Запуск на :8081")
	http.ListenAndServe(":8081", mux)
}
