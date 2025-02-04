package main

import (
	"fmt"
	"log"
	"net/http"

	"upskill-backend/internal/config"
	"upskill-backend/internal/db"
	"upskill-backend/internal/events"
	"upskill-backend/services/auth"
	"upskill-backend/services/badge"
	"upskill-backend/services/user"
)

func main() {
	cfg := config.NewConfig()

	database, err := db.InitDB(cfg.PgConnStr)
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}

	if err := db.CreateTables(database); err != nil {
		log.Fatalf("Ошибка при создании таблиц: %v", err)
	}

	kafkaWriter := events.InitKafkaWriter(cfg.KafkaAddr, cfg.TopicName)
	defer kafkaWriter.Close()

	go auth.StartAuthService(database, kafkaWriter)
	go user.StartUserService(database, kafkaWriter)
	go badge.StartBadgeService(database, kafkaWriter)

	log.Println("[Gateway] Слушаем на :8080...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "UpSkill Gateway with Kafka.")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
