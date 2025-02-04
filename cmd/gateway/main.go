package main

import (
	"fmt"
	"log"
	"net/http"

	"upskill-backend/internal/config"
	"upskill-backend/internal/db"
	"upskill-backend/internal/events"
	"upskill-backend/service/auth"
	"upskill-backend/service/badge"
	"upskill-backend/service/recommend"
	"upskill-backend/service/user"
)

func main() {
	cfg := config.GetConfig()

	database, err := db.InitDB(cfg.PgConnStr)
	if err != nil {
		log.Fatalf("[DB] Ошибка: %v", err)
	}
	if err := database.Ping(); err != nil {
		log.Fatalf("[DB] Не удаётся связаться: %v", err)
	}
	log.Println("[DB] Подключение к PostgreSQL успешно")

	if err := db.CreateTables(database); err != nil {
		log.Fatalf("[DB] Ошибка миграции: %v", err)
	}

	writer := events.NewKafkaWriter(cfg.KafkaAddr, cfg.KafkaTopic)
	defer writer.Close()

	go auth.StartAuthService(database, writer, cfg)
	go user.StartUserService(database, writer)
	go badge.StartBadgeService(database, writer)
	go recommend.StartRecommendService(database, writer)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "UpSkill Gateway with 'service/' folder.")
	})

	log.Println("[Gateway] видим на :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
