package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"

	"upskill-backend/internal/events"
)

func StartUserService(db *sql.DB, writer *kafka.Writer) {
	mux := http.NewServeMux()

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rows, err := db.Query(`SELECT id, first_name, last_name, email, created_at FROM users`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var result []map[string]interface{}
			for rows.Next() {
				var (
					id        int
					firstName string
					lastName  string
					email     string
					createdAt time.Time
				)
				if err := rows.Scan(&id, &firstName, &lastName, &email, &createdAt); err != nil {
					log.Println("[UserService] scan error", err)
					continue
				}
				result = append(result, map[string]interface{}{
					"id":         id,
					"first_name": firstName,
					"last_name":  lastName,
					"email":      email,
					"created_at": createdAt,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/users/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := r.URL.Query().Get("id")
		if userID == "" {
			http.Error(w, "Missing id", http.StatusBadRequest)
			return
		}
		var body struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		query := `UPDATE users SET first_name=$1, last_name=$2 WHERE id=$3`
		if _, err := db.Exec(query, body.FirstName, body.LastName, userID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go events.ProduceEvent(writer, "UserUpdated", fmt.Sprintf("User ID: %s updated", userID))
		w.Write([]byte("User updated"))
	})

	mux.HandleFunc("/users/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := r.URL.Query().Get("id")
		if userID == "" {
			http.Error(w, "Missing id", http.StatusBadRequest)
			return
		}
		if _, err := db.Exec(`DELETE FROM users WHERE id=$1`, userID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go events.ProduceEvent(writer, "UserDeleted", fmt.Sprintf("User ID: %s deleted", userID))
		w.Write([]byte("User deleted"))
	})

	log.Println("[UserService] Запуск на :8082")
	http.ListenAndServe(":8082", mux)
}
