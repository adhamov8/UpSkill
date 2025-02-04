package badge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/segmentio/kafka-go"

	"upskill-backend/internal/events"
)

func StartBadgeService(db *sql.DB, writer *kafka.Writer) {
	mux := http.NewServeMux()

	mux.HandleFunc("/badges", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rows, err := db.Query(`SELECT id, name, description FROM badges`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var badges []map[string]interface{}
			for rows.Next() {
				var (
					id          int
					name        string
					description string
				)
				if err := rows.Scan(&id, &name, &description); err != nil {
					log.Println("[BadgeService] scan error", err)
					continue
				}
				badges = append(badges, map[string]interface{}{
					"id":   id,
					"name": name,
					"desc": description,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(badges)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/badges/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Name string `json:"name"`
			Desc string `json:"desc"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		var badgeID int
		err := db.QueryRow(`INSERT INTO badges (name, description) VALUES ($1, $2) RETURNING id`,
			body.Name, body.Desc,
		).Scan(&badgeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go events.ProduceEvent(writer, "BadgeCreated", fmt.Sprintf("Badge ID: %d", badgeID))
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Badge created with ID=%d\n", badgeID)
	})

	log.Println("[BadgeService] Запуск на :8083")
	http.ListenAndServe(":8083", mux)
}
