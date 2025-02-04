package recommend

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

func StartRecommendService(db *sql.DB, writer *kafka.Writer) {
	mux := http.NewServeMux()

	mux.HandleFunc("/recommendations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var body struct {
				UserID int `json:"user_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			plan := fmt.Sprintf("Plan for user %d: Study Go, get badges!", body.UserID)

			_, err := db.Exec(`INSERT INTO recommendations (user_id, plan) VALUES ($1, $2)`, body.UserID, plan)
			if err != nil {
				http.Error(w, "DB error", http.StatusInternalServerError)
				return
			}
			go events.ProduceEvent(writer, "RecommendationCreated", fmt.Sprintf("User %d has new plan", body.UserID))
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]string{"plan": plan}
			json.NewEncoder(w).Encode(resp)

		case http.MethodGet:
			rows, err := db.Query(`SELECT user_id, plan, created_at FROM recommendations`)
			if err != nil {
				http.Error(w, "DB error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var result []map[string]interface{}
			for rows.Next() {
				var userID int
				var plan string
				var createdAt time.Time
				if err := rows.Scan(&userID, &plan, &createdAt); err != nil {
					log.Println("[RecommendationService] scan error", err)
					continue
				}
				result = append(result, map[string]interface{}{
					"user_id":    userID,
					"plan":       plan,
					"created_at": createdAt,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("[RecommendationService] Запуск на :8084")
	http.ListenAndServe(":8084", mux)
}
